# DevArch Application Migration Guide

**Version**: 1.0.0
**Last Updated**: 2025-12-03

This guide helps you migrate existing applications to DevArch's standardized `public/` directory structure.

## Table of Contents

- [Overview](#overview)
- [Before You Begin](#before-you-begin)
- [Migration Methods](#migration-methods)
- [Framework-Specific Migrations](#framework-specific-migrations)
- [Post-Migration Verification](#post-migration-verification)
- [Troubleshooting](#troubleshooting)

## Overview

DevArch requires all applications to serve content from a `public/` subdirectory. If your application:

- Builds to `dist/`, `.next/`, `build/`, or another directory
- Serves static files from the root directory
- Uses a different document root structure

You'll need to migrate it to use `public/` as the web server document root.

## Before You Begin

### Backup Your Application

```bash
# Create backup
cp -r apps/my-app apps/my-app.backup

# Or use git
cd apps/my-app
git add .
git commit -m "Backup before migration to public/ structure"
```

### Check Current Structure

```bash
# Identify current build output
cd apps/my-app
npm run build  # or equivalent

# Where did it output?
ls -la dist/     # React/Vue default
ls -la .next/    # Next.js default
ls -la build/    # Create React App default
ls -la out/      # Next.js export default
```

### Understand Your Framework

Determine:
- **Framework**: React, Next.js, Express, Django, etc.
- **Build tool**: Vite, Webpack, Next.js built-in, etc.
- **Current output directory**: Where builds currently go
- **Static file serving**: How static files are currently served

## Migration Methods

### Method 1: Automated Migration Script (Recommended)

```bash
# Run migration script
./scripts/migrate-app-structure.sh apps/my-app

# Follow prompts
# Script will:
# 1. Detect framework
# 2. Backup current configuration
# 3. Update build configs
# 4. Update static file serving
# 5. Rebuild application
# 6. Verify structure
```

### Method 2: Manual Migration

Follow framework-specific instructions below.

## Framework-Specific Migrations

### React + Vite

**Current Issue**: Vite builds to `dist/` by default.

**Migration Steps**:

1. **Locate vite.config.js**:
   ```bash
   cd apps/my-app
   ```

2. **Update vite.config.js**:
   ```javascript
   import { defineConfig } from 'vite'
   import react from '@vitejs/plugin-react'

   export default defineConfig({
     plugins: [react()],
     publicDir: 'static', // Rename from 'public' to avoid conflict
     build: {
       outDir: 'public',        // ← CHANGE FROM 'dist'
       emptyOutDir: false,      // Don't delete API folder if present
       sourcemap: true,
     }
   })
   ```

3. **Update .gitignore**:
   ```diff
   - /dist
   + /public/*.html
   + /public/*.js
   + /public/*.css
   + /public/assets/*.js
   + /public/assets/*.css
   + !public/api/
   ```

4. **Rename public/ to static/** (if it exists):
   ```bash
   # Vite uses 'public' for static assets by default
   # We need that for build output, so rename it
   if [ -d "public" ]; then
       mv public static
   fi
   ```

5. **Update index.html** references (if needed):
   ```html
   <!-- Update asset paths if using relative references -->
   ```

6. **Clean and rebuild**:
   ```bash
   rm -rf dist
   npm run build
   ```

7. **Verify**:
   ```bash
   ls -la public/
   # Should see index.html and assets/
   ```

**Common Issues**:

- **"public/ directory conflicts"**: Rename Vite's `public` to `static` and set `publicDir: 'static'`
- **Assets not found**: Check `publicDir` configuration
- **Build errors**: Clear `node_modules` and reinstall

### Next.js

**Current Issue**: Next.js builds to `.next/` or exports to `out/`.

**Migration Steps**:

1. **Update next.config.js**:
   ```javascript
   /** @type {import('next').NextConfig} */
   const nextConfig = {
     distDir: 'public/.next',   // ← BUILD DIRECTORY
     output: 'export',          // ← STATIC EXPORT MODE
     images: {
       unoptimized: true,       // Required for static export
     },
     trailingSlash: true,       // Better for static hosting
   }

   module.exports = nextConfig
   ```

2. **Alternative for SSR** (if not using static export):
   ```javascript
   const nextConfig = {
     distDir: 'public/.next',   // Build to public/
     output: 'standalone',      // For SSR mode
   }
   ```

3. **Update .gitignore**:
   ```diff
   - /.next
   - /out
   + /public/.next
   + /public/*.html
   + /public/_next/
   ```

4. **Clean and rebuild**:
   ```bash
   rm -rf .next out
   npm run build
   ```

5. **Verify**:
   ```bash
   ls -la public/
   # Static export: Should see *.html files
   # Standalone: Should see .next/ directory
   ```

**Common Issues**:

- **"Image optimization error"**: Set `images.unoptimized: true` for static export
- **"404 on navigation"**: Set `trailingSlash: true`
- **"Dynamic routes fail"**: Use `getStaticPaths` for static export or switch to standalone mode

### Express.js

**Current Issue**: Express may serve static files from root or `dist/`.

**Migration Steps**:

1. **Update server configuration**:
   ```javascript
   // server.js or app.js
   import express from 'express'
   import path from 'path'
   import { fileURLToPath } from 'url'

   const __dirname = path.dirname(fileURLToPath(import.meta.url))
   const app = express()

   // BEFORE:
   // app.use(express.static('dist'))
   // app.use(express.static(__dirname))

   // AFTER:
   app.use(express.static(path.join(__dirname, 'public')))

   // API routes
   app.use('/api', apiRoutes)

   // SPA fallback
   app.get('*', (req, res) => {
     res.sendFile(path.join(__dirname, 'public/index.html'))
   })
   ```

2. **If you have a separate frontend build**:
   ```bash
   # Update frontend build config (e.g., Vite) to output to public/
   # See React + Vite section above
   ```

3. **Move existing static files**:
   ```bash
   mkdir -p public
   # If static files were in dist/
   mv dist/* public/ 2>/dev/null || true
   # If static files were in root
   mv *.html public/ 2>/dev/null || true
   ```

4. **Update package.json** scripts if needed:
   ```json
   {
     "scripts": {
       "build": "vite build",  // Should output to public/
       "start": "node server.js"
     }
   }
   ```

5. **Test**:
   ```bash
   npm start
   curl http://localhost:8200
   ```

**Common Issues**:

- **"Static files 404"**: Check `express.static()` path is correct
- **"API routes not working"**: Ensure API routes are defined before SPA fallback (`app.get('*')`)
- **"SPA routing broken"**: Check SPA fallback sends `index.html` for all non-API routes

### Create React App

**Current Issue**: CRA builds to `build/` directory.

**Migration Steps**:

1. **Option A: Eject and configure Webpack** (Not recommended):
   ```bash
   npm run eject
   # Then modify webpack config...
   ```

2. **Option B: Use build script to move files** (Recommended):
   ```bash
   # Add post-build script to package.json
   ```

   ```json
   {
     "scripts": {
       "build": "react-scripts build && npm run move-to-public",
       "move-to-public": "rm -rf public && mv build public"
     }
   }
   ```

3. **Option C: Migrate to Vite** (Best long-term):
   ```bash
   # Follow Vite migration guide
   # Then use Vite config from React + Vite section
   ```

4. **Update .gitignore**:
   ```diff
   - /build
   + /public
   ```

5. **Build**:
   ```bash
   npm run build
   ```

**Recommendation**: Consider migrating from CRA to Vite for better performance and simpler configuration.

### Django

**Current Issue**: Static files in multiple locations, not centralized in `public/`.

**Migration Steps**:

1. **Update settings.py**:
   ```python
   import os
   from pathlib import Path

   BASE_DIR = Path(__file__).resolve().parent.parent

   # Static files (CSS, JavaScript, Images)
   STATIC_URL = '/static/'
   STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')  # ← CHANGE

   # Additional locations of static files
   STATICFILES_DIRS = [
       os.path.join(BASE_DIR, 'static'),  # Your source static files
   ]

   # Media files (user uploads)
   MEDIA_URL = '/media/'
   MEDIA_ROOT = os.path.join(BASE_DIR, 'public/media')  # ← CHANGE
   ```

2. **Create directory structure**:
   ```bash
   mkdir -p public/static
   mkdir -p public/media
   mkdir -p static  # Source static files
   ```

3. **Collect static files**:
   ```bash
   python manage.py collectstatic --noinput
   # This copies all static files to public/static/
   ```

4. **Update .gitignore**:
   ```diff
   + /public/static/
   + /public/media/
   - /staticfiles/
   ```

5. **For development with Django dev server**:
   ```python
   # urls.py
   from django.conf import settings
   from django.conf.urls.static import static

   urlpatterns = [
       # Your URL patterns
   ]

   if settings.DEBUG:
       urlpatterns += static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
       urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
   ```

6. **Test**:
   ```bash
   python manage.py runserver 0.0.0.0:8300
   ```

**Common Issues**:

- **"Static files not loading"**: Run `collectstatic` after any static file changes
- **"Admin styles missing"**: Ensure `django.contrib.staticfiles` is in `INSTALLED_APPS`
- **"Media files 404"**: Check `MEDIA_ROOT` and `MEDIA_URL` settings

### Flask

**Current Issue**: Flask may serve static files from default `static/` folder.

**Migration Steps**:

1. **Update Flask app configuration**:
   ```python
   from flask import Flask
   import os

   # BEFORE:
   # app = Flask(__name__)

   # AFTER:
   app = Flask(__name__,
               static_folder='public',
               static_url_path='')

   # OR if you need /static prefix:
   # app = Flask(__name__,
   #             static_folder='public',
   #             static_url_path='/static')

   @app.route('/')
   def index():
       return app.send_static_file('index.html')

   # API routes
   @app.route('/api/health')
   def health():
       return {'status': 'ok'}

   if __name__ == '__main__':
       app.run(host='0.0.0.0', port=8300)
   ```

2. **Move static files**:
   ```bash
   mkdir -p public
   # Move existing static files
   mv static/* public/ 2>/dev/null || true
   ```

3. **If using templates**, keep them separate:
   ```python
   app = Flask(__name__,
               static_folder='public',
               static_url_path='',
               template_folder='templates')
   ```

4. **Test**:
   ```bash
   python app.py
   curl http://localhost:8300
   ```

**Common Issues**:

- **"Static files 404"**: Check `static_folder` path is correct
- **"Templates not found"**: Ensure `template_folder` is set if using templates
- **"URL conflicts"**: Use `static_url_path` to namespace static files

### Laravel

**Good News**: Laravel **already uses** `public/` as document root!

**No migration needed** - Laravel is natively compatible.

**Verification**:
```bash
cd apps/my-laravel-app
ls -la public/
# Should see index.php, .htaccess, etc.
```

### WordPress

**Migration Steps**:

1. **Install WordPress in public/ subdirectory**:
   ```bash
   cd apps/my-wordpress-app
   mkdir -p public
   cd public

   # Download WordPress
   wget https://wordpress.org/latest.tar.gz
   tar -xzf latest.tar.gz --strip-components=1
   rm latest.tar.gz
   ```

2. **Or use install script**:
   ```bash
   ./scripts/wordpress/install-wordpress.sh my-wordpress-app
   ```

3. **Configure wp-config.php**:
   ```php
   // Database settings should point to DevArch databases
   define('DB_NAME', 'wordpress_myapp');
   define('DB_USER', 'devarch_user');
   define('DB_PASSWORD', 'devarch_pass');
   define('DB_HOST', 'mariadb');
   ```

4. **Verify structure**:
   ```bash
   ls -la apps/my-wordpress-app/public/
   # Should see WordPress files
   ```

### Go (Gin/Echo)

**Migration Steps**:

1. **Update static file serving**:

   **Gin**:
   ```go
   package main

   import "github.com/gin-gonic/gin"

   func main() {
       r := gin.Default()

       // Serve static files from public/
       r.Static("/assets", "./public/assets")
       r.StaticFile("/", "./public/index.html")
       r.StaticFile("/favicon.ico", "./public/favicon.ico")

       // API routes
       api := r.Group("/api")
       {
           api.GET("/health", healthHandler)
       }

       r.Run(":8400")
   }
   ```

   **Echo**:
   ```go
   package main

   import (
       "github.com/labstack/echo/v4"
       "github.com/labstack/echo/v4/middleware"
   )

   func main() {
       e := echo.New()
       e.Use(middleware.Logger())
       e.Use(middleware.Recover())

       // Serve static files from public/
       e.Static("/assets", "public/assets")
       e.File("/", "public/index.html")
       e.File("/favicon.ico", "public/favicon.ico")

       // API routes
       e.GET("/api/health", healthHandler)

       e.Start(":8400")
   }
   ```

2. **Create public/ structure**:
   ```bash
   mkdir -p public/assets
   touch public/index.html
   ```

3. **Build and test**:
   ```bash
   go run main.go
   curl http://localhost:8400
   ```

## Post-Migration Verification

After migrating, verify everything works:

### 1. Directory Structure

```bash
cd apps/my-app

# Verify public/ exists
ls -la public/

# Check for entry point
ls public/index.html || ls public/index.php

# Check for assets
ls public/assets/
```

### 2. Build Process

```bash
# Clean previous builds
rm -rf dist .next build out

# Run build
npm run build  # or equivalent

# Verify output location
ls -la public/
# Should contain built files
```

### 3. Local Testing

```bash
# Start dev server or backend runtime
npm run dev  # or npm start, python app.py, etc.

# Test access
curl http://localhost:{port}/

# Check specific assets
curl http://localhost:{port}/assets/css/app.css
```

### 4. Integration Testing

```bash
# Ensure backend runtime is running
./scripts/service-manager.sh status backend

# Access via domain (after configuring Nginx Proxy Manager)
curl https://my-app.test

# Or test HTTP
curl http://my-app.test
```

### 5. Asset Loading

Open browser developer tools and check:
- No 404 errors for CSS/JS/images
- Assets load from correct paths
- No CORS errors

## Troubleshooting

### Build Still Goes to dist/ or .next/

**Cause**: Build configuration not updated properly.

**Solution**:
1. Double-check build config file (`vite.config.js`, `next.config.js`, etc.)
2. Clear cache: `rm -rf node_modules/.vite`, `rm -rf .next`
3. Rebuild: `npm run build`
4. Verify: `ls -la public/`

### Public/ Directory Conflicts

**Cause**: Framework uses `public/` for source static files (like Vite).

**Solution**:
1. Rename source `public/` to `static/`
2. Update build config:
   ```javascript
   // vite.config.js
   publicDir: 'static',  // Source static files
   build: {
     outDir: 'public',   // Build output
   }
   ```

### Static Files Return 404

**Cause**: Web server or application not configured to serve from `public/`.

**Solution**:

1. **For development servers**: Check config serves from `public/`
2. **For production**: Verify Nginx Proxy Manager points to correct container
3. **For Express**: Check `express.static()` path
4. **For Flask**: Check `static_folder` parameter

### Application Shows Old Version

**Cause**: Cached build files or browser cache.

**Solution**:
```bash
# Clear build
rm -rf public/*.html public/*.js public/*.css public/assets

# Rebuild
npm run build

# Clear browser cache
# Or open in incognito/private window
```

### Permission Errors

**Cause**: File permissions on `public/` directory.

**Solution**:
```bash
# Fix permissions
chmod -R 755 apps/my-app/public/

# For PHP apps in container
podman exec php chown -R www-data:www-data /var/www/html/my-app/public
```

## Rollback Procedure

If migration fails and you need to rollback:

### From Backup

```bash
# Remove failed migration
rm -rf apps/my-app

# Restore from backup
cp -r apps/my-app.backup apps/my-app
```

### From Git

```bash
cd apps/my-app

# Check what changed
git status
git diff

# Rollback all changes
git checkout .

# Or rollback specific files
git checkout vite.config.js .gitignore
```

## Getting Help

If you encounter issues not covered in this guide:

1. **Check framework-specific documentation**:
   - `templates/{category}/{framework}/README.md`

2. **Review app structure standard**:
   - `APP_STRUCTURE.md`

3. **Check DevArch architecture**:
   - `CLAUDE.md`

4. **Test with template**:
   - Create new app from template to see working config
   - `./scripts/create-app.sh --list`
   - Compare your config with template config

## Success Checklist

After migration, verify:

- [ ] `public/` directory exists
- [ ] Entry point exists (`index.html` or `index.php`)
- [ ] Assets directory exists (`public/assets/`)
- [ ] Build configuration updated
- [ ] Build outputs to `public/`
- [ ] `.gitignore` updated
- [ ] Static file serving configured
- [ ] Application runs in development
- [ ] Application builds successfully
- [ ] Built files appear in `public/`
- [ ] Assets load without 404 errors
- [ ] Application works via Nginx Proxy Manager
- [ ] Domain resolves correctly (e.g., `my-app.test`)

## Additional Resources

- **App Structure Standard**: `APP_STRUCTURE.md`
- **Templates Documentation**: `templates/README.md`
- **DevArch Architecture**: `CLAUDE.md`
- **Framework Templates**: `templates/{category}/{framework}/`

## Version History

- **1.0.0** (2025-12-03): Initial migration guide
