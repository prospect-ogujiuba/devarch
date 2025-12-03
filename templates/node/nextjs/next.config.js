/** @type {import('next').NextConfig} */
const nextConfig = {
  // CRITICAL: Build to public/ directory for DevArch compatibility
  distDir: 'public/.next',

  // For static export (SPA mode)
  output: 'export',

  // Alternatively, for server-side rendering:
  // output: 'standalone',

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
