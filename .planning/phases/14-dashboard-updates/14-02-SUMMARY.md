---
phase: 14-dashboard-updates
plan: 02
subsystem: dashboard
tags: [dashboard, typescript, react, ui, env-files, networks, config-mounts]
requires:
  - phase-13 (streaming import)
  - phase-14-01 (instance override types)
provides:
  - service-level UI for env_files field
  - service-level UI for networks field
  - service-level UI for config_mounts field with resolution badges
affects:
  - phase-14-03 (instance override components will follow same pattern)
tech-stack:
  added: []
  patterns:
    - editable-* component pattern for service sub-resources
    - resolved/unresolved badge pattern for config mount provenance
key-files:
  created:
    - dashboard/src/components/services/editable-env-files.tsx
    - dashboard/src/components/services/editable-networks.tsx
    - dashboard/src/components/services/editable-config-mounts.tsx
  modified:
    - dashboard/src/types/api.ts
    - dashboard/src/features/services/queries.ts
    - dashboard/src/routes/services/$name.tsx
decisions:
  - id: service-env-files-placement
    summary: Placed EditableEnvFiles after EditableVolumes in info tab
    rationale: Env files relate to volume-like configuration; placing near volumes maintains related context
  - id: service-networks-placement
    summary: Placed EditableNetworks after EditableEnvFiles
    rationale: Networks are deployment-level config; grouped with env files before dependencies
  - id: service-config-mounts-placement
    summary: Placed EditableConfigMounts after Dependencies before Healthcheck
    rationale: Config mounts bridge config files and volumes; position reflects mount provenance concern
  - id: config-mount-badge-colors
    summary: Green for resolved, amber for unresolved
    rationale: Green = success (linked to DB config file), amber = warning (external path only)
metrics:
  duration: 134s
  completed: 2026-02-09
---

# Phase 14 Plan 02: Service-Level env_files/networks/config_mounts Components Summary

**One-liner:** Dashboard components for service-level env_files, networks, and config_mounts with resolved/unresolved badges for mount provenance.

## Execution Summary

Added three new editable-* components to service detail page, following established patterns from EditableDependencies (ordered list) and EditableVolumes (structured form). Components support full CRUD with proper TypeScript types and query hooks.

### Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Add TypeScript types, query hooks, and 3 editable components | c9faaac | 5 files (api.ts, queries.ts, 3 components) |
| 2 | Wire service-level components into service detail page | f97f45b | $name.tsx |

### Implementation Details

**TypeScript Types:**
- Added `ServiceConfigMount` interface with optional `config_file_id` for resolution tracking
- Added `env_files`, `networks`, `config_mounts` to `Service` and `EffectiveConfig` interfaces
- Added corresponding override flags to `OverridesApplied` interface

**Query Hooks:**
- `useUpdateEnvFiles` → PUT /services/{name}/env-files
- `useUpdateNetworks` → PUT /services/{name}/networks
- `useUpdateConfigMounts` → PUT /services/{name}/config-mounts
- All use `makeSubResourceMutation` pattern for consistency

**EditableEnvFiles Component:**
- View mode: Badge list showing file paths (returns null if empty)
- Edit mode: Input field per path with trash button, Add button appends empty string
- Preserves array order (matches sort_order column)

**EditableNetworks Component:**
- Identical pattern to EditableEnvFiles
- View mode: Badge list of network names
- Edit mode: Input per network with add/remove

**EditableConfigMounts Component:**
- View mode: Displays `source_path → target_path` with badges
  - Green "resolved" badge if `config_file_id` is set (linked to DB)
  - Amber "unresolved" badge if `config_file_id` is null (external path)
  - Gray "ro" badge if readonly flag set
- Edit mode: source/target inputs + readonly checkbox
- Note: `config_file_id` not user-editable (set server-side during import)

**Service Detail Page Integration:**
Order in info tab:
1. EditableDomains
2. EditablePorts
3. EditableVolumes
4. **EditableEnvFiles** (new)
5. **EditableNetworks** (new)
6. EditableDependencies
7. **EditableConfigMounts** (new)
8. EditableHealthcheck
9. EditableLabels

## Deviations from Plan

None - plan executed exactly as written.

## Technical Insights

**Component Pattern Consistency:**
All three components follow established editable-* patterns:
- `useEditableSection` hook for draft state management
- `EditableCard` wrapper for edit/view mode UI
- `FieldRow` / `FieldRowActions` for edit mode layout
- Null return when empty + not editing (reduces clutter)

**Config Mount Resolution Badge:**
The resolved/unresolved badge provides visual feedback on config mount provenance:
- Resolved = `config_file_id` points to DB-managed config file (materialized during generation)
- Unresolved = external path only (used as-is in compose YAML)

This pattern will be reused in Phase 14-03 for instance-level config mounts.

## Next Phase Readiness

**Dependencies satisfied for 14-03 (Instance Override Components):**
- [x] Service-level component patterns established
- [x] TypeScript types include all override tables
- [x] Query hook pattern proven

**Known considerations:**
- Instance components will need reset/override detection logic (OverridesApplied flags)
- Instance config mounts may have different resolution state than template (overrides can point to different config files)

## Verification Results

All verification criteria met:
- [x] `tsc --noEmit` compiles clean (verified after each task)
- [x] Service detail page imports all 3 components
- [x] Components placed in correct order per plan spec
- [x] EditableEnvFiles/EditableNetworks show badge lists
- [x] EditableConfigMounts shows resolved/unresolved badges
- [x] All follow Card + edit/view pattern
- [x] Query hooks call correct PUT endpoints via makeSubResourceMutation

## Performance

**Execution time:** 2m14s (134 seconds)
- Task 1: ~80s (types + 3 components + verification)
- Task 2: ~30s (wiring + verification)
- Summary creation: ~24s

**Efficiency:** Slightly faster than Phase 14-01 (3m19s) due to simpler implementation (no API changes, pure UI layer).
