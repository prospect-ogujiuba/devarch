# =================================================================
# DEVARCH NGINX PROXY MANAGER WITH CUSTOM CONFIGURATION
# =================================================================
# Custom Nginx Proxy Manager image with pre-configured DevArch
# routing and SSL certificate support for *.test domains.

FROM jc21/nginx-proxy-manager:latest

# Install additional packages for DevArch
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    nano \
    vim \
    net-tools \
    dnsutils \
    && rm -rf /var/lib/apt/lists/*

# Copy custom Nginx configuration for *.test domain routing
COPY nginx-custom.conf /data/nginx/custom/nginx-custom.conf

# Copy SSL setup script
COPY setup-ssl.sh /usr/local/bin/setup-ssl.sh
RUN chmod +x /usr/local/bin/setup-ssl.sh

# Copy startup script that configures *.test routing
COPY npm-startup.sh /usr/local/bin/npm-startup.sh
RUN chmod +x /usr/local/bin/npm-startup.sh

# Create directories for SSL certificates
RUN mkdir -p /etc/letsencrypt/live/wildcard.test \
    && mkdir -p /data/nginx/proxy_host \
    && mkdir -p /data/nginx/stream \
    && mkdir -p /data/logs

# Copy default proxy host configurations
COPY proxy-hosts/ /data/nginx/proxy_host/

# Set proper permissions
RUN chown -R root:root /data/nginx/custom \
    && chmod -R 644 /data/nginx/custom/*.conf

# Health check for DevArch-specific functionality
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:80 || exit 1

# Use custom startup script
ENTRYPOINT ["/usr/local/bin/npm-startup.sh"]