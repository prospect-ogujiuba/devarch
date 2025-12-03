# Quick Start: DevArch Application Templates

**Quick reference for creating new applications in DevArch**

## TL;DR

```bash
# Create a new app (interactive)
./scripts/create-app.sh

# Create a new app (non-interactive)
./scripts/create-app.sh --name my-app --template node/react-vite --port 8200

# List available templates
./scripts/create-app.sh --list
```

## Available Templates

### Node.js (Ports 8200-8299)

```bash
# React + Vite (SPA, hot reload, modern tooling)
./scripts/create-app.sh --name my-spa --template node/react-vite

# Next.js (SSR, static export, file-based routing)
./scripts/create-app.sh --name my-nextapp --template node/nextjs

# Express (API server, static file serving)
./scripts/create-app.sh --name my-api --template node/express
```

### PHP (Ports 8100-8199)

```bash
# Laravel (full-stack PHP framework)
./scripts/create-app.sh --name my-laravel --template php/laravel

# WordPress (CMS)
./scripts/create-app.sh --name my-blog --template php/wordpress

# Vanilla PHP (plain PHP)
./scripts/create-app.sh --name my-php --template php/vanilla
```

### Python (Ports 8300-8399)

```bash
# Django (full-stack Python framework)
./scripts/create-app.sh --name my-django --template python/django

# Flask (lightweight Python web framework)
./scripts/create-app.sh --name my-flask --template python/flask

# FastAPI (modern Python API framework)
./scripts/create-app.sh --name my-fastapi --template python/fastapi
```

### Go (Ports 8400-8499)

```bash
# Gin (high-performance HTTP framework)
./scripts/create-app.sh --name my-gin --template go/gin

# Echo (minimalist framework)
./scripts/create-app.sh --name my-echo --template go/echo
```

### .NET (Ports 8600-8699)

```bash
# ASP.NET Core (enterprise framework)
./scripts/create-app.sh --name my-dotnet --template dotnet/aspnet-core
```

## After Creating an App

### 1. Navigate to App

```bash
cd apps/my-app
```

### 2. Install Dependencies

```bash
# Node.js
npm install

# PHP
composer install

# Python
pip install -r requirements.txt

# Go
go mod download

# .NET
dotnet restore
```

### 3. Configure Environment

```bash
# Edit .env file (created from .env.example)
nano .env
```

### 4. Development Mode

```bash
# Node.js
npm run dev

# PHP (Laravel)
php artisan serve

# Python (Django)
python manage.py runserver 0.0.0.0:8300

# Python (Flask)
python app.py

# Go
go run main.go

# .NET
dotnet run
```

### 5. Build for Production

```bash
# Node.js (React, Next.js, Vue)
npm run build
# ✅ Outputs to public/

# Python (Django)
python manage.py collectstatic --noinput
# ✅ Outputs to public/static/

# Go
go build -o app main.go

# .NET
dotnet build
```

### 6. Verify public/ Directory

```bash
ls -la public/
# Should contain:
# - index.html (or index.php)
# - assets/ directory
# - Built files from your build process
```

### 7. Start Backend Runtime

```bash
# From DevArch root
./scripts/service-manager.sh start backend
```

### 8. Configure Nginx Proxy Manager

1. Access: http://localhost:81
2. Add Proxy Host:
   - **Domain**: `my-app.test`
   - **Forward Hostname**: `nodejs` (or `php`, `python`, etc.)
   - **Forward Port**: `8200` (your app's port)
3. Enable SSL certificate

### 9. Update Hosts File

```bash
sudo ./scripts/update-hosts.sh
```

### 10. Access Your App

- **Development**: http://localhost:{port}
- **Production**: https://my-app.test

## Port Ranges

| Runtime | Range | Default | Example |
|---------|-------|---------|---------|
| PHP | 8100-8199 | 8100 | Laravel: 8100 |
| Node.js | 8200-8299 | 8200 | React: 8200 |
| Python | 8300-8399 | 8300 | Django: 8300 |
| Go | 8400-8499 | 8400 | Gin: 8400 |
| .NET | 8600-8699 | 8600 | ASP.NET: 8600 |

## Common Commands

```bash
# List all templates
./scripts/create-app.sh --list

# Get help
./scripts/create-app.sh --help

# Check backend services
./scripts/service-manager.sh status backend

# View service logs
./scripts/service-manager.sh logs nodejs --follow

# Restart a service
./scripts/service-manager.sh restart nodejs
```

## Troubleshooting

### Build doesn't output to public/

Check your build configuration:
- **Vite**: `build.outDir: 'public'`
- **Next.js**: `distDir: 'public/.next'`
- **Webpack**: `output.path: 'public'`

### App not accessible at .test domain

1. Check Nginx Proxy Manager configuration
2. Verify backend runtime is running
3. Update hosts file: `sudo ./scripts/update-hosts.sh`
4. Clear browser cache

### Port already in use

Choose different port in same range:
```bash
# Check what's using the port
lsof -i :8200

# Use different port
./scripts/create-app.sh --name my-app --template node/react-vite --port 8201
```

### Assets not loading (404)

1. Verify `public/` directory exists and has files
2. Check build process completed successfully
3. Verify web server serves from `public/`

## Documentation

- **Complete Structure Guide**: `APP_STRUCTURE.md`
- **Template Details**: `TEMPLATES.md`
- **Migration Guide**: `MIGRATION_GUIDE.md`
- **Implementation Summary**: `IMPLEMENTATION_SUMMARY.md`

## Need Help?

1. Check template README: `templates/{category}/{framework}/README.md`
2. Review app structure: `APP_STRUCTURE.md`
3. Check DevArch docs: `CLAUDE.md`

## Examples

### Create React Dashboard

```bash
./scripts/create-app.sh --name dashboard --template node/react-vite --port 8200
cd apps/dashboard
npm install
npm run dev
# Access: http://localhost:5173 (dev) or http://localhost:8200 (prod)
```

### Create Express API

```bash
./scripts/create-app.sh --name api --template node/express --port 8201
cd apps/api
npm install
npm start
# Access: http://localhost:8201
```

### Create Django App

```bash
./scripts/create-app.sh --name admin --template python/django --port 8300
cd apps/admin
pip install -r requirements.txt
python manage.py migrate
python manage.py runserver 0.0.0.0:8300
# Access: http://localhost:8300
```

---

**Quick Start Complete!** 🚀

Your app is ready for development. Build outputs will go to `public/` and be served by the web server.
