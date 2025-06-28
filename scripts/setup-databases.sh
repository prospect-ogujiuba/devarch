#!/bin/zsh

# =============================================================================
# DATABASE SETUP SCRIPT - REFACTORED
# =============================================================================
# Initializes databases and creates required schemas/users for all services
# Now leverages config.sh and service-manager for better integration

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
opt_timeout=120

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Initializes databases and creates required schemas, users, and configurations
    for all microservices. Leverages config.sh and service-manager integration.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -m, --mariadb           Setup MariaDB databases
    -y, --mysql             Setup MySQL databases  
    -p, --postgres          Setup PostgreSQL databases
    -g, --mongodb           Setup MongoDB databases
    -a, --all               Setup all available databases
    -T, --timeout SECONDS   Database connection timeout (default: 120)
    -h, --help              Show this help message

DATABASES CREATED:
    MariaDB/MySQL:
        - npm (Nginx Proxy Manager)
        - matomo (Analytics)
        
    PostgreSQL:
        - metabase (Analytics)
        - nocodb (No-code database)
        
    MongoDB:
        - admin (Administrative)
        - Additional collections as needed

EXAMPLES:
    $0 -a                           # Setup all databases
    $0 -m -p                        # Setup MariaDB and PostgreSQL only
    $0 -s -e -a                     # Use sudo, show errors, setup all
    $0 --all --timeout 180          # Setup all with 3-minute timeout

NOTES:
    - Uses service discovery to detect available database services
    - Leverages config.sh utilities for container operations
    - Automatically detects which databases are actually running
    - Will skip setup for databases that aren't installed/running
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
# DATABASE DETECTION & VALIDATION
# =============================================================================

detect_available_databases() {
    local -a available_databases
    
    print_status "step" "Detecting available database services..."
    
    # Check each database service using config.sh functions
    local database_services=("mariadb" "mysql" "postgres" "mongodb")
    
    for db_service in "${database_services[@]}"; do
        if validate_service_exists "$db_service"; then
            local status=$(get_service_status "$db_service")
            if [[ "$status" == "running" ]]; then
                available_databases+=("$db_service")
                print_status "info" "✓ $db_service is running and available"
            else
                print_status "warning" "⚠️  $db_service exists but is not running (status: $status)"
            fi
        else
            print_status "info" "ℹ️  $db_service service not found (not installed)"
        fi
    done
    
    if [[ ${#available_databases[@]} -eq 0 ]]; then
        handle_error "No database services are currently running. Start database services first."
    fi
    
    print_status "success" "Found ${#available_databases[@]} running database service(s): ${available_databases[*]}"
    
    # Export for use by other functions
    AVAILABLE_DATABASES=("${available_databases[@]}")
}

validate_requested_databases() {
    print_status "step" "Validating requested database setups..."
    
    local -a requested_databases
    [[ "$opt_setup_mariadb" == "true" ]] && requested_databases+=("mariadb")
    [[ "$opt_setup_mysql" == "true" ]] && requested_databases+=("mysql")
    [[ "$opt_setup_postgres" == "true" ]] && requested_databases+=("postgres")
    [[ "$opt_setup_mongodb" == "true" ]] && requested_databases+=("mongodb")
    
    # Check if requested databases are actually available
    local -a unavailable_databases
    for db in "${requested_databases[@]}"; do
        if [[ ! " ${AVAILABLE_DATABASES[*]} " =~ " $db " ]]; then
            unavailable_databases+=("$db")
        fi
    done
    
    if [[ ${#unavailable_databases[@]} -gt 0 ]]; then
        print_status "warning" "Requested database(s) not available: ${unavailable_databases[*]}"
        print_status "info" "Available databases: ${AVAILABLE_DATABASES[*]}"
        print_status "info" "Continuing with available databases only..."
        
        # Disable setup for unavailable databases
        for db in "${unavailable_databases[@]}"; do
            case "$db" in
                "mariadb") opt_setup_mariadb=false ;;
                "mysql") opt_setup_mysql=false ;;
                "postgres") opt_setup_postgres=false ;;
                "mongodb") opt_setup_mongodb=false ;;
            esac
        done
    fi
}

# =============================================================================
# ENHANCED DATABASE WAITING FUNCTIONS
# =============================================================================

wait_for_database_ready() {
    local db_type="$1"
    local container_name="$2"
    
    print_status "step" "Waiting for $db_type ($container_name) to be ready..."
    
    # Use config.sh utilities for consistent behavior
    case "$db_type" in
        "mariadb"|"mysql")
            wait_for_mysql_ready "$container_name"
            ;;
        "postgres")
            wait_for_postgres_ready "$container_name"
            ;;
        "mongodb")
            # Use the specialized MongoDB function from config.sh
            wait_for_mongodb "$opt_timeout"
            ;;
        *)
            print_status "warning" "Unknown database type: $db_type"
            return 1
            ;;
    esac
}

wait_for_mysql_ready() {
    local container_name="$1"
    local counter=0
    
    while [[ $counter -lt $opt_timeout ]]; do
        # Try direct connection with root user
        if eval "$CONTAINER_CMD exec $container_name mariadb -u root -p\$MYSQL_ROOT_PASSWORD -e 'SELECT 1;' >/dev/null 2>&1"; then
            print_status "success" "$container_name is ready!"
            return 0
        fi
        
        # Fallback: Check if mysqladmin ping works
        if eval "$CONTAINER_CMD exec $container_name mysqladmin ping -h localhost -u root -p\$MYSQL_ROOT_PASSWORD >/dev/null 2>&1"; then
            print_status "success" "$container_name is ready!"
            return 0
        fi
        
        # Check if container is still running
        local status=$(get_service_status "$container_name")
        if [[ "$status" != "running" ]]; then
            handle_error "$container_name container is not running (status: $status)"
        fi
        
        print_status "info" "$container_name not ready yet... ($counter/$opt_timeout seconds)"
        sleep 3
        counter=$((counter + 3))
    done
    
    handle_error "$container_name connection timeout after $opt_timeout seconds"
}

wait_for_postgres_ready() {
    local container_name="$1"
    local counter=0
    
    while [[ $counter -lt $opt_timeout ]]; do
        # Check if PostgreSQL is ready
        if eval "$CONTAINER_CMD exec $container_name pg_isready -U postgres >/dev/null 2>&1"; then
            # Also verify we can actually connect
            if eval "$CONTAINER_CMD exec $container_name psql -U postgres -c 'SELECT 1;' >/dev/null 2>&1"; then
                print_status "success" "$container_name is ready!"
                return 0
            fi
        fi
        
        # Check if container is still running
        local status=$(get_service_status "$container_name")
        if [[ "$status" != "running" ]]; then
            handle_error "$container_name container is not running (status: $status)"
        fi
        
        print_status "info" "$container_name not ready yet... ($counter/$opt_timeout seconds)"
        sleep 3
        counter=$((counter + 3))
    done
    
    handle_error "$container_name connection timeout after $opt_timeout seconds"
}

# =============================================================================
# DATABASE SETUP FUNCTIONS
# =============================================================================

setup_mariadb() {
    if [[ ! " ${AVAILABLE_DATABASES[*]} " =~ " mariadb " ]]; then
        print_status "info" "MariaDB not available, skipping setup"
        return 0
    fi
    
    print_status "step" "Setting up MariaDB databases and users..."
    
    wait_for_database_ready "mariadb" "mariadb"
    
    # Enhanced MariaDB setup with better conflict handling
    local mariadb_setup_sql="
-- Create databases with IF NOT EXISTS to avoid conflicts
CREATE DATABASE IF NOT EXISTS ${DB_MYSQL_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS ${MATOMO_DATABASE_DBNAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Drop and recreate users to ensure clean state
DROP USER IF EXISTS '${MARIADB_USER}'@'%';
DROP USER IF EXISTS '${DB_MYSQL_USER}'@'%';
DROP USER IF EXISTS '${MATOMO_DATABASE_USERNAME}'@'%';

-- Create users
CREATE USER '${MARIADB_USER}'@'%' IDENTIFIED BY '${MARIADB_PASSWORD}';
CREATE USER '${DB_MYSQL_USER}'@'%' IDENTIFIED BY '${DB_MYSQL_PASSWORD}';
CREATE USER '${MATOMO_DATABASE_USERNAME}'@'%' IDENTIFIED BY '${MATOMO_DATABASE_PASSWORD}';

-- Grant privileges
GRANT ALL PRIVILEGES ON *.* TO '${MARIADB_USER}'@'%' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON ${DB_MYSQL_NAME}.* TO '${DB_MYSQL_USER}'@'%';
GRANT ALL PRIVILEGES ON ${MATOMO_DATABASE_DBNAME}.* TO '${MATOMO_DATABASE_USERNAME}'@'%';

FLUSH PRIVILEGES;
"
    
    # Use mariadb command with better error handling
    if eval "$CONTAINER_CMD exec -i mariadb mariadb -u root -p\$MYSQL_ROOT_PASSWORD 2>/dev/null << 'EOF'
$mariadb_setup_sql
EOF"; then
        print_status "success" "MariaDB databases and users created successfully"
    else
        print_status "error" "Failed to setup MariaDB databases"
        print_status "info" "Attempting to show MariaDB error details..."
        # Show what went wrong
        eval "$CONTAINER_CMD exec mariadb mariadb -u root -p\$MYSQL_ROOT_PASSWORD -e 'SHOW DATABASES;' 2>&1" || true
        return 1
    fi
}

setup_mysql() {
    if [[ ! " ${AVAILABLE_DATABASES[*]} " =~ " mysql " ]]; then
        print_status "info" "MySQL not available, skipping setup"
        return 0
    fi
    
    print_status "step" "Setting up MySQL databases and users..."
    
    wait_for_database_ready "mysql" "mysql"
    
    # Create databases and users for MySQL
    local mysql_setup_sql="
-- Drop and recreate users to ensure clean state
DROP USER IF EXISTS '${MYSQL_CUSTOM_USER}'@'%';

-- Create user
CREATE USER '${MYSQL_CUSTOM_USER}'@'%' IDENTIFIED BY '${MYSQL_CUSTOM_USER_PASSWORD}';

-- Grant full root-level privileges
GRANT ALL PRIVILEGES ON *.* TO '${MYSQL_CUSTOM_USER}'@'%' WITH GRANT OPTION;

-- Apply changes
FLUSH PRIVILEGES;
"
    
    if eval "$CONTAINER_CMD exec -i mysql mysql -u root -p\"\$MYSQL_ROOT_PASSWORD\" 2>/dev/null <<EOF
$mysql_setup_sql
EOF
    "; then
        print_status "success" "MySQL databases and users created successfully"
    else
        handle_error "Failed to setup MySQL databases"
    fi
}

setup_postgres() {
    if [[ ! " ${AVAILABLE_DATABASES[*]} " =~ " postgres " ]]; then
        print_status "info" "PostgreSQL not available, skipping setup"
        return 0
    fi
    
    print_status "step" "Setting up PostgreSQL databases and users..."
    
    wait_for_database_ready "postgres" "postgres"
    
    # Enhanced PostgreSQL setup with conflict resolution
    local postgres_setup_sql="
-- Drop and recreate privileged ${POSTGRES_CUSTOM_USER} first
DROP ROLE IF EXISTS ${POSTGRES_CUSTOM_USER};
CREATE ROLE ${POSTGRES_CUSTOM_USER} WITH LOGIN PASSWORD '${POSTGRES_CUSTOM_USER_PASSWORD}' SUPERUSER CREATEDB CREATEROLE;

-- Terminate existing connections to databases we might need to recreate
SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
WHERE datname IN ('${MB_DB_NAME}', '${NC_DATABASE_NAME}') AND pid <> pg_backend_pid();

-- Wait a moment for connections to close
SELECT pg_sleep(2);

-- Drop existing databases if they exist (CASCADE to handle dependencies)
DROP DATABASE IF EXISTS ${MB_DB_NAME};
DROP DATABASE IF EXISTS ${NC_DATABASE_NAME};

-- Drop existing users if they exist
DROP USER IF EXISTS ${MB_DB_USER};
DROP USER IF EXISTS ${NC_DATABASE_USER};

-- Create standard users
CREATE USER ${MB_DB_USER} WITH PASSWORD '${ADMIN_PASSWORD}';
CREATE USER ${NC_DATABASE_USER} WITH PASSWORD '${ADMIN_PASSWORD}';

-- Create databases
CREATE DATABASE ${MB_DB_NAME} OWNER ${MB_DB_USER};
CREATE DATABASE ${NC_DATABASE_NAME} OWNER ${NC_DATABASE_USER};

-- Grant additional privileges
GRANT ALL PRIVILEGES ON DATABASE ${MB_DB_NAME} TO ${MB_DB_USER};
GRANT ALL PRIVILEGES ON DATABASE ${NC_DATABASE_NAME} TO ${NC_DATABASE_USER};
"
    
    if eval "$CONTAINER_CMD exec -i postgres psql -U postgres 2>/dev/null <<EOF
$postgres_setup_sql
EOF
    "; then
        print_status "success" "PostgreSQL databases and users created successfully"
    else
        print_status "error" "Failed to setup PostgreSQL databases"
        print_status "info" "Attempting to show PostgreSQL error details..."
        eval "$CONTAINER_CMD exec postgres psql -U postgres -c '\l' 2>&1" || true
        return 1
    fi
}

setup_mongodb() {
    if [[ ! " ${AVAILABLE_DATABASES[*]} " =~ " mongodb " ]]; then
        print_status "info" "MongoDB not available, skipping setup"
        return 0
    fi
    
    print_status "step" "Setting up MongoDB root-like user..."

    wait_for_database_ready "mongodb" "mongodb"

    local mongodb_setup_js="
// Switch to the admin database
use admin;

// Drop the user if it already exists
try { db.dropUser('${MONGO_CUSTOM_USER}'); } catch(e) { print('${MONGO_CUSTOM_USER} did not exist'); }

// Create a new superuser with root privileges
db.createUser({
    user: '${MONGO_CUSTOM_USER}',
    pwd: '${MONGO_CUSTOM_USER_PASSWORD}',
    roles: [
        { role: 'root', db: 'admin' }
    ]
});

print('MongoDB superuser ${MONGO_CUSTOM_USER} created successfully');
"

    if eval "$CONTAINER_CMD exec -i mongodb mongosh --quiet 2>/dev/null <<EOF
$mongodb_setup_js
EOF
    "; then
        print_status "success" "MongoDB superuser created successfully"
    else
        print_status "error" "Failed to setup MongoDB user"
        print_status "info" "Attempting to show MongoDB error details..."
        eval "$CONTAINER_CMD exec mongodb mongosh --eval 'show users' 2>&1" || true
        return 1
    fi
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
    echo ""
    echo "  Databases to setup:"
    [[ "$opt_setup_mariadb" == "true" ]] && echo "    ✓ MariaDB"
    [[ "$opt_setup_mysql" == "true" ]] && echo "    ✓ MySQL"
    [[ "$opt_setup_postgres" == "true" ]] && echo "    ✓ PostgreSQL"
    [[ "$opt_setup_mongodb" == "true" ]] && echo "    ✓ MongoDB"
    echo ""
}

run_database_setup() {
    print_status "step" "Starting enhanced database setup process..."
    
    # Detect what's actually available
    detect_available_databases
    
    # Validate requested vs available
    validate_requested_databases
    
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
    print_status "info" "Management commands:"
    echo "  Check status: ./scripts/service-manager.sh status"
    echo "  View logs: ./scripts/service-manager.sh logs [database_name]"
    echo "  Restart service: ./scripts/service-manager.sh restart [database_name]"
}

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options (from config.sh)
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Show setup summary
    show_setup_summary
    
    # Run the enhanced database setup
    run_database_setup
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi