# Architecture Research: Service Composition in Go API

**Domain:** Service composition/stacking system for containerized environments
**Researched:** 2026-02-03
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      API Layer (Chi Router)                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────┐   │
│  │  Stack   │  │ Instance │  │  Wiring  │  │ Plan/Apply    │   │
│  │ Handlers │  │ Handlers │  │ Resolver │  │   Engine      │   │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └───────┬───────┘   │
│        │             │              │               │           │
├────────┴─────────────┴──────────────┴───────────────┴───────────┤
│                      Service Layer                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐  │
│  │  Config        │  │  Wiring Graph  │  │  Compose         │  │
│  │  Resolver      │  │  Builder       │  │  Generator       │  │
│  └────────┬───────┘  └────────┬───────┘  └─────────┬────────┘  │
│           │                   │                     │           │
├───────────┴───────────────────┴─────────────────────┴───────────┤
│                      Database Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐    │
│  │ Services │  │  Stacks  │  │ Instance │  │   Wiring     │    │
│  │ (base)   │  │          │  │ Overrides│  │   Contracts  │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| Stack Handlers | HTTP API for stack CRUD, plan/apply ops | Chi router handlers + JSON marshaling |
| Config Resolver | Merge template service + instance overrides | Recursive merge with precedence rules |
| Wiring Graph Builder | Resolve service dependencies + contracts | DAG construction + topological sort |
| Plan/Apply Engine | Preview changes, execute with locking | PostgreSQL advisory locks + transaction |
| Compose Generator | Generate stack-scoped or per-service YAML | DB query + YAML marshaling |
| Database Layer | Store templates, instances, overrides | PostgreSQL with JSONB for overrides |

## Recommended Project Structure

```
api/
├── internal/
│   ├── compose/
│   │   ├── generator.go         # Existing: per-service compose
│   │   ├── stack_generator.go   # NEW: stack-scoped compose
│   │   └── resolver.go          # NEW: config resolution (merge)
│   ├── wiring/
│   │   ├── graph.go             # NEW: dependency graph builder
│   │   ├── resolver.go          # NEW: auto-wire + explicit contracts
│   │   └── validator.go         # NEW: cycle detection, port conflicts
│   ├── plan/
│   │   ├── planner.go           # NEW: diff current vs desired state
│   │   ├── executor.go          # NEW: apply with locking
│   │   └── lock.go              # NEW: advisory lock helpers
│   ├── handlers/
│   │   ├── stacks.go            # NEW: stack CRUD
│   │   ├── instances.go         # NEW: instance CRUD
│   │   └── plans.go             # NEW: plan/apply endpoints
│   └── db/
│       └── queries.go           # Extend with stack/instance queries
├── migrations/
│   ├── 013_stacks.up.sql        # NEW: stacks table
│   ├── 014_instances.up.sql     # NEW: service_instances + overrides
│   ├── 015_wiring.up.sql        # NEW: wiring contracts
│   └── 016_plans.up.sql         # NEW: plan state tracking
└── pkg/
    └── models/
        └── models.go            # Extend with Stack, Instance, Contract
```

### Structure Rationale

- **compose/**: Extends existing generator with stack-scoped + resolution logic
- **wiring/**: New package for dependency graph + auto-wiring (separate concern)
- **plan/**: Plan/apply pattern isolated from execution logic
- **handlers/**: Clear separation stack vs instance vs plan endpoints

## Architectural Patterns

### Pattern 1: Copy-on-Write Override Schema

**What:** Template service defines defaults, instance stores only overrides

**When to use:** Multiple instances of same service with minor config differences

**Trade-offs:**
- PRO: Minimal storage, DRY configuration
- PRO: Template changes propagate to all instances (unless overridden)
- CON: Query complexity for effective config resolution
- CON: Migration complexity when template schema changes

**Schema Design:**

```sql
-- Base template (existing services table)
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    category_id INTEGER,
    image_name VARCHAR(256) NOT NULL,
    image_tag VARCHAR(64) DEFAULT 'latest',
    -- ... all existing fields
);

-- Stacks: logical grouping of service instances
CREATE TABLE stacks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    description TEXT,
    environment VARCHAR(32) DEFAULT 'development',
    compose_mode VARCHAR(32) DEFAULT 'stack-scoped',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Service instances: template reference + overrides
CREATE TABLE service_instances (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    service_id INTEGER REFERENCES services(id),
    instance_name VARCHAR(128) NOT NULL, -- e.g., "api-v1", "api-v2"

    -- Copy-on-write overrides (NULL = inherit from template)
    image_tag_override VARCHAR(64),
    restart_policy_override VARCHAR(32),
    command_override TEXT,
    user_spec_override VARCHAR(64),
    enabled_override BOOLEAN,

    -- JSONB for arbitrary compose overrides
    compose_overrides JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(stack_id, instance_name)
);

CREATE INDEX idx_instances_stack ON service_instances(stack_id);
CREATE INDEX idx_instances_service ON service_instances(service_id);

-- Override tables (copy-on-write pattern for relations)
CREATE TABLE instance_port_overrides (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    host_ip VARCHAR(45) DEFAULT '127.0.0.1',
    host_port INTEGER NOT NULL,
    container_port INTEGER NOT NULL,
    protocol VARCHAR(8) DEFAULT 'tcp',
    UNIQUE(instance_id, host_ip, host_port)
);

CREATE TABLE instance_env_overrides (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    key VARCHAR(256) NOT NULL,
    value TEXT,
    is_secret BOOLEAN DEFAULT false,
    UNIQUE(instance_id, key)
);

CREATE TABLE instance_volume_overrides (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    volume_type VARCHAR(16) NOT NULL,
    source VARCHAR(512) NOT NULL,
    target VARCHAR(512) NOT NULL,
    read_only BOOLEAN DEFAULT false
);
```

### Pattern 2: Effective Config Resolution

**What:** Merge template + instance overrides at query time or API layer

**When to use:** Every compose generation, API GET requests

**Trade-offs:**
- PRO: Single source of truth (no denormalization)
- PRO: Template changes immediately visible
- CON: Query complexity (LEFT JOINs + COALESCE)
- CON: Performance impact for large stacks

**Resolution Strategy:**

```go
// Option A: SQL-level resolution (performant, complex query)
func (r *Resolver) ResolveInstanceConfig(instanceID int) (*ResolvedConfig, error) {
    query := `
    SELECT
        si.instance_name,
        s.image_name,
        COALESCE(si.image_tag_override, s.image_tag) AS image_tag,
        COALESCE(si.restart_policy_override, s.restart_policy) AS restart_policy,
        COALESCE(si.command_override, s.command) AS command,
        s.id AS template_id
    FROM service_instances si
    JOIN services s ON si.service_id = s.id
    WHERE si.id = $1
    `
    // Then resolve ports, env, volumes with similar COALESCE logic
    // Merge template relations with override relations
}

// Option B: Application-level resolution (flexible, more queries)
func (r *Resolver) ResolveInstanceConfig(instanceID int) (*ResolvedConfig, error) {
    instance := r.LoadInstance(instanceID)
    template := r.LoadService(instance.ServiceID)
    overrides := r.LoadOverrides(instanceID)

    return r.Merge(template, instance, overrides)
}

// Merge precedence: instance overrides > template defaults
func (r *Resolver) Merge(template, instance, overrides) *ResolvedConfig {
    config := template.Clone()

    // Scalar overrides
    if instance.ImageTagOverride != nil {
        config.ImageTag = *instance.ImageTagOverride
    }

    // Collection merging (ports, env, volumes)
    config.Ports = r.MergePorts(template.Ports, overrides.Ports)
    config.EnvVars = r.MergeEnvVars(template.EnvVars, overrides.EnvVars)

    // JSONB compose_overrides applied last (highest precedence)
    config = r.ApplyComposeOverrides(config, instance.ComposeOverrides)

    return config
}
```

**Recommendation:** Start with application-level (Option B) for flexibility. Migrate to SQL-level (Option A) if profiling shows resolution is a bottleneck.

### Pattern 3: Service Wiring Graph Resolution

**What:** Build directed acyclic graph (DAG) of service dependencies, resolve wiring contracts

**When to use:** Stack generation, validation, startup order calculation

**Trade-offs:**
- PRO: Explicit contracts prevent port conflicts, wrong protocols
- PRO: Auto-wiring reduces boilerplate for standard patterns
- CON: Graph complexity grows with stack size
- CON: Requires cycle detection

**Schema:**

```sql
-- Explicit wiring contracts
CREATE TABLE wiring_contracts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    protocol VARCHAR(32) NOT NULL, -- 'http', 'grpc', 'postgres', 'redis'
    default_port INTEGER,
    description TEXT
);

-- Instance-to-instance wiring
CREATE TABLE instance_wiring (
    id SERIAL PRIMARY KEY,
    consumer_instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    provider_instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    contract_id INTEGER REFERENCES wiring_contracts(id),

    -- Auto-wired or explicit
    wiring_mode VARCHAR(16) DEFAULT 'auto', -- 'auto', 'explicit'

    -- Override connection details (explicit mode)
    connection_override JSONB, -- {host, port, path, credentials}

    UNIQUE(consumer_instance_id, provider_instance_id, contract_id)
);

CREATE INDEX idx_wiring_consumer ON instance_wiring(consumer_instance_id);
CREATE INDEX idx_wiring_provider ON instance_wiring(provider_instance_id);
```

**Wiring Resolution Algorithm:**

```go
type WiringGraph struct {
    nodes map[int]*InstanceNode // instanceID -> node
    edges []*WiringEdge
}

type InstanceNode struct {
    InstanceID   int
    InstanceName string
    ServiceType  string // postgres, redis, api, worker
}

type WiringEdge struct {
    From     int // consumer instance
    To       int // provider instance
    Contract *WiringContract
    Mode     string // auto, explicit
}

func (g *WiringGraph) Build(stackID int) error {
    instances := g.LoadStackInstances(stackID)

    // Add nodes
    for _, inst := range instances {
        g.AddNode(inst)
    }

    // Explicit wiring
    explicitWiring := g.LoadExplicitWiring(stackID)
    for _, wire := range explicitWiring {
        g.AddEdge(wire.ConsumerID, wire.ProviderID, wire.Contract, "explicit")
    }

    // Auto-wiring: infer from service types + standard contracts
    g.AutoWire(instances)

    // Validate
    if cycles := g.DetectCycles(); len(cycles) > 0 {
        return fmt.Errorf("dependency cycles detected: %v", cycles)
    }

    return nil
}

func (g *WiringGraph) AutoWire(instances []*Instance) {
    // Example: All API instances auto-wire to first postgres instance
    postgres := g.FindByServiceType("postgres")
    apis := g.FindByServiceType("api")

    postgresContract := g.LoadContract("postgres-tcp")
    for _, api := range apis {
        if !g.HasExplicitWiring(api.ID, postgres.ID) {
            g.AddEdge(api.ID, postgres.ID, postgresContract, "auto")
        }
    }

    // Similar logic for redis, rabbitmq, etc.
}

func (g *WiringGraph) DetectCycles() [][]int {
    // Tarjan's algorithm or DFS-based cycle detection
    visited := make(map[int]bool)
    recStack := make(map[int]bool)
    var cycles [][]int

    for nodeID := range g.nodes {
        if !visited[nodeID] {
            g.dfsCycleDetect(nodeID, visited, recStack, &cycles)
        }
    }

    return cycles
}

func (g *WiringGraph) TopologicalSort() ([]*InstanceNode, error) {
    // Kahn's algorithm for startup order
    inDegree := make(map[int]int)
    queue := []int{}
    sorted := []*InstanceNode{}

    // Calculate in-degrees
    for nodeID := range g.nodes {
        inDegree[nodeID] = g.InDegree(nodeID)
        if inDegree[nodeID] == 0 {
            queue = append(queue, nodeID)
        }
    }

    // Process queue
    for len(queue) > 0 {
        nodeID := queue[0]
        queue = queue[1:]
        sorted = append(sorted, g.nodes[nodeID])

        for _, edge := range g.OutEdges(nodeID) {
            inDegree[edge.To]--
            if inDegree[edge.To] == 0 {
                queue = append(queue, edge.To)
            }
        }
    }

    if len(sorted) != len(g.nodes) {
        return nil, fmt.Errorf("cycle detected")
    }

    return sorted, nil
}
```

### Pattern 4: Plan/Apply with Advisory Locking

**What:** Terraform-style plan (preview) → apply (execute) workflow with distributed locking

**When to use:** Multi-step operations requiring user confirmation, concurrent API access

**Trade-offs:**
- PRO: Prevents race conditions on stack modifications
- PRO: User visibility into changes before execution
- PRO: Audit trail of planned vs actual changes
- CON: Additional state management (plan lifecycle)
- CON: Lock contention under high concurrency

**Schema:**

```sql
CREATE TABLE stack_plans (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    plan_type VARCHAR(32) NOT NULL, -- 'create', 'update', 'destroy'

    -- Planned changes (JSON diff)
    changes JSONB NOT NULL,

    -- State tracking
    status VARCHAR(32) DEFAULT 'pending', -- 'pending', 'applied', 'failed', 'cancelled'
    created_by VARCHAR(128),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    applied_at TIMESTAMPTZ,
    error_message TEXT,

    -- Lock ownership
    lock_acquired BOOLEAN DEFAULT false,
    lock_acquired_at TIMESTAMPTZ,
    lock_released_at TIMESTAMPTZ
);

CREATE INDEX idx_plans_stack ON stack_plans(stack_id);
CREATE INDEX idx_plans_status ON stack_plans(status);
```

**Advisory Lock Pattern:**

```go
type PlanExecutor struct {
    db *sql.DB
}

// Advisory lock key derivation: hash stack_id to int64
func (e *PlanExecutor) lockKey(stackID int) int64 {
    // Use stack_id directly or hash for namespacing
    return int64(stackID)
}

// Plan: Generate diff without acquiring lock
func (e *PlanExecutor) Plan(stackID int, desiredState *StackConfig) (*Plan, error) {
    currentState := e.LoadCurrentState(stackID)
    changes := e.Diff(currentState, desiredState)

    plan := &Plan{
        StackID: stackID,
        Type:    "update",
        Changes: changes,
        Status:  "pending",
    }

    // Persist plan
    err := e.SavePlan(plan)
    return plan, err
}

// Apply: Acquire lock, execute, release
func (e *PlanExecutor) Apply(planID int) error {
    plan := e.LoadPlan(planID)

    // Start transaction
    tx, err := e.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Acquire advisory lock (blocking, with timeout)
    lockKey := e.lockKey(plan.StackID)
    acquired, err := e.tryLock(tx, lockKey, 30*time.Second)
    if err != nil {
        return fmt.Errorf("lock error: %w", err)
    }
    if !acquired {
        return fmt.Errorf("failed to acquire lock on stack %d", plan.StackID)
    }

    // Mark lock acquired
    plan.LockAcquired = true
    plan.LockAcquiredAt = time.Now()
    e.UpdatePlan(tx, plan)

    // Execute changes
    err = e.ExecuteChanges(tx, plan)
    if err != nil {
        plan.Status = "failed"
        plan.ErrorMessage = err.Error()
        e.UpdatePlan(tx, plan)
        return err
    }

    // Mark applied
    plan.Status = "applied"
    plan.AppliedAt = time.Now()
    e.UpdatePlan(tx, plan)

    // Commit releases advisory lock automatically
    return tx.Commit()
}

func (e *PlanExecutor) tryLock(tx *sql.Tx, key int64, timeout time.Duration) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // pg_try_advisory_xact_lock: released automatically on transaction end
    var acquired bool
    err := tx.QueryRowContext(ctx, "SELECT pg_try_advisory_xact_lock($1)", key).Scan(&acquired)
    return acquired, err
}

func (e *PlanExecutor) ExecuteChanges(tx *sql.Tx, plan *Plan) error {
    for _, change := range plan.Changes {
        switch change.Type {
        case "create_instance":
            err := e.CreateInstance(tx, change.Data)
            if err != nil {
                return fmt.Errorf("create instance: %w", err)
            }
        case "update_instance":
            err := e.UpdateInstance(tx, change.Data)
            if err != nil {
                return fmt.Errorf("update instance: %w", err)
            }
        case "delete_instance":
            err := e.DeleteInstance(tx, change.Data)
            if err != nil {
                return fmt.Errorf("delete instance: %w", err)
            }
        }
    }

    // Regenerate compose for stack
    return e.RegenerateCompose(tx, plan.StackID)
}
```

**Lock Guidelines:**

- Use `pg_try_advisory_xact_lock()` for transaction-scoped locks (auto-release)
- Use `pg_advisory_lock()` + `pg_advisory_unlock()` for session-scoped (manual release)
- Prefer transaction-scoped to prevent orphaned locks
- Set statement timeout to prevent indefinite blocking: `SET statement_timeout = '30s'`

### Pattern 5: Stack-Scoped vs Per-Service Compose

**What:** Generate single compose file for entire stack vs separate files per instance

**When to use:**
- Stack-scoped: Development, tightly coupled services, single docker-compose up
- Per-service: Production, independent deploy cycles, service isolation

**Trade-offs:**

| Aspect | Stack-Scoped | Per-Service |
|--------|--------------|-------------|
| Simplicity | Single file, easy orchestration | Multiple files, need coordination |
| Deploy granularity | All-or-nothing | Per-service rollout |
| Network isolation | Shared network | Can isolate networks |
| Dependency management | Built-in depends_on | Manual coordination |
| Compose reload | Restart entire stack | Restart single service |

**Implementation:**

```go
// Stack-scoped generator
func (g *StackGenerator) GenerateStack(stackID int) ([]byte, error) {
    instances := g.LoadInstances(stackID)
    stack := g.LoadStack(stackID)

    compose := &ComposeFile{
        Version:  "3.8",
        Networks: g.generateNetworks(stack),
        Volumes:  g.generateVolumes(instances),
        Services: make(map[string]*ServiceConfig),
    }

    // Resolve dependencies and generate startup order
    graph := g.BuildWiringGraph(stackID)
    sorted, err := graph.TopologicalSort()
    if err != nil {
        return nil, err
    }

    // Generate service configs in dependency order
    for _, node := range sorted {
        instance := instances[node.InstanceID]
        resolvedConfig := g.ResolveConfig(instance)

        svcConfig := g.GenerateServiceConfig(resolvedConfig)
        compose.Services[instance.InstanceName] = svcConfig
    }

    return yaml.Marshal(compose)
}

// Per-service generator (extend existing generator)
func (g *Generator) GenerateInstance(instanceID int) ([]byte, error) {
    instance := g.LoadInstance(instanceID)
    resolvedConfig := g.ResolveConfig(instance)

    // Generate single-service compose
    compose := &ComposeFile{
        Version:  "3.8",
        Networks: g.generateNetworks(instance.StackID),
        Services: map[string]*ServiceConfig{
            instance.InstanceName: g.GenerateServiceConfig(resolvedConfig),
        },
    }

    return yaml.Marshal(compose)
}
```

**Recommendation:** Support both modes with `stacks.compose_mode` column:
- `stack-scoped`: Default, generate one file
- `per-service`: Generate individual files in `compose/{stack_name}/{instance_name}.yml`

## Data Flow

### Stack Creation Flow

```
User: POST /api/stacks {name, description, instances}
    ↓
Handler: Validate stack definition
    ↓
Planner: Generate creation plan
    ↓ (writes to stack_plans table)
User: GET /api/plans/:id (review changes)
    ↓
User: POST /api/plans/:id/apply
    ↓
Executor: Acquire advisory lock (stack_id)
    ↓
Executor: Create stack + instances in transaction
    ↓
Resolver: Resolve all instance configs
    ↓
Wiring: Build dependency graph
    ↓
Generator: Generate compose YAML
    ↓
Materializer: Write config files to disk
    ↓
Executor: Commit transaction (releases lock)
    ↓
Response: Stack created, compose available
```

### Config Resolution Flow

```
Request: GET /api/instances/:id
    ↓
Handler: Load instance from DB
    ↓
Resolver: Load service template
    ↓
Resolver: Load overrides (ports, env, volumes)
    ↓
Resolver: Merge with precedence (overrides > template)
    ↓
Resolver: Apply JSONB compose_overrides
    ↓
Response: Resolved config JSON
```

### Wiring Resolution Flow

```
Trigger: Stack compose generation
    ↓
Graph Builder: Load all instances in stack
    ↓
Graph Builder: Load explicit wiring contracts
    ↓
Graph Builder: Apply auto-wiring rules
    ↓
Graph Validator: Detect cycles
    ↓
Graph Validator: Check port conflicts
    ↓
Topological Sort: Calculate startup order
    ↓
Compose Generator: Generate depends_on directives
    ↓
Compose Generator: Inject connection env vars
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-10 stacks | In-memory graph, synchronous plan/apply |
| 10-100 stacks | Connection pooling, advisory lock queue monitoring |
| 100+ stacks | Background job queue for plan/apply, caching resolved configs |

### Scaling Priorities

1. **First bottleneck:** Config resolution queries (many LEFT JOINs)
   - **Fix:** Cache resolved configs with cache invalidation on update
   - **Fix:** Pre-compute resolved configs on write (denormalize)

2. **Second bottleneck:** Lock contention on popular stacks
   - **Fix:** Queue plan applies instead of blocking HTTP requests
   - **Fix:** Split locks by operation type (read vs write)

3. **Third bottleneck:** Compose generation for large stacks (100+ instances)
   - **Fix:** Generate in background, stream results
   - **Fix:** Incremental regeneration (only changed instances)

## Anti-Patterns

### Anti-Pattern 1: Denormalizing Resolved Config

**What people do:** Store fully resolved config in `service_instances` table on every update

**Why it's wrong:**
- Breaks single source of truth
- Template changes don't propagate
- Staleness issues (resolved config out of sync with template)
- Storage bloat (duplicated data)

**Do this instead:**
- Resolve config at read time (query time or API layer)
- Cache resolved config with TTL + invalidation
- Use database views if query complexity becomes issue

### Anti-Pattern 2: Mixing Stack State with Instance State

**What people do:** Store `stack_id` in `services` table, making services stack-aware

**Why it's wrong:**
- Services are templates, shouldn't know about instances
- Can't reuse service across stacks
- Breaks copy-on-write pattern

**Do this instead:**
- Keep `services` table pure (templates only)
- All stack/instance relationships in `service_instances`
- Template → Instance is one-to-many relationship

### Anti-Pattern 3: Blocking HTTP Requests on Apply

**What people do:** Execute plan apply synchronously in HTTP handler

**Why it's wrong:**
- Long-running operations block connections
- No graceful timeout handling
- User sees gateway timeout, doesn't know if apply succeeded
- Advisory lock held for entire HTTP request lifecycle

**Do this instead:**
- POST /plans/:id/apply returns immediately with job ID
- Background worker executes apply
- GET /plans/:id polls for status
- WebSocket for real-time progress updates

### Anti-Pattern 4: Auto-Wiring Everything

**What people do:** Auto-wire all services based on type, no explicit contracts

**Why it's wrong:**
- Brittle (breaks when service types change)
- No visibility into actual dependencies
- Hard to debug connection issues
- Can't model complex wiring (multiple postgres instances)

**Do this instead:**
- Default to auto-wiring for simple cases (single DB per type)
- Require explicit wiring when multiple providers exist
- Validate wiring on plan (fail early)
- Document wiring contracts in API

### Anti-Pattern 5: Per-Instance Schema Tables

**What people do:** Create separate tables for each override type, growing schema explosively

**Why it's wrong:**
- Schema bloat (dozens of tables)
- Hard to query (JOIN hell)
- Difficult to add new override types

**Do this instead:**
- Use JSONB `compose_overrides` for arbitrary overrides
- Keep typed tables only for frequently queried fields (ports, env)
- Balance: type safety vs schema complexity

## Integration Points

### Docker/Podman API

| Operation | Integration Pattern | Notes |
|-----------|---------------------|-------|
| Start stack | Generate compose → docker-compose up | Use existing service-manager.sh wrapper |
| Stop stack | docker-compose down | Lock-free (read-only operation) |
| Reload service | Generate compose → docker-compose up -d service | Partial apply (single instance) |

### Secrets Management

| Phase | Pattern | Notes |
|-------|---------|-------|
| Import | Extract env vars with is_secret=true | Existing pattern in service_env_vars |
| Storage | Encrypted column or external vault | PostgreSQL pgcrypto or Vault integration |
| Resolution | Decrypt during config resolution | Inject into compose as plain env vars |
| Export | Redact secrets in compose export | Replace values with ***REDACTED*** |

**Recommendation:** Start with PostgreSQL `pgcrypto` for simplicity, add Vault support later if needed.

### Configuration Files

| Phase | Pattern | Notes |
|-------|---------|-------|
| Template | Store in service_config_files | Existing pattern |
| Override | Store in instance_config_file_overrides | New table, same structure |
| Materialization | Merge template + override → write to compose/{stack}/{instance}/ | Extend existing MaterializeConfigFiles |
| Encryption | Store encrypted, decrypt on materialization | For sensitive configs (TLS keys) |

## Build Order (Component Dependencies)

### Phase 1: Schema Foundation
Build order:
1. Stacks table + basic CRUD
2. Service instances table
3. Override tables (ports, env, volumes)

**Why first:** Foundation for all other components

### Phase 2: Config Resolution
Build order:
1. Application-level resolver (Go)
2. Resolved config API endpoint (GET /instances/:id/resolved)
3. Unit tests for merge precedence

**Why second:** Needed before compose generation works

### Phase 3: Compose Generation
Build order:
1. Extend existing generator with resolution hook
2. Stack-scoped generator
3. Materialization for instance config files

**Why third:** Depends on resolution, provides immediate value

### Phase 4: Wiring System
Build order:
1. Wiring contracts table + CRUD
2. Instance wiring table
3. Graph builder + topological sort
4. Auto-wiring rules

**Why fourth:** Complex, but decoupled from earlier phases

### Phase 5: Plan/Apply
Build order:
1. Plans table + basic CRUD
2. Diff algorithm (current vs desired)
3. Advisory lock wrapper
4. Apply executor

**Why fifth:** Most complex, depends on all previous components

### Phase 6: Advanced Features
Build order (any order):
- Secrets encryption
- Config file overrides
- WebSocket progress streaming
- Background job queue

**Why last:** Polish and performance, not MVP-critical

## Sources

- PostgreSQL Advisory Locks: https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS
- Go database/sql patterns: Standard library documentation
- Docker Compose Specification: https://docs.docker.com/compose/compose-file/
- Topological Sort: Standard graph algorithm (Kahn's, Tarjan's)
- Terraform Plan/Apply: https://www.terraform.io/docs/cli/commands/plan.html (pattern inspiration)

**Confidence Note:** Architecture patterns based on established software engineering practices (HIGH confidence). Specific Go implementation details based on standard library and existing codebase (HIGH confidence). Advisory lock patterns verified against PostgreSQL documentation (HIGH confidence).

---
*Architecture research for: Service Composition in Go API*
*Researched: 2026-02-03*
