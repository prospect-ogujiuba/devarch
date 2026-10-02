#!/usr/bin/env bash
set -Eeuo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="${DEVARCH_APPS_DIR:-$PROJECT_ROOT/apps}"
APP_COMPOSE="$PROJECT_ROOT/services-library/backend/node/app.compose.yml"
PLATFORM_LIBRARY="$PROJECT_ROOT/scripts/devarch/lib/platform.sh"

# shellcheck source=../devarch/lib/platform.sh
source "$PLATFORM_LIBRARY"

APP_NAME=""
PACKAGE_SCRIPT=devarch
PACKAGE_MANAGER=auto
REGISTER_HOSTS=true
DRY_RUN=false
RUNTIME=""
CONTAINER_USER=""
TARGET=""
COMPOSE=()
DEVARCH=devarch
CURRENT_STEP=bootstrap

log() { printf '[node] %s\n' "$*"; }
die() {
  devarch_progress "$CURRENT_STEP" fail "$*"
  printf '[node] error: %s\n' "$*" >&2
  exit 1
}

# step NAME FUNCTION: run one provisioning step and report its progress.
step() {
  CURRENT_STEP="$1"
  devarch_progress "$1" start
  "$2"
  devarch_progress "$1" done
}

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
  RUNTIME="${CONTAINER_RUNTIME:-podman}"
  [[ "$RUNTIME" == podman ]] || die "only Podman is supported (CONTAINER_RUNTIME=$RUNTIME)"
  command -v podman >/dev/null 2>&1 || die 'Podman is required'
  podman compose version >/dev/null 2>&1 || die 'podman compose is unavailable'
  COMPOSE=(podman compose)
  # Root in a rootless Podman container maps to the invoking host user.
  CONTAINER_USER="0:0"
  DEVARCH="$(devarch_bin)" || die 'the devarch CLI is required: (cd cli && go install ./cmd/devarch)'
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


start_services() {
  "$DEVARCH" up backend/node proxy/nginx-proxy-manager --build --no-hosts
  DEVARCH_NODE_APP_NAME="$APP_NAME" \
  DEVARCH_NODE_SCRIPT="$PACKAGE_SCRIPT" \
  DEVARCH_NODE_PACKAGE_MANAGER="$PACKAGE_MANAGER" \
  DEVARCH_NODE_CONTAINER_USER="$CONTAINER_USER" \
    "${COMPOSE[@]}" -p "devarch-node-$APP_NAME" -f "$APP_COMPOSE" up -d --build --force-recreate

  "$DEVARCH" compose nginx-proxy-manager -- exec -T nginx-proxy-manager nginx -t >/dev/null
  "$DEVARCH" compose nginx-proxy-manager -- exec -T nginx-proxy-manager nginx -s reload >/dev/null
}

register_host() {
  [[ "$REGISTER_HOSTS" == true ]] || return 0
  "$DEVARCH" hosts add "$APP_NAME.test"
}

main() {
  parse_args "$@"
  validate_inputs
  print_plan
  [[ "$DRY_RUN" == true ]] && return 0
  detect_runtime
  step services start_services
  step hosts register_host
  log "ready: https://$APP_NAME.test"
}

main "$@"
