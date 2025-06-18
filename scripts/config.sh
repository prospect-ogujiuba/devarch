#!/bin/zsh

# =============================================================================
# MICROSERVICES CONFIGURATION MANAGEMENT - ENHANCED
# =============================================================================
# Smart configuration with automatic path resolution and flexible overrides

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
# SERVICE CATEGORIES & COMPOSE FILES - SMART CONFIGURATION
# =============================================================================

# Smart service definitions - automatically resolves paths based on category
typeset -A SERVICE_CATEGORIES
SERVICE_CATEGORIES=(
    [proxy]="traefik.yml nginx-proxy-manager.yml"
    [database]="mariadb.yml mysql.yml postgres.yml mongodb.yml redis.yml"
    [exporters]="blackbox-exporter.yml mongodb-exporter.yml mysqld-exporter.yml node-exporter.yml postgres-exporter.yml redis-exporter.yml"
    [dbms]="adminer.yml phpmyadmin.yml mongo-express.yml metabase.yml nocodb.yml pgadmin.yml redis-commander.yml"
    [backend]="php.yml node.yml python.yml go.yml dotnet.yml"
    [analytics]="elasticsearch.yml kibana.yml logstash.yml grafana.yml prometheus.yml matomo.yml cadvisor.yml otel-collector.yml"
    [ai-services]="langflow.yml n8n.yml"
    [mail]="mailpit.yml"
    [project]="gitea.yml"
    [management]="portainer.yml"  # Add Portainer to management category
)

# Optional: Override category paths if you need different directory structure
typeset -A CATEGORY_PATH_OVERRIDES
CATEGORY_PATH_OVERRIDES=(
    # [category]="custom/path"
    # Example: [backend]="apps/backend"
    # Example: [legacy]="old-services"
    # Example: [management]="infrastructure/management"
)

# Optional: Full path overrides for specific services (for maximum flexibility)
typeset -A SERVICE_PATH_OVERRIDES
SERVICE_PATH_OVERRIDES=(
    # [service.yml]="full/custom/path/service.yml"
    # Example: [special-service.yml]="legacy/docker-compose.yml"
    # Example: [nginx.yml]="infrastructure/nginx/docker-compose.yml"
    # Example: [portainer.yml]="infrastructure/portainer/docker-compose.yml"
)

# Service startup order (critical for dependencies) - zsh array
SERVICE_STARTUP_ORDER=(
    "management"   # Add management first - Portainer should start early for container oversight
    "proxy"
    "database"
    "exporters"
    "dbms" 
    "backend"
    "analytics"
    "ai-services"
    "mail"
    "project"
)

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

# Function to validate all service files exist
validate_service_files() {
    local category="$1"
    local -a missing_files valid_files
    
    if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
        print_status "error" "Unknown service category: $category"
        return 1
    fi
    
    local service_files="${SERVICE_CATEGORIES[$category]}"
    local -a files
    files=(${=service_files})
    
    print_status "step" "Validating $category service files..."
    
    for service_file in "${files[@]}"; do
        local resolved_path
        resolved_path=$(resolve_service_path "$service_file" "$category")
        
        if [[ -f "$resolved_path" ]]; then
            valid_files+=("$service_file")
            if [[ "$VERBOSE_VALIDATION" == "true" ]]; then
                print_status "info" "✓ Found: $service_file -> $resolved_path"
            fi
        else
            missing_files+=("$service_file")
            print_status "warning" "❌ Missing: $service_file (expected at $resolved_path)"
        fi
    done
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        print_status "warning" "Category '$category' has ${#missing_files[@]} missing file(s): ${missing_files[*]}"
        print_status "info" "Available files: ${valid_files[*]}"
        return 1
    else
        print_status "success" "All $category service files found (${#valid_files[@]} files)"
        return 0
    fi
}

# =============================================================================
# ENHANCED SERVICE RESOLUTION FUNCTIONS
# =============================================================================
# Add these functions to config.sh after the existing resolve_service_path function

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

# =============================================================================
# CORE SERVICE MANAGEMENT FUNCTIONS
# =============================================================================
# Add these functions to config.sh after the enhanced service resolution functions

# Function to start a single service
start_single_service() {
    local service_name="$1"
    local force_recreate="${2:-false}"
    local service_path
    
    print_status "step" "Starting service: $service_name"
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        print_status "error" "Service '$service_name' not found"
        print_status "info" "Available services: $(list_all_service_names | tr '\n' ' ')"
        return 1
    fi
    
    # Build compose command
    local compose_args=("-f" "$service_path" "up" "-d")
    if [[ "$force_recreate" == "true" ]]; then
        compose_args+=("--force-recreate")
    fi
    
    # Execute command
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "success" "Service '$service_name' started successfully"
        return 0
    else
        print_status "error" "Failed to start service '$service_name'"
        return 1
    fi
}

# Function to stop a single service
stop_single_service() {
    local service_name="$1"
    local remove_volumes="${2:-false}"
    local timeout="${3:-30}"
    local service_path
    
    print_status "step" "Stopping service: $service_name"
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        print_status "error" "Service '$service_name' not found"
        return 1
    fi
    
    # Build compose command
    local compose_args=("-f" "$service_path" "down")
    compose_args+=("--timeout" "$timeout")
    
    if [[ "$remove_volumes" == "true" ]]; then
        compose_args+=("--volumes")
    fi
    
    # Execute command
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "success" "Service '$service_name' stopped successfully"
        return 0
    else
        print_status "warning" "Service '$service_name' may have already been stopped"
        return 1
    fi
}

# Function to rebuild a single service
rebuild_single_service() {
    local service_name="$1"
    local no_cache="${2:-false}"
    local service_path
    
    print_status "step" "Rebuilding service: $service_name"
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        print_status "error" "Service '$service_name' not found"
        return 1
    fi
    
    # Stop service first
    print_status "info" "Stopping service for rebuild..."
    stop_single_service "$service_name" "false" "10"
    
    # Build compose command
    local compose_args=("-f" "$service_path" "build")
    if [[ "$no_cache" == "true" ]]; then
        compose_args+=("--no-cache")
    fi
    
    # Build the service
    print_status "info" "Building service..."
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "info" "Starting rebuilt service..."
        start_single_service "$service_name" "true"
        return $?
    else
        print_status "error" "Failed to rebuild service '$service_name'"
        return 1
    fi
}

# Function to restart a single service
restart_single_service() {
    local service_name="$1"
    local service_path
    
    print_status "step" "Restarting service: $service_name"
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        print_status "error" "Service '$service_name' not found"
        return 1
    fi
    
    # Use compose restart for graceful restart
    if eval "$COMPOSE_CMD -f '$service_path' restart $ERROR_REDIRECT"; then
        print_status "success" "Service '$service_name' restarted successfully"
        return 0
    else
        print_status "warning" "Compose restart failed, trying stop/start..."
        stop_single_service "$service_name" "false" "10"
        start_single_service "$service_name" "false"
        return $?
    fi
}

# Function to show logs for a single service
show_service_logs() {
    local service_name="$1"
    local follow="${2:-false}"
    local tail_lines="${3:-100}"
    local service_path
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        print_status "error" "Service '$service_name' not found"
        return 1
    fi
    
    # Build logs command
    local compose_args=("-f" "$service_path" "logs" "--tail" "$tail_lines")
    if [[ "$follow" == "true" ]]; then
        compose_args+=("-f")
        print_status "info" "Following logs for '$service_name' (Ctrl+C to exit)..."
    else
        print_status "info" "Showing last $tail_lines lines for '$service_name'..."
    fi
    
    # Execute logs command (don't redirect - we want to see the logs)
    eval "$COMPOSE_CMD ${compose_args[*]}"
}

# Function to get service status
get_service_status() {
    local service_name="$1"
    local service_path
    
    # Get service path
    service_path=$(get_service_path "$service_name")
    if [[ $? -ne 0 ]]; then
        echo "NOT_FOUND"
        return 1
    fi
    
    # Get container name from service
    local container_name="$service_name"
    
    # Check if container exists and get status
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' $container_name 2>/dev/null")
    
    if [[ $? -eq 0 && -n "$container_status" ]]; then
        echo "$container_status"
        return 0
    else
        echo "STOPPED"
        return 1
    fi
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
                # Use a simple approach without persistent variables
                local file_path="$(resolve_service_path "$service_file" "$category" 2>/dev/null)"
                local status_icon="❌ MISSING"
                
                [[ -f "$file_path" ]] && status_icon="✅ EXISTS"
                
                printf "  %-25s -> %s [%s]\n" "$service_file" "$file_path" "$status_icon"
            done
        else
            echo "  No services defined"
        fi
        echo ""
    done
}

# Enhanced start function with smart path resolution
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
    
    local started=0 failed=0
    
    for service_file in "${files[@]}"; do
        local full_path
        full_path=$(resolve_service_path "$service_file" "$category")
        
        if [[ -f "$full_path" ]]; then
            print_status "info" "Starting services from $service_file..."
            if eval "$COMPOSE_CMD -f \"$full_path\" up -d $ERROR_REDIRECT"; then
                ((started++))
            else
                print_status "error" "Failed to start services from $service_file"
                ((failed++))
            fi
        else
            print_status "warning" "Service file not found: $full_path"
            ((failed++))
        fi
    done
    
    if [[ $failed -eq 0 ]]; then
        print_status "success" "$category services started successfully ($started/$((started + failed)))"
    else
        print_status "warning" "$category services partially started ($started/$((started + failed))) - $failed failed"
    fi
    
    return $failed
}

# Enhanced stop function with smart path resolution
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
        local full_path
        full_path=$(resolve_service_path "$service_file" "$category")
        
        if [[ -f "$full_path" ]]; then
            print_status "info" "Stopping services from $service_file..."
            eval "$COMPOSE_CMD -f \"$full_path\" down $ERROR_REDIRECT" || true
        else
            print_status "warning" "Service file not found: $full_path (skipping)"
        fi
    done
    
    print_status "success" "$category services stopped"
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
# DATABASE CREDENTIALS (unchanged)
# =============================================================================

export MARIADB_ROOT_PASSWORD="123456"
export MYSQL_ROOT_PASSWORD="123456"
export POSTGRES_PASSWORD="123456"
export MONGO_ROOT_PASSWORD="123456"
export ADMIN_PASSWORD="123456"

# =============================================================================
# SSL CONFIGURATION (unchanged)
# =============================================================================

export SSL_DOMAIN="*.test"
export SSL_CERT_DIR="/etc/letsencrypt/live/wildcard.test"
export SSL_DAYS_VALID="3650"  # 10 years for development

# =============================================================================
# RUNTIME DETECTION & COMMAND SETUP (unchanged)
# =============================================================================

# Detect if we should use sudo (for Docker on Linux)
detect_sudo_requirement() {
    if command -v podman >/dev/null 2>&1; then
        export USE_PODMAN=true
        export DEFAULT_SUDO=true
    elif command -v docker >/dev/null 2>&1; then
        export USE_PODMAN=true
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

# =============================================================================
# UTILITY FUNCTIONS (unchanged but enhanced error messages)
# =============================================================================

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
    if [[ "$USE_PODMAN" == "true" ]]; then
        # Podman has 'network exists' command
        if ! eval "$CONTAINER_CMD network exists \"$NETWORK_NAME\" $ERROR_REDIRECT"; then
            print_status "step" "Creating network: $NETWORK_NAME"
            eval "$CONTAINER_CMD network create --driver bridge \"$NETWORK_NAME\" $ERROR_REDIRECT"
            check_status "Failed to create network: $NETWORK_NAME"
            print_status "success" "Network created: $NETWORK_NAME"
        else
            print_status "info" "Network already exists: $NETWORK_NAME"
        fi
    else
        # Docker doesn't have 'network exists' - use 'network ls' instead
        if ! eval "$CONTAINER_CMD network ls --format '{{.Name}}' | grep -q '^$NETWORK_NAME$' $ERROR_REDIRECT"; then
            print_status "step" "Creating network: $NETWORK_NAME"
            eval "$CONTAINER_CMD network create --driver bridge \"$NETWORK_NAME\" $ERROR_REDIRECT"
            check_status "Failed to create network: $NETWORK_NAME"
            print_status "success" "Network created: $NETWORK_NAME"
        else
            print_status "info" "Network already exists: $NETWORK_NAME"
        fi
    fi
}

# =============================================================================
# TRAEFIK SPECIFIC FUNCTIONS (unchanged)
# =============================================================================

# Function to validate Traefik setup
validate_traefik_setup() {
    print_status "step" "Validating Traefik setup..."
    
    # Check if Traefik container exists
    if ! eval "$CONTAINER_CMD container exists traefik $ERROR_REDIRECT"; then
        print_status "error" "Traefik container not found"
        return 1
    fi
    
    # Check if container is running
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' traefik 2>/dev/null" || echo "unknown")
    
    if [[ "$container_status" != "running" ]]; then
        print_status "error" "Traefik container is not running. Status: $container_status"
        return 1
    fi
    
    print_status "success" "Traefik container is running"
    return 0
}

# Function to get Traefik dashboard URL
get_traefik_dashboard() {
    echo "https://traefik.test"
}

# Function to get Traefik API info
get_traefik_api() {
    local api_endpoint="http://localhost:8080/api/rawdata"
    if command -v curl >/dev/null 2>&1; then
        curl -s "$api_endpoint" 2>/dev/null || echo "Traefik API not accessible"
    else
        echo "curl not available - cannot access Traefik API"
    fi
}

# Function to show Traefik routes
show_traefik_routes() {
    print_status "info" "Traefik Routes:"
    
    if validate_traefik_setup; then
        # Try to get routes from API
        local routes_json
        routes_json=$(curl -s "http://localhost:8080/api/http/routers" 2>/dev/null)
        
        if [[ $? -eq 0 && -n "$routes_json" ]]; then
            echo "$routes_json" | grep -o '"name":"[^"]*"' | sed 's/"name":"//g' | sed 's/"//g' | sort
        else
            print_status "warning" "Could not fetch routes from Traefik API"
            echo "Access the dashboard at: $(get_traefik_dashboard)"
        fi
    fi
}

# Function to reload Traefik configuration
reload_traefik_config() {
    print_status "step" "Reloading Traefik configuration..."
    
    if validate_traefik_setup; then
        # Traefik automatically reloads Docker provider configurations
        # But we can restart the container to ensure everything is fresh
        eval "$CONTAINER_CMD restart traefik $ERROR_REDIRECT"
        check_status "Failed to restart Traefik"
        
        print_status "success" "Traefik configuration reloaded"
        
        # Wait a moment for startup
        sleep 3
        
        # Show routes after reload
        show_traefik_routes
    else
        print_status "error" "Cannot reload - Traefik is not running"
        return 1
    fi
}

# =============================================================================
# OS DETECTION (unchanged)
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
# INITIALIZATION (enhanced)
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
# EXPORTED FUNCTIONS (enhanced)
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
    export -f validate_traefik_setup 2>/dev/null || true
    export -f get_traefik_dashboard 2>/dev/null || true
    export -f get_traefik_api 2>/dev/null || true
    export -f show_traefik_routes 2>/dev/null || true
    export -f reload_traefik_config 2>/dev/null || true
    # Enhanced functions
    export -f resolve_service_path 2>/dev/null || true
    export -f validate_service_files 2>/dev/null || true
    export -f list_all_services 2>/dev/null || true
    export -f add_service_override 2>/dev/null || true
    export -f add_category_override 2>/dev/null || true
    export -f show_configuration 2>/dev/null || true
fi

print_status "info" "Smart configuration loaded successfully"
print_status "info" "Container runtime: $CONTAINER_RUNTIME (sudo: $DEFAULT_SUDO)"
print_status "info" "Project root: $PROJECT_ROOT"

# Optional: Validate all service files on load (uncomment if desired)
# VERBOSE_VALIDATION=false
# for category in "${SERVICE_STARTUP_ORDER[@]}"; do
#     validate_service_files "$category" >/dev/null 2>&1
# done