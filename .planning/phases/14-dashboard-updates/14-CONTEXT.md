# Phase 14: Dashboard Updates - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Dashboard UI exposes new v1.1 schema fields (env_files, networks, config mounts) for user inspection and editing — on both service templates and instances. Includes the full vertical slice: missing instance override tables (migration), Go API handlers/queries, TypeScript types, and React dashboard components. No new dashboard pages or navigation changes — fields integrate into existing service and instance detail views.

</domain>

<decisions>
## Implementation Decisions

### Instance override schema gap
- Add 3 new instance override tables via migration 010: `instance_env_files`, `instance_networks`, `instance_config_mounts`
- Tables mirror their service-level counterparts structurally (same columns, FK to service_instances instead of services)
- API endpoints follow existing override pattern (GET/PUT per field type on instance routes)
- Full vertical slice in one phase: migration → API handlers → dashboard UI
- Forward consideration: Phase 15 validation needs instance-level overrides for env_files/networks/config_mounts to verify parity. Without these tables, Phase 15 success criteria #1 has a gap.

### Display placement
- Env files section placed near env vars on service detail page (related data together)
- Networks section placed near existing network info
- Config mounts section placed near config files panel
- No new tabs or pages — fields integrate into existing tab structure

### Env files display
- Editable ordered list following `editable-dependencies.tsx` pattern
- Each item shows path string with trash button for removal, + Add button
- Sort order preserved (matches `sort_order` column from service_env_files)
- Service-level: new `editable-env-files.tsx` component
- Instance-level: new `override-env-files.tsx` using EditableCard + useOverrideSection pattern

### Networks display
- Editable list of network names following same list pattern as env files
- Service-level: new `editable-networks.tsx` component
- Instance-level: new `override-networks.tsx` using EditableCard + useOverrideSection pattern

### Config mount provenance display
- Show source_path → target_path like volumes display pattern
- When config_file_id is set: clickable link/badge to the linked config file
- When config_file_id is NULL: "unresolved" warning badge
- Full CRUD: users can add, edit, remove mounts (supports services not yet imported)
- Service-level: new `editable-config-mounts.tsx` component
- Instance-level: new `override-config-mounts.tsx` using EditableCard + useOverrideSection pattern
- Forward consideration: Phase 15 can check unresolved mount counts as part of parity validation

### CRUD form pattern
- Service-level: follow existing `editable-*` component pattern (inline edit/save/cancel within cards)
- Instance-level: follow existing `override-*` component pattern (EditableCard + useOverrideSection hook, template data read-only, overrides editable, Reset All button)
- Forward consideration: Phase 15 golden service verification can test override round-trips through these forms

### Effective config updates
- Add env_files, networks, config_mounts to EffectiveConfig API response and TypeScript type
- Add corresponding entries to OverridesApplied type
- Config mounts in effective view include resolution status (whether config_file_id resolved)
- Forward consideration: Phase 15 can use effective config API endpoint for automated parity checks across all field types

### Claude's Discretion
- Exact component layout and spacing within existing tabs
- Whether env_files sort order uses drag-and-drop or up/down arrows
- Badge styling for resolved vs unresolved config mount links
- API query structure (separate queries vs joined)
- Migration column types and defaults (follow existing patterns)

</decisions>

<specifics>
## Specific Ideas

- Existing override pattern is well-established (8 override component files, useOverrideSection hook) — new fields should be structurally identical
- Config mount form needs source_path, target_path, readonly checkbox, and optional config_file_id selector
- Env files are the simplest (just a path string + sort order) — good candidate for first implementation to establish the pattern

</specifics>

<deferred>
## Deferred Ideas

- DASH-F01 (inline config file editing with syntax highlighting) — future requirement, not in v1.1 scope
- DASH-F02 (visual network topology diagram per stack) — future requirement, not in v1.1 scope

</deferred>

---

*Phase: 14-dashboard-updates*
*Context gathered: 2026-02-09*
