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

# Source template library functions
TEMPLATE_LIB="${SCRIPT_DIR}/lib/app-templates.sh"
TEMPLATES_AVAILABLE=false
if [[ -f "$TEMPLATE_LIB" ]]; then
    source "$TEMPLATE_LIB"
    TEMPLATES_AVAILABLE=true
else
    print_status "warning" "Template library not found. Using legacy boilerplate generation."
fi

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

# Global variables for template configuration
USE_TEMPLATE=false
SELECTED_TEMPLATE=""

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
    --template TEMPLATE      Use specific template (e.g., node/react-vite)
    --list-templates         List all available templates
    -h, --help              Show this help message

INTERACTIVE MODE:
    Run without arguments for interactive prompts:
    ./create-app.sh

TEMPLATE MODE:
    Create app from template:
    ./create-app.sh --name my-app --template node/react-vite

LEGACY MODE:
    Create app without template (uses inline generation):
    ./create-app.sh --name my-app --template none

EXAMPLES:
    # Interactive mode
    ./create-app.sh

    # Create React app from template
    ./create-app.sh --name dashboard --template node/react-vite

    # Create Express API from template
    ./create-app.sh --name api-server --template node/express

    # Create Django app from template
    ./create-app.sh --name admin-panel --template python/django

    # List all templates
    ./create-app.sh --list-templates

AVAILABLE TEMPLATES:
    PHP:        php/laravel, php/wordpress, php/vanilla
    Node.js:    node/react-vite, node/nextjs, node/express, node/vue
    Python:     python/django, python/flask, python/fastapi
    Go:         go/gin, go/echo
    .NET:       dotnet/aspnet-core

For more information, see: templates/README.md

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

prompt_runtime() {
    local runtime=""

    echo "" >&2
    echo "Select runtime:" >&2
    echo "  1) PHP" >&2
    echo "  2) Node.js" >&2
    echo "  3) Python" >&2
    echo "  4) Go" >&2
    echo "  5) .NET" >&2
    echo "" >&2

    while true; do
        read -p "Runtime [1-5]: " choice

        case "$choice" in
            1) runtime="php"; break ;;
            2) runtime="node"; break ;;
            3) runtime="python"; break ;;
            4) runtime="go"; break ;;
            5) runtime="dotnet"; break ;;
            *) print_status "error" "Invalid choice. Please enter 1-5." ;;
        esac
    done

    echo "$runtime"
}

prompt_framework() {
    local runtime="$1"
    local framework=""

    echo "" >&2
    case "$runtime" in
        php)
            echo "Select PHP framework:" >&2
            echo "  1) Laravel" >&2
            echo "  2) WordPress" >&2
            echo "  3) Generic PHP" >&2
            echo "" >&2

            while true; do
                read -p "Framework [1-3]: " choice
                case "$choice" in
                    1) framework="laravel"; break ;;
                    2) framework="wordpress"; break ;;
                    3) framework="generic"; break ;;
                    *) print_status "error" "Invalid choice. Please enter 1-3." ;;
                esac
            done
            ;;

        node)
            echo "Select Node.js framework:" >&2
            echo "  1) Next.js" >&2
            echo "  2) React (Vite)" >&2
            echo "  3) Express" >&2
            echo "" >&2

            while true; do
                read -p "Framework [1-3]: " choice
                case "$choice" in
                    1) framework="nextjs"; break ;;
                    2) framework="react"; break ;;
                    3) framework="express"; break ;;
                    *) print_status "error" "Invalid choice. Please enter 1-3." ;;
                esac
            done
            ;;

        python)
            echo "Select Python framework:" >&2
            echo "  1) Django" >&2
            echo "  2) FastAPI" >&2
            echo "  3) Flask" >&2
            echo "" >&2

            while true; do
                read -p "Framework [1-3]: " choice
                case "$choice" in
                    1) framework="django"; break ;;
                    2) framework="fastapi"; break ;;
                    3) framework="flask"; break ;;
                    *) print_status "error" "Invalid choice. Please enter 1-3." ;;
                esac
            done
            ;;

        go)
            echo "Select Go framework:" >&2
            echo "  1) Standard (net/http)" >&2
            echo "  2) Gin" >&2
            echo "  3) Echo" >&2
            echo "" >&2

            while true; do
                read -p "Framework [1-3]: " choice
                case "$choice" in
                    1) framework="standard"; break ;;
                    2) framework="gin"; break ;;
                    3) framework="echo"; break ;;
                    *) print_status "error" "Invalid choice. Please enter 1-3." ;;
                esac
            done
            ;;

        dotnet)
            echo "Select .NET framework:" >&2
            echo "  1) ASP.NET Core Web API" >&2
            echo "  2) ASP.NET Core MVC" >&2
            echo "  3) Blazor Server" >&2
            echo "" >&2

            while true; do
                read -p "Framework [1-3]: " choice
                case "$choice" in
                    1) framework="webapi"; break ;;
                    2) framework="mvc"; break ;;
                    3) framework="blazor"; break ;;
                    *) print_status "error" "Invalid choice. Please enter 1-3." ;;
                esac
            done
            ;;
    esac

    echo "$framework"
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

prompt_template() {
    local runtime="$1"
    local framework="$2"
    local template_name=""

    # Don't offer templates if library is not available
    if [[ "$TEMPLATES_AVAILABLE" != "true" ]]; then
        echo ""
        return 0
    fi

    echo "" >&2
    echo "Use template? (Templates provide standardized structure with public/ directory)" >&2
    echo "" >&2

    # Show available templates for this runtime
    case "$runtime" in
        php)
            echo "Available PHP templates:" >&2
            echo "  1) php/laravel     - Laravel framework (recommended for Laravel)" >&2
            echo "  2) php/wordpress   - WordPress CMS (recommended for WordPress)" >&2
            echo "  3) php/vanilla     - Plain PHP application" >&2
            echo "  4) none            - Use legacy inline generation" >&2
            ;;
        node)
            echo "Available Node.js templates:" >&2
            echo "  1) node/react-vite - React + Vite SPA (recommended for React)" >&2
            echo "  2) node/nextjs     - Next.js with static export (recommended for Next.js)" >&2
            echo "  3) node/express    - Express.js server (recommended for Express)" >&2
            echo "  4) node/vue        - Vue.js SPA" >&2
            echo "  5) none            - Use legacy inline generation" >&2
            ;;
        python)
            echo "Available Python templates:" >&2
            echo "  1) python/django   - Django framework (recommended for Django)" >&2
            echo "  2) python/flask    - Flask framework (recommended for Flask)" >&2
            echo "  3) python/fastapi  - FastAPI framework (recommended for FastAPI)" >&2
            echo "  4) none            - Use legacy inline generation" >&2
            ;;
        go)
            echo "Available Go templates:" >&2
            echo "  1) go/gin          - Gin framework (recommended for Gin)" >&2
            echo "  2) go/echo         - Echo framework (recommended for Echo)" >&2
            echo "  3) none            - Use legacy inline generation" >&2
            ;;
        dotnet)
            echo "Available .NET templates:" >&2
            echo "  1) dotnet/aspnet-core - ASP.NET Core" >&2
            echo "  2) none               - Use legacy inline generation" >&2
            ;;
        *)
            echo "none"
            return 0
            ;;
    esac

    echo "" >&2
    read -p "Select template (or 'none' to skip): " choice

    case "$runtime" in
        php)
            case "$choice" in
                1) template_name="php/laravel" ;;
                2) template_name="php/wordpress" ;;
                3) template_name="php/vanilla" ;;
                4|none|n|N) template_name="none" ;;
                *)
                    print_status "warning" "Invalid choice, using legacy generation"
                    template_name="none"
                    ;;
            esac
            ;;
        node)
            case "$choice" in
                1) template_name="node/react-vite" ;;
                2) template_name="node/nextjs" ;;
                3) template_name="node/express" ;;
                4) template_name="node/vue" ;;
                5|none|n|N) template_name="none" ;;
                *)
                    print_status "warning" "Invalid choice, using legacy generation"
                    template_name="none"
                    ;;
            esac
            ;;
        python)
            case "$choice" in
                1) template_name="python/django" ;;
                2) template_name="python/flask" ;;
                3) template_name="python/fastapi" ;;
                4|none|n|N) template_name="none" ;;
                *)
                    print_status "warning" "Invalid choice, using legacy generation"
                    template_name="none"
                    ;;
            esac
            ;;
        go)
            case "$choice" in
                1) template_name="go/gin" ;;
                2) template_name="go/echo" ;;
                3|none|n|N) template_name="none" ;;
                *)
                    print_status "warning" "Invalid choice, using legacy generation"
                    template_name="none"
                    ;;
            esac
            ;;
        dotnet)
            case "$choice" in
                1) template_name="dotnet/aspnet-core" ;;
                2|none|n|N) template_name="none" ;;
                *)
                    print_status "warning" "Invalid choice, using legacy generation"
                    template_name="none"
                    ;;
            esac
            ;;
    esac

    echo "$template_name"
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
        node) container_name="node" ;;
        python) container_name="python" ;;
        go) container_name="go" ;;
        dotnet) container_name="dotnet" ;;
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

    case "$runtime" in
        php)
            install_php_framework "$app_name" "$framework"
            ;;
        node)
            install_node_framework "$app_name" "$framework"
            ;;
        python)
            install_python_framework "$app_name" "$framework"
            ;;
        go)
            install_go_framework "$app_name" "$framework"
            ;;
        dotnet)
            install_dotnet_framework "$app_name" "$framework"
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

install_node_framework() {
    local app_name="$1"
    local framework="$2"

    case "$framework" in
        nextjs)
            print_status "info" "Creating Next.js application (this may take 30-60 seconds)..."
            if $CONTAINER_CMD exec -w /var/www/html node npx -y create-next-app@latest "$app_name" --typescript --tailwind --eslint --app --src-dir --import-alias "@/*" --no-git 2>&1; then
                print_status "success" "Next.js application created"
                return 0
            else
                print_status "error" "Failed to create Next.js application"
                return 1
            fi
            ;;

        react)
            print_status "info" "Creating React application with Vite..."
            if $CONTAINER_CMD exec -w /var/www/html node npm create -y vite@latest "$app_name" -- --template react 2>&1; then
                print_status "success" "React application created"
                print_status "info" "Installing dependencies..."
                $CONTAINER_CMD exec -w "/var/www/html/${app_name}" node npm install 2>&1 | grep -v "npm WARN"
                return 0
            else
                print_status "error" "Failed to create React application"
                return 1
            fi
            ;;

        express)
            print_status "info" "Creating Express application..."
            local app_path="${APPS_DIR}/${app_name}"
            mkdir -p "$app_path"

            # Create package.json
            cat > "${app_path}/package.json" << EOF
{
  "name": "${app_name}",
  "version": "1.0.0",
  "description": "Express application",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "devDependencies": {
    "nodemon": "^3.0.1"
  }
}
EOF

            # Create index.js
            cat > "${app_path}/index.js" << 'EOF'
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/', (req, res) => {
    res.send(`
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Express App</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
                h1 { color: #333; }
                .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
            </style>
        </head>
        <body>
            <h1>Welcome to Your Express Application</h1>
            <div class="info">
                <h2>Server Information</h2>
                <p><strong>Node Version:</strong> ${process.version}</p>
                <p><strong>Platform:</strong> ${process.platform}</p>
                <p><strong>Port:</strong> ${PORT}</p>
            </div>
        </body>
        </html>
    `);
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(`Server running on http://0.0.0.0:${PORT}`);
});
EOF

            # Install dependencies in container
            print_status "info" "Installing dependencies..."
            $CONTAINER_CMD exec -w "/var/www/html/${app_name}" node npm install 2>&1 | grep -v "npm WARN"

            print_status "success" "Express application created"
            return 0
            ;;
    esac
}

install_python_framework() {
    local app_name="$1"
    local framework="$2"

    case "$framework" in
        django)
            print_status "info" "Creating Django project..."
            if $CONTAINER_CMD exec -w /var/www/html python django-admin startproject "$app_name" 2>&1; then
                print_status "success" "Django project created"

                # Update settings for development
                print_status "info" "Configuring Django for development..."
                $CONTAINER_CMD exec python bash -c "cd /var/www/html/${app_name} && python manage.py migrate" 2>&1 || true

                return 0
            else
                print_status "error" "Failed to create Django project"
                return 1
            fi
            ;;

        fastapi)
            print_status "info" "Creating FastAPI application..."
            local app_path="${APPS_DIR}/${app_name}"
            mkdir -p "$app_path"

            # Create requirements.txt
            cat > "${app_path}/requirements.txt" << EOF
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
EOF

            # Create main.py
            cat > "${app_path}/main.py" << 'EOF'
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
async def root():
    return """
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>FastAPI App</title>
        <style>
            body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
            h1 { color: #333; }
            .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
        </style>
    </head>
    <body>
        <h1>Welcome to Your FastAPI Application</h1>
        <div class="info">
            <h2>API Information</h2>
            <p><strong>Framework:</strong> FastAPI</p>
            <p><strong>Docs:</strong> <a href="/docs">/docs</a></p>
            <p><strong>ReDoc:</strong> <a href="/redoc">/redoc</a></p>
        </div>
    </body>
    </html>
    """

@app.get("/api/health")
async def health():
    return {"status": "healthy"}
EOF

            print_status "success" "FastAPI application created"
            print_status "info" "Run: uvicorn main:app --host 0.0.0.0 --port 8000 --reload"
            return 0
            ;;

        flask)
            print_status "info" "Creating Flask application..."
            local app_path="${APPS_DIR}/${app_name}"
            mkdir -p "$app_path"

            # Create requirements.txt
            cat > "${app_path}/requirements.txt" << EOF
Flask==3.0.0
Werkzeug==3.0.1
EOF

            # Create app.py
            cat > "${app_path}/app.py" << 'EOF'
from flask import Flask

app = Flask(__name__)

@app.route('/')
def index():
    return """
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Flask App</title>
        <style>
            body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
            h1 { color: #333; }
            .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
        </style>
    </head>
    <body>
        <h1>Welcome to Your Flask Application</h1>
        <div class="info">
            <h2>Application Information</h2>
            <p><strong>Framework:</strong> Flask</p>
            <p><strong>Status:</strong> Running</p>
        </div>
    </body>
    </html>
    """

@app.route('/api/health')
def health():
    return {"status": "healthy"}

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000, debug=True)
EOF

            print_status "success" "Flask application created"
            print_status "info" "Run: python app.py or flask run --host=0.0.0.0 --port=8000"
            return 0
            ;;
    esac
}

install_go_framework() {
    local app_name="$1"
    local framework="$2"
    local app_path="${APPS_DIR}/${app_name}"

    mkdir -p "$app_path"

    # Initialize go module
    print_status "info" "Initializing Go module..."
    $CONTAINER_CMD exec -w "/var/www/html/${app_name}" go go mod init "$app_name" 2>&1 || true

    case "$framework" in
        standard)
            print_status "info" "Creating standard Go application..."

            cat > "${app_path}/main.go" << 'EOF'
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/api/health", handleHealth)

    port := ":8080"
    fmt.Printf("Server starting on http://0.0.0.0%s\n", port)

    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal(err)
    }
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
    html := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Go App</title>
        <style>
            body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
            h1 { color: #333; }
            .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
        </style>
    </head>
    <body>
        <h1>Welcome to Your Go Application</h1>
        <div class="info">
            <h2>Application Information</h2>
            <p><strong>Framework:</strong> net/http (standard library)</p>
            <p><strong>Status:</strong> Running</p>
        </div>
    </body>
    </html>
    `
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, `{"status":"healthy"}`)
}
EOF

            print_status "success" "Standard Go application created"
            ;;

        gin)
            print_status "info" "Creating Gin application..."

            cat > "${app_path}/main.go" << 'EOF'
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Gin App</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
                h1 { color: #333; }
                .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
            </style>
        </head>
        <body>
            <h1>Welcome to Your Gin Application</h1>
            <div class="info">
                <h2>Application Information</h2>
                <p><strong>Framework:</strong> Gin</p>
                <p><strong>Status:</strong> Running</p>
            </div>
        </body>
        </html>
        `))
    })

    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    r.Run(":8080")
}
EOF

            # Install Gin
            print_status "info" "Installing Gin framework..."
            $CONTAINER_CMD exec -w "/var/www/html/${app_name}" go go get -u github.com/gin-gonic/gin 2>&1 || true
            $CONTAINER_CMD exec -w "/var/www/html/${app_name}" go go mod tidy 2>&1 || true

            print_status "success" "Gin application created"
            ;;

        echo)
            print_status "info" "Creating Echo application..."

            cat > "${app_path}/main.go" << 'EOF'
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    e.GET("/", func(c echo.Context) error {
        html := `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Echo App</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
                h1 { color: #333; }
                .info { background: #f0f0f0; padding: 15px; border-radius: 5px; }
            </style>
        </head>
        <body>
            <h1>Welcome to Your Echo Application</h1>
            <div class="info">
                <h2>Application Information</h2>
                <p><strong>Framework:</strong> Echo</p>
                <p><strong>Status:</strong> Running</p>
            </div>
        </body>
        </html>
        `
        return c.HTML(http.StatusOK, html)
    })

    e.GET("/api/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
    })

    e.Logger.Fatal(e.Start(":8080"))
}
EOF

            # Install Echo
            print_status "info" "Installing Echo framework..."
            $CONTAINER_CMD exec -w "/var/www/html/${app_name}" go go get -u github.com/labstack/echo/v4 2>&1 || true
            $CONTAINER_CMD exec -w "/var/www/html/${app_name}" go go mod tidy 2>&1 || true

            print_status "success" "Echo application created"
            ;;
    esac

    return 0
}

install_dotnet_framework() {
    local app_name="$1"
    local framework="$2"
    local app_path="${APPS_DIR}/${app_name}"

    mkdir -p "$app_path"

    case "$framework" in
        webapi)
            print_status "info" "Creating ASP.NET Core Web API..."
            if $CONTAINER_CMD exec -w /app dotnet dotnet new webapi -n "$app_name" -o "/app/$app_name" 2>&1; then
                print_status "success" "ASP.NET Core Web API created"
                return 0
            else
                print_status "error" "Failed to create Web API"
                return 1
            fi
            ;;

        mvc)
            print_status "info" "Creating ASP.NET Core MVC application..."
            if $CONTAINER_CMD exec -w /app dotnet dotnet new mvc -n "$app_name" -o "/app/$app_name" 2>&1; then
                print_status "success" "ASP.NET Core MVC application created"
                return 0
            else
                print_status "error" "Failed to create MVC application"
                return 1
            fi
            ;;

        blazor)
            print_status "info" "Creating Blazor Server application..."
            if $CONTAINER_CMD exec -w /app dotnet dotnet new blazorserver -n "$app_name" -o "/app/$app_name" 2>&1; then
                print_status "success" "Blazor Server application created"
                return 0
            else
                print_status "error" "Failed to create Blazor application"
                return 1
            fi
            ;;
    esac

    return 0
}

create_app_from_template() {
    local app_name="$1"
    local template="$2"
    local app_path="${APPS_DIR}/${app_name}"

    print_status "step" "Creating app from template: $template"

    # Validate template exists
    if ! validate_template "$template" 2>/dev/null; then
        print_status "error" "Template not found: $template"
        print_status "info" "Falling back to legacy generation"
        return 1
    fi

    # Copy template files
    if ! copy_template "$template" "$app_name" 2>/dev/null; then
        print_status "error" "Failed to copy template"
        return 1
    fi

    # Customize template with app-specific values
    print_status "step" "Customizing template for app: $app_name"
    customize_template "$app_name" 2>/dev/null || true

    # Ensure public directory exists
    ensure_public_directory "$app_name" 2>/dev/null || true

    # Verify template structure
    if verify_template_structure "$app_name" 2>/dev/null; then
        print_status "success" "Template applied successfully"
        return 0
    else
        print_status "warning" "Template structure has issues, but app was created"
        return 0
    fi
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
        node)
            # Node apps are configured during installation
            ;;
        python)
            # Python apps are configured during installation
            ;;
        go)
            # Go apps are configured during installation
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
            echo "  - Default credentials: admin / admin (change immediately!)"
            echo "  - Preset: $WORDPRESS_PRESET"
            ;;
        node-nextjs)
            echo "Next.js-specific:"
            echo "  - Start dev server: cd ${APPS_DIR}/${app_name} && npm run dev"
            echo "  - Build for production: npm run build && npm start"
            ;;
        node-react)
            echo "React-specific:"
            echo "  - Start dev server: cd ${APPS_DIR}/${app_name} && npm run dev"
            echo "  - Build for production: npm run build"
            ;;
        node-express)
            echo "Express-specific:"
            echo "  - Start server: cd ${APPS_DIR}/${app_name} && npm start"
            echo "  - Development mode: npm run dev"
            ;;
        python-django)
            echo "Django-specific:"
            echo "  - Run migrations: podman exec -w /var/www/html/${app_name} python python manage.py migrate"
            echo "  - Create superuser: podman exec -it -w /var/www/html/${app_name} python python manage.py createsuperuser"
            echo "  - Start server: python manage.py runserver 0.0.0.0:8000"
            ;;
        python-fastapi)
            echo "FastAPI-specific:"
            echo "  - Start server: uvicorn main:app --host 0.0.0.0 --port 8000 --reload"
            echo "  - API docs: http://${app_name}.test/docs"
            ;;
        python-flask)
            echo "Flask-specific:"
            echo "  - Start server: python app.py"
            echo "  - Or: flask run --host=0.0.0.0 --port=8000"
            ;;
        go-*)
            echo "Go-specific:"
            echo "  - Run app: cd ${APPS_DIR}/${app_name} && go run main.go"
            echo "  - Build binary: go build -o app main.go"
            ;;
        dotnet-webapi)
            echo ".NET Web API-specific:"
            echo "  - Run app: podman exec -w /app/${app_name} dotnet dotnet run"
            echo "  - API docs (Swagger): http://${app_name}.test/swagger"
            echo "  - Watch mode: dotnet watch run"
            ;;
        dotnet-mvc)
            echo ".NET MVC-specific:"
            echo "  - Run app: podman exec -w /app/${app_name} dotnet dotnet run"
            echo "  - Watch mode: dotnet watch run"
            echo "  - Scaffold: dotnet aspnet-codegenerator"
            ;;
        dotnet-blazor)
            echo "Blazor Server-specific:"
            echo "  - Run app: podman exec -w /app/${app_name} dotnet dotnet run"
            echo "  - Watch mode: dotnet watch run"
            echo "  - Hot reload enabled by default"
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
            --template)
                SELECTED_TEMPLATE="$2"
                USE_TEMPLATE=true
                shift 2
                ;;
            --list-templates)
                if [[ "$TEMPLATES_AVAILABLE" == "true" ]]; then
                    list_templates
                else
                    print_status "error" "Template library not available"
                fi
                exit 0
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

    # Interactive prompts (skip if already provided via CLI)
    if [[ -z "$APP_NAME" ]]; then
        APP_NAME=$(prompt_app_name)
    else
        print_status "info" "Using app name: $APP_NAME"
    fi

    # If template was specified via CLI, validate it and skip interactive prompts
    if [[ "$USE_TEMPLATE" == "true" ]] && [[ -n "$SELECTED_TEMPLATE" ]] && [[ "$SELECTED_TEMPLATE" != "none" ]]; then
        # Extract runtime from template name
        case "$SELECTED_TEMPLATE" in
            php/*) RUNTIME="php" ;;
            node/*) RUNTIME="node" ;;
            python/*) RUNTIME="python" ;;
            go/*) RUNTIME="go" ;;
            dotnet/*) RUNTIME="dotnet" ;;
            *)
                print_status "error" "Invalid template format: $SELECTED_TEMPLATE"
                print_status "info" "Template should be in format: runtime/framework (e.g., node/react-vite)"
                exit 1
                ;;
        esac

        # Extract framework from template
        FRAMEWORK=$(basename "$SELECTED_TEMPLATE")

        print_status "info" "Using template: $SELECTED_TEMPLATE"
        print_status "info" "Runtime: $RUNTIME, Framework: $FRAMEWORK"
    else
        # Interactive mode - prompt for runtime and framework
        RUNTIME=$(prompt_runtime)
        FRAMEWORK=$(prompt_framework "$RUNTIME")

        # Prompt for template selection (if templates available)
        if [[ "$TEMPLATES_AVAILABLE" == "true" ]]; then
            SELECTED_TEMPLATE=$(prompt_template "$RUNTIME" "$FRAMEWORK")
            if [[ -n "$SELECTED_TEMPLATE" ]] && [[ "$SELECTED_TEMPLATE" != "none" ]]; then
                USE_TEMPLATE=true
            fi
        fi
    fi

    # WordPress-specific prompts (only if not using template or using WordPress template)
    if [[ "$RUNTIME" == "php" ]] && [[ "$FRAMEWORK" == "wordpress" ]]; then
        if [[ "$USE_TEMPLATE" != "true" ]] || [[ "$SELECTED_TEMPLATE" == "php/wordpress" ]]; then
            WORDPRESS_PRESET=$(prompt_wordpress_preset)
            WORDPRESS_TITLE=$(prompt_site_title "$APP_NAME")
        fi
    fi

    echo ""
    if [[ "$USE_TEMPLATE" == "true" ]] && [[ "$SELECTED_TEMPLATE" != "none" ]]; then
        print_status "info" "Creating $FRAMEWORK application from template: $APP_NAME"
    else
        print_status "info" "Creating $FRAMEWORK application: $APP_NAME"
    fi
    echo ""

    # Validation
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

    # Creation steps
    # If using template, try template creation first
    if [[ "$USE_TEMPLATE" == "true" ]] && [[ "$SELECTED_TEMPLATE" != "none" ]]; then
        print_status "step" "Using template: $SELECTED_TEMPLATE"

        # Create directory first
        if ! create_directory "$APP_NAME"; then
            exit 1
        fi

        # Try to create from template
        if create_app_from_template "$APP_NAME" "$SELECTED_TEMPLATE"; then
            print_status "success" "App created from template successfully"

            # Skip legacy framework installation
            SKIP_FRAMEWORK_INSTALL=true
        else
            print_status "warning" "Template creation failed, falling back to legacy generation"
            USE_TEMPLATE=false
            SKIP_FRAMEWORK_INSTALL=false
        fi
    else
        # Legacy path - create directory
        if ! create_directory "$APP_NAME"; then
            exit 1
        fi
        SKIP_FRAMEWORK_INSTALL=false
    fi

    # Install framework using legacy method (if not already created from template)
    if [[ "$SKIP_FRAMEWORK_INSTALL" != "true" ]]; then
        if ! install_framework "$APP_NAME" "$RUNTIME" "$FRAMEWORK"; then
            print_status "error" "Framework installation failed"
            print_status "warning" "Directory created but framework not installed"
            exit 1
        fi
    fi

    # Configure app (always run this)
    configure_app "$APP_NAME" "$RUNTIME" "$FRAMEWORK"

    # Start backend service
    start_backend "$RUNTIME"

    # Success message
    if [[ "$USE_TEMPLATE" == "true" ]] && [[ "$SELECTED_TEMPLATE" != "none" ]]; then
        # Show template-specific next steps
        if [[ "$TEMPLATES_AVAILABLE" == "true" ]]; then
            local template_port=$(get_default_port "$SELECTED_TEMPLATE" 2>/dev/null || echo "8000")
            show_next_steps "$APP_NAME" "$SELECTED_TEMPLATE" "$template_port" 2>/dev/null || show_next_steps "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
        else
            show_next_steps "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
        fi
    else
        show_next_steps "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
    fi
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
