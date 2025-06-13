#!/bin/zsh

# =============================================================================
# DATABASE SETUP SCRIPT
# =============================================================================
# Initializes databases and creates required schemas/users for all services
# Supports MariaDB, MySQL, PostgreSQL, and MongoDB

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_setup_mariadb=false
opt_setup_mysql=false
opt_setup_postgres=false
opt_setup_mongodb=false
opt_setup_all=false
opt_create_test_data=false
opt_timeout=120

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Initializes databases and creates required schemas, users, and configurations
    for all microservices. Supports MariaDB, MySQL, PostgreSQL, and MongoDB.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -m, --mariadb           Setup MariaDB databases
    -y, --mysql             Setup MySQL databases  
    -p, --postgres          Setup PostgreSQL databases
    -g, --mongodb           Setup MongoDB databases
    -a, --all               Setup all available databases
    -t, --test-data         Create test data in databases
    -T, --timeout SECONDS   Database connection timeout (default: 120)
    -h, --help              Show this help message

DATABASES CREATED:
    MariaDB/MySQL:
        - npm (Nginx Proxy Manager)
        - matomo (Analytics)
        - keycloak (Authentication)
        
    PostgreSQL:
        - metabase (Analytics)
        - nocodb (No-code database)
        - keycloak (Authentication alternative)
        - odoo (ERP system)
        
    MongoDB:
        - admin (Administrative)
        - langflow (AI workflows)
        - n8n (Automation)

EXAMPLES:
    $0 -a                           # Setup all databases
    $0 -m -p                        # Setup MariaDB and PostgreSQL only
    $0 -s -e -a                     # Use sudo, show errors, setup all
    $0 --postgres --test-data       # Setup PostgreSQL with test data
    $0 --all --timeout 180          # Setup all with 3-minute timeout

NOTES:
    - Databases must be running before executing this script
    - Script will wait for database containers to be ready
    - Test data is useful for development environments
    - Use appropriate timeout for slower systems
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--sudo)
                opt_use_sudo=true
                shift
                ;;
            -e|--errors)
                opt_show_errors=true
                shift
                ;;
            -m|--mariadb)
                opt_setup_mariadb=true
                shift
                ;;
            -y|--mysql)
                opt_setup_mysql=true
                shift
                ;;
            -p|--postgres)
                opt_setup_postgres=true
                shift
                ;;
            -g|--mongodb)
                opt_setup_mongodb=true
                shift
                ;;
            -a|--all)
                opt_setup_all=true
                opt_setup_mariadb=true
                opt_setup_mysql=true
                opt_setup_postgres=true
                opt_setup_mongodb=true
                shift
                ;;
            -t|--test-data)
                opt_create_test_data=true
                shift
                ;;
            -T|--timeout)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_timeout="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric value"
                fi
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1. Use -h for help."
                ;;
        esac
    done
    
    # If no specific databases selected, show usage
    if [[ "$opt_setup_all" == "false" && 
          "$opt_setup_mariadb" == "false" && 
          "$opt_setup_mysql" == "false" && 
          "$opt_setup_postgres" == "false" && 
          "$opt_setup_mongodb" == "false" ]]; then
        print_status "error" "No databases specified for setup"
        show_usage
        exit 1
    fi
}

# =============================================================================
# DATABASE VALIDATION FUNCTIONS
# =============================================================================

wait_for_database() {
    local db_type="$1"
    local container_name="$2"
    local timeout="$opt_timeout"
    local counter=0
    
    print_status "step" "Waiting for $db_type ($container_name) to be ready..."
    
    while [[ $counter -lt $timeout ]]; do
        case "$db_type" in
            "mariadb"|"mysql")
                # Try direct connection with root user instead of healthcheck
                if eval "$CONTAINER_CMD exec $container_name mariadb -u root -p\$MYSQL_ROOT_PASSWORD -e 'SELECT 1;' >/dev/null 2>&1"; then
                    print_status "success" "$db_type is ready!"
                    return 0
                fi
                
                # Fallback: Check if mysqladmin ping works
                if eval "$CONTAINER_CMD exec $container_name mysqladmin ping -h localhost -u root -p\$MYSQL_ROOT_PASSWORD >/dev/null 2>&1"; then
                    print_status "success" "$db_type is ready!"
                    return 0
                fi
                ;;
            "postgres")
                # More robust postgres check
                if eval "$CONTAINER_CMD exec $container_name pg_isready -U postgres >/dev/null 2>&1"; then
                    # Also verify we can actually connect
                    if eval "$CONTAINER_CMD exec $container_name psql -U postgres -c 'SELECT 1;' >/dev/null 2>&1"; then
                        print_status "success" "$db_type is ready!"
                        return 0
                    fi
                fi
                ;;
            "mongodb")
                if eval "$CONTAINER_CMD exec $container_name mongosh --quiet --eval 'db.adminCommand(\"ping\")' >/dev/null 2>&1"; then
                    print_status "success" "$db_type is ready!"
                    return 0
                fi
                ;;
        esac
        
        # Check if container is still running
        if ! eval "$CONTAINER_CMD ps --format '{{.Names}}' | grep -q '^$container_name$' 2>/dev/null"; then
            handle_error "$container_name container is not running"
        fi
        
        print_status "info" "$db_type not ready yet... ($counter/$timeout seconds)"
        sleep 3  # Increased sleep time
        counter=$((counter + 3))
    done
    
    handle_error "$db_type connection timeout after $timeout seconds"
}

validate_database_containers() {
    print_status "step" "Validating database containers..."
    
    local -a required_containers=()
    
    [[ "$opt_setup_mariadb" == "true" ]] && required_containers+="mariadb"
    [[ "$opt_setup_mysql" == "true" ]] && required_containers+="mysql"
    [[ "$opt_setup_postgres" == "true" ]] && required_containers+="postgres"
    [[ "$opt_setup_mongodb" == "true" ]] && required_containers+="mongodb"
    
    for container in "${required_containers[@]}"; do
        if ! eval "$CONTAINER_CMD ps --format '{{.Names}}' | grep -q '^$container$' $ERROR_REDIRECT"; then
            handle_error "$container container is not running. Please start database services first."
        fi
    done
    
    print_status "success" "All required database containers are running"
}

# =============================================================================
# DATABASE SETUP FUNCTIONS
# =============================================================================

setup_mariadb() {
    print_status "step" "Setting up MariaDB databases and users..."
    
    wait_for_database "mariadb" "mariadb"
    
    # Enhanced MariaDB setup with better conflict handling
    local mariadb_setup_sql="
-- Create databases with IF NOT EXISTS to avoid conflicts
CREATE DATABASE IF NOT EXISTS npm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS matomo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS keycloak CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS microservices CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Drop and recreate users to ensure clean state
DROP USER IF EXISTS 'npm_user'@'%';
DROP USER IF EXISTS 'matomo_user'@'%';
DROP USER IF EXISTS 'keycloak_user'@'%';
DROP USER IF EXISTS 'app_user'@'%';

-- Create users
CREATE USER 'npm_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
CREATE USER 'matomo_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
CREATE USER 'keycloak_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
CREATE USER 'app_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';

-- Grant privileges
GRANT ALL PRIVILEGES ON npm.* TO 'npm_user'@'%';
GRANT ALL PRIVILEGES ON matomo.* TO 'matomo_user'@'%';
GRANT ALL PRIVILEGES ON keycloak.* TO 'keycloak_user'@'%';
GRANT ALL PRIVILEGES ON microservices.* TO 'app_user'@'%';

FLUSH PRIVILEGES;
"
    
    # Use mariadb command with better error handling
    if eval "$CONTAINER_CMD exec -i mariadb mariadb -u root -p\$MYSQL_ROOT_PASSWORD 2>/dev/null << 'EOF'
$mariadb_setup_sql
EOF"; then
        print_status "success" "MariaDB databases and users created successfully"
        
        # Create test data if requested
        if [[ "$opt_create_test_data" == "true" ]]; then
            create_mariadb_test_data
        fi
    else
        print_status "error" "Failed to setup MariaDB databases"
        print_status "info" "Attempting to show MariaDB error details..."
        # Show what went wrong
        eval "$CONTAINER_CMD exec mariadb mariadb -u root -p\$MYSQL_ROOT_PASSWORD -e 'SHOW DATABASES;' 2>&1" || true
        return 1
    fi
}

setup_mysql() {
    print_status "step" "Setting up MySQL databases and users..."
    
    wait_for_database "mysql" "mysql"
    
    # Create databases and users for MySQL
    local mysql_setup_sql="
-- WordPress Database
CREATE DATABASE IF NOT EXISTS wordpress CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'wordpress_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
GRANT ALL PRIVILEGES ON wordpress.* TO 'wordpress_user'@'%';

-- General MySQL application database
CREATE DATABASE IF NOT EXISTS mysql_apps CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'mysql_app_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
GRANT ALL PRIVILEGES ON mysql_apps.* TO 'mysql_app_user'@'%';

-- Laravel/PHP application database
CREATE DATABASE IF NOT EXISTS laravel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'laravel_user'@'%' IDENTIFIED BY '${ADMIN_PASSWORD}';
GRANT ALL PRIVILEGES ON laravel.* TO 'laravel_user'@'%';

FLUSH PRIVILEGES;
"
    
    if eval "$CONTAINER_CMD exec -i mysql mysql -u root -p\$MYSQL_ROOT_PASSWORD << 'EOF'
$mysql_setup_sql
EOF" $ERROR_REDIRECT; then
        print_status "success" "MySQL databases and users created successfully"
        
        # Create test data if requested
        if [[ "$opt_create_test_data" == "true" ]]; then
            create_mysql_test_data
        fi
    else
        handle_error "Failed to setup MySQL databases"
    fi
}

setup_postgres() {
    print_status "step" "Setting up PostgreSQL databases and users..."
    
    wait_for_database "postgres" "postgres"
    
    # Enhanced PostgreSQL setup with conflict resolution
    local postgres_setup_sql="
-- Terminate existing connections to databases we might need to recreate
SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
WHERE datname IN ('metabase', 'nocodb', 'keycloak', 'odoo') AND pid <> pg_backend_pid();

-- Wait a moment for connections to close
SELECT pg_sleep(2);

-- Drop existing databases if they exist (CASCADE to handle dependencies)
DROP DATABASE IF EXISTS metabase;
DROP DATABASE IF EXISTS nocodb; 
DROP DATABASE IF EXISTS keycloak;
DROP DATABASE IF EXISTS odoo;

-- Drop existing users if they exist
DROP USER IF EXISTS metabase_user;
DROP USER IF EXISTS nocodb_user;
DROP USER IF EXISTS keycloak_user;
DROP USER IF EXISTS odoo_user;

-- Create users first
CREATE USER metabase_user WITH PASSWORD '${ADMIN_PASSWORD}';
CREATE USER nocodb_user WITH PASSWORD '${ADMIN_PASSWORD}';
CREATE USER keycloak_user WITH PASSWORD '${ADMIN_PASSWORD}';
CREATE USER odoo_user WITH PASSWORD '${ADMIN_PASSWORD}';

-- Create databases
CREATE DATABASE metabase OWNER metabase_user;
CREATE DATABASE nocodb OWNER nocodb_user;
CREATE DATABASE keycloak OWNER keycloak_user;
CREATE DATABASE odoo OWNER odoo_user;

-- Grant additional privileges
ALTER USER odoo_user CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE metabase TO metabase_user;
GRANT ALL PRIVILEGES ON DATABASE nocodb TO nocodb_user;
GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak_user;
GRANT ALL PRIVILEGES ON DATABASE odoo TO odoo_user;
"
    
    if eval "$CONTAINER_CMD exec -i postgres psql -U postgres 2>/dev/null << 'EOF'
$postgres_setup_sql
EOF"; then
        print_status "success" "PostgreSQL databases and users created successfully"
        
        # Create test data if requested
        if [[ "$opt_create_test_data" == "true" ]]; then
            create_postgres_test_data
        fi
    else
        print_status "error" "Failed to setup PostgreSQL databases"
        print_status "info" "Attempting to show PostgreSQL error details..."
        # Show current state
        eval "$CONTAINER_CMD exec postgres psql -U postgres -c '\l' 2>&1" || true
        return 1
    fi
}

setup_mongodb() {
    print_status "step" "Setting up MongoDB databases and collections..."
    
    wait_for_database "mongodb" "mongodb"
    
    # Enhanced MongoDB setup
    local mongodb_setup_js="
// Switch to admin database for user management
use admin;

// Remove existing users if they exist
try { db.dropUser('langflow_user'); } catch(e) { print('langflow_user did not exist'); }
try { db.dropUser('n8n_user'); } catch(e) { print('n8n_user did not exist'); }
try { db.dropUser('app_user'); } catch(e) { print('app_user did not exist'); }

// Create application-specific users
db.createUser({
    user: 'langflow_user',
    pwd: '${ADMIN_PASSWORD}',
    roles: [
        { role: 'readWrite', db: 'langflow' },
        { role: 'dbAdmin', db: 'langflow' }
    ]
});

db.createUser({
    user: 'n8n_user', 
    pwd: '${ADMIN_PASSWORD}',
    roles: [
        { role: 'readWrite', db: 'n8n' },
        { role: 'dbAdmin', db: 'n8n' }
    ]
});

db.createUser({
    user: 'app_user',
    pwd: '${ADMIN_PASSWORD}',
    roles: [
        { role: 'readWrite', db: 'microservices' },
        { role: 'dbAdmin', db: 'microservices' }
    ]
});

// Create and initialize databases
use langflow;
db.flows.deleteMany({});  // Clear existing data
db.flows.insertOne({ name: 'sample_flow', created_at: new Date() });

use n8n;
db.workflows.deleteMany({});  // Clear existing data
db.workflows.insertOne({ name: 'sample_workflow', created_at: new Date() });

use microservices;
db.config.deleteMany({});  // Clear existing data
db.config.insertOne({ service: 'mongodb', initialized: true, created_at: new Date() });

print('MongoDB databases and users created successfully');
"
    
    if eval "$CONTAINER_CMD exec -i mongodb mongosh --quiet 2>/dev/null << 'EOF'
$mongodb_setup_js
EOF"; then
        print_status "success" "MongoDB databases and users created successfully"
        
        # Create test data if requested
        if [[ "$opt_create_test_data" == "true" ]]; then
            create_mongodb_test_data
        fi
    else
        print_status "error" "Failed to setup MongoDB databases"
        print_status "info" "Attempting to show MongoDB error details..."
        # Show current state
        eval "$CONTAINER_CMD exec mongodb mongosh --eval 'show dbs' 2>&1" || true
        return 1
    fi
}

# =============================================================================
# TEST DATA CREATION FUNCTIONS
# =============================================================================

create_mariadb_test_data() {
    print_status "step" "Creating MariaDB test data..."
    
    local test_data_sql="
USE microservices;

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (username, email) VALUES 
('admin', 'admin@site.test'),
('testuser', 'test@site.test'),
('developer', 'dev@site.test')
ON DUPLICATE KEY UPDATE username=VALUES(username);

CREATE TABLE IF NOT EXISTS services (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    status ENUM('running', 'stopped', 'error') DEFAULT 'running',
    port INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO services (name, status, port) VALUES
('nginx', 'running', 80),
('postgres', 'running', 5432),
('redis', 'running', 6379)
ON DUPLICATE KEY UPDATE name=VALUES(name);
"
    
    eval "$CONTAINER_CMD exec -i mariadb mariadb -u root -p\$MYSQL_ROOT_PASSWORD << 'EOF'
$test_data_sql
EOF" $ERROR_REDIRECT
    
    print_status "success" "MariaDB test data created"
}

create_mysql_test_data() {
    print_status "step" "Creating MySQL test data..."
    
    local test_data_sql="
USE mysql_apps;

CREATE TABLE IF NOT EXISTS applications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    framework VARCHAR(50),
    language VARCHAR(30),
    status ENUM('active', 'inactive', 'maintenance') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO applications (name, framework, language, status) VALUES
('API Gateway', 'Express', 'Node.js', 'active'),
('User Service', 'FastAPI', 'Python', 'active'),
('Analytics Dashboard', 'Laravel', 'PHP', 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name);
"
    
    eval "$CONTAINER_CMD exec -i mysql -u root -p\$MYSQL_ROOT_PASSWORD << 'EOF'
$test_data_sql
EOF" $ERROR_REDIRECT
    
    print_status "success" "MySQL test data created"
}

create_postgres_test_data() {
    print_status "step" "Creating PostgreSQL test data..."
    
    local test_data_sql="
\\c microservices;

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    technology VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO projects (name, description, technology, status) VALUES
('Microservices Platform', 'Docker-based microservices architecture', 'Docker', 'active'),
('Analytics Engine', 'Real-time data processing and visualization', 'Python', 'active'),
('API Gateway', 'Centralized API management and routing', 'Go', 'active')
ON CONFLICT (name) DO NOTHING;
"
    
    eval "$CONTAINER_CMD exec -i postgres psql -U postgres << 'EOF'
$test_data_sql
EOF" $ERROR_REDIRECT
    
    print_status "success" "PostgreSQL test data created"
}

create_mongodb_test_data() {
    print_status "step" "Creating MongoDB test data..."
    
    local test_data_js="
use microservices;

db.projects.insertMany([
    {
        name: 'Langflow AI Platform',
        type: 'ai-workflow',
        status: 'active',
        technologies: ['Python', 'LangChain', 'React'],
        created_at: new Date()
    },
    {
        name: 'n8n Automation Hub', 
        type: 'automation',
        status: 'active',
        technologies: ['Node.js', 'Vue.js', 'TypeScript'],
        created_at: new Date()
    },
    {
        name: 'Data Analytics Pipeline',
        type: 'analytics',
        status: 'active', 
        technologies: ['Python', 'Pandas', 'MongoDB'],
        created_at: new Date()
    }
]);

print('MongoDB test data created successfully');
"
    
    eval "$CONTAINER_CMD exec -i mongodb mongosh --quiet << 'EOF'
$test_data_js
EOF" $ERROR_REDIRECT
    
    print_status "success" "MongoDB test data created"
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_setup_summary() {
    print_status "info" "Database Setup Summary:"
    echo "  Container Runtime: $CONTAINER_RUNTIME"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Show Errors: $opt_show_errors"
    echo "  Connection Timeout: ${opt_timeout}s"
    echo "  Create Test Data: $opt_create_test_data"
    echo ""
    echo "  Databases to setup:"
    [[ "$opt_setup_mariadb" == "true" ]] && echo "    ✓ MariaDB"
    [[ "$opt_setup_mysql" == "true" ]] && echo "    ✓ MySQL"
    [[ "$opt_setup_postgres" == "true" ]] && echo "    ✓ PostgreSQL"
    [[ "$opt_setup_mongodb" == "true" ]] && echo "    ✓ MongoDB"
    echo ""
}

run_database_setup() {
    print_status "step" "Starting database setup process..."
    
    # Validate containers are running
    validate_database_containers
    
    # Setup each requested database type
    [[ "$opt_setup_mariadb" == "true" ]] && setup_mariadb
    [[ "$opt_setup_mysql" == "true" ]] && setup_mysql
    [[ "$opt_setup_postgres" == "true" ]] && setup_postgres
    [[ "$opt_setup_mongodb" == "true" ]] && setup_mongodb
    
    print_status "success" "Database setup completed successfully!"
}

main() {
    # Parse command line arguments (keeping existing argument parsing)
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Show setup summary
    show_setup_summary
    
    print_status "step" "Starting enhanced database setup process..."
    
    # Validate containers are running
    validate_database_containers
    
    # Setup each requested database type with enhanced error handling
    local setup_failed=0
    
    if [[ "$opt_setup_mariadb" == "true" ]]; then
        if ! setup_mariadb; then
            ((setup_failed++))
        fi
    fi
    
    if [[ "$opt_setup_mysql" == "true" ]]; then
        if ! setup_mysql; then
            ((setup_failed++))
        fi
    fi
    
    if [[ "$opt_setup_postgres" == "true" ]]; then
        if ! setup_postgres; then
            ((setup_failed++))
        fi
    fi
    
    if [[ "$opt_setup_mongodb" == "true" ]]; then
        if ! setup_mongodb; then
            ((setup_failed++))
        fi
    fi
    
    if [[ $setup_failed -eq 0 ]]; then
        print_status "success" "All database operations completed successfully!"
    else
        print_status "warning" "$setup_failed database setup(s) encountered issues"
        print_status "info" "You can re-run the setup script to retry failed operations"
    fi
    
    # Show connection information
    echo ""
    print_status "info" "Database Connection Information:"
    echo "  Default credentials: admin / $ADMIN_PASSWORD"
    echo "  MariaDB: localhost:8501"
    echo "  PostgreSQL: localhost:8502"
    echo "  MongoDB: localhost:8503"
    echo "  MySQL: localhost:8505"
    echo "  Redis: localhost:8504"
    echo ""
    print_status "info" "If you encountered errors, you can:"
    echo "  1. Check container logs: podman logs [container_name]"
    echo "  2. Re-run setup: $0 -a"
    echo "  3. Connect manually to debug: podman exec -it [container] [mysql/psql/mongosh]"
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi