# Phase 5: Compose Generation - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Stack compose generator produces single docker-compose YAML with all instances from effective configs, replacing per-service generation for stack workflows. Config files materialized to disk. Existing single-service compose generation preserved (backward compatible). Plan/apply workflow is Phase 6 — this phase generates YAML, not applies it.

</domain>

<decisions>
## Implementation Decisions

### Output structure
- Service keys: instance ID as key (e.g., `db-01`, `redis-cache`) — `container_name: devarch-{stack}-{instance}` set explicitly
- Network: declared as `external: true` (Phase 4 pre-creates networks via container.Client)
- depends_on: condition-based (`service_healthy`) when healthcheck exists on target, simple list otherwise
- Disabled instances: excluded entirely from YAML — depends_on references to disabled instances stripped
- Compose endpoint returns structured JSON `{ yaml: "...", warnings: [...] }` (not raw YAML text) so warnings about stripped dependencies, disabled references, port conflicts are surfaced

### Config file materialization
- Directory layout: `compose/stacks/{stack}/{instance}/` (separate from legacy `compose/{category}/{service}/`)
- Stale cleanup: delete `compose/stacks/{stack}/` before materializing (files are always generated from DB, never hand-edited)
- Atomicity: write to temp directory, then rename to final path (prevents half-written state on failure)

### Backward compatibility
- Existing `GET /api/v1/services/{name}/compose` preserved unchanged
- New endpoint: `GET /api/v1/stacks/{name}/compose` for stack-scoped generation
- Generator architecture: new stack method reusing existing serviceConfig struct and YAML marshaling internals
- Dual-mode: standalone services use per-service generation, stack instances use stack generation — both coexist

### Dashboard integration
- Compose tab on stack detail page — read-only CodeMirror editor with YAML syntax highlighting
- Download button → downloads `docker-compose.yml` file
- Validation warnings shown below YAML preview (missing deps, disabled instance references, port conflicts)
- Diff view deferred to Phase 6 (plan/apply scope)

### Claude's Discretion
- Exact serviceConfig struct reuse vs new stack-specific struct
- Warning severity levels and formatting
- CodeMirror configuration and theme for YAML preview
- Whether to create a separate `StackGenerator` type or add methods to existing `Generator`

</decisions>

<specifics>
## Specific Ideas

- Compose preview follows existing pattern from service detail page (`useServiceCompose` hook, `<pre>` block with YAML) — upgrade to CodeMirror for stack version
- Warning about disabled dependencies was driven by the "exclude disabled instances" decision — the two are coupled
- Temp-dir-then-rename pattern for materialization anticipates Phase 6 (plan/apply) needing consistent config state during compose up

</specifics>

<deferred>
## Deferred Ideas

- Diff view between generated YAML and running state — Phase 6 (plan/apply workflow)
- Compose validation against Docker/Podman spec — could be future enhancement
- Multi-stack compose (generating YAML across multiple stacks) — not in current scope

</deferred>

---

*Phase: 05-compose-generation*
*Context gathered: 2026-02-07*
