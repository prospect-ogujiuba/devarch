#!/bin/zsh
# Enhanced Database Setup Script for Microservices Architecture
# Sets up databases, users, and initial schemas for all services

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# =============================================================================
# SCRIPT-SPECIFIC CONFIGURATION
# =============================================================================

# Database setup options
SETUP_MARIADB=false
SETUP_MYSQL=false
SETUP_POSTGRES=false
SETUP_MONGODB=false
SETUP_REDIS=false
SETUP_ALL=false
CREATE_SAMPLE_DATA=false
BACKUP_EXISTING=false
RESTORE_FROM_BACKUP=false
BACKUP_FILE=""

# Database configurations
declare -A DB_CONFIGS=(
    ["mariadb_host"]="mariadb"
    ["mariadb_port"]="3306"
    ["mariadb_root_password"]="$DEFAULT_DB_PASSWORD"
    ["mysql_host"]="mysql"
    ["mysql_port"]="3306"
    ["mysql_root_password"]="$DEFAULT_DB_PASSWORD"
    ["postgres_host"]="postgres"
    ["postgres_port"]="5432"
    ["postgres_password"]="$DEFAULT_DB_PASSWORD"
    ["mongodb_host"]="mongodb"
    ["mongodb_port"]="27017"
    ["mongodb_root_password"]="$DEFAULT_DB_PASSWORD"
)

# =============================================================================
# HELP FUNCTION
# =============================================================================

show_help() {
    cat << EOF
Enhanced Database Setup Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -M              Setup MariaDB
    -Y              Setup MySQL
    -P              Setup PostgreSQL
    -O              Setup MongoDB
    -R              Setup Redis
    -A              Setup all databases
    -S              Create sample data
    -B              Backup existing databases before setup
    -F FILE         Restore from backup file
    -w SECONDS      Wait time for database readiness (default: 30)
    -h              Show this help message

DATABASE OPERATIONS:
    The script will create:
    - Required databases for each service
    - Service-specific users with appropriate permissions
    - Initial schemas and sample data (if requested)
    - Backup of existing data (if requested)

EXAMPLES:
    $0 -A                        # Setup all databases
    $0 -P -M                     # Setup only PostgreSQL and MariaDB
    $0 -A -S -v                  # Setup all with sample data, verbose
    $0 -B -A                     # Backup existing data then setup all
    $0 -F backup.sql -P          # Restore PostgreSQL from backup

DATABASES CREATED:
    MariaDB:     npm (Nginx Proxy Manager), matomo
    MySQL:       mysql (default), backup_db
    PostgreSQL:  metabase, nocodb, keycloak, odoo
    MongoDB:     admin (default), logs, analytics
    Redis:       (key-value store, no databases)

EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_database_args() {
    local OPTIND=1
    local wait_time=30
    
    while getopts "sevdr:MYPORAS:BF:w:h" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            M) SETUP_MARIADB=true ;;
            Y) SETUP_MYSQL=true ;;
            P) SETUP_POSTGRES=true ;;
            O) SETUP_MONGODB=true ;;
            R) SETUP_REDIS=true ;;
            A) SETUP_ALL=true ;;
            S) CREATE_SAMPLE_DATA=true ;;
            B) BACKUP_EXISTING=true ;;
            F) RESTORE_FROM_BACKUP=true; BACKUP_FILE="$OPTARG" ;;
            w) wait_time="$OPTARG" ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
    
    # Set all databases if -A flag is used
    if [ "$SETUP_ALL" = true ]; then
        SETUP_MARIADB=true
        SETUP_MYSQL=true
        SETUP_POSTGRES=true
        SETUP_MONGODB=true
        SETUP_REDIS=true
    fi
    
    # If no specific database selected, default to all
    if [ "$SETUP_MARIADB" = false ] && [ "$SETUP_MYSQL" = false ] && [ "$SETUP_POSTGRES" = false ] && [ "$SETUP_MONGODB" = false ] && [ "$SETUP_REDIS" = false ]; then
        log "INFO" "No specific database selected, setting up all databases"
        SETUP_ALL=true
        SETUP_MARIADB=true
        SETUP_MYSQL=true
        SETUP_POSTGRES=true
        SETUP_MONGODB=true
        SETUP_REDIS=true
    fi
    
    export DB_WAIT_TIME="$wait_time"
}

# =============================================================================
# DATABASE READINESS CHECKS
# =============================================================================

wait_for_database_ready() {
    local db_type="$1"
    local container_name="$2"
    local max_attempts="${DB_WAIT_TIME:-30}"
    
    log "INFO" "🔍 Waiting for $db_type database ($container_name) to be ready..."
    
    for ((i=1; i<=max_attempts; i++)); do
        case "$db_type" in
            "mariadb"|"mysql")
                if execute_command "Check $db_type readiness" \
                   "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name mysqladmin ping -h localhost --silent" false false; then
                    log "INFO" "✅ $db_type is ready (attempt $i/$max_attempts)"
                    return 0
                fi
                ;;
            "postgres")
                if execute_command "Check PostgreSQL readiness" \
                   "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name pg_isready -U postgres" false false; then
                    log "INFO" "✅ PostgreSQL is ready (attempt $i/$max_attempts)"
                    return 0
                fi
                ;;
            "mongodb")
                if execute_command "Check MongoDB readiness" \
                   "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name mongosh --quiet --eval 'db.adminCommand(\"ping\")'" false false; then
                    log "INFO" "✅ MongoDB is ready (attempt $i/$max_attempts)"
                    return 0
                fi
                ;;
            "redis")
                if execute_command "Check Redis readiness" \
                   "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name redis-cli ping" false false; then
                    log "INFO" "✅ Redis is ready (attempt $i/$max_attempts)"
                    return 0
                fi
                ;;
        esac
        
        log "DEBUG" "$db_type not ready, attempt $i/$max_attempts"
        sleep 2
    done
    
    log "ERROR" "❌ $db_type failed to become ready after $max_attempts attempts"
    return 1
}

# =============================================================================
# BACKUP FUNCTIONS
# =============================================================================

backup_database() {
    local db_type="$1"
    local container_name="$2"
    
    if [ "$BACKUP_EXISTING" = false ]; then
        return 0
    fi
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="$PROJECT_ROOT/backups"
    mkdir -p "$backup_dir"
    
    log "INFO" "💾 Creating backup of $db_type..."
    
    case "$db_type" in
        "mariadb"|"mysql")
            local backup_file="$backup_dir/${db_type}_backup_${timestamp}.sql"
            if [ "$DRY_RUN" = true ]; then
                log "INFO" "[DRY RUN] Would create backup: $backup_file"
            else
                execute_command "Backup $db_type" \
                    "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name mysqldump --all-databases -u root -p${DB_CONFIGS[${db_type}_root_password]} > $backup_file" \
                    "$SHOW_ERRORS" false
            fi
            ;;
        "postgres")
            local backup_file="$backup_dir/postgres_backup_${timestamp}.sql"
            if [ "$DRY_RUN" = true ]; then
                log "INFO" "[DRY RUN] Would create backup: $backup_file"
            else
                execute_command "Backup PostgreSQL" \
                    "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name pg_dumpall -U postgres > $backup_file" \
                    "$SHOW_ERRORS" false
            fi
            ;;
        "mongodb")
            local backup_dir_mongo="$backup_dir/mongodb_backup_${timestamp}"
            if [ "$DRY_RUN" = true ]; then
                log "INFO" "[DRY RUN] Would create backup: $backup_dir_mongo"
            else
                execute_command "Backup MongoDB" \
                    "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name mongodump --out /dump/backup_${timestamp}" \
                    "$SHOW_ERRORS" false
            fi
            ;;
    esac
}

# =============================================================================
# DATABASE SETUP FUNCTIONS
# =============================================================================

setup_mariadb() {
    local container_name="mariadb"
    
    log "INFO" "🔧 Setting up MariaDB databases..."
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$container_name" 2>/dev/null; then
        log "ERROR" "MariaDB container not found. Please start it first."
        return 1
    fi
    
    wait_for_database_ready "mariadb" "$container_name" || return 1
    backup_database "mariadb" "$container_name"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup MariaDB databases: npm, matomo"
        return 0
    fi
    
    # Create databases and users
    local sql_commands="
CREATE DATABASE IF NOT EXISTS npm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS matomo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'npm_user'@'%' IDENTIFIED BY '${DEFAULT_DB_PASSWORD}';
CREATE USER IF NOT EXISTS 'matomo_user'@'%' IDENTIFIED BY '${DEFAULT_DB_PASSWORD}';
GRANT ALL PRIVILEGES ON npm.* TO 'npm_user'@'%';
GRANT ALL PRIVILEGES ON matomo.* TO 'matomo_user'@'%';
FLUSH PRIVILEGES;
"
    
    execute_command "Setup MariaDB schemas" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name mysql -u root -p${DB_CONFIGS[mariadb_root_password]} -e \"$sql_commands\"" \
        "$SHOW_ERRORS" true
    
    if [ "$CREATE_SAMPLE_DATA" = true ]; then
        setup_mariadb_sample_data "$container_name"
    fi
    
    log "INFO" "✅ MariaDB setup completed"
}

setup_mysql() {
    local container_name="mysql"
    
    log "INFO" "🔧 Setting up MySQL databases..."
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$container_name" 2>/dev/null; then
        log "WARN" "MySQL container not found, skipping MySQL setup"
        return 0
    fi
    
    wait_for_database_ready "mysql" "$container_name" || return 1
    backup_database "mysql" "$container_name"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup MySQL databases: backup_db, analytics"
        return 0
    fi
    
    local sql_commands="
CREATE DATABASE IF NOT EXISTS backup_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS analytics CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'backup_user'@'%' IDENTIFIED BY '${DEFAULT_DB_PASSWORD}';
CREATE USER IF NOT EXISTS 'analytics_user'@'%' IDENTIFIED BY '${DEFAULT_DB_PASSWORD}';
GRANT ALL PRIVILEGES ON backup_db.* TO 'backup_user'@'%';
GRANT ALL PRIVILEGES ON analytics.* TO 'analytics_user'@'%';
FLUSH PRIVILEGES;
"
    
    execute_command "Setup MySQL schemas" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name mysql -u root -p${DB_CONFIGS[mysql_root_password]} -e \"$sql_commands\"" \
        "$SHOW_ERRORS" true
    
    log "INFO" "✅ MySQL setup completed"
}

setup_postgres() {
    local container_name="postgres"
    
    log "INFO" "🔧 Setting up PostgreSQL databases..."
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$container_name" 2>/dev/null; then
        log "ERROR" "PostgreSQL container not found. Please start it first."
        return 1
    fi
    
    wait_for_database_ready "postgres" "$container_name" || return 1
    backup_database "postgres" "$container_name"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup PostgreSQL databases: metabase, nocodb, keycloak, odoo"
        return 0
    fi
    
    # Create users and databases
    local sql_commands="
-- Create users
CREATE USER IF NOT EXISTS metabase_user WITH PASSWORD '${DEFAULT_DB_PASSWORD}';
CREATE USER IF NOT EXISTS nocodb_user WITH PASSWORD '${DEFAULT_DB_PASSWORD}';
CREATE USER IF NOT EXISTS keycloak_user WITH PASSWORD '${DEFAULT_DB_PASSWORD}';
CREATE USER IF NOT EXISTS odoo_user WITH PASSWORD '${DEFAULT_DB_PASSWORD}';

-- Create databases
CREATE DATABASE metabase OWNER metabase_user;
CREATE DATABASE nocodb OWNER nocodb_user;
CREATE DATABASE keycloak OWNER keycloak_user;
CREATE DATABASE odoo OWNER odoo_user;

-- Grant additional permissions
ALTER USER odoo_user WITH CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE metabase TO metabase_user;
GRANT ALL PRIVILEGES ON DATABASE nocodb TO nocodb_user;
GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak_user;
GRANT ALL PRIVILEGES ON DATABASE odoo TO odoo_user;
"
    
    execute_command "Setup PostgreSQL schemas" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name psql -U postgres -c \"$sql_commands\"" \
        "$SHOW_ERRORS" true
    
    if [ "$CREATE_SAMPLE_DATA" = true ]; then
        setup_postgres_sample_data "$container_name"
    fi
    
    log "INFO" "✅ PostgreSQL setup completed"
}

setup_mongodb() {
    local container_name="mongodb"
    
    log "INFO" "🔧 Setting up MongoDB databases..."
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$container_name" 2>/dev/null; then
        log "WARN" "MongoDB container not found, skipping MongoDB setup"
        return 0
    fi
    
    wait_for_database_ready "mongodb" "$container_name" || return 1
    backup_database "mongodb" "$container_name"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup MongoDB databases: logs, analytics, sessions"
        return 0
    fi
    
    # Create databases and users
    local mongo_script="
use logs;
db.createUser({
    user: 'logs_user',
    pwd: '${DEFAULT_DB_PASSWORD}',
    roles: [{role: 'readWrite', db: 'logs'}]
});

use analytics;
db.createUser({
    user: 'analytics_user', 
    pwd: '${DEFAULT_DB_PASSWORD}',
    roles: [{role: 'readWrite', db: 'analytics'}]
});

use sessions;
db.createUser({
    user: 'sessions_user',
    pwd: '${DEFAULT_DB_PASSWORD}',
    roles: [{role: 'readWrite', db: 'sessions'}]
});
"
    
    execute_command "Setup MongoDB schemas" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name mongosh --authenticationDatabase admin -u root -p ${DB_CONFIGS[mongodb_root_password]} --eval \"$mongo_script\"" \
        "$SHOW_ERRORS" true
    
    if [ "$CREATE_SAMPLE_DATA" = true ]; then
        setup_mongodb_sample_data "$container_name"
    fi
    
    log "INFO" "✅ MongoDB setup completed"
}

setup_redis() {
    local container_name="redis"
    
    log "INFO" "🔧 Setting up Redis configuration..."
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$container_name" 2>/dev/null; then
        log "WARN" "Redis container not found, skipping Redis setup"
        return 0
    fi
    
    wait_for_database_ready "redis" "$container_name" || return 1
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup Redis key namespaces"
        return 0
    fi
    
    # Redis doesn't have databases like SQL, but we can set up some initial keys for testing
    execute_command "Setup Redis test keys" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name redis-cli SET microservices:setup:timestamp $(date +%s)" \
        "$SHOW_ERRORS" false
    
    execute_command "Setup Redis health check key" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec $container_name redis-cli SET microservices:health:status 'ok'" \
        "$SHOW_ERRORS" false
    
    log "INFO" "✅ Redis setup completed"
}

# =============================================================================
# SAMPLE DATA FUNCTIONS
# =============================================================================

setup_mariadb_sample_data() {
    local container_name="$1"
    
    log "INFO" "📊 Creating MariaDB sample data..."
    
    local sample_sql="
USE npm;
INSERT IGNORE INTO test_table (id, name, created_at) VALUES 
(1, 'Sample Entry 1', NOW()),
(2, 'Sample Entry 2', NOW());

USE matomo;
CREATE TABLE IF NOT EXISTS sample_visits (
    id INT AUTO_INCREMENT PRIMARY KEY,
    visitor_ip VARCHAR(45),
    visit_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    page_url VARCHAR(255)
);
INSERT IGNORE INTO sample_visits (visitor_ip, page_url) VALUES 
('127.0.0.1', '/home'),
('127.0.0.1', '/about');
"
    
    execute_command "Create MariaDB sample data" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name mysql -u root -p${DB_CONFIGS[mariadb_root_password]} -e \"$sample_sql\"" \
        "$SHOW_ERRORS" false
}

setup_postgres_sample_data() {
    local container_name="$1"
    
    log "INFO" "📊 Creating PostgreSQL sample data..."
    
    local sample_sql="
\\c metabase;
CREATE TABLE IF NOT EXISTS sample_data (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    value INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);
INSERT INTO sample_data (name, value) VALUES 
('Sample Metric 1', 100),
('Sample Metric 2', 200);

\\c nocodb;
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100),
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
INSERT INTO projects (title, description) VALUES 
('Sample Project', 'This is a sample project for testing');
"
    
    execute_command "Create PostgreSQL sample data" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name psql -U postgres -c \"$sample_sql\"" \
        "$SHOW_ERRORS" false
}

setup_mongodb_sample_data() {
    local container_name="$1"
    
    log "INFO" "📊 Creating MongoDB sample data..."
    
    local mongo_script="
use logs;
db.application_logs.insertMany([
    {level: 'info', message: 'Application started', timestamp: new Date()},
    {level: 'error', message: 'Sample error message', timestamp: new Date()}
]);

use analytics;
db.events.insertMany([
    {event: 'page_view', page: '/home', timestamp: new Date()},
    {event: 'button_click', element: 'signup', timestamp: new Date()}
]);
"
    
    execute_command "Create MongoDB sample data" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $container_name mongosh --authenticationDatabase admin -u root -p ${DB_CONFIGS[mongodb_root_password]} --eval \"$mongo_script\"" \
        "$SHOW_ERRORS" false
}

# =============================================================================
# RESTORE FUNCTIONS
# =============================================================================

restore_from_backup() {
    if [ "$RESTORE_FROM_BACKUP" = false ] || [ -z "$BACKUP_FILE" ]; then
        return 0
    fi
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log "ERROR" "Backup file not found: $BACKUP_FILE"
        return 1
    fi
    
    log "INFO" "📥 Restoring from backup: $BACKUP_FILE"
    
    # Determine database type from filename
    local db_type=""
    if [[ "$BACKUP_FILE" == *"postgres"* ]]; then
        db_type="postgres"
    elif [[ "$BACKUP_FILE" == *"mariadb"* ]]; then
        db_type="mariadb"
    elif [[ "$BACKUP_FILE" == *"mysql"* ]]; then
        db_type="mysql"
    elif [[ "$BACKUP_FILE" == *"mongodb"* ]]; then
        db_type="mongodb"
    else
        log "ERROR" "Could not determine database type from backup filename"
        return 1
    fi
    
    case "$db_type" in
        "postgres")
            execute_command "Restore PostgreSQL backup" \
                "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i postgres psql -U postgres < $BACKUP_FILE" \
                "$SHOW_ERRORS" true
            ;;
        "mariadb"|"mysql")
            execute_command "Restore $db_type backup" \
                "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec -i $db_type mysql -u root -p${DB_CONFIGS[${db_type}_root_password]} < $BACKUP_FILE" \
                "$SHOW_ERRORS" true
            ;;
    esac
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

show_setup_summary() {
    local databases_to_setup=()
    
    [ "$SETUP_MARIADB" = true ] && databases_to_setup+=("MariaDB")
    [ "$SETUP_MYSQL" = true ] && databases_to_setup+=("MySQL")
    [ "$SETUP_POSTGRES" = true ] && databases_to_setup+=("PostgreSQL")
    [ "$SETUP_MONGODB" = true ] && databases_to_setup+=("MongoDB")
    [ "$SETUP_REDIS" = true ] && databases_to_setup+=("Redis")
    
    cat << EOF

📋 DATABASE SETUP SUMMARY
=========================

Databases to setup: ${databases_to_setup[*]}
Create sample data: $CREATE_SAMPLE_DATA
Backup existing: $BACKUP_EXISTING
Restore from backup: $RESTORE_FROM_BACKUP
$([ "$RESTORE_FROM_BACKUP" = true ] && echo "Backup file: $BACKUP_FILE")
Container runtime: $CONTAINER_RUNTIME
Wait time: ${DB_WAIT_TIME:-30} seconds

EOF
}

main() {
    parse_common_args "$@"
    parse_database_args "$@"
    
    show_setup_summary
    
    log "INFO" "🚀 Starting database setup..."
    
    # Restore from backup first if requested
    restore_from_backup
    
    # Setup databases in dependency order
    [ "$SETUP_POSTGRES" = true ] && setup_postgres
    [ "$SETUP_MARIADB" = true ] && setup_mariadb
    [ "$SETUP_MYSQL" = true ] && setup_mysql
    [ "$SETUP_MONGODB" = true ] && setup_mongodb
    [ "$SETUP_REDIS" = true ] && setup_redis
    
    log "INFO" "✅ Database setup completed successfully!"
    
    if [ "$DRY_RUN" = false ]; then
        log "INFO" "🔍 Database connection details:"
        log "INFO" "  PostgreSQL: localhost:8502 (user: postgres, password: $DEFAULT_DB_PASSWORD)"
        log "INFO" "  MariaDB: localhost:8501 (user: root, password: $DEFAULT_DB_PASSWORD)"
        log "INFO" "  MySQL: localhost:8505 (user: root, password: $DEFAULT_DB_PASSWORD)"
        log "INFO" "  MongoDB: localhost:8503 (user: root, password: $DEFAULT_DB_PASSWORD)"
        log "INFO" "  Redis: localhost:8504 (no authentication)"
    fi
}

main "$@"