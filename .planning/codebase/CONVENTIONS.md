# Coding Conventions

**Analysis Date:** 2026-01-07

## Naming Patterns

**Files:**
- kebab-case.sh for shell scripts (`service-manager.sh`, `runtime-switcher.sh`)
- PascalCase.jsx for React components (`ContainerCard.jsx`, `ActionButton.jsx`)
- camelCase.js for hooks and utilities (`useContainers.js`, `formatters.js`)
- snake_case.php for PHP libraries (`containers.php`, `services.php`)
- lowercase.yml for Compose files (`postgres.yml`, `redis.yml`)

**Functions:**
- camelCase for JavaScript (`formatRuntime`, `getContainerStatusBgClass`)
- camelCase for PHP (`getAllContainers`, `startService`)
- snake_case for Bash/Zsh (`print_status`, `get_service_files`)
- `handle*` prefix for React event handlers (`handleControl`, `handleSubmit`)

**Variables:**
- camelCase for JavaScript variables (`containerData`, `isRunning`)
- UPPER_SNAKE_CASE for constants (`RUNTIME_COLORS`, `STATUS_COLORS`, `REFRESH_INTERVAL`)
- UPPER_SNAKE_CASE for shell exports (`PROJECT_ROOT`, `NETWORK_NAME`)
- lowercase_snake_case for shell local vars (`service_file`, `category`)

**Types:**
- PascalCase for React contexts (`ThemeContext`, `ToastContext`)
- No TypeScript in dashboard (plain JavaScript)

## Code Style

**Formatting (JavaScript):**
- 2-space indentation
- Single quotes for strings
- Semicolons required
- No Prettier configured (ESLint only)

**Formatting (PHP):**
- 4-space indentation (PSR-12 style)
- Single quotes for strings
- DocBlocks for functions

**Formatting (Shell):**
- 4-space indentation
- Section headers with `=` decoration
- Comments above code blocks

**Linting:**
- ESLint 8.55.0 for JavaScript (`apps/dashboard/.eslintrc.cjs`)
- Extends: `eslint:recommended`, `plugin:react/recommended`, `plugin:react-hooks/recommended`
- `react/prop-types: off` (no PropTypes)
- No PHP linter configured

## Import Organization

**Order (JavaScript):**
1. React imports (`import { useState, useEffect } from 'react'`)
2. External packages (`import { Dialog } from '@headlessui/react'`)
3. Internal contexts (`import { useTheme } from '../contexts/ThemeContext'`)
4. Internal hooks (`import { useToast } from '../hooks/useToast'`)
5. Components (`import { SearchBar } from '../components/ui/SearchBar'`)
6. Utilities (`import { formatTime } from '../utils/formatters'`)

**Grouping:**
- Blank line between groups
- Alphabetical within groups (informal)

**Path Aliases:**
- None configured (relative paths only)

## Error Handling

**Patterns:**
- React: try/catch in async functions, error state in hooks
- PHP: try/catch at endpoint level, return error JSON
- Shell: Exit codes, stderr for errors

**Error Types:**
- JavaScript: throw Error, catch and display via Toast
- PHP: Return `{ success: false, error: "message" }`
- Shell: `echo "Error: ..." >&2; exit 1`

**Logging:**
- Development: console.log
- Production: Container stdout/stderr
- No structured logging framework

## Logging

**Framework:**
- JavaScript: console.log (development only)
- PHP: error_log for errors
- Container: Podman native logging

**Patterns:**
- Log at service boundaries
- Include context (container name, service name)
- No log levels configured

## Comments

**When to Comment:**
- Explain complex logic or algorithms
- Document business rules
- Avoid obvious comments

**JSDoc/TSDoc:**
- Optional (not widely used in dashboard)
- Used for exported hooks

**TODO Comments:**
- Format: `// TODO: description`
- Few found in codebase

**Shell Script Headers:**
```bash
# =============================================================================
# SECTION NAME
# =============================================================================
```

## Function Design

**Size:**
- Keep under 50 lines
- Extract helpers for complex logic

**Parameters:**
- Max 3-4 parameters
- Use options object for more
- Destructure in React components

**Return Values:**
- Explicit returns
- Return early for guard clauses
- Consistent response format in API

## Module Design

**Exports (JavaScript):**
- Named exports preferred (`export function useContainers()`)
- No default exports
- Context providers export both Provider and hook

**PHP Libraries:**
- Function-based (no classes)
- `require_once` for dependencies
- No namespacing

**Shell Scripts:**
- Source shared config via `. config.sh`
- Functions prefixed by purpose (`print_*`, `get_*`, `validate_*`)

## React Patterns

**Components:**
- Functional components only (no classes)
- Hooks for state and effects
- memo() for performance-critical components (`ContainerCard`)

**State Management:**
- useState for local state
- Context API for global state (Theme, Toast)
- No Redux or external state library

**Data Fetching:**
- Custom hooks encapsulate fetch logic
- Auto-refresh with configurable intervals
- Loading/error states in hooks

**Example Pattern:**
```javascript
export function useContainers(options = {}) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchData = useCallback(async () => {
    try {
      const response = await fetch('/api/containers');
      const json = await response.json();
      setData(json.data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, options.refreshInterval || 60000);
    return () => clearInterval(interval);
  }, [fetchData, options.refreshInterval]);

  return { data, loading, error, refetch: fetchData };
}
```

## PHP API Patterns

**Endpoint Structure:**
```php
<?php
require_once __DIR__ . '/../lib/common.php';
require_once __DIR__ . '/../lib/containers.php';

// Get request parameters
$filter = $_GET['filter'] ?? 'all';
$search = $_GET['search'] ?? '';

// Process request
$containers = getAllContainers();
$filtered = filterContainers($containers, $filter, $search);

// Return response
jsonResponse([
    'success' => true,
    'data' => ['containers' => $filtered],
    'meta' => ['count' => count($filtered)]
]);
```

**Response Format:**
- Always return JSON
- Structure: `{ success: bool, data: {}, meta: {}, error?: string }`
- HTTP 200 for all responses (error in body)

---

*Convention analysis: 2026-01-07*
*Update when patterns change*
