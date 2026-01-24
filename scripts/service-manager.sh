#!/bin/zsh

# =============================================================================
# UNIFIED MICROSERVICES MANAGER - THE ONE TOOL TO RULE THEM ALL
# =============================================================================
# Complete service management with podman terminology, dependency handling,
# bulk operations, and smart cleanup. Replaces start-services.sh and stop-services.sh

# Source the central configuration
. "$(dirname "$0")/config.sh"

SCRIPT_NAME="${0##*/}"

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
VERBOSE=0
QUIET=0

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

show_brief_usage() {
    echo "Usage: $SCRIPT_NAME COMMAND [TARGET] [OPTIONS]"
    echo ""
    echo "Commands: up, down, restart, rebuild, logs, status, ps, list"
    echo "          start, stop, start-all, stop-all"
    echo "          list-components, prune-components"
    echo ""
    echo "Run '$SCRIPT_NAME --help' for detailed usage"
}

show_usage() {
    cat << EOF
Usage: $SCRIPT_NAME COMMAND [TARGET] [OPTIONS]

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
    refresh                 Re-scan compose dirs for new services/categories
    check                   Verify prerequisites (container runtime, network, env)

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
    --refresh               Re-scan compose directories for new services/categories
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

SHELL ALIASES:
    Run from any directory by setting up shell aliases:

    ./scripts/setup-aliases.sh              # Interactive setup

    Or manually add to ~/.bashrc or ~/.zshrc:
        alias devarch='$PROJECT_ROOT/scripts/service-manager.sh'
        alias dvrc='$PROJECT_ROOT/scripts/service-manager.sh'
        alias dv='$PROJECT_ROOT/scripts/service-manager.sh'
        alias da='$PROJECT_ROOT/scripts/service-manager.sh'

    Then: source ~/.zshrc (or open new terminal)
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
        show_usage
        exit 0
    fi

    if [[ $# -eq 0 ]]; then
        show_brief_usage
        exit 0
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
                export VERBOSE=1
                shift
                ;;
            -v|--verbose)
                opt_verbose=true
                export VERBOSE=1
                shift
                ;;
            --refresh)
                refresh_service_discovery
                print_status "success" "Service discovery refreshed"
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
    local valid_commands=("up" "down" "restart" "rebuild" "logs" "status" "ps" "list" "refresh" "check" "start" "stop" "start-all" "stop-all" "list-components" "prune-components")
    [[ ! " ${valid_commands[*]} " =~ " $COMMAND " ]] && { print_status "error" "Invalid command: $COMMAND"; exit 1; }
}

validate_service_required() {
    local cmds=("up" "down" "restart" "rebuild" "logs")
    [[ " ${cmds[*]} " =~ " $COMMAND " && -z "$SERVICE_NAME" ]] && { print_status "error" "$COMMAND requires service name"; exit 1; }
}

validate_service_exists_if_provided() {
    [[ -n "$SERVICE_NAME" ]] && ! validate_service_exists "$SERVICE_NAME" && { print_status "error" "$SERVICE_NAME not found"; exit 1; }
}

validate_bulk_targets() {
    local -a all_targets
    [[ -n "$BULK_TARGETS" ]] && all_targets+=(${=BULK_TARGETS})
    [[ -n "$opt_categories_only" ]] && all_targets+=(${(s:,:)opt_categories_only})
    [[ -n "$opt_services_only" ]] && all_targets+=(${(s:,:)opt_services_only})

    for target in "${all_targets[@]}"; do
        [[ -n "${SERVICE_CATEGORIES[$target]}" ]] && continue
        validate_service_exists "$target" && continue
        print_status "error" "Invalid target: $target"
        exit 1
    done

    [[ -n "$BULK_TARGETS" && -n "$opt_categories_only" ]] && { print_status "error" "Cannot mix targets and --categories"; exit 1; }
    [[ -n "$BULK_TARGETS" && -n "$opt_services_only" ]] && { print_status "error" "Cannot mix targets and --services"; exit 1; }
}

validate_destructive_operations() {
    [[ "$opt_remove_volumes" != "true" || "$opt_dry_run" == "true" ]] && return 0

    if [[ -t 0 ]]; then
        printf "\033[33mWARN\033[0m --remove-volumes destroys all data. Type 'DELETE ALL DATA' to confirm: "
        read response
        [[ "$response" != "DELETE ALL DATA" ]] && { opt_remove_volumes=false; opt_preserve_volumes=true; opt_preserve_data=true; }
    else
        print_status "error" "Volume removal requires interactive confirmation"
        exit 1
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
    local rc=$?
    cleanup_service_resources "$SERVICE_NAME"
    return $rc
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
        local service_status=$(get_service_status "$SERVICE_NAME" | tr -d '\n')
        local category=$(find_service_category "$SERVICE_NAME")
        printf "%-20s %-12s %s\n" "$SERVICE_NAME" "$service_status" "$category"
    else
        printf "\033[1m%-20s %-12s %s\033[0m\n" "SERVICE" "STATUS" "CATEGORY"
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            local service_files="${SERVICE_CATEGORIES[$category]}"
            [[ -z "$service_files" ]] && continue
            local -a files
            files=(${=service_files})
            for service_file in "${files[@]}"; do
                local sname="${service_file%.yml}"
                local sstatus=$(get_service_status "$sname" 2>/dev/null | tr -d '\n')
                [[ -z "$sstatus" ]] && sstatus="stopped"
                local color=""
                case "$sstatus" in
                    running) color="\033[32m" ;;
                    exited|stopped|STOPPED) color="\033[31m"; sstatus="stopped" ;;
                    *) color="\033[33m" ;;
                esac
                printf "%b%-20s %-12s %s\033[0m\n" "$color" "$sname" "$sstatus" "$category"
            done
        done
    fi
}

cmd_check() {
    local has_runtime=false
    local exit_code=0

    # Container runtimes
    command -v podman &>/dev/null && { printf "\033[32m+\033[0m podman %s\n" "$(podman --version 2>/dev/null | cut -d' ' -f3)"; has_runtime=true; }
    command -v docker &>/dev/null && { printf "\033[32m+\033[0m docker %s\n" "$(docker --version 2>/dev/null | cut -d' ' -f3 | tr -d ',')"; has_runtime=true; }
    [[ "$has_runtime" == "false" ]] && { print_status "error" "No container runtime"; exit_code=1; }

    # Network
    eval "$CONTAINER_CMD network exists $NETWORK_NAME" 2>/dev/null && \
        printf "\033[32m+\033[0m network %s\n" "$NETWORK_NAME" || \
        printf "\033[33m-\033[0m network %s (will create)\n" "$NETWORK_NAME"

    # Project structure
    [[ -d "$COMPOSE_DIR" ]] && printf "\033[32m+\033[0m compose dir\n" || { printf "\033[31m!\033[0m compose dir missing\n"; exit_code=1; }
    [[ -f "$PROJECT_ROOT/.env" ]] && printf "\033[32m+\033[0m .env\n" || printf "\033[33m-\033[0m .env (optional)\n"

    return $exit_code
}

cmd_ps() {
    eval "$CONTAINER_CMD ps --filter 'network=$NETWORK_NAME' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'" 2>/dev/null || \
        print_status "warning" "Could not list containers"
}

cmd_list() {
    printf "\033[1m%-20s %s\033[0m\n" "SERVICE" "CATEGORY"
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        local service_files="${SERVICE_CATEGORIES[$category]}"
        [[ -z "$service_files" ]] && continue
        local -a files
        files=(${=service_files})
        for service_file in "${files[@]}"; do
            printf "%-20s %s\n" "${service_file%.yml}" "$category"
        done
    done
}

cmd_list_components() {
    printf "\033[1m=== IMAGES ===\033[0m\n"
    eval "$CONTAINER_CMD images"
    printf "\n\033[1m=== CONTAINERS ===\033[0m\n"
    eval "$CONTAINER_CMD ps -a"
    printf "\n\033[1m=== VOLUMES ===\033[0m\n"
    eval "$CONTAINER_CMD volume ls"
    printf "\n\033[1m=== NETWORKS ===\033[0m\n"
    eval "$CONTAINER_CMD network ls"
    [[ "$USE_PODMAN" == "true" ]] && { printf "\n\033[1m=== PODS ===\033[0m\n"; eval "$CONTAINER_CMD pod ps -a"; }
}

cmd_prune_components() {
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: would remove all containers, images, volumes, networks"
        return 0
    fi

    # Confirm unless force
    if [[ "$opt_force_recreate" != "true" && "$opt_force_services" != "true" ]]; then
        printf "\033[33mWARN\033[0m This removes ALL containers, images, volumes, networks. Continue? (yes/no): "
        read confirmation
        [[ "$confirmation" != "yes" ]] && { echo "Cancelled"; return 0; }
    fi

    eval "$CONTAINER_CMD container rm --all -f $ERROR_REDIRECT" && print_status "success" "containers"
    eval "$CONTAINER_CMD image rm --all -f $ERROR_REDIRECT" && print_status "success" "images"
    eval "$CONTAINER_CMD volume rm --all -f $ERROR_REDIRECT" && print_status "success" "volumes"
    eval "$CONTAINER_CMD network rm $NETWORK_NAME $ERROR_REDIRECT" && print_status "success" "network"
    [[ "$USE_PODMAN" == "true" ]] && eval "$CONTAINER_CMD pod rm --all -f $ERROR_REDIRECT" && print_status "success" "pods"
    print_status "success" "prune complete"
}

# =============================================================================
# BULK OPERATION COMMANDS
# =============================================================================

cmd_start() {
    resolve_bulk_targets "start"
    [[ ${#RESOLVED_CATEGORIES[@]} -eq 0 && ${#RESOLVED_SERVICES[@]} -eq 0 ]] && { print_status "warning" "No targets"; return 0; }

    ensure_network_exists

    for service in "${RESOLVED_SERVICES[@]}"; do
        start_individual_service "$service"
    done

    for category in "${RESOLVED_CATEGORIES[@]}"; do
        [[ "$opt_parallel_start" == "true" ]] && start_category_parallel "$category" || start_category_sequential "$category"
        [[ "$opt_wait_healthy" == "true" ]] && wait_for_category_health "$category"
    done
}

cmd_stop() {
    resolve_bulk_targets "stop"
    [[ ${#RESOLVED_CATEGORIES[@]} -eq 0 && ${#RESOLVED_SERVICES[@]} -eq 0 ]] && { print_status "warning" "No targets"; return 0; }

    for category in "${RESOLVED_CATEGORIES[@]}"; do
        stop_category "$category"
    done

    for service in "${RESOLVED_SERVICES[@]}"; do
        stop_individual_service "$service"
    done

    run_cleanup_operations
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
    start_container "$service_name"
}

stop_individual_service() {
    local service_name="$1"
    stop_container "$service_name" "$opt_timeout"
}

start_category_sequential() {
    local category="$1"
    local service_files=$(filter_services_for_category "$category")
    [[ -z "$service_files" ]] && return 0

    local -a files
    files=(${=service_files})
    local failed=0

    for service_file in "${files[@]}"; do
        start_individual_service "${service_file%.yml}" || ((failed++))
    done

    [[ $failed -gt 0 ]] && print_status "warning" "$category: $failed failed"
}

start_category_parallel() {
    local category="$1"
    local service_files=$(filter_services_for_category "$category")
    [[ -z "$service_files" ]] && return 0

    local -a files pids
    files=(${=service_files})

    for service_file in "${files[@]}"; do
        if [[ "$opt_dry_run" == "true" ]]; then
            start_individual_service "${service_file%.yml}"
        else
            start_individual_service "${service_file%.yml}" &
            pids+=($!)
        fi
    done

    [[ "$opt_dry_run" == "false" ]] && {
        local failed=0
        for pid in "${pids[@]}"; do wait "$pid" || ((failed++)); done
        [[ $failed -gt 0 ]] && print_status "warning" "$category: $failed failed"
    }
}

stop_category() {
    local category="$1"
    local service_files=$(filter_services_for_category "$category")
    [[ -z "$service_files" ]] && return 0

    local -a files
    files=(${=service_files})
    local failed=0

    for service_file in "${files[@]}"; do
        stop_individual_service "${service_file%.yml}" || ((failed++))
    done

    [[ $failed -gt 0 ]] && print_status "warning" "$category: $failed failed"
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
    [[ "$opt_wait_healthy" == "false" || "$opt_dry_run" == "true" ]] && return 0

    case "$category" in
        "database")
            local service_files=$(get_service_files "$category")
            [[ "$service_files" == *"mongodb.yml"* ]] && wait_for_mongodb "$opt_health_timeout"
            sleep 3
            ;;
        "proxy") sleep 5 ;;
        *) sleep 1 ;;
    esac
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

cleanup_service_resources() {
    local service_name="$1"
    [[ "$opt_dry_run" == "true" ]] && return 0
    [[ "$opt_cleanup_service_images" == "true" ]] && cleanup_service_images "$service_name"
    [[ "$opt_cleanup_service_volumes" == "true" && "$opt_preserve_data" == "false" ]] && cleanup_service_volumes "$service_name"
}

cleanup_service_images() {
    local service_name="$1"
    local images=$(eval "$CONTAINER_CMD images --filter 'label=com.docker.compose.service=$service_name' -q" 2>/dev/null)
    [[ -z "$images" ]] && return 0
    local removed=0
    for image in ${(f)images}; do
        eval "$CONTAINER_CMD rmi -f $image $ERROR_REDIRECT" && ((removed++))
    done
    [[ $VERBOSE -eq 1 && $removed -gt 0 ]] && print_status "info" "cleaned $removed images for $service_name"
}

cleanup_service_volumes() {
    local service_name="$1"
    local volumes=$(eval "$CONTAINER_CMD volume ls --filter 'label=com.docker.compose.service=$service_name' -q" 2>/dev/null)
    [[ -z "$volumes" ]] && return 0
    local removed=0
    for volume in ${(f)volumes}; do
        eval "$CONTAINER_CMD volume rm -f $volume $ERROR_REDIRECT" && ((removed++))
    done
    [[ $VERBOSE -eq 1 && $removed -gt 0 ]] && print_status "info" "cleaned $removed volumes for $service_name"
}

run_cleanup_operations() {
    [[ -n "$opt_cleanup_older_than" ]] && cleanup_old_resources "$opt_cleanup_older_than" "$opt_dry_run"
    [[ "$opt_cleanup_large_volumes" == "true" ]] && cleanup_large_volumes "$opt_max_volume_size" "$opt_max_volumes_remove" "$opt_dry_run"
    [[ "$opt_cleanup_orphans" == "true" ]] && cleanup_service_orphans "${opt_services_only:-}" "$opt_dry_run"
    [[ "$opt_remove_images" == "true" ]] && cleanup_images_global
    [[ "$opt_remove_volumes" == "true" ]] && cleanup_volumes_global
    [[ "$opt_remove_networks" == "true" ]] && cleanup_networks_global
    [[ "$opt_cleanup_orphans" == "true" || "$opt_remove_images" == "true" || "$opt_remove_volumes" == "true" ]] && cleanup_system_global
}

cleanup_images_global() {
    [[ "$opt_remove_images" == "false" ]] && return 0
    [[ "$opt_dry_run" == "true" ]] && { echo "DRY RUN: remove all images"; return 0; }
    local images=$(eval "$CONTAINER_CMD images -q $ERROR_REDIRECT")
    [[ -z "$images" ]] && return 0
    local removed=0
    for image in ${(f)images}; do
        eval "$CONTAINER_CMD rmi -f $image $ERROR_REDIRECT" && ((removed++))
    done
    [[ $removed -gt 0 ]] && print_status "success" "removed $removed images"
}

cleanup_volumes_global() {
    [[ "$opt_remove_volumes" == "false" ]] && return 0
    [[ "$opt_dry_run" == "true" ]] && { echo "DRY RUN: remove all volumes"; return 0; }
    local volumes=$(eval "$CONTAINER_CMD volume ls -q $ERROR_REDIRECT")
    [[ -z "$volumes" ]] && return 0
    local removed=0
    for volume in ${(f)volumes}; do
        eval "$CONTAINER_CMD volume rm -f $volume $ERROR_REDIRECT" && ((removed++))
    done
    [[ $removed -gt 0 ]] && print_status "success" "removed $removed volumes"
}

cleanup_networks_global() {
    [[ "$opt_remove_networks" == "false" ]] && return 0
    [[ "$opt_dry_run" == "true" ]] && { echo "DRY RUN: remove network $NETWORK_NAME"; return 0; }
    eval "$CONTAINER_CMD network exists $NETWORK_NAME $ERROR_REDIRECT" && \
        eval "$CONTAINER_CMD network rm $NETWORK_NAME $ERROR_REDIRECT" && \
        print_status "success" "removed network $NETWORK_NAME"
    eval "$CONTAINER_CMD network prune -f $ERROR_REDIRECT" || true
}

cleanup_system_global() {
    [[ "$opt_dry_run" == "true" ]] && { echo "DRY RUN: system prune"; return 0; }
    eval "$CONTAINER_CMD system prune -f $ERROR_REDIRECT" || true
}

cleanup_old_resources() {
    local max_age_days="$1"
    local dry_run="${2:-false}"
    [[ -z "$max_age_days" || ! "$max_age_days" =~ ^[0-9]+$ ]] && { print_status "error" "Invalid age: $max_age_days"; return 1; }
    [[ "$dry_run" == "true" ]] && { echo "DRY RUN: remove resources older than ${max_age_days}d"; return 0; }

    local cutoff_date=$(date -d "$max_age_days days ago" '+%Y-%m-%d' 2>/dev/null || date -v-${max_age_days}d '+%Y-%m-%d')
    local old_containers=$(eval "$CONTAINER_CMD ps -a --filter 'status=exited' --format '{{.ID}} {{.CreatedAt}}'" 2>/dev/null)
    local removed=0

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local cid=$(echo "$line" | awk '{print $1}')
        local cdate=$(echo "$line" | awk '{print $2}')
        [[ "$cdate" < "$cutoff_date" ]] && eval "$CONTAINER_CMD rm -f $cid $ERROR_REDIRECT" && ((removed++))
    done <<< "$old_containers"

    eval "$CONTAINER_CMD image prune -a --filter \"until=${max_age_days}h\" -f $ERROR_REDIRECT" || true
    eval "$CONTAINER_CMD volume prune --filter \"until=${max_age_days}h\" -f $ERROR_REDIRECT" || true
    [[ $removed -gt 0 ]] && print_status "success" "cleaned $removed old containers"
}

cleanup_large_volumes() {
    local max_size_mb="$1"
    local max_count="${2:-3}"
    local dry_run="${3:-false}"
    [[ -z "$max_size_mb" || ! "$max_size_mb" =~ ^[0-9]+$ ]] && { print_status "error" "Invalid size: $max_size_mb"; return 1; }
    [[ "$dry_run" == "true" ]] && { echo "DRY RUN: remove up to $max_count volumes > ${max_size_mb}MB"; return 0; }

    local volumes_info=$(eval "$CONTAINER_CMD volume ls --format '{{.Name}}'" 2>/dev/null)
    [[ -z "$volumes_info" ]] && return 0

    local -a large_volumes
    while IFS= read -r vname; do
        [[ -z "$vname" ]] && continue
        local vpath=$(eval "$CONTAINER_CMD volume inspect $vname --format '{{.Mountpoint}}'" 2>/dev/null)
        [[ -z "$vpath" ]] && continue
        local in_use=$(eval "$CONTAINER_CMD ps -a --filter \"volume=$vname\" --format '{{.Names}}'" 2>/dev/null)
        [[ -n "$in_use" ]] && continue
        local size_mb=$(sudo du -sm "$vpath" 2>/dev/null | awk '{print $1}' || echo "0")
        [[ "$size_mb" -gt "$max_size_mb" ]] && large_volumes+=("$vname:$size_mb")
    done <<< "$volumes_info"

    [[ ${#large_volumes[@]} -eq 0 ]] && return 0
    local -a sorted=($(printf '%s\n' "${large_volumes[@]}" | sort -t: -k2 -rn))
    local removed=0

    for vinfo in "${sorted[@]}"; do
        [[ $removed -ge $max_count ]] && break
        local vn="${vinfo%:*}"
        eval "$CONTAINER_CMD volume rm -f $vn $ERROR_REDIRECT" && ((removed++))
    done
    [[ $removed -gt 0 ]] && print_status "success" "removed $removed large volumes"
}

cleanup_service_orphans() {
    local target_services="$1"
    local dry_run="${2:-false}"
    [[ "$dry_run" == "true" ]] && { echo "DRY RUN: remove orphaned containers"; return 0; }

    local all_containers=$(eval "$CONTAINER_CMD ps -a --format '{{.Names}}'" 2>/dev/null)
    [[ -z "$all_containers" ]] && return 0

    local -a managed_services
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        local sfiles=$(get_service_files "$category")
        [[ -n "$sfiles" ]] && for sf in ${=sfiles}; do managed_services+=("${sf%.yml}"); done
    done
    [[ -n "$target_services" ]] && managed_services=(${(s:,:)target_services})

    local removed=0
    while IFS= read -r cname; do
        [[ -z "$cname" ]] && continue
        local is_managed=false
        for ms in "${managed_services[@]}"; do [[ "$cname" == "$ms" ]] && { is_managed=true; break; }; done
        eval "$CONTAINER_CMD inspect $cname --format '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}'" 2>/dev/null | grep -q "$NETWORK_NAME" && is_managed=true
        [[ "$is_managed" == "true" ]] && continue
        local has_labels=$(eval "$CONTAINER_CMD inspect $cname --format '{{index .Config.Labels \"com.docker.compose.project\"}}'" 2>/dev/null)
        [[ -z "$has_labels" ]] && eval "$CONTAINER_CMD rm -f $cname $ERROR_REDIRECT" && ((removed++))
    done <<< "$all_containers"
    [[ $removed -gt 0 ]] && print_status "success" "removed $removed orphans"
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
        "refresh")
            refresh_service_discovery
            print_status "success" "Service discovery refreshed"
            print_status "info" "Run 'list' to see updated services"
            ;;
        "check")
            cmd_check
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