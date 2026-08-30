# Phase 2 — Resolver and UI integration

Created: 2026-08-29
Purpose: Replace duplicated frontend heuristics with one normalized relationship source.

## Goal

Resolve safe container-to-app/service links once and consume them consistently across detail pages, rows, and filters.

## Scope

Backend label extraction/resolution, CLI convention fallback, additive response fields, frontend relationship lookup, linked chips, and fixtures.

## Backend

- Extract allowlisted label values while parsing Podman socket or structured CLI JSON responses, then discard all labels. Treat absent, null, non-map, and oversized label fields as no metadata.
- Build unique lookup maps for app names, service canonical IDs, and unique service short names.
- Resolve relationships after projects, services, and containers are discovered so every target is validated.
- Keep container ordering and existing fields unchanged; add normalized `compose` and `relationships` fields only.
- If CLI fallback lacks structured labels, emit convention relationships only; never parse formatted label strings.

## Frontend

- Replace `relatedContainers(name)` heuristic with relationship-ID lookup.
- App and service details show exact/convention related containers and a subtle source label only when confidence is not exact.
- Add the deferred Services runtime-match facet through the plan-02 view-state extension point. If Podman inventory is unavailable, disable the facet with an explanation instead of returning misleading zero matches.
- Container rows may show compact linked app/service chips; cap visible chips and provide accessible text.
- Detail links navigate to existing `/apps/...` or `/services/...` routes.

## Outputs

- Inventory relationship resolver and normalized container fields.
- Frontend ID-based relationship index.
- App/service details, runtime filters, and container linked chips using the same data.

## Acceptance criteria

- Exact metadata wins over naming conventions.
- Ambiguous short service names do not attach to multiple services.
- UI never displays raw label keys or values.
- Relationships remain useful when Podman labels are absent.
- Mobile rows wrap without horizontal overflow.
- No additional API request or polling is introduced.
- The runtime-match filter, detail counts, and container chips all consume the same relationship index.

## Verification

Podman JSON fixtures for Podman Compose, Docker Compose compatibility, rejection of not-yet-enabled explicit DevArch labels, ambiguity, sensitive labels, CLI fallback, and responsive linked-chip rendering.