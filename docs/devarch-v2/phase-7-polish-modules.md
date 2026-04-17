# Phase 7 — Polish + Decomposition of Advanced Modules

## Goal
Stabilize V2 for primary product direction while carving postponed concerns into explicit module boundaries instead of letting them leak into the core.

## Acceptance

- error handling and UX are polished across surfaces
- performance issues are identified and addressed
- release/docs/site materials are updated
- deferred concerns have explicit boundaries or follow-up plans
- V2 is shippable as the primary direction

## Recommended sequence

Polish core UX first, then performance and operational hardening, then module boundaries and release closeout.

---

### P7-UX-001 — Improve error and empty-state UX
- **Owner:** surgeon-ui
- **Depends on:** Phase 6 complete
- **Goal:** make failures understandable across CLI, API, and UI
- **Tasks:**
  1. Review common validation, runtime, and import failure messages.
  2. Improve user-facing errors and empty states.
  3. Keep technical detail available without overwhelming the default view.
- **Validation:**
  - UI/CLI/API smoke review
- **Done when:**
  - common failure cases are actionable and consistent

### P7-PERF-001 — Profile critical paths
- **Owner:** verifier
- **Depends on:** Phase 6 complete
- **Goal:** identify the slowest parts of load, resolve, plan, apply, and UI rendering flows
- **Tasks:**
  1. Define benchmark or profiling scenarios.
  2. Measure engine and UI hotspots.
  3. Produce a ranked optimization list.
- **Validation:**
  - profiling report
- **Done when:**
  - optimization work is evidence-based

### P7-PERF-002 — Optimize resolver and plan hot paths
- **Owner:** surgeon-engine
- **Depends on:** P7-PERF-001
- **Goal:** reduce unnecessary work in the highest-value engine paths
- **Tasks:**
  1. Address the top resolver and planner bottlenecks.
  2. Keep determinism and readability intact.
  3. Add benchmarks where helpful.
- **Validation:**
  - targeted benchmarks/tests
- **Done when:**
  - critical engine paths measurably improve

### P7-OPS-001 — Harden runtime and daemon behavior
- **Owner:** surgeon-runtime
- **Depends on:** P7-PERF-001
- **Goal:** improve reliability under routine local-operator conditions
- **Tasks:**
  1. Improve retry, reconnect, and cleanup behavior.
  2. Tighten daemon lifecycle and shutdown handling.
  3. Add regression tests for common failure modes.
- **Validation:**
  - runtime/API integration tests
- **Done when:**
  - local operations are more robust and predictable

### P7-MOD-001 — Define boundaries for deferred advanced modules
- **Owner:** architect
- **Depends on:** Phase 6 complete
- **Goal:** prevent postponed concerns from leaking back into the V2 core
- **Tasks:**
  1. Define boundary notes for AI, registry utilities, image utilities, scanning, and advanced team features.
  2. Clarify what stays out of the core milestone.
  3. Record integration points without implementing the modules fully.
- **Validation:**
  - ADR or RFC review
- **Done when:**
  - deferred work has explicit seams instead of vague TODOs

### P7-DOC-001 — Update product and migration docs/site
- **Owner:** scribe
- **Depends on:** P7-UX-001, P7-MOD-001
- **Goal:** align outward-facing docs with the shipped V2 direction
- **Tasks:**
  1. Update README, docs, and site-facing materials.
  2. Align onboarding and migration narratives.
  3. Remove stale descriptions of V2 scope where needed.
- **Validation:**
  - docs review
- **Done when:**
  - the repository tells a consistent V2 story

### P7-REL-001 — Run release closeout checklist
- **Owner:** orchestrator
- **Depends on:** P7-PERF-002, P7-OPS-001, P7-DOC-001
- **Goal:** close the rewrite with explicit ship/no-ship evidence
- **Tasks:**
  1. Re-check all phase acceptance gates.
  2. Record deferred work and residual risks.
  3. Publish a phase-closeout summary and release recommendation.
- **Validation:**
  - release checklist
  - final verifier evidence bundle
- **Done when:**
  - V2 can be recommended as the primary product direction with known risks documented

## Parallel-safe packets

After `P7-PERF-001`, these can run in parallel if they do not collide on files:

- `P7-PERF-002`
- `P7-OPS-001`
- `P7-MOD-001`

`P7-DOC-001` should wait until the major polish and module-boundary decisions are stable.

## Phase closeout criteria

This phase is done when:

- the release checklist passes
- deferred modules are clearly bounded
- docs are current
- remaining risks are documented rather than hidden
