# practical-use-case-and-library-design

Created: 2026-05-17T02:10:41.560Z
Purpose: Practical DevArch story, command walkthrough, expected outputs, and unified catalog/library design.

# DevArch practical use case: onboarding a shop app locally

## Story

Maya joins a small commerce team. The app has a Vite storefront, a Node API, Postgres, and Redis. Before DevArch, onboarding meant reading a README, installing databases, copying `.env` files, creating Docker networks, and hoping ports did not collide.

With DevArch, the team commits one workspace manifest. The manifest says what the local system needs, not how each developer should hand-wire it. Reusable catalog templates provide the standard service shapes. DevArch resolves the manifest, plans what must change, applies the runtime resources, then gives status/log/exec commands for daily work.

## Example workspace

Create `workspaces/shop-local/devarch.workspace.yaml`:

```yaml
apiVersion: devarch.io/alpha1
kind: Workspace
metadata:
  name: shop-local
runtime:
  provider: docker
  namespace: shop
resources:
  postgres:
    template:
      name: postgres
    env:
      POSTGRES_DB: shop
      POSTGRES_USER: shop
      POSTGRES_PASSWORD: devpass
  redis:
    template:
      name: redis
  api:
    template:
      name: node-api
    source:
      path: ../shop-api
    env:
      PORT: "3000"
    dependsOn:
      - postgres
      - redis
  web:
    template:
      name: vite-web
    source:
      path: ../shop-web
    env:
      PORT: "5173"
    dependsOn:
      - api
```

The important part: `api` imports `postgres` and `redis` contracts from the templates. DevArch can inject `DATABASE_URL`, `REDIS_URL`, host, port, and similar environment values without every project inventing its own wiring.

## Commands and expected outputs

Use the built-in catalog plus the folder containing workspace manifests:

```bash
export DEVARCH_ROOTS="--workspace-root ./workspaces --catalog-root ./catalog/builtin"
```

### 1. Check host readiness

```bash
devarch $DEVARCH_ROOTS doctor
```

Expected shape:

```txt
Doctor: pass
CHECK                 STATUS  SUMMARY
podman/docker binary  pass    podman version 5.0
socket                pass    socket ready
package               pass    package ok
```

If Docker/Podman is missing, expect a failed check with the failing command summary.

### 2. Inspect reusable templates

```bash
devarch $DEVARCH_ROOTS catalog list
```

Expected shape:

```txt
NAME         TAGS
node-api     backend,node,http
postgres     database,sql,postgres
redis        cache,queue,redis
vite-web     frontend,vite,http
nginx        proxy,http,nginx
```

```bash
devarch $DEVARCH_ROOTS catalog show postgres
```

Expected shape:

```txt
Template: postgres
Description: PostgreSQL database with persisted application data.
Ports: 5432
Exports: postgres
Environment: POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD
Volumes: /var/lib/postgresql/data (data)
```

### 3. Discover workspaces

```bash
devarch $DEVARCH_ROOTS workspace list
```

Expected shape:

```txt
NAME        PROVIDER  RESOURCES
shop-local  docker    api, postgres, redis, web
```

```bash
devarch $DEVARCH_ROOTS workspace open shop-local
```

Expected shape:

```txt
Workspace: shop-local
Manifest: workspaces/shop-local/devarch.workspace.yaml
Provider: docker
Resources: api, postgres, redis, web
```

### 4. Plan before changing the machine

```bash
devarch $DEVARCH_ROOTS workspace plan shop-local
```

Expected first run:

```txt
Workspace: shop-local
ACTION   TARGET          RUNTIME NAME           REASONS
create   network/shop    devarch_shop-local     missing network
create   volume/postgres devarch_shop_postgres  data volume missing
create   volume/redis    devarch_shop_redis     data volume missing
create   service/postgres devarch_shop_postgres missing container
create   service/redis    devarch_shop_redis    missing container
create   service/api      devarch_shop_api      missing container
create   service/web      devarch_shop_web      missing container
```

Expected later run after no changes:

```txt
Workspace: shop-local
ACTION  TARGET  REASONS
noop    all     runtime already matches desired graph
```

For automation, use JSON:

```bash
devarch $DEVARCH_ROOTS --json workspace plan shop-local
```

Expected JSON shape:

```json
{
  "workspace": "shop-local",
  "actions": [
    {"scope":"service","target":"postgres","kind":"create","reasons":["missing container"]}
  ]
}
```

### 5. Apply the plan

```bash
devarch $DEVARCH_ROOTS workspace apply shop-local
```

Expected shape:

```txt
Workspace: shop-local
OPERATION             STATUS  SUMMARY
network/create        pass    devarch_shop-local
volume/create         pass    devarch_shop_postgres
service/create        pass    postgres started
service/create        pass    redis started
service/create        pass    api started
service/create        pass    web started
```

### 6. Observe and work inside the environment

```bash
devarch $DEVARCH_ROOTS workspace status shop-local
```

Expected shape:

```txt
Workspace: shop-local
RESOURCE   STATUS   URL/PORT
postgres   running  localhost:5432
redis      running  localhost:6379
api        running  http://localhost:3000
web        running  http://localhost:5173
```

```bash
devarch $DEVARCH_ROOTS workspace logs --tail 20 shop-local api
```

Expected shape:

```txt
api | npm run dev
api | listening on 0.0.0.0:3000
api | connected to postgres
api | connected to redis
```

```bash
devarch $DEVARCH_ROOTS workspace exec shop-local api -- npm test
```

Expected shape:

```txt
> shop-api test
PASS src/health.test.ts
PASS src/orders.test.ts
```

### 7. Use scan to bootstrap a manifest

```bash
devarch $DEVARCH_ROOTS scan project ./shop-api
```

Expected shape for an Express app with Compose Postgres:

```txt
Project: ./shop-api
Type: node
Detected: package.json, compose.yml
Suggested templates: node-api, postgres
```

## What the practical value is

- New developers run one plan/apply workflow instead of reconstructing local infra.
- The team can review environment changes before they touch the runtime.
- Common services live once in the catalog, not copied across projects.
- Contracts make service wiring deterministic: `postgres` exports database env, `node-api` imports it.
- Status, logs, exec, and restart are standard across projects.

# Unifying catalog and library

## Current split

DevArch currently has the right primitives but the names imply two systems:

- **Catalog**: discovered by `--catalog-root`, stores v2 reusable `Template` documents like `postgres`, `redis`, `node-api`.
- **Library/import path**: legacy-facing commands such as `import v1-library <path>` suggest a separate source of reusable architecture definitions.

That split will confuse users. In practice, both are a library of reusable building blocks.

## Proposed model: one DevArch Library

Make **Library** the product concept and keep **catalog** as the internal/indexing term if useful.

A DevArch Library contains versioned, discoverable packages:

```txt
library/
  devarch.library.yaml
  templates/
    database/postgres/template.yaml
    cache/redis/template.yaml
    backend/node-api/template.yaml
  bundles/
    shop-stack/bundle.yaml
  contracts/
    postgres/contract.yaml
    redis/contract.yaml
  policies/
    local-dev/policy.yaml
```

Suggested package types:

1. **Template**: one resource shape; current catalog template.
2. **Bundle**: a reusable group of resources, e.g. `node-api + postgres + redis`.
3. **Contract**: named wiring surface, e.g. `postgres`, `redis`, `http`.
4. **Policy**: defaults and constraints, e.g. port ranges, allowed images, required health checks.
5. **Workspace preset**: partial workspace manifest for teams.

## CLI shape

Keep old catalog commands as aliases during transition, but teach users `library`:

```txt
devarch library list
devarch library search postgres
devarch library show template/postgres
devarch library show bundle/shop-stack
devarch library add ./company-devarch-library
devarch library update
devarch library validate
devarch library explain shop-local
```

Backward compatible aliases:

```txt
devarch catalog list        -> devarch library list --type template
devarch catalog show NAME   -> devarch library show template/NAME
--catalog-root PATH         -> --library-root PATH
```

## Resolver changes

Resolution should become:

1. Load workspace roots.
2. Load library roots.
3. Index all packages by canonical ID: `template/postgres`, `bundle/shop-stack`, `contract/postgres`.
4. Expand bundles into resources.
5. Merge workspace overrides.
6. Resolve imports/exports through contract package definitions.
7. Emit one effective graph for plan/apply/status.

Workspace references become explicit but can keep shorthand:

```yaml
resources:
  db:
    use: template/postgres
  cache:
    use: template/redis
  api:
    use: template/node-api
    dependsOn: [db, cache]
```

Bundle example:

```yaml
resources:
  app:
    use: bundle/node-api-with-postgres-redis
    with:
      api.source.path: ../shop-api
      postgres.env.POSTGRES_DB: shop
```

## Why this fits the rest of DevArch

- The planner still consumes a deterministic resolved graph.
- Apply/status/logs/exec do not need to know whether a resource came from a template or bundle.
- Contracts become first-class library documents instead of implicit strings inside templates.
- Project scan can suggest library packages, not just templates.
- Teams can publish one internal DevArch Library instead of copying catalog folders.

## Migration path

1. Rename user docs from catalog to library.
2. Add `--library-root` while preserving `--catalog-root`.
3. Add `devarch library ...` commands backed by the existing catalog index.
4. Introduce package IDs and package type metadata.
5. Add bundles after templates are stable.
6. Deprecate `import v1-library` into `library import v1 <path>`.

## Design principle

Users should think: **"My workspace uses packages from a DevArch Library."**

Implementers can still think: **"The resolver indexes templates/contracts/bundles into a catalog index and emits a graph."**
