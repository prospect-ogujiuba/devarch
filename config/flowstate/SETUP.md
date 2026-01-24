# FlowState Production Setup

## Prerequisites

- DevArch environment running on VPS
- PostgreSQL and Redis containers running
- Domain DNS pointing to VPS (flowstate.mxcro.com)

## Initial Setup

### 1. Create External Volume

```bash
podman volume create flowstate-storage
```

### 2. Configure Environment

```bash
cd /path/to/devarch/apps/flowstate
cp .env.production.example .env

# Edit .env with production values:
# - Generate APP_KEY: php artisan key:generate --show
# - Set DB_PASSWORD
# - Set REDIS_PASSWORD (if using)
# - Generate REVERB keys: php artisan reverb:generate-keys (or use random strings)
# - Set MAIL_* credentials
```

### 3. Update Client Environment

```bash
cd client
# Edit .env.production with your REVERB_APP_KEY
```

### 4. Create Database

```bash
podman exec -it postgres psql -U postgres -c "CREATE DATABASE flowstate;"
podman exec -it postgres psql -U postgres -c "CREATE USER flowstate WITH PASSWORD 'your-password';"
podman exec -it postgres psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE flowstate TO flowstate;"
```

### 5. Build and Start

```bash
cd /path/to/devarch

# Build the image
podman-compose -f compose/project/flowstate.yml build

# Start containers
podman-compose -f compose/project/flowstate.yml up -d

# Run migrations
podman exec -it flowstate-app php artisan migrate --force

# Create storage link
podman exec -it flowstate-app php artisan storage:link
```

### 6. Configure Nginx Proxy

Option A: Nginx Proxy Manager
- Add new proxy host
- Domain: flowstate.mxcro.com
- Forward to: flowstate-app:80
- Enable SSL (Let's Encrypt)
- Add custom location for WebSocket:
  - Location: /app
  - Forward to: flowstate-reverb:8080
  - Enable WebSocket support

Option B: Manual nginx config
- Copy `nginx-proxy.conf` to `/etc/nginx/sites-enabled/flowstate.conf`
- Run `certbot --nginx -d flowstate.mxcro.com`
- Reload nginx

## CI/CD Setup

### GitHub Secrets Required

Add these secrets to your GitHub repository:

- `VPS_HOST`: Your VPS IP or hostname
- `VPS_USER`: SSH username
- `VPS_SSH_KEY`: Private SSH key for deployment

### Update Deploy Script

Edit `/path/to/devarch/scripts/flowstate-deploy.sh`:
- Replace `/path/to/devarch` with actual path

## Useful Commands

```bash
# View logs
podman logs -f flowstate-app
podman logs -f flowstate-queue
podman logs -f flowstate-reverb

# Restart services
podman-compose -f compose/project/flowstate.yml restart

# Rebuild after code changes
podman-compose -f compose/project/flowstate.yml up -d --build

# Run artisan commands
podman exec -it flowstate-app php artisan <command>

# Clear caches
podman exec -it flowstate-app php artisan optimize:clear
```

## Troubleshooting

### Queue not processing
```bash
podman logs flowstate-queue
podman exec -it flowstate-queue php artisan queue:restart
```

### WebSocket not connecting
- Check Reverb logs: `podman logs flowstate-reverb`
- Verify REVERB_APP_KEY matches in both .env files
- Check nginx proxy passes /app to reverb container

### 502 Bad Gateway
- Check container is running: `podman ps`
- Check container logs: `podman logs flowstate-app`
- Verify network connectivity: `podman network inspect microservices-net`
