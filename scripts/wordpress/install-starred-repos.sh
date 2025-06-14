#!/bin/zsh
WWW_DIR="/var/www"

set -a  # Automatically export all variables
source $WWW_DIR/.env
set +a  # Stop automatically exporting variables

# Define variables
SITE_TITLE="Playground"
DB_NAME="playground"

DB_USER="${MARIADB_USER}"
DB_PASS="${MARIADB_PASSWORD}"
DB_HOST="${MARIADB_HOST}"
SITE_DIR="$WWW_DIR/html/$DB_NAME"
WP_DIR="$SITE_DIR/public"
SITE_URL="$DB_NAME.test"
ADMIN_USER="${ADMIN_USER}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"
ADMIN_EMAIL="${ADMIN_EMAIL}"

# GitHub repositories for plugins and themes
PLUGIN_REPOS=(
    "https://github.com/pronamic/gravityforms.git"
    "https://github.com/pronamic/advanced-custom-fields-pro.git"
    "https://github.com/pronamic/woocommerce-subscriptions.git"
    "https://github.com/pronamic/facetwp.git"
)

# Install plugins
cd "$WP_DIR/wp-content/plugins" || exit
for PLUGIN_REPO in "${PLUGIN_REPOS[@]}"; do
    PLUGIN_DIR=$(basename "$PLUGIN_REPO" .git)
    git clone "$PLUGIN_REPO" || { echo "Failed to clone plugin $PLUGIN_REPO"; }
    wp plugin activate "$PLUGIN_DIR" --path="$WP_DIR" --allow-root || { echo "Failed to activate plugin $PLUGIN_DIR"; }
done

echo "Starred Repos Installed and Activated"