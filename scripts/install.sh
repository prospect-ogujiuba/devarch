#!/bin/bash
# =================================================================
# DEVARCH INSTALLATION SCRIPT
# =================================================================
# This script sets up the complete DevArch development environment
# with all services, SSL certificates, and development tools.

set -euo pipefail

# =================================================================
# CONFIGURATION AND VARIABLES
# =================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_DIR="$PROJECT_ROOT/compose"
CONFIG_DIR="$PROJECT_ROOT/config"
APPS_DIR="$PROJECT_ROOT/apps"
LOGS_DIR="$PROJECT_ROOT/logs"

# Load environment variables
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
else
    echo "❌ Error: .env file not found in $PROJECT_ROOT"
    echo "Please ensure the .env file exists before running the installer."
    exit 1
fi

# Script options
VERBOSE=false
SKIP_CONFIRM=false
ENABLE_MONGODB=false
ENABLE_MONITORING=false
ENABLE_QUALITY=false
USE_SUDO=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# =================================================================
# UTILITY FUNCTIONS
# =================================================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${PURPLE}[$(date +'%Y-%m-%d %H:%M:%S')] DEBUG: $1${NC}"
    fi
}

# Progress indicator
show_progress() {
    local duration=$1
    local message=$2
    
    echo -n "$message"
    for ((i=0; i<duration; i++)); do
        echo -n "."
        sleep 1
    done
    echo " ✓"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Execute command with optional sudo
execute_cmd() {
    local cmd="$1"
    if [[ "$USE_SUDO" == "true" ]]; then
        sudo bash -c "$cmd"
    else
        bash -c "$cmd"
    fi
}

# =================================================================
# VALIDATION FUNCTIONS
# =================================================================
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check for required commands
    local required_commands=("podman" "curl" "openssl")
    local missing_commands=()
    
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            missing_commands+=("$cmd")
        fi
    done
    
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        error "Missing required commands: ${missing_commands[*]}"
        echo
        echo "Please install the missing commands and try again:"
        echo "- Podman: https://podman.io/getting-started/installation"
        echo "- curl: Usually available in system packages"
        echo "- openssl: Usually available in system packages"
        exit 1
    fi
    
    # Check podman version
    local podman_version
    podman_version=$(podman --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    debug "Detected Podman version: $podman_version"
    
    # Check if podman is running
    if ! podman info >/dev/null 2>&1; then
        error "Podman is not running or not accessible"
        echo "Please ensure Podman is installed and running, then try again."
        echo "You may need to start the Podman socket: systemctl --user start podman.socket"
        exit 1
    fi
    
    log "Prerequisites check passed ✓"
}

validate_environment() {
    log "Validating environment configuration..."
    
    # Check required environment variables
    local required_vars=(
        "COMPOSE_PROJECT_NAME"
        "NETWORK_NAME"
        "MYSQL_ROOT_PASSWORD"
        "POSTGRES_PASSWORD"
        "REDIS_PASSWORD"
        "ADMIN_USER"
        "ADMIN_PASSWORD"
    )
    
    local missing_vars=()
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        error "Missing required environment variables: ${missing_vars[*]}"
        echo "Please check your .env file and ensure all required variables are set."
        exit 1
    fi
    
    # Validate password strength (basic check)
    if [[ ${#ADMIN_PASSWORD} -lt 6 ]]; then
        warn "Admin password is less than 6 characters. Consider using a stronger password."
    fi
    
    log "Environment validation passed ✓"
}

# =================================================================
# DIRECTORY SETUP
# =================================================================
create_directories() {
    log "Creating directory structure..."
    
    local directories=(
        "$APPS_DIR"
        "$LOGS_DIR/nginx"
        "$LOGS_DIR/php"
        "$LOGS_DIR/supervisor"
        "$CONFIG_DIR/nginx"
        "$CONFIG_DIR/php"
        "$CONFIG_DIR/databases/mariadb"
        "$CONFIG_DIR/databases/postgres"
        "$CONFIG_DIR/databases/redis"
        "$CONFIG_DIR/databases/mongodb"
        "$CONFIG_DIR/ssl"
    )
    
    for dir in "${directories[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            debug "Created directory: $dir"
        fi
    done
    
    # Create example app structure
    create_example_app
    
    log "Directory structure created ✓"
}

create_example_app() {
    local example_dir="$APPS_DIR/welcome"
    
    if [[ ! -d "$example_dir" ]]; then
        mkdir -p "$example_dir/public"
        
        # Create a simple welcome page
        cat > "$example_dir/public/index.php" << 'EOF'
<?php
$apps = array_filter(glob('/var/www/html/*'), 'is_dir');
$apps = array_map('basename', $apps);
sort($apps);
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DevArch - Development Environment</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: #f8fafc; color: #2d3748; line-height: 1.6; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .header { text-align: center; margin-bottom: 3rem; }
        .logo { font-size: 3rem; font-weight: bold; color: #3182ce; margin-bottom: 1rem; }
        .subtitle { font-size: 1.2rem; color: #718096; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 2rem; margin-bottom: 3rem; }
        .card { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 6px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; }
        .card h3 { color: #2d3748; margin-bottom: 1rem; font-size: 1.3rem; }
        .card ul { list-style: none; }
        .card li { margin-bottom: 0.5rem; }
        .card a { color: #3182ce; text-decoration: none; font-weight: 500; }
        .card a:hover { color: #2c5aa0; text-decoration: underline; }
        .status { display: inline-block; padding: 0.25rem 0.75rem; border-radius: 20px; font-size: 0.875rem; font-weight: 500; }
        .status.running { background: #c6f6d5; color: #22543d; }
        .status.stopped { background: #fed7d7; color: #822727; }
        .apps-list { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .app-item { display: flex; justify-content: space-between; align-items: center; padding: 1rem; border-bottom: 1px solid #e2e8f0; }
        .app-item:last-child { border-bottom: none; }
        .app-name { font-weight: 600; color: #2d3748; }
        .app-url { font-family: monospace; background: #f7fafc; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.875rem; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🚀 DevArch</div>
            <div class="subtitle">Development Architecture Environment</div>
        </div>
        
        <div class="grid">
            <div class="card">
                <h3>🗄️ Database Tools</h3>
                <ul>
                    <li><a href="https://adminer.test">Adminer</a> - Universal database tool</li>
                    <li><a href="https://phpmyadmin.test">phpMyAdmin</a> - MySQL management</li>
                    <li><a href="https://metabase.test">Metabase</a> - Business intelligence</li>
                    <li><a href="https://redis.test">Redis Commander</a> - Redis management</li>
                </ul>
            </div>
            
            <div class="card">
                <h3>📧 Development Tools</h3>
                <ul>
                    <li><a href="https://mailpit.test">Mailpit</a> - Email testing</li>
                    <li><a href="https://swagger.test">Swagger UI</a> - API documentation</li>
                    <li><a href="https://nginx.test">Nginx Admin</a> - Proxy management</li>
                </ul>
            </div>
            
            <div class="card">
                <h3>⚙️ System Status</h3>
                <ul>
                    <li>PHP Version: <?= phpversion() ?> <span class="status running">Running</span></li>
                    <li>MariaDB: <span class="status running">Running</span></li>
                    <li>PostgreSQL: <span class="status running">Running</span></li>
                    <li>Redis: <span class="status running">Running</span></li>
                    <li>Mailpit: <span class="status running">Running</span></li>
                </ul>
            </div>
        </div>
        
        <div class="apps-list">
            <h3 style="margin-bottom: 1.5rem;">📱 Your Applications</h3>
            <?php if (count($apps) > 1): ?>
                <?php foreach ($apps as $app): ?>
                    <?php if ($app !== 'welcome'): ?>
                        <div class="app-item">
                            <div class="app-name"><?= ucfirst($app) ?></div>
                            <a href="https://<?= $app ?>.test" class="app-url">https://<?= $app ?>.test</a>
                        </div>
                    <?php endif; ?>
                <?php endforeach; ?>
            <?php else: ?>
                <div style="text-align: center; color: #718096; padding: 2rem;">
                    <p>No applications found. Create your first app in the <code>./apps/</code> directory!</p>
                    <p style="margin-top: 1rem;">Example: Create <code>./apps/my-app/public/index.php</code> and access it at <code>https://my-app.test</code></p>
                </div>
            <?php endif; ?>
        </div>
    </div>
</body>
</html>
EOF
        
        debug "Created welcome application"
    fi
}

# =================================================================
# NETWORK SETUP
# =================================================================
setup_network() {
    log "Setting up container network..."
    
    # Check if network already exists
    if podman network exists "$NETWORK_NAME" 2>/dev/null; then
        info "Network '$NETWORK_NAME' already exists"
    else
        debug "Creating network: $NETWORK_NAME"
        podman network create \
            --driver bridge \
            --subnet 172.20.0.0/16 \
            "$NETWORK_NAME"
        log "Network '$NETWORK_NAME' created ✓"
    fi
}

# =================================================================
# SERVICE DEPLOYMENT
# =================================================================
deploy_core_services() {
    log "Deploying core services..."
    
    cd "$COMPOSE_DIR"
    
    # Start core infrastructure
    info "Starting database services..."
    podman compose -f core.docker-compose.yml up -d mariadb postgres redis
    
    # Wait for databases to be ready
    show_progress 30 "Waiting for databases to initialize"
    
    # Start PHP runtime
    info "Starting PHP runtime..."
    podman compose -f core.docker-compose.yml up -d php-runtime
    
    # Start mail service
    info "Starting mail service..."
    podman compose -f core.docker-compose.yml up -d mailpit
    
    # Wait a bit more for services to stabilize
    show_progress 15 "Waiting for services to stabilize"
    
    log "Core services deployed ✓"
}

deploy_development_tools() {
    log "Deploying development tools..."
    
    cd "$COMPOSE_DIR"
    
    local profiles=()
    
    if [[ "$ENABLE_MONGODB" == "true" ]]; then
        profiles+=(--profile mongodb)
        info "MongoDB profile enabled"
    fi
    
    if [[ "$ENABLE_QUALITY" == "true" ]]; then
        profiles+=(--profile quality)
        info "Quality tools profile enabled"
    fi
    
    # Deploy development tools
    podman compose "${profiles[@]}" -f development.docker-compose.yml up -d
    
    show_progress 20 "Waiting for development tools to start"
    
    log "Development tools deployed ✓"
}

deploy_nginx_proxy() {
    log "Deploying Nginx proxy manager..."
    
    cd "$COMPOSE_DIR"
    
    # Start Nginx Proxy Manager
    podman compose -f core.docker-compose.yml up -d nginx-proxy-manager
    
    show_progress 30 "Waiting for Nginx to start"
    
    log "Nginx proxy deployed ✓"
}

# =================================================================
# SSL CERTIFICATE SETUP
# =================================================================
setup_ssl_certificates() {
    log "Setting up SSL certificates..."
    
    # Wait for Nginx to be fully ready
    local max_attempts=30
    local attempt=1
    
    while ! podman exec nginx-proxy-manager curl -f http://localhost:80 >/dev/null 2>&1; do
        if [[ $attempt -ge $max_attempts ]]; then
            error "Nginx container is not responding after $max_attempts attempts"
            return 1
        fi
        debug "Waiting for Nginx to be ready (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    # Create SSL directory in container
    podman exec nginx-proxy-manager mkdir -p /etc/letsencrypt/live/wildcard.test
    
    # Generate wildcard certificate
    info "Generating wildcard SSL certificate..."
    podman exec nginx-proxy-manager openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/letsencrypt/live/wildcard.test/privkey.pem \
        -out /etc/letsencrypt/live/wildcard.test/fullchain.pem \
        -subj "/C=${SSL_COUNTRY}/ST=${SSL_STATE}/L=${SSL_CITY}/O=${SSL_ORG}/CN=*.test" \
        -addext "subjectAltName=DNS:*.test,DNS:test,DNS:localhost,DNS:mailpit.test,DNS:adminer.test,DNS:phpmyadmin.test,DNS:metabase.test,DNS:redis.test,DNS:swagger.test,DNS:nginx.test,DNS:welcome.test"
    
    # Restart Nginx to load certificates
    info "Restarting Nginx to load certificates..."
    podman restart nginx-proxy-manager
    
    show_progress 10 "Waiting for Nginx to restart"
    
    log "SSL certificates configured ✓"
}

trust_ssl_certificates() {
    log "Installing SSL certificates on host system..."
    
    local cert_dir="/tmp/devarch-ssl-$$"
    mkdir -p "$cert_dir"
    
    # Copy certificate from container
    podman cp nginx-proxy-manager:/etc/letsencrypt/live/wildcard.test/fullchain.pem "$cert_dir/devarch.crt"
    
    # Install certificate based on OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command_exists update-ca-certificates; then
            execute_cmd "cp $cert_dir/devarch.crt /usr/local/share/ca-certificates/"
            execute_cmd "update-ca-certificates"
            log "SSL certificate installed on Linux ✓"
        else
            warn "Could not install SSL certificate automatically on this Linux distribution"
            warn "Please manually install: $cert_dir/devarch.crt"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists security; then
            security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "$cert_dir/devarch.crt"
            log "SSL certificate installed on macOS ✓"
        else
            warn "Could not install SSL certificate automatically on macOS"
            warn "Please manually install: $cert_dir/devarch.crt"
        fi
    else
        warn "Automatic SSL certificate installation not supported on this OS"
        warn "Please manually install: $cert_dir/devarch.crt"
    fi
    
    # Cleanup
    rm -rf "$cert_dir"
}

# =================================================================
# HEALTH CHECKS
# =================================================================
verify_deployment() {
    log "Verifying deployment..."
    
    local services=(
        "mariadb:3306"
        "postgres:5432"
        "redis:6379"
        "mailpit:8025"
        "nginx-proxy-manager:80"
    )
    
    local failed_services=()
    
    for service in "${services[@]}"; do
        local container="${service%:*}"
        local port="${service#*:}"
        
        if podman exec "$container" nc -z localhost "$port" 2>/dev/null; then
            debug "✓ $container is responding on port $port"
        else
            failed_services+=("$container")
            error "✗ $container is not responding on port $port"
        fi
    done
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        error "Some services failed health checks: ${failed_services[*]}"
        return 1
    fi
    
    log "All services are healthy ✓"
}

# =================================================================
# MAIN INSTALLATION FLOW
# =================================================================
show_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                      🚀 DevArch Installer                    ║"
    echo "║              Development Architecture Environment            ║"
    echo "║                                                              ║"
    echo "║  This installer will set up a complete development          ║"
    echo "║  environment with databases, tools, and SSL certificates.   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -y, --yes              Skip confirmation prompts"
    echo "  -s, --sudo             Use sudo for system operations"
    echo "  --enable-mongodb       Enable MongoDB services"
    echo "  --enable-monitoring    Enable monitoring services"
    echo "  --enable-quality       Enable code quality tools"
    echo "  -h, --help             Show this help message"
    echo
    echo "Examples:"
    echo "  $0                     # Basic installation"
    echo "  $0 -y --enable-mongodb # Auto-confirm with MongoDB"
    echo "  $0 -v -s               # Verbose output with sudo"
}

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -y|--yes)
                SKIP_CONFIRM=true
                shift
                ;;
            -s|--sudo)
                USE_SUDO=true
                shift
                ;;
            --enable-mongodb)
                ENABLE_MONGODB=true
                shift
                ;;
            --enable-monitoring)
                ENABLE_MONITORING=true
                shift
                ;;
            --enable-quality)
                ENABLE_QUALITY=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

confirm_installation() {
    if [[ "$SKIP_CONFIRM" == "true" ]]; then
        return 0
    fi
    
    echo "Installation Configuration:"
    echo "  • Project: $COMPOSE_PROJECT_NAME"
    echo "  • Network: $NETWORK_NAME"
    echo "  • MongoDB: $([ "$ENABLE_MONGODB" == "true" ] && echo "Enabled" || echo "Disabled")"
    echo "  • Monitoring: $([ "$ENABLE_MONITORING" == "true" ] && echo "Enabled" || echo "Disabled")"
    echo "  • Quality Tools: $([ "$ENABLE_QUALITY" == "true" ] && echo "Enabled" || echo "Disabled")"
    echo "  • Use Sudo: $([ "$USE_SUDO" == "true" ] && echo "Yes" || echo "No")"
    echo
    
    read -p "Do you want to continue with the installation? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
}

show_completion_message() {
    echo
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                   🎉 Installation Complete!                  ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${CYAN}Access your development environment:${NC}"
    echo
    echo -e "${YELLOW}Core Services:${NC}"
    echo "  🏠 Welcome Page:       https://welcome.test"
    echo "  📧 Email Testing:      https://mailpit.test"
    echo "  🔧 Nginx Admin:        https://nginx.test"
    echo
    echo -e "${YELLOW}Database Tools:${NC}"
    echo "  🗄️ Adminer:            https://adminer.test"
    echo "  🐬 phpMyAdmin:         https://phpmyadmin.test"
    echo "  📊 Metabase:           https://metabase.test"
    echo "  🔴 Redis Commander:    https://redis.test"
    
    if [[ "$ENABLE_MONGODB" == "true" ]]; then
        echo "  🍃 MongoDB Express:    https://mongodb.test"
    fi
    
    echo
    echo -e "${YELLOW}Development:${NC}"
    echo "  📝 API Docs:           https://swagger.test"
    
    if [[ "$ENABLE_QUALITY" == "true" ]]; then
        echo "  🔍 Code Quality:       https://sonarqube.test"
    fi
    
    echo
    echo -e "${YELLOW}Getting Started:${NC}"
    echo "  1. Create a new app:   mkdir -p ./apps/my-app/public"
    echo "  2. Add index file:     echo '<?php echo \"Hello World!\";' > ./apps/my-app/public/index.php"
    echo "  3. Access your app:    https://my-app.test"
    echo
    echo -e "${YELLOW}Management:${NC}"
    echo "  • Start services:      ./scripts/manage.sh start"
    echo "  • Stop services:       ./scripts/manage.sh stop"
    echo "  • View logs:           ./scripts/manage.sh logs [service]"
    echo "  • Service status:      ./scripts/manage.sh status"
    echo
    echo -e "${GREEN}Happy coding! 🚀${NC}"
}

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Show banner
    show_banner
    
    # Confirm installation
    confirm_installation
    
    # Run installation steps
    check_prerequisites
    validate_environment
    create_directories
    setup_network
    deploy_core_services
    deploy_development_tools
    deploy_nginx_proxy
    setup_ssl_certificates
    trust_ssl_certificates
    verify_deployment
    
    # Show completion message
    show_completion_message
}

# =================================================================
# ERROR HANDLING
# =================================================================
cleanup_on_error() {
    error "Installation failed. Cleaning up..."
    
    # Stop all services
    cd "$COMPOSE_DIR" 2>/dev/null || true
    podman compose -f core.docker-compose.yml down 2>/dev/null || true
    podman compose -f development.docker-compose.yml down 2>/dev/null || true
    
    # Remove network
    podman network rm "$NETWORK_NAME" 2>/dev/null || true
    
    echo "Cleanup completed. Please check the error messages above and try again."
    exit 1
}

# Set up error handling
trap cleanup_on_error ERR

# Run main function
main "$@"