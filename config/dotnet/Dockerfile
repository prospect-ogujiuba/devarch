FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    zsh \
    vim \
    nano \
    && rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install .NET tools
RUN dotnet tool install --global dotnet-ef && \
    dotnet tool install --global dotnet-aspnet-codegenerator && \
    dotnet tool install --global dotnet-dev-certs && \
    dotnet tool install --global dotnet-watch

# Add tools to PATH
ENV PATH="$PATH:/root/.dotnet/tools"

# Create app directory
WORKDIR /app

# Copy smart entrypoint script
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

# Create non-root user (with proper error handling)
RUN useradd --create-home --shell /bin/bash app || echo "User app already exists"
RUN chown -R app:app /app

# Expose common .NET ports
EXPOSE 80 443 5000 5001 8080

# Use smart entrypoint
ENTRYPOINT ["/usr/local/bin/smart-entrypoint.sh"]

# Default command
CMD ["dotnet", "run"]