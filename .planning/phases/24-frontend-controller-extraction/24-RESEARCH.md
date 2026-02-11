# Phase 24: Frontend Controller Extraction - Research

**Researched:** 2026-02-11
**Domain:** React architecture patterns, controller hooks, mutation abstraction
**Confidence:** HIGH

## Summary

Phase 24 extracts orchestration logic from detail pages into controller hooks, following modern React architectural patterns. The codebase already has foundational patterns (e.g., `useEditableSection`, feature-layer query hooks) but pages currently contain significant inline orchestration: query composition, state derivation, action handlers, and mutation boilerplate (toast + invalidation).

Controller hooks centralize this orchestration, making pages presentational and improving testability, reusability, and maintainability. Research confirms this aligns with 2026 React best practices: custom hooks for business logic separation, feature-sliced architecture, and DRY mutation patterns with TanStack Query.

**Primary recommendation:** Extract controller hooks per entity type (stack/instance/service), create shared mutation helper for toast+invalidation boilerplate, maintain feature layer as source of truth for queries/mutations.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| @tanstack/react-query | v5 | Server state management | Industry standard for async state, caching, invalidation in 2026 |
| React Hooks | 19 | Logic composition | Native React API for reusable stateful logic |
| TanStack Router | v1 | File-based routing | Type-safe routing with built-in search param management |
| sonner | latest | Toast notifications | Modern, accessible toast library |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| zod | latest | Validation | Already used for route search params, can extend to form validation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom controller hooks | Redux/Zustand | Overkill for this use case; TanStack Query already manages server state |
| Inline mutations | Redux Toolkit | RTK Query redundant with existing TanStack Query setup |
| Page-level orchestration | HOCs/render props | Hooks are more composable and align with 2026 patterns |

**Installation:**
No new dependencies required. All patterns use existing libraries.

## Architecture Patterns

### Recommended Project Structure
Current structure already follows feature-sliced principles:
```
dashboard/src/
├── features/               # Feature layer (queries, mutations)
│   ├── stacks/queries.ts
│   ├── instances/queries.ts
│   └── services/queries.ts
├── routes/                 # Page components (presentational)
│   ├── stacks/$name.tsx
│   ├── stacks/$name.instances.$instance.tsx
│   └── services/$name.tsx
├── hooks/                  # Shared hooks (editable-section, override-section)
├── components/             # Reusable UI components
└── lib/                    # Utilities (api, format)
```

**Proposed additions:**
```
dashboard/src/
├── features/
│   ├── stacks/
│   │   ├── queries.ts           # Existing
│   │   └── useStackDetailController.ts  # NEW: Stack detail orchestration
│   ├── instances/
│   │   ├── queries.ts           # Existing
│   │   └── useInstanceDetailController.ts  # NEW: Instance detail orchestration
│   └── services/
│       ├── queries.ts           # Existing
│       └── useServiceDetailController.ts  # NEW: Service detail orchestration
└── lib/
    └── mutations.ts         # NEW: Shared mutation helper
```

### Pattern 1: Controller Hook for Detail Pages
**What:** Hook that encapsulates query orchestration, derived state, and action handlers for entity detail pages.

**When to use:** Any detail page with multiple queries, derived state (e.g., computed status), or complex action logic.

**Example:**
```typescript
// features/stacks/useStackDetailController.ts
import { useStack, useInstances, useStackNetwork, useStackCompose } from './queries'
import { useMutationHelper } from '@/lib/mutations'

export function useStackDetailController(stackName: string) {
  // 1. Query orchestration
  const stack = useStack(stackName)
  const instances = useInstances(stackName)
  const network = useStackNetwork(stackName)
  const compose = useStackCompose(stackName)

  // 2. Derived state
  const runningContainerNames = new Set(network.data?.containers ?? [])
  const isLoading = stack.isLoading || instances.isLoading

  // 3. Action handlers with mutation helper
  const startStack = useMutationHelper('startStack', {
    successMessage: (name) => `Started ${name}`,
    invalidate: ['stacks']
  })

  // 4. Return unified interface
  return {
    // Data
    stack: stack.data,
    instances: instances.data ?? [],
    network: network.data,
    compose: compose.data,
    runningContainerNames,

    // Loading states
    isLoading,

    // Actions
    startStack: (name: string) => startStack.mutate(name)
  }
}
```

### Pattern 2: Shared Mutation Helper
**What:** Factory function that wraps TanStack Query's `useMutation` with standardized toast notifications and query invalidation.

**When to use:** All mutations to eliminate repetitive `onSuccess`/`onError` boilerplate.

**Example:**
```typescript
// lib/mutations.ts
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api, getErrorMessage } from './api'

interface MutationConfig<TData, TVariables> {
  mutationFn: (vars: TVariables) => Promise<TData>
  successMessage?: string | ((vars: TVariables, data: TData) => string)
  errorMessage?: string | ((error: unknown) => string)
  invalidate?: string[][]  // Query key patterns to invalidate
  onSuccess?: (data: TData, vars: TVariables) => void
}

export function useMutationHelper<TData, TVariables>(
  config: MutationConfig<TData, TVariables>
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: config.mutationFn,
    onSuccess: (data, vars) => {
      // Toast
      if (config.successMessage) {
        const msg = typeof config.successMessage === 'function'
          ? config.successMessage(vars, data)
          : config.successMessage
        toast.success(msg)
      }

      // Invalidate queries
      config.invalidate?.forEach((keyPattern) => {
        queryClient.invalidateQueries({ queryKey: keyPattern })
      })

      // Custom callback
      config.onSuccess?.(data, vars)
    },
    onError: (error) => {
      const msg = config.errorMessage
        ? typeof config.errorMessage === 'function'
          ? config.errorMessage(error)
          : config.errorMessage
        : getErrorMessage(error, 'Operation failed')
      toast.error(msg)
    }
  })
}

// Usage in queries.ts
export function useEnableStack() {
  return useMutationHelper({
    mutationFn: (name: string) => api.post(`/stacks/${name}/enable`).then(r => r.data),
    successMessage: (name) => `Enabled ${name}`,
    invalidate: [['stacks']]
  })
}
```

### Pattern 3: Instance Tab State Management
**What:** Controller hook manages tab-specific data loading and actions for instance detail pages with many tabs (14 override sections).

**When to use:** Detail pages with multiple tabs requiring different data.

**Example:**
```typescript
// features/instances/useInstanceDetailController.ts
export function useInstanceDetailController(
  stackName: string,
  instanceId: string,
  activeTab: InstanceTab
) {
  const instance = useInstance(stackName, instanceId)
  const templateService = useService(instance.data?.template_name ?? '', {
    enabled: !!instance.data?.template_name
  })

  // Lazy-load effective config only when needed
  const effectiveConfig = useEffectiveConfig(stackName, instanceId, {
    enabled: activeTab === 'effective'
  })

  // Unified update handler
  const updateInstance = useMutationHelper({
    mutationFn: (data) => api.put(`/stacks/${stackName}/instances/${instanceId}`, data),
    successMessage: 'Instance updated',
    invalidate: [
      ['stacks', stackName, 'instances', instanceId],
      ['stacks', stackName, 'instances'],
      ['stacks', stackName]
    ]
  })

  return {
    instance: instance.data,
    templateService: templateService.data,
    effectiveConfig: effectiveConfig.data,
    isLoading: instance.isLoading,
    updateInstance,
    toggleEnabled: () => updateInstance.mutate({
      enabled: !instance.data?.enabled
    })
  }
}
```

### Anti-Patterns to Avoid
- **Controller bloat:** Don't put all domain logic in one controller. Split by page/feature.
- **Tight coupling:** Controllers should depend on feature queries, not directly on API.
- **Premature abstraction:** Don't create shared controllers for dissimilar pages (stack vs instance have different needs).
- **Breaking single responsibility:** Controller coordinates queries/actions; it doesn't implement business rules (that stays in API/backend).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Mutation toast + invalidation | Custom wrapper around each mutation | Shared `useMutationHelper` factory | Eliminates 140+ instances of toast boilerplate; centralized error handling |
| Query orchestration | Inline `useQuery` calls in page components | Controller hooks per entity type | Makes pages presentational; improves testability; centralizes query dependencies |
| Derived state logic | Scattered computations in page components | Controller hook computed properties | Single source of truth for derived values (e.g., running container set) |
| Action handler props | Inline mutation callbacks with imperative logic | Controller-provided action methods | Declarative API; easier to test and refactor |

**Key insight:** Mutation boilerplate is repeated ~140 times across 10 files. A shared helper reduces this to declarative config. Controller hooks follow the same DRY principle for query orchestration.

## Common Pitfalls

### Pitfall 1: Over-Abstracting Controller Hooks
**What goes wrong:** Creating one generic `useDetailController<T>` that tries to handle all entity types leads to complex generics and awkward conditionals.

**Why it happens:** Desire to maximize code reuse without considering entity-specific needs (e.g., stack has network, service has logs).

**How to avoid:** Create entity-specific controller hooks. Extract shared patterns (like mutation helper) but keep controllers tailored to their domain.

**Warning signs:** Controller hooks with many conditional branches (`if (entityType === 'stack')`), excessive generic parameters, unused properties in return value.

### Pitfall 2: Invalidation Over-Fetching
**What goes wrong:** Mutation helper invalidates too many query keys, causing unnecessary refetches.

**Why it happens:** Conservative approach to cache invalidation ("invalidate everything related").

**How to avoid:** Use precise invalidation patterns. TanStack Query supports partial key matching (e.g., `['stacks', stackName]` invalidates all queries for that stack).

**Warning signs:** Network tab shows refetch storms after mutations; multiple identical requests for same resource.

### Pitfall 3: State Duplication Between Controller and Page
**What goes wrong:** Page component maintains local state that duplicates derived state from controller.

**Why it happens:** Incremental refactor leaves behind old state management code.

**How to avoid:** Controller should be single source of truth for all orchestration state. Remove local `useState` calls for data the controller can provide.

**Warning signs:** Two sources of truth for the same value; stale UI after mutation despite query invalidation.

### Pitfall 4: Breaking Existing Patterns
**What goes wrong:** Introducing controller hooks that conflict with existing `useEditableSection` / `useOverrideSection` patterns.

**Why it happens:** Not understanding the existing abstraction layer before adding new one.

**How to avoid:** Controllers should *use* existing hooks (like `useEditableSection`) for editable card state, not replace them. Controllers coordinate queries; editable hooks manage form state.

**Warning signs:** Duplication of edit/cancel/save logic; editable card components break after controller refactor.

## Code Examples

Verified patterns from codebase analysis:

### Current Pattern: Direct Query Usage in Page
```typescript
// routes/stacks/$name.tsx (current)
function StackDetailPage() {
  const { name } = Route.useParams()
  const { data: stack, isLoading } = useStack(name)
  const { data: instances = [] } = useInstances(name)
  const { data: networkStatus } = useStackNetwork(name)
  const { data: composeData } = useStackCompose(name)

  const enableStack = useEnableStack()
  const stopStack = useStopStack()

  const runningContainerNames = new Set(networkStatus?.containers ?? [])

  return (
    <div>
      {/* 730 lines of JSX */}
    </div>
  )
}
```

### Target Pattern: Controller Hook
```typescript
// features/stacks/useStackDetailController.ts (target)
export function useStackDetailController(stackName: string) {
  const stack = useStack(stackName)
  const instances = useInstances(stackName)
  const network = useStackNetwork(stackName)
  const compose = useStackCompose(stackName)

  const enableStack = useEnableStack()
  const stopStack = useStopStack()

  return {
    stack: stack.data,
    instances: instances.data ?? [],
    network: network.data,
    compose: compose.data,
    runningContainerNames: new Set(network.data?.containers ?? []),
    isLoading: stack.isLoading,
    enableStack: () => enableStack.mutate(stackName),
    stopStack: () => stopStack.mutate(stackName)
  }
}

// routes/stacks/$name.tsx (refactored)
function StackDetailPage() {
  const { name } = Route.useParams()
  const controller = useStackDetailController(name)

  if (controller.isLoading) return <Loader />
  if (!controller.stack) return <NotFound />

  return (
    <div>
      {/* JSX now uses controller.* instead of inline state */}
    </div>
  )
}
```

### Current Pattern: Mutation Boilerplate
```typescript
// features/stacks/queries.ts (current)
export function useEnableStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const response = await api.post(`/stacks/${name}/enable`)
      return response.data
    },
    onSuccess: (_data, name) => {
      toast.success(`Enabled ${name}`)
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error, name) => {
      toast.error(getErrorMessage(error, `Failed to enable ${name}`))
    },
  })
}

// Repeated 140+ times across features/
```

### Target Pattern: Mutation Helper
```typescript
// features/stacks/queries.ts (refactored)
export function useEnableStack() {
  return useMutationHelper({
    mutationFn: (name: string) => api.post(`/stacks/${name}/enable`).then(r => r.data),
    successMessage: (name) => `Enabled ${name}`,
    errorMessage: (error, name) => getErrorMessage(error, `Failed to enable ${name}`),
    invalidate: [['stacks']]
  })
}
```

### Existing Pattern: Editable Section Hook
```typescript
// hooks/use-editable-section.ts (keep this pattern)
export function useEditableSection<TDraft>(toDrafts: () => TDraft[]) {
  const [editing, setEditing] = useState(false)
  const [drafts, setDrafts] = useState<TDraft[]>([])

  const startEdit = () => {
    setDrafts(toDrafts())
    setEditing(true)
  }

  const cancel = () => setEditing(false)
  const update = (index: number, patch: Partial<TDraft>) => { /* ... */ }
  const add = (template: TDraft) => { /* ... */ }
  const remove = (index: number) => { /* ... */ }

  return { editing, drafts, startEdit, cancel, update, add, remove }
}

// Controller hooks should USE this, not replace it
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Container/Presentational components | Custom hooks for business logic | React 16.8 (2019), refined 2024-2026 | Hooks more composable than HOCs; better co-location |
| Redux for all state | TanStack Query for server state | 2020-2023 | Eliminates 80% of Redux boilerplate for CRUD apps |
| Class components with lifecycle methods | Functional components with hooks | React 16.8+ | Simpler mental model; better tree-shaking |
| useEffect for data fetching | TanStack Query + Suspense | React 18+, Query v5 | Declarative data fetching; eliminates race conditions |
| Global state for everything | Feature-sliced architecture | 2024-2026 | Better scalability; clearer dependencies |

**Deprecated/outdated:**
- **HOCs for reusable logic:** Replaced by custom hooks (more flexible composition)
- **Redux for server state:** TanStack Query handles this better with less boilerplate
- **Prop drilling for actions:** Controller hooks provide unified action interface
- **Manual query invalidation everywhere:** Mutation helper centralizes this pattern

## Open Questions

1. **Should controller hooks manage dialog state?**
   - What we know: Pages currently use `useState` for dialog open/closed (e.g., `editOpen`, `deleteOpen`)
   - What's unclear: Should controllers expose dialog control methods, or keep this in page component?
   - Recommendation: Keep dialog state in page. It's presentation-level concern, not business logic. Controllers shouldn't know about UI modals.

2. **How to handle optimistic updates?**
   - What we know: Codebase doesn't currently use optimistic updates
   - What's unclear: Should mutation helper support optimistic update patterns?
   - Recommendation: Add opt-in optimistic update config to mutation helper, but don't require it. Most mutations are fast enough without it.

3. **Should we extract list page controllers?**
   - What we know: Phase focuses on detail pages (stack/instance/service detail)
   - What's unclear: List pages (stacks index, services index) also have orchestration logic
   - Recommendation: Phase 24 focuses on detail pages per requirements. List controller extraction is future work (lower complexity, fewer queries per page).

## Sources

### Primary (HIGH confidence)
- Codebase analysis - `dashboard/src/routes/stacks/$name.tsx` (730 lines, orchestration mixed with presentation)
- Codebase analysis - `dashboard/src/features/*/queries.ts` (140+ mutation instances with identical toast/invalidation patterns)
- Codebase analysis - `dashboard/src/hooks/use-editable-section.ts` (existing abstraction for form state management)
- TanStack Query v5 documentation - Query invalidation patterns and `useMutation` API

### Secondary (MEDIUM confidence)
- [TanStack Query discussions on custom hooks](https://github.com/TanStack/query/discussions/3227)
- [Building Reusable Queries with TanStack Query (2026)](https://oluwadaprof.medium.com/building-reusable-queries-with-tanstack-query-a618c5bc82ff)
- [React Hooks Complete Guide 2026](https://inhaq.com/blog/mastering-react-hooks-the-ultimate-guide-for-building-modern-performant-uis.html)
- [Container/Presentational Pattern](https://www.patterns.dev/react/presentational-container-pattern/)
- [React Architecture: Business Logic Separation](https://profy.dev/article/react-architecture-business-logic-and-dependency-injection)
- [Feature-Sliced Design Architecture](https://medium.com/@codewithxohii/feature-sliced-design-architecture-in-react-with-typescript-a-comprehensive-guide-b2652283c6b2)

### Tertiary (LOW confidence)
- Dan Abramov's updated stance on container/presentational (noted hook-based approaches are preferred, but pattern not strictly enforced)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use, no new deps
- Architecture: HIGH - Patterns verified against existing codebase structure
- Mutation helper: HIGH - Direct observation of 140+ repetitions of same pattern
- Controller hooks: HIGH - Natural progression from existing feature-layer queries + editable-section hooks
- Pitfalls: MEDIUM-HIGH - Based on common React Query mistakes and codebase specifics

**Research date:** 2026-02-11
**Valid until:** 60 days (stable patterns, mature libraries, no major React/Query releases expected)
