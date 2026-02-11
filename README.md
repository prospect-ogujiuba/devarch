# DevArch User Guide

## What is DevArch?

DevArch is a local microservices development environment. You define containerized services (postgres, redis, nginx, etc.), group them into **stacks**, and DevArch handles isolation, networking, compose generation, and deployment. The only prerequisite is Docker or Podman.

**Core invariant:** Two stacks using the same service template never collide — each gets its own network, container names, and config.

---

## 1. Setup & First Launch

### Prerequisites
- Docker or Podman installed
- PostgreSQL (runs via compose)

### Start the infrastructure

```bash
# From project root — starts Postgres on :5433 and API on :8550
docker compose up

# Run database migrations (from api/)
cd api && go run ./cmd/migrate -migrations ./migrations

# Import the service catalog and config files (from api/)
go run ./cmd/import -compose-dir ../services-library -config-dir ../services-library/config

# Start the dashboard (from dashboard/)
cd dashboard && npm run dev
# Dashboard available at http://localhost:5174
```

### Authentication
- Set `DEVARCH_API_KEY` env var on the API to enable auth (if unset, auth is disabled)
- The dashboard login page asks for this key
- All API calls use `X-API-Key` header

---

## 2. Core Concepts

### Services (Templates)
173+ pre-built container definitions across 24 categories (databases, caching, proxy, monitoring, etc.). Each has: image (or custom Dockerfile via build context), ports, volumes, env vars, healthchecks, labels, domains, dependencies, config files. You **don't run services directly** — you instantiate them inside stacks.

Services can use pre-built images or custom Dockerfiles. Build contexts and non-standard compose keys are stored in a `compose_overrides` JSONB field and passed through to the generated compose YAML.

### Stacks
Isolated environment namespaces. Creating a stack gives you:
- A dedicated bridge network: `devarch-{stack}-net`
- Deterministic container naming: `devarch-{stack}-{instance}`
- Encrypted secrets, auto-wiring, and isolated compose YAML

### Instances
When you add a service to a stack, you create an **instance** — a copy-on-write deployment of that service template. Override anything per-instance: ports, volumes, env vars, networks, labels, domains, healthchecks, dependencies, config files, resource limits. The template stays pristine.

### Projects
Scanned from the `apps/` directory. WordPress, Laravel, and other project types are auto-detected with their dependencies, plugins, themes, and scripts. Projects can be linked to stacks.

### Categories
Organizational groups for services (DATABASES, CACHING, PROXY, etc.). Used for filtering and browsing.

---

## 3. Dashboard Pages

Navigate via the left sidebar:

### Overview (`/`)
Landing page with stat cards: total services, running/stopped counts, categories, avg CPU, total memory. Browse categories with search, sort, and status filters. Grid or table view.

### Stacks (`/stacks`)
**The primary workspace.** List all stacks with search, sort, status filters.

**Stack actions:** Create, Clone, Rename, Edit description, Enable/Disable, Delete, Create/Remove network, Start/Stop/Restart all instances, Export YAML, Import from file.

**Stack detail** (`/stacks/{name}`) has tabs:
- **Instances** — Grid of all service instances. Per-instance: start/stop/restart, enable/disable, duplicate, delete. Click through to instance detail.
- **Compose** — View/download the generated Docker Compose YAML (never hand-written, always generated from DB state).
- **Wiring** — Shows active wires between instances, injected env vars, warnings.
- **Deploy** — Generate a terraform-style plan (add/modify/remove preview), then Apply to execute.
- **Proxy** — Generate nginx reverse proxy config for the stack's domains.

### Instance Detail (`/stacks/{name}/instances/{instance}`)
Full override editor with tabs for every configurable aspect:
- **Info**: basics, container name, description, enabled toggle
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
- **Config Files**: manage config file content with syntax highlighting
- **Resources**: CPU/memory limits and reservations
- **Effective Config**: read-only merged view (template + your overrides)

Actions: Edit description, Duplicate, Rename, Delete.

### Services (`/services`)
Browse and manage the service template catalog. Search by name/image, filter by status/category, sort by name/status/CPU/memory. Grid or table view.

**Service detail** (`/services/{name}`) tabs:
- **Info**: status, image, restart policy, metrics (CPU/memory/network/uptime), inline editors for ports/volumes/env/deps/labels/domains/healthcheck/config mounts
- **Environment**: env var editor with secret flag
- **Logs**: live streaming container logs with search
- **Compose**: generated YAML preview
- **Files**: config file editor with syntax highlighting
- **Proxy**: reverse proxy config generation

**Create service** (`/services/new`): form with all fields — name, category, image, tag (with tag picker from registry), restart policy, command, ports, volumes, env vars, dependencies, healthcheck, labels, domains.

### Projects (`/projects`)
Browse auto-detected projects from `apps/` directory. Search by name/framework/language. Filter by type (wordpress/laravel) or language.

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
Search Docker images across registries (defaults to DockerHub). Browse results with star/pull counts. Used by the tag picker when creating/editing services.

### Settings (`/settings`)
- **Container Runtime**: see Docker/Podman status, versions, container counts. Switch between runtimes.
- **Podman Socket**: manage rootless/rootful socket status.
- **API Key**: view (obscured), copy, change.

---

## 4. Typical Workflow

### Creating a development environment from scratch

1. **Create a stack**: Stacks > "New Stack" > name it (e.g., `my-project`)
2. **Add instances**: In stack detail > Instances tab > "Add Instance" > pick a template (e.g., postgres, redis, nginx) > give it an instance ID
3. **Configure overrides**: Click into each instance > override ports (avoid collisions with other stacks), set env vars (database passwords, etc.), mount config files
4. **Review wiring**: Wiring tab shows auto-detected connections. If service A needs postgres and there's exactly one postgres instance, DevArch auto-wires it (injects `DB_HOST`, `DB_PORT` env vars)
5. **Plan**: Deploy tab > "Generate Plan" — see what containers will be created/modified/removed
6. **Apply**: "Apply Plan" — DevArch locks the stack, ensures the network exists, materializes config files to disk, and runs compose up
7. **Monitor**: View logs in Services > {name} > Logs tab. Check container status on Overview.

### Sharing an environment

```
# Export
GET /api/v1/stacks/{name}/export  ->  devarch.yml
POST /api/v1/stacks/{name}/lock   ->  devarch.lock

# Share devarch.yml + devarch.lock

# On another machine
POST /api/v1/stacks/import  (with devarch.yml)
# Then plan + apply
```

Secrets are redacted in exports (placeholders, not plaintext). The recipient provides their own secrets.

---

## 5. API Reference (Key Endpoints)

All endpoints are under `/api/v1/`. Auth via `X-API-Key` header. Rate limited: 10 req/sec, burst 50.

### Health & Status
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Liveness check |
| POST | `/auth/validate` | Validate API key |
| GET | `/status` | System overview (service counts, runtime info) |
| GET | `/runtime/status` | Docker/Podman availability |
| WS | `/ws/status` | Real-time container status (WebSocket) |

### Stacks
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/stacks` | List / Create |
| GET/PUT/DELETE | `/stacks/{name}` | Get / Update / Soft-delete |
| POST | `/stacks/{name}/clone` | Clone with new name |
| POST | `/stacks/{name}/enable` | Enable stack |
| POST | `/stacks/{name}/disable` | Disable (stops containers) |
| GET | `/stacks/{name}/plan` | Preview deployment changes |
| POST | `/stacks/{name}/apply` | Execute deployment |
| GET | `/stacks/{name}/compose` | Get generated compose YAML |
| GET | `/stacks/{name}/export` | Export as devarch.yml |
| POST | `/stacks/import` | Import devarch.yml (256MB limit) |
| POST | `/stacks/{name}/start\|stop\|restart` | Lifecycle operations |

### Instances (nested under stacks)
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/stacks/{name}/instances` | List / Create |
| GET/PUT/DELETE | `/stacks/{name}/instances/{id}` | Get / Update / Delete |
| PUT | `/stacks/{name}/instances/{id}/ports` | Override ports |
| PUT | `/stacks/{name}/instances/{id}/volumes` | Override volumes |
| PUT | `/stacks/{name}/instances/{id}/env-vars` | Override env vars |
| PUT | `/stacks/{name}/instances/{id}/networks` | Override networks |
| PUT | `/stacks/{name}/instances/{id}/resources` | Set CPU/memory limits |
| GET | `/stacks/{name}/instances/{id}/effective-config` | Merged template + overrides |

### Services (Templates)
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/services` | List / Create |
| GET/PUT/DELETE | `/services/{name}` | Get / Update / Delete |
| POST | `/services/{name}/start\|stop\|restart` | Container lifecycle |
| GET | `/services/{name}/logs` | Container logs |
| GET | `/services/{name}/metrics` | CPU/memory metrics |
| GET | `/services/{name}/compose` | Generated YAML |

### Wiring
| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/stacks/{name}/wires` | List / Create wire |
| POST | `/stacks/{name}/wires/resolve` | Auto-resolve all wires |
| DELETE | `/stacks/{name}/wires/{id}` | Remove wire |

### Projects, Categories, Networks, Registries
Similar CRUD patterns. Projects support `/scan` to detect new ones, `/projects/{name}/stack` to link to a stack.

---

## 6. URL Parameters & Navigation

The dashboard syncs state to URL for bookmarkability:

| Param | Purpose | Example |
|-------|---------|---------|
| `q` | Search text | `?q=postgres` |
| `sort` | Sort field | `?sort=name` |
| `dir` | Sort direction | `?dir=desc` |
| `view` | Display mode | `?view=table` (or `grid`) |
| `page` | Page number | `?page=2` |
| `size` | Items per page | `?size=50` |
| `status` | Status filter | `?status=running` |
| `tab` | Active tab | `?tab=compose` |

---

## 7. CLI Scripts

From the `scripts/` directory:

| Script | Purpose |
|--------|---------|
| `devarch` | Main CLI wrapper — thin API client |
| `service-manager.sh` | Start/stop/restart services, manage lifecycle |
| `config.sh` | Central configuration (paths, runtime, network name) |
| `runtime-switcher.sh` | Switch between Docker and Podman |
| `socket-manager.sh` | Manage Podman socket (rootless/rootful) |
| `wordpress/setup-wordpress.sh` | WordPress project setup (create, activate, deactivate, backup, restore) |
| `laravel/setup-laravel.sh` | Laravel project setup (new or clone from git) |

### Common CLI Commands

```bash
devarch doctor                          # System health check
devarch runtime status                  # Show current runtime
devarch service status                  # Show all service statuses
devarch service up <service>            # Start a service
devarch service down <service>          # Stop a service
devarch service logs <service> -f       # Follow logs
devarch init [devarch.yml]              # Bootstrap from export file
```

---

## 8. Key Things to Know for Testing

1. **Compose YAML is never stored** — always generated on-the-fly from DB. If something looks wrong in the compose, the issue is in the DB state or generator.

2. **Instance overrides always win** over template values. Check the "Effective Config" tab to see the merged result.

3. **Stack names are immutable** — renaming creates a new stack via clone + delete. Container names change accordingly.

4. **Soft-delete** — deleted stacks go to trash (`/stacks/trash`), can be restored.

5. **Advisory locking** — only one apply per stack at a time. If plan is stale (stack changed since plan was generated), apply is rejected.

6. **Secrets** — env vars marked `is_secret` are encrypted (AES-256-GCM) in DB, redacted in API responses and exports. Key stored at `~/.devarch/secret.key`.

7. **Networking** — each stack gets `devarch-{stack}-net`. Containers reference each other by instance name within the stack network.

8. **Auto-wiring** — if one instance exports "postgres" and another imports it, wiring is automatic. Ambiguous cases (two postgres instances) require explicit wiring.

9. **Import size** — stack imports support up to 256MB; all other endpoints are capped at 10MB.

10. **Runtime switching** — Docker and Podman are interchangeable via Settings. The API uses a container.Client abstraction layer.

---

## 9. Service Library Structure

Service templates live in `services-library/`, organized by category:

```
services-library/
├── database/          # postgres.yml, mysql.yml, redis.yml, ...
├── backend/           # php.yml, node.yml, python.yml, go.yml, ...
├── proxy/             # nginx.yml, traefik.yml, ...
├── ...                # 15+ category directories
└── config/            # config files imported into DB
    ├── php/           # Dockerfile, php.ini
    ├── nginx/         # nginx.conf, default.conf
    └── ...
```

### Custom Dockerfiles

Services can use `build:` instead of `image:` to build from a custom Dockerfile:

```yaml
services:
  php:
    build:
      context: ../config/php
      dockerfile: Dockerfile
    ports:
      - "127.0.0.1:8100:8000"
    volumes:
      - php_config_php_ini:/usr/local/etc/php/php.ini:ro
    depends_on:
      - mailpit
```

The importer stores the `build` key in `compose_overrides` and resolves relative paths. The generator includes it in the final compose output.

### Config Files

Config files in `services-library/config/{service-name}/` are imported into the `service_config_files` table. These are materialized to disk during stack apply and mounted into containers. Examples: `php.ini`, `nginx.conf`, Dockerfiles.

### Import

The import command is idempotent — safe to re-run at any time:

```bash
go run ./cmd/import -compose-dir ../services-library -config-dir ../services-library/config
```

Services are upserted by name. Child rows (ports, volumes, env vars, etc.) are deleted and re-inserted atomically per service.
