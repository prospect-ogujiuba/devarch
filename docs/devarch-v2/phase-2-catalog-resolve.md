# Phase 2 — Catalog + Workspace Resolver

## Goal
Build the file-based loading path and deterministic effective graph generation from manifests and templates.

## Acceptance

- templates can be discovered and indexed
- workspaces can be loaded and normalized
- template defaults merge with workspace overrides deterministically
- imports/exports resolve with clear ambiguity handling
- golden tests capture effective graph behavior

## Recommended sequence

Catalog and workspace loading come first. Resolver merge packets come next. Contract-linking lands only after effective resource shapes are stable.

---

### P2-CAT-001 — Implement template discovery
- **Owner:** surgeon-catalog
- **Depends on:** Phase 1 complete
- **Goal:** discover template documents from builtin and workspace-local catalog roots
- **Tasks:**
  1. Walk configured catalog roots for template files.
  2. Normalize and sort discovered paths.
  3. Add tests for discovery behavior.
- **Validation:**
  - `go test ./internal/catalog/...`
- **Done when:**
  - template discovery is deterministic and test-covered

### P2-CAT-002 — Build template index
- **Owner:** surgeon-catalog
- **Depends on:** P2-CAT-001
- **Goal:** load templates into an in-memory index by name, tags, and contracts
- **Tasks:**
  1. Load and validate template files.
  2. Index by name and useful lookup attributes.
  3. Reject duplicate template names clearly.
- **Validation:**
  - `go test ./internal/catalog/...`
- **Done when:**
  - later engine code can resolve templates by stable lookups

### P2-WS-001 — Implement workspace loader
- **Owner:** surgeon-engine
- **Depends on:** Phase 1 complete
- **Goal:** load the canonical workspace manifest from disk
- **Tasks:**
  1. Load `devarch.workspace.yaml` from a given path or directory.
  2. Validate documents against schema.
  3. Decode into normalized in-memory models.
- **Validation:**
  - `go test ./internal/workspace/...`
- **Done when:**
  - workspaces can be loaded reproducibly from disk

### P2-WS-002 — Add workspace normalization
- **Owner:** surgeon-engine
- **Depends on:** P2-WS-001
- **Goal:** normalize workspace data before resolution
- **Tasks:**
  1. Default common fields such as enabled resources.
  2. Normalize empty collections and consistent ordering hooks.
  3. Add tests for normalization behavior.
- **Validation:**
  - `go test ./internal/workspace/...`
- **Done when:**
  - resolver input is stable and predictable

### P2-RSLV-001 — Define effective graph model
- **Owner:** architect
- **Depends on:** P2-CAT-002, P2-WS-002
- **Goal:** specify the in-memory effective graph shape used by resolve, contracts, and plan
- **Tasks:**
  1. Define the effective resource model.
  2. Clarify what is preserved from templates vs materialized from workspace overrides.
  3. Record intentional omissions for later phases.
- **Validation:**
  - design review
- **Done when:**
  - implementation packets have a stable graph target

### P2-RSLV-002 — Merge env and runtime defaults
- **Owner:** surgeon-engine
- **Depends on:** P2-RSLV-001
- **Goal:** merge template runtime defaults and workspace env overrides
- **Tasks:**
  1. Implement precedence rules.
  2. Support secret-ref values without flattening away semantics.
  3. Add golden tests for override precedence.
- **Validation:**
  - `go test ./internal/resolve/...`
- **Done when:**
  - workspace values win consistently where intended

### P2-RSLV-003 — Merge ports, volumes, health, and develop hints
- **Owner:** surgeon-engine
- **Depends on:** P2-RSLV-002
- **Goal:** complete the first full resource merge path
- **Tasks:**
  1. Merge ports and volumes predictably.
  2. Carry health and develop sections into the effective graph.
  3. Add golden tests for structural merges.
- **Validation:**
  - `go test ./internal/resolve/...`
- **Done when:**
  - effective resources contain normalized merged config

### P2-CON-001 — Implement contract-link resolution
- **Owner:** surgeon-engine
- **Depends on:** P2-RSLV-003
- **Goal:** resolve imports and exports into concrete links
- **Tasks:**
  1. Auto-link exact single matches.
  2. Surface unresolved and ambiguous cases explicitly.
  3. Attach derived connection env for resolved links.
- **Validation:**
  - `go test ./internal/contracts/...`
- **Done when:**
  - contract resolution behavior matches the spec rules

### P2-GOLD-001 — Add effective graph golden fixtures
- **Owner:** surgeon-tests
- **Depends on:** P2-CON-001
- **Goal:** lock deterministic resolver output with golden tests
- **Tasks:**
  1. Serialize normalized effective graph outputs.
  2. Add fixtures for simple, project-backed, and ambiguous-link scenarios.
  3. Document how to update goldens intentionally.
- **Validation:**
  - `go test ./internal/resolve/... ./internal/contracts/...`
- **Done when:**
  - effective graph behavior is regression-protected

## Parallel-safe packets

These are safe in parallel after Phase 1:

- `P2-CAT-001`
- `P2-WS-001`

These can overlap after their prerequisites are satisfied:

- `P2-CAT-002`
- `P2-WS-002`

Do not parallelize `P2-RSLV-*` and `P2-CON-001` unless the architect explicitly splits file ownership.

## Handoff to Phase 3

Phase 3 starts only after:

- effective graph output is deterministic
- contract-link rules are test-covered
- golden fixtures exist for representative scenarios
