# DevArch App Management - Quick Reference

**TL;DR**: Comprehensive app lifecycle automation for PHP, Node, Python, and Go applications.

---

## Quick Start

### Create New App (Interactive)
```bash
./scripts/create-app.sh
```

### List All Apps
```bash
./scripts/list-apps.sh
./scripts/list-apps.sh --json          # JSON output
./scripts/list-apps.sh --runtime php   # Filter by runtime
```

### Update Hosts File
```bash
sudo ./scripts/update-hosts-enhanced.sh           # Update all
sudo ./scripts/update-hosts-enhanced.sh add myapp # Add specific app
./scripts/update-hosts-enhanced.sh scan           # Check status (no sudo)
```

### Setup Proxy Host
```bash
./scripts/setup-proxy-host.sh myapp
```

---

## Command Reference

### create-app.sh
**Purpose**: Create new app with framework boilerplate

```bash
./scripts/create-app.sh
```

**Interactive prompts for**:
- App name (lowercase, no spaces)
- Runtime: PHP, Node, Python, Go
- Framework: Laravel, WordPress, Next.js, React, Django, etc.

**Automatically**:
- Creates app directory
- Installs framework via CLI tools (in container)
- Configures app
- Adds to /etc/hosts
- Checks/starts backend service

---

### list-apps.sh
**Purpose**: List all apps with runtime/framework detection

```bash
./scripts/list-apps.sh [options]
```

**Options**:
- `--json` - JSON format (for scripts)
- `--csv` - CSV format
- `--runtime RUNTIME` - Filter (php/node/python/go)
- `--paths` - Show full paths
- `-v, --verbose` - Verbose output

**Examples**:
```bash
./scripts/list-apps.sh                    # Table view
./scripts/list-apps.sh --json             # JSON
./scripts/list-apps.sh --runtime php      # PHP apps only
./scripts/list-apps.sh --paths            # With paths
```

---

### update-hosts-enhanced.sh
**Purpose**: Manage /etc/hosts entries for apps

```bash
./scripts/update-hosts-enhanced.sh [action] [app] [options]
```

**Actions**:
- `update` - Update all (default)
- `add APP` - Add specific app
- `remove APP` - Remove specific app
- `scan` - Check status (no modification)
- `list` - List current entries

**Options**:
- `-n, --dry-run` - Preview changes
- `-v, --verbose` - Detailed output
- `--no-backup` - Skip backup

**Examples**:
```bash
sudo ./scripts/update-hosts-enhanced.sh             # Update all
sudo ./scripts/update-hosts-enhanced.sh add myapp   # Add one
sudo ./scripts/update-hosts-enhanced.sh remove old  # Remove one
./scripts/update-hosts-enhanced.sh scan             # Status check
./scripts/update-hosts-enhanced.sh list             # List entries
```

---

### setup-proxy-host.sh
**Purpose**: Configure NPM proxy hosts

```bash
./scripts/setup-proxy-host.sh <appname> [options]
```

**Options**:
- `-m, --manual` - Show manual instructions
- `-d, --domain` - Custom domain
- `-v, --verbose` - Verbose output

**Examples**:
```bash
./scripts/setup-proxy-host.sh myapp              # Auto-detect and guide
./scripts/setup-proxy-host.sh myapp -d custom.test  # Custom domain
./scripts/setup-proxy-host.sh myapp --manual     # Manual steps
```

---

### detect-app-runtime.sh
**Purpose**: Detect app runtime type

```bash
./scripts/detect-app-runtime.sh <appname> [options]
```

**Options**:
- `-v, --verbose` - Show details
- `-i, --info` - Backend info

**Examples**:
```bash
./scripts/detect-app-runtime.sh myapp       # Output: php
./scripts/detect-app-runtime.sh myapp -v    # Verbose
./scripts/detect-app-runtime.sh myapp -i    # Backend info
```

---

## Supported Frameworks

### PHP
- **Laravel**: Full installation via Composer
- **WordPress**: WP-CLI download + setup
- **Generic**: Basic PHP structure

### Node.js
- **Next.js**: With TypeScript + Tailwind
- **React**: Vite template
- **Express**: Basic server setup

### Python
- **Django**: Complete project structure
- **FastAPI**: With uvicorn
- **Flask**: Basic app structure

### Go
- **Standard**: net/http library
- **Gin**: Framework installed
- **Echo**: Framework installed

---

## Common Workflows

### Create & Deploy New App
```bash
# 1. Create app
./scripts/create-app.sh
# Enter: my-api, PHP, Laravel

# 2. Setup proxy (if needed)
./scripts/setup-proxy-host.sh my-api

# 3. Access
open http://my-api.test
```

### Check App Status
```bash
# List all apps
./scripts/list-apps.sh

# Check hosts sync
./scripts/update-hosts-enhanced.sh scan

# Verify backend running
./scripts/service-manager.sh status php
```

### Update Hosts After Changes
```bash
# Update all apps
sudo ./scripts/update-hosts-enhanced.sh

# Or add specific new app
sudo ./scripts/update-hosts-enhanced.sh add newapp
```

### Export App List for Scripts
```bash
# Get all apps as JSON
./scripts/list-apps.sh --json > apps.json

# Filter PHP apps
./scripts/list-apps.sh --runtime php --json

# Count apps
./scripts/list-apps.sh --json | jq 'length'

# Get app URLs
./scripts/list-apps.sh --json | jq -r '.[].url'
```

---

## Backend Management

### Start Backend Service
```bash
./scripts/service-manager.sh up php
./scripts/service-manager.sh up node
./scripts/service-manager.sh up python
./scripts/service-manager.sh up go
```

### Start All Backends
```bash
./scripts/service-manager.sh start backend
```

### Check Backend Status
```bash
./scripts/service-manager.sh status
./scripts/service-manager.sh status php
./scripts/service-manager.sh ps
```

### Restart Backend
```bash
./scripts/service-manager.sh restart php
```

---

## Runtime Detection

**File Markers**:

| Runtime | Markers |
|---------|---------|
| PHP | composer.json, index.php, wp-config.php, artisan |
| Node | package.json |
| Python | requirements.txt, pyproject.toml, manage.py |
| Go | go.mod, main.go |

**Priority**: PHP > Node > Python > Go

---

## Port Mapping

| Runtime | External | Internal | Container |
|---------|----------|----------|-----------|
| PHP     | 8100     | 8000     | php       |
| Node    | 8200     | 3000     | node      |
| Python  | 8300     | 8000     | python    |
| Go      | 8400     | 8080     | go        |

---

## Troubleshooting

### App Creation Fails
```bash
# Check backend running
./scripts/service-manager.sh status php

# Start backend
./scripts/service-manager.sh up php

# Check container logs
podman logs php
```

### Domain Not Resolving
```bash
# Check hosts file
./scripts/update-hosts-enhanced.sh scan

# Update hosts
sudo ./scripts/update-hosts-enhanced.sh

# Verify entry
cat /etc/hosts | grep myapp
```

### 502 Bad Gateway
```bash
# Check backend container
podman ps | grep php

# Check backend logs
podman logs php

# Restart backend
./scripts/service-manager.sh restart php

# Test backend directly
curl http://localhost:8100
```

### App Not Listed
```bash
# Check runtime detection
./scripts/detect-app-runtime.sh myapp -v

# Add required marker file:
# - PHP: touch apps/myapp/composer.json
# - Node: touch apps/myapp/package.json
# - Python: touch apps/myapp/requirements.txt
# - Go: touch apps/myapp/go.mod
```

---

## Integration Examples

### Bash Script
```bash
#!/bin/bash
# Get all PHP apps
APPS=$(./scripts/list-apps.sh --runtime php --json)
echo "$APPS" | jq -r '.[].name' | while read app; do
  echo "Processing: $app"
  # Do something with each app
done
```

### Python Script
```python
import json
import subprocess

# Get apps as JSON
result = subprocess.run(['./scripts/list-apps.sh', '--json'],
                       capture_output=True, text=True)
apps = json.loads(result.stdout)

# Process apps
for app in apps:
    print(f"{app['name']}: {app['runtime']}/{app['framework']}")
    if app['status'] == 'Stopped':
        print(f"  Warning: {app['name']} backend is stopped")
```

### Check All Apps Health
```bash
#!/bin/bash
for app in $(./scripts/list-apps.sh --json | jq -r '.[].name'); do
  url="http://${app}.test"
  if curl -sf "$url" >/dev/null 2>&1; then
    echo "✓ $app - OK"
  else
    echo "✗ $app - FAILED"
  fi
done
```

---

## File Locations

```
/home/fhcadmin/projects/devarch/
├── apps/                           # All applications
├── scripts/
│   ├── create-app.sh              # App creator
│   ├── list-apps.sh               # App lister
│   ├── update-hosts-enhanced.sh   # Hosts manager
│   ├── setup-proxy-host.sh        # Proxy configurator
│   ├── detect-app-runtime.sh      # Runtime detector
│   └── service-manager.sh         # Service manager
└── context/
    ├── app-management-guide.md    # Full documentation
    └── app-management-quickref.md # This file
```

---

## Tips

1. **Always start backend first**: Run `./scripts/service-manager.sh up php` before creating PHP apps

2. **Use scan to check sync**: Run `./scripts/update-hosts-enhanced.sh scan` to see what needs updating

3. **JSON for scripting**: Use `--json` flag for programmatic access to app data

4. **Filter for focused view**: Use `--runtime` to see apps for specific backend

5. **Dry run for safety**: Use `--dry-run` to preview hosts file changes

6. **Check logs on failure**: Container logs (`podman logs <container>`) show detailed errors

7. **Framework markers matter**: Ensure proper marker files exist for accurate detection

---

## One-Liner Examples

```bash
# Count apps by runtime
./scripts/list-apps.sh --json | jq -r '.[].runtime' | sort | uniq -c

# List stopped apps
./scripts/list-apps.sh --json | jq -r '.[] | select(.status=="Stopped") | .name'

# Get all app URLs
./scripts/list-apps.sh --json | jq -r '.[].url'

# Export to CSV
./scripts/list-apps.sh --csv > apps.csv

# Update hosts and show changes
sudo ./scripts/update-hosts-enhanced.sh --verbose

# Scan and count missing entries
./scripts/update-hosts-enhanced.sh scan | grep -c "not in hosts"
```

---

## Next Steps

After creating an app:
1. Configure NPM proxy host
2. Update /etc/hosts (if not automatic)
3. Access app via browser
4. Check dashboard: http://dashboard.test

---

**Full Documentation**: `/home/fhcadmin/projects/devarch/context/app-management-guide.md`

**Last Updated**: 2025-12-02
