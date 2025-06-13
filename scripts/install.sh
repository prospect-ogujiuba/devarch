#!/bin/zsh
# Enhanced Installation Script for Microservices Architecture
# Installs and configures the entire microservices environment with improved error handling and flexibility

# Source configuration and dependencies
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# =============================================================================
# SCRIPT-SPECIFIC CONFIGURATION
# =============================================================================

# Installation options
SKIP_CONFIRMATION=false
SKIP_DB_SETUP=false
SKIP_SSL_SETUP=false
SKIP_TRUST_SETUP=false
INSTALL_CATEGORIES=()
EXCLUDE_CATEGORIES=()
PARALLEL_INSTALL=false
INSTALL_TIMEOUT=300

# =============================================================================
# HELP FUNCTION
# =============================================================================

show_help() {
    cat << EOF
Enhanced Microservices Installation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -y              Skip confirmation prompts
    -D              Skip database setup
    -S              Skip SSL certificate setup
    -T              Skip trust certificate setup
    -c CATEGORY     Install only specific category (can be used multiple times)
    -x CATEGORY     Exclude specific category (can be used multiple times)
    -p              Enable parallel installation (faster but less debugging)
    -t SECONDS      Installation timeout in seconds (default: 300)
    -h              Show this help message

CATEGORIES:
    database        PostgreSQL, MySQL, MariaDB, MongoDB, Redis
    dbms           Adminer, pgAdmin, phpMyAdmin, Mongo Express, Metabase, NocoDB
    backend        PHP, Node.js, Python, Go, .NET development environments
    analytics      Elasticsearch, Kibana, Grafana, Prometheus, Matomo
    ai             Langflow, n8n workflow automation
    mail           Mailpit email testing
    project        Gitea project management
    erp            Odoo ERP system
    proxy          Nginx Proxy Manager

EXAMPLES:
    $0                           # Full installation with prompts
    $0 -y -v                     # Full installation, skip prompts, verbose
    $0 -c database -c dbms       # Install only database and dbms categories
    $0 -x ai -x erp             # Install everything except AI and ERP
    $0 -p -t 600                # Parallel installation with 10 minute timeout
    $0 -d -v                    # Dry run with verbose output to see what would happen

EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_install_args() {
    local OPTIND=1  # Reset OPTIND
    
    while getopts "sevdr:yDSTc:x:pt:h" opt; do
        case $opt in
            s|e|v|d|r) ;; # These are handled by parse_common_args
            y) SKIP_CONFIRMATION=true ;;
            D) SKIP_DB_SETUP=true ;;
            S) SKIP_SSL_SETUP=true ;;
            T) SKIP_TRUST_SETUP=true ;;
            c) INSTALL_CATEGORIES+=("$OPTARG") ;;
            x) EXCLUDE_CATEGORIES+=("$OPTARG") ;;
            p) PARALLEL_INSTALL=true ;;
            t) INSTALL_TIMEOUT="$OPTARG" ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_categories() {
    local valid_categories=(${(k)SERVICE_CATEGORIES})
    
    # Validate install categories
    for category in "${INSTALL_CATEGORIES[@]}"; do
        if [[ ! " ${valid_categories[@]} " =~ " ${category} " ]]; then
            log "ERROR" "Invalid category: $category"
            log "INFO" "Valid categories: ${valid_categories[*]}"
            exit 1
        fi
    done
    
    # Validate exclude categories
    for category in "${EXCLUDE_CATEGORIES[@]}"; do
        if [[ ! " ${valid_categories[@]} " =~ " ${category} " ]]; then
            log "ERROR" "Invalid exclude category: $category"
            log "INFO" "Valid categories: ${valid_categories[*]}"
            exit 1
        fi
    done
}

pre_installation_checks() {
    log "INFO" "Performing pre-installation checks..."
    
    # Check container runtime
    log "DEBUG" "Container runtime: $CONTAINER_RUNTIME"
    
    # Check if running as root (when using sudo)
    if [ "$USE_SUDO" = true ]; then
        if ! sudo -n true 2>/dev/null; then
            log "WARN" "Sudo access required but not available. Some operations may fail."
        fi
    fi
    
    # Check available disk space (at least 10GB recommended)
    local available_space=$(df "$PROJECT_ROOT" | awk 'NR==2 {print $4}')
    local required_space=10485760  # 10GB in KB
    
    if [ "$available_space" -lt "$required_space" ]; then
        log "WARN" "Low disk space detected. At least 10GB recommended for full installation."
    fi
    
    # Validate compose files
    local total_files=0
    local valid_files=0
    
    for category in "${(k)SERVICE_CATEGORIES[@]}"; do
        local files=(${=SERVICE_CATEGORIES[$category]})
        for file in "${files[@]}"; do
            total_files=$((total_files + 1))
            local compose_path="$COMPOSE_DIR/$file"
            if validate_compose_file "$compose_path"; then
                valid_files=$((valid_files + 1))
            fi
        done
    done
    
    log "INFO" "Compose file validation: $valid_files/$total_files files valid"
    
    if [ "$valid_files" -ne "$total_files" ]; then
        log "ERROR" "Some compose files are invalid. Please check the errors above."
        [ "$DRY_RUN" = false ] && exit 1
    fi
}

# =============================================================================
# INSTALLATION FUNCTIONS
# =============================================================================

determine_install_categories() {
    local categories_to_install=()
    
    if [ ${#INSTALL_CATEGORIES[@]} -gt 0 ]; then
        # Use only specified categories
        categories_to_install=("${INSTALL_CATEGORIES[@]}")
        log "INFO" "Installing only specified categories: ${categories_to_install[*]}"
    else
        # Use all categories except excluded ones
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then
                categories_to_install+=("$category")
            fi
        done
        
        if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then
            log "INFO" "Installing all categories except: ${EXCLUDE_CATEGORIES[*]}"
        else
            log "INFO" "Installing all categories"
        fi
    fi
    
    echo "${categories_to_install[@]}"
}

install_category() {
    local category="$1"
    local files=(${=SERVICE_CATEGORIES[$category]})
    
    log "INFO" "📦 Installing category: $category"
    
    for file in "${files[@]}"; do
        local compose_path="$COMPOSE_DIR/$file"
        local service_name=$(basename "$file" .yml)
        
        if [ ! -f "$compose_path" ]; then
            log "WARN" "Compose file not found: $compose_path, skipping"
            continue
        fi
        
        log "INFO" "  🚀 Starting service: $service_name"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "  [DRY RUN] Would execute: ${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path up -d"
            continue
        fi
        
        # Start the service
        local cmd="${SUDO_PREFIX}${CONTAINER_RUNTIME} compose -f $compose_path up -d"
        if [ "$SHOW_ERRORS" = true ]; then
            execute_command "Start $service_name" "$cmd" true false
        else
            execute_command "Start $service_name" "$cmd" false false
        fi
        
        # Wait a bit between services to avoid overwhelming the system
        [ "$PARALLEL_INSTALL" = false ] && sleep 2
    done
    
    # Health check for critical services
    if [[ "$category" == "database" ]]; then
        log "INFO" "  🔍 Performing health checks for database services..."
        sleep 5  # Give databases time to initialize
        
        for file in "${files[@]}"; do
            local service_name=$(basename "$file" .yml)
            check_service_health "$service_name" "$CONTAINER_RUNTIME" "$SUDO_PREFIX" 15 3
        done
    fi
}

install_services() {
    local categories_to_install=($(determine_install_categories))
    
    log "INFO" "Starting installation of ${#categories_to_install[@]} categories..."
    
    # Ensure network exists
    ensure_network "$CONTAINER_RUNTIME" "$SUDO_PREFIX"
    
    # Change to compose directory
    cd "$COMPOSE_DIR" || {
        log "ERROR" "Failed to change to compose directory: $COMPOSE_DIR"
        exit 1
    }
    
    if [ "$PARALLEL_INSTALL" = true ]; then
        log "INFO" "Using parallel installation mode"
        
        # Install categories in parallel (be careful with dependencies)
        for category in "${categories_to_install[@]}"; do
            install_category "$category" &
        done
        
        # Wait for all background jobs
        wait
    else
        # Sequential installation (safer for dependencies)
        for category in "${categories_to_install[@]}"; do
            install_category "$category"
            
            # Add delay between categories for system stability
            sleep 3
        done
    fi
}

setup_databases() {
    if [ "$SKIP_DB_SETUP" = true ]; then
        log "INFO" "Skipping database setup as requested"
        return 0
    fi
    
    log "INFO" "🗄️ Setting up databases..."
    
    # Wait for database services to be ready
    log "INFO" "Waiting for database services to initialize..."
    sleep 15
    
    # Run database setup script
    local db_setup_script="$SCRIPT_DIR/setup-databases.sh"
    if [ -f "$db_setup_script" ]; then
        local cmd="$db_setup_script"
        [ "$USE_SUDO" = true ] && cmd="$cmd -s"
        [ "$SHOW_ERRORS" = true ] && cmd="$cmd -e"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would execute: $cmd"
        else
            execute_command "Setup databases" "$cmd" "$SHOW_ERRORS" false
        fi
    else
        log "WARN" "Database setup script not found: $db_setup_script"
    fi
}

setup_ssl_certificates() {
    if [ "$SKIP_SSL_SETUP" = true ]; then
        log "INFO" "Skipping SSL setup as requested"
        return 0
    fi
    
    log "INFO" "🔐 Setting up SSL certificates..."
    
    # Wait for nginx-proxy-manager to be ready
    log "INFO" "Waiting for nginx-proxy-manager to start..."
    sleep 10
    
    local ssl_setup_script="$SCRIPT_DIR/setup-ssl.sh"
    if [ -f "$ssl_setup_script" ]; then
        local cmd="$ssl_setup_script"
        [ "$USE_SUDO" = true ] && cmd="$cmd -s"
        [ "$SHOW_ERRORS" = true ] && cmd="$cmd -e"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would execute: $cmd"
        else
            execute_command "Setup SSL certificates" "$cmd" "$SHOW_ERRORS" false
        fi
    else
        log "WARN" "SSL setup script not found: $ssl_setup_script"
    fi
}

trust_ssl_certificates() {
    if [ "$SKIP_TRUST_SETUP" = true ]; then
        log "INFO" "Skipping SSL trust setup as requested"
        return 0
    fi
    
    log "INFO" "🛡️ Installing SSL certificates in system trust store..."
    
    local trust_script="$SCRIPT_DIR/trust-host.sh"
    if [ -f "$trust_script" ]; then
        local cmd="$trust_script"
        [ "$USE_SUDO" = true ] && cmd="$cmd -s"
        [ "$SHOW_ERRORS" = true ] && cmd="$cmd -e"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would execute: $cmd"
        else
            execute_command "Trust SSL certificates" "$cmd" "$SHOW_ERRORS" false
        fi
    else
        log "WARN" "Trust setup script not found: $trust_script"
    fi
}

# =============================================================================
# CONFIRMATION AND SUMMARY
# =============================================================================

show_installation_summary() {
    local categories_to_install=($(determine_install_categories))
    
    cat << EOF

📋 INSTALLATION SUMMARY
======================

Project Root: $PROJECT_ROOT
Container Runtime: $CONTAINER_RUNTIME
Categories to Install: ${categories_to_install[*]}
$([ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ] && echo "Excluded Categories: ${EXCLUDE_CATEGORIES[*]}")

Installation Options:
- Skip Confirmation: $SKIP_CONFIRMATION
- Skip Database Setup: $SKIP_DB_SETUP
- Skip SSL Setup: $SKIP_SSL_SETUP
- Skip Trust Setup: $SKIP_TRUST_SETUP
- Parallel Install: $PARALLEL_INSTALL
- Use Sudo: $USE_SUDO
- Show Errors: $SHOW_ERRORS
- Dry Run: $DRY_RUN

Estimated Installation Time: $([ "$PARALLEL_INSTALL" = true ] && echo "5-10 minutes" || echo "10-15 minutes")

EOF
}

confirm_installation() {
    if [ "$SKIP_CONFIRMATION" = true ] || [ "$DRY_RUN" = true ]; then
        return 0
    fi
    
    echo "Do you want to proceed with the installation? (y/N)"
    read -r response
    
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log "INFO" "Installation cancelled by user"
        exit 0
    fi
}

# =============================================================================
# SUCCESS REPORTING
# =============================================================================

show_success_message() {
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "🔍 Dry run completed successfully!"
        return 0
    fi
    
    cat << EOF

🎉 INSTALLATION COMPLETED SUCCESSFULLY!
======================================

Your microservices environment is now running!

📊 Dashboard Access:
- Nginx Proxy Manager: https://nginx.test (admin panel: http://localhost:81)
- Grafana Monitoring: https://grafana.test (admin:123456)
- Metabase Analytics: https://metabase.test

🗄️ Database Management:
- Adminer: https://adminer.test
- phpMyAdmin: https://phpmyadmin.test  
- Mongo Express: https://mongodb.test
- pgAdmin: https://pgadmin.test

🚀 Development Tools:
- Mailpit: https://mailpit.test
- Gitea: https://gitea.test

💼 Business Applications:
- Odoo ERP: https://odoo.test

🤖 AI & Automation:
- n8n Workflows: https://n8n.test
- Langflow: https://langflow.test

🔧 Useful Commands:
- View services: $SCRIPT_DIR/show-services.sh
- Stop all: $SCRIPT_DIR/stop-services.sh
- Restart all: $SCRIPT_DIR/start-services.sh

📝 Next Steps:
1. Configure your proxy rules in Nginx Proxy Manager
2. Set up your first database connections
3. Deploy your applications to the apps/ directory
4. Configure monitoring dashboards in Grafana

EOF
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse arguments
    parse_common_args "$@"
    parse_install_args "$@"
    
    # Validate inputs
    validate_categories
    
    # Show summary
    show_installation_summary
    
    # Confirm installation
    confirm_installation
    
    # Pre-installation checks
    pre_installation_checks
    
    # Start installation
    log "INFO" "🚀 Starting microservices installation..."
    
    # Install services
    install_services
    
    # Post-installation setup
    setup_databases
    setup_ssl_certificates
    trust_ssl_certificates
    
    # Show success message
    show_success_message
    
    log "INFO" "✅ Installation process completed!"
}

# Trap to cleanup on exit
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ] && [ "$DRY_RUN" = false ]; then
        log "ERROR" "Installation failed with exit code $exit_code"
        log "INFO" "You can retry with: $0 ${original_args[*]}"
        log "INFO" "Or check logs in: $LOGS_DIR/scripts.log"
    fi
}

# Store original arguments for retry message
original_args=("$@")
trap cleanup EXIT

# Run main function
main "$@"