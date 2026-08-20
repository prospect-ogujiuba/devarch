#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="${DEVARCH_APPS_DIR:-$PROJECT_ROOT/apps}"
MANIFEST="$SCRIPT_DIR/maker-stack.json"
PROFILE=""
TARGET_VERSION=""
SITE_NAME=""
DRY_RUN=false
ALLOW_MAIN=false
ROLLBACK_ID=""
PACKAGES=(makerstarter makerblocks makermaker)
PACKAGE_TYPES=()
REPOSITORIES=()
REFS=()
COMMITS=()
HEALTH_FILES=()
INSTALLS=()
TARGETS=()
STAGES=()
BACKUP_DIR=""
PUBLISH_STARTED=false
LOCK_FILE=""

log() { printf '[maker-sync] %s\n' "$*"; }
die() { printf '[maker-sync] error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: scripts/wordpress/sync-maker.sh <site> --profile NAME [--to VERSION] [options]

Synchronize only MakerStarter, MakerBlocks, and MakerMaker core checkouts to an
exact stack release. Project-owned sibling workspaces are never read or changed.

Options:
  --profile NAME       Maker-enabled WordPress profile used by the site
  --to VERSION         Semantic stack version or channel (default: profile channel)
  --manifest FILE      Release manifest (default: scripts/wordpress/maker-stack.json)
  --allow-main         Permit a manifest package ref of main (local development only)
  --rollback ID        Restore core packages and lock from a retained rollback ID
  --dry-run            Validate, fetch, and print without changing the site
  -h, --help           Show this help
EOF
}

parse_args() {
  [[ $# -gt 0 ]] || { usage; exit 1; }
  [[ "$1" != -h && "$1" != --help ]] || { usage; exit 0; }
  [[ "$1" != -* ]] || die "site is required as the first argument"
  SITE_NAME="$1"; shift
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --profile) [[ $# -ge 2 ]] || die "$1 requires a value"; PROFILE="$2"; shift 2 ;;
      --to) [[ $# -ge 2 ]] || die "$1 requires a value"; TARGET_VERSION="$2"; shift 2 ;;
      --manifest) [[ $# -ge 2 ]] || die "$1 requires a value"; MANIFEST="$2"; shift 2 ;;
      --allow-main) ALLOW_MAIN=true; shift ;;
      --rollback) [[ $# -ge 2 ]] || die "$1 requires a value"; ROLLBACK_ID="$2"; shift 2 ;;
      --dry-run) DRY_RUN=true; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown option: $1" ;;
    esac
  done
}

validate_base() {
  [[ "$SITE_NAME" =~ ^[a-z0-9][a-z0-9-]{0,59}$ ]] || die "invalid site name: $SITE_NAME"
  [[ -d "$APPS_DIR/$SITE_NAME/wp-content" ]] || die "WordPress content directory not found: $APPS_DIR/$SITE_NAME/wp-content"
  LOCK_FILE="$APPS_DIR/$SITE_NAME/.devarch-maker.lock"
  if [[ -n "$ROLLBACK_ID" ]]; then
    [[ "$ROLLBACK_ID" =~ ^[0-9]{8}T[0-9]{6}Z-[0-9]+$ ]] || die "invalid rollback ID: $ROLLBACK_ID"
    return
  fi
  [[ -n "$PROFILE" ]] || die "--profile is required"
  [[ "$PROFILE" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || die "invalid profile name: $PROFILE"
  local profile_file="$SCRIPT_DIR/profiles/$PROFILE.profile"
  [[ -f "$profile_file" ]] || die "unknown profile: $PROFILE"
  local fragment
  fragment="$(awk '$1=="include" && $2=="maker-stack.fragment" {print $2; exit}' "$profile_file")"
  [[ -n "$fragment" ]] || die "profile '$PROFILE' is not Maker-enabled"
  if [[ -z "$TARGET_VERSION" ]]; then
    TARGET_VERSION="$(awk '$1=="maker-stack-channel" {print $2; exit}' "$SCRIPT_DIR/profiles/$fragment")"
    [[ -n "$TARGET_VERSION" ]] || die "--to is required because profile '$PROFILE' has no Maker stack channel"
  fi
  [[ -f "$MANIFEST" ]] || die "manifest not found: $MANIFEST"
  [[ -f "$LOCK_FILE" ]] || die "Maker lock file not found: $LOCK_FILE"
  command -v git >/dev/null 2>&1 || die "git is required"
  command -v php >/dev/null 2>&1 || die "PHP is required to validate the release manifest"
}

load_manifest() {
  local output
  output="$(php -r '
    $file=$argv[1]; $wanted=$argv[2];
    $data=json_decode(file_get_contents($file), true, 512, JSON_THROW_ON_ERROR);
    if (($data["schemaVersion"] ?? null) !== 1 || !is_array($data["channels"] ?? null) || !is_array($data["stacks"] ?? null)) throw new Exception("unsupported manifest schema");
    $version=$data["channels"][$wanted] ?? $wanted;
    if (!preg_match("/^[0-9]+\\.[0-9]+\\.[0-9]+$/", $version) || !isset($data["stacks"][$version])) throw new Exception("unknown stack version or channel: ".$wanted);
    $expected=["makerstarter"=>"theme", "makerblocks"=>"plugin", "makermaker"=>"plugin"];
    echo $version, "\n";
    foreach ($expected as $slug=>$type) {
      $p=$data["stacks"][$version]["packages"][$slug] ?? null;
      if (!is_array($p) || ($p["type"] ?? null) !== $type) throw new Exception("invalid package declaration: ".$slug);
      foreach (["repository","ref","commit","healthFile","install"] as $key) if (!is_string($p[$key] ?? null) || $p[$key] === "" || preg_match("/\\s/", $p[$key])) throw new Exception("invalid ".$key." for ".$slug);
      if (!preg_match("/^(v[0-9]+\\.[0-9]+\\.[0-9]+|main)$/", $p["ref"])) throw new Exception("untrusted ref for ".$slug);
      if (!preg_match("/^[0-9a-f]{40}$/", $p["commit"])) throw new Exception("invalid commit for ".$slug);
      if (!preg_match("#^[A-Za-z0-9._/-]+$#", $p["healthFile"]) || str_contains($p["healthFile"], "..")) throw new Exception("unsafe health file for ".$slug);
      if (!in_array($p["install"], ["none","composer-no-dev"], true)) throw new Exception("unsupported install contract for ".$slug);
      echo implode("\t", [$type,$p["repository"],$p["ref"],$p["commit"],$p["healthFile"],$p["install"]]), "\n";
    }
  ' "$MANIFEST" "$TARGET_VERSION")" || die "release manifest validation failed"
  mapfile -t lines <<< "$output"
  TARGET_VERSION="${lines[0]}"
  local i type repository ref commit health install
  for i in "${!PACKAGES[@]}"; do
    IFS=$'\t' read -r type repository ref commit health install <<< "${lines[$((i+1))]:-}"
    [[ -n "$install" ]] || die "manifest package data is incomplete: ${PACKAGES[$i]}"
    [[ "$ref" != main || "$ALLOW_MAIN" == true ]] || die "manifest uses main for ${PACKAGES[$i]}; pass --allow-main only for local development"
    PACKAGE_TYPES+=("$type"); REPOSITORIES+=("$repository"); REFS+=("$ref"); COMMITS+=("$commit"); HEALTH_FILES+=("$health"); INSTALLS+=("$install")
  done
}

set_targets() {
  local site="$APPS_DIR/$SITE_NAME" i target expected
  for i in "${!PACKAGES[@]}"; do
    if [[ "${PACKAGE_TYPES[$i]}" == theme ]]; then expected="$site/wp-content/themes/${PACKAGES[$i]}"; else expected="$site/wp-content/plugins/${PACKAGES[$i]}"; fi
    target="$expected"
    [[ -d "$target/.git" && ! -L "$target" ]] || die "declared core target is not a direct Git checkout: $target"
    [[ "$(cd "$(dirname "$target")" && pwd -P)/$(basename "$target")" == "$expected" ]] || die "core target escaped its declared directory: $target"
    TARGETS+=("$target")
  done
}

validate_current_state() {
  local i target remote repository
  [[ "$(head -n 1 "$LOCK_FILE")" == $'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at' ]] || die "unsupported Maker lock format: $LOCK_FILE"
  for i in "${!PACKAGES[@]}"; do
    target="${TARGETS[$i]}"; repository="${REPOSITORIES[$i]}"
    [[ -z "$(git -C "$target" status --porcelain)" ]] || die "dirty core repository refused: $target"
    remote="$(git -C "$target" remote get-url origin 2>/dev/null || true)"
    [[ "$remote" == "$repository" ]] || die "unknown core remote for ${PACKAGES[$i]}: $remote"
    awk -F '\t' -v repo="$repository" 'NR>1 && $1==repo {found=1} END {exit !found}' "$LOCK_FILE" || die "lock file does not declare ${PACKAGES[$i]} repository"
  done
}

prepare_stages() {
  local i package stage actual parent
  for i in "${!PACKAGES[@]}"; do
    package="${PACKAGES[$i]}"
    if [[ "$DRY_RUN" == true ]]; then parent="${TMPDIR:-/tmp}"; else parent="$(dirname "${TARGETS[$i]}")"; fi
    stage="$(mktemp -d "$parent/.${package}.devarch-sync.XXXXXX")"
    STAGES+=("$stage")
    log "stage $package: ${REFS[$i]} -> ${COMMITS[$i]}"
    git -C "$stage" init -q
    git -C "$stage" remote add origin "${REPOSITORIES[$i]}"
    git -C "$stage" fetch -q --depth 1 origin "${REFS[$i]}" || die "failed to fetch trusted ref ${REFS[$i]} for $package"
    actual="$(git -C "$stage" rev-parse 'FETCH_HEAD^{commit}')"
    [[ "$actual" == "${COMMITS[$i]}" ]] || die "ref/commit mismatch for $package: ${REFS[$i]} resolved to $actual"
    git -C "$stage" checkout -q --detach "$actual"
    [[ -f "$stage/${HEALTH_FILES[$i]}" ]] || die "staged health check failed for $package: missing ${HEALTH_FILES[$i]}"
    grep -Fq 'FRAMEWORK CORE — DO NOT EDIT; update from playground releases.' "$stage/CORE-BOUNDARY.md" || die "staged core marker check failed for $package"
    case "${INSTALLS[$i]}" in
      none) ;;
      composer-no-dev)
        command -v composer >/dev/null 2>&1 || die "release requires Composer for $package"
        local had_lock=false
        [[ ! -f "$stage/composer.lock" ]] || had_lock=true
        composer install --no-interaction --no-dev --working-dir="$stage" >/dev/null || die "dependency install failed for $package"
        [[ "$had_lock" == true ]] || rm -f "$stage/composer.lock"
        [[ -z "$(git -C "$stage" status --porcelain)" ]] || die "dependency install mutated tracked release files for $package"
        ;;
    esac
  done
}

cleanup_and_maybe_restore() {
  local status=$? i target backup
  if [[ $status -ne 0 && "$PUBLISH_STARTED" == true ]]; then
    log "update failed; restoring prior core versions and lock"
    for i in "${!PACKAGES[@]}"; do
      target="${TARGETS[$i]}"; backup="$BACKUP_DIR/core/${PACKAGES[$i]}"
      if [[ -e "$backup" ]]; then rm -rf "$target"; mv "$backup" "$target"; fi
    done
    if [[ -f "$BACKUP_DIR/lock.before" ]]; then cp "$BACKUP_DIR/lock.before" "$LOCK_FILE"; else rm -f "$LOCK_FILE"; fi
  fi
  for i in "${!STAGES[@]}"; do [[ -e "${STAGES[$i]}" ]] && rm -rf "${STAGES[$i]}"; done
  exit "$status"
}

write_lock() {
  local tmp="$LOCK_FILE.tmp.$$" installed i
  installed="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  {
    printf 'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at\n'
    for i in "${!PACKAGES[@]}"; do
      printf '%s\t%s\t%s\t%s\t%s\n' "${REPOSITORIES[$i]}" "${REFS[$i]}" "${COMMITS[$i]}" "${PACKAGE_TYPES[$i]}" "$installed"
    done
  } > "$tmp"
  mv "$tmp" "$LOCK_FILE"
}

publish_update() {
  local id i target
  id="$(date -u +%Y%m%dT%H%M%SZ)-$$"
  BACKUP_DIR="$APPS_DIR/$SITE_NAME/.devarch-maker-rollbacks/$id"
  mkdir -p "$BACKUP_DIR/core"
  cp "$LOCK_FILE" "$BACKUP_DIR/lock.before"
  printf 'stack_version\t%s\ncreated_at\t%s\n' "$TARGET_VERSION" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$BACKUP_DIR/metadata.tsv"
  PUBLISH_STARTED=true
  for i in "${!PACKAGES[@]}"; do
    target="${TARGETS[$i]}"
    mv "$target" "$BACKUP_DIR/core/${PACKAGES[$i]}"
    mv "${STAGES[$i]}" "$target"
  done
  write_lock
  for i in "${!PACKAGES[@]}"; do
    [[ "$(git -C "${TARGETS[$i]}" rev-parse HEAD)" == "${COMMITS[$i]}" ]]
    [[ -f "${TARGETS[$i]}/${HEALTH_FILES[$i]}" ]]
  done
  [[ "${DEVARCH_SYNC_FAIL_AFTER_PUBLISH:-false}" != true ]]
  PUBLISH_STARTED=false
  log "synchronized Maker stack $TARGET_VERSION; rollback ID: $id"
}

rollback_release() {
  local source="$APPS_DIR/$SITE_NAME/.devarch-maker-rollbacks/$ROLLBACK_ID" i target
  [[ -d "$source/core" && -f "$source/lock.before" ]] || die "rollback not found or incomplete: $ROLLBACK_ID"
  for i in "${!PACKAGES[@]}"; do
    if [[ "${PACKAGES[$i]}" == makerstarter ]]; then target="$APPS_DIR/$SITE_NAME/wp-content/themes/${PACKAGES[$i]}"; else target="$APPS_DIR/$SITE_NAME/wp-content/plugins/${PACKAGES[$i]}"; fi
    [[ -d "$source/core/${PACKAGES[$i]}/.git" && -d "$target/.git" && ! -L "$target" ]] || die "rollback package invalid: ${PACKAGES[$i]}"
    [[ -z "$(git -C "$target" status --porcelain)" ]] || die "dirty core repository refused during rollback: $target"
  done
  log "restore retained rollback: $ROLLBACK_ID"
  [[ "$DRY_RUN" == true ]] && return
  (
    local txn="$APPS_DIR/$SITE_NAME/.devarch-maker-rollback-txn.$$" published=false status stage
    local rollback_targets=() rollback_stages=()
    rollback_cleanup() {
      status=$?
      if [[ $status -ne 0 && "$published" == true ]]; then
        log "rollback failed; restoring the pre-rollback core and lock"
        for i in "${!PACKAGES[@]}"; do
          target="${rollback_targets[$i]}"
          if [[ -e "$txn/core/${PACKAGES[$i]}" ]]; then rm -rf "$target"; mv "$txn/core/${PACKAGES[$i]}" "$target"; fi
        done
        [[ -f "$txn/lock.current" ]] && cp "$txn/lock.current" "$LOCK_FILE"
      fi
      for stage in "${rollback_stages[@]}"; do [[ -e "$stage" ]] && rm -rf "$stage"; done
      [[ $status -eq 0 ]] && rm -rf "$txn"
      exit "$status"
    }
    trap rollback_cleanup EXIT
    mkdir -p "$txn/core"
    cp "$LOCK_FILE" "$txn/lock.current"
    for i in "${!PACKAGES[@]}"; do
      if [[ "${PACKAGES[$i]}" == makerstarter ]]; then target="$APPS_DIR/$SITE_NAME/wp-content/themes/${PACKAGES[$i]}"; else target="$APPS_DIR/$SITE_NAME/wp-content/plugins/${PACKAGES[$i]}"; fi
      stage="$(mktemp -d "$(dirname "$target")/.${PACKAGES[$i]}.devarch-rollback.XXXXXX")"
      cp -a "$source/core/${PACKAGES[$i]}/." "$stage/"
      [[ -f "$stage/CORE-BOUNDARY.md" ]] || return 1
      rollback_targets+=("$target"); rollback_stages+=("$stage")
    done
    published=true
    for i in "${!PACKAGES[@]}"; do
      target="${rollback_targets[$i]}"
      mv "$target" "$txn/core/${PACKAGES[$i]}"
      mv "${rollback_stages[$i]}" "$target"
    done
    cp "$source/lock.before" "$LOCK_FILE.tmp.$$"; mv "$LOCK_FILE.tmp.$$" "$LOCK_FILE"
    for target in "${rollback_targets[@]}"; do [[ -d "$target/.git" && -f "$target/CORE-BOUNDARY.md" ]]; done
    published=false
  )
  log "rollback restored: $ROLLBACK_ID"
}

main() {
  parse_args "$@"
  validate_base
  if [[ -n "$ROLLBACK_ID" ]]; then rollback_release; return; fi
  load_manifest
  set_targets
  "$SCRIPT_DIR/audit-maker.sh" "$SITE_NAME" --manifest "$MANIFEST" --sync-ready --quiet || die "Maker ownership audit failed; resolve reported core/workspace issues before synchronization"
  validate_current_state
  trap cleanup_and_maybe_restore EXIT
  log "site: $SITE_NAME; profile: $PROFILE; target stack: $TARGET_VERSION"
  prepare_stages
  if [[ "$DRY_RUN" == true ]]; then log "dry run complete; site and lock unchanged"; return; fi
  publish_update
}

main "$@"
