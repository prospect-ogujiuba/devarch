#!/usr/bin/env bash
# app-templates.sh - Shared functions for application template management
# Part of DevArch development environment

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEVARCH_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TEMPLATES_DIR="${DEVARCH_ROOT}/templates"
APPS_DIR="${DEVARCH_ROOT}/apps"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_step() {
    echo -e "${CYAN}==>${NC} $*"
}

# List all available templates
list_templates() {
    log_info "Available templates:\n"

    echo -e "${GREEN}PHP Applications:${NC}"
    echo "  1. php/laravel         - Laravel framework"
    echo "  2. php/wordpress       - WordPress CMS"
    echo "  3. php/vanilla         - Plain PHP application"

    echo -e "\n${GREEN}Node.js Applications:${NC}"
    echo "  4. node/react-vite     - React + Vite SPA"
    echo "  5. node/nextjs         - Next.js with static export"
    echo "  6. node/express        - Express.js server"
    echo "  7. node/vue            - Vue.js SPA"

    echo -e "\n${GREEN}Python Applications:${NC}"
    echo "  8. python/django       - Django framework"
    echo "  9. python/flask        - Flask framework"
    echo " 10. python/fastapi      - FastAPI framework"

    echo -e "\n${GREEN}Go Applications:${NC}"
    echo " 11. go/gin              - Gin framework"
    echo " 12. go/echo             - Echo framework"

    echo -e "\n${GREEN}.NET Applications:${NC}"
    echo " 13. dotnet/aspnet-core  - ASP.NET Core"
}

# Get template by number
get_template_by_number() {
    local num=$1
    case $num in
        1) echo "php/laravel" ;;
        2) echo "php/wordpress" ;;
        3) echo "php/vanilla" ;;
        4) echo "node/react-vite" ;;
        5) echo "node/nextjs" ;;
        6) echo "node/express" ;;
        7) echo "node/vue" ;;
        8) echo "python/django" ;;
        9) echo "python/flask" ;;
        10) echo "python/fastapi" ;;
        11) echo "go/gin" ;;
        12) echo "go/echo" ;;
        13) echo "dotnet/aspnet-core" ;;
        *) echo "" ;;
    esac
}

# Validate template exists
validate_template() {
    local template=$1
    local template_path="${TEMPLATES_DIR}/${template}"

    if [[ ! -d "$template_path" ]]; then
        log_error "Template not found: $template"
        log_info "Template path: $template_path"
        return 1
    fi

    return 0
}

# Validate app name
validate_app_name() {
    local name=$1

    # Check if name is empty
    if [[ -z "$name" ]]; then
        log_error "App name cannot be empty"
        return 1
    fi

    # Check if name contains only valid characters (lowercase, numbers, hyphens)
    if [[ ! "$name" =~ ^[a-z0-9-]+$ ]]; then
        log_error "App name must contain only lowercase letters, numbers, and hyphens"
        return 1
    fi

    # Check if app already exists
    if [[ -d "${APPS_DIR}/${name}" ]]; then
        log_error "App already exists: $name"
        return 1
    fi

    return 0
}

# Get default port for template type
get_default_port() {
    local template=$1

    case $template in
        php/*) echo "8100" ;;
        node/*) echo "8200" ;;
        python/*) echo "8300" ;;
        go/*) echo "8400" ;;
        dotnet/*) echo "8600" ;;
        *) echo "8000" ;;
    esac
}

# Get runtime name for template
get_runtime() {
    local template=$1

    case $template in
        php/*) echo "php" ;;
        node/*) echo "nodejs" ;;
        python/*) echo "python" ;;
        go/*) echo "golang" ;;
        dotnet/*) echo "dotnet" ;;
        *) echo "unknown" ;;
    esac
}

# Get package manager for template
get_package_manager() {
    local template=$1

    case $template in
        php/*) echo "composer" ;;
        node/*) echo "npm" ;;
        python/*) echo "pip" ;;
        go/*) echo "go" ;;
        dotnet/*) echo "dotnet" ;;
        *) echo "unknown" ;;
    esac
}

# Copy template to apps directory
copy_template() {
    local template=$1
    local app_name=$2
    local template_path="${TEMPLATES_DIR}/${template}"
    local app_path="${APPS_DIR}/${app_name}"

    log_step "Copying template: $template -> $app_name"

    # Create app directory
    mkdir -p "$app_path"

    # Copy template files
    if [[ -d "$template_path" ]]; then
        cp -r "${template_path}/." "${app_path}/"
        log_success "Template copied successfully"
    else
        log_error "Template directory not found: $template_path"
        return 1
    fi

    return 0
}

# Customize template files
customize_template() {
    local app_name=$1
    local app_path="${APPS_DIR}/${app_name}"
    local domain="${app_name}.test"

    log_step "Customizing template files"

    # Replace placeholders in package.json, composer.json, etc.
    find "$app_path" -type f \( -name "package.json" -o -name "composer.json" -o -name "README.md" \) -exec sed -i \
        -e "s/devarch-app/${app_name}/g" \
        -e "s/devarch-nextjs-app/${app_name}/g" \
        -e "s/devarch-express-app/${app_name}/g" \
        -e "s/DevArch App/${app_name}/g" \
        -e "s/your-app/${app_name}/g" \
        -e "s/your-app\.test/${domain}/g" \
        {} \; 2>/dev/null || true

    # Create .env from .env.example if it exists
    if [[ -f "${app_path}/.env.example" ]]; then
        cp "${app_path}/.env.example" "${app_path}/.env"
        log_success "Created .env from .env.example"
    fi

    log_success "Template customized successfully"
}

# Create public directory if it doesn't exist
ensure_public_directory() {
    local app_name=$1
    local app_path="${APPS_DIR}/${app_name}"

    if [[ ! -d "${app_path}/public" ]]; then
        log_warning "public/ directory not found, creating it"
        mkdir -p "${app_path}/public"

        # Create .gitkeep
        cat > "${app_path}/public/.gitkeep" <<EOF
# This directory is the web server document root
# Build outputs will be placed here
EOF
    fi
}

# Display post-installation instructions
show_next_steps() {
    local app_name=$1
    local template=$2
    local port=$3
    local runtime=$(get_runtime "$template")
    local pkg_mgr=$(get_package_manager "$template")
    local domain="${app_name}.test"

    echo ""
    log_success "Application created successfully!"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Next Steps:${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "1. Navigate to your app:"
    echo -e "   ${YELLOW}cd apps/${app_name}${NC}"
    echo ""
    echo "2. Install dependencies:"
    case $pkg_mgr in
        npm)
            echo -e "   ${YELLOW}npm install${NC}"
            ;;
        composer)
            echo -e "   ${YELLOW}composer install${NC}"
            ;;
        pip)
            echo -e "   ${YELLOW}pip install -r requirements.txt${NC}"
            ;;
        go)
            echo -e "   ${YELLOW}go mod download${NC}"
            ;;
        dotnet)
            echo -e "   ${YELLOW}dotnet restore${NC}"
            ;;
    esac
    echo ""
    echo "3. Configure environment variables:"
    echo -e "   ${YELLOW}# Edit .env file${NC}"
    echo ""
    echo "4. Build for production (if needed):"
    case $pkg_mgr in
        npm)
            echo -e "   ${YELLOW}npm run build${NC}"
            echo -e "   ${GREEN}# Builds to public/ directory${NC}"
            ;;
        *)
            echo -e "   ${GREEN}# See template README for build instructions${NC}"
            ;;
    esac
    echo ""
    echo "5. Start backend runtime:"
    echo -e "   ${YELLOW}./scripts/service-manager.sh start backend${NC}"
    echo ""
    echo "6. Configure Nginx Proxy Manager:"
    echo -e "   - Access: ${YELLOW}http://localhost:81${NC}"
    echo -e "   - Add proxy host: ${YELLOW}${domain}${NC}"
    echo -e "   - Forward to: ${YELLOW}http://${runtime}:${port}${NC}"
    echo -e "   - Enable SSL certificate"
    echo ""
    echo "7. Access your application:"
    echo -e "   - Development: ${YELLOW}http://localhost:${port}${NC}"
    echo -e "   - Production: ${YELLOW}https://${domain}${NC}"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "For more information:"
    echo -e "  ${YELLOW}cat apps/${app_name}/README.md${NC}"
    echo ""
}

# Verify template structure
verify_template_structure() {
    local app_name=$1
    local app_path="${APPS_DIR}/${app_name}"

    log_step "Verifying template structure"

    local issues=0

    # Check for public/ directory
    if [[ ! -d "${app_path}/public" ]]; then
        log_error "Missing public/ directory (required for web serving)"
        ((issues++))
    else
        log_success "public/ directory exists"
    fi

    # Check for README
    if [[ ! -f "${app_path}/README.md" ]]; then
        log_warning "Missing README.md"
    else
        log_success "README.md exists"
    fi

    # Check for .gitignore
    if [[ ! -f "${app_path}/.gitignore" ]]; then
        log_warning "Missing .gitignore"
    else
        log_success ".gitignore exists"
    fi

    if [[ $issues -gt 0 ]]; then
        log_error "Template structure verification failed ($issues issues)"
        return 1
    fi

    log_success "Template structure verified"
    return 0
}

# Export functions for use in other scripts
export -f list_templates
export -f get_template_by_number
export -f validate_template
export -f validate_app_name
export -f get_default_port
export -f get_runtime
export -f get_package_manager
export -f copy_template
export -f customize_template
export -f ensure_public_directory
export -f show_next_steps
export -f verify_template_structure
export -f log_info
export -f log_success
export -f log_warning
export -f log_error
export -f log_step
