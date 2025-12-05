# Quick Start: Creating DevArch Applications

**Quick reference for creating new applications in DevArch using JetBrains IDEs**

## TL;DR

**Use JetBrains IDEs for project creation**:
- PHPStorm (Laravel, WordPress)
- WebStorm (React, Next.js, Vue, Express)
- PyCharm (Django, Flask, FastAPI)
- GoLand (Go applications)
- Rider (ASP.NET Core)

**WordPress only**: Use custom script for custom repos/templates
```bash
./scripts/wordpress/install-wordpress.sh
```

## Creating Apps with JetBrains IDEs

### WebStorm (Node.js - Ports 8200-8299)

**React + Vite**:
1. File → New Project → React
2. Location: `/home/fhcadmin/projects/devarch/apps/my-spa`
3. Template: Vite
4. See: `docs/jetbrains/webstorm-react.md`

**Next.js**:
1. File → New Project → Next.js
2. Location: `/home/fhcadmin/projects/devarch/apps/my-nextapp`
3. See: `docs/jetbrains/webstorm-nextjs.md`

**Express**:
1. File → New Project → Express App
2. Location: `/home/fhcadmin/projects/devarch/apps/my-api`
3. See: `docs/jetbrains/webstorm-express.md`

### PHPStorm (PHP - Ports 8100-8199)

**Laravel**:
1. File → New Project → Laravel
2. Location: `/home/fhcadmin/projects/devarch/apps/my-laravel`
3. See: `docs/jetbrains/phpstorm-laravel.md`

**WordPress** (uses custom script):
```bash
./scripts/wordpress/install-wordpress.sh my-blog \
  --preset clean \
  --title "My Blog"
```
See: `docs/jetbrains/phpstorm-wordpress.md`

### PyCharm (Python - Ports 8300-8399)

**Django**:
1. File → New Project → Django
2. Location: `/home/fhcadmin/projects/devarch/apps/my-django`
3. See: `docs/jetbrains/pycharm-django.md`

**Flask**:
1. File → New Project → Flask
2. Location: `/home/fhcadmin/projects/devarch/apps/my-flask`
3. See: `docs/jetbrains/pycharm-flask.md`

**FastAPI**:
1. File → New Project → FastAPI
2. Location: `/home/fhcadmin/projects/devarch/apps/my-fastapi`
3. See: `docs/jetbrains/pycharm-fastapi.md`

### GoLand (Go - Ports 8400-8499)

**Go Application**:
1. File → New Project → Go
2. Location: `/home/fhcadmin/projects/devarch/apps/my-goapp`
3. Add Gin/Echo manually via `go get`
4. See: `docs/jetbrains/goland-gin.md`

### Rider (.NET - Ports 8600-8699)

**ASP.NET Core**:
1. File → New Solution → ASP.NET Core Web Application
2. Location: `/home/fhcadmin/projects/devarch/apps/my-dotnet`
3. See: `docs/jetbrains/rider-aspnet.md`

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
# WordPress installation (only framework needing custom script)
./scripts/wordpress/install-wordpress.sh --help

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

- **Complete Structure Guide**: `docs/APP_STRUCTURE.md`
- **JetBrains IDE Guides**: `docs/jetbrains/`
- **DevArch Architecture**: `CLAUDE.md`
- **Service Management**: `docs/SERVICE_MANAGER.md`

## Need Help?

1. Check JetBrains guide: `docs/jetbrains/{ide}-{framework}.md`
2. Review app structure: `docs/APP_STRUCTURE.md`
3. Check DevArch docs: `CLAUDE.md`
4. WordPress: `scripts/wordpress/install-wordpress.sh --help`

## Examples

### Create React Dashboard (WebStorm)

1. WebStorm → File → New Project → React
2. Location: `/home/fhcadmin/projects/devarch/apps/dashboard`
3. Template: Vite
4. Navigate to project:
   ```bash
   cd apps/dashboard
   npm install
   npm run dev
   # Access: http://localhost:5173 (dev)
   ```

### Create Express API (WebStorm)

1. WebStorm → File → New Project → Express App
2. Location: `/home/fhcadmin/projects/devarch/apps/api`
3. Navigate to project:
   ```bash
   cd apps/api
   npm install
   npm start
   # Access: http://localhost:8201
   ```

### Create Django App (PyCharm)

1. PyCharm → File → New Project → Django
2. Location: `/home/fhcadmin/projects/devarch/apps/admin`
3. Navigate to project:
   ```bash
   cd apps/admin
   pip install -r requirements.txt
   python manage.py migrate
   python manage.py runserver 0.0.0.0:8300
   # Access: http://localhost:8300
   ```

### Create WordPress (Custom Script)

```bash
./scripts/wordpress/install-wordpress.sh mysite \
  --preset clean \
  --title "My WordPress Site"
# Access: https://mysite.test
```

---

**Quick Start Complete!** 🚀

Your app is ready for development. Build outputs will go to `public/` and be served by the web server.
