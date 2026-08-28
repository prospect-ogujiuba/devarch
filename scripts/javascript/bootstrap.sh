#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROFILES_DIR="$SCRIPT_DIR/profiles"
APPS_DIR="$PROJECT_ROOT/apps"

usage() {
  cat <<'EOF'
Usage: scripts/javascript/bootstrap.sh <app-name> --framework <framework> --profile <profile> [options]
       scripts/javascript/bootstrap.sh <app-name> --profile <legacy-framework> [options]
       scripts/javascript/bootstrap.sh --list-frameworks
       scripts/javascript/bootstrap.sh --list-profiles [--framework <framework>]

Create a current JavaScript framework starter under apps/<app-name>.

Options:
  --framework NAME  JavaScript framework or framework/tool combination
  --profile NAME    Project profile for the selected framework
  --start           Start the generated app through scripts/node/bootstrap.sh
  --force           Back up and replace an existing app
  --no-hosts        With --start, skip hosts-file registration
  --dry-run         Validate and print the plan without changing files
  --list-frameworks List available frameworks
  --list-profiles   List curated framework/profile combinations
  -h, --help        Show this help

Compatibility: --profile <framework> without --framework selects that
framework's default project profile.
EOF
}

die() { printf 'javascript bootstrap: %s\n' "$*" >&2; exit 1; }

valid_slug() { [[ "$1" =~ ^[a-z0-9][a-z0-9-]*$ ]]; }

load_framework() {
  local requested=$1 metadata="$PROFILES_DIR/$1/framework.conf"
  [[ -f "$metadata" ]] || return 1
  unset FRAMEWORK_DESCRIPTION DEFAULT_PROFILE
  # shellcheck source=/dev/null
  source "$metadata"
  [[ -n ${FRAMEWORK_DESCRIPTION:-} && -n ${DEFAULT_PROFILE:-} ]] ||
    die "invalid framework metadata: $requested"
}

load_project_profile() {
  local requested_framework=$1 requested_profile=$2
  profile_file="$PROFILES_DIR/$requested_framework/$requested_profile.profile"
  [[ -f "$profile_file" ]] || return 1
  unset PROFILE_DESCRIPTION DEVARCH_SCRIPT POST_INSTALL SCAFFOLD
  unset -f configure_app 2>/dev/null || true
  # shellcheck source=/dev/null
  source "$profile_file"
  [[ -n ${PROFILE_DESCRIPTION:-} && -n ${DEVARCH_SCRIPT:-} ]] ||
    die "invalid profile: $requested_framework/$requested_profile"
  declare -p SCAFFOLD >/dev/null 2>&1 ||
    die "profile has no scaffold command: $requested_framework/$requested_profile"
  ((${#SCAFFOLD[@]} > 0)) ||
    die "profile has an empty scaffold command: $requested_framework/$requested_profile"
}

list_frameworks() {
  local directory framework
  printf '%-20s %-16s %s\n' FRAMEWORK DEFAULT DESCRIPTION
  for directory in "$PROFILES_DIR"/*; do
    [[ -d "$directory" && -f "$directory/framework.conf" ]] || continue
    framework=$(basename "$directory")
    load_framework "$framework"
    printf '%-20s %-16s %s\n' "$framework" "$DEFAULT_PROFILE" "$FRAMEWORK_DESCRIPTION"
  done
}

list_profiles() {
  local filter=${1:-} directory file listed_framework profile marker
  if [[ -n "$filter" ]]; then
    valid_slug "$filter" || die 'invalid framework name'
    load_framework "$filter" || die "unknown framework: $filter (use --list-frameworks)"
  fi
  printf '%-20s %-18s %-8s %s\n' FRAMEWORK PROFILE DEFAULT DESCRIPTION
  for directory in "$PROFILES_DIR"/*; do
    [[ -d "$directory" && -f "$directory/framework.conf" ]] || continue
    listed_framework=$(basename "$directory")
    [[ -z "$filter" || "$listed_framework" == "$filter" ]] || continue
    load_framework "$listed_framework"
    for file in "$directory"/*.profile; do
      [[ -f "$file" ]] || continue
      profile=$(basename "$file" .profile)
      load_project_profile "$listed_framework" "$profile"
      marker=''
      [[ "$profile" == "$DEFAULT_PROFILE" ]] && marker='yes'
      printf '%-20s %-18s %-8s %s\n' "$listed_framework" "$profile" "$marker" "$PROFILE_DESCRIPTION"
    done
  done
}

app_name=''
framework=''
profile=''
start=0
force=0
no_hosts=0
dry_run=0
list_frameworks_requested=0
list_profiles_requested=0
legacy_alias=0

while (($#)); do
  case "$1" in
    --framework) (($# >= 2)) || die '--framework requires a value'; framework=$2; shift 2 ;;
    --profile) (($# >= 2)) || die '--profile requires a value'; profile=$2; shift 2 ;;
    --start) start=1; shift ;;
    --force) force=1; shift ;;
    --no-hosts) no_hosts=1; shift ;;
    --dry-run) dry_run=1; shift ;;
    --list-frameworks) list_frameworks_requested=1; shift ;;
    --list-profiles) list_profiles_requested=1; shift ;;
    -h|--help) usage; exit 0 ;;
    --*) die "unknown option: $1" ;;
    *) [[ -z "$app_name" ]] || die "unexpected argument: $1"; app_name=$1; shift ;;
  esac
done

if ((list_frameworks_requested)); then
  [[ -z "$app_name" && $list_profiles_requested -eq 0 ]] || die '--list-frameworks cannot be combined with an app or --list-profiles'
  list_frameworks
  exit 0
fi
if ((list_profiles_requested)); then
  [[ -z "$app_name" ]] || die '--list-profiles cannot be combined with an app name'
  list_profiles "$framework"
  exit 0
fi

[[ -n "$app_name" ]] || die 'app name is required'
[[ "$app_name" =~ ^[a-z0-9]([a-z0-9-]{0,56}[a-z0-9])?$ ]] ||
  die 'app name must be lowercase DNS-safe text with at most 58 characters'

if [[ -z "$framework" ]]; then
  [[ -n "$profile" ]] || die '--framework and --profile are required (use --list-profiles)'
  valid_slug "$profile" || die 'invalid profile name'
  if load_framework "$profile"; then
    framework=$profile
    profile=$DEFAULT_PROFILE
    legacy_alias=1
  else
    die '--framework is required when selecting a project profile'
  fi
else
  valid_slug "$framework" || die 'invalid framework name'
  load_framework "$framework" || die "unknown framework: $framework (use --list-frameworks)"
  [[ -n "$profile" ]] || die "--profile is required for $framework (default: $DEFAULT_PROFILE)"
fi
valid_slug "$profile" || die 'invalid profile name'
load_project_profile "$framework" "$profile" ||
  die "unknown profile: $framework/$profile (use --list-profiles --framework $framework)"

app_dir="$APPS_DIR/$app_name"
if [[ -e "$app_dir" && $force -eq 0 ]]; then
  die "target already exists: $app_dir (use --force to back it up and replace it)"
fi

printf 'JavaScript app plan\n'
printf '  app: %s (https://%s.test)\n' "$app_name" "$app_name"
printf '  framework: %s — %s\n' "$framework" "$FRAMEWORK_DESCRIPTION"
printf '  profile: %s — %s\n' "$profile" "$PROFILE_DESCRIPTION"
((legacy_alias)) && printf '  compatibility alias: --profile %s -> --framework %s --profile %s\n' "$framework" "$framework" "$profile"
printf '  target: %s\n' "$app_dir"
if [[ -e "$app_dir" ]]; then
  printf '  backup existing target under: %s/apps/.devarch-backups/\n' "$PROJECT_ROOT"
fi
printf '  scaffold: '
printf '%q ' "${SCAFFOLD[@]/__APP__/$app_name}"
printf '\n  normalize package script: devarch=%q\n' "$DEVARCH_SCRIPT"
declare -F configure_app >/dev/null && printf '  apply profile-specific runtime configuration\n'
[[ ${POST_INSTALL:-0} == 1 ]] && printf '  install dependencies with npm\n'
if ((start)); then
  printf '  start isolated Node runtime through scripts/node/bootstrap.sh%s\n' "$([[ $no_hosts -eq 1 ]] && printf ' (without hosts registration)')"
fi

((dry_run)) && exit 0
command -v node >/dev/null 2>&1 || die 'node is required'
command -v npm >/dev/null 2>&1 || die 'npm is required'
mkdir -p "$APPS_DIR"

staging="$(mktemp -d "$APPS_DIR/.devarch-scaffold-${app_name}.XXXXXX")"
cleanup() { rm -rf -- "$staging"; }
trap cleanup EXIT

scaffold_command=("${SCAFFOLD[@]/__APP__/app}")
(
  cd "$staging"
  CI=1 "${scaffold_command[@]}"
)
[[ -f "$staging/app/package.json" ]] || die 'scaffolder did not create app/package.json'

if [[ ${POST_INSTALL:-0} == 1 ]]; then
  (cd "$staging/app" && npm install)
fi

DEVARCH_SCRIPT_VALUE="$DEVARCH_SCRIPT" node - "$staging/app/package.json" <<'NODE'
const fs = require('node:fs');
const path = process.argv[2];
const pkg = JSON.parse(fs.readFileSync(path, 'utf8'));
pkg.scripts ||= {};
pkg.scripts.devarch = process.env.DEVARCH_SCRIPT_VALUE;
fs.writeFileSync(path, `${JSON.stringify(pkg, null, 2)}\n`);
NODE
if declare -F configure_app >/dev/null; then
  configure_app "$staging/app" "$app_name"
fi
rm -rf -- "$staging/app/.git"

if [[ -e "$app_dir" ]]; then
  backup_dir="$APPS_DIR/.devarch-backups/${app_name}-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$(dirname "$backup_dir")"
  mv -- "$app_dir" "$backup_dir"
  printf 'Backed up existing app to %s\n' "$backup_dir"
fi
mv -- "$staging/app" "$app_dir"
printf 'Created %s with %s/%s.\n' "$app_dir" "$framework" "$profile"

if ((start)); then
  runtime_args=("$app_name")
  ((no_hosts)) && runtime_args+=(--no-hosts)
  "$PROJECT_ROOT/scripts/node/bootstrap.sh" "${runtime_args[@]}"
else
  printf 'Start it with: scripts/node/bootstrap.sh %s\n' "$app_name"
fi
