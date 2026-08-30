# Second-pass plan review

Reviewed: 2026-08-29
Decision: Approved after revisions
Scope: `docs/plans/devarch-home-qol/`

## Resolved findings

### High — No shared implementation/test contract

The feature plans required pure JavaScript tests and browser verification but did not select module boundaries or a runnable test architecture. Added [cross-cutting implementation contracts](cross-cutting-contracts.md) covering incremental ES-module extraction, Node `node:test`, Python `unittest`, Playwright browser coverage, HTTPS smoke tests, and phase-level file ownership.

### High — Runtime filter dependency was reversed

Plan 02 attempted to ship `hasRuntimeMatch` before plan 06 replaced the current naming heuristic. Plan 02 now reserves but does not ship that facet; plan 06 owns its relationship-backed implementation and unavailable-Podman state.

### High — Bookmark canonicalization could destroy valid links

The URL plan previously removed facet values absent from current inventory. A transient collector failure or catalog change could rewrite a bookmark destructively. Syntactically valid zero-count selections are now retained; limits, privacy wording, and a 2,048-byte query cap are explicit.

### High — Action-registry/editor scheme mismatch

The action registry claimed all detail actions but permitted only same-origin, HTTP, and copy actions while the app detail page already has a `vscode://file` handoff. Added an explicit non-palette `editor` action kind restricted to encoded workspace paths; arbitrary custom schemes remain forbidden.

### Medium — Vague quality and performance gates

Replaced subjective timing with shared median/p95 budgets, fixed fixture sizes, inventory response/size limits, Git deadlines, palette/filter budgets, storage limits, and browser breakpoint/zoom requirements.

### Medium — Optional API capabilities lacked versioning

Added a top-level `schemaVersion` migration contract, bounded capability states, legacy-version tolerance, additive-field deployment order, and sequential inventory-assembler integration for Git and relationship collectors.

### Medium — Refresh races and partial failure behavior were unspecified

Added one-in-flight request semantics, stale-response protection, last-successful-inventory preservation, optional collector isolation, and compatibility behavior during service restart.

### Medium — localStorage multi-tab behavior was unspecified

Added `storage`-event synchronization, load-validate-replace semantics, no echo writes, conflict behavior, and tests for invalid external-tab data.

### Medium — Security controls were distributed but incomplete

Added dynamic-DOM sink requirements, minimum CSP, URL/privacy boundaries, external-open protections, clipboard constraints, Host/local publication preservation, explicit API serialization, and negative tests.

### Medium — Copy-command scope had drifted

Removed the proposed container inspect shortcut because it approaches administration and can reveal environment secrets when run. Added descriptor schema versioning, provider compatibility checks, stale-target wording, bounded logs, and explicit state-changing labeling for the one existing copied start command.

### Low — Ranking and sorting varied by host locale

Specified explicit Unicode normalization and collator behavior, deterministic tie-breaks, non-virtualized 50-result palette rendering, and restrained live-region behavior.

## Residual risks

- Browser-critical phases require an environment with Playwright Chromium installed; the contract now blocks completion if only manual evidence exists.
- Git status cost varies by filesystem and repository size; hard deadlines and capability fallback bound the impact, but implementation benchmarks determine whether the proposed defaults need adjustment.
- Compose provider support differs; unsupported copy commands must be omitted rather than assumed.

## Final recommendation

Proceed in the program order. Treat `cross-cutting-contracts.md` as a mandatory prerequisite and update canonical plan status/evidence after each subphase. No plan now requires a Portainer-style management surface or background polling.
