# Phase 27: Frontend Controller Tests - Research

**Researched:** 2026-02-12
**Domain:** Vitest + React Testing Library + TanStack Query controller hook testing
**Confidence:** HIGH

## Summary

Phase 27 adds test coverage for three controller hooks created in Phase 24: `useStackDetailController`, `useInstanceDetailController`, and `useServiceDetailController`. These hooks orchestrate multiple TanStack Query hooks (queries + mutations), derive state from query results, and expose action handlers. Tests verify the orchestration contract: "given these query states, controller returns these values."

The testing approach mocks at the hook level using `vi.mock()` on query/mutation modules (not API-level mocking with MSW), uses `@testing-library/react`'s `renderHook` (v16+) with a QueryClient wrapper, and validates data passthrough, derived state computation, loading state aggregation, and mutation object presence. The existing test infrastructure (Vitest 3.2.4, jsdom environment, colocated test pattern) is already configured and ready to use.

**Primary recommendation:** Mock query/mutation hook modules with configurable return values, create reusable test wrapper providing fresh QueryClient, test orchestration logic without exercising actual API calls or mutation side effects.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
**Mock Strategy:**
- Mock at hook level using vi.mock() on query/mutation hook modules (useStack, useInstances, etc.)
- NOT API-level mocking (no MSW) — controllers orchestrate hooks, not API calls
- Each mock returns configurable query/mutation result objects (data, isLoading, error states)

**Test Infrastructure:**
- Shared test wrapper in src/test/test-utils.ts providing renderHook with fresh QueryClient
- No TanStack Router provider needed — controllers don't use router
- Use @testing-library/react's built-in renderHook (v16)

**File Organization:**
- Colocated with source: useStackDetailController.test.ts next to useStackDetailController.ts
- Follows existing pattern (entity-actions.test.tsx is colocated)

**Test Scope per Controller:**
- Data passthrough: Mock query returns flow to correct return properties
- Derived state: connectedContainers, runningContainerNames, status, image, healthStatus, uptime, metrics compute correctly
- Loading states: isLoading true when any query loading; false when all resolved
- Mutation exposure: All expected mutation objects present (verify existence, not internal behavior)
- Edge cases: undefined/null data, conditional query enabling (instance controller template_name dependency)

**Stack Controller Page Typos:**
- Fix ctrl.ctrl.* and rectrl.* typos as pre-task before writing tests
- Phase 24 verification gap, not new capability

**CI Integration:**
- Use existing npm run test:unit (vitest run)
- Ensure CI workflow runs dashboard tests (Phase 26 added API test CI — may need parallel dashboard step)

### Claude's Discretion
- Exact mock factory shape and helper utilities
- Test assertion granularity (how many derived state edge cases)
- Whether to group tests by concern (data/loading/mutations) or by scenario
- No snapshot testing — pure assertion-based

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| vitest | 3.2.4 | Test runner | Fast, Vite-native, Jest-compatible API with native ESM support |
| @testing-library/react | 16.3.0 | React testing utilities | Industry standard for testing React components and hooks, renderHook built-in for React 18+ |
| jsdom | 27.2.0 | DOM environment | Provides browser APIs for Node.js test environment |
| @tanstack/react-query | 5.90.19 | State management | Query/mutation hooks being tested |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @testing-library/user-event | 14.6.1 | User interaction simulation | Not needed for controller tests (no UI interaction) |
| @testing-library/jest-dom | 6.9.1 | DOM matchers | Provides vitest expect extensions like toBeInTheDocument |

**Installation:**
No new dependencies required. All testing infrastructure already configured in `vite.config.ts` and `src/test/setup.ts`.

## Architecture Patterns

### Recommended Project Structure
```
dashboard/src/
├── features/
│   ├── stacks/
│   │   ├── queries.ts                          # Query/mutation hooks
│   │   ├── useStackDetailController.ts         # Controller hook
│   │   └── useStackDetailController.test.ts    # NEW: Controller tests
│   ├── instances/
│   │   ├── queries.ts
│   │   ├── useInstanceDetailController.ts
│   │   └── useInstanceDetailController.test.ts # NEW
│   └── services/
│       ├── queries.ts
│       ├── useServiceDetailController.ts
│       └── useServiceDetailController.test.ts  # NEW
└── test/
    ├── setup.ts                               # Existing: imports jest-dom
    └── test-utils.ts                          # NEW: renderHook wrapper with QueryClient
```

### Pattern 1: Hook-Level Mocking with vi.mock()
**What:** Mock entire query/mutation hook modules to return configurable test data, avoiding API calls and mutation side effects.

**When to use:** Testing controller hooks that orchestrate multiple query/mutation hooks from feature modules.

**Example:**
```typescript
// useStackDetailController.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useStackDetailController } from './useStackDetailController'
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query'

// Mock the queries module - hoisted to top automatically
vi.mock('@/features/stacks/queries', () => ({
  useStack: vi.fn(),
  useInstances: vi.fn(),
  useStackNetwork: vi.fn(),
  useStackCompose: vi.fn(),
  useEnableStack: vi.fn(),
  useDisableStack: vi.fn(),
  // ... other mutations
}))

vi.mock('@/features/instances/queries', () => ({
  useInstances: vi.fn(),
}))

describe('useStackDetailController', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns loading state when queries are loading', () => {
    const { useStack } = await import('@/features/stacks/queries')

    vi.mocked(useStack).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as UseQueryResult)

    const { result } = renderHook(() => useStackDetailController('test-stack'), {
      wrapper: createQueryWrapper(),
    })

    expect(result.current.isLoading).toBe(true)
  })
})
```

### Pattern 2: Test Wrapper with Fresh QueryClient
**What:** Wrapper component providing QueryClientProvider with test-specific QueryClient configuration.

**When to use:** All `renderHook` calls testing hooks that depend on TanStack Query.

**Example:**
```typescript
// test/test-utils.ts
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,           // Disable retries to prevent timeouts
        gcTime: Infinity,       // Keep data in cache for test duration
      },
      mutations: {
        retry: false,
      },
    },
  })
}

export function createQueryWrapper() {
  const queryClient = createTestQueryClient()
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}
```

**Source:** [TanStack Query Testing Guide](https://tanstack.com/query/v5/docs/react/guides/testing), [Testing React Query by TkDodo](https://tkdodo.eu/blog/testing-react-query)

### Pattern 3: Derived State Validation
**What:** Test that controller hooks correctly compute derived values from query data.

**When to use:** When controllers compute status, aggregate data, or transform query results.

**Example:**
```typescript
it('computes derived state from query data', () => {
  const { useStack, useStackNetwork } = await import('@/features/stacks/queries')

  vi.mocked(useStack).mockReturnValue({
    data: { name: 'test-stack', enabled: true },
    isLoading: false,
  } as UseQueryResult)

  vi.mocked(useStackNetwork).mockReturnValue({
    data: { containers: ['container1', 'container2'] },
    isLoading: false,
  } as UseQueryResult)

  const { result } = renderHook(() => useStackDetailController('test-stack'), {
    wrapper: createQueryWrapper(),
  })

  expect(result.current.connectedContainers).toEqual(['container1', 'container2'])
  expect(result.current.runningContainerNames).toBeInstanceOf(Set)
  expect(result.current.runningContainerNames.size).toBe(2)
})
```

### Pattern 4: Mutation Object Presence Verification
**What:** Verify that controller exposes all expected mutation objects without testing internal behavior.

**When to use:** Controllers that aggregate multiple mutations for consumer convenience.

**Example:**
```typescript
it('exposes all expected mutations', () => {
  const { useEnableStack, useDisableStack, useStartStack } = await import('@/features/stacks/queries')

  // Mock mutations to return mutation result objects
  vi.mocked(useEnableStack).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as UseMutationResult)

  vi.mocked(useDisableStack).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as UseMutationResult)

  vi.mocked(useStartStack).mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  } as unknown as UseMutationResult)

  const { result } = renderHook(() => useStackDetailController('test-stack'), {
    wrapper: createQueryWrapper(),
  })

  expect(result.current.enableStack).toBeDefined()
  expect(result.current.disableStack).toBeDefined()
  expect(result.current.startStack).toBeDefined()
  expect(result.current.enableStack.mutate).toBeInstanceOf(Function)
})
```

**Source:** [Vitest Expect API](https://vitest.dev/api/expect), [Testing React hooks with Vitest](https://mayashavin.com/articles/test-react-hooks-with-vitest)

### Anti-Patterns to Avoid
- **Testing mutation side effects:** Don't test toast notifications, query invalidations, or API calls — these are tested by mutation hooks themselves
- **Snapshot testing:** Avoid snapshots for controller return objects — they're brittle and don't validate logic
- **Shared mock state:** Don't reuse mock return values across tests — creates coupling and flaky tests
- **Testing implementation details:** Don't test internal helper functions or hook call order — test the public API contract

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Query client test setup | Custom test provider/context management | `createTestQueryClient()` factory with `retry: false` | Prevents test timeouts from query retries, handles cache isolation |
| Module mocking utilities | Custom mock factories for each hook | `vi.mock()` with factory functions + `vi.mocked()` | Vitest handles hoisting, type inference, and mock state management |
| Async hook testing | Manual promise resolution tracking | `renderHook` from @testing-library/react | Handles React rendering lifecycle, provides `result.current` interface |
| Mock cleanup | Manual `mockClear()` calls per test | `vi.clearAllMocks()` in `beforeEach` | Clears all mocks globally, prevents dirty state between tests |

**Key insight:** Controller hooks are pure orchestration — no need for MSW, no-op container clients, or complex setup. Mock at the boundary (query/mutation hooks), validate outputs.

## Common Pitfalls

### Pitfall 1: Query Retry Timeouts
**What goes wrong:** Tests hang or timeout waiting for failed queries to retry 3 times with exponential backoff (TanStack Query default).

**Why it happens:** Test QueryClient uses production defaults including retry logic.

**How to avoid:** Create test-specific QueryClient with `retry: false` in `defaultOptions`.

**Warning signs:** Tests timeout after 2+ seconds, console shows retry warnings.

**Example:**
```typescript
// ❌ BAD: Uses production defaults
const queryClient = new QueryClient()

// ✅ GOOD: Disables retries for tests
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
})
```

**Source:** [TanStack Query Testing Guide](https://tanstack.com/query/v4/docs/react/guides/testing)

### Pitfall 2: Mock State Pollution
**What goes wrong:** Tests pass in isolation but fail when run together; changing test order breaks tests.

**Why it happens:** Mock return values from previous tests leak into subsequent tests via shared module state.

**How to avoid:** Use `vi.clearAllMocks()` in `beforeEach` to reset all mock state between tests.

**Warning signs:** Tests fail in suite but pass individually, random failures based on test order.

**Example:**
```typescript
// ❌ BAD: Mock state persists between tests
describe('controller tests', () => {
  it('test 1', () => {
    vi.mocked(useStack).mockReturnValue({ data: { name: 'stack1' } })
    // test...
  })

  it('test 2', () => {
    // useStack still returns stack1 from test 1!
    // test...
  })
})

// ✅ GOOD: Clear mocks between tests
describe('controller tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('test 1', () => {
    vi.mocked(useStack).mockReturnValue({ data: { name: 'stack1' } })
    // test...
  })

  it('test 2', () => {
    vi.mocked(useStack).mockReturnValue({ data: { name: 'stack2' } })
    // test starts clean
  })
})
```

**Source:** [Vitest Mocking Guide](https://vitest.dev/guide/mocking), [Best practices discussion](https://github.com/vitest-dev/vitest/discussions/4224)

### Pitfall 3: Testing Mutation Behavior Instead of Presence
**What goes wrong:** Tests call `mutate()` and assert on side effects (toasts, invalidations), coupling tests to mutation implementation.

**Why it happens:** Natural instinct to test "does it work" rather than "does it expose the API."

**How to avoid:** Verify mutation objects exist and have expected shape (`mutate` function, `isPending` state), but don't call them. Mutation behavior is tested in `queries.test.ts`.

**Warning signs:** Tests importing `toast` or `queryClient.invalidateQueries`, mocking side effects.

**Example:**
```typescript
// ❌ BAD: Testing mutation side effects
it('calls mutation and shows toast', async () => {
  const { result } = renderHook(...)
  await result.current.enableStack.mutate('test-stack')
  expect(toast.success).toHaveBeenCalledWith('Enabled test-stack')
})

// ✅ GOOD: Verify mutation exists and is callable
it('exposes enableStack mutation', () => {
  const { result } = renderHook(...)
  expect(result.current.enableStack).toBeDefined()
  expect(result.current.enableStack.mutate).toBeInstanceOf(Function)
  expect(result.current.enableStack.isPending).toBe(false)
})
```

### Pitfall 4: Forgetting Conditional Query Dependencies
**What goes wrong:** Tests fail with "query is disabled" or don't exercise conditional logic.

**Why it happens:** Controllers may enable queries conditionally (e.g., `useService` depends on `instance?.template_name`), but mocks don't respect this.

**How to avoid:** When testing conditional queries, mock prerequisite data to enable dependent queries, or explicitly test disabled state.

**Warning signs:** Queries unexpectedly return undefined, `enabled: false` not being tested.

**Example:**
```typescript
// useInstanceDetailController enables useService conditionally:
// useService(instance?.template_name ?? '', { enabled: !!instance?.template_name })

// ❌ BAD: Doesn't mock instance data, useService never enabled
it('loads template service', () => {
  vi.mocked(useInstance).mockReturnValue({ data: undefined })
  // useService won't run, template_name is undefined
})

// ✅ GOOD: Mocks prerequisite data to enable conditional query
it('loads template service when instance has template_name', () => {
  vi.mocked(useInstance).mockReturnValue({
    data: { template_name: 'nginx' },
    isLoading: false,
  })

  vi.mocked(useService).mockReturnValue({
    data: { name: 'nginx', image_name: 'nginx' },
    isLoading: false,
  })

  const { result } = renderHook(...)
  expect(result.current.templateService).toEqual({ name: 'nginx', ... })
})

// Also test disabled state
it('does not query service when instance has no template_name', () => {
  vi.mocked(useInstance).mockReturnValue({ data: null, isLoading: false })
  const { result } = renderHook(...)
  expect(result.current.templateService).toBeUndefined()
})
```

### Pitfall 5: vi.mock() Hoisting Gotchas
**What goes wrong:** Mocks don't work as expected, imports resolve before mocks are set up.

**Why it happens:** `vi.mock()` is hoisted to the top of the file before all imports, but developers expect sequential execution.

**How to avoid:** Call `vi.mock()` at module level (not inside describe/it), use dynamic imports (`await import()`) in tests when you need to access mocked exports.

**Warning signs:** Mocks not applied, "X is not a function" errors, imports resolving to real modules.

**Example:**
```typescript
// ❌ BAD: Mock inside describe, imports happen first
describe('controller', () => {
  vi.mock('@/features/stacks/queries') // Too late, already imported

  it('test', () => {
    // useStack is the real function, not mocked
  })
})

// ✅ GOOD: Mock at module level, dynamic import in test
import { renderHook } from '@testing-library/react'
import { useStackDetailController } from './useStackDetailController'

// Mock is hoisted automatically
vi.mock('@/features/stacks/queries', () => ({
  useStack: vi.fn(),
  useInstances: vi.fn(),
}))

describe('controller', () => {
  it('test', async () => {
    const { useStack } = await import('@/features/stacks/queries')
    vi.mocked(useStack).mockReturnValue({ data: {...} })
    // Now mock is applied
  })
})
```

**Source:** [Vitest Module Mocking](https://vitest.dev/guide/mocking/modules), [Hoisting behavior](https://v2.vitest.dev/guide/mocking)

## Code Examples

Verified patterns from official sources and codebase analysis:

### Complete Controller Test Suite Structure
```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useStackDetailController } from './useStackDetailController'
import { createQueryWrapper } from '@/test/test-utils'
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query'

// 1. Mock all dependencies at module level (hoisted automatically)
vi.mock('@/features/stacks/queries')
vi.mock('@/features/instances/queries')
vi.mock('@/features/proxy/queries')

describe('useStackDetailController', () => {
  // 2. Clear mocks before each test for isolation
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('data passthrough', () => {
    it('returns stack data from useStack query', async () => {
      const { useStack } = await import('@/features/stacks/queries')

      const mockStack = { name: 'test-stack', enabled: true }
      vi.mocked(useStack).mockReturnValue({
        data: mockStack,
        isLoading: false,
        error: null,
      } as UseQueryResult)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.stack).toEqual(mockStack)
    })
  })

  describe('derived state', () => {
    it('computes runningContainerNames Set from network data', async () => {
      const { useStackNetwork } = await import('@/features/stacks/queries')

      vi.mocked(useStackNetwork).mockReturnValue({
        data: { containers: ['web', 'db', 'cache'] },
        isLoading: false,
      } as UseQueryResult)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.runningContainerNames).toBeInstanceOf(Set)
      expect(result.current.runningContainerNames.size).toBe(3)
      expect(result.current.runningContainerNames.has('web')).toBe(true)
    })
  })

  describe('loading states', () => {
    it('aggregates isLoading from stack query', async () => {
      const { useStack } = await import('@/features/stacks/queries')

      vi.mocked(useStack).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as UseQueryResult)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.isLoading).toBe(true)
    })

    it('returns false when all queries resolved', async () => {
      const { useStack } = await import('@/features/stacks/queries')

      vi.mocked(useStack).mockReturnValue({
        data: { name: 'test-stack' },
        isLoading: false,
      } as UseQueryResult)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('mutation exposure', () => {
    it('exposes all stack lifecycle mutations', async () => {
      const {
        useEnableStack,
        useDisableStack,
        useStartStack,
        useStopStack,
        useRestartStack,
      } = await import('@/features/stacks/queries')

      // Mock all mutations to return mutation result shape
      const mockMutation = {
        mutate: vi.fn(),
        mutateAsync: vi.fn(),
        isPending: false,
      } as unknown as UseMutationResult

      vi.mocked(useEnableStack).mockReturnValue(mockMutation)
      vi.mocked(useDisableStack).mockReturnValue(mockMutation)
      vi.mocked(useStartStack).mockReturnValue(mockMutation)
      vi.mocked(useStopStack).mockReturnValue(mockMutation)
      vi.mocked(useRestartStack).mockReturnValue(mockMutation)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.enableStack).toBeDefined()
      expect(result.current.disableStack).toBeDefined()
      expect(result.current.startStack).toBeDefined()
      expect(result.current.stopStack).toBeDefined()
      expect(result.current.restartStack).toBeDefined()
    })
  })

  describe('edge cases', () => {
    it('handles undefined network data gracefully', async () => {
      const { useStackNetwork } = await import('@/features/stacks/queries')

      vi.mocked(useStackNetwork).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as UseQueryResult)

      const { result } = renderHook(
        () => useStackDetailController('test-stack'),
        { wrapper: createQueryWrapper() }
      )

      expect(result.current.connectedContainers).toEqual([])
      expect(result.current.runningContainerNames.size).toBe(0)
    })
  })
})
```

### Test Utils with QueryClient Wrapper
```typescript
// test/test-utils.ts
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,        // Prevent retry timeouts in tests
        gcTime: Infinity,    // Keep cache during test
      },
      mutations: {
        retry: false,
      },
    },
    logger: {
      log: () => {},        // Silence logs in tests
      warn: () => {},
      error: () => {},
    },
  })
}

export function createQueryWrapper() {
  const queryClient = createTestQueryClient()

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }
}

// Alternative: Export renderHook wrapper for convenience
export function renderHookWithQuery<TResult, TProps>(
  hook: (props: TProps) => TResult,
  options?: { initialProps?: TProps }
) {
  return renderHook(hook, {
    ...options,
    wrapper: createQueryWrapper(),
  })
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| @testing-library/react-hooks package | renderHook in @testing-library/react | React 18 (2022) | Simpler imports, fewer dependencies, built-in support |
| jest.mock() | vi.mock() | Vitest migration | Same API, faster execution, native ESM support |
| Manual QueryClient setup | Factory pattern with retry: false | TanStack Query v4+ (2023) | Prevents test timeouts from query retries |
| Testing mutation side effects | Testing mutation presence | Controller pattern (2024+) | Separates orchestration tests from behavior tests |

**Deprecated/outdated:**
- **@testing-library/react-hooks:** Deprecated in favor of built-in `renderHook` in @testing-library/react v13+
- **MSW for controller tests:** Overkill when mocking at hook level — reserve for integration/E2E tests
- **jest.resetModules():** Not needed with Vitest's ESM-native module system

## Open Questions

1. **Should service detail controller mock containerStatus derivation edge cases?**
   - What we know: `useServiceDetailController` derives `status`, `healthStatus`, `uptime` from `service?.status` and `service?.metrics`
   - What's unclear: How many null/undefined permutations to test (all null, partial data, etc.)
   - Recommendation: Test happy path (all data present) + one null/undefined case per derived field — comprehensive coverage without combinatorial explosion

2. **Should we test the order of query hook calls?**
   - What we know: Controllers call multiple hooks in sequence (e.g., `useStack`, `useInstances`, `useStackNetwork`)
   - What's unclear: Is call order part of the contract, or implementation detail?
   - Recommendation: Don't test call order — it's implementation detail. Test outputs only.

3. **How granular should loading state tests be?**
   - What we know: Controllers aggregate `isLoading` from one or more queries
   - What's unclear: Test every query loading independently, or just "any loading" and "all resolved"?
   - Recommendation: For stack controller (largest): test 2-3 query loading states. For instance controller (smallest): test single query loading state. Balance coverage with maintainability.

## Sources

### Primary (HIGH confidence)
- [Vitest Mocking Guide](https://vitest.dev/guide/mocking) - Module mocking, vi.mock(), hoisting behavior
- [Vitest Expect API](https://vitest.dev/api/expect) - toBeDefined, assertion matchers
- [TanStack Query Testing Guide](https://tanstack.com/query/v4/docs/react/guides/testing) - QueryClient wrapper pattern, retry configuration
- [Testing Library API](https://testing-library.com/docs/react-testing-library/api/) - renderHook with wrapper option
- Existing codebase: `dashboard/src/components/ui/entity-actions.test.tsx` - Colocated test pattern with Vitest

### Secondary (MEDIUM confidence)
- [Testing React Query by TkDodo](https://tkdodo.eu/blog/testing-react-query) - Practical patterns for testing TanStack Query hooks
- [Test React hooks with Vitest efficiently](https://mayashavin.com/articles/test-react-hooks-with-vitest) - renderHook usage, mock cleanup patterns
- [How To Test React Query Hook using Jest](https://medium.com/@ghewadesumit/how-to-test-react-query-hook-using-jest-11d01a0a0acd) - QueryClient wrapper examples (Jest, but pattern applies to Vitest)
- [vi.spyOn vs vi.mock discussion](https://github.com/vitest-dev/vitest/discussions/4224) - Best practices for mocking strategies

### Tertiary (LOW confidence)
- [GitHub Actions Parallel Testing](https://www.testmo.com/guides/github-actions-parallel-testing/) - CI workflow patterns for parallel jobs
- [Vitest Coverage Configuration](https://vitest.dev/guide/coverage) - Coverage reporting, thresholds, CI integration

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already installed and configured
- Architecture patterns: HIGH - Verified with official Vitest/Testing Library docs and existing test file
- Pitfalls: HIGH - Documented in official Vitest/TanStack Query guides with examples
- CI integration: MEDIUM - Need to verify existing workflow includes dashboard tests

**Research date:** 2026-02-12
**Valid until:** 2026-03-15 (30 days for stable testing stack)
