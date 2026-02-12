# Phase 25: WebSocket Expansion & Frontend Auth - Research

**Researched:** 2026-02-12
**Domain:** React Query invalidation patterns, WebSocket real-time updates
**Confidence:** HIGH

## Summary

Phase 25 extends WebSocket query invalidation to cover stack and instance queries for live container status updates. This is a frontend-only enhancement requiring one-line additions to the existing `useWebSocket` hook. WS auth (delivered in Phase 18) already works — verification is documentation, not new code.

The core pattern: when backend broadcasts `{type: "status"}` every 30s, frontend uses predicate matching to invalidate all `['stacks', ...]` query keys. React Query's deduplication and stale-while-revalidate behavior naturally throttle refetches. No backend changes required.

**Primary recommendation:** Add broad predicate invalidation matching `key[0] === 'stacks'` using the exact pattern already proven for service metrics. Keep existing 30s broadcast interval.

## User Constraints (from CONTEXT.md)

<user_constraints>

### Locked Decisions

**Invalidation strategy:**
- Broad invalidation via predicate matching all `['stacks', ...]` query keys on any `status` message
- No targeted per-stack/per-instance invalidation — 30s broadcast interval naturally throttles refetches
- React Query deduplicates and returns cached data for unchanged queries

**Message structure:**
- No changes to WS `StatusUpdate` format — keep `{type: "status", data: {containers: {...}}}`
- Frontend doesn't need to know which stacks changed — "something changed, refetch" is sufficient
- No new message types (`"stack-update"`, `"instance-update"` etc.) — single `"status"` type retained

**Existing invalidation:**
- Keep current `['services']`, `['status']`, `['categories']`, and service metrics predicate invalidation
- Stack/instance invalidation is purely additive alongside existing patterns

**WS auth (SC-3):**
- Already delivered by Phase 18 — `fetchWSToken()` + `?token=` query param + `ValidateWSToken()` server-side
- Reconnection in `useWebSocket` re-fetches tokens (handles 60s TTL expiry)
- Phase 25 verifies end-to-end but requires no new auth code

**Broadcast mechanism:**
- Keep existing 30s polling timer — no event-driven broadcasts on action completion
- "Live" in SC-4 refers to the existing WS mechanism, not sub-second responsiveness

**Scope:**
- Frontend-only phase — all changes in `dashboard/src/hooks/use-websocket.ts`
- No backend changes to sync manager, WS handler, or message format

### Claude's Discretion

- Exact predicate implementation (single predicate vs multiple `queryKey` entries)
- Whether to consolidate existing service invalidation into predicates for consistency

### Deferred Ideas (OUT OF SCOPE)

- Event-driven WS broadcasts on action completion (start/stop/restart/apply) for sub-second UI refresh
- Granular WS message types for targeted per-stack invalidation
- WS connection health indicator in dashboard UI

</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| @tanstack/react-query | 5.90.20 | Query invalidation & caching | Already in project — industry standard for React server state |

### Supporting
| Tool | Purpose | When to Use |
|------|---------|-------------|
| Native WebSocket API | Browser WS connection | Already used in `useWebSocket` hook |

**Installation:** No new packages required — all tools already in project.

## Architecture Patterns

### Current Codebase Structure
```
dashboard/src/
├── hooks/
│   └── use-websocket.ts          # WebSocket connection manager (WE CHANGE THIS)
├── features/
│   ├── stacks/queries.ts         # Stack query keys: ['stacks'], ['stacks', name], ['stacks', name, 'network'], etc.
│   └── instances/queries.ts      # Instance query keys: ['stacks', stackName, 'instances'], ['stacks', stackName, 'instances', id], etc.
└── lib/
    └── api.ts                    # fetchWSToken() — already implemented
```

### Pattern 1: Broad Predicate Invalidation (Recommended)

**What:** Single predicate matching all query keys starting with `'stacks'`

**When to use:** When broadcast messages don't contain granular change information and 30s interval naturally throttles

**Example:**
```typescript
// Already exists in use-websocket.ts for service metrics:
queryClient.invalidateQueries({ predicate: (q) => {
  const key = q.queryKey
  return Array.isArray(key) && key.length >= 3 && key[0] === 'services' && key[2] === 'metrics'
}})

// Add for stacks (all stack-related queries):
queryClient.invalidateQueries({ predicate: (q) => {
  const key = q.queryKey
  return Array.isArray(key) && key[0] === 'stacks'
}})
```

**Why this works:**
- React Query marks queries as stale but doesn't refetch if no components are currently subscribed
- Active queries (pages user is viewing) refetch automatically via `refetchInterval: 30000` coordination
- Cached data returned immediately while refetch happens in background (stale-while-revalidate)
- React Query deduplicates simultaneous requests for the same query key

Source: [TanStack Query Query Invalidation Docs](https://tanstack.com/query/latest/docs/framework/react/guides/query-invalidation)

### Pattern 2: Existing Stack Query Key Structure

Current query keys follow hierarchical pattern:
```typescript
['stacks']                                    // useStacks() - list all stacks
['stacks', name]                              // useStack(name) - single stack detail
['stacks', name, 'network']                   // useStackNetwork(name) - network status
['stacks', name, 'compose']                   // useStackCompose(name) - compose YAML
['stacks', name, 'wires']                     // useStackWires(name) - wiring state
['stacks', name, 'instances']                 // useInstances(stackName) - instance list
['stacks', name, 'instances', id]             // useInstance(stackName, id) - instance detail
['stacks', name, 'instances', id, 'resources'] // useResourceLimits(stackName, id) - resource config
```

All refetch every 30s when active:
```typescript
refetchInterval: 30000
```

Source: Verified in `dashboard/src/features/stacks/queries.ts` and `dashboard/src/features/instances/queries.ts`

### Pattern 3: WebSocket Message Flow

Backend broadcasts status every 30s (confirmed):
```go
// api/internal/sync/manager.go:140
func (m *Manager) containerStatusLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	// ...broadcasts StatusUpdate{Type: "status", Data: containers}
}
```

Frontend receives and processes:
```typescript
// dashboard/src/hooks/use-websocket.ts:49-60
ws.onmessage = (event) => {
  const message: WebSocketMessage = JSON.parse(event.data)
  if (message.type === 'status') {
    queryClient.invalidateQueries({ queryKey: ['services'] })
    queryClient.invalidateQueries({ queryKey: ['status'] })
    queryClient.invalidateQueries({ queryKey: ['categories'] })
    queryClient.invalidateQueries({ predicate: (q) => {
      const key = q.queryKey
      return Array.isArray(key) && key.length >= 3 && key[0] === 'services' && key[2] === 'metrics'
    }})
  }
}
```

### Pattern 4: Auth Token Flow (Already Implemented)

Phase 18 delivered complete WS auth:
```typescript
// dashboard/src/hooks/use-websocket.ts:30-38
async function connect() {
  if (!mounted) return

  const token = await fetchWSToken()  // Fetches HMAC-signed token (60s TTL)

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  let wsUrl = `${protocol}//${window.location.host}/api/v1/ws/status`
  if (token) {
    wsUrl += `?token=${encodeURIComponent(token)}`  // Appends token as query param
  }
  // ...WebSocket upgrade with token in URL
}
```

Backend validation (strict mode only):
```go
// api/internal/api/handlers/websocket.go:69-76
if h.secMode.RequiresWSAuth() {
	token := r.URL.Query().Get("token")
	apiKey := os.Getenv("DEVARCH_API_KEY")
	if err := security.ValidateWSToken(token, []byte(apiKey)); err != nil {
		respond.Unauthorized(w, r, "unauthorized: invalid or missing ws token")
		return
	}
}
```

Reconnection automatically re-fetches token — handles expiry naturally via exponential backoff retry loop.

Source: Verified in Phase 18 plans (18-01-PLAN.md, 18-02-PLAN.md) and current codebase.

### Anti-Patterns to Avoid

- **Overly-specific predicates:** Matching exact query keys defeats React Query's deduplication
- **Targeted invalidation without data:** Frontend can't know which specific stack changed from a broadcast container list
- **Multiple invalidation calls:** Single predicate more efficient than calling `invalidateQueries` for each query key pattern

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Targeted WS invalidation | Parse container list to extract stack names | Broad predicate matching | 30s interval + React Query deduplication already throttles effectively. Parsing adds complexity with no latency gain. |
| Connection state tracking | Custom reconnection logic | Existing exponential backoff | Already implemented with BASE_DELAY=3s, MAX_DELAY=30s. Fetches fresh token on each reconnect. |
| Granular message types | Backend changes for `stack-update` | Keep single `status` type | No sub-second requirement. Backend broadcasts already throttled at 30s. |

**Key insight:** React Query's intelligent caching makes "invalidate everything potentially affected" faster and simpler than "parse, filter, target." The 30s broadcast interval is the bottleneck, not invalidation granularity.

## Common Pitfalls

### Pitfall 1: Over-Invalidation Fear

**What goes wrong:** Developers hesitate to invalidate broad query key prefixes, fearing performance issues

**Why it happens:** Misunderstanding React Query's refetch behavior — only active queries refetch, and deduplication prevents duplicate requests

**How to avoid:**
- Invalidation only marks queries as stale (cheap operation)
- Refetches only occur if query is actively rendered (`useQuery` hook mounted)
- Multiple invalidations of same query key within short window deduplicated
- Stale data returned immediately while refetch happens in background

**Warning signs:** Adding complex filtering logic to avoid "unnecessary" invalidations

Source: [TanStack Query Invalidation Behavior](https://tanstack.com/query/latest/docs/framework/react/guides/query-invalidation)

### Pitfall 2: Mixing queryKey and predicate

**What goes wrong:** Using both `queryKey: ['stacks']` and `predicate` in same invalidation call

**Why it happens:** Uncertainty about which approach to use

**How to avoid:**
- `queryKey: ['stacks']` — exact or prefix match (matches `['stacks']`, `['stacks', 'foo']`, etc.)
- `predicate` — custom logic for complex matching (e.g., checking second element)
- Choose one per invalidation call — mixing causes AND logic (both must match)

**Warning signs:** Predicates that just check `key[0]` when `queryKey` would suffice

### Pitfall 3: Assuming Instant Refetch

**What goes wrong:** Expecting sub-second updates after WS message

**Why it happens:** Confusion between invalidation and refetch timing

**How to avoid:**
- Invalidation is immediate (marks query as stale)
- Refetch respects `refetchInterval` — if query just fetched, waits up to 30s
- For instant updates, use `setQueryData` to update cache directly (not needed here — 30s is acceptable per requirements)

**Warning signs:** Adding `refetchType: 'all'` thinking it makes refetches faster

## Code Examples

### Example 1: Adding Stack Invalidation to useWebSocket

Current implementation (service invalidation):
```typescript
// dashboard/src/hooks/use-websocket.ts:49-60
ws.onmessage = (event) => {
  try {
    const message: WebSocketMessage = JSON.parse(event.data)
    if (message.type === 'status') {
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      queryClient.invalidateQueries({ predicate: (q) => {
        const key = q.queryKey
        return Array.isArray(key) && key.length >= 3 && key[0] === 'services' && key[2] === 'metrics'
      }})
    }
  } catch {
    // ignore malformed messages
  }
}
```

**Add single line for stack invalidation:**
```typescript
ws.onmessage = (event) => {
  try {
    const message: WebSocketMessage = JSON.parse(event.data)
    if (message.type === 'status') {
      queryClient.invalidateQueries({ queryKey: ['services'] })
      queryClient.invalidateQueries({ queryKey: ['status'] })
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      queryClient.invalidateQueries({ predicate: (q) => {
        const key = q.queryKey
        return Array.isArray(key) && key.length >= 3 && key[0] === 'services' && key[2] === 'metrics'
      }})
      // NEW: Invalidate all stack and instance queries
      queryClient.invalidateQueries({ predicate: (q) => {
        const key = q.queryKey
        return Array.isArray(key) && key[0] === 'stacks'
      }})
    }
  } catch {
    // ignore malformed messages
  }
}
```

Matches all these keys:
- `['stacks']` — stack list
- `['stacks', 'prod']` — single stack
- `['stacks', 'prod', 'network']` — stack network
- `['stacks', 'prod', 'instances']` — instance list
- `['stacks', 'prod', 'instances', 'postgres-1']` — instance detail
- `['stacks', 'prod', 'instances', 'postgres-1', 'resources']` — instance resources

### Example 2: Alternative — Simple queryKey Prefix Match

If predicate feels heavy-handed, use prefix matching (simpler but matches fewer edge cases):
```typescript
queryClient.invalidateQueries({ queryKey: ['stacks'] })
```

**Tradeoff:** Simpler code, same practical effect. Predicate is more explicit about "all stacks-related queries."

### Example 3: Verification Test (Manual)

Phase 18 already delivered auth — verification is documentation:

**Test in strict mode:**
1. Set `SECURITY_MODE=strict` in `.env`
2. Set `DEVARCH_API_KEY=test-key`
3. Start API + dashboard
4. Open browser DevTools Network tab, filter WS
5. Verify WS upgrade request includes `?token=...` query param
6. Verify WS connects successfully (101 Switching Protocols)
7. Verify status messages received every ~30s

**Expected behavior:**
- Token fetched before WS connect (`POST /api/v1/auth/ws-token` with `X-API-Key` header)
- WS URL includes token: `ws://localhost:5174/api/v1/ws/status?token=<hex>.<hex>`
- Backend validates token, allows connection
- On reconnect (close tab, reopen), fresh token fetched automatically

**Test in dev-open mode:**
1. Set `SECURITY_MODE=dev-open` (or unset)
2. Start API + dashboard
3. Verify WS connects without token in URL
4. Verify status messages received every ~30s

No code changes required — Phase 18 delivered this.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual subscription per query | React Query auto-refetch with invalidation | TanStack Query v3+ (2021) | Developers trigger refetch via invalidation, React Query handles subscription management |
| WebSocket cache updates | Hybrid invalidation + polling | React Query v4+ (2022) | Stale-while-revalidate pattern allows simple invalidation without complex cache surgery |
| JWT-based WS auth | HMAC-signed query tokens | Phase 18 (2026-02-11) | Simpler implementation, no JWT library dependency, 60s TTL sufficient for upgrade handshake |

**Deprecated/outdated:**
- React Query v3 `refetchQueries()` — v4+ use `invalidateQueries()` (cleaner API)
- Synchronous cache updates via `setQueryData()` for WS — invalidation + refetch interval simpler for 30s broadcast cadence

## Open Questions

**None** — scope locked by user decisions. Implementation is straightforward one-line addition.

## Sources

### Primary (HIGH confidence)
- [TanStack Query Invalidation Docs](https://tanstack.com/query/latest/docs/framework/react/guides/query-invalidation) — predicate patterns, invalidation behavior
- [TanStack Query QueryClient Reference](https://tanstack.com/query/latest/docs/reference/QueryClient) — invalidateQueries API
- Codebase verification: `dashboard/src/hooks/use-websocket.ts`, `dashboard/src/features/stacks/queries.ts`, `api/internal/sync/manager.go`, Phase 18 plans

### Secondary (MEDIUM confidence)
- [TkDodo's Blog: Using WebSockets with React Query](https://tkdodo.eu/blog/using-web-sockets-with-react-query) — invalidation vs setQueryData patterns
- [LogRocket: TanStack Query and WebSockets](https://blog.logrocket.com/tanstack-query-websockets-real-time-react-data-fetching/) — real-time integration patterns

### Tertiary (LOW confidence)
- None — scope is narrow and well-documented in official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — TanStack Query v5 already in project, no new deps
- Architecture: HIGH — existing patterns verified in codebase, Phase 18 auth confirmed implemented
- Pitfalls: HIGH — official docs + established React Query community knowledge

**Research date:** 2026-02-12
**Valid until:** 30 days (stable API, unlikely to change)
