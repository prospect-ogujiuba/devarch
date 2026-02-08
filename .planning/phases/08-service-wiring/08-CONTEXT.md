# Phase 8: Service Wiring - Context

**Gathered:** 2026-02-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Contract-based service discovery within stacks. Templates declare what they provide (exports) and what they need (imports). DevArch auto-wires unambiguous matches, surfaces ambiguous ones for user resolution. Wires inject env vars into consumers, appear in plan diagnostics, and export to devarch.yml.

Scope: WIRE-01 through WIRE-08, MIGR-03. Graph visualization (AWIR-02), custom naming templates (AWIR-03), and role-based priority (AWIR-01) are v2.

</domain>

<decisions>
## Implementation Decisions

### Contract Model
- Contracts are **template-level only** — defined on service templates, not overridable per instance
- Export contracts: `{name, type, port, protocol}` (e.g., `{name: "database", type: "postgres", port: 5432, protocol: "tcp"}`)
- Import contracts: `{name, type, required, env_vars}` where `type` is **exact match** against exports (e.g., import `type: "postgres"` only matches postgres exports, not mysql)
- Import contracts have a `required` boolean — required imports generate plan warnings if unwired, optional imports do not
- **Forward-looking (Phase 9):** Template-level contracts keep the secret injection path clean — wire-injected secrets follow the same contract→env_var path, and `${SECRET:VAR_NAME}` redaction from Phase 7 naturally extends

### Auto-Wire Behavior
- Auto-wiring runs at **plan generation time** (not instance creation) — user reviews proposed wires in plan output before applying, aligning with Phase 6 plan/apply safety
- Auto-wire is **always on**, no per-stack toggle — users who want explicit-only can disconnect auto-wires
- One matching provider → auto-wire. **Multiple matches → leave unwired** with plan warning (e.g., "ambiguous: 2 postgres providers for laravel.database")
- Users can **disconnect any wire** (auto or explicit) — combined with WIRE-08 (user env overrides win), user can set DB_HOST manually for external services
- Explicit wires always override auto-wires for the same import contract
- **Forward-looking (v2 AWIR-01):** "Multiple match = unwired" is compatible with v2's role-based priority tiebreaker — v2 adds resolution rules, doesn't change the model

### Env Var Injection
- **Import contract defines env var names** — e.g., Laravel's import declares it needs `DB_HOST`, `DB_PORT`, `DB_DATABASE`
- Export contract provides values: hostname via **internal DNS** (`devarch-{stack}-{instance}`), **container port** (not host port) — containers talk on the stack bridge network (Phase 4)
- Merge priority: template env vars → **wire-injected env vars** → instance env var overrides (instance always wins per WIRE-08)
- **Forward-looking (v2 AWIR-03):** Custom env var naming templates extend this pattern — v2 adds a user-configurable mapping layer on top of import-defined names

### Wiring Diagnostics & Plan Integration
- Missing required contracts = **warning, not blocker** — users may deploy consumers before providers are ready
- Wires shown as **dedicated section** in plan output (not mixed into per-instance env var changes)
- Ambiguous contracts shown as warnings with context: "2 postgres providers available for laravel.database — create explicit wire to resolve"
- Orphaned wires (referencing deleted instances) cleaned up automatically
- **Forward-looking (Phase 9 SECR-03):** Separate wiring section in plan makes secret redaction boundary cleaner — wire-injected secrets appear as `***` in the wiring section

### Wiring Persistence & Export
- **All active wires stored in `service_instance_wires` table** — both auto-resolved and explicit
- `source` column distinguishes `auto` vs `explicit` (informational, no behavioral difference once stored)
- Auto-wiring is a resolution step that writes to DB, not a runtime computation — ensures export/import determinism (EXIM-05 round-trip stability)
- devarch.yml export includes all wires (existing `Wires []interface{}` stub gets typed)
- Import recreates wires from devarch.yml
- **Forward-looking:** Phase 7 established "export includes resolved specifics" — wires are another resolved specific

### Dashboard Wiring UX
- **New "Wiring" tab on stack detail page** (alongside Instances, Compose, Deploy)
- Table of all wires: consumer instance → provider instance, contract name, source badge (auto/explicit)
- Unresolved contracts (ambiguous) show badge + dropdown to select provider
- Disconnect button on any wire
- Create explicit wire via "Add wire" action
- **Forward-looking (v2 AWIR-02):** Graph visualization replaces/augments this table — building as a tab makes v2 a content swap, not layout restructure

### Claude's Discretion
- DB schema details for `service_exports`, `service_import_contracts`, `service_instance_wires`
- Auto-wire algorithm implementation
- Wire resolution ordering (alphabetical, creation order, etc. for determinism)
- Exact plan output formatting for wiring section
- Dashboard component structure and styling for wiring tab

</decisions>

<specifics>
## Specific Ideas

- Wiring should feel like "it just works" for simple stacks (1 db + 1 app = auto-wired, no user action needed)
- The plan review is the safety net — user sees proposed wires before they take effect
- Internal DNS names from Phase 4 (`devarch-{stack}-{instance}`) are the wiring primitive

</specifics>

<deferred>
## Deferred Ideas

- AWIR-01: Role-based auto-wire priority (`devarch.role=primary`) — v2
- AWIR-02: Wiring graph visualization in dashboard — v2
- AWIR-03: Custom env var naming templates for wire injection — v2
- Cross-stack wiring — v2 (MULT-02)

</deferred>

---

*Phase: 08-service-wiring*
*Context gathered: 2026-02-08*
