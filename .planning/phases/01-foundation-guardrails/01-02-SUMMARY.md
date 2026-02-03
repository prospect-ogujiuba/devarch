---
phase: 01-foundation-guardrails
plan: 02
subsystem: container-runtime
tags: [go, runtime-abstraction, podman, docker, refactor]
requires: [01-01]
provides:
  - Unified runtime interface implementation in container.Client
  - DEVARCH_RUNTIME env var support with prescriptive errors
  - Zero hardcoded exec.Command(podman|docker) outside container/client.go
affects: [01-03, 04-network-isolation]
tech-stack:
  added: []
  patterns: [runtime-interface, env-var-override, compile-time-checks]
key-files:
  created:
    - api/internal/container/client_test.go
  modified:
    - api/internal/container/client.go
    - api/internal/project/controller.go
    - api/internal/api/handlers/nginx.go
    - api/cmd/server/main.go
    - api/internal/api/routes.go
decisions:
  - name: DEVARCH_RUNTIME env var override
    rationale: Explicit runtime control needed for CI/CD and testing environments
    impact: Users can force specific runtime instead of auto-detect
  - name: Network stubs return ErrNotImplemented
    rationale: Signal future Phase 4 implementation without breaking current code
    impact: Network operations fail fast with clear message
  - name: Preserve backward compat methods (GetStatus/GetMetrics)
    rationale: Avoid breaking existing code during refactor
    impact: Both old and new interface methods coexist temporarily
metrics:
  duration: 3m
  completed: 2026-02-03
---

# Phase 01 Plan 02: Runtime Abstraction Refactor Summary

**One-liner:** Unified runtime client implements Runtime interface; DEVARCH_RUNTIME override; zero hardcoded podman/docker calls outside container/client.go

## What Was Built

Extended `container.Client` to implement the `Runtime` interface from 01-01, eliminating all hardcoded `exec.Command("podman"|"docker")` calls throughout the codebase.

**Key changes:**
1. **DEVARCH_RUNTIME env var support** — Explicit runtime override with prescriptive errors when specified runtime not found
2. **Runtime interface implementation** — Exec, ListContainersWithLabels, GetContainerStatus, GetContainerMetrics, RunCompose methods
3. **Network operation stubs** — CreateNetwork, RemoveNetwork, ListNetworks return ErrNotImplemented (Phase 4)
4. **Refactored hardcoded calls** — project.Controller and nginx.Handler now route through container.Client
5. **Compile-time interface check** — `var _ Runtime = (*Client)(nil)` ensures interface satisfaction

## Deviations from Plan

None — plan executed exactly as written.

## Technical Decisions

### DEVARCH_RUNTIME Override Behavior

**Decision:** Check env var first before auto-detection; fail with prescriptive error if set but runtime not found.

**Rationale:**
- CI/CD environments need explicit control
- Auto-detection works for local dev
- Prescriptive errors guide users to solution

**Implementation:**
```go
if envRuntime := os.Getenv("DEVARCH_RUNTIME"); envRuntime != "" {
    switch envRuntime {
    case "podman":
        if _, err := exec.LookPath("podman"); err != nil {
            return nil, fmt.Errorf("DEVARCH_RUNTIME=podman but podman not found — install podman or unset DEVARCH_RUNTIME for auto-detection")
        }
    // ...
    }
}
```

### Backward Compatibility Layer

**Decision:** Keep existing GetStatus/GetMetrics methods alongside new GetContainerStatus/GetContainerMetrics.

**Why:** Avoid breaking existing handlers (serviceHandler, statusHandler) during refactor. Will consolidate in future cleanup phase.

### Network Stubs with ErrNotImplemented

**Decision:** Network methods return `fmt.Errorf("not implemented: will be available in Phase 4")`.

**Why:**
- Clear signal to future implementers
- Fails fast rather than silently
- Satisfies Runtime interface without premature implementation

## Verification Results

```bash
✓ cd api && go build ./...          # Full API compiles
✓ cd api && go vet ./...            # No vet warnings
✓ go test ./internal/container/...  # All tests pass (docker test skipped when not installed)
✓ grep hardcoded runtime calls      # Zero results outside container/client.go
```

## Commits

| Commit | Type | Description | Files |
|--------|------|-------------|-------|
| 4202af6 | feat | Implement Runtime interface in container.Client | client.go, client_test.go |
| a496fff | refactor | Route all container ops through container.Client | controller.go, nginx.go, main.go, routes.go |

## Next Phase Readiness

**Blockers:** None

**Concerns:** None

**Validates:** BASE-03 (all container operations through container.Client)

**Enables:**
- Phase 4: Network isolation implementation (stubs ready)
- 01-03: Stack/instance container lifecycle operations can use Runtime interface

## Learnings

1. **Prescriptive errors work** — DEVARCH_RUNTIME error messages guide users exactly what to do
2. **Compile-time checks catch issues early** — Interface check prevented signature mismatches
3. **Backward compat preserved velocity** — Didn't need to refactor all handlers in one go
