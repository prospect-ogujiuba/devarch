# DevArch React + Vite Template

A modern React application template using Vite as the build tool, configured to work seamlessly with DevArch's standardized `public/` directory structure.

## Features

- React 18.3+ with React Router
- Vite 5+ for fast development and optimized builds
- ESLint for code quality
- Vitest for testing
- Path aliases (@components, @pages, @utils, @styles)
- Environment variable support
- API proxy configuration
- **Builds to `public/` directory for web server compatibility**

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Development Mode

```bash
npm run dev
```

Access at: http://localhost:5173

### 4. Build for Production

```bash
npm run build
```

This will build all assets to the `public/` directory, which is the web server document root.

### 5. Preview Production Build

```bash
npm run preview
```

## Directory Structure

```
react-vite-app/
├── public/              # WEB ROOT - Built assets go here
│   ├── index.html       # Built entry point
│   └── assets/          # Compiled JS, CSS, images
├── src/                 # Source code
│   ├── components/      # React components
│   ├── pages/           # Page components
│   ├── utils/           # Utility functions
│   ├── styles/          # CSS files
│   ├── App.jsx          # Root component
│   └── main.jsx         # Entry point
├── static/              # Static assets (copied as-is)
├── index.html           # Development HTML template
├── vite.config.js       # Vite configuration (outputs to public/)
├── package.json         # Dependencies
└── .env.example         # Environment variables template
```

## Important: Build Output Configuration

The `vite.config.js` is configured to build to `public/`:

```javascript
export default defineConfig({
  build: {
    outDir: 'public',
    emptyOutDir: false,
  }
})
```

**This is critical** for the app to work with DevArch's web server, which expects all apps to serve from the `public/` subdirectory.

## Environment Variables

All environment variables must be prefixed with `VITE_` to be available in the application:

```env
VITE_APP_NAME=My App
VITE_API_BASE_URL=http://localhost:8200/api
```

Access in code:

```javascript
const apiUrl = import.meta.env.VITE_API_BASE_URL
```

## Path Aliases

Configured aliases for cleaner imports:

- `@` → `/src`
- `@components` → `/src/components`
- `@pages` → `/src/pages`
- `@utils` → `/src/utils`
- `@styles` → `/src/styles`

Example:

```javascript
import Header from '@components/Header'
import { api } from '@utils/api'
```

## API Integration

### Development Proxy

The Vite dev server proxies API requests to avoid CORS issues:

```javascript
// vite.config.js
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8200',
      changeOrigin: true,
    }
  }
}
```

### Production API

In production, place API endpoints directly in the `public/api/` directory, or configure your backend runtime to handle API routes.

## DevArch Integration

### Port Assignment

Node.js apps use ports 8200-8299. Update in `.env`:

```env
VITE_PORT=8200
```

### Domain Configuration

After deployment, configure Nginx Proxy Manager:

1. Access: http://localhost:81
2. Add proxy host for `your-app.test`
3. Point to: `http://nodejs:8200`
4. Enable SSL certificate

### Container Access

The app will be accessible at:
- Development: http://localhost:5173
- Production: https://your-app.test

## Build Process

1. **Development**: Vite serves files from memory, no build needed
2. **Production**:
   ```bash
   npm run build
   ```
   - Compiles React to optimized JavaScript
   - Processes CSS and assets
   - Outputs everything to `public/`
   - Ready to serve via web server

## Testing

```bash
# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Generate coverage report
npm run test:coverage
```

## Linting

```bash
# Check for issues
npm run lint

# Auto-fix issues
npm run lint:fix
```

## Troubleshooting

### Build Not Appearing in public/

Check `vite.config.js` has:

```javascript
build: {
  outDir: 'public',
}
```

### Assets Not Loading

Ensure all asset imports use relative paths or path aliases.

### API Calls Failing

- Development: Check proxy configuration in `vite.config.js`
- Production: Verify API endpoint exists in `public/api/` or backend is configured

### Environment Variables Not Working

- Ensure variables are prefixed with `VITE_`
- Restart dev server after changing `.env`

## Customization

### Adding New Pages

1. Create component in `src/pages/`
2. Add route in `src/App.jsx`:

```javascript
import NewPage from './pages/NewPage'

<Route path="/new" element={<NewPage />} />
```

### Adding Dependencies

```bash
npm install package-name
```

### Updating Vite Config

Edit `vite.config.js` - but **keep `build.outDir: 'public'`** to maintain DevArch compatibility.

## Learn More

- [React Documentation](https://react.dev/)
- [Vite Documentation](https://vitejs.dev/)
- [React Router](https://reactrouter.com/)
- [DevArch Documentation](../../APP_STRUCTURE.md)

## Support

For issues specific to this template:
- Check DevArch documentation: `APP_STRUCTURE.md`
- Review template documentation: `templates/README.md`
- Verify build outputs to `public/` directory
