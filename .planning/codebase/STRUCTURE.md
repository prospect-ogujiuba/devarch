# Codebase Structure

**Analysis Date:** 2026-01-07

## Directory Layout

```
devarch/
├── apps/                    # Application projects
│   ├── dashboard/          # React management UI
│   ├── b2bcnc/            # WordPress multisite
│   ├── mxcros/            # Laravel + Node app
│   └── phplib/            # PHP shared libraries
├── compose/                # Docker Compose files (166 files, 24 categories)
│   ├── backend/           # Runtime services (php, node, python, go...)
│   ├── database/          # Data stores (postgres, mariadb, redis...)
│   ├── gateway/           # API gateways (kong, traefik, envoy...)
│   ├── management/        # Admin tools (portainer, devarch...)
│   └── [20 more categories]
├── config/                 # Service configurations (33 directories)
│   ├── devarch/           # Main management container
│   ├── php/               # PHP runtime
│   ├── node/              # Node.js runtime
│   └── [30 more services]
├── scripts/                # Management scripts
│   ├── service-manager.sh # Primary CLI tool
│   ├── config.sh          # Central configuration
│   └── [support scripts]
├── .env                    # Environment variables
├── .env.example           # Environment template
└── CLAUDE.md              # Project guidance
```

## Directory Purposes

**apps/dashboard/**
- Purpose: React management dashboard for DevArch
- Contains: Vite React app with Tailwind CSS
- Key files: `src/main.jsx`, `vite.config.js`, `package.json`
- Subdirectories:
  - `src/pages/` - Page components (Dashboard, Apps, Containers, Settings)
  - `src/components/` - 33 React components (ui, apps, containers, shared)
  - `src/hooks/` - 13 custom data-fetching hooks
  - `src/contexts/` - Global state (Theme, Toast)
  - `src/utils/` - Helper functions (formatters, colors)
  - `src/layouts/` - AppShell with sidebar navigation

**apps/b2bcnc/**
- Purpose: WordPress multisite B2B application
- Contains: WordPress installation with custom plugins/themes
- Key files: `wp-content/plugins/makerblocks/`, `wp-content/plugins/makermaker/`
- Subdirectories: Standard WordPress structure

**apps/mxcros/**
- Purpose: Laravel + Node hybrid application
- Contains: Laravel 12 with Vite asset compilation
- Key files: `composer.json`, `package.json`, `vite.config.js`

**config/devarch/**
- Purpose: Main management container configuration
- Contains: PHP API, Nginx config, Dockerfile
- Key files:
  - `api/public/index.php` - API router (124 routes)
  - `api/endpoints/*.php` - 16 endpoint handlers
  - `api/lib/*.php` - 8 business logic libraries
  - `Dockerfile` - Multi-language container
  - `nginx.conf` - Dashboard and API serving

**compose/**
- Purpose: Docker Compose service definitions
- Contains: 166 YAML files across 24 categories
- Key categories: backend (10), database (13), gateway (8), messaging (8), ci (8)
- Pattern: One file per service (`compose/{category}/{service}.yml`)

**scripts/**
- Purpose: Service orchestration and management
- Contains: Bash/Zsh scripts for service lifecycle
- Key files:
  - `service-manager.sh` - Primary CLI (up, down, rebuild, logs, status)
  - `config.sh` - Categories, ports, startup order, helpers
  - `runtime-switcher.sh` - Docker/Podman switching
  - `socket-manager.sh` - Socket management

## Key File Locations

**Entry Points:**
- `apps/dashboard/src/main.jsx` - React application entry
- `config/devarch/api/public/index.php` - PHP API router
- `scripts/service-manager.sh` - CLI entry point

**Configuration:**
- `.env` / `.env.example` - Environment variables
- `apps/dashboard/vite.config.js` - Vite bundler config
- `apps/dashboard/tailwind.config.js` - Tailwind CSS config
- `apps/dashboard/.eslintrc.cjs` - ESLint rules
- `scripts/config.sh` - Service categories and ports

**Core Logic:**
- `config/devarch/api/lib/containers.php` - Container inspection
- `config/devarch/api/lib/services.php` - Service management
- `apps/dashboard/src/hooks/useContainers.js` - Container data hook
- `apps/dashboard/src/hooks/useServiceManager.js` - Service control hook

**Testing:**
- `apps/mxcros/tests/` - Laravel PHPUnit tests
- `apps/b2bcnc/wp-content/plugins/makermaker/tests/` - WordPress Pest tests
- No React tests configured

**Documentation:**
- `CLAUDE.md` - Project guidance for AI
- `apps/b2bcnc/wp-content/plugins/makerblocks/CLAUDE.md` - Plugin guidance
- `apps/b2bcnc/wp-content/plugins/makermaker/CLAUDE.md` - Plugin guidance

## Naming Conventions

**Files:**
- kebab-case.sh - Shell scripts (`service-manager.sh`, `runtime-switcher.sh`)
- PascalCase.jsx - React components (`ContainerCard.jsx`, `AppShell.jsx`)
- camelCase.js - Hooks and utilities (`useContainers.js`, `formatters.js`)
- snake_case.php - PHP libraries (`containers.php`, `services.php`)
- lowercase.yml - Compose files (`postgres.yml`, `redis.yml`)

**Directories:**
- lowercase - All directories (`apps`, `config`, `compose`, `scripts`)
- kebab-case - Service names (`nginx-proxy-manager`)
- PascalCase - React component folders in some cases

**Special Patterns:**
- `use*.js` - React hooks (`useApps.js`, `useContainers.js`)
- `*Context.jsx` - React contexts (`ThemeContext.jsx`, `ToastContext.jsx`)
- `*.yml` - Compose files (one per service)
- `lib/*.php` - PHP libraries (shared business logic)
- `endpoints/*.php` - PHP endpoint handlers

## Where to Add New Code

**New React Feature:**
- Primary code: `apps/dashboard/src/pages/` or `src/components/`
- Data fetching: `apps/dashboard/src/hooks/`
- Global state: `apps/dashboard/src/contexts/`
- Utilities: `apps/dashboard/src/utils/`

**New API Endpoint:**
- Handler: `config/devarch/api/endpoints/{name}.php`
- Business logic: `config/devarch/api/lib/{name}.php`
- Route: Add case in `config/devarch/api/public/index.php`

**New Service:**
- Compose file: `compose/{category}/{service}.yml`
- Config files: `config/{service}/` (Dockerfile, configs)
- Category entry: `scripts/config.sh` SERVICE_CATEGORIES array

**New Script:**
- Location: `scripts/{name}.sh`
- Source config: `. "$(dirname "$0")/config.sh"`

## Special Directories

**compose/**
- Purpose: Service definitions (166 files)
- Source: Manually maintained YAML files
- Committed: Yes

**config/devarch/api/**
- Purpose: PHP API backend for dashboard
- Source: Custom PHP code
- Committed: Yes

**.planning/**
- Purpose: Project planning and codebase documentation
- Source: Generated by GSD workflow
- Committed: Yes (recommended)

**node_modules/, vendor/**
- Purpose: Package dependencies
- Source: npm install, composer install
- Committed: No (.gitignore)

---

*Structure analysis: 2026-01-07*
*Update when directory structure changes*
