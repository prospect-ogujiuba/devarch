# `devarch` CLI

Phase 5 exposes workflow services through one CLI binary. Command handlers are thin wrappers over `internal/appsvc`.

DevArch v2 product scope is the Go CLI workflow surface: `doctor`, `runtime`, `socket`, `catalog`, `scan`, `import`, and `workspace`. Legacy shell entrypoints are retired; AI, WordPress, Laravel, database bootstrap, and generated context scripts require separate product decisions before becoming v2 CLI commands.

## Global flags

Place global flags before the command:

```bash
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin --json workspace plan shop-local
```

- `--workspace-root` repeatable workspace discovery root
- `--catalog-root` repeatable catalog discovery root
- `--json` stable machine-readable output

## Operator workflow examples

```bash
devarch --json doctor
devarch runtime status
devarch socket status
devarch socket start
devarch socket stop
devarch --workspace-root ./examples/v2/workspaces workspace status shop-local
devarch --workspace-root ./examples/v2/workspaces workspace apply shop-local
devarch --workspace-root ./examples/v2/workspaces workspace logs shop-local api
devarch --workspace-root ./examples/v2/workspaces workspace exec shop-local api -- echo ok
```

## Legacy script migration

Use the Go `devarch` binary directly instead of retired shell shims:

| Removed script | Replacement |
| --- | --- |
| `scripts/devarch` | `devarch <command>` |
| `scripts/devarch-doctor.sh` | `devarch doctor` or `devarch --json doctor` |
| `scripts/runtime-switcher.sh` | `devarch runtime status` |
| `scripts/socket-manager.sh` | `devarch socket status|start|stop` |
| `scripts/service-manager.sh` | `devarch --workspace-root <root> workspace <status|logs|exec|restart|apply> ...` |
| `scripts/setup-aliases.sh` | Install the Go binary on `PATH`; no aliases are required. |

## Stable JSON contract

`--json` emits the same service-backed payload shapes used by the thin API where they already exist:

- `doctor`, `runtime status`, `socket status/start/stop`
- `workspace list/open/plan/apply/status/logs/exec/restart`
- `catalog list/show`
- `scan project`
- `import v1-stack`
- `import v1-library`

Human-readable output is operator-oriented and may change.

## Import output

`import v1-stack` and `import v1-library` now execute the Phase 6 importer slice.
JSON output returns structured statuses, diagnostics, and emitted artifact documents so callers can distinguish supported, lossy, and rejected mappings explicitly.
