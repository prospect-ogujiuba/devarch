# Podman-First Lightweight DevArch Redesign Implementation Plan

> **For Hermes:** Use subagent-driven-development or pi-coding-agent to implement this plan packet-by-packet. Do not expand the legacy `api/` monolith; root V2 is the product boundary.

**Goal:** Redesign DevArch so it is no longer a monolithic Go application that does everything, but a lightweight Podman control layer with one shared API/CLI surface for workflows currently handled by shell scripts.

**Architecture:** DevArch V2 remains manifest-first: workspace manifests and catalog templates are canonical desired state; runtime snapshots are derived from Podman. The CLI (`cmd/devarch`) and local daemon/API (`cmd/devarchd`, `internal/api`) must call a shared application service layer (`internal/appsvc`) instead of duplicating behavior. Script workflows move into small Go workflow packages and the legacy V1 `api/` tree is frozen as migration/reference material.

**Tech Stack:** Go, Podman CLI/API via testable command runners, chi-based local API, root V2 packages under `cmd/`, `internal/`, `schemas/`, `catalog/`, `examples/`, and `web/`.

---

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

---

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

---

## Non-Negotiable Boundaries

1. **Podman-first:** Podman is the default runtime. Docker support may stay as compatibility, but it must not drive the architecture.
2. **No V1 expansion:** Do not add new product behavior to `api/` unless explicitly needed for V1 import/parity.
3. **Shared service layer:** CLI and API must call the same `internal/appsvc` methods.
4. **Thin transports:** No business logic in `cmd/devarch`, `cmd/devarchd`, or `internal/api` handlers.
5. **Scripts become clients:** Shell scripts should either disappear or call `devarch`; they should not remain the real implementation.
6. **Manifests are canonical:** No DB-first desired-state storage in V2.
7. **Runtime state is derived:** Podman inspection, status, logs, events, and cache are runtime-derived, not source of truth.

---

## Phase 1 — Record the Podman-First Architecture Decision

### Task 1: Add ADR for Podman-first lightweight wrapper

**Objective:** Make the new direction explicit so future changes do not drift back into monolith/CRUD/script sprawl.

**Files:**
- Create: `docs/adr/0004-podman-first-lightweight-wrapper.md`

**Content outline:**

```markdown
# ADR 0004 — DevArch V2 is a Podman-first lightweight wrapper

- Status: Accepted
- Date: 2026-05-04

## Context

DevArch V1 grew into a DB-backed Go API with many CRUD handlers and separate scripts for operator workflows. The desired V2 product is a small local control plane over Podman, not another monolith.

## Decision

DevArch V2 is Podman-first. Workspace manifests and catalog templates are desired state. The root engine resolves manifests into runtime payloads, plans changes, and applies them through Podman. CLI and local API are transport layers over `internal/appsvc`. Script workflows move into small Go workflow services and scripts become compatibility shims.

## Consequences

- Root V2 packages are the product path.
- `api/` is frozen as V1 reference/import material.
- Docker remains optional compatibility only.
- Adding API endpoints requires an appsvc/workflow method first.
- Adding CLI commands requires the same shared method used by the API.
```

**Validation:**
- Read the ADR and confirm it does not contradict:
  - `docs/rfc/000-devarch-v2-charter.md`
  - `docs/adr/0003-v2-thin-local-api.md`

---

## Phase 2 — Introduce a Testable Podman Command Layer

### Task 2: Create `internal/podmanctl` command runner foundation

**Objective:** Extract raw Podman command construction out of the adapter so the runtime adapter remains a translator, not a command-string monolith.

**Files:**
- Create: `internal/podmanctl/runner.go`
- Create: `internal/podmanctl/runner_test.go`

**Implementation requirements:**
- Define a `Runner` interface equivalent to the existing command runner shape:
  - `Run(ctx context.Context, command string, args ...string) ([]byte, error)`
- Provide an `ExecRunner` implementation using `exec.CommandContext`.
- Provide a helper `Podman(ctx, runner, args...)` that always invokes `podman`.
- Tests must verify argument forwarding with a fake runner.

**Validation:**

```bash
go test ./internal/podmanctl/...
```

### Task 3: Add Podman network commands

**Objective:** Build reusable, tested Podman network helpers.

**Files:**
- Create: `internal/podmanctl/network.go`
- Create: `internal/podmanctl/network_test.go`

**Functions:**
- `NetworkExists(ctx, runner, name string) (bool, error)`
- `EnsureNetwork(ctx, runner, name string, labels map[string]string) error`
- `RemoveNetwork(ctx, runner, name string) error`

**Rules:**
- `EnsureNetwork` should be idempotent.
- Missing network on remove should not fail.
- Labels must be rendered as repeated `--label key=value` arguments.

**Validation:**

```bash
go test ./internal/podmanctl/...
```

### Task 4: Add Podman container command builder

**Objective:** Convert DevArch resource payloads into deterministic `podman run/create` arguments without executing real Podman in tests.

**Files:**
- Create: `internal/podmanctl/container.go`
- Create: `internal/podmanctl/container_test.go`

**Functions:**
- `BuildRunArgs(spec ContainerSpec) []string`
- `ApplyContainer(ctx, runner, spec ContainerSpec) error`
- `RemoveContainer(ctx, runner, name string) error`
- `RestartContainer(ctx, runner, name string) error`

**Spec fields should cover current `runtime.ResourceSpec`:**
- name
- image/build placeholder
- command/args
- env
- ports
- mounts/volumes
- labels
- network
- restart policy
- healthcheck where already modeled

**Validation:**

```bash
go test ./internal/podmanctl/...
```

---

## Phase 3 — Complete the Podman Runtime Adapter

### Task 5: Wire `internal/runtime/podman` to `internal/podmanctl`

**Objective:** Replace unsupported Podman mutation methods with real implementations.

**Files:**
- Modify: `internal/runtime/podman/adapter.go`
- Modify: `internal/runtime/podman/adapter_test.go`

**Required behavior:**
- `Capabilities()` returns:
  - `Inspect: true`
  - `Apply: true`
  - `Logs: true`
  - `Exec: true`
  - `Network: true`
- `EnsureNetwork` calls `podmanctl.EnsureNetwork`.
- `RemoveNetwork` calls `podmanctl.RemoveNetwork`.
- `ApplyResource` converts `runtime.ApplyResourceRequest` into a `podmanctl.ContainerSpec` and applies it.
- `RemoveResource` removes the runtime container by `ResourceRef.RuntimeName`.
- `RestartResource` restarts the runtime container by `ResourceRef.RuntimeName`.
- Unsupported behavior should be narrow and specific, not a blanket “apply mutations deferred”.

**Validation:**

```bash
go test ./internal/runtime/podman/...
go test ./internal/apply/... ./internal/runtime/...
```

### Task 6: Add an end-to-end Podman apply smoke fixture

**Objective:** Prove that a simple workspace can plan/apply/status through Podman without V1 API or scripts.

**Files:**
- Create or modify: `examples/v2/workspaces/podman-smoke/devarch.yaml`
- Create: `internal/appsvc/podman_smoke_test.go` or equivalent test gated behind an integration flag.

**Rules:**
- Keep the default test suite mock-based.
- Live Podman integration must be opt-in, e.g. `DEVARCH_INTEGRATION=podman go test ...`.

**Validation:**

```bash
go test ./internal/appsvc/...
DEVARCH_INTEGRATION=podman go test ./internal/appsvc/... -run PodmanSmoke
```

---

## Phase 4 — Move Script Workflows into Go Services

### Task 7: Create `internal/workflows` package foundation

**Objective:** Establish a home for workflows currently implemented as scripts.

**Files:**
- Create: `internal/workflows/model.go`
- Create: `internal/workflows/runner.go`
- Create: `internal/workflows/doc.go`

**Core model:**
- `Diagnostic`
- `CheckResult`
- `CommandResult`
- `WorkflowStatus`
- shared runner interface for host commands where unavoidable

**Validation:**

```bash
go test ./internal/workflows/...
```

### Task 8: Port doctor workflow

**Objective:** Replace `scripts/devarch-doctor.sh` with a Go workflow.

**Files:**
- Create: `internal/workflows/doctor.go`
- Create: `internal/workflows/doctor_test.go`
- Modify: `internal/appsvc/service.go`
- Modify: `internal/appsvc/model.go`

**App service method:**

```go
Doctor(ctx context.Context) (*workflows.DoctorReport, error)
```

**Checks:**
- `podman` available
- Podman socket status
- workspace roots readable
- catalog roots readable
- root V2 module builds or at least package discovery is valid

**Validation:**

```bash
go test ./internal/workflows/... ./internal/appsvc/...
```

### Task 9: Port runtime/socket workflow

**Objective:** Replace `scripts/runtime-switcher.sh` and `scripts/socket-manager.sh` with Go workflow methods.

**Files:**
- Create: `internal/workflows/runtime.go`
- Create: `internal/workflows/socket.go`
- Create tests for both.
- Modify: `internal/appsvc/service.go`
- Modify: `internal/appsvc/model.go`

**App service methods:**

```go
RuntimeStatus(ctx context.Context) (*workflows.RuntimeStatus, error)
SocketStatus(ctx context.Context) (*workflows.SocketStatus, error)
SocketStart(ctx context.Context) (*workflows.CommandResult, error)
SocketStop(ctx context.Context) (*workflows.CommandResult, error)
```

**Validation:**

```bash
go test ./internal/workflows/... ./internal/appsvc/...
```

### Task 10: Port service-manager workflow onto workspace runtime operations

**Objective:** Replace broad service-manager script behavior with workspace/resource operations backed by appsvc/runtime.

**Files:**
- Create: `internal/workflows/service.go`
- Modify: `internal/appsvc/service.go`
- Modify: `cmd/devarch/cli.go`
- Modify: `internal/api/runtime_handlers.go` or create `internal/api/workflow_handlers.go`

**Important:** Prefer existing workspace/resource terminology over reviving V1 “service CRUD”.

**Commands/endpoints should map to:**
- workspace status
- workspace apply
- resource logs
- resource exec
- resource restart

**Validation:**

```bash
go test ./cmd/devarch/... ./internal/api/... ./internal/appsvc/... ./internal/workflows/...
```

---

## Phase 5 — Expose Workflows Through CLI and API

### Task 11: Add CLI command groups for workflow surfaces

**Objective:** Make `devarch` the operator interface that scripts used to be.

**Files:**
- Modify: `cmd/devarch/cli.go`
- Modify: `cmd/devarch/cli_test.go`
- Modify: `cmd/devarch/README.md`

**Command groups:**

```bash
devarch doctor
devarch runtime status
devarch socket status
devarch socket start
devarch socket stop
devarch workspace status <name>
devarch workspace apply <name>
devarch workspace logs <name> <resource>
devarch workspace exec <name> <resource> -- <command...>
```

**Rules:**
- Support `--json` for automation.
- Keep command handlers thin.
- All behavior must call `internal/appsvc`.

**Validation:**

```bash
go test ./cmd/devarch/...
go run ./cmd/devarch --help
go run ./cmd/devarch --json doctor
```

### Task 12: Add thin local API workflow endpoints

**Objective:** Allow the UI and external tools to call the same workflow operations as the CLI.

**Files:**
- Create: `internal/api/workflow_handlers.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/server_test.go`

**Endpoints:**

```text
GET  /api/v1/doctor
GET  /api/v1/runtime/status
GET  /api/v1/socket/status
POST /api/v1/socket/start
POST /api/v1/socket/stop
```

**Rules:**
- Handlers call appsvc only.
- HTTP response shape mirrors CLI JSON shape where practical.
- No shell implementation hidden in handlers.

**Validation:**

```bash
go test ./internal/api/...
```

---

## Phase 6 — Retire Scripts into Compatibility Shims

### Task 13: Replace script implementations with `devarch` shims after parity

**Objective:** Stop scripts from being the source of truth.

**Files:**
- Modify: `scripts/devarch-doctor.sh`
- Modify: `scripts/socket-manager.sh`
- Modify: `scripts/runtime-switcher.sh`
- Modify: `scripts/service-manager.sh`
- Create: `docs/devarch-v2/script-migration.md`

**Example shim pattern:**

```bash
#!/usr/bin/env bash
set -euo pipefail
exec devarch doctor "$@"
```

**Validation:**

```bash
go test ./...
bash -n scripts/devarch-doctor.sh scripts/socket-manager.sh scripts/runtime-switcher.sh scripts/service-manager.sh
```

---

## Suggested Pi Delegation Packets

### Packet A — ADR + command layer

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement Phase 1 and Phase 2 from docs/plans/2026-05-04-podman-first-lightweight-redesign.md.
Do not touch legacy api/ except to read context.
Run go test ./internal/podmanctl/... and report files changed, tests run, and any blockers.
"
```

### Packet B — Podman adapter mutations

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement Phase 3 from docs/plans/2026-05-04-podman-first-lightweight-redesign.md.
Use internal/podmanctl for command construction.
Do not touch legacy api/.
Run go test ./internal/runtime/podman/... ./internal/apply/... ./internal/runtime/... and report files changed, tests run, and blockers.
"
```

### Packet C — workflow services

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement Phase 4 from docs/plans/2026-05-04-podman-first-lightweight-redesign.md.
Port doctor/runtime/socket workflows into internal/workflows and expose them through internal/appsvc.
Do not modify scripts yet except to read behavior.
Run go test ./internal/workflows/... ./internal/appsvc/... and report files changed, tests run, and blockers.
"
```

### Packet D — CLI/API exposure

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement Phase 5 from docs/plans/2026-05-04-podman-first-lightweight-redesign.md.
Keep CLI and API thin over internal/appsvc.
Run go test ./cmd/devarch/... ./internal/api/... ./internal/appsvc/... and report files changed, tests run, and blockers.
"
```

---

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

---

## Immediate Next Step

Start with **Packet A**. It is low-risk, establishes the architectural decision, and creates the reusable Podman command layer needed before mutating the runtime adapter.
