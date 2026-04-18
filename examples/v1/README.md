# DevArch V1 importer fixtures

These fixtures back the initial DevArch V2 Phase 6 importer/parity slice.

## Fixture matrix

| Fixture | Class | Covers | Expected outcome |
|---|---|---|---|
| `library/database/postgres/compose.yml` | simple | image/env/ports/volumes/health import from compose | imports as a valid V2 template |
| `library/backend/php/compose.yml` | edge-case | build context, working dir, bind + named volumes, labels, env files, networks, config files, `x-devarch-config`, container metadata | imports as a valid V2 template with explicit lossy diagnostics and compatibility data under `spec.develop.importv1` |
| `stacks/shop-export.yaml` | medium | stack metadata, instance overrides, secret refs, domains, config files, config mounts, explicit wires/imports/exports | imports as a valid V2 workspace and resolves against the builtin catalog |
| `stacks/rejected-missing-template.yaml` | intentionally unsupported | malformed export with an instance missing `template` | rejected with structured diagnostics |

## Known gaps in this slice

- V1 compose library inputs do not encode contract metadata, so imported templates currently emit empty `imports` / `exports` and warn explicitly.
- V1 stack network names, labels, env files, config files, config mounts, and other non-schema-native fields are preserved in compatibility buckets with diagnostics instead of silent normalization.
- The broader V1 `api/cmd/verify-parity` database-backed parity flow is reference material only. This slice uses fixture-backed import/resolve/render evidence inside `internal/importv1` tests.
