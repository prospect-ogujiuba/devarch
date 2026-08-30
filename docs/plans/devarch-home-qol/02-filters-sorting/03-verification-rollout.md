# Phase 3 — Accessibility, performance, and verification

Created: 2026-08-29
Purpose: Prove collection controls remain fast and comprehensible across inventory sizes and devices.

## Goal

Prove filtering and sorting are correct, accessible, responsive, stable, and fast across realistic inventory sizes.

## Scope

Pure pipeline coverage, DOM interactions, accessibility, synthetic performance, cross-route isolation, documentation, and staged rollout.

## Work

- Add tests for each allowed facet, sort direction, tie-break, combined search/filter/sort, invalid state, and empty inventory.
- Add DOM-level/manual checks for disclosure semantics, focus retention, chip removal, reset, and result-count announcements.
- Verify 1,000 synthetic entries meet the cross-cutting ≤50ms pure-transform and ≤100ms p95 rerender budgets across 20 runs; if rendering dominates, optimize DOM batching before considering pagination.
- Confirm filter changes do not write localStorage, fetch inventory, or alter other collection routes.
- Document collection defaults and supported facets.

## Outputs

- Complete filter/sort regression suite and synthetic benchmark.
- Keyboard/mobile accessibility checklist.
- Collection facet/default documentation.

## Acceptance criteria

- Pure view-model tests cover all branches.
- Existing route and detail-page tests remain green.
- At 320px controls do not overflow horizontally.
- Sort results remain stable across refreshes with equivalent inventory.
- No background network activity occurs.
- Tailwind output contains every dynamically used class through static source discovery.

## Verification

Run pure and integration tests, 1,000-item timing, 320px overflow checks, keyboard/screen-reader inspection, no-network assertions, and Tailwind build.

## Rollout

Ship one collection at a time behind the shared contract: Services first (category is clearest), Apps second, Containers last. Each page must pass its checks before the next adopts the controls.

## Exit evidence

Focused tests, synthetic timing, screenshots at required widths, keyboard checklist, and README updates.