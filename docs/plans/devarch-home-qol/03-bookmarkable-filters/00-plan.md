# Bookmarkable filters

Created: 2026-08-29
Status: Planned
Depends on: filters and sorting

## Outcome

Collection search, filters, and sort can be bookmarked, shared, and navigated with browser Back/Forward without changing inventory behavior.

## Scope

- Canonical query-string schema for collection view state.
- Bidirectional URL ↔ view-model synchronization.
- Back/Forward restoration and legacy URL compatibility.
- Optional copy-link affordance when non-default state is active.

## Subphases

1. [Canonical URL schema](01-url-contract.md)
2. [Router and controls integration](02-router-integration.md)
3. [Compatibility and verification](03-verification-rollout.md)

## Definition of done

- Reloading a filtered URL restores the same results.
- Invalid parameters are ignored and canonicalized safely.
- Back/Forward restores controls, focus-safe rendering, and results.
- Default state produces clean collection URLs.
- Query changes never fetch inventory.

## Non-goals

Server-side search, saved named views, compressed/opaque state, URL state on detail pages, or persistence of mobile disclosure state.