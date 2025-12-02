#!/bin/bash

# =============================================================================
# NPM PROXY HOST SETUP AUTOMATION
# =============================================================================
# Automates the creation of Nginx Proxy Manager proxy hosts for applications.
# Detects runtime type (PHP, Node, Python, Go) and configures appropriate
# backend routing.
#
# Usage: ./setup-proxy-host.sh <appname> [options]
# =============================================================================

# Configuration
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
APPS_DIR="${APPS_DIR:-$PROJECT_ROOT/apps}"
CONTEXT_DIR="${CONTEXT_DIR:-$PROJECT_ROOT/context}"

# NPM Configuration (adjust if needed)
NPM_HOST="localhost"
NPM_PORT="81"
NPM_API_URL="http://${NPM_HOST}:${NPM_PORT}/api"

# Detection script
DETECT_SCRIPT="${SCRIPT_DIR}/detect-app-runtime.sh"

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

print_status() {
    local level="$1"
    local message="$2"

    case "$level" in
        info)    echo "[INFO] $message" ;;
        success) echo "[SUCCESS] $message" ;;
        warning) echo "[WARNING] $message" ;;
        error)   echo "[ERROR] $message" ;;
        step)    echo "[STEP] $message" ;;
        *)       echo "$message" ;;
    esac
}

print_usage() {
    cat << EOF
Usage: $(basename "$0") <appname> [options]

Automates Nginx Proxy Manager proxy host setup for applications.

Arguments:
    appname     Name of the application directory in apps/

Options:
    -m, --manual        Output manual configuration instructions instead of API automation
    -v, --verbose       Verbose output
    -d, --domain DOMAIN Custom domain (default: appname.test)
    -h, --help          Show this help message

Examples:
    $(basename "$0") myapp                  # Auto-configure proxy for myapp.test
    $(basename "$0") myapp --manual         # Show manual NPM configuration steps
    $(basename "$0") myapp -d myapp.local   # Use custom domain

Integration:
    This script uses:
    - detect-app-runtime.sh for runtime detection
    - update-hosts.sh for /etc/hosts management

Backend Mapping:
    PHP:    php:8000
    Node:   node:3000
    Python: python:8000
    Go:     go:8080
EOF
}

get_backend_config() {
    local runtime="$1"

    case "$runtime" in
        php)
            echo "php 8000"
            ;;
        node)
            echo "node 3000"
            ;;
        python)
            echo "python 8000"
            ;;
        go)
            echo "go 8080"
            ;;
        *)
            echo "unknown 0"
            return 1
            ;;
    esac
}

print_manual_instructions() {
    local app_name="$1"
    local domain="$2"
    local backend_host="$3"
    local backend_port="$4"
    local runtime="$5"

    cat << EOF

================================================================================
MANUAL NPM CONFIGURATION INSTRUCTIONS
================================================================================

Application: $app_name
Runtime:     $runtime
Domain:      $domain
Backend:     $backend_host:$backend_port

Steps to configure in Nginx Proxy Manager UI (http://localhost:81):

1. Login to NPM Admin UI
   - URL: http://localhost:81
   - Default credentials: admin@example.com / changeme (if first time)

2. Create New Proxy Host
   - Click "Proxy Hosts" in the sidebar
   - Click "Add Proxy Host" button

3. Details Tab:
   - Domain Names: $domain
   - Scheme: http
   - Forward Hostname / IP: $backend_host
   - Forward Port: $backend_port
   - Cache Assets: ✓ (enabled)
   - Block Common Exploits: ✓ (enabled)
   - Websockets Support: ✓ (enabled if needed for your app)

4. SSL Tab:
   - SSL Certificate: Select your local certificate
     (Or None if testing without SSL)
   - Force SSL: ✓ (if using SSL)
   - HTTP/2 Support: ✓ (if using SSL)

5. Advanced Tab (optional):
   Add custom Nginx configuration if needed:

   location / {
       proxy_pass http://$backend_host:$backend_port;
       proxy_set_header Host \$host;
       proxy_set_header X-Real-IP \$remote_addr;
       proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto \$scheme;

       # Increase timeouts if needed
       proxy_connect_timeout 60s;
       proxy_send_timeout 60s;
       proxy_read_timeout 60s;
   }

6. Save the proxy host

7. Update /etc/hosts:
   Run: sudo ./scripts/update-hosts.sh

   Or manually add:
   127.0.0.1 $domain

8. Test the configuration:
   - Visit: http://$domain (or https://$domain if SSL enabled)
   - Check NPM logs if issues occur

================================================================================
Runtime-Specific Notes:
================================================================================

EOF

    case "$runtime" in
        php)
            cat << EOF
PHP Applications:
- Document root is typically in public/ subdirectory
- If app doesn't load, ensure PHP-FPM is running in the container
- Check container logs: podman logs php

EOF
            ;;
        node)
            cat << EOF
Node.js Applications:
- Ensure your Node app is listening on port 3000 inside the container
- Start your app: cd /app/$app_name && npm start
- Check container logs: podman logs node
- For development with hot reload, ensure port 5173 (Vite) is also exposed if needed

EOF
            ;;
        python)
            cat << EOF
Python Applications:
- Django: Ensure app runs on 0.0.0.0:8000 (not 127.0.0.1)
  Command: python manage.py runserver 0.0.0.0:8000
- Flask: Set FLASK_RUN_HOST=0.0.0.0 and FLASK_RUN_PORT=8000
  Command: flask run --host=0.0.0.0 --port=8000
- FastAPI: Use uvicorn with --host 0.0.0.0
  Command: uvicorn main:app --host 0.0.0.0 --port 8000
- Check container logs: podman logs python

EOF
            ;;
        go)
            cat << EOF
Go Applications:
- Ensure your Go app listens on 0.0.0.0:8080 (not 127.0.0.1)
- Build and run: cd /app/$app_name && go run main.go
- Check container logs: podman logs go

EOF
            ;;
    esac

    cat << EOF
================================================================================
Troubleshooting:
================================================================================

If the proxy host doesn't work:

1. Check backend container is running:
   podman ps | grep $backend_host

2. Check backend is accessible from NPM container:
   podman exec nginx-proxy-manager curl http://$backend_host:$backend_port

3. Check NPM logs:
   podman logs nginx-proxy-manager

4. Check backend container logs:
   podman logs $backend_host

5. Verify network connectivity:
   Both containers should be on 'microservices-net' network

6. Test direct access to backend:
   curl http://localhost:$(./scripts/detect-app-runtime.sh $app_name -i | grep -oP 'port=\K[0-9]+')

================================================================================

EOF
}

add_to_hosts_file() {
    local domain="$1"

    print_status "step" "Updating /etc/hosts file..."

    # Check if entry already exists
    if grep -q "127.0.0.1.*$domain" /etc/hosts 2>/dev/null; then
        print_status "info" "Entry already exists in /etc/hosts"
        return 0
    fi

    # Try to update using update-hosts.sh if available
    if [[ -f "${SCRIPT_DIR}/update-hosts.sh" ]]; then
        print_status "info" "Running update-hosts.sh to manage hosts file"
        "${SCRIPT_DIR}/update-hosts.sh"
    else
        print_status "warning" "update-hosts.sh not found, manual hosts file update needed"
        print_status "info" "Add this line to /etc/hosts:"
        echo "    127.0.0.1 $domain"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local app_name=""
    local manual_mode=false
    local verbose=false
    local custom_domain=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                print_usage
                exit 0
                ;;
            -m|--manual)
                manual_mode=true
                shift
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            -d|--domain)
                custom_domain="$2"
                shift 2
                ;;
            -*)
                print_status "error" "Unknown option: $1"
                print_usage
                exit 1
                ;;
            *)
                if [[ -z "$app_name" ]]; then
                    app_name="$1"
                else
                    print_status "error" "Multiple app names provided"
                    print_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate app name
    if [[ -z "$app_name" ]]; then
        print_status "error" "App name is required"
        print_usage
        exit 1
    fi

    # Check if app exists
    if [[ ! -d "${APPS_DIR}/${app_name}" ]]; then
        print_status "error" "App directory not found: ${APPS_DIR}/${app_name}"
        exit 1
    fi

    # Set domain
    local domain="${custom_domain:-${app_name}.test}"

    print_status "info" "Setting up proxy host for: $app_name"
    print_status "info" "Domain: $domain"
    echo ""

    # Detect runtime
    if [[ ! -x "$DETECT_SCRIPT" ]]; then
        print_status "error" "Runtime detection script not found or not executable: $DETECT_SCRIPT"
        exit 1
    fi

    print_status "step" "Detecting application runtime..."
    local runtime
    runtime=$("$DETECT_SCRIPT" "$app_name")

    if [[ "$runtime" == "unknown" ]]; then
        print_status "error" "Could not detect runtime for $app_name"
        print_status "info" "Ensure app has proper marker files (composer.json, package.json, etc.)"
        exit 1
    fi

    print_status "success" "Detected runtime: $runtime"
    echo ""

    # Get backend configuration
    local backend_config
    backend_config=$(get_backend_config "$runtime")
    local backend_host=$(echo "$backend_config" | cut -d' ' -f1)
    local backend_port=$(echo "$backend_config" | cut -d' ' -f2)

    print_status "info" "Backend configuration:"
    print_status "info" "  Container: $backend_host"
    print_status "info" "  Port: $backend_port"
    echo ""

    # Manual mode or API mode
    if [[ "$manual_mode" == "true" ]]; then
        print_manual_instructions "$app_name" "$domain" "$backend_host" "$backend_port" "$runtime"
    else
        print_status "warning" "NPM API automation is not yet implemented"
        print_status "info" "Showing manual configuration instructions instead..."
        echo ""
        print_manual_instructions "$app_name" "$domain" "$backend_host" "$backend_port" "$runtime"
    fi

    # Offer to update hosts file
    print_status "step" "Hosts file management"
    add_to_hosts_file "$domain"

    echo ""
    print_status "success" "Setup complete!"
    print_status "info" "Once NPM proxy host is configured, access your app at: http://$domain"
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
