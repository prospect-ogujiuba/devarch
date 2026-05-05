# 05 — Expose Workflows Through CLI and API

## Goal

Make `devarch` the operator interface and expose matching thin local API endpoints over the same appsvc methods.

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
