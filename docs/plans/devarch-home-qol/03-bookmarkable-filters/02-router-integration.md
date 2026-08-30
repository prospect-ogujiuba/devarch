# Phase 2 — Router and controls integration

Created: 2026-08-29
Purpose: Make URL state authoritative without causing navigation or focus churn.

## Goal

Connect collection controls and history so direct links, edits, Back, and Forward all restore the same view.

## Scope

Direct-load parsing, semantic history updates, Back/Forward restoration, copied view links, aliases, refresh reapplication, and router tests.

## Behavior

- On direct load or route navigation, parse and validate query syntax immediately against the route contract. Inventory later supplies labels/counts but must not invalidate a syntactically valid selected facet.
- On text input, update results immediately and use `history.replaceState` to avoid one history entry per keystroke.
- On discrete filter, sort, clear, or chip actions, use `history.pushState` so Back reverses meaningful changes.
- On `popstate`, restore controls and results without scrolling to top or refetching inventory.
- Navigating to another collection starts from that route's URL/default state; returning restores browser history naturally.
- Provide **Copy view link** only when state is non-default; use the existing clipboard/toast behavior and explain that search text becomes part of browser history and the copied URL.

## Outputs

- Router entry point that owns URL parsing and view-state creation.
- Control updates that dispatch semantic changes rather than manipulating history independently.
- Page title/result summary that reflects active query where useful but avoids noisy title changes per keystroke.

## Acceptance criteria

- Direct links render correctly after initial inventory load.
- Back/Forward restores checkboxes, selects, chips, search text, count, and results.
- Search typing does not flood browser history.
- Route aliases such as `/projects` canonicalize to `/apps` while retaining valid query state.
- Refresh preserves the current URL state and reapplies it to new inventory, including selected values with zero matches.

## Verification

Direct-load tests, navigation sequences, refresh with changed facets, aliases, copied links, keyboard focus, and no-fetch assertions.