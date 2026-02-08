# Phase 6: Plan/Apply Workflow - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Terraform-style safety mechanism for stack deployment. Users preview what will change (add/modify/remove) before applying, with advisory locking preventing concurrent modifications. Covers: plan generation, structured diffs, apply execution, staleness detection, dashboard plan/apply UX.

</domain>

<decisions>
## Implementation Decisions

### Plan Diff Structure
- Service-level diffs: each entry is one service instance with action (`+` add, `~` modify, `-` remove)
- Modifications include per-field detail (image, ports, env vars, volumes, etc.)
- Include `source` field on changes ("user", "template") — Phase 8 adds "wire" as a source
- Diff format should be reusable by Phase 7 import reconciliation preview

*Forward: Phase 7 (EXIM-02 import diffs reuse format), Phase 8 (WIRE-06 wiring diagnostics in plan)*

### Apply Flow & Error Handling
- Sequential flow per PLAN-03: lock → ensure network → materialize configs → compose up
- Network creation failure → unlock, return error (nothing to roll back)
- Config materialization failure → clean up partial config files, unlock, return error
- Compose up failure → leave configs in place for debugging, return compose stderr, unlock
- No automatic container rollback — compose handles its own partial state
- User inspects and fixes, like Terraform partial apply

*Forward: Phase 7 (BOOT-01 devarch init reuses apply flow — simpler error model = simpler init)*

### Plan Staleness & State Tracking
- Plans are ephemeral JSON returned from plan endpoint, not persisted in DB
- Client sends plan (including staleness token) to apply endpoint
- Staleness token: stack `updated_at` + hash of all instance `updated_at` values
- If any timestamp changed between plan and apply → reject as stale (HTTP 409)
- No plan table — plans are cheap to regenerate, no cleanup needed

*Forward: Phase 7 (LOCK-01/02 lockfile uses same hash concept), Phase 9 (SECR-03 no unredacted secrets persisted in plan table)*

### Dashboard Plan/Apply UX
- New tab on stack detail page (alongside Instances and Compose tabs)
- Tab name: "Deploy" (action-oriented)
- Flow: "Generate Plan" button → structured diff appears → "Apply" button → progress → result
- Diff color-coded: green for adds, yellow for modifications, red for removals
- Per-field detail visible for modifications (expandable)
- Apply progress indicator with real-time feedback

*Forward: Phase 8 (WIRE-06 wiring diagnostics need display space — tab provides it)*

### Advisory Lock Semantics
- Per-stack Postgres advisory locks using `pg_try_advisory_lock(stack.id)`
- Non-blocking: if lock held, immediately return HTTP 409 Conflict
- Message: "Stack is being applied by another session"
- Pattern follows existing `sync/manager.go` implementation
- No queueing, no waiting — fail fast with clear message

*Forward: Phase 7 (BOOT-01 init uses apply — must fail fast, not hang on lock)*

### Container State Inspection
- Query runtime for containers with `devarch.stack_id={stack}` labels via existing `ListContainersWithLabels`
- Compare running containers against desired instances for adds/removes
- For modifications: container inspect to get current image/env/ports, diff against effective config
- Runtime is source of truth for "what's running" — no separate state table
- Labels from Phase 4 (NETW-04) make this possible without new infrastructure

*Forward: Phase 8 (WIRE-04/05 wiring validation needs same runtime state inspection)*

### Claude's Discretion
- Exact diff JSON schema field naming
- How to render apply progress (streaming vs polling)
- Migration structure (if any new fields needed on stacks table)
- HTTP status codes for edge cases beyond 409
- Plan endpoint HTTP method (POST vs GET — both have valid arguments)

</decisions>

<specifics>
## Specific Ideas

- Diff format inspired by Terraform: `+` add, `~` modify, `-` remove with per-field detail
- Advisory lock pattern already proven in `internal/sync/manager.go` — reuse the approach
- Compose tab already exists on stack detail — Deploy tab follows same CodeMirror/tab pattern
- `GenerateStack()` already returns warnings array — plan can extend with diff data
- `MaterializeStackConfigs()` already handles atomic config file swap — apply reuses this

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-plan-apply-workflow*
*Context gathered: 2026-02-07*
