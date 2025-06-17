#!/bin/zsh

# =============================================================================
# SSL CERTIFICATE SETUP SCRIPT
# =============================================================================
# Generates and configures SSL certificates for all microservices
# Cross-platform support for Linux, macOS, Windows, and WSL

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_force_regenerate=false
opt_custom_domain="$SSL_DOMAIN"
opt_cert_days="$SSL_DAYS_VALID"
opt_key_size=4096
opt_export_formats=false
opt_verify_only=false

# Certificate paths
CERT_CONTAINER_DIR="$SSL_CERT_DIR"
CERT_TEMP_DIR="/tmp/ssl-microservices-$$"

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Generates wildcard SSL certificates for the microservices architecture.
    Creates certificates that work across all operating systems and browsers.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    -f, --force             Force regeneration of existing certificates
    -d, --domain DOMAIN     Custom domain (default: *.test)
    -k, --key-size SIZE     RSA key size in bits (default: 4096)
    -t, --days DAYS         Certificate validity in days (default: 3650)
    -x, --export            Export certificates in multiple formats
    -v, --verify            Only verify existing certificates
    -h, --help              Show this help message

SUPPORTED FORMATS:
    - PEM (default): fullchain.pem, privkey.pem
    - PFX/P12: For Windows IIS
    - DER: Binary format
    - CRT: Certificate only

EXAMPLES:
    $0                                  # Generate default wildcard certificate
    $0 -f                              # Force regenerate certificates
    $0 -d "*.local" -t 1825            # Custom domain, 5-year validity
    $0 -x                              # Generate and export all formats
    $0 -v                              # Verify existing certificates only
    $0 -s -e -f                        # Use sudo, show errors, force regen

CERTIFICATE LOCATIONS:
    Container: $CERT_CONTAINER_DIR/
    Exported:  $PROJECT_ROOT/ssl/

NOTES:
    - Certificates work with Chrome, Firefox, Safari, and Edge
    - Wildcard certificates cover all *.test subdomains
    - Generated certificates are suitable for development use
    - For production, consider using Let's Encrypt or commercial CA
EOF
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
            -k|--key-size)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_key_size="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric key size"
                fi
                ;;
            -t|--days)
                if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                    opt_cert_days="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a numeric days value"
                fi
                ;;
            -x|--export)
                opt_export_formats=true
                shift
                ;;
            -v|--verify)
                opt_verify_only=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1. Use -h for help."
                ;;
        esac
    done
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_nginx_container() {
    print_status "step" "Validating nginx-proxy-manager container..."
    
    if ! eval "$CONTAINER_CMD container exists nginx-proxy-manager $ERROR_REDIRECT"; then
        handle_error "nginx-proxy-manager container not found. Please start proxy services first."
    fi
    
    # Check if container is running
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' nginx-proxy-manager 2>/dev/null" || echo "unknown")
    
    if [[ "$container_status" != "running" ]]; then
        handle_error "nginx-proxy-manager container is not running. Current status: $container_status"
    fi
    
    print_status "success" "nginx-proxy-manager container is running"
}

validate_domain_format() {
    local domain="$opt_custom_domain"
    print_status "info" "Using domain: $domain"
}

check_existing_certificates() {
    if [[ "$opt_verify_only" == "true" ]]; then
        return 0
    fi
    
    local cert_exists=false
    
    if eval "$CONTAINER_CMD exec nginx-proxy-manager test -f $CERT_CONTAINER_DIR/fullchain.pem $ERROR_REDIRECT"; then
        cert_exists=true
        print_status "info" "Existing certificate found"
        
        if [[ "$opt_force_regenerate" == "false" ]]; then
            print_status "warning" "Certificate already exists. Use -f to force regeneration."
            verify_certificates
            exit 0
        else
            print_status "info" "Force regeneration requested"
        fi
    fi
}

# =============================================================================
# CERTIFICATE GENERATION FUNCTIONS
# =============================================================================

prepare_certificate_environment() {
    print_status "step" "Preparing certificate environment..."
    
    # Create certificate directory in container
    eval "$CONTAINER_CMD exec nginx-proxy-manager mkdir -p $CERT_CONTAINER_DIR $ERROR_REDIRECT"
    check_status "Failed to create certificate directory in container"
    
    # Create temporary directory on host
    mkdir -p "$CERT_TEMP_DIR"
    
    # Create export directory if needed
    if [[ "$opt_export_formats" == "true" ]]; then
        mkdir -p "$PROJECT_ROOT/ssl"
    fi
    
    print_status "success" "Certificate environment prepared"
}

generate_openssl_config() {
    local config_file="$CERT_TEMP_DIR/openssl.conf"
    local domain="$opt_custom_domain"
    local base_domain="${domain#\*.}"  # Remove *. prefix if present
    
    cat > "$config_file" << EOF
[req]
default_bits = $opt_key_size
prompt = no
default_md = sha256
req_extensions = v3_req
distinguished_name = dn

[dn]
C = US
ST = Development
L = Local
O = Microservices Development
OU = IT Department
CN = $domain

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = $domain
DNS.2 = $base_domain
DNS.3 = localhost
DNS.4 = *.localhost
EOF

    # Add common service subdomains
    local -a service_domains=(
        "nginx.$base_domain"
        "adminer.$base_domain"
        "phpmyadmin.$base_domain"
        "mongodb.$base_domain"
        "metabase.$base_domain"
        "nocodb.$base_domain"
        "pgadmin.$base_domain"
        "redis.$base_domain"
        "grafana.$base_domain"
        "prometheus.$base_domain"
        "matomo.$base_domain"
        "n8n.$base_domain"
        "langflow.$base_domain"
        "kibana.$base_domain"
        "elasticsearch.$base_domain"
        "mailpit.$base_domain"
        "gitea.$base_domain"
    )
    
    local dns_counter=5
    for service_domain in "${service_domains[@]}"; do
        echo "DNS.$dns_counter = $service_domain" >> "$config_file"
        ((dns_counter++))
    done
    
    print_status "success" "OpenSSL configuration generated with ${#service_domains[@]} service domains"
}

generate_certificate() {
    print_status "step" "Generating SSL certificate..."
    
    # First generate the config inside the container
    local container_config="/tmp/openssl.conf"
    local container_cert="$CERT_CONTAINER_DIR/fullchain.pem"
    local container_key="$CERT_CONTAINER_DIR/privkey.pem"
    
    # Copy our config to the container
    eval "$CONTAINER_CMD cp '$CERT_TEMP_DIR/openssl.conf' nginx-proxy-manager:/tmp/openssl.conf $ERROR_REDIRECT"
    check_status "Failed to copy OpenSSL config to container"
    
    # Generate certificate directly in the container
    local openssl_cmd="openssl req -x509 -nodes -days $opt_cert_days -newkey rsa:$opt_key_size"
    openssl_cmd="$openssl_cmd -keyout '$container_key' -out '$container_cert'"
    openssl_cmd="$openssl_cmd -config '$container_config' -extensions v3_req"
    
    if eval "$CONTAINER_CMD exec nginx-proxy-manager $openssl_cmd $ERROR_REDIRECT"; then
        print_status "success" "SSL certificate generated successfully"
    else
        print_status "error" "Failed to generate SSL certificate. Trying alternative method..."
        
        # Alternative: Generate on host and copy to container
        generate_certificate_on_host
    fi
    
    # Set proper permissions
    eval "$CONTAINER_CMD exec nginx-proxy-manager chmod 644 $container_cert $ERROR_REDIRECT"
    eval "$CONTAINER_CMD exec nginx-proxy-manager chmod 600 $container_key $ERROR_REDIRECT"
}

generate_certificate_on_host() {
    print_status "step" "Generating certificate on host as fallback..."
    
    local key_file="$CERT_TEMP_DIR/privkey.pem"
    local cert_file="$CERT_TEMP_DIR/fullchain.pem"
    local config_file="$CERT_TEMP_DIR/openssl.conf"
    
    # Generate certificate on host
    local openssl_cmd="openssl req -x509 -nodes -days $opt_cert_days -newkey rsa:$opt_key_size"
    openssl_cmd="$openssl_cmd -keyout '$key_file' -out '$cert_file'"
    openssl_cmd="$openssl_cmd -config '$config_file' -extensions v3_req"
    
    if eval "$openssl_cmd $ERROR_REDIRECT"; then
        print_status "success" "Certificate generated on host"
        
        # Copy to container
        eval "$CONTAINER_CMD cp '$cert_file' nginx-proxy-manager:$CERT_CONTAINER_DIR/fullchain.pem $ERROR_REDIRECT"
        check_status "Failed to copy certificate to container"
        
        eval "$CONTAINER_CMD cp '$key_file' nginx-proxy-manager:$CERT_CONTAINER_DIR/privkey.pem $ERROR_REDIRECT"
        check_status "Failed to copy private key to container"
        
        print_status "success" "Certificate copied to container"
    else
        handle_error "Failed to generate SSL certificate"
    fi
}

export_certificate_formats() {
    if [[ "$opt_export_formats" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Exporting certificate in multiple formats..."
    
    local export_dir="$PROJECT_ROOT/ssl"
    local domain_safe="${opt_custom_domain//\*/wildcard}"
    
    # Copy certificates from container to export directory
    eval "$CONTAINER_CMD cp nginx-proxy-manager:$CERT_CONTAINER_DIR/fullchain.pem '$export_dir/${domain_safe}.crt' $ERROR_REDIRECT"
    eval "$CONTAINER_CMD cp nginx-proxy-manager:$CERT_CONTAINER_DIR/privkey.pem '$export_dir/${domain_safe}.key' $ERROR_REDIRECT"
    
    # Generate additional formats
    local cert_file="$export_dir/${domain_safe}.crt"
    local key_file="$export_dir/${domain_safe}.key"
    
    # Export PFX format (for Windows)
    openssl pkcs12 -export -out "$export_dir/${domain_safe}.pfx" \
        -inkey "$key_file" -in "$cert_file" \
        -passout pass:password $ERROR_REDIRECT
    
    # Export DER format
    openssl x509 -outform der -in "$cert_file" -out "$export_dir/${domain_safe}.der" $ERROR_REDIRECT
    
    # Create combined file
    cat "$cert_file" "$key_file" > "$export_dir/${domain_safe}.pem"
    
    # Create info file
    cat > "$export_dir/certificate-info.txt" << EOF
SSL Certificate Information
==========================
Domain: $opt_custom_domain
Generated: $(date)
Validity: $opt_cert_days days
Key Size: $opt_key_size bits

Files:
- ${domain_safe}.crt    - Certificate (PEM format)
- ${domain_safe}.key    - Private Key (PEM format)
- ${domain_safe}.pem    - Combined Certificate + Key
- ${domain_safe}.pfx    - PKCS#12 format (password: password)
- ${domain_safe}.der    - Binary certificate format

Usage:
- Web servers: Use .crt and .key files
- Windows IIS: Use .pfx file
- Java applications: Convert .der to .jks format
EOF
    
    print_status "success" "Certificates exported to $export_dir"
}

restart_nginx_container() {
    print_status "step" "Restarting nginx-proxy-manager to apply certificates..."
    
    eval "$CONTAINER_CMD restart nginx-proxy-manager $ERROR_REDIRECT"
    check_status "Failed to restart nginx-proxy-manager"
    
    # Wait for container to be ready
    sleep 5
    
    # Verify container is running
    local max_wait=30
    local counter=0
    while [[ $counter -lt $max_wait ]]; do
        local container_status
        container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' nginx-proxy-manager 2>/dev/null" || echo "unknown")
        
        if [[ "$container_status" == "running" ]]; then
            print_status "success" "nginx-proxy-manager restarted successfully"
            return 0
        fi
        
        sleep 1
        ((counter++))
    done
    
    handle_error "nginx-proxy-manager failed to start after restart"
}

# =============================================================================
# CERTIFICATE VERIFICATION FUNCTIONS
# =============================================================================

verify_certificates() {
    print_status "step" "Verifying SSL certificates..."
    
    # Check if certificates exist in container
    if ! eval "$CONTAINER_CMD exec nginx-proxy-manager test -f $CERT_CONTAINER_DIR/fullchain.pem $ERROR_REDIRECT"; then
        handle_error "Certificate file not found in container"
    fi
    
    if ! eval "$CONTAINER_CMD exec nginx-proxy-manager test -f $CERT_CONTAINER_DIR/privkey.pem $ERROR_REDIRECT"; then
        handle_error "Private key file not found in container"
    fi
    
    # Copy certificate to temp for verification
    mkdir -p "$CERT_TEMP_DIR"
    eval "$CONTAINER_CMD cp nginx-proxy-manager:$CERT_CONTAINER_DIR/fullchain.pem $CERT_TEMP_DIR/verify.crt $ERROR_REDIRECT"
    
    # Verify certificate details
    local cert_info
    cert_info=$(openssl x509 -in "$CERT_TEMP_DIR/verify.crt" -noout -text 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        print_status "success" "Certificate is valid"
        
        # Extract key information
        local subject
        local validity
        local san_domains
        
        subject=$(echo "$cert_info" | grep "Subject:" | sed 's/.*Subject: //')
        validity=$(echo "$cert_info" | grep "Not After" | sed 's/.*Not After : //')
        san_domains=$(echo "$cert_info" | grep -A1 "Subject Alternative Name" | tail -1 | tr ',' '\n' | wc -l)
        
        print_status "info" "Certificate Details:"
        echo "  Subject: $subject"
        echo "  Valid Until: $validity"
        echo "  SAN Domains: $san_domains"
        
        # Check if certificate covers our domain
        if echo "$cert_info" | grep -q "$opt_custom_domain" $ERROR_REDIRECT; then
            print_status "success" "Certificate covers domain: $opt_custom_domain"
        else
            print_status "warning" "Certificate may not cover domain: $opt_custom_domain"
        fi
        
    else
        handle_error "Certificate verification failed"
    fi
}

test_certificate_installation() {
    print_status "step" "Testing certificate installation..."
    
    # Test if nginx can read the certificates
    if eval "$CONTAINER_CMD exec nginx-proxy-manager openssl x509 -in $CERT_CONTAINER_DIR/fullchain.pem -noout $ERROR_REDIRECT"; then
        print_status "success" "Certificate is readable by nginx"
    else
        print_status "warning" "Certificate may not be readable by nginx"
    fi
    
    # Test private key
    if eval "$CONTAINER_CMD exec nginx-proxy-manager openssl rsa -in $CERT_CONTAINER_DIR/privkey.pem -check -noout $ERROR_REDIRECT"; then
        print_status "success" "Private key is valid"
    else
        print_status "warning" "Private key validation failed"
    fi
    
    # Test key-certificate match
    local cert_hash key_hash
    cert_hash=$(eval "$CONTAINER_CMD exec nginx-proxy-manager openssl x509 -noout -modulus -in $CERT_CONTAINER_DIR/fullchain.pem | openssl md5" 2>/dev/null)
    key_hash=$(eval "$CONTAINER_CMD exec nginx-proxy-manager openssl rsa -noout -modulus -in $CERT_CONTAINER_DIR/privkey.pem | openssl md5" 2>/dev/null)
    
    if [[ "$cert_hash" == "$key_hash" && -n "$cert_hash" ]]; then
        print_status "success" "Certificate and private key match"
    else
        print_status "warning" "Certificate and private key may not match"
    fi
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

cleanup_temp_files() {
    if [[ -d "$CERT_TEMP_DIR" ]]; then
        rm -rf "$CERT_TEMP_DIR"
        print_status "info" "Temporary files cleaned up"
    fi
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_ssl_summary() {
    print_status "info" "SSL Certificate Setup Summary:"
    echo "  Container Runtime: $CONTAINER_RUNTIME"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Show Errors: $opt_show_errors"
    echo "  Domain: $opt_custom_domain"
    echo "  Key Size: $opt_key_size bits"
    echo "  Validity: $opt_cert_days days"
    echo "  Force Regenerate: $opt_force_regenerate"
    echo "  Export Formats: $opt_export_formats"
    echo "  Verify Only: $opt_verify_only"
    echo ""
}

main() {
    # Set up signal handlers for cleanup
    trap cleanup_temp_files EXIT INT TERM
    
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Show setup summary
    show_ssl_summary
    
    # Validate environment
    validate_nginx_container
    validate_domain_format
    
    if [[ "$opt_verify_only" == "true" ]]; then
        verify_certificates
        test_certificate_installation
        print_status "success" "Certificate verification completed!"
        return 0
    fi
    
    # Check existing certificates
    check_existing_certificates
    
    # Generate new certificates
    prepare_certificate_environment
    generate_openssl_config
    generate_certificate
    
    # Export in multiple formats if requested
    export_certificate_formats
    
    # Restart nginx to apply certificates
    restart_nginx_container
    
    # Verify installation
    verify_certificates
    test_certificate_installation
    
    print_status "success" "SSL certificate setup completed successfully!"
    
    # Show next steps
    echo ""
    print_status "info" "Next Steps:"
    echo "  1. Run trust-host.sh to install certificates in your system trust store"
    echo "  2. Test HTTPS access: curl -v https://nginx.test"
    echo "  3. Access services via HTTPS URLs (e.g., https://metabase.test)"
    
    if [[ "$opt_export_formats" == "true" ]]; then
        echo "  4. Check exported certificates in: $PROJECT_ROOT/ssl/"
    fi
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi