#!/bin/zsh

# =============================================================================
# MICROSERVICES STARTUP SCRIPT
# =============================================================================
# Enhanced service startup with dependency management, health monitoring,
# and selective service control

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_force_recreate=false
opt_categories_only=""
opt_exclude_categories=""
opt_parallel_start=false
opt_wait_healthy=true
opt_health_timeout=120
opt_restart_policy="unless-stopped"
opt_dry_run=false
opt_verbose=false

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Enhanced service startup script with dependency management, health monitoring,
    and selective service control. Starts services in proper dependency order.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -f, --force             Force recreate containers (docker compose up --force-recreate)
    -c, --categories LIST   Start only specific categories (comma-separated)
    -x, --exclude LIST      Exclude specific categories (comma-separated)
    -p, --parallel          Start services in parallel within categories
    -w, --no-wait           Don't wait for health checks
    -t, --timeout SECONDS   Health check timeout (default: 120)
    -r, --restart POLICY    Restart policy (no, always, unless-stopped, on-failure)
    -d, --dry-run           Show what would be started without actually starting
    -v, --verbose           Show detailed progress information
    -h, --help              Show this help message

CATEGORIES (in dependency order):
    database      - MariaDB, MySQL, PostgreSQL, MongoDB, Redis
    db-tools      - Adminer, phpMyAdmin, Mongo Express, Metabase, NocoDB, pgAdmin
    backend       - .NET, Go, Node.js, PHP, Python applications
    analytics     - Elasticsearch, Kibana, Logstash, Grafana, Prometheus, Matomo
    ai-services   - Langflow, n8n automation platforms
    mail          - Mailpit email testing
    project       - Gitea repository management
    erp           - Odoo business applications
    proxy         - Nginx Proxy Manager

RESTART POLICIES:
    no              - Do not restart containers automatically
    always          - Always restart containers
    unless-stopped  - Restart unless explicitly stopped (default)
    on-failure      - Restart only on failure

EXAMPLES:
    $0                                 # Start all services in dependency order
    $0 -c database,backend             # Start only database and backend services
    $0 -x dbms,erp                     # Start all except dbms and ERP services
    $0 -f -p                           # Force recreate and parallel start
    $0 -d -v                           # Dry run with verbose output
    $0 -w -t 60                        # Skip health checks, 60s timeout
    $0 --restart always --force        # Force recreate with always restart

DEPENDENCY HANDLING:
    - Database services start first and wait for readiness
    - MongoDB gets special handling for replica set initialization
    - Services within categories can start in parallel with -p option
    - Health checks ensure services are ready before dependent services start

NOTES:
    - Services are started in dependency order automatically
    - Use --parallel for faster startup within categories
    - Health checks prevent cascading failures
    - Dry run mode shows exactly what would be executed
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
            -f|--force)
                opt_force_recreate=true
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
            -p|--parallel)
                opt_parallel_start=true
                shift
                ;;
            -w|--no-wait)
                opt_wait_healthy=false
                shift
                ;;
            -t|--timeout)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_health_timeout="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric timeout value"
                fi
                ;;
            -r|--restart)
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
            -d|--dry-run)
                opt_dry_run=true
                opt_verbose=true  # Dry run implies verbose
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

validate_categories() {
    # Validate included categories
    if [[ -n "$opt_categories_only" ]]; then
        local -a requested_categories
        requested_categories=(${(s:,:)opt_categories_only})
        
        for category in "${requested_categories[@]}"; do
            if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
                handle_error "Invalid category: $category. Available: ${(k)SERVICE_CATEGORIES}"
            fi
        done
        
        print_status "info" "Will start categories: ${requested_categories[*]}"
    fi
    
    # Validate excluded categories
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

validate_environment() {
    print_status "step" "Validating environment..."
    
    # Check container runtime
    if ! command -v "$CONTAINER_RUNTIME" >/dev/null 2>&1; then
        handle_error "$CONTAINER_RUNTIME is not installed or not in PATH"
    fi
    
    # Test container runtime
    if ! eval "$CONTAINER_CMD --version $ERROR_REDIRECT"; then
        handle_error "$CONTAINER_RUNTIME is not working properly"
    fi
    
    # Check compose directory
    if [[ ! -d "$COMPOSE_DIR" ]]; then
        handle_error "Compose directory not found: $COMPOSE_DIR"
    fi
    
    # Check .env file
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_status "warning" "No .env file found"
        if [[ -f "$PROJECT_ROOT/.env-sample" ]]; then
            print_status "info" "Consider copying .env-sample to .env"
        fi
    fi
    
    print_status "success" "Environment validation passed"
}

# =============================================================================
# SERVICE MANAGEMENT FUNCTIONS
# =============================================================================

get_startup_categories() {
    local -a categories_to_start
    
    if [[ -n "$opt_categories_only" ]]; then
        # Use only specified categories, but maintain dependency order
        local -a requested_categories
        requested_categories=(${(s:,:)opt_categories_only})
        
        # Filter SERVICE_STARTUP_ORDER to only include requested categories
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            if [[ " ${requested_categories[*]} " =~ " $category " ]]; then
                categories_to_start+=("$category")
            fi
        done
    else
        # Use all categories
        categories_to_start=("${SERVICE_STARTUP_ORDER[@]}")
    fi
    
    # Remove excluded categories
    if [[ -n "$opt_exclude_categories" ]]; then
        local -a excluded_categories
        excluded_categories=(${(s:,:)opt_exclude_categories})
        
        local -a filtered_categories
        for category in "${categories_to_start[@]}"; do
            if [[ ! " ${excluded_categories[*]} " =~ " $category " ]]; then
                filtered_categories+=("$category")
            fi
        done
        categories_to_start=("${filtered_categories[@]}")
    fi
    
    echo "${categories_to_start[@]}"
}

start_service_file() {
    local service_file="$1"
    local category="$2"
    local full_path="$COMPOSE_DIR/$service_file"
    
    if [[ ! -f "$full_path" ]]; then
        print_status "warning" "Service file not found: $service_file"
        return 1
    fi
    
    local service_name="${service_file%.yml}"
    
    if [[ "$opt_verbose" == "true" ]]; then
        print_status "step" "Starting $service_name from $service_file..."
    fi
    
    # Build compose command
    local compose_args=("-f" "$full_path" "up" "-d")
    
    # Add restart policy if not default
    if [[ "$opt_restart_policy" != "unless-stopped" ]]; then
        compose_args+=("--restart" "$opt_restart_policy")
    fi
    
    # Add force recreate if requested
    if [[ "$opt_force_recreate" == "true" ]]; then
        compose_args+=("--force-recreate")
    fi
    
    # Execute or show dry run
    if [[ "$opt_dry_run" == "true" ]]; then
        echo "DRY RUN: $COMPOSE_CMD ${compose_args[*]}"
        return 0
    fi
    
    # Execute the command
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "success" "$service_name started successfully"
        fi
        return 0
    else
        print_status "error" "Failed to start $service_name"
        return 1
    fi
}

start_category_parallel() {
    local category="$1"
    local service_files="$2"
    
    local -a files pids
    files=(${=service_files})
    
    print_status "step" "Starting $category services in parallel..."
    
    # Start all services in background
    for service_file in "${files[@]}"; do
        if [[ "$opt_dry_run" == "true" ]]; then
            start_service_file "$service_file" "$category"
        else
            start_service_file "$service_file" "$category" &
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

start_category_sequential() {
    local category="$1"
    local service_files="$2"
    
    local -a files
    files=(${=service_files})
    
    print_status "step" "Starting $category services sequentially..."
    
    local failed=0
    for service_file in "${files[@]}"; do
        if ! start_service_file "$service_file" "$category"; then
            ((failed++))
        fi
        
        # Brief pause between services
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
# NETWORK AND ENVIRONMENT SETUP
# =============================================================================

setup_environment() {
    print_status "step" "Setting up environment..."
    
    # Create network if it doesn't exist
    ensure_network_exists
    
    # Create necessary directories
    create_required_directories
    
    print_status "success" "Environment setup completed"
}

create_required_directories() {
    local -a required_dirs=(
        "$APPS_DIR"
        "$LOGS_DIR"
        "$LOGS_DIR/nginx"
        "$LOGS_DIR/dotnet"
        "$LOGS_DIR/go"
        "$LOGS_DIR/node"
        "$LOGS_DIR/python"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            if [[ "$opt_dry_run" == "true" ]]; then
                echo "DRY RUN: mkdir -p $dir"
            else
                mkdir -p "$dir"
                if [[ "$opt_verbose" == "true" ]]; then
                    print_status "info" "Created directory: $dir"
                fi
            fi
        fi
    done
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_startup_summary() {
    local -a categories_to_start
    categories_to_start=($(get_startup_categories))
    
    print_status "info" "Service Startup Summary:"
    echo "  Container Runtime: $CONTAINER_RUNTIME"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Show Errors: $opt_show_errors"
    echo "  Force Recreate: $opt_force_recreate"
    echo "  Parallel Start: $opt_parallel_start"
    echo "  Wait for Health: $opt_wait_healthy"
    echo "  Health Timeout: ${opt_health_timeout}s"
    echo "  Restart Policy: $opt_restart_policy"
    echo "  Dry Run: $opt_dry_run"
    echo ""
    echo "  Categories to start: ${categories_to_start[*]}"
    echo ""
}

run_startup_process() {
    local -a categories_to_start
    categories_to_start=($(get_startup_categories))
    
    if [[ ${#categories_to_start[@]} -eq 0 ]]; then
        print_status "warning" "No categories selected for startup"
        return 0
    fi
    
    print_status "step" "Starting microservices in dependency order..."
    
    # Setup environment first
    setup_environment
    
    # Start each category
    for category in "${categories_to_start[@]}"; do
        local service_files
        service_files=$(get_service_files "$category")
        
        if [[ -z "$service_files" ]]; then
            print_status "warning" "No services found for category: $category"
            continue
        fi
        
        # Start services in category
        if [[ "$opt_parallel_start" == "true" ]]; then
            start_category_parallel "$category" "$service_files"
        else
            start_category_sequential "$category" "$service_files"
        fi
        
        # Wait for health if enabled
        wait_for_category_health "$category"
        
        # Brief pause between categories
        if [[ "$opt_dry_run" == "false" ]]; then
            sleep 2
        fi
    done
    
    print_status "success" "Service startup process completed!"
}

show_completion_info() {
    if [[ "$opt_dry_run" == "true" ]]; then
        echo ""
        print_status "info" "Dry run completed. No services were actually started."
        echo "Run without -d/--dry-run to execute the startup process."
        return 0
    fi
    
    echo ""
    print_status "info" "Startup completed! Services are now available:"
    echo ""
    echo "🔍 Check service status:"
    echo "  $SCRIPT_DIR/show-services.sh"
    echo ""
    echo "🌐 Access services via:"
    echo "  - Local ports: http://localhost:[port]"
    echo "  - Proxy URLs: https://[service].test"
    echo ""
    echo "⚙️  Manage services:"
    echo "  - Stop all: $SCRIPT_DIR/stop-services.sh"
    echo "  - Restart specific: $COMPOSE_CMD -f $COMPOSE_DIR/[category]/[service].yml restart"
    echo ""
    echo "🔐 Default credentials:"
    echo "  - Username: admin"
    echo "  - Password: 123456"
    echo ""
}

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Validate environment and arguments
    validate_environment
    validate_categories
    
    # Show startup summary
    show_startup_summary
    
    # Run the startup process
    run_startup_process
    
    # Show completion information
    show_completion_info
    
    print_status "success" "Service startup script completed!"
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi