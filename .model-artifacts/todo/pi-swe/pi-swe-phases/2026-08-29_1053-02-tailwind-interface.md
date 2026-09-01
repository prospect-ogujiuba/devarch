# 02-tailwind-interface

Created: 2026-08-29T10:53:32.811Z
Purpose: Specify the browser interface and launch documentation slice.

# Tailwind interface

Created: 2026-03-04
Purpose: Make DevArch inventory fast to search and open without becoming a management console.

## Goal
Deliver one polished Tailwind page for project, container, and service discovery.

## Scope
- Tailwind CLI isolated under `scripts/dashboard/`.
- Plain browser JavaScript for rendering, filtering, copying, and explicit refresh.
- Project `.test` links, HTTP container-port links, editor/folder handoff where supported, and native Compose command copying.
- Start/build helper and README documentation.

## Outputs
- Static HTML, JavaScript, Tailwind input and generated CSS.
- Dashboard package manifest and helper scripts.
- `scripts/dashboard/README.md` and root README link.

## Acceptance criteria
- No timer or automatic polling exists.
- Global search filters all three sections.
- Refresh is explicit and reports the last refresh time.
- Empty and Podman-unavailable states remain useful.
- UI works from the localhost Python server.

## Verification
Build Tailwind CSS, run tests, syntax-check Python and shell, and perform a representative HTTP smoke test.

## Non-goals
SPA framework, database, user accounts, container lifecycle actions, terminal, metrics, or remote runtime access.
