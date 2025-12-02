#!/bin/bash

# =============================================================================
# DEVARCH APP LISTER
# =============================================================================
# Lists all applications in the apps/ directory with runtime and framework
# detection. Supports table and JSON output formats with filtering.
#
# Usage: ./list-apps.sh [options]
# =============================================================================

# Configuration
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
APPS_DIR="${APPS_DIR:-$PROJECT_ROOT/apps}"
DETECT_SCRIPT="${SCRIPT_DIR}/detect-app-runtime.sh"

# Detect container runtime (don't source zsh config.sh in bash)
if command -v podman >/dev/null 2>&1; then
    CONTAINER_CMD="podman"
elif command -v docker >/dev/null 2>&1; then
    CONTAINER_CMD="docker"
else
    echo "[ERROR] Neither podman nor docker found!" >&2
    exit 1
fi

# Default options
OUTPUT_FORMAT="table"
FILTER_RUNTIME=""
SHOW_PATHS=false
VERBOSE=false

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

print_status() {
    local level="$1"
    local message="$2"

    case "$level" in
        info)    echo -e "\e[34m[INFO]\e[0m $message" >&2 ;;
        success) echo -e "\e[32m[SUCCESS]\e[0m $message" >&2 ;;
        warning) echo -e "\e[33m[WARNING]\e[0m $message" >&2 ;;
        error)   echo -e "\e[31m[ERROR]\e[0m $message" >&2 ;;
        *)       echo "$message" >&2 ;;
    esac
}

print_usage() {
    cat << EOF
Usage: $(basename "$0") [options]

Lists all applications in the apps/ directory with runtime detection.

Options:
    --json              Output in JSON format (for scripting)
    --runtime RUNTIME   Filter by runtime (php, node, python, go)
    --paths             Show full paths in table output
    -v, --verbose       Verbose output
    -h, --help          Show this help message

Examples:
    $(basename "$0")                    # List all apps in table format
    $(basename "$0") --json             # List all apps in JSON format
    $(basename "$0") --runtime php      # List only PHP apps
    $(basename "$0") --runtime node --json  # List Node apps as JSON
    $(basename "$0") --paths            # Show full paths

Output Formats:
    Table (default):    Human-readable table with columns
    JSON:               Machine-readable JSON array

Integration:
    This script uses detect-app-runtime.sh for runtime detection
EOF
}

# =============================================================================
# FRAMEWORK DETECTION
# =============================================================================

detect_framework() {
    local app_path="$1"
    local runtime="$2"

    case "$runtime" in
        php)
            if [[ -f "${app_path}/artisan" ]] || grep -q "laravel/framework" "${app_path}/composer.json" 2>/dev/null; then
                echo "Laravel"
            elif [[ -f "${app_path}/wp-config.php" ]] || [[ -f "${app_path}/wp-settings.php" ]]; then
                echo "WordPress"
            elif [[ -f "${app_path}/composer.json" ]]; then
                echo "PHP/Composer"
            else
                echo "Generic PHP"
            fi
            ;;
        node)
            if [[ -f "${app_path}/next.config.js" ]] || [[ -f "${app_path}/next.config.mjs" ]] || grep -q "\"next\"" "${app_path}/package.json" 2>/dev/null; then
                echo "Next.js"
            elif grep -q "\"vite\"" "${app_path}/package.json" 2>/dev/null; then
                if grep -q "\"react\"" "${app_path}/package.json" 2>/dev/null; then
                    echo "React (Vite)"
                else
                    echo "Vite"
                fi
            elif grep -q "\"express\"" "${app_path}/package.json" 2>/dev/null; then
                echo "Express"
            elif grep -q "\"react\"" "${app_path}/package.json" 2>/dev/null; then
                echo "React"
            elif grep -q "\"vue\"" "${app_path}/package.json" 2>/dev/null; then
                echo "Vue.js"
            else
                echo "Node.js"
            fi
            ;;
        python)
            if [[ -f "${app_path}/manage.py" ]]; then
                echo "Django"
            elif grep -q "fastapi" "${app_path}/requirements.txt" 2>/dev/null || grep -q "fastapi" "${app_path}/pyproject.toml" 2>/dev/null; then
                echo "FastAPI"
            elif grep -q "flask" "${app_path}/requirements.txt" 2>/dev/null || [[ -f "${app_path}/app.py" ]]; then
                echo "Flask"
            else
                echo "Python"
            fi
            ;;
        go)
            if grep -q "github.com/gin-gonic/gin" "${app_path}/go.mod" 2>/dev/null; then
                echo "Gin"
            elif grep -q "github.com/labstack/echo" "${app_path}/go.mod" 2>/dev/null; then
                echo "Echo"
            elif grep -q "github.com/gofiber/fiber" "${app_path}/go.mod" 2>/dev/null; then
                echo "Fiber"
            else
                echo "Go (net/http)"
            fi
            ;;
        *)
            echo "Unknown"
            ;;
    esac
}

# =============================================================================
# STATUS DETECTION
# =============================================================================

get_backend_status() {
    local runtime="$1"
    local container_name=""

    case "$runtime" in
        php) container_name="php" ;;
        node) container_name="node" ;;
        python) container_name="python" ;;
        go) container_name="go" ;;
        *) echo "Unknown"; return 1 ;;
    esac

    # Check if container is running
    if $CONTAINER_CMD ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container_name}$"; then
        echo "Active"
    else
        echo "Stopped"
    fi
}

# =============================================================================
# APP SCANNING
# =============================================================================

scan_apps() {
    local -a apps

    # Check if apps directory exists
    if [[ ! -d "$APPS_DIR" ]]; then
        print_status "error" "Apps directory not found: $APPS_DIR"
        return 1
    fi

    # Scan all directories in apps/
    for app_dir in "$APPS_DIR"/*; do
        # Skip if not a directory
        [[ ! -d "$app_dir" ]] && continue

        # Skip hidden directories
        local app_name=$(basename "$app_dir")
        [[ "$app_name" =~ ^\. ]] && continue

        # Detect runtime
        local runtime="unknown"
        if [[ -x "$DETECT_SCRIPT" ]]; then
            runtime=$("$DETECT_SCRIPT" "$app_name" 2>/dev/null) || runtime="unknown"
        fi

        # Filter by runtime if specified
        if [[ -n "$FILTER_RUNTIME" && "$runtime" != "$FILTER_RUNTIME" ]]; then
            continue
        fi

        # Detect framework
        local framework=$(detect_framework "$app_dir" "$runtime")

        # Get backend status
        local status=$(get_backend_status "$runtime")

        # Build app info
        local app_url="http://${app_name}.test"
        local app_path="$app_dir"

        # Store app info as pipe-separated string (avoid colon conflicts with URL)
        apps+=("${app_name}|${runtime}|${framework}|${status}|${app_url}|${app_path}")
    done

    # Return apps array
    printf '%s\n' "${apps[@]}"
}

# =============================================================================
# OUTPUT FORMATTERS
# =============================================================================

output_table() {
    local -a apps
    mapfile -t apps

    if [[ ${#apps[@]} -eq 0 ]]; then
        echo ""
        echo "No applications found in $APPS_DIR"
        echo ""
        return 0
    fi

    echo ""
    echo "DevArch Applications"
    echo "===================="
    echo ""

    # Table header
    if [[ "$SHOW_PATHS" == "true" ]]; then
        printf "%-20s %-10s %-20s %-10s %-30s %s\n" "NAME" "RUNTIME" "FRAMEWORK" "STATUS" "URL" "PATH"
        printf "%s\n" "$(printf '─%.0s' {1..140})"
    else
        printf "%-20s %-10s %-20s %-10s %s\n" "NAME" "RUNTIME" "FRAMEWORK" "STATUS" "URL"
        printf "%s\n" "$(printf '─%.0s' {1..85})"
    fi

    # Table rows
    for app_info in "${apps[@]}"; do
        IFS='|' read -r name runtime framework status url path <<< "$app_info"

        # Color status
        local status_display="$status"
        if [[ "$status" == "Active" ]]; then
            status_display="\e[32m${status}\e[0m"
        else
            status_display="\e[33m${status}\e[0m"
        fi

        if [[ "$SHOW_PATHS" == "true" ]]; then
            printf "%-20s %-10s %-20s %-10b %-30s %s\n" "$name" "$runtime" "$framework" "$status_display" "$url" "$path"
        else
            printf "%-20s %-10s %-20s %-10b %s\n" "$name" "$runtime" "$framework" "$status_display" "$url"
        fi
    done

    echo ""
    echo "Total: ${#apps[@]} application(s)"
    echo ""
}

output_json() {
    local -a apps
    mapfile -t apps

    echo "["

    local first=true
    for app_info in "${apps[@]}"; do
        IFS='|' read -r name runtime framework status url path <<< "$app_info"

        # Add comma separator except for first item
        if [[ "$first" == "true" ]]; then
            first=false
        else
            echo ","
        fi

        # JSON object
        cat << EOF
  {
    "name": "$name",
    "runtime": "$runtime",
    "framework": "$framework",
    "status": "$status",
    "url": "$url",
    "path": "$path"
  }
EOF
    done

    echo ""
    echo "]"
}

output_csv() {
    local -a apps
    mapfile -t apps

    # CSV header
    echo "name,runtime,framework,status,url,path"

    # CSV rows
    for app_info in "${apps[@]}"; do
        IFS='|' read -r name runtime framework status url path <<< "$app_info"
        echo "\"$name\",\"$runtime\",\"$framework\",\"$status\",\"$url\",\"$path\""
    done
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --json)
                OUTPUT_FORMAT="json"
                shift
                ;;
            --csv)
                OUTPUT_FORMAT="csv"
                shift
                ;;
            --runtime)
                if [[ -n "$2" && "$2" != -* ]]; then
                    FILTER_RUNTIME="$2"
                    shift 2
                else
                    print_status "error" "Option --runtime requires a runtime type"
                    exit 1
                fi
                ;;
            --paths)
                SHOW_PATHS=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            *)
                print_status "error" "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse arguments
    parse_arguments "$@"

    # Verbose output
    if [[ "$VERBOSE" == "true" ]]; then
        print_status "info" "Scanning apps directory: $APPS_DIR"
        if [[ -n "$FILTER_RUNTIME" ]]; then
            print_status "info" "Filtering by runtime: $FILTER_RUNTIME"
        fi
        print_status "info" "Output format: $OUTPUT_FORMAT"
        echo "" >&2
    fi

    # Scan apps
    local apps_output
    apps_output=$(scan_apps)

    if [[ $? -ne 0 ]]; then
        exit 1
    fi

    # Output in requested format
    case "$OUTPUT_FORMAT" in
        json)
            echo "$apps_output" | output_json
            ;;
        csv)
            echo "$apps_output" | output_csv
            ;;
        table|*)
            echo "$apps_output" | output_table
            ;;
    esac
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
