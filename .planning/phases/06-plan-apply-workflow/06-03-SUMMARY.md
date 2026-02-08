---
phase: 06-plan-apply-workflow
plan: 03
subsystem: ui
tags: [react, tanstack-query, typescript, dashboard, deploy-ui]

# Dependency graph
requires:
  - phase: 06-02
    provides: Plan/apply HTTP endpoints with structured diff and staleness tokens
provides:
  - Deploy tab UI for generating and applying stack plans
  - Color-coded diff visualization (adds/modifies/removes)
  - Staleness token handling with user feedback
affects: [phase-07, phase-08]

# Tech tracking
tech-stack:
  added: []
  patterns: [terraform-style-deploy-ux, ephemeral-plan-state, color-coded-diffs]

key-files:
  created: []
  modified:
    - dashboard/src/types/api.ts
    - dashboard/src/features/stacks/queries.ts
    - dashboard/src/routes/stacks/$name.tsx

key-decisions:
  - "Ephemeral plan state (useState, not cached) — plans are one-shot, not persisted"
  - "Color-coded border-left for visual hierarchy (green/yellow/red)"
  - "Per-field diff for modifications shows old -> new inline"

patterns-established:
  - "Terraform-style UX: explicit generate → preview → apply workflow"
  - "Staleness token validated server-side, 409 error shows regenerate message"
  - "Clear plan state after successful apply (user must regenerate)"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 6 Plan 3: Dashboard Deploy Tab Summary

**Deploy tab with Terraform-style plan preview, color-coded diffs (add/modify/remove), and staleness-aware apply execution**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T23:20:23Z
- **Completed:** 2026-02-07T23:22:03Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Deploy tab alongside Instances and Compose tabs on stack detail page
- Generate Plan button calls GET /stacks/{name}/plan and displays structured diff
- Color-coded changes: green for adds, yellow for modifications, red for removals
- Per-field modification detail (old value -> new value)
- Apply button sends staleness token, clears plan on success, handles 409 conflicts

## Task Commits

Each task was committed atomically:

1. **Task 1: Dashboard types and query hooks** - `1182bf15` (feat)
2. **Task 2: Deploy tab with diff visualization** - `4e290f77` (feat)

## Files Created/Modified
- `dashboard/src/types/api.ts` - Added PlanFieldChange, PlanChange, StackPlan, ApplyResult types
- `dashboard/src/features/stacks/queries.ts` - Added useGeneratePlan and useApplyPlan mutations with 409 handling
- `dashboard/src/routes/stacks/$name.tsx` - Added Deploy tab with diff preview and apply execution

## Decisions Made
- **Ephemeral plan storage:** currentPlan in useState, not query cache — plans are one-shot previews, not cacheable resources
- **Color-coded diffs:** Border-left visual hierarchy (green/yellow/red) aligns with standard diff conventions
- **Apply clears plan:** onSuccess sets currentPlan to null forcing regenerate before next apply (prevents stale token reuse)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 6 complete: plan/apply workflow fully functional
- UI surfaces structured diffs with per-field detail
- Staleness protection prevents concurrent modification conflicts
- Ready for Phase 7 (Wiring) — dependency wiring can reference this deploy workflow

---
*Phase: 06-plan-apply-workflow*
*Completed: 2026-02-07*
