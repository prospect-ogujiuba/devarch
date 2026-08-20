#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="${DEVARCH_APPS_DIR:-$PROJECT_ROOT/apps}"
MANIFEST="$SCRIPT_DIR/maker-stack.json"
SITE_NAME=""
PROFILE=""
TARGET_VERSION=""
PROJECT_SLUG=""
DRY_RUN=false
PACKAGES=(makerstarter makerblocks makermaker)
TYPES=(theme plugin plugin)
SUFFIXES=(theme blocks app)
TARGET_REPOSITORIES=()
BACKUP_DIR=""
DIRTY=false
LEGACY_MAKERMAKER=false

log() { printf '[maker-migrate] %s\n' "$*"; }
die() { printf '[maker-migrate] error: %s\n' "$*" >&2; exit 1; }
usage() {
  cat <<'EOF'
Usage: scripts/wordpress/migrate-maker.sh <site> --profile NAME --to VERSION [options]

Guarded existing-site migration. It inventories and backs up every core diff and
untracked file before creating project workspaces. Dirty or untrusted core stops
for human classification; clean trusted core is locked, synchronized, activated,
and audited. Nothing is silently reset, deleted, merged, or overwritten.

Options:
  --profile NAME       Maker-enabled profile used to create workspaces
  --to VERSION         Published semantic stack version or channel
  --manifest FILE      Stack manifest (default: scripts/wordpress/maker-stack.json)
  --project-slug SLUG  Override workspace prefix for an exceptional site
  --dry-run            Inventory and print the migration plan without writes
  -h, --help           Show this help
EOF
}

parse_args() {
  [[ $# -gt 0 ]] || { usage; exit 1; }
  [[ "$1" != -h && "$1" != --help ]] || { usage; exit 0; }
  SITE_NAME="$1"; shift
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --profile) [[ $# -ge 2 ]] || die "$1 requires a value"; PROFILE="$2"; shift 2 ;;
      --to) [[ $# -ge 2 ]] || die "$1 requires a value"; TARGET_VERSION="$2"; shift 2 ;;
      --manifest) [[ $# -ge 2 ]] || die "$1 requires a value"; MANIFEST="$2"; shift 2 ;;
      --project-slug) [[ $# -ge 2 ]] || die "$1 requires a value"; PROJECT_SLUG="$2"; shift 2 ;;
      --dry-run) DRY_RUN=true; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  PROJECT_SLUG="${PROJECT_SLUG:-$SITE_NAME}"
  [[ "$SITE_NAME" =~ ^[a-z0-9][a-z0-9-]{0,59}$ ]] || die "invalid site name: $SITE_NAME"
  [[ "$PROJECT_SLUG" =~ ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ ]] || die "invalid project slug: $PROJECT_SLUG"
  [[ "$PROFILE" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || die "--profile is required"
  [[ -n "$TARGET_VERSION" ]] || die "--to is required"
  [[ -f "$APPS_DIR/$SITE_NAME/wp-config.php" ]] || die "WordPress site not found: $APPS_DIR/$SITE_NAME"
  [[ -f "$MANIFEST" ]] || die "manifest not found: $MANIFEST"
}

core_path() {
  if [[ "$2" == theme ]]; then printf '%s/%s/wp-content/themes/%s' "$APPS_DIR" "$SITE_NAME" "$1"; else printf '%s/%s/wp-content/plugins/%s' "$APPS_DIR" "$SITE_NAME" "$1"; fi
}

workspace_path() {
  if [[ "$2" == theme ]]; then printf '%s/%s/wp-content/themes/%s-theme' "$APPS_DIR" "$SITE_NAME" "$PROJECT_SLUG"; else printf '%s/%s/wp-content/plugins/%s-%s' "$APPS_DIR" "$SITE_NAME" "$PROJECT_SLUG" "$2"; fi
}

load_target_repositories() {
  local output
  output="$(php -r '
    $m=json_decode(file_get_contents($argv[1]), true, 512, JSON_THROW_ON_ERROR); $wanted=$argv[2];
    $v=$m["channels"][$wanted] ?? $wanted; $s=$m["stacks"][$v] ?? null;
    if (!$s) throw new Exception("unknown stack: ".$wanted);
    echo $v, "\n"; foreach (["makerstarter","makerblocks","makermaker"] as $slug) echo $s["packages"][$slug]["repository"], "\n";
  ' "$MANIFEST" "$TARGET_VERSION")" || die "cannot resolve target stack: $TARGET_VERSION"
  mapfile -t lines <<< "$output"; TARGET_VERSION="${lines[0]}"
  TARGET_REPOSITORIES=("${lines[1]}" "${lines[2]}" "${lines[3]}")
}

inventory_core() {
  local i target status remote legacy="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins/makermaker"
  for i in "${!PACKAGES[@]}"; do
    target="$(core_path "${PACKAGES[$i]}" "${TYPES[$i]}")"
    if [[ "${PACKAGES[$i]}" == makermaker && ! -e "$target" && -d "$legacy/.git" && ! -L "$legacy" ]]; then
      target="$legacy"; LEGACY_MAKERMAKER=true; log "LEGACY makermaker: clean MU-plugin checkout will be promoted to the regular plugin directory after workspace backup"
    elif [[ ! -d "$target/.git" || -L "$target" ]]; then
      log "BLOCK ${PACKAGES[$i]}: not a direct Git checkout: $target"; DIRTY=true; continue
    fi
    remote="$(git -C "$target" remote get-url origin 2>/dev/null || true)"
    [[ "$remote" == "${TARGET_REPOSITORIES[$i]}" ]] || { log "BLOCK ${PACKAGES[$i]}: origin $remote does not match released origin ${TARGET_REPOSITORIES[$i]}"; DIRTY=true; }
    status="$(git -C "$target" status --porcelain)"
    if [[ -n "$status" ]]; then log "CLASSIFY ${PACKAGES[$i]}: modified/untracked core files require ownership decisions"; DIRTY=true; fi
    log "${PACKAGES[$i]} head=$(git -C "$target" rev-parse HEAD 2>/dev/null || printf missing) remote=${remote:-missing}"
  done
}

backup_inventory() {
  local id i target untracked legacy="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins/makermaker"
  id="$(date -u +%Y%m%dT%H%M%SZ)-$$"
  BACKUP_DIR="$APPS_DIR/.devarch-maker-migrations/$SITE_NAME-$id"
  if [[ "$DRY_RUN" == true ]]; then log "would write migration inventory and core backups: $BACKUP_DIR"; return; fi
  mkdir -p "$BACKUP_DIR/core"
  {
    printf 'site\t%s\nproject_slug\t%s\ntarget_stack\t%s\ncreated_at\t%s\n' "$SITE_NAME" "$PROJECT_SLUG" "$TARGET_VERSION" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  } > "$BACKUP_DIR/metadata.tsv"
  for i in "${!PACKAGES[@]}"; do
    target="$(core_path "${PACKAGES[$i]}" "${TYPES[$i]}")"
    if [[ "${PACKAGES[$i]}" == makermaker && ! -e "$target" && -d "$legacy/.git" ]]; then target="$legacy"; fi
    [[ -d "$target/.git" && ! -L "$target" ]] || continue
    git -C "$target" status --short > "$BACKUP_DIR/core/${PACKAGES[$i]}.status"
    git -C "$target" diff --binary HEAD > "$BACKUP_DIR/core/${PACKAGES[$i]}.patch"
    git -C "$target" remote -v > "$BACKUP_DIR/core/${PACKAGES[$i]}.remotes"
    git -C "$target" rev-parse HEAD > "$BACKUP_DIR/core/${PACKAGES[$i]}.head"
    untracked="$BACKUP_DIR/core/${PACKAGES[$i]}.untracked"
    git -C "$target" ls-files --others --exclude-standard -z > "$untracked.list"
    if [[ -s "$untracked.list" ]]; then tar -C "$target" --null -T "$untracked.list" -czf "$untracked.tgz"; fi
    rm -f "$untracked.list"
  done
  DEVARCH_APPS_DIR="$APPS_DIR" bash "$SCRIPT_DIR/audit-maker.sh" "$SITE_NAME" --manifest "$MANIFEST" > "$BACKUP_DIR/audit.txt" 2>&1 || true
  log "migration evidence preserved: $BACKUP_DIR"
}

provision_workspaces() {
  local output workspace i
  if [[ "$DRY_RUN" == true ]]; then
    log "would provision refusal-safe project workspaces through bootstrap --scaffolds-only"
    return
  fi
  if ! output="$(DEVARCH_APPS_DIR="$APPS_DIR" bash "$SCRIPT_DIR/bootstrap.sh" "$SITE_NAME" --profile "$PROFILE" --scaffolds-only --project-slug "$PROJECT_SLUG" 2>&1)"; then
    printf '%s\n' "$output" >&2
    [[ "$DIRTY" == true ]] || die "workspace provisioning failed; core backup remains at $BACKUP_DIR"
    log "workspace provisioning is partial; classification remains blocked and preserved core was not replaced"
  fi
  for i in "${!SUFFIXES[@]}"; do
    workspace="$(workspace_path "$SITE_NAME" "${SUFFIXES[$i]}")"
    if [[ -d "$workspace" ]]; then
      printf 'backup\t%s\ncreated_at\t%s\n' "$BACKUP_DIR" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$workspace/.devarch-workspace-backup"
    elif [[ "$DIRTY" == true ]]; then
      log "workspace still requires classified migration: $workspace"
    else
      die "expected workspace was not provisioned: $workspace"
    fi
  done
}

promote_legacy_makermaker() {
  [[ "$LEGACY_MAKERMAKER" == true ]] || return 0
  local legacy="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins/makermaker"
  local loader="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins/makermaker.php"
  local target="$APPS_DIR/$SITE_NAME/wp-content/plugins/makermaker"
  [[ -d "$legacy/.git" && ! -e "$target" ]] || die "legacy MakerMaker promotion precondition failed"
  [[ -f "$loader" ]] && cp "$loader" "$BACKUP_DIR/core/legacy-makermaker-loader.php"
  mv "$legacy" "$target"
  [[ ! -f "$loader" ]] || mv "$loader" "$BACKUP_DIR/core/legacy-makermaker-loader.active.php"
  if ! "${CONTAINER_RUNTIME:-podman}" exec --user "${WORDPRESS_CONTAINER_USER:-0:0}" -e HOME=/tmp php \
    wp --path="/var/www/html/$SITE_NAME" --allow-root plugin activate makermaker >/dev/null; then
    mv "$target" "$legacy"
    [[ ! -f "$BACKUP_DIR/core/legacy-makermaker-loader.active.php" ]] || mv "$BACKUP_DIR/core/legacy-makermaker-loader.active.php" "$loader"
    die "regular MakerMaker activation failed; legacy layout restored"
  fi
  log "promoted legacy MakerMaker MU checkout to regular plugin layout"
}

write_current_lock() {
  local lock="$APPS_DIR/$SITE_NAME/.devarch-maker.lock"
  local tmp="$lock.tmp.$$" i target remote commit ref installed
  installed="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  {
    printf 'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at\n'
    for i in "${!PACKAGES[@]}"; do
      target="$(core_path "${PACKAGES[$i]}" "${TYPES[$i]}")"; remote="$(git -C "$target" remote get-url origin)"; commit="$(git -C "$target" rev-parse HEAD)"
      ref="$(git -C "$target" describe --tags --exact-match 2>/dev/null || printf '%s' "$commit")"
      printf '%s\t%s\t%s\t%s\t%s\n' "$remote" "$ref" "$commit" "${TYPES[$i]}" "$installed"
    done
  } > "$tmp"
  mv "$tmp" "$lock"
}

main() {
  parse_args "$@"; load_target_repositories
  log "site=$SITE_NAME project=$PROJECT_SLUG target=$TARGET_VERSION"
  inventory_core; backup_inventory
  if [[ "$DRY_RUN" == true ]]; then log "dry run complete; no site, lock, activation, or backup changes made"; return; fi
  provision_workspaces
  if [[ "$DIRTY" == true ]]; then
    log "STOP: classify and move/back up reported core changes before rerunning; no core checkout or lock was replaced"
    exit 3
  fi
  promote_legacy_makermaker
  write_current_lock
  DEVARCH_APPS_DIR="$APPS_DIR" bash "$SCRIPT_DIR/sync-maker.sh" "$SITE_NAME" --profile "$PROFILE" --to "$TARGET_VERSION" --manifest "$MANIFEST"
  DEVARCH_APPS_DIR="$APPS_DIR" bash "$SCRIPT_DIR/audit-maker.sh" "$SITE_NAME" --manifest "$MANIFEST" --project-slug "$PROJECT_SLUG" --sync-ready --runtime-check
  log "migration complete: $SITE_NAME -> Maker stack $TARGET_VERSION"
}

main "$@"
