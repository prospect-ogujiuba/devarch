#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Default values
use_sudo=false
show_errors=false
setup_mariadb_flag=true
setup_postgres_flag=true

# Parse command line arguments
while getopts "semp" opt; do
    case $opt in
    s) use_sudo=true ;;
    e) show_errors=true ;;
    m) setup_mariadb_flag=true ;;
    p) setup_postgres_flag=true ;;
    ?)
        echo "Usage: $0 [-s] [-e] [-m] [-p]
Options:
    -s    Use sudo for commands
    -e    Show error messages
    -m    Setup MariaDB
    -p    Setup PostgreSQL" >&2
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

# Function to setup MariaDB
setup_mariadb() {
    echo "Setting up MariaDB..."
    eval "${sudo_prefix}podman exec -i mariadb bash << EOF ${error_redirect}
mariadb -u root -p\"\${MARIADB_ROOT_PASSWORD}\" << 'EOSQL'
CREATE DATABASE IF NOT EXISTS npm;
EOSQL
EOF"
    check_status "Failed to create MariaDB database"
}

# Function to setup PostgreSQL
setup_postgres() {
    echo "Setting up PostgreSQL..."
    eval "${sudo_prefix}podman exec -i postgres bash << EOF ${error_redirect}
PGPASSWORD=\"\${POSTGRES_PASSWORD}\" psql -U postgres << 'EOSQL'
CREATE USER odoo_user WITH PASSWORD '123456';
ALTER USER odoo_user WITH CREATEDB;
CREATE DATABASE metabase;
CREATE DATABASE keycloak;
EOSQL
EOF"
    check_status "Failed to create PostgreSQL schemas"
}

# Main execution
echo "Starting setup..."
[ "$setup_mariadb_flag" = true ] && setup_mariadb
[ "$setup_postgres_flag" = true ] && setup_postgres

echo "Setup completed successfully!"