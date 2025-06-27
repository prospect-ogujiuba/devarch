#!/bin/zsh

# =============================================================================
# WORDPRESS INSTALLATION SCRIPT - ENHANCED & UNIFIED
# =============================================================================
# Unified WordPress installer with multiple presets and config.sh integration

# Source the central configuration
. "$(dirname "$0")/../config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

SITE_NAME=""
SITE_TITLE=""
opt_preset="bare"
opt_force=false
opt_skip_email=false
opt_skip_permissions=false
opt_custom_domain=""

# WordPress presets
typeset -A WORDPRESS_PRESETS
WORDPRESS_PRESETS=(
    [bare]="Minimal WordPress with basic plugins"
    [clean]="WordPress with TypeRocket Pro and premium plugins"
    [custom]="WordPress with ACF, Gravity Forms, and custom setup"
    [loaded]="WordPress with all development tools and debugging plugins"
    [starred]="WordPress with starred repository plugins only"
)

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 <site-name> [OPTIONS]

DESCRIPTION:
    Unified WordPress installer with multiple presets and config.sh integration.
    Consolidates functionality from all individual WordPress install scripts.

ARGUMENTS:
    site-name           Name for the WordPress site (will be database name)

OPTIONS:
    -t, --title TITLE   Site title (defaults to formatted site name)
    -p, --preset TYPE   WordPress preset type (default: bare)
    -d, --domain DOMAIN Custom domain (defaults to <site-name>.test)
    -f, --force         Overwrite existing installation
    --skip-email        Skip sending installation emails
    --skip-permissions  Skip setting file permissions
    -h, --help          Show this help message

PRESETS:
$(for preset in "${(@k)WORDPRESS_PRESETS}"; do
    printf "    %-10s  %s\n" "$preset" "${WORDPRESS_PRESETS[$preset]}"
done)

EXAMPLES:
    $0 myblog                                    # Create basic WordPress site
    $0 myblog -p custom -t "My Custom Blog"     # Custom preset with title
    $0 devsite -p loaded --force                # Development site with all tools
    $0 client-site -d client.test -p clean      # Clean preset with custom domain

REQUIREMENTS:
    - PHP container must be running (use: service-manager.sh up php)
    - MariaDB container must be running (use: service-manager.sh up mariadb)
    - Apps directory: $APPS_DIR

NOTES:
    - Sites are created in: $APPS_DIR/<site-name>
    - Database name matches site name
    - Uses environment variables from: $PROJECT_ROOT/.env
    - Accessible at: https://<site-name>.test (or custom domain)
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ $# -eq 0 ]]; then
        print_status "error" "No site name specified"
        show_usage
        exit 1
    fi
    
    SITE_NAME="$1"
    shift
    
    # Default title is formatted site name
    SITE_TITLE=$(echo "$SITE_NAME" | sed 's/[_-]/ /g' | sed 's/\b\w/\U&/g')
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--title)
                if [[ -n "$2" && "$2" != -* ]]; then
                    SITE_TITLE="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a site title"
                fi
                ;;
            -p|--preset)
                if [[ -n "$2" && "$2" != -* ]]; then
                    if [[ -n "${WORDPRESS_PRESETS[$2]}" ]]; then
                        opt_preset="$2"
                        shift 2
                    else
                        handle_error "Invalid preset: $2. Available: ${(k)WORDPRESS_PRESETS}"
                    fi
                else
                    handle_error "Option $1 requires a preset type"
                fi
                ;;
            -d|--domain)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_custom_domain="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a domain name"
                fi
                ;;
            -f|--force)
                opt_force=true
                shift
                ;;
            --skip-email)
                opt_skip_email=true
                shift
                ;;
            --skip-permissions)
                opt_skip_permissions=true
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
    
    # Validate site name
    if [[ ! "$SITE_NAME" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        handle_error "Site name can only contain letters, numbers, hyphens, and underscores"
    fi
    
    # Set domain
    if [[ -z "$opt_custom_domain" ]]; then
        opt_custom_domain="${SITE_NAME}.test"
    fi
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_environment() {
    print_status "step" "Validating WordPress environment..."
    
    # Check if required services exist and are running
    local required_services=("php" "mariadb")
    
    for service in "${required_services[@]}"; do
        if ! validate_service_exists "$service"; then
            handle_error "$service service not found. Please ensure it's configured in your compose files."
        fi
        
        local service_status=$(get_service_status "$service")
        if [[ "$service_status" != "running" ]]; then
            print_status "warning" "$service container is not running (status: $service_status)"
            print_status "info" "Starting $service container..."
            
            if ! start_single_service "$service"; then
                handle_error "Failed to start $service container. Run: ./scripts/service-manager.sh up $service"
            fi
            
            # Wait a moment for container to be ready
            sleep 3
        else
            print_status "success" "$service container is running"
        fi
    done
    
    # Check apps directory exists
    if [[ ! -d "$APPS_DIR" ]]; then
        print_status "step" "Creating apps directory: $APPS_DIR"
        mkdir -p "$APPS_DIR"
    fi
    
    # Check .env file exists and source it
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        handle_error "Environment file not found: $PROJECT_ROOT/.env"
    fi
    
    # Source environment variables (using config.sh context)
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
    
    print_status "success" "Environment validation passed"
}

validate_site_setup() {
    local site_path="$APPS_DIR/$SITE_NAME"
    
    print_status "step" "Validating site setup..."
    
    # Check if site already exists
    if [[ -d "$site_path" ]]; then
        if [[ "$opt_force" == "true" ]]; then
            print_status "warning" "Site '$SITE_NAME' exists, removing due to --force flag"
            rm -rf "$site_path"
        else
            handle_error "Site '$SITE_NAME' already exists. Use --force to overwrite."
        fi
    fi
    
    print_status "success" "Site validation passed"
}

# =============================================================================
# WORDPRESS CORE INSTALLATION
# =============================================================================

install_wordpress_core() {
    local site_dir="$APPS_DIR/$SITE_NAME"
    local wp_dir="$site_dir/public"
    
    print_status "step" "Installing WordPress core..."
    
    # Create directory structure
    mkdir -p "$site_dir"
    
    # Download WordPress using config.sh container commands
    print_status "info" "Downloading WordPress..."
    if ! eval "$CONTAINER_CMD exec php wp core download --path='$wp_dir' --allow-root"; then
        handle_error "Failed to download WordPress"
    fi
    
    # Create wp-config.php
    print_status "info" "Creating wp-config.php..."
    local db_config_cmd="wp config create --dbname='$SITE_NAME' --dbuser='$MARIADB_USER' --dbpass='$MARIADB_PASSWORD' --dbhost='$MARIADB_HOST' --path='$wp_dir' --allow-root"
    
    if ! eval "$CONTAINER_CMD exec php $db_config_cmd"; then
        handle_error "Failed to create wp-config.php"
    fi
    
    # Set debugging
    eval "$CONTAINER_CMD exec php wp config set WP_DEBUG false --raw --path='$wp_dir' --allow-root"
    
    # Create database
    print_status "info" "Creating database..."
    if ! eval "$CONTAINER_CMD exec php wp db create --path='$wp_dir' --allow-root"; then
        handle_error "Failed to create database"
    fi
    
    # Install WordPress
    print_status "info" "Installing WordPress..."
    local install_cmd="wp core install --url='$opt_custom_domain' --title='$SITE_TITLE' --admin_user='$ADMIN_USER' --admin_password='$ADMIN_PASSWORD' --admin_email='$ADMIN_EMAIL' --path='$wp_dir' --allow-root"
    
    if ! eval "$CONTAINER_CMD exec php $install_cmd"; then
        handle_error "Failed to install WordPress"
    fi
    
    # Configure uploads
    eval "$CONTAINER_CMD exec php wp option update uploads_use_yearmonth_folders 0 --path='$wp_dir' --allow-root"
    
    print_status "success" "WordPress core installed successfully"
}

# =============================================================================
# PRESET-SPECIFIC CONFIGURATIONS
# =============================================================================

configure_wordpress_preset() {
    local wp_dir="$APPS_DIR/$SITE_NAME/public"
    
    print_status "step" "Configuring WordPress preset: $opt_preset"
    
    # Clean up default content
    cleanup_default_content "$wp_dir"
    
    # Apply preset-specific configuration
    case "$opt_preset" in
        "bare")
            configure_preset_bare "$wp_dir"
            ;;
        "clean")
            configure_preset_clean "$wp_dir"
            ;;
        "custom")
            configure_preset_custom "$wp_dir"
            ;;
        "loaded")
            configure_preset_loaded "$wp_dir"
            ;;
        "starred")
            configure_preset_starred "$wp_dir"
            ;;
        *)
            handle_error "Unknown preset: $opt_preset"
            ;;
    esac
    
    print_status "success" "Preset configuration completed"
}

cleanup_default_content() {
    local wp_dir="$1"
    
    print_status "info" "Cleaning up default content..."
    
    # Delete default posts and pages
    eval "$CONTAINER_CMD exec php wp post delete 1 --force --path='$wp_dir' --allow-root" 2>/dev/null || true
    eval "$CONTAINER_CMD exec php wp post delete 2 --force --path='$wp_dir' --allow-root" 2>/dev/null || true
    eval "$CONTAINER_CMD exec php wp post delete 3 --force --path='$wp_dir' --allow-root" 2>/dev/null || true
    
    # Delete default plugins
    eval "$CONTAINER_CMD exec php wp plugin delete akismet hello --path='$wp_dir' --allow-root" 2>/dev/null || true
}

configure_preset_bare() {
    local wp_dir="$1"
    
    print_status "info" "Configuring bare preset..."
    
    # Install minimal plugins
    local plugins=("all-in-one-wp-migration")
    
    install_github_plugins "$wp_dir" "${plugins[@]}"
    
    # Set up basic directory structure
    setup_basic_directories "$wp_dir"
}

configure_preset_clean() {
    local wp_dir="$1"
    
    print_status "info" "Configuring clean preset..."
    
    # Install clean preset plugins
    local plugins=(
        "typerocket-pro-v6"
        "all-in-one-wp-migration"
        "admin-site-enhancements-pro"
        "makermaker"
        "makerblocks"
    )
    
    install_github_plugins "$wp_dir" "${plugins[@]}"
    
    # Install clean preset themes
    local themes=("makerstarter")
    install_github_themes "$wp_dir" "${themes[@]}"
    
    # Delete default themes
    delete_default_themes "$wp_dir"
    
    # Set up directories and Galaxy files
    setup_basic_directories "$wp_dir"
    setup_galaxy_files "$wp_dir" "makermaker"
}

configure_preset_custom() {
    local wp_dir="$1"
    
    print_status "info" "Configuring custom preset..."
    
    # Install custom preset plugins
    local plugins=(
        "typerocket-pro-v6"
        "all-in-one-wp-migration"
        "admin-site-enhancements-pro"
        "advanced-custom-fields-pro"
        "acf-extended-pro"
        "makermaker"
        "makerblocks"
        "gravityforms"
        "manual-image-crop"
    )
    
    install_github_plugins "$wp_dir" "${plugins[@]}"
    
    # Install custom preset themes
    local themes=("makerstarter")
    install_github_themes "$wp_dir" "${themes[@]}"
    
    # Delete default themes
    delete_default_themes "$wp_dir"
    
    # Set up directories and Galaxy files
    setup_basic_directories "$wp_dir"
    setup_galaxy_files "$wp_dir" "typerocket-galaxy"
    setup_typerocket_integration "$wp_dir"
}

configure_preset_loaded() {
    local wp_dir="$1"
    
    print_status "info" "Configuring loaded preset..."
    
    # Install loaded preset plugins
    local plugins=(
        "typerocket-pro-v6"
        "all-in-one-wp-migration"
        "admin-site-enhancements-pro"
        "advanced-custom-fields-pro"
        "acf-extended-pro"
        "makermaker"
        "makerblocks"
        "gravityforms"
        "manual-image-crop"
    )
    
    install_github_plugins "$wp_dir" "${plugins[@]}"
    
    # Install WordPress repository plugins for development
    local wp_repo_plugins=(
        "debug-bar"
        "debug-bar-actions-and-filters-addon"
        "classic-editor"
        "default-featured-image"
        "plugin-inspector"
        "log-deprecated-notices"
        "query-monitor"
        "theme-check"
        "wordpress-beta-tester"
        "show-current-template"
        "theme-inspector"
        "view-admin-as"
    )
    
    print_status "info" "Installing WordPress repository plugins..."
    for plugin in "${wp_repo_plugins[@]}"; do
        eval "$CONTAINER_CMD exec php wp plugin install $plugin --path='$wp_dir' --allow-root" || true
    done
    
    # Install themes
    local themes=("makerstarter")
    install_github_themes "$wp_dir" "${themes[@]}"
    
    # Delete default themes
    delete_default_themes "$wp_dir"
    
    # Set up directories and Galaxy files
    setup_basic_directories "$wp_dir"
    setup_galaxy_files "$wp_dir" "makermaker"
}

configure_preset_starred() {
    local wp_dir="$1"
    
    print_status "info" "Configuring starred preset..."
    
    # Install starred repository plugins
    local plugins=(
        "gravityforms"
        "advanced-custom-fields-pro"
        "woocommerce-subscriptions"
        "facetwp"
    )
    
    install_github_plugins "$wp_dir" "${plugins[@]}"
    
    # Set up basic directories
    setup_basic_directories "$wp_dir"
}

# =============================================================================
# PLUGIN AND THEME INSTALLATION HELPERS
# =============================================================================

install_github_plugins() {
    local wp_dir="$1"
    shift
    local plugins=("$@")
    
    if [[ ${#plugins[@]} -eq 0 ]]; then
        return 0
    fi
    
    print_status "info" "Installing GitHub plugins..."
    
    # Change to plugins directory
    local plugins_dir="$wp_dir/wp-content/plugins"
    
    for plugin_name in "${plugins[@]}"; do
        print_status "info" "Installing plugin: $plugin_name"
        
        local plugin_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${plugin_name}.git"
        
        # Clone plugin repository
        if eval "$CONTAINER_CMD exec php git clone '$plugin_repo' '$plugins_dir/$plugin_name'"; then
            # Activate plugin
            if eval "$CONTAINER_CMD exec php wp plugin activate '$plugin_name' --path='$wp_dir' --allow-root"; then
                print_status "success" "Plugin $plugin_name installed and activated"
            else
                print_status "warning" "Plugin $plugin_name installed but failed to activate"
            fi
        else
            print_status "warning" "Failed to install plugin: $plugin_name"
        fi
    done
}

install_github_themes() {
    local wp_dir="$1"
    shift
    local themes=("$@")
    
    if [[ ${#themes[@]} -eq 0 ]]; then
        return 0
    fi
    
    print_status "info" "Installing GitHub themes..."
    
    # Change to themes directory
    local themes_dir="$wp_dir/wp-content/themes"
    
    for theme_name in "${themes[@]}"; do
        print_status "info" "Installing theme: $theme_name"
        
        local theme_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${theme_name}.git"
        
        # Clone theme repository
        if eval "$CONTAINER_CMD exec php git clone '$theme_repo' '$themes_dir/$theme_name'"; then
            # Activate theme
            if eval "$CONTAINER_CMD exec php wp theme activate '$theme_name' --path='$wp_dir' --allow-root"; then
                print_status "success" "Theme $theme_name installed and activated"
            else
                print_status "warning" "Theme $theme_name installed but failed to activate"
            fi
        else
            print_status "warning" "Failed to install theme: $theme_name"
        fi
    done
}

delete_default_themes() {
    local wp_dir="$1"
    
    print_status "info" "Removing default themes..."
    
    local default_themes=("twentytwentythree" "twentytwentyfour" "twentytwentyfive")
    
    for theme in "${default_themes[@]}"; do
        eval "$CONTAINER_CMD exec php wp theme delete '$theme' --path='$wp_dir' --allow-root" 2>/dev/null || true
    done
}

# =============================================================================
# DIRECTORY AND FILE SETUP HELPERS
# =============================================================================

setup_basic_directories() {
    local wp_dir="$1"
    
    print_status "info" "Setting up directory structure..."
    
    # Create and set permissions for AI1WM directories
    local ai1wm_backups="$wp_dir/wp-content/ai1wm-backups"
    local ai1wm_storage="$wp_dir/wp-content/plugins/all-in-one-wp-migration/storage"
    
    eval "$CONTAINER_CMD exec php mkdir -p '$ai1wm_backups'" || true
    eval "$CONTAINER_CMD exec php mkdir -p '$ai1wm_storage'" || true
    
    # Create uploads directory
    local uploads_dir="$wp_dir/wp-content/uploads"
    eval "$CONTAINER_CMD exec php mkdir -p '$uploads_dir'" || true
    
    if [[ "$opt_skip_permissions" == "false" ]]; then
        print_status "info" "Setting directory permissions..."
        
        # Set permissions using config.sh container commands
        eval "$CONTAINER_CMD exec php chmod -R 777 '$ai1wm_backups'" || true
        eval "$CONTAINER_CMD exec php chmod -R 777 '$ai1wm_storage'" || true
        eval "$CONTAINER_CMD exec php chmod -R 777 '$uploads_dir'" || true
        eval "$CONTAINER_CMD exec php chmod -R 777 '$wp_dir/wp-content/themes'" || true
        eval "$CONTAINER_CMD exec php chmod -R 777 '$wp_dir/wp-content/plugins'" || true
    fi
}

setup_galaxy_files() {
    local wp_dir="$1"
    local galaxy_type="$2"
    
    print_status "info" "Setting up Galaxy files..."
    
    case "$galaxy_type" in
        "makermaker")
            local galaxy_dir="$wp_dir/wp-content/plugins/makermaker/galaxy"
            local galaxy_file="$galaxy_dir/galaxy_makermaker"
            local galaxy_config="$galaxy_dir/galaxy-makermaker-config.php"
            
            # Copy Galaxy files
            eval "$CONTAINER_CMD exec php cp '$galaxy_file' '$wp_dir/'" || true
            
            # Update config with site name
            eval "$CONTAINER_CMD exec php sed \"s/\\$sitename = 'playground'/\\$sitename = '$SITE_NAME'/\" '$galaxy_config' > '$galaxy_config.tmp'" || true
            eval "$CONTAINER_CMD exec php mv '$galaxy_config.tmp' '$galaxy_config'" || true
            eval "$CONTAINER_CMD exec php cp '$galaxy_config' '$wp_dir/'" || true
            ;;
        "typerocket-galaxy")
            local typerocket_dir="$wp_dir/wp-content/plugins/typerocket-pro-v6/typerocket"
            local galaxy_file="$typerocket_dir/galaxy"
            local galaxy_config="$wp_dir/wp-content/plugins/makermaker/galaxy/galaxy-config.php"
            
            # Copy TypeRocket Galaxy files
            eval "$CONTAINER_CMD exec php cp '$galaxy_file' '$wp_dir/'" || true
            eval "$CONTAINER_CMD exec php cp '$galaxy_config' '$wp_dir/'" || true
            
            # Also copy Makermaker Galaxy files
            setup_galaxy_files "$wp_dir" "makermaker"
            ;;
    esac
}

setup_typerocket_integration() {
    local wp_dir="$1"
    
    print_status "info" "Setting up TypeRocket integration..."
    
    local typerocket_dir="$wp_dir/wp-content/plugins/typerocket-pro-v6/typerocket"
    local makerstarter_dir="$wp_dir/wp-content/themes/makerstarter"
    
    # Copy TypeRocket override folders to Maker Starter theme
    if [[ -d "$typerocket_dir" && -d "$makerstarter_dir" ]]; then
        eval "$CONTAINER_CMD exec php cp -R '$typerocket_dir/app' '$makerstarter_dir/'" || true
        eval "$CONTAINER_CMD exec php cp -R '$typerocket_dir/config' '$makerstarter_dir/'" || true
        eval "$CONTAINER_CMD exec php cp -R '$typerocket_dir/resources' '$makerstarter_dir/'" || true
        eval "$CONTAINER_CMD exec php cp -R '$typerocket_dir/routes' '$makerstarter_dir/'" || true
        eval "$CONTAINER_CMD exec php mkdir -p '$makerstarter_dir/storage'" || true
    fi
}

# =============================================================================
# FINALIZATION FUNCTIONS
# =============================================================================

finalize_installation() {
    local wp_dir="$APPS_DIR/$SITE_NAME/public"
    
    print_status "step" "Finalizing WordPress installation..."
    
    # Set permalink structure
    print_status "info" "Setting permalink structure..."
    eval "$CONTAINER_CMD exec php wp rewrite structure '%postname%' --path='$wp_dir' --allow-root"
    
    # Set final permissions if not skipped
    if [[ "$opt_skip_permissions" == "false" ]]; then
        print_status "info" "Setting final file permissions..."
        eval "$CONTAINER_CMD exec php chown -R www-data:www-data /var/www/html/"
        eval "$CONTAINER_CMD exec php chmod -R 777 /var/www/html/"
    fi
    
    # Send emails if not skipped
    if [[ "$opt_skip_email" == "false" ]]; then
        send_installation_emails "$wp_dir"
    fi
    
    print_status "success" "WordPress installation finalized"
}

send_installation_emails() {
    local wp_dir="$1"
    
    print_status "info" "Sending installation emails..."
    
    # Installation summary email
    local email_subject="WordPress Installation Summary"
    local email_body="$SITE_TITLE installed successfully at $APPS_DIR/$SITE_NAME and accessible at https://$opt_custom_domain"
    
    eval "$CONTAINER_CMD exec php wp eval \"wp_mail('$ADMIN_EMAIL', '$email_subject', '$email_body');\" --path='$wp_dir' --allow-root" || true
    
    # Test email
    local test_subject="Test Email from WordPress Setup Script"
    local test_body="This is a test email to verify that the mail configuration is working correctly.\\n\\nThanks,\\nThe Setup Script"
    
    eval "$CONTAINER_CMD exec php wp eval \"wp_mail('$ADMIN_EMAIL', '$test_subject', '$test_body');\" --path='$wp_dir' --allow-root" || true
    
    print_status "success" "Installation emails sent to $ADMIN_EMAIL"
}

# =============================================================================
# COMPLETION MESSAGE
# =============================================================================

show_completion_message() {
    echo ""
    echo "========================================================"
    print_status "success" "WordPress Installation Complete!"
    echo "========================================================"
    echo ""
    
    echo "Site Details:"
    echo "  Name: $SITE_NAME"
    echo "  Title: $SITE_TITLE"
    echo "  Preset: $opt_preset (${WORDPRESS_PRESETS[$opt_preset]})"
    echo "  Location: $APPS_DIR/$SITE_NAME"
    echo "  URL: https://$opt_custom_domain"
    echo ""
    
    echo "Database Information:"
    echo "  Database: $SITE_NAME"
    echo "  User: $MARIADB_USER"
    echo "  Password: $MARIADB_PASSWORD"
    echo "  Host: $MARIADB_HOST"
    echo ""
    
    echo "Admin Access:"
    echo "  Username: $ADMIN_USER"
    echo "  Password: $ADMIN_PASSWORD"
    echo "  Email: $ADMIN_EMAIL"
    echo "  Login: https://$opt_custom_domain/wp-admin"
    echo ""
    
    echo "Next Steps:"
    echo "  1. Add to hosts file (if not done already):"
    echo "     echo '127.0.0.1 $opt_custom_domain' | sudo tee -a /etc/hosts"
    echo ""
    echo "  2. Access your site:"
    echo "     https://$opt_custom_domain"
    echo ""
    
    echo "Management Commands:"
    echo "  View logs: ./scripts/service-manager.sh logs php"
    echo "  Restart services: ./scripts/service-manager.sh restart php mariadb"
    echo "  Database shell: $CONTAINER_CMD exec -it mariadb mysql -u root -p"
    echo "  PHP shell: $CONTAINER_CMD exec -it php zsh"
    echo ""
    
    echo "WordPress Commands:"
    echo "  WP-CLI: $CONTAINER_CMD exec php wp --path='$APPS_DIR/$SITE_NAME/public'"
    echo "  Update plugins: $CONTAINER_CMD exec php wp plugin update --all --path='$APPS_DIR/$SITE_NAME/public' --allow-root"
    echo "  Backup site: Use All-in-One WP Migration plugin"
    echo ""
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context (use config.sh defaults)
    setup_command_context "$DEFAULT_SUDO" "false"
    
    # Show installation summary
    print_status "info" "WordPress Installation Summary:"
    echo "  Site: $SITE_NAME ($SITE_TITLE)"
    echo "  Preset: $opt_preset"
    echo "  Domain: $opt_custom_domain"
    echo "  Location: $APPS_DIR/$SITE_NAME"
    echo ""
    
    # Validate environment and requirements
    validate_environment
    validate_site_setup
    
    # Install WordPress
    install_wordpress_core
    configure_wordpress_preset
    finalize_installation
    
    # Show completion message
    show_completion_message
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi