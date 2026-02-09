# Phase 12: Compose Generator Parity - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Generated compose output matches legacy 1:1 using new schema. Generator reads from service_env_files, service_networks, and service_config_mounts tables and emits those sections in YAML. Both single-service (generator.go) and stack (stack.go) generators updated. No schema changes, no new tables — pure read-and-emit.

</domain>

<decisions>
## Implementation Decisions

### Generator scope
- Update both generators together: generator.go (single-service) AND stack.go (stack)
- Both need identical queries for the 3 new tables
- Forward-looking: Phase 14 Dashboard surfaces fields for both services and instances; Phase 15 validates "all services" which includes stack generation

### Network emission
- DB-sourced only — read service_networks table, emit exactly what's stored
- No hardcoded fallback to stack default network
- Phase 11 importer already stores all network attachments including microservices-net
- Importer is sole data path (decision from 10-02); fallback would create dual source of truth

### env_file format
- Always emit as YAML list, even for single entries
- Simpler code (no string-vs-list branching)
- Phase 15 allows whitespace/ordering differences; list vs string is semantically identical to Docker/Podman
- Forward-looking: Phase 14 Dashboard can consistently render list widget

### Config mount emission
- Config mounts emitted as regular volume strings in the volumes: section
- Merged alongside service_volumes entries
- Format: source_path:target_path[:ro] matching legacy compose convention
- Legacy compose files use volumes for config mounts — separate section would break parity

### Config mount source paths
- Resolve source paths through MaterializeConfigFiles() output, not raw DB source_path
- Generator already calls MaterializeConfigFiles() which writes config content to disk
- Ensures generated compose references actual files on disk
- Forward-looking: Phase 14 config file edits go through same materialization path

### Parity verification
- Import-then-generate round-trip: import all 173 legacy compose files, generate for each, parse generated output, diff key fields against original parsed
- Build the round-trip comparison tool in this phase
- Forward-looking: Phase 15 reuses this tool directly for final parity verification across all services

### Claude's Discretion
- SQL query structure for 3 new table reads
- YAML field ordering in serviceConfig/stackServiceEntry structs
- Error handling for NULL config_file_id in config mounts
- Batch vs individual queries for new relations
- Exact diff reporting format for round-trip verification

</decisions>

<specifics>
## Specific Ideas

- All 6 decisions account for Phase 13 (no conflict — read-only), Phase 14 (Dashboard field exposure), and Phase 15 (parity verification reuse)
- The round-trip tool should compare: ports, volumes, env vars, env files, deps, healthchecks, labels, networks, config mappings — matching Phase 15's requirement list

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 12-compose-generator-parity*
*Context gathered: 2026-02-09*
