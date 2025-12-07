#!/bin/bash
# Discovers and runs .NET applications in /app directory
# Auto-discovers apps with .csproj files and runs them with dotnet run
# Part of DevArch Multi-Language Backend Architecture

set -e

APPS_DIR="/app"
BASE_PORT=8080

echo "🟣 .NET Apps Launcher - DevArch Multi-Language Backend"
echo "======================================================="
echo "🔍 Discovering .NET applications in $APPS_DIR..."
echo ""

port=$BASE_PORT
app_count=0
pids=()

for app_dir in "$APPS_DIR"/*/; do
    app_name=$(basename "$app_dir")

    # Skip special directories
    if [[ "$app_name" == "node_modules" ]] || [[ "$app_name" == "logs" ]] || [[ "$app_name" == ".git" ]]; then
        continue
    fi

    # Skip non-.NET apps (check for *.csproj or *.fsproj)
    csproj=$(find "$app_dir" -maxdepth 1 -name "*.csproj" -o -name "*.fsproj" 2>/dev/null | head -1)

    if [ -z "$csproj" ]; then
        continue
    fi

    project_file=$(basename "$csproj")
    echo "📦 Found .NET app: $app_name"
    echo "   Location: $app_dir"
    echo "   Project file: $project_file"

    cd "$app_dir"

    # Restore NuGet packages
    echo "   📥 Restoring NuGet packages..."
    dotnet restore "$project_file" --verbosity quiet 2>&1 | grep -E '(Restored|Failed|error)' || true

    # Build the project
    echo "   🔨 Building project..."
    if dotnet build "$project_file" --no-restore --verbosity quiet 2>&1 | grep -E '(Build succeeded|Build FAILED)'; then
        # Check if build succeeded
        if ! dotnet build "$project_file" --no-restore --verbosity quiet > /dev/null 2>&1; then
            echo "   ❌ Build failed for $app_name"
            echo ""
            continue
        fi

        echo "   ✅ Build successful"
        echo "   🚀 Starting on port $port..."

        # Run app in background with port override
        ASPNETCORE_URLS="http://*:$port" \
        DOTNET_ENVIRONMENT="Development" \
        PORT="$port" \
        APP_NAME="$app_name" \
        dotnet run --project "$project_file" --no-build --verbosity quiet > "/app/logs/$app_name.log" 2>&1 &

        pid=$!
        pids+=($pid)

        # Give it a moment to start
        sleep 2

        # Check if still running
        if kill -0 "$pid" 2>/dev/null; then
            echo "   ✓ Started with PID $pid"
            echo "   📍 Access: http://localhost:8600 (external) → :$port (internal)"
            echo "   📝 Logs: /app/logs/$app_name.log"
            echo ""

            ((app_count++))
            ((port++))
        else
            echo "   ❌ Failed to start - check logs at /app/logs/$app_name.log"
            echo ""
        fi
    else
        echo "   ❌ Build failed for $app_name"
        echo ""
        continue
    fi
done

echo "======================================================="
if [ $app_count -eq 0 ]; then
    echo "⚠️  No .NET applications found in $APPS_DIR"
    echo ""
    echo "💡 To create a .NET app:"
    echo "   1. Use Rider IDE: File → New Project → ASP.NET Core/Blazor"
    echo "   2. Location: $APPS_DIR/my-dotnet-app"
    echo "   3. Ensure port binding reads from environment:"
    echo "      builder.WebHost.UseUrls(\$\"http://*:{Environment.GetEnvironmentVariable(\"PORT\") ?? \"8080\"}\");"
    echo "   4. Container will auto-build and start app"
    echo ""
    echo "⏳ Container will stay alive (no apps to run)..."
    sleep infinity
else
    echo "✅ Successfully started $app_count .NET application(s)"
    echo "🔍 Process IDs: ${pids[*]}"
    echo ""
    echo "📊 Port mapping:"
    echo "   External (host) → Internal (container)"
    for ((i=0; i<app_count; i++)); do
        external_port=$((8600 + i))
        internal_port=$((8080 + i))
        echo "   localhost:$external_port → :$internal_port (app $((i+1)))"
    done
    echo ""
    echo "🔧 Container commands:"
    echo "   devarch exec dotnet ps aux | grep dotnet  - List running .NET apps"
    echo "   devarch logs dotnet -f                    - Follow container logs"
    echo "   devarch restart dotnet                    - Rebuild and restart all apps"
    echo ""
    echo "⏳ Keeping container alive..."
    echo "   All apps running in background"
    echo "   Press Ctrl+C to stop all apps"
    echo ""
fi

# Function to kill all child processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down .NET applications..."
    for pid in "${pids[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "   Stopping PID $pid..."
            kill "$pid" 2>/dev/null || true
        fi
    done
    # Also kill any remaining dotnet processes
    pkill -f "dotnet run" 2>/dev/null || true
    echo "✅ All .NET apps stopped"
    exit 0
}

trap cleanup SIGTERM SIGINT

# Wait for all background processes
wait
