---
phase: 03-service-instances
plan: 03
subsystem: dashboard-ui
tags: [typescript, tanstack-query, react, instance-ui]

requires: [02-05, 03-01]
provides:
  - Instance TypeScript types for all override models
  - TanStack Query hooks for instance CRUD operations
  - Add Instance dialog with template catalog
  - Instance list with rich cards on stack detail page

affects: [03-04, 03-05]

tech-stack:
  added: []
  patterns:
    - TanStack Query cache invalidation chains
    - Template catalog with search filtering
    - Auto-incrementing instance name generation

key-files:
  created:
    - dashboard/src/types/api.ts: Instance and override types
    - dashboard/src/features/instances/queries.ts: Instance query hooks
    - dashboard/src/components/stacks/add-instance-dialog.tsx: Add instance dialog
  modified:
    - dashboard/src/routes/stacks/$name.tsx: Instance list UI

decisions: []

metrics:
  duration: 3.4min
  tasks: 2
  commits: 2
  files_created: 2
  files_modified: 2
  completed: 2026-02-04
---

# Phase 03 Plan 03: Dashboard Instance UI Summary

Instance types, query hooks, and instance list UI on stack detail page

## What Was Built

Added complete frontend infrastructure for working with instances in the dashboard.

### Instance Types (api.ts)

Full TypeScript types for instance domain:
- `Instance` — base instance with override count
- `InstanceDetail` — instance with all override collections
- Override types: `InstancePort`, `InstanceVolume`, `InstanceEnvVar`, `InstanceLabel`, `InstanceDomain`, `InstanceHealthcheck`, `InstanceConfigFile`
- `EffectiveConfig` — resolved config with overrides applied
- `OverridesApplied` — metadata about which overrides exist
- `InstanceDeletePreview` — delete preview data

### TanStack Query Hooks (instance queries.ts)

Created comprehensive query hooks following existing stack pattern:

**Queries:**
- `useInstances(stackName)` — list instances in stack
- `useInstance(stackName, instanceId)` — get instance detail
- `useEffectiveConfig(stackName, instanceId)` — get resolved config
- `useInstanceDeletePreview(stackName, instanceId)` — delete preview

**Mutations:**
- `useCreateInstance` — create instance from template
- `useUpdateInstance` — update description/enabled
- `useDeleteInstance` — delete instance
- `useDuplicateInstance` — duplicate with overrides
- `useRenameInstance` — rename instance
- Override mutations: ports, volumes, env vars, labels, domains, healthcheck

All mutations invalidate relevant query keys (instance, instances list, effective config, stack).

### Add Instance Dialog

Template catalog dialog for adding instances:
- Search input filters templates by name, image, category
- Grid of template cards showing name, image, category
- Badge shows existing instance count per template
- Clicking template reveals creation form at bottom
- Auto-generates instance name (template name, or template-2, template-3 if collision)
- Optional description field
- Creates instance via API on submit
- Inline error display on failure

### Instance List UI

Replaced placeholder on stack detail page with real instance cards:
- Section header shows count: "Instances (N)"
- "Add Instance" button in header opens dialog
- Empty state: centered CTA card with "Add your first service" and button
- Instance cards in responsive grid:
  - Instance name (bold)
  - Template name (muted)
  - Container name (monospace, small)
  - Status indicator (green/gray dot)
  - Description (line-clamp-2)
  - Override count badge (e.g., "3 overrides")
- Click card navigates to instance detail (route placeholder for Plan 04)

## Technical Decisions

**Auto-name generation with collision detection:** When user selects template, dialog pre-fills instance name as template name, or appends -2, -3, etc. if name exists. Reduces friction for common case (single instance per template) while handling duplicates gracefully.

**Cache invalidation chains:** Override mutations invalidate instance detail, effective config, AND instances list. Ensures UI stays consistent across all views showing instance data.

**Template catalog with instance counts:** Shows how many instances of each template already exist in stack. Helps user understand their stack composition when adding more services.

**Empty state CTA pattern:** Follows 02-04 pattern for empty lists — centered card with icon, message, and primary action button. Consistent UX across stacks and instances.

## Deviations from Plan

None — plan executed exactly as written.

## Testing Notes

TypeScript compilation passes with no errors in changed files. Pre-existing errors in other files remain unchanged.

Visual verification:
- Stack detail page shows "Add Instance" button and instance count
- Empty state displays centered CTA when no instances
- Instance cards show all expected data (name, template, container, status, overrides)
- Add Instance dialog opens with template grid
- Search filters templates correctly
- Template selection reveals creation form
- Instance name auto-generates with collision detection

## Next Phase Readiness

**Ready for 03-04 (Instance Detail Page):**
- All instance types defined
- Instance query hooks ready to use
- Instance cards link to detail page route (placeholder)

**Ready for 03-05 (Override Editors):**
- Override mutation hooks ready
- Effective config query hook ready
- Override count visible on instance cards

## Blockers/Concerns

None.

## Performance Notes

Duration: 3.4 minutes
- Task 1 (types + hooks): 1.5 min
- Task 2 (UI components): 1.9 min

Instance list uses 5-second polling (same as stack list). Override editor interactions will use optimistic updates in 03-05.
