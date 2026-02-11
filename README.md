# DevArch User Guide

## What is DevArch?

DevArch is a local microservices development environment. You define containerized services (postgres, redis, nginx, etc.), group them into **stacks**, and DevArch handles isolation, networking, compose generation, and deployment. The only prerequisite is Docker or Podman.

**Core invariant:** Two stacks using the same service template never collide — each gets its own network, container names, and config.

**Architecture:** Go API + React dashboard + bash CLI. PostgreSQL is the source of truth; compose YAML is generated on-the-fly from DB state and never stored.

---

## 1. Setup & First Launch

### Prerequisites
- Docker or Podman installed
- The external bridge network must exist before starting

```bash
# Create the shared network (one-time)
podman network create microservices-net
# or: docker network create microservices-net
```

### Start with compose

The root `compose.yml` runs three containers:

| Container | Image | Host Port | Purpose |
|-----------|-------|-----------|---------|
| `devarch-db` | postgres:16-alpine | 5433 | PostgreSQL database |
| `devarch-api` | ./api (Dockerfile, dev target) | 8550 | Go API with air hot-reload |
| `devarch-app` | ./dashboard (Dockerfile) | 5174 | React dashboard (nginx) |

```bash
# From project root — start all three services
podman compose up
# or: docker compose up
```

The API entrypoint automatically runs migrations and starts the server with hot-reload via air. No manual migration step needed when using compose.

### Import the service catalog

```bash
# Run inside the API container
podman exec devarch-api devarch-import \
  -compose-dir /workspace/services-library \
  -config-dir /workspace/services-library/config
```

Or run the import command directly if developing locally (outside compose):

```bash
cd api
go run ./cmd/import -compose-dir ../services-library -config-dir ../services-library/config
```

### Local development (without compose)

```bash
# Start only the database
podman compose up devarch-db

# Run API directly (from api/)
cd api && go run ./cmd/server

# Run migrations manually (from api/)
go run ./cmd/migrate -migrations ./migrations

# Start the dashboard dev server (from dashboard/)
cd dashboard && npm run dev
# Dashboard at http://localhost:5174
```

### Authentication

- Set `DEVARCH_API_KEY` env var on the API to enable auth
- If `DEVARCH_API_KEY` is unset, authentication is disabled entirely (warning logged once)
- Uses constant-time comparison on the `X-API-Key` header
- The dashboard login page prompts for the key and stores it in session storage
- The compose.yml ships with a pre-configured API key

---

## 2. Core Concepts

### Services (Templates)

173 pre-built container definitions across 23 categories. Each service template defines: image (or custom Dockerfile via build context), ports, volumes, env vars, healthchecks, labels, domains, dependencies, config files, networks, restart policy, and command.

You **don't run services directly** — you instantiate them inside stacks. Services also support `compose_overrides` (JSONB) for non-standard compose keys like `build:`, `user:`, `cap_add:`, etc.

### Stacks

Isolated environment namespaces. Creating a stack gives you:
- A dedicated bridge network: `devarch-{stack}-net`
- Deterministic container naming: `devarch-{stack}-{instance}`
- Container labels: `devarch.stack_id={stack}`, `devarch.instance_id={instance}`
- Isolated compose YAML generation
- Advisory locking for safe concurrent operations

### Instances

When you add a service to a stack, you create an **instance** — a copy-on-write deployment of that service template. Override anything per-instance: ports, volumes, env vars, env files, networks, labels, domains, healthchecks, dependencies, config mounts, config files, resource limits (CPU/memory). The template stays pristine. Effective config = template values merged with instance overrides (overrides always win).

### Wiring

Services declare **export contracts** (what they provide, e.g., "postgres") and **import contracts** (what they need). When you add instances to a stack, DevArch can auto-resolve wires: if instance A needs "postgres" and exactly one instance provides it, the wire is created automatically. Wired connections inject environment variables (e.g., `DB_HOST`, `DB_PORT`) into consuming instances. Ambiguous cases (multiple providers) require explicit manual wiring.

### Projects

Scanned from the `apps/` directory. The scanner auto-detects project types:
- **WordPress** — detected via `wp-config.php`, extracts plugins/themes
- **Laravel** — detected via `artisan` file, extracts dependencies
- **Custom** — parses `package.json`, `composer.json` for metadata

Projects extract: dependencies, scripts, git info, version, language, framework. Projects can be linked to stacks to manage services through the stack workflow.

### Categories

Organizational groups for services. 23 categories: ai, analytics, backend, ci, collaboration, database, dbms, docs, erp, exporters, gateway, mail, management, messaging, project, proxy, registry, search, security, storage, support, testing, workflow.

---

## 3. Dashboard Pages

Navigate via the left sidebar. All list pages sync state to URL for bookmarkability (search, sort, pagination, filters, active tab).

### Login (`/login`)
API key entry form. Key stored in session storage for subsequent requests.

### Overview (`/`)
Stat cards: total services, running/stopped counts, categories. Browse categories with search, sort, and status filters.

### Stacks (`/stacks`)
**The primary workspace.** List all stacks with search, sort, status filters.

**Stack actions:** Create, Clone, Rename, Edit description, Enable/Disable, Delete (soft-delete to trash), Create/Remove network, Start/Stop/Restart all instances, Export YAML, Import from file.

**Stack detail** (`/stacks/{name}`) has tabs:
- **Instances** — Grid of all service instances. Per-instance: start/stop/restart, enable/disable, duplicate, delete. Click through to instance detail.
- **Compose** — View/download the generated Docker Compose YAML (always generated from DB state, never hand-written).
- **Wiring** — Shows active wires between instances, injected env vars, warnings for ambiguous/orphaned wires. Create manual wires or auto-resolve.
- **Deploy** — Generate a terraform-style plan (add/modify/remove preview), then Apply to execute. Stale plans (stack changed since generation) are rejected.
- **Proxy** — Generate reverse proxy config for the stack's domains.

### Instance Detail (`/stacks/{name}/instances/{instance}`)
Full override editor with tabs for every configurable aspect:
- **Info**: container name, description, enabled toggle
- **Ports**: override host/container port mappings, protocol
- **Volumes**: override mounts (bind/volume/tmpfs)
- **Environment**: override env vars (with secret masking)
- **Env Files**: override env file paths
- **Networks**: override network attachments
- **Labels**: override Docker labels
- **Domains**: override domains and proxy ports
- **Healthcheck**: override health check test, intervals, timeouts
- **Dependencies**: override service dependencies
- **Config Mounts**: override config file mounts
- **Config Files**: manage config file content with syntax highlighting (CodeMirror)
- **Resources**: CPU/memory limits and reservations
- **Effective Config**: read-only merged view (template + your overrides + wired env vars)

Actions: Edit description, Duplicate, Rename, Delete (with delete preview).

### Services (`/services`)
Browse and manage the service template catalog. Search by name/image, filter by status/category, sort by name/status. Grid or table view.

**Service detail** (`/services/{name}`) tabs:
- **Info**: status, image, restart policy, metrics, inline editors for ports/volumes/env/deps/labels/domains/healthcheck/config mounts
- **Environment**: env var editor with secret flag
- **Logs**: live streaming container logs with search
- **Compose**: generated YAML preview
- **Files**: config file editor with syntax highlighting
- **Proxy**: reverse proxy config generation

**Create service** (`/services/new`): form with all fields — name, category, image, tag (with tag picker from registry), restart policy, command, ports, volumes, env vars, dependencies, healthcheck, labels, domains.

### Projects (`/projects`)
Browse auto-detected projects from `apps/` directory. Search by name/framework/language. Filter by type or language.

**Project detail** (`/projects/{name}`) tabs:
- **Services**: stack-linked instances or compose services with status/controls
- **Info**: path, type, framework, language, package manager, version, domain, proxy port
- **Dependencies**: project dependencies by type
- **Plugins/Themes** (WordPress): installed plugins and themes
- **Scripts**: npm/composer scripts
- **Git**: remote URL, branch info
- **Proxy**: reverse proxy config

**Stack linking**: link a project to a stack to manage its services through the stack workflow.

### Categories (`/categories`)
Browse service categories with search, sort (by name, service count, startup order), and status filter.

### Networks (`/networks`)
Manage Docker/Podman networks. View managed vs. external vs. orphaned networks. Create new networks, bulk-remove orphaned ones.

### Registries (`/registries`)
Search Docker images across registries (DockerHub and GitHub Container Registry). Browse results with star/pull counts. Used by the tag picker when creating/editing services.

### Settings (`/settings`)
- **Container Runtime**: see Docker/Podman status, versions. Switch between runtimes.
- **Podman Socket**: manage rootless/rootful socket status.
- **API Key**: view (obscured), copy.

---

## 4. Typical Workflow

### Creating a development environment from scratch

1. **Create a stack**: Stacks > "New Stack" > name it (e.g., `my-project`)
2. **Add instances**: In stack detail > Instances tab > "Add Instance" > pick a template (e.g., postgres, redis, nginx) > give it an instance ID
3. **Configure overrides**: Click into each instance > override ports (avoid collisions with other stacks), set env vars (database passwords, etc.), mount config files
4. **Review wiring**: Wiring tab shows auto-detected connections. Click "Auto-Resolve" to wire instances that match import/export contracts
5. **Plan**: Deploy tab > "Generate Plan" — see what containers will be created/modified/removed
6. **Apply**: "Apply Plan" — DevArch locks the stack, ensures the network exists, materializes config files to disk, generates compose YAML, and runs compose up
7. **Monitor**: View container status in real-time via WebSocket updates. Check logs in Services > {name} > Logs tab.

### Sharing an environment

```
# Export stack
GET /api/v1/stacks/{name}/export  ->  devarch.yml

# Generate lock file (config hash + container IDs)
POST /api/v1/stacks/{name}/lock   ->  devarch.lock

# Share devarch.yml + devarch.lock

# On another machine — import, plan, apply
POST /api/v1/stacks/import  (with devarch.yml)
GET  /api/v1/stacks/{name}/plan
POST /api/v1/stacks/{name}/apply
```

Secrets are redacted in exports (env vars marked `is_secret` replaced with placeholders). The recipient provides their own secrets.

### Bootstrap via CLI

```bash
# Import + pull images + apply in one step
devarch init devarch.yml
```

---

## 5. API Reference

All endpoints under `/api/v1/` require `X-API-Key` header (when auth is enabled). Rate limited: 10 req/sec, burst 50 per IP. Default body size limit: 10MB (stack import: 256MB, configurable via `STACK_IMPORT_MAX_BYTES`).

### Health & Auth
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Liveness check (root-level, no auth) |
| POST | `/api/v1/auth/validate` | Validate API key (no rate limit) |

### Status & Sync
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/status` | System overview (service counts, runtime info) |
| POST | `/api/v1/sync` | Trigger manual sync |
| GET | `/api/v1/sync/jobs` | Background job status |
| WS | `/api/v1/ws/status` | Real-time container status (WebSocket) |

### Runtime
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/runtime/status` | Docker/Podman availability |
| POST | `/api/v1/runtime/switch` | Switch between runtimes |
| GET | `/api/v1/socket/status` | Podman socket status |
| POST | `/api/v1/socket/start` | Start Podman socket |

### Stacks
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/stacks` | List / Create |
| GET/PUT/DELETE | `/stacks/{name}` | Get / Update / Soft-delete |
| POST | `/stacks/{name}/clone` | Clone with new name |
| POST | `/stacks/{name}/rename` | Rename stack |
| POST | `/stacks/{name}/enable` | Enable stack |
| POST | `/stacks/{name}/disable` | Disable (stops containers) |
| POST | `/stacks/{name}/start\|stop\|restart` | Lifecycle operations |
| GET | `/stacks/{name}/network` | Network status |
| POST/DELETE | `/stacks/{name}/network` | Create / Remove network |
| GET | `/stacks/{name}/compose` | Get generated compose YAML |
| GET | `/stacks/{name}/plan` | Preview deployment changes |
| POST | `/stacks/{name}/apply` | Execute deployment |
| GET | `/stacks/{name}/export` | Export as devarch.yml |
| POST | `/stacks/import` | Import devarch.yml (256MB limit) |
| GET | `/stacks/{name}/delete-preview` | Preview what delete affects |
| POST | `/stacks/{name}/lock` | Generate lock file |
| POST | `/stacks/{name}/lock/validate` | Validate lock against state |
| POST | `/stacks/{name}/lock/refresh` | Refresh stale lock |
| POST | `/stacks/{name}/proxy-config` | Generate proxy config |

### Stack Trash (Soft-Delete Recovery)
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/stacks/trash` | List deleted stacks |
| POST | `/stacks/trash/{name}/restore` | Restore deleted stack |
| DELETE | `/stacks/trash/{name}` | Permanently delete |

### Instances (nested under stacks)
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/stacks/{name}/instances` | List / Create |
| GET/PUT/DELETE | `/stacks/{name}/instances/{id}` | Get / Update / Delete |
| POST | `/stacks/{name}/instances/{id}/duplicate` | Duplicate instance |
| PUT | `/stacks/{name}/instances/{id}/rename` | Rename instance |
| GET | `/stacks/{name}/instances/{id}/delete-preview` | Preview delete impact |
| POST | `/stacks/{name}/instances/{id}/start\|stop\|restart` | Lifecycle |
| PUT | `/stacks/{name}/instances/{id}/ports` | Override ports |
| PUT | `/stacks/{name}/instances/{id}/volumes` | Override volumes |
| PUT | `/stacks/{name}/instances/{id}/env-vars` | Override env vars |
| PUT | `/stacks/{name}/instances/{id}/env-files` | Override env files |
| PUT | `/stacks/{name}/instances/{id}/networks` | Override networks |
| PUT | `/stacks/{name}/instances/{id}/labels` | Override labels |
| PUT | `/stacks/{name}/instances/{id}/domains` | Override domains |
| PUT | `/stacks/{name}/instances/{id}/healthcheck` | Override healthcheck |
| PUT | `/stacks/{name}/instances/{id}/dependencies` | Override dependencies |
| PUT | `/stacks/{name}/instances/{id}/config-mounts` | Override config mounts |
| GET/PUT | `/stacks/{name}/instances/{id}/resources` | Get / Set CPU/memory limits |
| GET | `/stacks/{name}/instances/{id}/files` | List config files |
| GET/PUT/DELETE | `/stacks/{name}/instances/{id}/files/*` | Manage config files |
| GET | `/stacks/{name}/instances/{id}/effective-config` | Merged template + overrides |

### Wiring
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/stacks/{name}/wires` | List wires |
| POST | `/stacks/{name}/wires` | Create wire |
| POST | `/stacks/{name}/wires/resolve` | Auto-resolve all wires |
| POST | `/stacks/{name}/wires/cleanup` | Remove orphaned wires |
| DELETE | `/stacks/{name}/wires/{id}` | Remove wire |

### Services (Templates)
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/services` | List / Create |
| POST | `/services/bulk` | Bulk operations |
| GET/PUT/DELETE | `/services/{name}` | Get / Update / Delete |
| POST | `/services/{name}/start\|stop\|restart` | Container lifecycle |
| POST | `/services/{name}/rebuild` | Rebuild container |
| GET | `/services/{name}/status` | Container status |
| GET | `/services/{name}/logs` | Container logs |
| GET | `/services/{name}/metrics` | CPU/memory metrics |
| GET | `/services/{name}/compose` | Generated YAML |
| POST | `/services/{name}/validate` | Validate config |
| POST | `/services/{name}/export` | Export service |
| POST | `/services/{name}/materialize` | Materialize config files |
| GET | `/services/{name}/versions` | List config versions |
| GET | `/services/{name}/versions/{v}` | Get specific version |
| PUT | `/services/{name}/ports\|volumes\|env-vars\|env-files` | Update child rows |
| PUT | `/services/{name}/networks\|config-mounts\|dependencies` | Update child rows |
| PUT | `/services/{name}/healthcheck\|labels\|domains` | Update child rows |
| GET | `/services/{name}/files` | List config files |
| GET/PUT/DELETE | `/services/{name}/files/*` | Manage config files |
| GET/PUT | `/services/{name}/exports` | Export contracts |
| GET/PUT | `/services/{name}/imports` | Import contracts |
| GET | `/services/{name}/image` | Image metadata |
| GET | `/services/{name}/tags` | Image tags from registry |
| GET | `/services/{name}/vulnerabilities` | CVE data |
| POST | `/services/{name}/proxy-config` | Generate proxy config |

### Categories
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/categories` | List all |
| GET | `/categories/{name}` | Get one |
| GET | `/categories/{name}/services` | Services in category |
| POST | `/categories/{name}/start\|stop` | Start/stop all services |

### Projects
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/projects` | List all |
| POST | `/projects/scan` | Rescan apps directory |
| GET | `/projects/{name}` | Get project |
| PUT | `/projects/{name}/stack` | Link to stack |
| GET | `/projects/{name}/services` | Project services |
| GET | `/projects/{name}/status` | Project status |
| POST | `/projects/{name}/start\|stop\|restart` | Lifecycle |
| POST | `/projects/{name}/services/{svc}/start\|stop\|restart` | Per-service lifecycle |
| POST | `/projects/{name}/proxy-config` | Generate proxy config |

### Networks
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/networks` | List / Create |
| DELETE | `/networks/{name}` | Remove |
| POST | `/networks/bulk-remove` | Remove orphaned |

### Registries
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/registries` | List registries (DockerHub, GHCR) |
| GET | `/registries/{registry}/search` | Search images |
| GET | `/registries/{registry}/images/*` | Image details |

### Nginx
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/nginx/generate` | Generate all configs |
| POST | `/nginx/generate/{name}` | Generate for one |
| POST | `/nginx/reload` | Reload nginx |

### Proxy
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/proxy/types` | Available proxy types |

---

## 6. URL Parameters & Navigation

The dashboard syncs state to URL for bookmarkability. Defaults are omitted from the URL.

### List Controls
| Param | Purpose | Default | Example |
|-------|---------|---------|---------|
| `q` | Search text | — | `?q=postgres` |
| `sort` | Sort field | varies | `?sort=name` |
| `dir` | Sort direction | `asc` | `?dir=desc` |
| `view` | Display mode | `grid` | `?view=table` |
| `page` | Page number | `1` | `?page=2` |
| `size` | Items per page | `24` | `?size=50` |

Allowed `size` values: 12, 24, 50, 100, 200.

### Filters (per page)
| Page | Params | Example |
|------|--------|---------|
| Overview/Categories | `status` | `?status=running` |
| Services | `status`, `category` | `?status=running&category=database` |
| Projects | `type`, `language` | `?type=wordpress` |

### Tab Keys
| Route | Param | Example |
|-------|-------|---------|
| Stack/Service/Project detail | `tab` | `?tab=compose` |
| Instance detail (nested) | `instanceTab` | `?instanceTab=ports` |

---

## 7. CLI (`scripts/devarch`)

The CLI is a namespace-based dispatcher. All commands follow the pattern `devarch <namespace> <command> [args...]`.

### Namespace Reference

| Namespace | Aliases | Purpose |
|-----------|---------|---------|
| `service` | `svc`, `s` | Service lifecycle (up, down, restart, rebuild, logs, status) |
| `wp` | — | WP-CLI passthrough (runs in php container) |
| `artisan` | `art`, `a` | Laravel Artisan passthrough |
| `wordpress` | `wf` | WordPress workflows (install, backup, restore) |
| `laravel` | `lara`, `l` | Laravel project setup |
| `socket` | `sock` | Podman socket management |
| `runtime` | `rt` | Container runtime switching (Docker/Podman) |
| `db` | `database` | Database initialization |
| `context` | `ctx`, `c` | Generate project context files |
| `run` | `r` | Run command in any container |
| `shell` | `sh` | Open interactive shell in container |
| `init` | — | Bootstrap stack from devarch.yml |
| `doctor` | `doc` | System diagnostics |

### Examples

```bash
# Service management
devarch service up postgres
devarch service down php
devarch service logs nginx -f
devarch service status
devarch service rebuild php --no-cache

# WordPress CLI
devarch wp plugin list
devarch wp theme activate theme-name

# Laravel Artisan
devarch artisan migrate
devarch artisan -p myapp make:model User

# WordPress workflows
devarch wordpress install -n mysite -p bare
devarch wordpress backup -n mysite

# Runtime and socket
devarch runtime status
devarch runtime podman
devarch socket status
devarch socket start-rootless

# Container access
devarch run php bash
devarch run postgres psql -U postgres
devarch shell php

# Bootstrap and diagnostics
devarch init devarch.yml
devarch doctor

# Context generation (hosts file for WSL)
devarch context --apply-hosts
```

### Aliases

After running `scripts/setup-aliases.sh`, these aliases are available: `devarch`, `dvrc`, `dv`, `da`. Supports bash, zsh, and fish shells.

### Scripts Reference

| Script | Purpose |
|--------|---------|
| `devarch` | Main CLI dispatcher (namespace router) |
| `service-manager.sh` | Service lifecycle operations |
| `config.sh` | Central configuration (paths, runtime, network, env vars) |
| `runtime-switcher.sh` | Switch between Docker and Podman |
| `socket-manager.sh` | Manage Podman socket (rootless/rootful) |
| `setup-aliases.sh` | Install shell aliases (bash/zsh/fish) |
| `devarch-init.sh` | Bootstrap stack from devarch.yml (import + pull + apply) |
| `devarch-doctor.sh` | System health checks (runtime, API, disk, tools, ports) |
| `init-databases.sh` | Database initialization scripts |
| `generate-context.sh` | Generate project context files + hosts |
| `wordpress/wp-workflow.sh` | WordPress install, backup, restore |
| `laravel/setup-laravel.sh` | Laravel project creation |

---

## 8. API CLI Commands

The API ships with several CLI tools (available as binaries in the container or runnable via `go run`):

| Command | Usage | Purpose |
|---------|-------|---------|
| `cmd/server` | `go run ./cmd/server` | Start API server |
| `cmd/migrate` | `go run ./cmd/migrate -migrations ./migrations` | Run database migrations |
| `cmd/import` | `go run ./cmd/import -compose-dir <path> -config-dir <path>` | Import service templates from YAML |
| `cmd/export` | `go run ./cmd/export -out seed.sql` | Export DB state as SQL seed file |

Import flags:
- `-compose-dir` — path to service YAML directory (required)
- `-config-dir` — path to config files directory (optional)
- `-project-root` — project root for resolving relative paths
- `-count-only` — print service count and exit
- `-db` — database URL (or set `DATABASE_URL` env)

---

## 9. Environment Variables

### API Server
| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `DEVARCH_API_KEY` | (unset = auth disabled) | API key for authentication |
| `DEVARCH_RUNTIME` | (auto-detect) | Force `docker` or `podman` |
| `DEVARCH_USE_SUDO` | `false` | Run container commands with sudo |
| `CONTAINER_HOST` | (auto-detect) | Podman socket path |
| `PROJECT_ROOT` | `/app` | Project root inside container |
| `HOST_PROJECT_ROOT` | — | Host machine project path |
| `APPS_DIR` | `/workspace/apps` | Directory with project apps |
| `NGINX_GENERATED_DIR` | `/workspace/config/nginx/generated` | Output for nginx configs |
| `STACK_IMPORT_MAX_BYTES` | `268435456` (256MB) | Max stack import file size |

### CLI Scripts (config.sh)
| Variable | Default | Purpose |
|----------|---------|---------|
| `NETWORK_NAME` | `microservices-net` | Shared bridge network name |
| `CONTAINER_RUNTIME` | `podman` | Container runtime |
| `VERBOSE` | `0` | Enable verbose output |
| `QUIET` | `0` | Suppress non-error output |

Default passwords set in `config.sh` (override via `.env` at project root):
- `MARIADB_ROOT_PASSWORD`, `MYSQL_ROOT_PASSWORD`, `POSTGRES_PASSWORD`, `MONGO_ROOT_PASSWORD` — all default to `123456`
- `ADMIN_USER` / `ADMIN_PASSWORD` / `ADMIN_EMAIL` — `admin` / `123456` / `admin@test.local`

### Port Allocation Strategy (from config.sh)

| Language | Port Range | Primary | Vite | Other |
|----------|-----------|---------|------|-------|
| PHP | 8100-8199 | 8100 | 8102 | — |
| Node | 8200-8299 | 8200 | 8202 | graphql:8203, debug:9229 |
| Python | 8300-8399 | 8300 | — | flask:8301, jupyter:8302, flower:8303 |
| Go | 8400-8499 | 8400 | — | metrics:8401, debug:8402, pprof:8403 |
| .NET | 8600-8699 | 8600 | — | debug:8602, hot-reload:8603 |
| Rust | 8700-8799 | 8700 | — | debug:8702, metrics:8703 |

---

## 10. Service Library Structure

Service templates live in `services-library/`, organized by category:

```
services-library/
├── ai/              # ollama, localai, ...
├── analytics/       # matomo, plausible, ...
├── backend/         # php, node, python, go, rust, dotnet, java, ...
├── ci/              # jenkins, drone, ...
├── collaboration/   # nextcloud, ...
├── database/        # postgres, mysql, redis, mongo, neo4j, clickhouse, ...
├── dbms/            # pgadmin, phpmyadmin, adminer, ...
├── docs/            # mkdocs, docusaurus, ...
├── erp/             # odoo, erpnext, ...
├── exporters/       # prometheus exporters
├── gateway/         # api gateway services
├── mail/            # mailpit, mailhog, ...
├── management/      # portainer, ...
├── messaging/       # rabbitmq, kafka, ...
├── project/         # project management tools
├── proxy/           # nginx, traefik, caddy, haproxy, envoy, ...
├── registry/        # docker registry, harbor, ...
├── search/          # elasticsearch, meilisearch, typesense, ...
├── security/        # vault, keycloak, ...
├── storage/         # minio, ...
├── support/         # support tools
├── testing/         # selenium, ...
├── workflow/        # n8n, temporal, ...
└── config/          # config files imported into DB
    ├── php/          # Dockerfile, php.ini
    ├── nginx/        # nginx.conf, default.conf
    ├── prometheus/   # prometheus.yml
    └── ...           # 21 service config directories
```

### Service Template Format

```yaml
services:
  servicename:
    image: imagename:tag
    restart: unless-stopped
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - data_volume:/var/lib/data
    environment:
      KEY: value
    healthcheck:
      test: ["CMD-SHELL", "health-test"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - microservices-net

volumes:
  data_volume:

networks:
  microservices-net:
    external: true
```

### Custom Dockerfiles

Services can use `build:` instead of `image:` for custom Dockerfiles:

```yaml
services:
  php:
    build:
      context: ../config/php
      dockerfile: Dockerfile
    ports:
      - "127.0.0.1:8100:8000"
```

The importer stores the `build` key in `compose_overrides` and resolves relative paths. The generator includes it in the final compose output.

### Config Files

Config files in `services-library/config/{service-name}/` are imported into the `service_config_files` table. These are materialized to disk during stack apply and mounted into containers. Examples: `php.ini`, `nginx.conf`, Dockerfiles.

### Import

The import command is idempotent — safe to re-run at any time:

```bash
go run ./cmd/import -compose-dir ../services-library -config-dir ../services-library/config
```

Services are upserted by name. Child rows (ports, volumes, env vars, etc.) are deleted and re-inserted atomically per service. Config file imports resolve config mount foreign keys automatically.

---

## 11. Deployment System

### Plan

`GET /api/v1/stacks/{name}/plan` computes a diff between desired state (DB) and running state (containers). Returns grouped changes:
1. **Remove** — containers to stop/remove
2. **Modify** — containers to recreate with new config
3. **Add** — new containers to create

The plan differ matches running containers to instances using `devarch.stack_id` and `devarch.instance_id` labels.

### Apply

`POST /api/v1/stacks/{name}/apply` executes the plan:
1. Acquires advisory lock on the stack
2. Checks plan is not stale (stack unchanged since plan was generated)
3. Creates the stack network if it doesn't exist
4. Materializes config files to disk
5. Generates compose YAML from DB state
6. Runs `podman compose -f generated.yml up -d`
7. Updates container state in DB

### Lock Files

Lock files capture a snapshot of running state:
- Config hash of all instances
- Container IDs
- Network ID

Used to validate that apply is safe (no concurrent modifications). Locks can be validated and refreshed via API.

---

## 12. Secrets & Encryption

- Env vars can be marked `is_secret` when created
- Secret values are encrypted with **AES-256-GCM** before storing in PostgreSQL
- Encryption key: 32 random bytes at `~/.devarch/secret.key` (auto-generated on first run)
- In API responses: secrets are redacted (not returned in plaintext)
- In stack exports: secrets are replaced with placeholders
- Secrets are decrypted only for compose generation and container launch

---

## 13. Real-Time Status

A background **sync manager** runs continuously:
- Polls container state and metrics from the container runtime
- Updates `container_states` and `container_metrics` tables in DB
- Syncs image tags from registries (DockerHub, GHCR)
- Cleans up old metrics and config versions
- Broadcasts status updates via **WebSocket** at `/api/v1/ws/status`

The dashboard subscribes to the WebSocket for live container status updates without polling.

Manual sync can be triggered via `POST /api/v1/sync`. Job status visible at `GET /api/v1/sync/jobs`.

---

## 14. Database Schema

10 migration files (001-010) creating ~40 tables:

| Migration | Tables Created |
|-----------|---------------|
| 001 | `categories`, `services` |
| 002 | `service_ports`, `service_volumes`, `service_env_vars`, `service_env_files`, `service_dependencies`, `service_healthchecks`, `service_labels`, `service_domains`, `service_networks` |
| 003 | `service_config_files`, `service_config_mounts`, `service_config_versions` |
| 004 | `registries`, `images`, `image_tags`, `image_architectures`, `vulnerabilities`, `image_tag_vulnerabilities` |
| 005 | `projects`, `project_services` |
| 006 | `stacks`, `service_instances` |
| 007 | `instance_ports`, `instance_volumes`, `instance_env_vars`, `instance_env_files`, `instance_labels`, `instance_domains`, `instance_healthchecks`, `instance_config_files`, `instance_dependencies`, `instance_resource_limits` |
| 008 | `service_exports`, `service_import_contracts`, `service_instance_wires`, `sync_state`, `container_states`, `container_metrics` |
| 009 | Performance indexes |
| 010 | Additional instance override tables |

---

## 15. Key Things to Know

1. **Compose YAML is never stored** — always generated on-the-fly from DB. If something looks wrong in the compose, the issue is in the DB state or generator.

2. **Instance overrides always win** over template values. Check the "Effective Config" tab to see the merged result.

3. **Soft-delete** — deleted stacks go to trash (`/stacks/trash`), can be restored. Permanent delete available.

4. **Advisory locking** — only one apply per stack at a time. If plan is stale (stack changed since plan was generated), apply is rejected.

5. **Secrets** — env vars marked `is_secret` are encrypted (AES-256-GCM) in DB, redacted in API responses and exports. Key stored at `~/.devarch/secret.key`.

6. **Networking** — each stack gets `devarch-{stack}-net`. Containers reference each other by instance name within the stack network.

7. **Auto-wiring** — if one instance exports a contract type (e.g., "postgres") and another imports it, wiring can be auto-resolved. Ambiguous cases (two postgres instances) require explicit wiring.

8. **Import size** — stack imports support up to 256MB (configurable); all other endpoints are capped at 10MB.

9. **Runtime switching** — Docker and Podman are interchangeable via Settings or CLI. The API uses a container.Client abstraction layer that auto-detects the available runtime.

10. **Config versioning** — service config changes are versioned in `service_config_versions` with JSONB snapshots (default 10 versions retained).

11. **Vulnerability scanning** — optional CVE scanning via Trivy integration. Results stored per image tag.

12. **External network** — the `microservices-net` bridge network must exist before starting compose. Create it with `podman network create microservices-net`.

---

## 16. Technology Stack

| Component | Technology |
|-----------|------------|
| **API** | Go 1.22, chi router, lib/pq, gorilla/websocket, yaml.v3 |
| **Dashboard** | React 19, Vite 7, TanStack Router + Query, Tailwind CSS 4, Radix UI, CodeMirror 6, Zod, Axios, Sonner |
| **Database** | PostgreSQL 16 |
| **Container Runtime** | Docker or Podman (auto-detect, switchable) |
| **CLI** | Bash/Zsh scripts |
| **API Container** | Go 1.23 alpine + air hot-reload (dev), multi-stage build (production) |
| **Dashboard Container** | Node 22 build → nginx alpine |
| **Registries** | DockerHub, GitHub Container Registry (GHCR) |
