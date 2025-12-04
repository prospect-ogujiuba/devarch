#!/bin/zsh

# =============================================================================
# EXPRESS INSTALLATION SCRIPT
# =============================================================================
# Standalone installer for Express.js applications
# Called by create-app.sh or can be run independently

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT CONFIGURATION
# =============================================================================

APP_NAME=""
OPT_FORCE=false

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 <app-name> [OPTIONS]

DESCRIPTION:
    Creates a new Express.js application with standardized structure.

ARGUMENTS:
    app-name            Name for the application (lowercase, no spaces)

OPTIONS:
    -f, --force         Overwrite existing installation
    -h, --help          Show this help message

EXAMPLES:
    $0 my-api                    # Create new Express app
    $0 my-api --force           # Overwrite existing

REQUIREMENTS:
    - Node container must be running
    - Apps directory: $APPS_DIR

NOTES:
    - Apps are created in: $APPS_DIR/<app-name>
    - Public directory: $APPS_DIR/<app-name>/public (web root)
    - Accessible at: https://<app-name>.test
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ $# -eq 0 ]]; then
        print_status "error" "No app name specified"
        show_usage
        exit 1
    fi

    APP_NAME="$1"
    shift

    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                OPT_FORCE=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1"
                ;;
        esac
    done
}

# =============================================================================
# VALIDATION
# =============================================================================

validate_environment() {
    print_status "step" "Validating environment..."

    # Check Node container
    if ! validate_service_exists "node"; then
        handle_error "Node service not found"
    fi

    local service_status=$(get_service_status "node")
    if [[ "$service_status" != "running" ]]; then
        print_status "info" "Starting Node container..."
        start_single_service "node" || handle_error "Failed to start Node container"
        sleep 3
    fi

    # Check apps directory
    mkdir -p "$APPS_DIR"

    print_status "success" "Environment validated"
}

# =============================================================================
# INSTALLATION
# =============================================================================

install_express() {
    print_status "step" "Installing Express.js application..."

    local app_path="${APPS_DIR}/${APP_NAME}"
    mkdir -p "$app_path"

    # Create package.json
    cat > "${app_path}/package.json" << EOF
{
  "name": "${APP_NAME}",
  "version": "1.0.0",
  "description": "Express.js application",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "nodemon server.js"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "devDependencies": {
    "nodemon": "^3.0.1"
  }
}
EOF

    # Create server.js
    cat > "${app_path}/server.js" << 'EOF'
const express = require('express');
const path = require('path');
const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Serve static files from public directory
app.use(express.static('public'));

// API routes
app.get('/api/health', (req, res) => {
    res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        uptime: process.uptime()
    });
});

app.get('/api/info', (req, res) => {
    res.json({
        name: 'Express Application',
        version: '1.0.0',
        node: process.version,
        platform: process.platform
    });
});

// Fallback route - serves index.html from public
app.get('*', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

// Start server
app.listen(PORT, '0.0.0.0', () => {
    console.log(`Server running on http://0.0.0.0:${PORT}`);
    console.log(`Environment: ${process.env.NODE_ENV || 'development'}`);
});
EOF

    # Install dependencies in container (run as root to avoid permission issues)
    print_status "info" "Installing npm dependencies..."
    $CONTAINER_CMD exec --user root -w "/app/${APP_NAME}" node npm install 2>&1 | grep -v "npm WARN" || true

    print_status "success" "Express.js application installed"
}

ensure_public_directory() {
    local app_path="${APPS_DIR}/${APP_NAME}"
    local public_dir="${app_path}/public"

    print_status "step" "Creating public directory..."
    mkdir -p "$public_dir"

    # Create index.html
    cat > "${public_dir}/index.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Express Application</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 600px;
            width: 100%;
            padding: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 2.5em;
        }
        .subtitle {
            color: #667eea;
            font-size: 1.2em;
            margin-bottom: 30px;
        }
        .info-box {
            background: #f7f7f7;
            border-left: 4px solid #667eea;
            padding: 20px;
            border-radius: 4px;
            margin-bottom: 20px;
        }
        .info-box h2 {
            color: #333;
            font-size: 1.2em;
            margin-bottom: 15px;
        }
        .info-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #e0e0e0;
        }
        .info-item:last-child {
            border-bottom: none;
        }
        .info-label {
            font-weight: 600;
            color: #555;
        }
        .info-value {
            color: #667eea;
            font-family: 'Courier New', monospace;
        }
        .links {
            display: flex;
            gap: 15px;
            margin-top: 30px;
        }
        .link {
            flex: 1;
            text-align: center;
            padding: 12px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background 0.3s;
        }
        .link:hover {
            background: #5568d3;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Express.js</h1>
        <div class="subtitle">Your application is running!</div>

        <div class="info-box">
            <h2>Server Information</h2>
            <div class="info-item">
                <span class="info-label">Framework:</span>
                <span class="info-value">Express.js</span>
            </div>
            <div class="info-item">
                <span class="info-label">Status:</span>
                <span class="info-value">Running</span>
            </div>
            <div class="info-item">
                <span class="info-label">Environment:</span>
                <span class="info-value">Development</span>
            </div>
        </div>

        <div class="info-box">
            <h2>API Endpoints</h2>
            <div class="info-item">
                <span class="info-label">Health Check:</span>
                <span class="info-value">/api/health</span>
            </div>
            <div class="info-item">
                <span class="info-label">Server Info:</span>
                <span class="info-value">/api/info</span>
            </div>
        </div>

        <div class="links">
            <a href="/api/health" class="link">Health Check</a>
            <a href="/api/info" class="link">Server Info</a>
        </div>
    </div>
</body>
</html>
EOF

    # Create .gitkeep for empty directories
    mkdir -p "${public_dir}/css"
    mkdir -p "${public_dir}/js"
    mkdir -p "${public_dir}/images"
    touch "${public_dir}/css/.gitkeep"
    touch "${public_dir}/js/.gitkeep"
    touch "${public_dir}/images/.gitkeep"

    print_status "success" "Public directory created"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    parse_arguments "$@"
    setup_command_context "$DEFAULT_SUDO" "false"

    validate_environment

    install_express
    ensure_public_directory

    print_status "success" "Express.js app '$APP_NAME' created at ${APPS_DIR}/${APP_NAME}"
    echo ""
    print_status "info" "To start the server:"
    echo "  cd ${APPS_DIR}/${APP_NAME}"
    echo "  npm start              # Production mode"
    echo "  npm run dev            # Development mode with nodemon"
}

# Only run if executed directly
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi
