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
for repo in all-in-one-wp-migration admin-site-enhancements-pro typerocket-pro-v6 makermaker; do
  grep -q "$repo" <<<"$profile_output" || fail "clean profile should include $repo"
done
for removed_repo in makerblocks makerstarter; do
  if grep -q "$removed_repo" <<<"$profile_output"; then fail "clean profile should prune missing $removed_repo"; fi
done
grep -q 'must-use plugin' <<<"$profile_output" || fail "clean profile should install TypeRocket as an MU plugin"

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
[[ "$(profile_entries clean)" == $'github-mu-plugin typerocket-pro-v6\ngithub-mu-plugin makermaker\ngithub-plugin all-in-one-wp-migration inactive\ngithub-plugin admin-site-enhancements-pro' ]] || fail "clean profile should contain TypeRocket, MakerMaker, and accessible repositories"
[[ "$(profile_entries custom)" == $'github-mu-plugin typerocket-pro-v6\ngithub-mu-plugin makermaker\ngithub-plugin all-in-one-wp-migration inactive\ngithub-plugin admin-site-enhancements-pro\ngithub-plugin manual-image-crop' ]] || fail "custom profile should contain TypeRocket, MakerMaker, and accessible repositories"
[[ "$(profile_entries loaded | grep -c '^wp-plugin ')" -eq 12 ]] || fail "loaded profile should retain all 12 development plugins"

printf 'bootstrap tests passed\n'
