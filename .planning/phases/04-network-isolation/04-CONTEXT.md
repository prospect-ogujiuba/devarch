# Phase 4: Network Isolation - Context

**Gathered:** 2026-02-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Each stack runs on an isolated bridge network with deterministic container naming. No cross-stack contamination. Implementation must be runtime-agnostic: Docker and Podman (rootless and rootful) both get per-stack isolated networks with DNS-based service discovery. This phase builds the network primitives and container naming enforcement — actual container creation happens in Phase 6 (plan/apply).

</domain>

<decisions>
## Implementation Decisions

### Network Lifecycle
- EnsureNetwork is idempotent (create-if-not-exists, no-op if exists)
- Orphan cleanup: apply cleans up for its own stack, devarch doctor cleans up globally
- Network create/destroy timing: Claude's discretion (apply-time creation and destroy-on-stack-delete are the natural fit given stacks are DB records until Phase 6)

### Container Naming Edge Cases
- Reject at creation if combined devarch-{stack}-{instance} exceeds container name length limits — fail early with prescriptive error
- Network name configurable per stack via optional network_name field, defaults to devarch-{stack}-net
- Network name collision strategy (e.g., stack 'app' vs 'app-net'): Claude's discretion — pick safest approach
- Container name enforcement timing (compute+store now vs validate for Phase 6): Claude's discretion

### Runtime Parity (Docker vs Podman)
- Support both rootless and rootful Podman — detect mode and adapt network creation accordingly
- DNS-based service discovery required within stack networks — containers resolve each other by instance name
- Both Docker and Podman support DNS on user-defined bridge networks natively
- Interface design (extend Client vs separate NetworkManager): Claude's discretion based on existing code patterns
- Testing strategy: Claude's discretion

### Dashboard Visibility
- Dedicated Network tab on stack detail page — show network name, status, connected containers, DNS entries
- Stack list page shows network status per stack (badge/column: active/inactive/not created)
- Network icon (globe/network) for network-related indicators — different visual language from container status dots
- Instance card network connectivity indicators: Claude's discretion (may be premature before Phase 6 containers exist)

### Claude's Discretion
- Network create/destroy timing (recommended: create on apply, destroy on stack delete)
- Container name enforcement timing
- Network name collision prevention strategy
- Container client interface design (extend vs separate)
- Testing approach for dual-runtime
- Instance card network indicators (premature before containers exist)

</decisions>

<specifics>
## Specific Ideas

- DNS discovery is critical for Phase 8 wiring to feel natural — instance 'postgres' should be reachable by hostname from other instances in the same stack
- Network icon differentiates from container status dots (green/gray) — separate visual language for network vs container health
- Configurable network_name enables advanced users to control network naming while keeping sensible defaults

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-network-isolation*
*Context gathered: 2026-02-06*
