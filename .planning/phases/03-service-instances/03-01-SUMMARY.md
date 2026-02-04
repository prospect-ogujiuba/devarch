---
phase: 03-service-instances
plan: 01
subsystem: api
tags: [postgres, go, chi, instance-overrides, copy-on-write, soft-delete]

requires:
  - phase: 02-stack-crud
    provides: stacks table with soft-delete pattern, stack CRUD handlers
provides:
  - 7 instance override tables (ports, volumes, env_vars, labels, domains, healthchecks, config_files)
  - Instance CRUD API handler with create/list/get/update/delete/duplicate/rename endpoints
  - Routes under /stacks/{name}/instances/{instance}
  - Soft-delete pattern on service_instances
affects: [03-02, 03-03, 06-plan-apply]

tech-stack:
  added: []
  patterns:
    - "Copy-on-write override schema: separate override tables per resource type"
    - "Partial unique index for soft-delete: WHERE deleted_at IS NULL"
    - "Override count computed via subquery aggregation across all override tables"

key-files:
  created:
    - api/migrations/014_instance_overrides.up.sql
    - api/migrations/014_instance_overrides.down.sql
    - api/internal/api/handlers/instance.go
  modified:
    - api/internal/api/routes.go

key-decisions:
  - "Override tables match service table structure exactly (instance_ports mirrors service_ports)"
  - "Partial unique index on service_instances for soft-delete support"
  - "Container name pattern: devarch-{stack}-{instance}"
  - "Override count aggregated from all 7 tables in single query"

patterns-established:
  - "Instance handler follows stack handler patterns (getStackByName helper, response types)"
  - "Duplicate uses transaction to atomically copy instance + all overrides"
  - "Rename is direct UPDATE (instances are DB records, not containers yet)"

duration: 5min
completed: 2026-02-03
---

# Phase 03 Plan 01: Instance Override Schema & CRUD Summary

**7 override tables with ON DELETE CASCADE, instance CRUD handler with soft-delete and duplicate/rename support**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-03T21:46:10Z
- **Completed:** 2026-02-03T21:51:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Migration 014 creates 7 instance override tables mirroring service resource structure
- Instance CRUD handler with all lifecycle operations (create, list, get, update, delete, duplicate, rename)
- Routes wired under /stacks/{name}/instances/{instance}
- Soft-delete support via deleted_at column and partial unique index
- Override count computed from all tables for instance list/detail views

## Task Commits

Each task was committed atomically:

1. **Task 1: Migration 014 — instance override tables** - `f7bfe556` (feat)
2. **Task 2: Instance CRUD handler + route wiring** - `f2c449c2` (feat)

## Files Created/Modified
- `api/migrations/014_instance_overrides.up.sql` - 7 override tables, service_instances columns, partial unique index
- `api/migrations/014_instance_overrides.down.sql` - Rollback migration
- `api/internal/api/handlers/instance.go` - InstanceHandler with 8 endpoints
- `api/internal/api/routes.go` - Routes wired under /stacks/{name}/instances

## Decisions Made
- Override tables mirror service tables exactly (consistency, type safety, efficient queries)
- Partial unique index WHERE deleted_at IS NULL enables soft-delete while preventing duplicates
- Container name follows devarch-{stack}-{instance} pattern (consistent with existing container naming)
- Override count computed via subquery sum in single query (avoids N+1, shows override density)
- Duplicate copies all override records in transaction (atomic, all-or-nothing)
- Rename is direct UPDATE on instance_id + container_name (instances are DB records at this phase)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

Port conflict on 5433 during compose startup - used existing running database instead. No impact on execution.

## Next Phase Readiness

Override tables ready for data. Plan 02 will add PUT endpoints for modifying overrides (ports, volumes, env_vars, labels, domains, healthcheck, config_files).

Instance CRUD foundation complete. Next: override editing API + effective config resolver.

---
*Phase: 03-service-instances*
*Completed: 2026-02-03*
