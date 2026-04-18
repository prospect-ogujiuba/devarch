# DevArch V2 goldens

Phase 2 goldens lock the deterministic workspace -> resolve -> contracts pipeline.
Phase 3 goldens lock desired/runtime planning, render payloads, mocked apply execution, and canonical event serialization.

## Layout

- `phase2/*.resolved.golden.json` — serialized effective graph, resolved links, and diagnostics
- `phase2/fixtures/ambiguous-http/` — extra fixture manifests that exercise ambiguity behavior beyond the examples corpus
- `phase3/*.plan.golden.json` — deterministic desired-vs-empty-snapshot plans
- `phase3/*.render.golden.json` — provider-neutral apply render payloads
- `phase3/*.apply.golden.json` — mocked apply execution results
- `phase3/runtime-events.golden.json` — canonical event envelope serialization and ordering

## Updating goldens intentionally

1. Make the engine change.
2. Review the affected test output locally.
3. Re-run the relevant golden tests with updates enabled:
   - `DEVARCH_UPDATE_GOLDENS=1 go test ./internal/resolve/...`
   - `DEVARCH_UPDATE_GOLDENS=1 go test ./internal/plan/... ./internal/apply/... ./internal/events/...`
4. Review the resulting JSON diff to confirm resource order, action order, diagnostic order, event sequence, and path sanitization stayed deterministic.

## Guardrails

- Goldens must not contain machine-specific absolute paths.
- Resources must stay sorted by resource key.
- Links and diagnostics must stay sorted by consumer, contract, and provider.
- Secret refs may survive as structured values; composite strings must not silently flatten them.
- Runtime object names and workspace network names must remain distinct from logical resource host aliases.
- `compat-local` must stay explicitly unsupported in Phase 3 with no partial apply output.
