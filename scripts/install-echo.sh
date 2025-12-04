#!/bin/zsh

# =============================================================================
# ECHO INSTALLATION SCRIPT
# =============================================================================
# Standalone installer for Echo framework applications
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
    Creates a new Go application using the Echo web framework.

ARGUMENTS:
    app-name            Name for the application (lowercase, no spaces)

OPTIONS:
    -f, --force         Overwrite existing installation
    -h, --help          Show this help message

EXAMPLES:
    $0 my-api                    # Create new Echo app
    $0 my-api --force           # Overwrite existing

REQUIREMENTS:
    - Go container must be running
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

    if ! validate_service_exists "go"; then
        handle_error "Go service not found"
    fi

    local service_status=$(get_service_status "go")
    if [[ "$service_status" != "running" ]]; then
        print_status "info" "Starting Go container..."
        start_single_service "go" || handle_error "Failed to start Go container"
        sleep 3
    fi

    mkdir -p "$APPS_DIR"
    print_status "success" "Environment validated"
}

# =============================================================================
# INSTALLATION
# =============================================================================

install_echo_app() {
    print_status "step" "Installing Echo application..."

    local app_path="${APPS_DIR}/${APP_NAME}"
    mkdir -p "$app_path"

    # Initialize Go module
    print_status "info" "Initializing Go module..."
    $CONTAINER_CMD exec --user root -w "/app/${APP_NAME}" go go mod init "$APP_NAME" 2>&1 || true

    # Create main.go
    cat > "${app_path}/main.go" << 'EOF'
package main

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var startTime = time.Now()

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files from public directory
	e.Static("/static", "public")
	e.File("/", "public/index.html")

	// API routes
	api := e.Group("/api")
	{
		api.GET("/health", healthCheck)
		api.GET("/info", getInfo)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    float64   `json:"uptime"`
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(startTime).Seconds(),
	})
}

type InfoResponse struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Framework  string `json:"framework"`
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	NumCPU     int    `json:"num_cpu"`
	NumRoutine int    `json:"num_goroutine"`
}

func getInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, InfoResponse{
		Name:       "Echo Application",
		Version:    "1.0.0",
		Framework:  "Echo v4",
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		NumCPU:     runtime.NumCPU(),
		NumRoutine: runtime.NumGoroutine(),
	})
}
EOF

    # Install Echo framework (v4.12.0 works with Go 1.22)
    print_status "info" "Installing Echo framework..."
    $CONTAINER_CMD exec --user root -w "/app/${APP_NAME}" go go get github.com/labstack/echo/v4@v4.12.0 2>&1 || true

    # Run go mod tidy
    print_status "info" "Running go mod tidy..."
    $CONTAINER_CMD exec --user root -w "/app/${APP_NAME}" go go mod tidy 2>&1 || true

    # Create .env file
    cat > "${app_path}/.env" << EOF
PORT=8080
EOF

    # Create .gitignore
    cat > "${app_path}/.gitignore" << EOF
*.exe
*.exe~
*.dll
*.so
*.dylib
${APP_NAME}
*.test
*.out
go.work
.env
EOF

    print_status "success" "Echo application installed"
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
    <title>Echo Application</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1292ee 0%, #46b4e7 100%);
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
        h1 { color: #333; margin-bottom: 10px; font-size: 2.5em; }
        .subtitle { color: #1292ee; font-size: 1.2em; margin-bottom: 30px; }
        .info-box {
            background: #f7f7f7;
            border-left: 4px solid #1292ee;
            padding: 20px;
            border-radius: 4px;
            margin-bottom: 20px;
        }
        .info-box h2 { color: #333; font-size: 1.2em; margin-bottom: 15px; }
        .info-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #e0e0e0;
        }
        .info-item:last-child { border-bottom: none; }
        .info-label { font-weight: 600; color: #555; }
        .info-value { color: #1292ee; font-family: 'Courier New', monospace; }
        .links { display: flex; gap: 15px; margin-top: 30px; }
        .link {
            flex: 1;
            text-align: center;
            padding: 12px;
            background: #1292ee;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: background 0.3s;
        }
        .link:hover { background: #0e75c0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📢 Echo</h1>
        <div class="subtitle">High performance, minimalist Go web framework</div>

        <div class="info-box">
            <h2>Application Information</h2>
            <div class="info-item">
                <span class="info-label">Framework:</span>
                <span class="info-value">Echo v4</span>
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
            <div class="info-item">
                <span class="info-label">Static Files:</span>
                <span class="info-value">/static/*</span>
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

    mkdir -p "${public_dir}/css" "${public_dir}/js" "${public_dir}/images"
    touch "${public_dir}/css/.gitkeep" "${public_dir}/js/.gitkeep" "${public_dir}/images/.gitkeep"

    print_status "success" "Public directory created"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    parse_arguments "$@"
    setup_command_context "$DEFAULT_SUDO" "false"

    validate_environment

    install_echo_app
    ensure_public_directory

    print_status "success" "Echo app '$APP_NAME' created at ${APPS_DIR}/${APP_NAME}"
    echo ""
    print_status "info" "To start the server:"
    echo "  cd ${APPS_DIR}/${APP_NAME}"
    echo "  go run main.go"
}

# Only run if executed directly
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi
