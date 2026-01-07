# Architecture

**Analysis Date:** 2026-01-07

## Pattern Overview

**Overall:** Distributed Microservices Orchestration Platform with Hub-and-Spoke Pattern

**Key Characteristics:**
- Central management container (`devarch`) orchestrates all services
- Dual interface: CLI (`service-manager.sh`) + Web Dashboard (React)
- Declarative service definitions via Docker Compose YAML files
- Stateless API layer - state derived from container inspection
- Polyglot runtime support (PHP, Node, Python, Go, .NET, Rust)

## Layers

**Presentation Layer:**
- Purpose: User interface for service management
- Contains: React components, pages, layouts
- Location: `apps/dashboard/src/`
- Depends on: Service Integration layer (hooks)
- Used by: End users via browser

**Service Integration Layer:**
- Purpose: Data fetching, state management, business logic
- Contains: 13 custom React hooks, 2 Context providers
- Location: `apps/dashboard/src/hooks/`, `apps/dashboard/src/contexts/`
- Depends on: API layer via HTTP fetch
- Used by: Presentation layer components

**API Layer:**
- Purpose: REST endpoints for container/service operations
- Contains: PHP router, 16 endpoint handlers, 8 business logic libraries
- Location: `config/devarch/api/`
- Depends on: Orchestration layer (shell commands, Podman CLI)
- Used by: Dashboard hooks, external clients

**Orchestration Layer:**
- Purpose: Service lifecycle management, container control
- Contains: Shell scripts, service configuration, startup logic
- Location: `scripts/service-manager.sh`, `scripts/config.sh`
- Depends on: Docker Compose files, Podman/Docker runtime
- Used by: API layer, CLI users

**Infrastructure Layer:**
- Purpose: Container and service definitions
- Contains: 166 Docker Compose files across 24 categories
- Location: `compose/{category}/{service}.yml`
- Depends on: Docker/Podman runtime
- Used by: Orchestration layer

## Data Flow

**Container Listing Request:**

1. User navigates to Containers page
2. `useContainers()` hook calls `fetch('/api/containers')`
3. PHP router (`index.php`) routes to `endpoints/containers.php`
4. Endpoint calls `lib/containers.php` → `getAllContainers()`
5. Library executes `podman ps --format json`, parses output
6. Results cached (30s TTL), filtered/sorted per request params
7. JSON response: `{ success: true, data: { containers: [], stats: {} } }`
8. React updates state, renders container grid

**Service Control (Start):**

1. User clicks "Start" on service
2. `useServiceManager()` hook POSTs to `/api/services/start`
3. PHP endpoint calls `lib/services.php` → `startService()`
4. Library executes `service-manager.sh up {service}`
5. Script loads `compose/{category}/{service}.yml`
6. Podman-compose brings up containers
7. Response returned, dashboard refreshes container list

**State Management:**
- No database - all state from container inspection
- React useState for local component state
- React Context for global state (theme, toasts)
- localStorage for UI preferences (sidebar state, active tab)
- PHP static cache for Podman query results (30-60s TTL)

## Key Abstractions

**Service Categories:**
- Purpose: Group related services (database, backend, gateway, etc.)
- Location: `scripts/config.sh` lines 52-76
- Examples: 22 categories including database, backend, gateway, proxy, messaging, analytics, ci
- Pattern: Associative array mapping category → space-separated service names

**Service Startup Order:**
- Purpose: Dependency-aware service initialization
- Location: `scripts/config.sh` lines 93-117
- Pattern: Ordered array ensuring databases start before backends

**Custom Hooks:**
- Purpose: Encapsulate data fetching and business logic
- Location: `apps/dashboard/src/hooks/`
- Examples: `useApps.js`, `useContainers.js`, `useServiceManager.js`, `useBulkControl.js`
- Pattern: useState + useEffect + useCallback with auto-refresh intervals

**Port Allocation:**
- Purpose: Predictable port ranges per runtime
- Location: `scripts/config.sh` lines 9-22
- Pattern: 100-port ranges (PHP 8100-8199, Node 8200-8299, etc.)

**Response Format:**
- Purpose: Consistent API response structure
- Pattern: `{ success: boolean, data: {}, meta: {}, error?: string }`
- Used by: All 16 API endpoints

## Entry Points

**React Application:**
- Location: `apps/dashboard/src/main.jsx`
- Triggers: Browser navigation to dashboard
- Responsibilities: Mount React app, initialize providers (Theme, Toast)

**PHP API Router:**
- Location: `config/devarch/api/public/index.php`
- Triggers: HTTP requests to `/api/*`
- Responsibilities: Parse request, route to endpoint handler, return JSON

**Service Manager CLI:**
- Location: `scripts/service-manager.sh`
- Triggers: Terminal commands (`./service-manager.sh up postgres`)
- Responsibilities: Parse args, validate, execute compose operations

**Configuration:**
- Location: `scripts/config.sh`
- Triggers: Sourced by other scripts
- Responsibilities: Define categories, ports, paths, helper functions

## Error Handling

**Strategy:** Errors surface to boundaries, return structured responses

**Patterns:**
- PHP: try/catch at endpoint level, return `{ success: false, error: "message" }`
- React: Error state in hooks, display via Toast notifications
- Shell: Exit codes, status messages to stderr
- Podman: Error output captured and returned in API responses

## Cross-Cutting Concerns

**Logging:**
- Container logs via Podman native logging
- API access logged to stdout
- Dashboard: console.log in development only

**Validation:**
- PHP: `validateServiceName()`, `validateCategoryName()` regex patterns
- React: Form validation in components
- Shell: Input validation before operations

**Caching:**
- PHP static cache for container inspection (30-60s TTL)
- React localStorage for UI preferences
- No application-level cache (Redis available but not wired)

**Theme/UI State:**
- React Context API for global theme (dark/light)
- localStorage persistence for sidebar state, active tabs
- Tailwind dark mode via class strategy

---

*Architecture analysis: 2026-01-07*
*Update when major patterns change*
