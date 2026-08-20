#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLAYGROUND="${DEVARCH_PLAYGROUND_DIR:-$PROJECT_ROOT/apps/playground}"
MANIFEST="$SCRIPT_DIR/maker-stack.json"
VERSION=""
DRY_RUN=false
PACKAGES=(makerstarter makerblocks makermaker)
DIRS=(
  "$PLAYGROUND/wp-content/themes/makerstarter"
  "$PLAYGROUND/wp-content/plugins/makerblocks"
  "$PLAYGROUND/wp-content/plugins/makermaker"
)
TYPES=(theme plugin plugin)
HEALTH=(style.css makerblocks.php makermaker.php)
INSTALLS=(none none composer-no-dev)

log() { printf '[maker-release] %s\n' "$*"; }
die() { printf '[maker-release] error: %s\n' "$*" >&2; exit 1; }
usage() {
  cat <<'EOF'
Usage: scripts/wordpress/release-maker.sh VERSION [options]

Test the three playground Maker repositories independently, run the playground
integration matrix, create matching local semantic tags, and atomically add a
stack release to maker-stack.json. This command never pushes. Review and commit
the manifest, then push the three tags and DevArch manifest commit together.

Options:
  --manifest FILE   Manifest to update (default: scripts/wordpress/maker-stack.json)
  --dry-run         Run all release gates and print the tag/manifest plan only
  -h, --help        Show this help
EOF
}

parse_args() {
  [[ $# -gt 0 ]] || { usage; exit 1; }
  [[ "$1" != -h && "$1" != --help ]] || { usage; exit 0; }
  VERSION="$1"; shift
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --manifest) [[ $# -ge 2 ]] || die "$1 requires a value"; MANIFEST="$2"; shift 2 ;;
      --dry-run) DRY_RUN=true; shift ;;
      -h|--help) usage; exit 0 ;;
      *) die "unknown option: $1" ;;
    esac
  done
  [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "VERSION must be semantic X.Y.Z"
  [[ -f "$MANIFEST" ]] || die "manifest not found: $MANIFEST"
}

validate_repositories() {
  local i dir remote
  for i in "${!PACKAGES[@]}"; do
    dir="${DIRS[$i]}"
    [[ -d "$dir/.git" ]] || die "playground repository missing: $dir"
    [[ "$(git -C "$dir" branch --show-current)" == main ]] || die "playground release repository must be on main: $dir"
    [[ -z "$(git -C "$dir" status --porcelain)" ]] || die "dirty playground repository refused: $dir"
    remote="$(git -C "$dir" remote get-url origin 2>/dev/null || true)"
    [[ -n "$remote" && ! "$remote" =~ [[:space:]] ]] || die "invalid origin for ${PACKAGES[$i]}"
    ! git -C "$dir" rev-parse -q --verify "refs/tags/v$VERSION" >/dev/null || die "tag already exists for ${PACKAGES[$i]}: v$VERSION"
    grep -Fq 'FRAMEWORK CORE — DO NOT EDIT; update from playground releases.' "$dir/CORE-BOUNDARY.md" || die "core marker missing for ${PACKAGES[$i]}"
  done
}

run_package_tests() {
  log "test MakerStarter independently"
  (cd "${DIRS[0]}" && npm test)
  log "test MakerBlocks independently"
  (cd "${DIRS[1]}" && npm test)
  log "test MakerMaker independently"
  (cd "${DIRS[2]}" && composer test)
}

run_playground_matrix() {
  local runtime="${CONTAINER_RUNTIME:-}" user="${WORDPRESS_CONTAINER_USER:-0:0}" wp_path=/var/www/html/playground
  if [[ -z "$runtime" ]]; then
    if command -v podman >/dev/null 2>&1; then runtime=podman; elif command -v docker >/dev/null 2>&1; then runtime=docker; else die "Podman or Docker is required for the playground integration matrix"; fi
  fi
  "$runtime" exec --user "$user" -e HOME=/tmp php wp --path="$wp_path" --allow-root core is-installed >/dev/null || die "playground WordPress health check failed"
  "$runtime" exec --user "$user" -e HOME=/tmp php wp --path="$wp_path" --allow-root plugin is-active makerblocks >/dev/null || die "MakerBlocks is not active in playground"
  "$runtime" exec --user "$user" -e HOME=/tmp php wp --path="$wp_path" --allow-root plugin is-active makermaker >/dev/null || die "MakerMaker is not active in playground"
  "$runtime" exec --user "$user" -e HOME=/tmp php wp --path="$wp_path" --allow-root theme is-installed makerstarter >/dev/null || die "MakerStarter is not installed in playground"
  "$runtime" exec --user "$user" -e HOME=/tmp php wp --path="$wp_path" --allow-root makermaker register-galaxy >/dev/null || die "MakerMaker Galaxy integration failed"
  log "playground integration matrix passed"
}

update_manifest() {
  local tag="v$VERSION" released tmp i
  local repositories=() commits=()
  released="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  for i in "${!PACKAGES[@]}"; do
    repositories+=("$(git -C "${DIRS[$i]}" remote get-url origin)")
    commits+=("$(git -C "${DIRS[$i]}" rev-parse HEAD)")
  done
  if [[ "$DRY_RUN" == true ]]; then
    for i in "${!PACKAGES[@]}"; do log "would tag ${PACKAGES[$i]} $tag at ${commits[$i]}"; done
    log "would publish stack $VERSION to $MANIFEST"
    return
  fi
  tmp="$MANIFEST.tmp.$$"
  php -r '
    $file=$argv[1]; $out=$argv[2]; $version=$argv[3]; $released=$argv[4];
    $data=json_decode(file_get_contents($file), true, 512, JSON_THROW_ON_ERROR);
    if (($data["schemaVersion"] ?? null) !== 1 || isset($data["stacks"][$version])) throw new Exception("manifest version exists or schema is invalid");
    $slugs=["makerstarter","makerblocks","makermaker"];
    $types=["theme","plugin","plugin"];
    $health=["style.css","makerblocks.php","makermaker.php"];
    $installs=["none","none","composer-no-dev"];
    $packages=[];
    foreach ($slugs as $i=>$slug) $packages[$slug]=["type"=>$types[$i],"repository"=>$argv[5+$i*2],"ref"=>"v".$version,"commit"=>$argv[6+$i*2],"healthFile"=>$health[$i],"install"=>$installs[$i]];
    $data["stacks"][$version]=["releasedAt"=>$released,"packages"=>$packages];
    $data["channels"]["stable"]=$version;
    file_put_contents($out, json_encode($data, JSON_PRETTY_PRINT|JSON_UNESCAPED_SLASHES)."\n");
  ' "$MANIFEST" "$tmp" "$VERSION" "$released" \
    "${repositories[0]}" "${commits[0]}" "${repositories[1]}" "${commits[1]}" "${repositories[2]}" "${commits[2]}"
  local created=() dir
  for i in "${!PACKAGES[@]}"; do
    dir="${DIRS[$i]}"
    if git -C "$dir" tag -a "$tag" -m "release $tag"; then
      created+=("$dir")
    else
      for dir in "${created[@]}"; do git -C "$dir" tag -d "$tag" >/dev/null 2>&1 || true; done
      rm -f "$tmp"
      die "failed to create all package tags; created tags were removed"
    fi
  done
  if ! mv "$tmp" "$MANIFEST"; then
    for dir in "${created[@]}"; do git -C "$dir" tag -d "$tag" >/dev/null 2>&1 || true; done
    die "failed to publish manifest; created tags were removed"
  fi
  log "created local tags and published stack $VERSION to $MANIFEST"
  log "review and commit the manifest, then push all three tags and the DevArch commit"
}

main() {
  parse_args "$@"
  validate_repositories
  run_package_tests
  run_playground_matrix
  update_manifest
}

main "$@"
