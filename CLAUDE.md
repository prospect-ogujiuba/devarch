# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DevArch is a microservices development environment managing multiple backend runtimes, databases, and services via Docker/Podman Compose. The database is the source of truth — compose YAML is generated on-the-fly from DB state by the Go API.

## Essential Commands

```bash
# Service management (requires API running at localhost:8550)
./scripts/service-manager.sh up <service>           # Start service
./scripts/service-manager.sh down <service>         # Stop service
./scripts/service-manager.sh rebuild <service>      # Rebuild and restart (--no-cache)
./scripts/service-manager.sh logs <service> -f      # Follow logs
./scripts/service-manager.sh start <category>       # Start category
./scripts/service-manager.sh stop <category>        # Stop category
./scripts/service-manager.sh status                 # Show all statuses
./scripts/service-manager.sh list                   # List available services
./scripts/service-manager.sh compose <service>      # Show generated compose YAML
./scripts/service-manager.sh check                  # Verify runtime + API

# Import compose files into DB (one-time or re-import)
cd api && go run cmd/import/main.go --compose-dir ../compose --project-root .. --db $DATABASE_URL

# Dashboard development
cd apps/dashboard && npm run dev    # Start Vite dev server
cd apps/dashboard && npm run build  # Production build
cd apps/dashboard && npm run lint   # ESLint check
```

## Architecture

### Directory Structure
- `api/` - Go API backend (compose generation, service CRUD, container ops)
- `config/<service>/` - Dockerfiles and service configs
- `config/devarch-api/compose.yml` - Bootstrap compose for the API itself
- `scripts/` - CLI wrapper (service-manager.sh) and runtime config (config.sh)
- `apps/dashboard/` - React management UI (Vite + Tailwind + HeadlessUI)

### Key Components
- `api/internal/compose/generator.go` - Generates compose YAML from DB
- `api/internal/compose/importer.go` - Imports compose files into DB
- `api/internal/compose/parser.go` - Parses compose YAML files
- `scripts/config.sh` - Runtime detection, network config, env vars
- `scripts/service-manager.sh` - Thin API client CLI

### Service Categories (defined in database)
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
- Service configs: placed in `config/<service>/`
- Compose is generated from DB, not stored as files
