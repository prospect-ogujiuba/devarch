#!/bin/zsh

# =============================================================================
# MICROSERVICES INSTALLATION SCRIPT - SIMPLIFIED
# =============================================================================
# Streamlined installation using service-manager for all service operations

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_skip_confirmation=false
opt_skip_db_setup=false
opt_categories_only=""
opt_parallel=false
opt_force_recreate=false
opt_rebuild_services=false

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Complete installation script for the microservices architecture.
    Uses service-manager for robust dependency handling and service management.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -y, --yes               Skip confirmation prompts
    -d, --skip-db           Skip database initialization
    -c, --categories LIST   Install only specific categories (comma-separated)
    -p, --parallel          Start services in parallel within categories
    -f, --force             Force recreate containers during installation
    -r, --rebuild           Rebuild services before starting them
    -h, --help              Show this help message

CATEGORIES:
    $(printf '%s, ' "${SERVICE_STARTUP_ORDER[@]}" | sed 's/, $//')

EXAMPLES:
    $0                                   # Full installation with prompts
    $0 -y                                # Full installation, no prompts
    $0 -c database,backend               # Install only database and backend services
    $0 -s -e -y                          # Use sudo, show errors, skip prompts
    $0 -c proxy -d                       # Install only proxy, skip database setup
    $0 -p -f                             # Parallel installation with forced recreation

NOTES:
    - Services are installed in dependency order automatically via service-manager
    - Database services are always installed first when included
    - Parallel mode speeds up installation but may be harder to debug
    - Force recreation ensures clean containers but takes longer
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
            -y|--yes)
                opt_skip_confirmation=true
                shift
                ;;
            -d|--skip-db)
                opt_skip_db_setup=true
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
            -p|--parallel)
                opt_parallel=true
                shift
                ;;
            -f|--force)
                opt_force_recreate=true
                shift
                ;;
            -r|--rebuild)
                opt_rebuild_services=true
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

validate_environment() {
    print_status "step" "Validating environment..."
    
    # Check for container runtime
    if ! command -v "$CONTAINER_RUNTIME" >/dev/null 2>&1; then
        handle_error "$CONTAINER_RUNTIME is not installed or not in PATH"
    fi
    
    # Check service-manager exists
    if [[ ! -f "$SCRIPT_DIR/service-manager.sh" ]]; then
        handle_error "service-manager.sh not found at $SCRIPT_DIR/service-manager.sh"
    fi
    
    # Check compose functionality
    if ! eval "$CONTAINER_CMD --version $ERROR_REDIRECT"; then
        handle_error "Container runtime is not working properly"
    fi
    
    # Check project structure
    if [[ ! -d "$COMPOSE_DIR" ]]; then
        handle_error "Compose directory not found: $COMPOSE_DIR"
    fi
    
    # Check for .env file
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        if [[ -f "$PROJECT_ROOT/.env-sample" ]]; then
            print_status "warning" "No .env file found, copying from .env-sample"
            cp "$PROJECT_ROOT/.env-sample" "$PROJECT_ROOT/.env"
            print_status "success" ".env file created from sample"
        else
            handle_error "No .env file found and no .env-sample available"
        fi
    fi
    
    print_status "success" "Environment validation passed"
}

validate_categories() {
    if [[ -n "$opt_categories_only" ]]; then
        local -a requested_categories
        requested_categories=(${(s:,:)opt_categories_only})
        
        for category in "${requested_categories[@]}"; do
            if [[ -z "${SERVICE_CATEGORIES[$category]}" ]]; then
                print_status "error" "Invalid category: $category"
                print_status "info" "Available categories: ${(k)SERVICE_CATEGORIES}"
                exit 1
            fi
        done
        
        print_status "info" "Will install categories: ${requested_categories[*]}"
    fi
}

# =============================================================================
# INSTALLATION FUNCTIONS
# =============================================================================

show_installation_summary() {
    print_status "info" "Installation Summary:"
    echo "  Container Runtime: $CONTAINER_RUNTIME"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Show Errors: $opt_show_errors"
    echo "  Skip DB Setup: $opt_skip_db_setup"
    echo "  Parallel Mode: $opt_parallel"
    echo "  Force Recreate: $opt_force_recreate"
    echo "  Rebuild Services: $opt_rebuild_services"
    
    if [[ -n "$opt_categories_only" ]]; then
        echo "  Categories: $opt_categories_only"
    else
        echo "  Categories: All (${SERVICE_STARTUP_ORDER[*]})"
    fi
    
    echo ""
}

confirm_installation() {
    if [[ "$opt_skip_confirmation" == "false" ]]; then
        echo "This script will install and configure the microservices environment."
        echo "The installation will create containers, networks, and volumes."
        echo ""
        read "response?Do you want to continue? (y/N): "
        
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_status "info" "Installation cancelled by user"
            exit 0
        fi
    fi
}

install_services() {
    print_status "step" "Starting service installation via service-manager..."
    
    # Build service-manager arguments
    local -a service_manager_args
    
    # Determine the command
    if [[ -n "$opt_categories_only" ]]; then
        service_manager_args+=("start" "--categories" "$opt_categories_only")
    else
        service_manager_args+=("start-all")
    fi
    
    # Add service-manager options based on install options
    [[ "$opt_use_sudo" == "true" ]] && service_manager_args+=("--sudo")
    [[ "$opt_show_errors" == "true" ]] && service_manager_args+=("--errors")
    [[ "$opt_parallel" == "true" ]] && service_manager_args+=("--parallel")
    [[ "$opt_force_recreate" == "true" ]] && service_manager_args+=("--force-services")
    [[ "$opt_rebuild_services" == "true" ]] && service_manager_args+=("--rebuild-services")
    
    # Always wait for health checks during installation
    service_manager_args+=("--health-timeout" "120")
    
    print_status "info" "Executing: service-manager.sh ${service_manager_args[*]}"
    
    # Execute service-manager
    if "$SCRIPT_DIR/service-manager.sh" "${service_manager_args[@]}"; then
        print_status "success" "Service installation completed successfully!"
        
        # Handle database setup if needed
        handle_database_setup
    else
        handle_error "Service installation failed. Check logs for details."
    fi
}

handle_database_setup() {
    # Check if database category was installed
    local db_installed=false
    
    if [[ -n "$opt_categories_only" ]]; then
        # Check if database was in the requested categories
        if [[ "$opt_categories_only" == *"database"* ]]; then
            db_installed=true
        fi
    else
        # Full installation includes database
        db_installed=true
    fi
    
    if [[ "$db_installed" == "true" && "$opt_skip_db_setup" == "false" ]]; then
        print_status "step" "Waiting for databases to initialize..."
        sleep 10
        
        # Wait for MongoDB specifically if it's installed
        if [[ -n "${SERVICE_CATEGORIES[database]}" ]] && [[ "${SERVICE_CATEGORIES[database]}" == *"mongodb.yml"* ]]; then
            wait_for_mongodb 60
        fi
        
        run_database_setup
    else
        if [[ "$opt_skip_db_setup" == "true" ]]; then
            print_status "info" "Database setup skipped as requested"
        else
            print_status "info" "No database category installed, skipping database setup"
        fi
    fi
}

run_database_setup() {
    print_status "step" "Running database setup..."
    
    # Check if setup-databases.sh exists
    if [[ ! -f "$SCRIPT_DIR/setup-databases.sh" ]]; then
        print_status "warning" "setup-databases.sh not found, skipping database initialization"
        return 0
    fi
    
    local setup_args=()
    [[ "$opt_use_sudo" == "true" ]] && setup_args+=("--sudo")
    [[ "$opt_show_errors" == "true" ]] && setup_args+=("--errors")
    setup_args+=("--all")  # Setup all available databases
    
    if "$SCRIPT_DIR/setup-databases.sh" "${setup_args[@]}"; then
        print_status "success" "Database setup completed"
    else
        print_status "warning" "Database setup encountered issues but continuing..."
        print_status "info" "You can run database setup manually later: ./scripts/setup-databases.sh --all"
    fi
}

show_completion_message() {
    echo ""
    echo "========================================================"
    print_status "success" "Installation completed successfully!"
    echo "========================================================"
    echo ""
    
    echo "Your microservices environment is now ready!"
    echo ""
    echo "Quick commands to get started:"
    echo "  ./scripts/service-manager.sh status        # Check service status"
    echo "  ./scripts/service-manager.sh ps            # List running containers"
    echo "  ./scripts/service-manager.sh logs SERVICE  # View service logs"
    echo ""
    
    # Show service URLs if proxy is installed
    local proxy_installed=false
    if [[ -n "$opt_categories_only" ]]; then
        [[ "$opt_categories_only" == *"proxy"* ]] && proxy_installed=true
    else
        proxy_installed=true
    fi
    
    if [[ "$proxy_installed" == "true" ]]; then
        echo "Service URLs (add to /etc/hosts):"
        if [[ -f "$PROJECT_ROOT/context/hosts.txt" ]]; then
            echo "  Check: ./context/hosts.txt for complete host entries"
        fi
        echo "  Example: https://nginx.test (Nginx Proxy Manager)"
        echo "  Example: https://portainer.test (Container Management)"
        echo ""
    fi
    
    echo "For troubleshooting:"
    echo "  ./scripts/service-manager.sh --help        # View all options"
    echo "  ./scripts/service-manager.sh down SERVICE  # Stop specific service"
    echo "  ./scripts/service-manager.sh stop-all      # Stop all services"
    echo ""
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Validate environment and arguments
    validate_environment
    validate_categories
    
    # Show summary and get confirmation
    show_installation_summary
    confirm_installation
    
    # Start installation process
    print_status "step" "Starting microservices installation..."
    
    # Install services via service-manager
    install_services
    
    # Show completion message
    show_completion_message
    
    print_status "success" "Installation process completed!"
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi