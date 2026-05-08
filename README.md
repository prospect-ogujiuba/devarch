# DevArch

DevArch is a local Go CLI for defining, planning, applying, and observing development workspaces.

It solves a common local-development problem: projects grow a pile of fragile shell scripts, ad-hoc Compose files, aliases, and one-off container conventions. DevArch replaces that with declarative workspace manifests, reusable catalog templates, and a deterministic `plan -> apply -> status` workflow.

## What DevArch does

- Discovers v2 workspace manifests from configured roots.
- Resolves workspace resources against reusable catalog templates.
- Plans runtime changes before applying them.
- Applies local containers and networks through Podman-oriented runtime adapters.
- Shows status, logs, exec, and restart operations through one CLI.
- Imports selected v1 stack/library fixtures for migration work.

## Current architecture

The supported operator surface is the `devarch` binary in `cmd/devarch`.

DevArch CLI does **not** require the old DevArch API/app container. Runtime operations need local Podman, and socket commands manage/check `podman.socket`. The web UI/API surface is separate and currently expects a local `devarchd` if used; it is not required for CLI workflows.

## Install

From this repository:

```bash
go install ./cmd/devarch
export PATH="$HOME/go/bin:$PATH"
devarch --help
```

No aliases are required. Remove old aliases that point to `scripts/devarch`; that legacy shell entrypoint is retired.

## Command map

```txt
devarch doctor
devarch runtime status
devarch socket status|start|stop
devarch catalog list
devarch catalog show <template>
devarch scan project <path>
devarch import v1-stack <file>
devarch import v1-library <path>
devarch workspace list
devarch workspace open <name>
devarch workspace plan <name>
devarch workspace apply <name>
devarch workspace status <name>
devarch workspace logs <name> <resource>
devarch workspace exec <name> <resource> -- <command...>
devarch workspace restart <name> <resource>
```

Global flags must appear before the command:

```bash
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin --json workspace plan shop-local
```

## Documentation

- [Overview](docs/overview.md) — what DevArch is and the problem it solves.
- [Getting started](docs/getting-started.md) — install, verify, and first commands.
- [Concepts](docs/concepts.md) — workspaces, resources, templates, catalogs, runtimes, contracts.
- [Small DB tutorial](docs/tutorial-small-db.md) — end-to-end MariaDB + Adminer workflow.
- [DB admin/proxy stack from scratch](docs/tutorial-db-admin-proxy.md) — starts at `devarch: command not found`, then uses native DevArch commands for PostgreSQL and service-library import preview for MariaDB/Adminer/Nginx Proxy Manager.
- [Troubleshooting](docs/troubleshooting.md) — aliases, missing binary, Podman/socket, pcleanall, API/container questions.
- [CLI details](cmd/devarch/README.md) — command-specific notes and legacy migration table.

## Scope

Current product scope is intentionally narrow:

- one Go CLI entrypoint: `cmd/devarch`
- v2 workspace discovery, planning, applying, status, logs, exec, and restart
- built-in catalog inspection
- project scanning
- v1 import helpers for migration only
- Podman-oriented runtime/socket checks

Removed legacy surfaces:

- nested API service and daemon code as a required CLI dependency
- AI API/service code
- React/dashboard compose stack assumptions
- shell script product surface
- generated context snapshots
- WordPress/Laravel/app scaffolding workflows

If one of those capabilities becomes product scope again, reintroduce it as a small, tested Go CLI command with an explicit command contract.

## Development

```bash
go test ./...
go run ./cmd/devarch --help
```
