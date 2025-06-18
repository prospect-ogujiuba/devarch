#!/bin/zsh

# =============================================================================
# MICROSERVICES SERVICE MANAGER
# =============================================================================
# Individual service management with smart defaults and enhanced UX
# Replaces manual podman compose commands with intelligent automation

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

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 COMMAND [SERVICE|CATEGORY] [OPTIONS]

DESCRIPTION:
    Smart service manager for individual microservices. Replaces manual
    podman compose commands with intelligent automation and better UX.

COMMANDS:
    up SERVICE              Start a single service
    down SERVICE            Stop a single service
    restart SERVICE         Restart a single service
    rebuild SERVICE         Rebuild and restart a service
    logs SERVICE            Show service logs
    status [SERVICE]        Show service status (or all if no service)
    ps                      List all running services
    list                    List all available services

SERVICE MANAGEMENT OPTIONS:
    -f, --force             Force recreate containers (for up/rebuild)
    --no-cache              Don't use cache when rebuilding
    --remove-volumes        Remove volumes when stopping
    -t, --timeout SECONDS   Graceful shutdown timeout (default: 30)

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
    $0 up postgres                      # Start postgres service
    $0 down adminer --remove-volumes    # Stop adminer and remove volumes
    $0 rebuild nginx --no-cache         # Rebuild nginx without cache
    $0 restart grafana                  # Restart grafana service
    $0 logs postgres --follow           # Follow postgres logs
    $0 status                           # Show status of all services
    $0 ps                              # List running services
    $0 list                            # List all available services

SMART FEATURES:
    - Auto-detects service compose files
    - Provides helpful error messages with suggestions
    - Safe defaults (preserves data unless explicitly requested)
    - Enhanced progress feedback vs native podman compose
    - Category support (coming soon)

NOTES:
    - Service names correspond to compose file names (postgres.yml → postgres)
    - Use --remove-volumes with caution (destroys persistent data)
    - Dry run mode shows exact commands that would be executed
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
    
    # Second argument is service name (for most commands)
    if [[ -n "$1" && "$1" != -* ]]; then
        SERVICE_NAME="$1"
        shift
    fi
    
    # Parse remaining options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                opt_force_recreate=true
                shift
                ;;
            --no-cache)
                opt_no_cache=true
                shift
                ;;
            --remove-volumes)
                opt_remove_volumes=true
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
    local valid_commands=("up" "down" "restart" "rebuild" "logs" "status" "ps" "list")
    
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

validate_service_exists() {
    if [[ -n "$SERVICE_NAME" ]]; then
        if ! validate_service_exists "$SERVICE_NAME"; then
            print_status "error" "Service '$SERVICE_NAME' not found"
            print_status "info" "Available services: $(list_all_service_names | tr '\n' ' ')"
            exit 1
        fi
    fi
}

# =============================================================================
# COMMAND EXECUTION FUNCTIONS
# =============================================================================
# Add these functions to service-manager.sh after the validation functions

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
    
    # Use podman ps with our network filter
    if eval "$CONTAINER_CMD ps --filter 'network=$NETWORK_NAME' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' $ERROR_REDIRECT"; then
        eval "$CONTAINER_CMD ps --filter 'network=$NETWORK_NAME' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'"
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
    validate_service_exists
    
    # Execute command
    case "$COMMAND" in
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