#!/bin/bash

echo "🔍 Keycloak Smart Entrypoint: Initializing..."

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

# Set default admin credentials if not provided
export KEYCLOAK_ADMIN="${KEYCLOAK_ADMIN:-admin}"
export KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-123456}"

# Database connection settings
export KC_DB="${KC_DB:-postgres}"
export KC_DB_URL_HOST="${KC_DB_URL_HOST:-postgres}"
export KC_DB_URL_PORT="${KC_DB_URL_PORT:-5432}"
export KC_DB_URL_DATABASE="${KC_DB_URL_DATABASE:-keycloak}"
export KC_DB_USERNAME="${KC_DB_USERNAME:-keycloak_user}"
export KC_DB_PASSWORD="${KC_DB_PASSWORD:-123456}"

# Hostname settings
export KC_HOSTNAME="${KC_HOSTNAME:-keycloak.test}"
export KC_HOSTNAME_STRICT="${KC_HOSTNAME_STRICT:-false}"
export KC_HOSTNAME_STRICT_HTTPS="${KC_HOSTNAME_STRICT_HTTPS:-false}"

# Proxy settings
export KC_PROXY="${KC_PROXY:-edge}"
export KC_PROXY_HEADERS="${KC_PROXY_HEADERS:-forwarded}"

# HTTP settings
export KC_HTTP_ENABLED="${KC_HTTP_ENABLED:-true}"
export KC_HTTP_HOST="${KC_HTTP_HOST:-0.0.0.0}"
export KC_HTTP_PORT="${KC_HTTP_PORT:-8080}"

# Health and metrics
export KC_HEALTH_ENABLED="${KC_HEALTH_ENABLED:-true}"
export KC_METRICS_ENABLED="${KC_METRICS_ENABLED:-true}"

# =============================================================================
# DATABASE WAIT FUNCTION
# =============================================================================

wait_for_database() {
    echo "🔄 Waiting for PostgreSQL database to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if pg_isready -h "$KC_DB_URL_HOST" -p "$KC_DB_URL_PORT" -U "$KC_DB_USERNAME" >/dev/null 2>&1; then
            echo "✅ Database is ready!"
            return 0
        fi
        
        echo "⏳ Database not ready yet (attempt $attempt/$max_attempts)..."
        sleep 2
        ((attempt++))
    done
    
    echo "❌ Database connection timeout"
    return 1
}

# =============================================================================
# KEYCLOAK SETUP FUNCTIONS
# =============================================================================

setup_admin_user() {
    echo "🔧 Setting up admin user..."
    
    # Create admin user if it doesn't exist
    if [[ -n "$KEYCLOAK_ADMIN" && -n "$KEYCLOAK_ADMIN_PASSWORD" ]]; then
        echo "👤 Admin user: $KEYCLOAK_ADMIN"
        echo "🔐 Admin password: [REDACTED]"
    else
        echo "⚠️  No admin credentials provided"
    fi
}

import_realm_configuration() {
    echo "🔧 Checking for realm import configuration..."
    
    if [[ -f "/opt/keycloak/data/import/realm-import.json" ]]; then
        echo "📋 Found realm import file, will import on startup"
        export KC_IMPORT="--import-realm"
    else
        echo "ℹ️  No realm import file found"
        export KC_IMPORT=""
    fi
}

setup_themes() {
    echo "🎨 Setting up custom themes..."
    
    if [[ -d "/opt/keycloak/themes/microservices" ]]; then
        echo "✅ Custom microservices theme found"
    else
        echo "ℹ️  No custom themes found, using defaults"
    fi
}

# =============================================================================
# DEVELOPMENT FEATURES
# =============================================================================

enable_dev_features() {
    if [[ "$KEYCLOAK_ENVIRONMENT" == "development" ]]; then
        echo "🛠️  Enabling development features..."
        export KC_FEATURES="token-exchange,admin-fine-grained-authz,recovery-codes,update-email,preview"
        export KC_LOG_LEVEL="DEBUG"
    else
        echo "🏭 Production mode - limited features enabled"
        export KC_FEATURES="token-exchange,admin-fine-grained-authz,recovery-codes,update-email"
        export KC_LOG_LEVEL="INFO"
    fi
}

# =============================================================================
# MAIN SETUP SEQUENCE
# =============================================================================

main_setup() {
    echo "🚀 Starting Keycloak setup sequence..."
    
    # Wait for database
    if ! wait_for_database; then
        echo "❌ Failed to connect to database, continuing anyway..."
    fi
    
    # Setup functions
    setup_admin_user
    import_realm_configuration
    setup_themes
    enable_dev_features
    
    echo "✅ Keycloak setup complete!"
}

# =============================================================================
# COMMAND HANDLING
# =============================================================================

# Run setup
main_setup

# Handle different commands
case "$1" in
    "start"|"start-dev")
        echo "🔄 Starting Keycloak server..."
        
        # Build configuration arguments
        KC_ARGS=(
            "--hostname=$KC_HOSTNAME"
            "--hostname-strict=$KC_HOSTNAME_STRICT"
            "--hostname-strict-https=$KC_HOSTNAME_STRICT_HTTPS"
            "--proxy=$KC_PROXY"
            "--proxy-headers=$KC_PROXY_HEADERS"
            "--http-enabled=$KC_HTTP_ENABLED"
            "--http-host=$KC_HTTP_HOST"
            "--http-port=$KC_HTTP_PORT"
            "--health-enabled=$KC_HEALTH_ENABLED"
            "--metrics-enabled=$KC_METRICS_ENABLED"
            "--features=$KC_FEATURES"
            "--log-level=$KC_LOG_LEVEL"
        )
        
        # Add import flag if needed
        if [[ -n "$KC_IMPORT" ]]; then
            KC_ARGS+=("$KC_IMPORT")
        fi
        
        # Development vs production start
        if [[ "$1" == "start-dev" ]]; then
            echo "🛠️  Starting in development mode..."
            exec /opt/keycloak/bin/kc.sh start-dev "${KC_ARGS[@]}"
        else
            echo "🏭 Starting in production mode..."
            exec /opt/keycloak/bin/kc.sh start --optimized "${KC_ARGS[@]}"
        fi
        ;;
        
    "build")
        echo "🔨 Building Keycloak..."
        exec /opt/keycloak/bin/kc.sh build
        ;;
        
    "export")
        echo "📤 Exporting Keycloak configuration..."
        exec /opt/keycloak/bin/kc.sh export --dir /opt/keycloak/data/export
        ;;
        
    "import")
        echo "📥 Importing Keycloak configuration..."
        exec /opt/keycloak/bin/kc.sh import --dir /opt/keycloak/data/import
        ;;
        
    *)
        echo "🔧 Running custom command: $*"
        exec /opt/keycloak/bin/kc.sh "$@"
        ;;
esac