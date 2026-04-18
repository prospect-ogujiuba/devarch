# DevArch V2 Web UI

Workspace-first Phase 5 web app for DevArch V2.

## Scope

- top-level navigation stays limited to Workspaces, Catalog, Activity, and Settings
- workspace detail keeps the primary define -> plan -> apply -> observe flow in one place
- API integration stays thin against the Phase 4 `/api` surface
- V1 `dashboard/` is mined only for reusable interaction patterns, not route sprawl

## Commands

```bash
npm install
npm run test
npm run build
npm run dev
```

The dev server proxies `/api` requests to `http://127.0.0.1:7777` for a local `devarchd` instance.

## Current notes

- raw manifest editing is implemented as a validated local draft because the current Phase 4 API does not expose a manifest write endpoint yet
- workspace activity uses the workspace-scoped SSE endpoint exposed by Phase 4
