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

# Create a simple ASP.NET Core file server
RUN mkdir -p /app
WORKDIR /app

# Create a simple web server project
RUN dotnet new web -n FileServer
WORKDIR /app/FileServer

# Modify Program.cs to serve static files from /var/www/html
RUN cat > Program.cs << 'EOF'
var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

// Configure static file serving from /var/www/html
app.UseStaticFiles(new StaticFileOptions
{
    FileProvider = new Microsoft.Extensions.FileProviders.PhysicalFileProvider("/var/www/html"),
    RequestPath = ""
});

// Fallback for SPA applications
app.MapFallbackToFile("/var/www/html/index.html");

// Basic API endpoint
app.MapGet("/api/status", () => new { 
    Status = "Running", 
    Server = ".NET File Server",
    Time = DateTime.Now 
});

app.Run("http://0.0.0.0:80");
EOF

# Build the application
RUN dotnet build

# Create app directory
WORKDIR /var/www/html

# Create non-root user (with proper error handling)
RUN useradd --create-home --shell /bin/bash app || echo "User app already exists"
RUN chown -R app:app /var/www/html

# Expose common .NET ports
EXPOSE 80 443 5000 5001 8080

# Run the .NET file server
CMD ["dotnet", "run", "--project", "/app/FileServer", "--urls", "http://0.0.0.0:80"]