# Settings Page

**Path:** `dashboard/src/routes/settings/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Configuration page with three sections: Container Runtime, Podman Socket, and API Key management. Shows runtime/socket status, allows switching runtimes, managing sockets, and updating credentials.

## Route
- File route: `/settings/`

## Sections

### 1. RuntimeSection

**Display:**
- Current runtime badge (podman/docker) with success/destructive coloring
- Grid of available runtimes (podman, docker):
  - Installed: yes/no indicator
  - Version info (if installed)
  - Container counts for each runtime
  - CheckCircle (green) if responsive, XCircle (muted) if not

**Actions:**
- Switch button: if not current AND installed
- Switches with options: stop_services, preserve_data, update_config
- Loading state with spinner

**Network Info:**
- Microservices network name
- Missing indicator: yellow text span (was plain text)
- Running services count

### 2. SocketSection

**Display:**
- Active Podman socket status badge (rootless/rootful)
- Grid of available sockets (rootless, rootful):
  - Path: socket_path or "—"
  - Connectivity status
  - CheckCircle (green) if active, XCircle (muted) if not

**Actions:**
- Start button: if not active
- Starts with options: enable_lingering, stop_conflicting
- Loading state with spinner

**User Info:**
- User name with UID
- DOCKER_HOST env var (or "not set")

### 3. ApiKeySection

**Input:**
- Password input field (type="password")
- Placeholder: "Enter API key..."

**Actions:**
- Save button:
  - Normal state: "Save"
  - Saved state: "Saved" with green (success) variant
  - Auto-resets to "Save" after 2 seconds
- Logout button:
  - Calls `clearApiKey()`
  - Redirects to `/login`

**Help Text:**
- "Set the API key for authenticating with the Go API backend. Leave empty if auth is disabled."

## Mutations & Hooks

### Runtime
- `useRuntimeStatus()` — Fetch current runtime status
- `useSwitchRuntime()` — Switch between podman/docker

### Socket
- `useSocketStatus()` — Fetch socket status
- `useStartSocket()` — Start rootless/rootful socket

### API Key
- `getApiKey()` — Retrieve from localStorage
- `setApiKey()` — Save to localStorage
- `clearApiKey()` — Clear from localStorage

## UI Components
- `Card` — Section wrapper
- `CardHeader`, `CardTitle` — Section headers with icons
- `Badge` — Status display
- `Button` — Actions (switch, start, save, logout)
- `Input` — API key input
- `SectionLoader` — Loading placeholder for card

## Loading States
- `SectionLoader` component: "Loading..." spinner while data fetches
- Per-button loading: spinner icon on switch/start buttons when pending

## Recent Changes
- Missing network indicator: yellow text color (better visibility)

## Dependencies
- `useRuntimeStatus()`, `useSwitchRuntime()` — Runtime mutations
- `useSocketStatus()`, `useStartSocket()` — Socket mutations
- `getApiKey()`, `setApiKey()`, `clearApiKey()` — API key storage
- `Card`, `Badge`, `Button`, `Input` — UI components
- Icons: Monitor, Plug, Key, CheckCircle, XCircle, LogOut, Loader2

## Related Pages
- All pages — Use API key from settings for auth
