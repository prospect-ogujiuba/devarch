# Script Migration

Compatibility shell scripts now delegate to the root `devarch` CLI. New workflow logic belongs in Go appsvc/workflow packages, not shell.

## Command map

| Old command | New command | Compatibility notes |
| --- | --- | --- |
| `scripts/devarch-doctor.sh [args...]` | `devarch doctor [args...]` | Direct shim. |
| `scripts/socket-manager.sh status` / `s` | `devarch socket status` | Direct shim. |
| `scripts/socket-manager.sh start` / `start-rootless` / `sr` | `devarch socket start` | Starts the Podman-first socket workflow exposed by CLI. |
| `scripts/socket-manager.sh stop` | `devarch socket stop` | Direct shim. |
| `scripts/runtime-switcher.sh status` / `s` | `devarch runtime status` | Direct shim. |
| `scripts/service-manager.sh --workspace-root ROOT list` | `devarch --workspace-root ROOT workspace list` | `DEVARCH_WORKSPACE_ROOT=ROOT` may replace the flag. |
| `scripts/service-manager.sh --workspace-root ROOT status WORKSPACE` | `devarch --workspace-root ROOT workspace status WORKSPACE` | Service-manager now uses workspace terms. |
| `scripts/service-manager.sh --workspace-root ROOT up WORKSPACE` | `devarch --workspace-root ROOT workspace apply WORKSPACE` | Legacy `up` maps to workspace apply. |
| `scripts/service-manager.sh --workspace-root ROOT logs WORKSPACE RESOURCE [--tail N] [--follow]` | `devarch --workspace-root ROOT workspace logs [--tail N] [--follow] WORKSPACE RESOURCE` | Uses the CLI log flags. |
| `scripts/service-manager.sh --workspace-root ROOT restart WORKSPACE RESOURCE` | `devarch --workspace-root ROOT workspace restart WORKSPACE RESOURCE` | Resource name is explicit. |
| `scripts/service-manager.sh check` | `devarch doctor` | Runtime prerequisite checks moved to doctor. |

## Unsupported legacy behavior

These old commands now fail explicitly with exit code `2` instead of silently dropping behavior:

- `socket-manager.sh start-rootful` / `sf`: rootful socket mode is not exposed by `devarch socket` yet.
- `socket-manager.sh test` / `t`: use `devarch socket status` until a dedicated socket test command exists.
- `socket-manager.sh logs` / `l`: use `journalctl` directly until socket logs are exposed by CLI.
- `socket-manager.sh fix` / `f`: manual socket repair is not exposed by CLI yet.
- `socket-manager.sh nuke` / `n`: destructive reset is intentionally not exposed by the compatibility shim.
- `socket-manager.sh env` / `e`: use `devarch socket status` and Podman socket documentation.
- `runtime-switcher.sh podman`: retired; Podman is the devarch v2 default.
- `runtime-switcher.sh docker`: Docker switching is compatibility-only and not a devarch v2 operator workflow.
- `service-manager.sh down`: no `devarch workspace stop` command exists yet.
- `service-manager.sh rebuild`: no `devarch workspace rebuild` command exists yet.
- `service-manager.sh start|stop CATEGORY`: category orchestration is retired from this shim.
- `service-manager.sh compose`: compose generation is not exposed by `devarch workspace` yet.

## Deprecation timing

- Current phase: scripts remain as compatibility shims for existing callers.
- Next release after v2 operator workflow adoption: announce script deprecation in release notes.
- Two releases after announcement: remove script entry points or leave tiny error shims pointing to `devarch` commands.
