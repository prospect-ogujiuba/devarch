FROM mcr.microsoft.com/dotnet/sdk:8.0

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

# Set the working directory
WORKDIR /var/www/html

# Create non-root user
RUN useradd --create-home --shell /bin/bash app || echo "User app already exists"
RUN chown -R app:app /var/www/html

# Expose common .NET ports
EXPOSE 80 443 5000 5001 8080

# Install Python for a simple HTTP server (most reliable approach)
RUN apt-get update && apt-get install -y python3 && rm -rf /var/lib/apt/lists/*

# Use Python's built-in HTTP server - simple and reliable
CMD ["python3", "-m", "http.server", "80"]