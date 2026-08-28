#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ROUTER_COMPOSE="$PROJECT_ROOT/services-library/backend/node/compose.yml"
APP_COMPOSE="$PROJECT_ROOT/services-library/backend/node/app.compose.yml"
PROXY_CONTAINER=nginx-proxy-manager
RUNTIME="${CONTAINER_RUNTIME:-}"
[[ -n "$RUNTIME" ]] || { command -v podman >/dev/null 2>&1 && RUNTIME=podman || RUNTIME=docker; }

for command_name in "$RUNTIME" curl node; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'SKIP: %s is required for Node wildcard routing integration tests\n' "$command_name"
    exit 0
  fi
done
if [[ "$("$RUNTIME" inspect -f '{{.State.Running}}' "$PROXY_CONTAINER" 2>/dev/null || true)" != true ]]; then
  printf 'SKIP: running %s container is required\n' "$PROXY_CONTAINER"
  exit 0
fi

COMPOSE=("$RUNTIME" compose)
SUFFIX="$$"
APP_A="js-route-a-$SUFFIX"
APP_B="js-route-b-$SUFFIX"
STATIC_APP="js-static-$SUFFIX"
ROUTER_EXISTED=false
ROUTER_WAS_RUNNING=false
passed=0

if "$RUNTIME" inspect node >/dev/null 2>&1; then
  ROUTER_EXISTED=true
  if [[ "$("$RUNTIME" inspect -f '{{.State.Running}}' node 2>/dev/null || true)" == true ]]; then
    ROUTER_WAS_RUNNING=true
  fi
fi

cleanup() {
  local app
  for app in "$APP_A" "$APP_B"; do
    DEVARCH_NODE_APP_NAME="$app" "${COMPOSE[@]}" -p "devarch-node-$app" -f "$APP_COMPOSE" down -v >/dev/null 2>&1 || true
  done
  rm -rf "$PROJECT_ROOT/apps/$APP_A" "$PROJECT_ROOT/apps/$APP_B" "$PROJECT_ROOT/apps/$STATIC_APP"
  if [[ "$ROUTER_WAS_RUNNING" == false ]]; then
    if [[ "$ROUTER_EXISTED" == true ]]; then
      "${COMPOSE[@]}" -f "$ROUTER_COMPOSE" stop >/dev/null 2>&1 || true
    else
      "${COMPOSE[@]}" -f "$ROUTER_COMPOSE" down >/dev/null 2>&1 || true
    fi
  fi
  "$RUNTIME" exec "$PROXY_CONTAINER" nginx -s reload >/dev/null 2>&1 || true
}
trap cleanup EXIT

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { ((passed += 1)); }

create_runtime_app() {
  local app="$1" target
  target="$PROJECT_ROOT/apps/$app"
  mkdir -p "$target"
  printf '{"scripts":{"devarch":"node server.js"}}\n' > "$target/package.json"
  cat > "$target/server.js" <<'JS'
const http = require('node:http');
const app = '__DEVARCH_TEST_APP__';
const server = http.createServer((request, response) => {
  response.setHeader('content-type', 'text/plain');
  response.end(`${app}:${request.url}`);
});
server.on('upgrade', (request, socket) => {
  if (request.url === '/reject-socket') {
    socket.end('HTTP/1.1 401 Unauthorized\r\nContent-Length: 6\r\nConnection: close\r\n\r\ndenied');
    return;
  }
  socket.end('HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\naccepted');
});
server.listen(3000, '0.0.0.0');
JS
  sed -i "s/__DEVARCH_TEST_APP__/$app/" "$target/server.js"
}

start_runtime_app() {
  local app="$1"
  DEVARCH_NODE_APP_NAME="$app" \
  DEVARCH_NODE_SCRIPT=devarch \
  DEVARCH_NODE_PACKAGE_MANAGER=npm \
  DEVARCH_NODE_CONTAINER_USER=0:0 \
    "${COMPOSE[@]}" -p "devarch-node-$app" -f "$APP_COMPOSE" up -d --build >/dev/null
}

request() {
  local app="$1" path="$2" expected_body="$3" expected_status="$4" output status body
  output="$(mktemp)"
  status="$(curl -ksS -o "$output" -w '%{http_code}' --resolve "$app.test:443:127.0.0.1" "https://$app.test$path")" || {
    rm -f "$output"
    fail "$app.test$path request failed"
  }
  body="$(<"$output")"
  rm -f "$output"
  [[ "$status" == "$expected_status" ]] || fail "$app.test$path returned HTTP $status, expected $expected_status"
  [[ "$body" == "$expected_body" ]] || fail "$app.test$path returned '$body', expected '$expected_body'"
  pass
}

create_runtime_app "$APP_A"
create_runtime_app "$APP_B"
mkdir -p "$PROJECT_ROOT/apps/$STATIC_APP/out/nested" "$PROJECT_ROOT/apps/$STATIC_APP/out/_next/static"
printf '{}\n' > "$PROJECT_ROOT/apps/$STATIC_APP/package.json"
printf 'static-home\n' > "$PROJECT_ROOT/apps/$STATIC_APP/out/index.html"
printf 'static-about\n' > "$PROJECT_ROOT/apps/$STATIC_APP/out/about.html"
printf 'static-nested\n' > "$PROJECT_ROOT/apps/$STATIC_APP/out/nested/index.html"
printf 'static-missing\n' > "$PROJECT_ROOT/apps/$STATIC_APP/out/404.html"
printf 'static-asset\n' > "$PROJECT_ROOT/apps/$STATIC_APP/out/_next/static/app.js"

"${COMPOSE[@]}" -f "$ROUTER_COMPOSE" up -d --build >/dev/null
start_runtime_app "$APP_A"
start_runtime_app "$APP_B"
"$RUNTIME" exec "$PROXY_CONTAINER" nginx -t >/dev/null
"$RUNTIME" exec "$PROXY_CONTAINER" nginx -s reload >/dev/null
sleep 1

for app in "$APP_A" "$APP_B"; do
  ready=false
  for _ in {1..30}; do
    if [[ "$(curl -ksS -o /dev/null -w '%{http_code}' --resolve "$app.test:443:127.0.0.1" "https://$app.test/" 2>/dev/null)" == 200 ]]; then
      ready=true
      break
    fi
    sleep 2
  done
  [[ "$ready" == true ]] || fail "$app did not become ready"
done

request "$APP_A" / "$APP_A:/" 200
request "$APP_B" / "$APP_B:/" 200
request "$APP_A" /api/check "$APP_A:/api/check" 200
request "$APP_A" /favicon.ico "$APP_A:/favicon.ico" 200
request "$APP_A" /robots.txt "$APP_A:/robots.txt" 200
request "$STATIC_APP" / static-home 200
request "$STATIC_APP" /about static-about 200
request "$STATIC_APP" /nested/ static-nested 200
request "$STATIC_APP" /_next/static/app.js static-asset 200
request "$STATIC_APP" /missing static-missing 404

DEVARCH_TEST_HOST="$APP_A.test" node <<'JS'
const tls = require('node:tls');
const host = process.env.DEVARCH_TEST_HOST;
function upgrade(path, marker) {
  return new Promise((resolve, reject) => {
    let response = '';
    let settled = false;
    const finish = () => {
      if (settled) return;
      settled = true;
      socket.destroy();
      resolve(response);
    };
    const socket = tls.connect({ host: '127.0.0.1', port: 443, servername: host, rejectUnauthorized: false }, () => {
      socket.write(`GET ${path} HTTP/1.1\r\nHost: ${host}\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n`);
    });
    socket.setEncoding('utf8');
    socket.setTimeout(5000, () => {
      if (!settled) socket.destroy(new Error('upgrade timed out'));
    });
    socket.on('data', (chunk) => {
      response += chunk;
      if (response.endsWith(marker)) finish();
    });
    socket.on('end', finish);
    socket.on('error', reject);
  });
}
(async () => {
  const accepted = await upgrade('/custom-socket', 'accepted');
  if (!accepted.startsWith('HTTP/1.1 101') || !accepted.endsWith('accepted')) throw new Error(`unexpected accepted upgrade: ${accepted}`);
  const rejected = await upgrade('/reject-socket', 'denied');
  if (!rejected.startsWith('HTTP/1.1 401') || !rejected.endsWith('denied')) throw new Error(`unexpected rejected upgrade: ${rejected}`);
})().catch((error) => { console.error(error.message); process.exit(1); });
JS
pass
pass

if [[ -d "$PROJECT_ROOT/apps/playground" ]] && "$RUNTIME" inspect php >/dev/null 2>&1; then
  headers="$(curl -ksSI --resolve playground.test:443:127.0.0.1 https://playground.test/)"
  grep -Eqi '^x-detected-language:[[:space:]]*php' <<<"$headers" || fail 'existing playground.test no longer detects PHP'
  pass
fi

printf 'Node wildcard routing integration tests passed (%d assertions)\n' "$passed"
