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

    echo ""
    echo "Select runtime:"
    echo "  1) PHP"
    echo "  2) Node.js"
    echo "  3) Python"
    echo "  4) Go"
    echo ""

    while true; do
        read -p "Runtime [1-4]: " choice

        case "$choice" in
            1) runtime="php"; break ;;
            2) runtime="node"; break ;;
            3) runtime="python"; break ;;
            4) runtime="go"; break ;;
            *) print_status "error" "Invalid choice. Please enter 1-4." ;;
        esac
    done

    echo "$runtime"
}

prompt_framework() {
    local runtime="$1"
    local framework=""

    echo ""
    case "$runtime" in
        php)
            echo "Select PHP framework:"
            echo "  1) Laravel"
            echo "  2) WordPress"
            echo "  3) Generic PHP"
            echo ""

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
            echo "Select Node.js framework:"
            echo "  1) Next.js"
            echo "  2) React (Vite)"
            echo "  3) Express"
            echo ""

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
            echo "Select Python framework:"
            echo "  1) Django"
            echo "  2) FastAPI"
            echo "  3) Flask"
            echo ""

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
            echo "Select Go framework:"
            echo "  1) Standard (net/http)"
            echo "  2) Gin"
            echo "  3) Echo"
            echo ""

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
    esac

    echo "$framework"
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
            print_status "info" "Downloading WordPress core..."
            if $CONTAINER_CMD exec -w /var/www/html php wp core download --path="$app_name" --allow-root 2>&1; then
                print_status "success" "WordPress downloaded successfully"

                # Set permissions
                $CONTAINER_CMD exec php chown -R www-data:www-data "/var/www/html/${app_name}" 2>/dev/null || true

                print_status "info" "Run 'wp core config' and 'wp core install' inside the container to complete setup"
                return 0
            else
                print_status "error" "Failed to download WordPress"
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

add_hosts_entry() {
    local app_name="$1"
    local domain="${app_name}.test"

    print_status "step" "Adding hosts entry for $domain..."

    if [[ -f "${SCRIPT_DIR}/update-hosts.sh" ]]; then
        # Use update-hosts.sh to add the entry
        if sudo "${SCRIPT_DIR}/update-hosts.sh" add "$app_name" 2>/dev/null; then
            print_status "success" "Hosts entry added"
        else
            print_status "warning" "Could not automatically add hosts entry"
            print_status "info" "Run: sudo ./scripts/update-hosts.sh"
        fi
    else
        print_status "warning" "update-hosts.sh not found"
        print_status "info" "Manually add to /etc/hosts: 127.0.0.1 $domain"
    fi
}

setup_proxy() {
    local app_name="$1"
    local runtime="$2"

    print_status "step" "Proxy host setup instructions..."

    if [[ -f "${SCRIPT_DIR}/setup-proxy-host.sh" ]]; then
        print_status "info" "Run: ./scripts/setup-proxy-host.sh $app_name"
    else
        print_status "warning" "setup-proxy-host.sh not found"

        # Show manual instructions
        local backend_host=""
        local backend_port=""

        case "$runtime" in
            php) backend_host="php"; backend_port="8000" ;;
            node) backend_host="node"; backend_port="3000" ;;
            python) backend_host="python"; backend_port="8000" ;;
            go) backend_host="go"; backend_port="8080" ;;
        esac

        echo ""
        echo "Configure NPM proxy host:"
        echo "  Domain: ${app_name}.test"
        echo "  Forward to: http://${backend_host}:${backend_port}"
    fi
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
    echo "  1. Configure Nginx Proxy Manager:"
    echo "     ./scripts/setup-proxy-host.sh $app_name"
    echo ""
    echo "  2. Update /etc/hosts (if not done automatically):"
    echo "     sudo ./scripts/update-hosts.sh"
    echo ""
    echo "  3. Access your app:"
    echo "     http://${app_name}.test"
    echo ""
    echo "  4. View in dashboard:"
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
            echo "  - Complete setup: http://${app_name}.test"
            echo "  - Or use WP-CLI in container to configure"
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
    esac

    echo ""
    print_status "success" "Done!"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    print_header "DevArch App Creator"

    # Interactive prompts
    APP_NAME=$(prompt_app_name)
    RUNTIME=$(prompt_runtime)
    FRAMEWORK=$(prompt_framework "$RUNTIME")

    echo ""
    print_status "info" "Creating $FRAMEWORK application: $APP_NAME"
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
    if ! create_directory "$APP_NAME"; then
        exit 1
    fi

    if ! install_framework "$APP_NAME" "$RUNTIME" "$FRAMEWORK"; then
        print_status "error" "Framework installation failed"
        print_status "warning" "Directory created but framework not installed"
        exit 1
    fi

    configure_app "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
    add_hosts_entry "$APP_NAME"
    setup_proxy "$APP_NAME" "$RUNTIME"
    start_backend "$RUNTIME"

    # Success message
    show_next_steps "$APP_NAME" "$RUNTIME" "$FRAMEWORK"
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
