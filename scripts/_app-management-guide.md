# DevArch App Management Guide

Comprehensive guide to managing applications in the DevArch environment using automated scripts.

## Table of Contents

- [Overview](#overview)
- [Scripts Reference](#scripts-reference)
  - [create-app.sh](#create-appsh)
  - [list-apps.sh](#list-appssh)
  - [update-hosts-enhanced.sh](#update-hosts-enhancedsh)
  - [setup-proxy-host.sh](#setup-proxy-hostsh)
  - [detect-app-runtime.sh](#detect-app-runtimesh)
- [Common Workflows](#common-workflows)
- [Framework Support](#framework-support)
- [Troubleshooting](#troubleshooting)
- [Integration with Existing Tools](#integration-with-existing-tools)

---

## Overview

The DevArch app management automation provides a complete lifecycle toolkit for creating, managing, and deploying applications across multiple backend runtimes:

- **PHP**: Laravel, WordPress, Generic PHP
- **Node.js**: Next.js, React (Vite), Express
- **Python**: Django, FastAPI, Flask
- **Go**: Standard library, Gin, Echo

All scripts follow a simple, transparent approach with no hidden magic. They integrate seamlessly with the existing service-manager.sh and provide clear feedback at every step.

---

## Scripts Reference

### create-app.sh

**Purpose**: Interactive app creation with framework-aware boilerplate generation.

**Location**: `/home/fhcadmin/projects/devarch/scripts/create-app.sh`

**Usage**:
```bash
./scripts/create-app.sh
```

**Features**:
- Interactive prompts for app name, runtime, and framework
- Validates app names (lowercase, no spaces, valid characters)
- Checks if app already exists
- Installs framework boilerplate using official CLI tools inside containers
- Configures apps with DevArch-specific settings
- Adds .test domain to /etc/hosts
- Provides next steps and proxy configuration instructions
- Starts backend service if not running

**Interactive Flow**:
```
DevArch App Creator
===================

App name (lowercase, no spaces): my-api

Select runtime:
  1) PHP
  2) Node.js
  3) Python
  4) Go

Runtime [1-4]: 1

Select PHP framework:
  1) Laravel
  2) WordPress
  3) Generic PHP

Framework [1-3]: 1

Creating Laravel application 'my-api'...
✓ Directory created: apps/my-api
✓ Laravel installed via Composer
✓ Environment configured
✓ Hosts entry added
✓ Backend service 'php' is running

Next Steps:
  1. Configure Nginx Proxy Manager:
     ./scripts/setup-proxy-host.sh my-api

  2. Access your app:
     http://my-api.test

  3. View in dashboard:
     http://dashboard.test

Done!
```

**Framework Installation Details**:

**PHP**:
- **Laravel**: `composer create-project laravel/laravel <appname>` (inside PHP container)
- **WordPress**: `wp core download --path=<appname>` (using WP-CLI)
- **Generic**: Creates basic PHP structure with index.php and composer.json

**Node.js**:
- **Next.js**: `npx create-next-app@latest <appname> --typescript --tailwind`
- **React (Vite)**: `npm create vite@latest <appname> -- --template react`
- **Express**: Creates basic Express app with package.json and index.js

**Python**:
- **Django**: `django-admin startproject <appname>`
- **FastAPI**: Creates structure with main.py and requirements.txt
- **Flask**: Creates structure with app.py and requirements.txt

**Go**:
- **Standard**: `go mod init <appname>` with basic main.go
- **Gin**: Go module + Gin framework installation
- **Echo**: Go module + Echo framework installation

**Validation Rules**:
- App name must be lowercase
- Only letters, numbers, hyphens, and underscores allowed
- Cannot start with a number
- 2-50 characters length
- Must not already exist

**Container Execution**:
All framework installations execute inside the appropriate backend container:
```bash
# Example for Laravel
podman exec -w /var/www/html php composer create-project laravel/laravel my-app

# Example for Django
podman exec -w /var/www/html python django-admin startproject my-app
```

---

### list-apps.sh

**Purpose**: Command-line app listing with runtime detection and multiple output formats.

**Location**: `/home/fhcadmin/projects/devarch/scripts/list-apps.sh`

**Usage**:
```bash
./scripts/list-apps.sh [options]
```

**Options**:
- `--json` - Output in JSON format (for scripting)
- `--csv` - Output in CSV format
- `--runtime RUNTIME` - Filter by runtime (php, node, python, go)
- `--paths` - Show full paths in table output
- `-v, --verbose` - Verbose output
- `-h, --help` - Show help message

**Examples**:

1. **List all apps (table format)**:
```bash
./scripts/list-apps.sh
```

Output:
```
DevArch Applications
====================

NAME              RUNTIME   FRAMEWORK    STATUS   URL
─────────────────────────────────────────────────────────────────────────────────────
b2bcnc            php       WordPress    Active   http://b2bcnc.test
playground        php       WordPress    Active   http://playground.test
my-api            php       Laravel      Active   http://my-api.test
react-frontend    node      React        Stopped  http://react-frontend.test
dashboard         node      Next.js      Active   http://dashboard.test

Total: 5 application(s)
```

2. **JSON output (for scripting)**:
```bash
./scripts/list-apps.sh --json
```

Output:
```json
[
  {
    "name": "b2bcnc",
    "runtime": "php",
    "framework": "WordPress",
    "status": "Active",
    "url": "http://b2bcnc.test",
    "path": "/home/fhcadmin/projects/devarch/apps/b2bcnc"
  },
  {
    "name": "my-api",
    "runtime": "php",
    "framework": "Laravel",
    "status": "Active",
    "url": "http://my-api.test",
    "path": "/home/fhcadmin/projects/devarch/apps/my-api"
  }
]
```

3. **Filter by runtime**:
```bash
./scripts/list-apps.sh --runtime php
./scripts/list-apps.sh --runtime node --json
```

4. **Show paths**:
```bash
./scripts/list-apps.sh --paths
```

**Framework Detection**:

The script automatically detects frameworks based on file markers:

- **PHP**:
  - Laravel: `artisan` file or `laravel/framework` in composer.json
  - WordPress: `wp-config.php` or `wp-settings.php`
  - Composer: `composer.json` present
  - Generic: fallback

- **Node.js**:
  - Next.js: `next.config.js` or `"next"` in package.json
  - React: `"react"` in package.json (with Vite if also has `"vite"`)
  - Express: `"express"` in package.json
  - Vue.js: `"vue"` in package.json

- **Python**:
  - Django: `manage.py` present
  - FastAPI: `fastapi` in requirements.txt or pyproject.toml
  - Flask: `flask` in requirements.txt or `app.py` present

- **Go**:
  - Gin: `github.com/gin-gonic/gin` in go.mod
  - Echo: `github.com/labstack/echo` in go.mod
  - Fiber: `github.com/gofiber/fiber` in go.mod
  - Standard: fallback (net/http)

**Status Detection**:

Status is determined by checking if the backend container is running:
- **Active**: Backend container is running
- **Stopped**: Backend container is stopped

---

### update-hosts-enhanced.sh

**Purpose**: Enhanced hosts file management with auto-detection and granular control.

**Location**: `/home/fhcadmin/projects/devarch/scripts/update-hosts-enhanced.sh`

**Usage**:
```bash
sudo ./scripts/update-hosts-enhanced.sh [action] [app] [options]
```

**Actions**:
- `update` - Update all app entries (default)
- `add APP` - Add specific app entry
- `remove APP` - Remove specific app entry
- `scan` - Scan and report without modifying
- `list` - List all DevArch entries

**Options**:
- `-n, --dry-run` - Show changes without modifying
- `-v, --verbose` - Verbose output
- `--no-backup` - Skip backup creation
- `-h, --help` - Show help

**Examples**:

1. **Update all apps**:
```bash
sudo ./scripts/update-hosts-enhanced.sh update
# or simply:
sudo ./scripts/update-hosts-enhanced.sh
```

2. **Add single app**:
```bash
sudo ./scripts/update-hosts-enhanced.sh add my-api
```

3. **Remove single app**:
```bash
sudo ./scripts/update-hosts-enhanced.sh remove old-app
```

4. **Scan without modifying** (no sudo needed):
```bash
./scripts/update-hosts-enhanced.sh scan
```

Output:
```
Apps in /home/fhcadmin/projects/devarch/apps:
--------------------
  - b2bcnc
  - dashboard
  - my-api
  - playground

Total discovered: 4

Current DevArch entries in /etc/hosts:
----------------------------------------
  - b2bcnc.test
  - playground.test

Total in hosts: 2

Analysis:
---------
Apps not in hosts file:
  - dashboard
  - my-api

Run: sudo ./scripts/update-hosts-enhanced.sh update
```

5. **List current entries** (no sudo needed):
```bash
./scripts/update-hosts-enhanced.sh list
```

6. **Dry run**:
```bash
sudo ./scripts/update-hosts-enhanced.sh --dry-run --verbose
```

**Hosts File Management**:

The script uses section markers to manage DevArch entries:

```
# DevArch Apps - Start
# Auto-managed by DevArch - Do not edit manually
#
127.0.0.1    b2bcnc.test
127.0.0.1    dashboard.test
127.0.0.1    my-api.test
127.0.0.1    playground.test
#
# Total: 4 application(s)
# DevArch Apps - End
```

**Safety Features**:
- Creates timestamped backup before modification: `/etc/hosts.backup.YYYYMMDD_HHMMSS`
- Only modifies DevArch section (preserves all other entries)
- Validates operations before applying
- Supports dry-run mode to preview changes
- Requires sudo for modification (read-only operations don't need sudo)

**Auto-Detection**:

The script automatically:
1. Scans the `apps/` directory
2. Uses `detect-app-runtime.sh` to validate apps
3. Filters out hidden directories and non-app folders
4. Generates .test domains for all valid apps
5. Maintains alphabetical order

---

### setup-proxy-host.sh

**Purpose**: Automates Nginx Proxy Manager proxy host configuration.

**Location**: `/home/fhcadmin/projects/devarch/scripts/setup-proxy-host.sh`

**Usage**:
```bash
./scripts/setup-proxy-host.sh <appname> [options]
```

**Options**:
- `-m, --manual` - Output manual configuration instructions
- `-v, --verbose` - Verbose output
- `-d, --domain DOMAIN` - Custom domain (default: appname.test)
- `-h, --help` - Show help

**Examples**:

1. **Setup proxy for app**:
```bash
./scripts/setup-proxy-host.sh my-api
```

2. **Custom domain**:
```bash
./scripts/setup-proxy-host.sh my-api -d myapi.local
```

3. **Show manual instructions**:
```bash
./scripts/setup-proxy-host.sh my-api --manual
```

**Backend Mapping**:

The script automatically configures the correct backend based on runtime:

| Runtime | Container | Port |
|---------|-----------|------|
| PHP     | php       | 8000 |
| Node    | node      | 3000 |
| Python  | python    | 8000 |
| Go      | go        | 8080 |

**Manual Configuration Output**:

The script provides detailed instructions for NPM setup:
- Domain configuration
- Forward host/port settings
- SSL configuration (optional)
- Advanced Nginx directives
- Runtime-specific notes
- Troubleshooting steps

---

### detect-app-runtime.sh

**Purpose**: Detects application runtime type based on file markers.

**Location**: `/home/fhcadmin/projects/devarch/scripts/detect-app-runtime.sh`

**Usage**:
```bash
./scripts/detect-app-runtime.sh <appname> [options]
```

**Options**:
- `-v, --verbose` - Verbose output
- `-i, --info` - Show backend information
- `-h, --help` - Show help

**Examples**:

1. **Detect runtime**:
```bash
./scripts/detect-app-runtime.sh my-api
# Output: php
```

2. **With backend info**:
```bash
./scripts/detect-app-runtime.sh my-api -i
# Output: port=8100 container=php internal_port=8000
```

3. **Verbose mode**:
```bash
./scripts/detect-app-runtime.sh my-api -v
# Output:
# App: my-api
# Path: /home/fhcadmin/projects/devarch/apps/my-api
# Runtime: php
```

**Detection Logic**:

**Priority**: PHP > Node > Python > Go

**Markers**:
- **PHP**: composer.json, index.php, wp-config.php, artisan
- **Node**: package.json
- **Python**: requirements.txt, pyproject.toml, manage.py, main.py
- **Go**: go.mod, main.go

**Integration**:
This script is used by:
- `create-app.sh` (runtime validation)
- `list-apps.sh` (app discovery)
- `update-hosts-enhanced.sh` (app validation)
- `setup-proxy-host.sh` (backend configuration)

---

## Common Workflows

### Creating a New Application

**Complete workflow from creation to access**:

```bash
# 1. Create the app (interactive)
./scripts/create-app.sh

# Follow prompts to select:
# - App name: my-api
# - Runtime: PHP (1)
# - Framework: Laravel (1)

# 2. Script automatically:
#    - Creates apps/my-api
#    - Installs Laravel
#    - Configures app
#    - Adds to /etc/hosts
#    - Checks backend status

# 3. Setup NPM proxy (if needed)
./scripts/setup-proxy-host.sh my-api

# 4. Access your app
open http://my-api.test
```

### Listing All Applications

```bash
# Quick list (table)
./scripts/list-apps.sh

# Detailed with paths
./scripts/list-apps.sh --paths

# Export to JSON for scripting
./scripts/list-apps.sh --json > apps.json

# Filter by runtime
./scripts/list-apps.sh --runtime php
./scripts/list-apps.sh --runtime node
```

### Managing Hosts Entries

```bash
# Update all apps in hosts file
sudo ./scripts/update-hosts-enhanced.sh

# Check what needs updating (no modification)
./scripts/update-hosts-enhanced.sh scan

# Add specific app
sudo ./scripts/update-hosts-enhanced.sh add new-app

# Remove old app
sudo ./scripts/update-hosts-enhanced.sh remove old-app

# Preview changes without applying
sudo ./scripts/update-hosts-enhanced.sh --dry-run --verbose
```

### Setting Up Proxy Hosts

```bash
# Auto-detect and show instructions
./scripts/setup-proxy-host.sh my-app

# Custom domain
./scripts/setup-proxy-host.sh my-app -d custom.local

# Force manual mode
./scripts/setup-proxy-host.sh my-app --manual
```

### Managing Backend Services

```bash
# Start a backend service
./scripts/service-manager.sh up php
./scripts/service-manager.sh up node

# Check status
./scripts/service-manager.sh status php

# Restart backend
./scripts/service-manager.sh restart python

# Start all backends
./scripts/service-manager.sh start backend
```

---

## Framework Support

### PHP Applications

#### Laravel

**Creation**:
```bash
./scripts/create-app.sh
# Select PHP > Laravel
```

**What's installed**:
- Complete Laravel installation via Composer
- Generated application key
- Configured permissions for storage and cache

**Post-creation steps**:
1. Configure database in `.env`
2. Run migrations: `podman exec -w /var/www/html/my-app php php artisan migrate`
3. Access: `http://my-app.test`

**Development**:
```bash
# Run artisan commands
podman exec -w /var/www/html/my-app php php artisan <command>

# Queue worker
podman exec -w /var/www/html/my-app php php artisan queue:work

# Tinker
podman exec -it -w /var/www/html/my-app php php artisan tinker
```

#### WordPress

**Creation**:
```bash
./scripts/create-app.sh
# Select PHP > WordPress
```

**What's installed**:
- WordPress core files via WP-CLI
- Configured permissions

**Post-creation steps**:
1. Complete web-based setup: `http://my-wp.test`
2. Or use WP-CLI:
```bash
podman exec -w /var/www/html/my-wp php wp core config \
  --dbname=wordpress \
  --dbuser=root \
  --dbpass=123456 \
  --dbhost=mariadb \
  --allow-root

podman exec -w /var/www/html/my-wp php wp core install \
  --url=http://my-wp.test \
  --title="My Site" \
  --admin_user=admin \
  --admin_password=admin \
  --admin_email=admin@example.com \
  --allow-root
```

**Development**:
```bash
# WP-CLI commands
podman exec -w /var/www/html/my-wp php wp <command> --allow-root

# Install plugin
podman exec -w /var/www/html/my-wp php wp plugin install contact-form-7 --activate --allow-root

# Update WordPress
podman exec -w /var/www/html/my-wp php wp core update --allow-root
```

#### Generic PHP

**Creation**:
```bash
./scripts/create-app.sh
# Select PHP > Generic PHP
```

**Structure**:
```
my-php-app/
├── composer.json
├── public/
│   └── index.php
└── src/
```

**Development**:
- Add your PHP code in `src/`
- Entry point: `public/index.php`
- Use Composer for dependencies

---

### Node.js Applications

#### Next.js

**Creation**:
```bash
./scripts/create-app.sh
# Select Node > Next.js
```

**What's installed**:
- Next.js with TypeScript
- Tailwind CSS
- ESLint
- App Router
- Import alias (@/*)

**Development**:
```bash
# Start dev server
cd apps/my-nextapp
npm run dev

# Build for production
npm run build
npm start
```

**Port Configuration**:
- Development: Port 3000 (inside container)
- Production: Port 3000 (inside container)
- External access via NPM: http://my-nextapp.test

#### React (Vite)

**Creation**:
```bash
./scripts/create-app.sh
# Select Node > React (Vite)
```

**What's installed**:
- Vite with React template
- Fast development server
- HMR (Hot Module Replacement)

**Development**:
```bash
# Start dev server
cd apps/my-react-app
npm run dev

# Build for production
npm run build
npm run preview
```

#### Express

**Creation**:
```bash
./scripts/create-app.sh
# Select Node > Express
```

**Structure**:
```
my-express-app/
├── package.json
└── index.js
```

**Development**:
```bash
# Start server
cd apps/my-express-app
npm start

# Development mode with nodemon
npm run dev
```

---

### Python Applications

#### Django

**Creation**:
```bash
./scripts/create-app.sh
# Select Python > Django
```

**Post-creation steps**:
```bash
# Run migrations
podman exec -w /var/www/html/my-django python python manage.py migrate

# Create superuser
podman exec -it -w /var/www/html/my-django python python manage.py createsuperuser

# Start dev server
podman exec -w /var/www/html/my-django python python manage.py runserver 0.0.0.0:8000
```

**Access**:
- App: `http://my-django.test`
- Admin: `http://my-django.test/admin`

#### FastAPI

**Creation**:
```bash
./scripts/create-app.sh
# Select Python > FastAPI
```

**Structure**:
```
my-fastapi-app/
├── main.py
└── requirements.txt
```

**Development**:
```bash
# Install dependencies
podman exec -w /var/www/html/my-fastapi python pip install -r requirements.txt

# Start with uvicorn
podman exec -w /var/www/html/my-fastapi python uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

**Access**:
- App: `http://my-fastapi.test`
- API Docs: `http://my-fastapi.test/docs`
- ReDoc: `http://my-fastapi.test/redoc`

#### Flask

**Creation**:
```bash
./scripts/create-app.sh
# Select Python > Flask
```

**Structure**:
```
my-flask-app/
├── app.py
└── requirements.txt
```

**Development**:
```bash
# Install dependencies
podman exec -w /var/www/html/my-flask python pip install -r requirements.txt

# Start server
podman exec -w /var/www/html/my-flask python python app.py

# Or with Flask CLI
podman exec -w /var/www/html/my-flask python flask run --host=0.0.0.0 --port=8000
```

---

### Go Applications

#### Standard (net/http)

**Creation**:
```bash
./scripts/create-app.sh
# Select Go > Standard
```

**Structure**:
```
my-go-app/
├── go.mod
└── main.go
```

**Development**:
```bash
# Run app
podman exec -w /var/www/html/my-go go go run main.go

# Build binary
podman exec -w /var/www/html/my-go go go build -o app main.go

# Run binary
podman exec -w /var/www/html/my-go go ./app
```

#### Gin

**Creation**:
```bash
./scripts/create-app.sh
# Select Go > Gin
```

**What's installed**:
- Go module initialized
- Gin framework installed
- Basic route structure

#### Echo

**Creation**:
```bash
./scripts/create-app.sh
# Select Go > Echo
```

**What's installed**:
- Go module initialized
- Echo framework installed
- Middleware configured
- Basic route structure

---

## Troubleshooting

### App Creation Issues

**Problem**: "Backend service 'php' is not running"

**Solution**:
```bash
# Start the backend service
./scripts/service-manager.sh up php

# Or start all backends
./scripts/service-manager.sh start backend

# Then retry app creation
./scripts/create-app.sh
```

**Problem**: Framework installation fails

**Solution**:
```bash
# Check container logs
podman logs php
podman logs node
podman logs python
podman logs go

# Ensure container has internet access
podman exec php ping -c 3 google.com

# Retry with verbose output
# (The script shows installation output by default)
```

**Problem**: "App already exists"

**Solution**:
```bash
# List existing apps
./scripts/list-apps.sh

# Choose different name or remove existing app
rm -rf apps/old-app-name
```

### Hosts File Issues

**Problem**: Hosts entry not added

**Solution**:
```bash
# Check if running with sudo
sudo ./scripts/update-hosts-enhanced.sh

# Manually verify entry was added
./scripts/update-hosts-enhanced.sh list

# Check scan results
./scripts/update-hosts-enhanced.sh scan
```

**Problem**: Domain doesn't resolve

**Solution**:
```bash
# Verify entry in hosts file
cat /etc/hosts | grep devarch -i

# Test DNS resolution
ping my-app.test

# Clear DNS cache (if needed)
# Linux: sudo systemd-resolve --flush-caches
# macOS: sudo dscacheutil -flushcache
# Windows: ipconfig /flushdns
```

### Proxy Issues

**Problem**: App not accessible via domain

**Solution**:
```bash
# 1. Check backend is running
./scripts/service-manager.sh status php

# 2. Check container is accessible
podman exec nginx-proxy-manager curl http://php:8000

# 3. Check NPM logs
podman logs nginx-proxy-manager

# 4. Check app logs
podman logs php

# 5. Verify proxy host in NPM UI
open http://localhost:81
```

**Problem**: 502 Bad Gateway

**Solution**:
```bash
# Check backend container is running
podman ps | grep php

# Check if app is listening
podman exec php netstat -tulpn | grep 8000

# Restart backend
./scripts/service-manager.sh restart php

# Check network connectivity
podman exec nginx-proxy-manager ping php
```

### List Apps Issues

**Problem**: App not showing in list

**Solution**:
```bash
# Check if app has runtime markers
ls -la apps/my-app

# Test runtime detection
./scripts/detect-app-runtime.sh my-app -v

# Add required marker files:
# PHP: composer.json or index.php
# Node: package.json
# Python: requirements.txt or manage.py
# Go: go.mod or main.go
```

**Problem**: Wrong runtime detected

**Solution**:
```bash
# Check detection priority (PHP > Node > Python > Go)
./scripts/detect-app-runtime.sh my-app -v

# If app has mixed markers, highest priority wins
# Remove unwanted marker files to fix detection
```

---

## Integration with Existing Tools

### service-manager.sh Integration

The app management scripts integrate seamlessly with service-manager.sh:

**Starting backends before app creation**:
```bash
# Start required backend
./scripts/service-manager.sh up php

# Create app (will detect backend is running)
./scripts/create-app.sh
```

**Starting all backends**:
```bash
# Start all backend services
./scripts/service-manager.sh start backend

# This starts: php, node, python, go
```

**Checking status**:
```bash
# Check specific backend
./scripts/service-manager.sh status php

# Check all services
./scripts/service-manager.sh status
```

### Dashboard Integration

The app dashboard (`http://dashboard.test`) automatically discovers and displays all apps:

**How it works**:
1. Dashboard scans `apps/` directory
2. Uses `detect-app-runtime.sh` for runtime detection
3. Detects frameworks using same logic as `list-apps.sh`
4. Shows status based on backend container state
5. Provides quick links to apps

**Keeping in sync**:
```bash
# After creating new app
./scripts/create-app.sh

# Dashboard will automatically show it on next refresh
# No manual configuration needed
```

### Scripting and Automation

**Example: CI/CD Integration**

```bash
#!/bin/bash
# deploy-app.sh - Deploy app in CI/CD

APP_NAME="$1"
RUNTIME="$2"
FRAMEWORK="$3"

# Check if app exists
if ./scripts/list-apps.sh --json | jq -e ".[] | select(.name==\"$APP_NAME\")" > /dev/null; then
  echo "App $APP_NAME already exists, updating..."
else
  echo "Creating new app $APP_NAME..."
  # App creation would need non-interactive mode (future enhancement)
fi

# Update hosts
sudo ./scripts/update-hosts-enhanced.sh add "$APP_NAME"

# Setup proxy
./scripts/setup-proxy-host.sh "$APP_NAME"

# Get app info
./scripts/list-apps.sh --json | jq ".[] | select(.name==\"$APP_NAME\")"
```

**Example: Monitoring Script**

```bash
#!/bin/bash
# monitor-apps.sh - Check all apps status

# Get all apps as JSON
APPS=$(./scripts/list-apps.sh --json)

# Count by status
ACTIVE=$(echo "$APPS" | jq '[.[] | select(.status=="Active")] | length')
STOPPED=$(echo "$APPS" | jq '[.[] | select(.status=="Stopped")] | length')
TOTAL=$(echo "$APPS" | jq 'length')

echo "Apps Status:"
echo "  Active:  $ACTIVE"
echo "  Stopped: $STOPPED"
echo "  Total:   $TOTAL"

# List stopped apps
echo ""
echo "Stopped apps:"
echo "$APPS" | jq -r '.[] | select(.status=="Stopped") | "  - \(.name) (\(.runtime)/\(.framework))"'
```

**Example: Cleanup Script**

```bash
#!/bin/bash
# cleanup-old-apps.sh - Remove apps not accessed in 30 days

DAYS=30

# Find apps by last access time
for app_dir in apps/*; do
  APP_NAME=$(basename "$app_dir")
  LAST_ACCESS=$(stat -c %X "$app_dir" 2>/dev/null || stat -f %a "$app_dir" 2>/dev/null)
  NOW=$(date +%s)
  AGE=$(( (NOW - LAST_ACCESS) / 86400 ))

  if [[ $AGE -gt $DAYS ]]; then
    echo "App $APP_NAME not accessed in $AGE days"
    read -p "Remove? [y/N]: " confirm
    if [[ "$confirm" == "y" ]]; then
      rm -rf "$app_dir"
      sudo ./scripts/update-hosts-enhanced.sh remove "$APP_NAME"
      echo "Removed $APP_NAME"
    fi
  fi
done
```

---

## Best Practices

### App Naming

**Do**:
- Use lowercase names: `my-api`, `blog-backend`
- Use hyphens for readability: `user-management`
- Keep it short and descriptive: `crm`, `dashboard`, `api`

**Don't**:
- Use spaces: ~~`my api`~~
- Use uppercase: ~~`MyApp`~~
- Use special characters: ~~`my_app!`~~
- Start with numbers: ~~`1-app`~~

### Directory Organization

**Recommended structure**:
```
apps/
├── blog/              # PHP Laravel blog
├── api/               # Node Express API
├── dashboard/         # Node Next.js dashboard
├── analytics/         # Python Django analytics
└── webhook-service/   # Go webhook processor
```

### Hosts File Management

**Best practices**:
- Run `update-hosts-enhanced.sh` after creating/removing apps
- Use `scan` action periodically to check sync status
- Keep backups (script does this automatically)
- Don't manually edit DevArch section in /etc/hosts

### Backend Services

**Keep backends running**:
```bash
# Start all backends at system boot
./scripts/service-manager.sh start-all

# Or add to startup script
# Add to ~/.bashrc or system startup:
# cd /path/to/devarch && ./scripts/service-manager.sh start backend
```

### Monitoring

**Regular checks**:
```bash
# Weekly: Check app status
./scripts/list-apps.sh

# Weekly: Verify hosts file sync
./scripts/update-hosts-enhanced.sh scan

# Monthly: Review backend resource usage
./scripts/service-manager.sh ps
```

---

## Advanced Usage

### Batch Operations

**Create multiple apps**:
```bash
# Create multiple PHP apps
for app in blog forum shop; do
  echo -e "${app}\n1\n1" | ./scripts/create-app.sh
done
```

**Update hosts for specific apps**:
```bash
# Add multiple apps to hosts
for app in app1 app2 app3; do
  sudo ./scripts/update-hosts-enhanced.sh add "$app"
done
```

### Custom Domains

**Use custom domains instead of .test**:
```bash
# Create app
./scripts/create-app.sh
# Enter: myapp

# Setup proxy with custom domain
./scripts/setup-proxy-host.sh myapp -d myapp.local

# Manually add to hosts
echo "127.0.0.1 myapp.local" | sudo tee -a /etc/hosts
```

### Multi-Environment

**Development, Staging, Production**:
```bash
# Development
./scripts/create-app.sh  # myapp
./scripts/setup-proxy-host.sh myapp -d myapp-dev.test

# Staging
./scripts/create-app.sh  # myapp-staging
./scripts/setup-proxy-host.sh myapp-staging -d myapp-staging.test

# Production
./scripts/create-app.sh  # myapp-prod
./scripts/setup-proxy-host.sh myapp-prod -d myapp.com
```

---

## Reference

### Port Allocation

| Runtime | External Port | Internal Port | Container |
|---------|--------------|---------------|-----------|
| PHP     | 8100         | 8000          | php       |
| Node    | 8200         | 3000          | node      |
| Python  | 8300         | 8000          | python    |
| Go      | 8400         | 8080          | go        |

### File Locations

| Purpose | Location |
|---------|----------|
| Scripts | `/home/fhcadmin/projects/devarch/scripts/` |
| Apps | `/home/fhcadmin/projects/devarch/apps/` |
| Config | `/home/fhcadmin/projects/devarch/scripts/config.sh` |
| Hosts backup | `/etc/hosts.backup.YYYYMMDD_HHMMSS` |
| Dashboard | `http://dashboard.test` |
| NPM UI | `http://localhost:81` |

### Quick Reference

```bash
# Create app
./scripts/create-app.sh

# List apps
./scripts/list-apps.sh
./scripts/list-apps.sh --json
./scripts/list-apps.sh --runtime php

# Update hosts
sudo ./scripts/update-hosts-enhanced.sh
sudo ./scripts/update-hosts-enhanced.sh add myapp
sudo ./scripts/update-hosts-enhanced.sh remove oldapp
./scripts/update-hosts-enhanced.sh scan

# Setup proxy
./scripts/setup-proxy-host.sh myapp

# Detect runtime
./scripts/detect-app-runtime.sh myapp

# Manage services
./scripts/service-manager.sh start backend
./scripts/service-manager.sh up php
./scripts/service-manager.sh status
```

---

## Changelog

### Version 1.0 (Current)
- Initial release
- Full framework support for PHP, Node, Python, Go
- Auto-detection and hosts management
- Integration with service-manager.sh
- Comprehensive documentation

---

## Contributing

To add support for new frameworks:

1. Add detection logic to `detect-app-runtime.sh`
2. Add framework option to `create-app.sh`
3. Add installation function in `create-app.sh`
4. Update detection in `list-apps.sh`
5. Update this documentation

---

## Support

For issues or questions:
1. Check [Troubleshooting](#troubleshooting) section
2. Review script output (use `-v` for verbose mode)
3. Check container logs
4. Consult service-manager.sh documentation

---

**Last Updated**: 2025-12-02
**Version**: 1.0
**Maintainer**: DevArch Project
