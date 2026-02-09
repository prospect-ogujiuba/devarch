# Phase 10: Fresh Baseline Migrations - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Replace 23 incremental migrations with 9 domain-separated fresh DDL files. Each migration creates tables in final form (no ALTER patches). Pure schema work — no parser, generator, or dashboard changes. Old migrations deleted after new ones verified.

</domain>

<decisions>
## Implementation Decisions

### Domain Grouping (9 files, dependency order)
- 001: categories + services (incl. container_name_template) — core entities everything references
- 002: service runtime shape — ports, volumes, env_vars, env_files, deps, healthchecks, labels, domains, networks
- 003: config model — service_config_files + service_config_mounts (new provenance concept)
- 004: registry & images — registry, image, tag, vulnerability tables
- 005: projects & project_services — scanning/link layer
- 006: stacks core — stacks + service_instances with non-collision invariants
- 007: instance overrides — all instance override tables + resource_limits
- 008: wiring, contracts, sync, security — wiring, sync, encryption tables
- 009: performance indexes — all non-inline indexes consolidated in one file

### New Column/Table Designs

**container_name_template** (on services table):
- `TEXT NULL` — NULL means "not specified", generator omits it
- Stores raw template from compose file, not computed value
- Phase 11 parser writes it, Phase 12 generator reads it

**service_env_files** (new table in 002):
- id SERIAL PK, service_id INT NOT NULL FK, path TEXT NOT NULL, sort_order INT NOT NULL DEFAULT 0
- UNIQUE(service_id, path)
- sort_order preserves compose ordering (later files override earlier)
- Phase 14 dashboard can reorder via sort_order

**service_networks** (new table in 002):
- id SERIAL PK, service_id INT NOT NULL FK, network_name TEXT NOT NULL
- UNIQUE(service_id, network_name)
- Captures explicit network attachments from compose, name only
- Separate from stack-level network isolation (Phase 4)

**service_config_mounts** (new table in 003):
- id SERIAL PK, service_id INT NOT NULL FK, config_file_id INT NULL FK -> service_config_files, source_path TEXT NOT NULL, target_path TEXT NOT NULL, readonly BOOLEAN NOT NULL DEFAULT false
- UNIQUE(service_id, target_path)
- Nullable FK handles shared/mismatched config cases (php->nginx, python->supervisord)
- Full provenance: where config comes from (source), where it mounts (target), read-only semantics

### Down Migration Strategy
- Precise DROP in reverse dependency order, NO CASCADE
- Each down file drops ONLY its domain's objects
- 009_down drops only non-inline performance indexes; inline constraints (PK, UNIQUE) die with their table
- Precise drops prove the dependency graph is correct — Phase 15 uses down+up roundtrip as validation

### Old Migration Removal
- Delete ALL 23 old migration pairs after new 9 verified
- No staged removal, no archiving — clean break
- No seed data in any new migration (PARS-07) — importer is sole data path
- Seeds from old 004 + 011 deleted with everything else

### Write Strategy (overlap handling)
- Write new migrations with `new_` prefix (e.g., new_001_categories_services.up.sql)
- Verify new migrations work from zero database
- Delete old 001-023 files
- Rename new files to drop prefix
- Re-verify migrate-up still works

### Claude's Discretion
- Exact column types for existing tables (infer from current migrations)
- Constraint naming conventions
- Index selection for 009 (pick from current schema + new tables)
- Transaction wrapping per migration file

</decisions>

<specifics>
## Specific Ideas

- All decisions made with forward-looking alignment to Phases 11-15 — column shapes designed so parser (11), generator (12), and dashboard (14) consume them without schema changes
- Dependency ordering in migration numbering ensures migrate-up always works sequentially
- Nullable FK on config_mounts specifically handles PARS-05 shared config cases

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 10-fresh-baseline-migrations*
*Context gathered: 2026-02-09*
