---
phase: 12-compose-generator-parity
plan: 01
subsystem: compose
tags: [go, yaml, compose, generator, postgres]

# Dependency graph
requires:
  - phase: 10-v11-migrations
    provides: service_env_files, service_networks, service_config_mounts tables
  - phase: 11-parser-importer-updates
    provides: parser populates new tables during import
provides:
  - Generator queries service_env_files, service_networks, service_config_mounts
  - env_file emitted as YAML list
  - Networks DB-sourced (no hardcoded fallback)
  - Config mounts merged into volumes with materialized paths
affects: [13-nginx-subdomain-routing, 14-dashboard-field-exposure, 15-parity-verification]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Config mount source path resolution through MaterializeConfigFiles output"
    - "DB-sourced networks only (no fallback to default network)"
    - "env_file always emitted as YAML list"

key-files:
  created: []
  modified:
    - api/internal/compose/generator.go
    - api/internal/compose/stack.go
    - api/pkg/models/models.go

key-decisions:
  - "DB-sourced networks only - no hardcoded fallback maintains single source of truth"
  - "Config mounts use materialized paths from MaterializeConfigFiles output"
  - "env_file always emitted as YAML list for consistency"

patterns-established:
  - "Generator queries 3 new tables (service_env_files, service_networks, service_config_mounts) via loadServiceRelations"
  - "Stack generator follows instance-override-then-template pattern for 3 new fields"
  - "Config mount source paths resolve through MaterializeConfigFiles() output location"

# Metrics
duration: 3min
completed: 2026-02-09
---

# Phase 12 Plan 01: Generator Updates Summary

**Both generators query service_env_files, service_networks, service_config_mounts and emit as env_file YAML list, DB-sourced networks, and config volume mounts**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-09T21:06:51Z
- **Completed:** 2026-02-09T21:10:01Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Single-service generator queries and emits 3 new table fields
- Stack generator queries and emits 3 new table fields with instance-override pattern
- Config mounts merged into volumes section with materialized path resolution

## Task Commits

Each task was committed atomically:

1. **Task 1: Add env_file, networks, config mount emission to single-service generator** - `0272975` (feat)
2. **Task 2: Add env_file, networks, config mount emission to stack generator** - `d03fdb0` (feat)

## Files Created/Modified
- `api/internal/compose/generator.go` - Added EnvFile field, queries for service_env_files/networks/config_mounts, DB-sourced network emission, loadConfigMounts helper
- `api/internal/compose/stack.go` - Added EnvFile field, loadEffectiveEnvFiles/Networks/ConfigMounts methods, DB-sourced network map building, config mount merging into volumes
- `api/pkg/models/models.go` - Added EnvFiles and Networks fields to Service struct

## Decisions Made

**1. DB-sourced networks only (no hardcoded fallback)**
- Generator emits exactly what's in service_networks table
- Maintains single source of truth principle
- Importer is sole data path (decision from 10-02)

**2. Config mount source path resolution through MaterializeConfigFiles output**
- For mounts with resolved config_file_id: source becomes compose/{category}/{service}/{file_path}
- For mounts with NULL config_file_id: use raw source_path from DB
- Stack paths: compose/stacks/{stackName}/{instanceID}/{file_path}
- Ensures generated compose references actual materialized files

**3. env_file always emitted as YAML list**
- Simpler code path (no string-vs-list branching)
- Semantically identical to Docker/Podman
- Consistent with Phase 14 Dashboard rendering

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Generators emit all 3 new fields correctly
- Ready for Phase 13 (nginx subdomain routing - read-only, no conflict)
- Ready for Phase 14 (Dashboard field exposure - needs these queries as reference)
- Ready for Phase 15 (parity verification - will test these emissions)

---
*Phase: 12-compose-generator-parity*
*Completed: 2026-02-09*
