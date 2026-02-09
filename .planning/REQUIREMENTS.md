# Requirements: DevArch

**Defined:** 2026-02-09
**Core Value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.

## v1.1 Requirements

Requirements for Schema Reconciliation milestone. Each maps to roadmap phases.

### Schema

- [ ] **SCHM-01**: Fresh 001_categories_services migration creates categories and services tables in final form (incl. container_name_template)
- [ ] **SCHM-02**: Fresh 002_service_runtime_shape creates ports, volumes, env_vars, env_files, dependencies, healthchecks, labels, domains, networks tables
- [ ] **SCHM-03**: Fresh 003_service_config_model creates service_config_files and service_config_mounts tables with provenance semantics
- [ ] **SCHM-04**: Fresh 004_registry_and_images creates registry, image, tag, vulnerability tables
- [ ] **SCHM-05**: Fresh 005_projects_and_project_services creates project scanning/link tables
- [ ] **SCHM-06**: Fresh 006_stacks_core creates stacks and service_instances with non-collision invariants
- [ ] **SCHM-07**: Fresh 007_instance_overrides creates all instance override tables including resource_limits
- [ ] **SCHM-08**: Fresh 008_wiring_contracts_sync_security creates wiring, sync, and encryption support tables
- [ ] **SCHM-09**: Fresh 009_performance_indexes creates all non-inline indexes
- [ ] **SCHM-10**: Each migration has a matching down that drops only its domain objects in reverse dependency order
- [ ] **SCHM-11**: Old migrations (001-023) deleted from repository
- [ ] **SCHM-12**: Fresh migrate-up from zero succeeds with no errors

### Parser/Importer

- [ ] **PARS-01**: Parser preserves env_file (string and list forms) with order into service_env_files
- [ ] **PARS-02**: Parser preserves container_name into container_name_template column
- [ ] **PARS-03**: Parser preserves explicit network attachments into service_networks
- [ ] **PARS-04**: Config import derives mount source from actual volume paths, not config/<service> assumption
- [ ] **PARS-05**: Config import handles shared/mismatched mappings (php→nginx, python→supervisord, blackbox-exporter→prometheus, nginx-proxy-manager→nginx)
- [ ] **PARS-06**: Config import stores source path, target path, and read-only/mount semantics
- [ ] **PARS-07**: No seed data in any migration — catalog loaded exclusively through importer

### Generator

- [ ] **GENR-01**: Compose generator emits env_files faithfully from DB model
- [ ] **GENR-02**: Compose generator emits explicit network attachments from DB model
- [ ] **GENR-03**: Compose generator emits config mounts with correct provenance from DB model
- [ ] **GENR-04**: Generated compose matches legacy output 1:1 for all 166+ services

### Import Scalability

- [ ] **IMPT-01**: Stack import uses streaming multipart read (no ParseMultipartForm)
- [ ] **IMPT-02**: Route-specific cap via STACK_IMPORT_MAX_BYTES (default 256MB)
- [ ] **IMPT-03**: Global 10MB cap preserved for all other endpoints
- [ ] **IMPT-04**: Bulk import uses prepared statements + batched upserts within transaction
- [ ] **IMPT-05**: Import handles conflicts idempotently
- [ ] **IMPT-06**: Import avoids memory-heavy full buffering

### Dashboard

- [ ] **DASH-01**: Dashboard displays env_files for services and instances
- [ ] **DASH-02**: Dashboard displays network attachments for services and instances
- [ ] **DASH-03**: Dashboard displays config mount provenance (source, target, semantics)
- [ ] **DASH-04**: Dashboard forms support editing env_files, networks, config mounts

### Validation

- [ ] **VALD-01**: Legacy parity verified for ports, volumes, env vars, env files, deps, healthchecks, labels, networks, config mappings across all services
- [ ] **VALD-02**: Golden service parity explicitly verified: php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api
- [ ] **VALD-03**: Upload boundary test: <256MB import passes
- [ ] **VALD-04**: Upload boundary test: >256MB import rejected with clear error

## Future Requirements

### Dashboard Advanced
- **DASH-F01**: Inline config file editing with syntax highlighting for mounted configs
- **DASH-F02**: Visual network topology diagram per stack

## Out of Scope

| Feature | Reason |
|---------|--------|
| Compatibility migration from old chain | Fresh baseline only — old chain deleted |
| Seed data in migrations | Catalog loaded through importer, not migrations |
| Multi-database support | PostgreSQL only |
| Migration rollback tooling | Down migrations exist but no automated rollback UI |
| Real-time compose diff preview | Defer to future milestone |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SCHM-01 | Phase 10 | Pending |
| SCHM-02 | Phase 10 | Pending |
| SCHM-03 | Phase 10 | Pending |
| SCHM-04 | Phase 10 | Pending |
| SCHM-05 | Phase 10 | Pending |
| SCHM-06 | Phase 10 | Pending |
| SCHM-07 | Phase 10 | Pending |
| SCHM-08 | Phase 10 | Pending |
| SCHM-09 | Phase 10 | Pending |
| SCHM-10 | Phase 10 | Pending |
| SCHM-11 | Phase 10 | Pending |
| SCHM-12 | Phase 10 | Pending |
| PARS-01 | Phase 11 | Pending |
| PARS-02 | Phase 11 | Pending |
| PARS-03 | Phase 11 | Pending |
| PARS-04 | Phase 11 | Pending |
| PARS-05 | Phase 11 | Pending |
| PARS-06 | Phase 11 | Pending |
| PARS-07 | Phase 11 | Pending |
| GENR-01 | Phase 12 | Pending |
| GENR-02 | Phase 12 | Pending |
| GENR-03 | Phase 12 | Pending |
| GENR-04 | Phase 12 | Pending |
| IMPT-01 | Phase 13 | Pending |
| IMPT-02 | Phase 13 | Pending |
| IMPT-03 | Phase 13 | Pending |
| IMPT-04 | Phase 13 | Pending |
| IMPT-05 | Phase 13 | Pending |
| IMPT-06 | Phase 13 | Pending |
| DASH-01 | Phase 14 | Pending |
| DASH-02 | Phase 14 | Pending |
| DASH-03 | Phase 14 | Pending |
| DASH-04 | Phase 14 | Pending |
| VALD-01 | Phase 15 | Pending |
| VALD-02 | Phase 15 | Pending |
| VALD-03 | Phase 15 | Pending |
| VALD-04 | Phase 15 | Pending |

**Coverage:**
- v1.1 requirements: 36 total
- Mapped to phases: 36/36 (100%)
- Unmapped: 0

---
*Requirements defined: 2026-02-09*
*Last updated: 2026-02-09 — traceability complete*
