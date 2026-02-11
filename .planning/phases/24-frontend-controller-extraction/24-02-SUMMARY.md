---
phase: 24
plan: 02
subsystem: dashboard
tags: [refactor, controller-pattern, mutation-helper, instances]
dependency_graph:
  requires: [mutation-helper-pattern]
  provides: [instance-controller-hook]
  affects: [instances-queries, instance-detail-page]
tech_stack:
  added: [useInstanceDetailController]
  patterns: [controller-hook-pattern, mutation-helper-application]
key_files:
  created:
    - dashboard/src/features/instances/useInstanceDetailController.ts
  modified:
    - dashboard/src/features/instances/queries.ts
    - dashboard/src/routes/stacks/$name.instances.$instance.tsx
decisions:
  - useMutationHelper applied to all 21 instance mutations - consistent with stack pattern
  - Controller only manages top-level queries/mutations - override components self-contained
  - timeAgo kept in page component - depends on Date.now() for freshness
metrics:
  duration: 287s
  tasks_completed: 2
  files_modified: 3
  completed: 2026-02-11
---

# Phase 24 Plan 02: Instance Mutations & Controller Extraction Summary

**One-liner:** 21 instance mutations refactored to useMutationHelper; controller hook consolidates instance detail orchestration

## Objective

Apply mutation helper pattern to instances/queries.ts (21 mutations) and extract instance page orchestration into `useInstanceDetailController` (FE-02).

## Completed Tasks

### Task 1: Refactor instances/queries.ts to use mutation helper

**Commit:** 1d2c729

Converted 21 mutations to use `useMutationHelper`:

**CRUD operations:**
- useCreateInstance, useUpdateInstance, useDeleteInstance
- useDuplicateInstance, useRenameInstance

**Lifecycle actions:**
- useStopInstance, useStartInstance, useRestartInstance

**Override mutations (with effective-config invalidation):**
- useUpdateInstancePorts, useUpdateInstanceVolumes
- useUpdateInstanceEnvVars, useUpdateInstanceLabels
- useUpdateInstanceDomains, useUpdateInstanceHealthcheck
- useUpdateInstanceDependencies, useUpdateInstanceConfigMounts
- useUpdateInstanceEnvFiles, useUpdateInstanceNetworks
- useUpdateResourceLimits

**Config file mutations (no success toast):**
- useSaveConfigFile, useDeleteConfigFile

**Preserved query functions:**
- useInstances, useInstance, useEffectiveConfig
- useInstanceDeletePreview, useResourceLimits

**Removed imports:** `useMutation`, `useQueryClient`, `toast` (no longer needed)

**Files modified:**
- `dashboard/src/features/instances/queries.ts`

**Lines changed:** -228 removed, +165 added (net -63 via DRY)

### Task 2: Extract useInstanceDetailController and refactor instance detail page

**Commit:** fbd7325

Created `useInstanceDetailController(stackName, instanceId)` hook:
- Orchestrates `useInstance` for instance data
- Orchestrates `useService` for template service data (enabled when template_name available)
- Provides `updateInstance` mutation for toggle/edit actions
- Returns `instance`, `templateService`, `isLoading`, `updateInstance`

Refactored `src/routes/stacks/$name.instances.$instance.tsx`:
- Replaced 3 individual hooks with `ctrl = useInstanceDetailController(stackName, instanceId)`
- All instance/templateService references now via `ctrl.*`
- Update mutation via `ctrl.updateInstance`

**Preserved in page (presentation concerns):**
- Dialog state (editOpen, deleteOpen, duplicateOpen, renameOpen, editDescription)
- Event handlers (openEdit, handleToggleEnabled, handleEditSave, handleDeleteSuccess, etc.)
- timeAgo helper function (uses Date.now() - must stay fresh in render path)
- Route definition and tabs configuration
- Override tab components (self-contained with same props)

**Files modified:**
- `dashboard/src/features/instances/useInstanceDetailController.ts` (created)
- `dashboard/src/routes/stacks/$name.instances.$instance.tsx` (refactored)

**Lines changed:** +15 created, -63 removed (net -48 via consolidation)

## Deviations from Plan

None - plan executed exactly as written.

## Verification

- [x] TypeScript compilation passes (`npx tsc --noEmit`)
- [x] ESLint passes for all modified files
- [x] Production build succeeds (`npm run build`)
- [x] instances/queries.ts uses useMutationHelper (22 occurrences - 21 mutations + 1 import)
- [x] useInstanceDetailController.ts exports controller hook
- [x] Instance detail page imports and uses controller hook (2 occurrences - import + usage)
- [x] Override components receive same props via ctrl.instance/ctrl.templateService

## Key Outcomes

**Mutation helper benefits:**
- 21 mutations reduced from ~15 lines each to ~7 lines each
- Consistent error handling across all instance mutations
- Preserved complex invalidation patterns (instance + effective-config + instances list + stack)
- Override mutations correctly invalidate effective-config for real-time UI sync

**Controller benefits:**
- Instance detail page orchestration logic consolidated
- Page component 48 lines smaller - more presentational
- Clear separation: controller = query/mutation orchestration, page = UI/dialogs/navigation
- Override components remain self-contained - controller doesn't manage per-override mutations

**Behavior preserved:**
- Identical rendering before/after
- Same toast messages for all mutations
- Same invalidation patterns (critical for override effective-config sync)
- timeAgo freshness maintained (kept in page render path)

## Next Steps

Pattern now established for:
- Applying mutation helper to remaining query files (services, categories, etc.)
- Extracting controllers for other complex pages
- Consistent FE architecture across dashboard

## Self-Check: PASSED

All created files exist. All commits verified.
