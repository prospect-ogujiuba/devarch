# Phase 11: Parser & Importer Updates - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Import logic preserves all compose constructs (env_file, container_name, networks, config mounts with provenance) into new schema tables from Phase 10. No seed data in migrations — catalog loaded exclusively through importer.

Scope: parser correctness and DB persistence for PARS-01 through PARS-07. Does NOT include streaming (Phase 13), generator changes (Phase 12), or UI (Phase 14).

</domain>

<decisions>
## Implementation Decisions

### Data Model Bridging
- Add flat fields directly to ParsedService struct: EnvFiles []string, ContainerName string, Networks []string, ConfigMounts []ParsedConfigMount
- Matches existing pattern (ParsedPort, ParsedVolume, etc.) — no nested sub-structs
- New ParsedConfigMount type with source_path, target_path, readonly, config_file_id — maps 1:1 to service_config_mounts table
- Store both original relative paths AND normalized paths — Phase 15 needs originals for byte-for-byte parity
- Populate config_file_id FK during import (same transaction) — config_files already created by import, link mounts immediately

### Config Mount Provenance
- Volume-path-driven ownership: config_file record belongs to whichever config directory it lives in (nginx owns http.conf even when php mounts it)
- Mount record on consuming service links to owning service's config_file via FK — provenance is "php mounts nginx's http.conf"
- Replace rewriteConfigPaths() with provenance-aware logic — remove assumption-based rewrite entirely
- Path-pattern heuristic to classify volumes: source path containing 'config/' prefix = config mount, everything else = regular volume
- Missing config files (referenced in compose but not on disk): warn and skip mount, store service_config_mounts record with null config_file_id — Phase 15 catches these

### Env File Handling
- Store original relative paths as-is (e.g., "../../.env") — Phase 12 generator emits exact same string for parity
- Normalize string and list forms to []string internally — single string becomes 1-element list
- sort_order column preserves declaration ordering
- Store regardless of whether referenced file exists on disk — existence is runtime concern, not import concern

### Network Attachments
- Store network names only (service_networks.network_name) — no driver/ipam/config parsing
- Network configuration (top-level networks: block) is Phase 4's domain (already shipped)
- PARS-03 scope: which networks a service attaches to, nothing more

### Import Write Strategy
- DELETE + INSERT per service for new tables (env_files, networks, config_mounts) — matches existing pattern for ports/volumes/env_vars
- All service data in single transaction (ports, volumes, env_vars, env_files, networks, config_mounts) — atomic per-service import
- Keep []byte input interface — Phase 13 streaming is about HTTP multipart boundary, not compose file parsing (individual files are <10KB)
- No UPSERT/conflict handling — Phase 13 explicitly owns that (IMPT-05)

### Claude's Discretion
- Exact parsing logic for env_file interface{} type coercion (string vs []interface{})
- Config path pattern matching regex/logic details
- Error logging format for missing config files
- Internal function decomposition within importer.go

</decisions>

<specifics>
## Specific Ideas

- Cross-service config mappings that MUST work: php->nginx/custom/http.conf, python->supervisord/supervisord.conf, blackbox-exporter->prometheus/blackbox.yml, nginx-proxy-manager->nginx/
- All 173 compose files in old-compose-and-configs/compose/ use env_file: ../../.env (string form)
- All services attach to microservices-net explicitly
- container_name is present in all 173 compose files

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 11-parser-importer-updates*
*Context gathered: 2026-02-09*
