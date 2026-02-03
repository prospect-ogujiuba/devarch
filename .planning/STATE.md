# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 2 - Stack CRUD

## Current Position

Phase: 2 of 9 (Stack CRUD)
Plan: 5 of 5 in current phase
Status: Phase complete
Last activity: 2026-02-03 — Completed 02-05-PLAN.md

Progress: [███░░░░░░░] ~26% (7 plans complete of ~27 total)

## Performance Metrics

**Velocity:**
- Total plans completed: 7
- Average duration: 2.5 min
- Total execution time: 0.29 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 2/2 | 1min | 0.5min |
| 2 | 5/5 | 17.5min | 3.5min |

**Recent Trend:**
- Last 3 plans: 02-03 (1.2min), 02-04 (2.4min), 02-05 (11min)
- Trend: UI tasks take longer than API (expected)

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

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-03T22:13:50Z
Stopped at: Completed 02-05-PLAN.md (Stack detail & dialogs) — Phase 2 complete
Resume file: None

**Phase 2 complete.** Ready to begin Phase 3 (Service Instances).
