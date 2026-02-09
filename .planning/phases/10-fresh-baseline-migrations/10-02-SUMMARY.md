---
phase: 10-fresh-baseline-migrations
plan: 02
subsystem: database
tags: [postgres, migrations, sql, stacks, instances, wiring, soft-delete]

# Dependency graph
requires:
  - phase: 10-01
    provides: "Migrations 001-005 covering categories, services, service config"
provides:
  - "Migrations 006-009: stacks with soft delete, service instances, all instance overrides, wiring contracts, sync state, runtime state, performance indexes"
  - "Stacks table with source column and partial unique index (soft delete)"
  - "Service instances with description and soft delete"
  - "All 9 instance override tables including resource_limits"
  - "Wiring contracts (service_exports, service_import_contracts, service_instance_wires)"
  - "Runtime state tables (container_states, container_metrics)"
  - "Performance indexes (BRIN, autovacuum tuning)"
affects: [10-03-schema-validation, 10-04-migration-runner, API-rebuild]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Soft delete with partial unique indexes (WHERE deleted_at IS NULL)"
    - "Multi-table FK wiring (service_instance_wires references 4 tables)"
    - "Encryption-ready env vars (encrypted_value + encryption_version)"

key-files:
  created:
    - api/migrations/new_006_stacks_core.up.sql
    - api/migrations/new_006_stacks_core.down.sql
    - api/migrations/new_007_instance_overrides.up.sql
    - api/migrations/new_007_instance_overrides.down.sql
    - api/migrations/new_008_wiring_contracts_sync_security.up.sql
    - api/migrations/new_008_wiring_contracts_sync_security.down.sql
    - api/migrations/new_009_performance_indexes.up.sql
    - api/migrations/new_009_performance_indexes.down.sql
  modified: []

key-decisions:
  - "No sync_state seed data (importer is sole data path)"
  - "projects.stack_id FK added in 006 (not separate migration)"
  - "009 contains only non-inline indexes and autovacuum tuning"

patterns-established:
  - "Partial unique indexes for soft delete: CREATE UNIQUE INDEX ... WHERE deleted_at IS NULL"
  - "Down files use DROP IF EXISTS without CASCADE (explicit, testable)"
  - "Migration 009 is performance-only: indexes and autovacuum settings"

# Metrics
duration: 2min
completed: 2026-02-09
---

# Phase 10 Plan 02: Stacks & Instance Overrides Summary

**Fresh baseline migrations 006-009: stacks with soft delete, 9 instance override tables, wiring contracts, runtime state, BRIN indexes**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-09T18:12:14Z
- **Completed:** 2026-02-09T18:13:45Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Migrations 006-009 written in final form (no ALTER TABLE except projects FK)
- Stacks table with soft delete, source column, partial unique index
- Service instances with description, soft delete, partial unique index
- All 9 instance override tables created including resource_limits and encryption columns
- Wiring contracts (exports, imports, instance_wires) with multi-table FK references
- Runtime state (container_states, container_metrics) and sync_state table
- Performance indexes: BRIN on container_metrics, autovacuum tuning

## Task Commits

Each task was committed atomically:

1. **Task 1: Write migrations 006-007 (stacks core, instance overrides)** - `a17bae3` (feat)
2. **Task 2: Write migrations 008-009 (wiring/sync/security, performance indexes)** - `11b7070` (feat)

## Files Created/Modified

**Created:**
- `api/migrations/new_006_stacks_core.up.sql` - Stacks + service_instances with soft delete, projects.stack_id FK
- `api/migrations/new_006_stacks_core.down.sql` - Clean drops in reverse order
- `api/migrations/new_007_instance_overrides.up.sql` - 9 instance override tables (ports, volumes, env_vars, labels, domains, healthchecks, config_files, dependencies, resource_limits)
- `api/migrations/new_007_instance_overrides.down.sql` - Reverse order drops
- `api/migrations/new_008_wiring_contracts_sync_security.up.sql` - Wiring (exports, imports, wires), sync_state, runtime state
- `api/migrations/new_008_wiring_contracts_sync_security.down.sql` - Reverse order drops
- `api/migrations/new_009_performance_indexes.up.sql` - BRIN index, last_synced_at/created_at indexes, autovacuum tuning
- `api/migrations/new_009_performance_indexes.down.sql` - Index drops and RESET autovacuum

## Decisions Made

1. **No sync_state seed data** - Per phase decision, importer is sole data path; sync_state initialization is runtime concern
2. **projects.stack_id FK in 006** - Added in same migration that creates stacks table rather than separate migration
3. **009 is performance-only** - Contains only non-inline indexes and autovacuum tuning, no table creation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Migrations 001-009 complete (9 migration pairs, 18 files)
- Ready for schema validation (10-03)
- All tables in final form with proper indexes, constraints, soft delete patterns
- No seed data per phase decision

**Blocker:** None

---
*Phase: 10-fresh-baseline-migrations*
*Completed: 2026-02-09*
