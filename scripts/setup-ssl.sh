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

# Check if NPM container exists
if ! eval "${sudo_prefix}podman container exists nginx-proxy-manager ${error_redirect}"; then
    handle_error "nginx-proxy-manager container not found. Please start it first."
fi

echo "Creating certificate directory in container..."
eval "${sudo_prefix}podman exec nginx-proxy-manager mkdir -p /etc/letsencrypt/live/wildcard.test ${error_redirect}"
check_status "Failed to create certificate directory"

echo "Generating wildcard SSL certificate..."
# Generate a proper wildcard certificate that covers all *.test domains
eval "${sudo_prefix}podman exec nginx-proxy-manager openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
    -keyout /etc/letsencrypt/live/wildcard.test/privkey.pem \
    -out /etc/letsencrypt/live/wildcard.test/fullchain.pem \
    -subj \"/C=US/ST=Development/L=Local/O=Development/CN=*.test\" \
    -addext \"subjectAltName=DNS:*.test,DNS:test\" ${error_redirect}"
check_status "Failed to generate SSL certificate"

echo "Setting proper permissions on certificates..."
eval "${sudo_prefix}podman exec nginx-proxy-manager chmod 644 /etc/letsencrypt/live/wildcard.test/fullchain.pem ${error_redirect}"
eval "${sudo_prefix}podman exec nginx-proxy-manager chmod 600 /etc/letsencrypt/live/wildcard.test/privkey.pem ${error_redirect}"
check_status "Failed to set certificate permissions"

echo "Verifying certificate was created..."
if ! eval "${sudo_prefix}podman exec nginx-proxy-manager test -f /etc/letsencrypt/live/wildcard.test/fullchain.pem ${error_redirect}"; then
    handle_error "Certificate file was not created successfully"
fi

echo "Testing certificate validity..."
eval "${sudo_prefix}podman exec nginx-proxy-manager openssl x509 -in /etc/letsencrypt/live/wildcard.test/fullchain.pem -text -noout | grep -E '(CN=|DNS:)' ${error_redirect}"
check_status "Failed to verify certificate contents"

echo "Restarting NPM container to apply new certificate..."
eval "${sudo_prefix}podman restart nginx-proxy-manager ${error_redirect}"
check_status "Failed to restart nginx-proxy-manager"

echo "Waiting for NPM to start..."
sleep 10

echo "Testing if NPM is responding..."
until eval "${sudo_prefix}podman exec nginx-proxy-manager curl -k -s https://localhost:443 ${error_redirect}"; do
    echo "Waiting for NPM to be ready..."
    sleep 5
done

echo ""
echo "✅ SSL certificate generation completed successfully!"
echo ""
echo "Certificate details:"
eval "${sudo_prefix}podman exec nginx-proxy-manager openssl x509 -in /etc/letsencrypt/live/wildcard.test/fullchain.pem -text -noout | grep -E 'Subject:|DNS:' ${error_redirect}"
echo ""
echo "Next step: Run ./trust-host.sh to install the certificate in your system trust store"