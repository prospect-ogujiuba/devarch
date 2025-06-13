# =============================================================================
# stop-services.sh - Enhanced Service Stop Script  
# =============================================================================
#!/bin/zsh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# Script-specific options
STOP_CATEGORIES=()
EXCLUDE_CATEGORIES=()
REMOVE_VOLUMES=false
REMOVE_IMAGES=false
REMOVE_ORPHANS=false
FORCE_STOP=false

show_stop_help() {
    cat << EOF
Enhanced Service Stop Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -c CATEGORY     Stop only specific category (can be used multiple times)
    -x CATEGORY     Exclude specific category (can be used multiple times)
    -V              Remove volumes after stopping
    -I              Remove images after stopping
    -o              Remove orphaned containers
    -f              Force stop (kill instead of graceful stop)
    -h              Show this help message

CATEGORIES:
    database, dbms, backend, analytics, ai, mail, project, erp, proxy

EXAMPLES:
    $0                           # Stop all services gracefully
    $0 -c backend -c analytics   # Stop only backend and analytics
    $0 -V -I                     # Stop all and cleanup volumes/images
    $0 -f -v                     # Force stop with verbose output

EOF
}

parse_stop_args() {
    local OPTIND=1
    
    while getopts "sevdr:c:x:VIofh" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            c) STOP_CATEGORIES+=("$OPTARG") ;;
            x) EXCLUDE_CATEGORIES+=("$OPTARG") ;;
            V) REMOVE_VOLUMES=true ;;
            I) REMOVE_IMAGES=true ;;
            o) REMOVE_ORPHANS=true ;;
            f) FORCE_STOP=true ;;
            h) show_stop_help; exit 0 ;;
            ?) show_stop_help; exit 1 ;;
        esac
    done
}

determine_stop_categories() {
    local categories_to_stop=()
    
    if [ ${#STOP_CATEGORIES[@]} -gt 0 ]; then
        categories_to_stop=("${STOP_CATEGORIES[@]}")
        log "INFO" "Stopping only specified categories: ${categories_to_stop[*]}"
    else
        # Reverse order for graceful shutdown
        local reverse_order=()
        for ((i=${#SERVICE_STARTUP_ORDER[@]}-1; i>=0; i--)); do
            local category="${SERVICE_STARTUP_ORDER[$i]}"
            if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then
                reverse_order+=("$category")
            fi
        done
        categories_to_stop=("${reverse_order[@]}")
        
        if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then
            log "INFO" "Stopping all categories except: ${EXCLUDE_CATEGORIES[*]}"
        else
            log "INFO" "Stopping all categories"
        fi
    fi
    
    echo "${categories_to_stop[@]}"
}

stop_category() {
    local category="$1"
    local files=(${=SERVICE_CATEGORIES[$category]})
    
    log "INFO" "🛑 Stopping category: $category"
    
    for file in "${files[@]}"; do
        local compose_path="$COMPOSE_DIR/$file"
        local service_name=$(basename "$file" .yml)
        
        if [ ! -f "$compose_path" ]; then
            log "WARN" "Compose file not found: $compose_path, skipping"
            continue
        fi
        
        log "INFO" "  📦 $([ "$FORCE_STOP" = true ] && echo "Force stopping" || echo "Stopping") service: $service_name"
        
        if [ "$DRY_RUN" = true ]; then
            local action=$([ "$FORCE_STOP" = true ] && echo "kill" || echo "down")
            log "INFO" "  [DRY RUN] Would execute: ${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path $action"
            continue
        fi
        
        # Stop the service
        local action=$([ "$FORCE_STOP" = true ] && echo "kill" || echo "down")
        local cmd="${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path $action"
        
        # Add orphan removal if requested
        if [ "$REMOVE_ORPHANS" = true ] && [ "$action" = "down" ]; then
            cmd="$cmd --remove-orphans"
        fi
        
        execute_command "$([ "$FORCE_STOP" = true ] && echo "Force stop" || echo "Stop") $service_name" "$cmd" "$SHOW_ERRORS" false
        
        sleep 1  # Brief pause between services
    done
}

cleanup_resources() {
    if [ "$REMOVE_VOLUMES" = true ]; then
        log "INFO" "🗑️ Removing unused volumes..."
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would execute: ${SUDO_PREFIX}${CONTAINER_RUNTIME} volume prune -f"
        else
            execute_command "Remove unused volumes" "${SUDO_PREFIX}${CONTAINER_RUNTIME} volume prune -f" "$SHOW_ERRORS" false
        fi
    fi
    
    if [ "$REMOVE_IMAGES" = true ]; then
        log "INFO" "🗑️ Removing unused images..."
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would execute: ${SUDO_PREFIX}${CONTAINER_RUNTIME} image prune -f"
        else
            execute_command "Remove unused images" "${SUDO_PREFIX}${CONTAINER_RUNTIME} image prune -f" "$SHOW_ERRORS" false
        fi
    fi
}

main_stop() {
    parse_common_args "$@"
    parse_stop_args "$@"
    
    local categories_to_stop=($(determine_stop_categories))
    
    log "INFO" "🛑 $([ "$FORCE_STOP" = true ] && echo "Force stopping" || echo "Stopping") microservices..."
    
    # Change to compose directory
    cd "$COMPOSE_DIR" || {
        log "ERROR" "Failed to change to compose directory: $COMPOSE_DIR"
        exit 1
    }
    
    # Stop categories in reverse order
    for category in "${categories_to_stop[@]}"; do
        stop_category "$category"
        sleep 2  # Pause between categories
    done
    
    # Cleanup resources if requested
    cleanup_resources
    
    log "INFO" "✅ Service shutdown completed!"
}

# Only run if this script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main_stop "$@"
fi