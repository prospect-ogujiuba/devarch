# devarch CLI

`devarch` runs the service library by name. It is a thin layer over native `podman compose`: every action prints the command it runs, `--dry-run` prints without running, and nothing runs in the background. There is no daemon, API, or database; the only state is two small files under `~/.config/devarch/`.

The design and roadmap are in [`docs/plans/2026-10-02-devarch-cli-and-tui.md`](../docs/plans/2026-10-02-devarch-cli-and-tui.md).

## Install

Requires Go 1.24+ and rootless Podman with a Compose provider.

```bash
cd cli
go install ./cmd/devarch          # installs to $(go env GOPATH)/bin
devarch completion zsh > "${fpath[1]}/_devarch"   # or: bash, fish
```

Run `devarch` once from inside the checkout; it remembers the location in `~/.config/devarch/config.yml`. `DEVARCH_ROOT` overrides it.

## Interactive screen

Run `devarch` with no arguments in a terminal:

```text
DevArch  1 Services  2 Apps  3 Running  4 Doctor                       /maria
● database/mariadb  12.3    running            
──────────────────────────────────────────────────────────────────────────────
MariaDB · Shared MySQL-compatible database for WordPress and Laravel apps
database, sql, mysql · version 12.3 (MARIADB_VERSION; choices 12.3 11.8 11.4 10.11)
services-library/database/mariadb/compose.yml  (builtin)
containers: mariadb :8501
──────────────────────────────────────────────────────────────────────────────
$ (cd services-library/database/mariadb && podman compose up -d)
u up · d down · r restart · v version · l logs · o open · / filter · n new · H hosts · ? help · q quit
```

- **Services**: the whole catalog with live state. `/` filters, `u` `d` `r` start, stop and restart, `v` switches the version (asking before it recreates a running service), `l` follows logs inside the screen, `o` opens the service URL.
- **Apps**: `apps/` directories with their detected framework; `o` opens `https://<name>.test`, `e` opens the folder in VS Code.
- **Running**: every running container, including ones outside the catalog; `l` logs, `r` restart.
- **Doctor**: `devarch doctor` plus the catalog lint summary; `r` reruns it.
- **New project** (`n`): pick a recipe, fill in a form built from its `recipe.yml`, review the exact command, and watch its steps complete. Destructive flags such as `--force` need an explicit `y`.

The activity pane shows every native command and its output. State updates come from `podman events` while the screen is open; nothing runs after you quit. Hostname registration may need `sudo` or UAC, so the screen never runs it in the background: when a started service or new app is missing from the hosts file, it says so, and `H` runs `devarch hosts sync` interactively.

## Commands

```bash
devarch ls                          # catalog: version, state, title, URL
devarch ls queue                    # search IDs, titles, descriptions, tags
devarch ls --tag database --json
devarch up php mariadb nginx-proxy-manager --wait
devarch up --tag project-management
devarch ps                          # containers mapped to catalog services (-a for stopped)
devarch logs mariadb -f -n 100
devarch restart php
devarch down redis                  # volumes are kept
devarch down redis --volumes        # names the volumes and asks first (--yes to skip)
devarch compose postgres -- exec postgres psql -U postgres
devarch --dry-run up --tag project-management

devarch use php                     # list versions; * marks the selected one
devarch use php 8.3                 # record, rebuild, and recreate if running
devarch use postgres 17             # refused while a postgres volume exists (--force overrides)

devarch hosts sync                  # write every .test domain into the managed hosts block
devarch hosts add demo.test
devarch hosts list                  # show the block and whether it is current
devarch doctor                      # environment checks with fixes
devarch lint                        # catalog contract checks (--native, --strict, --json)
```

Services are named by canonical ID (`database/redis`) or a unique short name (`redis`). An ambiguous short name lists its candidates.

`up` creates the `microservices-net` network when it is missing, starts each service's `requires` first, retries a failed start once, and registers any unmapped `.test` hostname by synchronizing the hosts block (`--no-hosts` skips this). `--wait` blocks until every container passes its compose healthcheck and the optional `ready` probe.

`use` protects data volumes: a downgrade of an `upgrade-only` service, or a major-version change of a `same-major` service, is refused while its volume exists. Built services (`rebuild: true`) are rebuilt immediately, because `podman compose up -d` does not rebuild an existing image.

On WSL the hosts commands edit the Windows hosts file, the one the browser uses, through embedded PowerShell helpers that request UAC elevation; elsewhere they edit `/etc/hosts` (or `HOSTS_FILE`) with `sudo` when needed.

Any `devarch-<name>` executable on `PATH` runs as `devarch <name>`, with `DEVARCH_ROOT` set.

## Configuration

`~/.config/devarch/config.yml` holds a few settings. Every key is optional and defaults to the behavior described in this README:

```yaml
root: /home/me/devarch        # remembered checkout; DEVARCH_ROOT overrides it
runtime: podman               # podman or docker
apps_dir: apps                # where projects live; relative to the checkout, or absolute
network: microservices-net    # the shared external network every service joins
hosts:
  manage: true                # false: devarch never edits the hosts file
  address: 127.0.0.1          # where .test names point
editor: code                  # command that opens a project folder (Apps view, `e`)
```

```bash
devarch config                      # effective values and where each came from
devarch config get apps_dir
devarch config set editor 'code -n' # validates, keeps the file's comments
devarch config set hosts.manage false
```

How each setting travels:

- **`network`** and **`apps_dir`** reach compose files as `DEVARCH_NETWORK` and `DEVARCH_APPS_DIR`. Every service declares `name: ${DEVARCH_NETWORK:-microservices-net}` on the shared network (`devarch lint` enforces it), and the PHP, Nginx Proxy Manager and Node runtime mounts use `${DEVARCH_APPS_DIR:-../../../apps}`. DevArch only sets these when they differ from the defaults, so plain `podman compose` in a service directory keeps working. `devarch new` always passes both to the bootstraps; when running a bootstrap directly with a non-default `apps_dir`, export `DEVARCH_APPS_DIR` yourself.
- **`hosts.manage: false`** makes `up` skip registration, turns `hosts sync|add|remove` into no-ops that say so (bootstraps call `devarch hosts add`, so they follow the setting), and `doctor` reports the hosts file as not managed. Use it when DNS (dnsmasq, a router, Acrylic) already resolves `*.test`.
- **`hosts.address`** is the default for `--address` and what `doctor` and `hosts list` compare against.

Two things are deliberately not settings. The **`.test` suffix** is baked into the local certificate (`*.test`), the proxy's routing rules, every bootstrap's URLs and the hosts block, and `.test` is reserved (RFC 2606), so it never collides with real names. **Nginx Proxy Manager** is the one proxy because its config, the certificate path and the Node and PHP routing snippets are written for it; supporting a second proxy would mean a second copy of each.

## Hosts

`devarch hosts sync` writes every `.test` domain into one managed block: each catalog `container_name`, `.test` hosts listed in `x-devarch.urls`, every `apps/*` directory containing `index.php`, `public/index.php`, `public/index.html`, or `package.json`, and `devarch.test`. Only content between `# BEGIN DEVARCH HOSTS` and `# END DEVARCH HOSTS` is replaced, and an already-current block is not rewritten. Catalog services are included whether or not they are running.

`devarch hosts add NAME` idempotently maps one hostname to `127.0.0.1`: other mappings of that name are removed without touching other aliases on the same line, then one canonical line is appended. `devarch hosts remove NAME` unmaps one (Linux and macOS).

On Linux and macOS the commands edit `/etc/hosts` (or `HOSTS_FILE`) in place, using `sudo tee` when the file is not writable. Under WSL and Git Bash/MSYS they update `%SystemRoot%\System32\drivers\etc\hosts` through embedded PowerShell helpers that request UAC elevation and preserve the file's encoding and line endings. Run `devarch` as your normal user rather than with `sudo`; it calls Windows PowerShell through WSL's `/init` interop host so Wine or another `.exe` binfmt handler cannot intercept it. Set `DEVARCH_HOSTS_PLATFORM=unix|windows` to override detection.

## Service metadata (`x-devarch`)

Compose ignores top-level `x-` keys, so metadata lives in each service's `compose.yml`. Every field is optional; a service without the block still works with every command.

```yaml
x-devarch:
  title: PostgreSQL
  description: Relational database
  tags: [database, sql]
  urls: [https://pgadmin.test]
  notes: Shown after `up` and in listings
  version:
    var: POSTGRES_VERSION      # must appear in the file as ${POSTGRES_VERSION:-18.2}
    default: "18.2"
    choices: ["18.2", "17", "16", "15"]
    rebuild: false             # true when a Dockerfile takes the version as a build arg
    data: same-major           # "", upgrade-only, or same-major: data-volume compatibility
  ready:
    exec: [pg_isready, -U, postgres]   # extra probe inside the primary container
  requires: []                 # catalog IDs that `up` starts first
services:
  postgres:
    image: postgres:${POSTGRES_VERSION:-18.2}
```

Selected versions are stored in `~/.config/devarch/versions.env` and passed to `podman compose` as process environment, so running `podman compose up -d` by hand in the service directory still uses the default.

## Extending

Everything extends by adding a folder; there is no plugin API.

- **Services:** put `~/.config/devarch/services/<category>/<name>/compose.yml` (optionally with `x-devarch` metadata). It appears in `devarch ls` and the screen with source `user`, and a user service with the same ID as a built-in one replaces it, so you can customize a service without editing the repository.
- **Recipes:** put `~/.config/devarch/recipes/<name>/recipe.yml`; see [`recipes/README.md`](../recipes/README.md). A personal recipe with a built-in name replaces it.
- **Commands:** any executable named `devarch-<name>` on `PATH` runs as `devarch <name>` with `DEVARCH_ROOT` set, the way `git` runs `git-<name>`.

`DEVARCH_CONFIG_HOME` moves the whole `~/.config/devarch` directory, which is useful for trying changes in isolation.

## Development

```bash
cd cli
go test ./...
go vet ./...
gofmt -l .
```

Every process goes through `internal/runner`; tests use `runner.Fake` to assert the exact native commands.
