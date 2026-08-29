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
LOG_DIR="${DEVARCH_MATRIX_LOG_DIR:-$PROJECT_ROOT/.model-artifacts/logs/javascript-scaffold-matrix}"
SCAFFOLD_ATTEMPTS="${DEVARCH_MATRIX_ATTEMPTS:-2}"
SCAFFOLD_RESULT=''

log() { printf '[javascript-matrix] %s\n' "$*"; }
die() { printf '[javascript-matrix] error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: scripts/javascript/scaffold-matrix.sh list
       scripts/javascript/scaffold-matrix.sh scaffold <framework> <profile>
       scripts/javascript/scaffold-matrix.sh scaffold-all
       scripts/javascript/scaffold-matrix.sh start <framework> <profile> [node-bootstrap-options]
       scripts/javascript/scaffold-matrix.sh start-all [node-bootstrap-options]
       scripts/javascript/scaffold-matrix.sh stop <framework> <profile>

Discover the curated JavaScript profile matrix and manage deterministic
apps/showcase-<framework>-<profile> applications. Bulk commands run
sequentially: scaffold-all skips existing applications, while start-all
starts only applications that already contain package.json. Upstream scaffold
output is captured to a run log so terminal progress stays concise.

Environment:
  DEVARCH_MATRIX_APP_PREFIX  Application prefix (default: showcase)
  DEVARCH_MATRIX_ATTEMPTS    Attempts per profile in scaffold-all (default: 2)
  DEVARCH_MATRIX_LOG_DIR     Run log directory (default: .model-artifacts/logs/...)
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
  [[ "$SCAFFOLD_ATTEMPTS" =~ ^[1-9][0-9]*$ ]] || die 'DEVARCH_MATRIX_ATTEMPTS must be a positive integer'
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

scaffold_attempt() {
  local framework=$1 profile=$2 name target status
  name=$(app_name "$framework" "$profile")
  target="$APPS_DIR/$name"

  set +e
  "$SCAFFOLD_BOOTSTRAP" "$name" --framework "$framework" --profile "$profile"
  status=$?
  set -e
  ((status == 0)) || return "$status"
  [[ -f "$target/package.json" ]] || {
    printf '[javascript-matrix] error: scaffold completed without package.json: %s\n' "$target" >&2
    return 1
  }
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
  scaffold_attempt "$framework" "$profile" || die "scaffolder failed: $framework/$profile"
  SCAFFOLD_RESULT=created
}

new_run_log() {
  local base candidate suffix=1
  mkdir -p "$LOG_DIR"
  base="$LOG_DIR/$(date +%Y-%m-%d_%H%M)-scaffold-all"
  candidate="$base.log"
  while [[ -e "$candidate" ]]; do
    ((suffix += 1))
    candidate="$base-$suffix.log"
  done
  printf '%s\n' "$candidate"
}

scaffold_all() {
  local -a paths=() failures=()
  local path framework profile name target run_log attempt status
  local created=0 skipped=0 failed=0 current total
  mapfile -d '' paths < <(profile_paths)
  total=${#paths[@]}
  run_log=$(new_run_log)
  log "scaffold-all: total=$total attempts=$SCAFFOLD_ATTEMPTS log=$run_log"

  for current in "${!paths[@]}"; do
    path=${paths[$current]}
    framework=$(basename "$(dirname "$path")")
    profile=$(basename "$path" .profile)
    name=$(app_name "$framework" "$profile")
    target="$APPS_DIR/$name"
    current=$((current + 1))

    if [[ -f "$target/package.json" ]]; then
      printf '[javascript-matrix] [%d/%d] %s/%s — skipped\n' "$current" "$total" "$framework" "$profile"
      ((skipped += 1))
      continue
    fi
    if [[ -e "$target" ]]; then
      printf '[javascript-matrix] [%d/%d] %s/%s — failed (incomplete target)\n' "$current" "$total" "$framework" "$profile" >&2
      printf '\n===== %s/%s =====\nexisting target is incomplete: %s\n' "$framework" "$profile" "$target" >> "$run_log"
      failures+=("$framework/$profile")
      ((failed += 1))
      continue
    fi

    status=1
    for ((attempt = 1; attempt <= SCAFFOLD_ATTEMPTS; attempt++)); do
      printf '[javascript-matrix] [%d/%d] %s/%s — attempt %d/%d\n' \
        "$current" "$total" "$framework" "$profile" "$attempt" "$SCAFFOLD_ATTEMPTS"
      printf '\n===== %s/%s attempt %d/%d =====\n' \
        "$framework" "$profile" "$attempt" "$SCAFFOLD_ATTEMPTS" >> "$run_log"
      if scaffold_attempt "$framework" "$profile" >> "$run_log" 2>&1; then
        status=0
        break
      fi
      printf '[javascript-matrix] [%d/%d] %s/%s — attempt %d failed%s\n' \
        "$current" "$total" "$framework" "$profile" "$attempt" \
        "$([[ $attempt -lt $SCAFFOLD_ATTEMPTS ]] && printf ', retrying' || true)" >&2
    done

    if ((status == 0)); then
      printf '[javascript-matrix] [%d/%d] %s/%s — created\n' "$current" "$total" "$framework" "$profile"
      ((created += 1))
    else
      printf '[javascript-matrix] [%d/%d] %s/%s — failed; log tail follows\n' "$current" "$total" "$framework" "$profile" >&2
      tail -n 12 "$run_log" >&2
      failures+=("$framework/$profile")
      ((failed += 1))
    fi
  done

  log "complete: total=$total created=$created skipped=$skipped failed=$failed log=$run_log"
  if ((failed > 0)); then
    log "failed profiles: ${failures[*]}"
    return 1
  fi
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

start_all() {
  local -a paths=() failures=()
  local path framework profile name current total
  local started=0 skipped=0 failed=0
  [[ -x "$NODE_BOOTSTRAP" ]] || die "Node bootstrap is not executable: $NODE_BOOTSTRAP"
  mapfile -d '' paths < <(profile_paths)
  total=${#paths[@]}
  log "start-all: total=$total"

  for current in "${!paths[@]}"; do
    path=${paths[$current]}
    framework=$(basename "$(dirname "$path")")
    profile=$(basename "$path" .profile)
    name=$(app_name "$framework" "$profile")
    current=$((current + 1))

    if [[ ! -f "$APPS_DIR/$name/package.json" ]]; then
      ((skipped += 1))
      continue
    fi

    printf '[javascript-matrix] [%d/%d] %s/%s — starting\n' "$current" "$total" "$framework" "$profile"
    if "$NODE_BOOTSTRAP" "$name" "$@"; then
      ((started += 1))
    else
      printf '[javascript-matrix] [%d/%d] %s/%s — failed\n' "$current" "$total" "$framework" "$profile" >&2
      failures+=("$framework/$profile")
      ((failed += 1))
    fi
  done

  log "complete: total=$total started=$started skipped=$skipped failed=$failed"
  if ((failed > 0)); then
    log "failed applications: ${failures[*]}"
    return 1
  fi
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
    start-all)
      start_all "$@"
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
