#!/bin/zsh

# =============================================================================
# FASTAPI INSTALLATION SCRIPT
# =============================================================================
# Standalone installer for FastAPI applications
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
    Creates a new FastAPI application with standardized structure.

ARGUMENTS:
    app-name            Name for the application (lowercase, no spaces)

OPTIONS:
    -f, --force         Overwrite existing installation
    -h, --help          Show this help message

EXAMPLES:
    $0 my-api                    # Create new FastAPI app
    $0 my-api --force           # Overwrite existing

REQUIREMENTS:
    - Python container must be running
    - Apps directory: $APPS_DIR

NOTES:
    - Apps are created in: $APPS_DIR/<app-name>
    - Public directory: $APPS_DIR/<app-name>/public (web root for static files)
    - Accessible at: https://<app-name>.test
    - API docs available at: /docs and /redoc
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

install_fastapi() {
    print_status "step" "Installing FastAPI application..."

    local app_path="${APPS_DIR}/${APP_NAME}"
    mkdir -p "$app_path"

    # Create requirements.txt
    cat > "${app_path}/requirements.txt" << EOF
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
python-multipart==0.0.6
python-dotenv==1.0.0
EOF

    # Create main.py
    cat > "${app_path}/main.py" << 'EOF'
from fastapi import FastAPI, HTTPException
from fastapi.responses import HTMLResponse, FileResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
from datetime import datetime
import os
import sys

# Create FastAPI app
app = FastAPI(
    title="FastAPI Application",
    description="A modern, fast web framework for building APIs",
    version="1.0.0",
)

# Mount static files from public directory
app.mount("/static", StaticFiles(directory="public"), name="static")


# Models
class HealthResponse(BaseModel):
    status: str
    timestamp: str
    uptime: float


class InfoResponse(BaseModel):
    name: str
    version: str
    python_version: str
    fastapi_version: str
    environment: str


# Routes
@app.get("/", response_class=HTMLResponse)
async def read_root():
    """Serve the main page"""
    try:
        with open("public/index.html", "r") as f:
            return f.read()
    except FileNotFoundError:
        return HTMLResponse("""
            <html>
                <body>
                    <h1>FastAPI Application</h1>
                    <p>Welcome! API documentation available at <a href="/docs">/docs</a></p>
                </body>
            </html>
        """)


@app.get("/api/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    import time
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "uptime": time.process_time()
    }


@app.get("/api/info", response_model=InfoResponse)
async def get_info():
    """Application information endpoint"""
    return {
        "name": "FastAPI Application",
        "version": "1.0.0",
        "python_version": sys.version,
        "fastapi_version": "0.104.1",
        "environment": os.getenv("ENVIRONMENT", "development")
    }


# Example POST endpoint
class Item(BaseModel):
    name: str
    description: str | None = None
    price: float
    tax: float | None = None


@app.post("/api/items")
async def create_item(item: Item):
    """Create a new item"""
    return {
        "message": "Item created successfully",
        "item": item.dict()
    }


@app.get("/api/items/{item_id}")
async def read_item(item_id: int):
    """Get an item by ID"""
    if item_id < 1:
        raise HTTPException(status_code=400, detail="Invalid item ID")

    return {
        "item_id": item_id,
        "name": f"Item {item_id}",
        "description": "This is a sample item"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
EOF

    # Create .env file
    cat > "${app_path}/.env" << EOF
ENVIRONMENT=development
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

    print_status "success" "FastAPI application installed"
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
    <title>FastAPI Application</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #009688 0%, #26a69a 100%);
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
            color: #009688;
            font-size: 1.2em;
            margin-bottom: 30px;
        }
        .info-box {
            background: #f7f7f7;
            border-left: 4px solid #009688;
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
            color: #009688;
            font-family: 'Courier New', monospace;
        }
        .links {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-top: 30px;
        }
        .link {
            text-align: center;
            padding: 12px;
            background: #009688;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background 0.3s;
        }
        .link:hover {
            background: #00796b;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚡ FastAPI</h1>
        <div class="subtitle">Modern, fast (high-performance) web framework!</div>

        <div class="info-box">
            <h2>Application Information</h2>
            <div class="info-item">
                <span class="info-label">Framework:</span>
                <span class="info-value">FastAPI 0.104</span>
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
            <h2>API Features</h2>
            <div class="info-item">
                <span class="info-label">Interactive Docs:</span>
                <span class="info-value">/docs (Swagger UI)</span>
            </div>
            <div class="info-item">
                <span class="info-label">Alternative Docs:</span>
                <span class="info-value">/redoc (ReDoc)</span>
            </div>
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
            <a href="/docs" class="link">Swagger UI</a>
            <a href="/redoc" class="link">ReDoc</a>
            <a href="/api/health" class="link">Health Check</a>
            <a href="/api/info" class="link">Server Info</a>
        </div>
    </div>
</body>
</html>
EOF

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

    install_fastapi
    ensure_public_directory

    print_status "success" "FastAPI app '$APP_NAME' created at ${APPS_DIR}/${APP_NAME}"
    echo ""
    print_status "info" "To start the server:"
    echo "  cd ${APPS_DIR}/${APP_NAME}"
    echo "  uvicorn main:app --host 0.0.0.0 --port 8000 --reload"
    echo ""
    print_status "info" "API Documentation:"
    echo "  Swagger UI: http://localhost:8000/docs"
    echo "  ReDoc:      http://localhost:8000/redoc"
}

# Only run if executed directly
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi
