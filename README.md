# DevArch

DevArch is now a local Go CLI for planning and applying v2 development workspaces.

The supported operator surface is the `devarch` binary:

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

## Scope

Current product scope is intentionally narrow:

- one Go CLI entrypoint: `cmd/devarch`
- v2 workspace discovery, planning, applying, status, logs, exec, and restart
- built-in catalog inspection
- project scanning
- v1 import helpers for migration only
- Podman-oriented runtime/socket checks

Removed legacy surfaces:

- nested API service and daemon code
- AI API/service code
- React/dashboard compose stack assumptions
- shell script product surface
- generated context snapshots
- WordPress/Laravel/app scaffolding workflows

If one of those capabilities becomes product scope again, reintroduce it as a small, tested Go CLI command with an explicit command contract.

## Usage

```bash
go run ./cmd/devarch help
go run ./cmd/devarch --json doctor
go run ./cmd/devarch catalog list
go run ./cmd/devarch --workspace-root ./examples/v2/workspaces workspace status shop-local
```

Global flags must appear before the command:

```bash
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin --json workspace plan shop-local
```

## Development

```bash
go test ./...
go run ./cmd/devarch help
```

See `cmd/devarch/README.md` for CLI details and migration notes.
