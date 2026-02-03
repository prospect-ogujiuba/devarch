#!/bin/zsh

# =============================================================================
# PODMAN SOCKET MANAGER - REFACTORED WITH CONFIG INTEGRATION
# =============================================================================
# Enhanced socket management with config.sh integration and ecosystem awareness

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# ENHANCED SOCKET STATUS FUNCTIONS
# =============================================================================

cmd_status() {
    print_status "info" "Podman Socket Status"
    echo ""
    
    echo "Environment:"
    echo "  User: $(whoami) ($(id -u))"
    echo "  Project: $PROJECT_ROOT"
    echo "  Network: $NETWORK_NAME"
    echo ""
    
    echo "Socket Status:"
    local rootless_active=$(systemctl --user is-active podman.socket 2>/dev/null || echo "inactive")
    local rootful_active=$(sudo systemctl is-active podman.socket 2>/dev/null || echo "inactive")
    
    if [[ "$rootless_active" == "active" ]]; then
        print_status "success" "Rootless socket: $rootless_active"
    else
        print_status "warning" "Rootless socket: $rootless_active"
    fi
    
    if [[ "$rootful_active" == "active" ]]; then
        print_status "success" "Rootful socket: $rootful_active"
    else
        print_status "info" "Rootful socket: $rootful_active"
    fi
    
    echo ""
    echo "Socket Files:"
    local rootless_socket="/run/user/$(id -u)/podman/podman.sock"
    local rootful_socket="/run/podman/podman.sock"
    
    if [[ -S "$rootless_socket" ]]; then
        print_status "success" "Rootless file: exists ($rootless_socket)"
    else
        print_status "error" "Rootless file: missing ($rootless_socket)"
    fi
    
    if sudo test -S "$rootful_socket"; then
        print_status "success" "Rootful file: exists ($rootful_socket)"
    else
        print_status "info" "Rootful file: missing ($rootful_socket)"
    fi
    
    echo ""
    echo "Environment:"
    if [[ -n "$DOCKER_HOST" ]]; then
        print_status "info" "DOCKER_HOST: $DOCKER_HOST"
    else
        print_status "warning" "DOCKER_HOST: not set"
    fi
    
    echo ""
    test_socket_connectivity
    show_microservice_status
}

test_socket_connectivity() {
    echo "Connectivity Tests:"
    
    local rootless_socket="/run/user/$(id -u)/podman/podman.sock"
    if [[ -S "$rootless_socket" ]]; then
        if DOCKER_HOST="unix://$rootless_socket" podman version >/dev/null 2>&1; then
            print_status "success" "Rootless connectivity: working"
        else
            print_status "error" "Rootless connectivity: broken"
        fi
    else
        print_status "info" "Rootless connectivity: socket not found"
    fi
    
    if sudo test -S /run/podman/podman.sock; then
        if sudo DOCKER_HOST="unix:///run/podman/podman.sock" podman version >/dev/null 2>&1; then
            print_status "success" "Rootful connectivity: working"
        else
            print_status "error" "Rootful connectivity: broken"
        fi
    else
        print_status "info" "Rootful connectivity: socket not found"
    fi
}

show_microservice_status() {
    echo ""
    echo "Microservice Integration:"
    
    # Check if service-manager is available
    local service_manager="$SCRIPT_DIR/service-manager.sh"
    if [[ -f "$service_manager" ]]; then
        print_status "success" "Service-manager: available"
        
        # Show running services count if any
        local running_count
        running_count=$(podman ps --filter "network=$NETWORK_NAME" -q 2>/dev/null | wc -l || echo "0")
        
        if [[ "$running_count" -gt 0 ]]; then
            print_status "info" "Running services: $running_count containers"
            echo "  Use: ./scripts/service-manager.sh status"
        else
            print_status "info" "Running services: none"
        fi
    else
        print_status "warning" "Service-manager: not found"
    fi
    
    # Check network status
    if podman network exists "$NETWORK_NAME" 2>/dev/null; then
        print_status "success" "Project network: $NETWORK_NAME exists"
    else
        print_status "info" "Project network: $NETWORK_NAME not found"
    fi
}

# =============================================================================
# ENHANCED SOCKET MANAGEMENT
# =============================================================================

cmd_start_rootless() {
    print_status "step" "Starting rootless socket..."
    
    # Stop rootful first to avoid conflicts
    print_status "info" "Stopping rootful socket to avoid conflicts..."
    sudo systemctl stop podman.socket 2>/dev/null || true
    
    # Enable lingering for persistent user services
    print_status "step" "Enabling user lingering..."
    sudo loginctl enable-linger $(whoami)
    
    # Start rootless socket
    print_status "step" "Starting rootless podman socket..."
    systemctl --user enable --now podman.socket
    
    # Wait for socket to be ready
    print_status "step" "Waiting for socket to be ready..."
    local counter=0
    local max_wait=10
    
    while [[ $counter -lt $max_wait ]]; do
        if [[ -S "/run/user/$(id -u)/podman/podman.sock" ]]; then
            break
        fi
        sleep 1
        ((counter++))
    done
    
    if [[ -S "/run/user/$(id -u)/podman/podman.sock" ]]; then
        print_status "success" "Rootless socket started successfully"
        
        # Test connectivity
        if DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock" podman version >/dev/null 2>&1; then
            print_status "success" "Socket connectivity verified"
        else
            print_status "warning" "Socket file exists but connectivity test failed"
        fi
        
        # Update environment guidance
        echo ""
        print_status "info" "Environment setup:"
        echo "  export DOCKER_HOST=\"unix:///run/user/$(id -u)/podman/podman.sock\""
        echo ""
        print_status "info" "Or add to ~/.zshrc for persistence:"
        echo "  echo 'export DOCKER_HOST=\"unix:///run/user/\$(id -u)/podman/podman.sock\"' >> ~/.zshrc"
        
        return 0
    else
        print_status "error" "Socket failed to start within $max_wait seconds"
        print_status "info" "Check logs: journalctl --user -u podman.socket"
        return 1
    fi
}

cmd_start_rootful() {
    print_status "step" "Starting rootful socket..."
    
    # Stop rootless first to avoid conflicts
    print_status "info" "Stopping rootless socket to avoid conflicts..."
    systemctl --user stop podman.socket 2>/dev/null || true
    
    # Start rootful socket
    print_status "step" "Starting rootful podman socket..."
    sudo systemctl enable --now podman.socket
    
    # Wait for socket to be ready
    print_status "step" "Waiting for socket to be ready..."
    local counter=0
    local max_wait=10
    
    while [[ $counter -lt $max_wait ]]; do
        if sudo test -S /run/podman/podman.sock; then
            break
        fi
        sleep 1
        ((counter++))
    done
    
    if sudo test -S /run/podman/podman.sock; then
        print_status "success" "Rootful socket started successfully"
        
        # Test connectivity
        if sudo DOCKER_HOST="unix:///run/podman/podman.sock" podman version >/dev/null 2>&1; then
            print_status "success" "Socket connectivity verified"
        else
            print_status "warning" "Socket file exists but connectivity test failed"
        fi
        
        # Update environment guidance
        echo ""
        print_status "info" "Environment setup:"
        echo "  export DOCKER_HOST=\"unix:///run/podman/podman.sock\""
        echo ""
        print_status "warning" "Note: Rootful requires sudo for most operations"
        
        return 0
    else
        print_status "error" "Socket failed to start within $max_wait seconds"
        print_status "info" "Check logs: sudo journalctl -u podman.socket"
        return 1
    fi
}

cmd_stop() {
    print_status "step" "Stopping all podman sockets..."
    
    # Stop both socket types
    print_status "info" "Stopping rootless socket..."
    systemctl --user stop podman.socket 2>/dev/null || true
    
    print_status "info" "Stopping rootful socket..."
    sudo systemctl stop podman.socket 2>/dev/null || true
    
    # Wait a moment for proper shutdown
    sleep 2
    
    # Verify they're stopped
    local rootless_stopped=false rootful_stopped=false
    
    if [[ "$(systemctl --user is-active podman.socket 2>/dev/null)" != "active" ]]; then
        rootless_stopped=true
    fi
    
    if [[ "$(sudo systemctl is-active podman.socket 2>/dev/null)" != "active" ]]; then
        rootful_stopped=true
    fi
    
    if [[ "$rootless_stopped" == "true" && "$rootful_stopped" == "true" ]]; then
        print_status "success" "All sockets stopped successfully"
    else
        print_status "warning" "Some sockets may still be active"
        print_status "info" "Check status: $0 status"
    fi
}

cmd_test() {
    print_status "info" "Testing socket connectivity..."
    echo ""
    
    local rootless_socket="/run/user/$(id -u)/podman/podman.sock"
    local rootful_socket="/run/podman/podman.sock"
    
    # Test rootless socket
    if [[ -S "$rootless_socket" ]]; then
        print_status "step" "Testing rootless socket..."
        if DOCKER_HOST="unix://$rootless_socket" podman version >/dev/null 2>&1; then
            print_status "success" "Rootless socket responds correctly"
            
            # Test with our project network
            if DOCKER_HOST="unix://$rootless_socket" podman network exists "$NETWORK_NAME" 2>/dev/null; then
                print_status "success" "Project network '$NETWORK_NAME' accessible"
            else
                print_status "info" "Project network '$NETWORK_NAME' not found (create with: service-manager.sh)"
            fi
        else
            print_status "error" "Rootless socket not responding"
        fi
    else
        print_status "info" "Rootless socket not found"
    fi
    
    echo ""
    
    # Test rootful socket
    if sudo test -S "$rootful_socket"; then
        print_status "step" "Testing rootful socket..."
        if sudo DOCKER_HOST="unix://$rootful_socket" podman version >/dev/null 2>&1; then
            print_status "success" "Rootful socket responds correctly"
        else
            print_status "error" "Rootful socket not responding"
        fi
    else
        print_status "info" "Rootful socket not found"
    fi
    
    echo ""
    
    # Integration test with service-manager
    local service_manager="$SCRIPT_DIR/service-manager.sh"
    if [[ -f "$service_manager" ]]; then
        print_status "step" "Testing service-manager integration..."
        if "$service_manager" list >/dev/null 2>&1; then
            print_status "success" "Service-manager integration working"
        else
            print_status "warning" "Service-manager integration issues"
            print_status "info" "Try: $service_manager --help"
        fi
    else
        print_status "info" "Service-manager not found, skipping integration test"
    fi
}

cmd_logs() {
    local type="${1:-rootless}"
    
    case "$type" in
        rootless)
            print_status "info" "Showing rootless socket logs (Ctrl+C to exit)..."
            journalctl --user -u podman.socket -f
            ;;
        rootful)
            print_status "info" "Showing rootful socket logs (Ctrl+C to exit)..."
            sudo journalctl -u podman.socket -f
            ;;
        *)
            print_status "error" "Invalid log type: $type"
            print_status "info" "Usage: $0 logs [rootless|rootful]"
            ;;
    esac
}

cmd_fix() {
    print_status "step" "Fixing broken sockets..."
    
    # Kill conflicting processes
    print_status "info" "Stopping conflicting podman processes..."
    pkill -f "podman system service" 2>/dev/null || true
    
    # Stop all sockets cleanly
    print_status "info" "Stopping all sockets..."
    systemctl --user stop podman.socket 2>/dev/null || true
    sudo systemctl stop podman.socket 2>/dev/null || true
    
    # Wait for proper shutdown
    sleep 3
    
    # Fix permissions if sockets exist
    print_status "info" "Checking and fixing permissions..."
    local rootless_socket="/run/user/$(id -u)/podman/podman.sock"
    if [[ -S "$rootless_socket" ]]; then
        chmod 660 "$rootless_socket" 2>/dev/null || true
        print_status "info" "Fixed rootless socket permissions"
    fi
    
    if sudo test -S /run/podman/podman.sock; then
        sudo chmod 660 /run/podman/podman.sock 2>/dev/null || true
        print_status "info" "Fixed rootful socket permissions"
    fi
    
    # Restart rootless socket (recommended default)
    print_status "step" "Restarting rootless socket..."
    systemctl --user start podman.socket
    
    # Wait and test
    sleep 3
    print_status "step" "Testing fixed socket..."
    cmd_test
}

cmd_nuke() {
    print_status "warning" "Nuclear Reset - This will destroy everything and start fresh!"
    echo "This will:"
    echo "  - Stop all podman processes"
    echo "  - Remove all socket files"  
    echo "  - Reset podman configuration"
    echo "  - Start fresh rootless socket"
    echo ""
    
    if [[ -t 0 ]]; then
        read "response?Are you sure? Type 'NUKE' to confirm: "
        if [[ "$response" != "NUKE" ]]; then
            print_status "info" "Nuclear reset cancelled"
            return 0
        fi
    else
        print_status "error" "Nuclear reset requires interactive confirmation"
        return 1
    fi
    
    print_status "step" "Initiating nuclear reset..."
    
    # Stop everything
    print_status "info" "Stopping all podman services..."
    systemctl --user stop podman.socket 2>/dev/null || true
    sudo systemctl stop podman.socket 2>/dev/null || true
    pkill -f podman 2>/dev/null || true
    
    # Remove socket files
    print_status "info" "Removing socket files..."
    rm -rf "/run/user/$(id -u)/podman/" 2>/dev/null || true
    sudo rm -rf /run/podman/ 2>/dev/null || true
    
    # Wait for cleanup
    print_status "step" "Waiting for cleanup to complete..."
    sleep 5
    
    # Start fresh
    print_status "step" "Starting fresh rootless socket..."
    sudo loginctl enable-linger $(whoami)
    systemctl --user enable --now podman.socket
    
    # Wait and test
    sleep 3
    print_status "step" "Testing nuclear reset results..."
    cmd_test
    
    print_status "success" "Nuclear reset completed!"
    print_status "info" "You can now start services: ./scripts/service-manager.sh start-all"
}

cmd_env() {
    print_status "info" "Environment Setup Guide"
    echo ""
    
    # Detect active socket
    local active_socket="none"
    if systemctl --user is-active podman.socket >/dev/null 2>&1; then
        active_socket="rootless"
    elif sudo systemctl is-active podman.socket >/dev/null 2>&1; then
        active_socket="rootful"
    fi
    
    case "$active_socket" in
        "rootless")
            print_status "success" "Rootless socket is active (recommended)"
            echo ""
            echo "Current session:"
            echo "  export DOCKER_HOST=\"unix:///run/user/\$(id -u)/podman/podman.sock\""
            echo ""
            echo "Permanent setup (add to ~/.zshrc):"
            echo "  echo 'export DOCKER_HOST=\"unix:///run/user/\$(id -u)/podman/podman.sock\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
            ;;
        "rootful")
            print_status "info" "Rootful socket is active"
            echo ""
            echo "Current session:"
            echo "  export DOCKER_HOST=\"unix:///run/podman/podman.sock\""
            echo ""
            echo "Permanent setup (add to ~/.zshrc):"
            echo "  echo 'export DOCKER_HOST=\"unix:///run/podman/podman.sock\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
            ;;
        "none")
            print_status "warning" "No active sockets found"
            echo ""
            echo "Start a socket first:"
            echo "  $0 start-rootless    # Recommended"
            echo "  $0 start-rootful     # Alternative"
            ;;
    esac
    
    echo ""
    print_status "info" "Integration Commands:"
    echo "  ./scripts/service-manager.sh start-all    # Start all services"
    echo "  ./scripts/service-manager.sh status       # Check service status"
    echo "  ./scripts/runtime-switcher.sh status      # Check runtime status"
}

# =============================================================================
# USAGE AND HELP
# =============================================================================

usage() {
    cat << EOF
Podman Socket Manager - Enhanced with Config Integration

COMMANDS:
  status           Show socket status and connectivity
  start-rootless   Start rootless socket (recommended)
  start-rootful    Start rootful socket (needs sudo)
  stop             Stop all sockets
  test             Test socket connectivity and integration
  logs [type]      Show socket logs (rootless/rootful)
  fix              Try to fix broken sockets
  nuke             Nuclear reset - destroy and rebuild
  env              Show environment setup guide
  help             Show this help message

SHORTCUTS:
  s                status
  sr               start-rootless
  sf               start-rootful
  t                test
  l [type]         logs
  f                fix
  n                nuke
  e                env

EXAMPLES:
  $0 status                # Check what's running
  $0 start-rootless       # Start rootless (preferred)
  $0 test                 # Test if sockets work
  $0 fix                  # Fix broken sockets
  $0 logs rootless        # Show logs

INTEGRATION:
  - Uses config.sh for consistent messaging and paths
  - Integrates with service-manager for microservice status
  - Shows project network ($NETWORK_NAME) status
  - Provides environment setup for the full ecosystem

NOTES:
  - Rootless is recommended for development
  - Rootful requires sudo for most operations
  - Use 'nuke' for complete reset when things go wrong
  - Environment setup persists across shell sessions

EOF
}

# =============================================================================
# MAIN COMMAND DISPATCH
# =============================================================================

main() {
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
        help|h|-h|--help)   usage ;;
        *)                  
            print_status "error" "Unknown command: $1"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi