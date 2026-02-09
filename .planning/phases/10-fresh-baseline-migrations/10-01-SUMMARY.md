---
phase: 10-fresh-baseline-migrations
plan: 01
subsystem: database
tags: [migrations, schema, postgresql, baseline]
requires: []
provides: [baseline-migrations-001-005, final-form-tables]
affects: [10-02, 10-03]
tech-stack:
  added: []
  patterns: [domain-separated-ddl, fresh-baseline-migrations]
key-files:
  created:
    - api/migrations/new_004_registry_and_images.up.sql
    - api/migrations/new_004_registry_and_images.down.sql
    - api/migrations/new_005_projects_and_project_services.up.sql
    - api/migrations/new_005_projects_and_project_services.down.sql
  modified: []
decisions:
  - slug: migrations-001-003-pre-existing
    summary: Migrations 001-003 already created by plan 10-02
    impact: Skipped Task 1, only created 004-005
metrics:
  duration: 151s
  completed: 2026-02-09
---

# Phase 10 Plan 01: Fresh Baseline Migrations (001-005) Summary

Fresh baseline migrations 001-005 covering categories, services, runtime shape, config model, registry/images, and projects in final form with domain separation.

## What Was Built

Created migrations 004-005 (registry/images and projects). Migrations 001-003 were already present from a previous execution.

**Migration 004: Registry & Images**
- Registries table with rate limiting columns
- Images with repository metadata
- Image tags with digests and sync timestamps
- Image architectures for multi-platform support
- Vulnerabilities (CVE tracking)
- Image-tag-vulnerability junction table
- NO seed data (per PARS-07)

**Migration 005: Projects & Project Services**
- Projects table with all accumulated columns (compose_path, service_count, stack_id)
- stack_id column as plain INT (FK constraint deferred to 006)
- Project services for scanning/linking
- Indexes on project_type, language, stack_id

**Key Characteristics:**
- All tables in final form (no ALTER patches needed)
- Down files use precise DROP statements in reverse dependency order
- NO CASCADE in down files (proves correct dependency graph)
- container_name_template column in services (001)
- service_env_files and service_networks tables (002)
- service_config_mounts with nullable FK (003)

## Decisions Made

**1. Skipped creating 001-003 (already exist)**
- Found migrations 001-003 already committed (from plan 10-02 execution)
- Files matched exactly what would have been created
- Continued with Task 2 only (004-005)
- Classification: Deviation Rule 3 (work blocking completion already done)

## Implementation Notes

**Schema Design Patterns:**
- Domain-separated: Each migration creates one domain's tables
- Final-form: All columns present from creation, no incremental ALTERs
- Precise drops: Down files drop only domain objects in reverse dependency order

**Deferred Constraints:**
- projects.stack_id has no FK constraint yet (stacks table created in 006)
- Comment documents deferral reason

## Deviations from Plan

**1. [Rule 3 - Blocking] Migrations 001-003 pre-existing**
- **Found during:** Task 1 start
- **Issue:** Migrations 001-003 already committed by plan 10-02
- **Action:** Verified files match expected DDL, skipped Task 1
- **Files affected:** new_001, new_002, new_003 (all 6 files)
- **Commit:** a17bae3 (from 10-02)

## Files Changed

**Created (4 files):**
- `api/migrations/new_004_registry_and_images.up.sql` - Registry and image domain tables
- `api/migrations/new_004_registry_and_images.down.sql` - Reverse drops for 004
- `api/migrations/new_005_projects_and_project_services.up.sql` - Projects scanning layer
- `api/migrations/new_005_projects_and_project_services.down.sql` - Reverse drops for 005

**Pre-existing (6 files from 10-02):**
- `api/migrations/new_001_categories_services.up.sql`
- `api/migrations/new_001_categories_services.down.sql`
- `api/migrations/new_002_service_runtime_shape.up.sql`
- `api/migrations/new_002_service_runtime_shape.down.sql`
- `api/migrations/new_003_service_config_model.up.sql`
- `api/migrations/new_003_service_config_model.down.sql`

## Test Strategy

Verification performed:
- ✓ All 10 migration files exist (001-005 up+down)
- ✓ No ALTER TABLE in any up file
- ✓ No INSERT/seed data in any file
- ✓ No CASCADE in any down file
- ✓ container_name_template in new_001
- ✓ service_env_files and service_networks in new_002
- ✓ service_config_mounts with nullable config_file_id in new_003
- ✓ No seed data in new_004
- ✓ stack_id column in new_005

Next verification: Phase 15 will test migrate-down+up roundtrip.

## Next Phase Readiness

**Enables:**
- Phase 10 Plan 02: Can reference tables from 001-005
- Phase 10 Plan 03: Full migration set ready for cleanup/rename

**Blocks removed:**
- None (foundation layer complete)

**New blockers:**
- None

**Handoff notes:**
- Migrations use `new_` prefix (renamed in plan 03)
- Old migrations still present (deleted in plan 03)
- Projects.stack_id awaits FK constraint in 006

## Metrics

- **Duration:** 151 seconds (~2.5 minutes)
- **Files created:** 4
- **Lines of SQL:** ~130
- **Commits:** 1
- **Deviations:** 1 (pre-existing work)

---

*Completed: 2026-02-09*
*Phase: 10-fresh-baseline-migrations*
*Plan: 01 of 03*
