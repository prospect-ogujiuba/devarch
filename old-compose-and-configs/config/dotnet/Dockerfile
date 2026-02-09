# .NET 8.0 SDK for comprehensive development
FROM mcr.microsoft.com/dotnet/sdk:8.0-bookworm-slim

# Install system dependencies and database clients
RUN apt-get update && apt-get install -y \
    # Build tools
    build-essential \
    cmake \
    # Database clients (matching other backends)
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
    # Mail relay
    msmtp \
    msmtp-mta \
    # Additional libraries
    ca-certificates \
    libssl-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Add .NET tools to PATH (tools will be installed by user as needed)
ENV PATH="${PATH}:/root/.dotnet/tools"

# Configure msmtp for mail relay to mailpit
RUN echo "account default" > /etc/msmtprc \
    && echo "host mailpit" >> /etc/msmtprc \
    && echo "port 1025" >> /etc/msmtprc \
    && echo "from noreply@devarch.test" >> /etc/msmtprc \
    && chmod 644 /etc/msmtprc

# Install Oh My Zsh non-interactively
RUN RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s "$(which zsh)"

# Set working directory
WORKDIR /app

# Create non-root user (consistent with other backends)
RUN useradd -m -s /bin/zsh dotnet-user && \
    echo "dotnet-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R dotnet-user:dotnet-user /app /root/.dotnet

USER dotnet-user

# Add user's .NET tools to PATH (tools can be installed on-demand)
ENV PATH="${PATH}:/home/dotnet-user/.dotnet/tools"

# Configure environment for development
ENV ASPNETCORE_ENVIRONMENT=Development \
    DOTNET_WATCH_RESTART_ON_RUDE_EDIT=true \
    DOTNET_USE_POLLING_FILE_WATCHER=true \
    ASPNETCORE_URLS=http://+:8080 \
    ASPNETCORE_Kestrel__Endpoints__Http__Url=http://+:8080

# Expose ports
EXPOSE 8080 8081 10000 52323

# Default command
CMD ["dotnet", "--version"]
