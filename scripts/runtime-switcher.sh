#!/usr/bin/env bash
set -euo pipefail

# Compatibility shim: replacement command is `devarch runtime status`.
# Podman is the default runtime in devarch v2; Docker switching is not exposed.

usage() {
  cat <<'EOF'
Usage: runtime-switcher.sh COMMAND

Supported compatibility commands:
  status, s              -> devarch runtime status
  help, h, -h, --help    -> this help

Unsupported legacy commands (blocked, not silently ignored):
  podman                 Runtime switching is retired; Podman is the devarch v2 default.
  docker                 Docker switching is compatibility-only and not exposed by devarch CLI.
EOF
}

cmd="${1:-status}"
case "$cmd" in
  status|s)
    shift || true
    exec devarch runtime status "$@"
    ;;
  help|h|-h|--help)
    usage
    ;;
  podman|docker)
    printf 'runtime-switcher.sh: unsupported legacy command %q. See docs/devarch-v2/script-migration.md.\n' "$cmd" >&2
    exit 2
    ;;
  *)
    printf 'runtime-switcher.sh: unknown command %q\n\n' "$cmd" >&2
    usage >&2
    exit 1
    ;;
esac
