#!/bin/zsh

# =============================================================================
# WORDPRESS INSTALLATION SCRIPT - ENHANCED & UNIFIED
# =============================================================================
# Unified WordPress installer with centralized path and configuration management

# Source the central configuration
. "$(dirname "$0")/../config.sh"

# =============================================================================
# CENTRALIZED CONFIGURATION MANAGEMENT
# =============================================================================

# Site-specific configuration (set during argument parsing)
typeset -A SITE_CONFIG
SITE_CONFIG=(
    [name]=""
    [title]=""
    [domain]=""
    [preset]="bare"
    [force]=false
    [skip_email]=false
    [skip_permissions]=false
)

# Path configuration - centralized for easy management
typeset -A PATHS
PATHS=(
    # Host paths (outside containers)
    [host_apps]="$APPS_DIR"
    [host_site]=""  # Will be set to $APPS_DIR/$SITE_NAME
    [host_env]="$PROJECT_ROOT/.env"
    
    # Container paths (inside PHP container)
    [container_root]="/var/www/html"
    [container_site]=""  # Will be set to /var/www/html/$SITE_NAME
    [container_content]=""  # Will be set to /var/www/html/$SITE_NAME/wp-content
    [container_plugins]="" # Will be set to wp-content/plugins
    [container_mu_plugins]="" # Will be set to wp-content/mu-plugins
    [container_themes]=""  # Will be set to wp-content/themes
    [container_uploads]="" # Will be set to wp-content/uploads
)

# Database configuration
typeset -A DB_CONFIG
DB_CONFIG=(
    [host]="mariadb"
    [name]=""  # Will be set to site name
    [user]="root"
    [password]="admin1234567"
)

# WordPress admin configuration
typeset -A WP_ADMIN
WP_ADMIN=(
    [user]="$ADMIN_USER"
    [password]="$ADMIN_PASSWORD"
    [email]="$ADMIN_EMAIL"
)

# WordPress presets configuration
typeset -A WORDPRESS_PRESETS
WORDPRESS_PRESETS=(
    [bare]="Minimal WordPress with basic plugins"
    [clean]="WordPress with TypeRocket Pro and premium plugins"
    [custom]="WordPress with ACF, Gravity Forms, and custom setup"
    [loaded]="WordPress with all development tools and debugging plugins"
    [starred]="WordPress with starred repository plugins only"
)

# Plugin collections for each preset
typeset -A PRESET_PLUGINS
PRESET_PLUGINS=(
    [bare]="all-in-one-wp-migration"
    [clean]="all-in-one-wp-migration admin-site-enhancements-pro makermaker makerblocks"
    [custom]="all-in-one-wp-migration admin-site-enhancements-pro advanced-custom-fields-pro acf-extended-pro makermaker makerblocks gravityforms manual-image-crop"
    [loaded]="all-in-one-wp-migration admin-site-enhancements-pro advanced-custom-fields-pro acf-extended-pro makermaker makerblocks gravityforms manual-image-crop"
    [starred]="gravityforms advanced-custom-fields-pro woocommerce-subscriptions facetwp"
)

# Theme collections for each preset
typeset -A PRESET_THEMES
PRESET_THEMES=(
    [bare]=""
    [clean]="makerstarter"
    [custom]="makerstarter"
    [loaded]="makerstarter"
    [starred]=""
)

# WordPress repository plugins for loaded preset
WP_REPO_PLUGINS=(
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

# =============================================================================
# PATH MANAGEMENT FUNCTIONS
# =============================================================================

setup_site_paths() {
    local site_name="$1"
    
    # Set site-specific paths
    PATHS[host_site]="${PATHS[host_apps]}/$site_name"
    PATHS[container_site]="${PATHS[container_root]}/$site_name"
    PATHS[container_content]="${PATHS[container_site]}/wp-content"
    PATHS[container_plugins]="${PATHS[container_content]}/plugins"
    PATHS[container_mu_plugins]="${PATHS[container_content]}/mu-plugins"
    PATHS[container_themes]="${PATHS[container_content]}/themes"
    PATHS[container_uploads]="${PATHS[container_content]}/uploads"
    
    # Set database name
    DB_CONFIG[name]="$site_name"
}

get_path() {
    local path_key="$1"
    echo "${PATHS[$path_key]}"
}

get_db_config() {
    local config_key="$1"
    echo "${DB_CONFIG[$config_key]}"
}

get_wp_admin() {
    local admin_key="$1"
    echo "${WP_ADMIN[$admin_key]}"
}

# =============================================================================
# DATABASE HELPER FUNCTIONS
# =============================================================================

check_database_exists() {
    local db_name="$1"
    exec_php_wp "db check" 2>/dev/null
}

ensure_database_ready() {
    local db_name=$(get_db_config name)
    
    print_status "info" "Ensuring database is ready..."
    
    # First, try to connect to the database server
    if ! exec_php "mysql -h $(get_db_config host) -u $(get_db_config user) -p$(get_db_config password) -e 'SELECT 1;'" 2>/dev/null; then
        handle_error "Cannot connect to MariaDB server. Please check if the MariaDB container is running and credentials are correct."
    fi
    
    # Check if database exists and is accessible
    if check_database_exists "$db_name"; then
        print_status "success" "Database '$db_name' is ready"
        return 0
    fi
    
    # Try to create the database
    print_status "info" "Creating database '$db_name'..."
    if exec_php_wp "db create" 2>/dev/null; then
        print_status "success" "Database '$db_name' created successfully"
        return 0
    fi
    
    # If creation failed, check if it's because the database already exists
    local create_output=$(exec_php_wp "db create" 2>&1)
    if [[ "$create_output" == *"database exists"* ]] || [[ "$create_output" == *"ERROR 1007"* ]] || [[ "$create_output" == *"Can't create database"* ]]; then
        print_status "info" "Database '$db_name' already exists"
        
        # Verify we can actually use it
        if check_database_exists "$db_name"; then
            print_status "success" "Database '$db_name' is accessible"
            return 0
        else
            handle_error "Database '$db_name' exists but is not accessible. Please check database permissions."
        fi
    else
        handle_error "Failed to create database '$db_name': $create_output"
    fi
}

# =============================================================================
# CONTAINER COMMAND HELPERS
# =============================================================================

exec_php() {
    local command="$1"
    eval "$CONTAINER_CMD exec php $command"
}

exec_php_wp() {
    local wp_command="$1"
    local wp_path=$(get_path "container_site")
    exec_php "wp $wp_command --path='$wp_path' --allow-root"
}

exec_php_wp_quiet() {
    local wp_command="$1"
    local wp_path=$(get_path "container_site")
    exec_php "wp $wp_command --path='$wp_path' --allow-root" 2>/dev/null || true
}

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 <site-name> [OPTIONS]

DESCRIPTION:
    Unified WordPress installer with centralized configuration management.
    All paths and settings are managed in a single location for consistency.

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
    - Apps directory: $(get_path "host_apps")

NOTES:
    - Sites are created in: $(get_path "host_apps")/<site-name>
    - Database name matches site name
    - Uses environment variables from: ${PATHS[host_env]}
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
    
    local site_name="$1"
    shift
    
    # Validate site name
    if [[ ! "$site_name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        handle_error "Site name can only contain letters, numbers, hyphens, and underscores"
    fi
    
    # Set up site configuration
    SITE_CONFIG[name]="$site_name"
    SITE_CONFIG[title]=$(echo "$site_name" | sed 's/[_-]/ /g' | sed 's/\b\w/\U&/g')
    SITE_CONFIG[domain]="${site_name}.test"
    
    # Set up paths based on site name
    setup_site_paths "$site_name"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--title)
                if [[ -n "$2" && "$2" != -* ]]; then
                    SITE_CONFIG[title]="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a site title"
                fi
                ;;
            -p|--preset)
                if [[ -n "$2" && "$2" != -* ]]; then
                    if [[ -n "${WORDPRESS_PRESETS[$2]}" ]]; then
                        SITE_CONFIG[preset]="$2"
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
                    SITE_CONFIG[domain]="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a domain name"
                fi
                ;;
            -f|--force)
                SITE_CONFIG[force]=true
                shift
                ;;
            --skip-email)
                SITE_CONFIG[skip_email]=true
                shift
                ;;
            --skip-permissions)
                SITE_CONFIG[skip_permissions]=true
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
    local host_apps=$(get_path "host_apps")
    if [[ ! -d "$host_apps" ]]; then
        print_status "step" "Creating apps directory: $host_apps"
        mkdir -p "$host_apps"
    fi
    
    # Check .env file exists and source it
    local env_file="${PATHS[host_env]}"
    if [[ ! -f "$env_file" ]]; then
        handle_error "Environment file not found: $env_file"
    fi
    
    # Source environment variables
    set -a
    source "$env_file"
    set +a
    
    print_status "success" "Environment validation passed"
}

validate_site_setup() {
    local site_path=$(get_path "host_site")
    
    print_status "step" "Validating site setup..."
    
    # Check if site already exists
    if [[ -d "$site_path" ]]; then
        if [[ "${SITE_CONFIG[force]}" == "true" ]]; then
            print_status "warning" "Site '${SITE_CONFIG[name]}' exists, removing due to --force flag"
            rm -rf "$site_path"
        else
            handle_error "Site '${SITE_CONFIG[name]}' already exists. Use --force to overwrite."
        fi
    fi
    
    print_status "success" "Site validation passed"
}

# =============================================================================
# WORDPRESS CORE INSTALLATION
# =============================================================================

install_wordpress_core() {
    local container_site=$(get_path "container_site")
    local container_site=$(get_path "container_site")
    
    print_status "step" "Installing WordPress core..."
    
    # Create directory structure
    if ! exec_php "mkdir -p $container_site"; then
        handle_error "Failed to create WordPress directory"
    fi
    
    # Download WordPress
    print_status "info" "Downloading WordPress..."
    if ! exec_php "wp core download --path='$container_site' --allow-root"; then
        handle_error "Failed to download WordPress"
    fi
    
    # Create wp-config.php
    print_status "info" "Creating wp-config.php..."
    local db_config_cmd="config create --dbname='$(get_db_config name)' --dbuser='$(get_db_config user)' --dbpass='$(get_db_config password)' --dbhost='$(get_db_config host)'"
    
    if ! exec_php_wp "$db_config_cmd"; then
        handle_error "Failed to create wp-config.php"
    fi
    
    # Set debugging and other config
    exec_php_wp "config set WP_DEBUG false --raw"
    exec_php_wp "config set AUTOMATIC_UPDATER_DISABLED true --raw"
    
    # Ensure database is ready
    ensure_database_ready
    
    # Install WordPress (handle existing installation gracefully)
    print_status "info" "Installing WordPress..."
    local install_cmd="core install --url='${SITE_CONFIG[domain]}' --title='${SITE_CONFIG[title]}' --admin_user='$(get_wp_admin user)' --admin_password='$(get_wp_admin password)' --admin_email='$(get_wp_admin email)'"
    
    if exec_php_wp "$install_cmd" 2>/dev/null; then
        print_status "success" "WordPress installed successfully"
    else
        # Check if WordPress is already installed
        if exec_php_wp "core is-installed" 2>/dev/null; then
            print_status "info" "WordPress is already installed"
            
            # Update site URL and title if they're different
            local current_url=$(exec_php_wp "option get home" 2>/dev/null || echo "")
            local current_title=$(exec_php_wp "option get blogname" 2>/dev/null || echo "")
            
            if [[ "$current_url" != "https://${SITE_CONFIG[domain]}" ]]; then
                print_status "info" "Updating site URL to https://${SITE_CONFIG[domain]}"
                exec_php_wp "option update home 'https://${SITE_CONFIG[domain]}'" || true
                exec_php_wp "option update siteurl 'https://${SITE_CONFIG[domain]}'" || true
            fi
            
            if [[ "$current_title" != "${SITE_CONFIG[title]}" ]]; then
                print_status "info" "Updating site title to '${SITE_CONFIG[title]}'"
                exec_php_wp "option update blogname '${SITE_CONFIG[title]}'" || true
            fi
        else
            # Try installation again with more specific error handling
            local install_output=$(exec_php_wp "$install_cmd" 2>&1)
            handle_error "Failed to install WordPress: $install_output"
        fi
    fi
    
    # Configure uploads
    exec_php_wp "option update uploads_use_yearmonth_folders 0"
    
    print_status "success" "WordPress core installed successfully"
}

# =============================================================================
# PRESET-SPECIFIC CONFIGURATIONS
# =============================================================================

configure_wordpress_preset() {
    local preset="${SITE_CONFIG[preset]}"
    
    print_status "step" "Configuring WordPress preset: $preset"
    
    # Clean up default content
    cleanup_default_content

    # Install TypeRocket to mu-plugins
    install_typerocket_mu_plugin "$preset"
    
    # Install preset plugins
    install_preset_plugins "$preset"
    
    # Install preset themes
    install_preset_themes "$preset"
    
    # Apply preset-specific configuration
    case "$preset" in
        "clean")
            configure_preset_clean
            ;;
        "custom")
            configure_preset_custom
            ;;
        "loaded")
            configure_preset_loaded
            ;;
    esac
    
    print_status "success" "Preset configuration completed"
}

cleanup_default_content() {
    print_status "info" "Cleaning up default content..."
    
    # Delete default posts and pages
    exec_php_wp_quiet "post delete 1 --force"
    exec_php_wp_quiet "post delete 2 --force"
    exec_php_wp_quiet "post delete 3 --force"
    
    # Delete default plugins
    exec_php_wp_quiet "plugin delete akismet hello"
}

install_preset_plugins() {
    local preset="$1"
    local plugins_string="${PRESET_PLUGINS[$preset]}"
    
    if [[ -z "$plugins_string" ]]; then
        return 0
    fi
    
    # Convert string to array
    local plugins=(${(s: :)plugins_string})
    
    if [[ ${#plugins[@]} -gt 0 ]]; then
        install_github_plugins "${plugins[@]}"
    fi
}

install_preset_themes() {
    local preset="$1"
    local themes_string="${PRESET_THEMES[$preset]}"
    
    if [[ -z "$themes_string" ]]; then
        return 0
    fi
    
    # Convert string to array
    local themes=(${(s: :)themes_string})
    
    if [[ ${#themes[@]} -gt 0 ]]; then
        install_github_themes "${themes[@]}"
        delete_default_themes
    fi
}

configure_preset_clean() {
    print_status "info" "Configuring clean preset..."
    setup_basic_directories
    setup_galaxy_files "makermaker"
}

configure_preset_custom() {
    print_status "info" "Configuring custom preset..."
    setup_basic_directories
    setup_galaxy_files "typerocket-galaxy"
}

configure_preset_loaded() {
    print_status "info" "Configuring loaded preset..."
    
    # Install WordPress repository plugins
    print_status "info" "Installing WordPress repository plugins..."
    for plugin in "${WP_REPO_PLUGINS[@]}"; do
        exec_php_wp_quiet "plugin install $plugin"
    done
    
    setup_basic_directories
    setup_galaxy_files "makermaker"
}

# =============================================================================
# PLUGIN AND THEME INSTALLATION HELPERS
# =============================================================================

install_github_plugins() {
    local plugins=("$@")
    
    if [[ ${#plugins[@]} -eq 0 ]]; then
        return 0
    fi
    
    print_status "info" "Installing GitHub plugins..."
    
    local plugins_dir=$(get_path "container_plugins")
    
    for plugin_name in "${plugins[@]}"; do
        print_status "info" "Installing plugin: $plugin_name"
        
        local plugin_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${plugin_name}.git"
        
        # Clone plugin repository
        if exec_php "git clone '$plugin_repo' '$plugins_dir/$plugin_name'"; then
            
            # Run composer install for makermaker
            if [[ "$plugin_name" == "makermaker" ]]; then
                print_status "info" "Running composer install for makermaker plugin..."
                if exec_php "sh -c 'cd $plugins_dir/$plugin_name && composer install'"; then
                    print_status "success" "Composer dependencies installed for makermaker"
                else
                    print_status "warning" "Failed to install composer dependencies for makermaker"
                fi
            fi

            # Skip activation for all-in-one-wp-migration
            if [[ "$plugin_name" == "all-in-one-wp-migration" ]]; then
                print_status "success" "Plugin $plugin_name installed (not activated)"
            else
                # Activate plugin
                if exec_php_wp "plugin activate '$plugin_name'"; then
                    print_status "success" "Plugin $plugin_name installed and activated"
                else
                    print_status "warning" "Plugin $plugin_name installed but failed to activate"
                fi
            fi
        else
            print_status "warning" "Failed to install plugin: $plugin_name"
        fi
    done
}

install_github_themes() {
    local themes=("$@")
    
    if [[ ${#themes[@]} -eq 0 ]]; then
        return 0
    fi
    
    print_status "info" "Installing GitHub themes..."
    
    local themes_dir=$(get_path "container_themes")
    
    for theme_name in "${themes[@]}"; do
        print_status "info" "Installing theme: $theme_name"
        
        local theme_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/${theme_name}.git"
        
        # Clone theme repository
        if exec_php "git clone '$theme_repo' '$themes_dir/$theme_name'"; then
            # Activate theme
            if exec_php_wp "theme activate '$theme_name'"; then
                print_status "success" "Theme $theme_name installed and activated"
            else
                print_status "warning" "Theme $theme_name installed but failed to activate"
            fi
        else
            print_status "warning" "Failed to install theme: $theme_name"
        fi
    done
}

install_typerocket_mu_plugin() {
    local preset="$1"
    
    # Only install for presets that need TypeRocket
    if [[ "$preset" != "clean" && "$preset" != "custom" && "$preset" != "loaded" ]]; then
        return 0
    fi
    
    print_status "info" "Installing TypeRocket Pro to mu-plugins..."
    
    local mu_plugins_dir=$(get_path "container_mu_plugins")
    local typerocket_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/typerocket-pro-v6.git"
    
    # Create mu-plugins directory
    exec_php "mkdir -p '$mu_plugins_dir'" || true
    
    # Clone TypeRocket to mu-plugins
    if exec_php "git clone '$typerocket_repo' '$mu_plugins_dir/typerocket-pro-v6'"; then
        exec_php "mv '$mu_plugins_dir/typerocket-pro-v6/typerocket-pro-v6.php' '$mu_plugins_dir/'" || true
        print_status "success" "TypeRocket Pro installed to mu-plugins"
    else
        print_status "warning" "Failed to install TypeRocket Pro to mu-plugins"
    fi
}

delete_default_themes() {
    print_status "info" "Removing default themes..."
    
    local default_themes=("twentytwentythree" "twentytwentyfour" "twentytwentyfive")
    
    for theme in "${default_themes[@]}"; do
        exec_php_wp_quiet "theme delete '$theme'"
    done
}

# =============================================================================
# DIRECTORY AND FILE SETUP HELPERS
# =============================================================================

setup_basic_directories() {
    print_status "info" "Setting up directory structure..."
    
    local wp_content=$(get_path "container_content")
    local uploads=$(get_path "container_uploads")
    
    # Create and set permissions for AI1WM directories
    local ai1wm_backups="$wp_content/ai1wm-backups"
    local ai1wm_storage="$wp_content/plugins/all-in-one-wp-migration/storage"
    
    exec_php "mkdir -p '$ai1wm_backups'" || true
    exec_php "mkdir -p '$ai1wm_storage'" || true
    exec_php "mkdir -p '$uploads'" || true
    
    if [[ "${SITE_CONFIG[skip_permissions]}" == "false" ]]; then
        print_status "info" "Setting directory permissions..."
        
        exec_php "chmod -R 777 '$ai1wm_backups'" || true
        exec_php "chmod -R 777 '$ai1wm_storage'" || true
        exec_php "chmod -R 777 '$uploads'" || true
        exec_php "chmod -R 777 '$(get_path container_mu_plugins)'" || true
        exec_php "chmod -R 777 '$(get_path container_themes)'" || true
        exec_php "chmod -R 777 '$(get_path container_plugins)'" || true
    fi
}

setup_galaxy_files() {
    local galaxy_type="$1"
    local wp_dir=$(get_path "container_site")
    local site_name="${SITE_CONFIG[name]}"
    
    print_status "info" "Setting up Galaxy files..."
    
    case "$galaxy_type" in
        "makermaker")
            local galaxy_dir="$wp_dir/wp-content/plugins/makermaker/galaxy"
            local galaxy_file="$galaxy_dir/galaxy_makermaker"
            local galaxy_config="$galaxy_dir/galaxy-makermaker-config.php"
            
            exec_php "cp '$galaxy_file' '$wp_dir/'" || true
            
            # Fixed sed command with proper escaping
            exec_php "sed -i \"s/\\\$sitename = 'playground'/\\\$sitename = '$site_name'/g\" '$galaxy_config'" || true
            
            exec_php "cp '$galaxy_config' '$wp_dir/'" || true
            ;;
        "typerocket-galaxy")
            local typerocket_dir="$wp_dir/wp-content/mu-plugins/typerocket-pro-v6/typerocket"
            local galaxy_file="$typerocket_dir/galaxy"
            local galaxy_config="$wp_dir/wp-content/plugins/makermaker/galaxy/galaxy-config.php"
            
            exec_php "cp '$galaxy_file' '$wp_dir/'" || true
            exec_php "cp '$galaxy_config' '$wp_dir/'" || true
            
            # Also copy Makermaker Galaxy files
            setup_galaxy_files "makermaker"
            ;;
    esac
}

# =============================================================================
# FINALIZATION FUNCTIONS
# =============================================================================

finalize_installation() {
    print_status "step" "Finalizing WordPress installation..."
    
    # Set permalink structure
    print_status "info" "Setting permalink structure..."
    exec_php_wp "rewrite structure '%postname%'"
    
    # Set final permissions if not skipped
    if [[ "${SITE_CONFIG[skip_permissions]}" == "false" ]]; then
        print_status "info" "Setting final file permissions..."
        exec_php "chown -R www-data:www-data $(get_path container_site)"
        exec_php "chmod -R 777 $(get_path container_site)"
    fi
    
    # Send emails if not skipped
    if [[ "${SITE_CONFIG[skip_email]}" == "false" ]]; then
        send_installation_emails
    fi
    
    print_status "success" "WordPress installation finalized"
}

send_installation_emails() {
    print_status "info" "Sending installation emails..."
    
    local admin_email=$(get_wp_admin "email")
    local site_title="${SITE_CONFIG[title]}"
    local site_path=$(get_path "host_site")
    local domain="${SITE_CONFIG[domain]}"
    
    # Installation summary email
    local email_subject="WordPress Installation Summary"
    local email_body="$site_title installed successfully at $site_path and accessible at https://$domain"
    
    exec_php_wp_quiet "eval \"wp_mail('$admin_email', '$email_subject', '$email_body');\""
    
    # Test email
    local test_subject="Test Email from WordPress Setup Script"
    local test_body="This is a test email to verify that the mail configuration is working correctly.\\n\\nThanks,\\nThe Setup Script"
    
    exec_php_wp_quiet "eval \"wp_mail('$admin_email', '$test_subject', '$test_body');\""
    
    print_status "success" "Installation emails sent to $admin_email"
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
    echo "  Name: ${SITE_CONFIG[name]}"
    echo "  Title: ${SITE_CONFIG[title]}"
    echo "  Preset: ${SITE_CONFIG[preset]} (${WORDPRESS_PRESETS[${SITE_CONFIG[preset]}]})"
    echo "  Location: $(get_path host_site)"
    echo "  URL: https://${SITE_CONFIG[domain]}"
    echo ""
    
    echo "Database Information:"
    echo "  Database: $(get_db_config name)"
    echo "  User: $(get_db_config user)"
    echo "  Password: $(get_db_config password)"
    echo "  Host: $(get_db_config host)"
    echo ""
    
    echo "Admin Access:"
    echo "  Username: $(get_wp_admin user)"
    echo "  Password: $(get_wp_admin password)"
    echo "  Email: $(get_wp_admin email)"
    echo "  Login: https://${SITE_CONFIG[domain]}/wp-admin"
    echo ""
    
    echo "Next Steps:"
    echo "  1. Add to hosts file (if not done already):"
    echo "     echo '127.0.0.1 ${SITE_CONFIG[domain]}' | sudo tee -a /etc/hosts"
    echo ""
    echo "  2. Access your site:"
    echo "     https://${SITE_CONFIG[domain]}"
    echo ""
    
    echo "Management Commands:"
    echo "  View logs: ./scripts/service-manager.sh logs php"
    echo "  Restart services: ./scripts/service-manager.sh restart php mariadb"
    echo "  Database shell: $CONTAINER_CMD exec -it mariadb mysql -u root -p"
    echo "  PHP shell: $CONTAINER_CMD exec -it php zsh"
    echo ""
    
    echo "WordPress Commands:"
    echo "  WP-CLI: $CONTAINER_CMD exec php wp --path='$(get_path container_site)'"
    echo "  Update plugins: $CONTAINER_CMD exec php wp plugin update --all --path='$(get_path container_site)' --allow-root"
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
    echo "  Site: ${SITE_CONFIG[name]} (${SITE_CONFIG[title]})"
    echo "  Preset: ${SITE_CONFIG[preset]}"
    echo "  Domain: ${SITE_CONFIG[domain]}"
    echo "  Location: $(get_path host_site)"
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