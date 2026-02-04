# Phase 3: Service Instances - Context

**Gathered:** 2026-02-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can create service instances from templates within stacks, with full copy-on-write overrides on ports, volumes, env vars, labels, domains, healthchecks, and config files. Effective config = template + overrides (overrides win). Creating/starting containers is NOT this phase (that's plan/apply in Phase 6). This phase builds the data model, API, and dashboard UI for managing instance definitions and their overrides.

</domain>

<decisions>
## Implementation Decisions

### Instance Creation Flow
- Template selection via searchable grid dialog with icons/names
- Instance name auto-generated from template name, shown in dialog for user customization before confirming
- Lightweight creation: pick template + name + optional description, then navigate to instance detail for overrides
- Multiple instances of same template allowed (no restrictions)
- "Add instance" action lives on stack detail page only
- Instances displayed in creation order (oldest first)
- Instance card/row shows: name, template, status, key ports, override count (rich info density)
- Empty stack state: CTA-focused ("Add your first service" button + brief explanation)
- Template catalog shows all templates, badges existing ones ("1 instance")
- Optional description field on creation dialog
- New instances start with zero overrides (pure copy-on-write: template IS the default)

### Override Editing UX
- Tabbed layout on instance detail page (matching existing service detail pattern)
- Template value shown as muted/placeholder text alongside editable override field
- "Add" button per section for new overrides (e.g., new env var not in template)
- Explicit save button (changes staged until user confirms, not auto-save)
- Per-field reset icon + global "Reset all overrides" option
- Config file overrides use CodeMirror editor with syntax highlighting
- Overridden fields distinguished by subtle colored left border
- Port overrides show full host:container mapping (user controls both)
- Validation on both layers: client-side with Zod for instant feedback, server-side as safety net
- Volume overrides support both bind mounts and named volumes
- Dependency overrides NOT user-editable (template-only, wiring is Phase 8)
- Healthcheck overrides use individual field editing (command, interval, timeout, retries, start_period)
- devarch.* labels are read-only (system-managed), custom labels editable
- Domain overrides are domain list only (no path-based routing)

### Effective Config Preview
- Flat merged view showing final effective values (what will actually run)
- Dedicated "Effective Config" tab on instance detail page
- Structured UI format (read-only cards matching override editor layout)
- Overridden values marked with subtle indicator (colored border, same visual language)
- REST API endpoint: GET /stacks/{name}/instances/{instance}/effective-config
- Updates after save only (not live preview — different tabs)
- Copy button: YAML default with JSON toggle

### Instance Lifecycle Actions
- Soft delete (trash) matching stack delete pattern — recoverable
- Duplicate within same stack: copies all overrides, auto-increment name with rename prompt
- Always confirm removal with blast radius preview (instance name, template, override count)
- Per-instance enable/disable toggle (disabled instances excluded from compose generation)
- Direct DB rename (no clone+soft-delete needed — instances are just DB records at this phase)
- No move-between-stacks action (recreate if needed, or use export/import in Phase 7)
- Breadcrumb navigation: Stacks > {stack-name} > {instance-name}

### Claude's Discretion
- Loading skeleton design and exact tab order
- Spacing, typography, and exact component sizing
- Error state handling and messaging
- Progress/loading indicators during save operations
- Default active tab on instance detail page

</decisions>

<specifics>
## Specific Ideas

- Instance cards should be "rich" — show enough at a glance that users rarely need to click in just to check status
- Same visual language as stack detail page (colored borders for overrides = colored indicators for status)
- Follow existing service detail tab pattern (Info, Environment, Compose equivalent tabs)
- Existing editable components (healthcheck, ports, volumes, etc.) should be reusable or adapted for override editing

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-service-instances*
*Context gathered: 2026-02-03*
