---
phase: 24
plan: 01
subsystem: dashboard
tags: [refactor, controller-pattern, DRY, mutation-helper]
dependency_graph:
  requires: []
  provides: [mutation-helper-pattern, stack-controller-hook]
  affects: [stacks-queries, stack-detail-page]
tech_stack:
  added: [useMutationHelper]
  patterns: [controller-hook-pattern, mutation-factory]
key_files:
  created:
    - dashboard/src/lib/mutations.ts
    - dashboard/src/features/stacks/useStackDetailController.ts
  modified:
    - dashboard/src/features/stacks/queries.ts
    - dashboard/src/routes/stacks/$name.tsx
decisions:
  - Kept complex mutations (useApplyPlan, useExportStack, useImportStack, useGeneratePlan) manual - custom behavior doesn't fit helper pattern
  - Preserved InstanceCard component in page with own mutations - per-instance actions are self-contained
  - Dialog state remains in page - presentation concern, not controller business
metrics:
  duration: 393s
  tasks_completed: 2
  files_modified: 4
  completed: 2026-02-11
---

# Phase 24 Plan 01: Shared Mutation Helper & Stack Controller Extraction Summary

**One-liner:** Shared mutation helper factory eliminates toast+invalidation boilerplate; stack detail controller hook consolidates query orchestration

## Objective

Establish the mutation helper pattern (FE-04 foundation) and extract stack page orchestration into `useStackDetailController` (FE-01).

## Completed Tasks

### Task 1: Create shared mutation helper and refactor stacks/queries.ts

**Commit:** 6dcad6a

Created `lib/mutations.ts` with `useMutationHelper` factory:
- Wraps `useMutation` with configurable toast + invalidation
- Supports string or function-based success/error messages
- Accepts array of query keys to invalidate on success
- Optional onSuccess callback for custom logic

Refactored 18 standard mutations in `stacks/queries.ts`:
- useCreateStack, useUpdateStack, useDeleteStack
- useEnableStack, useDisableStack
- useCloneStack, useRenameStack
- useRestoreStack, usePermanentDeleteStack
- useStopStack, useStartStack, useRestartStack
- useCreateNetwork, useRemoveNetwork
- useResolveWires, useCreateWire, useDeleteWire, useCleanupOrphanedWires

Preserved 4 complex mutations as manual (custom behavior):
- useApplyPlan (409-specific error handling)
- useExportStack (blob download)
- useImportStack (multipart upload + 413 handling)
- useGeneratePlan (no success toast)

**Files modified:**
- `dashboard/src/lib/mutations.ts` (created)
- `dashboard/src/features/stacks/queries.ts` (refactored)

**Lines changed:** +49 created, -134 removed (net -85 via DRY)

### Task 2: Extract useStackDetailController and refactor stack detail page

**Commit:** 8bf8dea

Created `useStackDetailController(stackName)` hook:
- Orchestrates 4 query hooks (useStack, useInstances, useStackNetwork, useStackCompose)
- Instantiates 13 mutation hooks
- Derives connectedContainers and runningContainerNames
- Returns unified interface for page consumption

Refactored `src/routes/stacks/$name.tsx`:
- Replaced 17 individual hook calls with single `ctrl = useStackDetailController(name)`
- All stack/instances/networkStatus/composeData references now via `ctrl.*`
- All mutation references now via `ctrl.enableStack`, `ctrl.stopStack`, etc.

**Preserved in page (presentation concerns):**
- Dialog state (editOpen, deleteOpen, etc.)
- File upload ref and handlers
- currentPlan state (local UI state for deploy tab)
- InstanceCard component (self-contained with own mutations)
- timeAgo helper
- Route definition and tabs config

**Files modified:**
- `dashboard/src/features/stacks/useStackDetailController.ts` (created)
- `dashboard/src/routes/stacks/$name.tsx` (refactored)

**Lines changed:** +63 created, -92 removed (net -29 via consolidation)

## Deviations from Plan

None - plan executed exactly as written.

## Verification

- [x] TypeScript compilation passes (`npx tsc --noEmit`)
- [x] ESLint passes for all modified files
- [x] Production build succeeds (`npm run build`)
- [x] lib/mutations.ts exports useMutationHelper
- [x] stacks/queries.ts standard mutations use useMutationHelper
- [x] useStackDetailController.ts exports controller hook
- [x] Stack detail page imports and uses controller hook

## Key Outcomes

**Mutation helper benefits:**
- 18 mutations reduced from ~15 lines each to ~7 lines each
- Consistent error handling across all standard mutations
- Single source of truth for toast + invalidation pattern
- Type-safe factory with generic TData/TVariables

**Controller benefits:**
- Stack detail page orchestration logic consolidated in single hook
- Page component 29 lines smaller - more presentational
- Easier to test (mock controller vs individual hooks)
- Clear separation: controller = business logic, page = presentation

**Behavior preserved:**
- Identical rendering before/after
- Same toast messages
- Same invalidation patterns
- Same mutation behavior

## Next Steps

Pattern established for:
- Applying mutation helper to other query files (instances, services, etc.)
- Extracting controllers for other complex pages
- Consistent FE architecture across dashboard

## Self-Check: PASSED

All created files exist. All commits verified.
