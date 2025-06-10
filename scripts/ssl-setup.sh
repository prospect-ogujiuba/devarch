#!/bin/bash
# =================================================================
# DEVARCH SSL CERTIFICATE SETUP SCRIPT
# =================================================================
# This script manages SSL certificates for the DevArch development
# environment, including generation, installation, and trust management
# across different operating systems.

set -euo pipefail

# =================================================================
# CONFIGURATION AND VARIABLES
# =================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_DIR="$PROJECT_ROOT/config"
SSL_DIR="$CONFIG_DIR/ssl"

# Load environment variables
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
else
    echo "❌ Error: .env file not found in $PROJECT_ROOT"
    exit 1
fi

# Script options
VERBOSE=false
FORCE=false
SKIP_TRUST=false
REGENERATE=false
EXPORT_FORMAT=""

# Certificate configuration
CERT_VALIDITY_DAYS=${SSL_CERT_VALIDITY_DAYS:-365}
CERT_KEY_SIZE=${SSL_CERT_KEY_SIZE:-2048}
CERT_ALGORITHM=${SSL_CERT_ALGORITHM:-rsa}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Supported domains (add new domains here)
DOMAINS=(
    "*.test"
    "test"
    "localhost"
    "welcome.test"
    "mailpit.test"
    "adminer.test"
    "phpmyadmin.test"
    "metabase.test"
    "redis.test"
    "swagger.test"
    "nginx.test"
    "mongodb.test"
    "sonarqube.test"
    "gitiles.test"
    "grafana.test"
    "prometheus.test"
)

# =================================================================
# UTILITY FUNCTIONS
# =================================================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${PURPLE}[$(date +'%Y-%m-%d %H:%M:%S')] DEBUG: $1${NC}"
    fi
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command_exists lsb_release; then
            echo "$(lsb_release -si | tr '[:upper:]' '[:lower:]')"
        elif [[ -f /etc/os-release ]]; then
            echo "$(grep '^ID=' /etc/os-release | cut -d'=' -f2 | tr -d '"')"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Check if running in WSL
is_wsl() {
    [[ -n "${WSL_DISTRO_NAME:-}" ]] || [[ -f /proc/version && $(grep -i microsoft /proc/version) ]]
}

# =================================================================
# CERTIFICATE GENERATION
# =================================================================
generate_root_ca() {
    local ca_key="$SSL_DIR/devarch-ca.key"
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    local ca_config="$SSL_DIR/ca.conf"
    
    info "Generating Root Certificate Authority..."
    
    # Create CA configuration
    cat > "$ca_config" << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
prompt = no

[req_distinguished_name]
C = ${SSL_COUNTRY}
ST = ${SSL_STATE}
L = ${SSL_CITY}
O = ${SSL_ORG}
OU = Development CA
CN = DevArch Root CA
emailAddress = ca@devarch.test

[v3_ca]
basicConstraints = critical,CA:TRUE
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer:always
EOF
    
    # Generate CA private key
    debug "Generating CA private key..."
    openssl genrsa -out "$ca_key" 4096
    
    # Generate CA certificate
    debug "Generating CA certificate..."
    openssl req -new -x509 -key "$ca_key" -out "$ca_cert" \
        -days $((CERT_VALIDITY_DAYS * 10)) \
        -config "$ca_config" \
        -extensions v3_ca
    
    # Set proper permissions
    chmod 600 "$ca_key"
    chmod 644 "$ca_cert"
    
    success "Root CA generated successfully"
}

generate_wildcard_certificate() {
    local ca_key="$SSL_DIR/devarch-ca.key"
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    local cert_key="$SSL_DIR/wildcard.test.key"
    local cert_csr="$SSL_DIR/wildcard.test.csr"
    local cert_crt="$SSL_DIR/wildcard.test.crt"
    local cert_config="$SSL_DIR/wildcard.conf"
    
    info "Generating wildcard certificate for *.test domains..."
    
    # Create certificate configuration
    cat > "$cert_config" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = ${SSL_COUNTRY}
ST = ${SSL_STATE}
L = ${SSL_CITY}
O = ${SSL_ORG}
OU = Development
CN = *.test
emailAddress = admin@devarch.test

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
EOF
    
    # Add all domains to SAN
    local dns_count=1
    for domain in "${DOMAINS[@]}"; do
        echo "DNS.${dns_count} = ${domain}" >> "$cert_config"
        ((dns_count++))
    done
    
    # Generate certificate private key
    debug "Generating certificate private key..."
    if [[ "$CERT_ALGORITHM" == "ecdsa" ]]; then
        openssl ecparam -genkey -name secp384r1 -out "$cert_key"
    else
        openssl genrsa -out "$cert_key" "$CERT_KEY_SIZE"
    fi
    
    # Generate certificate signing request
    debug "Generating certificate signing request..."
    openssl req -new -key "$cert_key" -out "$cert_csr" \
        -config "$cert_config"
    
    # Generate certificate signed by CA
    debug "Signing certificate with CA..."
    openssl x509 -req -in "$cert_csr" -CA "$ca_cert" -CAkey "$ca_key" \
        -CAcreateserial -out "$cert_crt" \
        -days "$CERT_VALIDITY_DAYS" \
        -extensions v3_req \
        -extfile "$cert_config"
    
    # Create full chain certificate
    cat "$cert_crt" "$ca_cert" > "$SSL_DIR/fullchain.pem"
    cp "$cert_key" "$SSL_DIR/privkey.pem"
    
    # Set proper permissions
    chmod 600 "$cert_key" "$SSL_DIR/privkey.pem"
    chmod 644 "$cert_crt" "$SSL_DIR/fullchain.pem"
    
    # Clean up temporary files
    rm -f "$cert_csr"
    
    success "Wildcard certificate generated successfully"
}

copy_certificates_to_container() {
    info "Copying certificates to Nginx container..."
    
    # Wait for container to be running
    local max_attempts=30
    local attempt=1
    
    while ! podman container exists nginx-proxy-manager 2>/dev/null || \
          [[ "$(podman inspect nginx-proxy-manager --format '{{.State.Status}}')" != "running" ]]; do
        if [[ $attempt -ge $max_attempts ]]; then
            error "Nginx container is not running after $max_attempts attempts"
            return 1
        fi
        debug "Waiting for Nginx container (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    # Create certificate directory in container
    podman exec nginx-proxy-manager mkdir -p /etc/letsencrypt/live/wildcard.test
    
    # Copy certificates to container
    podman cp "$SSL_DIR/fullchain.pem" nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/
    podman cp "$SSL_DIR/privkey.pem" nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/
    
    # Set proper permissions in container
    podman exec nginx-proxy-manager chmod 644 /etc/letsencrypt/live/wildcard.test/fullchain.pem
    podman exec nginx-proxy-manager chmod 600 /etc/letsencrypt/live/wildcard.test/privkey.pem
    
    success "Certificates copied to container"
}

# =================================================================
# CERTIFICATE INSTALLATION
# =================================================================
install_ca_linux() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    local os_type
    os_type=$(detect_os)
    
    info "Installing CA certificate on Linux ($os_type)..."
    
    case "$os_type" in
        ubuntu|debian)
            sudo cp "$ca_cert" /usr/local/share/ca-certificates/devarch-ca.crt
            sudo update-ca-certificates
            ;;
        fedora|centos|rhel)
            sudo cp "$ca_cert" /etc/pki/ca-trust/source/anchors/devarch-ca.crt
            sudo update-ca-trust
            ;;
        arch|manjaro)
            sudo cp "$ca_cert" /etc/ca-certificates/trust-source/anchors/devarch-ca.crt
            sudo trust extract-compat
            ;;
        opensuse*)
            sudo cp "$ca_cert" /usr/share/pki/trust/anchors/devarch-ca.crt
            sudo update-ca-certificates
            ;;
        *)
            warn "Unsupported Linux distribution: $os_type"
            warn "Please manually install: $ca_cert"
            return 1
            ;;
    esac
    
    success "CA certificate installed on Linux"
}

install_ca_macos() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    
    info "Installing CA certificate on macOS..."
    
    # Add to system keychain
    sudo security add-trusted-cert -d -r trustRoot \
        -k /Library/Keychains/System.keychain "$ca_cert"
    
    # Also add to user keychain for applications
    security add-trusted-cert -d -r trustRoot \
        -k ~/Library/Keychains/login.keychain "$ca_cert"
    
    success "CA certificate installed on macOS"
}

install_ca_windows() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    
    info "Installing CA certificate on Windows..."
    
    if is_wsl; then
        # WSL environment - copy to Windows
        local windows_temp="/mnt/c/temp"
        mkdir -p "$windows_temp"
        cp "$ca_cert" "$windows_temp/devarch-ca.crt"
        
        warn "Certificate copied to C:\\temp\\devarch-ca.crt"
        warn "Please run the following command in Windows PowerShell as Administrator:"
        warn "Import-Certificate -FilePath 'C:\\temp\\devarch-ca.crt' -CertStoreLocation Cert:\\LocalMachine\\Root"
    else
        warn "Please manually install the CA certificate:"
        warn "1. Copy $ca_cert to Windows"
        warn "2. Run PowerShell as Administrator"
        warn "3. Execute: Import-Certificate -FilePath 'path\\to\\devarch-ca.crt' -CertStoreLocation Cert:\\LocalMachine\\Root"
    fi
}

install_ca_firefox() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    
    info "Installing CA certificate for Firefox..."
    
    # Find Firefox profiles
    local firefox_profiles=()
    
    case "$(detect_os)" in
        linux)
            if [[ -d "$HOME/.mozilla/firefox" ]]; then
                while IFS= read -r -d '' profile; do
                    firefox_profiles+=("$profile")
                done < <(find "$HOME/.mozilla/firefox" -name "*.default*" -type d -print0)
            fi
            ;;
        macos)
            if [[ -d "$HOME/Library/Application Support/Firefox/Profiles" ]]; then
                while IFS= read -r -d '' profile; do
                    firefox_profiles+=("$profile")
                done < <(find "$HOME/Library/Application Support/Firefox/Profiles" -name "*.default*" -type d -print0)
            fi
            ;;
        windows)
            if is_wsl && [[ -d "/mnt/c/Users/$USER/AppData/Roaming/Mozilla/Firefox/Profiles" ]]; then
                while IFS= read -r -d '' profile; do
                    firefox_profiles+=("$profile")
                done < <(find "/mnt/c/Users/$USER/AppData/Roaming/Mozilla/Firefox/Profiles" -name "*.default*" -type d -print0)
            fi
            ;;
    esac
    
    if [[ ${#firefox_profiles[@]} -eq 0 ]]; then
        warn "No Firefox profiles found"
        warn "Please manually import $ca_cert into Firefox:"
        warn "1. Go to about:preferences#privacy"
        warn "2. Scroll to Certificates -> View Certificates"
        warn "3. Authorities tab -> Import"
        warn "4. Select the CA certificate and trust for websites"
        return 0
    fi
    
    # Install certificate in each profile
    for profile in "${firefox_profiles[@]}"; do
        local profile_name
        profile_name=$(basename "$profile")
        debug "Installing certificate in Firefox profile: $profile_name"
        
        if command_exists certutil; then
            certutil -A -n "DevArch CA" -t "TCu,Cu,Tu" -i "$ca_cert" -d "$profile"
            success "Certificate installed in Firefox profile: $profile_name"
        else
            warn "certutil not found. Please install NSS tools to auto-install in Firefox"
            warn "Manual installation required for profile: $profile_name"
        fi
    done
}

# =================================================================
# CERTIFICATE VALIDATION
# =================================================================
validate_certificates() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    local cert_crt="$SSL_DIR/wildcard.test.crt"
    local cert_key="$SSL_DIR/wildcard.test.key"
    
    info "Validating certificates..."
    
    # Check if files exist
    local required_files=("$ca_cert" "$cert_crt" "$cert_key")
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            error "Certificate file not found: $file"
            return 1
        fi
    done
    
    # Validate CA certificate
    debug "Validating CA certificate..."
    if ! openssl x509 -in "$ca_cert" -text -noout >/dev/null 2>&1; then
        error "Invalid CA certificate"
        return 1
    fi
    
    # Validate wildcard certificate
    debug "Validating wildcard certificate..."
    if ! openssl x509 -in "$cert_crt" -text -noout >/dev/null 2>&1; then
        error "Invalid wildcard certificate"
        return 1
    fi
    
    # Validate private key
    debug "Validating private key..."
    if ! openssl rsa -in "$cert_key" -check -noout >/dev/null 2>&1; then
        error "Invalid private key"
        return 1
    fi
    
    # Check if certificate and key match
    debug "Verifying certificate and key match..."
    local cert_modulus key_modulus
    cert_modulus=$(openssl x509 -noout -modulus -in "$cert_crt" | openssl md5)
    key_modulus=$(openssl rsa -noout -modulus -in "$cert_key" | openssl md5)
    
    if [[ "$cert_modulus" != "$key_modulus" ]]; then
        error "Certificate and private key do not match"
        return 1
    fi
    
    # Check certificate validity
    debug "Checking certificate validity..."
    if ! openssl x509 -in "$cert_crt" -checkend 86400 >/dev/null 2>&1; then
        warn "Certificate expires within 24 hours"
    fi
    
    # Verify certificate chain
    debug "Verifying certificate chain..."
    if ! openssl verify -CAfile "$ca_cert" "$cert_crt" >/dev/null 2>&1; then
        error "Certificate chain validation failed"
        return 1
    fi
    
    success "All certificates are valid"
    
    # Show certificate information
    show_certificate_info
}

show_certificate_info() {
    local ca_cert="$SSL_DIR/devarch-ca.crt"
    local cert_crt="$SSL_DIR/wildcard.test.crt"
    
    echo
    echo -e "${CYAN}📋 Certificate Information:${NC}"
    echo
    
    # CA Certificate info
    echo -e "${YELLOW}Root CA Certificate:${NC}"
    openssl x509 -in "$ca_cert" -text -noout | grep -E "(Subject:|Not Before|Not After|Subject Alternative Name)" || true
    echo
    
    # Wildcard Certificate info
    echo -e "${YELLOW}Wildcard Certificate:${NC}"
    openssl x509 -in "$cert_crt" -text -noout | grep -E "(Subject:|Not Before|Not After|Subject Alternative Name)" || true
    echo
    
    # Supported domains
    echo -e "${YELLOW}Supported Domains:${NC}"
    for domain in "${DOMAINS[@]}"; do
        echo "  • $domain"
    done
}

# =================================================================
# CERTIFICATE EXPORT
# =================================================================
export_certificates() {
    local format="${1:-pem}"
    local export_dir="$SSL_DIR/export"
    
    info "Exporting certificates in $format format..."
    
    mkdir -p "$export_dir"
    
    case "$format" in
        pem)
            export_pem_format "$export_dir"
            ;;
        p12|pfx)
            export_p12_format "$export_dir"
            ;;
        der)
            export_der_format "$export_dir"
            ;;
        jks)
            export_jks_format "$export_dir"
            ;;
        *)
            error "Unsupported export format: $format"
            return 1
            ;;
    esac
    
    success "Certificates exported to: $export_dir"
}

export_pem_format() {
    local export_dir="$1"
    
    # Copy PEM files
    cp "$SSL_DIR/devarch-ca.crt" "$export_dir/ca.pem"
    cp "$SSL_DIR/wildcard.test.crt" "$export_dir/cert.pem"
    cp "$SSL_DIR/wildcard.test.key" "$export_dir/key.pem"
    cp "$SSL_DIR/fullchain.pem" "$export_dir/fullchain.pem"
    
    debug "PEM certificates exported"
}

export_p12_format() {
    local export_dir="$1"
    local p12_password="${SSL_P12_PASSWORD:-devarch123}"
    
    # Create PKCS#12 bundle
    openssl pkcs12 -export -out "$export_dir/certificate.p12" \
        -inkey "$SSL_DIR/wildcard.test.key" \
        -in "$SSL_DIR/wildcard.test.crt" \
        -certfile "$SSL_DIR/devarch-ca.crt" \
        -passout pass:"$p12_password"
    
    echo "$p12_password" > "$export_dir/p12-password.txt"
    
    debug "PKCS#12 certificate exported with password: $p12_password"
}

export_der_format() {
    local export_dir="$1"
    
    # Convert to DER format
    openssl x509 -outform DER -in "$SSL_DIR/devarch-ca.crt" -out "$export_dir/ca.der"
    openssl x509 -outform DER -in "$SSL_DIR/wildcard.test.crt" -out "$export_dir/cert.der"
    openssl rsa -outform DER -in "$SSL_DIR/wildcard.test.key" -out "$export_dir/key.der"
    
    debug "DER certificates exported"
}

export_jks_format() {
    local export_dir="$1"
    local jks_password="${SSL_JKS_PASSWORD:-devarch123}"
    
    if ! command_exists keytool; then
        error "keytool not found. Please install Java to export JKS format"
        return 1
    fi
    
    # First create PKCS#12, then convert to JKS
    export_p12_format "$export_dir"
    
    keytool -importkeystore \
        -srckeystore "$export_dir/certificate.p12" \
        -srcstoretype PKCS12 \
        -srcstorepass "${SSL_P12_PASSWORD:-devarch123}" \
        -destkeystore "$export_dir/certificate.jks" \
        -deststoretype JKS \
        -deststorepass "$jks_password"
    
    echo "$jks_password" > "$export_dir/jks-password.txt"
    
    debug "JKS certificate exported with password: $jks_password"
}

# =================================================================
# CERTIFICATE TESTING
# =================================================================
test_certificates() {
    info "Testing SSL certificates..."
    
    # Test domains
    local test_domains=("welcome.test" "mailpit.test" "adminer.test")
    local failed_tests=()
    
    for domain in "${test_domains[@]}"; do
        debug "Testing $domain..."
        
        # Test with curl (ignore connection refused, we only care about SSL)
        if curl -I --connect-timeout 5 --max-time 10 --cacert "$SSL_DIR/devarch-ca.crt" "https://$domain" >/dev/null 2>&1 || \
           curl -I --connect-timeout 5 --max-time 10 --cacert "$SSL_DIR/devarch-ca.crt" "https://$domain" 2>&1 | grep -q "SSL certificate problem" && false; then
            success "SSL test passed for $domain"
        else
            failed_tests+=("$domain")
            warn "SSL test failed for $domain"
        fi
    done
    
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        success "All SSL tests passed"
    else
        warn "SSL tests failed for: ${failed_tests[*]}"
        warn "This may be normal if services are not running"
    fi
}

# =================================================================
# MAIN OPERATIONS
# =================================================================
setup_certificates() {
    log "Setting up SSL certificates for DevArch..."
    
    # Create SSL directory
    mkdir -p "$SSL_DIR"
    
    # Check if certificates exist and are valid
    if [[ "$REGENERATE" != "true" ]] && [[ -f "$SSL_DIR/devarch-ca.crt" ]] && [[ -f "$SSL_DIR/wildcard.test.crt" ]]; then
        if validate_certificates 2>/dev/null; then
            info "Valid certificates already exist"
            if [[ "$FORCE" != "true" ]]; then
                read -p "Regenerate certificates anyway? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    info "Using existing certificates"
                    copy_certificates_to_container
                    return 0
                fi
            fi
        fi
    fi
    
    # Generate certificates
    generate_root_ca
    generate_wildcard_certificate
    
    # Validate generated certificates
    validate_certificates
    
    # Copy to container
    copy_certificates_to_container
    
    # Install on host system
    if [[ "$SKIP_TRUST" != "true" ]]; then
        install_certificates
    fi
    
    success "SSL certificates setup completed"
}

install_certificates() {
    local os_type
    os_type=$(detect_os)
    
    log "Installing certificates on host system ($os_type)..."
    
    case "$os_type" in
        ubuntu|debian|fedora|centos|rhel|arch|manjaro|opensuse*)
            install_ca_linux
            ;;
        macos)
            install_ca_macos
            ;;
        windows)
            install_ca_windows
            ;;
        *)
            warn "Unsupported operating system: $os_type"
            warn "Please manually install: $SSL_DIR/devarch-ca.crt"
            ;;
    esac
    
    # Install for Firefox
    if command_exists firefox || [[ -d "$HOME/.mozilla" ]]; then
        install_ca_firefox
    fi
}

remove_certificates() {
    if [[ "$FORCE" != "true" ]]; then
        read -p "Are you sure you want to remove all SSL certificates? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Operation cancelled"
            return 0
        fi
    fi
    
    log "Removing SSL certificates..."
    
    # Remove from container
    if podman container exists nginx-proxy-manager 2>/dev/null; then
        podman exec nginx-proxy-manager rm -rf /etc/letsencrypt/live/wildcard.test 2>/dev/null || true
        info "Certificates removed from container"
    fi
    
    # Remove local files
    if [[ -d "$SSL_DIR" ]]; then
        rm -rf "$SSL_DIR"/*
        success "Local certificate files removed"
    fi
    
    # TODO: Remove from system trust store (requires careful implementation)
    warn "System trust store cleanup not implemented"
    warn "You may need to manually remove the DevArch CA from your system"
    
    success "SSL certificates removed"
}

# =================================================================
# USAGE AND HELP
# =================================================================
show_usage() {
    echo -e "${CYAN}DevArch SSL Certificate Setup Script${NC}"
    echo
    echo "Usage: $0 <command> [options]"
    echo
    echo -e "${YELLOW}Commands:${NC}"
    echo "  setup                  Generate and install SSL certificates"
    echo "  install                Install existing certificates on system"
    echo "  validate               Validate existing certificates"
    echo "  test                   Test SSL certificate functionality"
    echo "  export <format>        Export certificates (pem|p12|der|jks)"
    echo "  info                   Show certificate information"
    echo "  remove                 Remove all certificates"
    echo
    echo -e "${YELLOW}Options:${NC}"
    echo "  -v, --verbose          Enable verbose output"
    echo "  -f, --force            Skip confirmation prompts"
    echo "  -r, --regenerate       Force regeneration of certificates"
    echo "  --skip-trust           Skip system trust installation"
    echo "  -h, --help             Show this help message"
    echo
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 setup               # Generate and install certificates"
    echo "  $0 validate            # Check if certificates are valid"
    echo "  $0 export p12          # Export as PKCS#12 format"
    echo "  $0 test                # Test SSL functionality"
    echo
    echo -e "${YELLOW}Certificate Files:${NC}"
    echo "  CA Certificate:        $SSL_DIR/devarch-ca.crt"
    echo "  Wildcard Certificate:  $SSL_DIR/wildcard.test.crt"
    echo "  Private Key:           $SSL_DIR/wildcard.test.key"
    echo "  Full Chain:            $SSL_DIR/fullchain.pem"
}

# =================================================================
# COMMAND PARSING
# =================================================================
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -r|--regenerate)
                REGENERATE=true
                shift
                ;;
            --skip-trust)
                SKIP_TRUST=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
}

# =================================================================
# MAIN FUNCTION
# =================================================================
main() {
    # Parse global options first
    parse_arguments "$@"
    
    # Remove parsed options from arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose|-f|--force|-r|--regenerate|--skip-trust|-h|--help)
                shift
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Ensure we have a command
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi
    
    local command="$1"
    shift
    
    # Check prerequisites
    if ! command_exists openssl; then
        error "OpenSSL is required but not installed"
        exit 1
    fi
    
    # Execute command
    case "$command" in
        setup)
            setup_certificates
            ;;
        install)
            install_certificates
            ;;
        validate)
            validate_certificates
            ;;
        test)
            test_certificates
            ;;
        export)
            if [[ $# -eq 0 ]]; then
                error "Export format required (pem|p12|der|jks)"
                exit 1
            fi
            export_certificates "$1"
            ;;
        info)
            show_certificate_info
            ;;
        remove)
            remove_certificates
            ;;
        *)
            error "Unknown command: $command"
            echo
            show_usage
            exit 1
            ;;
    esac
}

# =================================================================
# ERROR HANDLING
# =================================================================
handle_error() {
    local exit_code=$?
    error "SSL setup failed with exit code $exit_code"
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo "Stack trace:"
        local frame=0
        while caller $frame; do
            ((frame++))
        done
    fi
    
    # Cleanup on error
    if [[ -d "$SSL_DIR" && "$command" == "setup" ]]; then
        warn "Cleaning up incomplete certificate generation..."
        rm -f "$SSL_DIR"/*.{key,crt,csr,pem} 2>/dev/null || true
    fi
    
    exit $exit_code
}

# Set up error handling
trap handle_error ERR

# =================================================================
# SCRIPT EXECUTION
# =================================================================
# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi