#!/bin/zsh

# =============================================================================
# MICROSERVICES CONFIGURATION MANAGEMENT
# =============================================================================
# Central configuration file for all scripts in the microservices architecture
# This file should be sourced by all other scripts for consistent behavior
# Optimized for zsh but maintains compatibility

# =============================================================================
# SCRIPT METADATA & PATHS
# =============================================================================

# Get script location using zsh-specific method
SCRIPT_SOURCE="${(%):-%x}"

# Establish project structure
export PROJECT_ROOT=$(cd "$(dirname "$SCRIPT_SOURCE")/../" && pwd)
export SCRIPT_DIR="${PROJECT_ROOT}/scripts"
export COMPOSE_DIR="${PROJECT_ROOT}/compose"
export CONFIG_DIR="${PROJECT_ROOT}/config"
export APPS_DIR="${PROJECT_ROOT}/apps"
export LOGS_DIR="${PROJECT_ROOT}/logs"

# =============================================================================
# NETWORK & CONTAINER CONFIGURATION
# =============================================================================

export NETWORK_NAME="microservices-net"
export CONTAINER_RUNTIME="podman"  # Change to "docker" if using Docker instead

# =============================================================================
# SERVICE CATEGORIES & COMPOSE FILES
# =============================================================================

# Define service categories and their compose files using zsh associative arrays
typeset -A SERVICE_CATEGORIES
SERVICE_CATEGORIES=(
    [database]="database/mariadb.yml database/mysql.yml database/postgres.yml database/mongodb.yml database/redis.yml"
    [dbms]="dbms/adminer.yml dbms/phpmyadmin.yml dbms/mongo-express.yml dbms/metabase.yml dbms/nocodb.yml dbms/pgadmin.yml"
    [backend]="backend/dotnet.yml backend/go.yml backend/node.yml backend/php.yml backend/python.yml"
    [analytics]="analytics/elasticsearch.yml analytics/kibana.yml analytics/logstash.yml analytics/grafana.yml analytics/prometheus.yml analytics/matomo.yml"
    [ai-services]="ai/langflow.yml ai/n8n.yml"
    [mail]="mail/mailpit.yml"
    [project]="project/gitea.yml"
    [proxy]="proxy/nginx-proxy-manager.yml"
)

# Service startup order (critical for dependencies) - zsh array
SERVICE_STARTUP_ORDER=(
    "database"
    "dbms" 
    "backend"
    "analytics"
    "ai-services"
    "mail"
    "project"
    "proxy"
)

# =============================================================================
# DATABASE CREDENTIALS
# =============================================================================

export MARIADB_ROOT_PASSWORD="123456"
export MYSQL_ROOT_PASSWORD="123456"
export POSTGRES_PASSWORD="123456"
export MONGO_ROOT_PASSWORD="123456"
export ADMIN_PASSWORD="123456"

# =============================================================================
# SSL CONFIGURATION
# =============================================================================

export SSL_DOMAIN="*.test"
export SSL_CERT_DIR="/etc/letsencrypt/live/wildcard.test"
export SSL_DAYS_VALID="3650"  # 10 years for development

# =============================================================================
# RUNTIME DETECTION & COMMAND SETUP
# =============================================================================

# Detect if we should use sudo (for Docker on Linux)
detect_sudo_requirement() {
    if command -v podman >/dev/null 2>&1; then
        export USE_PODMAN=true
        export DEFAULT_SUDO=true
    elif command -v docker >/dev/null 2>&1; then
        export USE_PODMAN=false
        # Check if user is in docker group
        if groups | grep -q docker 2>/dev/null; then
            export DEFAULT_SUDO=false
        else
            export DEFAULT_SUDO=true
        fi
    else
        echo "Error: Neither podman nor docker found!"
        exit 1
    fi
}

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Set up command execution context
setup_command_context() {
    local use_sudo="${1:-$DEFAULT_SUDO}"
    local show_errors="${2:-false}"
    
    if [[ "$use_sudo" == "true" ]]; then
        export SUDO_PREFIX="sudo "
    else
        export SUDO_PREFIX=""
    fi
    
    if [[ "$show_errors" == "true" ]]; then
        export ERROR_REDIRECT=""
    else
        export ERROR_REDIRECT="2>/dev/null"
    fi
    
    # Set container command based on runtime
    if [[ "$USE_PODMAN" == "true" ]]; then
        export CONTAINER_CMD="${SUDO_PREFIX}podman"
        
        # Check for native podman compose support first
        if command -v "podman" >/dev/null 2>&1 && ${SUDO_PREFIX}podman compose --help >/dev/null 2>&1; then
            export COMPOSE_CMD="${SUDO_PREFIX}podman compose"
            export COMPOSE_PROVIDER="native"
        else
            # Fallback to podman-compose
            export COMPOSE_CMD="${SUDO_PREFIX}podman-compose"
            export COMPOSE_PROVIDER="external"
        fi
        
        # Set podman-specific environment variables
        export PODMAN_COMPOSE_WARNING_LOGS="false"
        export COMPOSE_IGNORE_ORPHANS="true"
        export PODMAN_USERNS="keep-id"
        
    else
        export CONTAINER_CMD="${SUDO_PREFIX}docker"
        export COMPOSE_CMD="${SUDO_PREFIX}docker compose"
        export COMPOSE_PROVIDER="native"
    fi
}

# Function to handle errors consistently
handle_error() {
    local message="$1"
    local exit_code="${2:-1}"
    echo "❌ Error: $message" >&2
    exit "$exit_code"
}

# Function to check command status
check_status() {
    local exit_code=$?
    local message="$1"
    if [[ $exit_code -ne 0 ]]; then
        handle_error "$message" "$exit_code"
    fi
}

# Function to print colored output
print_status() {
    local level="$1"
    local message="$2"
    
    case "$level" in
        "info")
            echo "ℹ️  $message"
            ;;
        "success")
            echo "✅ $message"
            ;;
        "warning")
            echo "⚠️  $message"
            ;;
        "error")
            echo "❌ $message"
            ;;
        "step")
            echo "🔄 $message"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Function to wait for service health
wait_for_service_health() {
    local service_name="$1"
    local timeout="${2:-60}"
    local counter=0
    
    print_status "step" "Waiting for $service_name to be healthy..."
    
    while [[ $counter -lt $timeout ]]; do
        local health_status
        health_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Health.Status}}' $service_name 2>/dev/null" || echo "unknown")
        
        case "$health_status" in
            "healthy")
                print_status "success" "$service_name is healthy!"
                return 0
                ;;
            "unhealthy")
                print_status "warning" "$service_name is unhealthy"
                return 1
                ;;
            *)
                print_status "info" "$service_name health status: ${health_status:-starting} ($counter/$timeout)"
                ;;
        esac
        
        sleep 2
        counter=$((counter + 2))
    done
    
    print_status "warning" "$service_name health check timeout, but continuing..."
    return 1
}

# Function to wait for MongoDB specifically (no built-in health check)
wait_for_mongodb() {
    local timeout="${1:-60}"
    local counter=0
    
    print_status "step" "Waiting for MongoDB to be ready..."
    
    while [[ $counter -lt $timeout ]]; do
        # Check if we can ping MongoDB
        if eval "$CONTAINER_CMD exec mongodb mongosh --quiet --eval 'db.adminCommand(\"ping\")' $ERROR_REDIRECT"; then
            print_status "success" "MongoDB is ready!"
            return 0
        fi
        
        # Check if container is still running
        if ! eval "$CONTAINER_CMD ps --format '{{.Names}}' | grep -q '^mongodb$' $ERROR_REDIRECT"; then
            print_status "error" "MongoDB container is not running"
            return 1
        fi
        
        print_status "info" "MongoDB not ready yet... ($counter/$timeout)"
        sleep 2
        counter=$((counter + 2))
    done
    
    print_status "warning" "MongoDB timeout, but continuing..."
    return 1
}

# Function to create network if it doesn't exist
ensure_network_exists() {
    if ! eval "$CONTAINER_CMD network exists \"$NETWORK_NAME\" $ERROR_REDIRECT"; then
        print_status "step" "Creating network: $NETWORK_NAME"
        eval "$CONTAINER_CMD network create --driver bridge \"$NETWORK_NAME\" $ERROR_REDIRECT"
        check_status "Failed to create network: $NETWORK_NAME"
        print_status "success" "Network created: $NETWORK_NAME"
    else
        print_status "info" "Network already exists: $NETWORK_NAME"
    fi
}

# Function to get service files for a category
get_service_files() {
    local category="$1"
    if [[ -n "${SERVICE_CATEGORIES[$category]}" ]]; then
        echo "${SERVICE_CATEGORIES[$category]}"
    else
        print_status "error" "Unknown service category: $category"
        return 1
    fi
}

# Function to start services from a category
start_service_category() {
    local category="$1"
    local service_files
    service_files=$(get_service_files "$category")
    
    if [[ -z "$service_files" ]]; then
        return 1
    fi
    
    print_status "step" "Starting $category services..."
    
    # Split service_files into array using zsh word splitting
    local -a files
    files=(${=service_files})
    
    for service_file in "${files[@]}"; do
        local full_path="$COMPOSE_DIR/$service_file"
        if [[ -f "$full_path" ]]; then
            print_status "info" "Starting services from $service_file..."
            eval "$COMPOSE_CMD -f \"$full_path\" up -d $ERROR_REDIRECT"
            check_status "Failed to start services from $service_file"
        else
            print_status "warning" "Service file not found: $full_path"
        fi
    done
    
    print_status "success" "$category services started successfully"
}

# Function to stop services from a category
stop_service_category() {
    local category="$1"
    local service_files
    service_files=$(get_service_files "$category")
    
    if [[ -z "$service_files" ]]; then
        return 1
    fi
    
    print_status "step" "Stopping $category services..."
    
    # Split service_files into array using zsh word splitting
    local -a files
    files=(${=service_files})
    
    for service_file in "${files[@]}"; do
        local full_path="$COMPOSE_DIR/$service_file"
        if [[ -f "$full_path" ]]; then
            print_status "info" "Stopping services from $service_file..."
            eval "$COMPOSE_CMD -f \"$full_path\" down $ERROR_REDIRECT" || true
        fi
    done
    
    print_status "success" "$category services stopped"
}

# =============================================================================
# OS DETECTION
# =============================================================================

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if grep -q Microsoft /proc/version 2>/dev/null; then
            export OS_TYPE="wsl"
        else
            export OS_TYPE="linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        export OS_TYPE="macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        export OS_TYPE="windows"
    else
        export OS_TYPE="unknown"
    fi
}

# =============================================================================
# INITIALIZATION
# =============================================================================

# Auto-detect environment when this script is sourced
detect_sudo_requirement
detect_os

# Set default command context
setup_command_context "$DEFAULT_SUDO" "false"

# Validate project structure
if [[ ! -d "$COMPOSE_DIR" ]]; then
    handle_error "Compose directory not found: $COMPOSE_DIR"
fi

if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
    print_status "warning" "Environment file not found: $PROJECT_ROOT/.env"
    print_status "info" "Consider copying .env-sample to .env"
fi

# =============================================================================
# EXPORTED FUNCTIONS
# =============================================================================

# Make utility functions available to other scripts
if [[ -n "$ZSH_VERSION" ]]; then
    # ZSH: Functions are automatically available in sourced scripts
    # No need to export individual functions
    :
else
    # Bash: Export functions explicitly
    export -f handle_error 2>/dev/null || true
    export -f check_status 2>/dev/null || true
    export -f print_status 2>/dev/null || true
    export -f wait_for_service_health 2>/dev/null || true
    export -f wait_for_mongodb 2>/dev/null || true
    export -f ensure_network_exists 2>/dev/null || true
    export -f get_service_files 2>/dev/null || true
    export -f start_service_category 2>/dev/null || true
    export -f stop_service_category 2>/dev/null || true
    export -f setup_command_context 2>/dev/null || true
fi

print_status "info" "Configuration loaded successfully"
print_status "info" "Container runtime: $CONTAINER_RUNTIME (sudo: $DEFAULT_SUDO)"
print_status "info" "Project root: $PROJECT_ROOT"