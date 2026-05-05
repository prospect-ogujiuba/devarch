# 00 — Overview

## Current Findings

- Root V2 already exists and matches the desired direction:
  - `cmd/devarch` — CLI surface.
  - `cmd/devarchd` — local daemon bootstrap.
  - `internal/appsvc` — shared service boundary.
  - `internal/runtime` — provider-neutral runtime interface.
  - `internal/runtime/podman` — current Podman adapter.
  - `internal/apply`, `internal/plan`, `internal/events`, `internal/cache` — engine pieces.
  - `internal/api` — thin local HTTP layer.
- Existing ADR/RFC already supports this direction:
  - `docs/rfc/000-devarch-v2-charter.md` says V2 is manifest-first, thin CLI/API over one engine, with `api/`, `dashboard/`, and `services-library/` as reference/migration sources.
  - `docs/adr/0003-v2-thin-local-api.md` says API handlers must delegate to shared engine services and avoid V1-style CRUD sprawl.
- Main gap: `internal/runtime/podman/adapter.go` can inspect/log/exec, but mutation methods are currently unsupported:
  - `EnsureNetwork`
  - `RemoveNetwork`
  - `ApplyResource`
  - `RemoveResource`
  - `RestartResource`
- Scripts that should become API/CLI-backed workflows:
  - `scripts/devarch-doctor.sh`
  - `scripts/socket-manager.sh`
  - `scripts/runtime-switcher.sh`
  - `scripts/service-manager.sh`
  - `scripts/init-databases.sh`
  - `scripts/generate-context.sh`
  - `scripts/ai-manager.sh`
  - `scripts/wordpress/*.sh`
  - `scripts/laravel/setup-laravel.sh`
- Legacy V1 Go API is large and should not keep accumulating product logic:
  - `api/internal/api/routes.go`
  - `api/internal/api/handlers/*`
  - `api/internal/orchestration/*`
  - `api/internal/project/*`
  - `api/internal/container/*`

## Target Package Shape

```text
cmd/devarch/          CLI transport only; calls internal/appsvc
cmd/devarchd/         local daemon bootstrap only; calls internal/api + appsvc
internal/api/         thin HTTP handlers over appsvc/workflows
internal/appsvc/      shared product/workflow facade for CLI + API
internal/runtime/     provider-neutral runtime contracts and models
internal/runtime/podman/ Podman adapter over podmanctl
internal/podmanctl/   low-level Podman command builder/runner package
internal/workflows/   Go replacements for script workflows
scripts/              temporary compatibility shims only
api/                  V1 reference/import source; do not expand
```

## Non-Negotiable Boundaries

1. **Podman-first:** Podman is the default runtime. Docker support may stay as compatibility, but it must not drive the architecture.
2. **No V1 expansion:** Do not add new product behavior to `api/` unless explicitly needed for V1 import/parity.
3. **Shared service layer:** CLI and API must call the same `internal/appsvc` methods.
4. **Thin transports:** No business logic in `cmd/devarch`, `cmd/devarchd`, or `internal/api` handlers.
5. **Scripts become clients:** Shell scripts should either disappear or call `devarch`; they should not remain the real implementation.
6. **Manifests are canonical:** No DB-first desired-state storage in V2.
7. **Runtime state is derived:** Podman inspection, status, logs, events, and cache are runtime-derived, not source of truth.

## Verification Checklist

Before considering the redesign complete:

- `go test ./...` passes or known legacy failures are documented.
- `go run ./cmd/devarch --help` shows the new operator surface.
- `go run ./cmd/devarch --json doctor` returns machine-readable diagnostics.
- Podman adapter reports apply/network capabilities.
- A simple workspace can `plan -> apply -> status` through root V2.
- API and CLI use the same appsvc methods.
- No new product routes/handlers were added to legacy `api/`.
- Shell scripts either call `devarch` or are explicitly documented as deprecated.
