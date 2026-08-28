#!/usr/bin/env bash
set -Eeuo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_TMP="$(mktemp -d)"
PROJECT_ROOT="$TEST_TMP/project"
BOOTSTRAP="$PROJECT_ROOT/scripts/javascript/bootstrap.sh"
passed=0

cleanup() { rm -rf "$TEST_TMP"; }
trap cleanup EXIT
pass() { ((passed += 1)); }
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_contains() { grep -Fq -- "$2" <<<"$1" || fail "$3 (missing '$2')"; pass; }
assert_absent() { ! grep -Fq -- "$2" <<<"$1" || fail "$3 (found '$2')"; pass; }

mkdir -p "$PROJECT_ROOT/scripts/javascript" "$PROJECT_ROOT/scripts/node" "$PROJECT_ROOT/apps" "$TEST_TMP/bin"
cp "$SOURCE_DIR/bootstrap.sh" "$BOOTSTRAP"
cp -R "$SOURCE_DIR/profiles" "$PROJECT_ROOT/scripts/javascript/profiles"
cat > "$PROJECT_ROOT/scripts/node/bootstrap.sh" <<'RUNTIME'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${FAKE_START_LOG:?}"
RUNTIME
chmod +x "$BOOTSTRAP" "$PROJECT_ROOT/scripts/node/bootstrap.sh"

cat > "$TEST_TMP/bin/npm" <<'FAKE'
#!/usr/bin/env bash
printf 'npm %s\n' "$*" >> "${FAKE_COMMAND_LOG:?}"
if [[ ${1:-} == create ]]; then
  mkdir -p app
  printf '{"name":"generated","scripts":{"dev":"vite"}}\n' > app/package.json
elif [[ ${1:-} == install ]]; then
  printf '{}\n' > package-lock.json
fi
FAKE
cat > "$TEST_TMP/bin/npx" <<'FAKE'
#!/usr/bin/env bash
printf 'npx %s\n' "$*" >> "${FAKE_COMMAND_LOG:?}"
mkdir -p app
printf '{"name":"generated","scripts":{}}\n' > app/package.json
FAKE
chmod +x "$TEST_TMP/bin/npm" "$TEST_TMP/bin/npx"
: > "$TEST_TMP/commands.log"
: > "$TEST_TMP/start.log"

run_bootstrap() {
  PATH="$TEST_TMP/bin:$PATH" FAKE_COMMAND_LOG="$TEST_TMP/commands.log" \
    FAKE_START_LOG="$TEST_TMP/start.log" bash "$BOOTSTRAP" "$@"
}

help_output="$(run_bootstrap --help)" || fail '--help should succeed'
for expected in --framework --profile --list-frameworks --list-profiles --start --force --dry-run; do
  assert_contains "$help_output" "$expected" 'help contract drifted'
done

frameworks_output="$(run_bootstrap --list-frameworks)" || fail '--list-frameworks should succeed'
for framework in angular astro next nuxt qwik react-router sveltekit vite-lit vite-preact vite-react vite-solid vite-vue; do
  grep -Eq "^${framework}[[:space:]]" <<<"$frameworks_output" || fail "framework list should include $framework"
  pass
done
assert_contains "$frameworks_output" 'vite-react           typescript' 'framework defaults should be visible'

profiles_output="$(run_bootstrap --list-profiles)" || fail '--list-profiles should succeed'
for combination in \
  'angular              spa' 'angular              ssr' \
  'astro                blog' 'next                 api' 'next                 fullstack' \
  'nuxt                 content' 'qwik                 library' \
  'react-router         custom-server' 'sveltekit            tested' \
  'vite-react           compiler' 'vite-vue             javascript'; do
  assert_contains "$profiles_output" "$combination" 'curated combination should be listed'
done
filtered_output="$(run_bootstrap --list-profiles --framework next)" || fail 'filtered profile listing should succeed'
assert_contains "$filtered_output" 'next                 api' 'filtered listing should include selected framework profiles'
if grep -Fq 'astro' <<<"$filtered_output"; then fail 'filtered listing should omit other frameworks'; fi
pass

if run_bootstrap Unsafe --framework vite-react --profile typescript --dry-run >"$TEST_TMP/error" 2>&1; then fail 'uppercase names must fail'; fi
pass
if run_bootstrap demo --framework missing --profile minimal --dry-run >"$TEST_TMP/error" 2>&1; then fail 'unknown frameworks must fail'; fi
pass
if run_bootstrap demo --framework next --profile missing --dry-run >"$TEST_TMP/error" 2>&1; then fail 'unknown project profiles must fail'; fi
pass
if run_bootstrap demo --framework next --dry-run >"$TEST_TMP/error" 2>&1; then fail 'explicit framework selection requires a profile'; fi
pass
if run_bootstrap demo --profile minimal --dry-run >"$TEST_TMP/error" 2>&1; then fail 'project profiles require a framework'; fi
pass
if run_bootstrap demo --dry-run >"$TEST_TMP/error" 2>&1; then fail 'missing framework and profile must fail'; fi
pass

plan="$(run_bootstrap demo --framework vite-react --profile compiler --dry-run)" || fail 'framework/profile dry-run should succeed'
for expected in \
  'app: demo (https://demo.test)' \
  'framework: vite-react' \
  'profile: compiler' \
  '--template react-compiler-ts' \
  'normalize package script' \
  'install dependencies with npm'; do
  assert_contains "$plan" "$expected" 'dry-run plan drifted'
done
[[ ! -e "$PROJECT_ROOT/apps/demo" ]] || fail 'dry-run created an app'
[[ ! -s "$TEST_TMP/commands.log" ]] || fail 'dry-run invoked a package manager'
pass

legacy_plan="$(run_bootstrap legacy --profile vite-react --dry-run)" || fail 'legacy framework profile alias should succeed'
assert_contains "$legacy_plan" 'profile: typescript' 'legacy alias should select the framework default'
assert_contains "$legacy_plan" 'compatibility alias: --profile vite-react' 'legacy alias should be visible'

api_plan="$(run_bootstrap api-demo --framework next --profile api --dry-run)" || fail 'Next API profile should resolve'
assert_contains "$api_plan" '--api' 'Next API profile should select route-handler-only scaffolding'
custom_server_plan="$(run_bootstrap custom-server-demo --framework react-router --profile custom-server --dry-run)" || fail 'React Router custom-server profile should resolve'
assert_contains "$custom_server_plan" 'node-custom-server' 'custom-server profile should select the maintained deployment template'
assert_contains "$custom_server_plan" 'cross-env\ NODE_ENV=development' 'custom-server profile should preserve its custom development server'
assert_contains "$custom_server_plan" 'apply profile-specific runtime configuration' 'custom-server profile should integrate HMR with the wildcard proxy'
for qwik_profile in app playground; do
  qwik_plan="$(run_bootstrap "qwik-$qwik_profile-demo" --framework qwik --profile "$qwik_profile" --dry-run)" || fail "Qwik $qwik_profile profile should resolve"
  assert_contains "$qwik_plan" 'devarch=vite\ --mode\ ssr\ --host\ 0.0.0.0\ --port\ 3000' "Qwik $qwik_profile should preserve the official SSR development mode"
done
nuxt_minimal_plan="$(run_bootstrap nuxt-minimal-demo --framework nuxt --profile minimal --dry-run)" || fail 'Nuxt minimal profile should resolve'
assert_contains "$nuxt_minimal_plan" '--template v4' 'Nuxt minimal should choose the official v4 template non-interactively'
assert_contains "$nuxt_minimal_plan" '--no-gitInit' 'Nuxt minimal should explicitly disable git initialization'
nuxt_content_plan="$(run_bootstrap nuxt-content-demo --framework nuxt --profile content --dry-run)" || fail 'Nuxt content profile should resolve'
assert_contains "$nuxt_content_plan" '--template content' 'Nuxt content should use the untouched official content template'
assert_contains "$nuxt_content_plan" '--no-gitInit' 'Nuxt content should explicitly disable git initialization'
assert_absent "$nuxt_content_plan" '--modules=@nuxt/content' 'Nuxt content should not retrofit the module onto the v4 template'
assert_contains "$nuxt_content_plan" 'apply profile-specific runtime configuration' 'Nuxt content should install its required SQLite runtime'
for angular_profile in spa ssr; do
  angular_plan="$(run_bootstrap "angular-$angular_profile-demo" --framework angular --profile "$angular_profile" --dry-run)" || fail "Angular $angular_profile profile should resolve"
  assert_contains "$angular_plan" 'devarch=ng\ serve\ --host\ 0.0.0.0\ --port\ 3000\ --allowed-hosts' "Angular $angular_profile should enable allowed hosts"
  assert_absent "$angular_plan" '--allowed-hosts\ all' "Angular $angular_profile should not parse all as a project name"
done
svelte_tested_plan="$(run_bootstrap svelte-tested-demo --framework sveltekit --profile tested --dry-run)" || fail 'SvelteKit tested profile should resolve'
assert_contains "$svelte_tested_plan" 'vitest=usages:unit\,component' 'SvelteKit tested should encode both Vitest usage values as one option value'

run_bootstrap nuxt-content-runtime --framework nuxt --profile content >/dev/null || fail 'Nuxt content scaffold should succeed'
grep -Fqx 'npm install better-sqlite3' "$TEST_TMP/commands.log" || fail 'Nuxt content should install better-sqlite3'
pass

run_bootstrap demo --framework vite-react --profile typescript >/dev/null || fail 'Vite scaffold should succeed'
[[ -f "$PROJECT_ROOT/apps/demo/package-lock.json" ]] || fail 'post-install should create a lockfile'
pass
node -e 'const p=require(process.argv[1]); if(p.scripts.devarch!=="vite --host 0.0.0.0 --port 3000") process.exit(1)' \
  "$PROJECT_ROOT/apps/demo/package.json" || fail 'devarch package script was not normalized'
pass

printf 'preserve\n' > "$PROJECT_ROOT/apps/demo/original.txt"
if run_bootstrap demo --framework next --profile fullstack >"$TEST_TMP/error" 2>&1; then fail 'existing apps must be preserved without --force'; fi
pass
run_bootstrap demo --profile next --force --start --no-hosts >/dev/null || fail 'legacy alias force replacement and start should succeed'
grep -Fqx 'demo --no-hosts' "$TEST_TMP/start.log" || fail '--start should hand off runtime arguments'
pass
grep -Fq 'allowedDevOrigins: ["demo.test"]' "$PROJECT_ROOT/apps/demo/next.config.ts" ||
  fail 'Next.js profile should allow its wildcard-proxy development origin'
pass
compgen -G "$PROJECT_ROOT/apps/.devarch-backups/demo-*" >/dev/null || fail '--force should create a timestamped backup'
pass
backup="$(compgen -G "$PROJECT_ROOT/apps/.devarch-backups/demo-*" | head -n1)"
[[ "$(<"$backup/original.txt")" == preserve ]] || fail 'force backup did not preserve the original app'
pass

printf 'javascript bootstrap tests passed (%d assertions)\n' "$passed"
