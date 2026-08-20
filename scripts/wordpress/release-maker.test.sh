#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RELEASE="$SCRIPT_DIR/release-maker.sh"
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT
PLAYGROUND="$ROOT/playground"
BIN="$ROOT/bin"
MANIFEST="$ROOT/maker-stack.json"
mkdir -p "$PLAYGROUND/wp-content/themes" "$PLAYGROUND/wp-content/plugins" "$BIN"
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }

for command in npm composer podman; do
  cat > "$BIN/$command" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$BIN/$command"
done

packages=(makerstarter makerblocks makermaker)
paths=("$PLAYGROUND/wp-content/themes/makerstarter" "$PLAYGROUND/wp-content/plugins/makerblocks" "$PLAYGROUND/wp-content/plugins/makermaker")
for i in "${!packages[@]}"; do
  dir="${paths[$i]}"; bare="$ROOT/${packages[$i]}.git"
  git init -q --bare "$bare"
  git init -q "$dir"
  git -C "$dir" config user.email tests@devarch.test
  git -C "$dir" config user.name 'DevArch Tests'
  git -C "$dir" branch -M main
  printf '# boundary\n\n**FRAMEWORK CORE — DO NOT EDIT; update from playground releases.**\n' > "$dir/CORE-BOUNDARY.md"
  case "${packages[$i]}" in makerstarter) touch "$dir/style.css";; makerblocks) touch "$dir/makerblocks.php";; makermaker) touch "$dir/makermaker.php";; esac
  git -C "$dir" add . && git -C "$dir" commit -qm initial
  git -C "$dir" remote add origin "$bare"
done
cat > "$MANIFEST" <<'EOF'
{
  "$schema": "./maker-stack.schema.json",
  "schemaVersion": 1,
  "channels": {},
  "stacks": {}
}
EOF
manifest_before="$(sha256sum "$MANIFEST" | cut -d' ' -f1)"

printf 'dirty\n' >> "${paths[1]}/makerblocks.php"
if PATH="$BIN:$PATH" DEVARCH_PLAYGROUND_DIR="$PLAYGROUND" bash "$RELEASE" 1.0.0 --manifest "$MANIFEST" --dry-run >/dev/null 2>&1; then fail 'dirty release repository should be refused'; fi
git -C "${paths[1]}" checkout -q -- makerblocks.php

dry_output="$(PATH="$BIN:$PATH" DEVARCH_PLAYGROUND_DIR="$PLAYGROUND" bash "$RELEASE" 1.0.0 --manifest "$MANIFEST" --dry-run)" || fail 'release dry-run should pass all gates'
grep -q 'playground integration matrix passed' <<< "$dry_output" || fail 'dry-run should run playground integration'
grep -q 'would publish stack 1.0.0' <<< "$dry_output" || fail 'dry-run should print manifest plan'
[[ "$(sha256sum "$MANIFEST" | cut -d' ' -f1)" == "$manifest_before" ]] || fail 'release dry-run changed manifest'
for dir in "${paths[@]}"; do ! git -C "$dir" rev-parse -q --verify refs/tags/v1.0.0 >/dev/null || fail 'release dry-run created a tag'; done

PATH="$BIN:$PATH" DEVARCH_PLAYGROUND_DIR="$PLAYGROUND" bash "$RELEASE" 1.0.0 --manifest "$MANIFEST" >/dev/null || fail 'release should create tags and manifest entry'
for dir in "${paths[@]}"; do git -C "$dir" rev-parse -q --verify refs/tags/v1.0.0 >/dev/null || fail 'release tag missing'; done
php -r '
  $m=json_decode(file_get_contents($argv[1]), true, 512, JSON_THROW_ON_ERROR);
  if (($m["channels"]["stable"] ?? null) !== "1.0.0") exit(1);
  foreach (["makerstarter","makerblocks","makermaker"] as $slug) {
    $p=$m["stacks"]["1.0.0"]["packages"][$slug] ?? null;
    if (!$p || $p["ref"] !== "v1.0.0" || !preg_match("/^[0-9a-f]{40}$/", $p["commit"])) exit(2);
  }
' "$MANIFEST" || fail 'release manifest entry is invalid'

printf 'release-maker tests passed\n'
