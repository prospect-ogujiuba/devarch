# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DevArch is a microservices development environment managing multiple backend runtimes, databases, and services via Docker/Podman Compose. One compose file per service.

## Essential Commands

```bash
# Service management (primary tool)
./scripts/service-manager.sh up <service>           # Start service
./scripts/service-manager.sh down <service>         # Stop service
./scripts/service-manager.sh rebuild <service>      # Rebuild and restart
./scripts/service-manager.sh logs <service> -f      # Follow logs
./scripts/service-manager.sh start <category>       # Start category (database, backend, etc.)
./scripts/service-manager.sh status                 # Show all statuses
./scripts/service-manager.sh list                   # List available services

# Common options
--sudo              # Use sudo for Docker/Podman
--no-cache          # Rebuild without cache
--remove-volumes    # Remove volumes on stop
--force-recreate    # Force container recreation

# Dashboard development
cd apps/dashboard && npm run dev    # Start Vite dev server
cd apps/dashboard && npm run build  # Production build
cd apps/dashboard && npm run lint   # ESLint check
```

## Architecture

### Directory Structure
- `compose/<category>/<service>.yml` - Compose files (one per service)
- `config/<service>/` - Dockerfiles and service configs
- `scripts/` - Management scripts (service-manager.sh is primary)
- `apps/dashboard/` - React management UI (Vite + Tailwind + HeadlessUI)
- `config/devarch/api/` - PHP API backend for dashboard

### Key Components
- `scripts/config.sh` - Central config: service categories, path resolution, runtime detection
- `scripts/service-manager.sh` - Unified CLI for all service operations
- `config/devarch/api/public/index.php` - API router for dashboard

### Service Categories (defined in config.sh)
database, storage, dbms, security, registry, gateway, proxy, management, backend, ci, project, mail, exporters, analytics, messaging, search, workflow, docs, testing, collaboration, erp, support, ai

### Port Allocation (backend runtimes)
- PHP: 8100-8199
- Node.js: 8200-8299
- Python: 8300-8399
- Go: 8400-8499
- .NET: 8600-8699
- Rust: 8700-8799

### Network
All services connect to `microservices-net` bridge network. Use service names as hostnames.

## Dashboard Tech Stack
React 18, Vite, Tailwind CSS, @headlessui/react, @heroicons/react

## Code Conventions
- Use existing HeadlessUI components
- Use @heroicons/react icons
- Service names: lowercase, hyphen-separated
- Compose files: one service per file, `<service>.yml`
- Configs: placed in `config/<service>/`
