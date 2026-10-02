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

## Development

```bash
cd cli
go test ./...
go vet ./...
gofmt -l .
```

Every process goes through `internal/runner`; tests use `runner.Fake` to assert the exact native commands.
