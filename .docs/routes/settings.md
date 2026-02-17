# Settings Page

**Path:** `dashboard/src/routes/settings/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Configuration page with three sections: Container Runtime, Podman Socket, and API Key management. Shows runtime status, socket connectivity, and allows switching runtimes and managing credentials.

## Sections

### RuntimeSection
- Current runtime (podman/docker) status badge
- Grid of available runtimes with status:
  - Installed check
  - Version info
  - Container counts per runtime
  - Switch button (if not current and installed)
- Microservices network info:
  - Network name
  - Missing indicator (yellow highlight, was plain text)
  - Running services count

### SocketSection
- Active Podman socket (rootless/rootful) status badge
- Grid of available sockets with status:
  - Path info
  - Connectivity status
  - Start button (if not active)
- User info: UID and DOCKER_HOST

### ApiKeySection
- Password input for API key
- Save button (changes to green "Saved" on success)
- Logout button (redirects to `/login`)

## Recent Changes
- Missing network indicator now styled in yellow: `<span className="text-yellow-500">missing</span>`

## Dependencies
- `useRuntimeStatus()`, `useSwitchRuntime()` — Runtime management
- `useSocketStatus()`, `useStartSocket()` — Socket management
- `getApiKey()`, `setApiKey()`, `clearApiKey()` — Local storage API key helpers
- `Card`, `Badge`, `Button`, `Input` — UI components

## Route
- Path: `/settings/`
