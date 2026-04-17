# Phase 6 — V1 Importers + Parity Harness

## Goal
Make migration practical by importing representative V1 material into V2 manifests and checking parity where meaningful.

## Acceptance

- V1 service library content can be converted into V2 templates
- V1 exported stacks can be converted into V2 workspace manifests
- parity fixtures compare V1 output to V2 effective graph or rendered runtime payloads
- migration docs explain supported, partial, and intentionally rejected cases

## Recommended sequence

Start with fixture curation, then template importer, then workspace importer, then parity harness. Rollout and migration docs finish the phase.

---

### P6-FIX-001 — Curate importer and parity fixture matrix
- **Owner:** surgeon-import
- **Depends on:** Phase 1 fixture extraction complete
- **Goal:** choose the representative scenarios used for import and parity work
- **Tasks:**
  1. Categorize fixtures by simple, medium, edge-case, and intentionally unsupported.
  2. Add one matrix showing what each fixture covers.
  3. Identify known hard cases early.
- **Validation:**
  - fixture review
- **Done when:**
  - importer work has a stable test corpus

### P6-IMP-001 — Convert V1 library entries into V2 templates
- **Owner:** surgeon-import
- **Depends on:** P6-FIX-001
- **Goal:** build the V1 catalog importer
- **Tasks:**
  1. Map service template fields into V2 template files.
  2. Preserve names, tags, contracts, defaults, and health data where possible.
  3. Record lossy mappings explicitly.
- **Validation:**
  - `go test ./internal/importv1/...`
- **Done when:**
  - representative V1 templates can be emitted as valid V2 templates

### P6-IMP-002 — Convert V1 stack exports into V2 workspaces
- **Owner:** surgeon-import
- **Depends on:** P6-IMP-001
- **Goal:** build the V1 stack importer
- **Tasks:**
  1. Map stacks, instances, and projects into workspace/resources.
  2. Preserve wiring as imports/exports where possible.
  3. Emit compatibility resource shapes for cases that cannot be represented natively yet.
- **Validation:**
  - `go test ./internal/importv1/...`
- **Done when:**
  - representative V1 stacks import into valid V2 manifests or fail clearly

### P6-IMP-003 — Add importer diagnostics and rejection reporting
- **Owner:** surgeon-import
- **Depends on:** P6-IMP-002
- **Goal:** make unsupported cases explicit and actionable
- **Tasks:**
  1. Emit structured warnings and rejection reasons.
  2. Distinguish between lossy conversion and hard failure.
  3. Add test coverage for diagnostics.
- **Validation:**
  - `go test ./internal/importv1/...`
- **Done when:**
  - operators can tell why an import succeeded partially or failed

### P6-PAR-001 — Build parity harness
- **Owner:** surgeon-tests
- **Depends on:** P6-IMP-002, Phase 2 complete, Phase 3 complete
- **Goal:** compare V1 and V2 outputs for agreed fixtures
- **Tasks:**
  1. Define the comparison target for each fixture: compose, graph, or rendered runtime payload.
  2. Run fixture comparisons automatically.
  3. Support explicit whitelisting for intentional deviations.
- **Validation:**
  - parity suite run
- **Done when:**
  - parity claims are backed by repeatable checks

### P6-PAR-002 — Document intentional deviations
- **Owner:** scribe
- **Depends on:** P6-PAR-001
- **Goal:** prevent parity work from becoming hidden scope creep
- **Tasks:**
  1. Record every approved deviation from V1 behavior.
  2. Explain whether it is a V2 simplification, postponement, or bug backlog item.
  3. Link deviations to fixture IDs where possible.
- **Validation:**
  - docs review
- **Done when:**
  - parity gaps are visible and intentional

### P6-MIG-001 — Write migration operator docs
- **Owner:** scribe
- **Depends on:** P6-IMP-003, P6-PAR-002
- **Goal:** document how to move from V1 to V2 safely
- **Tasks:**
  1. Write the importer workflow and prerequisites.
  2. Document supported inputs and common failures.
  3. Explain shadow-mode rollout guidance for representative workspaces.
- **Validation:**
  - docs review against importer behavior
- **Done when:**
  - migration has clear operator-facing guidance

### P6-ROL-001 — Run shadow-mode rollout trial
- **Owner:** verifier
- **Depends on:** P6-MIG-001
- **Goal:** test V2 on representative real workspaces before full product shift
- **Tasks:**
  1. Pick representative workspaces.
  2. Run V2 in shadow mode beside V1 outputs.
  3. Capture blockers, deviations, and adoption notes.
- **Validation:**
  - rollout report with evidence
- **Done when:**
  - the team has real-world migration confidence data

## Parallel-safe packets

After `P6-FIX-001`, these can run in parallel with separate file ownership:

- `P6-IMP-001`
- `P6-PAR-001` prep work on fixture harness scaffolding

After `P6-PAR-001`, docs packets can overlap:

- `P6-PAR-002`
- `P6-MIG-001`

## Handoff to Phase 7

Phase 7 starts only after:

- representative imports work or fail clearly
- parity coverage exists for agreed fixtures
- migration docs and a shadow rollout report are available
