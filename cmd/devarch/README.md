# `devarch` CLI

Phase 4 exposes the shared engine through one CLI binary.

## Global flags

Place global flags before the command:

```bash
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin --json workspace plan shop-local
```

- `--workspace-root` repeatable workspace discovery root
- `--catalog-root` repeatable catalog discovery root
- `--json` stable machine-readable output

## Stable JSON contract

`--json` emits the same service-backed payload shapes used by the thin API where they already exist:

- `workspace list/open/plan/apply/status/logs/exec`
- `catalog list/show`
- `scan project`
- `import v1-stack`
- `import v1-library`

Human-readable output is operator-oriented and may change.

## Import output

`import v1-stack` and `import v1-library` now execute the Phase 6 importer slice.
JSON output returns structured statuses, diagnostics, and emitted artifact documents so callers can distinguish supported, lossy, and rejected mappings explicitly.
