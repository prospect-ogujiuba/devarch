# DevArch V2 Worklog

Chronological notes captured by the repo-local pi worklog extension.

## 2026-04-17T17:15:58.439Z
- phase: phase-1
- packet: P1-SPEC-001,P1-SCHEMA-001,P1-SCHEMA-003,P1-EX-001,P1-REPO-001
- status: done
- summary: Implemented Phase 1 engine batch: added root github.com/prospect-ogujiuba/devarch module, workspace/plan schemas with internal/spec validation tests, three validating example workspaces, and placeholder V2 package boundaries.
- files: go.mod, cmd/devarch/main.go, cmd/devarchd/main.go, internal/spec/spec.go, internal/spec/spec_test.go, schemas/workspace.schema.json, schemas/plan.schema.json, examples/v2/workspaces/shop-local/devarch.workspace.yaml, examples/v2/workspaces/laravel-local/devarch.workspace.yaml, examples/v2/workspaces/compat-local/devarch.workspace.yaml, examples/v2/workspaces/compat-local/compose.yml, internal/workspace/doc.go, internal/resolve/doc.go, internal/contracts/doc.go, internal/plan/doc.go, internal/projectscan/doc.go, internal/catalog/doc.go, internal/api/doc.go, internal/apply/doc.go, internal/cache/doc.go, internal/events/doc.go, internal/importv1/doc.go, internal/runtime/doc.go, catalog/builtin/README.md, web/README.md

## 2026-04-17T17:53:51.364Z
- phase: phase-1
- packet: P1-SCHEMA-002
- status: done
- summary: Added template.schema.json and minimal internal/spec template validation helpers with schema tests covering valid, invalid, and contracts-block rejection cases.
- files: schemas/template.schema.json, internal/spec/spec.go, internal/spec/spec_test.go

## 2026-04-17T17:53:55.094Z
- phase: phase-1
- packet: P1-EX-002
- status: done
- summary: Seeded builtin template corpus for postgres, redis, node-api, laravel-app, vite-web, and nginx; added builtin template validation sweep and updated builtin catalog README.
- files: catalog/builtin/README.md, catalog/builtin/database/postgres/template.yaml, catalog/builtin/cache/redis/template.yaml, catalog/builtin/backend/node-api/template.yaml, catalog/builtin/backend/laravel-app/template.yaml, catalog/builtin/frontend/vite-web/template.yaml, catalog/builtin/proxy/nginx/template.yaml, internal/spec/spec_test.go

## 2026-04-17T17:54:59.333Z
- phase: phase-2
- packet: P2-CAT-001
- status: done
- summary: Implemented deterministic catalog discovery over resolved roots with canonical template.yaml filtering, path sorting, deduplication, and root-behavior tests.
- files: internal/catalog/discovery.go, internal/catalog/discovery_test.go

## 2026-04-17T17:57:17.084Z
- phase: phase-2
- packet: P2-CAT-002
- status: done
- summary: Added template loading/validation/indexing in internal/catalog with name, tag, import-contract, and export-contract lookups plus duplicate-name and invalid-template coverage.
- files: internal/catalog/index.go, internal/catalog/index_test.go

## 2026-04-17T18:46:53.892Z
- phase: phase-2
- packet: P2-WS-001/P2-WS-002/P2-RSLV-001/P2-RSLV-002/P2-RSLV-003/P2-CON-001/P2-GOLD-001
- status: done
- summary: Implemented Phase 2 workspace/resolver/contracts slice with typed workspace loading, deterministic effective graph resolution, contract-link interpolation/diagnostics, and phase2 goldens for shop-local, laravel-local, and ambiguous-http.
- files: internal/workspace/model.go, internal/workspace/load.go, internal/workspace/normalize.go, internal/workspace/load_test.go, internal/workspace/normalize_test.go, internal/resolve/model.go, internal/resolve/runtime.go, internal/resolve/merge.go, internal/resolve/resolve.go, internal/resolve/resolve_test.go, internal/resolve/golden_test.go, internal/contracts/model.go, internal/contracts/resolve.go, internal/contracts/interpolate.go, internal/contracts/resolve_test.go, testdata/goldens/README.md, testdata/goldens/phase2/fixtures/ambiguous-http/devarch.workspace.yaml, testdata/goldens/phase2/shop-local.resolved.golden.json, testdata/goldens/phase2/laravel-local.resolved.golden.json, testdata/goldens/phase2/ambiguous-http.resolved.golden.json

## 2026-04-17T19:52:53.286Z
- phase: phase-3
- packet: P3-RT-001/P3-RT-002/P3-PLAN-001/P3-APPLY-001/P3-APPLY-002/P3-RT-003/P3-CACHE-001/P3-EVT-001
- status: done
- summary: Implemented Phase 3 runtime/plan/apply/cache/events batch: runtime-owned desired/snapshot models, docker/podman inspect/logs/exec adapters, deterministic planner goldens, provider-neutral render/apply executor, optional SQLite cache, and canonical event bus/codecs with tests.
- files: internal/runtime/model.go, internal/runtime/desired.go, internal/runtime/adapter.go, internal/runtime/naming.go, internal/runtime/unsupported.go, internal/runtime/inspect.go, internal/runtime/observe.go, internal/runtime/docker/adapter.go, internal/runtime/docker/adapter_test.go, internal/runtime/podman/adapter.go, internal/runtime/podman/adapter_test.go, internal/runtime/inspect_test.go, internal/runtime/logs_exec_test.go, internal/plan/model.go, internal/plan/diff.go, internal/plan/reasons.go, internal/plan/golden_test.go, internal/apply/model.go, internal/apply/render.go, internal/apply/executor.go, internal/apply/render_test.go, internal/apply/executor_test.go, internal/cache/model.go, internal/cache/sqlite.go, internal/cache/sqlite_test.go, internal/events/model.go, internal/events/codec.go, internal/events/bus.go, internal/events/events_test.go, testdata/goldens/README.md, testdata/goldens/phase3/shop-local.plan.golden.json, testdata/goldens/phase3/laravel-local.plan.golden.json, testdata/goldens/phase3/compat-local.plan.golden.json, testdata/goldens/phase3/laravel-local.render.golden.json, testdata/goldens/phase3/shop-local.apply.golden.json, testdata/goldens/phase3/runtime-events.golden.json, go.mod, go.sum

## 2026-04-17T20:21:23.097Z
- phase: phase-4
- packet: P4-SVC-001/P4-API-001/P4-API-002/P4-DAEMON-001
- status: done
- summary: Scouted Phase 4 API slice context: reviewed Phase 4 docs plus current workspace/catalog/resolve/contracts/runtime/plan/apply/events/cache code; confirmed root internal/api and cmd/devarchd are still stubs, no shared service layer exists, and docker/podman adapters only cover inspect/logs/exec today.
- files: docs/devarch-v2/phase-4-cli-api.md, docs/devarch-v2/devarch-v2-pi-plan-spec.md, docs/adr/0003-v2-thin-local-api.md, docs/adr/0002-v2-runtime-derived-state.md, internal/workspace/load.go, internal/catalog/index.go, internal/resolve/resolve.go, internal/runtime/adapter.go, internal/runtime/docker/adapter.go, internal/runtime/podman/adapter.go, internal/apply/executor.go, internal/events/model.go, internal/api/doc.go, cmd/devarchd/main.go

## 2026-04-17T20:56:17.935Z
- phase: phase-4
- packet: P4-SVC-001/P4-API-001/P4-API-002/P4-DAEMON-001
- status: done
- summary: Implemented Phase 4 shared appsvc seam, thin catalog/workspace/runtime API routes, and devarchd bootstrap with capability-gated apply/logs/events/exec coverage and tests.
- files: internal/appsvc/service.go, internal/appsvc/discovery.go, internal/appsvc/model.go, internal/appsvc/service_test.go, internal/api/server.go, internal/api/catalog_handlers.go, internal/api/workspace_handlers.go, internal/api/runtime_handlers.go, internal/api/transport.go, internal/api/server_test.go, cmd/devarchd/main.go, cmd/devarchd/main_test.go, docs/devarch-v2/README.md, go.mod, go.sum

## 2026-04-17T21:49:18.801Z
- phase: phase-4
- packet: P4-CLI-001/P4-CLI-002/P4-CLI-003
- status: done
- summary: Implemented cmd/devarch Phase 4 CLI slice with workspace/catalog commands, shared --json output, project scan wiring, and deferred V1 import previews backed by appsvc/projectscan/importv1 tests.
- files: cmd/devarch/main.go, cmd/devarch/cli.go, cmd/devarch/cli_test.go, cmd/devarch/README.md, internal/appsvc/model.go, internal/appsvc/service.go, internal/appsvc/service_test.go, internal/projectscan/doc.go, internal/projectscan/scan.go, internal/projectscan/scan_test.go, internal/importv1/doc.go, internal/importv1/import.go, internal/importv1/import_test.go

## 2026-04-17T22:53:49.722Z
- phase: phase-5
- packet: P5-UI-001/P5-UI-002/P5-UI-003/P5-UI-004/P5-UI-005/P5-UI-006/P5-CAT-001
- status: done
- summary: Bootstrapped the Phase 5 web app under web/ with constrained four-area navigation, workspace-first detail tabs, catalog/activity/settings surfaces, thin Phase 4 API fetch hooks, validated local raw manifest drafting, and Vite/Vitest build wiring with a route smoke test.
- files: web/README.md, web/package.json, web/package-lock.json, web/vite.config.ts, web/tsconfig.json, web/index.html, web/src/app/App.tsx, web/src/app/AppShell.tsx, web/src/app/App.test.tsx, web/src/generated/api.ts, web/src/lib/api.ts, web/src/lib/format.ts, web/src/lib/settings.ts, web/src/lib/utils.ts, web/src/features/workspaces/hooks.ts, web/src/features/catalog/hooks.ts, web/src/features/activity/useWorkspaceEvents.ts, web/src/routes/WorkspacesPage.tsx, web/src/routes/CatalogPage.tsx, web/src/routes/ActivityPage.tsx, web/src/routes/SettingsPage.tsx, web/src/components/Card.tsx, web/src/components/CodeEditor.tsx, web/src/components/EmptyState.tsx, web/src/components/ErrorPanel.tsx, web/src/components/EventFeed.tsx, web/src/components/LoadingBlock.tsx, web/src/components/StatusBadge.tsx, web/src/index.css, web/src/main.tsx, web/src/test/setup.ts

## 2026-04-18T00:28:17.356Z
- phase: phase-6
- packet: P6-FIX-001/P6-IMP-001/P6-IMP-002/P6-IMP-003
- status: done
- summary: Implemented the initial Phase 6 import slice with fixture-backed V1 library/stack importers, explicit lossy/rejected diagnostics, emitted V2 artifact documents, and resolve-level parity evidence against builtin catalog fixtures.
- files: internal/importv1/import.go, internal/importv1/compose.go, internal/importv1/library.go, internal/importv1/stack.go, internal/importv1/import_test.go, internal/importv1/doc.go, internal/workspace/model.go, internal/appsvc/model.go, internal/appsvc/service_test.go, cmd/devarch/cli.go, cmd/devarch/cli_test.go, cmd/devarch/README.md, examples/v1/README.md, examples/v1/library/database/postgres/compose.yml, examples/v1/library/backend/php/compose.yml, examples/v1/library/backend/php/config/php.ini, examples/v1/library/backend/php/config/Dockerfile, examples/v1/stacks/shop-export.yaml, examples/v1/stacks/rejected-missing-template.yaml

## 2026-04-18T21:04:08.596Z
- phase: phase-0..phase-6
- status: done
- summary: Committed Phase 0-6 DevArch V2 baseline implementation snapshot as feat: bootstrap devarch v2 implementation baseline.
- files: .pi/settings.json, docs/rfc/000-devarch-v2-charter.md, docs/adr/0001-v2-manifest-first.md, go.mod, internal/spec/spec.go, internal/resolve/resolve.go, internal/appsvc/service.go, internal/api/server.go, cmd/devarch/cli.go, web/src/app/App.tsx, internal/importv1/import.go
