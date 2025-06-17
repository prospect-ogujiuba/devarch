#!/bin/bash

# =============================================================================
# PODMAN SOCKET SETUP FOR TRAEFIK
# =============================================================================
# Configure Podman to work with Traefik's Docker provider

set -e

print_status() {
    echo "🔧 $1"
}

print_success() {
    echo "✅ $1"
}

print_error() {
    echo "❌ $1"
}

# =============================================================================
# ENABLE PODMAN SOCKET
# =============================================================================

setup_podman_socket() {
    print_status "Setting up Podman socket for Traefik..."
    
    # Enable and start Podman socket (user mode)
    systemctl --user enable podman.socket
    systemctl --user start podman.socket
    
    # Verify socket is running
    if systemctl --user is-active podman.socket >/dev/null 2>&1; then
        print_success "Podman socket is active"
    else
        print_error "Failed to start Podman socket"
        exit 1
    fi
    
    # Show socket location
    SOCKET_PATH="/run/user/$(id -u)/podman/podman.sock"
    if [ -S "$SOCKET_PATH" ]; then
        print_success "Socket available at: $SOCKET_PATH"
    else
        print_error "Socket not found at expected location"
        exit 1
    fi
}

# =============================================================================
# ALTERNATIVE: SYSTEMD SOCKET (ROOT)
# =============================================================================

setup_system_socket() {
    print_status "Setting up system-wide Podman socket (requires sudo)..."
    
    # Enable and start system socket
    sudo systemctl enable podman.socket
    sudo systemctl start podman.socket
    
    # Verify socket is running
    if sudo systemctl is-active podman.socket >/dev/null 2>&1; then
        print_success "System Podman socket is active"
    else
        print_error "Failed to start system Podman socket"
        exit 1
    fi
    
    # Show socket location
    SOCKET_PATH="/run/podman/podman.sock"
    if sudo test -S "$SOCKET_PATH"; then
        print_success "Socket available at: $SOCKET_PATH"
        
        # Make socket accessible to your user
        sudo chmod 666 "$SOCKET_PATH"
        print_success "Socket permissions updated"
    else
        print_error "Socket not found at expected location"
        exit 1
    fi
}

# =============================================================================
# PODMAN API SERVICE (TCP)
# =============================================================================

setup_podman_api() {
    print_status "Setting up Podman API service on TCP..."
    
    # Start Podman API service
    podman system service --time=0 tcp://localhost:8888 &
    
    # Store PID for later cleanup
    echo $! > /tmp/podman-api.pid
    
    print_success "Podman API service started on tcp://localhost:8888"
    print_status "To stop: kill \$(cat /tmp/podman-api.pid)"
}

# =============================================================================
# TEST CONNECTIVITY
# =============================================================================

test_socket_connectivity() {
    local socket_path="$1"
    
    print_status "Testing socket connectivity..."
    
    # Test with curl
    if command -v curl >/dev/null 2>&1; then
        if curl -s --unix-socket "$socket_path" http://localhost/version >/dev/null 2>&1; then
            print_success "Socket responds to API calls"
            return 0
        fi
    fi
    
    # Test with podman directly
    if DOCKER_HOST="unix://$socket_path" podman version >/dev/null 2>&1; then
        print_success "Socket accessible via Podman"
        return 0
    fi
    
    print_error "Socket not responding to API calls"
    return 1
}

# =============================================================================
# CONFIGURE TRAEFIK LABELS FOR SERVICES
# =============================================================================

add_traefik_labels() {
    print_status "Example Traefik labels for your services..."
    
    cat << 'EOF'
# Add these labels to your service definitions:

services:
  your-service:
    # ... other config ...
    labels:
      # Enable Traefik
      - "traefik.enable=true"
      
      # HTTP Router
      - "traefik.http.routers.your-service.rule=Host(`your-service.test`)"
      - "traefik.http.routers.your-service.entrypoints=web"
      
      # HTTPS Router
      - "traefik.http.routers.your-service-secure.rule=Host(`your-service.test`)"
      - "traefik.http.routers.your-service-secure.entrypoints=web"
      - "traefik.http.routers.your-service-secure.tls=false"
      
      # Service definition
      - "traefik.http.services.your-service.loadbalancer.server.port=80"
      
      # Network
      - "traefik.docker.network=microservices-net"
      
      # Middleware (optional)
      - "traefik.http.routers.your-service-secure.middlewares=default-headers"

networks:
  microservices-net:
    external: true
EOF
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    echo "=== Podman Socket Setup for Traefik ==="
    echo ""
    
    echo "Choose setup method:"
    echo "1) User socket (recommended) - /run/user/\$(id -u)/podman/podman.sock"
    echo "2) System socket (requires sudo) - /run/podman/podman.sock"
    echo "3) TCP API service - tcp://localhost:8888"
    echo "4) Show example labels"
    echo ""
    
    read -p "Enter choice (1-4): " choice
    
    case $choice in
        1)
            setup_podman_socket
            SOCKET_PATH="/run/user/$(id -u)/podman/podman.sock"
            test_socket_connectivity "$SOCKET_PATH"
            echo ""
            echo "Update your Traefik compose file with:"
            echo "- /run/user/\$(id -u)/podman/podman.sock:/var/run/podman/podman.sock:ro"
            ;;
        2)
            setup_system_socket
            SOCKET_PATH="/run/podman/podman.sock"
            test_socket_connectivity "$SOCKET_PATH"
            echo ""
            echo "Update your Traefik compose file with:"
            echo "- /run/podman/podman.sock:/var/run/podman/podman.sock:ro"
            ;;
        3)
            setup_podman_api
            echo ""
            echo "Update your Traefik config with:"
            echo "endpoint: \"tcp://host.docker.internal:8888\""
            ;;
        4)
            add_traefik_labels
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi