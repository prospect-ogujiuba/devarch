# Multi-app JavaScript runtime

`scripts/node/bootstrap.sh` runs an existing `apps/<app-name>` JavaScript application in its own Node 22 container. A shared `node` router receives wildcard Nginx traffic and forwards `<app-name>.test` to the isolated `node-<app-name>` container over `microservices-net`.

This design can run multiple Next.js, Nuxt, Vite, Remix/React Router, Astro SSR, or generic Node HTTP applications concurrently. It does not combine Node with PHP-FPM and does not publish per-app host ports, so it works alongside WordPress and Laravel applications.

## Application contract

Each application must:

1. Live at `apps/<app-name>` and contain `package.json`.
2. Define a package script named `devarch` (or select another with `--script`).
3. Listen on `0.0.0.0:3000`.
4. Use a lowercase DNS-safe directory name of at most 58 characters, such as `store-front`.

The runtime sets `HOST`, `HOSTNAME`, `PORT`, `NUXT_HOST`, and `NUXT_PORT`. Explicit framework CLI flags are still recommended.

Examples:

```json
{
  "scripts": {
    "devarch": "next dev --hostname 0.0.0.0 --port 3000"
  }
}
```

```json
{
  "scripts": {
    "devarch": "vite --host 0.0.0.0 --port 3000"
  }
}
```

```json
{
  "scripts": {
    "devarch": "nuxt dev --host 0.0.0.0 --port 3000"
  }
}
```

A generic Node server should read `HOST` and `PORT` or otherwise bind the same address and port.

## Usage

```bash
scripts/node/bootstrap.sh my-next-app --dry-run
scripts/node/bootstrap.sh my-next-app
```

Package manager selection is inferred from `pnpm-lock.yaml`, `yarn.lock`, or npm by default:

```bash
scripts/node/bootstrap.sh my-app --package-manager pnpm
scripts/node/bootstrap.sh my-app --script dev
scripts/node/bootstrap.sh my-app --no-hosts
```

The bootstrap:

1. Validates the app name, `package.json`, and selected script.
2. Creates `microservices-net` when absent.
3. Starts/rebuilds the shared `node` router.
4. Starts Nginx Proxy Manager and reloads its validated configuration.
5. Recreates `node-<app-name>` with an isolated `node_modules` volume so runtime environment changes are applied.
6. Registers `<app-name>.test` unless `--no-hosts` is used.

Use the selected runtime's Compose project to inspect or stop one app:

```bash
DEVARCH_NODE_APP_NAME=my-app \
  podman compose -p devarch-node-my-app \
  -f services-library/backend/node/app.compose.yml ps

DEVARCH_NODE_APP_NAME=my-app \
  podman compose -p devarch-node-my-app \
  -f services-library/backend/node/app.compose.yml down
```

Pass the same `DEVARCH_NODE_PACKAGE_MANAGER`, `DEVARCH_NODE_SCRIPT`, and `DEVARCH_NODE_CONTAINER_USER` values when recreating directly through Compose. The bootstrap is preferred because it supplies these consistently.

## Routing behavior

Nginx still serves detected `out/`, `dist/`, `build/`, and `public/` files directly. A project containing `composer.json` or a PHP entry point retains PHP precedence and routes to `php:9000`. Other projects with `package.json` route through `node:3000`; the router validates the `.test` host and forwards to the matching isolated container. JavaScript `/api/*` and WebSocket/HMR requests follow the same app route.

The router never mounts the container socket and cannot create containers from HTTP requests. An app must be started explicitly before its hostname returns a successful response.

## Static Next.js exports

A Next.js project using `output: "export"` generates `out/` and can be served directly by Nginx without a running app container. Only use the runtime container when the application needs SSR, Server Actions, route handlers, WebSockets, or another long-running Node server.

## Notes

- These definitions target local development, not production deployment.
- Dependency installation runs when the app container starts and is cached in its named volume.
- Rootless Podman uses `0:0` inside the user namespace. Docker maps the process to the invoking host UID/GID to avoid root-owned build output.
- File watching defaults to polling for bind-mount compatibility; set `WATCHPACK_POLLING=false` when native filesystem events work reliably.
- Vite-based servers receive `__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS=<app-name>.test`, allowing only their matching wildcard-proxy hostname.

## Tests

```bash
node --test services-library/backend/node/config/router.test.js
bash scripts/node/bootstrap.test.sh
bash scripts/node/routing.test.sh
```

The routing integration test uses the active Podman/Docker development stack and skips when Nginx Proxy Manager is not running. It verifies two concurrent runtimes, API and metadata routes, accepted/rejected WebSocket upgrades, static clean URLs/assets/404s, and PHP coexistence.
