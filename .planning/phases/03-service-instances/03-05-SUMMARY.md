---
phase: 03-service-instances
plan: 05
subsystem: dashboard-effective-config-actions
tags: [react, effective-config, instance-actions, yaml-export, lifecycle]
requires: [03-04]
provides:
  - Effective config tab with YAML/JSON export
  - Instance lifecycle action dialogs (delete, duplicate, rename)
  - Enable/disable toggle
affects: []
tech-stack:
  added: [yaml]
  patterns: [override-indicator-borders, blast-radius-preview, format-toggle]
key-files:
  created:
    - dashboard/src/components/instances/effective-config-tab.tsx
    - dashboard/src/components/instances/instance-actions.tsx
  modified:
    - dashboard/src/routes/stacks/$name.instances.$instance.tsx
    - dashboard/src/routes/stacks/$name.tsx
decisions:
  - slug: yaml-library-for-config-export
    summary: Used 'yaml' npm package for YAML serialization in effective config copy
    rationale: Clean YAML output for compose-style preview
  - slug: format-toggle-yaml-json
    summary: YAML/JSON toggle with copy-to-clipboard
    rationale: Users may need either format for different workflows
metrics:
  duration: ~5min
  completed: 2026-02-04
  note: "Summary reconstructed — original lost to context window reset"
---

# Phase 3 Plan 5: Effective Config + Instance Lifecycle Summary

Effective config preview tab and instance lifecycle action dialogs

## One-liner

Effective config tab with YAML/JSON copy, delete/duplicate/rename/enable-disable dialogs wired into instance detail page

## What Was Built

### Effective Config Tab
- Structured read-only display matching override editor layout
- Sections: Image, Ports, Volumes, Environment, Labels, Domains, Healthcheck, Dependencies, Config Files
- Overridden sections marked with blue left border + "Overridden" badge
- YAML/JSON format toggle with copy-to-clipboard
- Uses `useEffectiveConfig` query hook
- Merges template + overrides with `overrides_applied` tracking

### Instance Action Dialogs
- **DeleteInstanceDialog**: AlertDialog with blast radius preview (instance name, template, override count, container name)
- **DuplicateInstanceDialog**: Input for new name (auto-filled "{instance}-copy"), copies all overrides
- **RenameInstanceDialog**: Input for new name, validates non-empty and different from current
- **Enable/Disable toggle**: Button in header, calls useUpdateInstance with toggled enabled state

### Instance Detail Page Integration
- Effective Config placeholder replaced with full implementation
- Header actions: enable/disable button + dropdown menu (edit description, duplicate, rename, delete)
- All dialogs wired with success navigation (delete→stack, duplicate→stack, rename→new URL)

## Deviations from Plan

- Added Outlet to stack detail page for TanStack Router nested routing (fix commit)
- Added null-safe defaults for effective config API response arrays (fix commit)

## Technical Notes

### YAML Export
Uses `yaml` npm package. Formats effective config as compose-style YAML:
```yaml
version: '3.8'
services:
  {instance_id}:
    image: {image}:{tag}
    container_name: {name}
    ports: [...]
    environment: {...}
```

### Override Indicators
`overrides_applied` object from API maps section names to booleans. Blue left border + badge on overridden sections.

## Next Phase Readiness

**Enables:**
- Phase 4: Network isolation (instances have all data needed for container naming)
- Phase 5: Compose generation (effective config provides merge output)
- Phase 6: Plan/apply (full instance definition visible before apply)

## State Changes

### Files Created
- `dashboard/src/components/instances/effective-config-tab.tsx`
- `dashboard/src/components/instances/instance-actions.tsx`

### Files Modified
- `dashboard/src/routes/stacks/$name.instances.$instance.tsx` — wired dialogs + effective config tab
- `dashboard/src/routes/stacks/$name.tsx` — added Outlet for nested routing

## Task Breakdown

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Effective config tab + instance action dialogs | 0273ff6 | effective-config-tab.tsx, instance-actions.tsx, $instance.tsx, $name.tsx |
| fix | Outlet for nested routing | 14abf09 | $name.tsx |
| fix | Null-safe effective config arrays | 01e17c9 | effective-config-tab.tsx |
