# DevArch V2 Task Breakdown

This folder turns `devarch-v2-pi-plan-spec.md` into flat, phase-by-phase execution packets.

## Folder shape

This task breakdown is intentionally **organized but not nested**:

- `README.md`
- `phase-0-pi-foundation.md`
- `phase-1-spec-schema.md`
- `phase-2-catalog-resolve.md`
- `phase-3-runtime-plan-apply.md`
- `phase-4-cli-api.md`
- `phase-5-web-ui.md`
- `phase-6-import-parity.md`
- `phase-7-polish-modules.md`

## How to use this breakdown

- Treat each packet as a small, independently shippable slice.
- Execute packets sequentially unless a phase file explicitly marks packets as parallel-safe.
- Keep packet scope to one primary domain owner.
- Prefer `scout -> architect -> surgeon-* -> reviewer -> verifier -> scribe` for implementation packets.
- Prefer `scout -> architect -> scribe` for planning or batch-definition packets.

## Packet format

Each phase file uses the same packet structure:

- **ID** — stable packet identifier
- **Owner** — primary surgeon or control role
- **Goal** — single clear outcome
- **Depends on** — packet prerequisites
- **Tasks** — 2-3 implementation actions
- **Validation** — checks the verifier should run
- **Done when** — concrete completion criteria

## Suggested execution order

1. `phase-0-pi-foundation.md`
2. `phase-1-spec-schema.md`
3. `phase-2-catalog-resolve.md`
4. `phase-3-runtime-plan-apply.md`
5. `phase-4-cli-api.md`
6. `phase-5-web-ui.md`
7. `phase-6-import-parity.md`
8. `phase-7-polish-modules.md`

## Global rules

- Manifest files remain the canonical desired state.
- Runtime data is derived or cached, never canonical.
- API work stays thin and engine-backed.
- UI stays workspace-first and avoids V1 CRUD sprawl.
- No packet should widen into postponed concerns unless the phase explicitly calls for module boundaries.

## Parallelization guidance

Parallelize only when packets:

- do not touch the same package boundary,
- do not depend on unresolved schema/model decisions,
- and have architect-approved file ownership splits.

Good parallel examples:

- catalog indexing + V1 fixture extraction
- UI logs view + API event endpoint polish
- importer mapping + parity fixture authoring

Bad parallel examples:

- resolver merge semantics + contract-link output changes in the same files
- CLI command design + engine API shape changes before engine interfaces settle
