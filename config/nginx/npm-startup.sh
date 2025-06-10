#!/bin/bash
# =================================================================
# DEVARCH NGINX PROXY MANAGER STARTUP SCRIPT
# =================================================================
# Custom startup script that configures NPM for DevArch routing
# and automatically sets up proxy hosts for development tools.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[NPM] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[NPM] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[NPM] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[NPM] INFO: $1${NC}"
}

# =================================================================
# CONFIGURATION VARIABLES
# =================================================================
DEVARCH_DOMAINS=(
    "welcome.test"
    "mailpit.test"
    "adminer.test"
    "phpmyadmin.test"
    "metabase.test"
    "redis.test"
    "swagger.test"
    "nginx.test"
)

# Wait for services to be available
wait_for_service() {
    local service_host="$1"
    local service_port="$2"
    local max_attempts=30
    local attempt=1
    
    info "Waiting for $service_host:$service_port to be available..."
    
    while ! nc -z "$service_host" "$service_port" 2>/dev/null; do
        if [[ $attempt -ge $max_attempts ]]; then
            warn "Service $service_host:$service_port not available after $max_attempts attempts"
            return 1
        fi
        
        sleep 2
        ((attempt++))
    done
    
    log "Service $service_host:$service_port is available"
    return 0
}

# Setup SSL certificates
setup_ssl_certificates() {
    log "Setting up SSL certificates for DevArch domains..."
    
    # Create certificate directory if it doesn't exist
    mkdir -p /etc/letsencrypt/live/wildcard.test
    
    # Check if certificates already exist
    if [[ -f "/etc/letsencrypt/live/wildcard.test/fullchain.pem" && 
          -f "/etc/letsencrypt/live/wildcard.test/privkey.pem" ]]; then
        info "SSL certificates already exist"
        return 0
    fi
    
    # Generate wildcard certificate for *.test domains
    info "Generating wildcard SSL certificate for *.test domains..."
    
    # Create SSL configuration
    cat > /tmp/ssl.conf << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
C = CA
ST = Ontario
L = Kanata
O = DevArch
OU = Development
CN = *.test
emailAddress = admin@devarch.test

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = *.test
DNS.2 = test
DNS.3 = localhost
DNS.4 = welcome.test
DNS.5 = mailpit.test
DNS.6 = adminer.test
DNS.7 = phpmyadmin.test
DNS.8 = metabase.test
DNS.9 = redis.test
DNS.10 = swagger.test
DNS.11 = nginx.test
EOF
    
    # Generate private key
    openssl genrsa -out /etc/letsencrypt/live/wildcard.test/privkey.pem 2048
    
    # Generate certificate
    openssl req -new -x509 \
        -key /etc/letsencrypt/live/wildcard.test/privkey.pem \
        -out /etc/letsencrypt/live/wildcard.test/fullchain.pem \
        -days 365 \
        -config /tmp/ssl.conf \
        -extensions v3_req
    
    # Set proper permissions
    chmod 600 /etc/letsencrypt/live/wildcard.test/privkey.pem
    chmod 644 /etc/letsencrypt/live/wildcard.test/fullchain.pem
    
    # Clean up
    rm -f /tmp/ssl.conf
    
    log "SSL certificates generated successfully"
}

# Configure proxy hosts for DevArch services
configure_proxy_hosts() {
    log "Configuring proxy hosts for DevArch services..."
    
    # Wait for database to be ready
    sleep 10
    
    # Default proxy host configurations will be loaded from mounted files
    # This function can be extended to programmatically create proxy hosts
    # via the NPM API if needed
    
    info "Proxy host configurations loaded from mounted files"
}

# Setup default admin user if not exists
setup_default_admin() {
    log "Setting up default admin user..."
    
    # This will be handled by NPM's default initialization
    # Admin credentials from environment variables:
    # - Default email: admin@example.com
    # - Default password: changeme
    
    info "Default admin user will be created by NPM on first startup"
    warn "Please change default admin credentials after first login"
}

# Configure DevArch-specific settings
configure_devarch_settings() {
    log "Applying DevArch-specific configurations..."
    
    # Enable custom Nginx configurations
    if [[ -f "/data/nginx/custom/nginx-custom.conf" ]]; then
        # Link or copy custom configuration
        ln -sf /data/nginx/custom/nginx-custom.conf /etc/nginx/conf.d/devarch.conf 2>/dev/null || true
        info "DevArch custom Nginx configuration enabled"
    fi
    
    # Create log directory
    mkdir -p /data/logs
    chmod 755 /data/logs
    
    # Set up log rotation
    cat > /etc/logrotate.d/devarch << EOF
/data/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        /usr/sbin/nginx -s reload > /dev/null 2>&1 || true
    endscript
}
EOF
    
    log "DevArch-specific configurations applied"
}

# Health check function
health_check() {
    local retries=5
    local retry=0
    
    while [[ $retry -lt $retries ]]; do
        if curl -f http://localhost:80 >/dev/null 2>&1; then
            return 0
        fi
        
        ((retry++))
        sleep 2
    done
    
    return 1
}

# Main startup sequence
main() {
    log "Starting DevArch Nginx Proxy Manager..."
    
    # Setup SSL certificates first
    setup_ssl_certificates
    
    # Configure DevArch settings
    configure_devarch_settings
    
    # Start NPM in background
    log "Starting Nginx Proxy Manager..."
    /usr/bin/node --max_old_space_size=250 --abort_on_uncaught_exception /app/index.js &
    NPM_PID=$!
    
    # Wait for NPM to start
    sleep 15
    
    # Configure proxy hosts
    configure_proxy_hosts
    
    # Setup default admin
    setup_default_admin
    
    # Perform health check
    if health_check; then
        log "DevArch Nginx Proxy Manager started successfully!"
        log "Access the admin interface at http://localhost:81"
        log "Default credentials: admin@example.com / changeme"
        warn "Remember to change the default password!"
    else
        error "Health check failed!"
        exit 1
    fi
    
    # Monitor NPM process
    wait $NPM_PID
}

# Error handling
trap 'error "Startup failed"; exit 1' ERR

# Run main function
main "$@"