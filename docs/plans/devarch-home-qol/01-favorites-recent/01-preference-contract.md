# Phase 1 — Preference and identity contract

Created: 2026-08-29
Purpose: Establish durable client-only storage before adding visible controls.

## Goal

Create one tested preference module that owns schema versioning, stable identities, safe reads/writes, and reconciliation with inventory.

## Scope

Versioned browser storage, identity validation, bounded recency, reconciliation, and pure tests. No visible controls or server state.

## Design

Use one key, `devarch.home.preferences.v1`, containing:

```json
{
  "version": 1,
  "favorites": ["app:storefront", "service:database/postgres"],
  "recent": [{"id": "app:storefront", "action": "detail", "at": "ISO-8601"}]
}
```

Rules:

- Use the program identity formats exactly: `app:<name>`, `service:<category>/<name>`, and `container:<immutable-id>`. Build them only from validated inventory fields; reject empty/control-character components and duplicate computed identities rather than guessing. Container identities are stable only for the current runtime object and are never favoriteable.
- Deduplicate favorites while preserving explicit order.
- Bound recent records to 20 and deduplicate by identity plus the allowlisted action kind (`detail`, `open-site`, or `open-port`). The action records attribution only: selecting a recent item navigates to its current entity detail and never silently replays an external URL, port open, copy, or command.
- Accept only known identity prefixes, exact component counts, allowlisted action kinds, and valid strings.
- Treat malformed JSON, unavailable storage, quota errors, and future versions as an empty/default state with a non-blocking console warning; the UI remains fully usable for the current tab.
- Synchronize valid external-tab changes through the browser `storage` event using load-validate-replace semantics. Do not echo writes or merge stale in-memory documents.
- Reconcile identities against current inventory for rendering; do not rewrite merely because a runtime item is temporarily absent.

## Outputs

- Extracted client preference/store functions in `scripts/dashboard/static/dashboard.js` or the mandatory `scripts/dashboard/static/core/preferences.mjs` boundary.
- Stable identity helpers for apps, services, and containers.
- Unit-testable pure serialization, validation, migration, and recency functions.

## Acceptance criteria

- Default, valid, corrupt, partially valid, and unavailable-storage cases are deterministic.
- Unknown fields are ignored.
- No absolute filesystem paths, commands, or inventory payloads are persisted.
- Storage writes occur only after an explicit pin/unpin or qualifying navigation/open action.
- Two open tabs converge after the latest completed valid write; a key removal is treated as an explicit default document, while malformed or unsupported external-tab data is ignored without clearing the current valid state.

## Verification

Pure-function tests; manual storage denial simulation; reload test; inspect stored JSON for bounded, non-sensitive content.

## Non-goals

Visible pin controls or dashboard sections in this phase.