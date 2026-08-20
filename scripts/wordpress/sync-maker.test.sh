#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYNC="$SCRIPT_DIR/sync-maker.sh"
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT
APPS="$ROOT/apps"
SITE="$APPS/consumer"
REMOTES="$ROOT/remotes"
MANIFEST="$ROOT/manifest.json"
mkdir -p "$SITE/wp-content/themes" "$SITE/wp-content/plugins" "$REMOTES"
touch "$SITE/wp-config.php"

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }

packages=(makerstarter makerblocks makermaker)
types=(theme plugin plugin)
health=(style.css makerblocks.php makermaker.php)
old_commits=()
new_commits=()
repos=()

for i in "${!packages[@]}"; do
  package="${packages[$i]}"; repo="$REMOTES/$package.git"; work="$ROOT/$package-work"
  git init -q --bare "$repo"
  git init -q "$work"
  git -C "$work" config user.email tests@devarch.test
  git -C "$work" config user.name 'DevArch Tests'
  printf '# %s\n\n**FRAMEWORK CORE — DO NOT EDIT; update from playground releases.**\n' "$package" > "$work/CORE-BOUNDARY.md"
  printf 'old\n' > "$work/${health[$i]}"
  git -C "$work" add . && git -C "$work" commit -qm old
  git -C "$work" tag -a v1.0.0 -m v1.0.0
  old_commits+=("$(git -C "$work" rev-parse HEAD)")
  printf 'new\n' > "$work/${health[$i]}"
  git -C "$work" add . && git -C "$work" commit -qm new
  git -C "$work" tag -a v1.1.0 -m v1.1.0
  new_commits+=("$(git -C "$work" rev-parse HEAD)")
  git -C "$work" remote add origin "$repo"
  git -C "$work" push -q origin HEAD v1.0.0 v1.1.0
  repos+=("$repo")
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/$package"; else target="$SITE/wp-content/plugins/$package"; fi
  git clone -q --branch v1.0.0 "$repo" "$target"
done

mkdir -p "$SITE/wp-content/themes/consumer-theme" "$SITE/wp-content/plugins/consumer-blocks" "$SITE/wp-content/plugins/consumer-app"
printf 'theme workspace\n' > "$SITE/wp-content/themes/consumer-theme/custom.txt"
printf 'blocks workspace\n' > "$SITE/wp-content/plugins/consumer-blocks/custom.txt"
printf 'app workspace\n' > "$SITE/wp-content/plugins/consumer-app/custom.txt"
for workspace in "$SITE/wp-content/themes/consumer-theme" "$SITE/wp-content/plugins/consumer-blocks" "$SITE/wp-content/plugins/consumer-app"; do
  printf '# PROJECT OWNED — EDIT HERE\n' > "$workspace/README.md"
done
workspace_hash() {
  find "$SITE/wp-content/themes/consumer-theme" "$SITE/wp-content/plugins/consumer-blocks" "$SITE/wp-content/plugins/consumer-app" -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -d' ' -f1
}
before_workspace="$(workspace_hash)"

{
  printf 'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at\n'
  for i in "${!packages[@]}"; do printf '%s\tv1.0.0\t%s\t%s\t2026-01-01T00:00:00Z\n' "${repos[$i]}" "${old_commits[$i]}" "${types[$i]}"; done
} > "$SITE/.devarch-maker.lock"
old_lock_hash="$(sha256sum "$SITE/.devarch-maker.lock" | cut -d' ' -f1)"

write_manifest() {
  local version="$1" ref="$2" commit_set="$3" health_override="${4:-}"
  local commits_name="$commit_set"
  declare -n commits="$commits_name"
  cat > "$MANIFEST" <<EOF
{
  "schemaVersion": 1,
  "channels": {"stable": "$version"},
  "stacks": {
    "$version": {
      "releasedAt": "2026-01-01T00:00:00Z",
      "packages": {
        "makerstarter": {"type":"theme","repository":"${repos[0]}","ref":"$ref","commit":"${commits[0]}","healthFile":"${health_override:-style.css}","install":"none"},
        "makerblocks": {"type":"plugin","repository":"${repos[1]}","ref":"$ref","commit":"${commits[1]}","healthFile":"makerblocks.php","install":"none"},
        "makermaker": {"type":"plugin","repository":"${repos[2]}","ref":"$ref","commit":"${commits[2]}","healthFile":"makermaker.php","install":"none"}
      }
    }
  }
}
EOF
}

write_manifest 1.1.0 v1.1.0 new_commits
dry_output="$(DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --manifest "$MANIFEST" --dry-run)" || fail 'profile-channel dry-run should succeed'
grep -q 'dry run complete; site and lock unchanged' <<< "$dry_output" || fail 'dry-run completion should be explicit'
[[ "$(workspace_hash)" == "$before_workspace" ]] || fail 'dry-run changed workspace bytes'
[[ "$(sha256sum "$SITE/.devarch-maker.lock" | cut -d' ' -f1)" == "$old_lock_hash" ]] || fail 'dry-run changed the lock'
for i in "${!packages[@]}"; do
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/${packages[$i]}"; else target="$SITE/wp-content/plugins/${packages[$i]}"; fi
  [[ "$(git -C "$target" rev-parse HEAD)" == "${old_commits[$i]}" ]] || fail 'dry-run changed a core checkout'
done

write_manifest 1.1.0 v1.1.0 new_commits missing-health.php
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'failed staged health check should stop synchronization'; fi
[[ "$(sha256sum "$SITE/.devarch-maker.lock" | cut -d' ' -f1)" == "$old_lock_hash" ]] || fail 'failed staged health check changed lock'

write_manifest 1.1.0 v1.1.0 new_commits
php -r '$p=$argv[1]; $s=file_get_contents($p); $q=chr(34); $s=preg_replace("/".$q."install".$q.":".$q."none".$q."/", $q."install".$q.":".$q."composer-no-dev".$q, $s, 1); file_put_contents($p, $s);' "$MANIFEST"
fail_bin="$ROOT/fail-bin"; mkdir -p "$fail_bin"
printf '#!/usr/bin/env bash\nexit 1\n' > "$fail_bin/composer"; chmod +x "$fail_bin/composer"
if PATH="$fail_bin:$PATH" DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'failed dependency install should stop synchronization'; fi
[[ "$(sha256sum "$SITE/.devarch-maker.lock" | cut -d' ' -f1)" == "$old_lock_hash" ]] || fail 'failed dependency install changed lock'
write_manifest 1.1.0 v1.1.0 new_commits

if DEVARCH_APPS_DIR="$APPS" DEVARCH_SYNC_FAIL_AFTER_PUBLISH=true bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then
  fail 'injected post-publication failure should fail'
fi
[[ "$(sha256sum "$SITE/.devarch-maker.lock" | cut -d' ' -f1)" == "$old_lock_hash" ]] || fail 'failed update did not restore lock'
for i in "${!packages[@]}"; do
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/${packages[$i]}"; else target="$SITE/wp-content/plugins/${packages[$i]}"; fi
  [[ "$(git -C "$target" rev-parse HEAD)" == "${old_commits[$i]}" ]] || fail 'failed update did not restore core checkout'
done

sync_output="$(DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST")" || fail 'synchronization should succeed'
rollback_id="$(sed -n 's/.*rollback ID: //p' <<< "$sync_output")"
[[ -n "$rollback_id" ]] || fail 'successful update should retain rollback metadata'
[[ "$(workspace_hash)" == "$before_workspace" ]] || fail 'synchronization changed workspace bytes'
for i in "${!packages[@]}"; do
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/${packages[$i]}"; else target="$SITE/wp-content/plugins/${packages[$i]}"; fi
  [[ "$(git -C "$target" rev-parse HEAD)" == "${new_commits[$i]}" ]] || fail 'core commit does not match manifest'
done

git -C "$SITE/wp-content/plugins/makerblocks" status --porcelain >/dev/null
printf 'dirty\n' >> "$SITE/wp-content/plugins/makerblocks/makerblocks.php"
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'dirty core should be refused'; fi
git -C "$SITE/wp-content/plugins/makerblocks" checkout -q -- makerblocks.php

original_remote="$(git -C "$SITE/wp-content/plugins/makermaker" remote get-url origin)"
git -C "$SITE/wp-content/plugins/makermaker" remote set-url origin "$REMOTES/unknown.git"
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'unknown remote should be refused'; fi
git -C "$SITE/wp-content/plugins/makermaker" remote set-url origin "$original_remote"

bad_commits=("${new_commits[@]}")
bad_commits[0]="${old_commits[0]}"
write_manifest 1.1.0 v1.1.0 bad_commits
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'ref/commit mismatch should be refused'; fi

DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --rollback "$rollback_id" >/dev/null || fail 'explicit rollback should succeed'
for i in "${!packages[@]}"; do
  if [[ "${types[$i]}" == theme ]]; then target="$SITE/wp-content/themes/${packages[$i]}"; else target="$SITE/wp-content/plugins/${packages[$i]}"; fi
  [[ "$(git -C "$target" rev-parse HEAD)" == "${old_commits[$i]}" ]] || fail 'explicit rollback did not restore prior core'
done
[[ "$(workspace_hash)" == "$before_workspace" ]] || fail 'rollback changed workspace bytes'

write_manifest 1.1.0 v1.1.0 new_commits
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 9.9.9 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'unknown incompatible stack version should be refused'; fi
sed -i 's/"ref":"v1.1.0"/"ref":"main"/g' "$MANIFEST"
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" consumer --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'main should be refused without explicit local-development opt-in'; fi
[[ "$(workspace_hash)" == "$before_workspace" ]] || fail 'version/ref refusal changed workspace bytes'

mkdir -p "$APPS/playground/wp-content"
if DEVARCH_APPS_DIR="$APPS" bash "$SYNC" playground --profile clean --to 1.1.0 --manifest "$MANIFEST" >/dev/null 2>&1; then fail 'consumer synchronization should refuse the playground release worktree'; fi

printf 'sync-maker tests passed\n'
