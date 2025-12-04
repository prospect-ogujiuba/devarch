#!/bin/zsh

# =============================================================================
# FLASK INSTALLATION SCRIPT
# =============================================================================
# Standalone installer for Flask applications
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
    Creates a new Flask application with standardized structure.

ARGUMENTS:
    app-name            Name for the application (lowercase, no spaces)

OPTIONS:
    -f, --force         Overwrite existing installation
    -h, --help          Show this help message

EXAMPLES:
    $0 my-app                    # Create new Flask app
    $0 my-app --force           # Overwrite existing

REQUIREMENTS:
    - Python container must be running
    - Apps directory: $APPS_DIR

NOTES:
    - Apps are created in: $APPS_DIR/<app-name>
    - Public directory: $APPS_DIR/<app-name>/public (web root for static files)
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

    # Check Python container
    if ! validate_service_exists "python"; then
        handle_error "Python service not found"
    fi

    local service_status=$(get_service_status "python")
    if [[ "$service_status" != "running" ]]; then
        print_status "info" "Starting Python container..."
        start_single_service "python" || handle_error "Failed to start Python container"
        sleep 3
    fi

    # Check apps directory
    mkdir -p "$APPS_DIR"

    print_status "success" "Environment validated"
}

# =============================================================================
# INSTALLATION
# =============================================================================

install_flask() {
    print_status "step" "Installing Flask application..."

    local app_path="${APPS_DIR}/${APP_NAME}"
    mkdir -p "$app_path"

    # Create requirements.txt
    cat > "${app_path}/requirements.txt" << EOF
Flask==3.0.0
Werkzeug==3.0.1
python-dotenv==1.0.0
EOF

    # Create app.py
    cat > "${app_path}/app.py" << 'EOF'
from flask import Flask, render_template, jsonify, send_from_directory
import os
from datetime import datetime

app = Flask(__name__,
            static_folder='public',
            static_url_path='')

# Configuration
app.config['DEBUG'] = os.getenv('FLASK_DEBUG', 'True') == 'True'
app.config['ENV'] = os.getenv('FLASK_ENV', 'development')

@app.route('/')
def index():
    """Serve the main page"""
    return send_from_directory('public', 'index.html')

@app.route('/api/health')
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'timestamp': datetime.utcnow().isoformat(),
        'environment': app.config['ENV']
    })

@app.route('/api/info')
def info():
    """Application information endpoint"""
    import sys
    return jsonify({
        'name': 'Flask Application',
        'version': '1.0.0',
        'python_version': sys.version,
        'flask_version': '3.0.0',
        'environment': app.config['ENV']
    })

@app.errorhandler(404)
def not_found(error):
    """Handle 404 errors"""
    return jsonify({
        'error': 'Not Found',
        'message': 'The requested resource was not found'
    }), 404

@app.errorhandler(500)
def internal_error(error):
    """Handle 500 errors"""
    return jsonify({
        'error': 'Internal Server Error',
        'message': 'An internal error occurred'
    }), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000, debug=True)
EOF

    # Create .env file
    cat > "${app_path}/.env" << EOF
FLASK_APP=app.py
FLASK_ENV=development
FLASK_DEBUG=True
EOF

    # Create .gitignore
    cat > "${app_path}/.gitignore" << EOF
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
venv/
ENV/
.env
.venv
*.egg-info/
dist/
build/
EOF

    print_status "success" "Flask application installed"
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
    <title>Flask Application</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
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
            color: #4facfe;
            font-size: 1.2em;
            margin-bottom: 30px;
        }
        .info-box {
            background: #f7f7f7;
            border-left: 4px solid #4facfe;
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
            color: #4facfe;
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
            background: #4facfe;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background 0.3s;
        }
        .link:hover {
            background: #3d8cd4;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🐍 Flask</h1>
        <div class="subtitle">Your application is running!</div>

        <div class="info-box">
            <h2>Application Information</h2>
            <div class="info-item">
                <span class="info-label">Framework:</span>
                <span class="info-value">Flask 3.0</span>
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

    # Create templates directory (Flask convention)
    mkdir -p "${app_path}/templates"
    touch "${app_path}/templates/.gitkeep"

    # Create static asset directories
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

    install_flask
    ensure_public_directory

    print_status "success" "Flask app '$APP_NAME' created at ${APPS_DIR}/${APP_NAME}"
    echo ""
    print_status "info" "To start the server:"
    echo "  cd ${APPS_DIR}/${APP_NAME}"
    echo "  python app.py                    # Direct execution"
    echo "  flask run --host=0.0.0.0 --port=8000  # Using Flask CLI"
}

# Only run if executed directly
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi
