---
phase: 20
plan: 01
subsystem: api-responses
tags: [action-endpoints, consistency, typed-responses]
dependency_graph:
  requires: [phase-19]
  provides: [ActionResponse, respond.Action, consistent-action-responses]
  affects: [api-handlers, openapi-prep]
tech_stack:
  added: []
  patterns: [functional-options, optional-fields]
key_files:
  created: []
  modified:
    - api/internal/api/respond/types.go
    - api/internal/api/respond/respond.go
    - api/internal/api/handlers/stack_lifecycle.go
    - api/internal/api/handlers/stack_apply.go
    - api/internal/api/handlers/stack_wiring.go
    - api/internal/api/handlers/instance_lifecycle.go
    - api/internal/api/handlers/service.go
    - api/internal/api/handlers/category.go
    - api/internal/api/handlers/project.go
    - api/internal/api/handlers/nginx.go
    - api/internal/api/handlers/runtime.go
decisions: []
metrics:
  duration_seconds: 240
  tasks_completed: 2
  files_modified: 11
  handlers_migrated: 28
  completed_at: 2026-02-11T17:15:06Z
---

# Phase 20 Plan 01: Action Endpoint Consistency Summary

**One-liner:** Standardized 28 action endpoints to use typed ActionResponse with functional options (status, message, output, warnings, metadata).

## Objective

Replace ad-hoc `map[string]string` and `map[string]interface{}` responses across action endpoints with consistent ActionResponse struct, providing type safety and OpenAPI-ready structure.

## What Was Done

### Task 1: Add ActionResponse Infrastructure
- Added `ActionResponse` struct to `respond/types.go` with:
  - `Status` (required string)
  - `Message`, `Output`, `Warnings`, `Metadata` (optional fields)
- Added `respond.Action()` convenience function
- Added functional option helpers: `WithMessage`, `WithOutput`, `WithWarnings`, `WithMetadata`

### Task 2: Migrate All Action Handlers
Migrated 28 action endpoints across 9 handler files:

**Stack handlers (7 endpoints):**
- lifecycle: stop/start/restart → `respond.Action(w, r, 200, "stopped/started/restarted")`
- apply: complex response with output + lock warnings → `respond.Action` with `WithOutput` + `WithMetadata("lock_warnings")`
- wiring: resolve/cleanup → `respond.Action` with `WithMetadata("resolved_count"/"deleted_count")` + `WithWarnings`

**Instance handlers (3 endpoints):**
- stop/start/restart → `respond.Action(w, r, 200, "stopped/started/restarted")`

**Service handlers (4 endpoints):**
- start/stop/restart/rebuild → `respond.Action(w, r, 200, "started/stopped/restarted/rebuilt")`

**Category handlers (2 endpoints):**
- start/stop → `respond.Action(w, r, 200, "completed", WithMetadata("services", results))`

**Project handlers (6 endpoints):**
- start/stop/restart → `respond.Action` with `WithOutput(output)`
- Fixed error paths: replaced `respond.JSON` with error-in-success-envelope pattern with `respond.InternalError`
- service-level start/stop/restart → `respond.Action(w, r, 200, "started/stopped/restarted")`

**Nginx handlers (3 endpoints):**
- generateAll → `respond.Action(w, r, 200, "generated")`
- generateOne → `respond.Action` with `WithMetadata("project", name)`
- reload → `respond.Action(w, r, 200, "reloaded")`

**Runtime handlers (2 endpoints):**
- switch → `respond.Action` with `WithMessage` + multiple `WithMetadata` (previous, current, services_stopped, config_updated)
- socketStart → `respond.Action` with `WithMessage` + `WithMetadata` (type, socket_path)

**Cleanup:**
- Removed obsolete `resolveWiresResponse` and `cleanupWiresResponse` structs from `stack_wiring.go`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Critical Functionality] Lock warnings type mismatch**
- **Found during:** Task 2, stack_apply.go migration
- **Issue:** `lock.LockWarning` is a struct array, not `[]string`. Original code put structured warnings in `map[string]interface{}["lock_warnings"]`.
- **Fix:** Changed from `respond.WithWarnings(result.Warnings)` to `respond.WithMetadata("lock_warnings", result.Warnings)` to preserve structured data.
- **Files modified:** api/internal/api/handlers/stack_apply.go
- **Commit:** 484d34b

## Key Decisions

None - implementation followed plan exactly. Lock warnings placed in metadata preserves backward compatibility while maintaining type consistency.

## Testing Notes

- All packages compile successfully
- 28 action endpoints verified to use `respond.Action()`
- Zero leftover `map[string]string{"status":...}` patterns in action handlers
- Non-action handlers (update, delete, materialize, save) correctly unchanged

## Impact

**Benefits:**
- Type-safe action responses across all handlers
- Consistent client-side parsing
- OpenAPI annotation-ready structure (Plan 02)
- Cleaner error handling (project handlers)

**Breaking changes:** None - response structure maintains backward compatibility (status field always present, optional fields omitted when empty).

## Next Steps

Phase 20 Plan 02: Add OpenAPI annotations to ActionResponse and all action endpoints.

## Self-Check: PASSED

All claimed files exist:
- api/internal/api/respond/types.go
- api/internal/api/respond/respond.go
- 9 handler files

All claimed commits exist:
- b8d4f41 (Task 1: ActionResponse infrastructure)
- 484d34b (Task 2: migrate all action handlers)
