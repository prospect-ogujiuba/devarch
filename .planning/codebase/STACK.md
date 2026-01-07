# Technology Stack

**Analysis Date:** 2026-01-07

## Languages

**Primary:**
- JavaScript/JSX - All dashboard frontend code (`apps/dashboard/src/`)
- PHP 8.2+ - API backend, Laravel, WordPress (`config/devarch/api/`, `apps/mxcros/`, `apps/b2bcnc/`)
- Bash/Zsh - Service management scripts (`scripts/*.sh`)

**Secondary:**
- TypeScript 3.6.4 - Laravel Mix compilation (`apps/mxcros/package.json`)
- SCSS/Sass 1.89.2+ - Styling for WordPress/Laravel (`apps/b2bcnc/wp-content/plugins/makerblocks/`)
- Python 3.12+ - Microservices support (`config/python/`)
- SQL - Database migrations and schemas

## Runtime

**Environment:**
- Node.js 20 LTS - Dashboard and build tooling (`config/node/Dockerfile`)
- PHP-FPM 8.3 - API and WordPress (`config/php/Dockerfile`)
- Python 3.12 - Data processing services (`config/python/Dockerfile`)
- Docker/Podman - Container orchestration with runtime switching

**Package Manager:**
- npm - JavaScript dependencies (lockfiles in each app)
- Composer - PHP dependencies (`apps/mxcros/`, `apps/b2bcnc/wp-content/plugins/makermaker/`)
- pip/Poetry 1.7.1 - Python packages (`config/python/requirements.txt`)

## Frameworks

**Core:**
- React 18.2.0 - Dashboard UI (`apps/dashboard/package.json`)
- Laravel 12.0 - PHP backend framework (`apps/mxcros/composer.json`)
- WordPress 6.2+ - CMS and Block Editor (`apps/b2bcnc/`)
- TypeRocket Pro v6 - WordPress MVC extension (`apps/b2bcnc/wp-content/mu-plugins/typerocket-pro-v6/`)

**Testing:**
- Pest 2.34+ - PHP testing (`apps/b2bcnc/wp-content/plugins/makermaker/composer.json`)
- PHPUnit 11.5.3 - Laravel testing (`apps/mxcros/composer.json`)
- pytest 7.4+ - Python testing (`config/python/requirements.txt`)
- ESLint 8.55.0 - JavaScript linting only (`apps/dashboard/.eslintrc.cjs`)

**Build/Dev:**
- Vite 5.0.8 - React bundler and dev server (`apps/dashboard/vite.config.js`)
- @wordpress/scripts 30.19.0 - Block and theme building (`apps/b2bcnc/wp-content/plugins/makerblocks/`)
- PostCSS 8.4.32 + Autoprefixer - CSS processing (`apps/dashboard/package.json`)

## Key Dependencies

**Critical:**
- @headlessui/react 2.2.9 - Accessible UI components (`apps/dashboard/package.json`)
- @heroicons/react 2.2.0 - Icon library (`apps/dashboard/package.json`)
- Tailwind CSS 3.4.0/4.0 - Utility-first CSS (`apps/dashboard/`, `apps/b2bcnc/`)

**Infrastructure:**
- Docker Compose - 166 service definitions in `compose/` (24 categories)
- Podman/Docker runtime switching - `scripts/runtime-switcher.sh`
- MariaDB, PostgreSQL, Redis - Databases (`compose/database/`)

**Python Services (available):**
- Flask 3.0+, FastAPI 0.100+, Django 4.2+ - API frameworks
- Celery 5.3+ with Redis - Async task queues
- SQLAlchemy 2.0+ - Python ORM

## Configuration

**Environment:**
- `.env` files at project root and per-app (`apps/mxcros/.env.example`)
- `VITE_API_BASE_URL` for dashboard API proxy (`apps/dashboard/vite.config.js`)
- Database credentials via `DB_*`, `MYSQL_*` env vars

**Build:**
- `apps/dashboard/vite.config.js` - React bundling, API proxy to `:8500`
- `apps/dashboard/.eslintrc.cjs` - ESLint with React plugins
- `apps/dashboard/tailwind.config.js` - Dark mode, custom colors
- `scripts/config.sh` - Service categories, port allocation, startup order

## Platform Requirements

**Development:**
- Linux/macOS/WSL (Docker/Podman required)
- Node.js 20+, PHP 8.2+, optional Python 3.12+
- Podman or Docker runtime

**Production:**
- Docker/Podman orchestration
- Multi-container deployment via Compose files
- Services connected via `microservices-net` bridge network

**Port Allocation:**
- PHP: 8100-8199
- Node.js: 8200-8299
- Python: 8300-8399
- Go: 8400-8499
- .NET: 8600-8699
- Rust: 8700-8799

---

*Stack analysis: 2026-01-07*
*Update after major dependency changes*
