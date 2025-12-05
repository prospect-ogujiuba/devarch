# DevArch Application Templates Guide

**Version**: 1.0.0
**Last Updated**: 2025-12-03

Complete guide to using application templates in DevArch for rapid application scaffolding with standardized `public/` directory structure.

## Table of Contents

- [Overview](#overview)
- [Available Templates](#available-templates)
- [Quick Start](#quick-start)
- [Template Structure](#template-structure)
- [Creating Apps from Templates](#creating-apps-from-templates)
- [Customizing Templates](#customizing-templates)
- [Template Development](#template-development)

## Overview

DevArch provides pre-configured application templates for common frameworks. Each template:

✅ Follows the standardized `public/` directory pattern
✅ Includes proper build configurations
✅ Contains example code and documentation
✅ Is ready to deploy in DevArch environment
✅ Supports immediate development workflow

### Benefits of Using Templates

- **Instant Setup**: Create production-ready apps in seconds
- **Best Practices**: Templates follow DevArch standards and framework conventions
- **No Configuration**: Build systems pre-configured to output to `public/`
- **Consistency**: All apps have similar structure, easier to maintain
- **Documentation**: Each template includes comprehensive README

## Available Templates

### PHP Applications

#### Laravel (`php/laravel`)
- **Framework**: Laravel 10+
- **Port Range**: 8100-8199
- **Package Manager**: Composer
- **Build Process**: None (interpreted)
- **Best For**: Full-stack PHP applications, APIs, admin panels

**Features**:
- Native `public/` directory support
- Eloquent ORM, Blade templating
- Artisan CLI commands
- Authentication scaffolding

#### WordPress (`php/wordpress`)
- **Framework**: WordPress 6+
- **Port Range**: 8100-8199
- **Package Manager**: Composer (optional)
- **Build Process**: None (interpreted)
- **Best For**: CMS, blogs, content sites

**Features**:
- WordPress installed in `public/` subdirectory
- Multiple preset configurations
- Plugin and theme support
- Database integration

#### Vanilla PHP (`php/vanilla`)
- **Framework**: Plain PHP 8.3
- **Port Range**: 8100-8199
- **Package Manager**: Composer (optional)
- **Build Process**: None (interpreted)
- **Best For**: Custom PHP applications, microservices

**Features**:
- Minimal setup
- PSR-4 autoloading
- Composer support
- Simple file structure

### Node.js Applications

#### React + Vite (`node/react-vite`)
- **Framework**: React 18+ with Vite 5
- **Port Range**: 8200-8299
- **Package Manager**: npm
- **Build Process**: Vite → `public/`
- **Best For**: Modern SPAs, dashboards, web apps

**Features**:
- ⚡ Fast development with Hot Module Replacement
- React Router for routing
- ESLint configured
- Path aliases (@components, @utils)
- Builds to `public/` directory

**Build Configuration**:
```javascript
build: {
  outDir: 'public',
  emptyOutDir: false,
}
```

#### Next.js (`node/nextjs`)
- **Framework**: Next.js 14+ with App Router
- **Port Range**: 8200-8299
- **Package Manager**: npm
- **Build Process**: Next.js → `public/`
- **Best For**: SSR applications, static sites, hybrid apps

**Features**:
- File-based routing
- Server Components
- Static export or standalone mode
- TypeScript support
- Builds to `public/.next/`

**Build Configuration**:
```javascript
distDir: 'public/.next',
output: 'export',
```

#### Express (`node/express`)
- **Framework**: Express.js 4+
- **Port Range**: 8200-8299
- **Package Manager**: npm
- **Build Process**: None (serves from `public/`)
- **Best For**: APIs, backend services, hybrid apps

**Features**:
- RESTful API structure
- Static file serving from `public/`
- CORS and security middleware
- API documentation routes
- Health check endpoints

**Static Serving**:
```javascript
app.use(express.static('public'))
```

#### Vue.js (`node/vue`)
- **Framework**: Vue 3+
- **Port Range**: 8200-8299
- **Package Manager**: npm
- **Build Process**: Vite → `public/`
- **Best For**: Progressive web apps, admin interfaces

**Features**:
- Composition API
- Vue Router
- Vite build system
- TypeScript support (optional)

### Python Applications

#### Django (`python/django`)
- **Framework**: Django 4+
- **Port Range**: 8300-8399
- **Package Manager**: pip
- **Build Process**: collectstatic → `public/static/`
- **Best For**: Full-stack web apps, admin systems

**Features**:
- Django ORM
- Admin interface
- Template system
- REST framework ready
- Static files in `public/static/`

**Static Configuration**:
```python
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')
```

#### Flask (`python/flask`)
- **Framework**: Flask 3+
- **Port Range**: 8300-8399
- **Package Manager**: pip
- **Build Process**: None (serves from `public/`)
- **Best For**: Microservices, APIs, small web apps

**Features**:
- Lightweight and flexible
- Jinja2 templating
- RESTful routing
- Serves from `public/` directory

**Static Configuration**:
```python
app = Flask(__name__, static_folder='public')
```

#### FastAPI (`python/fastapi`)
- **Framework**: FastAPI 0.104+
- **Port Range**: 8300-8399
- **Package Manager**: pip
- **Build Process**: None (API only)
- **Best For**: Modern APIs, microservices

**Features**:
- Automatic API documentation
- Type hints and validation
- Async support
- High performance

### Go Applications

#### Gin (`go/gin`)
- **Framework**: Gin 1.9+
- **Port Range**: 8400-8499
- **Package Manager**: go modules
- **Build Process**: Compile → binary
- **Best For**: High-performance APIs, microservices

**Features**:
- Fast HTTP router
- Middleware support
- JSON validation
- Serves static from `public/`

**Static Serving**:
```go
r.Static("/assets", "./public/assets")
```

#### Echo (`go/echo`)
- **Framework**: Echo 4+
- **Port Range**: 8400-8499
- **Package Manager**: go modules
- **Build Process**: Compile → binary
- **Best For**: RESTful APIs, web services

**Features**:
- Optimized router
- Middleware chain
- Data binding
- Comprehensive error handling

### .NET Applications

#### ASP.NET Core (`dotnet/aspnet-core`)
- **Framework**: ASP.NET Core 8+
- **Port Range**: 8600-8699
- **Package Manager**: dotnet/NuGet
- **Build Process**: Compile → `public/`
- **Best For**: Enterprise APIs, web applications

**Features**:
- MVC or Web API
- Entity Framework Core
- Dependency injection
- Swagger/OpenAPI

## Quick Start

### Interactive Mode (Recommended)

```bash
./scripts/create-app.sh
```

Follow the interactive prompts:
1. Select template from list
2. Enter application name
3. Specify port (or use default)
4. Confirm creation

### Non-Interactive Mode

```bash
# Syntax
./scripts/create-app.sh --name APP_NAME --template TEMPLATE --port PORT

# Examples
./scripts/create-app.sh --name my-dashboard --template node/react-vite

./scripts/create-app.sh --name api-server --template node/express --port 8201

./scripts/create-app.sh --name admin-panel --template python/django
```

### List Available Templates

```bash
./scripts/create-app.sh --list
```

## Template Structure

Every template follows this standard structure:

```
template-name/
├── public/                 # WEB ROOT (mandatory)
│   ├── .gitkeep           # Ensures directory is tracked
│   └── api/               # API endpoints (optional)
│
├── src/                   # Source code
│   ├── components/        # UI components
│   ├── pages/             # Pages/views
│   ├── utils/             # Utilities
│   └── styles/            # Stylesheets
│
├── config/                # Configuration files
├── scripts/               # Build/deployment scripts
├── tests/                 # Test files
│
├── .env.example           # Environment variables template
├── .gitignore             # Git ignore patterns
├── package.json           # Dependencies (Node.js)
├── README.md              # Template documentation
│
└── [build-config]         # Build configuration
    ├── vite.config.js
    ├── next.config.js
    └── etc.
```

### Key Files

#### .env.example

Every template includes `.env.example` with required variables:

```env
# Application
APP_NAME=my-app
APP_ENV=development

# Server
PORT=8200

# API
API_BASE_URL=http://localhost:8200/api

# Feature Flags
ENABLE_ANALYTICS=false
```

#### README.md

Each template includes comprehensive documentation:
- Quick start instructions
- Development workflow
- Build process
- DevArch integration
- Troubleshooting
- Framework-specific guides

#### Build Configuration

Templates include pre-configured build systems:

**React/Vue (Vite)**:
```javascript
// vite.config.js
export default defineConfig({
  build: {
    outDir: 'public',  // ← Configured
  }
})
```

**Next.js**:
```javascript
// next.config.js
module.exports = {
  distDir: 'public/.next',  // ← Configured
  output: 'export',
}
```

**Django**:
```python
# settings.py
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')  # ← Configured
```

## Creating Apps from Templates

### Step-by-Step Process

#### 1. Create App from Template

```bash
./scripts/create-app.sh --name my-app --template node/react-vite
```

This automatically:
- ✅ Copies template files
- ✅ Creates `apps/my-app/` directory
- ✅ Customizes configurations with app name
- ✅ Creates `.env` from `.env.example`
- ✅ Ensures `public/` directory exists

#### 2. Navigate to App

```bash
cd apps/my-app
```

#### 3. Install Dependencies

```bash
# Node.js
npm install

# PHP
composer install

# Python
pip install -r requirements.txt

# Go
go mod download

# .NET
dotnet restore
```

#### 4. Configure Environment

```bash
# Edit .env file
nano .env

# Set app-specific values
APP_NAME=my-app
PORT=8200
```

#### 5. Development Mode

```bash
# Node.js
npm run dev

# Python
python manage.py runserver 0.0.0.0:8300  # Django
python app.py                             # Flask

# Go
go run main.go

# .NET
dotnet run
```

#### 6. Build for Production

```bash
# Node.js (React, Next.js, Vue)
npm run build

# Python (Django)
python manage.py collectstatic --noinput

# Go
go build -o app main.go

# .NET
dotnet build
```

#### 7. Verify Build Output

```bash
ls -la public/
# Should contain built files
```

#### 8. Start Backend Runtime

```bash
./scripts/service-manager.sh start backend
```

#### 9. Configure Nginx Proxy Manager

Access http://localhost:81 and add proxy host:

- **Domain**: `my-app.test`
- **Forward Hostname**: `nodejs` (or `php`, `python`, etc.)
- **Forward Port**: `8200` (your app's port)
- **Enable SSL**: Yes

#### 10. Access Application

- Development: http://localhost:8200
- Production: https://my-app.test

### Post-Creation Checklist

After creating an app from template:

- [ ] Dependencies installed
- [ ] `.env` configured
- [ ] Development server runs
- [ ] Build completes successfully
- [ ] `public/` contains built files
- [ ] Backend runtime is running
- [ ] Nginx Proxy Manager configured
- [ ] Domain resolves correctly
- [ ] Application loads in browser
- [ ] No console errors
- [ ] Assets load correctly

## Customizing Templates

### Modifying Existing Templates

Templates are located in `/home/fhcadmin/projects/devarch/templates/`.

To customize a template:

1. **Navigate to template**:
   ```bash
   cd templates/node/react-vite
   ```

2. **Make changes**:
   - Update source code in `src/`
   - Modify build configs
   - Update README.md
   - Adjust .env.example

3. **Test template**:
   ```bash
   cd /home/fhcadmin/projects/devarch
   ./scripts/create-app.sh --name test-app --template node/react-vite
   cd apps/test-app
   npm install
   npm run build
   ```

4. **Verify**:
   - App builds successfully
   - `public/` contains files
   - App runs in browser

### Creating New Templates

To add a new framework template:

#### 1. Create Template Directory

```bash
mkdir -p templates/node/my-framework
cd templates/node/my-framework
```

#### 2. Create Standard Structure

```bash
mkdir -p public src/{components,pages,utils,styles} config scripts tests
touch public/.gitkeep
```

#### 3. Add Build Configuration

Create framework-specific config file:

```bash
touch vite.config.js  # or next.config.js, webpack.config.js, etc.
```

**Ensure build outputs to `public/`**:

```javascript
export default defineConfig({
  build: {
    outDir: 'public',  // ← CRITICAL
  }
})
```

#### 4. Create package.json

```json
{
  "name": "devarch-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "...",
    "build": "...",    // Must build to public/
    "preview": "..."
  },
  "dependencies": {
    // Framework dependencies
  }
}
```

#### 5. Add .env.example

```env
APP_NAME=devarch-app
PORT=8200
API_BASE_URL=http://localhost:8200/api
```

#### 6. Create .gitignore

```gitignore
node_modules/
../.env
.env.local

# Build outputs
public/*.html
public/*.js
public/assets/*.js
!public/api/
```

#### 7. Write README.md

Document:
- Quick start
- Development workflow
- Build process
- DevArch integration
- Troubleshooting

#### 8. Add Example Code

Create minimal working example in `src/`:
- Entry point
- Example component
- Basic styles

#### 9. Test Template

```bash
./scripts/create-app.sh --name test-template --template node/my-framework
cd apps/test-template
npm install
npm run build
ls -la public/  # Verify output
```

#### 10. Update Template List

Add your template to:
- `templates/README.md`
- `scripts/lib/app-templates.sh` (list_templates function)
- This document (TEMPLATES.md)

### Template Best Practices

When creating or modifying templates:

✅ **DO**:
- Always output builds to `public/`
- Include comprehensive README
- Add .env.example with all variables
- Use framework best practices
- Include example/starter code
- Document port ranges
- Add .gitignore patterns
- Test template thoroughly

❌ **DON'T**:
- Hardcode configuration values
- Include secrets or credentials
- Use non-standard directory structures
- Skip documentation
- Forget build verification
- Commit node_modules or build outputs

## Template Troubleshooting

### Template Not Listed

**Cause**: Template not registered in `scripts/lib/app-templates.sh`.

**Solution**:
Add template to `list_templates()` function in `scripts/lib/app-templates.sh`.

### App Creation Fails

**Cause**: Missing template files or incorrect structure.

**Solution**:
1. Verify template directory exists:
   ```bash
   ls -la templates/{category}/{framework}/
   ```

2. Check for required files:
   - README.md
   - .env.example
   - .gitignore
   - Build config (if needed)

### Build Doesn't Output to public/

**Cause**: Build configuration not set correctly.

**Solution**:
Check build config file and ensure:
- Vite: `build.outDir: 'public'`
- Next.js: `distDir: 'public/.next'`
- Webpack: `output.path: 'public'`
- Django: `STATIC_ROOT = 'public/static'`

### Port Conflicts

**Cause**: Port already in use.

**Solution**:
Choose different port in same range:
- PHP: 8100-8199
- Node.js: 8200-8299
- Python: 8300-8399
- Go: 8400-8499
- .NET: 8600-8699

## Additional Resources

- **App Structure Standard**: `APP_STRUCTURE.md`
- **Migration Guide**: `MIGRATION_GUIDE.md`
- **DevArch Architecture**: `CLAUDE.md`
- **Template Directory**: `/home/fhcadmin/projects/devarch/templates/`
- **Individual Template READMEs**: `templates/{category}/{framework}/README.md`

## Version History

- **1.0.0** (2025-12-03): Initial templates guide

## Support

For template-related issues:
1. Check this guide (TEMPLATES.md)
2. Review template README: `templates/{category}/{framework}/README.md`
3. Check app structure: `APP_STRUCTURE.md`
4. Review DevArch docs: `CLAUDE.md`
