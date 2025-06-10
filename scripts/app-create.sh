#!/bin/bash
# =================================================================
# DEVARCH APPLICATION CREATION HELPER
# =================================================================
# This script helps create new applications in the DevArch environment
# with templates for different frameworks and automatic setup.

set -euo pipefail

# =================================================================
# CONFIGURATION AND VARIABLES
# =================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
APPS_DIR="$PROJECT_ROOT/apps"
TEMPLATES_DIR="$PROJECT_ROOT/templates"

# Load environment variables
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
fi

# Script options
VERBOSE=false
FORCE=false
FRAMEWORK=""
APP_NAME=""
TEMPLATE=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Available frameworks
FRAMEWORKS=(
    "laravel"
    "wordpress"
    "nextjs"
    "django"
    "flask"
    "fastapi"
    "vue"
    "react"
    "static"
    "php"
    "node"
    "python"
)

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

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Validate app name
validate_app_name() {
    local name="$1"
    
    # Check if name is provided
    if [[ -z "$name" ]]; then
        error "Application name is required"
        return 1
    fi
    
    # Check if name contains only valid characters
    if [[ ! "$name" =~ ^[a-zA-Z0-9][a-zA-Z0-9_-]*$ ]]; then
        error "Invalid application name. Use only letters, numbers, hyphens, and underscores. Must start with letter or number."
        return 1
    fi
    
    # Check if directory already exists
    if [[ -d "$APPS_DIR/$name" && "$FORCE" != "true" ]]; then
        error "Application '$name' already exists. Use --force to overwrite."
        return 1
    fi
    
    return 0
}

# Check if framework is supported
validate_framework() {
    local framework="$1"
    
    if [[ -z "$framework" ]]; then
        return 0  # Framework is optional
    fi
    
    for valid_framework in "${FRAMEWORKS[@]}"; do
        if [[ "$framework" == "$valid_framework" ]]; then
            return 0
        fi
    done
    
    error "Unsupported framework: $framework"
    error "Supported frameworks: ${FRAMEWORKS[*]}"
    return 1
}

# =================================================================
# TEMPLATE CREATION FUNCTIONS
# =================================================================
create_laravel_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating Laravel application: $app_name"
    
    # Use Laravel installer
    info "Installing Laravel via Composer..."
    docker run --rm \
        -v "$APPS_DIR":/apps \
        -w /apps \
        composer:latest \
        composer create-project laravel/laravel "$app_name" --prefer-dist
    
    # Set proper permissions
    chmod -R 755 "$app_dir"
    chmod -R 777 "$app_dir/storage" "$app_dir/bootstrap/cache"
    
    # Create environment file
    cat > "$app_dir/.env" << EOF
APP_NAME=$app_name
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=https://${app_name}.test

LOG_CHANNEL=stack
LOG_DEPRECATIONS_CHANNEL=null
LOG_LEVEL=debug

DB_CONNECTION=mysql
DB_HOST=mariadb
DB_PORT=3306
DB_DATABASE=$DEFAULT_DATABASE
DB_USERNAME=$DEFAULT_DATABASE_USER
DB_PASSWORD=$DEFAULT_DATABASE_PASSWORD

BROADCAST_DRIVER=log
CACHE_DRIVER=redis
FILESYSTEM_DISK=local
QUEUE_CONNECTION=redis
SESSION_DRIVER=redis
SESSION_LIFETIME=120

MEMCACHED_HOST=127.0.0.1

REDIS_HOST=redis
REDIS_PASSWORD=$DEFAULT_DATABASE_PASSWORD
REDIS_PORT=6379

MAIL_MAILER=smtp
MAIL_HOST=mailpit
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS="noreply@${app_name}.test"
MAIL_FROM_NAME="\${APP_NAME}"

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=us-east-1
AWS_BUCKET=
AWS_USE_PATH_STYLE_ENDPOINT=false

PUSHER_APP_ID=
PUSHER_APP_KEY=
PUSHER_APP_SECRET=
PUSHER_HOST=
PUSHER_PORT=443
PUSHER_SCHEME=https
PUSHER_APP_CLUSTER=mt1

VITE_PUSHER_APP_KEY="\${PUSHER_APP_KEY}"
VITE_PUSHER_HOST="\${PUSHER_HOST}"
VITE_PUSHER_PORT="\${PUSHER_PORT}"
VITE_PUSHER_SCHEME="\${PUSHER_SCHEME}"
VITE_PUSHER_APP_CLUSTER="\${PUSHER_APP_CLUSTER}"
EOF
    
    # Generate app key
    docker run --rm \
        -v "$app_dir":/var/www/html \
        -w /var/www/html \
        php:8.3-cli \
        php artisan key:generate
    
    success "Laravel application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
}

create_wordpress_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating WordPress application: $app_name"
    
    # Download WordPress
    info "Downloading WordPress..."
    mkdir -p "$app_dir"
    cd "$app_dir"
    
    curl -O https://wordpress.org/latest.tar.gz
    tar -xzf latest.tar.gz --strip-components=1
    rm latest.tar.gz
    
    # Set proper permissions
    chmod -R 755 "$app_dir"
    chmod -R 777 "$app_dir/wp-content"
    
    # Create wp-config.php
    cat > "$app_dir/wp-config.php" << EOF
<?php
// WordPress Configuration for DevArch

// Database settings
define( 'DB_NAME', 'wordpress' );
define( 'DB_USER', '$DEFAULT_DATABASE_USER' );
define( 'DB_PASSWORD', '$DEFAULT_DATABASE_PASSWORD' );
define( 'DB_HOST', 'mariadb' );
define( 'DB_CHARSET', 'utf8mb4' );
define( 'DB_COLLATE', '' );

// Authentication keys and salts
define( 'AUTH_KEY',         'devarch-auth-key-change-this' );
define( 'SECURE_AUTH_KEY',  'devarch-secure-auth-key-change-this' );
define( 'LOGGED_IN_KEY',    'devarch-logged-in-key-change-this' );
define( 'NONCE_KEY',        'devarch-nonce-key-change-this' );
define( 'AUTH_SALT',        'devarch-auth-salt-change-this' );
define( 'SECURE_AUTH_SALT', 'devarch-secure-auth-salt-change-this' );
define( 'LOGGED_IN_SALT',   'devarch-logged-in-salt-change-this' );
define( 'NONCE_SALT',       'devarch-nonce-salt-change-this' );

// WordPress table prefix
\$table_prefix = 'wp_';

// WordPress debugging
define( 'WP_DEBUG', true );
define( 'WP_DEBUG_LOG', true );
define( 'WP_DEBUG_DISPLAY', false );

// WordPress URLs
define( 'WP_HOME', 'https://${app_name}.test' );
define( 'WP_SITEURL', 'https://${app_name}.test' );

// SSL settings
define( 'FORCE_SSL_ADMIN', true );
if ( isset( \$_SERVER['HTTP_X_FORWARDED_PROTO'] ) && \$_SERVER['HTTP_X_FORWARDED_PROTO'] === 'https' ) {
    \$_SERVER['HTTPS'] = 'on';
}

// File permissions
define( 'FS_METHOD', 'direct' );

// Memory limit
define( 'WP_MEMORY_LIMIT', '256M' );

// Absolute path to the WordPress directory
if ( ! defined( 'ABSPATH' ) ) {
    define( 'ABSPATH', __DIR__ . '/' );
}

// Sets up WordPress vars and included files
require_once ABSPATH . 'wp-settings.php';
EOF
    
    success "WordPress application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
    warn "Don't forget to create the 'wordpress' database before setup"
}

create_nextjs_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating Next.js application: $app_name"
    
    # Create Next.js app
    info "Creating Next.js app with create-next-app..."
    cd "$APPS_DIR"
    
    npx create-next-app@latest "$app_name" \
        --typescript \
        --tailwind \
        --eslint \
        --app \
        --src-dir \
        --import-alias "@/*"
    
    # Create .env.local
    cat > "$app_dir/.env.local" << EOF
# DevArch Environment Configuration
NEXTAUTH_URL=https://${app_name}.test
NEXTAUTH_SECRET=$JWT_SECRET

# Database URLs
DATABASE_URL=mysql://$DEFAULT_DATABASE_USER:$DEFAULT_DATABASE_PASSWORD@mariadb:3306/$DEFAULT_DATABASE
POSTGRES_URL=postgresql://$DEFAULT_DATABASE_USER:$DEFAULT_DATABASE_PASSWORD@postgres:5432/$DEFAULT_DATABASE

# Redis URL
REDIS_URL=redis://:$DEFAULT_DATABASE_PASSWORD@redis:6379

# Mail configuration
SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_FROM=noreply@${app_name}.test
EOF
    
    # Update package.json for DevArch
    cd "$app_dir"
    npm pkg set scripts.dev="next dev --hostname 0.0.0.0 --port 3000"
    
    success "Next.js application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
}

create_django_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating Django application: $app_name"
    
    # Create Django project
    info "Creating Django project..."
    mkdir -p "$app_dir"
    cd "$app_dir"
    
    python -m venv venv
    source venv/bin/activate
    pip install django psycopg2-binary python-decouple
    
    django-admin startproject "$app_name" .
    
    # Create .env file
    cat > "$app_dir/.env" << EOF
# DevArch Django Configuration
DEBUG=True
SECRET_KEY=your-secret-key-change-this
ALLOWED_HOSTS=localhost,127.0.0.1,${app_name}.test

# Database
DATABASE_URL=postgresql://$DEFAULT_DATABASE_USER:$DEFAULT_DATABASE_PASSWORD@postgres:5432/$DEFAULT_DATABASE

# Redis
REDIS_URL=redis://:$DEFAULT_DATABASE_PASSWORD@redis:6379

# Email
EMAIL_HOST=mailpit
EMAIL_PORT=1025
EMAIL_USE_TLS=False
EMAIL_USE_SSL=False
DEFAULT_FROM_EMAIL=noreply@${app_name}.test
EOF
    
    # Update settings.py
    cat > "$app_dir/${app_name}/settings.py" << EOF
import os
from decouple import config
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

SECRET_KEY = config('SECRET_KEY', default='your-secret-key-change-this')
DEBUG = config('DEBUG', default=True, cast=bool)
ALLOWED_HOSTS = config('ALLOWED_HOSTS', default='localhost,127.0.0.1').split(',')

INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
]

ROOT_URLCONF = '${app_name}.urls'

TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [BASE_DIR / 'templates'],
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
            ],
        },
    },
]

WSGI_APPLICATION = '${app_name}.wsgi.application'

# Database
import dj_database_url
DATABASES = {
    'default': dj_database_url.parse(config('DATABASE_URL'))
}

AUTH_PASSWORD_VALIDATORS = [
    {'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator'},
    {'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator'},
    {'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator'},
    {'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator'},
]

LANGUAGE_CODE = 'en-us'
TIME_ZONE = 'America/Toronto'
USE_I18N = True
USE_TZ = True

STATIC_URL = '/static/'
STATIC_ROOT = BASE_DIR / 'staticfiles'

MEDIA_URL = '/media/'
MEDIA_ROOT = BASE_DIR / 'media'

DEFAULT_AUTO_FIELD = 'django.db.models.BigAutoField'

# Email configuration
EMAIL_HOST = config('EMAIL_HOST', default='mailpit')
EMAIL_PORT = config('EMAIL_PORT', default=1025, cast=int)
EMAIL_USE_TLS = config('EMAIL_USE_TLS', default=False, cast=bool)
EMAIL_USE_SSL = config('EMAIL_USE_SSL', default=False, cast=bool)
DEFAULT_FROM_EMAIL = config('DEFAULT_FROM_EMAIL', default='noreply@${app_name}.test')

# Security settings for development
if DEBUG:
    CORS_ALLOW_ALL_ORIGINS = True
    CSRF_TRUSTED_ORIGINS = ['https://${app_name}.test', 'http://localhost:8000']
EOF

    # Create requirements.txt
    cat > "$app_dir/requirements.txt" << EOF
Django>=4.2,<5.0
psycopg2-binary>=2.9.0
python-decouple>=3.8
dj-database-url>=2.0.0
django-cors-headers>=4.0.0
pillow>=10.0.0
EOF

    # Create manage.py wrapper script
    cat > "$app_dir/run-dev.sh" << 'EOF'
#!/bin/bash
source venv/bin/activate
python manage.py migrate
python manage.py collectstatic --noinput
python manage.py runserver 0.0.0.0:8000
EOF
    chmod +x "$app_dir/run-dev.sh"
    
    deactivate
    
    success "Django application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
    info "To start development: cd $app_dir && ./run-dev.sh"
}

create_static_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating static HTML application: $app_name"
    
    mkdir -p "$app_dir/public"
    
    # Create index.html
    cat > "$app_dir/public/index.html" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$app_name - DevArch Application</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container {
            text-align: center;
            max-width: 600px;
            padding: 2rem;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
        h1 { font-size: 3rem; margin-bottom: 1rem; }
        p { font-size: 1.2rem; margin-bottom: 2rem; opacity: 0.9; }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-top: 2rem;
        }
        .feature {
            padding: 1rem;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 10px;
        }
        .code {
            background: rgba(0, 0, 0, 0.2);
            padding: 1rem;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            margin: 1rem 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 $app_name</h1>
        <p>Your new DevArch application is ready!</p>
        
        <div class="code">
            https://${app_name}.test
        </div>
        
        <div class="features">
            <div class="feature">
                <h3>📱 Responsive</h3>
                <p>Mobile-first design</p>
            </div>
            <div class="feature">
                <h3>⚡ Fast</h3>
                <p>Optimized performance</p>
            </div>
            <div class="feature">
                <h3>🔒 Secure</h3>
                <p>HTTPS by default</p>
            </div>
        </div>
        
        <p style="margin-top: 2rem; font-size: 1rem; opacity: 0.7;">
            Edit files in: <code>apps/$app_name/public/</code>
        </p>
    </div>
</body>
</html>
EOF

    # Create CSS file
    mkdir -p "$app_dir/public/css"
    cat > "$app_dir/public/css/style.css" << EOF
/* DevArch Application Styles */
:root {
    --primary-color: #667eea;
    --secondary-color: #764ba2;
    --accent-color: #f093fb;
    --text-color: #333;
    --bg-color: #f8f9fa;
}

/* Add your custom styles here */
EOF

    # Create JavaScript file
    mkdir -p "$app_dir/public/js"
    cat > "$app_dir/public/js/app.js" << EOF
// DevArch Application JavaScript
console.log('Welcome to $app_name - DevArch Application!');

// Add your custom JavaScript here
document.addEventListener('DOMContentLoaded', function() {
    console.log('Application loaded successfully');
});
EOF

    success "Static application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
}

create_react_app() {
    local app_name="$1"
    local app_dir="$APPS_DIR/$app_name"
    
    log "Creating React application: $app_name"
    
    cd "$APPS_DIR"
    npx create-react-app "$app_name" --template typescript
    
    cd "$app_dir"
    
    # Install additional DevArch-friendly packages
    npm install --save-dev @types/node
    
    # Create .env file
    cat > "$app_dir/.env" << EOF
# DevArch React Configuration
REACT_APP_API_URL=https://${app_name}.test/api
REACT_APP_WS_URL=wss://${app_name}.test/ws
GENERATE_SOURCEMAP=true
BROWSER=none
EOF

    # Update package.json
    npm pkg set scripts.start="react-scripts start"
    npm pkg set homepage="https://${app_name}.test"
    
    success "React application '$app_name' created successfully"
    info "Access your app at: https://${app_name}.test"
}

# =================================================================
# MAIN FUNCTIONS
# =================================================================
show_framework_menu() {
    echo -e "${CYAN}Available Frameworks:${NC}"
    echo
    printf "%-15s %s\n" "Framework" "Description"
    echo "─────────────────────────────────────────────────"
    printf "%-15s %s\n" "laravel" "PHP framework for web applications"
    printf "%-15s %s\n" "wordpress" "Popular CMS platform"
    printf "%-15s %s\n" "nextjs" "React framework with SSR"
    printf "%-15s %s\n" "django" "Python web framework"
    printf "%-15s %s\n" "flask" "Lightweight Python framework"
    printf "%-15s %s\n" "fastapi" "Modern Python API framework"
    printf "%-15s %s\n" "vue" "Progressive JavaScript framework"
    printf "%-15s %s\n" "react" "JavaScript library for UIs"
    printf "%-15s %s\n" "static" "Static HTML/CSS/JS site"
    printf "%-15s %s\n" "php" "Basic PHP application"
    printf "%-15s %s\n" "node" "Node.js application"
    printf "%-15s %s\n" "python" "Python application"
    echo
}

create_application() {
    local app_name="$1"
    local framework="$2"
    
    # Validate inputs
    validate_app_name "$app_name" || exit 1
    validate_framework "$framework" || exit 1
    
    # Create app directory if force is enabled
    if [[ "$FORCE" == "true" && -d "$APPS_DIR/$app_name" ]]; then
        warn "Removing existing application directory"
        rm -rf "$APPS_DIR/$app_name"
    fi
    
    # Create application based on framework
    case "$framework" in
        laravel)
            create_laravel_app "$app_name"
            ;;
        wordpress)
            create_wordpress_app "$app_name"
            ;;
        nextjs)
            create_nextjs_app "$app_name"
            ;;
        django)
            create_django_app "$app_name"
            ;;
        react)
            create_react_app "$app_name"
            ;;
        static|"")
            create_static_app "$app_name"
            ;;
        *)
            error "Framework '$framework' not implemented yet"
            exit 1
            ;;
    esac
    
    # Create .devarch file for framework detection
    cat > "$APPS_DIR/$app_name/.devarch" << EOF
{
    "name": "$app_name",
    "framework": "$framework",
    "created": "$(date -Iseconds)",
    "version": "1.0.0",
    "devarch_version": "1.0.0"
}
EOF
    
    echo
    success "Application '$app_name' created successfully!"
    echo
    info "Next steps:"
    echo "  1. Access your app at: https://${app_name}.test"
    echo "  2. Edit files in: $APPS_DIR/$app_name"
    echo "  3. Check logs: ./scripts/manage.sh logs php-runtime"
    echo
}

show_usage() {
    echo -e "${CYAN}DevArch Application Creator${NC}"
    echo
    echo "Usage: $0 <app-name> [framework] [options]"
    echo
    echo -e "${YELLOW}Arguments:${NC}"
    echo "  app-name               Name of the application (required)"
    echo "  framework              Framework to use (optional, defaults to static)"
    echo
    echo -e "${YELLOW}Options:${NC}"
    echo "  -f, --force           Overwrite existing application"
    echo "  -v, --verbose         Enable verbose output"
    echo "  -l, --list            List available frameworks"
    echo "  -h, --help            Show this help message"
    echo
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 my-blog wordpress          # Create WordPress blog"
    echo "  $0 my-api laravel             # Create Laravel API"
    echo "  $0 my-site nextjs             # Create Next.js site"
    echo "  $0 my-page                    # Create static HTML site"
    echo
}

list_existing_apps() {
    echo -e "${CYAN}Existing Applications:${NC}"
    echo
    
    if [[ ! -d "$APPS_DIR" || -z "$(ls -A "$APPS_DIR" 2>/dev/null)" ]]; then
        echo "No applications found."
        return 0
    fi
    
    printf "%-20s %-15s %-20s %s\n" "Name" "Framework" "Created" "URL"
    echo "─────────────────────────────────────────────────────────────────────────"
    
    for app_dir in "$APPS_DIR"/*; do
        if [[ -d "$app_dir" ]]; then
            local app_name framework created url
            app_name=$(basename "$app_dir")
            
            if [[ -f "$app_dir/.devarch" ]]; then
                framework=$(jq -r '.framework // "unknown"' "$app_dir/.devarch" 2>/dev/null || echo "unknown")
                created=$(jq -r '.created // "unknown"' "$app_dir/.devarch" 2>/dev/null | cut -d'T' -f1 || echo "unknown")
            else
                framework="unknown"
                created=$(stat -c %y "$app_dir" 2>/dev/null | cut -d' ' -f1 || echo "unknown")
            fi
            
            url="https://${app_name}.test"
            
            printf "%-20s %-15s %-20s %s\n" "$app_name" "$framework" "$created" "$url"
        fi
    done
}

# =================================================================
# COMMAND PARSING
# =================================================================
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                FORCE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -l|--list)
                show_framework_menu
                echo
                list_existing_apps
                exit 0
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            -*)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [[ -z "$APP_NAME" ]]; then
                    APP_NAME="$1"
                elif [[ -z "$FRAMEWORK" ]]; then
                    FRAMEWORK="$1"
                else
                    error "Too many arguments"
                    show_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
}

# =================================================================
# MAIN FUNCTION
# =================================================================
main() {
    # Parse arguments
    parse_arguments "$@"
    
    # Check if app name is provided
    if [[ -z "$APP_NAME" ]]; then
        error "Application name is required"
        echo
        show_usage
        exit 1
    fi
    
    # Default to static framework if none specified
    if [[ -z "$FRAMEWORK" ]]; then
        FRAMEWORK="static"
    fi
    
    # Ensure apps directory exists
    mkdir -p "$APPS_DIR"
    
    # Create the application
    create_application "$APP_NAME" "$FRAMEWORK"
}

# =================================================================
# SCRIPT EXECUTION
# =================================================================
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',
]

MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth