/** @type {import('next').NextConfig} */
const nextConfig = {
  // For static export (SPA mode)
  // Next.js will build to .next/ directory and export static files to out/
  // The build script will then copy out/ to public/ for web server
  output: 'export',

  // Image optimization (disable for static export)
  images: {
    unoptimized: true,
  },

  // Base path if app is not at root (optional)
  // basePath: '/my-app',

  // Asset prefix for CDN (optional)
  // assetPrefix: 'https://cdn.example.com',

  // Trailing slash
  trailingSlash: true,

  // React strict mode
  reactStrictMode: true,

  // Disable x-powered-by header
  poweredByHeader: false,

  // Environment variables available to the client
  env: {
    APP_NAME: process.env.NEXT_PUBLIC_APP_NAME || 'DevArch App',
    API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8200/api',
  },
}

module.exports = nextConfig
