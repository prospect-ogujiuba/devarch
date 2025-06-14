#!/bin/zsh
WWW_DIR="/var/www"

set -a  # Automatically export all variables
source $WWW_DIR/.env
set +a  # Stop automatically exporting variables

# Define variables
SITE_TITLE="WP Clean"
DB_NAME="wpclean"

DB_USER="${MARIADB_USER}"
DB_PASS="${MARIADB_PASSWORD}"
DB_HOST="${MARIADB_HOST}"
SITE_DIR="$WWW_DIR/html/$DB_NAME"
WP_DIR="$SITE_DIR/public"
SITE_URL="$DB_NAME.test"
ADMIN_USER="${ADMIN_USER}"
ADMIN_PASSWORD="${ADMIN_PASSWORD}"
ADMIN_EMAIL="${ADMIN_EMAIL}"

# Ensure parent directory exists
mkdir -p "$SITE_DIR"

# Download WordPress
wp core download --path="$WP_DIR" --allow-root || { echo "Failed to download WordPress"; }

# Create the wp-config.php file
wp config create --dbname="$DB_NAME" --dbuser="$DB_USER" --dbpass="$DB_PASS" --dbhost="$DB_HOST" --path="$WP_DIR" --allow-root || { echo "Failed to create wp-config.php"; }

# Set debugging
wp config set WP_DEBUG false --raw --path="$WP_DIR" --allow-root || { echo "Failed to set WP_DEBUG";}

# Create the database
wp db create --path="$WP_DIR" --allow-root || { echo "Failed to create database"; }

# Install WordPress
wp core install --url="$SITE_URL" --title="$SITE_TITLE" --admin_user="$ADMIN_USER" --admin_password="$ADMIN_PASSWORD" --admin_email="$ADMIN_EMAIL" --path="$WP_DIR" --allow-root || { echo "Failed to install WordPress"; }

# Set month- and year-based folders
wp option update uploads_use_yearmonth_folders 0 --path="$WP_DIR" --allow-root || { echo "Failed to set upload folder";}

echo "WordPress installed successfully!"

# Delete default post and page
wp post delete 1 --force --path="$WP_DIR" --allow-root || echo "Post ID 1 not found."
wp post delete 2 --force --path="$WP_DIR" --allow-root || echo "Post ID 2 not found."
wp post delete 3 --force --path="$WP_DIR" --allow-root || echo "Post ID 3 not found."

# Delete default plugins
wp plugin delete akismet hello --path="$WP_DIR" --allow-root || echo "No default plugins found to delete."

# GitHub repositories for plugins and themes
PLUGIN_REPOS=(
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/typerocket-pro-v6.git"
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/all-in-one-wp-migration.git"
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/admin-site-enhancements-pro.git"
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/makermaker.git"
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/makerblocks.git"
)

THEME_REPOS=(
    "https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/makerstarter.git"
)

DEFAULT_THEMES=(
    "twentytwentythree"
    "twentytwentyfour"
    "twentytwentyfive"
)

# Install plugins
cd "$WP_DIR/wp-content/plugins" || exit
for PLUGIN_REPO in "${PLUGIN_REPOS[@]}"; do
    PLUGIN_DIR=$(basename "$PLUGIN_REPO" .git)
    git clone "$PLUGIN_REPO" || { echo "Failed to clone plugin $PLUGIN_REPO"; }
    wp plugin activate "$PLUGIN_DIR" --path="$WP_DIR" --allow-root || { echo "Failed to activate plugin $PLUGIN_DIR"; }
done

# Install themes
cd "$WP_DIR/wp-content/themes" || exit
for THEME_REPO in "${THEME_REPOS[@]}"; do
    THEME_DIR=$(basename "$THEME_REPO" .git)
    git clone "$THEME_REPO" || { echo "Failed to clone theme $THEME_REPO"; }
    wp theme activate "$THEME_DIR" --path="$WP_DIR" --allow-root || { echo "Failed to activate theme $THEME_DIR"; }
done

# Delete themes
cd "$WP_DIR/wp-content/themes" || exit
for THEME in "${DEFAULT_THEMES[@]}"; do
    wp theme delete "$THEME" --path="$WP_DIR" --allow-root || { echo "Failed to delete theme $THEME_DIR"; }
done

# Create necessary directories and set permissions for All in One WP Migration
AI1WM_BACKUPS_DIR="$WP_DIR/wp-content/ai1wm-backups"
mkdir -p "$AI1WM_BACKUPS_DIR" && echo "AI1WM Backup Folder Created at $AI1WM_BACKUPS_DIR" || { echo "Failed to create $AI1WM_BACKUPS_DIR"; }
chmod -R 777 "$AI1WM_BACKUPS_DIR" && echo "AI1WM Backup Folder Permissions set (777)" || { echo "Failed to set permissions on $AI1WM_BACKUPS_DIR"; }

AI1WM_STORAGE_DIR="$WP_DIR/wp-content/plugins/all-in-one-wp-migration/storage"
mkdir -p "$AI1WM_STORAGE_DIR" && echo "AI1WM Storage Folder Created at $AI1WM_STORAGE_DIR" || { echo "Failed to create $AI1WM_STORAGE_DIR"; exit 1; }
chmod -R 777 "$AI1WM_STORAGE_DIR" && echo "AI1WM Storage Folder Permissions set (777)" || { echo "Failed to set permissions on $AI1WM_STORAGE_DIR"; exit 1; }

# Uploads Directory
UPLOADS_DIR="$WP_DIR/wp-content/uploads"
mkdir -p "$UPLOADS_DIR" && echo "Uploads Folder Created at $UPLOADS_DIR" || { echo "Failed to create $UPLOADS_DIR"; exit 1; }
chmod -R 777 "$UPLOADS_DIR" && echo "Uploads Folder Permissions set (777)" || { echo "Failed to set permissions on $UPLOADS_DIR"; exit 1; }

# Themes Directory
THEMES_DIR="$WP_DIR/wp-content/themes"
chmod -R 777 "$THEMES_DIR" && echo "Themes Folder Permissions set (777)" || { echo "Failed to set permissions on $THEMES_DIR"; exit 1; }

# Plugins Directory
PLUGINS_DIR="$WP_DIR/wp-content/plugins"
chmod -R 777 "$PLUGINS_DIR" && echo "Plugins Folder Permissions set (777)" || { echo "Failed to set permissions on $PLUGINS_DIR"; exit 1; }

# Copy Makermaker Galaxy Files to WordPress Root Directory
GALAXY_DIR="$WP_DIR/wp-content/plugins/makermaker/galaxy"
GALAXY_FILE="$GALAXY_DIR/galaxy_makermaker"
GALAXY_CONFIG="$GALAXY_DIR/galaxy-makermaker-config.php"
cp "$GALAXY_FILE" "$WP_DIR"
sed "s/\$sitename = 'playground'/\$sitename = '$DB_NAME'/" "$GALAXY_CONFIG" > "$GALAXY_CONFIG.tmp"
mv "$GALAXY_CONFIG.tmp" "$GALAXY_CONFIG"
cp "$GALAXY_CONFIG" "$WP_DIR"

# Set permalink structure
wp rewrite structure '%postname%' --path="$WP_DIR" --allow-root || { echo "Failed to set permalink structure"; }

# Send installation summary email
EMAIL_SUBJECT="WordPress Installation Summary"
EMAIL_BODY="$SITE_TITLE installed successfully at $SITE_DIR and accessible at http://$SITE_URL"

# Send email using sendmail
wp eval "wp_mail('$ADMIN_EMAIL', '$EMAIL_SUBJECT', '$EMAIL_BODY');" --path="$WP_DIR" --allow-root || { echo "Failed to send email"; }

# Send a sample test email
TEST_EMAIL_SUBJECT="Test Email from WordPress Setup Script"
TEST_EMAIL_BODY="This is a test email to verify that the mail configuration is working correctly.\n\nThanks,\nThe Setup Script"

wp eval "wp_mail('$ADMIN_EMAIL', '$TEST_EMAIL_SUBJECT', '$TEST_EMAIL_BODY');" --path="$WP_DIR" --allow-root || { echo "Failed to send test email"; }

echo "Installation summary email and test email have been sent to $ADMIN_EMAIL"

echo "$SITE_TITLE installed successfully at $SITE_DIR and accessible at http://$SITE_URL"