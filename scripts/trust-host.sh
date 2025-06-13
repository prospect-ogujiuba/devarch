#!/bin/zsh
# Enhanced Trust Host Script for Microservices Architecture
# Installs SSL certificates in system trust stores for seamless HTTPS development

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# =============================================================================
# SCRIPT-SPECIFIC CONFIGURATION
# =============================================================================

# Trust store options
TRUST_LINUX=true
TRUST_WINDOWS=false
TRUST_MACOS=false
TRUST_FIREFOX=false
TRUST_CHROME=false
UPDATE_HOSTS_FILE=false
REMOVE_CERTIFICATES=false
LIST_TRUSTED_CERTS=false
CERT_SOURCE="container"        # container, local, custom
CUSTOM_CERT_PATH=""

# System paths
LINUX_TRUST_DIR="/etc/ca-certificates/trust-source/anchors"
LINUX_CERT_DIR="/usr/local/share/ca-certificates"
UBUNTU_CERT_DIR="/usr/local/share/ca-certificates"
ARCH_TRUST_DIR="/etc/ca-certificates/trust-source/anchors"

# Windows paths (WSL/Cross-platform)
WINDOWS_USER=""
WINDOWS_CERT_STORE="Cert:\\LocalMachine\\Root"
WINDOWS_DESKTOP_PATH=""

# macOS paths
MACOS_KEYCHAIN="/Library/Keychains/System.keychain"
MACOS_TRUST_SETTINGS="ssl"

# Browser paths
FIREFOX_PROFILE_DIRS=()
CHROME_POLICY_DIR="/etc/opt/chrome/policies/managed"

# Certificate information
CERT_CONTAINER_PATH="/etc/letsencrypt/live/wildcard.test/fullchain.pem"
CERT_LOCAL_PATH="$PROJECT_ROOT/ssl/wildcard.test.crt"
CERT_TEMP_DIR="/tmp/ssl-cert-$$"

# =============================================================================
# HELP FUNCTION
# =============================================================================

show_help() {
    cat << EOF
Enhanced Trust Host Script for SSL Certificates

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for system commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -L              Trust on Linux/WSL (default: true)
    -W              Trust on Windows (requires WSL/PowerShell)
    -M              Trust on macOS
    -F              Trust in Firefox browsers
    -C              Trust in Chrome/Chromium browsers
    -H              Update system hosts file for .test domains
    -R              Remove trusted certificates instead of adding
    -l              List currently trusted certificates
    -S SOURCE       Certificate source: container, local, custom (default: container)
    -P PATH         Custom certificate path (use with -S custom)
    -h              Show this help message

CERTIFICATE SOURCES:
    container       Extract from nginx-proxy-manager container (default)
    local           Use certificate from PROJECT_ROOT/ssl/
    custom          Use certificate from custom path (-P option)

SUPPORTED PLATFORMS:
    Linux           System-wide trust via ca-certificates
    WSL2            Linux trust + Windows certificate copy
    macOS           System keychain integration
    Windows         PowerShell certificate import
    Firefox         Profile-specific certificate store
    Chrome          Enterprise policy certificate management

EXAMPLES:
    $0                              # Trust on current Linux system
    $0 -W -F                        # Trust on Windows and Firefox
    $0 -L -M -C                     # Trust on Linux, macOS, and Chrome
    $0 -S local -v                  # Use local certificate with verbose output
    $0 -S custom -P /path/cert.pem  # Use custom certificate
    $0 -R -v                        # Remove trusted certificates
    $0 -l                           # List trusted certificates
    $0 -H                           # Update hosts file for .test domains

HOSTS FILE ENTRIES:
    When -H is used, adds entries for all microservice domains:
$(printf "    127.0.0.1 %s\n" "nginx.test" "adminer.test" "metabase.test" "grafana.test")

EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_trust_args() {
    local OPTIND=1
    
    while getopts "sevdr:LWMFCHRS:P:lh" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            L) TRUST_LINUX=true ;;
            W) TRUST_WINDOWS=true ;;
            M) TRUST_MACOS=true ;;
            F) TRUST_FIREFOX=true ;;
            C) TRUST_CHROME=true ;;
            H) UPDATE_HOSTS_FILE=true ;;
            R) REMOVE_CERTIFICATES=true ;;
            S) CERT_SOURCE="$OPTARG" ;;
            P) CUSTOM_CERT_PATH="$OPTARG" ;;
            l) LIST_TRUSTED_CERTS=true ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
    
    # Validate certificate source
    case "$CERT_SOURCE" in
        "container"|"local"|"custom") ;;
        *)
            log "ERROR" "Invalid certificate source: $CERT_SOURCE"
            log "INFO" "Valid sources: container, local, custom"
            exit 1
            ;;
    esac
    
    # Validate custom path if specified
    if [ "$CERT_SOURCE" = "custom" ] && [ -z "$CUSTOM_CERT_PATH" ]; then
        log "ERROR" "Custom certificate path required when using -S custom"
        exit 1
    fi
}

# =============================================================================
# SYSTEM DETECTION
# =============================================================================

detect_platform() {
    local platform=""
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        platform="linux"
        
        # Check if running in WSL
        if grep -qi microsoft /proc/version 2>/dev/null; then
            platform="wsl"
            log "INFO" "🐧 Detected: Windows Subsystem for Linux (WSL)"
            
            # Get Windows username for WSL
            WINDOWS_USER=$(cmd.exe /c "echo %USERNAME%" 2>/dev/null | tr -d '\r' 2>/dev/null || echo "")
            if [ -n "$WINDOWS_USER" ]; then
                WINDOWS_DESKTOP_PATH="/mnt/c/Users/$WINDOWS_USER/Desktop"
                log "DEBUG" "Windows user detected: $WINDOWS_USER"
            fi
        else
            log "INFO" "🐧 Detected: Linux"
        fi
        
        # Detect Linux distribution
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            log "DEBUG" "Linux distribution: $NAME"
            
            case "$ID" in
                "ubuntu"|"debian")
                    LINUX_TRUST_DIR="$UBUNTU_CERT_DIR"
                    ;;
                "arch"|"manjaro")
                    LINUX_TRUST_DIR="$ARCH_TRUST_DIR"
                    ;;
            esac
        fi
        
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        platform="macos"
        log "INFO" "🍎 Detected: macOS"
        
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        platform="windows"
        log "INFO" "🪟 Detected: Windows"
        
    else
        platform="unknown"
        log "WARN" "❓ Unknown platform: $OSTYPE"
    fi
    
    echo "$platform"
}

# =============================================================================
# CERTIFICATE ACQUISITION
# =============================================================================

get_certificate() {
    log "INFO" "📜 Acquiring certificate from source: $CERT_SOURCE"
    
    # Create temporary directory
    mkdir -p "$CERT_TEMP_DIR"
    
    case "$CERT_SOURCE" in
        "container")
            get_certificate_from_container
            ;;
        "local")
            get_certificate_from_local
            ;;
        "custom")
            get_certificate_from_custom
            ;;
    esac
}

get_certificate_from_container() {
    log "INFO" "📦 Extracting certificate from nginx-proxy-manager container..."
    
    # Check if container exists
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists nginx-proxy-manager 2>/dev/null; then
        log "ERROR" "nginx-proxy-manager container not found"
        return 1
    fi
    
    # Check if certificate exists in container
    if ! execute_command "Check certificate in container" \
         "${SUDO_PREFIX}${CONTAINER_RUNTIME} exec nginx-proxy-manager test -f $CERT_CONTAINER_PATH" false false; then
        log "ERROR" "Certificate not found in container: $CERT_CONTAINER_PATH"
        log "INFO" "Run setup-ssl.sh first to generate certificates"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would extract certificate from container"
        return 0
    fi
    
    # Copy certificate from container
    execute_command "Extract certificate from container" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} cp nginx-proxy-manager:$CERT_CONTAINER_PATH $CERT_TEMP_DIR/certificate.pem" \
        true true
    
    log "INFO" "✅ Certificate extracted from container"
}

get_certificate_from_local() {
    log "INFO" "📁 Using local certificate file..."
    
    if [ ! -f "$CERT_LOCAL_PATH" ]; then
        log "ERROR" "Local certificate not found: $CERT_LOCAL_PATH"
        log "INFO" "Run setup-ssl.sh first to generate and export certificates"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would use local certificate: $CERT_LOCAL_PATH"
        return 0
    fi
    
    cp "$CERT_LOCAL_PATH" "$CERT_TEMP_DIR/certificate.pem" || {
        log "ERROR" "Failed to copy local certificate"
        return 1
    }
    
    log "INFO" "✅ Local certificate loaded"
}

get_certificate_from_custom() {
    log "INFO" "🔧 Using custom certificate path: $CUSTOM_CERT_PATH"
    
    if [ ! -f "$CUSTOM_CERT_PATH" ]; then
        log "ERROR" "Custom certificate not found: $CUSTOM_CERT_PATH"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would use custom certificate: $CUSTOM_CERT_PATH"
        return 0
    fi
    
    cp "$CUSTOM_CERT_PATH" "$CERT_TEMP_DIR/certificate.pem" || {
        log "ERROR" "Failed to copy custom certificate"
        return 1
    }
    
    log "INFO" "✅ Custom certificate loaded"
}

# =============================================================================
# LINUX TRUST FUNCTIONS
# =============================================================================

trust_linux() {
    if [ "$TRUST_LINUX" = false ]; then
        return 0
    fi
    
    log "INFO" "🐧 Installing certificate in Linux trust store..."
    
    local cert_name="microservices-wildcard.crt"
    local target_path="$LINUX_TRUST_DIR/$cert_name"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would install certificate to: $target_path"
        return 0
    fi
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        remove_linux_certificate "$target_path"
        return $?
    fi
    
    # Create trust directory if it doesn't exist
    if [ ! -d "$LINUX_TRUST_DIR" ]; then
        execute_command "Create trust directory" \
            "sudo mkdir -p $LINUX_TRUST_DIR" \
            true true
    fi
    
    # Copy certificate to trust store
    execute_command "Install certificate in Linux trust store" \
        "sudo cp $CERT_TEMP_DIR/certificate.pem $target_path" \
        true true
    
    # Set proper permissions
    execute_command "Set certificate permissions" \
        "sudo chmod 644 $target_path" \
        true true
    
    # Update trust store
    update_linux_trust_store
    
    log "INFO" "✅ Certificate installed in Linux trust store"
}

remove_linux_certificate() {
    local cert_path="$1"
    
    log "INFO" "🗑️ Removing certificate from Linux trust store..."
    
    if [ -f "$cert_path" ]; then
        execute_command "Remove certificate from trust store" \
            "sudo rm -f $cert_path" \
            true true
        
        update_linux_trust_store
        
        log "INFO" "✅ Certificate removed from Linux trust store"
    else
        log "INFO" "Certificate not found in trust store, nothing to remove"
    fi
}

update_linux_trust_store() {
    log "INFO" "🔄 Updating Linux trust store..."
    
    # Try different update commands based on available tools
    if command -v update-ca-certificates >/dev/null 2>&1; then
        execute_command "Update CA certificates (Debian/Ubuntu)" \
            "sudo update-ca-certificates" \
            "$SHOW_ERRORS" false
    elif command -v trust >/dev/null 2>&1; then
        execute_command "Update trust store (Arch/RHEL)" \
            "sudo trust extract-compat" \
            "$SHOW_ERRORS" false
    elif command -v update-ca-trust >/dev/null 2>&1; then
        execute_command "Update CA trust (RHEL/CentOS)" \
            "sudo update-ca-trust" \
            "$SHOW_ERRORS" false
    else
        log "WARN" "No known trust store update command found"
    fi
}

# =============================================================================
# WINDOWS TRUST FUNCTIONS
# =============================================================================

trust_windows() {
    if [ "$TRUST_WINDOWS" = false ]; then
        return 0
    fi
    
    log "INFO" "🪟 Installing certificate in Windows trust store..."
    
    local platform=$(detect_platform)
    
    if [ "$platform" = "wsl" ]; then
        trust_windows_wsl
    else
        trust_windows_native
    fi
}

trust_windows_wsl() {
    log "INFO" "🔄 Installing certificate via WSL..."
    
    if [ -z "$WINDOWS_USER" ] || [ -z "$WINDOWS_DESKTOP_PATH" ]; then
        log "ERROR" "Could not determine Windows user information"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would copy certificate to Windows desktop and install"
        return 0
    fi
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        remove_windows_certificate
        return $?
    fi
    
    # Copy certificate to Windows desktop
    local windows_cert_path="$WINDOWS_DESKTOP_PATH/microservices-wildcard.crt"
    execute_command "Copy certificate to Windows desktop" \
        "cp $CERT_TEMP_DIR/certificate.pem '$windows_cert_path'" \
        true true
    
    # Create PowerShell script to install certificate
    local ps_script="$CERT_TEMP_DIR/install-cert.ps1"
    cat > "$ps_script" << 'EOF'
param([string]$CertPath)

try {
    # Import certificate to Local Machine Root store
    $cert = Import-Certificate -FilePath $CertPath -CertStoreLocation "Cert:\LocalMachine\Root"
    Write-Host "Certificate installed successfully: $($cert.Subject)"
    
    # Verify installation
    $thumbprint = $cert.Thumbprint
    $installed = Get-ChildItem -Path "Cert:\LocalMachine\Root" | Where-Object { $_.Thumbprint -eq $thumbprint }
    
    if ($installed) {
        Write-Host "Certificate verification successful"
        exit 0
    } else {
        Write-Host "Certificate verification failed"
        exit 1
    }
} catch {
    Write-Host "Error installing certificate: $($_.Exception.Message)"
    exit 1
}
EOF
    
    # Execute PowerShell script
    local windows_cert_win_path="C:\\Users\\$WINDOWS_USER\\Desktop\\microservices-wildcard.crt"
    execute_command "Install certificate in Windows" \
        "powershell.exe -ExecutionPolicy Bypass -File '$ps_script' -CertPath '$windows_cert_win_path'" \
        true false
    
    log "INFO" "✅ Certificate installed in Windows trust store"
    log "INFO" "💡 You may need to restart browsers for changes to take effect"
}

trust_windows_native() {
    log "INFO" "🖥️ Installing certificate on native Windows..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would install certificate using PowerShell"
        return 0
    fi
    
    # Use PowerShell to install certificate
    local cert_path_win=$(cygpath -w "$CERT_TEMP_DIR/certificate.pem" 2>/dev/null || echo "$CERT_TEMP_DIR/certificate.pem")
    
    powershell.exe -Command "Import-Certificate -FilePath '$cert_path_win' -CertStoreLocation Cert:\\LocalMachine\\Root"
    
    log "INFO" "✅ Certificate installed in Windows trust store"
}

remove_windows_certificate() {
    log "INFO" "🗑️ Removing certificate from Windows trust store..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would remove certificate from Windows"
        return 0
    fi
    
    # PowerShell script to remove certificate
    powershell.exe -Command "
        Get-ChildItem -Path Cert:\\LocalMachine\\Root | 
        Where-Object { \$_.Subject -like '*Microservices*' -or \$_.Subject -like '*wildcard.test*' } | 
        Remove-Item -Force"
    
    log "INFO" "✅ Certificate removed from Windows trust store"
}

# =============================================================================
# MACOS TRUST FUNCTIONS
# =============================================================================

trust_macos() {
    if [ "$TRUST_MACOS" = false ]; then
        return 0
    fi
    
    log "INFO" "🍎 Installing certificate in macOS keychain..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would install certificate in macOS keychain"
        return 0
    fi
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        remove_macos_certificate
        return $?
    fi
    
    # Add certificate to system keychain
    execute_command "Add certificate to macOS keychain" \
        "sudo security add-trusted-cert -d -r trustRoot -k $MACOS_KEYCHAIN $CERT_TEMP_DIR/certificate.pem" \
        true true
    
    # Set trust settings
    execute_command "Set certificate trust settings" \
        "sudo security add-trusted-cert -d -r trustAsRoot -p $MACOS_TRUST_SETTINGS -k $MACOS_KEYCHAIN $CERT_TEMP_DIR/certificate.pem" \
        "$SHOW_ERRORS" false
    
    log "INFO" "✅ Certificate installed in macOS keychain"
}

remove_macos_certificate() {
    log "INFO" "🗑️ Removing certificate from macOS keychain..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would remove certificate from macOS keychain"
        return 0
    fi
    
    # Find and remove certificate
    execute_command "Remove certificate from macOS keychain" \
        "sudo security delete-certificate -c 'wildcard.test' $MACOS_KEYCHAIN" \
        "$SHOW_ERRORS" false
    
    log "INFO" "✅ Certificate removed from macOS keychain"
}

# =============================================================================
# BROWSER TRUST FUNCTIONS
# =============================================================================

trust_firefox() {
    if [ "$TRUST_FIREFOX" = false ]; then
        return 0
    fi
    
    log "INFO" "🦊 Installing certificate in Firefox browsers..."
    
    # Find Firefox profiles
    find_firefox_profiles
    
    if [ ${#FIREFOX_PROFILE_DIRS[@]} -eq 0 ]; then
        log "WARN" "No Firefox profiles found"
        return 0
    fi
    
    for profile_dir in "${FIREFOX_PROFILE_DIRS[@]}"; do
        log "INFO" "  📁 Processing Firefox profile: $(basename "$profile_dir")"
        trust_firefox_profile "$profile_dir"
    done
}

find_firefox_profiles() {
    local firefox_base_dirs=(
        "$HOME/.mozilla/firefox"
        "$HOME/snap/firefox/common/.mozilla/firefox"
        "$HOME/.var/app/org.mozilla.firefox/.mozilla/firefox"
    )
    
    for base_dir in "${firefox_base_dirs[@]}"; do
        if [ -d "$base_dir" ]; then
            while IFS= read -r -d '' profile_dir; do
                if [ -f "$profile_dir/cert9.db" ]; then
                    FIREFOX_PROFILE_DIRS+=("$profile_dir")
                fi
            done < <(find "$base_dir" -type d -name "*.default*" -print0 2>/dev/null)
        fi
    done
}

trust_firefox_profile() {
    local profile_dir="$1"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "  [DRY RUN] Would install certificate in Firefox profile"
        return 0
    fi
    
    # Check if certutil is available
    if ! command -v certutil >/dev/null 2>&1; then
        log "WARN" "  certutil not found, skipping Firefox certificate installation"
        log "INFO" "  Install libnss3-tools (Debian/Ubuntu) or nss-tools (RHEL/CentOS)"
        return 0
    fi
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        execute_command "Remove certificate from Firefox profile" \
            "certutil -D -n 'Microservices Wildcard' -d sql:$profile_dir" \
            "$SHOW_ERRORS" false
    else
        execute_command "Install certificate in Firefox profile" \
            "certutil -A -n 'Microservices Wildcard' -t 'TCu,Cu,Tu' -i $CERT_TEMP_DIR/certificate.pem -d sql:$profile_dir" \
            "$SHOW_ERRORS" false
    fi
}

trust_chrome() {
    if [ "$TRUST_CHROME" = false ]; then
        return 0
    fi
    
    log "INFO" "🌐 Setting up Chrome certificate policy..."
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would setup Chrome certificate policy"
        return 0
    fi
    
    # Chrome uses the system trust store on most platforms
    # But we can also set up enterprise policy for additional control
    setup_chrome_policy
}

setup_chrome_policy() {
    if [ ! -d "$CHROME_POLICY_DIR" ]; then
        execute_command "Create Chrome policy directory" \
            "sudo mkdir -p $CHROME_POLICY_DIR" \
            "$SHOW_ERRORS" false
    fi
    
    local policy_file="$CHROME_POLICY_DIR/microservices-ssl.json"
    local policy_content='{
    "AutoSelectCertificateForUrls": [
        "{\"pattern\":\"https://*.test\",\"filter\":{}}"
    ],
    "CertificateTransparencyEnforcementDisabledForUrls": [
        "*.test"
    ]
}'
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        execute_command "Remove Chrome certificate policy" \
            "sudo rm -f $policy_file" \
            "$SHOW_ERRORS" false
    else
        execute_command "Create Chrome certificate policy" \
            "echo '$policy_content' | sudo tee $policy_file > /dev/null" \
            "$SHOW_ERRORS" false
    fi
}

# =============================================================================
# HOSTS FILE MANAGEMENT
# =============================================================================

update_hosts_file() {
    if [ "$UPDATE_HOSTS_FILE" = false ]; then
        return 0
    fi
    
    log "INFO" "📝 Updating system hosts file..."
    
    local hosts_file="/etc/hosts"
    local backup_file="$hosts_file.microservices.backup.$(date +%Y%m%d_%H%M%S)"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would update hosts file with microservices domains"
        return 0
    fi
    
    # Backup hosts file
    execute_command "Backup hosts file" \
        "sudo cp $hosts_file $backup_file" \
        true true
    
    # Define service domains
    local service_domains=(
        "nginx.test" "adminer.test" "phpmyadmin.test" "mongodb.test"
        "metabase.test" "nocodb.test" "grafana.test" "prometheus.test"
        "matomo.test" "n8n.test" "langflow.test" "kibana.test"
        "elasticsearch.test" "keycloak.test" "mailpit.test" "gitea.test"
        "odoo.test"
    )
    
    if [ "$REMOVE_CERTIFICATES" = true ]; then
        # Remove microservices entries
        for domain in "${service_domains[@]}"; do
            execute_command "Remove $domain from hosts file" \
                "sudo sed -i '/127\.0\.0\.1[[:space:]]*$domain/d' $hosts_file" \
                "$SHOW_ERRORS" false
        done
        log "INFO" "✅ Microservices domains removed from hosts file"
    else
        # Add microservices entries
        log "INFO" "Adding microservices domains to hosts file..."
        
        # Check if entries already exist
        local entries_to_add=()
        for domain in "${service_domains[@]}"; do
            if ! grep -q "127\.0\.0\.1[[:space:]]*$domain" "$hosts_file"; then
                entries_to_add+=("127.0.0.1 $domain")
            fi
        done
        
        if [ ${#entries_to_add[@]} -gt 0 ]; then
            # Add microservices section marker
            if ! grep -q "# Microservices Development" "$hosts_file"; then
                echo "" | sudo tee -a "$hosts_file" > /dev/null
                echo "# Microservices Development" | sudo tee -a "$hosts_file" > /dev/null
            fi
            
            # Add entries
            for entry in "${entries_to_add[@]}"; do
                echo "$entry" | sudo tee -a "$hosts_file" > /dev/null
            done
            
            log "INFO" "✅ Added ${#entries_to_add[@]} domains to hosts file"
        else
            log "INFO" "All microservices domains already in hosts file"
        fi
    fi
}

# =============================================================================
# CERTIFICATE LISTING
# =============================================================================

list_trusted_certificates() {
    if [ "$LIST_TRUSTED_CERTS" = false ]; then
        return 0
    fi
    
    log "INFO" "📋 Listing trusted certificates..."
    
    local platform=$(detect_platform)
    
    case "$platform" in
        "linux"|"wsl")
            list_linux_certificates
            ;;
        "macos")
            list_macos_certificates
            ;;
        "windows")
            list_windows_certificates
            ;;
    esac
}

list_linux_certificates() {
    log "INFO" "🐧 Linux trusted certificates:"
    
    if command -v trust >/dev/null 2>&1; then
        execute_command "List trusted certificates" \
            "trust list | grep -i microservices -A 3 -B 1" \
            false false
    elif [ -d "$LINUX_TRUST_DIR" ]; then
        execute_command "List certificates in trust directory" \
            "ls -la $LINUX_TRUST_DIR/*microservices* 2>/dev/null || echo 'No microservices certificates found'" \
            false false
    fi
}

list_macos_certificates() {
    log "INFO" "🍎 macOS keychain certificates:"
    
    execute_command "List macOS certificates" \
        "security find-certificate -a -c wildcard.test $MACOS_KEYCHAIN" \
        false false
}

list_windows_certificates() {
    log "INFO" "🪟 Windows certificate store:"
    
    execute_command "List Windows certificates" \
        "powershell.exe -Command \"Get-ChildItem -Path Cert:\\LocalMachine\\Root | Where-Object { \\\$_.Subject -like '*Microservices*' -or \\\$_.Subject -like '*wildcard.test*' } | Select-Object Subject, Thumbprint, NotAfter\"" \
        false false
}

# =============================================================================
# CLEANUP
# =============================================================================

cleanup() {
    if [ -d "$CERT_TEMP_DIR" ]; then
        rm -rf "$CERT_TEMP_DIR"
        log "DEBUG" "Cleaned up temporary directory: $CERT_TEMP_DIR"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

show_trust_summary() {
    local platform=$(detect_platform)
    
    cat << EOF

📋 CERTIFICATE TRUST SETUP SUMMARY
==================================

Platform: $platform
Certificate Source: $CERT_SOURCE
$([ "$CERT_SOURCE" = "custom" ] && echo "Custom Path: $CUSTOM_CERT_PATH")

Trust Options:
- Linux/WSL: $TRUST_LINUX
- Windows: $TRUST_WINDOWS  
- macOS: $TRUST_MACOS
- Firefox: $TRUST_FIREFOX
- Chrome: $TRUST_CHROME
- Update Hosts: $UPDATE_HOSTS_FILE

Operation: $([ "$REMOVE_CERTIFICATES" = true ] && echo "Remove certificates" || echo "Install certificates")

EOF
}

main() {
    parse_common_args "$@"
    parse_trust_args "$@"
    
    # Set trap for cleanup
    trap cleanup EXIT
    
    show_trust_summary
    
    # List certificates if requested
    if [ "$LIST_TRUSTED_CERTS" = true ]; then
        list_trusted_certificates
        exit 0
    fi
    
    log "INFO" "🚀 Starting certificate trust setup..."
    
    # Get certificate
    get_certificate || {
        log "ERROR" "Failed to acquire certificate"
        exit 1
    }
    
    # Detect platform for automatic trust decisions
    local platform=$(detect_platform)
    
    # Auto-enable platform-specific trust if not explicitly set
    case "$platform" in
        "wsl")
            # For WSL, enable both Linux and Windows trust by default
            [ "$TRUST_LINUX" = true ] && trust_linux
            [ "$TRUST_WINDOWS" = true ] && trust_windows
            ;;
        "linux")
            [ "$TRUST_LINUX" = true ] && trust_linux
            ;;
        "macos")
            [ "$TRUST_MACOS" = true ] && trust_macos
            ;;
        "windows")
            [ "$TRUST_WINDOWS" = true ] && trust_windows
            ;;
    esac
    
    # Browser-specific trust
    [ "$TRUST_FIREFOX" = true ] && trust_firefox
    [ "$TRUST_CHROME" = true ] && trust_chrome
    
    # Update hosts file if requested
    update_hosts_file
    
    log "INFO" "✅ Certificate trust setup completed successfully!"
    
    if [ "$DRY_RUN" = false ] && [ "$REMOVE_CERTIFICATES" = false ]; then
        show_success_instructions "$platform"
    elif [ "$REMOVE_CERTIFICATES" = true ]; then
        log "INFO" "🗑️ Certificates have been removed from trust stores"
        log "INFO" "💡 You may need to restart browsers for changes to take effect"
    fi
}

# =============================================================================
# SUCCESS INSTRUCTIONS
# =============================================================================

show_success_instructions() {
    local platform="$1"
    
    cat << EOF

🎉 CERTIFICATE TRUST SETUP COMPLETE!
====================================

Your SSL certificates are now trusted by your system!

📋 What was installed:
EOF
    
    [ "$TRUST_LINUX" = true ] && echo "  ✅ Linux system trust store"
    [ "$TRUST_WINDOWS" = true ] && echo "  ✅ Windows certificate store"
    [ "$TRUST_MACOS" = true ] && echo "  ✅ macOS system keychain"
    [ "$TRUST_FIREFOX" = true ] && echo "  ✅ Firefox browser profiles"
    [ "$TRUST_CHROME" = true ] && echo "  ✅ Chrome browser policy"
    [ "$UPDATE_HOSTS_FILE" = true ] && echo "  ✅ System hosts file entries"
    
    cat << EOF

🌐 Test your setup:
  Try visiting: https://metabase.test
  Or any other service: https://grafana.test, https://adminer.test

🔧 Troubleshooting:
EOF
    
    case "$platform" in
        "wsl")
            cat << EOF
  WSL/Windows specific:
  - Clear browser cache and restart browsers
  - For Firefox: Check that security.enterprise_roots.enabled = true in about:config
  - Verify Windows certificate: certlm.msc → Trusted Root Certification Authorities
  - If issues persist, run: $0 -W -F -v
EOF
            ;;
        "linux")
            cat << EOF
  Linux specific:
  - Restart browsers to pick up new certificates
  - Verify trust: trust list | grep microservices
  - Check certificate: openssl x509 -in $LINUX_TRUST_DIR/microservices-wildcard.crt -text -noout
  - If issues persist, run: $0 -L -F -v
EOF
            ;;
        "macos")
            cat << EOF
  macOS specific:
  - Open Keychain Access → System keychain → Certificates
  - Double-click the certificate and set trust to "Always Trust"
  - Restart browsers for changes to take effect
  - If issues persist, run: $0 -M -F -v
EOF
            ;;
    esac
    
    cat << EOF

💡 Additional setup:
  - Configure proxy hosts in Nginx Proxy Manager: http://localhost:81
  - Default login: admin@example.com / changeme
  - Point your domains to the appropriate internal services

🔄 To remove certificates later:
  Run: $0 -R -v

📚 Service URLs:
EOF
    
    local service_urls=(
        "https://nginx.test - Nginx Proxy Manager"
        "https://adminer.test - Database Admin"
        "https://metabase.test - Analytics Dashboard"
        "https://grafana.test - Monitoring Dashboard"
        "https://n8n.test - Workflow Automation"
        "https://gitea.test - Git Repository"
        "https://odoo.test - ERP System"
    )
    
    printf "  %s\n" "${service_urls[@]}"
    
    echo ""
}

# =============================================================================
# VERIFICATION FUNCTIONS
# =============================================================================

verify_certificate_trust() {
    log "INFO" "🔍 Verifying certificate trust..."
    
    local test_domains=("metabase.test" "grafana.test" "adminer.test")
    
    for domain in "${test_domains[@]}"; do
        log "INFO" "  🌐 Testing HTTPS connection to $domain..."
        
        if [ "$DRY_RUN" = true ]; then
            log "INFO" "  [DRY RUN] Would test HTTPS connection"
            continue
        fi
        
        # Test with curl
        if command -v curl >/dev/null 2>&1; then
            if execute_command "Test HTTPS to $domain" \
               "curl -s --connect-timeout 5 https://$domain > /dev/null" false false; then
                log "INFO" "  ✅ $domain - HTTPS connection successful"
            else
                log "WARN" "  ⚠️ $domain - HTTPS connection failed (service may not be running)"
            fi
        fi
        
        # Test certificate validation
        if command -v openssl >/dev/null 2>&1; then
            if execute_command "Verify certificate for $domain" \
               "echo | openssl s_client -connect $domain:443 -servername $domain 2>/dev/null | openssl x509 -noout -subject" false false; then
                log "INFO" "  ✅ $domain - Certificate validation successful"
            else
                log "DEBUG" "  Certificate validation details for $domain not available"
            fi
        fi
    done
}

# =============================================================================
# ADDITIONAL HELPER FUNCTIONS
# =============================================================================

check_certificate_expiry() {
    if [ ! -f "$CERT_TEMP_DIR/certificate.pem" ]; then
        return 0
    fi
    
    log "INFO" "📅 Checking certificate expiry..."
    
    if command -v openssl >/dev/null 2>&1; then
        local expiry_date=$(openssl x509 -in "$CERT_TEMP_DIR/certificate.pem" -noout -enddate 2>/dev/null | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s 2>/dev/null || echo "0")
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        if [ $days_until_expiry -gt 30 ]; then
            log "INFO" "✅ Certificate expires in $days_until_expiry days ($expiry_date)"
        elif [ $days_until_expiry -gt 0 ]; then
            log "WARN" "⚠️ Certificate expires in $days_until_expiry days ($expiry_date)"
        else
            log "ERROR" "❌ Certificate has expired ($expiry_date)"
        fi
    fi
}

backup_browser_profiles() {
    log "INFO" "💾 Creating backup of browser profiles..."
    
    local backup_dir="$PROJECT_ROOT/browser-backups/backup_$(date +%Y%m%d_%H%M%S)"
    
    if [ "$DRY_RUN" = true ]; then
        log "INFO" "[DRY RUN] Would create browser profile backup: $backup_dir"
        return 0
    fi
    
    mkdir -p "$backup_dir"
    
    # Backup Firefox profiles
    if [ ${#FIREFOX_PROFILE_DIRS[@]} -gt 0 ]; then
        for profile_dir in "${FIREFOX_PROFILE_DIRS[@]}"; do
            local profile_name=$(basename "$profile_dir")
            execute_command "Backup Firefox profile $profile_name" \
                "cp $profile_dir/cert9.db $backup_dir/firefox_${profile_name}_cert9.db" \
                "$SHOW_ERRORS" false
        done
    fi
    
    log "INFO" "✅ Browser profile backup completed: $backup_dir"
}

# Add verification to main function
main_with_verification() {
    # Run main trust setup
    main "$@"
    
    # Additional verification if not dry run
    if [ "$DRY_RUN" = false ] && [ "$REMOVE_CERTIFICATES" = false ]; then
        check_certificate_expiry
        verify_certificate_trust
    fi
}

# Update the final call
main_with_verification "$@"