#!/bin/bash

# =============================================================================
# APP RUNTIME DETECTOR
# =============================================================================
# Detects the runtime type (PHP, Node, Python, Go) for an application
# based on file markers present in the app directory.
#
# Usage: ./detect-app-runtime.sh <appname>
# Output: php|node|python|go|unknown
#
# Detection Priority (if multiple markers found): PHP > Node > Python > Go
# =============================================================================

# Configuration
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
APPS_DIR="${APPS_DIR:-$PROJECT_ROOT/apps}"

# =============================================================================
# FUNCTIONS
# =============================================================================

print_usage() {
    cat << EOF
Usage: $(basename "$0") <appname>

Detects the runtime type for an application based on file markers.

Arguments:
    appname     Name of the application directory in apps/

Output:
    php         PHP application detected
    node        Node.js application detected
    python      Python application detected
    go          Go application detected
    unknown     Could not determine runtime type

Detection Logic:
    PHP:     composer.json, index.php, wp-config.php, artisan
    Node:    package.json
    Python:  requirements.txt, pyproject.toml, manage.py, main.py
    Go:      go.mod, main.go

Priority (if multiple markers found): PHP > Node > Python > Go

Examples:
    $(basename "$0") playground      # Detects runtime for apps/playground
    $(basename "$0") myapp           # Detects runtime for apps/myapp

Integration:
    This script is used by:
    - setup-proxy-host.sh (NPM proxy host automation)
    - update-hosts.sh (multi-backend hosts file management)
EOF
}

detect_runtime() {
    local app_name="$1"
    local app_path="${APPS_DIR}/${app_name}"

    # Check if app directory exists
    if [[ ! -d "$app_path" ]]; then
        echo "ERROR: App directory not found: $app_path" >&2
        echo "unknown"
        return 1
    fi

    # PHP detection (highest priority)
    if [[ -f "${app_path}/composer.json" ]] || \
       [[ -f "${app_path}/index.php" ]] || \
       [[ -f "${app_path}/public/index.php" ]] || \
       [[ -f "${app_path}/wp-config.php" ]] || \
       [[ -f "${app_path}/artisan" ]]; then
        echo "php"
        return 0
    fi

    # Node detection
    if [[ -f "${app_path}/package.json" ]]; then
        echo "node"
        return 0
    fi

    # Python detection
    if [[ -f "${app_path}/requirements.txt" ]] || \
       [[ -f "${app_path}/pyproject.toml" ]] || \
       [[ -f "${app_path}/manage.py" ]] || \
       [[ -f "${app_path}/main.py" ]]; then
        # Additional check for main.py to avoid false positives
        if [[ -f "${app_path}/main.py" ]]; then
            # Check if it's a Python file (has common Python imports)
            if grep -q -E "^(import|from|def|class)" "${app_path}/main.py" 2>/dev/null; then
                echo "python"
                return 0
            fi
        else
            echo "python"
            return 0
        fi
    fi

    # Go detection
    if [[ -f "${app_path}/go.mod" ]] || \
       [[ -f "${app_path}/main.go" ]]; then
        echo "go"
        return 0
    fi

    # No markers found
    echo "unknown"
    return 1
}

get_backend_info() {
    local runtime="$1"

    case "$runtime" in
        php)
            echo "port=8100 container=php internal_port=8000"
            ;;
        node)
            echo "port=8200 container=node internal_port=3000"
            ;;
        python)
            echo "port=8300 container=python internal_port=8000"
            ;;
        go)
            echo "port=8400 container=go internal_port=8080"
            ;;
        *)
            echo "port=unknown container=unknown internal_port=unknown"
            return 1
            ;;
    esac
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse arguments
    if [[ $# -eq 0 ]] || [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
        print_usage
        exit 0
    fi

    local app_name="$1"
    local verbose=false
    local show_backend_info=false

    # Check for flags
    shift
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -v|--verbose)
                verbose=true
                shift
                ;;
            -i|--info)
                show_backend_info=true
                shift
                ;;
            *)
                echo "ERROR: Unknown option: $1" >&2
                print_usage
                exit 1
                ;;
        esac
    done

    # Detect runtime
    local runtime
    runtime=$(detect_runtime "$app_name")
    local exit_code=$?

    # Output result
    if [[ "$verbose" == "true" ]]; then
        echo "App: $app_name" >&2
        echo "Path: ${APPS_DIR}/${app_name}" >&2
        echo "Runtime: $runtime" >&2

        if [[ "$show_backend_info" == "true" ]] && [[ "$runtime" != "unknown" ]]; then
            local backend_info
            backend_info=$(get_backend_info "$runtime")
            echo "Backend: $backend_info" >&2
        fi
        echo "" >&2
    fi

    if [[ "$show_backend_info" == "true" ]] && [[ "$runtime" != "unknown" ]]; then
        get_backend_info "$runtime"
    else
        echo "$runtime"
    fi

    return $exit_code
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
