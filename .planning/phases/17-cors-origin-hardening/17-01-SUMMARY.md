---
phase: 17-cors-origin-hardening
plan: 01
subsystem: security
tags: [cors, websocket, security, origin-validation]
dependency_graph:
  requires: []
  provides: [origin-allowlist, cors-hardening, websocket-hardening]
  affects: [api-security, dashboard-connectivity]
tech_stack:
  added: []
  patterns: [env-driven-security, allowlist-validation]
key_files:
  created: []
  modified:
    - api/internal/api/routes.go
    - api/internal/api/handlers/websocket.go
    - .env.example
    - compose.yml
decisions:
  - decision: "Use comma-separated ALLOWED_ORIGINS env var for both CORS and WebSocket"
    rationale: "Single source of truth prevents config drift"
    alternatives: ["Separate env vars", "Config file"]
  - decision: "Default to wildcard when ALLOWED_ORIGINS unset"
    rationale: "Dev-friendly - no breaking changes to existing setups"
    alternatives: ["Require explicit configuration", "Fail safe closed"]
  - decision: "Enable AllowCredentials only for non-wildcard origins"
    rationale: "CORS spec forbids AllowCredentials with wildcard origins"
    alternatives: ["Always enable", "Separate env var"]
metrics:
  duration_seconds: 98
  tasks_completed: 2
  files_modified: 4
  commits: 2
  completed_at: "2026-02-11T14:42:00Z"
---

# Phase 17 Plan 01: CORS Origin Hardening Summary

**One-liner:** Environment-driven CORS and WebSocket origin validation using ALLOWED_ORIGINS allowlist with dev-friendly wildcard default.

## What Was Built

**CORS Middleware Hardening:**
- Read `ALLOWED_ORIGINS` env var as comma-separated list
- Replace hardcoded `AllowedOrigins: []string{"*"}` with dynamic slice
- Set `AllowCredentials: true` when specific origins configured (required for auth headers)
- Default to wildcard `["*"]` when env var unset (backward compatible)

**WebSocket Upgrader Hardening:**
- Changed `NewWebSocketHandler` signature to accept `allowedOrigins []string`
- Replaced permissive `CheckOrigin: func(r *http.Request) bool { return true }` with allowlist logic
- Wildcard or empty allowlist permits all origins (dev mode)
- Missing `Origin` header permitted (non-browser clients like CLI tools)
- Disallowed origins rejected with log message

**Configuration:**
- Added `ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-}` to `compose.yml` devarch-api environment
- Documented in `.env.example` with `http://localhost:5174` example

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All verification criteria passed:
- ✅ `grep -c "ALLOWED_ORIGINS" api/internal/api/routes.go` returns 2 (env read + usage)
- ✅ `grep "ALLOWED_ORIGINS" compose.yml` shows env var forwarded
- ✅ `grep "ALLOWED_ORIGINS" .env.example` shows documentation
- ✅ `cd api && go build ./...` compiles cleanly
- ✅ `grep "allowedOrigins" api/internal/api/handlers/websocket.go` shows parameter usage
- ✅ `grep "NewWebSocketHandler.*allowedOrigins" api/internal/api/routes.go` shows updated call
- ✅ `grep "rejected connection" api/internal/api/handlers/websocket.go` confirms rejection logging
- ✅ No hardcoded `AllowedOrigins: []string{"*"}` in routes.go
- ✅ No hardcoded `return true` in websocket.go CheckOrigin

## Requirements Satisfied

- **SEC-02:** CORS middleware driven by ALLOWED_ORIGINS env var ✅
- **SEC-03:** WebSocket upgrader rejects disallowed origins ✅

## Success Criteria Met

- [x] SEC-02 satisfied: CORS middleware driven by ALLOWED_ORIGINS env var
- [x] SEC-03 satisfied: WebSocket upgrader rejects disallowed origins
- [x] Dev-friendly: unset ALLOWED_ORIGINS defaults to permissive (no breaking change)
- [x] Both CORS and WS share the same allowlist source

## Technical Implementation Notes

**CORS AllowCredentials Logic:**
```go
allowCredentials := len(allowedOrigins) > 0 && allowedOrigins[0] != "*"
```
This ensures `AllowCredentials: false` when wildcard (CORS spec requirement) and `true` when specific origins configured (needed for cookie/auth headers).

**WebSocket Origin Validation:**
- Empty `Origin` header is allowed (non-browser clients don't send it)
- Wildcard or empty allowlist is permissive
- Explicit allowlist requires exact match
- Rejections are logged: `websocket: rejected connection from origin "http://example.com"`

**Backward Compatibility:**
Existing deployments without `ALLOWED_ORIGINS` env var will default to wildcard behavior (no breaking change). Production deployments should set `ALLOWED_ORIGINS=http://localhost:5174` or their dashboard URL.

## Files Changed

### api/internal/api/routes.go
- Added `"strings"` import
- Read `ALLOWED_ORIGINS` env var after `STACK_IMPORT_MAX_BYTES` read
- Parse comma-separated list and trim whitespace
- Calculate `allowCredentials` flag based on wildcard check
- Replace hardcoded CORS `AllowedOrigins` and `AllowCredentials` with dynamic values
- Pass `allowedOrigins` to `NewWebSocketHandler`

### api/internal/api/handlers/websocket.go
- Changed `NewWebSocketHandler` signature to accept `allowedOrigins []string` parameter
- Replaced `CheckOrigin: func(r *http.Request) bool { return true }` with allowlist validation logic
- Added wildcard detection, missing Origin header handling, and rejection logging

### .env.example
- Added `ALLOWED_ORIGINS` section with `http://localhost:5174` example

### compose.yml
- Added `ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-}` to devarch-api environment (empty default)

## Self-Check

Verifying all claims:

**Created files:**
None claimed, none created. ✅

**Modified files:**
- api/internal/api/routes.go
- api/internal/api/handlers/websocket.go
- .env.example
- compose.yml

```bash
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/routes.go" ] && echo "FOUND: api/internal/api/routes.go" || echo "MISSING"
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/websocket.go" ] && echo "FOUND: api/internal/api/handlers/websocket.go" || echo "MISSING"
[ -f "/home/fhcadmin/projects/devarch/.env.example" ] && echo "FOUND: .env.example" || echo "MISSING"
[ -f "/home/fhcadmin/projects/devarch/compose.yml" ] && echo "FOUND: compose.yml" || echo "MISSING"
```

**Commits:**
- 1aec0aa: feat(17-01): wire ALLOWED_ORIGINS into CORS middleware
- b617d80: feat(17-01): enforce origin allowlist in WebSocket upgrader

```bash
git log --oneline --all | grep -q "1aec0aa" && echo "FOUND: 1aec0aa" || echo "MISSING"
git log --oneline --all | grep -q "b617d80" && echo "FOUND: b617d80" || echo "MISSING"
```

## Self-Check: PASSED
