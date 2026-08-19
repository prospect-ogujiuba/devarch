#!/usr/bin/env bash
set -Eeuo pipefail

SOURCE_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_PROJECT_ROOT="$(cd "$SOURCE_SCRIPT_DIR/../.." && pwd)"
TEST_TMP="$(mktemp -d)"
PROJECT_ROOT="$TEST_TMP/project"
SCRIPT_DIR="$PROJECT_ROOT/scripts/laravel"
BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"
APP_NAME="laravel-regression-$$"
APP_TARGET="$PROJECT_ROOT/apps/$APP_NAME"
RUNTIME_LOG="$TEST_TMP/runtime.log"
passed=0

cleanup() {
  rm -rf "$TEST_TMP"
}
trap cleanup EXIT

mkdir -p "$SCRIPT_DIR"
cp "$SOURCE_SCRIPT_DIR/bootstrap.sh" "$SCRIPT_DIR/bootstrap.sh"
cp -R "$SOURCE_SCRIPT_DIR/profiles" "$SCRIPT_DIR/profiles"
for compose_file in \
  services-library/backend/php/compose.yml \
  services-library/database/mariadb/compose.yml \
  services-library/database/redis/compose.yml \
  services-library/mail/mailpit/compose.yml \
  services-library/proxy/nginx-proxy-manager/compose.yml; do
  mkdir -p "$PROJECT_ROOT/$(dirname "$compose_file")"
  cp "$SOURCE_PROJECT_ROOT/$compose_file" "$PROJECT_ROOT/$compose_file"
done
[[ ! -e "$PROJECT_ROOT/.env" ]] || { printf 'FAIL: temporary project fixture must not contain .env\n' >&2; exit 1; }

pass() { ((passed += 1)); }
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_contains() { grep -Fq -- "$2" <<<"$1" || fail "$3 (missing '$2')"; pass; }
assert_absent() { ! grep -Fq -- "$2" <<<"$1" || fail "$3 (found '$2')"; pass; }

mkdir -p "$TEST_TMP/bin"
for runtime in podman docker; do
  cat > "$TEST_TMP/bin/$runtime" <<'RUNTIME'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${FAKE_RUNTIME_LOG:?}"
[[ "$#" -eq 2 && "$1" == compose && "$2" == version ]]
RUNTIME
  chmod +x "$TEST_TMP/bin/$runtime"
done
: > "$RUNTIME_LOG"

run_bootstrap() {
  local runtime="$1"
  shift
  PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME="$runtime" \
    bash "$BOOTSTRAP" "$@"
}

expect_failure() {
  local description="$1"
  shift
  if run_bootstrap podman "$@" >"$TEST_TMP/failure.out" 2>&1; then
    fail "$description"
  fi
  pass
}

# Informational modes are standalone and do not inspect a container runtime.
help_output="$(PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" bash "$BOOTSTRAP" --help)" || fail '--help should succeed'
for option in --version --profile --package --dev-package --packages-file --database --with-mailpit --with-redis --migrate --no-migrate --seed --force --no-hosts --dry-run; do
  assert_contains "$help_output" "$option" "help should document $option"
done
profiles_output="$(PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" bash "$BOOTSTRAP" --list-profiles)" || fail '--list-profiles should succeed'
for profile in bare standard loaded; do
  grep -Eq "^${profile}[[:space:]]" <<<"$profiles_output" || fail "profile list should include $profile"
  pass
done
[[ ! -s "$RUNTIME_LOG" ]] || fail 'informational modes must not inspect a container runtime'
pass

# Safety-critical input is rejected before any provisioning command.
expect_failure 'missing app name should fail'
expect_failure 'unsafe app name should fail' 'unsafe/name' --dry-run
expect_failure 'uppercase app name should fail' 'Unsafe' --dry-run
expect_failure 'unknown option should fail' demo --unknown --dry-run
expect_failure 'unsupported v1 option should fail' demo --repository example/repo --dry-run
expect_failure 'invalid database should fail' demo --database postgres --dry-run
expect_failure 'conflicting migration flags should fail' demo --migrate --no-migrate --dry-run
expect_failure 'unknown profile should fail' demo --profile missing --dry-run
expect_failure 'missing packages file should fail' demo --packages-file "$TEST_TMP/missing" --dry-run
expect_failure 'invalid package should fail' demo --package Invalid --dry-run
expect_failure 'invalid package constraint should fail' demo --package 'vendor/package:^^1' --dry-run
expect_failure 'invalid Laravel version should fail' demo --version '12 |||| 13' --dry-run

printf 'feature redis\n' > "$TEST_TMP/invalid.packages"
expect_failure 'package files must reject profile features' demo --packages-file "$TEST_TMP/invalid.packages" --dry-run

if PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME=podman \
  LARAVEL_APP_URL='ftp://invalid.test' bash "$BOOTSTRAP" demo --dry-run >"$TEST_TMP/failure.out" 2>&1; then
  fail 'non-HTTP Laravel URL should fail'
fi
pass
if PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME=podman \
  LARAVEL_CONTAINER_USER='1000' bash "$BOOTSTRAP" demo --dry-run >"$TEST_TMP/failure.out" 2>&1; then
  fail 'invalid container user should fail'
fi
pass
if PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME=invalid \
  bash "$BOOTSTRAP" demo --dry-run >"$TEST_TMP/failure.out" 2>&1; then
  fail 'invalid container runtime should fail'
fi
pass

# A representative loaded plan maps services, packages, migrations, and secrets.
cat > "$TEST_TMP/packages" <<'PACKAGES'
composer-package vendor/runtime ^1.0
composer-dev-package vendor/tool dev-main
PACKAGES
secret='laravel-regression-db-secret'
loaded_output="$(
  PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME=podman \
  LARAVEL_DB_ROOT_PASSWORD="$secret" \
  bash "$BOOTSTRAP" service-rich \
    --profile loaded \
    --version '^12.0' \
    --packages-file "$TEST_TMP/packages" \
    --package 'vendor/cli-runtime:>=2 <3' \
    --dev-package vendor/cli-tool \
    --seed \
    --dry-run
)" || fail 'representative loaded dry-run should succeed'
for expected in \
  'runtime: podman; container user: 0:0' \
  'profile: loaded' \
  'start and wait: php' \
  'start and wait: nginx-proxy-manager' \
  'start and wait: mariadb; create isolated database/user: laravel_service_rich / lv_service_rich' \
  'start and wait: mailpit' \
  'start and wait: redis (password redacted)' \
  'scaffold Laravel through Composer (constraint: ^12.0)' \
  'Composer production package: vendor/runtime:^1.0' \
  'Composer development package: vendor/tool:dev-main' \
  'Composer production package: vendor/cli-runtime:>=2 <3' \
  'Composer development package: vendor/cli-tool' \
  'configure APP_URL=https://service-rich.test and database=mariadb; generate APP_KEY' \
  'run migrations' \
  'run database seeder once' \
  'register local host: 127.0.0.1 service-rich.test'; do
  assert_contains "$loaded_output" "$expected" 'loaded dry-run plan drifted'
done
assert_absent "$loaded_output" "$secret" 'database root password must be redacted'
no_hosts_output="$(run_bootstrap podman no-hosts-app --no-hosts --dry-run)" || fail 'hosts opt-out dry-run should succeed'
assert_contains "$no_hosts_output" 'hosts registration skipped: no-hosts-app.test' 'hosts opt-out should be visible'

# SQLite/no-migrate omits database and optional-service provisioning.
sqlite_output="$(run_bootstrap podman sqlite-app --database sqlite --no-migrate --dry-run)" || fail 'SQLite dry-run should succeed'
assert_contains "$sqlite_output" 'database=sqlite' 'SQLite plan should select SQLite'
for unexpected in 'start and wait: mariadb' 'start and wait: mailpit' 'start and wait: redis' 'run migrations' 'run database seeder'; do
  assert_absent "$sqlite_output" "$unexpected" 'SQLite bare plan should omit unselected work'
done

# Runtime user mapping and force backup planning are visible without changing fixtures.
docker_output="$(run_bootstrap docker docker-app --database sqlite --dry-run)" || fail 'Docker dry-run should succeed'
assert_contains "$docker_output" "runtime: docker; container user: $(id -u):$(id -g)" 'Docker should map the host UID/GID'
mkdir -p "$APP_TARGET"
printf 'preserve me\n' > "$APP_TARGET/original.txt"
expect_failure 'existing target without force should fail' "$APP_NAME" --database sqlite --dry-run
force_output="$(run_bootstrap podman "$APP_NAME" --database sqlite --force --dry-run)" || fail 'forced replacement dry-run should succeed'
assert_contains "$force_output" 'backup existing target:' 'force should plan a backup'
[[ "$(<"$APP_TARGET/original.txt")" == 'preserve me' ]] || fail 'dry-run changed the existing application fixture'
[[ ! -e "$PROJECT_ROOT/apps/.devarch-recovery/$APP_NAME" ]] || fail 'dry-run created a recovery marker'
if compgen -G "$PROJECT_ROOT/apps/.devarch-backups/$APP_NAME-*" >/dev/null || \
   compgen -G "$PROJECT_ROOT/apps/.devarch-failed/$APP_NAME-*" >/dev/null; then
  fail 'dry-run created backup or failed-application state'
fi
pass

# The fake runtime rejects every mutation; only Compose capability checks are allowed.
if grep -Fvx 'compose version' "$RUNTIME_LOG" >"$TEST_TMP/unexpected-runtime.log"; then
  fail "dry-run attempted a container mutation: $(head -n1 "$TEST_TMP/unexpected-runtime.log")"
fi
pass

printf 'PASS: %d assertions\n' "$passed"
