# Phase 2 — Responsive filter controls

Created: 2026-08-29
Purpose: Make the view-state contract discoverable without overwhelming the collection pages.

## Goal

Add compact, route-specific controls that work well on mobile, desktop, keyboard, and touch.

## Scope

Route-specific desktop/mobile controls, active chips/counts, reset behavior, empty states, collection integration, and accessibility.

## UX

- Keep the search field as the first control.
- Desktop: show a compact facet row and sort select below search.
- Mobile: show an **Filters** button with active-count badge and an inline disclosure panel; avoid a full-screen modal unless content proves too dense.
- Use checkboxes for multi-select facets, a switch/checkbox for boolean facets, and a native select for sort.
- Show “N of M” result count and **Clear filters** only when state differs from defaults.
- Preserve focus when applying filters; do not scroll to the page top on each change.
- Reflect active facets as removable chips where space permits.

## Rendering

- Collection pages consume only the shared pipeline output.
- Dashboard preview sections retain search only and do not inherit collection filters.
- Filter options with zero inventory count remain hidden unless currently selected.
- Runtime errors do not affect the plan-02 controls. Plan 06 owns the later runtime-match control, capability state, and unavailable behavior.

## Outputs

- Shared responsive filter bar and mobile disclosure.
- Per-collection facet/sort definitions.
- Active chips, result summary, clear action, and filtered empty states.

## Acceptance criteria

- Every control has a visible label and programmatic name.
- Disclosure state and active count are announced correctly.
- Touch targets are at least 44px.
- Filtering produces no API request.
- Reset restores the exact route defaults.
- Empty result copy names the active constraint and provides reset.

## Verification

Keyboard walkthrough, mobile breakpoints, combined filters, reset behavior, and screen-reader name/state inspection.