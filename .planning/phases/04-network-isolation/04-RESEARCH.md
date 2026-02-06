# Phase 4: Network Isolation - Research

**Researched:** 2026-02-06
**Domain:** Container network management (Docker + Podman), Go SDK integration, bridge networks, DNS service discovery
**Confidence:** HIGH

## Summary

Network isolation for containerized stacks follows established patterns: user-defined bridge networks provide automatic DNS resolution, deterministic container naming enables predictable service discovery, and both Docker and Podman support equivalent functionality through their native APIs. The standard approach uses idempotent network creation (create-if-not-exists), container identity labels for orphan cleanup, and runtime-specific client abstractions to handle Docker/Podman differences.

Docker SDK for Go (`github.com/docker/docker/client`) provides `NetworkCreate`, `NetworkInspect`, `NetworkRemove`, and `NetworkList` methods. Podman exposes equivalent functionality via REST API at `/libpod/networks/*` endpoints. Both runtimes enforce DNS-safe naming (63 character limit per RFC 1123), support bridge driver with DNS resolution, and use labels for metadata.

Key insight: user-defined bridge networks are required for DNS resolution—default bridge networks only support IP-based communication. For rootless Podman, netavark (default since v4.0) provides bridge networking with DNS, while legacy slirp4netns isolates containers completely. Both Docker and Podman implement DNS via embedded resolvers (Docker at 127.0.0.11, Podman via aardvark-dns plugin).

**Primary recommendation:** Extend existing `container.Client` with `CreateNetwork`, `RemoveNetwork`, `ListNetworks` methods that wrap CLI commands (consistent with current architecture). Use idempotent creation (inspect then create on 404/not-found), inject identity labels on all operations, and rely on compose YAML's `networks: { external: true }` pattern (already in generator.go).

## Standard Stack

The established libraries/tools for container network management in Go:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/docker/docker/client` | Latest (Go 1.22+) | Docker SDK for Go - network operations | Official Docker client library, provides NetworkCreate/Inspect/Remove/List methods |
| `github.com/docker/docker/api/types/network` | Latest | Network type definitions for Docker API | Defines CreateOptions, CreateResponse, Inspect, Summary types used by client |
| Podman REST API (libpod) | v4.0+ | Podman network HTTP endpoints | `/libpod/networks/create`, `/libpod/networks/{name}/json`, `/libpod/networks/{name}` DELETE |
| `context` (stdlib) | Go 1.22 | Timeout/cancellation for network ops | Standard Go pattern for cancellation propagation and timeout enforcement |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/containerd/errdefs` | Latest | Error type checking for container APIs | Robust error handling (IsAlreadyExists, IsNotFound) instead of string matching |
| CLI commands (`docker network`, `podman network`) | Native | Direct network management via CLI | Fallback when SDK not available or for runtime-agnostic operations (current DevArch pattern) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| CLI exec (current pattern) | Native Go SDKs (Docker client, Podman bindings) | SDK adds dependency complexity but provides structured responses; CLI exec is simpler, version-agnostic, matches existing codebase pattern |
| Docker-only implementation | Dual-runtime (Docker + Podman) | Docker-only is simpler but violates project requirement for runtime parity |

**Installation:**
```bash
# Docker SDK (if migrating from CLI)
go get github.com/docker/docker/client
go get github.com/docker/docker/api/types/network

# Podman bindings (if migrating from CLI)
go get github.com/containers/podman/v5/pkg/bindings
go get github.com/containers/podman/v5/pkg/bindings/network
```

**Note:** DevArch currently uses CLI exec pattern (`execCommand("podman", "network", ...)`) which avoids SDK dependencies and maintains runtime-agnostic code. This research recommends continuing that pattern for consistency.

## Architecture Patterns

### Recommended Extension to Existing Container Client

Current DevArch pattern (from `api/internal/container/client.go`):
- `Client` struct with `runtime RuntimeType`, `composeCmd string`, `useSudo bool`
- `execCommand(args ...string)` wraps CLI invocation with sudo support
- Network stubs return `ErrNotImplemented` (lines 463-473)

**Pattern: Extend Client with network methods that follow existing CLI exec pattern**

```go
// In api/internal/container/client.go
func (c *Client) CreateNetwork(name string, labels map[string]string) error {
	// Check if exists first (idempotent)
	if exists, _ := c.networkExists(name); exists {
		return nil // no-op if exists
	}

	args := []string{"network", "create"}
	for k, v := range labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, name)

	_, err := c.execCommand(args...)
	return err
}

func (c *Client) RemoveNetwork(name string) error {
	args := []string{"network", "rm", name}
	_, err := c.execCommand(args...)
	// Ignore "not found" errors (idempotent deletion)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}
	return err
}

func (c *Client) ListNetworks() ([]string, error) {
	output, err := c.execCommand("network", "ls", "--format", "{{.Name}}")
	if err != nil {
		return nil, err
	}
	return c.parseNamesList(output)
}

func (c *Client) networkExists(name string) (bool, error) {
	output, err := c.execCommand("network", "inspect", name)
	if err != nil {
		return false, nil // Assume doesn't exist
	}
	return output != "", nil
}
```

### Pattern 1: Idempotent Network Lifecycle
**What:** Create-if-not-exists for network creation, no-op on already-exists errors
**When to use:** Every network creation path (stack apply, initialization)
**Example:**
```go
// Source: DevArch context + Docker/Podman best practices
func EnsureNetwork(c *Client, stackName string, labels map[string]string) error {
	networkName := fmt.Sprintf("devarch-%s-net", stackName)

	// Add stack identity labels
	labels["devarch.stack_id"] = stackName
	labels["devarch.managed"] = "true"

	return c.CreateNetwork(networkName, labels)
}
```

### Pattern 2: Deterministic Container Naming
**What:** `devarch-{stack}-{instance}` pattern enforced at compose generation
**When to use:** Generating compose YAML for stack instances
**Example:**
```go
// Source: Existing generator.go pattern + DNS-safe validation
func (g *Generator) GenerateForInstance(stack *Stack, instance *Instance) ([]byte, error) {
	containerName := fmt.Sprintf("devarch-%s-%s", stack.Name, instance.Name)

	// Validate DNS-safe name (63 char limit, RFC 1123)
	if len(containerName) > 63 {
		return nil, fmt.Errorf("container name '%s' exceeds 63 character DNS limit", containerName)
	}

	networkName := stack.NetworkName
	if networkName == "" {
		networkName = fmt.Sprintf("devarch-%s-net", stack.Name)
	}

	compose := generatedCompose{
		Networks: map[string]networkConfig{
			networkName: {External: true},
		},
		Services: map[string]serviceConfig{
			instance.Name: {
				ContainerName: containerName,
				Networks:      []string{networkName},
				Labels: []string{
					fmt.Sprintf("devarch.stack_id=%s", stack.ID),
					fmt.Sprintf("devarch.instance_id=%s", instance.ID),
					fmt.Sprintf("devarch.template_service_id=%s", instance.ServiceID),
				},
			},
		},
	}
	// ... rest of generation
}
```

### Pattern 3: Runtime-Specific Network Detection (Rootless Podman)
**What:** Detect rootless vs rootful mode and adapt network operations
**When to use:** Podman runtime initialization
**Example:**
```go
// Source: Podman security info API + rootless detection patterns
func (c *Client) detectRootlessMode() (bool, error) {
	if c.runtime != RuntimePodman {
		return false, nil // Docker is always rootful or handles via daemon
	}

	// Check via podman info
	output, err := c.execCommand("info", "--format", "{{.Host.Security.Rootless}}")
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(output) == "true", nil
}
```

### Anti-Patterns to Avoid
- **Using default bridge network:** DNS resolution doesn't work on default bridge—always create user-defined networks
- **Assuming network create is idempotent:** Both Docker and Podman return errors on duplicate create—must inspect first or handle error
- **Ignoring 63-char DNS limit:** Docker/Podman enforce RFC 1123 DNS label limits—validate at creation, not at container start
- **Network cleanup without label filtering:** Prune operations without labels can delete unrelated networks—always use `devarch.*` labels for orphan detection
- **Creating networks per instance:** Networks are stack-scoped, not instance-scoped—one network per stack, multiple instances connect to same network

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Network exists checking | Parse `docker network ls` output | `network inspect` with error handling | Inspect returns structured error codes; ls parsing is fragile across versions |
| DNS-safe name validation | Custom regex for container names | Validate length ≤63, alphanumeric+hyphen | RFC 1123 is well-defined; custom validation will miss edge cases |
| Rootless Podman detection | Parse `/etc/subuid`, check socket paths | `podman info --format {{.Host.Security.Rootless}}` | Info command abstracts socket detection, slirp4netns vs netavark differences |
| Network pruning/cleanup | Manual network deletion loops | `docker network prune --filter label=devarch.managed=true` | Prune command handles active endpoint checking, parallel safety |
| Error type checking | `strings.Contains(err.Error(), "exists")` | `errdefs.IsAlreadyExists(err)` (if using SDK) or known error strings | String matching breaks across Docker/Podman/versions; error types are stable |

**Key insight:** Container network management has subtle edge cases (active endpoints blocking removal, rootless networking modes, DNS resolver configuration) that aren't obvious from docs. Use built-in commands and structured APIs rather than custom logic.

## Common Pitfalls

### Pitfall 1: Network Already Exists Error Breaks Stack Apply
**What goes wrong:** Calling `network create` on an existing network returns error, causing stack apply to fail even though network is usable
**Why it happens:** Docker/Podman `network create` is not idempotent by default—returns error if network exists
**How to avoid:** Always check existence before creation (inspect then create) or handle "already exists" error gracefully
**Warning signs:** Stack apply fails on second run with "network already exists" error

### Pitfall 2: Container Name Exceeds DNS Limit
**What goes wrong:** Container name `devarch-my-very-long-stack-name-my-service-instance` exceeds 63 characters, Docker/Podman reject at creation
**Why it happens:** RFC 1123 DNS labels limited to 63 chars; Docker/Podman enforce this for DNS resolution compatibility
**How to avoid:** Validate `len(devarch-{stack}-{instance}) <= 63` at stack/instance creation time, reject early with prescriptive error
**Warning signs:** Container create fails with "invalid name" after stack configuration already saved

### Pitfall 3: Orphan Networks After Stack Deletion
**What goes wrong:** Deleting stack DB record leaves network and containers running; networks accumulate over time
**Why it happens:** Network lifecycle not tied to DB lifecycle—must explicitly remove on stack delete
**How to avoid:** Hook stack delete to remove network; use labels (`devarch.stack_id={id}`) for orphan detection in `devarch doctor`
**Warning signs:** `docker network ls` shows many `devarch-*-net` networks without corresponding stacks in DB

### Pitfall 4: Default Bridge Network Assumption
**What goes wrong:** Services can't resolve each other by hostname—expect DNS but only IP works
**Why it happens:** Default bridge network doesn't provide DNS resolution; only user-defined bridges have embedded DNS
**How to avoid:** Always use user-defined bridge networks (created via `network create`); never rely on default bridge
**Warning signs:** Containers ping by IP but `ping postgres` fails with "unknown host"

### Pitfall 5: Rootless Podman Network Creation Fails
**What goes wrong:** `podman network create` fails with permission error in rootless mode
**Why it happens:** Rootless Podman uses netavark which requires network to be explicitly created; no default network with DNS
**How to avoid:** Detect rootless mode (already supported via socket detection in existing code); ensure network create called before compose up
**Warning signs:** Podman works rootful but fails rootless with "network not found" or "permission denied"

### Pitfall 6: Network Removal Fails with Active Endpoints
**What goes wrong:** `docker network rm` fails with "network has active endpoints" when containers still connected
**Why it happens:** Docker/Podman prevent removing networks with connected containers (safety feature)
**How to avoid:** Stop/remove containers before removing network, or use orphan cleanup that lists containers by label first
**Warning signs:** Stack delete removes DB record but network persists; subsequent stack with same name fails

## Code Examples

Verified patterns from official sources and existing DevArch code:

### Network Creation with Labels (Docker/Podman CLI)
```bash
# Docker
docker network create \
  --driver bridge \
  --label devarch.stack_id=mystack \
  --label devarch.managed=true \
  devarch-mystack-net

# Podman (rootless netavark, default since 4.0)
podman network create \
  --driver bridge \
  --label devarch.stack_id=mystack \
  --label devarch.managed=true \
  devarch-mystack-net
```

### Idempotent Network Creation (Go CLI exec pattern)
```go
// Source: DevArch container/client.go pattern + idempotent network operations
func (c *Client) CreateNetwork(name string, labels map[string]string) error {
	// Check existence (idempotent)
	inspectArgs := []string{"network", "inspect", name}
	if output, err := c.execCommand(inspectArgs...); err == nil && output != "" {
		return nil // Already exists, no-op
	}

	// Create with labels
	args := []string{"network", "create", "--driver", "bridge"}
	for k, v := range labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, name)

	_, err := c.execCommand(args...)
	return err
}
```

### Compose YAML with External Network (Existing Pattern)
```yaml
# Source: api/internal/compose/generator.go (lines 76-81)
networks:
  devarch-mystack-net:
    external: true

services:
  devarch-mystack-postgres:
    container_name: devarch-mystack-postgres
    networks:
      - devarch-mystack-net
    labels:
      - devarch.stack_id=stack-uuid
      - devarch.instance_id=instance-uuid
      - devarch.template_service_id=service-uuid
```

### Network Cleanup by Label Filter
```bash
# List networks managed by DevArch
docker network ls --filter label=devarch.managed=true --format "{{.Name}}"

# Prune unused DevArch networks (safe, checks for active endpoints)
docker network prune --filter label=devarch.managed=true --force

# Remove specific stack's network
docker network rm devarch-mystack-net
```

### Container Name Validation (DNS-safe)
```go
// Source: RFC 1123 + Docker naming restrictions
func ValidateContainerName(stack, instance string) error {
	name := fmt.Sprintf("devarch-%s-%s", stack, instance)

	if len(name) > 63 {
		return fmt.Errorf(
			"container name '%s' exceeds 63 character limit (RFC 1123 DNS); "+
			"stack+instance length must be ≤54 chars (63 - len('devarch--'))",
			name,
		)
	}

	// Docker/Podman allow alphanumeric + hyphen + underscore + dot
	// First char must be alphanumeric
	if !isAlphanumeric(name[0]) {
		return fmt.Errorf("container name must start with alphanumeric character")
	}

	return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CNI (Container Network Interface) for Podman | Netavark (Podman 4.0+, default) | Podman v4.0 (2022) | Netavark provides better rootless support, DNS via aardvark-dns; CNI deprecated in v5.0 |
| Slirp4netns (rootless default) | Pasta (rootless default Podman 5.0+) | Podman v5.0 (2024) | Pasta offers better performance; slirp4netns still supported but not default |
| Docker Compose v1 (docker-compose) | Docker Compose v2 (docker compose) | 2021-2022 | v2 is native CLI plugin, better performance; v1 deprecated |
| `CheckDuplicate` in NetworkCreate | Deprecated (Docker API v1.44+) | Docker Engine 20.10+ | No longer needed; API handles duplicate detection automatically |

**Deprecated/outdated:**
- **CNI for Podman networking:** Replaced by netavark in v4.0, removed in v5.0—use netavark-compatible configurations
- **Slirp4netns as default rootless networking:** Pasta is now default (Podman 5.0+), though slirp4netns still works
- **Docker Compose v1 (`docker-compose`):** Use `docker compose` (v2, builtin to Docker CLI) for better integration
- **`--link` flag for container networking:** Deprecated since Docker 1.13—use user-defined networks instead

## Open Questions

Things that couldn't be fully resolved:

1. **Podman REST API error codes for network already exists**
   - What we know: Docker returns 409 Conflict for duplicate network names; Podman issue #17585 reports Podman returns 500 instead of 409
   - What's unclear: Whether this is fixed in recent Podman versions (5.0+), or if CLI exec pattern should be preferred over REST API
   - Recommendation: Use CLI exec pattern (current DevArch approach) to avoid Docker vs Podman API error code divergence

2. **Network name collision strategy**
   - What we know: Stack names are user-controlled; `devarch-app-net` could collide with stack named "app-net"
   - What's unclear: Should network name include stack ID (UUID) for uniqueness, or rely on name validation to prevent collision?
   - Recommendation: Store optional `network_name` field per stack (decided in CONTEXT.md), default to `devarch-{stack}-net`, validate uniqueness at stack creation

3. **Container name enforcement timing**
   - What we know: Container names computed from stack+instance, must validate DNS-safe (≤63 chars)
   - What's unclear: Should names be computed and stored in DB now (Phase 4), or deferred until Phase 6 (plan/apply)?
   - Recommendation: Validate name length constraints at stack/instance creation (Phase 4) to fail early; compute actual name at compose generation time (Phase 6)

## Sources

### Primary (HIGH confidence)
- [Docker Bridge Network Driver Documentation](https://docs.docker.com/engine/network/drivers/bridge/) - DNS resolution, user-defined vs default bridge
- [Podman Network Create Documentation](https://docs.podman.io/en/latest/markdown/podman-network-create.1.html) - Rootless networking, netavark, DNS via aardvark-dns
- [Docker Go SDK Network Types](https://pkg.go.dev/github.com/docker/docker/api/types/network) - CreateOptions, CreateResponse, IPAM structures
- [Docker Go SDK Client Package](https://pkg.go.dev/github.com/docker/docker/client) - NetworkCreate, NetworkInspect, NetworkRemove, NetworkList signatures
- [Podman Basic Networking Tutorial](https://github.com/containers/podman/blob/main/docs/tutorials/basic_networking.md) - Rootless networking modes (netavark, slirp4netns, pasta)
- DevArch existing code: `api/internal/container/client.go`, `api/internal/compose/generator.go` - Current CLI exec pattern, compose generation

### Secondary (MEDIUM confidence)
- [Docker Compose Networks Documentation](https://docs.docker.com/reference/compose-file/networks/) - External networks, top-level networks element
- [Docker Network Prune Documentation](https://docs.docker.com/reference/cli/docker/network/prune/) - Orphan cleanup, label filtering
- [Podman REST API Reference](https://docs.podman.io/en/latest/_static/api.html) - Network endpoints (libpod)
- [Go Context Package Documentation](https://pkg.go.dev/context) - Timeout/cancellation best practices
- [Docker Container Naming Best Practices](https://devtodevops.com/blog/docker-container-naming-convention/) - 63-char limit, alphanumeric requirements
- [Docker Object Labels Documentation](https://docs.docker.com/engine/manage-resources/labels/) - Label standards, key-value metadata

### Tertiary (LOW confidence, flagged for validation)
- WebSearch: Podman issue #17585 re: 409 vs 500 status code divergence—should verify in Podman 5.0+
- WebSearch: Pasta as default rootless networking (Podman 5.0+)—official docs still reference slirp4netns prominently
- Community discussions: Network idempotency issues in Ansible Podman collections—may not affect direct API usage

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Docker SDK and Podman REST API are official, well-documented, verified via pkg.go.dev and official docs
- Architecture: HIGH - Existing DevArch code provides clear CLI exec pattern; Docker/Podman network commands are stable across versions
- Pitfalls: HIGH - DNS resolution on default bridge, 63-char limit, network exists errors all verified in official documentation
- Rootless Podman: MEDIUM - Netavark vs slirp4netns transition is documented but pasta as new default (5.0+) is less clear; recommend testing

**Research date:** 2026-02-06
**Valid until:** 60 days (stable domain—container networking patterns change slowly; Docker/Podman APIs backward-compatible)
