---
name: devarch-v2-rules
description: Manifest-first working rules for the DevArch V2 rewrite. Use when planning, implementing, reviewing, or verifying any DevArch V2 packet.
---

# DevArch V2 Rules

These rules apply to every DevArch V2 packet and phase.

## Canonical product rules

1. **Manifest is truth**
   - Desired state lives in workspace manifests and catalog template files.
   - Runtime/cache/history data must never become the desired-state source.

2. **One engine, many surfaces**
   - CLI, API, and UI must call the same Go engine.
   - Transport code stays thin and should not reimplement engine logic.

3. **Workspace first**
   - Prefer workspace-centric flows over V1-style CRUD sprawl.
   - Avoid recreating admin-heavy pages or endpoints unless a phase explicitly requires them.

4. **Selective reuse only**
   - Mine `api/`, `dashboard/`, and `services-library/` for concepts, fixtures, and narrow implementation ideas.
   - Do not port the V1 database-first model, route tree, or handler sprawl forward as-is.

## Packet sizing rules

- One packet should have one primary owner.
- Keep packets independently reviewable and revertible.
- Prefer 2-3 concrete tasks and a clear validation command set.
- Escalate architectural surprises instead of widening scope silently.

## Domain boundaries

- `surgeon-engine`: root module, schemas, examples, `internal/spec`, `internal/workspace`, `internal/resolve`, `internal/contracts`, `internal/plan`, CLI engine wiring.
- `surgeon-catalog`: `catalog/`, builtin templates, `internal/catalog`, template indexing/discovery.
- `surgeon-runtime`: `internal/runtime`, `internal/apply`, `internal/cache`, runtime-backed operations.
- `surgeon-api`: `internal/api`, daemon bootstrap, event transport.
- `surgeon-ui`: `web/` only.
- `surgeon-import`: `internal/importv1`, importer fixtures, migration notes.
- `surgeon-tests`: cross-cutting test harnesses, goldens, parity fixtures.

## Delivery rules

- Read before editing.
- Prefer minimal diffs and keep file ownership obvious.
- Add or update tests/goldens when behavior contracts stabilize.
- Record intentional deviations explicitly in docs instead of hiding them in code.
- Do not overwrite unrelated user changes.

## Verification rules

- Schemas/examples: validate expected good and bad cases.
- Resolver/plan behavior: add deterministic golden coverage.
- Runtime/API/CLI: prove equivalent engine-backed behavior where the phase calls for it.
- UI: keep smoke coverage on the primary workspace flow.
- Import/parity: distinguish supported, lossy, and rejected cases.

## Phase gates

- Phase 1 before Phase 2: schemas, examples, and root module exist.
- Phase 2 before Phase 3: effective graph and contract-link behavior are deterministic and golden-tested.
- Phase 3 before Phase 4: plan/apply/status/logs/exec exist through runtime abstractions.
- Phase 4 before Phase 5: CLI/API share the same service boundary.
- Phase 5 before Phase 6: define -> plan -> apply -> observe works in the UI.
- Phase 6 before Phase 7: migration/parity evidence exists.

## Required references

Before widening scope, re-read:
- `docs/devarch-v2/devarch-v2-pi-plan-spec.md`
- `docs/devarch-v2/implementation-orchestration.md`
- the relevant `docs/devarch-v2/phase-*.md`
- `docs/rfc/000-devarch-v2-charter.md`
- `docs/adr/0001-v2-manifest-first.md`
- `docs/adr/0002-v2-runtime-derived-state.md`
- `docs/adr/0003-v2-thin-local-api.md`
