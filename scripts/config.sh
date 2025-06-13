#!/bin/zsh
# Enhanced Configuration Script for Microservices Architecture
# Provides centralized configuration, logging, and utility functions

# =============================================================================
# CORE CONFIGURATION
# =============================================================================

# Get script location portably
PROJECT_ROOT=$(cd "$(dirname "$0")/../" && pwd)
SCRIPT_DIR="${PROJECT_ROOT}/scripts"
COMPOSE_DIR="${PROJECT_ROOT}/compose"
APPS_DIR="${PROJECT_ROOT}/apps"
CONFIG_DIR="${PROJECT_ROOT}/config"
LOGS_DIR="${PROJECT_ROOT}/logs"

# Network and service configuration
NETWORK_NAME="microservices-net"
COMPOSE_PROJECT_NAME="microservices"

# Default credentials (override in .env for security)
DEFAULT_DB_PASSWORD="123456"
DEFAULT_ADMIN_USER="admin"
DEFAULT_ADMIN_PASSWORD="123456"
DEFAULT_ADMIN_EMAIL="admin@site.test"

# =============================================================================
# SERVICE CATEGORIES AND COMPOSE FILES
# =============================================================================

# Define service categories with their compose files
declare -A SERVICE_CATEGORIES=(
    ["database"]="database/postgres.yml database/mysql.yml database/mariadb.yml database/mongodb.yml database/redis.yml"
    ["dbms"]="dbms/adminer.yml dbms/pgadmin.yml dbms/phpmyadmin.yml dbms/mongo-express.yml dbms/metabase.yml dbms/nocodb.yml"
    ["backend"]="backend/php.yml backend/node.yml backend/python.yml backend/go.yml backend/dotnet.yml"
    ["analytics"]="analytics/elasticsearch.yml analytics/kibana.yml analytics/logstash.yml analytics/grafana.yml analytics/prometheus.yml analytics/matomo.yml"
    ["ai"]="ai/langflow.yml ai/n8n.yml"
    ["mail"]="mail/mailpit.yml"
    ["project"]="project/gitea.yml"
    ["erp"]="erp/odoo.yml"
    ["auth"]="auth/keycloak.yml"
    ["proxy"]="proxy/nginx-proxy-manager.yml"
)

# Service startup order (dependencies first)
SERVICE_STARTUP_ORDER=(
    "database"
    "dbms"
    "backend"
    "analytics"
    "ai"
    "mail"
    "project"
    "erp"
    "auth"
    "proxy"
)

# =============================================================================
# LOGGING CONFIGURATION
# =============================================================================

# Log levels
LOG_LEVEL_DEBUG=0
LOG_LEVEL_INFO=1
LOG_LEVEL_WARN=2
LOG_LEVEL_ERROR=3

# Current log level (set via environment or default to INFO)
CURRENT_LOG_LEVEL=${MS_LOG_LEVEL:-$LOG_LEVEL_INFO}

# Log colors
LOG_COLOR_DEBUG='\033[0;36m'    # Cyan
LOG_COLOR_INFO='\033[0;32m'     # Green
LOG_COLOR_WARN='\033[0;33m'     # Yellow
LOG_COLOR_ERROR='\033[0;31m'    # Red
LOG_COLOR_RESET='\033[0m'       # Reset

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Enhanced logging function
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local script_name=$(basename "$0")
    
    case "$level" in
        "DEBUG")
            [ $CURRENT_LOG_LEVEL -le $LOG_LEVEL_DEBUG ] && \
                echo -e "${LOG_COLOR_DEBUG}[$timestamp] [DEBUG] [$script_name] $message${LOG_COLOR_RESET}" ;;
        "INFO")
            [ $CURRENT_LOG_LEVEL -le $LOG_LEVEL_INFO ] && \
                echo -e "${LOG_COLOR_INFO}[$timestamp] [INFO] [$script_name] $message${LOG_COLOR_RESET}" ;;
        "WARN")
            [ $CURRENT_LOG_LEVEL -le $LOG_LEVEL_WARN ] && \
                echo -e "${LOG_COLOR_WARN}[$timestamp] [WARN] [$script_name] $message${LOG_COLOR_RESET}" >&2 ;;
        "ERROR")
            [ $CURRENT_LOG_LEVEL -le $LOG_LEVEL_ERROR ] && \
                echo -e "${LOG_COLOR_ERROR}[$timestamp] [ERROR] [$script_name] $message${LOG_COLOR_RESET}" >&2 ;;
    esac
    
    # Also log to file if logs directory exists
    [ -d "$LOGS_DIR" ] && echo "[$timestamp] [$level] [$script_name] $message" >> "$LOGS_DIR/scripts.log"
}

# Command execution with logging
execute_command() {
    local description="$1"
    local command="$2"
    local show_errors="${3:-false}"
    local critical="${4:-false}"
    
    log "DEBUG" "Executing: $description"
    log "DEBUG" "Command: $command"
    
    local error_redirect=""
    [ "$show_errors" = "false" ] && error_redirect="2>/dev/null"
    
    if eval "$command $error_redirect"; then
        log "INFO" "✅ $description - Success"
        return 0
    else
        local exit_code=$?
        if [ "$critical" = "true" ]; then
            log "ERROR" "❌ $description - Failed (exit code: $exit_code)"
            log "ERROR" "Command: $command"
            exit $exit_code
        else
            log "WARN" "⚠️  $description - Failed (exit code: $exit_code), continuing..."
            return $exit_code
        fi
    fi
}

# Container runtime detection
detect_container_runtime() {
    if command -v podman >/dev/null 2>&1; then
        echo "podman"
    elif command -v docker >/dev/null 2>&1; then
        echo "docker"
    else
        log "ERROR" "Neither podman nor docker found. Please install one of them."
        exit 1
    fi
}

# Network management
ensure_network() {
    local runtime="$1"
    local sudo_prefix="$2"
    
    log "DEBUG" "Checking if network '$NETWORK_NAME' exists"
    
    if ! execute_command "Check network existence" "${sudo_prefix}${runtime} network exists $NETWORK_NAME" false false; then
        execute_command "Create network" "${sudo_prefix}${runtime} network create --driver bridge $NETWORK_NAME" true true
    else
        log "INFO" "Network '$NETWORK_NAME' already exists"
    fi
}

# Service status checking
check_service_health() {
    local service_name="$1"
    local runtime="$2"
    local sudo_prefix="$3"
    local max_attempts="${4:-30}"
    local wait_seconds="${5:-2}"
    
    log "INFO" "Checking health of service: $service_name"
    
    for ((i=1; i<=max_attempts; i++)); do
        if execute_command "Check $service_name status" "${sudo_prefix}${runtime} container inspect $service_name --format '{{.State.Status}}'" false false; then
            local status=$(${sudo_prefix}${runtime} container inspect $service_name --format '{{.State.Status}}' 2>/dev/null)
            if [ "$status" = "running" ]; then
                log "INFO" "✅ $service_name is healthy (attempt $i/$max_attempts)"
                return 0
            fi
        fi
        
        log "DEBUG" "Service $service_name not ready, attempt $i/$max_attempts"
        sleep $wait_seconds
    done
    
    log "WARN" "Service $service_name health check timed out after $max_attempts attempts"
    return 1
}

# Compose file validation
validate_compose_file() {
    local compose_file="$1"
    
    if [ ! -f "$compose_file" ]; then
        log "ERROR" "Compose file not found: $compose_file"
        return 1
    fi
    
    # Basic YAML validation (check if it can be parsed)
    if command -v yq >/dev/null 2>&1; then
        if ! yq eval '.' "$compose_file" >/dev/null 2>&1; then
            log "ERROR" "Invalid YAML syntax in: $compose_file"
            return 1
        fi
    fi
    
    log "DEBUG" "Compose file validated: $compose_file"
    return 0
}

# Environment setup
setup_environment() {
    # Create necessary directories
    local dirs=("$LOGS_DIR" "$APPS_DIR")
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            log "INFO" "Creating directory: $dir"
            mkdir -p "$dir" || {
                log "ERROR" "Failed to create directory: $dir"
                exit 1
            }
        fi
    done
    
    # Check for .env file
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        if [ -f "$PROJECT_ROOT/.env-sample" ]; then
            log "WARN" ".env file not found, copying from .env-sample"
            cp "$PROJECT_ROOT/.env-sample" "$PROJECT_ROOT/.env"
        else
            log "ERROR" "No .env or .env-sample file found"
            exit 1
        fi
    fi
}

# Parse common command line arguments
parse_common_args() {
    # Reset variables
    USE_SUDO=false
    SHOW_ERRORS=false
    VERBOSE=false
    DRY_RUN=false
    CONTAINER_RUNTIME=""
    
    while getopts "sevdr:" opt; do
        case $opt in
            s) USE_SUDO=true ;;
            e) SHOW_ERRORS=true ;;
            v) VERBOSE=true; CURRENT_LOG_LEVEL=$LOG_LEVEL_DEBUG ;;
            d) DRY_RUN=true ;;
            r) CONTAINER_RUNTIME="$OPTARG" ;;
            ?) 
                echo "Common options:"
                echo "  -s    Use sudo for container commands"
                echo "  -e    Show error messages"
                echo "  -v    Verbose output (debug mode)"
                echo "  -d    Dry run (show commands without executing)"
                echo "  -r    Container runtime (docker/podman)"
                return 1
                ;;
        esac
    done
    
    # Auto-detect container runtime if not specified
    if [ -z "$CONTAINER_RUNTIME" ]; then
        CONTAINER_RUNTIME=$(detect_container_runtime)
    fi
    
    # Set up command prefix
    SUDO_PREFIX=""
    [ "$USE_SUDO" = true ] && SUDO_PREFIX="sudo "
    
    log "DEBUG" "Configuration: runtime=$CONTAINER_RUNTIME, sudo=$USE_SUDO, errors=$SHOW_ERRORS, verbose=$VERBOSE, dry_run=$DRY_RUN"
}

# =============================================================================
# EXPORT VARIABLES AND FUNCTIONS
# =============================================================================

# Export all variables for use in other scripts
export PROJECT_ROOT SCRIPT_DIR COMPOSE_DIR APPS_DIR CONFIG_DIR LOGS_DIR
export NETWORK_NAME COMPOSE_PROJECT_NAME
export DEFAULT_DB_PASSWORD DEFAULT_ADMIN_USER DEFAULT_ADMIN_PASSWORD DEFAULT_ADMIN_EMAIL
export CURRENT_LOG_LEVEL LOG_LEVEL_DEBUG LOG_LEVEL_INFO LOG_LEVEL_WARN LOG_LEVEL_ERROR

# For zsh, we need to handle function exports differently
# Functions will be available when this script is sourced
# Other scripts should source this file rather than rely on exported functions

# Initialize environment on source
setup_environment

log "DEBUG" "Configuration loaded successfully"