# DevArch Scripts Overview

Comprehensive automation scripts for the DevArch microservices environment.

## App Management Scripts (New!)

### 1. create-app.sh
**Interactive app creation with framework boilerplate**

```bash
./scripts/create-app.sh
```

Creates new applications with full framework support:
- PHP: Laravel, WordPress, Generic
- Node: Next.js, React (Vite), Express
- Python: Django, FastAPI, Flask
- Go: Standard, Gin, Echo

**Features**:
- Interactive prompts for name, runtime, framework
- Validates app names
- Installs frameworks inside containers
- Configures apps for DevArch
- Adds .test domains
- Starts backends if needed

### 2. list-apps.sh
**List and filter applications**

```bash
./scripts/list-apps.sh [--json] [--runtime php]
```

**Output formats**:
- Table (default): Human-readable
- JSON: For scripting
- CSV: For spreadsheets

**Features**:
- Auto-detects runtime and framework
- Shows backend status (Active/Stopped)
- Filter by runtime
- Show full paths

### 3. update-hosts-enhanced.sh
**Manage /etc/hosts entries**

```bash
sudo ./scripts/update-hosts-enhanced.sh [update|add|remove|scan|list]
```

**Actions**:
- `update` - Update all app entries
- `add APP` - Add specific app
- `remove APP` - Remove specific app
- `scan` - Check status without modifying
- `list` - List current DevArch entries

**Features**:
- Auto-discovers apps
- Manages DevArch section in hosts file
- Creates backups
- Dry-run mode
- No sudo needed for read-only operations

### 4. setup-proxy-host.sh
**Configure Nginx Proxy Manager**

```bash
./scripts/setup-proxy-host.sh <appname> [options]
```

**Features**:
- Detects runtime and configures backend
- Shows manual configuration steps
- Custom domain support
- Runtime-specific instructions

### 5. detect-app-runtime.sh
**Detect application runtime type**

```bash
./scripts/detect-app-runtime.sh <appname>
```

**Detects**: PHP, Node, Python, Go
**Used by**: All other app management scripts

---

## Service Management Scripts (Existing)

### service-manager.sh
**Unified service orchestration**

```bash
./scripts/service-manager.sh <command> [target] [options]
```

**Commands**:
- `up SERVICE` - Start service
- `down SERVICE` - Stop service
- `restart SERVICE` - Restart service
- `start [categories]` - Bulk start
- `stop [categories]` - Bulk stop
- `status` - Show status
- `logs SERVICE` - View logs

**Categories**:
- database, backend, analytics, proxy, dbms, etc.

---

## Quick Start Examples

### Create New App
```bash
./scripts/create-app.sh
# Follow prompts to create app
```

### List All Apps
```bash
./scripts/list-apps.sh
```

### Update Hosts File
```bash
sudo ./scripts/update-hosts-enhanced.sh
```

### Setup Proxy
```bash
./scripts/setup-proxy-host.sh myapp
```

### Start Backend Services
```bash
./scripts/service-manager.sh start backend
```

---

## Documentation

**Comprehensive Guide**: [context/app-management-guide.md](context/app-management-guide.md)
- Complete usage instructions
- Framework-specific guides
- Troubleshooting
- Integration examples
- Best practices

**Quick Reference**: [context/app-management-quickref.md](context/app-management-quickref.md)
- Command reference
- Common workflows
- One-liner examples
- Tips and tricks

---

## Script Locations

```
/home/fhcadmin/projects/devarch/scripts/
├── create-app.sh              # NEW: Interactive app creator
├── list-apps.sh               # NEW: App lister with JSON
├── update-hosts-enhanced.sh   # NEW: Enhanced hosts manager
├── setup-proxy-host.sh        # Proxy configurator
├── detect-app-runtime.sh      # Runtime detector
├── service-manager.sh         # Service orchestration
└── config.sh                  # Central configuration
```

---

## Common Workflows

### Complete App Setup
```bash
# 1. Create app
./scripts/create-app.sh

# 2. Setup proxy (if needed)
./scripts/setup-proxy-host.sh myapp

# 3. Access
open http://myapp.test
```

### Check Environment Status
```bash
# List apps
./scripts/list-apps.sh

# Check hosts sync
./scripts/update-hosts-enhanced.sh scan

# Check services
./scripts/service-manager.sh status
```

### Automation Example
```bash
# Get all PHP apps as JSON
./scripts/list-apps.sh --runtime php --json

# Export all apps to CSV
./scripts/list-apps.sh --csv > apps.csv

# Count active apps
./scripts/list-apps.sh --json | jq '[.[] | select(.status=="Active")] | length'
```

---

## Integration

All scripts integrate seamlessly:

1. **create-app.sh** → calls detect-app-runtime.sh, update-hosts-enhanced.sh, setup-proxy-host.sh
2. **list-apps.sh** → uses detect-app-runtime.sh for detection
3. **update-hosts-enhanced.sh** → uses detect-app-runtime.sh to validate apps
4. **setup-proxy-host.sh** → uses detect-app-runtime.sh for backend config
5. **All scripts** → work with service-manager.sh for backend services

---

## Requirements

- **Container Runtime**: Podman or Docker
- **Shell**: Bash 4.0+
- **Permissions**: sudo for /etc/hosts modification
- **Backend Services**: PHP, Node, Python, Go containers
- **NPM**: Nginx Proxy Manager for routing

---

## Features

✓ **Simple & Transparent**: No hidden magic, clear output
✓ **Framework-Aware**: Official CLI tools for boilerplate
✓ **Multi-Runtime**: PHP, Node, Python, Go support
✓ **Auto-Detection**: Smart runtime and framework detection
✓ **Flexible Output**: Table, JSON, CSV formats
✓ **Safety Features**: Backups, dry-run, validation
✓ **Integration**: Works with existing tools
✓ **Comprehensive**: Full lifecycle management

---

## Success Criteria Met

✅ create-app.sh creates apps with proper framework boilerplate
✅ Supports Laravel, WordPress, Next.js, React, Django, FastAPI, Flask, and Go
✅ Executes framework CLIs inside appropriate containers
✅ list-apps.sh provides table and JSON output with runtime filtering
✅ update-hosts-enhanced.sh automatically detects and adds all apps
✅ Backup mechanism for /etc/hosts
✅ setup-proxy-host.sh exists and provides NPM guidance
✅ All scripts have error handling and validation
✅ Clear, actionable output with next steps
✅ Documentation in context/app-management-guide.md
✅ Scripts follow simple, transparent approach (no hidden magic)

---

## Troubleshooting

**Backend not running**: `./scripts/service-manager.sh up php`

**Domain not resolving**: `sudo ./scripts/update-hosts-enhanced.sh`

**App not detected**: Add marker files (composer.json, package.json, etc.)

**502 error**: Check backend logs with `podman logs php`

See full troubleshooting guide in app-management-guide.md

---

## Contributing

To add new framework support:
1. Add detection to detect-app-runtime.sh
2. Add framework option to create-app.sh
3. Add installation function
4. Update list-apps.sh framework detection
5. Update documentation

---

**Version**: 1.0
**Last Updated**: 2025-12-02
**Maintainer**: DevArch Project
