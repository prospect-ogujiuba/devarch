# DevArch Next.js Template

A Next.js 14+ application template configured for static export with DevArch's standardized `public/` directory structure.

## Features

- Next.js 14+ with App Router
- Static export to `public/` directory
- TypeScript support (optional)
- File-based routing
- API routes support
- Environment variable configuration
- **Builds to `public/` directory for web server compatibility**

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env.local
# Edit .env.local with your configuration
```

### 3. Development Mode

```bash
npm run dev
```

Access at: http://localhost:3000

### 4. Build for Production

```bash
npm run build
```

This exports the static site to `public/` directory.

### 5. Preview Production Build

```bash
npm start
```

## Directory Structure

```
nextjs-app/
├── .next/              # Next.js build cache (gitignored)
├── out/                # Static export output (gitignored)
├── public/             # WEB ROOT - Final production files
│   ├── _next/         # Static assets (copied from out/)
│   └── *.html         # Exported HTML pages (copied from out/)
├── src/
│   └── app/            # App Router
│       ├── layout.jsx  # Root layout
│       ├── page.jsx    # Home page
│       ├── globals.css # Global styles
│       └── about/
│           └── page.jsx
├── next.config.js      # Next.js configuration
└── package.json
```

## Important: Build Output Configuration

The build process works in two stages:

1. **Next.js builds and exports** to `out/` directory
2. **Build script copies** `out/` contents to `public/`

```javascript
// next.config.js
const nextConfig = {
  output: 'export', // Static export mode
  images: {
    unoptimized: true,
  },
}
```

```json
// package.json
{
  "scripts": {
    "build": "next build && rm -rf public/* && cp -r out/* public/"
  }
}
```

**This ensures** the `public/` directory serves as the web root without causing recursive build issues.

## Static Export vs Server-Side Rendering

### Static Export (Default)

```javascript
// next.config.js
output: 'export',
```

- Pre-renders all pages at build time
- No Node.js server required
- Works with DevArch's standard web server
- **Recommended for DevArch**

### Server-Side Rendering (Alternative)

```javascript
// next.config.js
output: 'standalone',
```

- Requires Node.js runtime
- Dynamic rendering per request
- Use with Node.js backend container
- Port 8200-8299 for Node apps

## Environment Variables

Next.js environment variables must be prefixed with `NEXT_PUBLIC_` to be available to the client:

```env
NEXT_PUBLIC_APP_NAME=My App
NEXT_PUBLIC_API_URL=http://localhost:8200/api
```

Access in code:

```javascript
const apiUrl = process.env.NEXT_PUBLIC_API_URL
```

## File-based Routing

Next.js uses file-based routing:

- `src/app/page.jsx` → `/`
- `src/app/about/page.jsx` → `/about`
- `src/app/blog/[slug]/page.jsx` → `/blog/:slug`

## API Routes

Create API endpoints in `src/app/api/`:

```javascript
// src/app/api/hello/route.js
export async function GET(request) {
  return Response.json({ message: 'Hello from API' })
}
```

Access at: `/api/hello`

**Note**: API routes don't work with static export. For APIs in static export mode, place endpoints in `public/api/` or use a separate backend.

## DevArch Integration

### Port Assignment

Node.js apps use ports 8200-8299:

```env
PORT=8200
```

### Domain Configuration

Configure Nginx Proxy Manager:

1. Access: http://localhost:81
2. Add proxy host for `your-app.test`
3. For static export: Point to `http://nginx:80`
4. For SSR: Point to `http://nodejs:8200`
5. Enable SSL certificate

### Container Access

- Development: http://localhost:3000
- Production (static): https://your-app.test
- Production (SSR): https://your-app.test (via Node container)

## Build Process

### Static Export

```bash
npm run build
```

1. Builds Next.js application to `.next/` directory
2. Pre-renders all pages to `out/` directory
3. Cleans `public/` directory
4. Copies `out/` contents to `public/`
5. Ready to serve via web server from `public/`

### Server-Side Rendering

```bash
npm run build
npm start
```

Runs Node.js server on specified port.

## Image Optimization

For static export, images must be unoptimized:

```javascript
// next.config.js
images: {
  unoptimized: true,
}
```

Alternatively, use external image optimization service.

## Testing

Add testing dependencies:

```bash
npm install --save-dev @testing-library/react @testing-library/jest-dom jest
```

Configure and run tests:

```bash
npm test
```

## Troubleshooting

### Build Not Appearing in public/

Check that:

1. `next.config.js` has `output: 'export'`
2. `package.json` build script includes the copy command:
   ```json
   "build": "next build && rm -rf public/* && cp -r out/* public/"
   ```
3. The `out/` directory exists after build
4. No permission issues when copying to `public/`

### Recursive Directory Error (ENAMETOOLONG)

If you encounter this error, ensure you're **NOT** using `distDir: 'public/.next'` in `next.config.js`. The correct configuration uses the default `.next/` directory and copies from `out/` to `public/`.

### 404 on Navigation

Ensure `trailingSlash: true` in `next.config.js` for static export.

### Images Not Loading

- Check `images.unoptimized: true` for static export
- Verify image paths are correct
- Use `next/image` component with `unoptimized` prop

### API Routes Not Working

API routes don't work with `output: 'export'`. Either:
- Use separate backend for APIs
- Switch to `output: 'standalone'` for SSR

## Customization

### Adding New Pages

Create new file in `src/app/`:

```bash
mkdir -p src/app/contact
touch src/app/contact/page.jsx
```

### Adding Layouts

Create `layout.jsx` in any directory:

```javascript
export default function ContactLayout({ children }) {
  return <div className="contact-layout">{children}</div>
}
```

### Adding Metadata

```javascript
export const metadata = {
  title: 'Page Title',
  description: 'Page description',
}
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [App Router](https://nextjs.org/docs/app)
- [Static Exports](https://nextjs.org/docs/app/building-your-application/deploying/static-exports)
- [DevArch Documentation](../../APP_STRUCTURE.md)

## Support

For template-specific issues:
- Check DevArch documentation: `APP_STRUCTURE.md`
- Review template documentation: `templates/README.md`
- Verify build outputs to `public/` directory
