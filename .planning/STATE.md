# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 3 - Service Instances

## Current Position

Phase: 3 of 9 (Service Instances)
Plan: 1 of 5 in current phase
Status: In progress
Last activity: 2026-02-03 — Completed 03-01-PLAN.md

Progress: [███░░░░░░░] ~30% (8 plans complete of ~27 total)

## Performance Metrics

**Velocity:**
- Total plans completed: 8
- Average duration: 2.3 min
- Total execution time: 0.37 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 2/2 | 1min | 0.5min |
| 2 | 5/5 | 17.5min | 3.5min |
| 3 | 1/5 | 5min | 5min |

**Recent Trend:**
- Last 3 plans: 02-04 (2.4min), 02-05 (11min), 03-01 (5min)
- Trend: API schema + handlers avg 3-5min

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Full copy-on-write overrides (including healthchecks, config files) — users expect full control
- Isolation as core value over auto-wiring or declarative config — wiring is useless if stacks collide
- Auto-wire + explicit contracts (both layers) — simple stacks stay simple, complex stacks are possible
- Encryption at rest from v1 — avoids painful retrofit, builds trust for adoption
- Per-phase dashboard UI (not a dedicated UI phase) — early feedback loop, testable increments
- devarch.yml for sharing + backup (both equally) — portable definitions that also serve as state backup
- Export/Import moved before Wiring (Phase 7) — validate "shareable env" loop early, wiring can follow
- Lockfile concept (devarch.lock) — manifest (devarch.yml) + lockfile (resolved ports/digests/versions) = deterministic reproduction
- devarch init + devarch doctor — one-command bootstrap + diagnostics for teammate handoff
- Export includes resolved specifics (host ports, image digests, template versions) — prevents import divergence

**From 01-01:**
- Label prefix "devarch." for namespace isolation
- DNS-safe naming (63 char limit) prevents network collision
- Prescriptive validation errors include Slugify suggestions

**From 01-02:**
- DEVARCH_RUNTIME env var for explicit runtime control
- Network stubs return ErrNotImplemented (Phase 4 ready)
- Backward compat preserved during refactor (GetStatus/GetMetrics)

**From 02-01:**
- Soft-delete pattern using deleted_at with partial unique index on active stacks only
- Stack name is immutable identifier (used in URL routes)
- Running count placeholder set to 0 until Phase 3+ wires container client queries

**From 02-02:**
- Rename implemented as atomic clone + soft-delete transaction
- Restore checks for active name conflicts with prescriptive error
- Trash routes registered before /{name} routes to avoid chi parameter conflicts

**From 02-03:**
- Stack hooks follow existing service hooks pattern for consistency
- 5-second polling for real-time updates (WebSocket extension deferred)
- Stacks positioned second in navigation (after Overview, before Services)

**From 02-04:**
- Grid as default view for stacks (better visual hierarchy)
- Color-coded status indicators (green/yellow/gray) for running status
- Create/clone/rename dialogs deferred to 02-05

**From 02-05:**
- Delete cascade preview pattern shows blast radius before destructive ops
- Rename UX hides clone+soft-delete implementation (feels first-class)
- Clone creates records only, doesn't start containers (aligns with plan/apply workflow)
- Disable dialog enumerates containers by name for transparency
- All actions accessible from both list and detail views

**From 03-01:**
- Override tables mirror service tables exactly (consistency, type safety, efficient queries)
- Partial unique index WHERE deleted_at IS NULL enables soft-delete while preventing duplicates
- Container name follows devarch-{stack}-{instance} pattern
- Override count computed via subquery sum in single query (avoids N+1)
- Duplicate copies all override records in transaction (atomic, all-or-nothing)
- Rename is direct UPDATE on instance_id + container_name (instances are DB records at this phase)

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-03T21:51:00Z
Stopped at: Completed 03-01-PLAN.md (Instance override schema & CRUD) — Phase 3 started
Resume file: None
