#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

help_output="$(bash "$BOOTSTRAP" --help)" || fail "--help should succeed"
for option in --plugin --dry-run --no-hosts --profile --restore; do
  grep -q -- "$option" <<<"$help_output" || fail "help should document $option"
done

profiles_output="$(bash "$BOOTSTRAP" --list-profiles)" || fail "--list-profiles should succeed"
for profile in bare clean custom loaded; do
  grep -q "^$profile" <<<"$profiles_output" || fail "profile list should include $profile"
done

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
generic_profile_root="$(mktemp -d)"
typerocket_install_root="$(mktemp -d)"
nested_site="$SCRIPT_DIR/../../apps/nested-restore-test"
cleanup() {
  rm -f "$restore_archive"
  rm -rf "$generic_profile_root" "$typerocket_install_root" "$nested_site"
}
trap cleanup EXIT
mkdir -p "$nested_site/wp-content/plugins/example"
touch "$nested_site/wp-config.php"

restore_output="$(
  cd "$nested_site/wp-content/plugins/example"
  bash "$BOOTSTRAP" --restore "$restore_archive" --dry-run
)" || fail "nested WordPress restore dry-run should succeed"
grep -q 'site: nested-restore-test' <<<"$restore_output" || fail "site name should be discovered from a nested WordPress directory"
grep -q 'create native AIOWM safety backup' <<<"$restore_output" || fail "existing restore targets should receive a safety backup"
grep -q 'ai1wm backup' <<<"$restore_output" || fail "safety backup should use native AIOWM WP-CLI"
grep -q 'ai1wm restore' <<<"$restore_output" || fail "restore should use native AIOWM WP-CLI"
grep -q 'wp-content/ai1wm-backups' <<<"$restore_output" || fail "restore should prepare the backup directory"

dry_run_output="$(
  ADMIN_PASSWORD='not-printed-secret' \
  MARIADB_ROOT_PASSWORD='not-printed-db-secret' \
  bash "$BOOTSTRAP" demo-site \
    --title 'Demo Site' \
    --plugin wp:query-monitor \
    --plugin git:https://github.com/example/example-plugin.git \
    --dry-run
)" || fail "representative dry-run should succeed"

for expected in \
  query-monitor example-plugin \
  'start PHP, MariaDB, and Nginx Proxy Manager services' \
  'wait for PHP/WP-CLI, MariaDB, and Nginx Proxy Manager readiness' \
  'URL: https://demo-site.test' \
  'register local host: 127.0.0.1 demo-site.test' \
  'config set FS_METHOD direct' \
  'option update uploads_use_yearmonth_folders 0' \
  'rewrite structure /%postname%/ --hard' \
  'post delete 1 --force' \
  'chmod -R a+rwX'; do
  grep -q "$expected" <<<"$dry_run_output" || fail "dry-run should include: $expected"
done
if grep -Eq 'not-printed-secret|not-printed-db-secret|wp server|WORDPRESS_PORT' <<<"$dry_run_output"; then
  fail "dry-run leaked a secret or used the obsolete standalone server"
fi
grep -q -- '--user 0:0' <<<"$dry_run_output" || fail "Podman should use the bind-mount owner mapping"

no_hosts_output="$(bash "$BOOTSTRAP" no-hosts-site --no-hosts --dry-run)" || fail "hosts opt-out should succeed"
grep -q 'hosts registration skipped: no-hosts-site.test' <<<"$no_hosts_output" || fail "hosts opt-out should be visible"

docker_output="$(CONTAINER_RUNTIME=docker bash "$BOOTSTRAP" demo-docker --dry-run)" || fail "Docker dry-run should succeed"
grep -q -- "--user $(id -u):$(id -g)" <<<"$docker_output" || fail "Docker should use the host UID/GID"

profile_output="$(GITHUB_USER=example bash "$BOOTSTRAP" profile-site --profile clean --dry-run)" || fail "clean profile should succeed"
for repo in all-in-one-wp-migration admin-site-enhancements-pro; do
  grep -q "$repo" <<<"$profile_output" || fail "clean profile should include $repo"
done
if grep -q 'plugin activate all-in-one-wp-migration' <<<"$profile_output"; then
  fail "the migration plugin should remain inactive in profiles"
fi

for profile in clean custom loaded; do
  typerocket_output="$(GITHUB_USER=example bash "$BOOTSTRAP" "typerocket-$profile" --profile "$profile" --dry-run)" || fail "$profile profile should install TypeRocket"
  grep -q 'clone Git must-use plugin: typerocket-pro-v6' <<<"$typerocket_output" || fail "$profile profile should include TypeRocket Pro v6"
  grep -q 'typerocket-pro-v6/typerocket/galaxy' <<<"$typerocket_output" || fail "$profile profile should copy the TypeRocket Galaxy launcher"
  grep -q "apps/typerocket-$profile/galaxy" <<<"$typerocket_output" || fail "$profile profile should copy Galaxy to the WordPress root"
done
bare_output="$(GITHUB_USER=example bash "$BOOTSTRAP" typerocket-bare --profile bare --dry-run)" || fail "bare profile should succeed"
if grep -q 'typerocket-pro-v6' <<<"$bare_output"; then
  fail "bare profile should not include TypeRocket Pro v6"
fi

fake_typerocket="$typerocket_install_root/real-site/wp-content/mu-plugins/typerocket-pro-v6"
mkdir -p "$fake_typerocket/typerocket"
printf '%s\n' '<?php // TypeRocket MU entry' > "$fake_typerocket/typerocket-pro-v6.php"
printf '%s\n' '#!/usr/bin/env php' > "$fake_typerocket/typerocket/galaxy"
chmod +x "$fake_typerocket/typerocket/galaxy"
(
  source "$BOOTSTRAP"
  APPS_DIR="$typerocket_install_root"
  SITE_NAME=real-site
  DRY_RUN=false
  MU_PLUGIN_SOURCES=('git@github.com:example/typerocket-pro-v6.git')
  install_mu_plugins
) >/dev/null || fail "TypeRocket MU-plugin files should install"
cmp -s "$fake_typerocket/typerocket-pro-v6.php" "$typerocket_install_root/real-site/wp-content/mu-plugins/typerocket-pro-v6.php" || fail "TypeRocket MU entry should be copied to mu-plugins"
cmp -s "$fake_typerocket/typerocket/galaxy" "$typerocket_install_root/real-site/galaxy" || fail "TypeRocket Galaxy launcher should be copied to the WordPress root"
[[ -x "$typerocket_install_root/real-site/galaxy" ]] || fail "TypeRocket Galaxy launcher should remain executable"

loaded_output="$(GITHUB_USER=example bash "$BOOTSTRAP" loaded-site --profile loaded --dry-run)" || fail "loaded profile should succeed"
grep -q 'query-monitor' <<<"$loaded_output" || fail "loaded profile should include development plugins"
[[ "$(grep -Ec '^[[:space:]]*wp-plugin ' "$SCRIPT_DIR/profiles/loaded.profile")" -eq 12 ]] || fail "loaded profile should retain 12 WordPress.org plugins"

printf 'wp-plugin query-monitor\n' > "$generic_profile_root/base.fragment"
printf 'include base.fragment\ngithub-mu-plugin utility-mu\ngithub-theme utility-theme\n' > "$generic_profile_root/generic.profile"
generic_profile_output="$(
  source "$BOOTSTRAP"
  PROFILE_DIR="$generic_profile_root"
  PROFILE=generic
  GITHUB_USER=example
  DRY_RUN=true
  SITE_NAME=generic-site
  RUNTIME=podman
  CONTAINER_USER=0:0
  load_profile
  install_mu_plugins
  install_plugins
  install_themes
)" || fail "generic profile directives should load and install"
for expected in \
  'clone Git must-use plugin: utility-mu' \
  'utility-mu.php' \
  'install WordPress.org plugin: query-monitor' \
  'clone Git theme: utility-theme' \
  'theme activate utility-theme' \
  'theme delete --all'; do
  grep -q "$expected" <<<"$generic_profile_output" || fail "generic profile should include: $expected"
done

if bash "$BOOTSTRAP" demo --profile does-not-exist --dry-run >/dev/null 2>&1; then
  fail "unknown profiles should be rejected"
fi
if bash "$BOOTSTRAP" demo --plugin 'git:https://token@github.com/example/private.git' --dry-run >/dev/null 2>&1; then
  fail "credential-bearing Git URLs should be rejected"
fi
for removed_option in --scaffolds-only --project-slug --php-namespace --js-namespace; do
  if bash "$BOOTSTRAP" demo "$removed_option" value --dry-run >/dev/null 2>&1; then
    fail "$removed_option should be rejected"
  fi
done

cleanup
trap - EXIT
printf 'bootstrap tests passed\n'
