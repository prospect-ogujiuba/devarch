# 02 — Introduce a Testable Podman Command Layer

## Goal

Extract raw Podman command construction out of the adapter so the runtime adapter remains a translator, not a command-string monolith.

### Task 2: Create `internal/podmanctl` command runner foundation

**Objective:** Create the testable command-runner foundation for low-level Podman operations.

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
