#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"; MIGRATE="$SCRIPT_DIR/migrate-maker.sh"
ROOT="$(mktemp -d)"; trap 'rm -rf "$ROOT"' EXIT
APPS="$ROOT/apps"; SITE="$APPS/migrate-site"; REMOTES="$ROOT/remotes"; MANIFEST="$ROOT/manifest.json"; BIN="$ROOT/bin"
mkdir -p "$SITE/wp-content/themes" "$SITE/wp-content/plugins" "$REMOTES" "$BIN"; touch "$SITE/wp-config.php"
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
printf '#!/usr/bin/env bash\nexit 0\n' > "$BIN/podman"; chmod +x "$BIN/podman"
packages=(makerstarter makerblocks makermaker); types=(theme plugin plugin); health=(style.css makerblocks.php makermaker.php); commits=(); repos=()
for i in "${!packages[@]}"; do
  package="${packages[$i]}"; bare="$REMOTES/$package.git"; work="$ROOT/$package-work"
  git init -q --bare "$bare"; git init -q "$work"; git -C "$work" config user.email tests@devarch.test; git -C "$work" config user.name 'DevArch Tests'
  printf '# boundary\n\n**FRAMEWORK CORE — DO NOT EDIT; update from playground releases.**\n' > "$work/CORE-BOUNDARY.md"; touch "$work/${health[$i]}"
  git -C "$work" add .; git -C "$work" commit -qm initial; git -C "$work" tag -a v1.0.0 -m v1.0.0; git -C "$work" remote add origin "$bare"; git -C "$work" push -q origin HEAD v1.0.0
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/$package"; else target="$SITE/wp-content/plugins/$package"; fi
  git clone -q --branch v1.0.0 "$bare" "$target"; commits+=("$(git -C "$target" rev-parse HEAD)"); repos+=("$bare")
done
for workspace in "$SITE/wp-content/themes/migrate-site-theme" "$SITE/wp-content/plugins/migrate-site-blocks" "$SITE/wp-content/plugins/migrate-site-app"; do mkdir -p "$workspace"; printf '# PROJECT OWNED — EDIT HERE\n' > "$workspace/README.md"; done
cat > "$MANIFEST" <<EOF
{"schemaVersion":1,"channels":{"stable":"1.0.0"},"stacks":{"1.0.0":{"releasedAt":"2026-01-01T00:00:00Z","packages":{
"makerstarter":{"type":"theme","repository":"${repos[0]}","ref":"v1.0.0","commit":"${commits[0]}","healthFile":"style.css","install":"none"},
"makerblocks":{"type":"plugin","repository":"${repos[1]}","ref":"v1.0.0","commit":"${commits[1]}","healthFile":"makerblocks.php","install":"none"},
"makermaker":{"type":"plugin","repository":"${repos[2]}","ref":"v1.0.0","commit":"${commits[2]}","healthFile":"makermaker.php","install":"none"}}}}}
EOF
common=(migrate-site --profile clean --to stable --manifest "$MANIFEST")
env_vars=(DEVARCH_APPS_DIR="$APPS" ADMIN_PASSWORD=test GITHUB_USER=example CONTAINER_RUNTIME=podman PATH="$BIN:$PATH")

env "${env_vars[@]}" bash "$MIGRATE" "${common[@]}" --dry-run >/dev/null || fail 'migration dry-run should succeed'
[[ ! -e "$SITE/.devarch-maker.lock" && ! -d "$APPS/.devarch-maker-migrations" ]] || fail 'dry-run changed migration state'

printf 'project change\n' >> "$SITE/wp-content/themes/makerstarter/style.css"
set +e
env "${env_vars[@]}" bash "$MIGRATE" "${common[@]}" >/dev/null 2>&1
status=$?
set -e
[[ "$status" -eq 3 ]] || fail 'dirty core migration should stop after preservation with exit 3'
backup="$(find "$APPS/.devarch-maker-migrations" -mindepth 1 -maxdepth 1 -type d | head -1)"
[[ -f "$backup/core/makerstarter.patch" && -s "$backup/core/makerstarter.patch" ]] || fail 'dirty core patch was not preserved'
[[ ! -e "$SITE/.devarch-maker.lock" ]] || fail 'dirty migration wrote a lock'
for workspace in "$SITE/wp-content/themes/migrate-site-theme" "$SITE/wp-content/plugins/migrate-site-blocks" "$SITE/wp-content/plugins/migrate-site-app"; do [[ -f "$workspace/.devarch-workspace-backup" ]] || fail 'workspace backup receipt missing'; done

git -C "$SITE/wp-content/themes/makerstarter" checkout -q -- style.css
output="$(env "${env_vars[@]}" bash "$MIGRATE" "${common[@]}")" || fail 'clean trusted migration should complete'
grep -q 'migration complete' <<< "$output" || fail 'migration completion should be explicit'
[[ -f "$SITE/.devarch-maker.lock" ]] || fail 'completed migration lock missing'
DEVARCH_APPS_DIR="$APPS" bash "$SCRIPT_DIR/audit-maker.sh" migrate-site --manifest "$MANIFEST" --sync-ready --quiet || fail 'completed migration should pass ownership audit'
for i in "${!packages[@]}"; do
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/${packages[$i]}"; else target="$SITE/wp-content/plugins/${packages[$i]}"; fi
  [[ "$(git -C "$target" rev-parse HEAD)" == "${commits[$i]}" ]] || fail 'completed migration core commit mismatch'
done

LEGACY="$APPS/legacy-site"; mkdir -p "$LEGACY/wp-content/themes" "$LEGACY/wp-content/plugins" "$LEGACY/wp-content/mu-plugins"; touch "$LEGACY/wp-config.php"
git clone -q --branch v1.0.0 "${repos[0]}" "$LEGACY/wp-content/themes/makerstarter"
git clone -q --branch v1.0.0 "${repos[1]}" "$LEGACY/wp-content/plugins/makerblocks"
git clone -q --branch v1.0.0 "${repos[2]}" "$LEGACY/wp-content/mu-plugins/makermaker"
printf '<?php // legacy loader\n' > "$LEGACY/wp-content/mu-plugins/makermaker.php"
for workspace in "$LEGACY/wp-content/themes/legacy-site-theme" "$LEGACY/wp-content/plugins/legacy-site-blocks" "$LEGACY/wp-content/plugins/legacy-site-app"; do mkdir -p "$workspace"; printf '# PROJECT OWNED — EDIT HERE\n' > "$workspace/README.md"; done
legacy_output="$(env "${env_vars[@]}" bash "$MIGRATE" legacy-site --profile clean --to stable --manifest "$MANIFEST")" || fail 'clean legacy MU-plugin migration should complete'
grep -q 'promoted legacy MakerMaker' <<< "$legacy_output" || fail 'legacy layout promotion should be explicit'
[[ -d "$LEGACY/wp-content/plugins/makermaker/.git" && ! -e "$LEGACY/wp-content/mu-plugins/makermaker" && ! -e "$LEGACY/wp-content/mu-plugins/makermaker.php" ]] || fail 'legacy MakerMaker layout was not safely promoted'

printf 'migrate-maker tests passed\n'
