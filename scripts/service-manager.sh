#!/bin/zsh

# =============================================================================
# UNIFIED MICROSERVICES MANAGER - THE ONE TOOL TO RULE THEM ALL
# =============================================================================
# Complete service management with podman terminology, dependency handling,
# bulk operations, and smart cleanup. Replaces start-services.sh and stop-services.sh

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_force_recreate=false
opt_no_cache=false
opt_remove_volumes=false
opt_follow_logs=false
opt_tail_lines=100
opt_timeout=30
opt_dry_run=false
opt_verbose=false

# Bulk operation options
opt_categories_only=""
opt_exclude_categories=""
opt_services_only=""
opt_except_services=""
opt_parallel_start=false
opt_wait_healthy=true
opt_health_timeout=120
opt_restart_policy="unless-stopped"
opt_rebuild_services=false
opt_force_services=false

# Cleanup options
opt_remove_images=false
opt_remove_networks=false
opt_cleanup_orphans=false
opt_preserve_volumes=true
opt_cleanup_service_images=false
opt_cleanup_service_volumes=false
opt_preserve_data=true
opt_cleanup_older_than=""
opt_cleanup_large_volumes=false
opt_max_volume_size="500"
opt_max_volumes_remove="3"

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 COMMAND [TARGET] [OPTIONS]

DESCRIPTION:
    Unified service management tool using podman terminology. Handles individual 
    services, bulk operations, dependency order, and smart cleanup.

INDIVIDUAL SERVICE COMMANDS:
    up SERVICE              Start a single service
    down SERVICE            Stop a single service  
    restart SERVICE         Restart a single service
    rebuild SERVICE         Rebuild and restart a service
    logs SERVICE            Show service logs
    status [SERVICE]        Show service status (or all if no service)
    ps                      List running services
    list                    List all available services

SYSTEM MANAGEMENT COMMANDS:
    list-components         List all podman/docker components (images, containers, volumes, networks, pods)
    prune-components        Remove ALL components (WARNING: destructive operation!)

BULK OPERATION COMMANDS:
    start [TARGETS]         Start services/categories in dependency order
    stop [TARGETS]          Stop services/categories in reverse dependency order  
    start-all               Start all services in dependency order
    stop-all                Stop all services in reverse dependency order

TARGETS (for bulk commands):
    CATEGORIES              database, backend, analytics, proxy, etc.
    SERVICES                postgres, nginx, redis, etc.
    
INDIVIDUAL SERVICE OPTIONS:
    -f, --force             Force recreate containers (for up/rebuild)
    --no-cache              Don't use cache when rebuilding
    --remove-volumes        Remove volumes when stopping
    -t, --timeout SECONDS   Graceful shutdown timeout (default: 30)

BULK OPERATION OPTIONS:
    -c, --categories LIST   Target specific categories (comma-separated)
    -x, --exclude LIST      Exclude specific categories (comma-separated)
    --services LIST         Target specific services (comma-separated)
    --except-services LIST  Exclude specific services
    -p, --parallel          Start services in parallel within categories
    -w, --no-wait           Don't wait for health checks
    --health-timeout SEC    Health check timeout (default: 120)
    --restart POLICY        Restart policy (no, always, unless-stopped, on-failure)
    --rebuild-services      Rebuild services before starting them
    --force-services        Force recreate service containers

CLEANUP OPTIONS:
    -i, --remove-images         Remove all container images
    -v, --remove-volumes        Remove all volumes (WARNING: destroys data)
    -n, --remove-networks       Remove created networks
    --preserve-volumes          Preserve all volumes (safe default)
    --preserve-data             Never touch any data volumes (safest)
    --cleanup-service-images    Remove images only for stopped services
    --cleanup-service-volumes   Remove volumes only for stopped services
    --cleanup-orphans          Remove containers not managed by compose files
    --cleanup-older-than AGE    Remove resources older than AGE (e.g., 7d, 2w)
    --cleanup-large-volumes     Remove large unused volumes to free space
    --max-volume-size SIZE      Max volume size in MB for removal (default: 500)
    --max-volumes COUNT         Max number of volumes to remove (default: 3)

LOG OPTIONS:
    --follow                Follow logs in real-time (Ctrl+C to exit)
    --tail LINES            Number of log lines to show (default: 100)

GENERAL OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -d, --dry-run           Show what would be executed
    -v, --verbose           Show detailed progress information
    -h, --help              Show this help message

EXAMPLES:
    # Individual service management
    $0 up postgres                          # Start postgres
    $0 down nginx --remove-volumes          # Stop nginx and remove volumes
    $0 rebuild php --no-cache               # Rebuild php without cache
    $0 logs redis --follow                  # Follow redis logs
    
    # System management
    $0 list-components                      # List all components (images, containers, volumes, etc.)
    $0 prune-components --force             # Remove ALL components without confirmation
    $0 prune-components --dry-run           # Preview what would be removed
    $0 prune-components --sudo              # Prune with sudo (for rootful podman/docker)
    
    # Bulk operations
    $0 start database backend               # Start database then backend
    $0 stop --categories analytics,proxy    # Stop analytics and proxy categories
    $0 start-all --parallel                 # Start all services with parallel startup
    $0 stop-all --preserve-data             # Stop all but preserve all data
    
    # Service filtering
    $0 start --services postgres,redis      # Start only postgres and redis
    $0 stop database --except-services mysql # Stop database category except mysql
    
    # Smart cleanup
    $0 stop-all --cleanup-orphans --cleanup-older-than 7d
    $0 down nginx --cleanup-service-images  # Stop nginx and clean its images

CATEGORIES (dependency order):
    database      - MariaDB, MySQL, PostgreSQL, MongoDB, Redis
    backend       - PHP, Python, Node.js applications  
    analytics     - Grafana, Prometheus, Matomo
    proxy         - Nginx Proxy Manager

SMART FEATURES:
    - Dependency order automatically maintained
    - Safe defaults (preserves data unless explicitly requested)
    - Enhanced progress feedback vs native podman compose
    - Auto-detects service compose files
    - Helpful error messages with suggestions

NOTES:
    - Services start in dependency order, stop in reverse order
    - Use --remove-volumes with caution (destroys persistent data)
    - Dry run mode shows exact commands that would be executed
    - Health checks prevent cascading failures in bulk operations
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ $# -eq 0 ]]; then
        print_status "error" "No command specified"
        show_usage
        exit 1
    fi
    
    # First argument is always the command
    COMMAND="$1"
    shift
    
    # For individual service commands, second argument is service name
    local individual_commands=("up" "down" "restart" "rebuild" "logs")
    if [[ " ${individual_commands[*]} " =~ " $COMMAND " ]]; then
        if [[ -n "$1" && "$1" != -* ]]; then
            SERVICE_NAME="$1"
            shift
        fi
    fi
    
    # For bulk commands, collect targets
    local bulk_commands=("start" "stop")
    if [[ " ${bulk_commands[*]} " =~ " $COMMAND " ]]; then
        # Collect non-option arguments as targets
        local -a targets
        while [[ -n "$1" && "$1" != -* ]]; do
            targets+=("$1")
            shift
        done
        BULK_TARGETS="${targets[*]}"
    fi
    
    # Parse remaining options
    while [[ $# -gt 0 ]]; do
        case $1 in
            # Individual service options
            -f|--force)
                opt_force_recreate=true
                opt_force_services=true
                shift
                ;;
            --no-cache)
                opt_no_cache=true
                shift
                ;;
            --remove-volumes)
                opt_remove_volumes=true
                opt_preserve_volumes=false
                opt_preserve_data=false
                shift
                ;;
            --follow)
                opt_follow_logs=true
                shift
                ;;
            --tail)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_tail_lines="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric value"
                fi
                ;;
            -t|--timeout)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_timeout="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric timeout value"
                fi
                ;;
                
            # Bulk operation options
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
            -p|--parallel)
                opt_parallel_start=true
                shift
                ;;
            -w|--no-wait)
                opt_wait_healthy=false
                shift
                ;;
            --health-timeout)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_health_timeout="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric timeout value"
                fi
                ;;
            --restart)
                if [[ -n "$2" && "$2" != -* ]]; then
                    case "$2" in
                        no|always|unless-stopped|on-failure)
                            opt_restart_policy="$2"
                            ;;
                        *)
                            handle_error "Invalid restart policy: $2"
                            ;;
                    esac
                    shift 2
                else
                    handle_error "Option $1 requires a restart policy"
                fi
                ;;
            --rebuild-services)
                opt_rebuild_services=true
                shift
                ;;
            --force-services)
                opt_force_services=true
                shift
                ;;
                
            # Cleanup options
            -i|--remove-images)
                opt_remove_images=true
                shift
                ;;
            -v|--remove-volumes)
                opt_remove_volumes=true
                opt_preserve_volumes=false
                opt_preserve_data=false
                shift
                ;;
            -n|--remove-networks)
                opt_remove_networks=true
                shift
                ;;
            --preserve-volumes)
                opt_preserve_volumes=true
                opt_remove_volumes=false
                shift
                ;;
            --preserve-data)
                opt_preserve_data=true
                opt_remove_volumes=false
                opt_cleanup_service_volumes=false
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
            --cleanup-orphans)
                opt_cleanup_orphans=true
                shift
                ;;
            --cleanup-older-than)
                if [[ -n "$2" && "$2" != -* ]]; then
                    local age_value="${2%d}"
                    age_value="${age_value%w}"
                    if [[ "$2" == *"w" ]]; then
                        age_value=$((age_value * 7))
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
                
            # General options
            -s|--sudo)
                opt_use_sudo=true
                shift
                ;;
            -e|--errors)
                opt_show_errors=true
                shift
                ;;
            -d|--dry-run)
                opt_dry_run=true
                opt_verbose=true
                shift
                ;;
            -v|--verbose)
                opt_verbose=true
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

validate_command() {
    local valid_commands=("up" "down" "restart" "rebuild" "logs" "status" "ps" "list" "start" "stop" "start-all" "stop-all" "list-components" "prune-components")
    
    if [[ ! " ${valid_commands[*]} " =~ " $COMMAND " ]]; then
        print_status "error" "Invalid command: $COMMAND"
        print_status "info" "Valid commands: ${valid_commands[*]}"
        exit 1
    fi
}

validate_service_required() {
    local commands_requiring_service=("up" "down" "restart" "rebuild" "logs")
    
    if [[ " ${commands_requiring_service[*]} " =~ " $COMMAND " && -z "$SERVICE_NAME" ]]; then
        print_status "error" "Command '$COMMAND' requires a service name"
        print_status "info" "Available services: $(list_all_service_names | tr '\n' ' ')"
        exit 1
    fi
}

validate_service_exists_if_provided() {
    if [[ -n "$SERVICE_NAME" ]] && ! validate_service_exists "$SERVICE_NAME"; then
        print_status "error" "Service '$SERVICE_NAME' not found"
        print_status "info" "Available services: $(list_all_service_names | tr '\n' ' ')"
        exit 1
    fi
}

validate_bulk_targets() {
    # Parse targets from command line or options
    local -a all_targets
    
    # Add targets from command line
    if [[ -n "$BULK_TARGETS" ]]; then
        all_targets+=($BULK_TARGETS)
    fi
    
    # Add targets from --categories
    if [[ -n "$opt_categories_only" ]]; then
        local -a categories
        categories=(${(s:,:)opt_categories_only})
        all_targets+=("${categories[@]}")
    fi
    
    # Add targets from --services  
    if [[ -n "$opt_services_only" ]]; then
        local -a services
        services=(${(s:,:)opt_services_only})
        all_targets+=("${services[@]}")
    fi
    
    # Validate each target
    for target in "${all_targets[@]}"; do
        # Check if it's a valid category
        if [[ -n "${SERVICE_CATEGORIES[$target]}" ]]; then
            continue
        fi
        
        # Check if it's a valid service
        if validate_service_exists "$target"; then
            continue
        fi
        
        # If neither, it's invalid
        print_status "error" "Invalid target: '$target'"
        print_status "info" "Valid categories: ${(k)SERVICE_CATEGORIES}"
        print_status "info" "Valid services: $(list_all_service_names | tr '\n' ' ')"
        exit 1
    done
    
    # Validate conflicting options
    if [[ -n "$BULK_TARGETS" && -n "$opt_categories_only" ]]; then
        print_status "error" "Cannot use both command-line targets and --categories option"
        exit 1
    fi
    
    if [[ -n "$BULK_TARGETS" && -n "$opt_services_only" ]]; then
        print_status "error" "Cannot use both command-line targets and --services option"
        exit 1
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
        
        if [[ -t 0 ]]; then
            read "response?Are you absolutely sure? Type 'DELETE ALL DATA' to confirm: "
            if [[ "$response" != "DELETE ALL DATA" ]]; then
                print_status "info" "Volume removal cancelled"
                opt_remove_volumes=false
                opt_preserve_volumes=true
                opt_preserve_data=true
            fi
        else
            print_status "error" "Non-interactive mode: volume removal requires explicit confirmation"
            exit 1
        fi
    fi
}

# =============================================================================
# TARGET RESOLUTION FOR BULK OPERATIONS
# =============================================================================

resolve_bulk_targets() {
    local operation="$1"  # "start" or "stop"
    local -a resolved_categories resolved_services
    
    # Determine target sources
    local -a target_list
    if [[ -n "$BULK_TARGETS" ]]; then
        target_list=(${=BULK_TARGETS})
    elif [[ -n "$opt_categories_only" ]]; then
        target_list=(${(s:,:)opt_categories_only})
    elif [[ -n "$opt_services_only" ]]; then
        target_list=(${(s:,:)opt_services_only})
    else
        # Default to all categories
        target_list=("${SERVICE_STARTUP_ORDER[@]}")
    fi
    
    # Separate categories from services
    for target in "${target_list[@]}"; do
        if [[ -n "${SERVICE_CATEGORIES[$target]}" ]]; then
            resolved_categories+=("$target")
        else
            resolved_services+=("$target")
        fi
    done
    
    # Apply exclusions
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories final_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        for category in "${resolved_categories[@]}"; do
            if [[ ! " ${excluded_categories[*]} " =~ " $category " ]]; then
                final_categories+=("$category")
            fi
        done
        resolved_categories=("${final_categories[@]}")
    fi
    
    if [[ -n "$opt_except_services" ]]; then
        local -a excluded_services final_services
        excluded_services=(${(s:,:)opt_except_services})
        
        for service in "${resolved_services[@]}"; do
            if [[ ! " ${excluded_services[*]} " =~ " $service " ]]; then
                final_services+=("$service")
            fi
        done
        resolved_services=("${final_services[@]}")
    fi
    
    # Order categories correctly
    local -a ordered_categories
    if [[ "$operation" == "start" ]]; then
        # Use startup order
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            if [[ " ${resolved_categories[*]} " =~ " $category " ]]; then
                ordered_categories+=("$category")
            fi
        done
    else
        # Use reverse startup order for stop
        for ((i = ${#SERVICE_STARTUP_ORDER[@]}; i > 0; i--)); do
            local category="${SERVICE_STARTUP_ORDER[i]}"
            if [[ " ${resolved_categories[*]} " =~ " $category " ]]; then
                ordered_categories+=("$category")
            fi
        done
    fi
    
    # Export resolved targets
    RESOLVED_CATEGORIES=("${ordered_categories[@]}")
    RESOLVED_SERVICES=("${resolved_services[@]}")
}

# =============================================================================
# INDIVIDUAL SERVICE COMMANDS (Enhanced from original)
# =============================================================================

cmd_up() {
    if [[ "$opt_dry_run" == "true" ]]; then
        local service_path
        service_path=$(get_service_path "$SERVICE_NAME")
        if [[ $? -eq 0 ]]; then
            local compose_args=("-f" "$service_path" "up" "-d")
            [[ "$opt_force_recreate" == "true" ]] && compose_args+=("--force-recreate")
            echo "DRY RUN: $COMPOSE_CMD ${compose_args[*]}"
        else
            echo "DRY RUN: Service '$SERVICE_NAME' not found"
        fi
        return 0
    fi
    
    start_single_service "$SERVICE_NAME" "$opt_force_recreate"
}

cmd_down() {
    if [[ "$opt_dry_run" == "true" ]]; then
        local service_path=$(get_service_path "$SERVICE_NAME")
        local compose_args=("-f" "$service_path" "down" "--timeout" "$opt_timeout")
        [[ "$opt_remove_volumes" == "true" ]] && compose_args+=("--volumes")
        echo "DRY RUN: $COMPOSE_CMD ${compose_args[*]}"
        return 0
    fi
    
    stop_single_service "$SERVICE_NAME" "$opt_remove_volumes" "$opt_timeout"
    
    # Run cleanup for this specific service
    if [[ $? -eq 0 ]]; then
        cleanup_service_resources "$SERVICE_NAME"
    fi
}

cmd_restart() {
    if [[ "$opt_dry_run" == "true" ]]; then
        local service_path=$(get_service_path "$SERVICE_NAME")
        echo "DRY RUN: $COMPOSE_CMD -f '$service_path' restart"
        return 0
    fi
    
    restart_single_service "$SERVICE_NAME"
}

cmd_rebuild() {
    if [[ "$opt_dry_run" == "true" ]]; then
        local service_path=$(get_service_path "$SERVICE_NAME")
        echo "DRY RUN: $COMPOSE_CMD -f '$service_path' down --timeout 10"
        local build_args=("-f" "$service_path" "build")
        [[ "$opt_no_cache" == "true" ]] && build_args+=("--no-cache")
        echo "DRY RUN: $COMPOSE_CMD ${build_args[*]}"
        echo "DRY RUN: $COMPOSE_CMD -f '$service_path' up -d --force-recreate"
        return 0
    fi
    
    rebuild_single_service "$SERVICE_NAME" "$opt_no_cache"
}

cmd_logs() {
    if [[ "$opt_dry_run" == "true" ]]; then
        local service_path=$(get_service_path "$SERVICE_NAME")
        local log_args=("-f" "$service_path" "logs" "--tail" "$opt_tail_lines")
        [[ "$opt_follow_logs" == "true" ]] && log_args+=("-f")
        echo "DRY RUN: $COMPOSE_CMD ${log_args[*]}"
        return 0
    fi
    
    show_service_logs "$SERVICE_NAME" "$opt_follow_logs" "$opt_tail_lines"
}

cmd_status() {
    if [[ -n "$SERVICE_NAME" ]]; then
        # Show status for specific service
        local service_status=$(get_service_status "$SERVICE_NAME")
        local category=$(find_service_category "$SERVICE_NAME")
        
        print_status "info" "Service: $SERVICE_NAME"
        echo "  Category: $category"
        echo "  Status: $service_status"
        echo "  Path: $(get_service_path "$SERVICE_NAME")"
    else
        # Show status for all services
        print_status "info" "Service Status Overview:"
        echo ""
        
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            echo "📂 $category:"
            local service_files="${SERVICE_CATEGORIES[$category]}"
            local -a files
            files=(${=service_files})
            
            for service_file in "${files[@]}"; do
                local service_name="${service_file%.yml}"
                local service_status=$(get_service_status "$service_name" 2>/dev/null || echo "")
                
                case "$service_status" in
                    "running")
                        echo "  ✅ $service_name"
                        ;;
                    "STOPPED"|"unknown")
                        echo "  ❌ $service_name"
                        ;;
                    *)
                        echo "  ⚠️  $service_name ($service_status)"
                        ;;
                esac
            done
            echo ""
        done
    fi
}

cmd_ps() {
    print_status "info" "Running Services:"
    echo ""

    if output=$(eval "$CONTAINER_CMD ps --filter 'network=$NETWORK_NAME' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'" 2>/dev/null); then
        echo "$output"
    else
        print_status "warning" "Could not list running containers"
    fi
}

cmd_list() {
    print_status "info" "Available Services:"
    echo ""
    
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        echo "📂 $category:"
        local service_files="${SERVICE_CATEGORIES[$category]}"
        local -a files
        files=(${=service_files})
        
        for service_file in "${files[@]}"; do
            local service_name="${service_file%.yml}"
            echo "  • $service_name"
        done
        echo ""
    done
    
    echo "Total services: $(list_all_service_names | wc -l)"
}

cmd_list_components() {
    print_status "info" "Listing all Podman/Docker components:"
    echo ""
    
    print_status "step" "Images:"
    eval "$CONTAINER_CMD images"
    echo ""
    
    print_status "step" "Containers:"
    eval "$CONTAINER_CMD ps -a"
    echo ""
    
    print_status "step" "Volumes:"
    eval "$CONTAINER_CMD volume ls"
    echo ""
    
    print_status "step" "Networks:"
    eval "$CONTAINER_CMD network ls"
    echo ""
    
    # Only list pods if using podman
    if [[ "$USE_PODMAN" == "true" ]]; then
        print_status "step" "Pods:"
        eval "$CONTAINER_CMD pod ps -a"
        echo ""
    fi
}

cmd_prune_components() {
    print_status "warning" "⚠️  WARNING: This will remove ALL containers, images, volumes, networks, and pods!"
    echo ""
    
    if [[ "$opt_dry_run" == "true" ]]; then
        print_status "info" "DRY RUN: Would execute the following:"
        echo "  - Remove all images"
        echo "  - Remove all containers"
        echo "  - Remove all volumes"
        echo "  - Remove network: $NETWORK_NAME"
        if [[ "$USE_PODMAN" == "true" ]]; then
            echo "  - Remove all pods"
        fi
        return 0
    fi
    
    # Confirm with user unless force flag is set
    if [[ "$opt_force_recreate" != "true" && "$opt_force_services" != "true" ]]; then
        echo -n "Are you sure you want to prune all components? (yes/no): "
        read confirmation
        if [[ "$confirmation" != "yes" ]]; then
            print_status "info" "Prune operation cancelled"
            return 0
        fi
    fi
    
    print_status "step" "Pruning all components..."
    echo ""
    
    # Remove all containers first
    print_status "step" "Removing all containers..."
    if eval "$CONTAINER_CMD container rm --all -f $ERROR_REDIRECT"; then
        print_status "success" "All containers removed"
    else
        print_status "warning" "Some containers may not have been removed"
    fi
    
    # Remove all images
    print_status "step" "Removing all images..."
    if eval "$CONTAINER_CMD image rm --all -f $ERROR_REDIRECT"; then
        print_status "success" "All images removed"
    else
        print_status "warning" "Some images may not have been removed"
    fi
    
    # Remove all volumes
    print_status "step" "Removing all volumes..."
    if eval "$CONTAINER_CMD volume rm --all -f $ERROR_REDIRECT"; then
        print_status "success" "All volumes removed"
    else
        print_status "warning" "Some volumes may not have been removed"
    fi
    
    # Remove project network
    print_status "step" "Removing network: $NETWORK_NAME..."
    if eval "$CONTAINER_CMD network rm $NETWORK_NAME $ERROR_REDIRECT"; then
        print_status "success" "Network $NETWORK_NAME removed"
    else
        print_status "warning" "Network $NETWORK_NAME may not exist or couldn't be removed"
    fi
    
    # Remove all pods (podman only)
    if [[ "$USE_PODMAN" == "true" ]]; then
        print_status "step" "Removing all pods..."
        if eval "$CONTAINER_CMD pod rm --all -f $ERROR_REDIRECT"; then
            print_status "success" "All pods removed"
        else
            print_status "warning" "Some pods may not have been removed"
        fi
    fi
    
    echo ""
    print_status "success" "Prune operation completed!"
}

# =============================================================================
# BULK OPERATION COMMANDS
# =============================================================================

cmd_start() {
    resolve_bulk_targets "start"
    
    if [[ ${#RESOLVED_CATEGORIES[@]} -eq 0 && ${#RESOLVED_SERVICES[@]} -eq 0 ]]; then
        print_status "warning" "No targets resolved for startup"
        return 0
    fi
    
    print_status "step" "Starting services in dependency order..."
    
    # Setup environment first
    ensure_network_exists
    
    # Start individual services first
    for service in "${RESOLVED_SERVICES[@]}"; do
        start_individual_service "$service"
        sleep 1
    done
    
    # Start categories in dependency order
    for category in "${RESOLVED_CATEGORIES[@]}"; do
        if [[ "$opt_parallel_start" == "true" ]]; then
            start_category_parallel "$category"
        else
            start_category_sequential "$category"
        fi
        
        # Wait for health if enabled
        wait_for_category_health "$category"
        sleep 2
    done
    
    print_status "success" "Bulk startup completed!"
}

cmd_stop() {
    resolve_bulk_targets "stop"
    
    if [[ ${#RESOLVED_CATEGORIES[@]} -eq 0 && ${#RESOLVED_SERVICES[@]} -eq 0 ]]; then
        print_status "warning" "No targets resolved for shutdown"
        return 0
    fi
    
    print_status "step" "Stopping services in reverse dependency order..."
    
    # Stop categories in reverse dependency order
    for category in "${RESOLVED_CATEGORIES[@]}"; do
        stop_category "$category"
        sleep 1
    done
    
    # Stop individual services
    for service in "${RESOLVED_SERVICES[@]}"; do
        stop_individual_service "$service"
    done
    
    # Run cleanup operations
    run_cleanup_operations
    
    print_status "success" "Bulk shutdown completed!"
}

cmd_start_all() {
    # Set all categories as targets
    BULK_TARGETS=""
    opt_categories_only=""
    opt_services_only=""
    
    # Override with all categories
    RESOLVED_CATEGORIES=("${SERVICE_STARTUP_ORDER[@]}")
    RESOLVED_SERVICES=()
    
    # Apply exclusions if any
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories final_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        for category in "${RESOLVED_CATEGORIES[@]}"; do
            if [[ ! " ${excluded_categories[*]} " =~ " $category " ]]; then
                final_categories+=("$category")
            fi
        done
        RESOLVED_CATEGORIES=("${final_categories[@]}")
    fi
    
    cmd_start
}

cmd_stop_all() {
    # Set all categories as targets in reverse order
    BULK_TARGETS=""
    opt_categories_only=""
    opt_services_only=""
    
    # Override with all categories in reverse order
    local -a reversed_categories
    for ((i = ${#SERVICE_STARTUP_ORDER[@]}; i > 0; i--)); do
        reversed_categories+=("${SERVICE_STARTUP_ORDER[i]}")
    done
    RESOLVED_CATEGORIES=("${reversed_categories[@]}")
    RESOLVED_SERVICES=()
    
    # Apply exclusions if any
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories final_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        for category in "${RESOLVED_CATEGORIES[@]}"; do
            if [[ ! " ${excluded_categories[*]} " =~ " $category " ]]; then
                final_categories+=("$category")
            fi
        done
        RESOLVED_CATEGORIES=("${final_categories[@]}")
    fi
    
    cmd_stop
}

# =============================================================================
# ENHANCED BULK OPERATION HELPERS
# =============================================================================

start_individual_service() {
    local service_name="$1"
    
    print_status "step" "Starting individual service: $service_name"
    
    if [[ "$opt_rebuild_services" == "true" ]]; then
        rebuild_single_service "$service_name" "$opt_no_cache"
    else
        local force_recreate="$opt_force_recreate"
        [[ "$opt_force_services" == "true" ]] && force_recreate="true"
        start_single_service "$service_name" "$force_recreate"
    fi
}

stop_individual_service() {
    local service_name="$1"
    
    print_status "step" "Stopping individual service: $service_name"
    
    local remove_volumes="$opt_remove_volumes"
    if [[ "$opt_preserve_data" == "true" || "$opt_preserve_volumes" == "true" ]]; then
        remove_volumes="false"
    fi
    
    stop_single_service "$service_name" "$remove_volumes" "$opt_timeout"
    
    # Cleanup for this specific service
    if [[ $? -eq 0 ]]; then
        cleanup_service_resources "$service_name"
    fi
}

start_category_sequential() {
    local category="$1"
    local service_files
    
    service_files=$(filter_services_for_category "$category")
    
    if [[ -z "$service_files" ]]; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "info" "No services to start in $category (filtered out)"
        fi
        return 0
    fi
    
    print_status "step" "Starting $category services sequentially..."
    
    local -a files
    files=(${=service_files})
    
    local failed=0
    for service_file in "${files[@]}"; do
        local service_name="${service_file%.yml}"
        if ! start_individual_service "$service_name"; then
            ((failed++))
        fi
        
        if [[ "$opt_dry_run" == "false" ]]; then
            sleep 1
        fi
    done
    
    if [[ $failed -eq 0 ]]; then
        print_status "success" "$category services started successfully"
    else
        print_status "warning" "$failed service(s) in $category failed to start"
    fi
}

start_category_parallel() {
    local category="$1"
    local service_files
    
    service_files=$(filter_services_for_category "$category")
    
    if [[ -z "$service_files" ]]; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "info" "No services to start in $category (filtered out)"
        fi
        return 0
    fi
    
    print_status "step" "Starting $category services in parallel..."
    
    local -a files pids
    files=(${=service_files})
    
    # Start all services in background
    for service_file in "${files[@]}"; do
        local service_name="${service_file%.yml}"
        if [[ "$opt_dry_run" == "true" ]]; then
            start_individual_service "$service_name"
        else
            start_individual_service "$service_name" &
            pids+=($!)
        fi
    done
    
    if [[ "$opt_dry_run" == "false" ]]; then
        # Wait for all parallel starts to complete
        local failed=0
        for pid in "${pids[@]}"; do
            if ! wait "$pid"; then
                ((failed++))
            fi
        done
        
        if [[ $failed -eq 0 ]]; then
            print_status "success" "$category services started successfully"
        else
            print_status "warning" "$failed service(s) in $category failed to start"
        fi
    fi
}

stop_category() {
    local category="$1"
    local service_files
    
    service_files=$(filter_services_for_category "$category")
    
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
        local service_name="${service_file%.yml}"
        if ! stop_individual_service "$service_name"; then
            ((failed++))
        fi
    done
    
    if [[ $failed -eq 0 ]]; then
        print_status "success" "$category services stopped successfully"
    else
        print_status "warning" "$failed service(s) in $category had issues stopping"
    fi
}

filter_services_for_category() {
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

wait_for_category_health() {
    local category="$1"
    
    if [[ "$opt_wait_healthy" == "false" || "$opt_dry_run" == "true" ]]; then
        return 0
    fi
    
    print_status "step" "Waiting for $category services to be healthy..."
    
    # Special handling for different categories
    case "$category" in
        "database")
            # Wait for MongoDB specifically if it's in this category
            local service_files
            service_files=$(get_service_files "$category")
            if [[ "$service_files" == *"mongodb.yml"* ]]; then
                wait_for_mongodb "$opt_health_timeout"
            fi
            
            # Brief wait for other databases
            sleep 5
            ;;
        "proxy")
            # Wait longer for nginx proxy manager
            sleep 10
            ;;
        *)
            # Standard wait for other services
            sleep 3
            ;;
    esac
    
    print_status "success" "$category services are ready"
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

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

run_cleanup_operations() {
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Running cleanup operations..."
    fi
    
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
    
    # Global cleanup operations - NOW PROPERLY CONDITIONAL
    if [[ "$opt_remove_images" == "true" ]]; then
        cleanup_images_global
    fi
    
    if [[ "$opt_remove_volumes" == "true" ]]; then
        cleanup_volumes_global
    fi
    
    if [[ "$opt_remove_networks" == "true" ]]; then
        cleanup_networks_global
    fi
    
    # System cleanup - should probably have its own flag or be tied to other cleanup options
    if [[ "$opt_cleanup_orphans" == "true" || "$opt_remove_images" == "true" || "$opt_remove_volumes" == "true" ]]; then
        cleanup_system_global
    fi
}

cleanup_images_global() {
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

cleanup_volumes_global() {
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

cleanup_networks_global() {
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
    
    if [[ "$opt_remove_networks" == "false" ]]; then
        return 0
    else
        eval "$CONTAINER_CMD network prune -f $ERROR_REDIRECT" || true
    fi
    # Cleanup unused networks
}

cleanup_system_global() {
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $CONTAINER_CMD system prune -f"
        return 0
    fi
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Cleaning up system resources..."
    fi
    
    # Prune unused containers, networks, and build cache
    eval "$CONTAINER_CMD system prune -f $ERROR_REDIRECT" || true
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "success" "System cleanup completed"
    fi
}

cleanup_old_resources() {
    local max_age_days="$1"
    local dry_run="${2:-false}"
    
    if [[ -z "$max_age_days" || ! "$max_age_days" =~ ^[0-9]+$ ]]; then
        print_status "error" "Invalid age specified for cleanup: $max_age_days"
        return 1
    fi
    
    print_status "step" "Cleaning up resources older than $max_age_days days..."
    
    if [[ "$dry_run" == "true" ]]; then
        echo "DRY RUN: Would remove containers, images, and volumes older than $max_age_days days"
        return 0
    fi
    
    local removed_containers=0 removed_images=0 removed_volumes=0
    local cutoff_date=$(date -d "$max_age_days days ago" '+%Y-%m-%d')
    
    # Clean up old containers (stopped ones only)
    print_status "info" "Removing containers older than $cutoff_date..."
    local old_containers
    old_containers=$(eval "$CONTAINER_CMD ps -a --filter 'status=exited' --format '{{.ID}} {{.CreatedAt}}' 2>/dev/null" || echo "")
    
    if [[ -n "$old_containers" ]]; then
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                local container_id=$(echo "$line" | awk '{print $1}')
                local created_date=$(echo "$line" | awk '{print $2}')
                
                # Simple date comparison (assumes YYYY-MM-DD format)
                if [[ "$created_date" < "$cutoff_date" ]]; then
                    if eval "$CONTAINER_CMD rm -f $container_id $ERROR_REDIRECT"; then
                        ((removed_containers++))
                    fi
                fi
            fi
        done <<< "$old_containers"
    fi
    
    # Clean up old images (unused ones only)
    print_status "info" "Removing unused images older than $cutoff_date..."
    if eval "$CONTAINER_CMD image prune -a --filter \"until=${max_age_days}h\" -f $ERROR_REDIRECT"; then
        # Count is approximate since we can't get exact numbers from prune
        removed_images=1
    fi
    
    # Clean up old volumes (unused ones only)
    print_status "info" "Removing unused volumes older than $cutoff_date..."
    if eval "$CONTAINER_CMD volume prune --filter \"until=${max_age_days}h\" -f $ERROR_REDIRECT"; then
        # Count is approximate since we can't get exact numbers from prune
        removed_volumes=1
    fi
    
    if [[ $removed_containers -gt 0 || $removed_images -gt 0 || $removed_volumes -gt 0 ]]; then
        print_status "success" "Cleanup completed: ~$removed_containers containers, $removed_images image groups, $removed_volumes volume groups"
    else
        print_status "info" "No old resources found to clean up"
    fi
}

cleanup_large_volumes() {
    local max_size_mb="$1"
    local max_count="${2:-3}"
    local dry_run="${3:-false}"
    
    if [[ -z "$max_size_mb" || ! "$max_size_mb" =~ ^[0-9]+$ ]]; then
        print_status "error" "Invalid size specified for volume cleanup: $max_size_mb"
        return 1
    fi
    
    print_status "step" "Finding volumes larger than ${max_size_mb}MB..."
    
    if [[ "$dry_run" == "true" ]]; then
        echo "DRY RUN: Would remove up to $max_count volumes larger than ${max_size_mb}MB"
        return 0
    fi
    
    # Get list of volumes with sizes
    local volumes_info
    volumes_info=$(eval "$CONTAINER_CMD volume ls --format '{{.Name}}' 2>/dev/null" || echo "")
    
    if [[ -z "$volumes_info" ]]; then
        print_status "info" "No volumes found"
        return 0
    fi
    
    local -a large_volumes
    local removed_count=0
    
    # Check each volume size
    while IFS= read -r volume_name; do
        if [[ -n "$volume_name" ]]; then
            # Get volume mount point to check size
            local volume_path
            volume_path=$(eval "$CONTAINER_CMD volume inspect $volume_name --format '{{.Mountpoint}}' 2>/dev/null" || echo "")
            
            if [[ -n "$volume_path" ]]; then
                # Check if volume is in use
                local in_use
                in_use=$(eval "$CONTAINER_CMD ps -a --filter \"volume=$volume_name\" --format '{{.Names}}' 2>/dev/null" || echo "")
                
                if [[ -z "$in_use" ]]; then
                    # Get size in MB
                    local size_mb
                    if [[ "$USE_PODMAN" == "true" ]]; then
                        # For podman, use du on the volume path
                        size_mb=$(sudo du -sm "$volume_path" 2>/dev/null | awk '{print $1}' || echo "0")
                    else
                        # For docker, also use du but might need different approach
                        size_mb=$(sudo du -sm "$volume_path" 2>/dev/null | awk '{print $1}' || echo "0")
                    fi
                    
                    if [[ "$size_mb" -gt "$max_size_mb" ]]; then
                        large_volumes+=("$volume_name:$size_mb")
                    fi
                fi
            fi
        fi
    done <<< "$volumes_info"
    
    if [[ ${#large_volumes[@]} -eq 0 ]]; then
        print_status "info" "No large unused volumes found"
        return 0
    fi
    
    # Sort by size (descending) and remove largest ones first
    local -a sorted_volumes
    sorted_volumes=($(printf '%s\n' "${large_volumes[@]}" | sort -t: -k2 -rn))
    
    print_status "info" "Found ${#sorted_volumes[@]} large volume(s), removing up to $max_count..."
    
    for volume_info in "${sorted_volumes[@]}"; do
        if [[ $removed_count -ge $max_count ]]; then
            break
        fi
        
        local volume_name="${volume_info%:*}"
        local volume_size="${volume_info#*:}"
        
        print_status "step" "Removing volume: $volume_name (${volume_size}MB)"
        
        if eval "$CONTAINER_CMD volume rm -f $volume_name $ERROR_REDIRECT"; then
            ((removed_count++))
            print_status "success" "Removed $volume_name (${volume_size}MB)"
        else
            print_status "warning" "Failed to remove $volume_name"
        fi
    done
    
    if [[ $removed_count -gt 0 ]]; then
        print_status "success" "Removed $removed_count large volume(s)"
    fi
}

cleanup_service_orphans() {
    local target_services="$1"
    local dry_run="${2:-false}"
    
    print_status "step" "Cleaning up orphaned containers..."
    
    if [[ "$dry_run" == "true" ]]; then
        echo "DRY RUN: Would remove containers not managed by compose files"
        return 0
    fi
    
    # Get all running containers
    local all_containers
    all_containers=$(eval "$CONTAINER_CMD ps -a --format '{{.Names}}' 2>/dev/null" || echo "")
    
    if [[ -z "$all_containers" ]]; then
        print_status "info" "No containers found"
        return 0
    fi
    
    # Get list of managed services from our categories
    local -a managed_services
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        local service_files
        service_files=$(get_service_files "$category")
        if [[ -n "$service_files" ]]; then
            local -a files
            files=(${=service_files})
            for service_file in "${files[@]}"; do
                local service_name="${service_file%.yml}"
                managed_services+=("$service_name")
            done
        fi
    done
    
    # If target_services specified, only check those
    if [[ -n "$target_services" ]]; then
        local -a target_list
        target_list=(${(s:,:)target_services})
        managed_services=("${target_list[@]}")
    fi
    
    local removed_count=0
    
    # Check each container
    while IFS= read -r container_name; do
        if [[ -n "$container_name" ]]; then
            # Skip if it's in our managed services list
            local is_managed=false
            for managed_service in "${managed_services[@]}"; do
                if [[ "$container_name" == "$managed_service" ]]; then
                    is_managed=true
                    break
                fi
            done
            
            # Skip containers in our project network (likely managed)
            if eval "$CONTAINER_CMD inspect $container_name --format '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' 2>/dev/null | grep -q $NETWORK_NAME"; then
                is_managed=true
            fi
            
            if [[ "$is_managed" == "false" ]]; then
                # Check if container has compose labels (indicating it's managed by compose)
                local has_compose_labels
                has_compose_labels=$(eval "$CONTAINER_CMD inspect $container_name --format '{{index .Config.Labels \"com.docker.compose.project\"}}' 2>/dev/null" || echo "")
                
                if [[ -z "$has_compose_labels" ]]; then
                    print_status "step" "Removing orphaned container: $container_name"
                    
                    if eval "$CONTAINER_CMD rm -f $container_name $ERROR_REDIRECT"; then
                        ((removed_count++))
                        print_status "info" "Removed orphaned container: $container_name"
                    else
                        print_status "warning" "Failed to remove container: $container_name"
                    fi
                fi
            fi
        fi
    done <<< "$all_containers"
    
    if [[ $removed_count -gt 0 ]]; then
        print_status "success" "Removed $removed_count orphaned container(s)"
    else
        print_status "info" "No orphaned containers found"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse arguments
    parse_arguments "$@"
    
    # Set up command context
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Validate command
    validate_command
    
    # Validate service name is provided when required
    validate_service_required
    
    # Validate service exists (if service name provided)
    validate_service_exists_if_provided
    
    # Validate bulk targets for bulk operations
    local bulk_commands=("start" "stop" "start-all" "stop-all")
    if [[ " ${bulk_commands[*]} " =~ " $COMMAND " ]]; then
        validate_bulk_targets
        validate_destructive_operations
    fi
    
    # Execute command
    case "$COMMAND" in
        # Individual service commands
        "up")
            cmd_up
            ;;
        "down")
            cmd_down
            ;;
        "restart")
            cmd_restart
            ;;
        "rebuild")
            cmd_rebuild
            ;;
        "logs")
            cmd_logs
            ;;
        "status")
            cmd_status
            ;;
        "ps")
            cmd_ps
            ;;
        "list")
            cmd_list
            ;;
        "list-components")
            cmd_list_components
            ;;
        "prune-components")
            cmd_prune_components
            ;;
        # Bulk operation commands
        "start")
            cmd_start
            ;;
        "stop")
            cmd_stop
            ;;
        "start-all")
            cmd_start_all
            ;;
        "stop-all")
            cmd_stop_all
            ;;
        *)
            handle_error "Unknown command: $COMMAND"
            ;;
    esac
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi