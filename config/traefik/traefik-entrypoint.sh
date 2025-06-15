#!/bin/bash

echo "🚀 Traefik Smart Entrypoint: Initializing..."

# Function to generate self-signed certificates
generate_ssl_certificates() {
    if [ ! -f "/ssl/cert.pem" ] || [ ! -f "/ssl/key.pem" ]; then
        echo "🔐 Generating self-signed SSL certificates..."
        
        # Create certificate configuration
        cat > /tmp/cert.conf << EOF
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
DNS.5 = nginx.test
DNS.6 = traefik.test
DNS.7 = adminer.test
DNS.8 = phpmyadmin.test
DNS.9 = mongodb.test
DNS.10 = metabase.test
DNS.11 = nocodb.test
DNS.12 = pgadmin.test
DNS.13 = grafana.test
DNS.14 = prometheus.test
DNS.15 = matomo.test
DNS.16 = n8n.test
DNS.17 = langflow.test
DNS.18 = kibana.test
DNS.19 = elasticsearch.test
DNS.20 = mailpit.test
DNS.21 = gitea.test
EOF

        # Generate private key and certificate
        openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
            -keyout /ssl/key.pem \
            -out /ssl/cert.pem \
            -config /tmp/cert.conf \
            -extensions v3_req

        # Set proper permissions
        chmod 600 /ssl/key.pem
        chmod 644 /ssl/cert.pem
        chown traefik:traefik /ssl/key.pem /ssl/cert.pem 2>/dev/null || true
        
        echo "✅ SSL certificates generated successfully"
        rm -f /tmp/cert.conf
    else
        echo "ℹ️  SSL certificates already exist"
    fi
}

# Function to create default dynamic configuration
create_dynamic_config() {
    if [ ! -f "/etc/traefik/dynamic/services.yml" ]; then
        echo "🔧 Creating default dynamic configuration..."
        
        cat > /etc/traefik/dynamic/services.yml << 'EOF'
# Dynamic configuration for microservices
http:
  routers:
    # NOTE: Traefik dashboard is handled by api@internal service automatically
    # No need to manually define traefik router when using api@internal

    # Nginx Proxy Manager (if running)
    nginx-proxy-manager:
      rule: "Host(`nginx.test`)"
      service: "nginx-proxy-manager"
      tls: {}

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
    # Nginx Proxy Manager (only if it's running alongside Traefik)
    nginx-proxy-manager:
      loadBalancer:
        servers:
          - url: "http://nginx-proxy-manager:81"

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
        
        echo "✅ Dynamic configuration created"
    else
        echo "ℹ️  Dynamic configuration already exists"
    fi
}

# Function to create static configuration without Docker provider
create_static_config() {
    if [ ! -f "/etc/traefik/static/traefik.yml" ]; then
        echo "🔧 Creating static configuration..."
        
        cat > /etc/traefik/static/traefik.yml << 'EOF'
# Static configuration - File provider only (no Docker/Podman)
global:
  checkNewVersion: false
  sendAnonymousUsage: false

# API and Dashboard configuration
api:
  dashboard: true
  debug: true
  insecure: true

# Entry points
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

# Certificate resolvers and TLS
certificatesResolvers:
  default:
    acme:
      email: admin@site.test
      storage: /data/acme.json
      caServer: https://acme-staging-v02.api.letsencrypt.org/directory
      httpChallenge:
        entryPoint: web

# Provider configuration - FILE ONLY (no Docker provider to avoid socket errors)
providers:
  file:
    directory: /etc/traefik/dynamic
    watch: true

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

# Ping endpoint for health checks
ping: {}

# Server transport
serversTransport:
  insecureSkipVerify: true
EOF
        
        echo "✅ Static configuration created"
    else
        echo "ℹ️  Static configuration already exists"
    fi
}

# Function to set up ACME storage
setup_acme_storage() {
    if [ ! -f "/data/acme.json" ]; then
        echo "🔐 Setting up ACME storage..."
        touch /data/acme.json
        chmod 600 /data/acme.json
        chown traefik:traefik /data/acme.json 2>/dev/null || true
        echo "✅ ACME storage initialized"
    fi
}

# Function to create log files
setup_logging() {
    mkdir -p /var/log/traefik
    touch /var/log/traefik/traefik.log
    touch /var/log/traefik/access.log
    chown -R traefik:traefik /var/log/traefik 2>/dev/null || true
    chmod 644 /var/log/traefik/*.log
    echo "✅ Logging configured"
}

# Run setup functions
generate_ssl_certificates
create_static_config
create_dynamic_config
setup_acme_storage
setup_logging

echo "✅ Traefik Smart Entrypoint: Setup complete!"

# Clear any conflicting environment variables
unset TRAEFIK_METRICS_PROMETHEUS_ADDLABELS

# Execute the original command with proper config file
if [[ $# -eq 0 ]]; then
    exec traefik --configFile=/etc/traefik/static/traefik.yml
else
    # If arguments are passed, check if they're traefik config flags
    if [[ "$1" == "--"* ]]; then
        # Arguments are config flags, prepend traefik command
        exec traefik --configFile=/etc/traefik/static/traefik.yml "$@"
    else
        # Normal command execution
        exec "$@"
    fi
fi