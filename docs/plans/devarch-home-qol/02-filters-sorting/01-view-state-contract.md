# Phase 1 — View-state and collection contract

Created: 2026-08-29
Purpose: Define one deterministic filtering pipeline before adding controls.

## Goal

Represent collection search, filters, and sort as validated data rather than page-specific DOM state.

## Scope

Pure route-aware view state, allowed facets/sorts, deterministic processing, facet counts, and performance fixtures. No rendered controls or URL persistence.

## Design

Create a route-aware view model:

```text
collection: apps | containers | services
query: normalized string
filters: map of allowlisted keys to string arrays/booleans
sort: allowlisted key
direction: asc | desc
```

Processing order is search → filters → stable sort. Use one explicit `Intl.Collator('en', {numeric: true, sensitivity: 'base'})`; tie-break collator equality by raw stable identity so tests do not depend on host locale. Derive available facet values from current inventory, not hardcoded framework/category lists.

### Allowed facets

- Apps: `kind[]`.
- Containers: `state[]`, `hasPublishedPort`.
- Services: `category[]`. Reserve `hasRuntimeMatch` for plan 06; do not implement it with the current name heuristic.

### Allowed sorts

- Apps: `name`, `kind`.
- Containers: `name`, `state`, `image`.
- Services: `name`, `category`.

## Outputs

- Pure filter/sort functions and per-collection defaults.
- Facet-count derivation that can show counts without mutating the inventory.
- Tests for normalization, invalid values, stable ordering, and combined criteria.

## Acceptance criteria

- Inventory arrays are never sorted or mutated in place.
- Unknown filter/sort keys fall back safely.
- Case, punctuation, and locale behavior are documented and deterministic.
- The contract can accept a future `hasRuntimeMatch` facet only through an additive definition supplied by plan 06 and the shared relationship index; plan 02 does not ship that facet.
- The pure transformation completes within 50ms p95 across 20 runs of the shared 1,000-item synthetic fixture.
- The pipeline remains fast for at least 1,000 items per collection.

## Verification

Pure tests with mixed fixtures; mutation guards; benchmark-style timing with synthetic inventory.

## Non-goals

URL serialization, rendered controls, or local persistence in this phase.