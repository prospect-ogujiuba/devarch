---
phase: 25-websocket-expansion-frontend-auth
plan: 01
subsystem: ui
tags: [websocket, react-query, tanstack-query, invalidation, real-time]

# Dependency graph
requires:
  - phase: 18-api-key-websocket-auth
    provides: WS auth token flow (fetchWSToken + ?token= query param)
  - phase: 24-frontend-controller-extraction
    provides: Query structure patterns for stacks and instances
provides:
  - WebSocket status messages invalidate all stack and instance queries
  - Real-time stack detail page updates on container status changes
  - Real-time instance detail page updates on container status changes
affects: [future-websocket-message-types, real-time-ui-features]

# Tech tracking
tech-stack:
  added: []
  patterns: [predicate-based query invalidation for nested query keys]

key-files:
  created: []
  modified: [dashboard/src/hooks/use-websocket.ts]

key-decisions:
  - "Single predicate covers all stack and instance queries via ['stacks', ...] prefix"
  - "Predicate-based invalidation used instead of multiple specific queryKey calls"

patterns-established:
  - "Predicate invalidation for query hierarchies: key[0] === 'stacks' covers ['stacks'], ['stacks', name], ['stacks', name, 'instances'], etc."

# Metrics
duration: 32s
completed: 2026-02-12
---

# Phase 25 Plan 01: WebSocket Stack/Instance Invalidation Summary

**WebSocket status messages now invalidate all stack and instance queries via single predicate, enabling real-time UI updates for stack/instance detail pages**

## Performance

- **Duration:** 32s
- **Started:** 2026-02-12T13:39:11Z
- **Completed:** 2026-02-12T13:39:43Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Added predicate-based invalidation for all `['stacks', ...]` query keys to WS status handler
- Covers stack queries: list, detail, network, compose, wires, trash
- Covers instance queries: list, detail, effective-config, resources
- Preserved existing service/status/category/metrics invalidation unchanged

## Task Commits

Each task was committed atomically:

1. **Task 1: Add stacks predicate invalidation to WS status handler** - `3d92306` (feat)

## Files Created/Modified
- `dashboard/src/hooks/use-websocket.ts` - Added predicate invalidation for all stack and instance queries

## Decisions Made
- Used single predicate `key[0] === 'stacks'` to cover all stack and instance queries - consolidates invalidation logic instead of separate calls for each query key pattern
- Verified Phase 18 WS auth (fetchWSToken + ?token= param) still intact - no auth changes needed

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

WebSocket invalidation complete for stacks and instances (FE-05 closed). Ready for additional WebSocket message types or real-time features in future phases.

---
*Phase: 25-websocket-expansion-frontend-auth*
*Completed: 2026-02-12*


## Self-Check: PASSED

All files and commits verified.
