# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

DevArch — local microservices dev environment. Go API + React dashboard + bash CLI. DB (Postgres) is source of truth; compose YAML generated on-the-fly. Runtime: Podman or Docker.

Current milestone: **Stacks & Instances** — isolated, composable service groups. Core invariant: two stacks using the same service template must never collide.

## Commands

### API (from `api/`)
```bash
go run ./cmd/server                              # start API server
go run ./cmd/migrate -migrations ./migrations     # run migrations
go run ./cmd/import -compose-dir ../apps          # import service templates
```

### Dashboard (from `dashboard/`)
```bash
npm run dev            # vite dev server on :5174
npm run build          # production build
npm run build:strict   # tsc + vite build
npm run lint           # eslint
```

### Docker Compose (from root)
```bash
docker compose up      # postgres :5433, api :8550
```

## Architecture

**API** (`api/`): Go 1.22, chi router, lib/pq, gorilla/websocket, yaml.v3
- `cmd/server` — entry point, wires DB + container client + router
- `cmd/migrate` — migration runner
- `cmd/import`, `cmd/export` — compose import/export
- `internal/api/routes.go` — all route definitions, chi middleware chain
- `internal/api/handlers/` — HTTP handlers (service CRUD is the largest at ~40KB)
- `internal/compose/` — YAML generation (`generator.go`), parsing, importing, validation
- `internal/container/client.go` — Docker/Podman abstraction layer
- `internal/podman/` — Podman-specific implementation
- `internal/nginx/` — nginx config generation
- `internal/project/` — project controller
- `migrations/` — 12 SQL migrations (001-012)

**Dashboard** (`dashboard/`): React 19, Vite, TanStack Router (file-based) + Query, Tailwind 4, Radix UI, Zod, CodeMirror
- `src/routes/` — page components (services, projects, categories, settings)
- `src/components/services/` — service-specific UI (editors, tables, log viewer)
- `src/types/api.ts` — API type definitions
- `@` alias maps to `src/`
- Vite proxies `/api` → `localhost:8550`

**Scripts** (`scripts/`): bash CLI wrapper (`devarch`), service manager, runtime switcher, config management

**Apps** (`apps/`): sample containerized applications mounted read-only into API container

## Key Patterns

- API key auth via `X-API-Key` header + rate limiting middleware on all `/api/v1` routes
- Services identified by `{name}` in URL paths, not IDs
- Compose YAML is never stored — always generated from DB state via `compose/generator.go`
- WebSocket at `/api/v1/ws/status` for real-time container status
- DB connection string via `DATABASE_URL` env var, API port via `PORT`

## Planning

`.planning/` contains project management artifacts (PROJECT.md, REQUIREMENTS.md, ROADMAP.md, STATE.md). 48 requirements across categories: BASE, STCK, INST, NETW, COMP, PLAN, WIRE, EXIM, SECR, RESC, MIGR. 9-phase roadmap.
