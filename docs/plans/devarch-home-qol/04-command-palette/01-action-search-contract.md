# Phase 1 — Action registry and search contract

Created: 2026-08-29
Purpose: Prevent action labels and safety rules from diverging between cards, details, and the palette.

## Goal

Represent every safe dashboard action as validated data with one execution boundary.

## Scope

Safe action descriptors, registry generation, execution constraints, query normalization, ranking, result bounds, and pure tests. No modal UI.

## Action model

Each action includes:

```text
id, entityId, group, label, keywords, iconKey,
kind: navigate | open | editor | copy,
payload, availability, paletteEligible, recencyPolicy
```

Rules:

- `navigate` accepts only same-origin dashboard paths.
- `open` accepts only parsed `http:`/`https:` URLs derived from inventory fields and always uses `noopener,noreferrer`.
- `editor` accepts only the existing encoded `vscode://file` workspace action, is never palette-eligible, and remains visible on app details. No arbitrary custom scheme is accepted.
- `copy` payloads come from explicit action builders, never arbitrary labels or HTML.
- Registry generation is pure and inventory-driven.
- Cards/detail pages request actions by ID instead of rebuilding payloads.
- Plan 04 registers only actions already visible when it lands. Plan 07 extends the registry with copy-only native commands; plan 04 must not predesign or duplicate those command payloads.

## Search and ranking

Normalize with Unicode NFKD, remove combining marks for matching only, and lowercase with an explicit English locale while retaining original labels for display. Rank in order: exact name, name prefix, token prefix, substring, keyword substring. Apply small favorite and recent boosts only after textual relevance so unrelated favorites never outrank a strong match. Tie-break with the shared explicit collator, group priority, label, then action ID.

## Outputs

- Action registry builders for route navigation, app/service/container actions, and global routes.
- Pure query normalization and ranking functions.
- Rank the full bounded registry but render at most the top 50 results. Do not virtualize the palette in this program; a 50-row cap keeps keyboard position, `aria-activedescendant`, and result counts predictable.

## Acceptance criteria

- Duplicate action IDs fail tests.
- Invalid URLs/paths/actions are omitted rather than executable, and non-palette-eligible actions never enter palette results.
- Empty query returns a bounded mix of favorites, recent items, and global navigation.
- Search remains deterministic with equivalent inventory.
- Plan 04 adds no lifecycle verbs/actions. Plan 07 may later register only the pre-existing copied service-start command as a `copy` action classified `lifecycle`; it remains palette-ineligible and visually state-changing everywhere it is shown.

## Verification

Registry snapshots by entity type, malicious/invalid payload fixtures, ranking table tests, and 1,000-entity timing.