# Phase 21: Deploy Orchestration Service - Research

**Researched:** 2026-02-11
**Domain:** Go service layer architecture / business logic extraction
**Confidence:** HIGH

## Summary

Phase 21 extracts deploy orchestration logic (plan generation, apply execution, wiring resolution) from HTTP handlers into a dedicated application service layer. Current implementation has ~1841 lines across 8 stack handler files with orchestration logic tightly coupled to HTTP concerns.

The Go ecosystem strongly converges on service layer patterns for business logic extraction. Standard approach: handlers parse HTTP requests → delegate to service methods → format responses. Services hold business logic, coordinate DB transactions, and orchestrate multi-step operations. This pattern enables testing without HTTP overhead and reuse across different interfaces (API, CLI).

DevArch already follows dependency injection via constructor functions (`NewStackHandler(db, containerClient)`). Extraction requires creating `internal/orchestration` package with service interface, moving plan/apply/wiring logic from handlers, and injecting orchestration service into StackHandler.

**Primary recommendation:** Use "fat service" pattern with explicit struct dependencies (db, containerClient), not interface-based repositories. DevArch is small-to-medium scale, pragmatism over abstraction matches existing codebase style.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| database/sql | stdlib | Database access | Go standard, already in use |
| No DI framework | N/A | Manual dependency injection | Explicit wiring, no magic, testable |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| testify/mock | v1.x | Service mocking | Unit testing handlers with mocked services |
| sqlmock | v1.x | DB mocking | Unit testing service layer without real DB |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual DI | go.uber.org/dig | Overkill for single implementation per interface; adds magic |
| Fat service | Repository pattern | Extra abstraction layer unnecessary for current scale |
| Concrete types | Interface-first | Premature abstraction; add interfaces when multiple implementations needed |

**Installation:**
```bash
# No new dependencies required
# Testing libraries (optional, if adding tests):
go get github.com/stretchr/testify/mock
go get github.com/DATA-DOG/go-sqlmock
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/
├── orchestration/          # NEW: deploy orchestration service
│   ├── service.go          # Service struct + methods
│   └── service_test.go     # Service unit tests
├── api/handlers/
│   └── stack_*.go          # Handlers delegate to orchestration service
├── plan/                   # Existing: types, diff computation, token validation
├── wiring/                 # Existing: auto-wire resolution, validation
└── compose/                # Existing: YAML generation
```

### Pattern 1: Service Layer with Explicit Dependencies

**What:** Service struct holds concrete dependencies (db, containerClient), exposes methods for orchestration operations

**When to use:** Small-to-medium apps where abstraction overhead outweighs benefits, single implementation per interface

**Example:**
```go
// internal/orchestration/service.go
package orchestration

import (
    "database/sql"
    "github.com/priz/devarch-api/internal/container"
    "github.com/priz/devarch-api/internal/plan"
    "github.com/priz/devarch-api/internal/wiring"
)

type Service struct {
    db              *sql.DB
    containerClient *container.Client
}

func NewService(db *sql.DB, cc *container.Client) *Service {
    return &Service{
        db:              db,
        containerClient: cc,
    }
}

// GeneratePlan orchestrates plan generation: load stack, compute diff, resolve wiring, load resources
func (s *Service) GeneratePlan(stackName string) (*plan.Plan, error) {
    // Move logic from handlers/stack_plan.go:Plan() here
    // Returns plan.Plan or error
}

// ApplyPlan orchestrates plan application: validate token, create network, generate compose, run compose
func (s *Service) ApplyPlan(stackName string, token string, lockFile *lock.LockFile) (string, error) {
    // Move logic from handlers/stack_apply.go:Apply() here
    // Returns compose output or error
}

// ResolveWiring orchestrates auto-wire resolution: load providers/consumers, compute candidates, persist
func (s *Service) ResolveWiring(stackName string) ([]wiring.WireCandidate, []string, error) {
    // Move logic from handlers/stack_wiring.go:ResolveWires() here
    // Returns candidates, warnings, error
}
```

Source: [The 'fat service' pattern for Go web applications](https://www.alexedwards.net/blog/the-fat-service-pattern)

### Pattern 2: Handler Delegation

**What:** Handlers parse HTTP input into request structs, delegate to service, format response envelope

**When to use:** Always — separates transport concerns from business logic

**Example:**
```go
// internal/api/handlers/stack_plan.go (after refactoring)
func (h *StackHandler) Plan(w http.ResponseWriter, r *http.Request) {
    stackName := chi.URLParam(r, "name")

    // Delegate to service
    planResp, err := h.orchestrationService.GeneratePlan(stackName)
    if err != nil {
        // Error handling based on error type
        if errors.Is(err, orchestration.ErrStackNotFound) {
            respond.NotFound(w, r, "stack", stackName)
            return
        }
        respond.InternalError(w, r, err)
        return
    }

    respond.JSON(w, r, http.StatusOK, planResp)
}
```

Source: [Using the Service Object Pattern in Go](https://www.calhoun.io/using-the-service-object-pattern-in-go/)

### Pattern 3: Transaction Management in Service Layer

**What:** Service methods own transaction lifecycle, commit/rollback logic stays in service not handler

**When to use:** Multi-step operations requiring atomicity (e.g., wiring resolution with DB writes)

**Example:**
```go
func (s *Service) ResolveWiring(stackName string) ([]wiring.WireCandidate, []string, error) {
    stackID, err := s.loadStackID(stackName)
    if err != nil {
        return nil, nil, err
    }

    providers, consumers, existingWires, err := s.loadWiringData(stackID)
    if err != nil {
        return nil, nil, err
    }

    candidates, warnings := wiring.ResolveAutoWires(stackName, providers, consumers, existingWires)

    // Transaction lives in service layer
    tx, err := s.db.Begin()
    if err != nil {
        return nil, nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    if err := s.persistWireCandidates(tx, stackID, candidates); err != nil {
        return nil, nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, nil, fmt.Errorf("commit wiring: %w", err)
    }

    return candidates, warnings, nil
}
```

Source: [Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)

### Anti-Patterns to Avoid

- **God Service:** Don't create single service for all stack operations. Orchestration service focuses on deploy workflow (plan/apply/wiring), not CRUD or lifecycle.
- **Handler Logic Leakage:** Don't leave complex business logic in handlers. If it's more than parsing/formatting, move to service.
- **Premature Interfaces:** Don't create `IOrchestrationService` interface until second implementation needed. YAGNI principle.
- **Deep Nesting:** Don't create layers for layers' sake (handlers → controllers → services → repositories). Two layers sufficient: handlers (transport) + services (business logic).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dependency injection framework | Custom DI container with reflection | Manual constructor functions | Go idiom: explicit > implicit; constructor functions self-documenting |
| Database abstraction | Generic repository interface | Concrete *sql.DB | Single DB implementation, no swapping needed; stdlib sufficient |
| Service mocking | Manual mock implementations | testify/mock | Handles method call verification, return value stubbing, argument matchers |
| Advisory locks | Custom mutex coordination | Postgres pg_advisory_lock | Already in use in apply handler; DB-level locks survive process crashes |

**Key insight:** Go favors explicitness and simplicity. Don't build abstraction layers for theoretical future needs. Extract when pain is real, not anticipated.

## Common Pitfalls

### Pitfall 1: Interface Abstraction Too Early

**What goes wrong:** Creating `IOrchestrationService` interface when only one concrete implementation exists

**Why it happens:** Coming from Java/C# background where "code to interfaces" is gospel; premature optimization for testing

**How to avoid:** Use concrete `*orchestration.Service` type in handler. Add interface only when second implementation appears (e.g., mock for testing, alternative orchestrator)

**Warning signs:** Interface has same name as struct with "I" prefix; interface defined in same package as implementation; "future-proofing" mentioned in commit message

### Pitfall 2: Granular Method Explosion

**What goes wrong:** Service has 20+ tiny methods (loadStackID, loadProviders, loadConsumers, buildWiringSection, persistWires, etc.) all public

**Why it happens:** Over-applying single responsibility principle; making every helper public "for reusability"

**How to avoid:** Keep helpers private (lowercase). Public methods should represent complete business operations: GeneratePlan, ApplyPlan, ResolveWiring. Private helpers compose these.

**Warning signs:** Public methods with "load", "build", "persist" in name; method call chains spanning 5+ service methods; difficulty summarizing what service does

### Pitfall 3: Transaction Boundary Confusion

**What goes wrong:** Handler starts transaction, passes to service; or service returns open transaction to handler

**Why it happens:** Unclear ownership of transaction lifecycle

**How to avoid:** Service owns transaction lifecycle completely. Handler never sees tx. Service begins, commits/rolls back, returns result.

**Warning signs:** `tx *sql.Tx` in handler code; service methods accepting `tx` parameter from outside; rollback deferred in handler

### Pitfall 4: Mixing HTTP Concerns into Service

**What goes wrong:** Service methods accept `http.Request`, call `respond.JSON()`, return `http.ResponseWriter`

**Why it happens:** Incremental extraction without full separation; handler logic "bleeding" into service

**How to avoid:** Service methods accept/return Go types only (structs, primitives). HTTP parsing/rendering stays in handler. Service can return custom error types for handler to map to status codes.

**Warning signs:** Service imports `net/http`; service calls responder functions; service logs request IDs

## Code Examples

Verified patterns from official sources:

### Service Construction and Injection

```go
// cmd/server/main.go
func main() {
    db, _ := sql.Open("postgres", dbURL)
    containerClient, _ := container.NewClient()

    // Create orchestration service
    orchestrationService := orchestration.NewService(db, containerClient)

    // Inject into handlers
    stackHandler := handlers.NewStackHandler(db, containerClient, orchestrationService)

    router := api.NewRouter(stackHandler, ...)
    // ...
}

// internal/api/handlers/stack.go
type StackHandler struct {
    db                   *sql.DB
    containerClient      *container.Client
    orchestrationService *orchestration.Service  // NEW
}

func NewStackHandler(db *sql.DB, cc *container.Client, os *orchestration.Service) *StackHandler {
    return &StackHandler{
        db:                   db,
        containerClient:      cc,
        orchestrationService: os,
    }
}
```

Source: [Dependency Injection in Go: Patterns & Best Practices](https://www.glukhov.org/post/2025/12/dependency-injection-in-go/)

### Service Method Error Handling

```go
// internal/orchestration/service.go
var (
    ErrStackNotFound = errors.New("stack not found")
    ErrStackDisabled = errors.New("stack is disabled")
    ErrStalePlan     = errors.New("plan is stale")
)

func (s *Service) GeneratePlan(stackName string) (*plan.Plan, error) {
    stackID, enabled, err := s.loadStackMeta(stackName)
    if err == sql.ErrNoRows {
        return nil, ErrStackNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("load stack: %w", err)
    }

    if !enabled {
        return nil, ErrStackDisabled
    }

    // ... orchestration logic
    return planResp, nil
}

// Handler maps service errors to HTTP responses
func (h *StackHandler) Plan(w http.ResponseWriter, r *http.Request) {
    planResp, err := h.orchestrationService.GeneratePlan(stackName)
    if err != nil {
        switch {
        case errors.Is(err, orchestration.ErrStackNotFound):
            respond.NotFound(w, r, "stack", stackName)
        case errors.Is(err, orchestration.ErrStackDisabled):
            respond.Conflict(w, r, "stack is disabled — enable it first")
        default:
            respond.InternalError(w, r, err)
        }
        return
    }
    respond.JSON(w, r, http.StatusOK, planResp)
}
```

Source: [How to implement Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)

### Testing Service Layer

```go
// internal/orchestration/service_test.go
func TestService_GeneratePlan(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    containerClient := &container.Client{} // or mock
    service := orchestration.NewService(db, containerClient)

    // Setup mock expectations
    rows := sqlmock.NewRows([]string{"id", "enabled"}).AddRow(1, true)
    mock.ExpectQuery("SELECT id, enabled FROM stacks").
        WithArgs("test-stack").
        WillReturnRows(rows)

    plan, err := service.GeneratePlan("test-stack")

    assert.NoError(t, err)
    assert.NotNil(t, plan)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

Source: [Simple clean Go REST API architecture with dependency injection](https://irahardianto.github.io/service-pattern-go/)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Fat handlers with inline business logic | Service layer extraction | 2020-2022 trend | Testability, reusability, SoC |
| Interface-first repository pattern | Concrete types, add interfaces when needed | 2023-2025 pragmatism shift | Less boilerplate, faster iteration |
| Global DI frameworks (wire, dig) | Manual constructor injection | Always valid | Explicitness wins for small-medium apps |
| God services (everything in one service) | Domain-specific services | 2024-2025 refinement | Clear boundaries, focused responsibility |

**Deprecated/outdated:**
- **Global state patterns:** Services as package-level vars instead of structs with dependencies. Modern Go uses constructor functions returning pointers.
- **Deep layering:** Handler → Controller → Service → Repository → DAO. Two layers (handlers + services) sufficient unless microservices architecture.

## Open Questions

1. **Should orchestration service be split into multiple services (plan, apply, wiring)?**
   - What we know: Single orchestration service covers all three operations (plan, apply, wiring)
   - What's unclear: Whether these should be separate services or one orchestration service with multiple methods
   - Recommendation: Start with single `orchestration.Service`. Operations are tightly related (plan → apply workflow). Split only if service grows beyond ~500 lines or operations diverge in dependencies.

2. **Do we need orchestration service interface for testing?**
   - What we know: Handlers could be tested by mocking orchestration service
   - What's unclear: Current codebase has no handler tests; adding interfaces adds boilerplate without proven need
   - Recommendation: Skip interface initially. Use concrete `*orchestration.Service`. Add `IOrchestrationService` interface in Phase 26 (testing phase) if handler tests require mocking.

3. **Should service methods return domain errors or HTTP-aware errors?**
   - What we know: Service should be transport-agnostic (no HTTP dependencies)
   - What's unclear: How to communicate error types (404 vs 409 vs 500) without HTTP coupling
   - Recommendation: Service returns sentinel errors (`var ErrStackNotFound = errors.New(...)`). Handler maps errors to HTTP status codes using `errors.Is()` checks.

## Sources

### Primary (HIGH confidence)
- [The 'fat service' pattern for Go web applications](https://www.alexedwards.net/blog/the-fat-service-pattern) - Alex Edwards (authoritative Go educator)
- [Using the Service Object Pattern in Go](https://www.calhoun.io/using-the-service-object-pattern-in-go/) - Jon Calhoun
- [How to implement Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/) - Three Dots Labs
- [Simple clean Go REST API architecture with dependency injection](https://irahardianto.github.io/service-pattern-go/) - Reference implementation

### Secondary (MEDIUM confidence)
- [Dependency Injection in Go: Patterns & Best Practices](https://www.glukhov.org/post/2025/12/dependency-injection-in-go/) - Rost Glukhov (Dec 2025)
- [Go Project Structure: Practices & Patterns](https://www.glukhov.org/post/2025/12/go-project-structure/) - Rost Glukhov (Dec 2025)
- [Clean Architecture in Golang](https://baktiayp.medium.com/clean-architecture-in-golang-7bccf122f88b) - Bakti Pratama (Oct 2025)
- [Saga pattern in Go: building resilient distributed transactions](https://blog.devgenius.io/saga-pattern-in-go-building-resilient-distributed-transactions-with-orchestration-19d9746d8b85) - Mykola Guley (Feb 2026)

### Tertiary (LOW confidence)
- None - all sources verified with official docs or authoritative Go educators

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib only, no third-party dependencies required
- Architecture: HIGH - multiple authoritative sources converge on same patterns (fat service, manual DI, concrete types)
- Pitfalls: HIGH - well-documented in Go community (interface premature optimization, transaction boundaries)

**Research date:** 2026-02-11
**Valid until:** 60 days (stable patterns, not fast-moving)
