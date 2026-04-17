# Phase 1 — Spec + Schema + Repo Skeleton

## Goal
Lock the V2 data model, examples, fixture direction, and repository boundaries before building deeper engine behavior.

## Acceptance

- workspace and template schemas validate examples
- representative examples exist for core manifest shapes
- the root V2 module and package boundaries are committed
- fixture strategy for V1 extraction is defined and started

## Recommended sequence

Run schema packets first, then examples and fixtures, then repo skeleton cleanup.

---

### P1-SPEC-001 — Finalize workspace manifest shape
- **Owner:** architect
- **Depends on:** Phase 0 complete
- **Goal:** settle the first usable workspace manifest contract
- **Tasks:**
  1. Review the spec’s workspace fields and normalize required vs optional fields.
  2. Record any intentional omissions or simplifications.
  3. Align examples and schema expectations.
- **Validation:**
  - doc review
  - schema/examples consistency check
- **Done when:**
  - workspace manifest shape is stable enough for engine packets

### P1-SCHEMA-001 — Define `workspace.schema.json`
- **Owner:** surgeon-engine
- **Depends on:** P1-SPEC-001
- **Goal:** implement the first workspace schema draft
- **Tasks:**
  1. Encode top-level workspace requirements.
  2. Encode normalized resource fields.
  3. Support secret refs, ports, volumes, imports, exports, and health structures.
- **Validation:**
  - `go test ./internal/spec/...`
  - validate example workspaces
- **Done when:**
  - workspace schema validates expected examples and rejects malformed docs

### P1-SCHEMA-002 — Define `template.schema.json`
- **Owner:** surgeon-catalog
- **Depends on:** P1-SCHEMA-001
- **Goal:** implement the first template schema draft
- **Tasks:**
  1. Encode metadata and runtime shape.
  2. Encode ports, volumes, imports, exports, health, and develop fields.
  3. Align contract structures with workspace expectations.
- **Validation:**
  - `go test ./internal/spec/...`
  - validate builtin template examples
- **Done when:**
  - template schema supports the first catalog corpus

### P1-SCHEMA-003 — Define `plan.schema.json`
- **Owner:** surgeon-engine
- **Depends on:** P1-SCHEMA-001
- **Goal:** establish the initial plan document contract for testing and transport
- **Tasks:**
  1. Define plan metadata and action list shape.
  2. Capture action types and human-readable reasons.
  3. Add validation coverage in spec tests.
- **Validation:**
  - `go test ./internal/spec/...`
- **Done when:**
  - plan docs can be validated and reused by later phases

### P1-EX-001 — Add sample workspace manifests
- **Owner:** surgeon-engine
- **Depends on:** P1-SCHEMA-001
- **Goal:** create representative example workspaces
- **Tasks:**
  1. Add a simple app workspace.
  2. Add a project-backed workspace.
  3. Add a compatibility workspace with a compose-style source.
- **Validation:**
  - validate examples against schema
- **Done when:**
  - at least three examples cover the intended workspace shapes

### P1-EX-002 — Add builtin sample templates
- **Owner:** surgeon-catalog
- **Depends on:** P1-SCHEMA-002
- **Goal:** seed the initial catalog with representative templates
- **Tasks:**
  1. Add database, cache, backend, frontend, and proxy templates.
  2. Encode basic contracts and health examples.
  3. Keep each template plain-file and human-readable.
- **Validation:**
  - validate templates against schema
- **Done when:**
  - at least five sample templates exist and validate

### P1-FIX-001 — Extract V1 fixture corpus
- **Owner:** surgeon-import
- **Depends on:** P1-SPEC-001
- **Goal:** collect representative V1 material for later importer and parity work
- **Tasks:**
  1. Export representative V1 stacks.
  2. Snapshot representative V1 template/library inputs.
  3. Add a fixture inventory note describing coverage gaps.
- **Validation:**
  - fixture files present
  - sample set reviewed for breadth
- **Done when:**
  - later migration packets have real source material

### P1-REPO-001 — Create root V2 module skeleton
- **Owner:** surgeon-engine
- **Depends on:** P1-SCHEMA-001, P1-SCHEMA-002
- **Goal:** establish the package layout for V2 without migrating V1 code wholesale
- **Tasks:**
  1. Add `cmd/`, `internal/`, `schemas/`, `catalog/`, and `web/` root boundaries.
  2. Add minimal README or placeholder markers where needed.
  3. Keep V2 additions isolated from V1 packages.
- **Validation:**
  - `go test ./...` for the new root module scope where practical
- **Done when:**
  - V2 package boundaries are obvious and ready for Phase 2 work

## Parallel-safe packets

After `P1-SPEC-001`, these can run in parallel with architect approval:

- `P1-SCHEMA-002`
- `P1-SCHEMA-003`
- `P1-FIX-001`

After schemas settle, these can run in parallel:

- `P1-EX-001`
- `P1-EX-002`
- `P1-REPO-001`

## Handoff to Phase 2

Phase 2 starts only after:

- schemas are committed
- examples validate
- repo boundaries exist
- at least one V1 fixture set has been extracted
