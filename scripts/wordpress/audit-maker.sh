#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="${DEVARCH_APPS_DIR:-$PROJECT_ROOT/apps}"
MANIFEST="$SCRIPT_DIR/maker-stack.json"
SITE_NAME=""
ALL=false
QUIET=false
SYNC_READY=false
RUNTIME_CHECK=false
PROJECT_SLUG=""
FAILURES=0
WARNINGS=0
PACKAGES=(makerstarter makerblocks makermaker)
TYPES=(theme plugin plugin)
WORKSPACE_SUFFIXES=(theme blocks app)

log() { [[ "$QUIET" == true ]] || printf '[maker-audit] %s\n' "$*"; }
fail() { FAILURES=$((FAILURES + 1)); printf '[maker-audit] FAIL: %s\n' "$*" >&2; }
warn() { WARNINGS=$((WARNINGS + 1)); [[ "$QUIET" == true ]] || printf '[maker-audit] WARN: %s\n' "$*"; }
die() { printf '[maker-audit] error: %s\n' "$*" >&2; exit 2; }

usage() {
  cat <<'EOF'
Usage: scripts/wordpress/audit-maker.sh <site>|--all [options]

Report only: inventory Maker core ownership, exact lock state, workspace markers,
and activation readiness. No site, Git, database, or lock data is changed.

Options:
  --manifest FILE       Stack manifest used for trusted repository inventory
  --project-slug SLUG   Override the workspace prefix for one exceptional site
  --sync-ready          Treat missing workspace markers as blocking failures
  --runtime-check       Verify active child theme/plugins through the PHP container
  --quiet               Print failures only
  -h, --help            Show this help
EOF
}

parse_args() {
  [[ $# -gt 0 ]] || { usage; exit 2; }
  [[ "$1" != -h && "$1" != --help ]] || { usage; exit 0; }
  if [[ "$1" == --all ]]; then ALL=true; shift; else SITE_NAME="$1"; shift; fi
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --manifest) [[ $# -ge 2 ]] || die "$1 requires a value"; MANIFEST="$2"; shift 2 ;;
      --project-slug) [[ $# -ge 2 ]] || die "$1 requires a value"; PROJECT_SLUG="$2"; shift 2 ;;
      --sync-ready) SYNC_READY=true; shift ;;
      --runtime-check) RUNTIME_CHECK=true; shift ;;
      --quiet) QUIET=true; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ -f "$MANIFEST" ]] || die "manifest not found: $MANIFEST"
  [[ "$ALL" == false || -z "$PROJECT_SLUG" ]] || die "--project-slug cannot be combined with --all"
}

core_path() {
  local site="$1" package="$2" type="$3"
  if [[ "$type" == theme ]]; then printf '%s/%s/wp-content/themes/%s' "$APPS_DIR" "$site" "$package"; else printf '%s/%s/wp-content/plugins/%s' "$APPS_DIR" "$site" "$package"; fi
}

workspace_path() {
  local site="$1" slug="$2" suffix="$3"
  if [[ "$suffix" == theme ]]; then printf '%s/%s/wp-content/themes/%s-theme' "$APPS_DIR" "$site" "$slug"; else printf '%s/%s/wp-content/plugins/%s-%s' "$APPS_DIR" "$site" "$slug" "$suffix"; fi
}

classify_change() {
  local package="$1" path="$2" destination
  case "$package" in
    makerstarter) destination='<site>-theme';;
    makerblocks) destination='<site>-blocks';;
    makermaker) destination='<site>-app';;
  esac
  case "$path" in
    CORE-BOUNDARY.md|README.md|tests/*|scaffolds/*|templates/create-block/*|app/Generator/*|app/Commands/*)
      printf 'review as reusable core change through playground';;
    *) printf 'classify for %s or reusable core before synchronization' "$destination";;
  esac
}

manifest_has_repository() {
  local repository="$1"
  php -r '
    $m=json_decode(file_get_contents($argv[1]), true, 512, JSON_THROW_ON_ERROR);
    foreach (($m["stacks"] ?? []) as $stack) foreach (($stack["packages"] ?? []) as $p) if (($p["repository"] ?? null) === $argv[2]) exit(0);
    exit(1);
  ' "$MANIFEST" "$repository"
}

runtime_check_site() {
  local site="$1" slug="$2" runtime="${CONTAINER_RUNTIME:-podman}" path="/var/www/html/$site"
  command -v "$runtime" >/dev/null 2>&1 || { fail "$site: runtime unavailable for activation checks: $runtime"; return; }
  "$runtime" exec --user "${WORDPRESS_CONTAINER_USER:-0:0}" -e HOME=/tmp php wp --path="$path" --allow-root theme is-active "$slug-theme" >/dev/null 2>&1 || fail "$site: child theme is not active: $slug-theme"
  "$runtime" exec --user "${WORDPRESS_CONTAINER_USER:-0:0}" -e HOME=/tmp php wp --path="$path" --allow-root plugin is-active "$slug-blocks" >/dev/null 2>&1 || fail "$site: blocks workspace is not active: $slug-blocks"
  "$runtime" exec --user "${WORDPRESS_CONTAINER_USER:-0:0}" -e HOME=/tmp php wp --path="$path" --allow-root plugin is-active "$slug-app" >/dev/null 2>&1 || fail "$site: app workspace is not active: $slug-app"
}

audit_site() {
  local site="$1" slug="${PROJECT_SLUG:-$1}"
  local root="$APPS_DIR/$site" lock="$APPS_DIR/$site/.devarch-maker.lock"
  local i package type target remote head locked_commit workspace suffix status path suggestion
  [[ "$site" =~ ^[a-z0-9][a-z0-9-]{0,59}$ ]] || { fail "$site: invalid site name"; return; }
  [[ "$slug" =~ ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ ]] || { fail "$site: invalid project slug: $slug"; return; }
  [[ -f "$root/wp-config.php" ]] || { fail "$site: WordPress site not found"; return; }
  log "site=$site project=$slug"
  if [[ ! -f "$lock" ]]; then
    fail "$site: Maker lock missing: $lock (bootstrap or migrate before synchronization)"
  elif [[ "$(head -n 1 "$lock")" != $'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at' ]]; then
    fail "$site: unsupported Maker lock format: $lock"
  fi
  for i in "${!PACKAGES[@]}"; do
    package="${PACKAGES[$i]}"; type="${TYPES[$i]}"; suffix="${WORKSPACE_SUFFIXES[$i]}"
    target="$(core_path "$site" "$package" "$type")"
    if [[ ! -d "$target/.git" || -L "$target" ]]; then fail "$site: core target must be a direct Git checkout: $target"; continue; fi
    remote="$(git -C "$target" remote get-url origin 2>/dev/null || true)"
    head="$(git -C "$target" rev-parse HEAD 2>/dev/null || true)"
    [[ -n "$remote" ]] || fail "$site: $package has no origin remote"
    manifest_has_repository "$remote" || fail "$site: $package origin is not trusted by any published stack: $remote"
    status="$(git -C "$target" status --porcelain)"
    if [[ -n "$status" ]]; then
      fail "$site: dirty core blocks synchronization: $target"
      while IFS= read -r line; do
        path="${line:3}"; suggestion="$(classify_change "$package" "$path")"
        [[ "$QUIET" == true ]] || printf '[maker-audit]   %s: %s -> %s\n' "$package" "$path" "$suggestion"
      done <<< "$status"
    fi
    if [[ -f "$lock" ]]; then
      locked_commit="$(awk -F '\t' -v repo="$remote" 'NR>1 && $1==repo {print $3; exit}' "$lock")"
      [[ -n "$locked_commit" ]] || fail "$site: lock has no row for $package origin: $remote"
      [[ -z "$locked_commit" || "$head" == "$locked_commit" ]] || fail "$site: $package HEAD $head does not match lock $locked_commit"
    fi
    workspace="$(workspace_path "$site" "$slug" "$suffix")"
    if [[ ! -d "$workspace" ]]; then
      if [[ "$SYNC_READY" == true ]]; then fail "$site: workspace missing: $workspace"; else warn "$site: workspace missing: $workspace"; fi
    elif [[ ! -f "$workspace/README.md" ]] || ! grep -Fq 'PROJECT OWNED — EDIT HERE' "$workspace/README.md"; then
      fail "$site: workspace ownership marker missing: $workspace/README.md"
    elif [[ ! -d "$workspace/.git" && ! -f "$workspace/.devarch-workspace-backup" ]]; then
      warn "$site: workspace is marked but not independently versioned/backed up: $workspace"
    fi
    log "$package head=$head lock=${locked_commit:-missing} remote=$remote"
  done
  [[ "$RUNTIME_CHECK" == false ]] || runtime_check_site "$site" "$slug"
}

is_maker_site() {
  local site="$1" count=0 i
  [[ -f "$APPS_DIR/$site/wp-config.php" ]] || return 1
  for i in "${!PACKAGES[@]}"; do [[ -d "$(core_path "$site" "${PACKAGES[$i]}" "${TYPES[$i]}")" ]] && count=$((count + 1)); done
  ((count > 0))
}

main() {
  parse_args "$@"
  local site found=false
  if [[ "$ALL" == true ]]; then
    while IFS= read -r site; do
      is_maker_site "$site" || continue
      found=true; audit_site "$site"
    done < <(find "$APPS_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
    [[ "$found" == true ]] || warn "no Maker sites found under $APPS_DIR"
  else
    audit_site "$SITE_NAME"
  fi
  log "summary: failures=$FAILURES warnings=$WARNINGS"
  ((FAILURES == 0))
}

main "$@"
