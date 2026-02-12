---
phase: 27-frontend-controller-tests
plan: 02
subsystem: frontend
tags: [testing, controller-hooks, vitest, ci-workflow]
dependency_graph:
  requires: [27-01, phase-26]
  provides: [service-controller-tests, dashboard-ci-workflow, complete-controller-test-coverage]
  affects: [dashboard-test-infrastructure, ci-pipeline]
tech_stack:
  added: [dashboard-tests.yml]
  patterns: [hook-testing, mock-factories, github-actions]
key_files:
  created:
    - dashboard/src/features/services/useServiceDetailController.test.ts
    - .github/workflows/dashboard-tests.yml
  modified: []
decisions:
  - Service controller tests follow same patterns as instance/stack tests (hook-level mocking)
  - CI workflow triggers only on dashboard/** path changes for efficiency
  - All 3 controller hooks now have complete test coverage (TEST-03 complete)
metrics:
  duration: 125
  completed_date: 2026-02-12
  tasks: 2
  files: 2
  tests_added: 14
---

# Phase 27 Plan 02: Service Controller Tests & CI Summary

**Complete controller hook test coverage and CI workflow for dashboard unit tests**

## What Was Built

Created comprehensive test coverage for service detail controller (richest derived state logic) and GitHub Actions workflow to run dashboard tests on PRs.

### Service Controller Tests (14 tests)

**useServiceDetailController.test.ts** — 9 test groups covering all derived state values:

**Mock factory:**
- `mockService(overrides)` — Returns full Service object with status and metrics populated
- Default values: nginx:latest, running status, restart_count: 0, 6 metrics populated
- Overrides support Partial<Service> for test-specific variations

**Test coverage:**

1. **Data passthrough** — Verifies service, composeYaml, isLoading, composeLoading returned from underlying query hooks (useService, useServiceCompose)

2. **Derived state: status** — Two cases:
   - Service has `status.status: 'running'` → returns `'running'`
   - Service has no status object → returns `'stopped'` (fallback)

3. **Derived state: image** — Two cases:
   - Service present with `image_name: 'nginx', image_tag: 'alpine'` → returns `'nginx:alpine'`
   - Service undefined → returns `''` (empty string)

4. **Derived state: healthStatus** — Three cases covering fallback logic:
   - Service has `status.health_status: 'healthy'` → returns `'healthy'`
   - No health_status but healthcheck object exists → returns `'configured'`
   - No health_status and no healthcheck → returns `'none'`

5. **Derived state: uptime** — Mocks computeUptime to return 3600, verifies:
   - `result.current.uptime === 3600`
   - `computeUptime` called with `service.status.started_at`

6. **Derived state: metrics** — All 5 metric values verified when metrics object present:
   - cpuPct: 45.2
   - memUsed: 128
   - memLimit: 512
   - rxBytes: 1024
   - txBytes: 2048

7. **Metrics defaults when undefined** — Service with no metrics object:
   - All 5 metric values default to 0 (cpuPct, memUsed, memLimit, rxBytes, txBytes)
   - Verifies the `?? 0` fallback logic for each metric field

8. **Loading states** — Two cases:
   - useService returns `isLoading: true` → `result.current.isLoading === true`
   - useServiceCompose returns `isLoading: true` → `result.current.composeLoading === true`

9. **Mutation exposure** — Three mutations verified present with mutate functions:
   - deleteService
   - updateService
   - generateProxyConfig

### CI Workflow

**.github/workflows/dashboard-tests.yml** — GitHub Actions workflow:

**Trigger:**
- `pull_request` event
- `paths: ['dashboard/**']` — only runs when dashboard code changes

**Job: unit-tests**
- `runs-on: ubuntu-latest`
- `actions/checkout@v4` — checkout repo
- `actions/setup-node@v4` — Node 20 with npm cache (cache-dependency-path: dashboard/package-lock.json)
- `npm ci` in dashboard/ — deterministic dependency install
- `npm run test:unit` — runs vitest unit tests

**Pattern:** Matches integration-tests.yml structure for consistency

## Deviations from Plan

None — plan executed exactly as written.

## Verification

All tests pass via vitest:

```bash
cd dashboard && npm run test:unit
```

**Results:**
- Test Files: 4 passed (4)
- Tests: 32 passed (32) — 14 service + 7 instance + 8 stack + 3 entity-actions
- Duration: ~9.7s

**CI workflow:**
- File exists at `.github/workflows/dashboard-tests.yml`
- YAML structure valid
- Triggers on `pull_request` with `paths: ['dashboard/**']`
- Runs `npm run test:unit` in dashboard working directory

## Commits

| Task | Description                        | Commit  | Files |
| ---- | ---------------------------------- | ------- | ----- |
| 1    | Service controller tests           | 127f439 | 1     |
| 2    | CI workflow for dashboard tests    | a64f9d0 | 1     |

## Self-Check

**Created files:**
- ✅ FOUND: dashboard/src/features/services/useServiceDetailController.test.ts
- ✅ FOUND: .github/workflows/dashboard-tests.yml

**Commits:**
- ✅ FOUND: 127f439
- ✅ FOUND: a64f9d0

**Self-Check: PASSED**

## Impact

**Test Coverage Complete:**
- All 3 controller hooks tested: instance (7), stack (8), service (14)
- TEST-03 requirement complete: "Controller hooks have unit test coverage"
- Total controller tests: 29 tests covering data passthrough, derived state, loading, mutations

**Derived State Verification:**
- Service controller has richest derived state: 9 values (status, image, healthStatus, uptime, cpuPct, memUsed, memLimit, rxBytes, txBytes)
- All fallback logic tested (status → 'stopped', image → '', healthStatus → 'configured'/'none', metrics → 0)
- Metric defaults critical for UI stability when container not reporting metrics

**CI Integration:**
- Dashboard tests now run automatically on PRs touching dashboard code
- Catches regressions before merge
- Path-based triggering avoids running tests on API-only changes

**Patterns Established:**
- Mock at hook level (vi.mock) for clean isolation from query implementation
- Mock factory pattern reduces boilerplate, improves readability
- Consistent test structure across all 3 controller hooks

## Next Steps

Phase 27 complete. Recommend for Phase 28:
1. Test coverage for remaining dashboard features (if any)
2. Integration tests for controller hook composition
3. E2E tests for critical user flows
