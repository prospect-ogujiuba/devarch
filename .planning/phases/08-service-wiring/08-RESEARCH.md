# Phase 8: Service Wiring - Research

**Researched:** 2026-02-08
**Domain:** Service discovery, dependency resolution, contract-based wiring
**Confidence:** HIGH

## Summary

Service wiring for local dev tools is NOT about adopting production service mesh patterns (Consul, etcd, sidecars). It's about leveraging Docker Compose's built-in DNS (127.0.0.11) and internal networking, then adding a lightweight contract resolution layer on top.

The standard approach: store contract metadata in your database, resolve dependencies with topological sort, inject environment variables with the resolved hostnames (service names = DNS names), and let Docker's internal DNS handle the rest. No external libraries needed for service discovery itself.

**Primary recommendation:** Hand-roll the contract resolution logic using topological sort for dependency ordering. Use Docker Compose's native DNS for runtime discovery. Use existing Go standard library + lib/pq for persistence. Don't introduce DI frameworks (Wire/Dig/Fx) — this is declarative config from DB, not application dependency injection.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Contract Model:**
- Contracts are template-level only — defined on service templates, not overridable per instance
- Export contracts: `{name, type, port, protocol}` (e.g., `{name: "database", type: "postgres", port: 5432, protocol: "tcp"}`)
- Import contracts: `{name, type, required, env_vars}` where `type` is exact match against exports
- Import contracts have a `required` boolean — required imports generate plan warnings if unwired, optional imports do not

**Auto-Wire Behavior:**
- Auto-wiring runs at plan generation time (not instance creation)
- Auto-wire is always on, no per-stack toggle
- One matching provider → auto-wire. Multiple matches → leave unwired with plan warning
- Users can disconnect any wire (auto or explicit)
- Explicit wires always override auto-wires for the same import contract

**Env Var Injection:**
- Import contract defines env var names (e.g., Laravel's import declares DB_HOST, DB_PORT, DB_DATABASE)
- Export provides values: hostname via internal DNS (`devarch-{stack}-{instance}`), container port (not host port)
- Merge priority: template env vars → wire-injected env vars → instance env var overrides (instance always wins per WIRE-08)

**Wiring Diagnostics & Plan Integration:**
- Missing required contracts = warning, not blocker
- Wires shown as dedicated section in plan output
- Ambiguous contracts shown as warnings with context
- Orphaned wires (referencing deleted instances) cleaned up automatically

**Wiring Persistence & Export:**
- All active wires stored in `service_instance_wires` table — both auto-resolved and explicit
- `source` column distinguishes `auto` vs `explicit` (informational)
- Auto-wiring is a resolution step that writes to DB, not a runtime computation
- devarch.yml export includes all wires

**Dashboard Wiring UX:**
- New "Wiring" tab on stack detail page (alongside Instances, Compose, Deploy)
- Table of all wires: consumer instance → provider instance, contract name, source badge
- Unresolved contracts show badge + dropdown to select provider
- Disconnect button on any wire
- Create explicit wire via "Add wire" action

### Claude's Discretion
- DB schema details for `service_exports`, `service_import_contracts`, `service_instance_wires`
- Auto-wire algorithm implementation
- Wire resolution ordering (alphabetical, creation order, etc. for determinism)
- Exact plan output formatting for wiring section
- Dashboard component structure and styling for wiring tab

### Deferred Ideas (OUT OF SCOPE)
- AWIR-01: Role-based auto-wire priority (`devarch.role=primary`) — v2
- AWIR-02: Wiring graph visualization in dashboard — v2
- AWIR-03: Custom env var naming templates for wire injection — v2
- Cross-stack wiring — v2 (MULT-02)

</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Docker Compose | 2.22+ | Container orchestration, internal DNS | Built-in service discovery via DNS, healthcheck conditions, native to local dev |
| lib/pq | current | PostgreSQL driver | Already in use, stores contract metadata and wiring state |
| Go stdlib | 1.22+ | Topological sort, dependency graph | No external DI libraries needed for declarative resolution |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| gopkg.in/yaml.v3 | current | YAML generation | Already in use for compose generation |
| golang-migrate | current | DB migrations | Already in use for schema changes (migration 015) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled topological sort | Third-party graph library | Overkill — topological sort is ~30 lines, no complex graph algorithms needed |
| Docker Compose DNS | Consul/etcd service registry | Massive complexity increase for zero benefit in local dev context |
| DB-stored wires | Runtime resolution | Loses export/import determinism (EXIM-05), breaks plan review safety |
| Manual DI (current) | uber-go/dig or google/wire | Wrong abstraction — this is declarative config resolution, not application DI |

**Installation:**
```bash
# No new dependencies needed
# Already using: lib/pq, yaml.v3, golang-migrate
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/wiring/
├── resolver.go          # Auto-wire resolution logic (topological sort)
├── contract_matcher.go  # Type-based contract matching
├── env_injector.go      # Env var injection from wires
└── validator.go         # Missing/ambiguous contract detection

api/migrations/
└── 015_service_wiring.up.sql  # Contract + wire tables (MIGR-03)

dashboard/src/routes/stacks/$stackId/
└── wiring.tsx           # Wiring tab component
```

### Pattern 1: Docker Compose Internal DNS for Service Discovery

**What:** Docker Compose automatically creates an embedded DNS server (127.0.0.11) that resolves service names to container IPs. Containers on the same network use service names as hostnames.

**When to use:** ALL container-to-container communication in local dev. This is the standard, not an option.

**Example:**
```yaml
# docker-compose.yml
services:
  laravel:
    environment:
      DB_HOST: postgres  # Resolves via internal DNS to postgres container IP
      DB_PORT: 5432      # Container port, not host port
  postgres:
    container_name: devarch-default-db
    ports:
      - "5433:5432"  # Host 5433 for external access, containers use 5432
```

**Sources:**
- [Docker Compose Networking | Docker Docs](https://docs.docker.com/compose/how-tos/networking/)
- [Docker Compose Tip #6: Service discovery and internal DNS](https://lours.me/posts/compose-tip-006-service-discovery/)
- [Container Hostnames and DNS with Docker Compose | Baeldung](https://www.baeldung.com/ops/docker-compose-hostnames-dns)

**DevArch-specific:** Container names follow `devarch-{stack}-{instance}` pattern (Phase 4). Wire-injected env vars use these names for DB_HOST, REDIS_HOST, etc.

### Pattern 2: Topological Sort for Dependency Resolution

**What:** DAG traversal algorithm that orders nodes so every dependency is processed before its dependents. Used by Docker Compose, build systems (Bazel, Maven), and compiler dependency resolution.

**When to use:** Auto-wire resolution — process providers before consumers to detect ambiguous matches and circular dependencies.

**Example:**
```go
// Kahn's Algorithm (BFS-based) — recommended for DevArch
func ResolveWiringOrder(instances []Instance, imports []Import) ([]WireCandidate, error) {
    // 1. Build in-degree map (count of unmet dependencies per instance)
    inDegree := make(map[int]int)
    adjList := make(map[int][]int) // provider -> consumers

    // 2. Initialize queue with instances that have no unmet imports
    queue := []int{}
    for _, inst := range instances {
        if inDegree[inst.ID] == 0 {
            queue = append(queue, inst.ID)
        }
    }

    // 3. Process queue, wiring one provider at a time
    var wired []WireCandidate
    for len(queue) > 0 {
        providerID := queue[0]
        queue = queue[1:]

        // Match exports to imports, decrement in-degree for consumers
        for _, consumerID := range adjList[providerID] {
            inDegree[consumerID]--
            if inDegree[consumerID] == 0 {
                queue = append(queue, consumerID)
            }
        }
    }

    // 4. If any instance still has in-degree > 0, circular dependency exists
    for id, degree := range inDegree {
        if degree > 0 {
            return nil, fmt.Errorf("circular dependency detected for instance %d", id)
        }
    }

    return wired, nil
}
```

**Sources:**
- [Topological Sorting Explained: Dependency Resolution | Medium](https://medium.com/@amit.anjani89/topological-sorting-explained-a-step-by-step-guide-for-dependency-resolution-1a6af382b065)
- [Service Startup and Dependency Management | docker/compose](https://deepwiki.com/docker/compose/3.3.1-service-startup-and-dependency-management)
- [Understanding Topological Sort | Oreate AI](https://www.oreateai.com/blog/understanding-topological-sort-a-key-to-dependency-resolution/fdd7415bc1b745d1fcbea02c710cb09b)

**DevArch-specific:** Run topological sort during plan generation. Output is list of `(consumer_instance_id, provider_instance_id, contract_name)` tuples. Store in `service_instance_wires` table for plan review and apply execution.

### Pattern 3: Environment Variable Merge Priority

**What:** Docker Compose resolves env vars from multiple sources with a defined precedence order. Later sources override earlier ones.

**When to use:** Always. Critical for wire-injected vars to merge correctly with template defaults and instance overrides.

**Precedence (lowest to highest):**
1. Template-level env vars (from `service_env_vars` table)
2. Wire-injected env vars (from wiring resolution + contract `env_vars` mapping)
3. Instance-level env var overrides (from `instance_env_vars` table)

**Example:**
```go
func MergeEnvVars(template []EnvVar, wired map[string]string, overrides []EnvVar) map[string]string {
    result := make(map[string]string)

    // 1. Template defaults
    for _, ev := range template {
        result[ev.Key] = ev.Value
    }

    // 2. Wire-injected (e.g., DB_HOST from postgres export)
    for k, v := range wired {
        result[k] = v
    }

    // 3. Instance overrides (user explicitly set)
    for _, ev := range overrides {
        result[ev.Key] = ev.Value
    }

    return result
}
```

**Sources:**
- [Best practices | Docker Docs](https://docs.docker.com/compose/how-tos/environment-variables/best-practices/)
- [Set environment variables | Docker Docs](https://docs.docker.com/compose/how-tos/environment-variables/set-environment-variables/)

**DevArch-specific:** WIRE-08 requirement — instance env overrides MUST win. This lets users disconnect from auto-wired DB and point DB_HOST to external service via override.

### Pattern 4: Healthcheck with depends_on Conditions

**What:** Docker Compose 2.1+ supports conditional dependencies — wait for service to be healthy, not just started.

**When to use:** Optional enhancement for DevArch Phase 8+. Not required for wiring MVP but natural extension of dependency metadata.

**Example:**
```yaml
services:
  laravel:
    depends_on:
      postgres:
        condition: service_healthy
  postgres:
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
```

**Sources:**
- [Docker Compose Health Checks | Last9](https://last9.io/blog/docker-compose-health-checks/)
- [How to Use Docker Compose depends_on with Health Checks](https://oneuptime.com/blog/post/2026-01-16-docker-compose-depends-on-healthcheck/view)
- [Forget wait-for-it, use healthcheck and depends_on](https://www.denhox.com/posts/forget-wait-for-it-use-docker-compose-healthcheck-and-depends-on-instead/)

**DevArch-specific:** Healthcheck configs already stored in template `service_healthchecks` and instance `instance_healthchecks` tables (migration 014). Could extend compose generator to add `depends_on: {condition: service_healthy}` when consumer wired to provider with healthcheck.

### Anti-Patterns to Avoid

**External Service Registries (Consul, etcd) for Local Dev:**
- Why it's bad: Massive complexity (run separate registry cluster, configure registrator sidecars, manage service registration lifecycle)
- What to do: Use Docker Compose's built-in DNS. It's free, automatic, and designed for this exact use case.

**Runtime Wire Resolution (not storing in DB):**
- Why it's bad: Breaks export/import determinism (EXIM-05), loses plan review safety (users can't see proposed wires before apply)
- What to do: Auto-wire writes to `service_instance_wires` table. Plan endpoint queries table and shows proposed changes.

**DI Frameworks for Config Resolution:**
- Why it's bad: uber-go/dig and google/wire solve application-level dependency injection (struct initialization order). This is declarative config resolution (which instance should wire to which).
- What to do: Hand-roll the resolver with topological sort. It's ~100 lines and has zero transitive dependencies.

**Hard-Coding DNS Names in Templates:**
- Why it's bad: Breaks multi-stack isolation (two stacks using same DB name collide)
- What to do: Wire-injected env vars use `devarch-{stack}-{instance}` pattern from Phase 4. Each stack gets unique container names.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Container DNS resolution | Custom DNS server | Docker Compose built-in DNS (127.0.0.11) | Docker's DNS handles service name → IP mapping, round-robin load balancing for scaled services, automatic updates when containers restart |
| Service dependency ordering in compose | Custom orchestrator | Docker Compose `depends_on` | Compose already does topological sort via `InDependencyOrder()` for service startup |
| Environment variable interpolation in compose | Custom templating | Docker Compose `${VAR}` syntax | Compose handles shell var substitution, .env file loading, precedence rules |
| Database migrations | Custom migration runner | golang-migrate | Already in use, battle-tested, handles up/down/version tracking |

**Key insight:** Docker Compose is NOT just a container runner — it's a complete local dev orchestration platform with DNS, networking, dependency ordering, healthchecks, and env var management built-in. The only custom logic needed is contract matching and wire resolution (which DB to inject for which consumer).

## Common Pitfalls

### Pitfall 1: Using Host Ports Instead of Container Ports for Wires

**What goes wrong:** Wire injector sets `DB_PORT=5433` (host port) instead of `DB_PORT=5432` (container port). Laravel container tries to connect to postgres on port 5433 and fails — postgres container is listening on 5432.

**Why it happens:** Confusion between external access (host machine → container via `ports: ["5433:5432"]`) and internal access (container → container via bridge network).

**How to avoid:** Wire-injected env vars ALWAYS use container ports from export contracts. Host ports are for external access only (developer's browser, external DB clients).

**Warning signs:**
- Connection refused errors when app tries to connect to DB
- Wire shows correct hostname but wrong port
- Connection works from host machine (using host port) but not from container

**DevArch-specific:** Export contract stores `port: 5432` (container port). This goes directly into `DB_PORT`. Host port (from `service_ports.host_port`) is ignored for wiring.

### Pitfall 2: Circular Dependencies Not Detected Until Apply

**What goes wrong:** User creates explicit wires: A imports from B, B imports from A. Plan succeeds, apply hangs because Docker Compose can't start either service (each depends_on the other).

**Why it happens:** Wire creation API doesn't validate for cycles. Cycle only detected when compose generator tries to build `depends_on` chains.

**How to avoid:** Run topological sort during plan generation. If sort fails (any instance has in-degree > 0 after processing), reject the plan with clear error: "Circular dependency: laravel → postgres → laravel".

**Warning signs:**
- `docker compose up` hangs during startup
- Logs show services waiting for each other
- No clear error message, just timeout

**DevArch-specific:** Plan endpoint must call `DetectCircularDependencies(wires)` BEFORE returning success. This is a blocker, not a warning.

### Pitfall 3: Ambiguous Contracts Not Surfaced in Plan

**What goes wrong:** Stack has 2 postgres instances (prod-db, analytics-db). Laravel auto-wires to prod-db alphabetically. User doesn't notice in plan output, deploys, laravel connects to wrong DB.

**Why it happens:** Auto-wire silently picks first match when multiple providers exist, instead of leaving unwired and warning.

**How to avoid:** When multiple providers match import contract type, DO NOT auto-wire. Leave contract unwired, add to plan warnings: "Ambiguous: 2 postgres providers for laravel.database — create explicit wire".

**Warning signs:**
- Plan shows no warnings despite multiple potential providers
- Services connect to unexpected instances after deploy
- No clear indication in UI which provider was selected

**DevArch-specific:** Auto-wire algorithm MUST have `if len(matchingProviders) > 1 { return nil, "ambiguous" }`. Ambiguous contracts appear in dedicated plan section, NOT silently resolved.

### Pitfall 4: Orphaned Wires After Instance Deletion

**What goes wrong:** User deletes postgres instance. Wires pointing to that instance remain in `service_instance_wires` table. Next plan shows laravel still wired to deleted DB.

**Why it happens:** Wire table has foreign key to `service_instances(id)`, but deletion doesn't cascade or clean up orphans.

**How to avoid:** Two options: (1) Cascade delete on foreign key, or (2) Filter deleted instances during plan generation. Option 2 is safer — keeps audit trail of what was wired before deletion.

**Warning signs:**
- Plan shows wires to non-existent instances
- Wire resolution throws "provider not found" errors
- Export includes references to deleted instances

**DevArch-specific:** Plan query should be `WHERE provider_instance_id IN (SELECT id FROM service_instances WHERE deleted_at IS NULL)`. Migration 014 added `deleted_at` soft delete pattern — wire queries must respect it.

### Pitfall 5: Template-Level Contract Changes Break Existing Wires

**What goes wrong:** User edits postgres template, changes export type from "postgres" to "postgresql". All existing wires stop matching because import still looks for "postgres".

**Why it happens:** Wires store `(consumer_id, provider_id, contract_name)` but don't store contract type. Type comes from template metadata, which changed.

**How to avoid:** Wire table should store snapshot of matched contract at wire creation time: `provider_contract_type`, `consumer_contract_type`. Breaking changes to templates don't auto-break wires, but plan shows mismatch warning.

**Warning signs:**
- Previously working wires suddenly fail to match
- Plan shows missing contracts despite wires existing
- No clear error about what changed

**DevArch-specific:** Consider adding `matched_export_type` and `matched_import_type` to `service_instance_wires` schema for audit trail. Alternative: Template contract changes are rare in practice — may be acceptable to just detect mismatch in plan and prompt user to recreate wire.

## Code Examples

### Example 1: Auto-Wire Resolution with Ambiguity Detection

```go
// Source: Pattern synthesis from Docker Compose dependency ordering + topological sort research
package wiring

type Contract struct {
    Name string
    Type string
}

type Provider struct {
    InstanceID   int
    ContractName string
    ContractType string
    Port         int
    Protocol     string
}

type Consumer struct {
    InstanceID   int
    ContractName string
    ContractType string
    Required     bool
    EnvVars      map[string]string // e.g., {"DB_HOST": "{{hostname}}", "DB_PORT": "{{port}}"}
}

type WireCandidate struct {
    ConsumerInstanceID int
    ProviderInstanceID int
    ContractName       string
    EnvVarInjections   map[string]string
    Source             string // "auto" or "explicit"
    AmbiguityReason    string // Non-empty if multiple matches
}

func ResolveAutoWires(providers []Provider, consumers []Consumer, existingWires []Wire) ([]WireCandidate, []string, error) {
    candidates := []WireCandidate{}
    warnings := []string{}

    for _, consumer := range consumers {
        // Skip if explicit wire already exists
        if hasExplicitWire(consumer, existingWires) {
            continue
        }

        // Find matching providers by contract type (exact match)
        matches := []Provider{}
        for _, p := range providers {
            if p.ContractType == consumer.ContractType {
                matches = append(matches, p)
            }
        }

        switch len(matches) {
        case 0:
            // Missing required contract
            if consumer.Required {
                warnings = append(warnings, fmt.Sprintf(
                    "Missing required contract: instance %d needs %s (type: %s)",
                    consumer.InstanceID, consumer.ContractName, consumer.ContractType))
            }
        case 1:
            // Unambiguous — auto-wire
            provider := matches[0]
            hostname := fmt.Sprintf("devarch-%s-%s", stackName, provider.InstanceName)

            envVars := make(map[string]string)
            for k, template := range consumer.EnvVars {
                v := strings.ReplaceAll(template, "{{hostname}}", hostname)
                v = strings.ReplaceAll(v, "{{port}}", strconv.Itoa(provider.Port))
                envVars[k] = v
            }

            candidates = append(candidates, WireCandidate{
                ConsumerInstanceID: consumer.InstanceID,
                ProviderInstanceID: provider.InstanceID,
                ContractName:       consumer.ContractName,
                EnvVarInjections:   envVars,
                Source:             "auto",
            })
        default:
            // Ambiguous — leave unwired
            warnings = append(warnings, fmt.Sprintf(
                "Ambiguous contract: %d providers match %s (type: %s) for instance %d — create explicit wire",
                len(matches), consumer.ContractName, consumer.ContractType, consumer.InstanceID))
        }
    }

    return candidates, warnings, nil
}

func hasExplicitWire(consumer Consumer, wires []Wire) bool {
    for _, w := range wires {
        if w.ConsumerInstanceID == consumer.InstanceID &&
           w.ContractName == consumer.ContractName &&
           w.Source == "explicit" {
            return true
        }
    }
    return false
}
```

### Example 2: Environment Variable Injection with Merge Priority

```go
// Source: Docker Compose env var best practices + DevArch merge priority (WIRE-08)
package compose

func BuildEnvVars(templateVars []ServiceEnvVar, wiredVars map[string]string, instanceVars []InstanceEnvVar) map[string]string {
    result := make(map[string]string)

    // Priority 1: Template defaults
    for _, ev := range templateVars {
        if ev.IsSecret {
            // Phase 7 pattern: secrets not exposed in compose, injected via runtime
            continue
        }
        result[ev.Key] = ev.Value
    }

    // Priority 2: Wire-injected (from resolved contracts)
    for k, v := range wiredVars {
        result[k] = v
    }

    // Priority 3: Instance overrides (highest priority per WIRE-08)
    for _, ev := range instanceVars {
        if ev.IsSecret {
            continue
        }
        result[ev.Key] = ev.Value
    }

    return result
}
```

### Example 3: Plan Output Format for Wiring Section

```go
// Source: Pattern from Phase 6 plan/apply + Docker Compose dependency diagnostics
package handlers

type PlanWiringSection struct {
    ActiveWires      []WirePlanEntry   `json:"active_wires"`
    ProposedChanges  []WireChange      `json:"proposed_changes"`
    Warnings         []WiringWarning   `json:"warnings"`
}

type WirePlanEntry struct {
    ConsumerInstance string `json:"consumer_instance"`
    ProviderInstance string `json:"provider_instance"`
    ContractName     string `json:"contract_name"`
    Source           string `json:"source"` // "auto" or "explicit"
    InjectedEnvVars  map[string]string `json:"injected_env_vars"`
}

type WireChange struct {
    Action           string `json:"action"` // "add", "remove", "update"
    ConsumerInstance string `json:"consumer_instance"`
    ProviderInstance string `json:"provider_instance"`
    ContractName     string `json:"contract_name"`
    Reason           string `json:"reason"` // e.g., "auto-wired (unambiguous match)"
}

type WiringWarning struct {
    Severity string `json:"severity"` // "warning" or "error"
    Message  string `json:"message"`
    Instance string `json:"instance"`
    Contract string `json:"contract,omitempty"`
}

// Example plan response
{
  "wiring": {
    "active_wires": [
      {
        "consumer_instance": "laravel",
        "provider_instance": "postgres",
        "contract_name": "database",
        "source": "auto",
        "injected_env_vars": {
          "DB_HOST": "devarch-default-postgres",
          "DB_PORT": "5432",
          "DB_DATABASE": "postgres"
        }
      }
    ],
    "proposed_changes": [
      {
        "action": "add",
        "consumer_instance": "laravel",
        "provider_instance": "redis",
        "contract_name": "cache",
        "reason": "auto-wired (unambiguous match)"
      }
    ],
    "warnings": [
      {
        "severity": "warning",
        "message": "Ambiguous contract: 2 postgres providers available — create explicit wire",
        "instance": "api",
        "contract": "database"
      }
    ]
  }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| wait-for-it.sh scripts | Docker Compose healthcheck + depends_on conditions | Compose 2.1+ (2020) | Built-in health-aware dependency management, no external scripts needed |
| External service registries (Consul) for local dev | Docker Compose internal DNS | Compose 1.10+ (2017) | Automatic service discovery via DNS names, zero config required |
| Manual env var wiring in .env files | Provider services pattern with dynamic injection | Compose 2.23+ (2023) | Dependent services receive auto-injected env vars from providers |
| Separate up/down scripts for startup order | depends_on with topological sort | Compose core feature | Guaranteed dependency-order startup/shutdown |
| Runtime DI frameworks in Go | Compile-time code generation (Wire) or manual DI | Wire v0.5+ (2023) | For app-level DI only, NOT config resolution |

**Deprecated/outdated:**
- wait-for-it.sh, dockerize, wait-for: Replaced by `depends_on: {condition: service_healthy}` in Compose 2.1+
- Docker Compose v1 (docker-compose): Replaced by Docker Compose v2 (docker compose, plugin-based) in 2021
- Hard-coded service names in app config: Replaced by env var injection pattern with dynamic values

**Current 2026 best practices:**
- Use Compose internal DNS for all service-to-service communication
- Healthchecks + depends_on conditions for reliable startup ordering
- Separate .env files per environment (dev, test, prod) for config management
- Never store secrets in env vars — use Docker secrets or external secret managers
- Reference containers by service name, not IP, for resilience to restarts

## Open Questions

1. **Should wire table store contract type snapshot?**
   - What we know: Wires currently store `(consumer_id, provider_id, contract_name)`. Contract type comes from template metadata.
   - What's unclear: If user changes postgres template export from type "postgres" to "postgresql", existing wires break (import type no longer matches).
   - Recommendation: Add `provider_contract_type` and `consumer_contract_type` to `service_instance_wires` for audit trail. Plan can detect mismatch and warn. LOW priority — template contract changes are rare.

2. **Should depends_on be generated from wires?**
   - What we know: Docker Compose `depends_on` controls startup order. Wires define logical dependencies (laravel needs postgres).
   - What's unclear: Should compose generator add `depends_on: [postgres]` for laravel based on wire existence? Or is wiring purely env var injection?
   - Recommendation: YES, generate depends_on from wires. Benefits: (1) correct startup order, (2) cleaner `docker compose logs` output (postgres starts first), (3) enables healthcheck conditions for v2. Implementation: compose generator adds `depends_on: [provider_container_name]` for each wire.

3. **Should env var templates support custom transformations?**
   - What we know: Import contract defines `env_vars: {"DB_HOST": "{{hostname}}", "DB_PORT": "{{port}}"}`. Resolver does string replacement.
   - What's unclear: Do users need transformations like `"DB_URL": "postgres://{{hostname}}:{{port}}/{{database}}"` or `"REDIS_URL": "redis://{{hostname}}:{{port}}/0"`?
   - Recommendation: Start with simple key-value mapping (DB_HOST, DB_PORT as separate vars). Connection string templates (DB_URL) are v2 feature (AWIR-03). Most frameworks accept separate host/port vars.

## Sources

### Primary (HIGH confidence)
- [Docker Compose Networking | Docker Docs](https://docs.docker.com/compose/how-tos/networking/) — Internal DNS, service discovery
- [Docker Compose Environment Variables Best Practices | Docker Docs](https://docs.docker.com/compose/how-tos/environment-variables/best-practices/) — Merge priority, security patterns
- [Define services in Docker Compose | Docker Docs](https://docs.docker.com/reference/compose-file/services/) — Service configuration reference
- [Control startup order | Docker Docs](https://docs.docker.com/compose/how-tos/startup-order/) — depends_on patterns
- [Topological Sorting Explained | Medium](https://medium.com/@amit.anjani89/topological-sorting-explained-a-step-by-step-guide-for-dependency-resolution-1a6af382b065) — Dependency resolution algorithm
- [golang-migrate/migrate | GitHub](https://github.com/golang-migrate/migrate) — Database migration patterns

### Secondary (MEDIUM confidence)
- [How to Implement Docker Compose Service Dependencies](https://oneuptime.com/blog/post/2026-01-30-how-to-implement-docker-compose-service-dependencies/view) — 2026 patterns
- [Docker Compose Health Checks | Last9](https://last9.io/blog/docker-compose-health-checks/) — Healthcheck integration
- [Dependency Injection in Go | DEV Community](https://dev.to/rezende79/dependency-injection-in-go-comparing-wire-dig-fx-more-3nkj) — DI framework comparison (why NOT to use for this)
- [Pact Contract Testing | Pact Docs](https://docs.pact.io/) — Consumer-driven contract pattern (inspiration, not direct use)

### Tertiary (LOW confidence)
- Service mesh patterns (Istio, Consul) — NOT applicable to local dev, included for contrast only

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — Docker Compose DNS and depends_on are documented, battle-tested features
- Architecture: HIGH — Topological sort is well-established algorithm, env var merge priority is in official docs
- Pitfalls: MEDIUM — Based on synthesis of Docker Compose gotchas + DevArch's multi-stack context, not direct experience with DevArch Phase 8

**Research date:** 2026-02-08
**Valid until:** 2026-04-08 (60 days — Docker Compose and Go patterns are stable)
