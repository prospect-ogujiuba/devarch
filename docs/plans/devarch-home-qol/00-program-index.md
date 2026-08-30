# DevArch Home quality-of-life program

Created: 2026-08-29
Status: Planned
Owner surface: `scripts/dashboard/`

## Purpose

Second-pass decision record: [review findings and resolved gaps](second-pass-review.md).

Improve discovery speed and daily ergonomics without turning DevArch Home into a container administration product. Every feature remains read-only or copy-only, preserves manual inventory refresh, and hands operational work to native tools.

## Product guardrails

- No in-dashboard start, stop, restart, delete, exec, log-streaming, image, volume, network, or resource-management execution. The existing copy-only service start command may remain, clearly labeled as state-changing; no additional lifecycle command catalog is added.
- No database, account system, remote access, telemetry, background polling, or hidden filesystem mutation.
- Browser preferences remain local to the browser.
- Server inventory remains secret-safe and explicitly allowlisted.
- Commands are displayed or copied, never executed by the dashboard.
- Existing `/`, `/apps`, `/containers`, `/services`, and detail URLs remain compatible.
- Mobile, keyboard, and screen-reader behavior are acceptance requirements rather than later cleanup.

## Plans

Read [Cross-cutting implementation contracts](cross-cutting-contracts.md) before starting any feature phase; its module, test, security, refresh, performance, accessibility, and rollout decisions are mandatory.

1. [Favorites and recent items](01-favorites-recent/00-plan.md)
2. [Filters and sorting](02-filters-sorting/00-plan.md)
3. [Bookmarkable filters](03-bookmarkable-filters/00-plan.md)
4. [Command palette](04-command-palette/00-plan.md)
5. [Project Git context](05-git-context/00-plan.md)
6. [Runtime relationships](06-runtime-relationships/00-plan.md)
7. [Copy-only native commands](07-copy-only-commands/00-plan.md)

## Recommended implementation order

The numbered order is intentional. Establish only the cross-cutting module/test boundary needed by the first active phase; do not perform a big-bang refactor.

- Favorites establish stable identities, the client preference store, and recency event semantics without duplicating the later action registry.
- Filters establish normalized collection view state. The `hasRuntimeMatch` service facet is deferred until runtime relationships land; plan 02 must not create a second heuristic.
- Bookmarkable filters make that state durable in URLs.
- The command palette consumes the shared action registry, routes, favorites, and recency data.
- Git context and runtime relationships extend the inventory API independently after client foundations stabilize.
- Copy-only commands consume the final app/service/container relationships and shared action registry.

Git context and runtime relationship collectors may be developed independently after plan 04 only after the backend module boundary exists. Their integration into inventory serialization and shared capability fields is sequential; do not make concurrent edits to `server.py`/the inventory assembler. Copy-only commands remain last to avoid duplicating action definitions.

## Shared architecture decisions

### Client state

Use a versioned `localStorage` document only for preferences and recency. URL state remains authoritative for shareable collection views. Runtime inventory remains authoritative from `/api/inventory`. Syntactically valid URL facets remain authoritative even when the current inventory has zero matching values; a transient collector/refresh difference must not erase a bookmark.

### Action registry

Introduce one client-side action registry before the command palette. Cards, detail pages, and the palette must consume the same action definitions so labels, URLs, copy payloads, availability, and analytics-free recency updates do not drift.

### Server inventory

All new backend fields must be additive, deterministic, bounded, and secret-safe. Never return arbitrary Git config, Compose labels, environment variables, container commands, or filesystem contents.

### Refresh behavior

Initial page load and the existing explicit **Refresh** action are the only inventory fetch triggers. Local UI preference changes may rerender immediately but must not fetch inventory. Only one request may be active; stale responses cannot replace newer state, and a failed refresh preserves the last successful inventory with a visible stale/error status.

## Program-level definition of done

- The cross-cutting contracts and all seven plans meet their acceptance criteria.
- Existing dashboard tests remain green and each feature adds focused tests.
- Tailwind builds without warnings and generated CSS is updated.
- Direct HTTPS smoke tests cover collection and detail routes.
- Keyboard-only navigation covers menus, filters, palette, pins, copy actions, and dismissal.
- Mobile layouts are verified at 320, 375, 768, and desktop widths.
- No network request repeats without an explicit page load or user refresh.
- `scripts/dashboard/README.md` documents preference storage, URL parameters, keyboard shortcuts, and copy-only safety.

## Explicitly deferred

Container lifecycle controls, embedded logs, shell access, runtime metrics, notifications, multi-host support, user-defined scripts, service installation wizards, Compose editing, and Portainer-style administration remain out of scope for the entire program.
