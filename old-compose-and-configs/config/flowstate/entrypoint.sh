#!/bin/sh
set -e

echo "Ensuring Laravel directories exist..."
mkdir -p /var/www/html/storage/framework/{views,cache,sessions}
mkdir -p /var/www/html/storage/logs
mkdir -p /var/www/html/bootstrap/cache

echo "Fixing Laravel permissions..."
chown -R www-data:www-data /var/www/html/storage /var/www/html/bootstrap/cache 2>/dev/null || true
chmod -R 775 /var/www/html/storage /var/www/html/bootstrap/cache 2>/dev/null || true

echo "Starting supervisor..."
exec "$@"
