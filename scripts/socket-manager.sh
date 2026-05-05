#!/usr/bin/env bash
set -euo pipefail

# Compatibility shim: replacement commands are `devarch socket status|start|stop`.

usage() {
  cat <<'EOF'
Usage: socket-manager.sh COMMAND [ARGS]

Supported compatibility commands:
  status, s              -> devarch socket status
  start, start-rootless, sr
                         -> devarch socket start
  stop                   -> devarch socket stop
  help, h, -h, --help    -> this help

Unsupported legacy commands (blocked, not silently ignored):
  start-rootful, sf      Rootful socket mode is not exposed by devarch CLI yet.
  test, t                Use `devarch socket status` until a CLI test command exists.
  logs, l                Use journalctl directly until devarch exposes socket logs.
  fix, f                Manual repair is not exposed by devarch CLI yet.
  nuke, n               Destructive reset is intentionally not exposed by this shim.
  env, e                Use `devarch socket status` and Podman socket docs.
EOF
}

unsupported() {
  printf 'socket-manager.sh: unsupported legacy command %q. See docs/devarch-v2/script-migration.md.\n' "$1" >&2
  exit 2
}

cmd="${1:-status}"
case "$cmd" in
  status|s)
    shift || true
    exec devarch socket status "$@"
    ;;
  start|start-rootless|sr)
    shift || true
    exec devarch socket start "$@"
    ;;
  stop)
    shift || true
    exec devarch socket stop "$@"
    ;;
  help|h|-h|--help)
    usage
    ;;
  start-rootful|sf|test|t|logs|l|fix|f|nuke|n|env|e)
    unsupported "$cmd"
    ;;
  *)
    printf 'socket-manager.sh: unknown command %q\n\n' "$cmd" >&2
    usage >&2
    exit 1
    ;;
esac
