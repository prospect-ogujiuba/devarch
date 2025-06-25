#!/bin/zsh

# =============================================================================
# SIMPLE MKCERT SSL SETUP SCRIPT
# =============================================================================
# Clean, simple SSL certificate generation using only mkcert
# Certificates placed directly in ./config/traefik/certs/

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS
# =============================================================================

opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_force_regenerate=false
opt_custom_domain="*.test"

# Certificate paths
CERT_OUTPUT_DIR="$PROJECT_ROOT/config/traefik/certs"

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Simple SSL certificate generation using mkcert with automatic trust.
    Certificates are placed in ./config/traefik/certs/ for immediate use.

OPTIONS:
    -s, --sudo              Use sudo for mkcert installation if needed
    -e, --errors            Show detailed error messages
    -f, --force             Force regeneration of existing certificates
    -d, --domain DOMAIN     Custom domain (default: *.test)
    -h, --help              Show this help message

EXAMPLES:
    $0                      # Generate *.test certificate
    $0 -f                   # Force regenerate existing certificate
    $0 -d "*.local"         # Generate *.local certificate
    $0 -s                   # Use sudo for mkcert installation

NOTES:
    - Requires mkcert (will attempt to install if missing)
    - Certificates are automatically trusted by your system/browsers
    - Output: ./config/traefik/certs/local.crt and local.key
EOF
}

# =============================================================================
# MAIN FUNCTIONS
# =============================================================================

check_mkcert() {
    if command -v mkcert >/dev/null 2>&1; then
        return 0
    else
        print_status "warning" "mkcert not found - installing..."
        install_mkcert
        return $?
    fi
}

install_mkcert() {
    case "$OS_TYPE" in
        "linux"|"wsl")
            if command -v snap >/dev/null 2>&1; then
                [[ "$opt_use_sudo" == "true" ]] && sudo snap install mkcert || snap install mkcert
            else
                # Download binary
                curl -fsSL "https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-v1.4.4-linux-amd64" -o "/tmp/mkcert"
                chmod +x "/tmp/mkcert"
                [[ "$opt_use_sudo" == "true" ]] && sudo mv "/tmp/mkcert" "/usr/local/bin/mkcert" || mv "/tmp/mkcert" "$HOME/.local/bin/mkcert"
            fi
            ;;
        "macos")
            if command -v brew >/dev/null 2>&1; then
                brew install mkcert
            else
                handle_error "Please install Homebrew first: https://brew.sh"
            fi
            ;;
        *)
            handle_error "Unsupported OS for automatic mkcert installation"
            ;;
    esac
}

generate_certificates() {
    print_status "step" "Generating SSL certificates with mkcert for dual domain structure..."
    
    # Create output directory
    mkdir -p "$CERT_OUTPUT_DIR"
    cd "$CERT_OUTPUT_DIR"
    
    # Install mkcert CA if not already done
    print_status "info" "Setting up mkcert Certificate Authority..."
    mkcert -install
    
    print_status "info" "Generating certificate for infrastructure (.test) and development (.dev.test) domains..."
    
    # Generate certificate with BOTH domain patterns
    mkcert -cert-file "local.crt" -key-file "local.key" \
        "*.test" \
        "*.dev.test" \
        "test" \
        "dev.test" \
        "localhost" \
        "127.0.0.1" \
        "::1" \
        "nginx.test" \
        "npm.test" \
        "traefik.test" \
        "adminer.test" \
        "grafana.test" \
        "prometheus.test" \
        "postgres.test" \
        "redis.test" \
        "mongodb.test" \
        "mailpit.test" \
        "gitea.test" \
        "matomo.test" \
        "metabase.test" \
        "nocodb.test" \
        "n8n.test" \
        "langflow.test" \
        "projects.dev.test"
    
    if [[ $? -eq 0 ]]; then
        print_status "success" "Certificates generated successfully!"
        print_status "info" "Certificate covers:"
        echo "  🏗️  Infrastructure: *.test (managed by Traefik)"
        echo "  🧪 Development: *.dev.test (managed via NPM)"
        echo "  📄 $CERT_OUTPUT_DIR/local.crt"
        echo "  🔑 $CERT_OUTPUT_DIR/local.key"
        return 0
    else
        handle_error "Certificate generation failed"
    fi
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--sudo)
                opt_use_sudo=true
                shift
                ;;
            -e|--errors)
                opt_show_errors=true
                shift
                ;;
            -f|--force)
                opt_force_regenerate=true
                shift
                ;;
            -d|--domain)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_custom_domain="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a domain value"
                fi
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
# MAIN EXECUTION
# =============================================================================

main() {
    parse_arguments "$@"
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    print_status "step" "Setting up SSL certificates for dual domain architecture..."
    
    # Check for existing certificates
    if [[ -f "$CERT_OUTPUT_DIR/local.crt" && "$opt_force_regenerate" == "false" ]]; then
        print_status "warning" "Certificates already exist. Use -f to force regeneration."
        exit 0
    fi
    
    # Check/install mkcert
    check_mkcert
    
    # Generate certificates
    generate_certificates
    
    print_status "success" "SSL setup completed for dual domain architecture!"
    echo ""
    echo "🔗 Test your infrastructure services:"
    echo "   https://npm.test (Nginx Proxy Manager)"
    echo "   https://traefik.test (Traefik Dashboard)"
    echo "   https://grafana.test (Grafana)"
    echo ""
    echo "🔗 Test your development projects:"
    echo "   https://projects.dev.test (Project listing)"
    echo "   https://[project-name].dev.test (Individual projects)"
    echo ""
    echo "✅ Certificates automatically trusted by your system and browsers!"
}

# Only run main if script is executed directly
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi