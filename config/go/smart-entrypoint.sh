#!/bin/bash

echo "🔍 Go Smart Entrypoint: Detecting project structure..."

# Function to detect and setup Go projects
detect_go_projects() {
    for app_dir in /app/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"
            
            cd "$app_dir"
            
            # Go Module Detection
            if [ -f "go.mod" ]; then
                echo "📦 Go module project detected in $app_name"
                
                # Download dependencies
                echo "📥 Downloading Go modules..."
                go mod download
                go mod tidy
                
                # Detect project type by checking imports and files
                if grep -r "github.com/gin-gonic/gin" . >/dev/null 2>&1; then
                    echo "🍸 Gin framework detected in $app_name"
                    
                elif grep -r "github.com/gorilla/mux" . >/dev/null 2>&1; then
                    echo "🦍 Gorilla Mux framework detected in $app_name"
                    
                elif grep -r "github.com/labstack/echo" . >/dev/null 2>&1; then
                    echo "📢 Echo framework detected in $app_name"
                    
                elif grep -r "github.com/gofiber/fiber" . >/dev/null 2>&1; then
                    echo "🚀 Fiber framework detected in $app_name"
                    
                elif grep -r "net/http" . >/dev/null 2>&1; then
                    echo "🌐 Standard net/http detected in $app_name"
                    
                elif grep -r "github.com/spf13/cobra" . >/dev/null 2>&1; then
                    echo "🐍 Cobra CLI application detected in $app_name"
                    
                else
                    echo "📄 Generic Go application detected in $app_name"
                fi
                
                # Build the application for production
                if [ "$GO_ENV" = "production" ]; then
                    echo "🔨 Building Go application..."
                    go build -o app .
                fi
                
            # Standalone Go files
            elif ls *.go >/dev/null 2>&1; then
                echo "🐹 Standalone Go files detected in $app_name"
                
                # Initialize go module if main.go exists
                if [ -f "main.go" ] && [ ! -f "go.mod" ]; then
                    echo "🔧 Initializing Go module..."
                    go mod init "$app_name"
                    go mod tidy
                fi
                
            else
                echo "❓ No recognizable Go project structure in $app_name"
            fi
            
            # Set proper permissions
            chown -R app:app "$app_dir"
        fi
    done
}

# Run detection
detect_go_projects

echo "✅ Go Smart Entrypoint: Setup complete!"

# If we're in a specific app directory, use it
if [ -f "/app/main.go" ] || [ -f "/app/go.mod" ]; then
    cd /app
elif [ -d "/app" ]; then
    # Find the first app with Go files and cd into it
    for app_dir in /app/*/; do
        if [ -f "$app_dir/main.go" ] || [ -f "$app_dir/go.mod" ]; then
            cd "$app_dir"
            echo "🎯 Switching to $app_dir for execution"
            break
        fi
    done
fi

# Switch to app user for security
exec su-exec app "$@" 2>/dev/null || exec "$@"