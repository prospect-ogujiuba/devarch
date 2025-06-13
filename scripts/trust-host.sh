#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Default values
use_sudo=false
show_errors=false
windows_cert=false

# Parse command line arguments
while getopts "sew" opt; do
    case $opt in
    s) use_sudo=true ;;
    e) show_errors=true ;;
    w) windows_cert=true ;;
    ?)
        echo "Usage: $0 [-s] [-e] [-w]
Options:
    -s    Use sudo for commands
    -e    Show error messages
    -w    Copy certificate to Windows" >&2
        exit 1
        ;;
    esac
done

# Set up command prefix and error redirection
sudo_prefix=""
error_redirect="2>/dev/null"

[ "$use_sudo" = true ] && sudo_prefix="sudo "
[ "$show_errors" = true ] && error_redirect=""

# Define paths
TEMP_DIR="/tmp/ssl-cert-$$"
LINUX_TRUST_DIR="/etc/ca-certificates/trust-source/anchors"

# Define Windows paths if needed
if [ "$windows_cert" = true ]; then
    WSL_WINDOWS_USER=$(cmd.exe /c "echo %USERNAME%" 2>/dev/null | tr -d '\r')
    WINDOWS_CERT_PATH="/mnt/c/Users/Prospect/Desktop/wildcard.test.crt"
fi

# Create temporary directory
mkdir -p "${TEMP_DIR}"

# Check if NPM container exists and copy certificates
if eval "${sudo_prefix}podman container exists nginx-proxy-manager ${error_redirect}"; then
    echo "Installing certificates for Linux/WSL2..."

    # First copy to temporary location
    echo "Copying certificate from container..."
    eval "${sudo_prefix}podman cp nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/fullchain.pem ${TEMP_DIR}/fullchain.pem ${error_redirect}"

    if [ -f "${TEMP_DIR}/fullchain.pem" ]; then
        # Copy to trust store with sudo
        echo "Installing certificate to system trust store..."
        eval "sudo cp ${TEMP_DIR}/fullchain.pem ${LINUX_TRUST_DIR}/wildcard.test.crt ${error_redirect}"
        eval "sudo chmod 644 ${LINUX_TRUST_DIR}/wildcard.test.crt ${error_redirect}"

        # Update trust store
        echo "Updating trust store..."
        eval "sudo trust extract-compat ${error_redirect}"
        eval "sudo update-ca-trust ${error_redirect}"
    else
        echo "Error: Failed to copy certificate from container"
        rm -rf "${TEMP_DIR}"
        exit 1
    fi

    if [ "$windows_cert" = true ]; then
        echo "Copying certificate for Windows..."
        # Copy certificate to Windows Desktop
        eval "cp ${TEMP_DIR}/fullchain.pem \"${WINDOWS_CERT_PATH}\" ${error_redirect}"
    fi

    # Cleanup
    rm -rf "${TEMP_DIR}"
else
    echo "Error: nginx-proxy-manager container not found. Please start it first."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Print success message and next steps
cat <<EOF
SSL certificates have been installed successfully!

Certificate Status:
- Linux/WSL2: Installed in ${LINUX_TRUST_DIR}/wildcard.test.crt
EOF

if [ "$windows_cert" = true ]; then
    cat <<EOF
- Windows: Copied to Desktop as wildcard.test.crt

For Windows (PowerShell Admin):
1. Open PowerShell as Administrator
2. Run the following command:
   Import-Certificate -FilePath "C:\\Users\\Prospect\\Desktop\\wildcard.test.crt" -CertStoreLocation Cert:\\LocalMachine\\Root

For Firefox (Additional Setup):
1. Open Firefox
2. Go to Settings → Privacy & Security → View Certificates
3. Click "Import" in Authorities tab
4. Browse to: C:\\Users\\Prospect\\Desktop\\wildcard.test.crt
5. Check "Trust this CA to identify websites"

Note: Chrome and Edge will use the Windows certificate store automatically.
EOF
fi

cat <<EOF

To verify the setup:
1. In WSL2/Linux: curl -v https://metabase.test
2. In a browser: visit https://metabase.test

Troubleshooting:
- Clear browser cache and restart if you still see certificate warnings
EOF

if [ "$windows_cert" = true ]; then
    cat <<EOF
- For Firefox, verify in about:config that security.enterprise_roots.enabled is true
- Check certificate presence in Windows Certificate Manager (certlm.msc)
EOF
fi

cat <<EOF
- Verify Linux trust with: trust list | grep "wildcard.test"
EOF
