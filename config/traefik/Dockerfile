FROM traefik:v3.4.1

# Install additional packages
RUN apk add --no-cache \
    zsh \
    git \
    curl \
    wget \
    bash \
    openssl \
    ca-certificates

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Create necessary directories
RUN mkdir -p /etc/traefik/static \
             /etc/traefik/dynamic \
             /ssl \
             /data \
             /var/log/traefik

# Set proper permissions
RUN chmod 755 /etc/traefik/static /etc/traefik/dynamic
RUN chmod 700 /ssl /data
RUN chmod 755 /var/log/traefik

# Create traefik user
RUN addgroup -g 1000 traefik && \
    adduser -u 1000 -G traefik -s /bin/sh -D traefik

# Set ownership
RUN chown -R traefik:traefik /data /var/log/traefik /ssl

# Use custom entrypoint
ENTRYPOINT ["/usr/local/bin/traefik-entrypoint.sh"]
CMD ["traefik"]