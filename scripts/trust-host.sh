#!/bin/zsh

# =============================================================================
# SSL CERTIFICATE TRUST INSTALLATION SCRIPT
# =============================================================================
# Cross-platform SSL certificate trust installation for all operating systems
# Supports Linux, macOS, Windows, and WSL with automatic detection

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_copy_to_windows=false
opt_install_firefox=false
opt_install_chrome=false
opt_install_system=true
opt_backup_existing=true
opt_cert_name="wildcard.test"
opt_export_formats=false
opt_verify_installation=true

# Certificate paths
CERT_CONTAINER_PATH="$SSL_CERT_DIR/fullchain.pem"
TEMP_CERT_DIR="/tmp/ssl-trust-$$"

# OS-specific paths
declare -A TRUST_PATHS=(
    [linux_system]="/usr/local/share/ca-certificates"
    [linux_update]="update-ca-certificates"
    [macos_system]="/System/Library/Keychains/SystemRootCertificates.keychain"
    [macos_login]="/Library/Keychains/System.keychain"
    [windows_user]="Cert:\\CurrentUser\\Root"
    [windows_machine]="Cert:\\LocalMachine\\Root"
)

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Cross-platform SSL certificate trust installation script. Automatically
    detects your operating system and installs SSL certificates in the
    appropriate trust stores for browsers and system applications.

OPTIONS:
    -s, --sudo              Use sudo for system operations
    -e, --errors            Show detailed error messages
    -w, --windows           Copy certificate to Windows (WSL only)
    -f, --firefox           Install certificate for Firefox
    -c, --chrome            Install certificate for Chrome/Chromium
    --no-system             Skip system trust store installation
    --no-backup             Skip backing up existing certificates
    -n, --name NAME         Certificate name (default: wildcard.test)
    -x, --export            Export certificates in multiple formats
    --no-verify             Skip installation verification
    -h, --help              Show this help message

SUPPORTED PLATFORMS:
    Linux (Ubuntu/Debian) - system ca-certificates
    Linux (RHEL/CentOS)   - ca-trust anchors
    Linux (Arch)          - ca-certificates-utils
    macOS                 - Keychain Access
    Windows               - Certificate Store (via PowerShell)
    WSL                   - Linux + Windows integration

BROWSER SUPPORT:
    Chrome/Edge           - Uses system trust store automatically
    Firefox               - Requires separate installation (--firefox)
    Safari                - Uses system trust store (macOS)

EXAMPLES:
    $0                              # Auto-detect OS and install
    $0 -w                          # WSL: Install on Linux + copy to Windows
    $0 -f -c                       # Install for Firefox and Chrome
    $0 --no-system -f              # Only Firefox, skip system
    $0 -s -x                       # Use sudo, export multiple formats
    $0 --name custom.local         # Custom certificate name

INSTALLATION LOCATIONS:
    Linux:    /usr/local/share/ca-certificates/
    macOS:    System Keychain
    Windows:  LocalMachine\\Root (requires admin)
    Firefox:  Browser-specific certificate store

NOTES:
    - System installation usually requires administrator privileges
    - Firefox uses its own certificate store separate from system
    - Windows installation requires PowerShell with admin rights
    - Verification tests actual HTTPS connections to confirm installation
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--sudo)
                opt_use_sudo=true
                shift
                ;;
            -e|--errors)
                opt_show_errors=true
                shift
                ;;
            -w|--windows)
                opt_copy_to_windows=true
                shift
                ;;
            -f|--firefox)
                opt_install_firefox=true
                shift
                ;;
            -c|--chrome)
                opt_install_chrome=true
                shift
                ;;
            --no-system)
                opt_install_system=false
                shift
                ;;
            --no-backup)
                opt_backup_existing=false
                shift
                ;;
            -n|--name)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_cert_name="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a certificate name"
                fi
                ;;
            -x|--export)
                opt_export_formats=true
                shift
                ;;
            --no-verify)
                opt_verify_installation=false
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1. Use -h for help."
                ;;
        esac
    done
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

validate_environment() {
    print_status "step" "Validating environment and detecting OS..."
    
    # Check if nginx container exists and is running
    if ! eval "$CONTAINER_CMD container exists nginx-proxy-manager $ERROR_REDIRECT"; then
        handle_error "nginx-proxy-manager container not found. Run setup-ssl.sh first."
    fi
    
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' nginx-proxy-manager 2>/dev/null" || echo "unknown")
    
    if [[ "$container_status" != "running" ]]; then
        handle_error "nginx-proxy-manager container is not running"
    fi
    
    # Check if certificate exists in container
    if ! eval "$CONTAINER_CMD exec nginx-proxy-manager test -f $CERT_CONTAINER_PATH $ERROR_REDIRECT"; then
        handle_error "SSL certificate not found in container. Run setup-ssl.sh first."
    fi
    
    print_status "success" "Environment validation passed"
    print_status "info" "Detected OS: $OS_TYPE"
}

detect_linux_distribution() {
    if [[ "$OS_TYPE" != "linux" && "$OS_TYPE" != "wsl" ]]; then
        return 0
    fi
    
    if command -v lsb_release >/dev/null 2>&1; then
        LINUX_DISTRO=$(lsb_release -si 2>/dev/null | tr '[:upper:]' '[:lower:]')
    elif [[ -f /etc/os-release ]]; then
        LINUX_DISTRO=$(grep '^ID=' /etc/os-release | cut -d'=' -f2 | tr -d '"' | tr '[:upper:]' '[:lower:]')
    elif [[ -f /etc/redhat-release ]]; then
        LINUX_DISTRO="rhel"
    else
        LINUX_DISTRO="unknown"
    fi
    
    print_status "info" "Linux distribution: $LINUX_DISTRO"
}

check_prerequisites() {
    case "$OS_TYPE" in
        "linux"|"wsl")
            detect_linux_distribution
            
            case "$LINUX_DISTRO" in
                "ubuntu"|"debian")
                    if ! command -v update-ca-certificates >/dev/null 2>&1; then
                        handle_error "ca-certificates package not installed. Run: apt install ca-certificates"
                    fi
                    ;;
                "rhel"|"centos"|"fedora")
                    if ! command -v update-ca-trust >/dev/null 2>&1; then
                        handle_error "ca-certificates package not installed. Run: yum install ca-certificates"
                    fi
                    ;;
                "arch")
                    if ! command -v trust >/dev/null 2>&1; then
                        handle_error "ca-certificates-utils package not installed. Run: pacman -S ca-certificates-utils"
                    fi
                    ;;
            esac
            ;;
        "macos")
            if ! command -v security >/dev/null 2>&1; then
                handle_error "macOS security command not available"
            fi
            ;;
        "windows")
            # Windows-specific checks would go here
            ;;
    esac
}

# =============================================================================
# CERTIFICATE PREPARATION FUNCTIONS
# =============================================================================

prepare_certificate_files() {
    print_status "step" "Preparing certificate files..."
    
    # Create temporary directory
    mkdir -p "$TEMP_CERT_DIR"
    
    # Copy certificate from container
    eval "$CONTAINER_CMD cp nginx-proxy-manager:$CERT_CONTAINER_PATH $TEMP_CERT_DIR/cert.pem $ERROR_REDIRECT"
    check_status "Failed to copy certificate from container"
    
    # Verify certificate
    if ! openssl x509 -in "$TEMP_CERT_DIR/cert.pem" -noout -text >/dev/null 2>&1; then
        handle_error "Invalid certificate file"
    fi
    
    print_status "success" "Certificate files prepared"
}

backup_existing_certificates() {
    if [[ "$opt_backup_existing" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Backing up existing certificates..."
    
    local backup_dir="$PROJECT_ROOT/ssl-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"
    
    case "$OS_TYPE" in
        "linux"|"wsl")
            case "$LINUX_DISTRO" in
                "ubuntu"|"debian")
                    if [[ -f "/usr/local/share/ca-certificates/$opt_cert_name.crt" ]]; then
                        cp "/usr/local/share/ca-certificates/$opt_cert_name.crt" "$backup_dir/"
                    fi
                    ;;
                "rhel"|"centos"|"fedora")
                    if [[ -f "/etc/pki/ca-trust/source/anchors/$opt_cert_name.crt" ]]; then
                        cp "/etc/pki/ca-trust/source/anchors/$opt_cert_name.crt" "$backup_dir/"
                    fi
                    ;;
            esac
            ;;
    esac
    
    print_status "success" "Existing certificates backed up to: $backup_dir"
}

# =============================================================================
# PLATFORM-SPECIFIC INSTALLATION FUNCTIONS
# =============================================================================

install_linux_system_trust() {
    if [[ "$opt_install_system" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Installing certificate to Linux system trust store..."
    
    case "$LINUX_DISTRO" in
        "ubuntu"|"debian")
            # Copy certificate to ca-certificates directory
            local cert_path="/usr/local/share/ca-certificates/$opt_cert_name.crt"
            
            if [[ "$opt_use_sudo" == "true" ]]; then
                sudo cp "$TEMP_CERT_DIR/cert.pem" "$cert_path"
                sudo chmod 644 "$cert_path"
                sudo update-ca-certificates
            else
                cp "$TEMP_CERT_DIR/cert.pem" "$cert_path" 2>/dev/null || {
                    print_status "error" "Permission denied. Try with --sudo option"
                    return 1
                }
                update-ca-certificates
            fi
            ;;
            
        "rhel"|"centos"|"fedora")
            # Copy certificate to ca-trust anchors
            local cert_path="/etc/pki/ca-trust/source/anchors/$opt_cert_name.crt"
            
            if [[ "$opt_use_sudo" == "true" ]]; then
                sudo cp "$TEMP_CERT_DIR/cert.pem" "$cert_path"
                sudo chmod 644 "$cert_path"
                sudo update-ca-trust
            else
                cp "$TEMP_CERT_DIR/cert.pem" "$cert_path" 2>/dev/null || {
                    print_status "error" "Permission denied. Try with --sudo option"
                    return 1
                }
                update-ca-trust
            fi
            ;;
            
        "arch")
            # Copy certificate and update trust
            local cert_path="/etc/ca-certificates/trust-source/anchors/$opt_cert_name.crt"
            
            if [[ "$opt_use_sudo" == "true" ]]; then
                sudo cp "$TEMP_CERT_DIR/cert.pem" "$cert_path"
                sudo chmod 644 "$cert_path"
                sudo trust extract-compat
            else
                cp "$TEMP_CERT_DIR/cert.pem" "$cert_path" 2>/dev/null || {
                    print_status "error" "Permission denied. Try with --sudo option"
                    return 1
                }
                trust extract-compat
            fi
            ;;
            
        *)
            print_status "warning" "Unsupported Linux distribution: $LINUX_DISTRO"
            return 1
            ;;
    esac
    
    print_status "success" "Certificate installed to Linux system trust store"
}

install_macos_system_trust() {
    if [[ "$opt_install_system" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Installing certificate to macOS system trust store..."
    
    # Import certificate to system keychain
    if security add-trusted-cert -d -r trustRoot -k "${TRUST_PATHS[macos_login]}" "$TEMP_CERT_DIR/cert.pem" 2>/dev/null; then
        print_status "success" "Certificate installed to macOS system trust store"
    else
        # Try with sudo
        if [[ "$opt_use_sudo" == "true" ]]; then
            if sudo security add-trusted-cert -d -r trustRoot -k "${TRUST_PATHS[macos_login]}" "$TEMP_CERT_DIR/cert.pem"; then
                print_status "success" "Certificate installed to macOS system trust store (with sudo)"
            else
                print_status "error" "Failed to install certificate to macOS trust store"
                return 1
            fi
        else
            print_status "error" "Permission denied. Try with --sudo option"
            return 1
        fi
    fi
}

install_windows_system_trust() {
    if [[ "$opt_copy_to_windows" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Installing certificate to Windows system trust store..."
    
    # Copy certificate to Windows accessible location
    local windows_cert_path="/mnt/c/temp/$opt_cert_name.crt"
    mkdir -p "/mnt/c/temp"
    cp "$TEMP_CERT_DIR/cert.pem" "$windows_cert_path"
    
    # Generate PowerShell script for certificate installation
    cat > "/mnt/c/temp/install-cert.ps1" << 'EOF'
param(
    [string]$CertPath = "C:\temp\wildcard.test.crt"
)

try {
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($CertPath)
    $store = New-Object System.Security.Cryptography.X509Certificates.X509Store([System.Security.Cryptography.X509Certificates.StoreName]::Root, [System.Security.Cryptography.X509Certificates.StoreLocation]::LocalMachine)
    $store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite)
    $store.Add($cert)
    $store.Close()
    Write-Host "Certificate installed successfully"
    exit 0
} catch {
    Write-Error "Failed to install certificate: $_"
    exit 1
}
EOF
    
    print_status "success" "Certificate copied to Windows. Run as Administrator in PowerShell:"
    echo "  PowerShell -ExecutionPolicy Bypass -File C:\\temp\\install-cert.ps1"
    echo ""
    print_status "info" "Alternative: Double-click C:\\temp\\$opt_cert_name.crt and install to 'Trusted Root Certification Authorities'"
}

install_firefox_trust() {
    if [[ "$opt_install_firefox" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Installing certificate for Firefox..."
    
    # Find Firefox profiles
    local firefox_dir=""
    case "$OS_TYPE" in
        "linux"|"wsl")
            firefox_dir="$HOME/.mozilla/firefox"
            ;;
        "macos")
            firefox_dir="$HOME/Library/Application Support/Firefox/Profiles"
            ;;
        "windows")
            firefox_dir="/mnt/c/Users/$USER/AppData/Roaming/Mozilla/Firefox/Profiles"
            ;;
    esac
    
    if [[ ! -d "$firefox_dir" ]]; then
        print_status "warning" "Firefox profile directory not found: $firefox_dir"
        return 1
    fi
    
    # Find all Firefox profiles
    local -a profiles
    profiles=($(find "$firefox_dir" -name "*.default*" -type d 2>/dev/null))
    
    if [[ ${#profiles[@]} -eq 0 ]]; then
        print_status "warning" "No Firefox profiles found"
        return 1
    fi
    
    print_status "info" "Found ${#profiles[@]} Firefox profile(s)"
    
    # Install certificate to each profile (requires certutil)
    if command -v certutil >/dev/null 2>&1; then
        for profile in "${profiles[@]}"; do
            if [[ -f "$profile/cert9.db" ]]; then
                certutil -A -n "$opt_cert_name" -t "TCu,Cu,Tu" -i "$TEMP_CERT_DIR/cert.pem" -d sql:"$profile" 2>/dev/null && \
                print_status "success" "Certificate installed to Firefox profile: $(basename "$profile")"
            fi
        done
    else
        print_status "warning" "certutil not found. Install nss-tools package for automatic Firefox installation"
        print_status "info" "Manual Firefox installation:"
        echo "  1. Open Firefox"
        echo "  2. Go to Settings → Privacy & Security → View Certificates"
        echo "  3. Click 'Import' in Authorities tab"
        echo "  4. Select: $TEMP_CERT_DIR/cert.pem"
        echo "  5. Check 'Trust this CA to identify websites'"
    fi
}

install_chrome_trust() {
    if [[ "$opt_install_chrome" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Configuring Chrome/Chromium trust..."
    
    case "$OS_TYPE" in
        "linux"|"wsl")
            # Chrome on Linux uses system certificate store
            print_status "info" "Chrome on Linux uses system certificate store"
            print_status "info" "Certificate should work automatically after system installation"
            ;;
        "macos")
            # Chrome on macOS uses system keychain
            print_status "info" "Chrome on macOS uses system keychain"
            print_status "info" "Certificate should work automatically after system installation"
            ;;
        "windows")
            # Chrome on Windows uses system certificate store
            print_status "info" "Chrome on Windows uses system certificate store"
            print_status "info" "Certificate should work automatically after Windows installation"
            ;;
    esac
    
    print_status "success" "Chrome trust configuration completed"
}

# =============================================================================
# VERIFICATION FUNCTIONS
# =============================================================================

verify_installation() {
    if [[ "$opt_verify_installation" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Verifying certificate installation..."
    
    # Test if certificate is trusted by checking common services
    local -a test_urls=(
        "https://nginx.test"
        "https://metabase.test"
        "https://grafana.test"
    )
    
    local successful_tests=0
    for url in "${test_urls[@]}"; do
        if curl -s --max-time 5 --connect-timeout 3 "$url" >/dev/null 2>&1; then
            ((successful_tests++))
            if [[ "$opt_verbose" == "true" ]]; then
                print_status "success" "✓ $url"
            fi
        else
            if [[ "$opt_verbose" == "true" ]]; then
                print_status "warning" "✗ $url (service may not be running)"
            fi
        fi
    done
    
    if [[ $successful_tests -gt 0 ]]; then
        print_status "success" "Certificate verification passed ($successful_tests/$((${#test_urls[@]})) services accessible)"
    else
        print_status "warning" "Certificate verification inconclusive (services may not be running)"
    fi
    
    # Test with openssl
    if echo | openssl s_client -connect localhost:443 -servername nginx.test 2>/dev/null | grep -q "Verify return code: 0"; then
        print_status "success" "OpenSSL verification passed"
    else
        print_status "info" "OpenSSL verification may require service restart"
    fi
}

export_certificate_formats() {
    if [[ "$opt_export_formats" == "false" ]]; then
        return 0
    fi
    
    print_status "step" "Exporting certificate in multiple formats..."
    
    local export_dir="$PROJECT_ROOT/ssl-export"
    mkdir -p "$export_dir"
    
    # Copy original PEM
    cp "$TEMP_CERT_DIR/cert.pem" "$export_dir/$opt_cert_name.crt"
    
    # Create DER format
    openssl x509 -outform der -in "$TEMP_CERT_DIR/cert.pem" -out "$export_dir/$opt_cert_name.der" 2>/dev/null
    
    # Create installation instructions
    cat > "$export_dir/installation-instructions.txt" << EOF
SSL Certificate Installation Instructions
========================================

Certificate: $opt_cert_name
Generated: $(date)

Manual Installation:

LINUX (Ubuntu/Debian):
sudo cp $opt_cert_name.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

LINUX (RHEL/CentOS):
sudo cp $opt_cert_name.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust

macOS:
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $opt_cert_name.crt

Windows (PowerShell as Admin):
Import-Certificate -FilePath "$opt_cert_name.crt" -CertStoreLocation Cert:\\LocalMachine\\Root

Firefox:
1. Settings → Privacy & Security → View Certificates
2. Authorities tab → Import
3. Select $opt_cert_name.crt
4. Check "Trust this CA to identify websites"
EOF
    
    print_status "success" "Certificates exported to: $export_dir"
}

# =============================================================================
# CLEANUP FUNCTIONS
# =============================================================================

cleanup_temp_files() {
    if [[ -d "$TEMP_CERT_DIR" ]]; then
        rm -rf "$TEMP_CERT_DIR"
    fi
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

show_trust_summary() {
    print_status "info" "SSL Certificate Trust Installation Summary:"
    echo "  Operating System: $OS_TYPE"
    echo "  Certificate Name: $opt_cert_name"
    echo "  Use Sudo: $opt_use_sudo"
    echo "  Install System Trust: $opt_install_system"
    echo "  Install Firefox: $opt_install_firefox"
    echo "  Install Chrome: $opt_install_chrome"
    echo "  Copy to Windows: $opt_copy_to_windows"
    echo "  Export Formats: $opt_export_formats"
    echo "  Verify Installation: $opt_verify_installation"
    echo ""
}

main() {
    # Set up signal handlers for cleanup
    trap cleanup_temp_files EXIT INT TERM
    
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Show trust summary
    show_trust_summary
    
    # Validate environment
    validate_environment
    check_prerequisites
    
    # Prepare certificate files
    prepare_certificate_files
    backup_existing_certificates
    
    # Install certificates based on OS
    case "$OS_TYPE" in
        "linux")
            install_linux_system_trust
            ;;
        "wsl")
            install_linux_system_trust
            install_windows_system_trust
            ;;
        "macos")
            install_macos_system_trust
            ;;
        "windows")
            install_windows_system_trust
            ;;
        *)
            print_status "error" "Unsupported operating system: $OS_TYPE"
            exit 1
            ;;
    esac
    
    # Install browser-specific certificates
    install_firefox_trust
    install_chrome_trust
    
    # Export certificates if requested
    export_certificate_formats
    
    # Verify installation
    verify_installation
    
    print_status "success" "SSL certificate trust installation completed!"
    
    # Show completion info
    echo ""
    print_status "info" "Next Steps:"
    echo "  1. Restart your browser to use the new certificate"
    echo "  2. Test HTTPS access: curl -v https://nginx.test"
    echo "  3. Access services via HTTPS URLs"
    
    if [[ "$OS_TYPE" == "wsl" && "$opt_copy_to_windows" == "true" ]]; then
        echo "  4. Run the PowerShell script in Windows as Administrator"
    fi
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi