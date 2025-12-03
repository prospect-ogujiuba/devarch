# DevArch Dashboard

A simple, self-contained PHP application dashboard that automatically scans the apps/ directory, detects application types, categorizes them by runtime and framework, displays their status, and provides a clean minimal UI.

## Features

- **Automatic App Detection**: Scans `/var/www/html/` and identifies all applications
- **Runtime Detection**: Identifies PHP, Node.js, Python, and Go applications
- **Framework Recognition**: Detects specific frameworks (Laravel, WordPress, React, Django, etc.)
- **Status Indicators**: Shows app status with color-coded visual indicators
- **Categorization**: Filter apps by runtime type (PHP, Node, Python, Go)
- **Clean UI**: Minimal, modern design with responsive layout
- **Zero Dependencies**: Pure PHP 8.3 with inline CSS, no external packages

## Access

After setting up the NPM proxy host for the dashboard, access it at:

```
https://dashboard.test
```

Or run directly via PHP container:

```bash
podman exec php php /var/www/html/dashboard/public/index.php
```

## How It Works

### Detection Logic

The dashboard uses file markers to detect application types:

#### PHP Applications
- **Laravel**: `artisan` file + `app/` directory
- **WordPress**: `wp-config.php` + `wp-content/` directory
- **Symfony**: `symfony/framework-bundle` in composer.json
- **Generic PHP**: `composer.json`, `index.php`

#### Node.js Applications
- **Next.js**: `next.config.js` or `next` in package.json
- **React**: `react` in package.json dependencies
- **Express**: `express` in package.json dependencies
- **Vue.js**: `vue` in package.json dependencies
- **Generic Node**: `package.json`

#### Python Applications
- **Django**: `manage.py` file
- **Flask**: `flask` in requirements.txt
- **FastAPI**: `fastapi` in requirements.txt
- **Generic Python**: `requirements.txt`, `pyproject.toml`

#### Go Applications
- **Gin**: `gin-gonic/gin` in go.mod
- **Echo**: `labstack/echo` in go.mod
- **Fiber**: `gofiber/fiber` in go.mod
- **Generic Go**: `go.mod`, `main.go`

### Status Detection

The dashboard determines app status based on:

- **Ready** (yellow): App has proper structure files and appears configured
- **Unknown** (gray): Cannot determine status (default for most apps)

Future enhancements could include:
- Container status checking
- Port availability testing
- Health endpoint pinging

### URL Generation

Each app is assigned a `.test` domain URL:
- Pattern: `https://{appname}.test`
- Example: `b2bcnc.test`, `playground.test`

## Architecture

### Single-File Design

The dashboard is implemented as a single PHP file (`public/index.php`) for maximum simplicity:

```
apps/dashboard/
├── public/
│   └── index.php          # Complete dashboard (HTML + CSS + PHP)
└── README.md              # This file
```

### Key Components

1. **Configuration**: Base paths and exclusions
2. **Detection Functions**: Runtime and framework identification
3. **Data Collection**: Scans apps and builds metadata
4. **Rendering**: HTML output with inline CSS

### No External Dependencies

- No Composer packages
- No npm packages
- No CDN resources
- Pure PHP 8.3 (match expressions, readonly, etc.)
- Modern CSS (Grid, Flexbox, CSS variables)
- No JavaScript required (filtering via URL params)

## Usage

### Viewing All Apps

Simply navigate to `https://dashboard.test` to see all detected applications.

### Filtering by Runtime

Use the filter buttons at the top of the page, or navigate directly:

- All apps: `https://dashboard.test`
- PHP only: `https://dashboard.test?filter=php`
- Node.js only: `https://dashboard.test?filter=node`
- Python only: `https://dashboard.test?filter=python`
- Go only: `https://dashboard.test?filter=go`

### Statistics

The dashboard displays summary statistics:
- Total number of applications
- Breakdown by runtime (PHP, Node, Python, Go)
- Color-coded for easy identification

## Extending the Dashboard

### Adding New Runtime Detection

Edit the `detectRuntime()` function to add new runtime types:

```php
function detectRuntime(string $appPath): string {
    // Add your detection logic
    if (file_exists("$appPath/your-marker-file")) {
        return 'your-runtime';
    }

    // ... existing logic
}
```

### Adding Framework Detection

Edit the appropriate framework detection function (`detectPhpFramework`, `detectNodeFramework`, etc.):

```php
function detectPhpFramework(string $appPath): string {
    // Check for your framework
    if (file_exists("$appPath/your-framework-marker")) {
        return 'YourFramework';
    }

    // ... existing logic
}
```

### Customizing Colors

Edit the CSS variables in the `<style>` section:

```css
:root {
    --color-php: #8892BF;      /* PHP badge color */
    --color-node: #68A063;     /* Node badge color */
    --color-python: #3776AB;   /* Python badge color */
    --color-go: #00ADD8;       /* Go badge color */
    --color-active: #22c55e;   /* Active status */
    --color-ready: #eab308;    /* Ready status */
    --color-stopped: #94a3b8;  /* Stopped status */
}
```

### Adding Status Checks

Enhance the `getAppStatus()` function to add real status checking:

```php
function getAppStatus(string $appPath, string $runtime): string {
    // Check container status
    // Ping health endpoints
    // Check port availability

    return 'active' | 'ready' | 'stopped' | 'unknown';
}
```

## Design Philosophy

This dashboard follows the same principles as the `serverinfo` app:

1. **Simplicity First**: Single file, minimal code
2. **Zero Dependencies**: No external packages or CDNs
3. **Self-Contained**: Works out of the box
4. **Pure PHP**: Standard library functions only
5. **Clean UI**: Modern, minimal design
6. **Responsive**: Works on all devices

## Comparison to serverinfo

| Feature | serverinfo | dashboard |
|---------|------------|-----------|
| Purpose | PHP configuration | App inventory |
| Files | 1 (index.php) | 1 (index.php) |
| Dependencies | None | None |
| Approach | Call phpinfo() | Scan and detect |
| UI | PHP default | Custom minimal |
| Framework | None | None |

## Future Enhancements

Potential improvements for future versions:

1. **Real-time Status**: Container status checking via podman commands
2. **Port Checking**: Verify if ports are responding
3. **Log Viewing**: Display recent app logs
4. **Resource Usage**: Show CPU/memory usage per app
5. **Quick Actions**: Start/stop/restart apps
6. **Version Detection**: Show framework versions
7. **Health Endpoints**: Ping app health URLs
8. **Search/Sort**: Search by name, sort by various fields
9. **Dark Mode**: Theme toggle
10. **Export**: Export app inventory as JSON/CSV

## Troubleshooting

### Dashboard shows no apps

- Verify the dashboard is running in the PHP container
- Check that `APPS_BASE_PATH` is set correctly (`/var/www/html`)
- Ensure app directories are readable

### Wrong framework detected

- Check detection order in `detectFramework()` functions
- Verify marker files exist in the app directory
- Review detection logic for your specific framework

### Styling issues

- Ensure modern browser (supports CSS Grid and Flexbox)
- Check for browser extensions that might interfere
- Verify inline CSS is not being stripped

## Technical Details

- **PHP Version**: 8.3+
- **Path**: `/var/www/html/dashboard/public/index.php`
- **Container**: PHP container
- **Port**: 8100 (PHP container external port)
- **Internal Port**: 8000 (PHP-FPM)
- **URL Pattern**: `https://dashboard.test`

## License

Part of the DevArch development environment.

## Credits

Inspired by the simplicity of `phpinfo()` and the `serverinfo` app.
