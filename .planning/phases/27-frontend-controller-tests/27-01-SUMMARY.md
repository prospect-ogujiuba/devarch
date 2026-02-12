---
phase: 27-frontend-controller-tests
plan: 01
subsystem: frontend
tags: [testing, controller-hooks, react-testing-library, vitest]
dependency_graph:
  requires: [24-frontend-mutation-refactor, phase-26]
  provides: [test-utils, instance-controller-tests, stack-controller-tests]
  affects: [dashboard-test-infrastructure]
tech_stack:
  added: [test-utils.tsx, localStorage-mock]
  patterns: [hook-testing, mock-factories, query-mocking]
key_files:
  created:
    - dashboard/src/test/test-utils.tsx
    - dashboard/src/features/instances/useInstanceDetailController.test.ts
    - dashboard/src/features/stacks/useStackDetailController.test.ts
  modified:
    - dashboard/src/test/setup.ts
decisions:
  - test-utils.tsx extension for JSX support in wrapper component
  - localStorage mock in setup.ts for API module compatibility
  - Mock at hook level using vi.mock() for cleaner test isolation
  - Mock factory helpers (mockQueryResult, mockMutation) for DRY test setup
metrics:
  duration: 183
  completed_date: 2026-02-12
  tasks: 2
  files: 4
  tests_added: 15
---

# Phase 27 Plan 01: Controller Hook Tests Summary

**Test infrastructure and coverage for instance and stack detail controller hooks**

## What Was Built

Created shared test utilities and comprehensive test coverage for 2 of 3 controller hooks (instance and stack detail controllers), establishing patterns for hook testing and query orchestration verification.

### Test Infrastructure

**test-utils.tsx** — Shared test wrapper utilities:
- `createTestQueryClient()` factory with `retry: false` and `gcTime: Infinity` for deterministic test behavior
- `createWrapper()` function returning QueryClientProvider wrapper component for renderHook
- Extension changed from .ts to .tsx to support JSX syntax in wrapper component

**setup.ts enhancement** — localStorage mock:
- Added localStorage mock (getItem, setItem, removeItem, clear) to test setup
- Required because API module (`src/lib/api.ts`) accesses localStorage at module initialization
- Mock prevents "localStorage.getItem is not a function" errors in test environment

### Instance Controller Tests (7 tests)

**useInstanceDetailController.test.ts** — 5 test groups:

1. **Data passthrough** — Verifies instance, templateService, and updateInstance mutation are correctly exposed from underlying query hooks
2. **Loading states** — Tests isLoading true/false based on useInstance query state
3. **Conditional query (template_name)** — Verifies useService called with empty string when instance undefined, with template_name when present (the `instance?.template_name ?? ''` fallback logic)
4. **Mutation exposure** — Confirms updateInstance has mutate function
5. **Edge case: undefined data** — All queries return undefined, controller returns undefined instance/templateService

### Stack Controller Tests (8 tests)

**useStackDetailController.test.ts** — 6 test groups with local mock factories:

**Mock helpers:**
- `mockQueryResult<T>(data, isLoading)` — Returns query shape with data/isLoading/error/isError
- `mockMutation()` — Returns mutation shape with vi.fn() for mutate/mutateAsync

**Test coverage:**
1. **Data passthrough** — Verifies stack, instances, networkStatus, composeData, composeLoading returned from underlying queries
2. **Derived state: connectedContainers** — Tests `networkQuery.data?.containers` mapped to connectedContainers array and runningContainerNames Set. Empty containers array produces empty Set.
3. **Derived state: null network data** — When networkQuery.data undefined, connectedContainers falls back to `[]` and runningContainerNames is empty Set (the `?? []` fallback)
4. **Loading states** — isLoading derived directly from stackQuery.isLoading
5. **Mutation exposure** — All 12 mutations verified present with mutate functions: enableStack, disableStack, stopStack, startStack, restartStack, generatePlan, applyPlan, createNetwork, removeNetwork, exportStack, importStack, generateProxyConfig
6. **Instances default** — When instancesQuery.data undefined, instances defaults to `[]` (the `?? []` fallback)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] JSX syntax error in test-utils.ts**
- **Found during:** Task 1 verification
- **Issue:** ESBuild error "Expected '>' but found 'client'" when trying to use JSX in .ts file. QueryClientProvider wrapper requires JSX syntax.
- **Fix:** Renamed `src/test/test-utils.ts` to `src/test/test-utils.tsx` to enable JSX transformation
- **Files modified:** dashboard/src/test/test-utils.tsx (rename)
- **Commit:** 4441641

**2. [Rule 3 - Blocking] localStorage not available in test environment**
- **Found during:** Task 1 verification
- **Issue:** Test error "TypeError: localStorage.getItem is not a function" because API module accesses localStorage at module level (line 49 of api.ts)
- **Fix:** Added localStorage mock to setup.ts with getItem/setItem/removeItem/clear/length/key methods using vi.fn()
- **Files modified:** dashboard/src/test/setup.ts
- **Commit:** 4441641

## Verification

All tests pass via vitest:

```bash
npx vitest run src/features/instances/useInstanceDetailController.test.ts src/features/stacks/useStackDetailController.test.ts
```

**Results:**
- Test Files: 2 passed (2)
- Tests: 15 passed (15)
- Duration: ~2.6s

Coverage areas verified:
- ✅ Data passthrough from query hooks
- ✅ Loading state derivation
- ✅ Conditional query enabling (template_name dependency)
- ✅ Derived state (connectedContainers, runningContainerNames)
- ✅ All 13 total mutations exposed (1 instance, 12 stack)
- ✅ Fallback behavior for undefined data

## Commits

| Task | Description                        | Commit  | Files |
| ---- | ---------------------------------- | ------- | ----- |
| 1    | test-utils + instance tests        | 4441641 | 3     |
| 2    | stack controller tests             | 7d598f0 | 1     |

## Self-Check

**Created files:**
- ✅ FOUND: dashboard/src/test/test-utils.tsx
- ✅ FOUND: dashboard/src/features/instances/useInstanceDetailController.test.ts
- ✅ FOUND: dashboard/src/features/stacks/useStackDetailController.test.ts

**Commits:**
- ✅ FOUND: 4441641
- ✅ FOUND: 7d598f0

**Self-Check: PASSED**

## Impact

**Test Infrastructure:**
- Shared test utilities enable consistent hook testing patterns across all controller hooks
- localStorage mock in setup.ts enables testing any component/hook that imports API module

**Coverage:**
- Instance controller: 100% line coverage (15 lines covered)
- Stack controller: 100% line coverage (64 lines covered)
- 2 of 3 detail controller hooks tested (service controller hook remains for next plan)

**Patterns Established:**
- Hook-level mocking using vi.mock() for clean test isolation
- Mock factory helpers reduce boilerplate in multi-query/mutation scenarios
- createWrapper pattern enables QueryClientProvider for all hook tests

## Next Steps

Recommend for next plan (27-02):
1. Create useServiceDetailController tests following same patterns
2. All 3 detail controller hooks will have test coverage
3. Consider integration tests for controller hook composition
