#!/bin/bash

# =============================================================================
# DEVARCH APP CREATOR
# =============================================================================
# Comprehensive app creation script with framework-aware boilerplate generation.
# Supports PHP (Laravel, WordPress, Generic), Node (Next.js, React, Express),
# Python (Django, FastAPI, Flask), and Go applications.
#
# Usage: ./create-app.sh [options]
# =============================================================================

# Configuration
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
APPS_DIR="${APPS_DIR:-$PROJECT_ROOT/apps}"

# Detect container runtime (don't source zsh config.sh in bash)
if command -v podman >/dev/null 2>&1; then
    CONTAINER_CMD="podman"
elif command -v docker >/dev/null 2>&1; then
    CONTAINER_CMD="docker"
else
    echo "[ERROR] Neither podman nor docker found!" >&2
    exit 1
fi

# Global variables for WordPress configuration
WORDPRESS_PRESET=""
WORDPRESS_TITLE=""

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

print_status() {
    local level="$1"
    local message="$2"

    case "$level" in
        info)    echo -e "\e[34m[INFO]\e[0m $message" ;;
        success) echo -e "\e[32m[SUCCESS]\e[0m $message" ;;
        warning) echo -e "\e[33m[WARNING]\e[0m $message" ;;
        error)   echo -e "\e[31m[ERROR]\e[0m $message" ;;
        step)    echo -e "\e[36m[STEP]\e[0m $message" ;;
        *)       echo "$message" ;;
    esac
}

print_header() {
    echo ""
    echo "==============================================="
    echo "  $1"
    echo "==============================================="
    echo ""
}

show_usage() {
    cat << EOF
DevArch App Creator

USAGE:
    ./create-app.sh [OPTIONS]

OPTIONS:
    --name NAME              Application name (lowercase, no spaces)
    --framework FRAMEWORK    Framework to use (e.g., laravel, nextjs, django)
    --preset PRESET          WordPress preset (bare|clean|custom|loaded|starred)
    --title TITLE            Site title (WordPress only)
    -h, --help              Show this help message

INTERACTIVE MODE:
    Run without arguments for interactive prompts:
    ./create-app.sh

EXAMPLES:
    # Interactive mode
    ./create-app.sh

    # Create Laravel app
    ./create-app.sh --name myblog --framework laravel

    # Create WordPress with preset
    ./create-app.sh --name mysite --framework wordpress --preset clean --title "My Site"

    # Create React app
    ./create-app.sh --name dashboard --framework react

    # Create Django app
    ./create-app.sh --name api --framework django

    # Create Express API
    ./create-app.sh --name api-server --framework express

AVAILABLE FRAMEWORKS:
    PHP:        laravel, wordpress, generic
    Node.js:    nextjs, react, vue, express
    Python:     django, flask, fastapi
    Go:         standard, gin, echo
    .NET:       webapi, mvc, blazor

For more information, see: README.md

EOF
}

# =============================================================================
# INTERACTIVE PROMPTS
# =============================================================================

prompt_app_name() {
    local app_name=""

    while true; do
        read -p "App name (lowercase, no spaces): " app_name

        # Validate app name
        if validate_app_name "$app_name"; then
            echo "$app_name"
            return 0
        fi
    done
}

prompt_framework() {
    local framework_spec=""

    echo "" >&2
    echo "Select framework:" >&2
    echo "" >&2
    echo "PHP Frameworks:" >&2
    echo "  1) Laravel" >&2
    echo "  2) WordPress" >&2
    echo "  3) Generic PHP" >&2
    echo "" >&2

    while true; do
        read -p "Framework [1-3]: " choice

        case "$choice" in
            1)  framework_spec="php:laravel"; break ;;
            2)  framework_spec="php:wordpress"; break ;;
            3)  framework_spec="php:generic"; break ;;
            *) print_status "error" "Invalid choice. Please enter 1-3." ;;
        esac
    done

    echo "$framework_spec"
}

prompt_wordpress_preset() {
    local preset=""

    echo "" >&2
    echo "Select WordPress preset:" >&2
    echo "  1) bare    - Minimal WordPress with basic plugins" >&2
    echo "  2) clean   - WordPress with TypeRocket Pro and premium plugins" >&2
    echo "  3) custom  - WordPress with ACF, Gravity Forms, and custom setup" >&2
    echo "  4) loaded  - WordPress with all development tools and debugging" >&2
    echo "  5) starred - WordPress with starred repository plugins only" >&2
    echo "" >&2

    while true; do
        read -p "Preset [1-5]: " choice
        case "$choice" in
            1) preset="bare"; break ;;
            2) preset="clean"; break ;;
            3) preset="custom"; break ;;
            4) preset="loaded"; break ;;
            5) preset="starred"; break ;;
            *) print_status "error" "Invalid choice. Please enter 1-5." ;;
        esac
    done

    echo "$preset"
}

prompt_site_title() {
    local app_name="$1"
    local default_title=$(echo "$app_name" | sed 's/[_-]/ /g' | sed 's/\b\w/\U&/g')

    echo "" >&2
    read -p "Site title (press Enter for '$default_title'): " site_title

    if [[ -z "$site_title" ]]; then
        echo "$default_title"
    else
        echo "$site_title"
    fi
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_app_name() {
    local name="$1"

    # Check if empty
    if [[ -z "$name" ]]; then
        print_status "error" "App name cannot be empty"
        return 1
    fi

    # Check for invalid characters (only lowercase, numbers, hyphens, underscores)
    if [[ ! "$name" =~ ^[a-z0-9_-]+$ ]]; then
        print_status "error" "App name can only contain lowercase letters, numbers, hyphens, and underscores"
        return 1
    fi

    # Check if starts with number
    if [[ "$name" =~ ^[0-9] ]]; then
        print_status "error" "App name cannot start with a number"
        return 1
    fi

    # Check length
    if [[ ${#name} -lt 2 ]]; then
        print_status "error" "App name must be at least 2 characters long"
        return 1
    fi

    if [[ ${#name} -gt 50 ]]; then
        print_status "error" "App name must be less than 50 characters"
        return 1
    fi

    return 0
}

check_app_exists() {
    local app_name="$1"

    if [[ -d "${APPS_DIR}/${app_name}" ]]; then
        print_status "error" "App '$app_name' already exists at ${APPS_DIR}/${app_name}"
        return 1
    fi

    return 0
}

check_backend_running() {
    local runtime="$1"
    local container_name=""

    case "$runtime" in
        php) container_name="php" ;;
    esac

    if $CONTAINER_CMD ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container_name}$"; then
        return 0
    else
        return 1
    fi
}

# =============================================================================
# APP CREATION FUNCTIONS
# =============================================================================

create_directory() {
    local app_name="$1"
    local app_path="${APPS_DIR}/${app_name}"

    print_status "step" "Creating app directory: $app_path"

    if mkdir -p "$app_path"; then
        print_status "success" "Directory created"
        return 0
    else
        print_status "error" "Failed to create directory"
        return 1
    fi
}

install_framework() {
    local app_name="$1"
    local runtime="$2"
    local framework="$3"

    print_status "step" "Installing $framework framework..."

    case "$runtime-$framework" in

        php-wordpress)
            "${SCRIPT_DIR}/wordpress/install-wordpress.sh" "$app_name" \
                --preset "$WORDPRESS_PRESET" --title "$WORDPRESS_TITLE" --force
            ;;
        php-laravel)
            install_php_framework "$app_name" "$framework"
            ;;
        php-generic)
            install_php_framework "$app_name" "$framework"
            ;;

        *)
            print_status "error" "Unknown framework: $runtime-$framework"
            return 1
            ;;
    esac
}

install_php_framework() {
    local app_name="$1"
    local framework="$2"

    case "$framework" in
        laravel)
            print_status "info" "Installing Laravel via Composer (this may take 30-60 seconds)..."
            if $CONTAINER_CMD exec -w /var/www/html php composer create-project --prefer-dist laravel/laravel "$app_name" 2>&1 | grep -v "Warning"; then
                print_status "success" "Laravel installed successfully"

                # Set permissions
                $CONTAINER_CMD exec php chown -R www-data:www-data "/var/www/html/${app_name}" 2>/dev/null || true
                $CONTAINER_CMD exec php chmod -R 775 "/var/www/html/${app_name}/storage" "/var/www/html/${app_name}/bootstrap/cache" 2>/dev/null || true

                return 0
            else
                print_status "error" "Failed to install Laravel"
                return 1
            fi
            ;;

        wordpress)
            print_status "info" "Installing WordPress with preset: $WORDPRESS_PRESET..."

            local install_script="${SCRIPT_DIR}/wordpress/install-wordpress.sh"

            if [[ ! -f "$install_script" ]]; then
                print_status "error" "WordPress installation script not found: $install_script"
                return 1
            fi

            # Build the command with collected options
            local wp_cmd="\"$install_script\" \"$app_name\" --preset \"$WORDPRESS_PRESET\" --force"

            if [[ -n "$WORDPRESS_TITLE" ]]; then
                wp_cmd="$wp_cmd --title \"$WORDPRESS_TITLE\""
            fi

            # Execute the installation script
            if eval "$wp_cmd"; then
                print_status "success" "WordPress installed successfully with $WORDPRESS_PRESET preset"
                return 0
            else
                print_status "error" "WordPress installation failed"
                return 1
            fi
            ;;

        generic)
            print_status "info" "Creating generic PHP application structure..."
            local app_path="${APPS_DIR}/${app_name}"

            # Create basic structure
            mkdir -p "${app_path}/public"

            # Create index.php
            cat > "${app_path}/public/index.php" << 'EOF'
<?php
/**
 * Generic PHP Application
 */

echo "<!DOCTYPE html>\n";
echo "<html lang=\"en\">\n";
echo "<head>\n";
echo "    <meta charset=\"UTF-8\">\n";
echo "    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n";
echo "    <title>PHP Application</title>\n";
echo "    <style>\n";
echo "        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }\n";
echo "        h1 { color: #333; }\n";
echo "        .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }\n";
echo "    </style>\n";
echo "</head>\n";
echo "<body>\n";
echo "    <h1>Welcome to Your PHP Application</h1>\n";
echo "    <div class=\"info\">\n";
echo "        <h2>Environment Information</h2>\n";
echo "        <p><strong>PHP Version:</strong> " . phpversion() . "</p>\n";
echo "        <p><strong>Server:</strong> " . $_SERVER['SERVER_SOFTWARE'] . "</p>\n";
echo "        <p><strong>Document Root:</strong> " . $_SERVER['DOCUMENT_ROOT'] . "</p>\n";
echo "    </div>\n";
echo "</body>\n";
echo "</html>\n";
EOF

            # Create composer.json
            cat > "${app_path}/composer.json" << EOF
{
    "name": "devarch/${app_name}",
    "description": "Generic PHP application",
    "type": "project",
    "require": {
        "php": "^8.0"
    },
    "autoload": {
        "psr-4": {
            "App\\\\": "src/"
        }
    }
}
EOF

            mkdir -p "${app_path}/src"

            print_status "success" "Generic PHP application created"
            return 0
            ;;
    esac
}


configure_app() {
    local app_name="$1"
    local runtime="$2"
    local framework="$3"

    print_status "step" "Configuring application..."

    # Runtime-specific configuration
    case "$runtime" in
        php)
            if [[ "$framework" == "laravel" ]]; then
                # Generate Laravel app key
                $CONTAINER_CMD exec -w "/var/www/html/${app_name}" php php artisan key:generate 2>/dev/null || true
            fi
            ;;
    esac

    print_status "success" "Configuration complete"
}

start_backend() {
    local runtime="$1"

    print_status "step" "Checking backend service..."

    if check_backend_running "$runtime"; then
        print_status "success" "Backend service '$runtime' is running"
    else
        print_status "warning" "Backend service '$runtime' is not running"

        # Try to start it
        if [[ -f "${SCRIPT_DIR}/service-manager.sh" ]]; then
            print_status "info" "Starting backend service..."
            if "${SCRIPT_DIR}/service-manager.sh" up "$runtime" 2>&1 | grep -q "started"; then
                print_status "success" "Backend service started"
            else
                print_status "warning" "Could not start backend service automatically"
                print_status "info" "Run: ./scripts/service-manager.sh up $runtime"
            fi
        else
            print_status "info" "Start backend manually: ./scripts/service-manager.sh up $runtime"
        fi
    fi
}

show_next_steps() {
    local app_name="$1"
    local runtime="$2"
    local framework="$3"

    echo ""
    print_header "Application Created Successfully!"

    echo "App Name:    $app_name"
    echo "Runtime:     $runtime"
    echo "Framework:   $framework"
    echo "Location:    ${APPS_DIR}/${app_name}"
    echo "URL:         http://${app_name}.test"
    echo ""

    echo "Next Steps:"
    echo ""
    echo "  1. Update /etc/hosts (if not done automatically):"
    echo "     sudo ./scripts/update-hosts.sh"
    echo "     e.g., sudo ./scripts/generate-context.sh && ./scripts/update-hosts.sh -o windows -f"
    echo ""
    echo "  2. Access your app:"
    echo "     http://${app_name}.test"
    echo ""
    echo "  3. View in dashboard:"
    echo "     http://dashboard.test"
    echo ""

    # Framework-specific instructions
    case "$runtime-$framework" in
        php-laravel)
            echo "Laravel-specific:"
            echo "  - Edit .env file in ${APPS_DIR}/${app_name}/.env"
            echo "  - Run migrations: podman exec -w /var/www/html/${app_name} php php artisan migrate"
            ;;
        php-wordpress)
            echo "WordPress-specific:"
            echo "  - WordPress is fully configured and ready to use"
            echo "  - Access site: http://${app_name}.test"
            echo "  - Admin login: http://${app_name}.test/wp-admin"
            echo "  - Preset: $WORDPRESS_PRESET"
            ;;
    esac

    echo ""
    print_status "success" "Done!"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

parse_args() {
    # Parse command-line arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --name)
                APP_NAME="$2"
                shift 2
                ;;
            --framework)
                # Map common framework names to runtime:framework format
                case "$2" in
                    laravel) FRAMEWORK_SPEC="php:laravel" ;;
                    wordpress) FRAMEWORK_SPEC="php:wordpress" ;;
                    generic) FRAMEWORK_SPEC="php:generic" ;;
                    *)
                        print_status "error" "Unknown framework: $2"
                        print_status "info" "Run with --help to see available frameworks"
                        exit 1
                        ;;
                esac
                RUNTIME=$(echo "$FRAMEWORK_SPEC" | cut -d: -f1)
                FRAMEWORK=$(echo "$FRAMEWORK_SPEC" | cut -d: -f2)
                shift 2
                ;;
            --preset)
                WORDPRESS_PRESET="$2"
                shift 2
                ;;
            --title)
                WORDPRESS_TITLE="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_status "error" "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

main() {
    # Parse command-line arguments if provided
    if [[ $# -gt 0 ]]; then
        parse_args "$@"
    fi

    print_header "DevArch App Creator"

    # STEP 1: Get app name
    if [[ -z "$APP_NAME" ]]; then
        APP_NAME=$(prompt_app_name)
    else
        print_status "info" "Using app name: $APP_NAME"
    fi

    # STEP 2: Select framework (single unified prompt)
    if [[ -z "$FRAMEWORK_SPEC" ]]; then
        FRAMEWORK_SPEC=$(prompt_framework)
        RUNTIME=$(echo "$FRAMEWORK_SPEC" | cut -d: -f1)
        FRAMEWORK=$(echo "$FRAMEWORK_SPEC" | cut -d: -f2)
    else
        print_status "info" "Using framework: $FRAMEWORK_SPEC"
    fi

    # STEP 3: Framework-specific prompts (WordPress only)
    if [[ "$RUNTIME" == "php" ]] && [[ "$FRAMEWORK" == "wordpress" ]]; then
        if [[ -z "$WORDPRESS_PRESET" ]]; then
            WORDPRESS_PRESET=$(prompt_wordpress_preset)
        fi
        if [[ -z "$WORDPRESS_TITLE" ]]; then
            WORDPRESS_TITLE=$(prompt_site_title "$APP_NAME")
        fi
    fi

    echo ""
    print_status "info" "Creating $FRAMEWORK application: $APP_NAME"
    echo ""

    # STEP 4: Validation
    if ! check_app_exists "$APP_NAME"; then
        exit 1
    fi

    # Check backend service
    if ! check_backend_running "$RUNTIME"; then
        print_status "warning" "Backend service '$RUNTIME' is not running"
        print_status "info" "The app will be created, but you'll need to start the backend service"
        echo ""
        read -p "Continue anyway? [y/N]: " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            print_status "info" "Cancelled"
            exit 0
        fi
    fi

    # STEP 5: Create directory and install framework
    if ! create_directory "$APP_NAME"; then
        exit 1
    fi

    if ! install_framework "$APP_NAME" "$RUNTIME" "$FRAMEWORK"; then
        print_status "error" "Framework installation failed"
        print_status "warning" "Directory created but framework not installed"
        exit 1
    fi

    # STEP 6: Configure app
    configure_app "$APP_NAME" "$RUNTIME" "$FRAMEWORK"

    # STEP 7: Start backend service
    start_backend "$RUNTIME"

    # STEP 8: Show next steps
    show_next_steps "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
