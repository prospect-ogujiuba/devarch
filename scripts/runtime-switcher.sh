#!/bin/bash

# =============================================================================
# CONTAINER RUNTIME SWITCHER
# =============================================================================
# Switches between Podman and Docker as drop-in replacements

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
CONFIG_FILE="$SCRIPT_DIR/config.sh"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_status() {
    local level="$1"
    local message="$2"
    
    case "$level" in
        "info")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
        "success")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "error")
            echo -e "${RED}❌ $message${NC}"
            ;;
    esac
}

get_current_runtime() {
    # Check Docker first (more reliable detection)
    if command -v docker >/dev/null 2>&1; then
        # Check if Docker daemon is running
        if sudo systemctl is-active docker >/dev/null 2>&1 || pgrep -f "dockerd" >/dev/null 2>&1; then
            # Try a simple docker command to confirm it's working
            if sudo docker info >/dev/null 2>&1; then
                echo "docker"
                return
            fi
        fi
    fi
    
    # Check Podman
    if command -v podman >/dev/null 2>&1; then
        # Check if podman socket is active or podman is responsive
        if systemctl --user is-active podman.socket >/dev/null 2>&1 || podman info >/dev/null 2>&1; then
            echo "podman"
            return
        fi
    fi
    
    echo "none"
}

stop_all_services() {
    print_status "info" "Stopping all microservices before runtime switch..."
    
    # Use the existing stop-services.sh script if available
    if [[ -f "$SCRIPT_DIR/stop-services.sh" ]]; then
        "$SCRIPT_DIR/stop-services.sh" --force --timeout 10 2>/dev/null || true
    else
        print_status "warning" "stop-services.sh not found, attempting manual cleanup..."
        cleanup_runtime_containers
    fi
}

cleanup_runtime_containers() {
    # Clean up both runtimes to ensure clean state
    cleanup_runtime "podman"
    cleanup_runtime "docker"
}

cleanup_runtime() {
    local runtime="$1"
    print_status "info" "Cleaning up $runtime resources..."
    
    if [[ "$runtime" == "podman" ]]; then
        # Stop all podman containers
        local containers=$(podman ps -q 2>/dev/null || true)
        if [[ -n "$containers" ]]; then
            podman stop $containers 2>/dev/null || true
            podman rm $containers 2>/dev/null || true
        fi
        
        # Remove networks
        podman network rm microservices-net 2>/dev/null || true
        
    elif [[ "$runtime" == "docker" ]]; then
        # Stop all docker containers  
        local containers=$(sudo docker ps -q 2>/dev/null || true)
        if [[ -n "$containers" ]]; then
            sudo docker stop $containers 2>/dev/null || true
            sudo docker rm $containers 2>/dev/null || true
        fi
        
        # Remove networks
        sudo docker network rm microservices-net 2>/dev/null || true
    fi
}

switch_to_docker() {
    print_status "info" "Switching to Docker..."
    
    # Stop podman services
    if command -v podman >/dev/null 2>&1; then
        systemctl --user stop podman.socket 2>/dev/null || true
        pkill -f podman 2>/dev/null || true
        sleep 2  # Give it time to stop
    fi
    
    # Start Docker
    if command -v docker >/dev/null 2>&1; then
        sudo systemctl start docker
        sleep 3  # Give Docker time to start
        
        # Verify Docker is working
        if sudo docker info >/dev/null 2>&1; then
            # Update config.sh
            if [[ -f "$CONFIG_FILE" ]]; then
                sed -i 's/export CONTAINER_RUNTIME="podman"/export CONTAINER_RUNTIME="docker"/' "$CONFIG_FILE"
                sed -i 's/export USE_PODMAN=true/export USE_PODMAN=false/' "$CONFIG_FILE"
            fi
            
            print_status "success" "Docker is now active"
            return 0
        else
            print_status "error" "Docker failed to start properly"
            return 1
        fi
    else
        print_status "error" "Docker not found. Please install Docker first."
        return 1
    fi
}

switch_to_podman() {
    print_status "info" "Switching to Podman..."
    
    # Stop Docker
    if command -v docker >/dev/null 2>&1; then
        sudo systemctl stop docker 2>/dev/null || true
        pkill -f dockerd 2>/dev/null || true
        sleep 2  # Give it time to stop
    fi
    
    # Start Podman
    if command -v podman >/dev/null 2>&1; then
        systemctl --user start podman.socket 2>/dev/null || true
        sleep 2  # Give Podman time to start
        
        # Verify Podman is working
        if podman info >/dev/null 2>&1; then
            # Update config.sh
            if [[ -f "$CONFIG_FILE" ]]; then
                sed -i 's/export CONTAINER_RUNTIME="docker"/export CONTAINER_RUNTIME="podman"/' "$CONFIG_FILE"
                sed -i 's/export USE_PODMAN=false/export USE_PODMAN=true/' "$CONFIG_FILE"
            fi
            
            print_status "success" "Podman is now active"
            return 0
        else
            print_status "error" "Podman failed to start properly"
            return 1
        fi
    else
        print_status "error" "Podman not found. Please install Podman first."
        return 1
    fi
}

show_status() {
    local current=$(get_current_runtime)
    
    echo -e "${BLUE}=== Container Runtime Status ===${NC}"
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
        echo -e "  ${GREEN}✓${NC} Podman installed"
    else
        echo -e "  ${RED}✗${NC} Podman not installed"
    fi
    
    if command -v docker >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Docker installed"
    else
        echo -e "  ${RED}✗${NC} Docker not installed"
    fi
    
    echo ""
    echo "Container Status:"
    
    # Show Podman containers
    if command -v podman >/dev/null 2>&1; then
        local podman_count=$(podman ps -q 2>/dev/null | wc -l || echo "0")
        if [[ "$podman_count" -gt 0 ]]; then
            echo -e "  ${YELLOW}⚠️${NC}  $podman_count containers running in Podman"
        else
            echo -e "  ${GREEN}✓${NC} No containers running in Podman"
        fi
    fi
    
    # Show Docker containers  
    if command -v docker >/dev/null 2>&1; then
        local docker_count=$(sudo docker ps -q 2>/dev/null | wc -l || echo "0")
        if [[ "$docker_count" -gt 0 ]]; then
            echo -e "  ${YELLOW}⚠️${NC}  $docker_count containers running in Docker"
        else
            echo -e "  ${GREEN}✓${NC} No containers running in Docker"
        fi
    fi
    
    echo ""
}

usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  docker     Switch to Docker"
    echo "  podman     Switch to Podman"
    echo "  status     Show current runtime status"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 docker     # Switch to Docker"
    echo "  $0 podman     # Switch to Podman"
    echo "  $0 status     # Show current status"
    echo ""
    echo "Notes:"
    echo "  - All running containers will be stopped during the switch"
    echo "  - The config.sh file will be updated automatically"
    echo "  - Restart your services after switching runtimes"
}

# Main execution
case "${1:-status}" in
    "docker")
        stop_all_services
        switch_to_docker
        echo ""
        print_status "info" "To start services with Docker, run: ./scripts/start-services.sh"
        ;;
    "podman")
        stop_all_services
        switch_to_podman
        echo ""
        print_status "info" "To start services with Podman, run: ./scripts/start-services.sh"
        ;;
    "status")
        show_status
        ;;
    "help"|"-h"|"--help")
        usage
        ;;
    *)
        print_status "error" "Unknown command: $1"
        usage
        exit 1
        ;;
esac