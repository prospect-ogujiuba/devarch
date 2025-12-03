# DevArch Dashboard 2.0

A modern React-based application dashboard that automatically scans the apps/ directory, detects application types, categorizes them by runtime and framework, displays their status, and provides an enhanced interactive UI with dark mode, search, sorting, and real-time updates.

## Features

### Core Features (from v1.0)
- **Automatic App Detection**: Scans `/var/www/html/` and identifies all applications
- **Runtime Detection**: Identifies PHP, Node.js, Python, Go, and .NET applications
- **Framework Recognition**: Detects specific frameworks (Laravel, WordPress, React, Django, Gin, ASP.NET Core, etc.)
- **Status Indicators**: Shows app status with color-coded visual indicators
- **Categorization**: Filter apps by runtime type (PHP, Node, Python, Go, .NET)
- **Clean UI**: Modern, minimal design with responsive layout

### New Features (v2.0)
- **Search Functionality**: Instant search across app names, frameworks, and runtimes
- **Advanced Sorting**: Sort apps by name, runtime, framework, or status (ascending/descending)
- **Dark Mode**: Persistent theme toggle with system preference detection
- **Real-time Updates**: Auto-refresh app data every 30 seconds without page reload
- **Detailed View**: Click any app card to see expanded details in a modal
- **Enhanced UX**: Smooth transitions, loading states, error handling, and responsive design
- **Modern Tech Stack**: React 18, Vite, Tailwind CSS for excellent performance

## Access

### Development Mode (Vite Dev Server)

```bash
# Install dependencies (first time only)
cd /home/fhcadmin/projects/devarch/apps/dashboard
npm install

# Start development server
npm run dev
```

Access at: `http://localhost:5173` (from host) or `http://localhost:8102` (mapped port)

### Production Mode

After setting up the NPM proxy host for the dashboard, access it at:

```
https://dashboard.test
```

## Architecture

### Modern React Application

The dashboard v2.0 is a complete rewrite using modern web technologies:

```
apps/dashboard/
├── api/                      # PHP REST API Backend
│   ├── apps.php             # Main API endpoint
│   └── lib/
│       └── detection.php    # Detection logic library
├── src/                     # React Frontend
│   ├── components/          # UI Components
│   │   ├── AppCard.jsx
│   │   ├── AppDetailModal.jsx
│   │   ├── AppsGrid.jsx
│   │   ├── EmptyState.jsx
│   │   ├── ErrorMessage.jsx
│   │   ├── FilterBar.jsx
│   │   ├── Header.jsx
│   │   ├── LoadingSpinner.jsx
│   │   ├── SearchBar.jsx
│   │   ├── SortControls.jsx
│   │   ├── StatCard.jsx
│   │   └── ThemeToggle.jsx
│   ├── contexts/            # React Contexts
│   │   └── ThemeContext.jsx
│   ├── hooks/               # Custom Hooks
│   │   ├── useApps.js
│   │   ├── useDebounce.js
│   │   └── useLocalStorage.js
│   ├── utils/               # Helper Functions
│   │   ├── colors.js
│   │   └── formatters.js
│   ├── App.jsx              # Main App Component
│   ├── main.jsx             # Entry Point
│   └── index.css            # Global Styles
├── public/                  # Static Assets
│   └── index.php           # Legacy v1.0 (for reference)
├── index.html              # HTML Entry Point
├── package.json            # Dependencies
├── vite.config.js          # Vite Configuration
├── tailwind.config.js      # Tailwind Configuration
├── postcss.config.js       # PostCSS Configuration
├── .eslintrc.cjs          # ESLint Configuration
├── .env.example           # Environment Variables Template
└── README.md              # This file
```

### Technology Stack

**Frontend:**
- **React 18**: Modern UI framework with hooks and concurrent features
- **Vite**: Lightning-fast build tool and dev server
- **Tailwind CSS**: Utility-first CSS framework for rapid UI development
- **Modern JavaScript**: ES modules, async/await, arrow functions

**Backend:**
- **PHP 8.3**: REST API with JSON responses
- **Detection Library**: Preserved all detection logic from v1.0

**Development:**
- **ESLint**: Code quality and consistency
- **Hot Module Replacement**: Instant feedback during development
- **Auto-refresh**: Background updates without disrupting user interaction

## How It Works

### Detection Logic

The dashboard uses file markers to detect application types (unchanged from v1.0):

#### PHP Applications
- **Laravel**: `artisan` file + `app/` directory
- **WordPress**: `wp-config.php` + `wp-content/` directory
- **Symfony**: `symfony/framework-bundle` in composer.json
- **Slim**: `slim/slim` in composer.json
- **Generic PHP**: `composer.json`, `index.php`

#### Node.js Applications
- **Next.js**: `next.config.js` or `next` in package.json
- **Nuxt.js**: `nuxt` in package.json
- **React**: `react` in package.json dependencies
- **Vue.js**: `vue` in package.json dependencies
- **Angular**: `@angular/core` in package.json
- **Express**: `express` in package.json dependencies
- **Fastify**: `fastify` in package.json dependencies
- **Koa**: `koa` in package.json dependencies
- **Generic Node**: `package.json`

#### Python Applications
- **Django**: `manage.py` file
- **Flask**: `flask` in requirements.txt
- **FastAPI**: `fastapi` in requirements.txt
- **Tornado**: `tornado` in requirements.txt
- **Bottle**: `bottle` in requirements.txt
- **Generic Python**: `requirements.txt`, `pyproject.toml`

#### Go Applications
- **Gin**: `gin-gonic/gin` in go.mod
- **Echo**: `labstack/echo` in go.mod
- **Fiber**: `gofiber/fiber` in go.mod
- **Gorilla Mux**: `gorilla/mux` in go.mod
- **Chi**: `chi` in go.mod
- **Generic Go**: `go.mod`, `main.go`

#### .NET Applications
- **Blazor WebAssembly**: `Microsoft.AspNetCore.Components.WebAssembly` in .csproj
- **Blazor Server**: `Microsoft.AspNetCore.Components` in .csproj
- **ASP.NET Core Web API**: `Swashbuckle.AspNetCore` in .csproj
- **ASP.NET Core**: `Microsoft.NET.Sdk.Web` in .csproj
- **.NET Worker Service**: `Microsoft.Extensions.Hosting` in .csproj
- **Generic .NET**: `.csproj`, `.sln`, `.fsproj`, `Program.cs`

### API Architecture

**Endpoint**: `/api/apps.php`

**Query Parameters:**
- `filter`: Filter by runtime (php, node, python, go, dotnet, all)
- `search`: Search query for name, framework, or runtime
- `sort`: Sort field (name, runtime, framework, status)
- `order`: Sort order (asc, desc)

**Response Format:**
```json
{
  "success": true,
  "data": {
    "apps": [...],
    "stats": {...},
    "count": 10,
    "total": 15
  },
  "meta": {
    "timestamp": 1234567890,
    "base_path": "/var/www/html",
    "filter": "all",
    "search": null,
    "sort": "name",
    "order": "asc"
  }
}
```

### Status Detection

The dashboard determines app status based on:

- **Ready** (green): App has proper structure files and appears configured
- **Unknown** (gray): Cannot determine status (default for most apps)

Future enhancements could include:
- Container status checking via Podman API
- Port availability testing
- Health endpoint pinging
- Real-time container logs

### URL Generation

Each app is assigned a `.test` domain URL:
- Pattern: `https://{appname}.test`
- Example: `b2bcnc.test`, `playground.test`

## Usage

### Installation

```bash
# Navigate to dashboard directory
cd /home/fhcadmin/projects/devarch/apps/dashboard

# Install dependencies
npm install

# Copy environment variables (optional)
cp .env.example .env

# Edit .env if needed
nano .env
```

### Development

```bash
# Start dev server (with hot reload)
npm run dev

# Lint code
npm run lint

# Build for production
npm run build

# Preview production build
npm run preview
```

### Environment Variables

Create a `.env` file from `.env.example`:

```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8100/api

# Auto-refresh interval (milliseconds)
VITE_REFRESH_INTERVAL=30000

# Feature flags
VITE_ENABLE_AUTO_REFRESH=true
VITE_ENABLE_DARK_MODE=true
```

### User Interface

**Header:**
- Application title and description
- Last updated timestamp
- Manual refresh button
- Dark mode toggle

**Statistics Panel:**
- Total applications count
- Breakdown by runtime (PHP, Node, Python, Go, .NET)
- Color-coded values matching runtime colors

**Search & Controls:**
- Search bar with instant results
- Sort dropdown (name, runtime, framework, status)
- Sort order toggle (ascending/descending)

**Filters:**
- All Apps, PHP, Node.js, Python, Go, .NET
- One-click filtering with visual feedback
- Persistent selection (saved to localStorage)

**App Cards:**
- App name in monospace font
- Status indicator (colored dot)
- Runtime badge (color-coded)
- Framework badge
- Application URL
- Open button (opens in new tab)
- Details button (opens modal)

**Detail Modal:**
- Full app information
- Status, runtime, framework
- Application URL (clickable)
- File path
- Quick actions (open app, close modal)
- Keyboard shortcut (ESC to close)

### Keyboard Shortcuts

- **ESC**: Close detail modal
- **Ctrl/Cmd + K**: Focus search bar (browser default)

### Features in Detail

#### Search
- Searches across app names, frameworks, and runtimes
- Instant results with debouncing (300ms)
- Case-insensitive matching
- Clear button to reset search

#### Sorting
- Sort by: Name, Runtime, Framework, Status
- Toggle between ascending and descending order
- Visual indicator for current sort direction
- Persistent across page reloads

#### Dark Mode
- Toggle between light and dark themes
- Respects system preference on first load
- Persistent selection (saved to localStorage)
- Smooth transitions between themes
- All components fully themed

#### Auto-refresh
- Background updates every 30 seconds (configurable)
- No UI disruption or scroll position changes
- Manual refresh button for immediate updates
- Loading indicator during refresh
- Error handling with retry option

#### Responsive Design
- Mobile-first approach
- Breakpoints: sm (640px), md (768px), lg (1024px)
- Collapsible navigation on mobile
- Touch-friendly tap targets
- Optimized grid layouts for all screen sizes

## Extending the Dashboard

### Adding a New Component

Create a new component in `src/components/`:

```jsx
// src/components/MyComponent.jsx
export function MyComponent({ prop1, prop2 }) {
  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg p-4">
      {/* Your component JSX */}
    </div>
  )
}
```

Import and use it in `App.jsx` or other components.

### Adding a New Hook

Create a custom hook in `src/hooks/`:

```javascript
// src/hooks/useMyHook.js
import { useState, useEffect } from 'react'

export function useMyHook(param) {
  const [state, setState] = useState(null)

  useEffect(() => {
    // Hook logic
  }, [param])

  return state
}
```

### Customizing Colors

Edit `tailwind.config.js` to add/modify colors:

```javascript
theme: {
  extend: {
    colors: {
      runtime: {
        php: '#8892BF',
        node: '#68A063',
        // Add more
      }
    }
  }
}
```

Or edit `src/utils/colors.js` for runtime-specific colors.

### Adding API Parameters

Modify `src/hooks/useApps.js` to add new API parameters:

```javascript
const params = new URLSearchParams()
params.append('myParam', value)
```

Update `api/apps.php` to handle the new parameter.

### Adding Detection for New Runtime

1. Edit `api/lib/detection.php`
2. Add detection logic to `detectRuntime()`
3. Create detection function `detectYourRuntimeFramework()`
4. Update `detectFramework()` to call your function
5. Add color mapping in `src/utils/colors.js`
6. Update filter options in `src/components/FilterBar.jsx`

Example:

```php
// In api/lib/detection.php
function detectRuntime(string $appPath): string {
    // Rust detection
    if (file_exists("$appPath/Cargo.toml")) {
        return 'rust';
    }

    // ... existing logic
}

function detectRustFramework(string $appPath): string {
    if (file_exists("$appPath/Cargo.toml")) {
        $cargo = file_get_contents("$appPath/Cargo.toml");
        if (stripos($cargo, 'actix-web') !== false) return 'Actix Web';
        if (stripos($cargo, 'rocket') !== false) return 'Rocket';
    }
    return 'Rust';
}
```

## Comparison: v1.0 vs v2.0

| Feature | v1.0 (PHP) | v2.0 (React) |
|---------|------------|--------------|
| Files | 1 (index.php) | 20+ (modular) |
| Tech Stack | Pure PHP | React + Vite + Tailwind |
| Dependencies | None | npm packages |
| UI Framework | Inline CSS | Tailwind CSS |
| Search | No | Yes (instant) |
| Sort | No | Yes (multi-field) |
| Dark Mode | No | Yes (persistent) |
| Auto-refresh | No | Yes (30s interval) |
| Detail View | No | Yes (modal) |
| Mobile UX | Basic | Optimized |
| Performance | Server-side | Client-side caching |
| Extensibility | Limited | Highly modular |

## Performance Optimization

**Frontend:**
- Component memoization with React.memo (where needed)
- Debounced search to reduce re-renders
- Lazy loading for modal components
- Optimized Tailwind build (purges unused CSS)
- Vite's code splitting and tree shaking

**Backend:**
- Efficient file system scanning
- Cached composer.json/package.json parsing
- Minimal JSON payload sizes

**Network:**
- Auto-refresh runs in background
- Failed requests don't block UI
- API responses include metadata for debugging

## Troubleshooting

### Dashboard shows no apps

**Check API connectivity:**
```bash
curl http://localhost:8100/api/apps.php
```

**Verify PHP container is running:**
```bash
podman ps | grep php
```

**Check API endpoint:**
- Open browser dev tools (F12)
- Go to Network tab
- Refresh dashboard
- Look for `/api/apps.php` request
- Check response status and data

### Development server won't start

**Check Node.js version:**
```bash
node --version  # Should be 18+
```

**Clear node_modules and reinstall:**
```bash
rm -rf node_modules package-lock.json
npm install
```

**Check port availability:**
```bash
lsof -i :5173  # See what's using port 5173
```

### Dark mode not working

**Clear localStorage:**
```javascript
// In browser console
localStorage.removeItem('devarch-theme')
location.reload()
```

### Auto-refresh not working

**Check environment variable:**
```bash
# In .env file
VITE_ENABLE_AUTO_REFRESH=true
```

**Check browser console for errors:**
- Open dev tools (F12)
- Look for network errors or JavaScript exceptions

### Wrong framework detected

- Check detection order in `api/lib/detection.php`
- Verify marker files exist in the app directory
- Review detection logic for your specific framework
- Test API directly: `curl http://localhost:8100/api/apps.php?filter=all`

### Styling issues

- Clear browser cache (Ctrl+Shift+R / Cmd+Shift+R)
- Check for browser extensions that might interfere
- Verify Tailwind classes are being generated
- Run `npm run build` to see if there are any warnings

## Technical Details

**Frontend:**
- React 18.2.0
- Vite 5.0.8
- Tailwind CSS 3.4.0
- Dev Server Port: 5173 (internal), 8102 (external)

**Backend:**
- PHP 8.3+
- Path: `/var/www/html/dashboard/api/`
- Container: PHP container
- Port: 8100 (PHP container external port)

**Deployment:**
- Production builds to `dist/` directory
- Can be served by any static file server
- API proxied through Nginx Proxy Manager

## Migration from v1.0

The v1.0 PHP file is preserved at `public/index.php` for reference. To migrate:

1. Install dependencies: `npm install`
2. Configure environment variables: `cp .env.example .env`
3. Test API endpoint: `curl http://localhost:8100/api/apps.php`
4. Start dev server: `npm run dev`
5. Update Nginx Proxy Manager to point to React build or dev server

Both versions can run simultaneously for testing:
- v1.0: `https://dashboard.test` (via PHP)
- v2.0: `http://localhost:5173` (dev server)

## Future Enhancements

Potential improvements for future versions:

1. **Container Integration**: Real-time status via Podman API
2. **Log Streaming**: View live logs from running containers
3. **Resource Monitoring**: CPU/memory usage per app
4. **Quick Actions**: Start/stop/restart apps from UI
5. **Health Checks**: Automated endpoint testing
6. **Version Detection**: Display framework/runtime versions
7. **Export/Import**: Export app inventory as JSON/CSV
8. **Notifications**: Desktop notifications for status changes
9. **Multi-user**: User authentication and personalization
10. **Analytics**: Usage tracking and insights dashboard

## Contributing

When contributing to the dashboard:

1. Follow React best practices and hooks patterns
2. Use functional components (no class components)
3. Keep components small and focused (single responsibility)
4. Write descriptive prop names and add comments
5. Use Tailwind utilities (avoid custom CSS)
6. Test on multiple screen sizes
7. Ensure dark mode compatibility
8. Run linter before committing: `npm run lint`

## License

Part of the DevArch development environment.

## Credits

**Version 2.0**: Modern React rewrite with enhanced UX
**Version 1.0**: Original PHP implementation

Inspired by the simplicity of `phpinfo()` and the `serverinfo` app, evolved into a full-featured application management dashboard.
