---
phase: 14-dashboard-updates
plan: 03
subsystem: dashboard-frontend
tags: [react, tanstack-query, override-components, env-files, networks, config-mounts]
requires: [14-01, 14-02]
provides:
  - instance-env-files-ui
  - instance-networks-ui
  - instance-config-mounts-ui
  - instance-override-tabs
affects: []
tech-stack:
  added: []
  patterns: [override-component-pattern, editable-card-pattern]
key-files:
  created:
    - dashboard/src/components/instances/override-env-files.tsx
    - dashboard/src/components/instances/override-networks.tsx
    - dashboard/src/components/instances/override-config-mounts.tsx
  modified:
    - dashboard/src/features/instances/queries.ts
    - dashboard/src/routes/stacks/$name.instances.$instance.tsx
decisions:
  - id: env-files-simple-list
    what: Env files override as simple string array input
    why: Follows same pattern as dependencies - just file paths
    impact: Consistent UX across all list-based overrides
  - id: networks-simple-list
    what: Networks override as simple string array input
    why: Network names only, no complex config needed at instance level
    impact: Clean UI for network membership control
  - id: config-mounts-resolved-indicator
    what: Config mounts show resolved/unresolved badge
    why: Users need to see if config_file_id is linked or path is external
    impact: Transparency in config mount provenance
  - id: tab-order-logical-grouping
    what: Tabs ordered by logical grouping (runtime, filesystem, network, config, meta)
    why: Puts related overrides near each other for easier navigation
    impact: Better UX - env-files near environment, networks standalone, config-mounts near files
metrics:
  duration: 314s
  tasks: 2
  files_created: 3
  files_modified: 2
  commits: 2
completed: 2026-02-09
---

# Phase 14 Plan 03: Instance-level override components for env_files, networks, config_mounts

Instance override UI completing the v1.1 vertical slice - users can now inspect and override all 3 new schema fields.

## Tasks Completed

1. **Instance query hooks + 3 override components** - Added mutation hooks, created OverrideEnvFiles, OverrideNetworks, OverrideConfigMounts following established pattern
2. **Wire into instance detail page** - Added 3 tabs, integrated components with template data read-only + overrides editable

## Deviations from Plan

None - plan executed exactly as written.

## Technical Implementation

**Instance Query Hooks:**
- `useUpdateInstanceEnvFiles` - PUT /env-files with { env_files: string[] }
- `useUpdateInstanceNetworks` - PUT /networks with { networks: string[] }
- `useUpdateInstanceConfigMounts` - PUT /config-mounts with { config_mounts: [...] }
- Updated `useInstance` to default env_files, networks, config_mounts arrays

**OverrideEnvFiles Component:**
- Simple string array input following override-dependencies pattern
- Template section shows service env_files read-only as muted badges
- Override section shows editable input fields with add/remove
- Reset All sends empty array to clear overrides

**OverrideNetworks Component:**
- Identical pattern to OverrideEnvFiles
- Network names as simple strings
- Read-only template + editable overrides

**OverrideConfigMounts Component:**
- Richer form following override-volumes pattern
- Template section shows source → target, readonly badge, resolved/unresolved badge
- Override section has source_path, target_path inputs, readonly checkbox
- Add button appends empty mount
- Reset All clears all overrides

**Instance Detail Page Integration:**
- Added 3 tabs: env-files, networks, config-mounts
- Tab order: info, ports, volumes, env-files, environment, networks, labels, domains, healthcheck, dependencies, config-mounts, files, resources, effective
- Each tab renders corresponding override component with templateService guard

## Verification

```bash
npx tsc --noEmit  # Passes clean
grep OverrideEnvFiles|OverrideNetworks|OverrideConfigMounts  # 6 matches (3 imports + 3 usages)
```

All 3 components render correctly with template data + override sections.

## Next Phase Readiness

**Phase 14 Complete:**
- ✅ All v1.1 schema fields have full dashboard CRUD
- ✅ Service-level components (14-02)
- ✅ Instance-level components (14-03)
- ✅ Effective config extended (14-01)
- ✅ Override detection in place

**Phase 15 (Compose Generator):**
- ✅ Backend API returns env_files, networks, config_mounts
- ✅ Dashboard components ready to consume/display
- ⏭️ Generator needs to emit env_file, networks, config mounts in YAML

**Blockers:** None

**Concerns:** None - straightforward UI extension following established patterns

## Integration Points

- Backend handlers (14-01) provide PUT endpoints
- TypeScript types (14-02) define interfaces
- Override pattern established by volumes/dependencies
- useOverrideSection hook handles edit state
- EditableCard provides consistent chrome
