---
phase: 11-parser-importer-updates
plan: 02
subsystem: parser-importer
tags: [compose-parser, importer, config-mounts, provenance, cross-service]

# Dependency graph
requires:
  - phase: 11-parser-importer-updates
    plan: 01
    provides: "Parser and importer for env_files, container_name, networks"
provides:
  - "ParsedConfigMount type with owner and relpath extraction"
  - "Config mount provenance links mounts to owning service's config_file via FK"
  - "Cross-service config mount resolution (php→nginx, blackbox-exporter→prometheus)"
affects: [12-generator-updates, compose-generation, config-management]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Parse-time classification over runtime guessing", "Post-import FK resolution for cross-references"]

key-files:
  created: []
  modified:
    - api/internal/compose/parser.go
    - api/internal/compose/importer.go
    - api/cmd/import/main.go

key-decisions:
  - "Config volumes classified by config/ prefix in source path, not by assumption"
  - "NULL config_file_id for missing files, resolved in post-import pass"
  - "ResolveConfigMountLinks() called after ImportAllConfigFiles() completes"

patterns-established:
  - "classifyConfigMounts() extracts owner from path: config/{owner}/{relpath}"
  - "Config mounts removed from Volumes slice, tracked in ConfigMounts field"
  - "Post-import FK resolution pattern for cross-service references"

# Metrics
duration: 4min
completed: 2026-02-09
---

# Phase 11 Plan 02: Config Mount Provenance Summary

**ParsedConfigMount type derives ownership from actual volume paths; cross-service FK linking resolves php→nginx, blackbox-exporter→prometheus mappings**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-09T20:19:44Z
- **Completed:** 2026-02-09T20:23:44Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- ParsedConfigMount type added with ConfigOwner and ConfigRelPath fields
- Config volumes classified at parse time by config/ prefix in source path
- Config mounts written to service_config_mounts table with source_path, target_path, readonly
- Cross-service provenance resolved via config_file_id FK using ResolveConfigMountLinks()
- rewriteConfigPaths() removed entirely, replaced by parse-time classification

## Task Commits

Each task was committed atomically:

1. **Task 1: Add ParsedConfigMount type and classify config volumes in parser** - `d0917da` (feat)
2. **Task 2: Provenance-aware config mount import with cross-service FK linking** - `c7ed444` (feat)

## Files Created/Modified
- `api/internal/compose/parser.go` - Added ParsedConfigMount type, classifyConfigMounts() function, ConfigMounts field on ParsedService
- `api/internal/compose/importer.go` - Removed rewriteConfigPaths/rewriteConfigSource, added config mount import to importService(), added ResolveConfigMountLinks()
- `api/cmd/import/main.go` - Call ResolveConfigMountLinks() after ImportAllConfigFiles()

## Decisions Made

None - followed plan as specified

## Deviations from Plan

None - plan executed exactly as written

## Issues Encountered

None - full import of 173 compose files + 38 config files succeeded on first attempt

## Verification Results

Full import succeeded with correct data:
- 34 total config mounts imported
- 23 config mounts with resolved config_file_id FK
- 11 config mounts with NULL config_file_id (missing files, warnings logged)
- Cross-service mappings verified:
  - blackbox-exporter→prometheus/blackbox.yml (config_file_id=21) ✓
  - php→nginx/custom/http.conf (NULL, file missing) ✓
  - python→supervisord/supervisord.conf (NULL, service missing) ✓
- Regular volumes (apps, scripts, named volumes) NOT in service_config_mounts ✓
- rewriteConfigPaths and rewriteConfigSource removed from codebase ✓

## User Setup Required

None - no external service configuration required

## Next Phase Readiness

Parser and importer now handle config mount provenance correctly. Ready for generator updates to materialize config directories at runtime.

---
*Phase: 11-parser-importer-updates*
*Completed: 2026-02-09*
