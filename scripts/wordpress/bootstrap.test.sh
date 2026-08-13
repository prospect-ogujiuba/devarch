#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

help_output="$(bash "$BOOTSTRAP" --help)" || fail "--help should succeed"
grep -q 'Usage:' <<<"$help_output" || fail "help should include usage"
grep -q -- '--plugin' <<<"$help_output" || fail "help should document plugins"
grep -q -- '--dry-run' <<<"$help_output" || fail "help should document dry-run"
grep -q -- '--profile' <<<"$help_output" || fail "help should document profiles"
grep -q -- '--restore' <<<"$help_output" || fail "help should document native AIOWM restore"

profiles_output="$(bash "$BOOTSTRAP" --list-profiles)" || fail "--list-profiles should succeed"
for profile in bare clean custom loaded; do
  grep -q "^$profile" <<<"$profiles_output" || fail "profile list should include $profile"
done
if grep -q '^starred' <<<"$profiles_output"; then
  fail "empty starred profile should be removed"
fi

if bash "$BOOTSTRAP" 'unsafe/site' --dry-run >/dev/null 2>&1; then
  fail "unsafe site names should be rejected"
fi
if bash "$BOOTSTRAP" restore-site --restore missing.wpress --dry-run >/dev/null 2>&1; then
  fail "missing restore archives should be rejected"
fi
if bash "$BOOTSTRAP" restore-site --restore bootstrap.test.sh --dry-run >/dev/null 2>&1; then
  fail "restore archives must use the .wpress extension"
fi

restore_archive="$(mktemp --suffix=.wpress)"
nested_site="$SCRIPT_DIR/../../apps/nested-restore-test"
cleanup() {
  rm -f "$restore_archive"
  rm -rf "$nested_site"
}
trap cleanup EXIT
mkdir -p "$nested_site/wp-content/plugins/example"
touch "$nested_site/wp-config.php"

restore_output="$(
  cd "$nested_site/wp-content/plugins/example"
  bash "$BOOTSTRAP" --restore "$restore_archive" --dry-run
)" || fail "nested WordPress restore dry-run should succeed"
grep -q 'site: nested-restore-test' <<<"$restore_output" || fail "site name should be discovered from a nested WordPress directory"
grep -q 'create native AIOWM safety backup' <<<"$restore_output" || fail "existing restore targets should receive a native AIOWM safety backup"
grep -q 'ai1wm backup' <<<"$restore_output" || fail "existing safety backup should use native AIOWM WP-CLI"
grep -q 'git clone --depth 1.*all-in-one-wp-migration' <<<"$restore_output" || fail "restore should install the established native-CLI AIOWM repository"
grep -q 'plugin activate all-in-one-wp-migration' <<<"$restore_output" || fail "restore should activate AIOWM"
grep -q 'wp-content/ai1wm-backups' <<<"$restore_output" || fail "restore should prepare the AIOWM backups directory"
grep -q 'all-in-one-wp-migration/storage' <<<"$restore_output" || fail "restore should prepare AIOWM storage"
grep -q 'chmod -R a+rwX' <<<"$restore_output" || fail "AIOWM working directories should be writable"
grep -q 'ai1wm restore' <<<"$restore_output" || fail "restore should use native AIOWM WP-CLI"

dry_run_output="$(
  ADMIN_PASSWORD='not-printed-secret' \
  MARIADB_ROOT_PASSWORD='not-printed-db-secret' \
  bash "$BOOTSTRAP" demo-site \
    --title 'Demo Site' \
    --plugin wp:query-monitor \
    --plugin git:https://github.com/example/example-plugin.git \
    --dry-run
)" || fail "representative dry-run should succeed"

grep -q 'query-monitor' <<<"$dry_run_output" || fail "dry-run should plan WordPress.org plugin installation"
grep -q 'example-plugin' <<<"$dry_run_output" || fail "dry-run should plan Git plugin installation"
grep -q 'start PHP and MariaDB services' <<<"$dry_run_output" || fail "dry-run should plan service startup"
grep -q 'URL: https://demo-site.test' <<<"$dry_run_output" || fail "default URL should use wildcard .test routing"
grep -q 'existing .test reverse proxy' <<<"$dry_run_output" || fail "completion should identify infrastructure routing"
grep -q 'config set FS_METHOD direct' <<<"$dry_run_output" || fail "bootstrap should enable direct filesystem changes without FTP credentials"
grep -q 'post delete 1 --force' <<<"$dry_run_output" || fail "bootstrap should delete the default WordPress post"
grep -q 'chmod -R a+rwX.*/wp-content' <<<"$dry_run_output" || fail "bootstrap should keep local wp-content writable by PHP"
if grep -Eq 'not-printed-secret|not-printed-db-secret' <<<"$dry_run_output"; then
  fail "dry-run must not print secrets"
fi
if grep -Eq 'wp server|WORDPRESS_PORT' <<<"$dry_run_output"; then
  fail "bootstrap should not bypass the existing reverse proxy"
fi

grep -q -- '--user 0:0' <<<"$dry_run_output" || fail "rootless Podman should use the bind-mount owner mapping"
if bash "$BOOTSTRAP" demo-port --port 18080 --dry-run >/dev/null 2>&1; then
  fail "obsolete standalone WP server port should be rejected"
fi
if bash "$BOOTSTRAP" unsafe_name --dry-run >/dev/null 2>&1; then
  fail "site names incompatible with .test host routing should be rejected"
fi

docker_output="$(CONTAINER_RUNTIME=docker bash "$BOOTSTRAP" demo-docker --dry-run)" || fail "Docker dry-run should succeed"
grep -q -- "--user $(id -u):$(id -g)" <<<"$docker_output" || fail "Docker should use the host UID/GID"

profile_output="$(GITHUB_USER=example bash "$BOOTSTRAP" profile-site --profile clean --dry-run)" || fail "clean profile dry-run should succeed"
for repo in all-in-one-wp-migration admin-site-enhancements-pro typerocket-pro-v6 makermaker makerblocks makerstarter; do
  grep -q "$repo" <<<"$profile_output" || fail "clean profile should include $repo"
done
grep -q 'clone Git must-use plugin: typerocket-pro-v6' <<<"$profile_output" || fail "clean profile should install TypeRocket as an MU plugin"
grep -q 'write portable MakerMaker-owned site Galaxy launcher and resolver' <<<"$profile_output" || fail "site Galaxy launcher should be MakerMaker-owned"
grep -q 'GalaxyContext::siteLauncher.*siteConfig' <<<"$profile_output" || fail "site Galaxy launcher/config should use shared portable sources"
grep -q 'clone Git plugin: makermaker' <<<"$profile_output" || fail "clean profile should install MakerMaker as a plugin"
grep -q 'register MakerMaker Galaxy command idempotently using runtime TypeRocket path' <<<"$profile_output" || fail "clean profile should register MakerMaker with runtime TypeRocket discovery"
grep -q 'makermaker register-galaxy' <<<"$profile_output" || fail "Galaxy registration should use MakerMaker's repeatable registrar"
if grep -q 'register-galaxy.*typerocket-path=' <<<"$profile_output"; then fail "bootstrap should not embed a TypeRocket registration path"; fi
grep -q 'makermaker register-plugin-galaxy.*plugin=makermaker.*namespace=Maker/MakerMaker' <<<"$profile_output" || fail "clean profile should backfill MakerMaker's plugin-specific Galaxy context"
grep -q 'clone Git plugin: makerblocks' <<<"$profile_output" || fail "clean profile should install MakerBlocks as a plugin"
grep -q 'clone Git theme: makerstarter' <<<"$profile_output" || fail "clean profile should install MakerStarter as a theme"
grep -q 'theme delete --all' <<<"$profile_output" || fail "custom-theme profiles should delete bundled inactive themes"

bare_output="$(GITHUB_USER=example bash "$BOOTSTRAP" bare-site --profile bare --dry-run)" || fail "bare profile dry-run should succeed"
grep -q 'post delete 1 --force' <<<"$bare_output" || fail "bare profile should delete the default WordPress post"
if grep -q 'theme delete --all' <<<"$bare_output"; then
  fail "profiles without a custom theme should retain bundled themes"
fi

loaded_output="$(GITHUB_USER=example bash "$BOOTSTRAP" loaded-site --profile loaded --dry-run)" || fail "loaded profile dry-run should succeed"
grep -q 'query-monitor' <<<"$loaded_output" || fail "loaded profile should include WordPress.org development plugins"

if bash "$BOOTSTRAP" starred-site --profile starred --dry-run >/dev/null 2>&1; then
  fail "removed starred profile should be rejected"
fi

if bash "$BOOTSTRAP" demo --profile does-not-exist --dry-run >/dev/null 2>&1; then
  fail "unknown profiles should be rejected"
fi

if bash "$BOOTSTRAP" demo --plugin 'git:https://token@github.com/example/private.git' --dry-run >/dev/null 2>&1; then
  fail "credential-bearing Git URLs should be rejected"
fi

profile_entries() {
  grep -Ev '^[[:space:]]*(#|$)' "$SCRIPT_DIR/profiles/$1.profile"
}

[[ "$(profile_entries bare)" == 'github-plugin all-in-one-wp-migration inactive' ]] || fail "bare profile drifted from history"
expected_custom_repos=$'github-mu-plugin typerocket-pro-v6\ngithub-plugin makermaker\ngithub-plugin makerblocks\ngithub-theme makerstarter\ngithub-plugin all-in-one-wp-migration inactive\ngithub-plugin admin-site-enhancements-pro'
[[ "$(profile_entries clean)" == "$expected_custom_repos" ]] || fail "clean profile should contain TypeRocket and all three custom plugins/themes"
[[ "$(profile_entries custom)" == "$expected_custom_repos"$'\ngithub-plugin manual-image-crop' ]] || fail "custom profile should contain TypeRocket and all three custom plugins/themes"
for profile in clean custom loaded; do
  entries="$(profile_entries "$profile")"
  grep -q '^github-mu-plugin typerocket-pro-v6$' <<<"$entries" || fail "$profile profile should contain TypeRocket Pro v6"
  grep -q '^github-plugin makermaker$' <<<"$entries" || fail "$profile profile should contain MakerMaker as a regular plugin"
  grep -q '^github-plugin makerblocks$' <<<"$entries" || fail "$profile profile should contain MakerBlocks"
  grep -q '^github-theme makerstarter$' <<<"$entries" || fail "$profile profile should contain MakerStarter"
done
[[ "$(profile_entries loaded | grep -c '^wp-plugin ')" -eq 12 ]] || fail "loaded profile should retain all 12 development plugins"

cleanup
trap - EXIT
printf 'bootstrap tests passed\n'
