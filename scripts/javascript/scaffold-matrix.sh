#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROFILES_DIR="$SCRIPT_DIR/profiles"
SCAFFOLD_BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"
NODE_BOOTSTRAP="$PROJECT_ROOT/scripts/node/bootstrap.sh"
APP_COMPOSE="$PROJECT_ROOT/services-library/backend/node/app.compose.yml"
APPS_DIR="$PROJECT_ROOT/apps"
APP_PREFIX="${DEVARCH_MATRIX_APP_PREFIX:-showcase}"
SCAFFOLD_RESULT=''

log() { printf '[javascript-matrix] %s\n' "$*"; }
die() { printf '[javascript-matrix] error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: scripts/javascript/scaffold-matrix.sh list
       scripts/javascript/scaffold-matrix.sh scaffold <framework> <profile>
       scripts/javascript/scaffold-matrix.sh scaffold-all
       scripts/javascript/scaffold-matrix.sh start <framework> <profile> [node-bootstrap-options]
       scripts/javascript/scaffold-matrix.sh stop <framework> <profile>

Discover the curated JavaScript profile matrix and manage deterministic
apps/showcase-<framework>-<profile> applications. scaffold-all runs
sequentially and skips applications that already contain package.json.

Environment:
  DEVARCH_MATRIX_APP_PREFIX  Application prefix (default: showcase)
  CONTAINER_RUNTIME          podman or docker (auto-detected for stop)
EOF
}

valid_slug() { [[ "$1" =~ ^[a-z0-9][a-z0-9-]*$ ]]; }

app_name() {
  printf '%s-%s-%s\n' "$APP_PREFIX" "$1" "$2"
}

profile_paths() {
  find "$PROFILES_DIR" -mindepth 2 -maxdepth 2 -type f -name '*.profile' -print0 | sort -z
}

validate_setup() {
  [[ -d "$PROFILES_DIR" ]] || die "profiles directory does not exist: $PROFILES_DIR"
  [[ -x "$SCAFFOLD_BOOTSTRAP" ]] || die "scaffold bootstrap is not executable: $SCAFFOLD_BOOTSTRAP"
  valid_slug "$APP_PREFIX" || die 'DEVARCH_MATRIX_APP_PREFIX must be lowercase DNS-safe text'
}

validate_combination() {
  local framework=$1 profile=$2
  valid_slug "$framework" || die "invalid framework: $framework"
  valid_slug "$profile" || die "invalid profile: $profile"
  [[ -f "$PROFILES_DIR/$framework/framework.conf" ]] ||
    die "unknown framework: $framework (use 'scaffold-matrix.sh list')"
  [[ -f "$PROFILES_DIR/$framework/$profile.profile" ]] ||
    die "unknown profile: $framework/$profile (use 'scaffold-matrix.sh list')"

  local name
  name=$(app_name "$framework" "$profile")
  ((${#name} <= 58)) || die "generated application name exceeds 58 characters: $name"
}

list_matrix() {
  local path framework profile name state total=0
  printf '%-20s %-18s %-42s %s\n' FRAMEWORK PROFILE APPLICATION STATE
  while IFS= read -r -d '' path; do
    framework=$(basename "$(dirname "$path")")
    profile=$(basename "$path" .profile)
    name=$(app_name "$framework" "$profile")
    state=missing
    [[ -f "$APPS_DIR/$name/package.json" ]] && state=scaffolded
    printf '%-20s %-18s %-42s %s\n' "$framework" "$profile" "$name" "$state"
    ((total += 1))
  done < <(profile_paths)
  printf '\nTotal: %d profiles\n' "$total"
}

scaffold_one() {
  local framework=$1 profile=$2 name target
  validate_combination "$framework" "$profile"
  name=$(app_name "$framework" "$profile")
  target="$APPS_DIR/$name"

  if [[ -f "$target/package.json" ]]; then
    log "skip $framework/$profile: apps/$name already contains package.json"
    SCAFFOLD_RESULT=skipped
    return 0
  fi
  [[ ! -e "$target" ]] || die "existing target is incomplete: $target"

  log "scaffold $framework/$profile as apps/$name"
  "$SCAFFOLD_BOOTSTRAP" "$name" --framework "$framework" --profile "$profile"
  [[ -f "$target/package.json" ]] || die "scaffold completed without package.json: $target"
  SCAFFOLD_RESULT=created
}

scaffold_all() {
  local path framework profile created=0 skipped=0 total=0
  while IFS= read -r -d '' path; do
    framework=$(basename "$(dirname "$path")")
    profile=$(basename "$path" .profile)
    scaffold_one "$framework" "$profile"
    case "$SCAFFOLD_RESULT" in
      created) ((created += 1)) ;;
      skipped) ((skipped += 1)) ;;
    esac
    ((total += 1))
  done < <(profile_paths)
  log "complete: total=$total created=$created skipped=$skipped"
}

start_one() {
  local framework=$1 profile=$2
  shift 2
  validate_combination "$framework" "$profile"
  [[ -x "$NODE_BOOTSTRAP" ]] || die "Node bootstrap is not executable: $NODE_BOOTSTRAP"

  local name
  name=$(app_name "$framework" "$profile")
  [[ -f "$APPS_DIR/$name/package.json" ]] ||
    die "application is not scaffolded: apps/$name (run scaffold first)"
  "$NODE_BOOTSTRAP" "$name" "$@"
}

detect_runtime() {
  if [[ -n ${CONTAINER_RUNTIME:-} ]]; then
    case "$CONTAINER_RUNTIME" in
      podman|docker) printf '%s\n' "$CONTAINER_RUNTIME" ;;
      *) die 'CONTAINER_RUNTIME must be podman or docker' ;;
    esac
  elif command -v podman >/dev/null 2>&1; then
    printf 'podman\n'
  elif command -v docker >/dev/null 2>&1; then
    printf 'docker\n'
  else
    die 'Podman or Docker is required to stop an application'
  fi
}

stop_one() {
  local framework=$1 profile=$2
  validate_combination "$framework" "$profile"
  [[ -f "$APP_COMPOSE" ]] || die "Node app Compose file does not exist: $APP_COMPOSE"

  local name runtime
  name=$(app_name "$framework" "$profile")
  runtime=$(detect_runtime)
  log "stop apps/$name"
  DEVARCH_NODE_APP_NAME="$name" \
    "$runtime" compose -p "devarch-node-$name" -f "$APP_COMPOSE" down
}

main() {
  (($# > 0)) || { usage; exit 2; }
  case "$1" in
    -h|--help|help) usage; exit 0 ;;
  esac
  validate_setup

  local command=$1
  shift
  case "$command" in
    list)
      (($# == 0)) || die 'list takes no arguments'
      list_matrix
      ;;
    scaffold)
      (($# == 2)) || die 'scaffold requires <framework> <profile>'
      scaffold_one "$1" "$2"
      ;;
    scaffold-all)
      (($# == 0)) || die 'scaffold-all takes no arguments'
      scaffold_all
      ;;
    start)
      (($# >= 2)) || die 'start requires <framework> <profile>'
      start_one "$@"
      ;;
    stop)
      (($# == 2)) || die 'stop requires <framework> <profile>'
      stop_one "$1" "$2"
      ;;
    *)
      die "unknown command: $command"
      ;;
  esac
}

main "$@"
