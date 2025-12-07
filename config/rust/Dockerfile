# Rust 1.75 (latest stable) for comprehensive development
FROM rust:1.75-bookworm

# Install system dependencies and database clients
RUN apt-get update && apt-get install -y \
    # Build tools
    build-essential \
    cmake \
    pkg-config \
    libssl-dev \
    # Database clients
    default-mysql-client \
    postgresql-client \
    redis-tools \
    # Network tools
    curl \
    wget \
    netcat-openbsd \
    # Development tools
    git \
    vim \
    nano \
    zsh \
    sudo \
    htop \
    # Additional libraries
    ca-certificates \
    libpq-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Rust tools
RUN cargo install \
    cargo-watch \
    cargo-edit \
    cargo-expand \
    sqlx-cli \
    diesel_cli --no-default-features --features postgres,mysql

# Install Oh My Zsh
RUN RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s "$(which zsh)"

# Set working directory
WORKDIR /app

# Create non-root user
RUN useradd -m -s /bin/zsh rust-user && \
    echo "rust-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R rust-user:rust-user /app /usr/local/cargo

USER rust-user

# Configure Rust environment
ENV RUST_BACKTRACE=1 \
    CARGO_HOME=/usr/local/cargo

# Expose ports
EXPOSE 8080 8081 8082 8083

# Default command
CMD ["rustc", "--version"]
