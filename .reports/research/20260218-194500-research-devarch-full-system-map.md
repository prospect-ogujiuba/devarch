# Research Report: DevArch Full System Map

**Topic:** Complete system map for AlphaOne harness redesign — every piece of functionality, config, data flow, and dependency mapped
**Files analyzed:** 200+
**Agents:** 4 parallel researchers (API, Dashboard, Scripts/Infra, Planning/Design)

## Summary

DevArch is a ~25K LOC Go + ~18K LOC TypeScript local microservices dev environment at v1.1.1. Three milestones shipped (91 requirements, 28 phases). Core invariant: **stack isolation** — two stacks using the same service template must never collide on naming, networking, or ports. The system uses PostgreSQL as single source of truth, generates Compose YAML on-the-fly, and abstracts over Docker/Podman. The codebase has clear seams (identity, orchestration, compose generation, security) but also significant technical debt (N+1 queries, dead code paths, broken column references, disabled event loop).

---

## I. Architecture Overview

### System Components

| Component | Tech | LOC | Purpose |
|-----------|------|-----|---------|
| API | Go 1.24, chi, lib/pq | ~25K | Orchestration brain, REST+WS |
| Dashboard | React 19, TanStack, Tailwind 4 | ~18K | Management UI |
| CLI | zsh/bash | ~3K | Shell wrapper, API client |
| Services Library | YAML | ~160 templates | Service template catalog |
| Infrastructure | Docker/Podman Compose | - | Container orchestration |

### Data Flow

```
User → CLI/Dashboard → API (chi router) → PostgreSQL (source of truth)
                                        → Compose Generator (on-the-fly YAML)
                                        → Container Client (Podman/Docker CLI)
                                        → Podman Socket (status/metrics/exec)
```

### Core Invariants

1. **Stack isolation**: `devarch-{stack}-{instance}` naming, partial unique indexes, advisory locks
2. **DB is truth**: Compose YAML never stored, always generated from DB state
3. **Override-first**: Instance tables override template; env vars merge (template → wired → instance)
4. **Security modes**: dev-open / dev-keyed / strict — validated at startup

---

## II. Go API — Complete Map

### Entry Points

| Binary | Source | Purpose |
|--------|--------|---------|
| `devarch-api` | `cmd/server/main.go:68` | HTTP server, WS, sync loops |
| `devarch-migrate` | `cmd/migrate/main.go:19` | Schema migrations (up/down/status/create-db) |
| `devarch-import` | `cmd/import/main.go:14` | Service template import from YAML library |
| `devarch-export` | `cmd/export/main.go:14` | SQL seed file generation (categories+services only) |

### Server Startup Sequence (`cmd/server/main.go:68`)

1. Logger init (JSON slog, LOG_LEVEL env)
2. DB connect + ping (DATABASE_URL env)
3. Encryption key load/generate (`~/.devarch/secret.key`, AES-256-GCM)
4. Podman socket client (auto-detect socket path)
5. Container CLI client (auto-detect podman/docker)
6. Project controller + scanner (scan APPS_DIR)
7. Nginx generator (NGINX_GENERATED_DIR)
8. Registry manager (DockerHub + GHCR clients)
9. Security mode parse + validate (SECURITY_MODE env)
10. Sync manager start (3 goroutines: status/metrics/cleanup)
11. Router construction + HTTP server start

### Middleware Chain (`internal/api/routes.go:51`)

**Global (all routes):**
1. `RequestID` — generates X-Request-ID
2. `SlogMiddleware` — structured JSON logging with request_id in context
3. `RecoverEnvelope` — panic → 500 JSON response
4. `RealIP` — X-Forwarded-For → RemoteAddr
5. `CORS` — ALLOWED_ORIGINS env, AllowCredentials when non-wildcard

**Per /api/v1 group:**
6. `APIKeyAuth` — X-API-Key header, constant-time compare, skip in dev-open
7. `RateLimit` — 10 req/s, burst 50, in-memory token bucket, 10K visitor cap
8. `MaxBodySize` — 10MB default (256MB for stack import)

### Complete Route Inventory

**Public (no auth):**
- `GET /health` — healthcheck
- `GET /swagger/*` — OpenAPI docs
- `POST /api/v1/auth/validate` — validate API key
- `POST /api/v1/auth/ws-token` — HMAC-SHA256 WS token (60s TTL)

**Services (~30 endpoints):**
- CRUD: `GET/POST /services`, `GET/PUT/DELETE /services/{name}`
- Lifecycle: `POST /services/{name}/start|stop|restart|rebuild`
- Status: `GET /services/{name}/status|logs|metrics|compose|versions`
- Sub-resources (PUT): ports, volumes, env-vars, env-files, networks, config-mounts, dependencies, healthcheck, labels, domains
- Config files: `GET/PUT/DELETE /services/{name}/files/*`
- Exports/imports: `GET/PUT /services/{name}/exports|imports`
- Registry: `GET /services/{name}/image|tags|vulnerabilities`
- Proxy: `POST /services/{name}/proxy-config`
- Bulk: `POST /services/bulk`, `POST /services/import-library`

**Stacks (~30 endpoints):**
- CRUD: `GET/POST /stacks`, `GET/PUT/DELETE /stacks/{name}`
- Lifecycle: `POST /stacks/{name}/start|stop|restart|enable|disable|clone|rename`
- Trash: `GET /stacks/trash`, `POST /stacks/trash/{name}/restore`, `DELETE /stacks/trash/{name}`
- Compose: `GET /stacks/{name}/compose|plan|export`, `POST /stacks/{name}/apply`
- Network: `GET/POST/DELETE /stacks/{name}/network`
- Lock: `POST /stacks/{name}/lock|lock/validate|lock/refresh`
- Import: `POST /stacks/import` (256MB, outside group)
- Proxy: `POST /stacks/{name}/proxy-config`

**Instances (~25 endpoints per stack):**
- CRUD: `GET/POST /stacks/{name}/instances`, `GET/PUT/DELETE .../instances/{instance}`
- Lifecycle: `POST .../instances/{instance}/start|stop|restart|duplicate|rename`
- Override PUT: ports, volumes, env-vars, env-files, networks, config-mounts, labels, domains, healthcheck, dependencies
- Resources: `GET/PUT .../instances/{instance}/resources`
- Config files: `GET/PUT/DELETE .../instances/{instance}/files/*`
- Effective: `GET .../instances/{instance}/effective-config`
- Compose/logs: `GET .../instances/{instance}/compose|logs`
- Proxy: `POST .../instances/{instance}/proxy-config`

**Wires:**
- `GET /stacks/{name}/wires`
- `POST /stacks/{name}/wires` (create explicit)
- `POST /stacks/{name}/wires/resolve` (auto-wire)
- `POST /stacks/{name}/wires/cleanup`
- `DELETE /stacks/{name}/wires/{wireId}`

**Other:**
- Categories: CRUD + `/categories/{name}/services|start|stop` (8 endpoints)
- Projects: CRUD + scan + status + lifecycle (12 endpoints)
- Networks: list, create, bulk-remove, remove
- Registries: list, search, image route
- Images: list, inspect, remove, history, pull (streaming NDJSON), prune
- Containers: `GET /containers/{name}/exec` (WebSocket terminal)
- Runtime: status, switch, socket-status, socket-start
- Nginx: generate-all, generate-one, reload
- Status: overview, trigger-sync, sync-jobs
- WebSocket: `GET /ws/status` (real-time container status broadcast)

### Database Schema — 39+ Tables, 13 Migrations

| Migration | Tables | Purpose |
|-----------|--------|---------|
| 001 | categories, services | Service catalog |
| 002 | service_ports, service_volumes, service_env_vars, service_env_files, service_dependencies, service_healthchecks, service_labels, service_domains, service_networks | Service runtime shape |
| 003 | service_config_files, service_config_mounts, service_config_versions | Config model |
| 004 | registries, images, image_tags, image_architectures, vulnerabilities, image_tag_vulnerabilities | Registry |
| 005 | projects, project_services | Projects |
| 006 | stacks, service_instances | Stacks core (soft-delete) |
| 007 | instance_ports, instance_volumes, instance_env_vars, instance_labels, instance_domains, instance_healthchecks, instance_config_files, instance_dependencies, instance_resource_limits | Instance overrides |
| 008 | service_exports, service_import_contracts, service_instance_wires, sync_state, container_states, container_metrics | Wiring + sync |
| 009 | (indexes only) | BRIN on metrics, autovacuum tuning |
| 010 | instance_env_files, instance_networks, instance_config_mounts | New instance overrides |
| 011 | (FK constraint) | Projects require stack |
| 012 | (data patch) | Flowstate build context fix |
| 013 | sync_jobs | Persistent sync job history |

**Key Constraints:**
- `stacks.name`: UNIQUE WHERE deleted_at IS NULL (partial index)
- `service_instances(stack_id, instance_id)`: UNIQUE WHERE deleted_at IS NULL
- `service_instance_wires(stack_id, consumer_instance_id, import_contract_id)`: UNIQUE
- `service_domains.domain`: UNIQUE globally
- `service_ports(service_id, host_ip, host_port)`: UNIQUE per service

### Internal Packages

| Package | Purpose | Key Types/Functions |
|---------|---------|-------------------|
| `internal/api/routes.go` | Route definitions, middleware wiring | `SetupRoutes()` |
| `internal/api/handlers/` | HTTP handlers (~24 files, ~40KB for services alone) | One file per domain |
| `internal/api/respond/` | Response envelope helpers | `JSON()`, `Error()`, `Action()` |
| `internal/api/middleware/` | Auth, rate limit, logging, recovery | `APIKeyAuth()`, `RateLimit()`, `SlogMiddleware()` |
| `internal/compose/generator.go` | Service-level YAML generation | `Generator.Generate()` |
| `internal/compose/stack.go` | Stack-level YAML generation + effective config merge | `Generator.GenerateStack()`, `loadInstanceEffectiveConfigWithSecrets()` |
| `internal/compose/importer.go` | Service library YAML import | `Importer.ImportAll()` |
| `internal/container/client.go` | Docker/Podman CLI wrapper | `Client` struct, `StartService()`, `StopStack()`, etc. |
| `internal/podman/client.go` | Podman socket HTTP client | `Client` struct, image/exec/metrics operations |
| `internal/orchestration/service.go` | Plan/apply/wire resolution (transport-agnostic) | `Service.GeneratePlan()`, `Service.ApplyPlan()` |
| `internal/orchestration/errors.go` | Sentinel errors for HTTP status mapping | `ErrStackNotFound`, `ErrStalePlan`, etc. |
| `internal/identity/service.go` | Pure naming functions (no DB) | `ContainerName()`, `NetworkName()` |
| `internal/identity/validation.go` | Stack/instance name validation | `ValidateName()`, reserved words |
| `internal/identity/labels.go` | DevArch label management | `ManagedLabels()`, label prefix `devarch.` |
| `internal/wiring/resolver.go` | Auto-wire by contract type matching | `Resolver.Resolve()` |
| `internal/wiring/env_injector.go` | Template variable substitution | `{{hostname}}`, `{{port}}`, `{{protocol}}`, `{{name}}` |
| `internal/security/mode.go` | Security mode parsing/validation | `ParseMode()`, three modes |
| `internal/security/token.go` | HMAC-SHA256 WS token gen/validate | `GenerateToken()`, `ValidateToken()` |
| `internal/crypto/cipher.go` | AES-256-GCM encrypt/decrypt | `Cipher.Encrypt()`, `Cipher.Decrypt()` |
| `internal/crypto/keymanager.go` | Key file management | `LoadOrGenerateKey()` at `~/.devarch/secret.key` |
| `internal/sync/manager.go` | Background loops (status/metrics/cleanup) | `Manager.Start()`, 3 goroutines |
| `internal/nginx/generator.go` | Nginx config generation | `Generator.GenerateAll()`, framework-aware templates |
| `internal/project/controller.go` | Project lifecycle | `Controller.EnsureStack()`, `Controller.Scan()` |
| `internal/plan/staleness.go` | SHA-256 plan token generation | `ComputeToken()` |
| `internal/lock/` | Lock file generation/validation | `Generator.Generate()`, `Validator.Validate()` |
| `internal/export/` | devarch.yml import/export | `Exporter.Export()`, `Importer.Import()` |
| `pkg/models/models.go` | All DB model structs | `Service`, `Stack`, `ServiceInstance`, etc. |
| `pkg/registry/dockerhub/` | DockerHub API client | `Client.Search()`, `Client.GetTags()` |
| `pkg/registry/ghcr/` | GitHub Container Registry client | `Client.Search()`, `Client.GetTags()` |

### Go Dependencies (Direct)

| Package | Version | Purpose |
|---------|---------|---------|
| go-chi/chi/v5 | v5.1.0 | HTTP router |
| go-chi/cors | v1.2.1 | CORS middleware |
| gorilla/websocket | v1.5.3 | WebSocket |
| lib/pq | v1.10.9 | Postgres driver |
| stretchr/testify | v1.11.1 | Test assertions |
| swaggo/http-swagger/v2 | v2.0.2 | Swagger UI |
| swaggo/swag | v1.16.6 | OpenAPI generation |
| testcontainers-go | v0.40.0 | Integration test containers |
| yaml.v3 | v3.0.1 | YAML gen/parse |

---

## III. React Dashboard — Complete Map

### Build Stack

| Tool | Config | Purpose |
|------|--------|---------|
| Vite 7 | `vite.config.ts` | Build, dev server (:5174), proxy `/api`→:8550 |
| Tailwind 4 | CSS-first in `index.css` | Styling (oklch theme, dark mode) |
| TanStack Router | `@tanstack/router-plugin/vite` | File-based routing |
| TanStack Query | staleTime: 30s, retry: 1 | Server state |
| TypeScript | strict, ES2022, `@`→`src/` | Type system |

### Routes (File-Based)

| Route | Search Params | Purpose |
|-------|--------------|---------|
| `/` | - | Overview |
| `/login` | - | API key auth |
| `/services` | q, sort, dir, view, page, size, category, status | Service list |
| `/services/new` | image_name, image_tag | Create service |
| `/services/$name` | tab (info/env/logs/compose/files/proxy) | Service detail |
| `/stacks` | q, sort, dir, view, page, size | Stack list |
| `/stacks/$name` | tab (instances/compose/wiring/deploy/proxy) | Stack detail |
| `/stacks/$name/instances/$instance` | instanceTab (17 values) | Instance detail |
| `/categories` | q, sort, dir, view | Category list |
| `/projects` | q, sort, dir, view | Project list |
| `/projects/$name` | tab | Project detail |
| `/networks` | q, status | Network list |
| `/images` | (local state, not URL-synced) | Image list |
| `/registries` | q, registry | Registry search |
| `/registries/$registry/$` | (splat for image name) | Image tags |
| `/settings` | - | Runtime/socket config |

**Instance 17 Tabs:** info, ports, volumes, env-files, environment, networks, labels, domains, healthcheck, dependencies, config-mounts, files, resources, effective, logs, compose, proxy

### Key Patterns

| Pattern | Implementation | Location |
|---------|---------------|----------|
| Server state | TanStack Query hooks per feature domain | `src/features/*/queries.ts` |
| Mutations | `useMutationHelper` wrapper (toast + invalidation) | `src/lib/mutations.ts` |
| URL sync | `useUrlSyncedListControls` + `useUrlPagination` | `src/hooks/use-url-synced-list-controls.ts` |
| Editing | `useEditableSection` / `useOverrideSection` (draft-based) | `src/hooks/use-editable-section.ts` |
| Controllers | `use*DetailController` aggregating queries+mutations | `src/features/*/use*DetailController.ts` |
| Sub-resources | `makeSubResourceMutation` factory | `src/features/services/queries.ts:212` |
| List pages | `ListPageScaffold` component | `src/components/ui/list-page-scaffold.tsx` |
| Entity actions | `LifecycleButtons`, `EnableToggle`, `MoreActionsMenu` | `src/components/ui/entity-actions.tsx` |
| Auth | localStorage `devarch-api-key`, axios interceptor (401→redirect) | `src/lib/api.ts` |
| WebSocket | `/ws/status` with JWT token, invalidates query cache | `src/hooks/use-websocket.ts` |
| Terminal | xterm.js + WebSocket exec | `src/components/terminal/container-terminal.tsx` |
| Theme | localStorage `devarch-theme`, dark/light/system | `src/lib/theme.tsx` |

### Frontend Dependencies (Key)

| Package | Purpose |
|---------|---------|
| @tanstack/react-router | File-based routing + search params |
| @tanstack/react-query | Server state management |
| axios | HTTP client with envelope unwrap |
| zod | Route search param validation |
| sonner | Toast notifications |
| @radix-ui/* (8 packages) | Alert Dialog, Checkbox, Dialog, Dropdown Menu, Label, Select, Slot, Tabs |
| @codemirror/* (5 packages) | Code editor (JSON, YAML, XML) |
| @xterm/xterm | Container terminal |
| lucide-react | Icons |
| class-variance-authority | Component variants |

### Dead Code

- `features/containers/queries.ts` — unused by any route
- `useTrashStacks/useRestoreStack/usePermanentDeleteStack` — no UI
- `useImageInspect/useImageHistory` — never imported
- `useProjectServiceControl` — never imported
- `useCleanupOrphanedWires` — no UI action
- `yaml` and `cmdk` npm packages — not imported

---

## IV. Scripts/CLI — Complete Map

### CLI Entry Point (`scripts/devarch`, zsh)

| Namespace | Aliases | Handler | Purpose |
|-----------|---------|---------|---------|
| service | svc, s | `service-manager.sh` | Service lifecycle (API client) |
| wp | - | Podman exec WP-CLI passthrough | WordPress CLI |
| artisan | art, a | Podman exec artisan passthrough | Laravel CLI |
| wordpress | wf | `wordpress/wp-workflow.sh` | WP install/backup/restore |
| laravel | lara, l | `laravel/setup-laravel.sh` | Laravel project setup |
| socket | sock | `socket-manager.sh` | Podman socket management |
| runtime | rt | `runtime-switcher.sh` | Docker/Podman switching |
| db | database | `init-databases.sh` | DB initialization |
| context | ctx, c | `generate-context.sh` | Context file generation |
| run | r | - | Arbitrary container exec |
| shell | sh | - | Interactive shell (zsh→bash→sh fallback) |
| init | - | `devarch-init.sh` | Bootstrap from devarch.yml |
| doctor | doc | `devarch-doctor.sh` | System diagnostics |

### Config (`scripts/config.sh`)

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONTAINER_RUNTIME` | podman | Runtime binary |
| `NETWORK_NAME` | microservices-net | Shared network |
| `COMPOSE_IGNORE_ORPHANS` | true | Suppress warnings |
| `PODMAN_USERNS` | keep-id | User namespace |
| Port ranges | PHP:8100-8199, Node:8200-8299, Python:8300-8399, Go:8400-8499 | (documented, not enforced) |
| Default passwords | 123456 for all DBs | Overridable by .env |
| `OS_TYPE` | auto-detect | WSL/Linux/macOS |

### Service Library (`services-library/`)

**24 categories, ~160 templates.** Template format:
```yaml
services:
  {name}:
    container_name: {name}
    image: {image}:{tag}
    ports:
      - "127.0.0.1:{host}:{container}"
    volumes:
      - {name}_data:/path
    networks:
      microservices-net:
        external: true
    restart: unless-stopped
    x-devarch-config:          # optional
      {filename}: {container-path}
volumes:
  {name}_data: null
```

**Categories:** ai (7), analytics, apps, backend (12), ci, collaboration, database (14), dbms, docs, erp, exporters, gateway, mail, management, messaging, project, proxy, registry, search, security, storage, support, testing, workflow

---

## V. Configuration — Complete Map

### API Environment Variables

| Env Var | Default | Required | Purpose |
|---------|---------|----------|---------|
| `DATABASE_URL` | `postgres://devarch:devarch@localhost:5432/devarch` | Yes | DB connection |
| `PORT` | 8080 | No | Listen port |
| `DEVARCH_API_KEY` | (none) | For dev-keyed/strict | Auth key |
| `SECURITY_MODE` | dev-open | No | Auth behavior |
| `ALLOWED_ORIGINS` | * | No | CORS origins |
| `STACK_IMPORT_MAX_BYTES` | 256MB | No | Import body cap |
| `PROJECT_ROOT` | (none) | For apply | Config materialization root |
| `HOST_PROJECT_ROOT` | (none) | For apply | Host path for volume mounts |
| `WORKSPACE_ROOT` | (none) | For apply | Workspace alias |
| `APPS_DIR` | /workspace/apps | No | Project scan dir |
| `NGINX_GENERATED_DIR` | /workspace/config/nginx/generated | No | Nginx output |
| `DEVARCH_RUNTIME` | (auto-detect) | No | Force podman/docker |
| `DEVARCH_USE_SUDO` | false | No | Prepend sudo to compose |
| `CONTAINER_HOST` | (auto-detect socket) | No | Podman socket path |
| `LOG_LEVEL` | info | No | debug/warn/error/info |
| `DEVARCH_METRICS_INTERVAL` | 30s | No | Metrics poll interval |
| `DEVARCH_METRICS_RETENTION` | 3d | No | Metrics DB retention |
| `DEVARCH_CLEANUP_INTERVAL` | 1h | No | Cleanup job interval |
| `DEVARCH_CLEANUP_BATCH` | 50000 | No | Cleanup batch size |
| `DEVARCH_CONFIG_VERSIONS_MAX` | 25 | No | Config version cap |
| `DEVARCH_REGISTRY_IMAGE_RETENTION` | 30d | No | Registry data retention |
| `DEVARCH_VULN_ORPHAN_RETENTION` | 30d | No | Vuln data retention |
| `DEVARCH_SOFT_DELETE_RETENTION` | 30d | No | Soft-delete purge window |

### Infrastructure (`compose.yml`)

3 containers: devarch-db (postgres:16-alpine :5433), devarch-api (Go, hot-reload via air :8550→:8080), devarch-app (nginx serving dashboard/dist :5174→:80)

---

## VI. Requirements — Complete Map

### v1.0: Stacks & Instances (66 requirements)

| Category | Count | Focus |
|----------|-------|-------|
| BASE | 3 | Foundation |
| STCK | 8 | Stack CRUD, isolation, naming |
| INST | 12 | Instance CRUD, overrides, lifecycle |
| NETW | 4 | Stack networking |
| COMP | 3 | Compose generation |
| PLAN | 5 | Plan/apply workflow |
| WIRE | 8 | Contract-based auto-wiring |
| EXIM | 6 | Export/import (devarch.yml) |
| BOOT | 3 | Bootstrap/lock files |
| LOCK | 3 | Lock validation |
| SECR | 4 | Secrets encryption |
| RESC | 3 | Resource limits |
| MIGR | 5 | DB migrations |

### v1.1: Schema Reconciliation (36 requirements)

| Category | Count | Focus |
|----------|-------|-------|
| SCHM | 12 | Domain-separated migrations (39 tables) |
| PARS | 7 | Template parsing |
| GENR | 4 | Generator pipeline |
| IMPT | 6 | Import pipeline |
| DASH | 4 | Dashboard sync |
| VALD | 4 | Validation |

### v1.1.1: Architecture Hardening (25 requirements)

| Category | Count | Focus |
|----------|-------|-------|
| SEC | 5 | Security modes, WS tokens, rate limiting |
| API | 4 | Response envelopes, structured logging |
| BE | 3 | Orchestration extraction, sentinel errors |
| PERF | 3 | Advisory locks, cleanup optimization |
| FE | 5 | Controller hooks, mutation helpers, URL sync |
| TEST | 3 | Integration + frontend tests |
| OPS | 2 | CI workflows |

---

## VII. Risks & Technical Debt

### Critical

| Issue | File | Impact |
|-------|------|--------|
| `loadAllProviders/loadAllConsumers` reference `si.service_id` (non-existent column; should be `template_service_id`) | `orchestration/service.go:496` | Wire resolution (`POST /wires/resolve`) crashes at runtime |
| `container_states` sync tracks by `services.name`, not instance container names | `sync/manager.go:176` | Instance containers never reflected in status |
| AES key at `~/.devarch/secret.key` not volume-mounted | `crypto/keymanager.go:13` | Container recreation loses all encrypted env vars |

### Medium

| Issue | File |
|-------|------|
| N+1 DB queries in compose generation (~10 queries per instance) | `compose/stack.go:463` |
| Migration 012 embeds app-specific data in schema migration | `migrations/012` |
| compose_overrides merge is shallow (marshal→unmarshal→merge) | `compose/generator.go:196` |
| Services fetched with limit=500 in single request (no server pagination) | `dashboard/features/services/queries.ts:15` |
| No React ErrorBoundary anywhere | `dashboard/src/routes/__root.tsx` |
| Auth endpoints not rate-limited | `api/routes.go:87` |
| http.Server 15s write timeout may kill slow compose applies | `cmd/server/main.go:172` |
| Runtime-switcher calls non-existent `stop-all` subcommand | `scripts/runtime-switcher.sh:57` |
| setup-laravel.sh / install-wordpress.sh call dead functions | `scripts/laravel/setup-laravel.sh:125` |

### Low / Dead Code

| Issue | Location |
|-------|----------|
| `GetCategoryRunningCount` always returns 0 | `container/client.go:362` |
| Event loop disabled (30s polling only) | `sync/manager.go:142` |
| `ErrNotImplemented` defined but never used | `container/client.go:18` |
| Dashboard dead code: containers feature, trash UI, image inspect/history, project service control, orphaned wires cleanup | Various `features/*/queries.ts` |
| Unused npm deps: yaml, cmdk | `dashboard/package.json` |
| Stale context tmp files | `context/*.tmp.*` |
| generate-context.sh targets non-existent compose/ and config/ dirs | `scripts/generate-context.sh:6` |
| Port range documentation doesn't match actual template ports | `scripts/config.sh:10-18` |

---

## VIII. AlphaOne Redesign — Key Seams & Recommendations

### Preserve As-Is (Clean Abstractions)

1. **Identity package** — pure functions, no DB, deterministic naming (`internal/identity/`)
2. **Orchestration service boundary** — Go types only, no net/http, sentinel errors (`internal/orchestration/`)
3. **Security modes** — three-tier with startup validation (`internal/security/`)
4. **Response envelope** — SuccessEnvelope/ErrorEnvelope/ActionResponse (`internal/api/respond/`)
5. **Plan/apply with staleness tokens** — SHA-256 token invalidation on any mutation
6. **Advisory lock per stack** — pg_try_advisory_lock prevents concurrent applies
7. **Soft-delete with partial unique indexes** — stacks and instances
8. **Wiring contract system** — type-based auto-wire with explicit override

### Redesign Targets

1. **DB access layer** — replace raw sql.DB N+1 with batched queries or lightweight query builder
2. **Container abstraction** — merge CLI client + socket client into single interface
3. **Compose generation** — replace marshal/unmarshal merge with proper deep-merge
4. **Sync manager** — re-enable event-driven updates, track instance containers (not just service templates)
5. **Config management** — single config struct loaded at startup instead of scattered os.Getenv
6. **Scripts** — many broken references; CLI is a thin API client now, could be simplified further
7. **Dashboard state** — standardize Images/Registries pages to use URL-sync pattern; add ErrorBoundary
8. **Dead code** — significant cleanup opportunity in both API and dashboard

### Feedback Loop Architecture (AlphaOne)

The system's natural feedback loops are:
- **Plan → Apply → Status** (generate plan, execute, monitor via WS/sync)
- **Template → Instance → Override → Effective Config** (layered composition)
- **Wire Resolve → Env Inject → Compose Generate** (dependency resolution → config materialization)
- **Import → Validate → Materialize → Run** (service library → running containers)

For an AlphaOne harness that "learns from 0":
- The plan/apply cycle is the primary action-feedback loop
- The wiring system provides automatic discovery (contract matching)
- The effective config merge provides the composability primitive
- The sync manager provides the observation channel (container states → DB → WS → dashboard)

---

## IX. Complete File Index

### API (`api/`)
```
cmd/server/main.go          — server entry, wiring
cmd/migrate/main.go         — migration runner
cmd/import/main.go          — service template import
cmd/export/main.go          — SQL seed export
internal/api/routes.go      — all route definitions
internal/api/handlers/       — 24 handler files
internal/api/respond/        — response envelope
internal/api/middleware/      — auth, rate limit, logging, recovery
internal/compose/generator.go — service YAML generation
internal/compose/stack.go    — stack YAML generation + effective config
internal/compose/importer.go  — service library import
internal/compose/validator.go  — template validation
internal/container/client.go  — CLI wrapper (compose operations)
internal/container/types.go   — Runtime interface
internal/podman/client.go    — socket HTTP client (status/metrics/exec)
internal/orchestration/       — plan/apply/wire (transport-agnostic)
internal/identity/            — naming, validation, labels (pure functions)
internal/wiring/              — resolver, env injector
internal/security/            — modes, token
internal/crypto/              — AES-256-GCM cipher, key manager
internal/sync/manager.go     — background loops
internal/nginx/generator.go  — nginx config gen
internal/project/controller.go — project lifecycle
internal/plan/staleness.go   — SHA-256 plan token
internal/lock/                — lock file gen/validate
internal/export/              — devarch.yml import/export
pkg/models/models.go         — all DB model structs
pkg/registry/                — dockerhub + ghcr clients
migrations/001-013           — 13 SQL migrations
tests/integration/           — 8 integration test files
```

### Dashboard (`dashboard/`)
```
src/routes/                  — 16 file-based routes
src/features/*/queries.ts    — per-domain TanStack Query hooks
src/features/*/use*Controller — detail page controllers
src/components/ui/           — Radix wrappers, scaffolds, entity actions
src/components/services/     — editors, tables, log viewer
src/components/stacks/       — wiring tab, plan tab
src/components/terminal/     — xterm container terminal
src/components/proxy/        — proxy config panel
src/components/layout/       — header, sidebar
src/hooks/                   — editable sections, URL sync, websocket, pagination
src/lib/api.ts               — axios singleton, auth
src/lib/mutations.ts         — useMutationHelper
src/lib/pagination.ts        — useUrlPagination
src/lib/theme.tsx            — dark/light/system
src/lib/nav-items.ts         — navigation definitions
src/types/api.ts             — API type definitions
```

### Scripts (`scripts/`)
```
devarch                      — CLI entry point (zsh)
config.sh                    — shared config/state
service-manager.sh           — API client for service lifecycle
runtime-switcher.sh          — docker/podman switching
socket-manager.sh            — podman socket management
devarch-init.sh              — bootstrap from devarch.yml
devarch-doctor.sh            — system diagnostics
generate-context.sh          — context file generation
init-databases.sh            — DB initialization
setup-aliases.sh             — shell alias setup
laravel/setup-laravel.sh     — Laravel project setup
wordpress/wp-workflow.sh     — WP install/backup/restore
wordpress/install-wordpress.sh — WP preset installer
```
