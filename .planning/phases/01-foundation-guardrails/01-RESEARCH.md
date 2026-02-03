# Phase 1: Foundation & Guardrails - Research

**Researched:** 2026-02-03
**Domain:** Container runtime abstraction, validation, label management
**Confidence:** MEDIUM

## Summary

Phase 1 establishes the foundational primitives for stack isolation: identity labels, validation helpers, and runtime abstraction. The existing codebase has two parallel container client implementations: (1) a shell-exec wrapper in `internal/container/client.go` that shells out to docker/podman CLI, and (2) a HTTP-based Podman client in `internal/podman/client.go` that speaks to the Podman REST API. Phase 1 must unify these under a single abstraction that supports both Docker and Podman with feature parity.

Key findings: Kubernetes validation patterns provide battle-tested DNS-safe naming rules. Container label filtering is standard practice with well-defined syntax across Docker/Podman APIs. The Podman Go bindings (`github.com/containers/podman/v5`) offer native API access, but Docker API compatibility means a unified HTTP client approach is viable.

**Primary recommendation:** Extend existing `internal/container/client.go` to be the unified abstraction layer. Add label-based filtering and network/pod stubs. Keep `internal/podman/` as implementation detail. Remove hardcoded `exec.Command("podman")` calls in `project/controller.go` and `nginx.go`.

## Standard Stack

The established libraries/tools for container runtime abstraction in Go:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| stdlib `os/exec` | Go 1.22 | CLI execution fallback | Already in use, zero dependencies, universal runtime support |
| stdlib `net/http` | Go 1.22 | API client transport | Powers Podman HTTP API client, Docker socket compatibility |
| `github.com/containers/podman/v5` | v5.x | Official Podman Go bindings | Canonical Podman API, pod management, rootless support |
| `github.com/docker/docker` | v25.x | Official Docker client SDK | Docker Engine API, compatible with Podman Docker socket emulation |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gopkg.in/yaml.v3` | v3.x | Compose YAML generation | Already in project, needed for compose generation |
| stdlib `regexp` | Go 1.22 | Validation patterns | DNS-safe name validation, no external deps |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Dual HTTP clients (Podman REST + Docker API) | Single exec-based wrapper | HTTP APIs are faster, more robust, provide structured errors vs parsing CLI output |
| Custom abstraction | containerd client directly | containerd is lower-level, doesn't support Compose, requires CRI knowledge |
| go-playground/validator | stdlib validation functions | validator adds dependency for simple DNS-safe regex, overkill for 3-4 validation rules |

**Installation:**
```bash
# Existing project already has yaml.v3 and stdlib
# If adding Podman bindings (optional, current HTTP client works):
go get github.com/containers/podman/v5
```

## Architecture Patterns

### Recommended Abstraction Structure
```
api/internal/
├── container/
│   ├── client.go          # Unified interface + factory
│   ├── labels.go          # Label constants + helpers (NEW)
│   ├── validation.go      # Name validation (NEW)
│   └── types.go           # Shared types (NEW)
├── podman/                # Podman HTTP client implementation (existing)
│   ├── client.go
│   ├── containers.go
│   ├── pods.go            # Pod operations (STUB in Phase 1)
│   └── networks.go        # Network operations (STUB in Phase 1)
└── docker/                # Docker HTTP client (future, optional)
    └── client.go
```

### Pattern 1: Unified Client Interface
**What:** Single `container.Client` interface with runtime-specific implementations behind it
**When to use:** All container operations in handlers, controllers, generators

**Example:**
```go
// internal/container/client.go
package container

type Client interface {
    // Existing operations
    RuntimeName() string
    StartService(name string, composeYAML []byte) error
    GetStatus(name string) (*models.ContainerState, error)
    ListContainers() ([]string, error)

    // NEW: Label-based queries (Phase 1)
    ListContainersWithLabels(labels map[string]string) ([]string, error)

    // NEW: Network operations (stub in Phase 1, implement Phase 4)
    EnsureNetwork(name string, driver string) error
    RemoveNetwork(name string) error

    // NEW: Pod operations (stub in Phase 1, implement Phase 4+)
    CreatePod(name string, labels map[string]string) error
    RemovePod(name string) error
}

// Factory auto-detects runtime
func NewClient() (Client, error) {
    // Prefer Podman, fallback Docker, check DEVARCH_RUNTIME env override
    runtime := os.Getenv("DEVARCH_RUNTIME")
    if runtime == "" {
        if _, err := exec.LookPath("podman"); err == nil {
            runtime = "podman"
        } else if _, err := exec.LookPath("docker"); err == nil {
            runtime = "docker"
        }
    }

    switch runtime {
    case "podman":
        return newPodmanClient()
    case "docker":
        return newDockerClient()
    default:
        return nil, fmt.Errorf("no container runtime found")
    }
}
```

### Pattern 2: Label Management Constants
**What:** Centralized label constants prevent typos, enable IDE autocomplete
**When to use:** Everywhere labels are applied or queried

**Example:**
```go
// internal/container/labels.go
package container

const (
    // Identity labels - core primitive for stack isolation
    LabelStackID           = "devarch.stack_id"
    LabelInstanceID        = "devarch.instance_id"
    LabelTemplateServiceID = "devarch.template_service_id"

    // Metadata labels
    LabelManagedBy = "devarch.managed_by"
    LabelVersion   = "devarch.version"
)

// BuildIdentityLabels creates standard label set for stack resources
func BuildIdentityLabels(stackID, instanceID, templateServiceID string) map[string]string {
    return map[string]string{
        LabelStackID:           stackID,
        LabelInstanceID:        instanceID,
        LabelTemplateServiceID: templateServiceID,
        LabelManagedBy:         "devarch",
        LabelVersion:           "1.0",
    }
}

// MatchStackLabels returns filter for querying all containers in a stack
func MatchStackLabels(stackID string) map[string]string {
    return map[string]string{
        LabelStackID: stackID,
    }
}
```

### Pattern 3: DNS-Safe Validation
**What:** RFC 1123 compliant validation (Kubernetes namespace rules)
**When to use:** Before creating stacks, instances, networks

**Example:**
```go
// internal/container/validation.go
package container

import (
    "fmt"
    "regexp"
    "strings"
)

const (
    MaxNameLength = 63
    MinNameLength = 1
)

var (
    // DNS label validation: lowercase alphanumeric + hyphens, no leading/trailing hyphens
    dnsLabelRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

    reservedNames = map[string]bool{
        "default": true,
        "devarch": true,
        "system":  true,
        "none":    true,
        "all":     true,
    }
)

type ValidationError struct {
    Field   string
    Value   string
    Problem string
    Suggest string
}

func (e ValidationError) Error() string {
    if e.Suggest != "" {
        return fmt.Sprintf("%s '%s' is invalid: %s. Try: %s", e.Field, e.Value, e.Problem, e.Suggest)
    }
    return fmt.Sprintf("%s '%s' is invalid: %s", e.Field, e.Value, e.Problem)
}

// ValidateStackName validates stack name with prescriptive errors
func ValidateStackName(name string) error {
    if name == "" {
        return ValidationError{
            Field:   "stack name",
            Value:   name,
            Problem: "cannot be empty",
            Suggest: "use lowercase letters and hyphens (e.g., 'my-stack')",
        }
    }

    if len(name) > MaxNameLength {
        return ValidationError{
            Field:   "stack name",
            Value:   name,
            Problem: fmt.Sprintf("exceeds %d characters", MaxNameLength),
            Suggest: name[:MaxNameLength],
        }
    }

    if reservedNames[name] {
        return ValidationError{
            Field:   "stack name",
            Value:   name,
            Problem: "reserved name",
            Suggest: name + "-stack",
        }
    }

    if !dnsLabelRegex.MatchString(name) {
        return ValidationError{
            Field:   "stack name",
            Value:   name,
            Problem: "must use lowercase letters, numbers, hyphens only (no leading/trailing hyphens)",
            Suggest: slugify(name),
        }
    }

    return nil
}

// ValidateInstanceID validates instance ID (same rules as stack name)
func ValidateInstanceID(id string) error {
    // Same logic as ValidateStackName
    return ValidateStackName(id) // reuse logic, adjust error field
}

// slugify converts arbitrary string to DNS-safe slug
func slugify(s string) string {
    s = strings.ToLower(s)
    s = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(s, "-")
    s = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(s, "")
    if len(s) > MaxNameLength {
        s = s[:MaxNameLength]
    }
    return s
}
```

### Pattern 4: Runtime Detection with Override
**What:** Auto-detect Podman/Docker, allow env var override
**When to use:** Client initialization

**Example:**
```go
// Auto-detect with Podman preferred
func detectRuntime() string {
    if override := os.Getenv("DEVARCH_RUNTIME"); override != "" {
        return override // docker or podman
    }

    // Prefer Podman (project decision: Podman first-class citizen)
    if _, err := exec.LookPath("podman"); err == nil {
        return "podman"
    }

    if _, err := exec.LookPath("docker"); err == nil {
        return "docker"
    }

    return ""
}
```

### Anti-Patterns to Avoid
- **Hardcoded runtime calls:** Never call `exec.Command("podman", ...)` directly in handlers/controllers. Always route through `container.Client`.
- **Uppercase labels:** Docker/Podman label keys are case-sensitive. Use lowercase consistently (`devarch.stack_id`, not `DevArch.StackID`).
- **Missing label validation:** Don't assume label values are safe. Validate before injecting into container configs.
- **String concatenation for names:** Use `fmt.Sprintf("devarch-%s-%s", stackID, instanceID)` not `"devarch-" + stackID + "-" + instanceID` (easier to audit naming pattern).

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Name validation | Custom validation rules | Kubernetes DNS label pattern (RFC 1123) | Battle-tested, matches ecosystem expectations, prevents subtle edge cases |
| Label filtering API | Custom query builder | Podman/Docker native filter syntax (`label=key=value`) | Consistent with CLI, documented, handles escaping |
| Container name conflicts | Uniqueness checks in DB | Deterministic naming + label queries | Name collisions fail loudly at runtime, labels provide safe querying |
| Network creation idempotency | Manual existence checks | API-level create-if-not-exists (Docker/Podman handle this) | Race conditions, atomic operations |
| Rootless socket detection | Hardcoded paths | Check `$XDG_RUNTIME_DIR/podman/podman.sock` then fallback | Respects user environment, works across distros |
| Validation error messages | Generic "invalid name" | Prescriptive errors: "must be lowercase. Try: 'my-stack'" | User fixes issue immediately instead of guessing |

**Key insight:** Container runtime abstraction complexity comes from subtle behavioral differences (Podman pods vs Docker networks, rootless permissions, socket paths). Use standard library detection patterns and official client libraries where possible. Don't parse CLI output when HTTP APIs exist.

## Common Pitfalls

### Pitfall 1: Label Filter Exclusivity
**What goes wrong:** Using multiple label filters with same key doesn't work as expected (exclusive in Podman, not inclusive)
**Why it happens:** Podman docs state "Filters with the same key work inclusive with the only exception being `label` which is exclusive"
**How to avoid:** Use single label filter with multiple key-value pairs passed as map, or combine results client-side
**Warning signs:** Label queries return zero results when multiple filters applied

### Pitfall 2: Rootless Port Binding
**What goes wrong:** Containers fail to bind ports < 1024 in rootless mode
**Why it happens:** Non-root users can't bind privileged ports without sysctl/capabilities changes
**How to avoid:** Validate port ranges in stack instances (warn if < 1024 and runtime is rootless), document port mapping strategy
**Warning signs:** Container starts but port binding fails silently or with cryptic error

### Pitfall 3: Hardcoded Podman Calls
**What goes wrong:** Docker users can't use DevArch, code breaks when switching runtimes
**Why it happens:** `exec.Command("podman", ...)` bypasses abstraction layer (found in `project/controller.go` line 83, `nginx.go` line 40)
**How to avoid:** Grep codebase for `exec.Command.*podman|docker`, route through `container.Client`
**Warning signs:** Runtime detection exists but some operations still shell out to hardcoded command

### Pitfall 4: Case-Sensitive Label Keys
**What goes wrong:** Labels with mixed case (DevArch.StackID) don't match queries (devarch.stack_id)
**Why it happens:** Docker/Podman treat label keys as case-sensitive strings
**How to avoid:** Use lowercase constants, enforce at validation layer, never accept user-provided label keys
**Warning signs:** Containers exist but label queries return empty results

### Pitfall 5: Name Validation After Creation
**What goes wrong:** Invalid names reach container runtime, fail with cryptic errors
**Why it happens:** Validation happens in DB layer but not before calling runtime
**How to avoid:** Validate at API boundary (handler input validation) before touching DB or runtime
**Warning signs:** DB has record but container creation failed, orphaned DB rows

### Pitfall 6: Socket Path Assumptions
**What goes wrong:** Rootless Podman client can't find socket at `/var/run/podman/podman.sock`
**Why it happens:** Rootless socket lives at `$XDG_RUNTIME_DIR/podman/podman.sock` (e.g., `/run/user/1000/podman/podman.sock`)
**How to avoid:** Check env var `CONTAINER_HOST`, then `XDG_RUNTIME_DIR`, then system paths (existing `podman/client.go` does this correctly)
**Warning signs:** Podman CLI works but HTTP client can't connect

## Code Examples

Verified patterns from official sources and existing codebase:

### Label-Based Container Listing (Podman HTTP API)
```go
// Source: Existing internal/podman/containers.go + official Podman API docs
func (c *Client) ListContainersWithLabels(ctx context.Context, labels map[string]string) ([]string, error) {
    // Build filter JSON: {"label": ["key=value", "key2=value2"]}
    labelFilters := make([]string, 0, len(labels))
    for k, v := range labels {
        labelFilters = append(labelFilters, fmt.Sprintf("%s=%s", k, v))
    }

    filterJSON, _ := json.Marshal(map[string][]string{
        "label": labelFilters,
    })

    path := fmt.Sprintf("/libpod/containers/json?filters=%s", url.QueryEscape(string(filterJSON)))

    var containers []Container
    if err := c.getJSON(ctx, path, &containers); err != nil {
        return nil, err
    }

    names := make([]string, 0, len(containers))
    for _, container := range containers {
        if len(container.Names) > 0 {
            names = append(names, container.Names[0])
        }
    }

    return names, nil
}
```

### Deterministic Container Naming
```go
// Source: Project requirements + Kubernetes naming patterns
func ContainerName(stackID, instanceID string) string {
    // Pattern: devarch-{stack}-{instance}
    // Both IDs are pre-validated as DNS-safe
    return fmt.Sprintf("devarch-%s-%s", stackID, instanceID)
}

func NetworkName(stackID string) string {
    // Pattern: devarch-{stack}-net
    return fmt.Sprintf("devarch-%s-net", stackID)
}

func PodName(stackID string) string {
    // Pattern: devarch-{stack}-pod (Podman only)
    return fmt.Sprintf("devarch-%s-pod", stackID)
}
```

### Routing Operations Through Abstraction
```go
// BEFORE (project/controller.go line 83):
inspectOut, err := exec.Command("podman", "inspect", "--format", "{{.State.Status}}", cName).Output()

// AFTER (using container.Client):
state, err := c.containerClient.GetStatus(cName)
if err != nil {
    return "not-created", nil
}
return state.Status, nil
```

### Migration Schema (Phase 1 Addition)
```sql
-- Migration 013: Stack and instance tables (schema defined in Phase 1, used in Phase 2)
CREATE TABLE stacks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(63) UNIQUE NOT NULL,  -- DNS-safe, validated
    description TEXT,
    network_name VARCHAR(63),           -- devarch-{name}-net
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_stacks_name ON stacks(name);
CREATE INDEX idx_stacks_enabled ON stacks(enabled);

CREATE TABLE service_instances (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    instance_id VARCHAR(63) NOT NULL,   -- DNS-safe, validated, unique per stack
    template_service_id INTEGER REFERENCES services(id),
    container_name VARCHAR(127),        -- devarch-{stack}-{instance}
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(stack_id, instance_id)
);

CREATE INDEX idx_service_instances_stack ON service_instances(stack_id);
CREATE INDEX idx_service_instances_template ON service_instances(template_service_id);
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Shell exec to docker/podman CLI | HTTP API clients (Docker Engine API, Podman REST API) | 2020-2022 | Structured errors, async ops, event streams, no CLI output parsing |
| Manual label management | OCI standard labels (`org.opencontainers.image.*`) | 2021-present | Tooling compatibility, metadata standards |
| Per-container networks | Pod-based shared network namespace (Podman 2.0+) | 2020 | Closer to Kubernetes patterns, simpler service discovery |
| Docker Compose v1 (python) | Docker Compose v2 (Go) / podman-compose | 2022-2023 | Better performance, native integration |
| Mixed-case label keys | Lowercase convention | Ongoing | Prevents query mismatches |

**Deprecated/outdated:**
- `docker-compose` (python): Replaced by `docker compose` (v2) or `podman compose`
- Hardcoded `/var/run/docker.sock`: Use `DOCKER_HOST` or `CONTAINER_HOST` env vars
- `podman/v2` Go bindings: Use `podman/v5` (current as of 2026)

## Open Questions

Things that couldn't be fully resolved:

1. **Docker API client implementation strategy**
   - What we know: Docker Engine API is HTTP-based, Podman socket can emulate it
   - What's unclear: Should we use `github.com/docker/docker` client or write minimal HTTP client like Podman implementation?
   - Recommendation: Start with Podman-only (existing HTTP client works), add Docker client in Phase 2 if needed. Test Docker via Podman socket emulation first.

2. **Validation library trade-off**
   - What we know: go-playground/validator is popular but adds dependency for simple regex checks
   - What's unclear: Does stdlib validation scale to future requirements (config file validation, complex rules)?
   - Recommendation: Use stdlib regex for Phase 1 (4 validation rules). Revisit in Phase 3 when override validation complexity increases.

3. **Pod vs network abstraction timing**
   - What we know: Podman uses pods (shared network namespace), Docker uses bridge networks. Phase 4 implements networking.
   - What's unclear: Should Phase 1 define pod interface stubs or wait until Phase 4?
   - Recommendation: Define stub methods in `container.Client` interface (Phase 1), implement in Phase 4. Prevents interface changes mid-milestone.

4. **Label namespace collision**
   - What we know: `devarch.*` labels are our convention, OCI standard is `org.opencontainers.image.*`
   - What's unclear: Should we use both namespaces or just devarch?
   - Recommendation: Use `devarch.*` for operational labels (stack_id, instance_id). Add OCI labels later for image metadata (version, commit). No collision.

## Sources

### Primary (HIGH confidence)
- [Podman API Documentation](https://docs.podman.io/en/latest/_static/api.html) - Label filtering syntax, container operations
- [Podman ps man page](https://docs.podman.io/en/latest/markdown/podman-ps.1.html) - Label filter behavior, exclusivity rule
- [Kubernetes validation source](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go) - DNS-safe validation patterns
- Existing codebase (`api/internal/container/client.go`, `api/internal/podman/`) - Current implementation patterns

### Secondary (MEDIUM confidence)
- [Kubernetes namespace naming best practices](https://cloudfleet.ai/blog/cloud-native-how-to/2024-11-kubernetes-namespaces-best-practices/) - DNS label rules, 63 char limit, reserved names
- [Container label best practices](https://mihirpopat.medium.com/unlocking-the-power-of-dockerfile-label-a-best-practice-guide-for-clean-and-maintainable-images-a0c408714667) - Lowercase convention, metadata governance
- [golang-migrate guide](https://betterstack.com/community/guides/scaling-go/golang-migrate/) - Migration patterns, versioning strategies
- [Podman Go bindings](https://pkg.go.dev/github.com/containers/podman/v5) - Official API documentation

### Tertiary (LOW confidence)
- WebSearch results on container runtime abstraction patterns (search tool unavailable, couldn't verify 2026 SOTA)
- Community discussions on Docker/Podman compatibility (no single authoritative source found)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib and existing dependencies well-documented, Podman bindings official
- Architecture: MEDIUM - Patterns verified in existing code, but refactor strategy based on analysis not official docs
- Pitfalls: HIGH - Found in existing code (`project/controller.go`, `nginx.go`), documented in Podman/Docker official sources
- Validation rules: HIGH - RFC 1123 + Kubernetes validation is authoritative source
- Label filtering: HIGH - Verified in Podman official docs
- DB schema: MEDIUM - Based on existing migration patterns, not yet implemented

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - stable technology domain, stdlib patterns don't change rapidly)
