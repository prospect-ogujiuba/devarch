# Filters and sorting

Created: 2026-08-29
Status: Planned
Depends on: existing Apps, Containers, and Services collection pages

## Outcome

Each collection page provides compact, responsive filters and predictable sorting appropriate to its data, while retaining the existing text search.

## Scope

- Apps: filter by detected framework/type; sort by name or type.
- Containers: filter by state/health and published-port availability; sort by name, state, or image.
- Services: filter by category; sort by name or category. Add runtime-match filtering only in plan 06 after normalized relationships exist.
- Shared normalized view-state and collection pipeline.
- Result counts, active-filter summary, and one-click reset.

## Subphases

1. [View-state and collection contract](01-view-state-contract.md)
2. [Responsive filter controls](02-controls-and-rendering.md)
3. [Accessibility, performance, and verification](03-verification-rollout.md)

## Definition of done

- Search, filters, and sort compose deterministically.
- Controls are usable at 320px and by keyboard.
- Empty states distinguish “inventory empty” from “filters matched nothing.”
- No filter triggers inventory refresh or persists implicitly outside the current page.

## Non-goals

Arbitrary query syntax, saved filter sets, server-side filtering, multi-column tables, pagination, or lifecycle actions.