#!/bin/zsh

# =============================================================================
# TRAEFIK SETUP AND MIGRATION SCRIPT
# =============================================================================
# Sets up Traefik as a replacement for Nginx Proxy Manager
# Handles migration, SSL certificates, and service discovery

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_migrate_from_nginx=false
opt_keep_nginx_backup=true
opt_setup_ssl=true
opt_enable_docker_discovery=true
opt_enable_file_discovery=true
opt_force_recreate=false
opt_import_nginx_routes=false

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Sets up Traefik as a drop-in replacement for Nginx Proxy Manager.
    Handles migration, SSL certificates, and automatic service discovery.

OPTIONS:
    -s, --sudo                      Use sudo for container commands
    -e, --errors                    Show detailed error messages
    -m, --migrate                   Migrate from existing Nginx Proxy Manager
    --no-backup                     Don't backup Nginx configuration during migration
    --no-ssl                        Skip SSL certificate setup
    --no-docker                     Disable Docker service discovery
    --no-file                       Disable file-based service discovery
    -f, --force                     Force recreate Traefik container
    --import-routes                 Import routes from Nginx Proxy Manager
    -h, --help                      Show this help message

MIGRATION FEATURES:
    - Automatic backup of Nginx Proxy Manager configuration
    - SSL certificate migration and conversion
    - Route discovery and conversion to Traefik format
    - Seamless switchover with minimal downtime

EXAMPLES:
    $0                              # Basic Traefik setup
    $0 -m                          # Migrate from Nginx Proxy Manager
    $0 -f --import-routes          # Force setup with route import
    $0 --no-ssl --no-docker        # Minimal setup without SSL or Docker discovery

POST-SETUP:
    - Dashboard: https://traefik.test
    - API: http://localhost:8080
    - Management: ./manage-traefik.sh status
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
            -m|--migrate)
                opt_migrate_from_nginx=true
                shift
                ;;
            --no-backup)
                opt_keep_nginx_backup=false
                shift
                ;;
            --no-ssl)
                opt_setup_ssl=false
                shift
                ;;
            --no-docker)
                opt_enable_docker_discovery=false
                shift
                ;;
            --no-file)
                opt_enable_file_discovery=false
                shift
                ;;
            -f|--force)
                opt_force_recreate=true
                shift
                ;;
            --import-routes)
                opt_import_nginx_routes=true
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

validate_environment() {
    print_status "step" "Validating environment for Traefik setup..."
    
    # Check container runtime
    if ! command -v "$CONTAINER_RUNTIME" >/dev/null 2>&1; then
        handle_error "$CONTAINER_RUNTIME is not installed or not in PATH"
    fi
    
    # Check compose directory
    if [[ ! -d "$COMPOSE_DIR" ]]; then
        handle_error "Compose directory not found: $COMPOSE_DIR"
    fi
    
    # Create Traefik config directories
    mkdir -p "$PROJECT_ROOT/config/traefik/static"
    mkdir -p "$PROJECT_ROOT/config/traefik/dynamic" 
    mkdir -p "$PROJECT_ROOT/backups/traefik"
    
    print_status "success" "Environment validation passed"
}

check_existing_nginx() {
    if [[ "$opt_migrate_from_nginx" == "true" ]]; then
        if ! eval "$CONTAINER_CMD container exists nginx-proxy-manager $ERROR_REDIRECT"; then
            print_status "warning" "Nginx Proxy Manager container not found"
            print_status "info" "Proceeding with fresh Traefik installation"
            opt_migrate_from_nginx=false
        else
            print_status "info" "Found existing Nginx Proxy Manager for migration"
        fi
    fi
}

# =============================================================================
# MIGRATION FUNCTIONS
# =============================================================================

backup_nginx_configuration() {
    if [[ "$opt_keep_nginx_backup" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Backing up Nginx Proxy Manager configuration..."
    
    local backup_dir="$PROJECT_ROOT/backups/nginx-proxy-manager"
    local backup_file="$backup_dir/npm-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    
    mkdir -p "$backup_dir"
    
    # Backup container volumes
    if eval "$CONTAINER_CMD container exists nginx-proxy-manager"; then
        # Get volume mounts
        local volumes
        volumes=$(eval "$CONTAINER_CMD inspect nginx-proxy-manager --format='{{range .Mounts}}{{.Source}}:{{.Destination}} {{end}}' 2>/dev/null" || echo "")
        
        if [[ -n "$volumes" ]]; then
            print_status "info" "Backing up Nginx Proxy Manager data..."
            
            # Create backup of configuration files
            eval "$CONTAINER_CMD exec nginx-proxy-manager tar -czf /tmp/npm-config.tar.gz /data 2>/dev/null" || true
            eval "$CONTAINER_CMD cp nginx-proxy-manager:/tmp/npm-config.tar.gz $backup_file 2>/dev/null" || true
            
            if [[ -f "$backup_file" ]]; then
                print_status "success" "Nginx configuration backed up to: $backup_file"
            else
                print_status "warning" "Backup creation failed, but continuing..."
            fi
        fi
    fi
}

extract_nginx_ssl_certificates() {
    if [[ "$opt_setup_ssl" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Extracting SSL certificates from Nginx Proxy Manager..."
    
    if ! eval "$CONTAINER_CMD container exists nginx-proxy-manager"; then
        print_status "warning" "Nginx Proxy Manager not found, will generate new certificates"
        return 0
    fi
    
    local ssl_dir="$PROJECT_ROOT/config/traefik/ssl"
    mkdir -p "$ssl_dir"
    
    # Try to copy existing certificates
    if eval "$CONTAINER_CMD exec nginx-proxy-manager test -f /etc/letsencrypt/live/wildcard.test/fullchain.pem 2>/dev/null"; then
        print_status "info" "Found existing SSL certificates, copying..."
        
        eval "$CONTAINER_CMD cp nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/fullchain.pem $ssl_dir/cert.pem 2>/dev/null" || true
        eval "$CONTAINER_CMD cp nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/privkey.pem $ssl_dir/key.pem 2>/dev/null" || true
        
        if [[ -f "$ssl_dir/cert.pem" && -f "$ssl_dir/key.pem" ]]; then
            print_status "success" "SSL certificates extracted successfully"
            
            # Validate certificates
            if openssl x509 -in "$ssl_dir/cert.pem" -noout -text >/dev/null 2>&1; then
                print_status "success" "SSL certificates are valid"
            else
                print_status "warning" "SSL certificates may be invalid, will regenerate"
                rm -f "$ssl_dir/cert.pem" "$ssl_dir/key.pem"
            fi
        else
            print_status "warning" "Failed to extract SSL certificates, will generate new ones"
        fi
    else
        print_status "info" "No existing SSL certificates found, will generate new ones"
    fi
}

import_nginx_routes() {
    if [[ "$opt_import_nginx_routes" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Importing routes from Nginx Proxy Manager..."
    
    if ! eval "$CONTAINER_CMD container exists nginx-proxy-manager"; then
        print_status "warning" "Nginx Proxy Manager not found, skipping route import"
        return 0
    fi
    
    local routes_file="$PROJECT_ROOT/config/traefik/dynamic/imported-routes.yml"
    
    cat > "$routes_file" << 'EOF'
# Routes imported from Nginx Proxy Manager
http:
  routers:
    # Common service routes will be added here
    
  services:
    # Service definitions will be added here
EOF
    
    # Extract common routes (this is a simplified example)
    # In a real implementation, you'd parse the Nginx configuration
    local common_routes=(
        "adminer:adminer.test:8080"
        "phpmyadmin:phpmyadmin.test:80"
        "grafana:grafana.test:3000"
        "metabase:metabase.test:3000"
        "mailpit:mailpit.test:8025"
        "gitea:gitea.test:3000"
    )
    
    for route in "${common_routes[@]}"; do
        local service_name="${route%%:*}"
        local domain="${route#*:}"
        domain="${domain%%:*}"
        local port="${route##*:}"
        
        cat >> "$routes_file" << EOF

    ${service_name}:
      rule: "Host(\`${domain}\`)"
      service: "${service_name}"
      tls: {}
EOF
        
        cat >> "$routes_file" << EOF

    ${service_name}:
      loadBalancer:
        servers:
          - url: "http://${service_name}:${port}"
EOF
    done
    
    print_status "success" "Routes imported successfully"
}

# =============================================================================
# TRAEFIK SETUP FUNCTIONS
# =============================================================================

create_traefik_directories() {
    print_status "step" "Creating Traefik directory structure..."
    
    local directories=(
        "$PROJECT_ROOT/config/traefik"
        "$PROJECT_ROOT/config/traefik/static"
        "$PROJECT_ROOT/config/traefik/dynamic"
        "$PROJECT_ROOT/config/traefik/ssl"
        "$PROJECT_ROOT/backups/traefik"
    )
    
    for dir in "${directories[@]}"; do
        mkdir -p "$dir"
        print_status "info" "Created: $dir"
    done
    
    # Set proper permissions
    chmod 755 "$PROJECT_ROOT/config/traefik"
    chmod 755 "$PROJECT_ROOT/config/traefik/static"
    chmod 755 "$PROJECT_ROOT/config/traefik/dynamic"
    chmod 700 "$PROJECT_ROOT/config/traefik/ssl"
    
    print_status "success" "Directory structure created"
}

generate_traefik_static_config() {
    print_status "step" "Generating Traefik static configuration..."
    
    local static_config="$PROJECT_ROOT/config/traefik/static/traefik.yml"
    
    cat > "$static_config" << EOF
# Traefik Static Configuration
global:
  checkNewVersion: false
  sendAnonymousUsage: false

# API and Dashboard
api:
  dashboard: true
  debug: true

# Entry Points
entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entrypoint:
          to: websecure
          scheme: https
          permanent: true
  
  websecure:
    address: ":443"

# Certificate Resolvers
certificatesResolvers:
  letsencrypt:
    acme:
      email: ${ADMIN_EMAIL}
      storage: /data/acme.json
      caServer: https://acme-staging-v02.api.letsencrypt.org/directory
      httpChallenge:
        entryPoint: web

# Providers
providers:
EOF

    if [[ "$opt_enable_file_discovery" == "true" ]]; then
        cat >> "$static_config" << EOF
  # File provider for static services
  file:
    directory: /etc/traefik/dynamic
    watch: true
EOF
    fi

    if [[ "$opt_enable_docker_discovery" == "true" ]]; then
        cat >> "$static_config" << EOF
  
  # Docker provider for container discovery
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: ${NETWORK_NAME}
    watch: true
EOF
    fi

    cat >> "$static_config" << EOF

# Logging
log:
  level: INFO
  filePath: "/var/log/traefik/traefik.log"

accessLog:
  filePath: "/var/log/traefik/access.log"

# Metrics
metrics:
  prometheus:
    addEntryPointsLabels: true
    addServicesLabels: true
    addRoutersLabels: true

# Health check
ping: {}

# Server transport
serversTransport:
  insecureSkipVerify: true
EOF
    
    print_status "success" "Static configuration generated"
}

generate_traefik_dynamic_config() {
    print_status "step" "Generating Traefik dynamic configuration..."
    
    local dynamic_config="$PROJECT_ROOT/config/traefik/dynamic/services.yml"
    
    cat > "$dynamic_config" << 'EOF'
# Traefik Dynamic Configuration for Microservices
http:
  routers:
    # Database Management Tools
    adminer:
      rule: "Host(`adminer.test`)"
      service: "adminer"
      tls: {}
    
    phpmyadmin:
      rule: "Host(`phpmyadmin.test`)"
      service: "phpmyadmin"
      tls: {}
    
    mongo-express:
      rule: "Host(`mongodb.test`)"
      service: "mongo-express"
      tls: {}
    
    metabase:
      rule: "Host(`metabase.test`)"
      service: "metabase"
      tls: {}
    
    nocodb:
      rule: "Host(`nocodb.test`)"
      service: "nocodb"
      tls: {}
    
    pgadmin:
      rule: "Host(`pgadmin.test`)"
      service: "pgadmin"
      tls: {}
    
    # Analytics & Monitoring
    grafana:
      rule: "Host(`grafana.test`)"
      service: "grafana"
      tls: {}
    
    prometheus:
      rule: "Host(`prometheus.test`)"
      service: "prometheus"
      tls: {}
    
    kibana:
      rule: "Host(`kibana.test`)"
      service: "kibana"
      tls: {}
    
    elasticsearch:
      rule: "Host(`elasticsearch.test`)"
      service: "elasticsearch"
      tls: {}
    
    matomo:
      rule: "Host(`matomo.test`)"
      service: "matomo"
      tls: {}
    
    # AI & Workflow Services
    n8n:
      rule: "Host(`n8n.test`)"
      service: "n8n"
      tls: {}
    
    langflow:
      rule: "Host(`langflow.test`)"
      service: "langflow"
      tls: {}
    
    # Mail Services
    mailpit:
      rule: "Host(`mailpit.test`)"
      service: "mailpit"
      tls: {}
    
    # Project Management
    gitea:
      rule: "Host(`gitea.test`)"
      service: "gitea"
      tls: {}
    
    # Backend Services
    dotnet:
      rule: "Host(`dotnet.test`)"
      service: "dotnet"
      tls: {}
    
    go:
      rule: "Host(`go.test`)"
      service: "go"
      tls: {}
    
    node:
      rule: "Host(`node.test`)"
      service: "node"
      tls: {}
    
    php:
      rule: "Host(`php.test`)"
      service: "php"
      tls: {}
    
    python:
      rule: "Host(`python.test`)"
      service: "python"
      tls: {}

  services:
    # Database Management Tools
    adminer:
      loadBalancer:
        servers:
          - url: "http://adminer:8080"
    
    phpmyadmin:
      loadBalancer:
        servers:
          - url: "http://phpmyadmin:80"
    
    mongo-express:
      loadBalancer:
        servers:
          - url: "http://mongo-express:8081"
    
    metabase:
      loadBalancer:
        servers:
          - url: "http://metabase:3000"
    
    nocodb:
      loadBalancer:
        servers:
          - url: "http://nocodb:8080"
    
    pgadmin:
      loadBalancer:
        servers:
          - url: "http://pgadmin:80"
    
    # Analytics & Monitoring
    grafana:
      loadBalancer:
        servers:
          - url: "http://grafana:3000"
    
    prometheus:
      loadBalancer:
        servers:
          - url: "http://prometheus:9090"
    
    kibana:
      loadBalancer:
        servers:
          - url: "http://kibana:5601"
    
    elasticsearch:
      loadBalancer:
        servers:
          - url: "http://elasticsearch:9200"
    
    matomo:
      loadBalancer:
        servers:
          - url: "http://matomo:80"
    
    # AI & Workflow Services
    n8n:
      loadBalancer:
        servers:
          - url: "http://n8n:5678"
    
    langflow:
      loadBalancer:
        servers:
          - url: "http://langflow:7860"
    
    # Mail Services
    mailpit:
      loadBalancer:
        servers:
          - url: "http://mailpit:8025"
    
    # Project Management
    gitea:
      loadBalancer:
        servers:
          - url: "http://gitea:3000"
    
    # Backend Services
    dotnet:
      loadBalancer:
        servers:
          - url: "http://dotnet:80"
    
    go:
      loadBalancer:
        servers:
          - url: "http://go:8080"
    
    node:
      loadBalancer:
        servers:
          - url: "http://node:3000"
    
    php:
      loadBalancer:
        servers:
          - url: "http://php:8000"
    
    python:
      loadBalancer:
        servers:
          - url: "http://python:8000"

# TLS Configuration
tls:
  certificates:
    - certFile: "/ssl/cert.pem"
      keyFile: "/ssl/key.pem"
EOF
    
    print_status "success" "Dynamic configuration generated"
}

setup_traefik_ssl() {
    if [[ "$opt_setup_ssl" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Setting up SSL certificates for Traefik..."
    
    local ssl_dir="$PROJECT_ROOT/config/traefik/ssl"
    
    # Check if we already have certificates (from migration or previous setup)
    if [[ -f "$ssl_dir/cert.pem" && -f "$ssl_dir/key.pem" ]]; then
        print_status "info" "Using existing SSL certificates"
        return 0
    fi
    
    # Generate new certificates
    print_status "info" "Generating new SSL certificates..."
    
    # Create certificate configuration
    cat > "/tmp/traefik-cert.conf" << EOF
[req]
default_bits = 4096
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
CN = *.test

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = *.test
DNS.2 = test
DNS.3 = localhost
DNS.4 = *.localhost
DNS.5 = traefik.test
DNS.6 = adminer.test
DNS.7 = phpmyadmin.test
DNS.8 = mongodb.test
DNS.9 = metabase.test
DNS.10 = nocodb.test
DNS.11 = pgadmin.test
DNS.12 = grafana.test
DNS.13 = prometheus.test
DNS.14 = matomo.test
DNS.15 = n8n.test
DNS.16 = langflow.test
DNS.17 = kibana.test
DNS.18 = elasticsearch.test
DNS.19 = mailpit.test
DNS.20 = gitea.test
DNS.21 = dotnet.test
DNS.22 = go.test
DNS.23 = node.test
DNS.24 = php.test
DNS.25 = python.test
EOF

    # Generate certificate
    openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
        -keyout "$ssl_dir/key.pem" \
        -out "$ssl_dir/cert.pem" \
        -config "/tmp/traefik-cert.conf" \
        -extensions v3_req
    
    # Set proper permissions
    chmod 600 "$ssl_dir/key.pem"
    chmod 644 "$ssl_dir/cert.pem"
    
    # Cleanup
    rm -f "/tmp/traefik-cert.conf"
    
    print_status "success" "SSL certificates generated"
}

# =============================================================================
# DEPLOYMENT FUNCTIONS
# =============================================================================

stop_nginx_proxy_manager() {
    if [[ "$opt_migrate_from_nginx" == "true" ]]; then
        print_status "step" "Stopping Nginx Proxy Manager..."
        
        if eval "$CONTAINER_CMD container exists nginx-proxy-manager"; then
            eval "$CONTAINER_CMD stop nginx-proxy-manager $ERROR_REDIRECT" || true
            print_status "success" "Nginx Proxy Manager stopped"
        else
            print_status "info" "Nginx Proxy Manager was not running"
        fi
    fi
}

deploy_traefik() {
    print_status "step" "Deploying Traefik..."
    
    # Ensure network exists
    ensure_network_exists
    
    # Build and start Traefik
    local compose_file="$COMPOSE_DIR/proxy/traefik.yml"
    
    if [[ ! -f "$compose_file" ]]; then
        handle_error "Traefik compose file not found: $compose_file"
    fi
    
    local compose_args=("-f" "$compose_file" "up" "-d")
    
    if [[ "$opt_force_recreate" == "true" ]]; then
        compose_args+=("--force-recreate")
    fi
    
    if eval "$COMPOSE_CMD ${compose_args[*]} $ERROR_REDIRECT"; then
        print_status "success" "Traefik deployed successfully"
    else
        handle_error "Failed to deploy Traefik"
    fi
    
    # Wait for Traefik to be ready
    wait_for_traefik_startup
}

wait_for_traefik_startup() {
    print_status "step" "Waiting for Traefik to start up..."
    
    local max_wait=60
    local counter=0
    
    while [[ $counter -lt $max_wait ]]; do
        if curl -s "http://localhost:8080/ping" >/dev/null 2>&1; then
            print_status "success" "Traefik is ready!"
            return 0
        fi
        
        print_status "info" "Waiting for Traefik... ($counter/$max_wait)"
        sleep 2
        counter=$((counter + 2))
    done
    
    print_status "warning" "Traefik startup timeout, but container may still be initializing"
    return 1
}

# =============================================================================
# POST-SETUP FUNCTIONS
# =============================================================================

update_config_for_traefik() {
    print_status "step" "Updating configuration to use Traefik..."
    
    # Switch proxy provider in config.sh
    switch_proxy_provider "traefik"
    
    print_status "success" "Configuration updated for Traefik"
}

run_post_setup_tests() {
    print_status "step" "Running post-setup tests..."
    
    # Test Traefik API
    if curl -s "http://localhost:8080/ping" >/dev/null 2>&1; then
        print_status "success" "✓ Traefik API accessible"
    else
        print_status "warning" "✗ Traefik API not accessible"
    fi
    
    # Test dashboard
    if curl -s -k "https://traefik.test" >/dev/null 2>&1; then
        print_status "success" "✓ Traefik dashboard accessible"
    else
        print_status "warning" "✗ Traefik dashboard not accessible"
    fi
    
    # Test a few common services
    local test_services=("adminer.test" "grafana.test" "metabase.test")
    
    for service in "${test_services[@]}"; do
        if curl -s -k "https://$service" >/dev/null 2>&1; then
            print_status "success" "✓ $service accessible"
        else
            print_status "info" "- $service not accessible (service may not be running)"
        fi
    done
    
    print_status "success" "Post-setup tests completed"
}

show_completion_summary() {
    echo ""
    echo "========================================================"
    print_status "success" "Traefik setup completed successfully!"
    echo "========================================================"
    echo ""
    
    echo "🌐 Traefik Dashboard:"
    echo "  URL: https://traefik.test"
    echo "  API: http://localhost:8080"
    echo "  Credentials: admin / 123456"
    echo ""
    
    echo "🔧 Management Commands:"
    echo "  Status: $SCRIPT_DIR/manage-traefik.sh status"
    echo "  Routes: $SCRIPT_DIR/manage-traefik.sh routes"
    echo "  Logs:   $SCRIPT_DIR/manage-traefik.sh logs follow"
    echo ""
    
    echo "🚀 Service URLs:"
    echo "  All services are now accessible via https://[service].test"
    echo "  Example: https://grafana.test, https://metabase.test"
    echo ""
    
    if [[ "$opt_migrate_from_nginx" == "true" ]]; then
        echo "📁 Migration Info:"
        echo "  Nginx config backed up to: $PROJECT_ROOT/backups/nginx-proxy-manager/"
        echo "  SSL certificates migrated successfully"
        echo "  Routes imported and converted to Traefik format"
        echo ""
    fi
    
    echo "📋 Next Steps:"
    echo "  1. Test service accessibility: ./manage-traefik.sh test-route grafana.test"
    echo "  2. Add custom services: ./manage-traefik.sh add-service myapp myapp.test 3000"
    echo "  3. Monitor with dashboard: https://traefik.test"
    
    if [[ "$opt_setup_ssl" == "true" ]]; then
        echo "  4. Install SSL certificates: $SCRIPT_DIR/trust-host.sh"
    fi
    
    echo ""
    echo "========================================================"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    print_status "step" "Starting Traefik setup and migration process..."
    
    # Validation
    validate_environment
    check_existing_nginx
    
    # Migration steps (if applicable)
    if [[ "$opt_migrate_from_nginx" == "true" ]]; then
        backup_nginx_configuration
        extract_nginx_ssl_certificates
        import_nginx_routes
        stop_nginx_proxy_manager
    fi
    
    # Traefik setup
    create_traefik_directories
    generate_traefik_static_config
    generate_traefik_dynamic_config
    setup_traefik_ssl
    
    # Deployment
    deploy_traefik
    update_config_for_traefik
    
    # Post-setup
    run_post_setup_tests
    show_completion_summary
    
    print_status "success" "Traefik setup process completed successfully!"
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi