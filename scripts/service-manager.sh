#!/usr/bin/env bash
set -euo pipefail

# Compatibility shim: replacement commands are `devarch --workspace-root <root> workspace ...`.
# Set DEVARCH_WORKSPACE_ROOT or pass --workspace-root before the legacy command.

workspace_root="${DEVARCH_WORKSPACE_ROOT:-}"

usage() {
  cat <<'EOF'
Usage: service-manager.sh [--workspace-root ROOT] COMMAND [TARGET] [OPTIONS]

Supported compatibility commands:
  list                         -> devarch --workspace-root ROOT workspace list
  status WORKSPACE             -> devarch --workspace-root ROOT workspace status WORKSPACE
  up WORKSPACE                 -> devarch --workspace-root ROOT workspace apply WORKSPACE
  logs WORKSPACE RESOURCE      -> devarch --workspace-root ROOT workspace logs WORKSPACE RESOURCE
  restart WORKSPACE RESOURCE   -> devarch --workspace-root ROOT workspace restart WORKSPACE RESOURCE
  check                        -> devarch doctor
  help, -h, --help             -> this help

Options forwarded for logs:
  --tail N, --follow

Unsupported legacy commands (blocked, not silently ignored):
  down                         No devarch workspace stop command exists yet.
  rebuild                      No devarch workspace rebuild command exists yet.
  start, stop CATEGORY         Category orchestration is retired from this shim.
  compose                      Compose generation is not exposed by devarch CLI.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace-root)
      [[ $# -lt 2 ]] && { echo 'service-manager.sh: --workspace-root requires a value' >&2; exit 1; }
      workspace_root="$2"
      shift 2
      ;;
    -h|--help|help)
      usage
      exit 0
      ;;
    *)
      break
      ;;
  esac
done

cmd="${1:-help}"
[[ $# -gt 0 ]] && shift || true

need_workspace_root() {
  if [[ -z "$workspace_root" ]]; then
    echo 'service-manager.sh: workspace commands require --workspace-root ROOT or DEVARCH_WORKSPACE_ROOT.' >&2
    exit 2
  fi
}

run_workspace() {
  need_workspace_root
  exec devarch --workspace-root "$workspace_root" workspace "$@"
}

unsupported() {
  printf 'service-manager.sh: unsupported legacy command %q. See docs/devarch-v2/script-migration.md.\n' "$1" >&2
  exit 2
}

case "$cmd" in
  list)
    run_workspace list "$@"
    ;;
  status)
    [[ $# -lt 1 ]] && { echo 'service-manager.sh: status requires WORKSPACE.' >&2; exit 1; }
    run_workspace status "$@"
    ;;
  up)
    [[ $# -lt 1 ]] && { echo 'service-manager.sh: up requires WORKSPACE.' >&2; exit 1; }
    run_workspace apply "$@"
    ;;
  logs)
    run_workspace logs "$@"
    ;;
  restart)
    [[ $# -lt 2 ]] && { echo 'service-manager.sh: restart requires WORKSPACE RESOURCE.' >&2; exit 1; }
    run_workspace restart "$@"
    ;;
  check)
    exec devarch doctor "$@"
    ;;
  down|rebuild|start|stop|compose)
    unsupported "$cmd"
    ;;
  help|h|-h|--help)
    usage
    ;;
  *)
    printf 'service-manager.sh: unknown command %q\n\n' "$cmd" >&2
    usage >&2
    exit 1
    ;;
esac
