#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Default values
use_sudo=false
show_errors=false
remove_images=false
remove_volumes=false

# Parse command line arguments
while getopts "seiv" opt; do
    case $opt in
    s) use_sudo=true ;;
    e) show_errors=true ;;
    i) remove_images=true ;;
    v) remove_volumes=true ;;
    ?)
        echo "Usage: $0 [-s] [-e] [-i] [-v]
Options:
    -s    Use sudo for commands
    -e    Show error messages
    -i    Remove all images
    -v    Remove all volumes" >&2
        exit 1
        ;;
    esac
done

# Set up command prefix and error redirection
sudo_prefix=""
error_redirect="2>/dev/null"

[ "$use_sudo" = true ] && sudo_prefix="sudo "
[ "$show_errors" = true ] && error_redirect=""

cd "$COMPOSE_DIR"

# Function to stop a compose file with proper error handling
stop_service() {
    local compose_file="$1"
    echo "Stopping services from $compose_file..."
    eval "${sudo_prefix}podman compose -f \"$compose_file\" down ${error_redirect}" || true
}

# Stop services in reverse order
stop_service frontend.docker-compose.yml
stop_service erp.docker-compose.yml
stop_service ai-services.docker-compose.yml
stop_service project.docker-compose.yml
stop_service mail.docker-compose.yml
stop_service analytics.docker-compose.yml
stop_service backend.docker-compose.yml
stop_service db-tools.docker-compose.yml
stop_service database.docker-compose.yml

echo "All services stopped"

# Remove images if requested
if [ "$remove_images" = true ]; then
    echo "Removing all images..."
    eval "${sudo_prefix}podman rmi -f \$(${sudo_prefix}podman images -q) ${error_redirect}" || true
    echo "All images removed"
fi

# Remove volumes if requested
if [ "$remove_volumes" = true ]; then
    echo "Removing all volumes..."
    eval "${sudo_prefix}podman volume rm -f \$(${sudo_prefix}podman volume ls -q) ${error_redirect}" || true
    echo "All volumes removed"
fi

echo "All Services Stopped Successfully"