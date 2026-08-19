#!/usr/bin/env bash
set -Eeuo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOOTSTRAP="$TEST_DIR/../bootstrap.sh"
# shellcheck source=../bootstrap.sh
source "$BOOTSTRAP"

TEST_TMP="$(mktemp -d)"
trap 'rm -rf "$TEST_TMP"' EXIT
passed=0

pass() { ((passed += 1)); }
fail() { printf 'FAIL: %s\n' "$1" >&2; exit 1; }
assert_eq() { [[ "$1" == "$2" ]] || fail "$3 (expected '$2', got '$1')"; pass; }
assert_contains() { [[ "$1" == *"$2"* ]] || fail "$3 (missing '$2')"; pass; }
expect_failure() { if ( "$@" ) >/dev/null 2>&1; then fail "expected failure: $*"; fi; pass; }

# Composer constraints: accepted common Composer forms and rejected malformed input.
valid_constraints=('^12.0' '~12.1' '>=12 <13' '12 || 13' '12 | 13' '1.0 - 2.0' 'dev-main' '12.*' '*')
invalid_constraints=(banana '^^12' '12 |||| 13' '12 || || 13' '>=' '12,' ',12' '|12' '12|' '1.0 - banana' '12.*.3' '12.x.3')
for constraint in "${valid_constraints[@]}"; do
  validate_composer_constraint "$constraint"
  pass
done
for constraint in "${invalid_constraints[@]}"; do
  expect_failure validate_composer_constraint "$constraint"
done
if command -v composer >/dev/null 2>&1; then
  for constraint in "${valid_constraints[@]}"; do
    printf '{"name":"devarch/preflight","description":"preflight","license":"proprietary","require":{"laravel/laravel":"%s"}}' "$constraint" > "$TEST_TMP/composer.json"
    composer validate "$TEST_TMP/composer.json" --no-check-all --no-check-publish --no-interaction >/dev/null 2>&1 || fail "Composer rejected valid table entry: $constraint"
    pass
  done
  for constraint in "${invalid_constraints[@]}"; do
    printf '{"name":"devarch/preflight","description":"preflight","license":"proprietary","require":{"laravel/laravel":"%s"}}' "$constraint" > "$TEST_TMP/composer.json"
    if composer validate "$TEST_TMP/composer.json" --no-check-all --no-check-publish --no-interaction >/dev/null 2>&1; then fail "Composer accepted invalid table entry: $constraint"; fi
    pass
  done
fi
version_marker="$PROJECT_ROOT/apps/.devarch-recovery/phase02-version-preflight-test"
[[ ! -e "$version_marker" ]] || fail "pre-existing test recovery marker: $version_marker"
for constraint in banana '^^12' '12 |||| 13'; do
  expect_failure "$BOOTSTRAP" phase02-version-preflight-test --version "$constraint" --dry-run
  [[ ! -e "$version_marker" ]] || fail 'invalid constraint mutated recovery state'
  pass
done

# Repository dotenv parsing and generated dotenv encoding round-trip.
parse_repository_env_value TEST '"quoted \"value\" \\ path \$cash" # trailing comment'
assert_eq "$PARSED_ENV_VALUE" 'quoted "value" \ path $cash' 'double-quoted repository dotenv decoding'
original="$PARSED_ENV_VALUE"
encode_dotenv "$original"
parse_repository_env_value TEST "$DOTENV_ENCODED"
assert_eq "$PARSED_ENV_VALUE" "$original" 'repository-to-Laravel dotenv round-trip'
printf 'APP_NAME=old\nAPP_ENV=local\n' > "$TEST_TMP/application.env"
dotenv_set APP_NAME "$original" "$TEST_TMP/application.env"
assert_contains "$(<"$TEST_TMP/application.env")" "APP_NAME=$DOTENV_ENCODED" 'dotenv_set writes encoded value under set -u'
TARGET="$TEST_TMP/runtime-writable"
DATABASE=sqlite
mkdir -p "$TARGET/storage/framework/views" "$TARGET/bootstrap/cache" "$TARGET/database"
touch "$TARGET/database/database.sqlite"
chmod -R 755 "$TARGET/storage" "$TARGET/bootstrap/cache" "$TARGET/database"
chmod 644 "$TARGET/database/database.sqlite"
make_runtime_writable
assert_eq "$(stat -c '%a' "$TARGET/storage/framework/views")" 777 'Laravel storage is writable by PHP-FPM'
assert_eq "$(stat -c '%a' "$TARGET/bootstrap/cache")" 777 'Laravel bootstrap cache is writable by PHP-FPM'
assert_eq "$(stat -c '%a' "$TARGET/database")" 777 'Laravel SQLite directory is writable by PHP-FPM'
assert_eq "$(stat -c '%a' "$TARGET/database/database.sqlite")" 666 'Laravel SQLite database is writable by PHP-FPM'
parse_repository_env_value TEST "'single \\'quote\\' and \\\\ path' # comment"
assert_eq "$PARSED_ENV_VALUE" "single 'quote' and \\ path" 'single-quoted repository dotenv decoding'
parse_repository_env_value TEST 'plain # trailing comment'
assert_eq "$PARSED_ENV_VALUE" plain 'unquoted repository dotenv comment stripping'
parse_repository_env_value TEST 'plain#literal'
assert_eq "$PARSED_ENV_VALUE" 'plain#literal' 'unquoted literal hash preservation'
expect_failure parse_repository_env_value TEST $'bad\rvalue'
expect_failure parse_repository_env_value TEST $'bad\nvalue'
expect_failure parse_repository_env_value TEST $'bad\tvalue'
expect_failure parse_repository_env_value TEST $'bad\177value'
printf 'LARAVEL_APP_NAME=ok\0bad\n' > "$TEST_TMP/nul.env"
expect_failure bash -c 'source "$1"; ENV_FILE="$2"; load_repository_env' _ "$BOOTSTRAP" "$TEST_TMP/nul.env"
printf 'LARAVEL_APP_NAME=bad\t\n' > "$TEST_TMP/control.env"
expect_failure bash -c 'source "$1"; ENV_FILE="$2"; load_repository_env' _ "$BOOTSTRAP" "$TEST_TMP/control.env"
printf '# unrelated\tcomment\nUNRELATED="ignored\177value"\nLARAVEL_APP_NAME="Accepted" # note\n' > "$TEST_TMP/repository.env"
ENV_FILE="$TEST_TMP/repository.env"
unset LARAVEL_APP_NAME
load_repository_env
assert_eq "$LARAVEL_APP_NAME" Accepted 'only supported repository dotenv values are validated'

# A durable guard exists before mutation; failed state updates leave it blocking retries.
RECOVERY_MARKER="$TEST_TMP/recovery/demo"
create_recovery_guard
assert_eq "$(<"$RECOVERY_MARKER")" 'provisioning in progress' 'durable recovery guard creation'
LARAVEL_TEST_FAIL_RECOVERY_MARKER_WRITE=1
expect_failure persist_recovery_state 'recovery incomplete'
unset LARAVEL_TEST_FAIL_RECOVERY_MARKER_WRITE
assert_eq "$(<"$RECOVERY_MARKER")" 'provisioning in progress' 'failed marker update preserves durable guard'
APP_NAME=demo
APPS_DIR="$TEST_TMP"
TARGET="$APPS_DIR/demo"
DATABASE=sqlite
FORCE=true
DRY_RUN=true
RECOVERY_MARKER="$APPS_DIR/.devarch-recovery/demo"
mkdir -p "$(dirname "$RECOVERY_MARKER")"
printf 'provisioning in progress\n' > "$RECOVERY_MARKER"
expect_failure validate_config
rm -f "$RECOVERY_MARKER"
LARAVEL_TEST_FAIL_RECOVERY_MARKER_CREATE=1
expect_failure create_recovery_guard
unset LARAVEL_TEST_FAIL_RECOVERY_MARKER_CREATE

# Persistent Docker/Podman detection and ordered-plan checks.
mkdir -p "$TEST_TMP/bin"
for runtime in docker podman; do
  cat > "$TEST_TMP/bin/$runtime" <<'RUNTIME'
#!/usr/bin/env bash
[[ "$1 $2" == 'compose version' ]]
RUNTIME
  chmod +x "$TEST_TMP/bin/$runtime"
done
PATH="$TEST_TMP/bin:$PATH"
for runtime in docker podman; do
  RUNTIME="$runtime"
  CONTAINER_USER=""
  COMPOSE=()
  detect_runtime
  assert_eq "${COMPOSE[*]}" "$runtime compose" "$runtime Compose provider"
  [[ "$runtime" != podman ]] || assert_eq "$CONTAINER_USER" '0:0' 'Podman bind-mount user'
done
TARGET="$TEST_TMP/demo"
BACKUP_PATH="$TEST_TMP/backup"
DATABASE=mariadb
WITH_MAILPIT=true
WITH_REDIS=true
MIGRATE=true
SEED=true
VERSION='^12.0'
APP_URL='https://demo.test'
DB_NAME=laravel_demo
DB_USER=lv_demo
plan="$(print_plan)"
assert_contains "$plan" 'ensure external network: microservices-net' 'ordered plan network step'
assert_contains "$plan" 'start and wait: php' 'ordered plan PHP step'
assert_contains "$plan" 'start and wait: mariadb' 'ordered plan MariaDB step'
assert_contains "$plan" 'start and wait: redis' 'ordered plan Redis step'
assert_contains "$plan" 'scaffold Laravel through Composer' 'ordered plan Composer step'

# Redis values stay centralized and match the Compose contract.
assert_eq "$REDIS_HOST" redis 'Redis container hostname'
assert_eq "$REDIS_PASSWORD" devarch 'Redis Compose password'
assert_eq "$REDIS_PORT" 6379 'Redis container port'
assert_eq "$REDIS_DB" 0 'Redis default DB'
assert_eq "$REDIS_CACHE_DB" 1 'Redis cache DB'
grep -Eq '127\.0\.0\.1:8504:6379' "$REDIS_COMPOSE" || fail 'Redis Compose port contract drifted'
grep -Eq 'redis-server --requirepass devarch' "$REDIS_COMPOSE" || fail 'Redis Compose password contract drifted'
pass
pass

printf 'PASS: %d assertions\n' "$passed"
