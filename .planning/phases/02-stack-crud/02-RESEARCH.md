# Phase 2: Stack CRUD - Research

**Researched:** 2026-02-03
**Domain:** Go API CRUD with PostgreSQL soft-delete, React dashboard with TanStack Query + Router
**Confidence:** HIGH

## Summary

Phase 2 implements Stack CRUD operations following established patterns from the existing service CRUD handlers (api/internal/api/handlers/service.go ~40KB). The existing codebase provides proven patterns: chi router with database/sql + lib/pq, deferred tx.Rollback(), chi.URLParam for name-based routing, gorilla/websocket for broadcasts, and TanStack Query mutations with invalidateQueries.

The primary challenge is soft-delete implementation without ORM. PostgreSQL supports soft-delete patterns via deleted_at timestamps with partial indexes for active rows only. Clone operations require application-level cascade logic (no built-in PostgreSQL cascade for copying records across FK relationships). WebSocket extension for stack events reuses existing /ws/status infrastructure.

**Primary recommendation:** Extend existing patterns (service.go handlers, useServices query hooks, useListControls view toggle) rather than introducing new libraries. Add deleted_at timestamp to stacks table, partial indexes for unique constraints on active rows, and manual cascade logic for clone operations.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/go-chi/chi | v5 | HTTP router | Already used, lightweight, composable, idiomatic Go |
| database/sql | stdlib | DB interface | Go standard, connection pooling built-in |
| github.com/lib/pq | latest | PostgreSQL driver | Already used (note: maintenance mode, but stable for sql.DB) |
| gorilla/websocket | latest | WebSocket server | Already used for /ws/status broadcasts |
| @tanstack/react-query | v5 | Server state | Already used, handles mutations + cache invalidation |
| @tanstack/react-router | latest | File-based routing | Already used, typed routes at build time |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| sonner | latest | Toast notifications | Already used for mutation feedback |
| Radix UI | latest | Dialog, dropdown | Already used for confirmation dialogs |
| Tailwind 4 | v4 | Styling | Already used |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| lib/pq | pgx/v5 | pgx faster + better maintained, but requires migration of existing code |
| Manual soft-delete | GORM soft_delete | GORM adds ORM layer, conflicts with existing sql.DB patterns |
| WebSocket broadcast | Server-Sent Events (SSE) | SSE simpler for one-way updates, but existing WS infra already works |

**Installation:**
```bash
# All dependencies already in project
# No new Go packages required
# No new npm packages required
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/api/handlers/
├── stack.go              # New: Stack CRUD handlers
├── service.go            # Reference pattern (existing ~40KB)
├── websocket.go          # Extend for stack events
└── ...

dashboard/src/
├── features/stacks/
│   ├── queries.ts        # useStacks, useCreateStack, useCloneStack, etc.
│   └── types.ts          # Stack type definitions
├── routes/stacks/
│   ├── index.tsx         # List page (card + table toggle)
│   ├── $name.tsx         # Detail page
│   └── trash.tsx         # Trash view (optional)
├── components/stacks/
│   ├── stack-table.tsx   # Table view component
│   ├── stack-grid.tsx    # Card grid component
│   └── ...
└── hooks/
    └── use-list-controls.ts  # Already exists, reuse
```

### Pattern 1: Chi Router CRUD Handler
**What:** RESTful handler struct with DB + container client dependencies
**When to use:** All entity CRUD operations

**Example:**
```go
// Source: Existing api/internal/api/handlers/service.go pattern
type StackHandler struct {
    db              *sql.DB
    containerClient *container.Client
}

func NewStackHandler(db *sql.DB, cc *container.Client) *StackHandler {
    return &StackHandler{db: db, containerClient: cc}
}

func (h *StackHandler) List(w http.ResponseWriter, r *http.Request) {
    query := `SELECT id, name, description, enabled, created_at, updated_at
              FROM stacks WHERE deleted_at IS NULL`
    // Apply filters, sorting, pagination...
    rows, err := h.db.Query(query, args...)
    // Scan rows, return JSON
}

func (h *StackHandler) Get(w http.ResponseWriter, r *http.Request) {
    name := chi.URLParam(r, "name")
    var s Stack
    err := h.db.QueryRow(`SELECT ... FROM stacks WHERE name = $1 AND deleted_at IS NULL`, name).Scan(...)
    // Return JSON
}
```

### Pattern 2: Deferred Rollback Transaction Pattern
**What:** Begin tx, defer rollback, commit on success
**When to use:** Any multi-step DB operation (create with cascade, delete with cascade, clone)

**Example:**
```go
// Source: Official Go docs + existing service.go pattern
func (h *StackHandler) Create(w http.ResponseWriter, r *http.Request) {
    tx, err := h.db.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback() // Safe even if Commit succeeds

    var stackID int
    err = tx.QueryRow(`INSERT INTO stacks (name, description) VALUES ($1, $2) RETURNING id`,
        req.Name, req.Description).Scan(&stackID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return // Rollback called by defer
    }

    if err := tx.Commit(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Success
}
```

### Pattern 3: Soft Delete with Partial Indexes
**What:** deleted_at timestamp column, exclude from queries, partial unique indexes
**When to use:** All stack queries, unique constraints on active rows only

**Example:**
```sql
-- Migration 013
CREATE TABLE stacks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(63) NOT NULL,
    description TEXT DEFAULT '',
    enabled BOOLEAN DEFAULT true,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partial index: unique name only for active stacks
CREATE UNIQUE INDEX uq_stacks_name_active
    ON stacks(name) WHERE deleted_at IS NULL;

-- Index for trash queries
CREATE INDEX idx_stacks_deleted_at ON stacks(deleted_at);
```

**Queries:**
```go
// List active stacks
SELECT * FROM stacks WHERE deleted_at IS NULL

// Soft delete
UPDATE stacks SET deleted_at = NOW() WHERE name = $1 AND deleted_at IS NULL

// Restore from trash
UPDATE stacks SET deleted_at = NULL WHERE name = $1 AND deleted_at IS NOT NULL

// Permanent delete
DELETE FROM stacks WHERE name = $1 AND deleted_at IS NOT NULL
```

### Pattern 4: Clone with Manual Cascade
**What:** Application-level transaction copying parent + child records
**When to use:** Clone operations (no built-in PostgreSQL cascade for copies)

**Example:**
```go
// Clone stack with instances
func (h *StackHandler) Clone(w http.ResponseWriter, r *http.Request) {
    sourceName := chi.URLParam(r, "name")
    var req struct{ NewName string }
    json.NewDecoder(r.Body).Decode(&req)

    tx, err := h.db.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    // Copy parent record
    var newStackID int
    err = tx.QueryRow(`
        INSERT INTO stacks (name, description, network_name, enabled)
        SELECT $1, description, $2, enabled
        FROM stacks WHERE name = $3 AND deleted_at IS NULL
        RETURNING id`,
        req.NewName, req.NewName+"-net", sourceName).Scan(&newStackID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Copy child records (Phase 3: instances + overrides)
    // This phase: no instances yet, clone is records-only

    if err := tx.Commit(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Success
}
```

### Pattern 5: TanStack Query Mutations with Cache Invalidation
**What:** useMutation hooks with onSuccess invalidateQueries
**When to use:** All mutations (create, update, delete, clone, enable/disable)

**Example:**
```typescript
// Source: Existing dashboard/src/features/services/queries.ts
export function useCreateStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: CreateStackRequest) => {
      const response = await api.post('/stacks', data)
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack created')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
    onError: (error) => {
      toast.error('Failed to create stack')
    },
  })
}

export function useCloneStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ name, newName }: { name: string; newName: string }) => {
      const response = await api.post(`/stacks/${name}/clone`, { new_name: newName })
      return response.data
    },
    onSuccess: () => {
      toast.success('Stack cloned')
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
  })
}
```

### Pattern 6: Card + Table View Toggle with Shared State
**What:** useListControls hook persists view mode, filters, sort to localStorage
**When to use:** List pages with multiple view modes

**Example:**
```typescript
// Source: Existing dashboard/src/hooks/use-list-controls.ts + routes/services/index.tsx
const controls = useListControls({
  storageKey: 'stacks',
  items: stacks,
  searchFn: (s, q) => s.name.toLowerCase().includes(q.toLowerCase()),
  filterFns: {
    enabled: (s, v) => v === 'all' || (v === 'true') === s.enabled,
  },
  sortFns: {
    name: (a, b) => a.name.localeCompare(b.name),
    created: (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  },
  defaultSort: 'name',
  defaultView: 'grid',
})

// Both views use controls.filtered
{controls.viewMode === 'table' ? (
  <StackTable stacks={controls.filtered} />
) : (
  <StackGrid stacks={controls.filtered} />
)}
```

### Pattern 7: WebSocket Broadcast Extension
**What:** Extend existing /ws/status WebSocket handler to include stack events
**When to use:** Real-time updates for stack creation, deletion, enable/disable

**Example:**
```go
// Source: Existing api/internal/api/handlers/websocket.go
// Extend syncManager to track stack state changes
type StatusUpdate struct {
    Services []ServiceStatus `json:"services"`
    Stacks   []StackStatus   `json:"stacks"` // Add stack status
}

// Broadcast on stack mutations
func (h *StackHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ... create stack ...

    // Notify WebSocket clients
    h.wsHandler.Broadcast(map[string]interface{}{
        "type": "stack_created",
        "stack": stackData,
    })
}
```

### Anti-Patterns to Avoid

- **Hard-coded `deleted_at IS NULL` checks everywhere:** Create helper function `ActiveStacksQuery()` returning base WHERE clause
- **Forgetting partial indexes on soft-delete:** Unique constraints MUST include `WHERE deleted_at IS NULL` or allow duplicates in trash
- **Shadowing error variables in transactions:** Always check `if err != nil` immediately after each tx operation
- **Using `git add .` or `git add -A`:** Stage specific files to avoid committing .env or credentials

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| View state persistence | Custom localStorage wrapper | Existing useListControls hook | Already handles sort, filter, view, search state with localStorage |
| Confirmation dialogs | Custom modal component | Radix AlertDialog | Already imported, accessible, keyboard nav built-in |
| Toast notifications | Custom snackbar | sonner toast | Already used project-wide, consistent UX |
| WebSocket connection | New WS implementation | Extend existing /ws/status | Infrastructure exists, clients already connected |
| Transaction rollback | Manual rollback tracking | defer tx.Rollback() | Go idiom, safe even if Commit succeeds |
| Validation errors | Generic "invalid input" | Prescriptive messages with fix suggestions | Existing pattern: "Name must be DNS-safe. Try: my-stack" |

**Key insight:** This codebase already has 40KB service CRUD handler with all necessary patterns. Clone and adapt rather than reinvent.

## Common Pitfalls

### Pitfall 1: Forgetting deleted_at IS NULL in Queries
**What goes wrong:** Soft-deleted records appear in lists, detail pages return deleted items, clone copies deleted stack
**Why it happens:** Easy to forget WHERE clause when querying by name or ID
**How to avoid:** Create constant or helper function for active filter
**Warning signs:** Users report "deleted" stacks still showing up

**Prevention:**
```go
const ActiveStacksFilter = "deleted_at IS NULL"

func (h *StackHandler) List(w http.ResponseWriter, r *http.Request) {
    query := fmt.Sprintf("SELECT ... FROM stacks WHERE %s", ActiveStacksFilter)
}
```

### Pitfall 2: Missing Partial Unique Index
**What goes wrong:** Cannot reuse stack name after soft-delete (unique constraint violation)
**Why it happens:** Standard UNIQUE INDEX applies to all rows including deleted_at IS NOT NULL
**How to avoid:** Always use `WHERE deleted_at IS NULL` in unique indexes on soft-deleted tables
**Warning signs:** "duplicate key" error when creating stack with name previously deleted

**Solution:**
```sql
-- WRONG: applies to all rows
CREATE UNIQUE INDEX uq_stacks_name ON stacks(name);

-- RIGHT: applies only to active rows
CREATE UNIQUE INDEX uq_stacks_name_active ON stacks(name) WHERE deleted_at IS NULL;
```

### Pitfall 3: Cascade Delete Without Confirmation
**What goes wrong:** User clicks delete, all instances and containers gone without warning
**Why it happens:** Backend deletes with ON DELETE CASCADE, frontend confirms but doesn't show blast radius
**How to avoid:** GET /stacks/:name/delete-preview returns affected resources, frontend shows in dialog
**Warning signs:** User complaints about data loss

**Pattern:**
```go
// Preview endpoint
func (h *StackHandler) DeletePreview(w http.ResponseWriter, r *http.Request) {
    name := chi.URLParam(r, "name")
    var preview struct {
        Stack      string   `json:"stack"`
        Instances  int      `json:"instance_count"`
        Containers []string `json:"container_names"`
    }
    // Query counts and names
    json.NewEncoder(w).Encode(preview)
}
```

### Pitfall 4: Clone Fails Silently on Name Conflict
**What goes wrong:** Clone with existing name returns 500, user doesn't know why
**Why it happens:** Unique constraint violation not mapped to user-friendly error
**How to avoid:** Check pq error code, return prescriptive message with suggestion
**Warning signs:** Generic "internal server error" on clone

**Pattern:**
```go
if err != nil {
    if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique_violation
        http.Error(w, fmt.Sprintf("Stack name '%s' already exists. Try '%s-copy'",
            req.NewName, req.NewName), http.StatusConflict)
        return
    }
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

### Pitfall 5: Disable Without Stopping Containers
**What goes wrong:** Stack.enabled = false but containers still running
**Why it happens:** enabled flag is DB-only, doesn't trigger container stop
**How to avoid:** Disable endpoint must: 1) list running containers 2) stop via container client 3) set enabled=false
**Warning signs:** Disabled stacks show running containers

**Pattern:**
```go
func (h *StackHandler) Disable(w http.ResponseWriter, r *http.Request) {
    // 1. Get stack
    // 2. Query running containers for this stack
    // 3. Call containerClient.StopService for each
    // 4. UPDATE stacks SET enabled = false
    // 5. Broadcast WebSocket event
}
```

### Pitfall 6: Transaction Rollback Error Ignored in Defer
**What goes wrong:** Linter complains about unhandled defer error
**Why it happens:** `defer tx.Rollback()` doesn't check error
**How to avoid:** This is intentional — even if Rollback fails, tx is invalid and not committed
**Warning signs:** Linter warnings, developers add incorrect error handling

**Correct pattern:**
```go
// From official Go docs: safe to ignore rollback error
defer tx.Rollback() // No error check needed

// Alternative if you must handle it:
defer func() {
    if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
        log.Printf("rollback error: %v", err)
    }
}()
```

### Pitfall 7: Rename as Separate Operation
**What goes wrong:** Rename appears as separate feature, but stacks table has name as primary key
**Why it happens:** Name is immutable identifier (chi.URLParam), not arbitrary field
**How to avoid:** Rename = clone + soft-delete in single transaction, expose as "Rename" in UI
**Warning signs:** Users expect PATCH /stacks/:name with new name in body

**Pattern:**
```go
// Rename endpoint wraps clone + delete
func (h *StackHandler) Rename(w http.ResponseWriter, r *http.Request) {
    oldName := chi.URLParam(r, "name")
    var req struct{ NewName string }
    json.NewDecoder(r.Body).Decode(&req)

    tx, _ := h.db.Begin()
    defer tx.Rollback()

    // Clone to new name
    var newID int
    tx.QueryRow(`INSERT INTO stacks ... SELECT ... WHERE name = $1`, oldName).Scan(&newID)

    // Soft-delete old
    tx.Exec(`UPDATE stacks SET deleted_at = NOW() WHERE name = $1`, oldName)

    tx.Commit()
    // Returns new stack as if rename happened
}
```

## Code Examples

Verified patterns from official sources and existing codebase:

### Chi Router Registration
```go
// Source: Existing api/internal/api/routes.go
r.Route("/api/v1", func(r chi.Router) {
    r.Use(mw.APIKeyAuth)
    r.Use(mw.RateLimit(10, 50))

    r.Route("/stacks", func(r chi.Router) {
        r.Get("/", stackHandler.List)
        r.Post("/", stackHandler.Create)

        r.Route("/{name}", func(r chi.Router) {
            r.Get("/", stackHandler.Get)
            r.Put("/", stackHandler.Update)
            r.Delete("/", stackHandler.Delete) // Soft delete

            r.Post("/enable", stackHandler.Enable)
            r.Post("/disable", stackHandler.Disable)
            r.Post("/clone", stackHandler.Clone)

            r.Get("/delete-preview", stackHandler.DeletePreview)
        })

        r.Get("/trash", stackHandler.ListTrash)
        r.Post("/trash/{name}/restore", stackHandler.Restore)
        r.Delete("/trash/{name}", stackHandler.PermanentDelete)
    })
})
```

### Prescriptive Validation Errors
```go
// Pattern from existing service validation
func validateStackName(name string) error {
    if len(name) < 2 || len(name) > 63 {
        return fmt.Errorf("Stack name must be 2-63 characters")
    }
    if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`).MatchString(name) {
        slug := strings.ToLower(name)
        slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "-")
        slug = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(slug, "")
        return fmt.Errorf("Stack name must be DNS-safe (lowercase, alphanumeric, hyphens). Try: %s", slug)
    }
    return nil
}
```

### TanStack Query List with Polling
```typescript
// Source: Existing dashboard/src/features/services/queries.ts
export function useStacks() {
  return useQuery({
    queryKey: ['stacks'],
    queryFn: async () => {
      const response = await api.get<Stack[]>('/stacks?include=summary')
      return response.data
    },
    refetchInterval: 5000, // Poll for real-time updates
  })
}
```

### Delete Confirmation with Blast Radius
```typescript
// Pattern: fetch preview, show in dialog, confirm, mutate
const { mutate: deleteStack } = useDeleteStack()
const [preview, setPreview] = useState<DeletePreview | null>(null)

async function handleDelete() {
  const res = await api.get(`/stacks/${name}/delete-preview`)
  setPreview(res.data)
  setShowDeleteDialog(true)
}

function confirmDelete() {
  deleteStack(name, {
    onSuccess: () => navigate('/stacks')
  })
}

// In dialog
<AlertDialogDescription>
  This will delete stack "{preview.stack}" and:
  <ul>
    <li>{preview.instance_count} instances</li>
    <li>Stop {preview.containers.length} containers: {preview.containers.join(', ')}</li>
  </ul>
</AlertDialogDescription>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| lib/pq recommended | pgx/v5 preferred for new projects | 2024-2025 | lib/pq in maintenance mode but stable; migration not required |
| Hard delete with FK CASCADE | Soft delete with deleted_at | Ongoing | Enables trash/restore, audit trails, but complicates queries |
| GORM for soft delete | Manual deleted_at + partial indexes | N/A | Avoid ORM in database/sql projects, explicit control |
| Imperative query invalidation | TanStack Query v5 auto-invalidation | 2023 | Cleaner mutation code, less manual cache management |
| Class components | React 19 hooks | 2024 | Function components standard, better TS inference |

**Deprecated/outdated:**
- GORM soft_delete plugin: Not compatible with database/sql approach
- TanStack Query v4: v5 is current, breaking changes in mutation API
- Separate state management for view toggle: useListControls hook consolidates localStorage persistence

## Open Questions

Things that couldn't be fully resolved:

1. **Trash retention policy**
   - What we know: Soft delete is clear, trash view is decided
   - What's unclear: Auto-purge timing (30 days? 90 days? manual only?)
   - Recommendation: Start with manual-only purge, add scheduled cleanup in later phase if needed

2. **Name reuse after soft delete**
   - What we know: Partial unique index allows name reuse for active stacks
   - What's unclear: Should restore conflict with newer stack using same name?
   - Recommendation: Restore checks for name conflict, prompts user to rename if needed

3. **Clone behavior with Phase 3 instances**
   - What we know: Phase 2 clone is records-only, no containers started
   - What's unclear: How Phase 3 instance overrides are cloned (shallow vs deep copy)
   - Recommendation: Clone copies all override records as-is, no container materialization

4. **WebSocket event granularity**
   - What we know: Existing /ws/status broadcasts service status every 5s
   - What's unclear: Should stack events be immediate or batched with service events?
   - Recommendation: Add stack events to existing 5s broadcast cycle, same StatusUpdate payload

5. **Enable without starting containers**
   - What we know: Disable stops containers, Enable prompts "Start now?"
   - What's unclear: If user says "no", does enabled=true but containers stopped?
   - Recommendation: Enabled=true updates DB regardless, prompt is convenience for immediate start

## Sources

### Primary (HIGH confidence)
- Official Go documentation: [Executing transactions](https://go.dev/doc/database/execute-transactions)
- TanStack Query v5: [Optimistic Updates](https://tanstack.com/query/v5/docs/react/guides/optimistic-updates) | [Mutations](https://tanstack.com/query/v5/docs/react/guides/mutations)
- Existing codebase: api/internal/api/handlers/service.go (CRUD reference pattern)
- Existing codebase: dashboard/src/hooks/use-list-controls.ts (view toggle pattern)
- Existing codebase: api/internal/api/handlers/websocket.go (broadcast pattern)

### Secondary (MEDIUM confidence)
- Evil Martians: [Soft deletion with PostgreSQL](https://evilmartians.com/chronicles/soft-deletion-with-postgresql-but-with-logic-on-the-database) - Database-level soft delete patterns
- DEV Community: [Soft delete cascade in PostgreSQL](https://dev.to/yugabyte/soft-delete-cascade-in-postgresql-and-yugabytedb-166n) - ON UPDATE CASCADE for soft deletes
- OneUpTime: [Go WebSocket with Gorilla](https://oneuptime.com/blog/post/2026-02-01-go-websocket-gorilla/view) - Hub pattern for broadcasts (2026-02-01)
- OneUpTime: [PostgreSQL connection pooling in Go](https://oneuptime.com/blog/post/2026-01-07-go-postgresql-connection-pooling/view) - database/sql pooling config (2026-01-07)
- Medium: [Handling Database Transactions in Go](https://medium.com/@cosmicray001/handling-database-transactions-in-go-with-rollback-and-commit-e35c1830b825) - Deferred rollback pattern
- Dashbit: [Soft deletes with Ecto and PostgreSQL](https://dashbit.co/blog/soft-deletes-with-ecto) - Partial index patterns

### Tertiary (LOW confidence)
- LogRocket: [Tables to grids with React compound components](https://blog.logrocket.com/converting-tables-to-grids-with-react-compound-components/) - View toggle patterns
- DEV Community: [Go validation with validator v10](https://dev.to/kittipat1413/a-guide-to-input-validation-in-go-with-validator-v10-56bp) - User-friendly error messages
- Various: PostgreSQL table cloning articles - No built-in cascade copy, manual required

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use, versions confirmed
- Architecture: HIGH - Existing handlers provide proven patterns
- Pitfalls: MEDIUM - Soft delete pitfalls verified, cascade pitfalls inferred from FK behavior
- Code examples: HIGH - Extracted from existing codebase + official docs

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - stable stack, patterns unlikely to change)
