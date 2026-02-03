---
phase: 02-stack-crud
plan: 03
subsystem: dashboard-data-layer
status: complete
completed: 2026-02-03

requires:
  - 02-01 (Stack API handlers for type contracts)

provides:
  - Stack TypeScript types matching API response shape
  - TanStack Query hooks for all stack CRUD operations
  - Stacks navigation link

affects:
  - 02-04 (Stack list/detail UI will consume these hooks)
  - 02-05 (Stack forms will use create/update mutations)

tech-stack:
  added: []
  patterns:
    - TanStack Query v5 mutations with cache invalidation
    - Toast notifications via sonner for mutation feedback

key-files:
  created:
    - dashboard/src/features/stacks/queries.ts
  modified:
    - dashboard/src/types/api.ts
    - dashboard/src/components/layout/header.tsx

decisions:
  - decision: Stack hooks follow existing service hooks pattern
    rationale: Consistency with established codebase patterns
    impact: Maintainable, predictable mutation behavior

metrics:
  tasks: 3
  commits: 3
  files_created: 1
  files_modified: 2
  duration: 69 seconds

tags: [dashboard, typescript, tanstack-query, stacks, data-layer]
---

# Phase 02 Plan 03: Stack Dashboard Data Layer Summary

**One-liner:** Stack TypeScript types, 13 TanStack Query hooks (4 queries, 9 mutations), and Stacks navigation link.

## What Was Built

Created complete dashboard data layer for stacks:

1. **TypeScript types** (api.ts)
   - Stack interface with all fields from API response
   - StackInstance interface for nested instance data
   - DeletePreview interface for cascade confirmation

2. **Query hooks** (queries.ts)
   - `useStacks()` - List all stacks with 5s polling
   - `useStack(name)` - Get single stack detail with 5s polling
   - `useTrashStacks()` - List soft-deleted stacks
   - `useDeletePreview(name)` - Get cascade impact before delete

3. **Mutation hooks** (queries.ts)
   - `useCreateStack()` - POST /stacks
   - `useUpdateStack()` - PUT /stacks/{name}
   - `useDeleteStack()` - DELETE /stacks/{name} (soft delete)
   - `useEnableStack()` - POST /stacks/{name}/enable
   - `useDisableStack()` - POST /stacks/{name}/disable
   - `useCloneStack()` - POST /stacks/{name}/clone
   - `useRenameStack()` - POST /stacks/{name}/rename
   - `useRestoreStack()` - POST /stacks/trash/{name}/restore
   - `usePermanentDeleteStack()` - DELETE /stacks/trash/{name}

4. **Navigation** (header.tsx)
   - Added Stacks link with Layers icon
   - Positioned second in nav (after Overview, before Services)

All mutations invalidate query cache on success and show toast notifications. Pattern matches existing service hooks exactly.

## Decisions Made

**1. Follow existing service hooks pattern**
- Rationale: Established pattern in codebase works well, maintains consistency
- Impact: Predictable behavior, familiar to codebase contributors
- Alternatives considered: None - clear pattern to follow

**2. 5-second polling for real-time updates**
- Rationale: Match existing service polling interval, WebSocket extension comes later
- Impact: Real-time feel without WebSocket complexity yet
- Alternatives considered: No polling (stale data), WebSocket (Phase 2 doesn't require it)

**3. Stacks positioned second in navigation**
- Rationale: Stacks are primary workflow in Phase 2+, more important than Services
- Impact: Prominent placement reflects new stack-first architecture
- Alternatives considered: After Services (less visible), configurable order (over-engineering)

## Deviations from Plan

None - plan executed exactly as written.

## Technical Notes

### Type Safety
- All interfaces match Go API response shape from 02-01
- Optional fields (deleted_at, instances) properly typed
- Generic error handling with fallback messages

### Cache Invalidation Strategy
- All mutations invalidate `['stacks']` query key
- Trash operations also invalidate `['stacks', 'trash']`
- Individual stack updates invalidate both specific stack and list

### Error Handling
- Toast notifications for all mutation outcomes
- Error messages from API response or fallback string
- No retry logic (user-initiated mutations should be explicit)

## Verification Results

```bash
✓ TypeScript compilation passes with no new errors
✓ Stack, StackInstance, DeletePreview types exported from api.ts
✓ 13 hooks exported from features/stacks/queries.ts
✓ Stacks navigation link appears in header.tsx
```

## Next Phase Readiness

**Ready for 02-04 (Stack List/Detail UI):**
- All query hooks available for data fetching
- Types properly exported for component consumption
- Navigation link in place for routing

**Blockers:** None

**Concerns:** None - clean data layer foundation

## Task Breakdown

| Task | Description | Files | Commit |
|------|-------------|-------|--------|
| 1 | Add Stack types to api.ts | dashboard/src/types/api.ts | 297c966 |
| 2 | Create stack query and mutation hooks | dashboard/src/features/stacks/queries.ts | fa23e93 |
| 3 | Add Stacks to navigation | dashboard/src/components/layout/header.tsx | b50fbb3 |

**Total:** 3 tasks, 3 commits, 69 seconds

---

*Phase: 02-stack-crud*
*Plan: 03*
*Status: Complete*
*Date: 2026-02-03*
