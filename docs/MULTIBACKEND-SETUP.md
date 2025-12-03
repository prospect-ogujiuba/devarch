# Multi-Backend Routing Setup - Installation Notes

This document describes the changes made to enable .test domain routing and SMTP integration for Node, Python, and Go applications.

## Changes Made

### 1. SMTP Configuration for All Backends

**Node.js (`compose/backend/node.yml` + `config/node/Dockerfile`):**
- Added msmtp package for mail relay
- Added environment variables: SMTP_HOST, SMTP_PORT, MAIL_FROM
- Configured msmtp to relay to mailpit:1025

**Python (`compose/backend/python.yml`):**
- Added Django/Flask email environment variables
- EMAIL_BACKEND, EMAIL_HOST, EMAIL_PORT, EMAIL_USE_TLS, EMAIL_FROM
- Added generic SMTP_HOST and SMTP_PORT for other frameworks

**Go (`compose/backend/go.yml`):**
- Added environment variables: SMTP_HOST, SMTP_PORT, SMTP_FROM
- Go uses native net/smtp package (no additional packages needed)

All backends now send emails to Mailpit (localhost:9200 UI, localhost:9201 SMTP).

### 2. Runtime Detection Script

**Created: `scripts/detect-app-runtime.sh`**

Automatically detects app type based on marker files:
- PHP: composer.json, wp-config.php, artisan, index.php
- Node: package.json
- Python: requirements.txt, pyproject.toml, manage.py, main.py
- Go: go.mod, main.go

Priority if multiple markers found: PHP > Node > Python > Go

Usage:
```bash
./scripts/detect-app-runtime.sh myapp           # Returns: php|node|python|go|unknown
./scripts/detect-app-runtime.sh myapp -v        # Verbose mode
./scripts/detect-app-runtime.sh myapp -v -i     # With backend info
```

### 3. Proxy Host Setup Script

**Created: `scripts/setup-proxy-host.sh`**

Automates NPM proxy host setup process:
- Detects runtime type
- Provides detailed manual NPM configuration instructions
- Shows correct backend host:port mapping
- Updates /etc/hosts file
- Runtime-specific setup notes

Usage:
```bash
./scripts/setup-proxy-host.sh myapp           # Full setup with instructions
./scripts/setup-proxy-host.sh myapp --manual  # Manual mode (same for now)
```

### 4. Nginx Backend Router Configuration

**Created: `config/nginx/custom/backend-router.conf`** (see /tmp/backend-router.conf)

Defines upstream backends and provides routing configuration:
- Upstream definitions for all four backends
- Common proxy settings
- Usage instructions for NPM integration

**NOTE:** This file needs to be manually copied to `/home/fhcadmin/projects/devarch/config/nginx/custom/backend-router.conf` with proper permissions:
```bash
sudo cp /tmp/backend-router.conf /home/fhcadmin/projects/devarch/config/nginx/custom/backend-router.conf
sudo chown root:root /home/fhcadmin/projects/devarch/config/nginx/custom/backend-router.conf
```

### 5. Documentation

**Created: `context/multibackend-routing-guide.md`**

Comprehensive guide covering:
- Architecture overview
- SMTP configuration for each runtime
- Runtime detection details
- Step-by-step setup for new apps
- Email testing with Mailpit
- Troubleshooting guide
- Quick reference commands

### 6. Hosts File Management

**No changes needed** - The existing `update-hosts.sh` and `generate-context.sh` already support all app types in the apps/ directory regardless of runtime. They automatically add .test domains for all apps.

## Backend Port Mapping

| Runtime | Container | Internal Port | External Port | NPM Forward To |
|---------|-----------|---------------|---------------|----------------|
| PHP     | php       | 8000          | 8100          | php:8000       |
| Node    | node      | 3000          | 8200          | node:3000      |
| Python  | python    | 8000          | 8300          | python:8000    |
| Go      | go        | 8080          | 8400          | go:8080        |

## Mailpit Access

- **Web UI:** http://localhost:9200
- **SMTP Port:** localhost:9201 (for external testing)
- **Internal:** mailpit:1025 (for containers)

## Container Rebuild Required

After these changes, rebuild the Node container to include msmtp:

```bash
cd /home/fhcadmin/projects/devarch

# Rebuild Node container
podman-compose -f compose/backend/node.yml build --no-cache

# Restart the container
podman-compose -f compose/backend/node.yml up -d
```

Python and Go containers don't need rebuilding (environment variables only).

## Testing the Setup

1. **Test Runtime Detection:**
```bash
./scripts/detect-app-runtime.sh playground    # Should return: php
./scripts/detect-app-runtime.sh b2bcnc        # Should return: php
```

2. **Setup a New App:**
```bash
# Create a test Node app
mkdir -p apps/testnode
cd apps/testnode
npm init -y
echo "console.log('Hello DevArch');" > index.js
cd ../..

# Run setup
./scripts/setup-proxy-host.sh testnode

# Follow the outputted NPM configuration instructions
```

3. **Test Email (after container rebuild):**
```bash
# Enter Node container
podman exec -it node zsh

# Test email with msmtp
echo "Subject: Test Email" | msmtp --host=mailpit --port=1025 --from=test@devarch.test recipient@example.com

# Check Mailpit UI
# Visit: http://localhost:9200
```

## Next Steps

1. Copy backend-router.conf to proper location (see note above)
2. Rebuild Node container to include msmtp
3. Test with a new Node/Python/Go application
4. Review the full guide: `context/multibackend-routing-guide.md`

## Files Modified

- `compose/backend/node.yml` - Added SMTP environment variables
- `compose/backend/python.yml` - Added SMTP environment variables
- `compose/backend/go.yml` - Added SMTP environment variables
- `config/node/Dockerfile` - Added msmtp installation and configuration

## Files Created

- `scripts/detect-app-runtime.sh` - Runtime detection utility
- `scripts/setup-proxy-host.sh` - NPM proxy host setup automation
- `config/nginx/custom/backend-router.conf` - Nginx upstream configuration (in /tmp)
- `context/multibackend-routing-guide.md` - Complete user guide
- `MULTIBACKEND-SETUP.md` - This file

## Design Philosophy

The implementation follows the user's requirements:

1. **Simple and Efficient:** File-based detection, no complex magic
2. **Nginx Does Heavy Lifting:** All routing handled by NPM
3. **Flat Structure:** Apps stay in `apps/appname/`, not `apps/runtime/appname/`
4. **Automatic Detection:** Based on standard marker files
5. **Unified Email Testing:** All backends → Mailpit

## Known Limitations

1. **NPM API Not Implemented:** Currently using manual NPM UI configuration
   - Future enhancement: Automate via NPM REST API
2. **Single Runtime per App:** Detection prioritizes one runtime if multiple markers found
3. **Manual nginx config file placement:** Requires sudo to place backend-router.conf

## Support

For issues or questions:
1. Check `context/multibackend-routing-guide.md` for troubleshooting
2. View container logs: `podman logs <container-name>`
3. Test connectivity: `podman exec nginx-proxy-manager curl http://node:3000`

---

**Setup Date:** 2025-12-02
**Version:** 1.0
