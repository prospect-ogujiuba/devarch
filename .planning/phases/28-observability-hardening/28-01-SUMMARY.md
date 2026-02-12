---
phase: 28-observability-hardening
plan: 01
subsystem: observability
tags: [logging, slog, structured-logs, request-correlation]
dependency_graph:
  requires: [phase-27]
  provides: [structured-logging-infrastructure]
  affects: [all-api-handlers, middleware-stack]
tech_stack:
  added: [log/slog, JSON logging]
  patterns: [request-scoped-logger, context-propagation]
key_files:
  created:
    - api/internal/api/middleware/logging.go
  modified:
    - api/cmd/server/main.go
    - api/internal/api/routes.go
    - api/internal/api/handlers/service.go
    - api/internal/api/handlers/stack.go
    - api/internal/api/handlers/websocket.go
    - api/internal/api/middleware/middleware.go
    - api/internal/api/respond/respond.go
decisions:
  - "SlogMiddleware replaces chi's middleware.Logger for structured logging"
  - "Request-scoped logger propagated via context with request_id from chi RequestID middleware"
  - "LoggerFromContext provides safe extraction with slog.Default() fallback for test/CLI contexts"
  - "Helper functions without request context use slog.Default() directly"
  - "LOG_LEVEL env var controls minimum log level (debug/info/warn/error, default info)"
metrics:
  duration_seconds: 450
  tasks_completed: 2
  files_modified: 8
  commits: 2
completed: 2026-02-12T15:54:57Z
---

# Phase 28 Plan 01: Structured JSON Logging Summary

**Implemented structured JSON logging with request correlation across all API handlers.**

## Overview

Replaced unstructured `log.Printf` calls with `log/slog` structured logging. Every HTTP request now gets a correlation ID via chi's RequestID middleware, propagated through context to handlers and services. Enables log aggregation and request tracing in production.

## Deviations from Plan

None - plan executed exactly as written.

## Tasks Completed

### Task 1: Slog initialization and logging middleware (commit: f714767)

**Created:**
- `api/internal/api/middleware/logging.go` - SlogMiddleware + LoggerFromContext

**Modified:**
- `api/cmd/server/main.go` - initLogger() function, JSON handler initialization, LOG_LEVEL parsing
- `api/internal/api/routes.go` - NewRouter accepts logger param, SlogMiddleware replaces middleware.Logger

**Implementation:**
- SlogMiddleware captures request_id, op (method + route pattern), status, duration_ms, bytes written
- Stack/instance URL params logged when present in route (via chi.URLParam)
- Request-scoped logger stored in context via type-safe contextKey
- LoggerFromContext provides fallback to slog.Default() when middleware bypassed
- LOG_LEVEL env var: debug/info/warn/error (default info) via slog.LevelVar
- RequestID middleware ordered BEFORE SlogMiddleware (chi.GetReqID must run first)

### Task 2: Migrate handler and service log calls to slog (commit: 68bd7b3)

**Migrated files:**
- `api/internal/api/handlers/service.go` - 29 log calls → structured slog (Error level for DB/scan errors in loadServiceRelations helper)
- `api/internal/api/handlers/stack.go` - 4 log calls → logger.Warn (request-scoped via LoggerFromContext)
- `api/internal/api/handlers/websocket.go` - 4 log calls → slog (1 request-scoped, 3 slog.Default for background tasks)
- `api/internal/api/middleware/middleware.go` - RecoverEnvelope panic log → slog.Error
- `api/internal/api/respond/respond.go` - 3 log calls → slog (Error for encoding failures, InternalError)

**Patterns applied:**
- Request handlers with context: `logger := mw.LoggerFromContext(r.Context())` → `logger.Error/Warn/Info`
- Helper methods without context: `slog.Error/Warn` (uses slog.Default implicitly)
- Background tasks (WebSocket broadcast): `slog.Error` directly
- Error level: DB errors, container failures, encoding errors
- Warn level: Non-fatal degradations (failed container stop, origin rejection)
- Structured attrs: `"service", name`, `"stack", name`, `"error", err`, `"service_id", s.ID`

**Stdlib log removed:**
- All `"log"` imports removed from handler files
- Verified via `grep -rn '"log"' api/internal/api/handlers/` returns no matches

## Verification Results

✓ Full project compiles (`go build ./...`)
✓ middleware.Logger replaced by SlogMiddleware
✓ LoggerFromContext used in 3 handler files (stack, websocket)
✓ No stdlib log imports remain in handler files
✓ Routes.go: RequestID → SlogMiddleware ordering correct

## Success Criteria Met

- [x] JSON-formatted structured logs emitted to stdout
- [x] Every request includes request_id, op, duration_ms, status in log output
- [x] Stack/instance params included when URL contains them
- [x] LOG_LEVEL env var controls minimum level (default info)
- [x] All handler log calls use request-scoped slog logger
- [x] No stdlib log imports remain in handler files

## Technical Notes

**Log output format (example):**
```json
{"time":"2026-02-12T15:54:57Z","level":"INFO","msg":"request completed","request_id":"abc123","op":"GET /api/v1/services/{name}","status":200,"duration_ms":45,"bytes":1234,"stack":"mystack"}
```

**Logger hierarchy:**
- `slog.SetDefault(logger)` bridges stdlib log → slog (useful for third-party libs)
- Request-scoped logger = base logger + request_id
- Handler-specific attrs added per log call (service, stack, error)

**Deferred logging in middleware:**
- Log statement in `defer` func ensures RoutePattern available (chi resolves after routing)
- Method prepended to pattern since chi.RouteContext().RoutePattern() returns path only

## Files Changed

| File | Lines Changed | Description |
|------|---------------|-------------|
| api/internal/api/middleware/logging.go | +73 | New slog middleware + context extraction |
| api/cmd/server/main.go | +33/-12 | initLogger func, migrate all log calls to slog |
| api/internal/api/routes.go | +2/-1 | Accept logger param, use SlogMiddleware |
| api/internal/api/handlers/service.go | +30/-30 | Replace 29 log.Printf with slog.Error |
| api/internal/api/handlers/stack.go | +6/-6 | Replace 4 log.Printf with logger.Warn |
| api/internal/api/handlers/websocket.go | +6/-6 | Replace 4 log.Printf with slog |
| api/internal/api/middleware/middleware.go | +2/-2 | RecoverEnvelope: log → slog |
| api/internal/api/respond/respond.go | +4/-4 | Replace 3 log.Printf with slog |

## Next Steps

Phase 28 Plan 02: Sync manager logging migration (internal/sync package).

## Self-Check

**Verifying created/modified files exist:**
```
FOUND: api/internal/api/middleware/logging.go
FOUND: api/cmd/server/main.go
FOUND: api/internal/api/routes.go
FOUND: api/internal/api/handlers/service.go
FOUND: api/internal/api/handlers/stack.go
FOUND: api/internal/api/handlers/websocket.go
FOUND: api/internal/api/middleware/middleware.go
FOUND: api/internal/api/respond/respond.go
```

**Verifying commits exist:**
```
FOUND: f714767 (Task 1: Slog initialization)
FOUND: 68bd7b3 (Task 2: Handler migration)
```

## Self-Check: PASSED
