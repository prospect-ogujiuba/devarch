#!/bin/zsh
# =============================================================================
# comprehensive-fix.sh - Fix all script issues
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🔧 Applying comprehensive fixes..."

# 1. Fix config.sh export statements
echo "📝 Fixing config.sh export statements..."

# Create backup
cp "$PROJECT_ROOT/scripts/config.sh" "$PROJECT_ROOT/scripts/config.sh.backup.$(date +%Y%m%d_%H%M%S)"

# Remove invalid export lines (around lines 284-291)
sed -i '/^export -f.*log execute_command.*$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f.*validate_compose_file.*$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f log$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f execute_command$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f detect_container_runtime$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f ensure_network$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f check_service_health$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f validate_compose_file$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f setup_environment$/d' "$PROJECT_ROOT/scripts/config.sh"
sed -i '/^export -f parse_common_args$/d' "$PROJECT_ROOT/scripts/config.sh"

# Add proper exports at the end (only if they don't exist)
if ! grep -q "# Export functions for use in other scripts" "$PROJECT_ROOT/scripts/config.sh"; then
    cat >> "$PROJECT_ROOT/scripts/config.sh" << 'EOF'

# Export functions for use in other scripts
export -f log
export -f execute_command
export -f detect_container_runtime
export -f ensure_network
export -f check_service_health
export -f validate_compose_file
export -f setup_environment
export -f parse_common_args
EOF
fi

echo "✅ config.sh exports fixed"

# 2. Fix install.sh determine_install_categories function
echo "📝 Fixing install.sh category determination..."

# Create backup
cp "$PROJECT_ROOT/scripts/install.sh" "$PROJECT_ROOT/scripts/install.sh.backup.$(date +%Y%m%d_%H%M%S)"

# Fix the determine_install_categories function
cat > /tmp/determine_install_categories_fix.sh << 'EOF'
determine_install_categories() {
    local categories_to_install=()
    
    if [ ${#INSTALL_CATEGORIES[@]} -gt 0 ]; then
        # Use only specified categories
        categories_to_install=("${INSTALL_CATEGORIES[@]}")
        log "INFO" "Installing only specified categories: ${categories_to_install[*]}"
    else
        # Use all categories except excluded ones
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do
            if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then
                categories_to_install+=("$category")
            fi
        done
        
        if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then
            log "INFO" "Installing all categories except: ${EXCLUDE_CATEGORIES[*]}"
        else
            log "INFO" "Installing all categories"
        fi
    fi
    
    echo "${categories_to_install[@]}"
}
EOF

# Replace the function in install.sh
sed -i '/^determine_install_categories() {/,/^}/c\
determine_install_categories() {\
    local categories_to_install=()\
    \
    if [ ${#INSTALL_CATEGORIES[@]} -gt 0 ]; then\
        # Use only specified categories\
        categories_to_install=("${INSTALL_CATEGORIES[@]}")\
        log "INFO" "Installing only specified categories: ${categories_to_install[*]}"\
    else\
        # Use all categories except excluded ones\
        for category in "${SERVICE_STARTUP_ORDER[@]}"; do\
            if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then\
                categories_to_install+=("$category")\
            fi\
        done\
        \
        if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then\
            log "INFO" "Installing all categories except: ${EXCLUDE_CATEGORIES[*]}"\
        else\
            log "INFO" "Installing all categories"\
        fi\
    fi\
    \
    echo "${categories_to_install[@]}"\
}' "$PROJECT_ROOT/scripts/install.sh"

echo "✅ install.sh category determination fixed"

# 3. Fix install.sh install_category function to skip health checks in dry run
echo "📝 Fixing install.sh health checks in dry mode..."

# Replace the install_category function to skip health checks in dry run
sed -i '/# Health check for critical services/,/done/c\
    # Health check for critical services (skip in dry run)\
    if [[ "$category" == "database" ]] && [ "$DRY_RUN" = false ]; then\
        log "INFO" "  🔍 Performing health checks for database services..."\
        sleep 5  # Give databases time to initialize\
        \
        for file in "${files[@]}"; do\
            local service_name=$(basename "$file" .yml)\
            check_service_health "$service_name" "$CONTAINER_RUNTIME" "$SUDO_PREFIX" 15 3\
        done\
    elif [[ "$category" == "database" ]] && [ "$DRY_RUN" = true ]; then\
        log "INFO" "  [DRY RUN] Would perform health checks for database services"\
    fi' "$PROJECT_ROOT/scripts/install.sh"

echo "✅ install.sh health checks fixed"

# 4. Update SERVICE_CATEGORIES to match actual directory structure
echo "📝 Updating SERVICE_CATEGORIES in config.sh..."

# Check if keycloak is in auth/ directory and update accordingly
if [ -f "$PROJECT_ROOT/compose/auth/keycloak.yml" ]; then
    KEYCLOAK_PATH="auth/keycloak.yml"
    PROXY_FILES="proxy/nginx-proxy-manager.yml"
else
    KEYCLOAK_PATH="proxy/keycloak.yml"
    PROXY_FILES="proxy/nginx-proxy-manager.yml proxy/keycloak.yml"
fi

# Replace SERVICE_CATEGORIES with correct structure
sed -i '/^declare -A SERVICE_CATEGORIES=/,/^)/c\
declare -A SERVICE_CATEGORIES=(\
    ["database"]="database/postgres.yml database/mysql.yml database/mariadb.yml database/mongodb.yml database/redis.yml"\
    ["dbms"]="dbms/adminer.yml dbms/pgadmin.yml dbms/phpmyadmin.yml dbms/mongo-express.yml dbms/metabase.yml dbms/nocodb.yml"\
    ["backend"]="backend/php.yml backend/node.yml backend/python.yml backend/go.yml backend/dotnet.yml"\
    ["analytics"]="analytics/elasticsearch.yml analytics/kibana.yml analytics/logstash.yml analytics/grafana.yml analytics/prometheus.yml analytics/matomo.yml"\
    ["ai"]="ai/langflow.yml ai/n8n.yml"\
    ["mail"]="mail/mailpit.yml"\
    ["project"]="project/gitea.yml"\
    ["erp"]="erp/odoo.yml"\
    ["auth"]="'$KEYCLOAK_PATH'"\
    ["proxy"]="'$PROXY_FILES'"\
)' "$PROJECT_ROOT/scripts/config.sh"

# 5. Update SERVICE_STARTUP_ORDER to include auth if keycloak is in auth/
if [ -f "$PROJECT_ROOT/compose/auth/keycloak.yml" ]; then
    sed -i '/^SERVICE_STARTUP_ORDER=/,/)/c\
SERVICE_STARTUP_ORDER=(\
    "database"\
    "dbms"\
    "backend"\
    "analytics"\
    "ai"\
    "mail"\
    "project"\
    "erp"\
    "auth"\
    "proxy"\
)' "$PROJECT_ROOT/scripts/config.sh"
fi

echo "✅ SERVICE_CATEGORIES updated"

# 6. Clean up any broken array syntax
echo "📝 Cleaning up syntax issues..."

# Fix any potential array syntax issues
sed -i 's/Installing all categories /Installing all categories: /' "$PROJECT_ROOT/scripts/install.sh" 2>/dev/null || true

echo "✅ Syntax cleanup completed"

# 7. Verify the fixes
echo "🔍 Verifying fixes..."

echo "📁 Category structure:"
for category in database dbms backend analytics ai mail project erp auth proxy; do
    if [ -d "$PROJECT_ROOT/compose/$category" ]; then
        file_count=$(find "$PROJECT_ROOT/compose/$category" -name "*.yml" | wc -l)
        echo "  ✅ $category: $file_count files"
    else
        echo "  ⚠️  $category: directory not found"
    fi
done

echo ""
echo "🎉 Comprehensive fixes completed!"
echo ""
echo "📋 Summary of fixes:"
echo "  ✅ Fixed config.sh export statements"
echo "  ✅ Fixed install.sh category determination"
echo "  ✅ Fixed health checks in dry run mode"
echo "  ✅ Updated SERVICE_CATEGORIES paths"
echo "  ✅ Updated SERVICE_STARTUP_ORDER"
echo "  ✅ Cleaned up syntax issues"
echo ""
echo "💡 Test with: ./scripts/install.sh -sed -v"

# Clean up temp files
rm -f /tmp/determine_install_categories_fix.sh