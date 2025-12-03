# DevArch Dashboard 2.0 - Quick Start Guide

Get the modern React dashboard up and running in minutes.

## Prerequisites

- Node.js 18+ installed in the Node container or on host
- PHP 8.3+ container running
- Apps directory populated at `/var/www/html/`

## Installation

### Option 1: Run in Node Container (Recommended)

```bash
# Enter the Node container
podman exec -it node bash

# Navigate to dashboard directory
cd /var/www/html/dashboard

# Install dependencies
npm install

# Start development server
npm run dev -- --host 0.0.0.0
```

Access at: `http://localhost:8200` (Node container external port)

### Option 2: Run on Host

```bash
# Navigate to dashboard directory
cd /home/fhcadmin/projects/devarch/apps/dashboard

# Install dependencies
npm install

# Start development server
npm run dev
```

Access at: `http://localhost:5173`

## Testing the API

Before running the React app, verify the PHP API is working:

```bash
# Test API endpoint
curl http://localhost:8100/api/apps.php

# Should return JSON with apps data
# Example response:
# {
#   "success": true,
#   "data": {
#     "apps": [...],
#     "stats": {...}
#   }
# }
```

If you get an error:
1. Verify PHP container is running: `podman ps | grep php`
2. Check file permissions: `ls -la /home/fhcadmin/projects/devarch/apps/dashboard/api/`
3. Test detection library: `podman exec php php /var/www/html/dashboard/api/apps.php`

## Configuration

### Basic Setup

Copy the environment example:

```bash
cp .env.example .env
```

Default values work out of the box. Optional customizations:

```bash
# API URL (change if PHP container uses different port)
VITE_API_BASE_URL=http://localhost:8100/api

# Auto-refresh interval in milliseconds (default: 30 seconds)
VITE_REFRESH_INTERVAL=30000

# Enable/disable features
VITE_ENABLE_AUTO_REFRESH=true
VITE_ENABLE_DARK_MODE=true
```

### Vite Proxy Configuration

The `vite.config.js` is pre-configured to proxy API requests:

```javascript
proxy: {
  '/api': {
    target: 'http://localhost:8100',
    changeOrigin: true,
  }
}
```

This means you can use relative URLs like `/api/apps.php` in development.

## First Run

1. **Start PHP container** (if not already running):
   ```bash
   podman start php
   ```

2. **Verify apps exist**:
   ```bash
   ls -la /home/fhcadmin/projects/devarch/apps/
   ```

3. **Test API**:
   ```bash
   curl http://localhost:8100/api/apps.php | jq
   ```

4. **Install React dependencies**:
   ```bash
   cd /home/fhcadmin/projects/devarch/apps/dashboard
   npm install
   ```

5. **Start dev server**:
   ```bash
   npm run dev
   ```

6. **Open in browser**:
   - Navigate to `http://localhost:5173`
   - You should see the dashboard with all detected apps

## Common Issues

### Port Already in Use

If port 5173 is in use:

```bash
# Find what's using the port
lsof -i :5173

# Kill the process or use a different port
npm run dev -- --port 3000
```

### API Returns Empty Apps

Check that apps exist in the scanned directory:

```bash
# List apps
ls /home/fhcadmin/projects/devarch/apps/

# Should see directories like: b2bcnc, playground, dashboard, etc.
```

### CORS Errors

If you see CORS errors in the browser console:

1. The API already includes CORS headers
2. Verify API is accessible: `curl http://localhost:8100/api/apps.php`
3. Check Vite proxy configuration in `vite.config.js`

### Dark Mode Not Persisting

Clear localStorage and try again:

```javascript
// In browser console
localStorage.clear()
location.reload()
```

## Development Workflow

### Hot Reload

Vite provides instant hot module replacement. Make changes to any `.jsx` file and see them immediately in the browser without refresh.

### Linting

Check code quality:

```bash
npm run lint
```

Fix auto-fixable issues:

```bash
npm run lint -- --fix
```

### Building for Production

```bash
# Create optimized production build
npm run build

# Output will be in dist/ directory
ls -la dist/

# Preview production build locally
npm run preview
```

## Project Structure Overview

```
apps/dashboard/
├── api/                    # PHP Backend
│   ├── apps.php           # REST API endpoint
│   └── lib/
│       └── detection.php  # Detection logic
├── src/                   # React Frontend
│   ├── components/        # UI components (12 files)
│   ├── contexts/         # Theme context
│   ├── hooks/            # Custom hooks (3 files)
│   ├── utils/            # Utilities (2 files)
│   ├── App.jsx           # Main component
│   └── main.jsx          # Entry point
├── package.json          # Dependencies
├── vite.config.js        # Vite config
└── tailwind.config.js    # Tailwind config
```

## Next Steps

1. **Explore the UI**: Try search, filters, sorting, and dark mode
2. **Check Auto-refresh**: Wait 30 seconds to see data update
3. **Click App Cards**: Open detail modals to see full info
4. **Responsive Testing**: Resize browser to see mobile layout
5. **Customize**: Edit `tailwind.config.js` to change colors/theme

## Deployment

### Static Build Deployment

1. Build the production bundle:
   ```bash
   npm run build
   ```

2. Serve the `dist/` directory with any static file server

3. Ensure API endpoint is accessible at `/api/apps.php`

### Nginx Proxy Manager Setup

Create a proxy host for `dashboard.test`:

1. **Proxy to React dev server**:
   - Forward to: `http://node:5173`
   - Enable websockets for HMR

2. **Or proxy to production build**:
   - Serve static files from `dist/`
   - Proxy `/api/*` to `http://php:8100/api/`

Example Nginx config:

```nginx
location / {
    root /var/www/html/dashboard/dist;
    try_files $uri $uri/ /index.html;
}

location /api/ {
    proxy_pass http://php:8100/api/;
    proxy_set_header Host $host;
}
```

## Useful Commands

```bash
# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linter
npm run lint

# Test API
curl http://localhost:8100/api/apps.php | jq

# Clear node_modules
rm -rf node_modules package-lock.json && npm install

# Watch logs (if running in container)
podman logs -f node
```

## Getting Help

1. Check the main [README.md](./README.md) for detailed documentation
2. Review [CLAUDE.md](/home/fhcadmin/projects/devarch/CLAUDE.md) for DevArch architecture
3. Check browser console (F12) for JavaScript errors
4. Check network tab for API request/response details
5. Verify PHP container logs: `podman logs php`

## Success Checklist

- [ ] Node.js 18+ installed
- [ ] PHP container running
- [ ] Dependencies installed (`npm install`)
- [ ] API responds with JSON data
- [ ] Dev server starts without errors
- [ ] Dashboard loads in browser
- [ ] Apps are detected and displayed
- [ ] Search, filter, and sort work
- [ ] Dark mode toggles successfully
- [ ] Detail modals open when clicking apps
- [ ] Auto-refresh updates data every 30s

Once all items are checked, you're ready to use the dashboard!
