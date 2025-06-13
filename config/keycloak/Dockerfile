FROM quay.io/keycloak/keycloak:23.0

# Switch to root for package installation
USER root

# Install additional packages using dnf (not microdnf)
# Keycloak image is based on RHEL UBI which uses dnf
RUN dnf update -y && \
    dnf install -y \
        curl \
        wget \
        vim \
        nano \
        findutils \
        git \
        zsh \
    && dnf clean all

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Create custom keycloak configuration directory
RUN mkdir -p /opt/keycloak/conf/custom

# Copy any custom configuration files (if they exist)
COPY custom-keycloak.conf /opt/keycloak/conf/custom/ || true

# Set proper permissions
RUN chown -R keycloak:keycloak /opt/keycloak/conf/custom

# Switch back to keycloak user for security
USER keycloak

# Set working directory
WORKDIR /opt/keycloak

# Expose Keycloak ports
EXPOSE 8080 8443 9000

# Use the default Keycloak entrypoint
ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]