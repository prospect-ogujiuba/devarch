#!/bin/zsh

# =============================================================================
# MICROSERVICES SHUTDOWN SCRIPT
# =============================================================================
# Enhanced service shutdown with selective control, cleanup options,
# and graceful dependency handling

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_categories_only=""
opt_exclude_categories=""
opt_remove_images=false
opt_remove_volumes=false
opt_remove_networks=false
opt_force_stop=false
opt_graceful_timeout=30
opt_dry_run=false
opt_verbose=false
opt_cleanup_orphans=true
opt_services_only=""              # Comma-separated list of specific services
opt_except_services=""            # Comma-separated list of services to exclude
opt_preserve_volumes=false        # Preserve volumes (opposite of remove_volumes)
opt_cleanup_service_images=false  # Only remove images for stopped services
opt_cleanup_service_volumes=false # Only remove volumes for stopped services
opt_preserve_data=false           # Never touch named volumes with data
opt_cleanup_older_than=""         # Age in days (e.g., "7d", "30d")
opt_cleanup_large_volumes=false   # Remove large unused volumes
opt_max_volume_size="500"         # Maximum volume size in MB to remove
opt_max_volumes_remove="3"        # Maximum number of volumes to remove

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Enhanced service shutdown script with selective control, cleanup options,
    and graceful dependency handling. Stops services in reverse dependency order.

OPTIONS:
    -s, --sudo                  Use sudo for container commands
    -e, --errors                Show detailed error messages
    -c, --categories LIST       Stop only specific categories (comma-separated)
    -x, --exclude LIST          Exclude specific categories (comma-separated)
    -i, --remove-images         Remove all container images after stopping
    -v, --remove-volumes        Remove all volumes (WARNING: destroys data)
    -n, --remove-networks       Remove created networks
    --services LIST             Stop only specific services (comma-separated)
    --except-services LIST      Exclude specific services from shutdown
    --preserve-volumes          Preserve all volumes (safe default)
    --preserve-data             Never touch any data volumes (safest option)
    --cleanup-service-images    Remove images only for stopped services
    --cleanup-service-volumes   Remove volumes only for stopped services
    --cleanup-orphans          Remove containers not managed by compose files
    -f, --force                Force stop containers (SIGKILL)
    -t, --timeout SECONDS      Graceful shutdown timeout (default: 30)
    -d, --dry-run              Show what would be stopped without execution
    --verbose                  Show detailed progress information
    --no-orphans               Skip cleanup of orphaned containers
    -h, --help                 Show this help message

CATEGORIES (stopped in reverse dependency order):
    proxy         - Nginx Proxy Manager
    project       - Gitea repository management
    mail          - Mailpit email testing
    ai-services   - Langflow, n8n automation platforms
    analytics     - Elasticsearch, Kibana, Logstash, Grafana, Prometheus, Matomo
    backend       - .NET, Go, Node.js, PHP, Python applications
    db-tools      - Adminer, phpMyAdmin, Mongo Express, Metabase, NocoDB, pgAdmin
    database      - MariaDB, MySQL, PostgreSQL, MongoDB, Redis

CLEANUP OPTIONS:
    --remove-images    Remove all downloaded container images
    --remove-volumes   Remove all data volumes (DESTROYS ALL DATA)
    --remove-networks  Remove created Docker networks

SMART CLEANUP OPTIONS:
    --cleanup-service-images    Remove images only for stopped services
    --cleanup-service-volumes   Remove volumes only for stopped services
    --cleanup-older-than AGE    Remove resources older than AGE (e.g., 7d, 2w)
    --cleanup-large-volumes     Remove large unused volumes to free space
    --max-volume-size SIZE      Max volume size in MB to consider for removal (default: 500)
    --max-volumes COUNT         Max number of volumes to remove (default: 3)
    --cleanup-orphans          Remove containers not managed by compose files

EXAMPLES:
    $0                                 # Stop all services gracefully
    $0 -c backend,analytics            # Stop only backend and analytics
    $0 -x database                     # Stop all except database services
    $0 -f -t 10                        # Force stop with 10s timeout
    $0 -i -v                           # Stop and cleanup images/volumes
    $0 -d --verbose                    # Dry run with detailed output
    $0 --remove-volumes                # Stop all and remove data volumes
    $0 --services postgres,redis           # Stop only postgres and redis
    $0 --categories database --preserve-volumes  # Stop database but keep data
    $0 --services nginx --cleanup-service-images # Stop nginx and clean its images
    $0 --except-services postgres --preserve-data # Stop all except postgres, preserve data

SHUTDOWN ORDER:
    Services are stopped in reverse dependency order to prevent
    connection errors and ensure graceful shutdown.

NOTES:
    - Use --remove-volumes with extreme caution (destroys all data)
    - Force stop may cause data corruption in databases
    - Dry run shows exactly what would be executed
    - Orphan cleanup removes containers not managed by compose files
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
            -c|--categories)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_categories_only="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a comma-separated list of categories"
                fi
                ;;
            --services)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_services_only="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a comma-separated list of services"
                fi
                ;;
            --except-services)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_except_services="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a comma-separated list of services"
                fi
                ;;
            --preserve-volumes)
                opt_preserve_volumes=true
                opt_remove_volumes=false  # Override any previous --remove-volumes
                shift
                ;;
            --cleanup-service-images)
                opt_cleanup_service_images=true
                shift
                ;;
            --cleanup-service-volumes)
                opt_cleanup_service_volumes=true
                shift
                ;;
            --preserve-data)
                opt_preserve_data=true
                opt_remove_volumes=false
                opt_cleanup_service_volumes=false
                shift
                ;;
            --cleanup-older-than)
                if [[ -n "$2" && "$2" != -* ]]; then
                    # Parse format like "7d", "30d", "1w" 
                    local age_value="${2%d}"  # Remove 'd' suffix
                    age_value="${age_value%w}"  # Remove 'w' suffix
                    if [[ "$2" == *"w" ]]; then
                        age_value=$((age_value * 7))  # Convert weeks to days
                    fi
                    opt_cleanup_older_than="$age_value"
                    shift 2
                else
                    handle_error "Option $1 requires an age value (e.g., 7d, 2w)"
                fi
                ;;
            --cleanup-large-volumes)
                opt_cleanup_large_volumes=true
                shift
                ;;
            --max-volume-size)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_max_volume_size="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric size in MB"
                fi
                ;;
            --max-volumes)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_max_volumes_remove="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric count"
                fi
                ;;
            -x|--exclude)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_exclude_categories="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a comma-separated list of categories"
                fi
                ;;
            -i|--remove-images)
                opt_remove_images=true
                shift
                ;;
            -v|--remove-volumes)
                opt_remove_volumes=true
                shift
                ;;
            -n|--remove-networks)
                opt_remove_networks=true
                shift
                ;;
            -f|--force)
                opt_force_stop=true
                shift
                ;;
            -t|--timeout)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_graceful_timeout="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric timeout value"
                fi
                ;;
            -d|--dry-run)
                opt_dry_run=true
                opt_verbose=true
                shift
                ;;
            --verbose)
                opt_verbose=true
                shift
                ;;
            --no-orphans)
                opt_cleanup_orphans=false
                shift
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
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_categories() {
    if [[ -n "$opt_categories_only" ]]; then
        local -a requested_categories
        requested_categories=(${(s:,:)opt_categories_only})
        
        for category in "${requested_categories[@]}"; do
            if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
                handle_error "Invalid category: $category. Available: ${(k)SERVICE_CATEGORIES}"
            fi
        done
        
        print_status "info" "Will stop categories: ${requested_categories[*]}"
    fi
    
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        for category in "${excluded_categories[@]}"; do
            if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
                handle_error "Invalid exclude category: $category. Available: ${(k)SERVICE_CATEGORIES}"
            fi
        done
        
        print_status "info" "Will exclude categories: ${excluded_categories[*]}"
    fi
}

# Function to validate individual services (similar to start-services.sh)
validate_individual_services() {
    # Validate --services option
    if [[ -n "$opt_services_only" ]]; then
        local -a requested_services
        requested_services=(${(s:,:)opt_services_only})
        
        for service in "${requested_services[@]}"; do
            if ! validate_service_exists "$service"; then
                print_status "error" "Service '$service' not found"
                print_status "info" "Available services: $(list_all_service_names | tr '\n' ' ')"
                exit 1
            fi
        done
        
        print_status "info" "Will stop services: ${requested_services[*]}"
    fi
    
    # Validate --except-services option
    if [[ -n "$opt_except_services" ]]; then
        local -a excluded_services
        excluded_services=(${(s:,:)opt_except_services})
        
        for service in "${excluded_services[@]}"; do
            if ! validate_service_exists "$service"; then
                print_status "warning" "Excluded service '$service' not found (ignoring)"
            fi
        done
        
        print_status "info" "Will exclude services: ${excluded_services[*]}"
    fi
    
    # Validate conflicting options
    if [[ -n "$opt_services_only" && -n "$opt_categories_only" ]]; then
        handle_error "Cannot use both --services and --categories options together"
    fi
}

# Function to filter services for shutdown (similar to start-services.sh)
filter_services_for_shutdown() {
    local category="$1"
    local service_files
    service_files=$(get_service_files "$category")
    
    local -a all_files filtered_files
    all_files=(${=service_files})
    
    # If --services specified, only include those services that belong to this category
    if [[ -n "$opt_services_only" ]]; then
        local -a requested_services
        requested_services=(${(s:,:)opt_services_only})
        
        for service_file in "${all_files[@]}"; do
            local service_name="${service_file%.yml}"
            if [[ " ${requested_services[*]} " =~ " $service_name " ]]; then
                filtered_files+=("$service_file")
            fi
        done
    else
        # Start with all files in category
        filtered_files=("${all_files[@]}")
    fi
    
    # Remove excluded services
    if [[ -n "$opt_except_services" ]]; then
        local -a excluded_services final_files
        excluded_services=(${(s:,:)opt_except_services})
        
        for service_file in "${filtered_files[@]}"; do
            local service_name="${service_file%.yml}"
            if [[ ! " ${excluded_services[*]} " =~ " $service_name " ]]; then
                final_files+=("$service_file")
            fi
        done
        filtered_files=("${final_files[@]}")
    fi
    
    # Return filtered list
    echo "${filtered_files[*]}"
}

# Enhanced service stopping function
stop_individual_service() {
    local service_file="$1"
    local category="$2"
    
    local service_name="${service_file%.yml}"
    local service_path
    service_path=$(resolve_service_path "$service_file" "$category")
    
    if [[ ! -f "$service_path" ]]; then
        print_status "warning" "Service file not found: $service_file"
        return 1
    fi
    
    # Determine volume removal strategy
    local remove_volumes="false"
    if [[ "$opt_preserve_data" == "true" || "$opt_preserve_volumes" == "true" ]]; then
        remove_volumes="false"
    elif [[ "$opt_remove_volumes" == "true" ]]; then
        remove_volumes="true"
    fi
    
    # Stop the service
    stop_single_service "$service_name" "$remove_volumes" "$opt_graceful_timeout"
    local stop_result=$?
    
    # Smart cleanup for this specific service
    if [[ $stop_result -eq 0 ]]; then
        cleanup_service_resources "$service_name"
    fi
    
    return $stop_result
}

# Function to cleanup resources for a specific service
cleanup_service_resources() {
    local service_name="$1"
    
    if [[ "$opt_dry_run" == "true" ]]; then
        if [[ "$opt_cleanup_service_images" == "true" ]]; then
            echo "DRY RUN: Remove images for service: $service_name"
        fi
        if [[ "$opt_cleanup_service_volumes" == "true" ]]; then
            echo "DRY RUN: Remove volumes for service: $service_name"
        fi
        return 0
    fi
    
    # Cleanup service-specific images
    if [[ "$opt_cleanup_service_images" == "true" ]]; then
        cleanup_service_images "$service_name"
    fi
    
    # Cleanup service-specific volumes
    if [[ "$opt_cleanup_service_volumes" == "true" && "$opt_preserve_data" == "false" ]]; then
        cleanup_service_volumes "$service_name"
    fi
}

# Function to cleanup images for a specific service
cleanup_service_images() {
    local service_name="$1"
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Cleaning up images for service: $service_name"
    fi
    
    # Get images related to this service
    local images
    images=$(eval "$CONTAINER_CMD images --filter 'label=com.docker.compose.service=$service_name' -q" 2>/dev/null || echo "")
    
    if [[ -n "$images" ]]; then
        local image_list
        image_list=(${(f)images})
        
        local removed=0
        for image in "${image_list[@]}"; do
            if eval "$CONTAINER_CMD rmi -f $image $ERROR_REDIRECT"; then
                ((removed++))
            fi
        done
        
        if [[ "$opt_verbose" == "true" && $removed -gt 0 ]]; then
            print_status "success" "Removed $removed image(s) for $service_name"
        fi
    fi
}

# Function to cleanup volumes for a specific service
cleanup_service_volumes() {
    local service_name="$1"
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Cleaning up volumes for service: $service_name"
    fi
    
    # Get volumes related to this service
    local volumes
    volumes=$(eval "$CONTAINER_CMD volume ls --filter 'label=com.docker.compose.service=$service_name' -q" 2>/dev/null || echo "")
    
    if [[ -n "$volumes" ]]; then
        local volume_list
        volume_list=(${(f)volumes})
        
        local removed=0
        for volume in "${volume_list[@]}"; do
            if eval "$CONTAINER_CMD volume rm -f $volume $ERROR_REDIRECT"; then
                ((removed++))
            fi
        done
        
        if [[ "$opt_verbose" == "true" && $removed -gt 0 ]]; then
            print_status "success" "Removed $removed volume(s) for $service_name"
        fi
    fi
}

validate_destructive_operations() {
    if [[ "$opt_remove_volumes" == "true" && "$opt_dry_run" == "false" ]]; then
        print_status "warning" "WARNING: --remove-volumes will destroy ALL data!"
        echo "This will permanently delete:"
        echo "  - Database data (PostgreSQL, MySQL, MongoDB)"
        echo "  - Application files and configurations"
        echo "  - Log files and user uploads"
        echo ""
        
        if [[ -t 0 ]]; then  # Check if running interactively
            read "response?Are you absolutely sure? Type 'DELETE ALL DATA' to confirm: "
            if [[ "$response" != "DELETE ALL DATA" ]]; then
                print_status "info" "Volume removal cancelled"
                opt_remove_volumes=false
            fi
        else
            print_status "error" "Non-interactive mode: volume removal requires explicit confirmation"
            exit 1
        fi
    fi
}

# =============================================================================
# SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

get_shutdown_categories() {
    local -a categories_to_stop
    
    if [[ -n "$opt_categories_only" ]]; then
        local -a requested_categories
        requested_categories=(${(s:,:)opt_categories_only})
        
        # Reverse the startup order and filter
        local -a reversed_order
        for ((i = ${#SERVICE_STARTUP_ORDER[@]}; i > 0; i--)); do
            reversed_order+=("${SERVICE_STARTUP_ORDER[i]}")
        done
        
        for category in "${reversed_order[@]}"; do
            if [[ " ${requested_categories[*]} " =~ " $category " ]]; then
                categories_to_stop+=("$category")
            fi
        done
    else
        # Reverse the startup order for shutdown
        for ((i = ${#SERVICE_STARTUP_ORDER[@]}; i > 0; i--)); do
            categories_to_stop+=("${SERVICE_STARTUP_ORDER[i]}")
        done
    fi
    
    # Remove excluded categories
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        local -a filtered_categories
        for category in "${categories_to_stop[@]}"; do
            if [[ ! " ${excluded_categories[*]} " =~ " $category " ]]; then
                filtered_categories+=("$category")
            fi
        done
        categories_to_stop=("${filtered_categories[@]}")
    fi
    
    echo "${categories_to_stop[@]}"
}

stop_service_file() {
    local service_file="$1"
    local category="$2"
    local full_path="$(resolve_service_path "$service_file" "$category")"
    
    if [[ ! -f "$full_path" ]]; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "warning" "Service file not found: $service_file"
        fi
        return 1
    fi
    
    local service_name="${service_file%.yml}"
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Stopping $service_name from $service_file..."
    fi
    
    # Build compose command
    local compose_args=("-f" "$full_path" "down")
    
    # Add timeout if not default
    if [[ "$opt_graceful_timeout" != "30" ]]; then
        compose_args+=("--timeout" "$opt_graceful_timeout")
    fi
    
    # Add remove orphans
    if [[ "$opt_cleanup_orphans" == "true" ]]; then
        compose_args+=("--remove-orphans")
    fi
    
    # Execute or show dry run
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $COMPOSE_CMD ${compose_args[*]}"
        return 0
    fi
    
    # Execute the command
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "success" "$service_name stopped successfully"
        fi
        return 0
    else
        print_status "warning" "Failed to stop $service_name (may already be stopped)"
        return 1
    fi
}

stop_category() {
    local category="$1"
    local service_files
    
    # Use filtered services instead of all services
    service_files=$(filter_services_for_shutdown "$category")
    
    if [[ -z "$service_files" ]]; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "info" "No services to stop in $category (filtered out)"
        fi
        return 0
    fi
    
    print_status "step" "Stopping $category services..."
    
    local -a files
    files=(${=service_files})
    
    local failed=0
    for service_file in "${files[@]}"; do
        if ! stop_individual_service "$service_file" "$category"; then
            ((failed++))
        fi
    done
    
    if [[ $failed -eq 0 ]]; then
        print_status "success" "$category services stopped successfully"
    else
        print_status "warning" "$failed service(s) in $category had issues stopping"
    fi
}

force_stop_containers() {
    if [[ "$opt_force_stop" == "false" || "$opt_dry_run" == "true" ]]; then
        return 0
    fi
    
    print_status "step" "Force stopping any remaining containers..."
    
    # Get all running containers related to our services
    local containers
    containers=$(eval "$CONTAINER_CMD ps -q --filter 'network=$NETWORK_NAME' $ERROR_REDIRECT" || echo "")
    
    if [[ -n "$containers" ]]; then
        local container_list
        container_list=(${(f)containers})
        
        for container in "${container_list[@]}"; do
            if [[ "$opt_verbose" == "true" ]]; then
                local name
                name=$(eval "$CONTAINER_CMD inspect --format='{{.Name}}' $container 2>/dev/null" | sed 's/^///' || echo "unknown")
                print_status "warning" "Force stopping container: $name"
            fi
            
            eval "$CONTAINER_CMD kill $container $ERROR_REDIRECT" || true
        done
        
        print_status "success" "Force stopped remaining containers"
    else
        print_status "info" "No containers to force stop"
    fi
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

cleanup_images() {
    if [[ "$opt_remove_images" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Removing container images..."
    
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $CONTAINER_CMD rmi \$($CONTAINER_CMD images -q)"
        return 0
    fi
    
    local images
    images=$(eval "$CONTAINER_CMD images -q $ERROR_REDIRECT" || echo "")
    
    if [[ -n "$images" ]]; then
        local image_list
        image_list=(${(f)images})
        
        local removed=0
        for image in "${image_list[@]}"; do
            if eval "$CONTAINER_CMD rmi -f $image $ERROR_REDIRECT"; then
                ((removed++))
            fi
        done
        
        print_status "success" "Removed $removed container images"
    else
        print_status "info" "No images to remove"
    fi
}

cleanup_volumes() {
    if [[ "$opt_remove_volumes" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Removing data volumes..."
    
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $CONTAINER_CMD volume rm \$($CONTAINER_CMD volume ls -q)"
        return 0
    fi
    
    local volumes
    volumes=$(eval "$CONTAINER_CMD volume ls -q $ERROR_REDIRECT" || echo "")
    
    if [[ -n "$volumes" ]]; then
        local volume_list
        volume_list=(${(f)volumes})
        
        local removed=0
        for volume in "${volume_list[@]}"; do
            if eval "$CONTAINER_CMD volume rm -f $volume $ERROR_REDIRECT"; then
                ((removed++))
            fi
        done
        
        print_status "success" "Removed $removed data volumes"
    else
        print_status "info" "No volumes to remove"
    fi
}

cleanup_networks() {
    if [[ "$opt_remove_networks" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Removing networks..."
    
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $CONTAINER_CMD network rm $NETWORK_NAME"
        return 0
    fi
    
    # Remove our specific network
    if eval "$CONTAINER_CMD network exists $NETWORK_NAME $ERROR_REDIRECT"; then
        if eval "$CONTAINER_CMD network rm $NETWORK_NAME $ERROR_REDIRECT"; then
            print_status "success" "Removed network: $NETWORK_NAME"
        else
            print_status "warning" "Failed to remove network: $NETWORK_NAME"
        fi
    else
        print_status "info" "Network $NETWORK_NAME does not exist"
    fi
    
    # Cleanup unused networks
    eval "$CONTAINER_CMD network prune -f $ERROR_REDIRECT" || true
}

cleanup_system() {
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $CONTAINER_CMD system prune -f"
        return 0
    fi
    
    print_status "step" "Cleaning up system resources..."
    
    # Prune unused containers, networks, and build cache
    eval "$CONTAINER_CMD system prune -f $ERROR_REDIRECT" || true
    
    print_status "success" "System cleanup completed"
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_shutdown_summary() {
    local -a categories_to_stop
    categories_to_stop=($(get_shutdown_categories))
    
    print_status "info" "Service Shutdown Summary:"
    echo "  Container Runtime: $CONTAINER_RUNTIME"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Show Errors: $opt_show_errors"
    echo "  Force Stop: $opt_force_stop"
    echo "  Graceful Timeout: ${opt_graceful_timeout}s"
    echo "  Dry Run: $opt_dry_run"
    echo ""
    
    # Show what will be affected
    if [[ -n "$opt_services_only" ]]; then
        echo "  Target: Specific services only"
        echo "  Services: $opt_services_only"
    elif [[ -n "$opt_categories_only" ]]; then
        echo "  Target: Specific categories only"
        echo "  Categories: $opt_categories_only"
    else
        echo "  Target: All services"
        echo "  Categories: ${categories_to_stop[*]}"
    fi
    
    if [[ -n "$opt_except_services" ]]; then
        echo "  Excluding services: $opt_except_services"
    fi
    
    if [[ -n "$opt_exclude_categories" ]]; then
        echo "  Excluding categories: $opt_exclude_categories"
    fi
    
    echo ""
    echo "  Data Safety:"
    if [[ "$opt_preserve_data" == "true" ]]; then
        echo "    🛡️  PRESERVE DATA: All data volumes will be preserved"
    elif [[ "$opt_preserve_volumes" == "true" ]]; then
        echo "    🛡️  PRESERVE VOLUMES: All volumes will be preserved"
    elif [[ "$opt_remove_volumes" == "true" ]]; then
        echo "    ⚠️  REMOVE VOLUMES: Data volumes will be destroyed"
    else
        echo "    ✅ SAFE DEFAULT: Volumes will be preserved"
    fi
    
    echo ""
    echo "  Cleanup Operations:"
    [[ "$opt_remove_images" == "true" ]] && echo "    🗑️  Remove all container images"
    [[ "$opt_cleanup_service_images" == "true" ]] && echo "    🗑️  Remove images for stopped services only"
    [[ "$opt_cleanup_service_volumes" == "true" ]] && echo "    🗑️  Remove volumes for stopped services only"
    [[ "$opt_remove_networks" == "true" ]] && echo "    🗑️  Remove networks"
    [[ "$opt_cleanup_orphans" == "true" ]] && echo "    🧹 Clean up orphaned containers"
    
    if [[ "$opt_remove_images" == "false" && "$opt_cleanup_service_images" == "false" && 
          "$opt_remove_volumes" == "false" && "$opt_cleanup_service_volumes" == "false" && 
          "$opt_remove_networks" == "false" ]]; then
        echo "    ✅ No cleanup operations (containers only)"
    fi
    
    echo ""
}

run_shutdown_process() {
    local -a categories_to_stop
    categories_to_stop=($(get_shutdown_categories))
    
    if [[ ${#categories_to_stop[@]} -eq 0 ]]; then
        print_status "warning" "No categories selected for shutdown"
        return 0
    fi
    
    print_status "step" "Stopping microservices in reverse dependency order..."
    
    # Stop each category
    for category in "${categories_to_stop[@]}"; do
        local service_files
        service_files=$(filter_services_for_shutdown "$category")
        
        # Skip categories with no services (after filtering)
        if [[ -z "$service_files" ]]; then
            if [[ "$opt_verbose" == "true" ]]; then
                print_status "info" "Skipping $category (no services to stop after filtering)"
            fi
            continue
        fi
        
        # Stop services in category
        stop_category "$category"
        
        # Brief pause between categories
        if [[ "$opt_dry_run" == "false" ]]; then
            sleep 1
        fi
    done
    
    # Force stop any remaining containers if requested
    force_stop_containers
    
    # Run cleanup operations
    run_cleanup_operations
    
    print_status "success" "Service shutdown process completed!"
}

run_cleanup_operations() {
    print_status "step" "Running smart cleanup operations..."
    
    # Age-based cleanup
    if [[ -n "$opt_cleanup_older_than" ]]; then
        cleanup_old_resources "$opt_cleanup_older_than" "$opt_dry_run"
    fi
    
    # Large volume cleanup
    if [[ "$opt_cleanup_large_volumes" == "true" ]]; then
        cleanup_large_volumes "$opt_max_volume_size" "$opt_max_volumes_remove" "$opt_dry_run"
    fi
    
    # Service-specific orphan cleanup
    if [[ "$opt_cleanup_orphans" == "true" ]]; then
        local target_services=""
        if [[ -n "$opt_services_only" ]]; then
            target_services="$opt_services_only"
        fi
        cleanup_service_orphans "$target_services" "$opt_dry_run"
    fi
    
    # Existing global cleanup (images, volumes, networks)
    cleanup_images
    cleanup_volumes  
    cleanup_networks
    cleanup_system
    
    print_status "success" "Smart cleanup operations completed"
}

show_completion_info() {
    if [[ "$opt_dry_run" == "true" ]]; then
        echo ""
        print_status "info" "Dry run completed. No services were actually stopped."
        echo "Run without -d/--dry-run to execute the shutdown process."
        return 0
    fi
    
    echo ""
    print_status "info" "Shutdown completed!"
    echo ""
    
    # Show what was actually done
    if [[ -n "$opt_services_only" ]]; then
        echo "🛑 Stopped specific services: $opt_services_only"
    elif [[ -n "$opt_categories_only" ]]; then
        echo "🛑 Stopped categories: $opt_categories_only"
    else
        echo "🛑 Stopped all services"
    fi
    
    if [[ "$opt_preserve_data" == "true" || "$opt_preserve_volumes" == "true" ]]; then
        echo "🛡️  Data preserved: All volumes kept safe"
    elif [[ "$opt_remove_volumes" == "true" ]]; then
        echo "⚠️  Data removed: Volumes were destroyed"
    fi
    
    echo ""
    echo "🔄 Management commands:"
    echo "  - Start services: $SCRIPT_DIR/start-services.sh"
    echo "  - Start specific: $SCRIPT_DIR/start-services.sh --services [service,...]"
    echo "  - Check status: $SCRIPT_DIR/service-manager.sh status"
    echo "  - View services: $SCRIPT_DIR/show-services.sh"
    echo ""
}

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Validate environment and arguments
    validate_categories
    validate_individual_services
    validate_destructive_operations
    
    # Show shutdown summary
    show_shutdown_summary
    
    # Run the shutdown process
    run_shutdown_process
    
    # Show completion information
    show_completion_info
    
    print_status "success" "Service shutdown script completed!"
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi