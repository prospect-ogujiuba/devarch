# 03 — Complete the Podman Runtime Adapter

## Goal

Replace unsupported Podman mutation methods with real implementations backed by `internal/podmanctl`.

### Task 5: Wire `internal/runtime/podman` to `internal/podmanctl`

**Objective:** Make the Podman adapter capable of real network/container mutations.

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
