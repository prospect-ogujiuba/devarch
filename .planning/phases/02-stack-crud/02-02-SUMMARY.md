---
phase: 02-stack-crud
plan: 02
subsystem: api
tags: [go, chi, stack-management, soft-delete, transactions]

# Dependency graph
requires:
  - phase: 02-01
    provides: StackHandler with CRUD + soft-delete migration
provides:
  - Enable/disable stack operations with container stopping
  - Clone stack (copy record and instances)
  - Rename stack (clone + soft-delete in transaction)
  - Trash operations (list/restore/permanent delete with name conflict handling)
  - Delete preview showing cascade impact
  - All stack routes registered at /api/v1/stacks
affects: [02-03, 02-04, 02-05, stack-ui, stack-operations]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Transaction-based operations (rename = clone + delete)
    - Trash pattern with name conflict detection on restore
    - Delete preview for blast radius visibility
    - Container name generation for stack instances

key-files:
  created: []
  modified:
    - api/internal/api/handlers/stack.go
    - api/internal/api/routes.go

key-decisions:
  - "Rename implemented as atomic clone + soft-delete transaction"
  - "Restore checks for active name conflicts with prescriptive error"
  - "Delete preview builds container names without side effects"
  - "Trash routes registered before /{name} routes to avoid chi parameter conflicts"

patterns-established:
  - "Transaction pattern: tx.BeginTx + defer tx.Rollback() + tx.Commit()"
  - "Container name format: container.ContainerName(stackName, instanceID)"
  - "Trash route ordering: /trash before /{name} to prevent chi routing issues"

# Metrics
duration: 2min
completed: 2026-02-03
---

# Phase 2 Plan 2: Stack Operations & Routes Summary

**Stack enable/disable/clone/rename/trash operations with full HTTP route wiring at /api/v1/stacks**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-03T21:56:16Z
- **Completed:** 2026-02-03T21:58:30Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Enable/disable operations with container stopping logic
- Clone creates full stack copy with all instances
- Rename atomically clones and soft-deletes original in single transaction
- Trash operations (list/restore/permanent delete) with name conflict handling
- Delete preview shows cascade impact without side effects
- All stack routes registered and accessible at /api/v1/stacks

## Task Commits

Each task was committed atomically:

1. **Task 1: Add advanced stack operations** - `639c570` (feat)
2. **Task 2: Register stack routes and wire in server** - `7bd1d0a` (feat)

## Files Created/Modified
- `api/internal/api/handlers/stack.go` - Added 8 methods: Enable, Disable, Clone, Rename, DeletePreview, ListTrash, Restore, PermanentDelete
- `api/internal/api/routes.go` - Registered stackHandler with 13 routes under /api/v1/stacks

## Decisions Made
- **Transaction-based rename:** Implemented as clone + soft-delete in single tx for atomicity
- **Trash route ordering:** Registered /trash routes before /{name} routes to prevent chi treating "trash" as URL parameter
- **Name conflict handling:** Restore operation checks for active stack with same name and returns prescriptive 409 error
- **Container name generation:** Used container.ContainerName(stackName, fmt.Sprintf("%d", instanceID)) for proper string formatting

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed ContainerName string conversion**
- **Found during:** Task 1 (Disable and DeletePreview implementation)
- **Issue:** container.ContainerName expects (string, string) but instanceID was int, causing compile error
- **Fix:** Added fmt.Sprintf("%d", instanceID) to convert int to string
- **Files modified:** api/internal/api/handlers/stack.go (2 locations)
- **Verification:** go build ./... passes
- **Committed in:** 639c570 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Type conversion fix necessary for compilation. No scope changes.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Stack operations API complete. Ready for:
- Phase 2 Plan 3: Dashboard UI integration
- Phase 2 Plan 4+: Instance management and runtime operations

**Blockers:** None

**Notes:**
- Container stopping in Disable is placeholder (Phase 3+ will wire actual containerClient.Stop calls)
- Clone copies instances but doesn't clone container state (stateless operation)
- Permanent delete cascades to service_instances via ON DELETE CASCADE
- All 13 stack routes now accessible at /api/v1/stacks/*

---
*Phase: 02-stack-crud*
*Completed: 2026-02-03*
