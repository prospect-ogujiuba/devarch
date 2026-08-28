#!/usr/bin/env bash
set -Eeuo pipefail

SOURCE_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_PROJECT_ROOT="$(cd "$SOURCE_SCRIPT_DIR/../.." && pwd)"
TEST_TMP="$(mktemp -d)"
PROJECT_ROOT="$TEST_TMP/project"
SCRIPT_DIR="$PROJECT_ROOT/scripts/node"
BOOTSTRAP="$SCRIPT_DIR/bootstrap.sh"
RUNTIME_LOG="$TEST_TMP/runtime.log"
passed=0

cleanup() { rm -rf "$TEST_TMP"; }
trap cleanup EXIT
pass() { ((passed += 1)); }
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_contains() { grep -Fq -- "$2" <<<"$1" || fail "$3 (missing '$2')"; pass; }

mkdir -p "$SCRIPT_DIR" "$PROJECT_ROOT/apps/demo" "$PROJECT_ROOT/scripts/hosts"
cp "$SOURCE_SCRIPT_DIR/bootstrap.sh" "$BOOTSTRAP"
cp "$SOURCE_PROJECT_ROOT/scripts/hosts/register-host.sh" "$PROJECT_ROOT/scripts/hosts/register-host.sh"
for file in \
  services-library/backend/node/compose.yml \
  services-library/backend/node/app.compose.yml \
  services-library/proxy/nginx-proxy-manager/compose.yml; do
  mkdir -p "$PROJECT_ROOT/$(dirname "$file")"
  cp "$SOURCE_PROJECT_ROOT/$file" "$PROJECT_ROOT/$file"
done

cat > "$PROJECT_ROOT/apps/demo/package.json" <<'JSON'
{"scripts":{"devarch":"next dev --hostname 0.0.0.0 --port 3000"}}
JSON
: > "$PROJECT_ROOT/apps/demo/package-lock.json"
mkdir -p "$TEST_TMP/bin"
cat > "$TEST_TMP/bin/podman" <<'RUNTIME'
#!/usr/bin/env bash
printf 'app=%s script=%s manager=%s user=%s | %s\n' \
  "${DEVARCH_NODE_APP_NAME:-}" "${DEVARCH_NODE_SCRIPT:-}" \
  "${DEVARCH_NODE_PACKAGE_MANAGER:-}" "${DEVARCH_NODE_CONTAINER_USER:-}" "$*" \
  >> "${FAKE_RUNTIME_LOG:?}"
if [[ "$1 $2" == "compose version" ]]; then exit 0; fi
if [[ "$1 $2" == "network inspect" ]]; then exit 0; fi
exit 0
RUNTIME
chmod +x "$TEST_TMP/bin/podman"
: > "$RUNTIME_LOG"

run_bootstrap() {
  PATH="$TEST_TMP/bin:$PATH" FAKE_RUNTIME_LOG="$RUNTIME_LOG" CONTAINER_RUNTIME=podman \
    bash "$BOOTSTRAP" "$@"
}

help_output="$(PATH="$TEST_TMP/bin:$PATH" bash "$BOOTSTRAP" --help)" || fail '--help should succeed'
for text in '--script' '--package-manager' '--no-hosts' '--dry-run'; do
  assert_contains "$help_output" "$text" 'help should document the public contract'
done

if run_bootstrap 'Unsafe/name' --dry-run >"$TEST_TMP/error" 2>&1; then fail 'unsafe app names must fail'; fi
pass
if run_bootstrap missing --dry-run >"$TEST_TMP/error" 2>&1; then fail 'missing app directories must fail'; fi
pass

max_name="$(printf 'a%.0s' {1..58})"
too_long_name="${max_name}a"
mkdir -p "$PROJECT_ROOT/apps/$max_name" "$PROJECT_ROOT/apps/$too_long_name"
cp "$PROJECT_ROOT/apps/demo/package.json" "$PROJECT_ROOT/apps/$max_name/package.json"
cp "$PROJECT_ROOT/apps/demo/package.json" "$PROJECT_ROOT/apps/$too_long_name/package.json"
run_bootstrap "$max_name" --dry-run >/dev/null || fail '58-character app names should remain valid after the node- prefix'
pass
if run_bootstrap "$too_long_name" --dry-run >"$TEST_TMP/error" 2>&1; then fail '59-character app names must fail the portable DNS boundary'; fi
pass

output="$(run_bootstrap demo --dry-run)" || fail 'representative dry-run should succeed'
for text in \
  'app: demo (https://demo.test)' \
  'container: node-demo' \
  'package manager: npm' \
  'package script: devarch' \
  'start shared Node router and wildcard proxy' \
  'start isolated app runtime on microservices-net'; do
  assert_contains "$output" "$text" 'dry-run plan drifted'
done
[[ ! -s "$RUNTIME_LOG" ]] || fail 'dry-run must not inspect or mutate the container runtime'
pass

if run_bootstrap demo --script missing --dry-run >"$TEST_TMP/error" 2>&1; then fail 'undefined package scripts must fail'; fi
pass

: > "$RUNTIME_LOG"
run_bootstrap demo --no-hosts >/dev/null || fail 'representative provisioning should succeed with the recording runtime'
runtime_calls="$(<"$RUNTIME_LOG")"
assert_contains "$runtime_calls" 'app=demo script=devarch manager=npm user=0:0' 'app Compose environment drifted'
assert_contains "$runtime_calls" '-p devarch-node-demo' 'app runtime must use an isolated Compose project'
assert_contains "$runtime_calls" 'up -d --build --force-recreate' 'app runtime must apply Compose environment changes'
app_compose="$(<"$PROJECT_ROOT/services-library/backend/node/app.compose.yml")"
assert_contains "$app_compose" '__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS' 'app runtime must trust its proxied Vite hostname'
assert_contains "$runtime_calls" 'nginx -t' 'proxy configuration must be validated before reload'
assert_contains "$runtime_calls" 'nginx -s reload' 'proxy configuration must be reloaded'

printf 'node bootstrap tests passed (%d assertions)\n' "$passed"
