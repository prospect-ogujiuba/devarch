# External Integrations

**Analysis Date:** 2026-01-07

## APIs & External Services

**DevArch Management API:**
- Purpose: Container and service management for dashboard
- Base URL: Configurable via `VITE_API_BASE_URL` (defaults to `/api`)
- Backend: PHP router at `config/devarch/api/public/index.php`
- Endpoints:
  - `/api/containers` - Container listing, stats, filtering
  - `/api/services/start`, `/api/services/stop` - Service control
  - `/api/categories`, `/api/category-containers` - Service grouping
  - `/api/apps`, `/api/domains` - Application registry
  - `/api/logs` - Container log streaming
  - `/api/control` - Container start/stop/restart
  - `/api/bulk/*` - Bulk operations

**TypeRocket REST API:**
- Purpose: WordPress data access for makerblocks plugin
- Base URL: `/tr-api/rest/{resource}/{id?}/actions/{action?}`
- Auth: Nonce-based (`X-TypeRocket-Nonce` header)
- Resources: teams, services, service-categories, service-bundles, pricing-models, contact-submissions

## Data Storage

**Databases:**
- MariaDB - Primary WordPress database (`compose/database/mariadb.yml`)
  - Host: `mariadb` (Docker network)
  - Port: 3306
  - Credentials: `MYSQL_USER`, `MYSQL_PASSWORD` env vars

- PostgreSQL - Analytics and NoCode platforms (`compose/database/postgres.yml`)
  - Host: `postgres`
  - Port: 5432
  - Databases: postgres, metabase, nocodb, openproject

- Redis - Caching, sessions, task queues (`compose/database/redis.yml`)
  - Host: `redis`
  - Port: 6379
  - Auth: `SHARED_DB_PASSWORD` env var

**File Storage:**
- Local filesystem via Docker volumes
- WordPress uploads in `apps/b2bcnc/wp-content/uploads/`

**Caching:**
- PHP API: Static cache in `lib/containers.php` (30-60s TTL)
- Redis available for application-level caching
- React: localStorage for UI preferences (`apps/dashboard/src/hooks/useLocalStorage.js`)

## Authentication & Identity

**Auth Provider:**
- WordPress nonces for CSRF protection
- TypeRocket Auth for WordPress admin policies
- No external OAuth configured at platform level

**API Authentication:**
- Dashboard API: No authentication (local development tool)
- TypeRocket REST: Nonce validation via `window.siteData.nonce`

## Monitoring & Observability

**Error Tracking:**
- Not configured - relies on container logs

**Analytics:**
- Metabase available (`compose/analytics/metabase.yml`)
- Uses PostgreSQL for data storage

**Logs:**
- Container logs via `/api/logs` endpoint
- Podman/Docker native logging
- Laravel Pail for log viewing (`apps/mxcros/composer.json`)
- PHP xdebug for debugging (`config/php/Dockerfile`)

## CI/CD & Deployment

**Hosting:**
- Self-hosted via Docker/Podman Compose
- Management container: `config/devarch/Dockerfile`
- All services on `microservices-net` bridge network

**CI Pipeline:**
- Drone CI available (`compose/ci/drone-runner.yml`)
- Woodpecker CI available (`compose/ci/woodpecker-server.yml`)
- GitLab Runner available (`compose/ci/gitlab-runner.yml`)
- Concourse CI available (`compose/ci/concourse-web.yml`)

**Available CI Services:**
- 8 CI/CD platforms in `compose/ci/`

## Environment Configuration

**Development:**
- Required env vars in `.env.example`:
  - `ADMIN_EMAIL`, `ADMIN_PASSWORD`
  - `MYSQL_*` credentials
  - `GITHUB_USER`, `GITHUB_TOKEN`
  - `SHARED_DB_PASSWORD`
- Secrets location: `.env` (gitignored), `.env.example` committed
- Dashboard proxy: `/api` → `http://localhost:8500`

**Production:**
- Same compose-based deployment
- Environment-specific `.env` files
- No secrets management platform configured

## Webhooks & Callbacks

**Incoming:**
- Not detected at platform level

**Outgoing:**
- Contact submission workflow in TypeRocket
- Admin actions: `admin_post_update_contact_submission_status`

## Message Queues

**Available (not all active):**
- Celery + Redis - Python async tasks (`config/python/requirements.txt`)
- RabbitMQ (`compose/messaging/rabbitmq.yml`)
- Apache Kafka + Zookeeper (`compose/messaging/kafka.yml`)
- NATS (`compose/messaging/nats.yml`)
- Apache Pulsar (`compose/messaging/pulsar.yml`)
- Apache ActiveMQ (`compose/messaging/activemq.yml`)

## Mail Services

**Development:**
- Mailpit - Local mail relay (`config/php/Dockerfile`, `config/node/Dockerfile`)
- msmtp configured in containers to forward to `mailpit` service

## Container Management

**Runtime:**
- Podman/Docker with automatic detection (`scripts/config.sh`)
- Runtime switching via `scripts/runtime-switcher.sh`
- Socket management via `scripts/socket-manager.sh`

**Management UIs (available):**
- Portainer (`compose/management/portainer.yml`)
- Yacht (`compose/management/yacht.yml`)
- Dockge (`compose/management/dockge.yml`)
- Rancher (`compose/management/rancher.yml`)

## Proxy & Gateway

**Reverse Proxy:**
- NGINX Proxy Manager available (`compose/management/`)
- Credentials from env: `INITIAL_ADMIN_EMAIL`, `INITIAL_ADMIN_PASSWORD`

**API Gateways (available):**
- KrakenD, Kong, Traefik, Envoy, Tyk, Gravitee, APISIX
- Located in `compose/gateway/`

---

*Integration audit: 2026-01-07*
*Update when adding/removing external services*
