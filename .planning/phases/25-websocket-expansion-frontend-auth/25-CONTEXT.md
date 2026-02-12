# Phase 25: WebSocket Expansion & Frontend Auth - Context

**Gathered:** 2026-02-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Extend WebSocket query invalidation to cover stack and instance queries so live container status updates refresh stack/instance detail pages. Verify WS auth (delivered in Phase 18) works end-to-end. No backend WS changes — this is a frontend invalidation expansion.

</domain>

<decisions>
## Implementation Decisions

### Invalidation strategy
- Broad invalidation via predicate matching all `['stacks', ...]` query keys on any `status` message
- No targeted per-stack/per-instance invalidation — 30s broadcast interval naturally throttles refetches
- React Query deduplicates and returns cached data for unchanged queries

### Message structure
- No changes to WS `StatusUpdate` format — keep `{type: "status", data: {containers: {...}}}`
- Frontend doesn't need to know which stacks changed — "something changed, refetch" is sufficient
- No new message types (`"stack-update"`, `"instance-update"` etc.) — single `"status"` type retained

### Existing invalidation
- Keep current `['services']`, `['status']`, `['categories']`, and service metrics predicate invalidation
- Stack/instance invalidation is purely additive alongside existing patterns

### WS auth (SC-3)
- Already delivered by Phase 18 — `fetchWSToken()` + `?token=` query param + `ValidateWSToken()` server-side
- Reconnection in `useWebSocket` re-fetches tokens (handles 60s TTL expiry)
- Phase 25 verifies end-to-end but requires no new auth code

### Broadcast mechanism
- Keep existing 30s polling timer — no event-driven broadcasts on action completion
- "Live" in SC-4 refers to the existing WS mechanism, not sub-second responsiveness

### Scope
- Frontend-only phase — all changes in `dashboard/src/hooks/use-websocket.ts`
- No backend changes to sync manager, WS handler, or message format

### Claude's Discretion
- Exact predicate implementation (single predicate vs multiple `queryKey` entries)
- Whether to consolidate existing service invalidation into predicates for consistency

</decisions>

<specifics>
## Specific Ideas

- Use same predicate pattern already used for service metrics: `Array.isArray(key) && key[0] === 'stacks'`
- SC-3 verification may just be a manual test or documented check rather than new code

</specifics>

<deferred>
## Deferred Ideas

- Event-driven WS broadcasts on action completion (start/stop/restart/apply) for sub-second UI refresh
- Granular WS message types for targeted per-stack invalidation
- WS connection health indicator in dashboard UI

</deferred>

---

*Phase: 25-websocket-expansion-frontend-auth*
*Context gathered: 2026-02-12*
