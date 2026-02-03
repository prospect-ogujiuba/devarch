---
phase: 02-stack-crud
plan: 01
subsystem: api
tags: [go, postgresql, chi, soft-delete, crud]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: container.ValidateName for DNS-safe name validation
provides:
  - Migration 013 with soft-delete support (deleted_at + partial unique index)
  - StackHandler with Create/List/Get/Update/Delete operations
  - Prescriptive validation errors using container.ValidateName
affects: [02-02, 02-03, 02-04, 02-05, stack-api, stack-ui]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Soft-delete with deleted_at timestamp and partial unique indexes
    - Name immutability (rename requires clone + soft-delete)
    - Active-only queries with WHERE deleted_at IS NULL

key-files:
  created:
    - api/internal/api/handlers/stack.go
  modified:
    - api/migrations/013_stacks_instances.up.sql
    - api/migrations/013_stacks_instances.down.sql

key-decisions:
  - "Soft-delete pattern using deleted_at with partial unique index on active stacks only"
  - "Stack name is immutable identifier (used in URL routes)"
  - "Running count placeholder set to 0 until Phase 3+ wires container client queries"

patterns-established:
  - "Partial unique indexes: CREATE UNIQUE INDEX name ON stacks(name) WHERE deleted_at IS NULL"
  - "Active stack queries always include WHERE deleted_at IS NULL filter"
  - "container.ValidateName provides prescriptive errors with Slugify suggestions"

# Metrics
duration: 1min
completed: 2026-02-03
---

# Phase 2 Plan 1: Stack CRUD API Summary

**Soft-delete migration and StackHandler with Create/List/Get/Update/Delete using container.ValidateName for prescriptive errors**

## Performance

- **Duration:** ~1 min
- **Started:** 2026-02-03T21:52:31Z
- **Completed:** 2026-02-03T21:53:47Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Migration 013 updated with soft-delete support (deleted_at column + partial unique index)
- StackHandler implemented with full CRUD operations following ServiceHandler patterns
- Name validation using existing container.ValidateName with prescriptive error messages

## Task Commits

Each task was committed atomically:

1. **Task 1: Update migration 013 for soft-delete** - `d394839` (feat)
2. **Task 2: Implement StackHandler with core CRUD** - `eb63c53` (feat)

## Files Created/Modified
- `api/migrations/013_stacks_instances.up.sql` - Added deleted_at column, partial unique index for active stacks, deleted_at index
- `api/migrations/013_stacks_instances.down.sql` - Updated to drop indexes before tables
- `api/internal/api/handlers/stack.go` - New StackHandler with Create/List/Get/Update/Delete operations

## Decisions Made
- **Soft-delete implementation:** Used deleted_at TIMESTAMPTZ with partial unique index (WHERE deleted_at IS NULL) to allow name reuse after deletion
- **Name immutability:** Stack name is the URL identifier, making it immutable. Rename will be clone + soft-delete (Phase 2 later plan)
- **Running count placeholder:** Set to 0 in List/Get operations until Phase 3+ implements instance management and container queries
- **Validation reuse:** Used existing container.ValidateName for consistent prescriptive error messages with Slugify suggestions

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Stack CRUD API foundation complete. Ready for:
- Phase 2 Plan 2: Stack routing integration (wire StackHandler into chi router)
- Phase 2 Plan 3+: Dashboard UI, enable/disable, clone, trash operations

**Blockers:** None

**Notes:**
- Migration 013 already existed from Phase 1, was updated with soft-delete support
- Container stopping logic in Delete is placeholder (Phase 3 will implement actual instance management)
- Instance counts work via LEFT JOIN to service_instances table

---
*Phase: 02-stack-crud*
*Completed: 2026-02-03*
