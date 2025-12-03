# DevArch Application Templates

This directory contains standardized templates for creating new applications in DevArch. All templates follow the **mandatory `public/` directory pattern** required by the web server configuration.

## Why Templates?

DevArch's web server expects all applications to serve content from a `public/` subdirectory:
- **Host path**: `/home/fhcadmin/projects/devarch/apps/{app-name}/public/`
- **Container mount**: `/var/www/html/{app-name}/public/`
- **Web server DocumentRoot**: `/var/www/html/{app-name}/public/`

Without proper `public/` directory structure, applications will fail to serve correctly through the web server.

## Available Templates

### PHP Applications
- **laravel/** - Laravel framework with standard `public/index.php` structure
- **wordpress/** - WordPress with installation in `public/` subdirectory
- **vanilla/** - Plain PHP application with `public/` web root

### Node.js Applications
- **nextjs/** - Next.js with static export or standalone mode building to `public/`
- **react-vite/** - React + Vite SPA with build output to `public/`
- **express/** - Express.js server with static file serving from `public/`
- **vue/** - Vue.js SPA with build output to `public/`

### Python Applications
- **django/** - Django framework with static files in `public/static/`
- **flask/** - Flask application with `public/` as static folder
- **fastapi/** - FastAPI with static file mounting from `public/`

### Go Applications
- **gin/** - Gin framework with static assets in `public/`
- **echo/** - Echo framework with static file serving

### .NET Applications
- **aspnet-core/** - ASP.NET Core with wwwroot mapped to `public/`

## Using Templates

### Quick Start with create-app.sh

```bash
# Interactive mode (recommended)
./scripts/create-app.sh

# Non-interactive mode
./scripts/create-app.sh --name my-app --template node/react-vite --port 8200

# List available templates
./scripts/create-app.sh --list
```

### Manual Template Usage

1. **Copy template to apps directory**:
   ```bash
   cp -r templates/node/react-vite apps/my-new-app
   cd apps/my-new-app
   ```

2. **Customize configuration**:
   - Update `package.json` (or `composer.json`, `requirements.txt`, etc.)
   - Copy `.env.example` to `.env` and configure
   - Update app name and description in README.md

3. **Install dependencies**:
   ```bash
   npm install  # or composer install, pip install -r requirements.txt, etc.
   ```

4. **Build the application**:
   ```bash
   npm run build  # This MUST output to public/ directory
   ```

5. **Configure Nginx Proxy Manager**:
   - Add proxy host for `my-new-app.test`
   - Point to backend runtime (e.g., http://nodejs:8200 for Node apps)
   - Configure SSL certificate

## Template Structure Standard

Every template follows this structure:

```
template-name/
├── public/              # MANDATORY - Web server document root
│   ├── index.html       # Entry point (for static/SPA apps)
│   ├── index.php        # Entry point (for PHP apps)
│   ├── assets/          # Built assets (CSS, JS, images)
│   ├── api/             # API endpoints (optional)
│   └── .htaccess        # Apache/Nginx config (optional)
├── src/                 # Source code (for compiled apps)
│   ├── components/      # UI components
│   ├── pages/           # Page components
│   ├── utils/           # Utility functions
│   └── ...
├── config/              # Application configuration
├── scripts/             # Build and deployment scripts
├── tests/               # Test files
├── .env.example         # Environment variables template
├── .gitignore           # Git ignore patterns
├── package.json         # Node.js dependencies (or equivalent)
├── README.md            # Template-specific documentation
└── [framework-config]   # vite.config.js, next.config.js, etc.
```

## Build Configuration Requirements

All templates MUST configure their build systems to output to `public/`:

### Vite (React, Vue)
```js
// vite.config.js
export default defineConfig({
  build: {
    outDir: 'public',
    emptyOutDir: false, // Preserve API folder if present
  }
})
```

### Next.js
```js
// next.config.js
module.exports = {
  distDir: 'public',
  output: 'export', // for static export
}
```

### Webpack (Custom React)
```js
// webpack.config.js
module.exports = {
  output: {
    path: path.resolve(__dirname, 'public'),
  }
}
```

### Django
```python
# settings.py
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')
STATICFILES_DIRS = [os.path.join(BASE_DIR, 'static')]
```

### Flask
```python
# app.py
app = Flask(__name__, static_folder='public', static_url_path='')
```

## Port Allocation by Runtime

Each backend runtime has a dedicated 100-port range:

- **PHP**: 8100-8199
- **Node.js**: 8200-8299
- **Python**: 8300-8399
- **Go**: 8400-8499
- **.NET**: 8600-8699

When creating an app, choose an available port within the appropriate range.

## Testing Your Template

After creating an app from a template:

1. **Build verification**:
   ```bash
   npm run build  # or equivalent
   ls -la public/  # Verify files are present
   ```

2. **Start backend runtime**:
   ```bash
   ./scripts/service-manager.sh start backend
   ```

3. **Configure proxy**:
   - Access Nginx Proxy Manager: http://localhost:81
   - Add proxy host for `your-app.test`
   - Point to backend container

4. **Test in browser**:
   ```bash
   curl https://your-app.test
   # or open in browser
   ```

## Adding New Templates

To add a new framework template:

1. Create directory in appropriate category:
   ```bash
   mkdir -p templates/node/my-framework
   ```

2. Create standard structure with `public/` directory

3. Add build configuration that outputs to `public/`

4. Include `.env.example` with required variables

5. Write comprehensive README.md

6. Test the template by creating a new app from it

7. Update this README.md with the new template

## Migration Guide

For existing apps that don't follow the `public/` pattern, see `MIGRATION_GUIDE.md`.

## Support

For issues or questions:
- Check template-specific README.md
- Review `APP_STRUCTURE.md` for structure standards
- See `CLAUDE.md` for DevArch architecture
- Run `./scripts/create-app.sh --help`
