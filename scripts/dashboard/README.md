# DevArch Home

DevArch Home is a small, read-only localhost dashboard for discovering DevArch projects, running Podman containers, and Compose services. It helps you open a local site or port and copy native commands; it does not manage containers.

## Install and open

Install the persistent user service and reload the running Nginx Proxy Manager configuration:

```bash
scripts/dashboard/install-service.sh
```

Then open <https://devarch.test>. The installer enables `devarch-dashboard.service` for the current user, so no terminal needs to remain open. Useful service commands are:

```bash
systemctl --user status devarch-dashboard.service
systemctl --user restart devarch-dashboard.service
journalctl --user -u devarch-dashboard.service
```

The compiled Tailwind CSS is tracked, so normal use needs only Python 3, Podman, and user-systemd. `devarch.test` must resolve to `127.0.0.1`; `scripts/hosts/sync-hosts.sh` manages that entry. The existing Nginx Proxy Manager container terminates local HTTPS and proxies the domain to the host service.

For temporary direct development, run `scripts/dashboard/start.sh` and open <http://127.0.0.1:7411>. The direct server binds to `127.0.0.1` by default; use `--port` to choose another port.

Inventory loads once when the page opens. It never polls. Use **Refresh** when projects or containers change.

The responsive navigation provides separate **Apps**, **Containers**, and **Services** pages. Every app and catalog service also has a bookmarkable detail URL, such as `/apps/storefront` or `/services/database/postgres`. Legacy `/projects` links remain compatible and redirect in the browser to `/apps`.

## What it discovers

- **Projects** — directories directly under `apps/`, with common framework detection and an inferred `https://<name>.test` URL.
- **Running containers** — the secret-safe subset of `podman ps --format json`: name, image, state, status, ID, and published ports.
- **Service library** — `services-library/<category>/<service>/compose.yml` entries with a copyable native startup command.

The folder action uses the `vscode://file` URL scheme supported by VS Code and compatible editors. Container port links are best-effort HTTP links; not every published port speaks HTTP.

If Podman is unavailable, projects and catalog services remain usable and the page shows the runtime error.

## Tailwind development

Frontend dependencies are isolated in this directory:

```bash
cd scripts/dashboard
npm install
npm run watch
```

Create the minified tracked stylesheet before committing UI changes:

```bash
npm run build
```

## Tests

```bash
python -m unittest discover -s scripts/dashboard/tests -v
bash -n scripts/dashboard/start.sh scripts/dashboard/install-service.sh
python -m py_compile scripts/dashboard/server.py
```

## Scope and safety

The dashboard has no lifecycle controls, terminal, metrics, authentication, database, or remote runtime connection. The persistent service listens for the local proxy on port 7411, but rejects HTTP hostnames other than `devarch.test`, `dashboard.test`, and loopback names. Nginx publishes the intended HTTPS endpoint only on `127.0.0.1`.
