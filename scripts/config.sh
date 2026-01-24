#!/bin/zsh

# =============================================================================
# MICROSERVICES CONFIGURATION MANAGEMENT - FIXED
# =============================================================================
# Smart configuration with NO INFINITE LOOPS - removed service-manager delegation

# =============================================================================
# BACKEND SERVICE PORT ALLOCATION STRATEGY
# =============================================================================
# Each runtime has a dedicated 100-port range for clean separation and
# simultaneous operation without conflicts:
#
# PHP:    8100-8199 (main: 8100, vite: 8102)
# Node:   8200-8299 (main: 8200, secondary: 8201, vite: 8202, graphql: 8203, debug: 9229)
# Python: 8300-8399 (main: 8300, flask: 8301, jupyter: 8302, flower: 8303)
# Go:     8400-8499 (main: 8400, metrics: 8401, debug: 8402, pprof: 8403)
# .NET:   8600-8699 (main: 8600, secondary: 8601, debug: 8602, hot-reload: 8603)
# Rust:   8700-8799 (main: 8700, secondary: 8701, debug: 8702, metrics: 8703)
#
# This allocation ensures all backend services can run simultaneously without
# port conflicts, supporting the microservices architecture effectively.

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
# SERVICE CATEGORIES & COMPOSE FILES - AUTO-DISCOVERY
# =============================================================================

# Service startup order (defines categories AND their dependency order)
# New categories are auto-discovered but appended at end (lowest priority)
SERVICE_STARTUP_ORDER=(
    "database"
    "storage"
    "dbms"
    "erp"
    "security"
    "registry"
    "gateway"
    "proxy"
    "management"
    "backend"
    "ci"
    "project"
    "mail"
    "exporters"
    "analytics"
    "messaging"
    "search"
    "workflow"
    "docs"
    "testing"
    "collaboration"
    "ai"
    "support"
)

# Optional: Override category paths if you need different directory structure
typeset -A CATEGORY_PATH_OVERRIDES
CATEGORY_PATH_OVERRIDES=(
    # [category]="custom/path"
    # Example: [backend]="apps/backend"
)

# Optional: Full path overrides for specific services (for maximum flexibility)
typeset -A SERVICE_PATH_OVERRIDES
SERVICE_PATH_OVERRIDES=(
    # [service.yml]="full/custom/path/service.yml"
    # Example: [special-service.yml]="legacy/docker-compose.yml"
)

# Auto-discovered service categories (populated by init_service_categories)
typeset -gA SERVICE_CATEGORIES

# Flag to track if discovery has run
_SERVICE_CATEGORIES_INITIALIZED=false

# =============================================================================
# SERVICE AUTO-DISCOVERY
# =============================================================================

# Discover services from compose directory structure
init_service_categories() {
    local force="${1:-}"
    [[ "$_SERVICE_CATEGORIES_INITIALIZED" == "true" && "$force" != "--force" ]] && return 0

    SERVICE_CATEGORIES=()
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        _discover_category_services "$category"
    done

    # Auto-discover new categories
    for dir in "$COMPOSE_DIR"/*(/N); do
        local category="${dir:t}"
        (( ${SERVICE_STARTUP_ORDER[(Ie)$category]} )) && continue
        _discover_category_services "$category"
        SERVICE_STARTUP_ORDER+=("$category")
    done

    _SERVICE_CATEGORIES_INITIALIZED=true
}

# Internal: Discover services for a single category
_discover_category_services() {
    local category="$1"
    local category_path="$COMPOSE_DIR/$category"

    # Check for path override
    [[ -n "${CATEGORY_PATH_OVERRIDES[$category]}" ]] && \
        category_path="$COMPOSE_DIR/${CATEGORY_PATH_OVERRIDES[$category]}"

    # Skip if directory doesn't exist
    [[ ! -d "$category_path" ]] && return 0

    # Collect all .yml files
    local services=""
    for yml_file in "$category_path"/*.yml(N); do
        [[ -f "$yml_file" ]] && services+="${yml_file:t} "
    done

    # Store (trim trailing space)
    SERVICE_CATEGORIES[$category]="${services% }"
}

# Get list of all known categories
get_all_categories() {
    init_service_categories
    echo "${SERVICE_STARTUP_ORDER[@]}"
}

# Get services for a category (triggers discovery if needed)
get_category_services() {
    local category="$1"
    init_service_categories
    echo "${SERVICE_CATEGORIES[$category]}"
}

# Refresh service discovery (force re-scan)
refresh_service_discovery() {
    init_service_categories --force
}

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

# Minimal output functions - ANSI colors, no emojis
# Respects VERBOSE and QUIET flags
VERBOSE=${VERBOSE:-0}
QUIET=${QUIET:-0}

print_status() {
    local level="$1"
    local message="$2"

    case "$level" in
        "info")
            [[ $VERBOSE -eq 1 ]] && printf "\033[36m--\033[0m %s\n" "$message"
            ;;
        "success")
            [[ $QUIET -eq 0 ]] && printf "\033[32mOK\033[0m %s\n" "$message"
            ;;
        "warning")
            printf "\033[33mWARN\033[0m %s\n" "$message"
            ;;
        "error")
            printf "\033[31mERR\033[0m %s\n" "$message" >&2
            ;;
        "step")
            [[ $VERBOSE -eq 1 ]] && printf "\033[34m..\033[0m %s\n" "$message"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# =============================================================================
# SMART PATH RESOLUTION FUNCTIONS
# =============================================================================

# Function to resolve the full path for a service file
resolve_service_path() {
    local service_file="$1"
    local category="$2"
    local resolved_path=""
    
    # Check for specific service override first
    if [[ -n "${SERVICE_PATH_OVERRIDES[$service_file]}" ]]; then
        resolved_path="$COMPOSE_DIR/${SERVICE_PATH_OVERRIDES[$service_file]}"
        echo "$resolved_path"
        return 0
    fi
    
    # Check for category path override
    local category_path="$category"
    if [[ -n "${CATEGORY_PATH_OVERRIDES[$category]}" ]]; then
        category_path="${CATEGORY_PATH_OVERRIDES[$category]}"
    fi
    
    # Construct the standard path
    local full_path="$COMPOSE_DIR/$category_path/$service_file"
    
    # Verify the file exists, if not try fallback locations
    if [[ -f "$full_path" ]]; then
        resolved_path="$full_path"
        echo "$resolved_path"
        return 0
    fi
    
    # Fallback 1: Try without category subdirectory (flat structure)
    local flat_path="$COMPOSE_DIR/$service_file"
    if [[ -f "$flat_path" ]]; then
        resolved_path="$flat_path"
        echo "$resolved_path"
        return 0
    fi
    
    # Fallback 2: Search in all subdirectories
    local found_path
    found_path=$(find "$COMPOSE_DIR" -name "$service_file" -type f 2>/dev/null | head -1)
    if [[ -n "$found_path" ]]; then
        resolved_path="$found_path"
        echo "$resolved_path"
        return 0
    fi
    
    # Return the expected path even if it doesn't exist (for error reporting)
    resolved_path="$full_path"
    echo "$resolved_path"
    return 1
}

# Enhanced function to get service files for a category with full paths
get_service_files() {
    local category="$1"
    local return_paths="${2:-false}"  # if true, return full paths instead of just filenames
    
    if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
        print_status "error" "Unknown service category: $category"
        return 1
    fi
    
    local service_files="${SERVICE_CATEGORIES[$category]}"
    
    if [[ "$return_paths" == "true" ]]; then
        # Return full resolved paths
        local -a files paths
        files=(${=service_files})
        
        for service_file in "${files[@]}"; do
            local resolved_path
            resolved_path=$(resolve_service_path "$service_file" "$category")
            paths+=("$resolved_path")
        done
        
        echo "${paths[@]}"
    else
        # Return just the filenames (original behavior)
        echo "$service_files"
    fi
}

# Validate all service files exist
validate_service_files() {
    local category="$1"
    [[ -z "${SERVICE_CATEGORIES[$category]}" ]] && { print_status "error" "Unknown category: $category"; return 1; }

    local -a files=(${=${SERVICE_CATEGORIES[$category]}})
    local -a missing
    for sf in "${files[@]}"; do
        local rp=$(resolve_service_path "$sf" "$category")
        [[ ! -f "$rp" ]] && missing+=("$sf")
    done

    [[ ${#missing[@]} -gt 0 ]] && { print_status "warning" "$category missing: ${missing[*]}"; return 1; }
    return 0
}

# =============================================================================
# SERVICE DISCOVERY FUNCTIONS
# =============================================================================

# Function to find which category a service belongs to
find_service_category() {
    local service_name="$1"
    local service_file="${service_name}.yml"
    
    # Search through all categories
    for category in "${(@k)SERVICE_CATEGORIES}"; do
        local service_files="${SERVICE_CATEGORIES[$category]}"
        if [[ "$service_files" == *"$service_file"* ]]; then
            echo "$category"
            return 0
        fi
    done
    
    # Not found in any category
    return 1
}

# Function to get service file path for a given service name
get_service_path() {
    local service_name="$1"
    local category
    
    # Find which category this service belongs to
    category=$(find_service_category "$service_name")
    if [[ $? -eq 0 ]]; then
        local service_file="${service_name}.yml"
        resolve_service_path "$service_file" "$category"
        return 0
    else
        # Service not found
        return 1
    fi
}

# Function to validate if a service exists
validate_service_exists() {
    local service_name="$1"
    local service_path
    
    service_path=$(get_service_path "$service_name")
    if [[ $? -eq 0 && -f "$service_path" ]]; then
        return 0
    else
        return 1
    fi
}

# Function to list all available services
list_all_service_names() {
    local -a all_services
    
    for category in "${(@k)SERVICE_CATEGORIES}"; do
        local service_files="${SERVICE_CATEGORIES[$category]}"
        local -a files
        files=(${=service_files})
        
        for service_file in "${files[@]}"; do
            local service_name="${service_file%.yml}"
            all_services+=("$service_name")
        done
    done
    
    # Remove duplicates and sort
    printf '%s\n' "${all_services[@]}" | sort -u
}

# Function to list all available service files with their resolved paths
list_all_services() {
    echo "=== SERVICE FILE MAPPING ==="
    echo ""
    
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        echo "Category: $category"
        echo "----------------------------------------"
        
        if [[ -n "${SERVICE_CATEGORIES[$category]}" ]]; then
            local service_files="${SERVICE_CATEGORIES[$category]}"
            local -a files
            files=(${=service_files})
            
            for service_file in "${files[@]}"; do
                local file_path="$(resolve_service_path "$service_file" "$category" 2>/dev/null)"
                local status_icon="-"
                [[ -f "$file_path" ]] && status_icon="+"
                printf "  %s %-20s %s\n" "$status_icon" "${service_file%.yml}" "$file_path"
            done
        else
            echo "  No services defined"
        fi
        echo ""
    done
}

# =============================================================================
# DIRECT SERVICE OPERATIONS (NO DELEGATION TO AVOID LOOPS)
# =============================================================================

# Function to get service status (lightweight, direct implementation)
get_service_status() {
    local service_name="$1"

    # Resolve service path to validate existence
    local service_path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        echo "NOT_FOUND"
        return 1
    fi

    # Try to find the real container name via compose labels
    # Works for both docker and podman compose
    local cname=""
    cname=$(eval "$CONTAINER_CMD ps --filter 'label=com.docker.compose.service=$service_name' --format '{{.Names}}' 2>/dev/null | head -n1")

    # Fallback to literal service name if label search returns nothing
    [[ -z "$cname" ]] && cname="$service_name"

    # Query status
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' \"$cname\" 2>/dev/null")

    if [[ -n "$container_status" ]]; then
        echo "$container_status"
        return 0
    else
        echo "STOPPED"
        return 1
    fi
}

# Direct service operations - thin wrappers around compose
start_single_service() {
    local service_name="$1"
    local force_recreate="${2:-false}"
    local service_path=$(get_service_path "$service_name")
    [[ $? -ne 0 ]] && { print_status "error" "$service_name not found"; return 1; }

    print_status "step" "up $service_name"
    local compose_args=("-f" "$service_path" "up" "-d")
    [[ "$force_recreate" == "true" ]] && compose_args+=("--force-recreate")

    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "success" "$service_name"
        return 0
    else
        print_status "error" "$service_name failed"
        return 1
    fi
}

stop_single_service() {
    local service_name="$1"
    local remove_volumes="${2:-false}"
    local timeout="${3:-30}"
    local service_path=$(get_service_path "$service_name")
    [[ $? -ne 0 ]] && { print_status "error" "$service_name not found"; return 1; }

    print_status "step" "down $service_name"
    local compose_args=("-f" "$service_path" "down" "--timeout" "$timeout")
    [[ "$remove_volumes" == "true" ]] && compose_args+=("--volumes")

    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "success" "$service_name stopped"
        return 0
    else
        print_status "error" "$service_name stop failed"
        return 1
    fi
}

restart_single_service() {
    local service_name="$1"
    stop_single_service "$service_name" && start_single_service "$service_name"
}

start_container() {
    local service_name="$1"
    if eval "$CONTAINER_CMD start $service_name $ERROR_REDIRECT"; then
        print_status "success" "$service_name"
        return 0
    else
        print_status "error" "$service_name start failed"
        return 1
    fi
}

stop_container() {
    local service_name="$1"
    local timeout="${2:-30}"
    if eval "$CONTAINER_CMD stop -t $timeout $service_name $ERROR_REDIRECT"; then
        print_status "success" "$service_name stopped"
        return 0
    else
        print_status "error" "$service_name stop failed"
        return 1
    fi
}

rebuild_single_service() {
    local service_name="$1"
    local no_cache="${2:-false}"
    local service_path=$(get_service_path "$service_name")
    [[ $? -ne 0 ]] && { print_status "error" "$service_name not found"; return 1; }

    stop_single_service "$service_name"
    local build_args=("-f" "$service_path" "build")
    [[ "$no_cache" == "true" ]] && build_args+=("--no-cache")

    if eval "$COMPOSE_CMD ${build_args[*]} $ERROR_REDIRECT"; then
        start_single_service "$service_name" "true"
        return 0
    else
        print_status "error" "$service_name build failed"
        return 1
    fi
}

show_service_logs() {
    local service_name="$1"
    local follow="${2:-false}"
    local tail_lines="${3:-100}"
    local service_path=$(get_service_path "$service_name")
    [[ $? -ne 0 ]] && { print_status "error" "$service_name not found"; return 1; }

    local log_args=("-f" "$service_path" "logs" "--tail" "$tail_lines")
    [[ "$follow" == "true" ]] && log_args+=("-f")
    eval "$COMPOSE_CMD ${log_args[*]}"
}

# Category operations
start_service_category() {
    local category="$1"
    [[ -z "${SERVICE_CATEGORIES[$category]}" ]] && { print_status "error" "Unknown category: $category"; return 1; }
    local -a files=(${=${SERVICE_CATEGORIES[$category]}})
    local failed=0
    for sf in "${files[@]}"; do start_single_service "${sf%.yml}" || ((failed++)); done
    [[ $failed -gt 0 ]] && print_status "warning" "$category: $failed failed"
    return $([[ $failed -eq 0 ]] && echo 0 || echo 1)
}

stop_service_category() {
    local category="$1"
    [[ -z "${SERVICE_CATEGORIES[$category]}" ]] && { print_status "error" "Unknown category: $category"; return 1; }
    local -a files=(${=${SERVICE_CATEGORIES[$category]}})
    local failed=0
    for sf in "${files[@]}"; do stop_single_service "${sf%.yml}" || ((failed++)); done
    [[ $failed -gt 0 ]] && print_status "warning" "$category: $failed failed"
    return $([[ $failed -eq 0 ]] && echo 0 || echo 1)
}

# =============================================================================
# CONFIGURATION UTILITIES
# =============================================================================

# Function to add a service override
add_service_override() {
    local service_file="$1"
    local custom_path="$2"
    
    SERVICE_PATH_OVERRIDES[$service_file]="$custom_path"
    print_status "success" "Added service override: $service_file -> $custom_path"
}

# Function to add a category override
add_category_override() {
    local category="$1"
    local custom_path="$2"
    
    CATEGORY_PATH_OVERRIDES[$category]="$custom_path"
    print_status "success" "Added category override: $category -> $custom_path"
}

# Function to show current configuration
show_configuration() {
    echo "=== SMART CONFIGURATION STATUS ==="
    echo ""
    echo "Project Root: $PROJECT_ROOT"
    echo "Compose Dir:  $COMPOSE_DIR"
    echo "Container Runtime: $CONTAINER_RUNTIME"
    echo ""
    
    echo "Service Categories:"
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        local count
        count=$(echo "${SERVICE_CATEGORIES[$category]}" | wc -w)
        printf "  %-15s: %d services\n" "$category" "$count"
    done
    echo ""
    
    if [[ ${#CATEGORY_PATH_OVERRIDES[@]} -gt 0 ]]; then
        echo "Category Path Overrides:"
        for category in "${(@k)CATEGORY_PATH_OVERRIDES}"; do
            printf "  %-15s -> %s\n" "$category" "${CATEGORY_PATH_OVERRIDES[$category]}"
        done
        echo ""
    fi
    
    if [[ ${#SERVICE_PATH_OVERRIDES[@]} -gt 0 ]]; then
        echo "Service Path Overrides:"
        for service in "${(@k)SERVICE_PATH_OVERRIDES}"; do
            printf "  %-25s -> %s\n" "$service" "${SERVICE_PATH_OVERRIDES[$service]}"
        done
        echo ""
    fi
}

# =============================================================================
# DATABASE CREDENTIALS & ENVIRONMENT VARIABLES
# =============================================================================

# Source .env file if it exists
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a; source "$PROJECT_ROOT/.env"; set +a
fi

# Fallback defaults if .env doesn't exist or variables are missing
export MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-123456}"
export MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-123456}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-123456}"
export MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-123456}"
export ADMIN_PASSWORD="${ADMIN_PASSWORD:-123456}"
export ADMIN_USER="${ADMIN_USER:-admin}"
export ADMIN_EMAIL="${ADMIN_EMAIL:-admin@test.local}"
export MARIADB_USER="${MARIADB_USER:-mariadb_user}"
export MARIADB_PASSWORD="${MARIADB_PASSWORD:-123456}"
export MARIADB_HOST="${MARIADB_HOST:-mariadb}"

# Additional .env variables that might be used by database setup
export DB_MYSQL_NAME="${DB_MYSQL_NAME:-npm}"
export DB_MYSQL_USER="${DB_MYSQL_USER:-npm_user}"
export DB_MYSQL_PASSWORD="${DB_MYSQL_PASSWORD:-${ADMIN_PASSWORD}}"
export MATOMO_DATABASE_DBNAME="${MATOMO_DATABASE_DBNAME:-matomo}"
export MATOMO_DATABASE_USERNAME="${MATOMO_DATABASE_USERNAME:-matomo_user}"
export MATOMO_DATABASE_PASSWORD="${MATOMO_DATABASE_PASSWORD:-${ADMIN_PASSWORD}}"
export MYSQL_CUSTOM_USER="${MYSQL_CUSTOM_USER:-mysql_user}"
export MYSQL_CUSTOM_USER_PASSWORD="${MYSQL_CUSTOM_USER_PASSWORD:-${ADMIN_PASSWORD}}"
export POSTGRES_CUSTOM_USER="${POSTGRES_CUSTOM_USER:-postgres_user}"
export POSTGRES_CUSTOM_USER_PASSWORD="${POSTGRES_CUSTOM_USER_PASSWORD:-${ADMIN_PASSWORD}}"
export MB_DB_NAME="${MB_DB_NAME:-metabase}"
export MB_DB_USER="${MB_DB_USER:-metabase_user}"
export NC_DATABASE_NAME="${NC_DATABASE_NAME:-nocodb}"
export NC_DATABASE_USER="${NC_DATABASE_USER:-nocodb_user}"
export MONGO_CUSTOM_USER="${MONGO_CUSTOM_USER:-mongodb_user}"
export MONGO_CUSTOM_USER_PASSWORD="${MONGO_CUSTOM_USER_PASSWORD:-${ADMIN_PASSWORD}}"

# GitHub integration variables (if you use them)
export GITHUB_TOKEN="${GITHUB_TOKEN:-}"
export GITHUB_USER="${GITHUB_USER:-}"

# =============================================================================
# RUNTIME DETECTION & COMMAND SETUP
# =============================================================================

# Detect if we should use sudo (for Docker on Linux)
detect_sudo_requirement() {
    if command -v podman >/dev/null 2>&1; then
        export USE_PODMAN=true
        export DEFAULT_SUDO=false
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

# Set up command execution context
setup_command_context() {
    local want_sudo="${1:-$DEFAULT_SUDO}"   # "true" / "false"
    local show_errors="${2:-false}"

    # stderr handling
    if [[ "$show_errors" == "true" ]]; then
        export ERROR_REDIRECT=""
    else
        export ERROR_REDIRECT="2>/dev/null"
    fi

    # Decide engine
    if [[ "$USE_PODMAN" == "true" ]]; then
        local rootless_has_net=""
        local rootful_has_net=""

        # Does our project network exist rootless?
        if podman network exists "$NETWORK_NAME" >/dev/null 2>&1; then
            rootless_has_net="yes"
        fi
        # Does it exist rootful?
        if sudo -n podman network exists "$NETWORK_NAME" >/dev/null 2>&1; then
            rootful_has_net="yes"
        fi

        # Choose namespace:
        # 1) Prefer the one that already holds our project network
        # 2) If neither exists yet, honor the user's flag
        if [[ -n "$rootless_has_net" && -z "$rootful_has_net" ]]; then
            export SUDO_PREFIX=""
        elif [[ -z "$rootless_has_net" && -n "$rootful_has_net" ]]; then
            export SUDO_PREFIX="sudo "
        else
            # nothing running yet
            [[ "$want_sudo" == "true" ]] && export SUDO_PREFIX="sudo " || export SUDO_PREFIX=""
        fi

        export CONTAINER_CMD="${SUDO_PREFIX}podman"

        # Prefer native 'podman compose' if available under the chosen namespace
        if ${SUDO_PREFIX}podman compose version >/dev/null 2>&1; then
            export COMPOSE_CMD="${SUDO_PREFIX}podman compose"
            export COMPOSE_PROVIDER="native"
        else
            export COMPOSE_CMD="${SUDO_PREFIX}podman-compose"
            export COMPOSE_PROVIDER="external"
        fi

        export PODMAN_COMPOSE_WARNING_LOGS="false"
        export COMPOSE_IGNORE_ORPHANS="true"
        export PODMAN_USERNS="keep-id"

    else
        # Docker: keep your current behavior
        if [[ "$want_sudo" == "true" ]]; then
            export SUDO_PREFIX="sudo "
        else
            export SUDO_PREFIX=""
        fi
        export CONTAINER_CMD="${SUDO_PREFIX}docker"
        export COMPOSE_CMD="${SUDO_PREFIX}docker compose"
        export COMPOSE_PROVIDER="native"
    fi
}

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Function to handle errors consistently
handle_error() {
    local message="$1"
    local exit_code="${2:-1}"
    printf "\033[31mERR\033[0m %s\n" "$message" >&2
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

# Wait for service health
wait_for_service_health() {
    local service_name="$1"
    local timeout="${2:-60}"
    local counter=0

    while [[ $counter -lt $timeout ]]; do
        local hs=$(eval "$CONTAINER_CMD inspect --format='{{.State.Health.Status}}' $service_name" 2>/dev/null || echo "unknown")
        case "$hs" in
            "healthy") return 0 ;;
            "unhealthy") print_status "warning" "$service_name unhealthy"; return 1 ;;
        esac
        sleep 2
        counter=$((counter + 2))
    done
    return 1
}

# Wait for MongoDB (no built-in health check)
wait_for_mongodb() {
    local timeout="${1:-60}"
    local counter=0

    while [[ $counter -lt $timeout ]]; do
        eval "$CONTAINER_CMD exec mongodb mongosh --quiet --eval 'db.adminCommand(\"ping\")' $ERROR_REDIRECT" && return 0
        eval "$CONTAINER_CMD ps --format '{{.Names}}'" 2>/dev/null | grep -q '^mongodb$' || { print_status "error" "mongodb not running"; return 1; }
        sleep 2
        counter=$((counter + 2))
    done
    return 1
}

# Create network if it doesn't exist
ensure_network_exists() {
    if [[ "$USE_PODMAN" == "true" ]]; then
        eval "$CONTAINER_CMD network exists \"$NETWORK_NAME\" $ERROR_REDIRECT" && return 0
    else
        eval "$CONTAINER_CMD network ls --format '{{.Name}}'" 2>/dev/null | grep -q "^$NETWORK_NAME$" && return 0
    fi
    eval "$CONTAINER_CMD network create --driver bridge \"$NETWORK_NAME\" $ERROR_REDIRECT" || { print_status "error" "network create failed"; return 1; }
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
[[ ! -d "$COMPOSE_DIR" ]] && handle_error "Compose directory not found: $COMPOSE_DIR"

# Initialize service discovery (silent)
init_service_categories