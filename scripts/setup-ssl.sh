#!/bin/zsh
# Enhanced SSL Setup Script for Microservices Architecture
# Generates and manages SSL certificates for local development and production

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# =============================================================================
# SCRIPT-SPECIFIC CONFIGURATION
# =============================================================================

# SSL configuration options
CERT_TYPE="wildcard"           # wildcard, individual, letsencrypt
CERT_DOMAIN="*.test"
CERT_COUNTRY="US"
CERT_STATE="Development" 
CERT_CITY="Local"
CERT_ORG="Microservices"
CERT_UNIT="Development"
CERT_VALIDITY_DAYS=3650
KEY_SIZE=4096
FORCE_REGENERATE=false
BACKUP_EXISTING=true
SETUP_NGINX_CONFIG=true
VALIDATE_CERTS=true

# Certificate paths
CERT_BASE_DIR="/etc/letsencrypt/live"
WILDCARD_CERT_DIR="$CERT_BASE_DIR/wildcard.test"
CUSTOM_CERT_DIR="$PROJECT_ROOT/ssl"

# Service domains that need SSL
declare -a SERVICE_DOMAINS=(
    "nginx.test"
    "adminer.test" 
    "phpmyadmin.test"
    "mongodb.test"
    "metabase.test"
    "nocodb.test"
    "grafana.test"
    "prometheus.test"
    "matomo.test"
    "n8n.test"
    "langflow.test"
    "kibana.test"
    "elasticsearch.test"
    "keycloak.test"
    "mailpit.test"
    "gitea.test"
    "odoo.test"
)

# =============================================================================
# HELP FUNCTION
# =============================================================================

show_help() {
    cat << EOF
Enhanced SSL Certificate Setup Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -t TYPE         Certificate type: wildcard, individual, letsencrypt (default: wildcard)
    -D DOMAIN       Domain for certificate (default: *.test)
    -C COUNTRY      Country code (default: US)
    -S STATE        State/Province (default: Development)
    -L CITY         City/Location (default: Local)
    -O ORG          Organization (default: Microservices)
    -U UNIT         Organizational Unit (default: Development)
    -V DAYS         Certificate validity in days (default: 3650)
    -K SIZE         RSA key size (default: 4096)
    -f              Force regenerate existing certificates
    -B              Skip backup of existing certificates
    -N              Skip Nginx configuration setup
    -T              Skip certificate validation
    -h              Show this help message

CERTIFICATE TYPES:
    wildcard        Single certificate for *.test (recommended for development)
    individual      Separate certificates for each service
    letsencrypt     Use Let's Encrypt (requires valid domain and internet)

EXAMPLES:
    $0                              # Generate wildcard certificate for *.test
    $0 -t individual -v             # Generate individual certs, verbose output
    $0 -D "*.mycompany.local" -f    # Custom domain, force regenerate
    $0 -t letsencrypt -D "api.mycompany.com"  # Let's Encrypt certificate
    $0 -V 365 -K 2048              # 1 year validity, 2048-bit key

GENERATED CERTIFICATES:
    The script creates certificates for all microservices:
$(printf "    %s\n" "${SERVICE_DOMAINS[@]}")

EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_ssl_args() {
    local OPTIND=1
    
    while getopts "sevdr:t:D:C:S:L:O:U:V:K:fBNTh" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            t) CERT_TYPE="$OPTARG" ;;
            D) CERT_DOMAIN="$OPTARG" ;;
            C) CERT_COUNTRY="$OPTARG" ;;
            S) CERT_STATE="$OPTARG" ;;
            L) CERT_CITY="$OPTARG" ;;
            O) CERT_ORG="$OPTARG" ;;
            U) CERT_UNIT="$OPTARG" ;;
            V) CERT_VALIDITY_DAYS="$OPTARG" ;;
            K) KEY_SIZE="$OPTARG" ;;
            f) FORCE_REGENERATE=true ;;
            B) BACKUP_EXISTING=false ;;
            N) SETUP_NGINX_CONFIG=false ;;
            T) VALIDATE_CERTS=false ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
    
    # Validate certificate type
    case "$CERT_TYPE" in
        "wildcard"|"individual"|"letsencrypt") ;;
        *) 
            log "ERROR" "Invalid certificate type: $CERT_TYPE"
            log "INFO" "Valid types: wildcard, individual, letsencrypt"
            exit 1
            ;;
    esac
    
    # Validate key size
    case "$KEY_SIZE" in
        2048|4096|8192) ;;
        *)
            log "ERROR" "Invalid key size: $KEY_SIZE"
            log "INFO" "Valid sizes: 2048, 4096, 8192"
            exit 1
            ;;
    esac
}

# =============================================================================
# PREREQUISITE CHECKS
# =============================================================================

check_prerequisites() {
    log "INFO" "🔍 Checking SSL setup prerequisites..."
    
    # Check if nginx-proxy-manager container exists
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists nginx-proxy-manager 2>/dev/null; then
        log "ERROR" "nginx-proxy-manager container not found. Please start it first."
        exit 1
    fi
    
    # Check if container is running
    local npm_status=$(${SUDO_PREFIX}${CONTAINER_RUNTIME} container inspect nginx-proxy-manager --format '{{.State.Status}}' 2>/dev/null)
    if [ "$npm_status" != "running" ]; then
        log "ERROR" "nginx-proxy-manager container is not running (status: $npm_status)"
        exit 1
    fi
    
    # Check for openssl in container
    if ! execute_command "Check OpenSSL availability" \
         "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager which openssl" false false; then
        log "ERROR" "OpenSSL not found in nginx-proxy-manager container"
        exit 1
    fi
    
    # Create SSL directories
    if [ "$DRY_RUN" = false ]; then
        execute_command "Create certificate directories" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager mkdir -p $CERT_BASE_DIR $WILDCARD_CERT_DIR" \
            true true
    fi
    
    # Create local SSL directory if needed
    if [ ! -d "$CUSTOM_CERT_DIR" ]; then
        log "INFO" "Creating local SSL directory: $CUSTOM_CERT_DIR"
        mkdir -p "$CUSTOM_CERT_DIR" || {
            log "ERROR" "Failed to create SSL directory: $CUSTOM_CERT_DIR"
            exit 1
        }
    fi
    
    log "INFO" "✅ Prerequisites check completed"
}

# =============================================================================
# BACKUP FUNCTIONS
# =============================================================================

backup_existing_certificates() {
    if [ "$BACKUP_EXISTING" = false ]; then
        return 0
    fi
    
    log "INFO" "💾 Backing up existing certificates..."
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="$PROJECT_ROOT/ssl-backups/backup_$timestamp"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would create backup in: $backup_dir"
        return 0
    fi
    
    mkdir -p "$backup_dir"
    
    # Backup from container
    if execute_command "Check for existing certificates in container" \
       "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager test -d $CERT_BASE_DIR" false false; then
        
        execute_command "Backup container certificates" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$CERT_BASE_DIR/. $backup_dir/container/" \
            "$SHOW_ERRORS" false
    fi
    
    # Backup local certificates
    if [ -d "$CUSTOM_CERT_DIR" ] && [ "$(ls -A $CUSTOM_CERT_DIR 2>/dev/null)" ]; then
        execute_command "Backup local certificates" \
            "cp -r $CUSTOM_CERT_DIR/* $backup_dir/local/" \
            "$SHOW_ERRORS" false
    fi
    
    log "INFO" "✅ Certificate backup completed: $backup_dir"
}

# =============================================================================
# CERTIFICATE GENERATION FUNCTIONS
# =============================================================================

generate_wildcard_certificate() {
    log "INFO" "🔐 Generating wildcard certificate for: $CERT_DOMAIN"
    
    local cert_path="$WILDCARD_CERT_DIR/fullchain.pem"
    local key_path="$WILDCARD_CERT_DIR/privkey.pem"
    
    # Check if certificate already exists
    if [ "$FORCE_REGENERATE" = false ]; then
        if execute_command "Check existing wildcard certificate" \
           "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager test -f $cert_path" false false; then
            log "INFO" "Wildcard certificate already exists. Use -f to force regeneration."
            return 0
        fi
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would generate wildcard certificate: $cert_path"
        return 0
    fi
    
    # Generate Subject Alternative Names for all service domains
    local san_list=""
    for domain in "${SERVICE_DOMAINS[@]}"; do
        san_list="$san_list,DNS:$domain"
    done
    # Remove leading comma
    san_list=${san_list#,}
    
    # Generate the certificate
    local openssl_cmd="openssl req -x509 -nodes -days $CERT_VALIDITY_DAYS -newkey rsa:$KEY_SIZE \
        -keyout $key_path \
        -out $cert_path \
        -subj \"/C=$CERT_COUNTRY/ST=$CERT_STATE/L=$CERT_CITY/O=$CERT_ORG/OU=$CERT_UNIT/CN=$CERT_DOMAIN\" \
        -addext \"subjectAltName=$san_list\""
    
    execute_command "Generate wildcard SSL certificate" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager bash -c \"$openssl_cmd\"" \
        true true
    
    # Set proper permissions
    execute_command "Set certificate permissions" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager chmod 644 $cert_path" \
        true true
    
    execute_command "Set private key permissions" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager chmod 600 $key_path" \
        true true
    
    log "INFO" "✅ Wildcard certificate generated successfully"
}

generate_individual_certificates() {
    log "INFO" "🔐 Generating individual certificates for each service..."
    
    for domain in "${SERVICE_DOMAINS[@]}"; do
        log "INFO" "  📄 Generating certificate for: $domain"
        
        local domain_dir="$CERT_BASE_DIR/$domain"
        local cert_path="$domain_dir/fullchain.pem"
        local key_path="$domain_dir/privkey.pem"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "  [DRY RUN] Would generate certificate: $cert_path"
            continue
        fi
        
        # Create domain directory
        execute_command "Create directory for $domain" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager mkdir -p $domain_dir" \
            true true
        
        # Check if certificate already exists
        if [ "$FORCE_REGENERATE" = false ]; then
            if execute_command "Check existing certificate for $domain" \
               "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager test -f $cert_path" false false; then
                log "INFO" "  Certificate for $domain already exists, skipping"
                continue
            fi
        fi
        
        # Generate the certificate
        local openssl_cmd="openssl req -x509 -nodes -days $CERT_VALIDITY_DAYS -newkey rsa:$KEY_SIZE \
            -keyout $key_path \
            -out $cert_path \
            -subj \"/C=$CERT_COUNTRY/ST=$CERT_STATE/L=$CERT_CITY/O=$CERT_ORG/OU=$CERT_UNIT/CN=$domain\" \
            -addext \"subjectAltName=DNS:$domain\""
        
        execute_command "Generate SSL certificate for $domain" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager bash -c \"$openssl_cmd\"" \
            "$SHOW_ERRORS" false
        
        # Set proper permissions
        execute_command "Set permissions for $domain certificate" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager chmod 644 $cert_path && ${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager chmod 600 $key_path" \
            "$SHOW_ERRORS" false
        
        sleep 1  # Brief pause between certificates
    done
    
    log "INFO" "✅ Individual certificates generated successfully"
}

setup_letsencrypt_certificate() {
    log "INFO" "🔐 Setting up Let's Encrypt certificate for: $CERT_DOMAIN"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup Let's Encrypt certificate for: $CERT_DOMAIN"
        return 0
    fi
    
    # Check if certbot is available
    if ! execute_command "Check for certbot" \
         "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager which certbot" false false; then
        
        log "INFO" "Installing certbot in nginx-proxy-manager container..."
        execute_command "Install certbot" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager apt-get update && ${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager apt-get install -y certbot" \
            true true
    fi
    
    # Generate Let's Encrypt certificate
    # Note: This requires the domain to be publicly accessible
    local certbot_cmd="certbot certonly --standalone \
        --agree-tos \
        --no-eff-email \
        --email admin@${CERT_DOMAIN#*.} \
        -d ${CERT_DOMAIN#*} \
        --cert-path $CERT_BASE_DIR/${CERT_DOMAIN#*.}/fullchain.pem \
        --key-path $CERT_BASE_DIR/${CERT_DOMAIN#*.}/privkey.pem"
    
    execute_command "Generate Let's Encrypt certificate" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager $certbot_cmd" \
        true false
    
    log "INFO" "✅ Let's Encrypt certificate setup completed"
}

# =============================================================================
# NGINX CONFIGURATION FUNCTIONS
# =============================================================================

setup_nginx_ssl_config() {
    if [ "$SETUP_NGINX_CONFIG" = false ]; then
        log "INFO" "Skipping Nginx SSL configuration as requested"
        return 0
    fi
    
    log "INFO" "🔧 Setting up Nginx SSL configuration..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup Nginx SSL configuration"
        return 0
    fi
    
    # Create SSL configuration snippet
    local ssl_config="/data/nginx/custom/ssl-params.conf"
    local ssl_config_content="# SSL Configuration for Microservices
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
ssl_session_tickets off;
ssl_stapling on;
ssl_stapling_verify on;

# Security headers
add_header Strict-Transport-Security \"max-age=31536000; includeSubDomains\" always;
add_header X-Content-Type-Options nosniff always;
add_header X-Frame-Options DENY always;
add_header X-XSS-Protection \"1; mode=block\" always;
add_header Referrer-Policy \"no-referrer-when-downgrade\" always;
"
    
    execute_command "Create SSL configuration snippet" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager bash -c \"echo '$ssl_config_content' > $ssl_config\"" \
        true true
    
    log "INFO" "✅ Nginx SSL configuration completed"
}

# =============================================================================
# CERTIFICATE VALIDATION FUNCTIONS
# =============================================================================

validate_certificates() {
    if [ "$VALIDATE_CERTS" = false ]; then
        log "INFO" "Skipping certificate validation as requested"
        return 0
    fi
    
    log "INFO" "🔍 Validating generated certificates..."
    
    case "$CERT_TYPE" in
        "wildcard")
            validate_wildcard_certificate
            ;;
        "individual")
            validate_individual_certificates
            ;;
        "letsencrypt")
            validate_letsencrypt_certificate
            ;;
    esac
}

validate_wildcard_certificate() {
    local cert_path="$WILDCARD_CERT_DIR/fullchain.pem"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would validate wildcard certificate"
        return 0
    fi
    
    # Check certificate exists
    if ! execute_command "Check wildcard certificate exists" \
         "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager test -f $cert_path" false false; then
        log "ERROR" "Wildcard certificate not found: $cert_path"
        return 1
    fi
    
    # Validate certificate details
    log "INFO" "📋 Certificate details:"
    execute_command "Show certificate subject" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -subject" \
        true false
    
    execute_command "Show certificate dates" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -dates" \
        true false
    
    execute_command "Show certificate SAN" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -ext subjectAltName" \
        true false
    
    # Test certificate validity
    if execute_command "Verify certificate" \
       "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -checkend 86400" false false; then
        log "INFO" "✅ Certificate is valid and not expiring within 24 hours"
    else
        log "WARN" "⚠️ Certificate validation failed or expiring soon"
    fi
}

validate_individual_certificates() {
    local failed_count=0
    
    for domain in "${SERVICE_DOMAINS[@]}"; do
        local cert_path="$CERT_BASE_DIR/$domain/fullchain.pem"
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "[DRY RUN] Would validate certificate for $domain"
            continue
        fi
        
        if execute_command "Validate certificate for $domain" \
           "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -checkend 86400" false false; then
            log "INFO" "  ✅ $domain certificate is valid"
        else
            log "WARN" "  ⚠️ $domain certificate validation failed"
            failed_count=$((failed_count + 1))
        fi
    done
    
    if [ $failed_count -gt 0 ]; then
        log "WARN" "$failed_count certificate(s) failed validation"
    else
        log "INFO" "✅ All individual certificates validated successfully"
    fi
}

validate_letsencrypt_certificate() {
    local cert_path="$CERT_BASE_DIR/${CERT_DOMAIN#*.}/fullchain.pem"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would validate Let's Encrypt certificate"
        return 0
    fi
    
    if execute_command "Validate Let's Encrypt certificate" \
       "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager openssl x509 -in $cert_path -noout -checkend 86400" false false; then
        log "INFO" "✅ Let's Encrypt certificate is valid"
    else
        log "WARN" "⚠️ Let's Encrypt certificate validation failed"
    fi
}

# =============================================================================
# EXPORT FUNCTIONS
# =============================================================================

export_certificates_to_host() {
    log "INFO" "📤 Exporting certificates to host system..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would export certificates to: $CUSTOM_CERT_DIR"
        return 0
    fi
    
    case "$CERT_TYPE" in
        "wildcard")
            execute_command "Export wildcard certificate" \
                "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$WILDCARD_CERT_DIR/fullchain.pem $CUSTOM_CERT_DIR/wildcard.test.crt" \
                "$SHOW_ERRORS" false
            
            execute_command "Export wildcard private key" \
                "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$WILDCARD_CERT_DIR/privkey.pem $CUSTOM_CERT_DIR/wildcard.test.key" \
                "$SHOW_ERRORS" false
            ;;
        "individual")
            for domain in "${SERVICE_DOMAINS[@]}"; do
                local domain_cert_dir="$CUSTOM_CERT_DIR/$domain"
                mkdir -p "$domain_cert_dir"
                
                execute_command "Export certificate for $domain" \
                    "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$CERT_BASE_DIR/$domain/fullchain.pem $domain_cert_dir/fullchain.pem" \
                    "$SHOW_ERRORS" false
                
                execute_command "Export private key for $domain" \
                    "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$CERT_BASE_DIR/$domain/privkey.pem $domain_cert_dir/privkey.pem" \
                    "$SHOW_ERRORS" false
            done
            ;;
    esac
    
    log "INFO" "✅ Certificates exported to: $CUSTOM_CERT_DIR"
}

# =============================================================================
# RESTART SERVICES
# =============================================================================

restart_nginx_proxy_manager() {
    log "INFO" "🔄 Restarting nginx-proxy-manager to apply SSL configuration..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would restart nginx-proxy-manager"
        return 0
    fi
    
    execute_command "Restart nginx-proxy-manager" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} restart nginx-proxy-manager" \
        true true
    
    # Wait for service to be ready
    log "INFO" "⏳ Waiting for nginx-proxy-manager to start..."
    sleep 10
    
    # Test if service is responding
    local max_attempts=12
    for ((i=1; i<=max_attempts; i++)); do
        if execute_command "Test nginx-proxy-manager readiness" \
           "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager curl -k -s https://localhost:443" false false; then
            log "INFO" "✅ nginx-proxy-manager is ready"
            return 0
        fi
        
        log "DEBUG" "nginx-proxy-manager not ready, attempt $i/$max_attempts"
        sleep 5
    done
    
    log "WARN" "⚠️ nginx-proxy-manager may not be fully ready yet"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

show_ssl_summary() {
    cat << EOF

📋 SSL CERTIFICATE SETUP SUMMARY
================================

Certificate Type: $CERT_TYPE
Domain: $CERT_DOMAIN
Country: $CERT_COUNTRY
State/Province: $CERT_STATE
City: $CERT_CITY
Organization: $CERT_ORG
Organizational Unit: $CERT_UNIT
Validity Period: $CERT_VALIDITY_DAYS days
RSA Key Size: $KEY_SIZE bits

Options:
- Force Regenerate: $FORCE_REGENERATE
- Backup Existing: $BACKUP_EXISTING
- Setup Nginx Config: $SETUP_NGINX_CONFIG
- Validate Certificates: $VALIDATE_CERTS

Services Covered: ${#SERVICE_DOMAINS[@]} domains
$(printf "  %s\n" "${SERVICE_DOMAINS[@]}")

EOF
}

main() {
    parse_common_args "$@"
    parse_ssl_args "$@"
    
    show_ssl_summary
    
    log "INFO" "🚀 Starting SSL certificate setup..."
    
    # Prerequisites check
    check_prerequisites
    
    # Backup existing certificates
    backup_existing_certificates
    
    # Generate certificates based on type
    case "$CERT_TYPE" in
        "wildcard")
            generate_wildcard_certificate
            ;;
        "individual")
            generate_individual_certificates
            ;;
        "letsencrypt")
            setup_letsencrypt_certificate
            ;;
    esac
    
    # Setup Nginx SSL configuration
    setup_nginx_ssl_config
    
    # Validate certificates
    validate_certificates
    
    # Export certificates to host
    export_certificates_to_host
    
    # Restart nginx-proxy-manager
    restart_nginx_proxy_manager
    
    log "INFO" "✅ SSL certificate setup completed successfully!"
    
    if [ "$DRY_RUN" = false ]; then
        cat << EOF

🎉 SSL SETUP COMPLETE!
=====================

Your SSL certificates are now ready for use.

📋 Next Steps:
1. Run the trust-host.sh script to install certificates in your system trust store
2. Configure proxy hosts in Nginx Proxy Manager (http://localhost:81)
3. Test HTTPS access to your services

🔍 Certificate Locations:
- Container: $CERT_BASE_DIR/
- Host: $CUSTOM_CERT_DIR/

🌐 Test URLs:
$(printf "  https://%s\n" "${SERVICE_DOMAINS[@]}")

EOF
    fi
}

main "$@"