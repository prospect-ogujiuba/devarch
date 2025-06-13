#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Default values
use_sudo=false
show_errors=false

# Parse command line arguments
while getopts "se" opt; do
    case $opt in
    s) use_sudo=true ;;
    e) show_errors=true ;;
    ?)
        echo "Usage: $0 [-s] [-e]
Options:
    -s    Use sudo for commands
    -e    Show error messages" >&2
        exit 1
        ;;
    esac
done

# Set up command prefix and error redirection
sudo_prefix=""
error_redirect="2>/dev/null"

[ "$use_sudo" = true ] && sudo_prefix="sudo "
[ "$show_errors" = true ] && error_redirect=""

# Create network if it doesn't exist
eval "${sudo_prefix}podman network exists \"$NETWORK_NAME\" ${error_redirect}" ||
    eval "${sudo_prefix}podman network create --driver bridge \"$NETWORK_NAME\" ${error_redirect}"

cd "$COMPOSE_DIR"

# Function to start a compose file with proper error handling
start_service() {
    local compose_file="$1"
    echo "Starting services from $compose_file..."
    eval "${sudo_prefix}podman compose -f \"$compose_file\" up -d ${error_redirect}" || true
}

# Start services in order
start_service database.docker-compose.yml
start_service db-tools.docker-compose.yml
start_service backend.docker-compose.yml
start_service analytics.docker-compose.yml
start_service mail.docker-compose.yml
start_service project.docker-compose.yml
start_service ai-services.docker-compose.yml
start_service erp.docker-compose.yml
start_service frontend.docker-compose.yml

echo "All services started successfully"