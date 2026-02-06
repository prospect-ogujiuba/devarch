# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-03)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 4 - Network Isolation

## Current Position

Phase: 4 of 9 (Network Isolation)
Plan: 04-01 complete (1 of ~3 in phase)
Status: In progress
Last activity: 2026-02-06 — Completed 04-01-PLAN.md (network operations backend)

Progress: [█████░░░░░] ~48% (13 plans complete of ~27 total)

## Performance Metrics

**Velocity:**
- Total plans completed: 13
- Average duration: 3.0 min
- Total execution time: ~0.7 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 2/2 | 1min | 0.5min |
| 2 | 5/5 | 17.5min | 3.5min |
| 3 | 5/5 | ~21min | ~4.2min |
| 4 | 1/3 | 4.1min | 4.1min |

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

**From 03-02:**
- Override PUT endpoints follow service handler pattern (DELETE + INSERT transaction)
- Config files use UPSERT pattern with ON CONFLICT for idempotent updates
- System labels (devarch.*) validated at API layer and auto-injected in effective config
- Dependencies read-only in effective config per INST-05 requirement
- Merge semantics: full replacement for ports/volumes/domains/healthcheck, key-based for env/labels, path-based for config files

**From 03-03:**
- Auto-name generation with collision detection for instances (template-2, template-3)
- Cache invalidation chains ensure override mutations invalidate instance detail, effective config, and list
- Template catalog shows instance counts per template (helps user understand stack composition)
- Empty state CTA pattern consistent with stack grid (centered card with icon and action)

**From 03-04:**
- Override editor UX pattern: template values muted, overrides with blue left border
- Explicit save (not auto-save) with dirty tracking prevents accidental changes
- Per-field reset (X icon) + Reset All button gives granular control
- Config files use CodeMirror with language detection (JSON/YAML/XML)
- Template config files shown read-only as reference above editable override

**From 04-01:**
- Idempotent CreateNetwork via inspect-then-create pattern (Docker/Podman return errors on duplicate)
- Graceful RemoveNetwork ignores not-found errors (safe delete operations)
- ListNetworks filters by devarch.managed_by label (orphan detection boundary)
- Container name validation at instance creation (127-char limit, fail early with prescriptive error)
- Network name auto-computed but overridable (devarch-{stack}-net default)
- Clone/Rename recompute network_name for new stack (prevents shared networks)

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-06
Stopped at: Completed 04-01-PLAN.md (network operations backend, 2 tasks, 2 commits)
Resume file: None
