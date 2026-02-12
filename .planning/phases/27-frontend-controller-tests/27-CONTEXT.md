# Phase 27: Frontend Controller Tests - Context

**Gathered:** 2026-02-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Test coverage for 3 controller hooks created in Phase 24: useStackDetailController, useInstanceDetailController, useServiceDetailController. Tests verify query orchestration, state derivation, loading states, and action handler delegation. No new features — test-only phase.

</domain>

<decisions>
## Implementation Decisions

### Mock Strategy
- Mock at hook level using vi.mock() on query/mutation hook modules (useStack, useInstances, etc.)
- NOT API-level mocking (no MSW) — controllers orchestrate hooks, not API calls
- Each mock returns configurable query/mutation result objects (data, isLoading, error states)

### Test Infrastructure
- Shared test wrapper in src/test/test-utils.ts providing renderHook with fresh QueryClient
- No TanStack Router provider needed — controllers don't use router
- Use @testing-library/react's built-in renderHook (v16)

### File Organization
- Colocated with source: useStackDetailController.test.ts next to useStackDetailController.ts
- Follows existing pattern (entity-actions.test.tsx is colocated)

### Test Scope per Controller
- **Data passthrough:** Mock query returns flow to correct return properties
- **Derived state:** connectedContainers, runningContainerNames, status, image, healthStatus, uptime, metrics compute correctly
- **Loading states:** isLoading true when any query loading; false when all resolved
- **Mutation exposure:** All expected mutation objects present (verify existence, not internal behavior)
- **Edge cases:** undefined/null data, conditional query enabling (instance controller template_name dependency)

### Stack Controller Page Typos
- Fix ctrl.ctrl.* and rectrl.* typos as pre-task before writing tests
- Phase 24 verification gap, not new capability

### CI Integration
- Use existing npm run test:unit (vitest run)
- Ensure CI workflow runs dashboard tests (Phase 26 added API test CI — may need parallel dashboard step)

### Claude's Discretion
- Exact mock factory shape and helper utilities
- Test assertion granularity (how many derived state edge cases)
- Whether to group tests by concern (data/loading/mutations) or by scenario
- No snapshot testing — pure assertion-based

</decisions>

<specifics>
## Specific Ideas

- useMutationHelper is NOT tested directly — it's infrastructure, not a controller
- Controller tests should prove the orchestration contract: "given these query states, controller returns these values"
- Instance controller is smallest (15 lines, 2 queries, 1 mutation) — good starting point

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 27-frontend-controller-tests*
*Context gathered: 2026-02-12*
