---
phase: 10-fresh-baseline-migrations
plan: 03
subsystem: database
tags: [migrations, schema, postgresql, baseline, verification]

# Dependency graph
requires:
  - phase: 10-01
    provides: "Baseline migrations 001-005 with new_ prefix"
  - phase: 10-02
    provides: "Migrations 006-009 with new_ prefix"
provides:
  - "Clean migrations directory: 9 domain-separated pairs (001-009)"
  - "Verified migrate-up from zero database succeeds"
  - "23 old migration pairs deleted permanently"
  - "Schema with 39 tables, all new columns/tables verified"
affects: [phase-11-api-rebuild, phase-12-importer, phase-13-dashboard]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Fresh baseline over incremental patches"
    - "Domain-separated migrations"

key-files:
  created: []
  modified:
    - api/migrations/001_categories_services.up.sql
    - api/migrations/001_categories_services.down.sql
    - api/migrations/002_service_runtime_shape.up.sql
    - api/migrations/002_service_runtime_shape.down.sql
    - api/migrations/003_service_config_model.up.sql
    - api/migrations/003_service_config_model.down.sql
    - api/migrations/004_registry_and_images.up.sql
    - api/migrations/004_registry_and_images.down.sql
    - api/migrations/005_projects_and_project_services.up.sql
    - api/migrations/005_projects_and_project_services.down.sql
    - api/migrations/006_stacks_core.up.sql
    - api/migrations/006_stacks_core.down.sql
    - api/migrations/007_instance_overrides.up.sql
    - api/migrations/007_instance_overrides.down.sql
    - api/migrations/008_wiring_contracts_sync_security.up.sql
    - api/migrations/008_wiring_contracts_sync_security.down.sql
    - api/migrations/009_performance_indexes.up.sql
    - api/migrations/009_performance_indexes.down.sql

key-decisions:
  - "Verified new_ migrations in isolation before deleting old ones"
  - "Double verification: before and after rename to final names"

patterns-established:
  - "Test migrations against fresh DB before committing cleanup"
  - "Verify table count and key columns after migrate-up"

# Metrics
duration: 48min
completed: 2026-02-09
---

# Phase 10 Plan 03: Fresh Baseline Migrations Swap Summary

**Replaced 23 old patch migrations with 9 fresh domain-separated migrations, verified migrate-up from zero creates 39 tables**

## Performance

- **Duration:** 48 minutes
- **Started:** 2026-02-09T13:50:07Z
- **Completed:** 2026-02-09T14:37:50Z
- **Tasks:** 1
- **Files modified:** 64 (46 deleted, 18 renamed)

## Accomplishments

- Verified new_ migrations work in isolation (moved old files aside temporarily)
- Created 39 tables successfully from fresh database
- Deleted all 23 old migration pairs (001-023 patterns, 46 files total)
- Renamed 9 new_ migration pairs to final names (001-009)
- Re-verified: migrate-up from zero succeeds with final names
- Verified key columns: container_name_template, service_env_files, service_networks, service_config_mounts

## Task Commits

Each task was committed atomically:

1. **Task 1: Verify new migrations from zero, then swap** - `7800ff0` (feat)

## Files Created/Modified

**Renamed (18 files):**
- `api/migrations/001_categories_services.{up,down}.sql` - Renamed from new_001
- `api/migrations/002_service_runtime_shape.{up,down}.sql` - Renamed from new_002
- `api/migrations/003_service_config_model.{up,down}.sql` - Renamed from new_003
- `api/migrations/004_registry_and_images.{up,down}.sql` - Renamed from new_004
- `api/migrations/005_projects_and_project_services.{up,down}.sql` - Renamed from new_005
- `api/migrations/006_stacks_core.{up,down}.sql` - Renamed from new_006
- `api/migrations/007_instance_overrides.{up,down}.sql` - Renamed from new_007
- `api/migrations/008_wiring_contracts_sync_security.{up,down}.sql` - Renamed from new_008
- `api/migrations/009_performance_indexes.{up,down}.sql` - Renamed from new_009

**Deleted (46 files):**
- All old migrations 001-023 (both up and down files)

## Decisions Made

1. **Isolated verification before deletion** - Moved old migrations to /tmp before testing new_ files to ensure they work independently on fresh DB
2. **Double verification** - Tested both with new_ prefix and after rename to final names, ensuring no issues with naming

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Verification Results

All verification criteria met:
- ✓ `ls api/migrations/*.sql | wc -l` = 18
- ✓ All files start with 001-009 pattern
- ✓ No new_ prefix files remain
- ✓ No old 010-023 files remain
- ✓ Migrate-up from zero succeeded (before rename)
- ✓ Migrate-up from zero succeeded (after rename)
- ✓ 39 tables created
- ✓ container_name_template column exists in services
- ✓ service_env_files table exists
- ✓ service_networks table exists
- ✓ service_config_mounts table exists

## Next Phase Readiness

**Phase 10 complete** - All success criteria met (SCHM-01 through SCHM-12):
- Fresh baseline migrations work from zero
- Domain separation achieved
- No ALTER patches needed
- All tables in final form
- Soft delete patterns established
- Performance indexes in place

**Enables:**
- Phase 11: API rebuild can reference clean schema
- Phase 12: Importer can work with final-form tables
- Phase 13: Dashboard can query stable schema

**Blockers removed:**
- Migration fragility from 23 patches
- Schema inconsistencies
- Missing columns/tables

**Handoff notes:**
- api/migrations/ contains exactly 18 files
- Running migrate-up from zero is now safe and tested
- No seed data in migrations (importer is data path)
- All FKs properly ordered, no CASCADE in down files

---

*Phase: 10-fresh-baseline-migrations*
*Completed: 2026-02-09*
