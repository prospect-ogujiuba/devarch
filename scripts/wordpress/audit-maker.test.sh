#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AUDIT="$SCRIPT_DIR/audit-maker.sh"
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT
APPS="$ROOT/apps"; SITE="$APPS/audit-site"; REMOTES="$ROOT/remotes"; MANIFEST="$ROOT/manifest.json"
mkdir -p "$SITE/wp-content/themes" "$SITE/wp-content/plugins" "$REMOTES"
touch "$SITE/wp-config.php"
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
packages=(makerstarter makerblocks makermaker); types=(theme plugin plugin); commits=(); repos=()
for i in "${!packages[@]}"; do
  package="${packages[$i]}"; bare="$REMOTES/$package.git"; work="$ROOT/$package-work"
  git init -q --bare "$bare"; git init -q "$work"
  git -C "$work" config user.email tests@devarch.test; git -C "$work" config user.name 'DevArch Tests'
  printf '# boundary\n\n**FRAMEWORK CORE — DO NOT EDIT; update from playground releases.**\n' > "$work/CORE-BOUNDARY.md"
  touch "$work/$package.php"; git -C "$work" add .; git -C "$work" commit -qm initial; git -C "$work" tag v1.0.0
  git -C "$work" remote add origin "$bare"; git -C "$work" push -q origin HEAD v1.0.0
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/$package"; else target="$SITE/wp-content/plugins/$package"; fi
  git clone -q --branch v1.0.0 "$bare" "$target"
  commits+=("$(git -C "$target" rev-parse HEAD)"); repos+=("$bare")
done
for workspace in "$SITE/wp-content/themes/audit-site-theme" "$SITE/wp-content/plugins/audit-site-blocks" "$SITE/wp-content/plugins/audit-site-app"; do
  mkdir -p "$workspace"; printf '# PROJECT OWNED — EDIT HERE\n' > "$workspace/README.md"; touch "$workspace/.devarch-workspace-backup"
done
cat > "$MANIFEST" <<EOF
{"schemaVersion":1,"channels":{"stable":"1.0.0"},"stacks":{"1.0.0":{"releasedAt":"2026-01-01T00:00:00Z","packages":{
"makerstarter":{"type":"theme","repository":"${repos[0]}","ref":"v1.0.0","commit":"${commits[0]}","healthFile":"makerstarter.php","install":"none"},
"makerblocks":{"type":"plugin","repository":"${repos[1]}","ref":"v1.0.0","commit":"${commits[1]}","healthFile":"makerblocks.php","install":"none"},
"makermaker":{"type":"plugin","repository":"${repos[2]}","ref":"v1.0.0","commit":"${commits[2]}","healthFile":"makermaker.php","install":"none"}}}}}
EOF
{
  printf 'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at\n'
  for i in "${!packages[@]}"; do printf '%s\tv1.0.0\t%s\t%s\t2026-01-01T00:00:00Z\n' "${repos[$i]}" "${commits[$i]}" "${types[$i]}"; done
} > "$SITE/.devarch-maker.lock"

DEVARCH_APPS_DIR="$APPS" bash "$AUDIT" audit-site --manifest "$MANIFEST" --sync-ready >/dev/null || fail 'clean locked marked site should pass audit'

printf 'site edit\n' >> "$SITE/wp-content/themes/makerstarter/makerstarter.php"
dirty_output="$(DEVARCH_APPS_DIR="$APPS" bash "$AUDIT" audit-site --manifest "$MANIFEST" --sync-ready 2>&1 || true)"
grep -q 'dirty core blocks synchronization' <<< "$dirty_output" || fail 'dirty core failure should be actionable'
grep -q 'classify for <site>-theme' <<< "$dirty_output" || fail 'dirty theme classification should name child workspace'
git -C "$SITE/wp-content/themes/makerstarter" checkout -q -- makerstarter.php

rm "$SITE/wp-content/plugins/audit-site-blocks/README.md"
marker_output="$(DEVARCH_APPS_DIR="$APPS" bash "$AUDIT" audit-site --manifest "$MANIFEST" --sync-ready 2>&1 || true)"
grep -q 'workspace ownership marker missing' <<< "$marker_output" || fail 'missing workspace marker should fail'
printf '# PROJECT OWNED — EDIT HERE\n' > "$SITE/wp-content/plugins/audit-site-blocks/README.md"

sed -i "s/${commits[2]}/0000000000000000000000000000000000000000/" "$SITE/.devarch-maker.lock"
lock_output="$(DEVARCH_APPS_DIR="$APPS" bash "$AUDIT" audit-site --manifest "$MANIFEST" --sync-ready 2>&1 || true)"
grep -q 'does not match lock' <<< "$lock_output" || fail 'lock mismatch should fail'

printf 'audit-maker tests passed\n'
