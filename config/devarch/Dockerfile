# DevArch Manager Container
# Purpose: Multi-language management container for DevArch platform
# Provides: Container orchestration (Podman), all language runtimes, database clients
# Use cases: Dashboard API, cross-language scripting, environment management

FROM debian:bookworm-slim

# Version arguments for maintainability
ARG PHP_VERSION=8.3
ARG NODE_VERSION=20.x
ARG PYTHON_VERSION=3.11
ARG GO_VERSION=1.21.5
ARG DOTNET_VERSION=8.0
ARG RUST_VERSION=stable

# Install base dependencies, Podman, and development tools
# Order: System packages (least frequently changed) first
RUN apt-get update && apt-get install -y \
    # Core system utilities
    sudo \
    curl \
    wget \
    git \
    vim \
    nano \
    htop \
    unzip \
    ca-certificates \
    gnupg2 \
    lsb-release \
    software-properties-common \
    apt-transport-https \
    # Shell
    zsh \
    # Container management
    podman \
    podman-compose \
    # Database clients
    default-mysql-client \
    postgresql-client \
    redis-tools \
    # Build essentials for language installations
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install PHP 8.3
RUN curl -sSL https://packages.sury.org/php/README.txt | bash - \
    && apt-get update \
    && apt-get install -y \
        php${PHP_VERSION}-cli \
        php${PHP_VERSION}-common \
        php${PHP_VERSION}-curl \
        php${PHP_VERSION}-mbstring \
        php${PHP_VERSION}-xml \
        php${PHP_VERSION}-zip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer
ENV COMPOSER_ALLOW_SUPERUSER=1

# Install Node.js 20 LTS
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION} | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Python 3.11 (Debian bookworm default)
RUN apt-get update && apt-get install -y \
    python3 \
    python3-dev \
    python3-venv \
    python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Install .NET SDK 8.0
RUN wget https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb -O packages-microsoft-prod.deb \
    && dpkg -i packages-microsoft-prod.deb \
    && rm packages-microsoft-prod.deb \
    && apt-get update \
    && apt-get install -y dotnet-sdk-${DOTNET_VERSION} \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain ${RUST_VERSION} \
    && /root/.cargo/bin/rustup component add rustfmt clippy
ENV PATH="/root/.cargo/bin:${PATH}"

# Install Oh My Zsh and set as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
    && chsh -s $(which zsh)

# Create devarch user with sudo access
RUN useradd -m -s /bin/zsh -u 1000 devarch \
    && echo "devarch ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/devarch \
    && chmod 0440 /etc/sudoers.d/devarch

# Configure Podman for devarch user
RUN mkdir -p /home/devarch/.config && chown -R devarch:devarch /home/devarch/.config

# Set up Oh My Zsh for devarch user
USER devarch
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
USER root

# Set working directory
WORKDIR /workspace

# Expose port for potential web UI
EXPOSE 8000

# Default command: keep container running with zsh available
CMD ["/bin/zsh", "-c", "tail -f /dev/null"]
