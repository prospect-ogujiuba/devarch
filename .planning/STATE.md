# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-09)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 12 - Compose Generator Parity

## Current Position

Phase: 12 of 15 (Compose Generator Parity)
Plan: 2 of 2 (phase 12 complete)
Status: Phase complete
Last activity: 2026-02-09 — Completed 12-02-PLAN.md

Progress: [████████░░] 80% (12/15 phases complete, 34/TBD plans total)

## Performance Metrics

**Velocity:**
- Total plans completed: 30 (from v1.0)
- Average duration: ~3.2 minutes per plan (v1.0)
- Total execution time: ~1.6 hours (v1.0)

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (1-9) | 30 | ~1.6h | ~3.2min |
| v1.1 (10-15) | 5 | ~61min | ~12min |

**Recent Trend:**
- Phase 10 complete (3 plans, 53min total)
- Phase 11 complete (2 plans, 9min total)
- Phase 12 complete (2 plans, 6min total)
- v1.0 shipped successfully on 2026-02-09

*Updated after each plan completion*

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

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-09
Stopped at: Completed 12-02-PLAN.md (phase 12 complete)
Resume file: None
Next: Phase 13 - Nginx Subdomain Routing
