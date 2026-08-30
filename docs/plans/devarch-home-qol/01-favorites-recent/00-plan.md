# Favorites and recent items

Created: 2026-08-29
Status: Planned
Depends on: existing route/detail pages

## Outcome

Users can pin frequently used apps and services and quickly return to the current detail page for entities they recently viewed or used to open sites and ports. All state stays in the current browser; recent entries never replay an external action automatically.

## Scope

- Favorite apps and catalog services from cards and detail pages.
- Recent item recording for detail-page navigation and explicit external-open actions.
- A compact dashboard section for favorites and recent items.
- A versioned, corruption-tolerant local preference store.
- Missing inventory items degrade to removable stale entries rather than broken pages.

## Subphases

1. [Preference and identity contract](01-preference-contract.md)
2. [Pinning and recent-item experience](02-ui-and-behavior.md)
3. [Hardening, migration, and verification](03-verification-rollout.md)

## Data decisions

Stable identities are `app:<name>`, `service:<category>/<name>`, and `container:<immutable-id>`. Favorites support apps and services only in the first release because container IDs are ephemeral. Recent records may include containers but are reconciled against each inventory refresh. Their action kind is attribution for ordering/display; activating a recent entry navigates to the entity detail and does not reopen a stored site or port.

## Definition of done

- Pins survive reloads without a server write.
- Recent ordering is deterministic and bounded.
- Storage denial/corruption never prevents dashboard rendering.
- Favorites and recents are usable with keyboard and touch.
- Removing a workspace or service does not create an unrecoverable entry.
- No inventory polling or server preference endpoint is added.

## Non-goals

Cloud sync, cross-browser sync, favorite containers, folders/tags, drag-and-drop ordering, activity telemetry, and server-side user profiles.