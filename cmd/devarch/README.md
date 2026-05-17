# `devarch` CLI

DevArch exposes workflow services through one CLI binary. Command handlers are thin wrappers over `internal/appsvc`.

DevArch product scope is the Go CLI workflow surface: `doctor`, `runtime`, `socket`, `catalog`, `scan`, and `workspace`. Legacy shell entrypoints, AI APIs, WordPress/Laravel scaffolding, database bootstrap scripts, and generated context scripts are outside this CLI surface.

## Global flags

Place global flags before the command:

```bash
devarch --workspace-root ./examples/workspaces --catalog-root ./catalog/builtin --json workspace plan shop-local
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
devarch --workspace-root ./examples/workspaces workspace status shop-local
devarch --workspace-root ./examples/workspaces workspace apply shop-local
devarch --workspace-root ./examples/workspaces workspace logs shop-local api
devarch --workspace-root ./examples/workspaces workspace exec shop-local api -- echo ok
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

Human-readable output is operator-oriented and may change.

