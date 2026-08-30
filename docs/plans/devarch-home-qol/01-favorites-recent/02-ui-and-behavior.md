# Phase 2 — Pinning and recent-item experience

Created: 2026-08-29
Purpose: Expose the preference contract through small, consistent discovery controls.

## Goal

Add favorite toggles and useful dashboard shortcuts without crowding cards or detail pages.

## Scope

Favorite controls, dashboard favorite/recent sections, recency event hooks, stale-item behavior, responsive layout, and interaction checks.

## UX

- Add a star toggle to app/service cards and detail headers with `aria-pressed` and explicit labels such as “Add Postgres to favorites.”
- Emit recency through a small typed event/function (`recordRecent(identity, action)`) rather than embedding storage writes in card handlers. The later action registry must call this boundary instead of replacing or duplicating it.
- Add **Favorites** and **Recently opened** sections near the top of the dashboard, below summary stats.
- Favorites render in stored order; recent items render newest first.
- Record recency on app/service/container detail navigation and explicit “Open site”/“Open port” actions, not on passive rendering or search. Recent rows may label the attributed action, but activating one always navigates to the entity's current detail; it never replays the prior external action.
- A stale favorite shows its identity and a remove action; stale recents are hidden.
- Empty states explain how to add a favorite without occupying excessive vertical space.

## Responsive behavior

- Dashboard shortcuts use a horizontally compact list on mobile and a grid on larger screens.
- Pin controls have at least a 44px touch target without making the icon visually oversized.
- Toast feedback supplements but does not replace the persistent pressed state.

## Outputs

- Reusable favorite button and compact recent-item renderer.
- Shared recency recording hook used by routed and external actions.
- Dashboard sections and detail/card integrations.

## Acceptance criteria

- Pin state updates immediately and survives reload.
- The same item cannot appear twice in either section.
- Keyboard activation and screen-reader state are correct.
- External-open failures do not prevent recording the user's explicit action.
- Search does not silently alter favorite ordering.
- Pin/recent changes made in a second tab update this tab without an inventory request or focus loss.

## Verification

Keyboard and touch checks; app and service pin/unpin flows; recent detail/site/port actions; reload and mobile layout checks.