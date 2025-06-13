#!/bin/bash

echo "🔍 .NET Smart Entrypoint: Detecting project structure..."

# Function to detect and setup .NET projects
detect_dotnet_projects() {
    for app_dir in /app/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"
            
            cd "$app_dir"
            
            # .NET Solution Detection
            if ls *.sln >/dev/null 2>&1; then
                echo "📊 .NET Solution detected in $app_name"
                
                # Restore packages for solution
                echo "📥 Restoring NuGet packages for solution..."
                dotnet restore
                
                # Build solution
                if [ "$ASPNETCORE_ENVIRONMENT" = "Production" ]; then
                    echo "🔨 Building solution for production..."
                    dotnet build --configuration Release --no-restore
                fi
                
            # .NET Project Detection
            elif ls *.csproj >/dev/null 2>&1 || ls *.fsproj >/dev/null 2>&1 || ls *.vbproj >/dev/null 2>&1; then
                echo "📦 .NET Project detected in $app_name"
                
                # Restore packages for project
                echo "📥 Restoring NuGet packages..."
                dotnet restore
                
                # Detect project type by examining project file
                if grep -q "Microsoft.AspNetCore" *.csproj 2>/dev/null; then
                    echo "🌐 ASP.NET Core Web Application detected"
                    
                    # Check for specific frameworks
                    if grep -q "Microsoft.AspNetCore.Mvc" *.csproj 2>/dev/null; then
                        echo "🎭 ASP.NET Core MVC detected"
                    elif grep -q "Microsoft.AspNetCore.Blazor" *.csproj 2>/dev/null; then
                        echo "⚡ Blazor application detected"
                    elif grep -q "Microsoft.AspNetCore.ApiController" *.csproj 2>/dev/null; then
                        echo "🔌 ASP.NET Core Web API detected"
                    fi
                    
                    # Setup development certificates
                    dotnet dev-certs https --trust 2>/dev/null || true
                    
                elif grep -q "Microsoft.NET.Sdk.Worker" *.csproj 2>/dev/null; then
                    echo "⚙️  .NET Worker Service detected"
                    
                elif grep -q "Microsoft.WindowsDesktop.App" *.csproj 2>/dev/null; then
                    echo "🖥️  .NET Desktop Application detected"
                    
                elif grep -q "OutputType.*Exe" *.csproj 2>/dev/null; then
                    echo "⚡ .NET Console Application detected"
                    
                elif grep -q "OutputType.*Library" *.csproj 2>/dev/null; then
                    echo "📚 .NET Class Library detected"
                    
                else
                    echo "📄 Generic .NET Project detected"
                fi
                
                # Build project
                if [ "$ASPNETCORE_ENVIRONMENT" = "Production" ]; then
                    echo "🔨 Building project for production..."
                    dotnet build --configuration Release --no-restore
                fi
                
                # Run Entity Framework migrations if present
                if grep -q "Microsoft.EntityFrameworkCore" *.csproj 2>/dev/null; then
                    echo "🗃️  Entity Framework detected, running migrations..."
                    dotnet ef database update 2>/dev/null || true
                fi
                
            # Standalone C# files
            elif ls *.cs >/dev/null 2>&1; then
                echo "📄 Standalone C# files detected in $app_name"
                
                # Try to find a main method
                if grep -r "static void Main" *.cs >/dev/null 2>&1; then
                    echo "🎯 Main method found, this appears to be a console application"
                fi
                
            else
                echo "❓ No recognizable .NET project structure in $app_name"
            fi
            
            # Set proper permissions
            chown -R app:app "$app_dir"
        fi
    done
}

# Run detection
detect_dotnet_projects

echo "✅ .NET Smart Entrypoint: Setup complete!"

# If we're in a specific app directory, use it
if ls /app/*.sln >/dev/null 2>&1 || ls /app/*.csproj >/dev/null 2>&1; then
    cd /app
elif [ -d "/app" ]; then
    # Find the first app with .NET files and cd into it
    for app_dir in /app/*/; do
        if ls "$app_dir"*.sln >/dev/null 2>&1 || ls "$app_dir"*.csproj >/dev/null 2>&1; then
            cd "$app_dir"
            echo "🎯 Switching to $app_dir for execution"
            break
        fi
    done
fi

# Switch to app user for security
exec su-exec app "$@" 2>/dev/null || exec "$@"