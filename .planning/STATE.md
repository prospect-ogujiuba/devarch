# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-09)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 15 - Validation & Parity

## Current Position

Phase: 15 of 15 (Validation & Parity)
Plan: 2 of 2
Status: Phase complete
Last activity: 2026-02-10 — Phase 15 Plan 02 complete

Progress: [██████████] 100% (15/15 phases complete, 41/TBD plans total)

## Performance Metrics

**Velocity:**
- Total plans completed: 30 (from v1.0)
- Average duration: ~3.2 minutes per plan (v1.0)
- Total execution time: ~1.6 hours (v1.0)

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (1-9) | 30 | ~1.6h | ~3.2min |
| v1.1 (10-15) | 9 | ~71min | ~7.9min |

**Recent Trend:**
- Phase 10 complete (3 plans, 53min total)
- Phase 11 complete (2 plans, 9min total)
- Phase 12 complete (2 plans, 6min total)
- Phase 13 complete (2 plans, 5min total)
- Phase 14 complete (3 plans, ~11min total)
- v1.0 shipped successfully on 2026-02-09

*Updated after each plan completion*

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 13 P01 | 2m22s | 2 | 3 |
| Phase 13 P02 | 2m47s | 1 | 1 |
| Phase 14 P01 | 3m19s | 2 | 8 |
| Phase 14 P02 | 2m14s | 2 | 6 |
| Phase 14 P03 | 5m14s | 2 | 5 |
| Phase 15 P01 | ~6m | 2 | 4 |
| Phase 15 P02 | 12m40s | 2 | 2 |
| Phase 15 P01 | 7m29s | 2 | 2 |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- v1.1: Fresh baseline over incremental migration — 23 patch migrations create fragility; clean slate is simpler to maintain
- v1.1: Domain-separated DDL files — Clean boundaries, each table created once in final form
- v1.1: Dedicated service_config_mounts table — Config provenance needs its own model, not overloaded config_files
- v1.1: Streaming multipart for large imports — ParseMultipartForm buffers entire body; streaming avoids OOM
- 10-01: Migrations 001-003 pre-existing from 10-02 — Skipped Task 1, created only 004-005
- 10-02: No sync_state seed data — Importer is sole data path; sync_state init is runtime concern
- 10-02: projects.stack_id FK in 006 — Added in same migration that creates stacks table
- 10-02: Migration 009 is performance-only — Contains only non-inline indexes and autovacuum tuning
- 10-03: Isolated verification before deletion — Moved old migrations to temp, tested new_ files independently
- 10-03: Double verification pattern — Tested before and after rename to ensure no naming issues
- 11-01: Store env_file paths as-is — No filesystem validation during import
- 11-01: Extract network names only — Ignore network config, store names in service_networks
- 11-01: sort_order preserves declaration order — service_env_files uses loop index for proper ordering
- 11-02: Config volumes classified by config/ prefix — Parse-time classification replaces runtime guessing
- 11-02: NULL config_file_id for missing files — Warnings logged, FKs resolved in post-import pass
- 11-02: ResolveConfigMountLinks() post-import — Called after ImportAllConfigFiles() completes
- 12-01: DB-sourced networks only — Generator emits exactly what's in service_networks; no hardcoded fallback maintains single source of truth
- 12-01: Config mount materialized paths — Mounts with resolved config_file_id use compose/{category}/{service}/{file_path}; NULL uses raw source_path
- 12-01: env_file always YAML list — Simpler code path, semantically identical, consistent Dashboard rendering
- 12-02: Volume comparison skips source path — Compares target + read_only only; source path resolution differences are artifacts, not semantic mismatches
- 12-02: Config mount validation via volume targets — Generator merges config mounts into volumes; validate via target presence + MaterializeConfigFiles paths
- [Phase 13]: Streaming multipart via multipart.NewReader eliminates full-body buffering
- [Phase 13]: io.LimitReader caps YAML part size, not multipart overhead
- [Phase 13]: Route-level MaxBodySize(256MB) overrides global 10MB via r.With()
- [Phase 13-02]: Use xmax = 0 for insert detection — PostgreSQL xmax indicates transaction ID of deleting/updating transaction; 0 means row was inserted
- [Phase 13-02]: Delete-then-reinsert for override tables — Lack natural unique keys; atomic within transaction, simpler than complex upserts
- [Phase 14-01]: Instance override tables mirror service tables — Created instance_env_files, instance_networks, instance_config_mounts following established override pattern
- [Phase 14-01]: ServiceConfigMount with optional config_file_id — Flexible model supports both DB-managed and external config paths
- [Phase 14-01]: Override detection in effective config — Boolean flags per field let dashboard highlight instance-specific overrides
- [Phase 14-03]: Tab order logical grouping — Groups related overrides (runtime, filesystem, network, config) for better navigation UX
- [Phase 14-03]: Config mounts show resolved/unresolved badge — Transparency in config_file_id linkage vs external paths
- [Phase 14-02]: Service-level component placement order — EnvFiles after Volumes, Networks after EnvFiles, ConfigMounts after Dependencies; groups related config types
- [Phase 14-02]: Config mount resolution badges — Green for resolved (config_file_id set), amber for unresolved (external path only); visual provenance feedback
- [Phase 15-02]: MaxBodySize scope matters — Chi r.Use() applies to all children; r.With() adds middleware on top. Multiple MaxBytesReader wrappers enforce smallest limit. Solution: register large-limit routes outside restrictive parent groups
- [Phase 15-01]: Whitelist zammad dependencies on external services — zammad references zammad-db and zammad-elasticsearch which are not standalone services; external dependencies provided at runtime
- [Phase 15-01]: Stderr for diagnostics, stdout for structured output — All importer warnings now use fmt.Fprintf(os.Stderr) to preserve clean JSON output

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-10
Stopped at: Phase 15 Plan 02 complete — verify-boundary validates import size boundaries
Resume file: None
Next: Phase 15 complete
