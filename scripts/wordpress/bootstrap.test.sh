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
grep -q -- '--no-hosts' <<<"$help_output" || fail "help should document hosts opt-out"
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
grep -q 'start PHP, MariaDB, and Nginx Proxy Manager services' <<<"$dry_run_output" || fail "dry-run should plan service startup"
grep -q 'services-library/proxy/nginx-proxy-manager/compose.yml up -d' <<<"$dry_run_output" || fail "dry-run should start Nginx Proxy Manager"
grep -q 'wait for PHP/WP-CLI, MariaDB, and Nginx Proxy Manager readiness' <<<"$dry_run_output" || fail "dry-run should wait for proxy readiness"
grep -q 'URL: https://demo-site.test' <<<"$dry_run_output" || fail "default URL should use wildcard .test routing"
grep -q 'Nginx Proxy Manager .test reverse proxy' <<<"$dry_run_output" || fail "completion should identify infrastructure routing"
grep -q 'register local host: 127.0.0.1 demo-site.test' <<<"$dry_run_output" || fail "dry-run should plan hosts registration"
no_hosts_output="$(bash "$BOOTSTRAP" no-hosts-site --no-hosts --dry-run)" || fail "hosts opt-out dry-run should succeed"
grep -q 'hosts registration skipped: no-hosts-site.test' <<<"$no_hosts_output" || fail "hosts opt-out should be visible"
grep -q 'config set FS_METHOD direct' <<<"$dry_run_output" || fail "bootstrap should enable direct filesystem changes without FTP credentials"
grep -q 'option update uploads_use_yearmonth_folders 0' <<<"$dry_run_output" || fail "bootstrap should disable year/month upload folders"
grep -q 'find .*wp-content/uploads .* -empty -delete' <<<"$dry_run_output" || fail "bootstrap should remove empty upload subdirectories"
grep -q "rewrite structure /%postname%/ --hard" <<<"$dry_run_output" || fail "bootstrap should use post-name permalinks"
grep -q 'post delete 1 --force' <<<"$dry_run_output" || fail "bootstrap should delete the default WordPress post"
grep -q 'chmod -R a+rwX.*/wp-content' <<<"$dry_run_output" || fail "bootstrap should keep local wp-content writable by PHP"
if grep -Eq 'not-printed-secret|not-printed-db-secret' <<<"$dry_run_output"; then
  fail "dry-run must not print secrets"
fi
if grep -Eq 'wp server|WORDPRESS_PORT' <<<"$dry_run_output"; then
  fail "bootstrap should not bypass Nginx Proxy Manager"
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
expected_clean=$'github-mu-plugin typerocket-pro-v6\ninclude maker-stack.fragment\ngithub-plugin all-in-one-wp-migration inactive\ngithub-plugin admin-site-enhancements-pro'
[[ "$(profile_entries clean)" == "$expected_clean" ]] || fail "clean profile should use the shared Maker stack fragment"
[[ "$(profile_entries custom)" == "$expected_clean"$'\nwp-plugin manual-image-crop' ]] || fail "custom profile should extend clean without duplicating Maker declarations"
for profile in clean custom loaded; do
  entries="$(profile_entries "$profile")"
  grep -q '^github-mu-plugin typerocket-pro-v6$' <<<"$entries" || fail "$profile profile should contain TypeRocket Pro v6"
  [[ "$(grep -c '^include maker-stack.fragment$' <<<"$entries")" -eq 1 ]] || fail "$profile should reference the shared Maker stack exactly once"
done
maker_stack="$(grep -Ev '^[[:space:]]*(#|$)' "$SCRIPT_DIR/profiles/maker-stack.fragment")"
[[ "$maker_stack" == $'maker-stack-channel stable\nmaker-core plugin makermaker\nmaker-core plugin makerblocks\nmaker-core theme makerstarter\nmaker-workspace child-theme makerstarter\nmaker-workspace blocks-plugin makerblocks\nmaker-workspace app-plugin makermaker' ]] || fail "shared Maker stack declarations drifted"
[[ "$(profile_entries loaded | grep -c '^wp-plugin ')" -eq 12 ]] || fail "loaded profile should retain all 12 development plugins"

for label in 'core install marker: makermaker' 'core install marker: makerblocks' 'core install marker: makerstarter' \
  'workspace create [child-theme]' 'workspace create [blocks-plugin]' 'workspace create [app-plugin via MakerMaker]' \
  'write Maker core lock manifest'; do
  grep -Fq "$label" <<<"$profile_output" || fail "Maker dry-run should label $label"
done
grep -q 'theme activate profile-site-theme' <<<"$profile_output" || fail "Maker profile should activate the child theme"
grep -q 'plugin activate profile-site-blocks' <<<"$profile_output" || fail "Maker profile should activate the project blocks plugin"
grep -q 'makermaker create profile-site-app.*--activate' <<<"$profile_output" || fail "Maker profile should generate and activate the app plugin through MakerMaker"

if GITHUB_USER=example bash "$BOOTSTRAP" demo-site --profile clean --project-slug Bad_Slug --dry-run >/dev/null 2>&1; then
  fail "invalid project slugs should be rejected"
fi
if GITHUB_USER=example bash "$BOOTSTRAP" demo-site --profile clean --php-namespace Single --dry-run >/dev/null 2>&1; then
  fail "single-segment PHP namespaces should be rejected"
fi
if GITHUB_USER=example bash "$BOOTSTRAP" demo-site --profile clean --js-namespace Bad/Namespace --dry-run >/dev/null 2>&1; then
  fail "invalid JS namespaces should be rejected"
fi
override_output="$(GITHUB_USER=example bash "$BOOTSTRAP" demo-site --profile clean --scaffolds-only --project-slug client-web --php-namespace 'Client\Web' --js-namespace client-blocks --dry-run)" || fail "scaffold-only override dry-run should succeed"
grep -Fq 'slug=client-web PHP=Client\Web JS=client-blocks' <<<"$override_output" || fail "explicit project identity overrides should be normalized into the plan"

scaffold_root="$(mktemp -d)"
(
  source "$BOOTSTRAP"
  APPS_DIR="$scaffold_root/apps"
  SITE_NAME=unit-site
  SITE_TITLE='Unit Site'
  PROJECT_SLUG=unit-site
  PHP_NAMESPACE='Maker\UnitSite'
  JS_NAMESPACE=unit-site
  DRY_RUN=false
  MAKER_WORKSPACES=('child-theme:makerstarter' 'blocks-plugin:makerblocks' 'app-plugin:makermaker')
  mkdir -p "$APPS_DIR/$SITE_NAME/wp-content/themes/makerstarter/scaffolds" "$APPS_DIR/$SITE_NAME/wp-content/plugins/makerblocks/scaffolds"
  cp -a "$PROJECT_ROOT/apps/playground/wp-content/themes/makerstarter/scaffolds/child-theme" "$APPS_DIR/$SITE_NAME/wp-content/themes/makerstarter/scaffolds/"
  cp -a "$PROJECT_ROOT/apps/playground/wp-content/plugins/makerblocks/scaffolds/project-plugin" "$APPS_DIR/$SITE_NAME/wp-content/plugins/makerblocks/scaffolds/"
  wp_exec() {
    local container_site="$1"; shift
    if [[ "${1:-}" == makermaker && "${2:-}" == create ]]; then
      mkdir -p "$APPS_DIR/$SITE_NAME/wp-content/plugins/${3}"
      printf '# Generated plugin\n' > "$APPS_DIR/$SITE_NAME/wp-content/plugins/${3}/README.md"
    fi
  }
  provision_maker_workspaces
  grep -q '^# PROJECT OWNED — EDIT HERE' "$APPS_DIR/$SITE_NAME/wp-content/themes/unit-site-theme/README.md"
  grep -q '^Template: makerstarter$' "$APPS_DIR/$SITE_NAME/wp-content/themes/unit-site-theme/style.css"
  grep -q '^# PROJECT OWNED — EDIT HERE' "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-blocks/README.md"
  grep -q 'unit-site namespace' "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-blocks/unit-site-blocks.php"
  grep -q '^# PROJECT OWNED — EDIT HERE' "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-app/README.md"
  before="$(find "$APPS_DIR/$SITE_NAME/wp-content/themes/unit-site-theme" "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-blocks" "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-app" -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum)"
  provision_maker_workspaces
  after="$(find "$APPS_DIR/$SITE_NAME/wp-content/themes/unit-site-theme" "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-blocks" "$APPS_DIR/$SITE_NAME/wp-content/plugins/unit-site-app" -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum)"
  [[ "$before" == "$after" ]]
  PROJECT_SLUG=interrupted
  if ( DEVARCH_SCAFFOLD_FAIL_AFTER_STAGE=true publish_child_theme_workspace ) >/dev/null 2>&1; then exit 1; fi
  [[ ! -e "$APPS_DIR/$SITE_NAME/wp-content/themes/interrupted-theme" ]]
  ! find "$APPS_DIR/$SITE_NAME/wp-content/themes" -maxdepth 1 -name 'interrupted-theme.devarch-stage.*' | grep -q .
) || fail "Maker workspace publication should be marked, idempotent, and rollback-safe"
rm -rf "$scaffold_root"

cleanup
trap - EXIT
printf 'bootstrap tests passed\n'
