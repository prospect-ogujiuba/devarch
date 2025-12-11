/**
 * PM2 Ecosystem Configuration for Multi-App Node.js Container
 *
 * Automatically discovers and runs Node.js applications in /app directory.
 * Each app should have:
 * - package.json with "scripts": { "start": "..." } or "dev": "..."
 * - Or standalone server.js/index.js entry point
 *
 * Apps are assigned sequential ports starting from 3000.
 *
 * DevArch Multi-Language Backend Architecture
 */

const fs = require('fs');
const path = require('path');

const APPS_DIR = '/app';
const BASE_PORT = 3000;

/**
 * Discovers Node.js applications by scanning for package.json files
 * @returns {Array} PM2 app configurations
 */
function discoverApps() {
  const apps = [];

  console.log('🔍 PM2: Discovering Node.js applications...');

  // Read directory entries
  let entries;
  try {
    entries = fs.readdirSync(APPS_DIR, { withFileTypes: true });
  } catch (error) {
    console.error(`❌ Failed to read ${APPS_DIR}:`, error.message);
    return apps;
  }

  let portCounter = BASE_PORT;

  for (const entry of entries) {
    // Skip non-directories and special directories
    if (!entry.isDirectory() || entry.name.startsWith('.') || entry.name === 'node_modules' || entry.name === 'logs') {
      continue;
    }

    const appPath = path.join(APPS_DIR, entry.name);
    const packageJsonPath = path.join(appPath, 'package.json');

    // Skip if no package.json (not a Node.js app)
    if (!fs.existsSync(packageJsonPath)) {
      continue;
    }

    // Skip static SPAs - they don't need a dev server
    // If app has dist/index.html or build/index.html, it's a static build
    const isStaticSPA = fs.existsSync(path.join(appPath, 'dist/index.html')) ||
                        fs.existsSync(path.join(appPath, 'build/index.html')) ||
                        fs.existsSync(path.join(appPath, 'out/index.html'));

    if (isStaticSPA) {
      console.log(`⏭️  Skipping ${entry.name} - static SPA (has dist/build/out folder)`);
      continue;
    }

    let packageJson;
    try {
      packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    } catch (error) {
      console.warn(`⚠️  Skipping ${entry.name} - invalid package.json:`, error.message);
      continue;
    }

    // Determine script to run
    let script = 'node';
    let args = '';
    let interpreter = 'node';
    let interpreterArgs = '';

    // Detect Next.js apps early for script selection
    const isNextJs = packageJson.dependencies?.next || packageJson.devDependencies?.next;
    const isDevelopment = process.env.NODE_ENV === 'development';

    if (packageJson.scripts) {
      // For Next.js in development, prefer 'dev' over 'start'
      // (start requires build first, dev doesn't)
      if (isNextJs && isDevelopment && packageJson.scripts.dev) {
        script = 'npm';
        args = 'run dev';
      } else if (packageJson.scripts.start) {
        // Use npm start for production or non-Next.js apps
        script = 'npm';
        args = 'start';
      } else if (packageJson.scripts.dev) {
        // Use npm run dev for development
        script = 'npm';
        args = 'run dev';
      } else if (packageJson.scripts['start:prod']) {
        // Some frameworks use start:prod
        script = 'npm';
        args = 'run start:prod';
      }
    }

    // Fallback to direct file execution
    if (script === 'node') {
      if (fs.existsSync(path.join(appPath, 'server.js'))) {
        args = 'server.js';
      } else if (fs.existsSync(path.join(appPath, 'index.js'))) {
        args = 'index.js';
      } else if (fs.existsSync(path.join(appPath, 'app.js'))) {
        args = 'app.js';
      } else if (fs.existsSync(path.join(appPath, 'main.js'))) {
        args = 'main.js';
      } else if (packageJson.main) {
        // Use package.json main field
        args = packageJson.main;
      } else {
        console.warn(`⚠️  Skipping ${entry.name} - no start script or entry point found`);
        continue;
      }
    }

    // Check for TypeScript apps (Next.js, NestJS, etc.)
    const hasTypeScript = fs.existsSync(path.join(appPath, 'tsconfig.json'));

    console.log(`📦 Found: ${entry.name}`);
    console.log(`   Script: ${script} ${args}`);
    console.log(`   Port: ${portCounter}`);
    console.log(`   TypeScript: ${hasTypeScript ? 'Yes' : 'No'}`);
    console.log(`   Framework: ${isNextJs ? 'Next.js' : 'Standard Node.js'}`);

    // Create PM2 app configuration
    apps.push({
      name: entry.name,
      script,
      args,
      cwd: appPath,
      env: {
        PORT: portCounter,
        NODE_ENV: process.env.NODE_ENV || 'development',
        APP_NAME: entry.name,
        // Next.js needs to bind to 0.0.0.0 to accept external connections
        HOST: isNextJs ? '0.0.0.0' : undefined,
        HOSTNAME: isNextJs ? '0.0.0.0' : undefined,
        // Pass through other environment variables
        ...process.env
      },
      // Development settings
      watch: process.env.PM2_WATCH !== 'false',  // Enable watch by default
      ignore_watch: [
        'node_modules',
        'dist',
        'build',
        '.next',
        'out',
        '.git',
        'logs',
        '*.log',
        '.env.local'
      ],
      // Resource management
      max_memory_restart: process.env.PM2_MAX_MEMORY || '500M',
      // Logging
      error_file: `/app/logs/${entry.name}-error.log`,
      out_file: `/app/logs/${entry.name}-out.log`,
      merge_logs: true,
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      // Auto-restart on crash
      autorestart: true,
      max_restarts: 10,
      min_uptime: '10s',
      // Kill timeout
      kill_timeout: 5000,
      // Shutdown signal
      shutdown_with_message: true
    });

    portCounter++;
  }

  console.log('');
  console.log('========================================');
  if (apps.length === 0) {
    console.log('⚠️  No Node.js applications found');
    console.log('');
    console.log('💡 To create a Node.js app:');
    console.log('   1. Use WebStorm IDE: File → New Project → Node.js/Express/Next.js');
    console.log(`   2. Location: ${APPS_DIR}/my-node-app`);
    console.log('   3. Container will auto-detect and start app');
    console.log('');
  } else {
    console.log(`✅ Discovered ${apps.length} Node.js application(s)`);
    console.log('');
    console.log('📊 Port mapping:');
    apps.forEach((app, index) => {
      const externalPort = 8200 + index;
      console.log(`   ${app.name}: localhost:${externalPort} → :${app.env.PORT}`);
    });
    console.log('');
    console.log('🔧 PM2 commands:');
    console.log('   devarch exec node pm2 list          - List all apps');
    console.log('   devarch exec node pm2 logs          - View all logs');
    console.log('   devarch exec node pm2 restart all   - Restart all apps');
    console.log('   devarch exec node pm2 monit          - Monitor resources');
    console.log('');
  }
  console.log('========================================');

  return apps;
}

module.exports = {
  apps: discoverApps()
};
