# =============================================================================
# start-services.sh - Enhanced Service Startup Script
# =============================================================================
#!/bin/zsh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# Script-specific options
START_CATEGORIES=()
EXCLUDE_CATEGORIES=()
RESTART_MODE=false
HEALTH_CHECK=true
WAIT_FOR_HEALTHY=false

show_help() {
    cat << EOF
Enhanced Service Startup Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages  
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -c CATEGORY     Start only specific category (can be used multiple times)
    -x CATEGORY     Exclude specific category (can be used multiple times)
    -R              Restart services instead of just starting
    -H              Skip health checks
    -w              Wait for services to be healthy before continuing
    -h              Show this help message

CATEGORIES:
    database, dbms, backend, analytics, ai, mail, project, erp, proxy

EXAMPLES:
    $0                           # Start all services
    $0 -c database -c dbms       # Start only database services
    $0 -R -v                     # Restart all services with verbose output
    $0 -x ai -x erp             # Start all except AI and ERP services

EOF
}

parse_start_args() {
    local OPTIND=1
    
    while getopts "sevdr:c:x:RHwh" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            c) START_CATEGORIES+=("$OPTARG") ;;
            x) EXCLUDE_CATEGORIES+=("$OPTARG") ;;
            R) RESTART_MODE=true ;;
            H) HEALTH_CHECK=false ;;
            w) WAIT_FOR_HEALTHY=true ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
}

determine_start_categories() {
    local categories_to_start=()
    
    if [ ${#START_CATEGORIES[@]} -gt 0 ]; then
        categories_to_start=("${START_CATEGORIES[@]}")
        log "INFO" "Starting only specified categories: ${categories_to_start[*]}"
    else
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then
                categories_to_start+=("$category")
            fi
        done
        
        if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then
            log "INFO" "Starting all categories except: ${EXCLUDE_CATEGORIES[*]}"
        else
            log "INFO" "Starting all categories"
        fi
    fi
    
    echo "${categories_to_start[@]}"
}

start_category() {
    local category="$1"
    local files=(${=SERVICE_CATEGORIES[$category]})
    
    log "INFO" "🚀 Starting category: $category"
    
    for file in "${files[@]}"; do
        local compose_path="$COMPOSE_DIR/$file"
        local service_name=$(basename "$file" .yml)
        
        if [ ! -f "$compose_path" ]; then
            log "WARN" "Compose file not found: $compose_path, skipping"
            continue
        fi
        
        log "INFO" "  📦 $([ "$RESTART_MODE" = true ] && echo "Restarting" || echo "Starting") service: $service_name"
        
        if [ "$DRY_RUN" = true ]; then
            local action=$([ "$RESTART_MODE" = true ] && echo "restart" || echo "up -d")
            log "INFO" "  [DRY RUN] Would execute: ${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path $action"
            continue
        fi
        
        # Start or restart the service
        local action=$([ "$RESTART_MODE" = true ] && echo "restart" || echo "up -d")
        local cmd="${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path $action"
        
        execute_command "$([ "$RESTART_MODE" = true ] && echo "Restart" || echo "Start") $service_name" "$cmd" "$SHOW_ERRORS" false
        
        # Health check if enabled
        if [ "$HEALTH_CHECK" = true ] && [ "$WAIT_FOR_HEALTHY" = true ]; then
            log "INFO" "  🔍 Checking health of $service_name..."
            check_service_health "$service_name" "$CONTAINER_RUNTIME" "$SUDO_PREFIX" 10 2
        fi
        
        sleep 1  # Brief pause between services
    done
}

main() {
    parse_common_args "$@"
    parse_start_args "$@"
    
    local categories_to_start=($(determine_start_categories))
    
    log "INFO" "$([ "$RESTART_MODE" = true ] && echo "🔄 Restarting" || echo "🚀 Starting") microservices..."
    
    # Ensure network exists
    ensure_network "$CONTAINER_RUNTIME" "$SUDO_PREFIX"
    
    # Change to compose directory
    cd "$COMPOSE_DIR" || {
        log "ERROR" "Failed to change to compose directory: $COMPOSE_DIR"
        exit 1
    }
    
    # Start categories in order
    for category in "${categories_to_start[@]}"; do
        start_category "$category"
        sleep 2  # Pause between categories
    done
    
    # Final health check if requested
    if [ "$HEALTH_CHECK" = true ] && [ "$WAIT_FOR_HEALTHY" = false ]; then
        log "INFO" "🔍 Performing final health checks..."
        sleep 5
        
        # Check critical services
        local critical_services=("postgres" "mariadb" "nginx-proxy-manager")
        for service in "${critical_services[@]}"; do
            if ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service" 2>/dev/null; then
                check_service_health "$service" "$CONTAINER_RUNTIME" "$SUDO_PREFIX" 5 2
            fi
        done
    fi
    
    log "INFO" "✅ Service startup completed!"
    
    if [ "$DRY_RUN" = false ]; then
        log "INFO" "💡 Run 'show-services.sh' to see all available services"
    fi
}

main "$@"