# dashboard-pages-responsive-nav

Created: 2026-08-29T17:19:43.155Z
Purpose: Define the dedicated routes and responsive navigation implementation slice.

# Dashboard pages and responsive navigation

Created: 2026-08-29
Purpose: Turn the single inventory surface into navigable pages without adding a frontend framework.

## Goal
Provide dedicated project and service detail URLs, useful collection pages, and a mobile-friendly navigation shell.

## Scope
- Client-side route rendering for `/`, `/apps`, `/apps/<name>`, `/containers`, `/services`, and `/services/<category>/<name>`; legacy `/projects` routes may remain aliases.
- Server fallback to the tracked HTML shell for supported browser routes.
- Responsive desktop navigation and accessible mobile menu.
- Link existing project and service cards to their detail pages.
- Preserve initial inventory load plus explicit manual refresh only.

## Outputs
- Updated dashboard HTML, JavaScript, Tailwind asset, server route behavior, tests, and documentation.

## Acceptance criteria
- Direct requests to supported routes return the HTML shell.
- Detail pages show identity, location, status/action context, and native handoff actions.
- Mobile menu exposes all collection routes and closes after navigation.
- Search and empty states remain useful on collection pages.
- No SPA framework or automatic polling is introduced.

## Verification
Focused route tests, complete dashboard test suite, JavaScript syntax, Tailwind build, systemd restart, and HTTPS smoke requests for collection/detail routes.

## Non-goals
Container detail pages, lifecycle controls, server-side templates, or a framework migration.
