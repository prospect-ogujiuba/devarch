# Roadmap: DevArch

## Milestones

- ✅ **v1.0 Stacks & Instances** — Phases 1-9 (shipped 2026-02-09)
- 🚧 **v1.1 Schema Reconciliation** — Phases 10-15 (in progress)

## Phases

<details>
<summary>✅ v1.0 Stacks & Instances (Phases 1-9) — SHIPPED 2026-02-09</summary>

- [x] Phase 1: Foundation & Guardrails (2/2 plans) — 2026-02-03
- [x] Phase 2: Stack CRUD (5/5 plans) — 2026-02-03
- [x] Phase 3: Service Instances (5/5 plans) — 2026-02-04
- [x] Phase 4: Network Isolation (2/2 plans) — 2026-02-06
- [x] Phase 5: Compose Generation (2/2 plans) — 2026-02-07
- [x] Phase 6: Plan/Apply Workflow (3/3 plans) — 2026-02-07
- [x] Phase 7: Export/Import & Bootstrap (4/4 plans) — 2026-02-08
- [x] Phase 8: Service Wiring (4/4 plans) — 2026-02-08
- [x] Phase 9: Secrets & Resources (3/3 plans) — 2026-02-08

Full details: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)

</details>

### 🚧 v1.1 Schema Reconciliation (In Progress)

**Milestone Goal:** Replace fragmented 23-migration chain with clean, domain-separated fresh baseline — 1:1 legacy parity, no seeds, no patch artifacts.

#### Phase 10: Fresh Baseline Migrations

**Goal**: Database schema recreated from scratch with domain-separated DDL files in final form
**Depends on**: Nothing (migration reset)
**Requirements**: SCHM-01, SCHM-02, SCHM-03, SCHM-04, SCHM-05, SCHM-06, SCHM-07, SCHM-08, SCHM-09, SCHM-10, SCHM-11, SCHM-12
**Success Criteria** (what must be TRUE):
  1. Fresh migrate-up from zero database succeeds with no errors
  2. All 9 domain-separated migration files (001-009) exist with final-form DDL
  3. Each migration has matching down file that drops its objects in reverse dependency order
  4. Old migrations (001-023) deleted from repository
  5. Schema includes new columns (container_name_template, service_config_mounts table, service_env_files, service_networks)
**Plans:** 3 plans

Plans:
- [x] 10-01-PLAN.md — Write fresh migrations 001-005 (categories, services, runtime, config, registry, projects)
- [x] 10-02-PLAN.md — Write fresh migrations 006-009 (stacks, instance overrides, wiring/sync, indexes)
- [x] 10-03-PLAN.md — Verify migrate-up, delete old 23 migrations, rename, re-verify

#### Phase 11: Parser & Importer Updates

**Goal**: Import logic preserves all compose constructs into new schema with correct provenance
**Depends on**: Phase 10 (new schema must exist)
**Requirements**: PARS-01, PARS-02, PARS-03, PARS-04, PARS-05, PARS-06, PARS-07
**Success Criteria** (what must be TRUE):
  1. Parser preserves env_file (string and list forms) with order into service_env_files table
  2. Parser preserves container_name into container_name_template column
  3. Parser preserves explicit network attachments into service_networks table
  4. Config import derives mount source from actual volume paths, not assumptions
  5. Config import handles shared/mismatched mappings (php→nginx, python→supervisord, etc.)
  6. No seed data in any migration — catalog loaded exclusively through importer
**Plans:** 2 plans

Plans:
- [x] 11-01-PLAN.md — Parse env_file, container_name, networks + write to DB tables
- [x] 11-02-PLAN.md — Provenance-aware config mount classification replacing rewriteConfigPaths

#### Phase 12: Compose Generator Parity

**Goal**: Generated compose output matches legacy 1:1 using new schema
**Depends on**: Phase 11 (parser must populate new schema)
**Requirements**: GENR-01, GENR-02, GENR-03, GENR-04
**Success Criteria** (what must be TRUE):
  1. Compose generator emits env_files section from service_env_files table
  2. Compose generator emits explicit networks section from service_networks table
  3. Compose generator emits config mounts with correct source/target from service_config_mounts table
  4. Generated compose for all 166+ services matches legacy output byte-for-byte (modulo whitespace/ordering)
**Plans:** 2 plans

Plans:
- [ ] 12-01-PLAN.md — Update both generators to emit env_file, networks, config mounts from new DB tables
- [ ] 12-02-PLAN.md — Round-trip parity verification tool (import-then-generate diff)

#### Phase 13: Import Scalability

**Goal**: Stack import handles large payloads without memory exhaustion
**Depends on**: Phase 11 (importer must use new parser)
**Requirements**: IMPT-01, IMPT-02, IMPT-03, IMPT-04, IMPT-05, IMPT-06
**Success Criteria** (what must be TRUE):
  1. Stack import uses streaming multipart read (no ParseMultipartForm buffering)
  2. Stack import endpoint has dedicated 256MB cap via STACK_IMPORT_MAX_BYTES
  3. All other endpoints retain global 10MB cap
  4. Bulk import uses prepared statements and batched upserts within transaction
  5. Import handles conflicts idempotently (same import twice succeeds)
**Plans:** 2 plans

Plans:
- [ ] 13-01-PLAN.md — Streaming multipart handler + route-level size cap
- [ ] 13-02-PLAN.md — Prepared statements + upsert idempotency

#### Phase 14: Dashboard Updates

**Goal**: Dashboard UI exposes new schema fields for user inspection and editing
**Depends on**: Phase 12 (API must serve new fields), Phase 13 (import flow stable)
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04
**Success Criteria** (what must be TRUE):
  1. Dashboard displays env_files list for services and instances
  2. Dashboard displays network attachments for services and instances
  3. Dashboard displays config mount provenance (source path, target path, read-only/mount type)
  4. Dashboard forms support adding/editing/removing env_files, networks, and config mounts
**Plans**: TBD

Plans:
- [ ] 14-01: [TBD during planning]

#### Phase 15: Validation & Parity

**Goal**: Legacy parity verified across all services, boundary cases tested
**Depends on**: Phase 12 (generator must be updated), Phase 13 (import must scale), Phase 14 (UI must be functional)
**Requirements**: VALD-01, VALD-02, VALD-03, VALD-04
**Success Criteria** (what must be TRUE):
  1. Legacy parity verified for ports, volumes, env vars, env files, deps, healthchecks, labels, networks, config mappings across all services
  2. Golden service parity explicitly verified: php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api
  3. Import boundary test: 200MB payload succeeds
  4. Import boundary test: 300MB payload rejected with clear error message
**Plans**: TBD

Plans:
- [ ] 15-01: [TBD during planning]

## Progress

**Execution Order:**
Phases execute in numeric order: 10 → 11 → 12 → 13 → 14 → 15

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 10. Fresh Baseline Migrations | v1.1 | 3/3 | ✓ Complete | 2026-02-09 |
| 11. Parser & Importer Updates | v1.1 | 2/2 | ✓ Complete | 2026-02-09 |
| 12. Compose Generator Parity | v1.1 | 0/? | Not started | - |
| 13. Import Scalability | v1.1 | 0/? | Not started | - |
| 14. Dashboard Updates | v1.1 | 0/? | Not started | - |
| 15. Validation & Parity | v1.1 | 0/? | Not started | - |

---
*Created: 2026-02-03*
*Last updated: 2026-02-09 — Phase 11 complete*
