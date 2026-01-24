#!/bin/bash
set -e

DEVARCH_DIR="/path/to/devarch"
APP_DIR="$DEVARCH_DIR/apps/flowstate"
COMPOSE_FILE="$DEVARCH_DIR/compose/project/flowstate.yml"

echo "=== FlowState Deployment ==="
echo "Started at: $(date)"

cd "$APP_DIR"

echo "1. Pulling latest code..."
git pull origin main

echo "2. Installing PHP dependencies..."
composer install --no-dev --optimize-autoloader --no-interaction

echo "3. Building frontend..."
cd client
npm ci --production=false
npm run build
cd ..

echo "4. Running migrations..."
php artisan migrate --force

echo "5. Clearing caches..."
php artisan config:cache
php artisan route:cache
php artisan view:cache
php artisan event:cache

echo "6. Restarting containers..."
cd "$DEVARCH_DIR"

# Gracefully stop queue worker (finish current job)
podman exec flowstate-queue php artisan queue:restart 2>/dev/null || true
sleep 5

# Rebuild and restart containers
podman-compose -f "$COMPOSE_FILE" build --no-cache
podman-compose -f "$COMPOSE_FILE" up -d

echo "7. Waiting for health checks..."
sleep 10

# Verify containers are running
if podman ps | grep -q flowstate-app; then
    echo "flowstate-app: RUNNING"
else
    echo "flowstate-app: FAILED"
    exit 1
fi

if podman ps | grep -q flowstate-queue; then
    echo "flowstate-queue: RUNNING"
else
    echo "flowstate-queue: FAILED"
    exit 1
fi

if podman ps | grep -q flowstate-reverb; then
    echo "flowstate-reverb: RUNNING"
else
    echo "flowstate-reverb: FAILED"
    exit 1
fi

echo "=== Deployment Complete ==="
echo "Finished at: $(date)"
