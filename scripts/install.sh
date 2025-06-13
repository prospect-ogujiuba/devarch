#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Default values
use_sudo=false
show_errors=false
skip_confirmation=false
skip_db_setup=false

# Parse command line arguments
while getopts "seydt" opt; do
    case $opt in
    s) use_sudo=true ;;
    e) show_errors=true ;;
    y) skip_confirmation=true ;;
    d) skip_db_setup=true ;;
    ?)
        echo "Usage: $0 [-s] [-e] [-y] [-d]
Options:
    -s    Use sudo for commands
    -e    Show error messages
    -d    Skip database setup
    -y    Skip confirmation prompts" >&2
        exit 1
        ;;
    esac
done

# Set up command prefix and error redirection
sudo_prefix=""
error_redirect="2>/dev/null"

[ "$use_sudo" = true ] && sudo_prefix="sudo "
[ "$show_errors" = true ] && error_redirect=""

# Function to handle errors
handle_error() {
    echo "Error: $1"
    exit 1
}

# Function to check command status
check_status() {
    if [ $? -ne 0 ]; then
        handle_error "$1"
    fi
}

# Confirmation prompt
if [ "$skip_confirmation" = false ]; then
    echo "This script will install and set up the entire environment."
    echo "Do you want to continue? (y/n)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Installation aborted."
        exit 0
    fi
fi

# Create network if it doesn't exist
eval "${sudo_prefix}podman network exists \"$NETWORK_NAME\" ${error_redirect}" ||
eval "${sudo_prefix}podman network create --driver bridge \"$NETWORK_NAME\" ${error_redirect}"

# Change to compose directory
echo "Changing to compose directory..."
cd "$COMPOSE_DIR" || handle_error "Failed to change to compose directory"

# Start database services
echo "Starting database services..."
eval "${sudo_prefix}podman compose -f database.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start database services"
echo "Database services started successfully"

# Allow for database initialization
echo "Waiting for databases to initialize..."
sleep 15

# Run database setup if not skipped
if [ "$skip_db_setup" = false ]; then
    echo "Setting up databases..."
    eval "${sudo_prefix}$SCRIPT_DIR/setup-databases.sh -m -p ${error_redirect}"
    check_status "Failed to setup databases"
    echo "Databases setup completed"
fi

# Start database tools
echo "Starting database tools..."
eval "${sudo_prefix}podman compose -f db-tools.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start database tools"
echo "Database tools started successfully"

# Start backend services
echo "Starting backend services..."
eval "${sudo_prefix}podman compose -f backend.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start backend services"
echo "Backend services started successfully"

# Start analytics services
echo "Starting analytics services..."
eval "${sudo_prefix}podman compose -f analytics.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start analytics services"
echo "Analytics services started successfully"

# Start mail services
echo "Starting mail services..."
eval "${sudo_prefix}podman compose -f mail.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start mail services"
echo "Mail services started successfully"

# Start project services
echo "Starting project services..."
eval "${sudo_prefix}podman compose -f project.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start project services"
echo "Project services started successfully"

# Start AI services
echo "Starting AI services..."
eval "${sudo_prefix}podman compose -f ai-services.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start AI services"
echo "AI services started successfully"

# Start ERP services
echo "Starting ERP services..."
eval "${sudo_prefix}podman compose -f erp.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start ERP services"
echo "ERP services started successfully"

# Start frontend services
echo "Starting frontend services..."
eval "${sudo_prefix}podman compose -f frontend.docker-compose.yml up -d ${error_redirect}"
check_status "Failed to start frontend services"
echo "Frontend services started successfully"

# Allow for services to start
echo "Waiting for all services to start..."
sleep 15

# Setup SSL certificates
echo "Setting up SSL certificates..."
eval "${sudo_prefix}$SCRIPT_DIR/setup-ssl.sh ${error_redirect}"
check_status "Failed to setup SSL certificates"
echo "SSL certificates setup completed"

# Trust SSL certificates
echo "Installing SSL certificates for host system..."
eval "${sudo_prefix}$SCRIPT_DIR/trust-host.sh ${error_redirect}"
check_status "Failed to trust SSL certificates"
echo "SSL certificates trusted by host system"

echo "========================================================"
echo "Installation completed successfully!"
echo "You can now access your services at their respective URLs:"
echo "- Nginx Proxy Manager: https://nginx.test"
echo "- Odoo ERP: https://odoo.test"
echo "- OpenProject: https://openproject.test"
echo "- Metabase: https://metabase.test"
echo "- Grafana: https://grafana.test"
echo "- Matomo: https://matomo.test"
echo "- Adminer: https://adminer.test"
echo "- phpMyAdmin: https://phpmyadmin.test"
echo "- Mongo Express: https://mongodb.test"
echo "- NocoDB: https://nocodb.test"
echo "- Mailpit: https://mailpit.test"
echo "- WordPress: https://wploaded.test, https://wpclean.test, https://wpcustom.test, https://wpbare.test"
echo "- AI services: https://n8n.test, https://langflow.test, https://kibana.test, https://portainer.test"
echo "========================================================"