# Phase 1: Foundation & Guardrails - Context

**Gathered:** 2026-02-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Establish isolation primitives — identity labels, name validation, and runtime abstraction — that all stack operations depend on. No stack CRUD, no UI, no service assignment. Build the foundation, don't use it yet.

</domain>

<decisions>
## Implementation Decisions

### Identity Scheme
- Stack ID = stack name (immutable). "Rename" = clone to new name + delete old stack.
- Instance ID = user-provided alias with template name + counter fallback (e.g., "main-db" or "mysql-1")
- Label prefix: `devarch.*` (e.g., `devarch.stack_id`, `devarch.instance_id`, `devarch.template_service_id`)
- Container naming: `devarch-{stack}-{instance}` (hyphen-separated)

### Validation Rules
- DNS-safe strict: lowercase alphanumeric + hyphens, no leading/trailing hyphens, max 63 chars
- Reserved name blocklist: `default`, `devarch`, `system`, `none`, `all`
- Prescriptive error messages: state what's wrong AND suggest a fix (e.g., `"My App" is invalid — try: my-app`)
- Stack names globally unique (single-developer local tool, no user scoping)

### Runtime Abstraction
- **Podman is first-class citizen**. Docker fully supported but secondary.
- Auto-detect with Podman preferred. `DEVARCH_RUNTIME=docker|podman` env var override.
- Unified interface with internal branching — callers never branch on runtime
- Isolation primitive is per-stack network name + deterministic container naming, implemented identically on Docker and Podman.
- Podman pods are a future internal optimization (optional): they may group containers, but the external contract remains "per-stack isolated network + labels".
- Pods, quadlets, and rootless concepts modeled in the abstraction interface from Phase 1
- Compose remains the deployment mechanism for now. Quadlets as future export option — interface designed so adding quadlet support is additive, not a rewrite.

### Migration Path
- Clean break: existing services untouched. Phase 1 builds primitives only.
- DB migration included: stack and instance table schema created in Phase 1 so Phase 2 can immediately build CRUD.
- Pod/network interfaces defined now as stubs. Implementations land in Phase 4+.
- Runtime abstraction includes "list unlabeled containers" capability for Phase 2 migration support.

### Existing Code Audit
- All existing direct runtime calls (hardcoded `exec.Command` for podman/docker) must be routed through `container.Client` in Phase 1.
- This satisfies the success criterion: "no hardcoded podman exec.Command."
- Researcher identifies all call sites; planner includes as refactor tasks.

### Claude's Discretion
- Refactor vs extend existing `internal/container/client.go` — decided during research based on code quality
- Exact interface method signatures for pod/network operations
- DB schema design for stack/instance tables
- Stub implementation details for not-yet-implemented operations

</decisions>

<specifics>
## Specific Ideas

- "I want a tool to make application development extremely accessible and powerful, running damn near parity with Podman and leveraging its best features in a nice interface (CLI, UI, and DevArch ecosystem as a whole)"
- Podman features (pods, quadlets, rootless, socket activation) should be first-class concepts in the abstraction, not hidden behind a lowest-common-denominator Docker interface
- The abstraction should model Podman natively — when quadlet support is added later, it should expose what the abstraction already knows, not bolt on something foreign

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation-guardrails*
*Context gathered: 2026-02-03*
