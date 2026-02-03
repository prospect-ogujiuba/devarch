# Phase 2: Stack CRUD - Context

**Gathered:** 2026-02-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can create, list, view, edit, delete, enable/disable, and clone stacks via API and dashboard. Stack name is immutable ID; "rename" is clone + soft-delete. All CRUD operations have corresponding dashboard UI. Instances and network isolation are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Dashboard List View
- Card grid + table view with toggle — both share sorting, filtering, pagination, and search context
- Card content is rich: name, description, status counts, instance names, last activity, quick actions (enable/disable/delete)
- Table rows are clickable (navigate to detail) AND have action buttons for quick ops
- Status shown as detailed counts with color coding (e.g., "3/5 running")

### Stack Detail Page
- Full-blown details — show everything the DB stores and everything available from runtime
- All stack metadata, all instances with their statuses, network info when available

### Empty State
- CTA + brief explanation of what stacks are + prominent "Create your first stack" button

### Disable / Enable
- Disable: stop containers with confirmation dialog showing what will be affected ("3 containers will be stopped: mysql, redis, laravel")
- Re-enable: prompt "Start containers now?" — user chooses whether to start immediately

### Delete (Soft Delete)
- Confirm dialog → soft delete (move to trash). Containers stopped on soft delete
- Confirmation shows cascade summary — blast radius before committing (N instances, N containers, network)
- From trash: permanently delete or restore

### Clone
- Copies everything: instances, overrides, description. New name required
- Records only — no containers started. User applies/starts when ready

### Rename
- Explicit "Rename" action in UI — clone + soft-delete old stack behind the scenes
- If clone fails, original is untouched

### Action Placement
- Clone, rename, delete, enable/disable available on stack detail page header AND list context menu

### API Design
- List returns all stacks (client-side sort/filter/search — local dev tool, few stacks)
- Status summary per stack includes counts + instance summaries (name, status)
- Validation errors are prescriptive with fix suggestions (consistent with Phase 1 slugify pattern)
- Stack events extend existing `/ws/status` WebSocket for real-time dashboard updates

### Claude's Discretion
- Create stack flow UX (modal vs page vs inline)
- List page layout (sidebar/preview panel vs separate detail route)
- Trash UX (separate view vs filtered in list)
- Trash retention policy (auto-purge timing vs manual only)
- Name reuse policy after soft delete
- Restore behavior (records only vs prompt to recreate containers)

</decisions>

<specifics>
## Specific Ideas

- Both card and table views must maintain identical sorting, filtering, pagination, and search state when toggling between them
- Disable confirmation should enumerate affected containers by name, not just count
- Rename should feel like a first-class action, not "clone then delete"

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-stack-crud*
*Context gathered: 2026-02-03*
