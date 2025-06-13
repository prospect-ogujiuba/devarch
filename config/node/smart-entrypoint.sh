#!/bin/bash

echo "🔍 Node.js Smart Entrypoint: Detecting project structure..."

# Function to detect and setup Node.js projects
detect_node_projects() {
    for app_dir in /app/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"
            
            cd "$app_dir"
            
            # Check if package.json exists
            if [ -f "package.json" ]; then
                echo "📦 package.json found in $app_name"
                
                # Install dependencies
                echo "📥 Installing dependencies..."
                npm ci 2>/dev/null || npm install
                
                # Detect project type from package.json
                if grep -q "\"next\"" package.json; then
                    echo "⚡ Next.js project detected in $app_name"
                    # Build Next.js app if not in development
                    if [ "$NODE_ENV" = "production" ]; then
                        npm run build
                    fi
                    
                elif grep -q "\"@nestjs/core\"" package.json; then
                    echo "🐱 NestJS project detected in $app_name"
                    # Build NestJS app if TypeScript
                    if [ -f "tsconfig.json" ]; then
                        npm run build 2>/dev/null || true
                    fi
                    
                elif grep -q "\"react\"" package.json && grep -q "\"react-dom\"" package.json; then
                    echo "⚛️  React project detected in $app_name"
                    # Build React app for production
                    if [ "$NODE_ENV" = "production" ]; then
                        npm run build 2>/dev/null || true
                    fi
                    
                elif grep -q "\"express\"" package.json; then
                    echo "🚂 Express.js project detected in $app_name"
                    # No build step needed for basic Express
                    
                elif grep -q "\"fastify\"" package.json; then
                    echo "⚡ Fastify project detected in $app_name"
                    # No build step needed for basic Fastify
                    
                elif [ -f "tsconfig.json" ]; then
                    echo "📘 TypeScript project detected in $app_name"
                    # Build TypeScript project
                    npm run build 2>/dev/null || npx tsc 2>/dev/null || true
                    
                else
                    echo "📄 Generic Node.js project detected in $app_name"
                fi
                
                # Make sure start script exists
                if ! grep -q "\"start\"" package.json; then
                    echo "⚠️  No start script found in $app_name package.json"
                    # Try to detect main file
                    if [ -f "server.js" ]; then
                        echo "🔧 Using server.js as entry point"
                    elif [ -f "app.js" ]; then
                        echo "🔧 Using app.js as entry point"
                    elif [ -f "index.js" ]; then
                        echo "🔧 Using index.js as entry point"
                    fi
                fi
                
            elif [ -f "server.js" ] || [ -f "app.js" ] || [ -f "index.js" ]; then
                echo "📄 Standalone Node.js file detected in $app_name"
                # Create a basic package.json if none exists
                if [ ! -f "package.json" ]; then
                    echo "🔧 Creating basic package.json"
                    cat > package.json << EOF
{
  "name": "$app_name",
  "version": "1.0.0",
  "main": "$(ls *.js | head -1)",
  "scripts": {
    "start": "node $(ls *.js | head -1)"
  }
}
EOF
                fi
                
            else
                echo "❓ No recognizable Node.js project structure in $app_name"
            fi
        fi
    done
}

# Run detection
detect_node_projects

echo "✅ Node.js Smart Entrypoint: Setup complete!"

# If we're in a specific app directory and have package.json, use it
if [ -f "/app/package.json" ]; then
    cd /app
elif [ -d "/app" ]; then
    # Find the first app with package.json and cd into it
    for app_dir in /app/*/; do
        if [ -f "$app_dir/package.json" ]; then
            cd "$app_dir"
            echo "🎯 Switching to $app_dir for execution"
            break
        fi
    done
fi

# Execute the original command
exec "$@"