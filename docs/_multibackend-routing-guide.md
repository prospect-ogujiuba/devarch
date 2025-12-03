# Multi-Backend Routing Guide

Complete guide for configuring .test domain routing and SMTP integration for Node, Python, and Go applications in the DevArch environment.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Backend Configuration](#backend-configuration)
4. [Runtime Detection](#runtime-detection)
5. [Setting Up a New App](#setting-up-a-new-app)
6. [Email Testing with Mailpit](#email-testing-with-mailpit)
7. [Troubleshooting](#troubleshooting)

---

## Overview

DevArch now supports seamless .test domain routing for applications across four backend runtimes:

- **PHP** (Laravel, WordPress, Symfony, etc.)
- **Node.js** (Express, NestJS, Next.js, etc.)
- **Python** (Django, Flask, FastAPI, etc.)
- **Go** (Gin, Echo, Fiber, etc.)

All applications use the same simple pattern:
- Apps live in flat `apps/` directory structure
- Nginx Proxy Manager handles SSL and routing
- Mailpit provides unified email testing
- Detection is automatic based on file markers

---

## Architecture

### Port Mapping

| Runtime | Container | Internal Port | External Port | Access URL |
|---------|-----------|---------------|---------------|------------|
| PHP     | php       | 8000          | 8100          | localhost:8100 |
| Node    | node      | 3000          | 8200          | localhost:8200 |
| Python  | python    | 8000          | 8300          | localhost:8300 |
| Go      | go        | 8080          | 8400          | localhost:8400 |

### Network Flow

```
Browser Request (appname.test)
    ↓
Nginx Proxy Manager (ports 80/443)
    ↓
Runtime Detection & Routing
    ↓
Backend Container (internal port)
    ↓
Application Response
```

### Directory Structure

```
devarch/
├── apps/
│   ├── mywordpress/          # PHP - has composer.json or wp-config.php
│   ├── mynodeapp/            # Node - has package.json
│   ├── mydjangoapp/          # Python - has requirements.txt or manage.py
│   └── mygoapp/              # Go - has go.mod
├── config/
│   └── nginx/
│       └── custom/
│           └── backend-router.conf
└── scripts/
    ├── detect-app-runtime.sh
    ├── setup-proxy-host.sh
    ├── update-hosts.sh
    └── generate-context.sh
```

---

## Backend Configuration

### SMTP Configuration (Mailpit)

All backend containers are configured with SMTP environment variables pointing to Mailpit:

#### Node.js
- `SMTP_HOST=mailpit`
- `SMTP_PORT=1025`
- `MAIL_FROM=noreply@devarch.test`
- msmtp installed for sendmail compatibility

**Usage Example (nodemailer):**
```javascript
const nodemailer = require('nodemailer');

const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST,
  port: process.env.SMTP_PORT,
  secure: false,
});

await transporter.sendMail({
  from: process.env.MAIL_FROM,
  to: 'user@example.com',
  subject: 'Test Email',
  text: 'Hello from Node.js!',
});
```

#### Python
- `EMAIL_BACKEND=django.core.mail.backends.smtp.EmailBackend`
- `EMAIL_HOST=mailpit`
- `EMAIL_PORT=1025`
- `EMAIL_USE_TLS=False`
- `EMAIL_FROM=noreply@devarch.test`

**Usage Example (Django):**
```python
from django.core.mail import send_mail

send_mail(
    'Test Email',
    'Hello from Django!',
    'noreply@devarch.test',
    ['user@example.com'],
    fail_silently=False,
)
```

**Usage Example (Flask):**
```python
from flask import Flask
from flask_mail import Mail, Message

app = Flask(__name__)
app.config['MAIL_SERVER'] = os.getenv('EMAIL_HOST')
app.config['MAIL_PORT'] = int(os.getenv('EMAIL_PORT'))
app.config['MAIL_USE_TLS'] = False
app.config['MAIL_USE_SSL'] = False

mail = Mail(app)

msg = Message('Test Email',
              sender='noreply@devarch.test',
              recipients=['user@example.com'])
msg.body = 'Hello from Flask!'
mail.send(msg)
```

#### Go
- `SMTP_HOST=mailpit`
- `SMTP_PORT=1025`
- `SMTP_FROM=noreply@devarch.test`

**Usage Example (net/smtp):**
```go
package main

import (
    "net/smtp"
    "os"
)

func main() {
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    from := os.Getenv("SMTP_FROM")

    msg := []byte("To: user@example.com\r\n" +
        "Subject: Test Email\r\n" +
        "\r\n" +
        "Hello from Go!\r\n")

    err := smtp.SendMail(
        smtpHost+":"+smtpPort,
        nil,
        from,
        []string{"user@example.com"},
        msg,
    )
}
```

### Accessing Mailpit UI

All test emails can be viewed at:
- **URL:** http://localhost:9200
- **SMTP Port:** localhost:9201 (for external testing)

---

## Runtime Detection

### Detection Markers

The `detect-app-runtime.sh` script automatically identifies app types based on these files:

#### PHP Detection
- `composer.json`
- `index.php`
- `public/index.php`
- `wp-config.php`
- `artisan` (Laravel)

#### Node Detection
- `package.json`

#### Python Detection
- `requirements.txt`
- `pyproject.toml`
- `manage.py` (Django)
- `main.py` (with Python imports)

#### Go Detection
- `go.mod`
- `main.go`

### Detection Priority

If multiple markers are found (e.g., a Node app with a Python requirements.txt for tooling), the priority is:

**PHP > Node > Python > Go**

This matches the most common to least common runtime in typical web development environments.

### Using the Detection Script

```bash
# Detect runtime for an app
./scripts/detect-app-runtime.sh myapp
# Output: php|node|python|go|unknown

# Verbose mode
./scripts/detect-app-runtime.sh myapp -v
# Shows app path and detection details

# Get backend info
./scripts/detect-app-runtime.sh myapp -i
# Output: port=8100 container=php internal_port=8000
```

---

## Setting Up a New App

### Step 1: Create Your App

Place your application in the `apps/` directory:

```bash
# For Node.js app
mkdir apps/mynodeapp
cd apps/mynodeapp
npm init -y

# For Python Django app
mkdir apps/mydjangoapp
cd apps/mydjangoapp
django-admin startproject myproject .

# For Go app
mkdir apps/mygoapp
cd apps/mygoapp
go mod init mygoapp
```

### Step 2: Run Setup Script

```bash
./scripts/setup-proxy-host.sh mynodeapp
```

This will:
1. Detect the runtime type
2. Show NPM configuration instructions
3. Provide backend routing details
4. Offer to update /etc/hosts

### Step 3: Configure Nginx Proxy Manager

**Option A: Manual Configuration (Recommended for now)**

The setup script outputs detailed instructions. Key steps:

1. Access NPM UI: http://localhost:81
2. Create new Proxy Host
3. Set domain: `mynodeapp.test`
4. Set forward to detected backend (e.g., `node:3000` for Node apps)
5. Enable SSL if desired
6. Save

**Option B: Manual NPM Configuration Template**

```
Domain Names: mynodeapp.test
Scheme: http
Forward Hostname: node
Forward Port: 3000
Cache Assets: ✓
Block Common Exploits: ✓
Websockets Support: ✓ (if needed)
```

### Step 4: Update Hosts File

```bash
# Generate updated context (includes hosts.txt)
./scripts/generate-context.sh

# Update system hosts file
sudo ./scripts/update-hosts.sh
```

Or manually add to `/etc/hosts`:
```
127.0.0.1 mynodeapp.test
```

### Step 5: Start Your Application

#### PHP Apps
Already running via PHP-FPM. Just add your code to `apps/myapp/public/`.

#### Node Apps
```bash
# Enter container
podman exec -it node zsh

# Navigate to your app
cd /app/mynodeapp

# Install dependencies
npm install

# Start your app on port 3000
npm start
# or
PORT=3000 npm run dev
```

#### Python Apps
```bash
# Enter container
podman exec -it python zsh

# Navigate to your app
cd /app/mydjangoapp

# Install dependencies
pip install -r requirements.txt

# Run Django on 0.0.0.0:8000 (important!)
python manage.py runserver 0.0.0.0:8000

# Or for Flask
FLASK_RUN_HOST=0.0.0.0 FLASK_RUN_PORT=8000 flask run

# Or for FastAPI
uvicorn main:app --host 0.0.0.0 --port 8000
```

#### Go Apps
```bash
# Enter container
podman exec -it go zsh

# Navigate to your app
cd /app/mygoapp

# Run your app on 0.0.0.0:8080
go run main.go
```

### Step 6: Access Your App

Visit `http://mynodeapp.test` (or `https://` if SSL enabled)

---

## Email Testing with Mailpit

### Why Mailpit?

Mailpit captures all outgoing emails from your applications, allowing you to:
- Test email functionality without sending real emails
- Inspect email content (HTML, text, headers)
- Debug email issues
- Test across all backend runtimes consistently

### Testing Email Functionality

1. **Send a test email from your app** using the examples in the SMTP Configuration section

2. **View the email in Mailpit UI:**
   - Visit http://localhost:9200
   - See all captured emails in the inbox
   - Click to view full email content

3. **Verify email contents:**
   - Check subject, body, headers
   - Test HTML rendering
   - Verify attachments if any

### Per-Runtime Email Examples

All examples send to Mailpit automatically via environment variables configured in the compose files.

---

## Troubleshooting

### App Not Accessible at .test Domain

**Check 1: Is the domain in /etc/hosts?**
```bash
grep mynodeapp.test /etc/hosts
```
If not found, run:
```bash
sudo ./scripts/update-hosts.sh
```

**Check 2: Is NPM proxy host configured?**
- Visit http://localhost:81
- Check "Proxy Hosts" for your app
- Verify domain name and backend settings

**Check 3: Is the backend container running?**
```bash
podman ps | grep node  # or php, python, go
```

**Check 4: Is your app running inside the container?**
```bash
# For Node
podman exec node curl localhost:3000

# For Python
podman exec python curl localhost:8000

# For Go
podman exec go curl localhost:8080
```

### Runtime Detection Issues

**Problem:** `detect-app-runtime.sh` returns "unknown"

**Solution:** Ensure your app has proper marker files:
- Node: Add `package.json`
- Python: Add `requirements.txt` or `pyproject.toml`
- Go: Add `go.mod`
- PHP: Add `composer.json` or ensure `index.php` exists

### Port Conflicts

**Problem:** Port already in use

**Solution:** Check if another service is using the port:
```bash
# Check what's using port 8200 (Node)
sudo lsof -i :8200

# Or use netstat
netstat -tulpn | grep 8200
```

### NPM Can't Reach Backend

**Problem:** NPM shows 502 Bad Gateway

**Check 1: Are containers on the same network?**
```bash
podman network inspect microservices-net
# Both nginx-proxy-manager and backend containers should be listed
```

**Check 2: Can NPM reach the backend?**
```bash
podman exec nginx-proxy-manager curl http://node:3000
```

**Check 3: Is your app listening on 0.0.0.0?**

Applications must listen on `0.0.0.0`, not `127.0.0.1`:
- Node: `app.listen(3000, '0.0.0.0')`
- Python: `python manage.py runserver 0.0.0.0:8000`
- Go: `http.ListenAndServe("0.0.0.0:8080", handler)`

### Emails Not Appearing in Mailpit

**Check 1: Is Mailpit running?**
```bash
podman ps | grep mailpit
```

**Check 2: Can your backend reach Mailpit?**
```bash
podman exec node ping mailpit
podman exec python ping mailpit
podman exec go ping mailpit
```

**Check 3: Are SMTP environment variables set?**
```bash
podman exec node env | grep SMTP
podman exec python env | grep EMAIL
podman exec go env | grep SMTP
```

**Check 4: Check Mailpit logs:**
```bash
podman logs mailpit
```

### Container Logs

View logs for debugging:

```bash
# NPM logs
podman logs nginx-proxy-manager

# Backend logs
podman logs php
podman logs node
podman logs python
podman logs go

# Mailpit logs
podman logs mailpit
```

---

## Quick Reference Commands

```bash
# Detect app runtime
./scripts/detect-app-runtime.sh myapp

# Setup proxy host
./scripts/setup-proxy-host.sh myapp

# Update hosts file
sudo ./scripts/update-hosts.sh

# Generate context (including hosts.txt)
./scripts/generate-context.sh

# Enter containers
podman exec -it node zsh
podman exec -it python zsh
podman exec -it go zsh
podman exec -it php zsh

# View logs
podman logs -f nginx-proxy-manager
podman logs -f node
podman logs -f python
podman logs -f go

# Test backend connectivity
podman exec nginx-proxy-manager curl http://node:3000
podman exec nginx-proxy-manager curl http://python:8000
podman exec nginx-proxy-manager curl http://go:8080
podman exec nginx-proxy-manager curl http://php:8000

# Access Mailpit UI
http://localhost:9200
```

---

## Advanced Configuration

### Custom Nginx Configuration

For apps requiring special Nginx configuration, add custom config in NPM UI under the "Advanced" tab:

```nginx
location / {
    proxy_pass http://node:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket support
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";

    # Increased timeouts for long-running requests
    proxy_connect_timeout 300s;
    proxy_send_timeout 300s;
    proxy_read_timeout 300s;
}
```

### Multiple Ports Per App

If your app uses multiple ports (e.g., main app + API), configure multiple proxy hosts:

```
myapp.test → node:3000 (main app)
api.myapp.test → node:4000 (API)
```

### SSL Configuration

NPM supports custom SSL certificates. For local development:

1. Use the provided local certificate in `/config/nginx/certs/`
2. Or generate a new certificate for your domain
3. Configure in NPM SSL tab when creating proxy host

---

## Best Practices

1. **Keep apps in flat structure:** Don't create `apps/php/`, `apps/node/` subdirectories
2. **Use marker files:** Always include `package.json`, `requirements.txt`, etc.
3. **Listen on 0.0.0.0:** Never bind to 127.0.0.1 in containerized apps
4. **Test emails early:** Use Mailpit from the start to avoid production email issues
5. **Use .test domain:** Consistent with existing DevArch setup
6. **Document custom configs:** If you need special Nginx rules, document them
7. **Check logs first:** Most issues are visible in container logs

---

## Next Steps

- **Add more apps:** The system scales to any number of apps across all runtimes
- **NPM API automation:** Future enhancement for automated proxy host creation
- **SSL automation:** Consider Let's Encrypt integration for production domains
- **Monitoring:** Add health checks and monitoring for production apps

---

**Last Updated:** 2025-12-02
**Version:** 1.0
**Maintainer:** DevArch Team
