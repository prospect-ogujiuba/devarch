# Phase 3 — Plan + Apply + Runtime Adapters

## Goal
Build the operational core that compares desired vs runtime state and performs apply, status, logs, and exec flows.

## Acceptance

- runtime adapters exist for Docker and Podman
- plan output explains add/modify/remove/restart decisions
- apply can materialize a simple workspace end to end
- logs, status, and exec work through the runtime abstraction
- event emission exists for long-running operations

## Recommended sequence

Define runtime interfaces first. Implement status inspection and planner before apply. Add logs/exec/events after the basic adapter contract is stable.

---

### P3-RT-001 — Define runtime adapter contract
- **Owner:** architect
- **Depends on:** Phase 2 complete
- **Goal:** settle the shared adapter interface for Docker and Podman
- **Tasks:**
  1. Define operations for inspect, apply primitives, logs, exec, and network helpers.
  2. Define the runtime snapshot shape consumed by the planner.
  3. Record unsupported-operation behavior clearly.
- **Validation:**
  - design review
- **Done when:**
  - both adapters can target one common interface

### P3-RT-002 — Implement runtime snapshot inspection
- **Owner:** surgeon-runtime
- **Depends on:** P3-RT-001
- **Goal:** collect runtime state for a workspace and its resources
- **Tasks:**
  1. Inspect containers, health, ports, and IDs.
  2. Normalize results into a runtime snapshot model.
  3. Add adapter contract tests against fixtures or mocks.
- **Validation:**
  - `go test ./internal/runtime/...`
- **Done when:**
  - planner input can be built from runtime inspection

### P3-PLAN-001 — Build planner diff engine
- **Owner:** surgeon-engine
- **Depends on:** P3-RT-002
- **Goal:** compare effective graph vs runtime snapshot
- **Tasks:**
  1. Compute add, modify, remove, restart, and noop decisions.
  2. Attach human-readable reasons to actions.
  3. Add planner diff tests and golden outputs.
- **Validation:**
  - `go test ./internal/plan/...`
- **Done when:**
  - plans are reproducible and explainable

### P3-APPLY-001 — Render runtime payloads
- **Owner:** surgeon-runtime
- **Depends on:** P3-PLAN-001
- **Goal:** convert effective graph resources into runtime-ready payloads
- **Tasks:**
  1. Materialize container/runtime settings from the graph.
  2. Handle networks, env, mounts, and ports.
  3. Keep rendered payloads testable without full runtime execution.
- **Validation:**
  - `go test ./internal/apply/... ./internal/runtime/...`
- **Done when:**
  - apply has a stable render step before side effects

### P3-APPLY-002 — Implement apply executor
- **Owner:** surgeon-runtime
- **Depends on:** P3-APPLY-001
- **Goal:** execute apply actions against the selected runtime
- **Tasks:**
  1. Create and update resources according to plan output.
  2. Handle stop/remove flows where required.
  3. Emit structured operation progress events.
- **Validation:**
  - `go test ./internal/apply/...`
  - sample workspace end-to-end apply smoke check
- **Done when:**
  - a simple workspace can be planned and applied successfully

### P3-RT-003 — Add logs and exec primitives
- **Owner:** surgeon-runtime
- **Depends on:** P3-RT-002
- **Goal:** provide runtime-backed logs and terminal access
- **Tasks:**
  1. Stream or fetch logs per workspace/resource.
  2. Provide exec session hooks.
  3. Normalize adapter differences where possible.
- **Validation:**
  - `go test ./internal/runtime/...`
- **Done when:**
  - later CLI/API/UI surfaces can expose logs and exec consistently

### P3-EVT-001 — Implement event stream model
- **Owner:** surgeon-api
- **Depends on:** P3-APPLY-002, P3-RT-003
- **Goal:** define and emit event messages for apply and runtime observation
- **Tasks:**
  1. Define event envelope and event types.
  2. Publish apply progress and relevant runtime changes.
  3. Add tests for event serialization and sequencing.
- **Validation:**
  - `go test ./internal/events/... ./internal/api/...`
- **Done when:**
  - consumers can subscribe to structured workspace events

### P3-CACHE-001 — Add optional runtime cache/history boundary
- **Owner:** surgeon-runtime
- **Depends on:** P3-PLAN-001
- **Goal:** create the non-canonical cache boundary for snapshots and apply history
- **Tasks:**
  1. Define what runtime data is cacheable.
  2. Add minimal SQLite-backed persistence where useful.
  3. Keep cache usage optional and replaceable.
- **Validation:**
  - `go test ./internal/cache/...`
- **Done when:**
  - runtime history can be persisted without becoming desired state

## Parallel-safe packets

After `P3-RT-001`, these can run in parallel if they touch separate files:

- `P3-RT-002`
- `P3-RT-003`

After `P3-PLAN-001`, these can overlap with care:

- `P3-APPLY-001`
- `P3-CACHE-001`

## Handoff to Phase 4

Phase 4 starts only after:

- simple workspace plan/apply works end to end
- runtime status/logs/exec primitives exist
- planner outputs are stable and test-covered
