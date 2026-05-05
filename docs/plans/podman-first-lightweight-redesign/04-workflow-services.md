# 04 — Move Script Workflows into Go Services

## Goal

Create a Go workflow layer for behavior currently trapped in shell scripts, then expose those workflows through `internal/appsvc`.

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
