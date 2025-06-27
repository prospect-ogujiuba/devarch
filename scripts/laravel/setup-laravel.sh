#!/bin/zsh

# =============================================================================
# LARAVEL PROJECT SETUP SCRIPT - ENHANCED
# =============================================================================
# Creates new Laravel projects or clones existing ones with config.sh integration

# Source the central configuration
. "$(dirname "$0")/../config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

PROJECT_NAME=""
GIT_URL=""
opt_force=false
opt_skip_npm=false
opt_skip_migrate=false

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 <project-name> [OPTIONS]

DESCRIPTION:
    Creates new Laravel projects or clones existing repositories.
    Uses config.sh for consistent container operations and path management.

ARGUMENTS:
    project-name        Name of the Laravel project to create

OPTIONS:
    -r, --repo URL      Clone from existing Git repository
    -f, --force         Overwrite existing project directory
    --skip-npm          Skip NPM dependency installation
    --skip-migrate      Skip database migrations
    -h, --help          Show this help message

EXAMPLES:
    $0 my-blog                                              # Create fresh Laravel project
    $0 my-blog -r https://github.com/user/laravel-blog.git # Clone existing repo
    $0 my-new-app --force                                   # Overwrite existing project
    $0 api-project --skip-npm --skip-migrate               # Minimal setup

REQUIREMENTS:
    - PHP container must be running (use: service-manager.sh up php)
    - Apps directory: $APPS_DIR

NOTES:
    - Projects are created in: $APPS_DIR/<project-name>
    - Accessible at: https://<project-name>.test
    - Uses container runtime: $CONTAINER_RUNTIME
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ $# -eq 0 ]]; then
        print_status "error" "No project name specified"
        show_usage
        exit 1
    fi
    
    PROJECT_NAME="$1"
    shift
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -r|--repo)
                if [[ -n "$2" && "$2" != -* ]]; then
                    GIT_URL="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a repository URL"
                fi
                ;;
            -f|--force)
                opt_force=true
                shift
                ;;
            --skip-npm)
                opt_skip_npm=true
                shift
                ;;
            --skip-migrate)
                opt_skip_migrate=true
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
    
    # Validate project name
    if [[ -z "$PROJECT_NAME" ]]; then
        handle_error "Project name is required"
    fi
    
    # Validate project name format
    if [[ ! "$PROJECT_NAME" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        handle_error "Project name can only contain letters, numbers, hyphens, and underscores"
    fi
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_environment() {
    print_status "step" "Validating Laravel setup environment..."
    
    # Check if PHP service exists and is running
    if ! validate_service_exists "php"; then
        handle_error "PHP service not found. Please ensure PHP service is configured in your compose files."
    fi
    
    local php_status=$(get_service_status "php")
    if [[ "$php_status" != "running" ]]; then
        print_status "warning" "PHP container is not running (status: $php_status)"
        print_status "info" "Starting PHP container..."
        
        if ! start_single_service "php"; then
            handle_error "Failed to start PHP container. Run: ./scripts/service-manager.sh up php"
        fi
        
        # Wait a moment for container to be ready
        sleep 3
    else
        print_status "success" "PHP container is running"
    fi
    
    # Check apps directory exists
    if [[ ! -d "$APPS_DIR" ]]; then
        print_status "step" "Creating apps directory: $APPS_DIR"
        mkdir -p "$APPS_DIR"
    fi
    
    print_status "success" "Environment validation passed"
}

validate_project_setup() {
    local project_path="$APPS_DIR/$PROJECT_NAME"
    
    print_status "step" "Validating project setup..."
    
    # Check if project already exists
    if [[ -d "$project_path" ]]; then
        if [[ "$opt_force" == "true" ]]; then
            print_status "warning" "Project '$PROJECT_NAME' exists, removing due to --force flag"
            rm -rf "$project_path"
        else
            handle_error "Project '$PROJECT_NAME' already exists. Use --force to overwrite."
        fi
    fi
    
    # Validate git URL if provided
    if [[ -n "$GIT_URL" ]]; then
        if ! echo "$GIT_URL" | grep -qE '^https?://|^git@'; then
            handle_error "Invalid Git URL format: $GIT_URL"
        fi
    fi
    
    print_status "success" "Project validation passed"
}

# =============================================================================
# PROJECT CREATION FUNCTIONS
# =============================================================================

create_laravel_project() {
    print_status "step" "Setting up Laravel project: $PROJECT_NAME"
    
    # Navigate to apps directory (using config.sh variable)
    cd "$APPS_DIR" || handle_error "Failed to access apps directory: $APPS_DIR"
    
    if [[ -n "$GIT_URL" ]]; then
        create_from_repository
    else
        create_fresh_project
    fi
    
    # Navigate to project directory
    cd "$PROJECT_NAME" || handle_error "Failed to enter project directory"
    
    setup_project_environment
    install_dependencies
    configure_application
    
    print_status "success" "Laravel project '$PROJECT_NAME' created successfully!"
    show_completion_message
}

create_from_repository() {
    print_status "step" "Cloning project from repository..."
    print_status "info" "Repository: $GIT_URL"
    
    if ! git clone "$GIT_URL" "$PROJECT_NAME"; then
        handle_error "Failed to clone repository: $GIT_URL"
    fi
    
    print_status "success" "Repository cloned successfully"
}

create_fresh_project() {
    print_status "step" "Creating fresh Laravel project..."
    
    # Execute Laravel installer inside PHP container using config.sh variables
    local create_command="laravel new $PROJECT_NAME"
    
    if ! eval "$CONTAINER_CMD exec -w /var/www/html php zsh -c '$create_command'"; then
        handle_error "Failed to create Laravel project. Ensure Laravel installer is available in PHP container."
    fi
    
    print_status "success" "Fresh Laravel project created"
}

setup_project_environment() {
    print_status "step" "Setting up project environment..."
    
    # Copy .env file
    if [[ -f ".env.example" ]]; then
        cp .env.example .env
        print_status "success" "Environment file created from .env.example"
    elif [[ -f ".env.sample" ]]; then
        cp .env.sample .env
        print_status "success" "Environment file created from .env.sample"
    else
        print_status "warning" "No .env template found"
    fi
    
    # Set proper permissions using config.sh container commands
    print_status "step" "Setting file permissions..."
    eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php zsh -c 'chmod -R 775 /var/www/html/$PROJECT_NAME'"
    eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php zsh -c 'chown -R www-data:www-data .'"
    
    print_status "success" "Project environment configured"
}

install_dependencies() {
    print_status "step" "Installing project dependencies..."
    
    # Install Composer dependencies
    print_status "info" "Installing Composer dependencies..."
    if ! eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php composer install"; then
        handle_error "Composer install failed"
    fi
    
    # Install NPM dependencies (unless skipped)
    if [[ "$opt_skip_npm" == "false" ]]; then
        print_status "info" "Installing NPM dependencies..."
        if ! eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php npm install"; then
            print_status "warning" "NPM install failed, continuing..."
        else
            print_status "success" "NPM dependencies installed"
        fi
    else
        print_status "info" "Skipping NPM installation as requested"
    fi
    
    print_status "success" "Dependencies installed successfully"
}

configure_application() {
    print_status "step" "Configuring Laravel application..."
    
    # Generate application key (only if .env exists and APP_KEY is empty)
    if [[ -f '.env' ]] && ! grep -q 'APP_KEY=.*[^=]' .env; then
        print_status "info" "Generating application key..."
        if ! eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php php artisan key:generate"; then
            print_status "warning" "Key generation failed"
        else
            print_status "success" "Application key generated"
        fi
    fi
    
    # Build assets (unless NPM was skipped)
    if [[ "$opt_skip_npm" == "false" ]]; then
        print_status "info" "Building frontend assets..."
        if eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php npm run build"; then
            print_status "success" "Assets built successfully"
        else
            print_status "warning" "Asset build failed, continuing..."
        fi
    fi
    
    # Run migrations (unless skipped and database is configured)
    if [[ "$opt_skip_migrate" == "false" ]]; then
        if grep -q 'DB_DATABASE=.*[^=]' .env 2>/dev/null; then
            print_status "info" "Running database migrations..."
            if eval "$CONTAINER_CMD exec -w /var/www/html/$PROJECT_NAME php php artisan migrate --force"; then
                print_status "success" "Database migrations completed"
            else
                print_status "warning" "Migration failed (database might not be configured)"
            fi
        else
            print_status "info" "No database configured, skipping migrations"
        fi
    else
        print_status "info" "Skipping migrations as requested"
    fi
    
    print_status "success" "Application configuration completed"
}

# =============================================================================
# COMPLETION AND SUMMARY
# =============================================================================

show_completion_message() {
    echo ""
    echo "========================================================"
    print_status "success" "Laravel Project Setup Complete!"
    echo "========================================================"
    echo ""
    
    echo "Project Details:"
    echo "  Name: $PROJECT_NAME"
    echo "  Location: $APPS_DIR/$PROJECT_NAME"
    echo "  URL: https://${PROJECT_NAME}.test"
    echo ""
    
    if [[ -n "$GIT_URL" ]]; then
        echo "  Source: $GIT_URL"
        echo ""
    fi
    
    echo "Next Steps:"
    echo "  1. Add to hosts file (if not done already):"
    echo "     echo '127.0.0.1 ${PROJECT_NAME}.test' | sudo tee -a /etc/hosts"
    echo ""
    echo "  2. Configure your .env file if needed:"
    echo "     cd $APPS_DIR/$PROJECT_NAME"
    echo "     nano .env"
    echo ""
    echo "  3. Access your application:"
    echo "     https://${PROJECT_NAME}.test"
    echo ""
    
    echo "Management Commands:"
    echo "  View logs: ./scripts/service-manager.sh logs php"
    echo "  Restart PHP: ./scripts/service-manager.sh restart php"
    echo "  Project shell: $CONTAINER_CMD exec -it php zsh"
    echo ""
    
    # Show database setup if database services are running
    local db_services=("mariadb" "mysql" "postgres")
    local running_db=""
    
    for db in "${db_services[@]}"; do
        if [[ "$(get_service_status "$db" 2>/dev/null)" == "running" ]]; then
            running_db="$db"
            break
        fi
    done
    
    if [[ -n "$running_db" ]]; then
        echo "Database Information:"
        echo "  Running database: $running_db"
        echo "  Setup databases: ./scripts/setup-databases.sh -a"
        echo "  Default credentials: admin / 123456"
        echo ""
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context (use config.sh defaults)
    setup_command_context "$DEFAULT_SUDO" "false"
    
    # Validate environment and requirements
    validate_environment
    validate_project_setup
    
    # Create the Laravel project
    create_laravel_project
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi