#!/bin/bash
# Discovers, builds, and runs Go applications in /app directory
# Auto-discovers apps with go.mod, compiles binaries, and launches them
# Part of DevArch Multi-Language Backend Architecture

set -e

APPS_DIR="/app"
BASE_PORT=8080

echo "🔵 Go Apps Launcher - DevArch Multi-Language Backend"
echo "====================================================="
echo "🔍 Discovering Go applications in $APPS_DIR..."
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

    # Skip non-Go apps (check for go.mod)
    if [ ! -f "$app_dir/go.mod" ]; then
        continue
    fi

    echo "📦 Found Go app: $app_name"
    echo "   Location: $app_dir"
    echo "   Building..."

    cd "$app_dir"

    # Download dependencies
    echo "   📥 Downloading dependencies..."
    go mod download 2>&1 | grep -E '(go: downloading|already)' | head -5 || true

    # Build binary (place in app directory for easier access)
    echo "   🔨 Compiling binary..."
    if go build -o "$app_name" . 2>&1; then
        if [ ! -f "$app_name" ]; then
            echo "   ❌ Build failed - binary not found: $app_name"
            echo ""
            continue
        fi

        echo "   ✅ Build successful: $app_name"
        echo "   🚀 Starting on port $port..."

        # Run in background with port override
        PORT=$port ./"$app_name" > "/app/logs/$app_name.log" 2>&1 &
        pid=$!
        pids+=($pid)

        echo "   ✓ Started with PID $pid"
        echo "   📍 Access: http://localhost:8400 (external) → :$port (internal)"
        echo "   📝 Logs: /app/logs/$app_name.log"
        echo ""

        ((app_count++))
        ((port++))
    else
        echo "   ❌ Compilation failed for $app_name"
        echo ""
        continue
    fi
done

echo "====================================================="
if [ $app_count -eq 0 ]; then
    echo "⚠️  No Go applications found in $APPS_DIR"
    echo ""
    echo "💡 To create a Go app:"
    echo "   1. Use GoLand IDE: File → New Project → Go"
    echo "   2. Location: $APPS_DIR/my-go-app"
    echo "   3. Initialize module: go mod init my-go-app"
    echo "   4. Create main.go with web server"
    echo "   5. Container will auto-build and start app"
    echo ""
    echo "⏳ Container will stay alive (no apps to run)..."
    sleep infinity
else
    echo "✅ Successfully started $app_count Go application(s)"
    echo "🔍 Process IDs: ${pids[*]}"
    echo ""
    echo "📊 Port mapping:"
    echo "   External (host) → Internal (container)"
    for ((i=0; i<app_count; i++)); do
        external_port=$((8400 + i))
        internal_port=$((8080 + i))
        echo "   localhost:$external_port → :$internal_port (app $((i+1)))"
    done
    echo ""
    echo "🔧 Container commands:"
    echo "   devarch exec go ps aux | grep go     - List running Go apps"
    echo "   devarch logs go -f                   - Follow container logs"
    echo "   devarch restart go                   - Rebuild and restart all apps"
    echo ""
    echo "⏳ Keeping container alive..."
    echo "   All apps running in background"
    echo "   Press Ctrl+C to stop all apps"
    echo ""
fi

# Function to kill all child processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down Go applications..."
    for pid in "${pids[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "   Stopping PID $pid..."
            kill "$pid" 2>/dev/null || true
        fi
    done
    echo "✅ All Go apps stopped"
    exit 0
}

trap cleanup SIGTERM SIGINT

# Wait for all background processes
wait
