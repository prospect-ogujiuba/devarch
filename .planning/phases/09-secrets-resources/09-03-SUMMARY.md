---
phase: 09-secrets-resources
plan: 03
subsystem: dashboard
tags: [ui, secrets, resources, consistency]
dependency_graph:
  requires: [09-01]
  provides: [consistent secret masking, resource limits UI]
  affects: [instance detail page, effective config tab, all env var components]
tech_stack:
  added: []
  patterns: [consistent secret redaction, optional EditableCard props]
key_files:
  created:
    - dashboard/src/components/instances/resource-limits.tsx
  modified:
    - dashboard/src/types/api.ts
    - dashboard/src/features/instances/queries.ts
    - dashboard/src/components/services/editable-env-vars.tsx
    - dashboard/src/routes/stacks/$name.instances.$instance.tsx
    - dashboard/src/components/instances/effective-config-tab.tsx
    - dashboard/src/components/ui/editable-card.tsx
decisions:
  - Use 8 bullet characters (••••••••) for all secret masking in dashboard
  - Make EditableCard onAdd prop optional to support single-object editing
  - Resource limits displayed in both dedicated Resources tab and effective config
  - API warnings from resource limits update shown in amber/yellow text
metrics:
  duration: 276
  completed: 2026-02-09T04:02:32Z
---

# Phase 09 Plan 03: Dashboard Secret Consistency & Resource Limits UI Summary

**One-liner:** Standardized secret display to 8 bullet characters across all env var components and added full-featured resource limits editor with validation warnings.

## What Was Built

### Task 1: Secret Display Standardization & Types

**types/api.ts:**
- Added `ResourceLimits` interface with optional cpu_limit, cpu_reservation, memory_limit, memory_reservation
- Added `ResourceLimitsResponse` interface with limits object and optional warnings array

**features/instances/queries.ts:**
- Added `useResourceLimits(stackName, instanceId)` query hook - returns ResourceLimits or null if 404
- Added `useUpdateResourceLimits(stackName, instanceId)` mutation hook - invalidates resource limits, instance detail, and effective config on success
- Imports ResourceLimits and ResourceLimitsResponse types from api.ts

**editable-env-vars.tsx:**
- Changed secret masking from `'********'` (8 asterisks) to `'••••••••'` (8 bullets, \u2022)
- Read-only display now consistent with override-env-vars.tsx which already used bullets

### Task 2: Resource Limits Editor & Integration

**resource-limits.tsx (new):**
- Created ResourceLimits component following EditableCard pattern
- Props: `{ stackName: string, instanceId: string }`
- Read mode: displays CPU/memory limits in 2-column grid, or "No resource limits configured" if empty
- Edit mode: 4 input fields (CPU Limit/Reservation, Memory Limit/Reservation) with placeholders and descriptions
- Save button calls useUpdateResourceLimits mutation
- Clear All button sends empty object to remove all limits
- Validation warnings from API response displayed in amber text below form
- Uses single-item array pattern with useEditableSection hook

**$name.instances.$instance.tsx:**
- Added "Resources" tab to instance detail page (between "Config Files" and "Effective Config")
- Updated instanceTabs array and validateSearch enum to include 'resources'
- Added TabsContent for resources with ResourceLimits component

**effective-config-tab.tsx:**
- Added useResourceLimits hook call
- Added "Resource Limits" card section that displays when any limits are set
- Shows CPU limits as "X.X cores", memory limits as-is (e.g., "512m", "1g")
- Section only appears if at least one limit is configured

**editable-card.tsx:**
- Made `onAdd` prop optional (was required)
- Add button only rendered when onAdd prop is provided
- Enables single-object editing pattern (resource limits) without showing unnecessary Add button

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Functionality] Made EditableCard onAdd prop optional**
- **Found during:** Task 2 implementation
- **Issue:** EditableCard required onAdd prop and always showed Add button, but resource limits editor edits a single object (not a list) and doesn't need Add functionality. This prevented proper use of EditableCard for single-object editing.
- **Fix:** Made onAdd optional in EditableCardProps, conditionally render Add button only when onAdd is provided.
- **Files modified:** dashboard/src/components/ui/editable-card.tsx
- **Commit:** 0c225cc9

## Verification Results

**Build:** ✅ `npm run build:strict` passes

**Secret masking consistency:** ✅ Both service template env vars (editable-env-vars.tsx) and instance override env vars (override-env-vars.tsx) now use `'••••••••'` (8 bullets)

**Resource limits types:** ✅ ResourceLimits and ResourceLimitsResponse exported from api.ts

**Query hooks:** ✅ useResourceLimits and useUpdateResourceLimits compile and follow TanStack Query patterns

**Resources tab:** ✅ Added to instance detail page between Config Files and Effective Config

**Effective config integration:** ✅ Resource limits section appears when limits are set, hidden when not configured

**EditableCard flexibility:** ✅ onAdd prop now optional, supports both list-based editing (ports, volumes) and single-object editing (resource limits)

## Technical Notes

**Secret masking character:** \u2022 (bullet) renders as • - visually distinct from asterisks, consistent with override components

**Resource limits validation:** API validation warnings (e.g., "memory limit very low") passed through ResourceLimitsResponse and displayed in component

**EditableCard pattern:** Making onAdd optional maintains backward compatibility (all existing uses provide onAdd) while enabling new use cases

**Query invalidation cascade:** useUpdateResourceLimits invalidates resources, instance detail, and effective config queries to ensure all views refresh

**404 handling:** useResourceLimits returns null when resource limits endpoint returns 404 (no limits configured), not an error

## Next Steps (Plan 02)

Plan 02 will implement the API endpoints that these UI components depend on:
- GET `/api/v1/stacks/{stack}/instances/{instance}/resources`
- PUT `/api/v1/stacks/{stack}/instances/{instance}/resources`
- Resource limits integration into compose generation
- Validation warnings for unreasonable limits

## Self-Check: PASSED

All created files exist:
- ✅ dashboard/src/components/instances/resource-limits.tsx

All modified files have changes:
- ✅ dashboard/src/types/api.ts (ResourceLimits types added)
- ✅ dashboard/src/features/instances/queries.ts (resource limits hooks added)
- ✅ dashboard/src/components/services/editable-env-vars.tsx (bullets not asterisks)
- ✅ dashboard/src/routes/stacks/$name.instances.$instance.tsx (Resources tab added)
- ✅ dashboard/src/components/instances/effective-config-tab.tsx (resource limits section added)
- ✅ dashboard/src/components/ui/editable-card.tsx (onAdd made optional)

All commits exist:
- ✅ 5e1d741c: feat(09-03): add resource limits types and standardize secret display
- ✅ 0c225cc9: feat(09-03): add resource limits editor and instance detail integration
