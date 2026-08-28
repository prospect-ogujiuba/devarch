#!/usr/bin/env bash
set -Eeuo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="$PROJECT_ROOT/apps"
ROUTER_COMPOSE="$PROJECT_ROOT/services-library/backend/node/compose.yml"
APP_COMPOSE="$PROJECT_ROOT/services-library/backend/node/app.compose.yml"
PROXY_COMPOSE="$PROJECT_ROOT/services-library/proxy/nginx-proxy-manager/compose.yml"
HOSTS_HELPER="$PROJECT_ROOT/scripts/hosts/register-host.sh"

APP_NAME=""
PACKAGE_SCRIPT=devarch
PACKAGE_MANAGER=auto
REGISTER_HOSTS=true
DRY_RUN=false
RUNTIME=""
CONTAINER_USER=""
TARGET=""
COMPOSE=()

log() { printf '[node] %s\n' "$*"; }
die() { printf '[node] error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: scripts/node/bootstrap.sh <app-name> [options]

Run an existing apps/<app-name> JavaScript application in an isolated Node
container and expose it through the wildcard https://<app-name>.test proxy.
The package script must bind its HTTP server to 0.0.0.0:3000.

Options:
  --script NAME                 Package script to run (default: devarch)
  --package-manager auto|npm|pnpm|yarn
                                Select installer/runner (default: auto)
  --no-hosts                    Do not register <app-name>.test
  --dry-run                     Validate and print a mutation-free plan
  --help                        Show this help (must be the sole argument)
EOF
}

usage_error() {
  printf '[node] error: %s\n' "$1" >&2
  usage >&2
  exit 2
}

parse_args() {
  if [[ "${1:-}" == --help ]]; then
    [[ $# -eq 1 ]] || usage_error '--help must be used alone'
    usage
    exit 0
  fi
  [[ $# -gt 0 ]] || usage_error 'app-name is required'

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --script)
        [[ $# -ge 2 && -n "$2" ]] || usage_error '--script requires a value'
        PACKAGE_SCRIPT="$2"
        shift 2
        ;;
      --package-manager)
        [[ $# -ge 2 && -n "$2" ]] || usage_error '--package-manager requires a value'
        PACKAGE_MANAGER="$2"
        shift 2
        ;;
      --no-hosts) REGISTER_HOSTS=false; shift ;;
      --dry-run) DRY_RUN=true; shift ;;
      --*) usage_error "unknown option: $1" ;;
      *)
        [[ -z "$APP_NAME" ]] || usage_error "unexpected extra positional argument: $1"
        APP_NAME="$1"
        shift
        ;;
    esac
  done
}

validate_inputs() {
  [[ "$APP_NAME" =~ ^[a-z0-9]([a-z0-9-]{0,56}[a-z0-9])?$ ]] || \
    die 'app-name must be a lowercase DNS label of at most 58 characters (letters, numbers, and interior hyphens)'
  [[ "$PACKAGE_SCRIPT" =~ ^[A-Za-z0-9:_-]+$ ]] || die 'package script contains unsupported characters'
  case "$PACKAGE_MANAGER" in auto|npm|pnpm|yarn) ;; *) die 'package manager must be auto, npm, pnpm, or yarn' ;; esac

  TARGET="$APPS_DIR/$APP_NAME"
  [[ -d "$TARGET" ]] || die "application directory does not exist: apps/$APP_NAME"
  [[ -f "$TARGET/package.json" ]] || die "package.json does not exist: apps/$APP_NAME/package.json"

  command -v python3 >/dev/null 2>&1 || die 'python3 is required to validate package.json'
  python3 - "$TARGET/package.json" "$PACKAGE_SCRIPT" <<'PY' || die "package.json does not define scripts.$PACKAGE_SCRIPT"
import json, sys
try:
    data = json.load(open(sys.argv[1], encoding='utf-8'))
except (OSError, json.JSONDecodeError):
    raise SystemExit(1)
script = data.get('scripts', {}).get(sys.argv[2])
raise SystemExit(0 if isinstance(script, str) and script.strip() else 1)
PY

  if [[ "$PACKAGE_MANAGER" == auto ]]; then
    if [[ -f "$TARGET/pnpm-lock.yaml" ]]; then PACKAGE_MANAGER=pnpm
    elif [[ -f "$TARGET/yarn.lock" ]]; then PACKAGE_MANAGER=yarn
    else PACKAGE_MANAGER=npm
    fi
  fi
}

detect_runtime() {
  if [[ -n "${CONTAINER_RUNTIME:-}" ]]; then
    case "$CONTAINER_RUNTIME" in podman|docker) RUNTIME="$CONTAINER_RUNTIME" ;; *) die 'CONTAINER_RUNTIME must be podman or docker' ;; esac
  elif command -v podman >/dev/null 2>&1; then RUNTIME=podman
  elif command -v docker >/dev/null 2>&1; then RUNTIME=docker
  else die 'Podman or Docker is required'
  fi

  command -v "$RUNTIME" >/dev/null 2>&1 || die "$RUNTIME is not installed"
  "$RUNTIME" compose version >/dev/null 2>&1 || die "$RUNTIME compose is unavailable"
  COMPOSE=("$RUNTIME" compose)
  if [[ "$RUNTIME" == docker ]]; then CONTAINER_USER="$(id -u):$(id -g)"
  else CONTAINER_USER="0:0"
  fi
}

print_plan() {
  log "app: $APP_NAME (https://$APP_NAME.test)"
  log "container: node-$APP_NAME"
  log "package manager: $PACKAGE_MANAGER"
  log "package script: $PACKAGE_SCRIPT"
  log 'start shared Node router and wildcard proxy'
  log 'start isolated app runtime on microservices-net'
  if [[ "$REGISTER_HOSTS" == true ]]; then log "register local host: 127.0.0.1 $APP_NAME.test"
  else log "hosts registration skipped: $APP_NAME.test"
  fi
}

ensure_network() {
  if ! "$RUNTIME" network inspect microservices-net >/dev/null 2>&1; then
    "$RUNTIME" network create microservices-net >/dev/null
  fi
}

start_services() {
  "${COMPOSE[@]}" -f "$ROUTER_COMPOSE" up -d --build
  "${COMPOSE[@]}" -f "$PROXY_COMPOSE" up -d
  DEVARCH_NODE_APP_NAME="$APP_NAME" \
  DEVARCH_NODE_SCRIPT="$PACKAGE_SCRIPT" \
  DEVARCH_NODE_PACKAGE_MANAGER="$PACKAGE_MANAGER" \
  DEVARCH_NODE_CONTAINER_USER="$CONTAINER_USER" \
    "${COMPOSE[@]}" -p "devarch-node-$APP_NAME" -f "$APP_COMPOSE" up -d --build --force-recreate

  "${COMPOSE[@]}" -f "$PROXY_COMPOSE" exec -T nginx-proxy-manager nginx -t >/dev/null
  "${COMPOSE[@]}" -f "$PROXY_COMPOSE" exec -T nginx-proxy-manager nginx -s reload >/dev/null
}

register_host() {
  [[ "$REGISTER_HOSTS" == true ]] || return 0
  [[ -x "$HOSTS_HELPER" ]] || die "hosts helper is not executable: $HOSTS_HELPER"
  "$HOSTS_HELPER" "$APP_NAME.test"
}

main() {
  parse_args "$@"
  validate_inputs
  print_plan
  [[ "$DRY_RUN" == true ]] && return 0
  detect_runtime
  ensure_network
  start_services
  register_host
  log "ready: https://$APP_NAME.test"
}

main "$@"
