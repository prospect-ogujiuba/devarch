#!/bin/zsh

# =============================================================================
# CONTAINER RUNTIME SWITCHER - REFACTORED
# =============================================================================
# Switches between Podman and Docker with enhanced service-manager integration

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# ENHANCED RUNTIME DETECTION
# =============================================================================

get_current_runtime() {
    # Use config.sh detection first
    detect_sudo_requirement
    
    # Check Docker first (more reliable detection)
    if command -v docker >/dev/null 2>&1; then
        # Check if Docker daemon is running
        if sudo systemctl is-active docker >/dev/null 2>&1 || pgrep -f "dockerd" >/dev/null 2>&1; then
            # Try a simple docker command to confirm it's working
            if sudo docker info >/dev/null 2>&1; then
                echo "docker"
                return 0
            fi
        fi
    fi
    
    # Check Podman
    if command -v podman >/dev/null 2>&1; then
        # Check if podman socket is active or podman is responsive
        if systemctl --user is-active podman.socket >/dev/null 2>&1 || podman info >/dev/null 2>&1; then
            echo "podman"
            return 0
        fi
    fi
    
    echo "none"
    return 1
}

# =============================================================================
# ENHANCED SERVICE MANAGEMENT
# =============================================================================

stop_all_services() {
    print_status "info" "Stopping all microservices before runtime switch..."
    
    # Use service-manager for robust service shutdown
    local service_manager="$SCRIPT_DIR/service-manager.sh"
    if [[ -f "$service_manager" ]]; then
        print_status "step" "Using service-manager for clean shutdown..."
        
        # Use service-manager with cleanup options for thorough shutdown
        if "$service_manager" stop-all --timeout 10 --preserve-data --cleanup-orphans 2>/dev/null; then
            print_status "success" "Services stopped cleanly via service-manager"
        else
            print_status "warning" "Service-manager shutdown had issues, falling back to manual cleanup"
            cleanup_runtime_containers
        fi
    else
        print_status "warning" "service-manager.sh not found, using manual cleanup"
        cleanup_runtime_containers
    fi
}

cleanup_runtime_containers() {
    print_status "step" "Performing manual container cleanup..."
    
    # Clean up both runtimes to ensure clean state
    cleanup_runtime "podman"
    cleanup_runtime "docker"
}

cleanup_runtime() {
    local runtime="$1"
    
    if [[ "$runtime" == "podman" ]]; then
        print_status "info" "Cleaning up podman resources..."
        
        # Stop all podman containers
        local containers=$(podman ps -q 2>/dev/null || echo "")
        if [[ -n "$containers" ]]; then
            local container_list=(${(f)containers})
            print_status "step" "Stopping ${#container_list[@]} podman container(s)..."
            podman stop ${container_list[@]} 2>/dev/null || true
            podman rm ${container_list[@]} 2>/dev/null || true
        fi
        
        # Remove networks (preserve data by default)
        podman network rm $NETWORK_NAME 2>/dev/null || true
        
    elif [[ "$runtime" == "docker" ]]; then
        print_status "info" "Cleaning up docker resources..."
        
        # Stop all docker containers
        local containers=$(sudo docker ps -q 2>/dev/null || echo "")
        if [[ -n "$containers" ]]; then
            local container_list=(${(f)containers})
            print_status "step" "Stopping ${#container_list[@]} docker container(s)..."
            sudo docker stop ${container_list[@]} 2>/dev/null || true
            sudo docker rm ${container_list[@]} 2>/dev/null || true
        fi
        
        # Remove networks (preserve data by default)
        sudo docker network rm $NETWORK_NAME 2>/dev/null || true
    fi
}

# =============================================================================
# RUNTIME SWITCHING FUNCTIONS
# =============================================================================

switch_to_docker() {
    print_status "info" "Switching to Docker..."
    
    # Stop podman services
    if command -v podman >/dev/null 2>&1; then
        print_status "step" "Stopping podman services..."
        systemctl --user stop podman.socket 2>/dev/null || true
        pkill -f podman 2>/dev/null || true
        sleep 2  # Give it time to stop
    fi
    
    # Start Docker
    if command -v docker >/dev/null 2>&1; then
        print_status "step" "Starting docker daemon..."
        sudo systemctl start docker
        sleep 3  # Give Docker time to start
        
        # Verify Docker is working
        if sudo docker info >/dev/null 2>&1; then
            # Update config.sh using proper replacement
            update_config_runtime "docker"
            
            print_status "success" "Docker is now active"
            return 0
        else
            print_status "error" "Docker failed to start properly"
            print_status "info" "Check docker service: sudo systemctl status docker"
            return 1
        fi
    else
        print_status "error" "Docker not found. Please install Docker first."
        print_status "info" "Install guide: https://docs.docker.com/engine/install/"
        return 1
    fi
}

switch_to_podman() {
    print_status "info" "Switching to Podman..."
    
    # Stop Docker
    if command -v docker >/dev/null 2>&1; then
        print_status "step" "Stopping docker daemon..."
        sudo systemctl stop docker 2>/dev/null || true
        pkill -f dockerd 2>/dev/null || true
        sleep 2  # Give it time to stop
    fi
    
    # Start Podman
    if command -v podman >/dev/null 2>&1; then
        print_status "step" "Starting podman socket..."
        systemctl --user start podman.socket 2>/dev/null || true
        sleep 2  # Give Podman time to start
        
        # Verify Podman is working
        if podman info >/dev/null 2>&1; then
            # Update config.sh using proper replacement
            update_config_runtime "podman"
            
            print_status "success" "Podman is now active"
            return 0
        else
            print_status "error" "Podman failed to start properly"
            print_status "info" "Check podman socket: systemctl --user status podman.socket"
            return 1
        fi
    else
        print_status "error" "Podman not found. Please install Podman first."
        print_status "info" "Install guide: https://podman.io/getting-started/installation"
        return 1
    fi
}

# =============================================================================
# CONFIG FILE MANAGEMENT
# =============================================================================

update_config_runtime() {
    local new_runtime="$1"
    local config_file="$SCRIPT_DIR/config.sh"
    
    if [[ ! -f "$config_file" ]]; then
        print_status "warning" "Config file not found: $config_file"
        return 1
    fi
    
    print_status "step" "Updating config.sh for $new_runtime runtime..."
    
    case "$new_runtime" in
        "docker")
            sed -i 's/export CONTAINER_RUNTIME="podman"/export CONTAINER_RUNTIME="docker"/' "$config_file"
            sed -i 's/export USE_PODMAN=true/export USE_PODMAN=false/' "$config_file"
            ;;
        "podman")
            sed -i 's/export CONTAINER_RUNTIME="docker"/export CONTAINER_RUNTIME="podman"/' "$config_file"
            sed -i 's/export USE_PODMAN=false/export USE_PODMAN=true/' "$config_file"
            ;;
        *)
            print_status "error" "Unknown runtime: $new_runtime"
            return 1
            ;;
    esac
    
    print_status "success" "Config updated for $new_runtime"
    
    # Reload configuration
    . "$config_file"
}

# =============================================================================
# STATUS AND INFORMATION FUNCTIONS
# =============================================================================

show_status() {
    local current=$(get_current_runtime)
    
    echo "=== Container Runtime Status ==="
    echo ""
    
    case "$current" in
        "podman")
            print_status "success" "Currently using: Podman"
            ;;
        "docker")
            print_status "success" "Currently using: Docker"
            ;;
        "none")
            print_status "warning" "No active container runtime detected"
            ;;
    esac
    
    echo ""
    echo "Available runtimes:"
    if command -v podman >/dev/null 2>&1; then
        print_status "success" "✓ Podman installed"
    else
        print_status "error" "✗ Podman not installed"
    fi
    
    if command -v docker >/dev/null 2>&1; then
        print_status "success" "✓ Docker installed"
    else
        print_status "error" "✗ Docker not installed"
    fi
    
    echo ""
    show_container_status
    show_service_status
}

show_container_status() {
    echo "Container Status:"
    
    # Show Podman containers using config.sh approach
    if command -v podman >/dev/null 2>&1; then
        local podman_count=$(podman ps -q 2>/dev/null | wc -l || echo "0")
        if [[ "$podman_count" -gt 0 ]]; then
            print_status "warning" "⚠️  $podman_count containers running in Podman"
        else
            print_status "success" "✓ No containers running in Podman"
        fi
    fi
    
    # Show Docker containers
    if command -v docker >/dev/null 2>&1; then
        local docker_count=$(sudo docker ps -q 2>/dev/null | wc -l || echo "0")
        if [[ "$docker_count" -gt 0 ]]; then
            print_status "warning" "⚠️  $docker_count containers running in Docker"
        else
            print_status "success" "✓ No containers running in Docker"
        fi
    fi
    
    echo ""
}

show_service_status() {
    echo "Microservice Status:"
    
    # Use service-manager if available for better service status
    local service_manager="$SCRIPT_DIR/service-manager.sh"
    if [[ -f "$service_manager" ]]; then
        local running_services=$("$service_manager" ps 2>/dev/null | grep -c "Up" || echo "0")
        if [[ "$running_services" -gt 0 ]]; then
            print_status "info" "📦 $running_services microservice(s) running"
            print_status "info" "Run './scripts/service-manager.sh status' for details"
        else
            print_status "success" "✓ No microservices currently running"
        fi
    else
        # Fallback: check for containers in our network
        local network_containers
        if [[ "$current" == "podman" ]]; then
            network_containers=$(podman ps --filter "network=$NETWORK_NAME" -q 2>/dev/null | wc -l || echo "0")
        elif [[ "$current" == "docker" ]]; then
            network_containers=$(sudo docker ps --filter "network=$NETWORK_NAME" -q 2>/dev/null | wc -l || echo "0")
        else
            network_containers="0"
        fi
        
        if [[ "$network_containers" -gt 0 ]]; then
            print_status "info" "📦 $network_containers microservice(s) running"
        else
            print_status "success" "✓ No microservices currently running"
        fi
    fi
    
    echo ""
}

# =============================================================================
# USAGE AND HELP
# =============================================================================

usage() {
    cat << EOF
Container Runtime Switcher - Enhanced

COMMANDS:
  docker     Switch to Docker
  podman     Switch to Podman  
  status     Show current runtime status
  help       Show this help message

EXAMPLES:
  $0 docker     # Switch to Docker
  $0 podman     # Switch to Podman
  $0 status     # Show current status

FEATURES:
  - Uses service-manager for clean service shutdown
  - Automatic config.sh updates
  - Enhanced status reporting
  - Graceful error handling
  - Preserves data during switches

NOTES:
  - All running services will be stopped during the switch
  - The config.sh file will be updated automatically
  - Use 'service-manager.sh start-all' to restart services after switching
  - Data volumes are preserved during runtime switches

POST-SWITCH COMMANDS:
  ./scripts/service-manager.sh start-all    # Restart all services
  ./scripts/service-manager.sh status       # Check service status
  ./scripts/service-manager.sh ps           # List running containers

EOF
}

# =============================================================================
# MAIN COMMAND DISPATCH
# =============================================================================

main() {
    case "${1:-status}" in
        "docker")
            stop_all_services
            if switch_to_docker; then
                echo ""
                print_status "info" "To start services with Docker:"
                echo "  ./scripts/service-manager.sh start-all"
            fi
            ;;
        "podman")
            stop_all_services
            if switch_to_podman; then
                echo ""
                print_status "info" "To start services with Podman:"
                echo "  ./scripts/service-manager.sh start-all"
            fi
            ;;
        "status"|"s")
            show_status
            ;;
        "help"|"h"|"-h"|"--help")
            usage
            ;;
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