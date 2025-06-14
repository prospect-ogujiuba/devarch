#!/bin/zsh

# =============================================================================
# MICROSERVICES INSTALLATION SCRIPT
# =============================================================================
# Streamlined installation script that leverages other modular scripts
# for a clean, maintainable setup process

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
opt_skip_ssl=false
opt_skip_trust=false
opt_quick_mode=false
opt_categories_only=""

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Complete installation script for the microservices architecture.
    Installs services in proper dependency order with optional configurations.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -y, --yes               Skip confirmation prompts
    -q, --quick             Quick mode: skip SSL and trust setup
    -d, --skip-db           Skip database initialization
    --skip-ssl              Skip SSL certificate generation
    --skip-trust            Skip SSL certificate trust installation
    -c, --categories LIST   Install only specific categories (comma-separated)
    -h, --help              Show this help message

CATEGORIES:
    database, db-tools, backend, analytics, ai-services, mail, project, erp, proxy

EXAMPLES:
    $0                                    # Full installation with prompts
    $0 -y                                # Full installation, no prompts
    $0 -q                                # Quick installation (no SSL setup)
    $0 -c database,backend               # Install only database and backend services
    $0 -s -e -y                          # Use sudo, show errors, skip prompts
    $0 --categories proxy --skip-db      # Install only proxy, skip database setup

NOTES:
    - Services are installed in dependency order automatically
    - Database services are always installed first when included
    - SSL certificates work across all operating systems
    - Use --quick mode for development-only setups
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
            -q|--quick)
                opt_quick_mode=true
                opt_skip_ssl=true
                opt_skip_trust=true
                shift
                ;;
            -d|--skip-db)
                opt_skip_db_setup=true
                shift
                ;;
            --skip-ssl)
                opt_skip_ssl=true
                shift
                ;;
            --skip-trust)
                opt_skip_trust=true
                shift
                ;;
            -c|--categories)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_categories_only="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a value"
                fi
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
    echo "  Skip SSL Setup: $opt_skip_ssl"
    echo "  Skip Trust Setup: $opt_skip_trust"
    
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
        echo "The installation will create containers, networks, and SSL certificates."
        echo ""
        read "response?Do you want to continue? (y/N): "
        
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_status "info" "Installation cancelled by user"
            exit 0
        fi
    fi
}

install_services() {
    local -a categories_to_install
    
    if [[ -n "$opt_categories_only" ]]; then
        categories_to_install=(${(s:,:)opt_categories_only})
    else
        categories_to_install=("${SERVICE_STARTUP_ORDER[@]}")
    fi
    
    print_status "step" "Starting service installation..."
    
    # Ensure network exists first
    ensure_network_exists
    
    # Install each category in order
    for category in "${categories_to_install[@]}"; do
        print_status "step" "Installing $category services..."
        start_service_category "$category"
        
        # Special handling for database category
        if [[ "$category" == "database" ]]; then
            # Wait for MongoDB specifically if it's being installed
            if [[ -n "${SERVICE_CATEGORIES[database]}" ]] && [[ "${SERVICE_CATEGORIES[database]}" == *"mongodb.yml"* ]]; then
                wait_for_mongodb
            fi
            
            # Wait a bit for other databases to initialize
            print_status "step" "Waiting for databases to initialize..."
            sleep 10
            
            # Run database setup if not skipped
            if [[ "$opt_skip_db_setup" == "false" ]]; then
                run_database_setup
            fi
        fi
        
        # Brief pause between categories
        sleep 2
    done
    
    print_status "success" "Service installation completed!"
}

run_database_setup() {
    print_status "step" "Running database setup..."
    
    local setup_args=()
    [[ "$opt_use_sudo" == "true" ]] && setup_args+=("-s")
    [[ "$opt_show_errors" == "true" ]] && setup_args+=("-e")
    setup_args+=("-m" "-p")  # Setup both MariaDB and PostgreSQL
    
    if "$SCRIPT_DIR/setup-databases.sh" "${setup_args[@]}"; then
        print_status "success" "Database setup completed"
    else
        print_status "warning" "Database setup encountered issues but continuing..."
    fi
}

setup_ssl_certificates() {
    if [[ "$opt_skip_ssl" == "true" ]]; then
        print_status "info" "Skipping SSL certificate generation"
        return 0
    fi
    
    print_status "step" "Setting up SSL certificates..."
    
    local ssl_args=()
    [[ "$opt_use_sudo" == "true" ]] && ssl_args+=("-s")
    [[ "$opt_show_errors" == "true" ]] && ssl_args+=("-e")
    
    if "$SCRIPT_DIR/setup-ssl.sh" "${ssl_args[@]}"; then
        print_status "success" "SSL certificates generated successfully"
    else
        print_status "warning" "SSL setup encountered issues but continuing..."
    fi
}

install_ssl_trust() {
    if [[ "$opt_skip_trust" == "true" ]]; then
        print_status "info" "Skipping SSL certificate trust installation"
        return 0
    fi
    
    print_status "step" "Installing SSL certificates for host system..."
    
    local trust_args=()
    [[ "$opt_use_sudo" == "true" ]] && trust_args+=("-s")
    [[ "$opt_show_errors" == "true" ]] && trust_args+=("-e")
    
    # Add Windows certificate copying if on WSL
    [[ "$OS_TYPE" == "wsl" ]] && trust_args+=("-w")
    
    if "$SCRIPT_DIR/trust-host.sh" "${trust_args[@]}"; then
        print_status "success" "SSL certificates trusted by host system"
    else
        print_status "warning" "SSL trust setup encountered issues but continuing..."
    fi
}

show_completion_message() {
    echo ""
    echo "========================================================"
    print_status "success" "Installation completed successfully!"
    echo "========================================================"
    echo ""
    
    if [[ "$opt_quick_mode" == "true" ]]; then
        echo "Quick mode installation completed. Services are accessible via localhost ports."
        echo "Run the following for SSL setup later:"
        echo "  $SCRIPT_DIR/setup-ssl.sh && $SCRIPT_DIR/trust-host.sh"
        echo ""
    fi
    
    echo "You can now access your services:"
    echo ""
    echo "🌐 Web Interfaces:"
    echo "  - Nginx Proxy Manager: https://nginx.test (or http://localhost:81)"
    echo "  - Grafana Dashboard: https://grafana.test (or http://localhost:9001)"
    echo "  - Metabase Analytics: https://metabase.test (or http://localhost:8085)"
    echo ""
    echo "🗄️  Database Tools:"
    echo "  - Adminer: https://adminer.test (or http://localhost:8082)"
    echo "  - phpMyAdmin: https://phpmyadmin.test (or http://localhost:8083)"
    echo "  - Mongo Express: https://mongodb.test (or http://localhost:8084)"
    echo ""
    echo "🤖 AI & Analytics:"
    echo "  - n8n Workflows: https://n8n.test (or http://localhost:9100)"
    echo "  - Langflow: https://langflow.test (or http://localhost:9110)"
    echo "  - Kibana: https://kibana.test (or http://localhost:9120)"
    echo ""
    echo "📊 Business Applications:"
    echo "  - Odoo ERP: https://odoo.test (or http://localhost:9300)"
    echo "  - Matomo Analytics: https://matomo.test (or http://localhost:9010)"
    echo ""
    echo "💻 Development Tools:"
    echo "  - Mailpit: https://mailpit.test (or http://localhost:9200)"
    echo "  - Gitea: https://gitea.test (or http://localhost:9210)"
    echo ""
    echo "📋 Management Commands:"
    echo "  - View all services: $SCRIPT_DIR/show-services.sh"
    echo "  - Stop all services: $SCRIPT_DIR/stop-services.sh"
    echo "  - Start services: $SCRIPT_DIR/start-services.sh"
    echo ""
    echo "🔧 Default Credentials:"
    echo "  - Username: admin"
    echo "  - Password: 123456"
    echo "  - Email: admin@site.test"
    echo ""
    echo "========================================================"
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
    
    # Install services
    install_services
    
    # Setup SSL if not skipped
    setup_ssl_certificates
    
    # Trust SSL certificates if not skipped
    install_ssl_trust
    
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