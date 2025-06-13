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
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -c, --categories LIST   Stop only specific categories (comma-separated)
    -x, --exclude LIST      Exclude specific categories (comma-separated)
    -i, --remove-images     Remove all container images after stopping
    -v, --remove-volumes    Remove all volumes (WARNING: destroys data)
    -n, --remove-networks   Remove created networks
    -f, --force             Force stop containers (SIGKILL)
    -t, --timeout SECONDS   Graceful shutdown timeout (default: 30)
    -d, --dry-run           Show what would be stopped without execution
    --verbose               Show detailed progress information
    --no-orphans           Skip cleanup of orphaned containers
    -h, --help              Show this help message

CATEGORIES (stopped in reverse dependency order):
    proxy         - Nginx Proxy Manager
    auth          - Keycloak identity provider
    erp           - Odoo business applications
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

EXAMPLES:
    $0                                  # Stop all services gracefully
    $0 -c backend,analytics            # Stop only backend and analytics
    $0 -x database                     # Stop all except database services
    $0 -f -t 10                        # Force stop with 10s timeout
    $0 -i -v                           # Stop and cleanup images/volumes
    $0 -d --verbose                    # Dry run with detailed output
    $0 --remove-volumes                # Stop all and remove data volumes

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
    local full_path="$COMPOSE_DIR/$service_file"
    
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
    service_files=$(get_service_files "$category")
    
    if [[ -z "$service_files" ]]; then
        print_status "warning" "No services found for category: $category"
        return 0
    fi
    
    print_status "step" "Stopping $category services..."
    
    local -a files
    files=(${=service_files})
    
    local failed=0
    for service_file in "${files[@]}"; do
        if ! stop_service_file "$service_file" "$category"; then
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
    echo "  Remove Images: $opt_remove_images"
    echo "  Remove Volumes: $opt_remove_volumes"
    echo "  Remove Networks: $opt_remove_networks"
    echo "  Cleanup Orphans: $opt_cleanup_orphans"
    echo "  Dry Run: $opt_dry_run"
    echo ""
    echo "  Categories to stop: ${categories_to_stop[*]}"
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
        stop_category "$category"
        
        # Brief pause between categories
        if [[ "$opt_dry_run" == "false" ]]; then
            sleep 1
        fi
    done
    
    # Force stop any remaining containers
    force_stop_containers
    
    # Cleanup operations
    cleanup_images
    cleanup_volumes
    cleanup_networks
    cleanup_system
    
    print_status "success" "Service shutdown process completed!"
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
    echo "🔍 Check remaining services:"
    echo "  $SCRIPT_DIR/show-services.sh"
    echo ""
    echo "🚀 Restart services:"
    echo "  $SCRIPT_DIR/start-services.sh"
    echo ""
    
    if [[ "$opt_remove_volumes" == "true" ]]; then
        echo "⚠️  Data volumes were removed - all data has been deleted!"
        echo "   Run setup scripts to reinitialize:"
        echo "   $SCRIPT_DIR/setup-databases.sh -a"
        echo ""
    fi
}

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Validate arguments and warn about destructive operations
    validate_categories
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