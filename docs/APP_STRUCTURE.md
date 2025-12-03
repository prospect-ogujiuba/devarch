# DevArch Application Structure Standard

**Version**: 1.0.0
**Last Updated**: 2025-12-03

This document defines the standardized application structure for all applications in the DevArch development environment.

## Table of Contents

- [Overview](#overview)
- [Why the public/ Standard](#why-the-public-standard)
- [Standard Directory Structure](#standard-directory-structure)
- [Framework-Specific Configurations](#framework-specific-configurations)
- [Build Process Requirements](#build-process-requirements)
- [Port Assignment Standards](#port-assignment-standards)
- [Creating New Applications](#creating-new-applications)
- [Migrating Existing Applications](#migrating-existing-applications)

## Overview

DevArch requires **all applications** to follow a standardized `public/` directory pattern. This is a fundamental requirement of the web server configuration and ensures consistent behavior across all framework types.

### Critical Requirement

**The `public/` subdirectory is the web server document root for all applications.**

```
apps/{app-name}/public/  ← WEB SERVER SERVES FROM HERE
```

Failure to follow this structure will result in the application not being accessible through the web server.

## Why the public/ Standard

### Technical Reasons

1. **Web Server Configuration**: Nginx and Apache are configured to serve from `/var/www/html/{app-name}/public/`
2. **Security**: Source code outside `public/` is not directly accessible via HTTP
3. **Container Mounts**: Host directories are mounted to containers expecting `public/` as document root
4. **Consistency**: All frameworks work the same way, simplifying deployment and troubleshooting

### Container Architecture

```
Host                          Container
─────────────────────────────────────────────────────────────
/home/fhcadmin/projects/      /var/www/html/{app}/
devarch/apps/{app}/public/    public/
                              ↑
                              Web server DocumentRoot
```

### Framework Compatibility

| Framework | Native Structure | DevArch Compatible |
|-----------|------------------|-------------------|
| Laravel   | Has `public/`    | ✅ Yes (native)   |
| WordPress | Installed in root | ✅ Yes (with setup) |
| Next.js   | Builds to `.next/` | ❌ No (needs config) |
| React/Vite | Builds to `dist/` | ❌ No (needs config) |
| Express   | Serves from root | ❌ No (needs config) |
| Django    | Static in various dirs | ❌ No (needs config) |

**Solution**: Configure all frameworks to build/serve from `public/` directory.

## Standard Directory Structure

Every application MUST follow this structure:

```
apps/{app-name}/
├── public/                  # MANDATORY - Web server document root
│   ├── index.html          # Entry point (static/SPA apps)
│   ├── index.php           # Entry point (PHP apps)
│   ├── assets/             # Built assets (CSS, JS, images)
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   ├── api/                # API endpoints (optional)
│   └── .htaccess           # Server configuration (optional)
│
├── src/                    # Source code (for compiled apps)
│   ├── components/         # UI components
│   ├── pages/              # Page components/templates
│   ├── utils/              # Utility functions
│   ├── styles/             # Source styles
│   └── ...
│
├── config/                 # Application configuration
│   ├── database.php        # Database config
│   ├── services.php        # Service config
│   └── ...
│
├── scripts/                # Application-specific scripts
│   ├── build.sh           # Build script
│   ├── deploy.sh          # Deployment script
│   └── ...
│
├── tests/                  # Test files
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── docs/                   # Application documentation
│
├── .env.example            # Environment variables template
├── .gitignore              # Git ignore patterns
├── package.json            # Node.js dependencies (if Node app)
├── composer.json           # PHP dependencies (if PHP app)
├── requirements.txt        # Python dependencies (if Python app)
├── go.mod                  # Go dependencies (if Go app)
├── {app-name}.csproj       # .NET project file (if .NET app)
│
├── README.md               # Application documentation
└── [build-config]          # Build configuration files
    ├── vite.config.js      # Vite config
    ├── next.config.js      # Next.js config
    ├── webpack.config.js   # Webpack config
    └── ...
```

### public/ Directory Requirements

The `public/` directory **MUST** contain:

1. **Entry Point**:
   - `index.html` for static/SPA apps
   - `index.php` for PHP apps
   - Built application files for framework apps

2. **Assets**: All static assets that should be served
   - CSS files (compiled)
   - JavaScript files (compiled/bundled)
   - Images, fonts, icons
   - Media files

3. **Optional**:
   - `.htaccess` for Apache/Nginx configuration
   - `robots.txt` for SEO
   - `sitemap.xml` for SEO
   - API endpoints (if not using separate backend)

### What Goes OUTSIDE public/

- **Source Code**: Original, uncompiled source files
- **Configuration**: Framework and build configurations
- **Dependencies**: node_modules, vendor, etc.
- **Tests**: Test files and fixtures
- **Documentation**: Development documentation
- **Build Tools**: Build scripts, tooling configs

## Framework-Specific Configurations

### React + Vite

**vite.config.js**:
```javascript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  publicDir: 'static', // Static assets source
  build: {
    outDir: 'public',        // ← BUILD TO public/
    emptyOutDir: false,      // Don't delete existing files
    sourcemap: true,
  }
})
```

**package.json**:
```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",   // Builds to public/
    "preview": "vite preview"
  }
}
```

### Next.js

**next.config.js**:
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  distDir: 'public/.next',   // ← BUILD TO public/
  output: 'export',          // Static export mode
  images: {
    unoptimized: true,       // Required for static export
  },
  trailingSlash: true,
}

module.exports = nextConfig
```

**package.json**:
```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",    // Builds to public/.next
    "start": "next start"
  }
}
```

### Express.js

**server.js**:
```javascript
import express from 'express'
import path from 'path'

const app = express()

// Serve static files from public/
app.use(express.static(path.join(__dirname, 'public')))

// API routes
app.use('/api', apiRoutes)

// SPA fallback
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'public/index.html'))
})

app.listen(8200, '0.0.0.0')
```

### Laravel

Laravel **already** uses `public/` as document root. No configuration needed.

**Directory Structure**:
```
laravel-app/
├── public/              # ✅ Already correct
│   └── index.php
├── app/
├── config/
└── ...
```

### WordPress

WordPress should be **installed inside** `public/` directory.

**Structure**:
```
wordpress-app/
└── public/              # WordPress installation
    ├── wp-admin/
    ├── wp-content/
    ├── wp-includes/
    ├── index.php
    └── wp-config.php
```

**Installation**:
```bash
cd apps/my-wordpress-app
mkdir -p public
cd public
wget https://wordpress.org/latest.tar.gz
tar -xzf latest.tar.gz --strip-components=1
rm latest.tar.gz
```

### Django

**settings.py**:
```python
import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

# Static files configuration
STATIC_URL = '/static/'
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')  # ← COLLECT TO public/
STATICFILES_DIRS = [
    os.path.join(BASE_DIR, 'static'),
]

# Media files
MEDIA_URL = '/media/'
MEDIA_ROOT = os.path.join(BASE_DIR, 'public/media')
```

**Collect Static Files**:
```bash
python manage.py collectstatic --noinput
# Outputs to public/static/
```

### Flask

**app.py**:
```python
from flask import Flask

# Configure Flask to use public/ directory
app = Flask(__name__,
            static_folder='public',      # ← SERVE FROM public/
            static_url_path='')

@app.route('/')
def index():
    return app.send_static_file('index.html')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8300)
```

### Go (Gin)

**main.go**:
```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Serve static files from public/
    r.Static("/assets", "./public/assets")
    r.StaticFile("/", "./public/index.html")

    // API routes
    api := r.Group("/api")
    {
        api.GET("/health", healthHandler)
    }

    r.Run(":8400")
}
```

### ASP.NET Core

**Program.cs**:
```csharp
var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

// Serve static files from wwwroot (rename to public)
app.UseStaticFiles(new StaticFileOptions
{
    FileProvider = new PhysicalFileProvider(
        Path.Combine(builder.Environment.ContentRootPath, "public")),
    RequestPath = ""
});

app.Run();
```

**Or rename wwwroot to public** in project file:

**.csproj**:
```xml
<PropertyGroup>
  <TargetFramework>net8.0</TargetFramework>
  <DefaultItemExcludes>$(DefaultItemExcludes);public\**</DefaultItemExcludes>
</PropertyGroup>

<ItemGroup>
  <Content Include="public\**" CopyToOutputDirectory="PreserveNewest" />
</ItemGroup>
```

## Build Process Requirements

### Build Output Location

**All build processes MUST output to the `public/` directory.**

### Build Script Standard

Every application should include a build script:

**scripts/build.sh**:
```bash
#!/usr/bin/env bash
set -euo pipefail

echo "Building application..."

# Clean previous build (optional)
# rm -rf public/assets

# Run build process
npm run build  # or appropriate build command

# Verify output
if [[ ! -d "public" ]] || [[ -z "$(ls -A public)" ]]; then
    echo "ERROR: Build failed - public/ directory is empty"
    exit 1
fi

echo "Build complete! Assets in public/"
```

### Verification

After building, verify:

```bash
# Check public/ exists
ls -la public/

# Check for entry point
ls public/index.html  # or index.php

# Check for assets
ls public/assets/

# Test file serving (if dev server running)
curl http://localhost:{port}/
```

## Port Assignment Standards

DevArch uses **dedicated 100-port ranges** for each runtime to eliminate conflicts.

| Runtime | Port Range | Default | Container | Example Ports |
|---------|------------|---------|-----------|---------------|
| PHP | 8100-8199 | 8100 | php | 8100 (Laravel), 8101 (WordPress) |
| Node.js | 8200-8299 | 8200 | nodejs | 8200 (React), 8201 (Next.js), 8202 (Express) |
| Python | 8300-8399 | 8300 | python | 8300 (Django), 8301 (Flask), 8302 (FastAPI) |
| Go | 8400-8499 | 8400 | golang | 8400 (Gin), 8401 (Echo) |
| .NET | 8600-8699 | 8600 | dotnet | 8600 (Web API), 8601 (MVC) |

### Port Selection

When creating a new app:

1. **Choose runtime** based on framework
2. **Select available port** in runtime's range
3. **Configure application** to use that port
4. **Update environment variables**:
   ```env
   PORT=8200
   VITE_PORT=8200
   ```

### Port Conflicts

If you encounter "port already in use":

```bash
# Check what's using the port
lsof -i :8200

# Or with netstat
netstat -tulpn | grep 8200

# Choose different port in same range
PORT=8201 npm run dev
```

## Creating New Applications

### Using Templates (Recommended)

1. **List available templates**:
   ```bash
   ./scripts/create-app.sh --list
   ```

2. **Create app interactively**:
   ```bash
   ./scripts/create-app.sh
   ```

3. **Create app non-interactively**:
   ```bash
   ./scripts/create-app.sh --name my-app --template node/react-vite --port 8200
   ```

4. **Navigate and build**:
   ```bash
   cd apps/my-app
   npm install
   npm run build
   ```

5. **Verify structure**:
   ```bash
   ls -la public/  # Should contain built files
   ```

### Manual Creation

1. **Create directory**:
   ```bash
   mkdir -p apps/my-app/public
   cd apps/my-app
   ```

2. **Initialize project**:
   ```bash
   npm init -y  # or composer init, etc.
   ```

3. **Configure build to output to public/**:
   - Edit `vite.config.js`, `next.config.js`, etc.
   - Set `outDir: 'public'` or equivalent

4. **Create entry point**:
   ```bash
   touch public/index.html  # or index.php
   ```

5. **Add .env.example**:
   ```bash
   cat > .env.example <<EOF
   APP_NAME=my-app
   PORT=8200
   EOF
   ```

6. **Test build**:
   ```bash
   npm install
   npm run build
   ls public/  # Verify output
   ```

## Migrating Existing Applications

If you have an existing app that doesn't follow the `public/` standard:

### Quick Migration

```bash
./scripts/migrate-app-structure.sh apps/my-app
```

### Manual Migration

#### For Apps Building to dist/

1. **Update build config**:
   ```javascript
   // vite.config.js
   build: {
     outDir: 'public',  // Changed from 'dist'
   }
   ```

2. **Rebuild**:
   ```bash
   rm -rf dist
   npm run build
   ls public/  # Verify
   ```

3. **Update .gitignore**:
   ```
   # Old
   /dist

   # New
   /public/*.html
   /public/assets
   !public/api
   ```

#### For Next.js Apps

1. **Update next.config.js**:
   ```javascript
   distDir: 'public/.next',
   output: 'export',
   ```

2. **Rebuild**:
   ```bash
   rm -rf .next
   npm run build
   ```

3. **Verify**:
   ```bash
   ls public/.next
   ```

#### For Express Apps

1. **Update static file serving**:
   ```javascript
   // Old
   app.use(express.static('dist'))

   // New
   app.use(express.static('public'))
   ```

2. **Move static files**:
   ```bash
   mv dist public
   ```

#### For Django Apps

1. **Update settings.py**:
   ```python
   STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')
   ```

2. **Collect static**:
   ```bash
   python manage.py collectstatic --noinput
   ```

### Migration Checklist

- [ ] Update build configuration
- [ ] Update static file serving (if applicable)
- [ ] Rebuild application
- [ ] Verify `public/` contains all necessary files
- [ ] Update `.gitignore`
- [ ] Update documentation
- [ ] Test application loads correctly
- [ ] Update CI/CD scripts if needed

## Troubleshooting

### Application Not Loading

**Symptom**: Accessing `https://my-app.test` returns 404 or blank page.

**Checks**:

1. **Verify public/ exists**:
   ```bash
   ls -la apps/my-app/public/
   ```

2. **Check for entry point**:
   ```bash
   ls apps/my-app/public/index.html
   # or
   ls apps/my-app/public/index.php
   ```

3. **Verify build process**:
   ```bash
   cd apps/my-app
   npm run build  # or equivalent
   ls public/  # Should show files
   ```

4. **Check web server config**:
   - Nginx Proxy Manager should point to correct container and port
   - DocumentRoot should be `/var/www/html/my-app/public`

### Assets Not Loading

**Symptom**: HTML loads but CSS/JS/images don't load (404).

**Checks**:

1. **Verify assets exist**:
   ```bash
   ls apps/my-app/public/assets/
   ```

2. **Check asset paths** in HTML:
   ```html
   <!-- Absolute paths (recommended) -->
   <link href="/assets/css/app.css" rel="stylesheet">

   <!-- NOT relative paths -->
   <link href="./assets/css/app.css" rel="stylesheet">
   ```

3. **Check build output**:
   - Build process should copy/compile assets to `public/assets/`

### Build Outputs to Wrong Directory

**Symptom**: Build process creates `dist/` or `.next/` instead of `public/`.

**Solution**:

1. **Update build config**:
   - Vite: `build.outDir: 'public'`
   - Next.js: `distDir: 'public/.next'`
   - Webpack: `output.path: 'public'`

2. **Rebuild**:
   ```bash
   rm -rf dist .next
   npm run build
   ```

## Best Practices

### Development Workflow

1. **Start with template**:
   ```bash
   ./scripts/create-app.sh
   ```

2. **Develop in `src/`**, not `public/`

3. **Build frequently** to test integration:
   ```bash
   npm run build && curl http://localhost:8200
   ```

4. **Use dev server** for hot reload:
   ```bash
   npm run dev
   ```

5. **Build for production**:
   ```bash
   npm run build
   ```

### Git Practices

**DO** commit:
- Source code (`src/`)
- Build configurations
- `.env.example`
- Documentation

**DON'T** commit:
- Built files in `public/` (except static assets)
- `node_modules/`, `vendor/`
- `.env` (actual secrets)

**.gitignore** example:

```gitignore
# Build outputs
public/*.html
public/*.js
public/*.css
public/assets/*.js
public/assets/*.css

# Keep static files
!public/api/
!public/robots.txt
!public/favicon.ico

# Dependencies
node_modules/
vendor/

# Environment
../.env
.env.local
```

### Testing

Test your app works with `public/` structure:

```bash
# Build
npm run build

# Verify public/ has files
ls -la public/

# Start backend runtime
./scripts/service-manager.sh start backend

# Test direct access
curl http://localhost:8200

# Test via proxy (after configuring)
curl https://my-app.test
```

## Additional Resources

- **Templates**: See `templates/README.md` for available templates
- **Migration Guide**: See `MIGRATION_GUIDE.md` for detailed migration steps
- **DevArch Architecture**: See `CLAUDE.md` for overall system architecture
- **Troubleshooting**: See specific framework documentation in `templates/{category}/{framework}/`

## Version History

- **1.0.0** (2025-12-03): Initial standardization document

## Support

For issues or questions:
1. Check this document for framework-specific configuration
2. Review template README: `templates/{category}/{framework}/README.md`
3. Check migration guide: `MIGRATION_GUIDE.md`
4. Review DevArch architecture: `CLAUDE.md`
