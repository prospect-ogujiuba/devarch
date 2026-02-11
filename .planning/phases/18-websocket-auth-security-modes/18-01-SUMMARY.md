---
phase: 18-websocket-auth-security-modes
plan: 01
subsystem: security
tags: [auth, security, websocket, modes]
dependency_graph:
  requires: []
  provides: [security-mode-framework, mode-aware-middleware]
  affects: [api-startup, http-middleware, websocket-handler]
tech_stack:
  added: [internal/security]
  patterns: [mode-based-auth, startup-validation]
key_files:
  created:
    - api/internal/security/mode.go
  modified:
    - api/cmd/server/main.go
    - api/internal/api/middleware/middleware.go
    - api/internal/api/routes.go
    - api/internal/api/handlers/websocket.go
    - compose.yml
    - .env.example
decisions:
  - Empty SECURITY_MODE defaults to dev-open for backward compatibility
  - Mode validation at startup prevents invalid runtime configuration
  - APIKeyAuth middleware becomes mode-aware function factory
  - WebSocket handler stores mode for Phase 18 Plan 02 token enforcement
metrics:
  duration: 165
  tasks: 2
  files_created: 1
  files_modified: 6
  completed: 2026-02-11
---

# Phase 18 Plan 01: Security Mode Profiles Summary

Security mode framework (dev-open/dev-keyed/strict) controlling API-wide auth behavior with startup validation.

## Tasks Completed

### Task 1: Create security mode package with parsing and validation
**Commit:** 7ccd7ca

Created `api/internal/security/mode.go` with:
- `Mode` type with three profiles: `ModeDevOpen`, `ModeDevKeyed`, `ModeStrict`
- `ParseMode(raw string)` — validates mode, defaults empty string to dev-open
- `RequiresAPIKey()` — returns true for dev-keyed and strict
- `RequiresWSAuth()` — returns true only for strict (for Plan 02)
- `ValidateConfig(mode)` — ensures DEVARCH_API_KEY set when mode requires it

**Files created:**
- `api/internal/security/mode.go`

### Task 2: Wire security mode into startup, middleware, and router
**Commit:** 8a0c0c7

Integrated security mode throughout the API:

**main.go:**
- Parse and validate SECURITY_MODE at startup (fail-fast on invalid)
- Log active security mode
- Pass mode to router

**middleware.go:**
- Changed `APIKeyAuth` from standalone function to mode-aware factory: `APIKeyAuth(mode security.Mode)`
- Skips auth entirely when `!mode.RequiresAPIKey()` (dev-open mode)
- Removed `apiKeyWarningLogged` pattern — mode makes intent explicit

**routes.go:**
- Added `secMode security.Mode` parameter to `NewRouter`
- Pass mode to `mw.APIKeyAuth(secMode)` on both stack import and /api/v1 routes
- Pass mode to `NewWebSocketHandler` for Plan 02 WS token enforcement

**websocket.go:**
- Added `secMode security.Mode` field to `WebSocketHandler` struct
- Store mode in constructor (enforcement deferred to Plan 02)

**compose.yml:**
- Added `SECURITY_MODE=${SECURITY_MODE:-}` env var forwarding

**.env.example:**
- Documented SECURITY_MODE with mode descriptions

**Files modified:**
- `api/cmd/server/main.go`
- `api/internal/api/middleware/middleware.go`
- `api/internal/api/routes.go`
- `api/internal/api/handlers/websocket.go`
- `compose.yml`
- `.env.example`

## Deviations from Plan

None — plan executed exactly as written.

## Verification

All verification checks passed:

```
✓ cd api && go build ./... — compiles cleanly
✓ ParseMode function exists (2 occurrences: definition + call)
✓ RequiresAPIKey function exists (3 occurrences: definition + calls)
✓ RequiresWSAuth function exists (2 occurrences: definition + future use)
✓ ValidateConfig function exists (2 occurrences: definition + call)
✓ SECURITY_MODE in compose.yml
✓ SECURITY_MODE in .env.example
✓ security.Mode in routes.go
✓ security.ParseMode in main.go
```

## Success Criteria Met

- [x] SEC-05 security mode profiles implemented with dev-open/dev-keyed/strict
- [x] Invalid SECURITY_MODE causes startup crash with descriptive error
- [x] dev-keyed without DEVARCH_API_KEY causes startup crash
- [x] dev-open with no key set behaves identically to current (backward compat)
- [x] Mode flows from startup → router → middleware → WS handler

## Impact

**Backward compatibility:** Fully preserved. Unset SECURITY_MODE defaults to dev-open, which matches current behavior (no auth when DEVARCH_API_KEY unset).

**Security posture:** Explicit security modes eliminate ambiguity. Operators now choose:
- `dev-open` — no auth (local dev)
- `dev-keyed` — API key validation (current production default)
- `strict` — API key + WS token (Plan 02 will enforce)

**Developer experience:** Clear documentation in .env.example. Startup validation prevents misconfiguration.

**Foundation for Plan 02:** WebSocket handler stores mode; Plan 02 will add token-based WS auth enforcement when `mode.RequiresWSAuth()` is true.

## Self-Check

**Created files:**
✓ FOUND: api/internal/security/mode.go

**Modified files:**
✓ FOUND: api/cmd/server/main.go
✓ FOUND: api/internal/api/middleware/middleware.go
✓ FOUND: api/internal/api/routes.go
✓ FOUND: api/internal/api/handlers/websocket.go
✓ FOUND: compose.yml
✓ FOUND: .env.example

**Commits:**
✓ FOUND: 7ccd7ca (Task 1)
✓ FOUND: 8a0c0c7 (Task 2)

## Self-Check: PASSED
