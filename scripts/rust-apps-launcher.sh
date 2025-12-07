#!/bin/bash
# Discovers and runs Rust applications in /app directory
# Auto-discovers apps with Cargo.toml, builds in release mode, and runs binaries

set -e

APPS_DIR="/app"
BASE_PORT=8080

echo "🦀 Rust Apps Launcher - DevArch Multi-Language Backend"
echo "========================================================"
echo "🔍 Discovering Rust applications in $APPS_DIR..."
echo ""

port=$BASE_PORT
app_count=0
pids=()

for app_dir in "$APPS_DIR"/*/; do
    app_name=$(basename "$app_dir")

    # Skip non-Rust apps (check for Cargo.toml)
    if [ ! -f "$app_dir/Cargo.toml" ]; then
        continue
    fi

    echo "📦 Found Rust app: $app_name"
    echo "   Location: $app_dir"
    echo "   Building in release mode..."

    cd "$app_dir"

    # Build in release mode for better performance
    if cargo build --release 2>&1 | grep -E '(Finished|error)'; then
        # Extract binary name from Cargo.toml
        binary_name=$(grep -E '^name = ' Cargo.toml | head -1 | sed 's/name = "\(.*\)"/\1/')
        binary_path="target/release/$binary_name"

        if [ ! -f "$binary_path" ]; then
            echo "   ❌ Build failed - binary not found: $binary_path"
            echo ""
            continue
        fi

        echo "   ✅ Build successful: $binary_name"
        echo "   🚀 Starting on port $port..."

        # Run in background with port override
        PORT=$port ./"$binary_path" > "/app/logs/$app_name.log" 2>&1 &
        pid=$!
        pids+=($pid)

        echo "   ✓ Started with PID $pid"
        echo "   📍 Access: http://localhost:8700 (external) → :$port (internal)"
        echo "   📝 Logs: /app/logs/$app_name.log"
        echo ""

        ((app_count++))
        ((port++))
    else
        echo "   ❌ Build failed for $app_name"
        echo ""
        continue
    fi
done

echo "========================================================"
if [ $app_count -eq 0 ]; then
    echo "⚠️  No Rust applications found in $APPS_DIR"
    echo ""
    echo "💡 To create a Rust app:"
    echo "   1. Use RustRover IDE: File → New Project → Rust Binary"
    echo "   2. Location: $APPS_DIR/my-rust-app"
    echo "   3. Restart container: devarch restart rust"
    echo ""
    echo "⏳ Container will stay alive (no apps to run)..."
    sleep infinity
else
    echo "✅ Successfully started $app_count Rust application(s)"
    echo "🔍 Process IDs: ${pids[*]}"
    echo ""
    echo "📊 Port mapping:"
    echo "   External (host) → Internal (container)"
    echo "   localhost:8700  → :8080  (app 1)"
    echo "   localhost:8701  → :8081  (app 2)"
    echo "   localhost:8702  → :8082  (debugger)"
    echo "   localhost:8703  → :8083  (metrics)"
    echo ""
    echo "⏳ Keeping container alive..."
    echo "   All apps running in background"
    echo "   Press Ctrl+C to stop all apps"
    echo ""
fi

# Function to kill all child processes on exit
cleanup() {
    echo ""
    echo "🛑 Shutting down Rust applications..."
    for pid in "${pids[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "   Stopping PID $pid..."
            kill "$pid" 2>/dev/null || true
        fi
    done
    echo "✅ All Rust apps stopped"
    exit 0
}

trap cleanup SIGTERM SIGINT

# Wait for all background processes
wait
