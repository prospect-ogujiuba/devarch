---
phase: 18-websocket-auth-security-modes
plan: 02
subsystem: security
tags: [auth, websocket, token, hmac]
dependency_graph:
  requires: [18-01]
  provides: [ws-token-auth, hmac-signing]
  affects: [websocket-handler, dashboard-ws-client]
tech_stack:
  added: [internal/security/token.go]
  patterns: [hmac-sha256-signing, token-based-ws-auth]
key_files:
  created:
    - api/internal/security/token.go
  modified:
    - api/internal/api/handlers/auth.go
    - api/internal/api/routes.go
    - api/internal/api/handlers/websocket.go
    - dashboard/src/lib/api.ts
    - dashboard/src/hooks/use-websocket.ts
decisions:
  - HMAC-SHA256 signing using DEVARCH_API_KEY as secret (no additional secret needed)
  - 60-second token TTL (sufficient for WS upgrade handshake)
  - Token format {hex-payload}.{hex-mac} for simplicity (no JWT dependency)
  - Empty token response when auth disabled (backward compat)
metrics:
  duration: 137
  tasks: 2
  files_created: 1
  files_modified: 5
  completed: 2026-02-11
---

# Phase 18 Plan 02: Signed WS Token Authentication Summary

JWT auth with refresh rotation using jose library

## Tasks Completed

### Task 1: Create HMAC token package and wire issuance endpoint
**Commit:** ba24f4e

Created token generation and validation infrastructure:

**api/internal/security/token.go:**
- `GenerateWSToken(secret, ttl)` — creates HMAC-SHA256 signed token with format `{hex-payload}.{hex-mac}`
- `ValidateWSToken(token, secret)` — verifies signature using constant-time comparison, checks expiry
- Token payload: JSON with `exp` unix timestamp
- 60-second TTL (short-lived, survives WS upgrade only)

**api/internal/api/handlers/auth.go:**
- Added `WSToken` handler method
- Validates X-API-Key header before issuing token
- Returns `{"token":""}` when DEVARCH_API_KEY unset (auth disabled)
- Uses DEVARCH_API_KEY as HMAC secret (no separate secret needed)

**api/internal/api/routes.go:**
- Registered `/api/v1/auth/ws-token` route alongside `/api/v1/auth/validate`
- Placed outside middleware group (handler does own key validation)

**Files created:**
- `api/internal/security/token.go`

**Files modified:**
- `api/internal/api/handlers/auth.go`
- `api/internal/api/routes.go`

### Task 2: Enforce WS token in handler and update dashboard client
**Commit:** 57be674

Integrated token validation into WebSocket flow:

**api/internal/api/handlers/websocket.go:**
- Check `h.secMode.RequiresWSAuth()` before upgrade
- If true: extract `token` query param, validate using `security.ValidateWSToken`
- If invalid/missing: return 401 BEFORE upgrade (prevents WS connection)
- If false (dev-open, dev-keyed): proceed directly to upgrade (backward compat)

**dashboard/src/lib/api.ts:**
- Added `fetchWSToken()` helper
- Fetches token from `/api/v1/auth/ws-token` with X-API-Key header
- Returns empty string if no key set or on error (graceful degradation)

**dashboard/src/hooks/use-websocket.ts:**
- Updated `connect()` to async
- Fetch WS token before creating WebSocket
- Append `?token={token}` to WS URL if token non-empty
- Reconnect with exponential backoff automatically fetches fresh token (handles expiry)

**Files modified:**
- `api/internal/api/handlers/websocket.go`
- `dashboard/src/lib/api.ts`
- `dashboard/src/hooks/use-websocket.ts`

## Deviations from Plan

None — plan executed exactly as written.

## Verification

All verification checks passed:

```
✓ cd api && go build ./... — compiles cleanly
✓ cd dashboard && npx tsc --noEmit — type-checks
✓ GenerateWSToken function exists in token.go
✓ ValidateWSToken function exists in token.go
✓ RequiresWSAuth check in websocket.go
✓ ws-token route registered in routes.go
✓ fetchWSToken helper in api.ts
✓ token query param in use-websocket.ts
```

## Success Criteria Met

- [x] SEC-04 satisfied: browser WS clients authenticate via signed query token
- [x] Token is HMAC-SHA256 signed with 60s TTL, validated server-side before WS upgrade
- [x] strict mode rejects WS without valid token
- [x] dev-open and dev-keyed allow WS without token (backward compat)
- [x] Dashboard fetches fresh token on each WS connect/reconnect
- [x] No new external dependencies (stdlib crypto only)

## Impact

**Security posture:** WebSocket connections now respect security modes from Plan 01:
- `dev-open` — no WS auth (current behavior)
- `dev-keyed` — no WS auth (current behavior)
- `strict` — WS token required (new enforcement)

**Authentication flow:**
1. Dashboard calls `/api/v1/auth/ws-token` with X-API-Key
2. API validates key, issues HMAC-SHA256 signed token (60s TTL)
3. Dashboard appends token to WS URL query param
4. API validates token before upgrade (strict mode only)
5. Reconnect automatically fetches fresh token (handles expiry)

**Backward compatibility:** Fully preserved. dev-open and dev-keyed modes skip WS token validation. strict mode is opt-in.

**Developer experience:** No config changes needed for existing deployments. strict mode requires setting SECURITY_MODE=strict.

**Token design rationale:**
- HMAC-SHA256 over JWT: simpler, no external deps, sufficient for short-lived tokens
- DEVARCH_API_KEY as secret: one secret to manage, key only signs (not encrypts)
- 60s TTL: minimal window, fresh token on each connect/reconnect
- Query param delivery: standard for browser WS (can't set headers in WebSocket constructor)

## Self-Check

**Created files:**
✓ FOUND: api/internal/security/token.go

**Modified files:**
✓ FOUND: api/internal/api/handlers/auth.go
✓ FOUND: api/internal/api/routes.go
✓ FOUND: api/internal/api/handlers/websocket.go
✓ FOUND: dashboard/src/lib/api.ts
✓ FOUND: dashboard/src/hooks/use-websocket.ts

**Commits:**
✓ FOUND: ba24f4e (Task 1)
✓ FOUND: 57be674 (Task 2)

## Self-Check: PASSED
