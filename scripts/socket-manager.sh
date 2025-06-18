#!/bin/zsh

# =============================================================================
# PODMAN SOCKET MANAGER - NO BULLSHIT VERSION (ZSH OPTIMIZED)
# =============================================================================

cmd_status() {
    echo "=== SOCKET STATUS ==="
    echo "User: $(whoami) ($(id -u))"
    echo "Rootless socket: $(systemctl --user is-active podman.socket)"
    echo "Rootful socket: $(sudo systemctl is-active podman.socket 2>/dev/null || echo 'inactive')"
    echo "Rootless file: $(test -S /run/user/$(id -u)/podman/podman.sock && echo 'exists' || echo '❌ MISSING')"
    echo "Rootful file: $(sudo test -S /run/podman/podman.sock && echo 'exists' || echo '❌ MISSING')"
    echo "DOCKER_HOST: ${DOCKER_HOST:-not set}"
    echo ""
    
    # Quick connectivity test
    if test -S /run/user/$(id -u)/podman/podman.sock; then
        if DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock" podman version >/dev/null 2>&1; then
            echo "Rootless connectivity: WORKING"
        else
            echo "Rootless connectivity: BROKEN"
        fi
    fi
    
    if sudo test -S /run/podman/podman.sock; then
        if sudo DOCKER_HOST="unix:///run/podman/podman.sock" podman version >/dev/null 2>&1; then
            echo "Rootful connectivity: WORKING"
        else
            echo "Rootful connectivity: BROKEN"
        fi
    fi
}

cmd_start_rootless() {
    print "Starting rootless socket..."
    
    # Stop rootful first
    sudo systemctl stop podman.socket 2>/dev/null
    
    # Enable lingering
    sudo loginctl enable-linger $(whoami)
    
    # Start rootless
    systemctl --user enable --now podman.socket
    
    print "Waiting for socket..."
    sleep 2
    
    if [[ -S /run/user/$(id -u)/podman/podman.sock ]]; then
        print "✅ Rootless socket started successfully"
        print "Export this: export DOCKER_HOST=\"unix:///run/user/$(id -u)/podman/podman.sock\""
    else
        print "❌ Socket failed to start"
        return 1
    fi
}

cmd_start_rootful() {
    print "Starting rootful socket..."
    
    # Stop rootless first
    systemctl --user stop podman.socket 2>/dev/null
    
    # Start rootful
    sudo systemctl enable --now podman.socket
    
    print "Waiting for socket..."
    sleep 2
    
    if sudo test -S /run/podman/podman.sock; then
        print "✅ Rootful socket started successfully"
        print "Export this: export DOCKER_HOST=\"unix:///run/podman/podman.sock\""
    else
        print "❌ Socket failed to start"
        return 1
    fi
}

cmd_stop() {
    print "Stopping all sockets..."
    systemctl --user stop podman.socket 2>/dev/null
    sudo systemctl stop podman.socket 2>/dev/null
    print "✅ All sockets stopped"
}

cmd_test() {
    print "=== TESTING SOCKETS ==="
    
    if [[ -S /run/user/$(id -u)/podman/podman.sock ]]; then
        print "Testing rootless..."
        if DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock" podman version >/dev/null 2>&1; then
            print "✅ Rootless socket responds"
        else
            print "❌ Rootless socket not responding"
        fi
    else
        print "⚪ Rootless socket not found"
    fi
    
    if sudo test -S /run/podman/podman.sock; then
        print "Testing rootful..."
        if sudo DOCKER_HOST="unix:///run/podman/podman.sock" podman version >/dev/null 2>&1; then
            print "✅ Rootful socket responds"
        else
            print "❌ Rootful socket not responding"
        fi
    else
        print "⚪ Rootful socket not found"
    fi
}

cmd_logs() {
    local type="${1:-rootless}"
    print "Showing $type socket logs (Ctrl+C to exit)..."
    
    case "$type" in
        rootless)
            journalctl --user -u podman.socket -f
            ;;
        rootful)
            sudo journalctl -u podman.socket -f
            ;;
        *)
            print "Usage: logs [rootless|rootful]"
            ;;
    esac
}

cmd_fix() {
    print "=== FIXING BROKEN SOCKETS ==="
    
    print "Killing conflicting processes..."
    pkill -f "podman system service" 2>/dev/null || true
    
    print "Stopping all sockets..."
    systemctl --user stop podman.socket 2>/dev/null
    sudo systemctl stop podman.socket 2>/dev/null
    
    print "Checking permissions..."
    if [[ -S /run/user/$(id -u)/podman/podman.sock ]]; then
        chmod 660 /run/user/$(id -u)/podman/podman.sock
        print "Fixed rootless permissions"
    fi
    
    if sudo test -S /run/podman/podman.sock; then
        sudo chmod 660 /run/podman/podman.sock
        print "Fixed rootful permissions"
    fi
    
    print "Restarting rootless socket..."
    systemctl --user start podman.socket
    
    sleep 2
    cmd_test
}

cmd_nuke() {
    print "=== NUCLEAR RESET ==="
    print "This will destroy everything and start fresh!"
    read -q "?Are you sure? (y/N): "
    print
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print "Cancelled"
        return 0
    fi
    
    print "Stopping everything..."
    systemctl --user stop podman.socket 2>/dev/null
    sudo systemctl stop podman.socket 2>/dev/null
    pkill -f podman 2>/dev/null || true
    
    print "Removing socket files..."
    rm -rf /run/user/$(id -u)/podman/ 2>/dev/null || true
    sudo rm -rf /run/podman/ 2>/dev/null || true
    
    print "Waiting 5 seconds..."
    sleep 5
    
    print "Starting fresh rootless socket..."
    sudo loginctl enable-linger $(whoami)
    systemctl --user enable --now podman.socket
    
    sleep 3
    cmd_test
}

cmd_env() {
    print "=== ENVIRONMENT SETUP ==="
    
    if systemctl --user is-active podman.socket >/dev/null 2>&1; then
        print "Rootless socket is active. Add to ~/.zshrc:"
        print 'export DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock"'
        print ""
        print "Or run now:"
        print "export DOCKER_HOST=\"unix:///run/user/$(id -u)/podman/podman.sock\""
    elif sudo systemctl is-active podman.socket >/dev/null 2>&1; then
        print "Rootful socket is active. Add to ~/.zshrc:"
        print 'export DOCKER_HOST="unix:///run/podman/podman.sock"'
        print ""
        print "Or run now:"
        print "export DOCKER_HOST=\"unix:///run/podman/podman.sock\""
    else
        print "No active sockets found. Start one first:"
        print "$0 start-rootless"
    fi
}

cmd_traefik() {
    print "=== TRAEFIK INTEGRATION ==="
    
    if systemctl --user is-active podman.socket >/dev/null 2>&1; then
        print "Using rootless socket. Add to traefik compose volumes:"
        print "  - /run/user/$(id -u)/podman/podman.sock:/var/run/docker.sock:ro"
    elif sudo systemctl is-active podman.socket >/dev/null 2>&1; then
        print "Using rootful socket. Add to traefik compose volumes:"
        print "  - /run/podman/podman.sock:/var/run/docker.sock:ro"
    else
        print "No active sockets. Start one first:"
        print "$0 start-rootless"
        return 1
    fi
    
    print ""
    print "Test traefik can see the socket:"
    print "podman exec traefik ls -la /var/run/docker.sock"
}

usage() {
    cat << EOF
Podman Socket Manager

COMMANDS:
  status           Show socket status and connectivity
  start-rootless   Start rootless socket (recommended)
  start-rootful    Start rootful socket (needs sudo)
  stop             Stop all sockets
  test             Test socket connectivity
  logs [type]      Show socket logs (rootless/rootful)
  fix              Try to fix broken sockets
  nuke             Nuclear reset - destroy and rebuild
  env              Show environment setup
  traefik          Show traefik integration info

EXAMPLES:
  $0 status                # Check what's running
  $0 start-rootless       # Start rootless (preferred)
  $0 test                 # Test if sockets work
  $0 fix                  # Fix broken sockets
  $0 logs rootless        # Show logs

EOF
}

# Main command dispatch
case "${1:-status}" in
    status|s)           cmd_status ;;
    start-rootless|sr)  cmd_start_rootless ;;
    start-rootful|sf)   cmd_start_rootful ;;
    stop)               cmd_stop ;;
    test|t)             cmd_test ;;
    logs|l)             cmd_logs "$2" ;;
    fix|f)              cmd_fix ;;
    nuke|n)             cmd_nuke ;;
    env|e)              cmd_env ;;
    traefik|tr)         cmd_traefik ;;
    help|h|-h|--help)   usage ;;
    *)                  echo "Unknown command: $1"; usage; exit 1 ;;
esac