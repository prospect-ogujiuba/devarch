# Technology Stack Research

**Project:** DevArch - Stacks/Instances/Wiring Milestone
**Domain:** Local microservices development orchestration
**Researched:** 2026-02-03
**Overall Confidence:** HIGH

## Executive Summary

DevArch follows a database-as-source-of-truth architecture for service composition. The upcoming stacks milestone adds service instance isolation, auto-wiring, and declarative configuration management. Stack recommendations prioritize Go stdlib where possible, minimal dependencies, and proven patterns from Docker Compose, Kubernetes, and Terraform ecosystems.

**Key principle:** Avoid premature abstraction. Use stdlib crypto, database/sql advisory locks, and gopkg.in/yaml.v3 (already in use). Add targeted libraries only when stdlib insufficient.

## Core Technologies (Already in Stack)

### Backend Foundation

| Technology | Version | Purpose | Rationale |
|------------|---------|---------|-----------|
| Go | 1.25+ | API server, compose generation | Latest stable. Go 1.25.6 currently in use. Fast compilation, excellent stdlib for systems programming |
| PostgreSQL | 14+ | Source of truth database | Already in use. Advisory locks native support, JSONB for flexible config storage, mature ecosystem |
| chi | v5.1.0 | HTTP router | Lightweight, stdlib-compatible, already in codebase |
| database/sql | stdlib | Database driver interface | Standard Go database abstraction |
| lib/pq | v1.10.9 | PostgreSQL driver | Currently in go.mod. Mature, stable driver |
| gopkg.in/yaml.v3 | v3.0.1 | YAML generation/parsing | Already used in compose/generator.go. Industry standard for YAML in Go |

### Frontend Foundation

| Technology | Version | Purpose | Rationale |
|------------|---------|---------|-----------|
| React | 19.2.0 | UI framework | Latest stable. Already in dashboard |
| Vite | 7.2.4 | Build tool | Fast HMR, excellent DX, currently in use |
| TanStack Router | 1.154.12 | Client-side routing | Type-safe, file-based routing, already in stack |
| TanStack Query | 5.90.19 | Server state management | De facto standard for data fetching/caching in React |
| Tailwind CSS | 4.0.0 | Styling | Utility-first, rapid development, already integrated |
| Radix UI | Various | Headless components | Accessible primitives, composable, currently in use |
| Zod | 4.3.6 | Schema validation | Type-safe validation, pairs well with TypeScript |

## New Libraries for Stacks Milestone

### Compose YAML Generation

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| gopkg.in/yaml.v3 | v3.0.1 (existing) | YAML marshal/unmarshal | HIGH |

**Recommendation:** Continue using gopkg.in/yaml.v3. Already in codebase, handles complex structs well, supports custom marshalers.

**Pattern:**
```go
type ComposeFile struct {
    Services map[string]ServiceConfig `yaml:"services"`
    Networks map[string]NetworkConfig `yaml:"networks"`
}

yaml.Marshal(&composeFile)
```

**Why not alternatives:**
- `yaml.v2`: Deprecated, v3 is the current standard
- `goyaml.v2`: Unmaintained fork
- Custom YAML writers: Reinventing the wheel, error-prone

### Advisory Locking in Go + PostgreSQL

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| database/sql | stdlib | Execute advisory lock queries | HIGH |
| lib/pq | v1.10.9 (existing) | PostgreSQL driver with full advisory lock support | HIGH |

**Recommendation:** Use PostgreSQL advisory locks via database/sql with raw SQL. No additional library needed.

**Pattern:**
```go
// Session-level advisory lock (auto-released on connection close)
_, err := db.Exec("SELECT pg_advisory_lock($1)", stackID)

// Try lock (non-blocking)
var locked bool
err := db.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&locked)

// Transaction-level lock (auto-released on commit/rollback)
_, err := tx.Exec("SELECT pg_advisory_xact_lock($1)", stackID)

// Release session lock
_, err := db.Exec("SELECT pg_advisory_unlock($1)", stackID)
```

**Why this approach:**
- Native PostgreSQL feature (since 8.2)
- No external dependencies
- Automatic cleanup on connection/transaction end
- Lock IDs are int64, perfect for service/stack IDs

**Advisory lock use cases in stacks:**
- Prevent concurrent plan/apply on same stack
- Serialize stack state mutations
- Coordinate multi-step operations (create stack → wire services → apply)

**Alternatives considered:**
| Approach | Why Not |
|----------|---------|
| File-based locking | Doesn't work in distributed/containerized environments |
| Redis locks (redlock) | Adds dependency, overkill for single-DB architecture |
| etcd/Consul | Heavy infrastructure, DevArch targets local dev |

**Confidence:** HIGH - Advisory locks are battle-tested in Postgres. Pattern used in Rails, Django, Laravel for background job coordination.

### AES-256-GCM Encryption

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| crypto/aes | stdlib | AES block cipher | HIGH |
| crypto/cipher | stdlib | GCM mode wrapper | HIGH |
| crypto/rand | stdlib | Cryptographically secure random | HIGH |

**Recommendation:** Use Go standard library crypto packages. No external dependencies needed for AES-256-GCM.

**Pattern:**
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
)

// Encrypt
func Encrypt(plaintext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key) // key must be 32 bytes for AES-256
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }

    // Prepend nonce to ciphertext (standard pattern)
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}

// Decrypt
func Decrypt(ciphertext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

**Key management pattern:**
```go
// Store 32-byte key in environment variable (base64 encoded)
// Generate once: openssl rand -base64 32

import "encoding/base64"

keyStr := os.Getenv("DEVARCH_ENCRYPTION_KEY")
key, err := base64.StdEncoding.DecodeString(keyStr)
if len(key) != 32 {
    return errors.New("encryption key must be 32 bytes for AES-256")
}
```

**Why stdlib:**
- FIPS 140-2 validated implementations available
- Audited by Go security team
- No supply chain risk from third-party crypto
- GCM provides authenticated encryption (prevents tampering)

**Alternatives considered:**
| Library | Why Not |
|---------|---------|
| github.com/FiloSottile/age | Good for file encryption, overkill for DB field encryption |
| github.com/gtank/cryptopasta | Deprecated, recommends using stdlib directly |
| golang.org/x/crypto/nacl/secretbox | XSalsa20-Poly1305 is good, but AES-256-GCM more widely recognized for compliance |

**Confidence:** HIGH - stdlib crypto is the Go-recommended approach for AES-GCM. Pattern from Go crypto documentation.

**Storage pattern for encrypted secrets:**
```sql
CREATE TABLE stack_secrets (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    key VARCHAR(256) NOT NULL,
    encrypted_value BYTEA NOT NULL,  -- Store ciphertext as BYTEA
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(stack_id, key)
);
```

### Copy-on-Write Config Override Patterns

**Recommendation:** Use PostgreSQL JSONB inheritance pattern + application-level merge logic. No library needed.

**Database schema pattern:**
```sql
-- Base service definition
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    base_config JSONB NOT NULL,  -- Canonical service config
    ...
);

-- Stack instances with config overrides
CREATE TABLE stack_instances (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    service_id INTEGER REFERENCES services(id),
    instance_name VARCHAR(256) NOT NULL,  -- e.g., "postgres-app1"
    config_overrides JSONB DEFAULT '{}',  -- Only changed fields
    ...
);
```

**Application-level merge pattern:**
```go
import "encoding/json"

// Merge base config with instance overrides (deep merge)
func MergeConfig(base, overrides json.RawMessage) (json.RawMessage, error) {
    var baseMap, overrideMap map[string]interface{}

    json.Unmarshal(base, &baseMap)
    json.Unmarshal(overrides, &overrideMap)

    merged := deepMerge(baseMap, overrideMap)
    return json.Marshal(merged)
}

func deepMerge(base, override map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})

    // Copy base
    for k, v := range base {
        result[k] = v
    }

    // Apply overrides (recursively for nested maps)
    for k, v := range override {
        if existing, ok := result[k]; ok {
            if existingMap, ok := existing.(map[string]interface{}); ok {
                if overrideMap, ok := v.(map[string]interface{}); ok {
                    result[k] = deepMerge(existingMap, overrideMap)
                    continue
                }
            }
        }
        result[k] = v
    }

    return result
}
```

**Optimized alternative for simple cases:**
Use PostgreSQL `jsonb_set()` or `||` operator for shallow merges:
```sql
-- Shallow merge at query time
SELECT
    s.name,
    s.base_config || si.config_overrides AS effective_config
FROM services s
JOIN stack_instances si ON si.service_id = s.id
WHERE si.stack_id = $1;
```

**Why this pattern:**
- Copy-on-write semantics: Base config never mutates
- Storage efficient: Only store deltas in stack_instances
- Version control friendly: Can track override changes independently
- Familiar to Docker Compose users (base + override files pattern)

**Inspiration from ecosystem:**
- Docker Compose: `docker-compose.yml` + `docker-compose.override.yml`
- Terraform: `terraform.tfvars` overrides
- Kubernetes: Kustomize overlays

**Alternative libraries considered:**
| Library | Why Not |
|---------|---------|
| github.com/imdario/mergo | Good, but 100 lines of Go for JSON merge is straightforward |
| github.com/jeremywohl/flatten | For flattening JSON; not needed for override pattern |
| JSONB operators in Postgres | Good for simple merges, limited for complex nested config |

**Confidence:** HIGH - JSONB + shallow application merge is standard practice for config management in Go APIs.

## Service Discovery & Wiring

**Context:** Stacks feature needs auto-wiring (service A discovers service B's connection details within stack).

| Approach | Recommendation | Confidence |
|----------|----------------|------------|
| Environment variable injection | PRIMARY | HIGH |
| DNS-based discovery | NOT NEEDED | HIGH |
| Service mesh | OVERKILL | HIGH |

**Recommended pattern:**
```go
// Generate stack-scoped env vars for wiring
func GenerateWiring(stack *Stack, instances []Instance) map[string]string {
    wiring := make(map[string]string)

    for _, instance := range instances {
        prefix := strings.ToUpper(instance.ServiceName)

        // Standard connection vars
        wiring[prefix+"_HOST"] = instance.ContainerName
        wiring[prefix+"_PORT"] = strconv.Itoa(instance.Port)

        // Service-specific (e.g., for databases)
        if instance.Type == "database" {
            wiring[prefix+"_URL"] = fmt.Sprintf("postgres://%s:%d/db",
                instance.ContainerName, instance.Port)
        }
    }

    return wiring
}
```

**Why environment variables:**
- Twelve-Factor App standard
- Docker Compose native pattern
- Works across all runtimes (PHP, Node, Go, Python)
- No service discovery daemon needed

**DNS in Docker networks:**
Docker already provides DNS (service name → IP resolution). Wiring layer adds semantic meaning:
- `POSTGRES_HOST=myapp-db` (semantic) vs `ping myapp-db` (low-level DNS)

**Confidence:** HIGH - Environment variable injection is the industry standard for local dev orchestration.

## Plan/Apply Workflow Pattern

**Context:** Terraform-style workflow for stack changes.

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| Standard Go | stdlib | State diffing logic | HIGH |
| encoding/json | stdlib | Serialize plan for preview | HIGH |

**Recommendation:** Implement plan/apply as application logic, not library dependency.

**Pattern:**
```go
type Plan struct {
    StackID     int
    ToCreate    []Instance
    ToUpdate    []InstanceDiff
    ToDelete    []Instance
    ToWire      []WiringChange
}

type InstanceDiff struct {
    Instance    Instance
    ConfigDiff  map[string]ConfigChange
}

type ConfigChange struct {
    Field    string
    OldValue interface{}
    NewValue interface{}
}

// Generate plan (read-only, safe to run repeatedly)
func GeneratePlan(stack *Stack, desired []Instance) (*Plan, error) {
    current := loadCurrentInstances(stack.ID)
    plan := &Plan{StackID: stack.ID}

    // Diff logic
    plan.ToCreate = diffCreate(current, desired)
    plan.ToUpdate = diffUpdate(current, desired)
    plan.ToDelete = diffDelete(current, desired)
    plan.ToWire = diffWiring(current, desired)

    return plan, nil
}

// Apply plan (mutates state, uses advisory lock)
func ApplyPlan(db *sql.DB, plan *Plan) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Advisory lock for stack (transaction-scoped)
    _, err = tx.Exec("SELECT pg_advisory_xact_lock($1)", plan.StackID)
    if err != nil {
        return err
    }

    // Apply changes
    for _, instance := range plan.ToCreate {
        if err := createInstance(tx, instance); err != nil {
            return err
        }
    }
    // ... updates, deletes, wiring

    return tx.Commit()
}
```

**Why not Terraform libraries:**
- Terraform SDK is for provider development, not state diffing
- State diffing is 200 lines of Go for our use case
- Avoids HashiCorp dependency

**Confidence:** MEDIUM - Pattern is well-established (Terraform, Pulumi, CloudFormation), but implementation details matter. Phase-specific research recommended for diff algorithm edge cases.

## Declarative Export/Import

**Context:** Stacks should be exportable/importable as YAML (like `docker-compose.yml`).

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| gopkg.in/yaml.v3 | v3.0.1 | YAML marshal/unmarshal | HIGH |

**Recommendation:** Reuse existing YAML pipeline from compose generation.

**Export pattern:**
```go
type StackExport struct {
    Name      string                    `yaml:"name"`
    Instances map[string]InstanceConfig `yaml:"instances"`
    Wiring    map[string]string         `yaml:"wiring,omitempty"`
    Secrets   []string                  `yaml:"secrets,omitempty"` // Keys only, not values
}

func ExportStack(stack *Stack) ([]byte, error) {
    export := buildExport(stack)
    return yaml.Marshal(export)
}
```

**Import pattern:**
```go
func ImportStack(data []byte) (*Stack, error) {
    var export StackExport
    if err := yaml.Unmarshal(data, &export); err != nil {
        return nil, err
    }

    // Validate
    if err := validateExport(&export); err != nil {
        return nil, err
    }

    // Generate plan from export
    plan := exportToPlan(&export)

    return plan, nil
}
```

**Confidence:** HIGH - YAML import/export is straightforward with gopkg.in/yaml.v3. Already proven in codebase for compose files.

## Database Migrations

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| Raw SQL migrations | N/A | Schema versioning | HIGH |

**Recommendation:** Continue using numbered SQL migration files. No migration framework needed.

**Current pattern:** `/api/migrations/*.up.sql` and `*.down.sql` files.

**Why no migration framework:**
- Current approach is explicit and debuggable
- 12 migrations so far without issues
- Framework adds dependency for marginal benefit

**If framework needed later:**
- `github.com/golang-migrate/migrate` (most popular)
- `github.com/pressly/goose` (simpler API)

**Confidence:** HIGH - Raw SQL migrations work well at this scale.

## Testing Libraries

| Library | Version | Purpose | Confidence |
|---------|---------|---------|------------|
| testing | stdlib | Unit tests | HIGH |
| testcontainers-go | v0.27+ | Integration tests with real Postgres | MEDIUM |

**Recommendation for stacks milestone:**
- Unit tests: stdlib `testing` package
- Integration tests: Consider `testcontainers-go` for testing advisory locks, JSONB merges with real Postgres

**Pattern:**
```go
import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestAdvisoryLocks(t *testing.T) {
    ctx := context.Background()

    // Spin up real Postgres
    pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:15-alpine",
            Env: map[string]string{"POSTGRES_PASSWORD": "test"},
            ExposedPorts: []string{"5432/tcp"},
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    defer pgContainer.Terminate(ctx)

    // Test advisory lock behavior
    db := connectToContainer(pgContainer)
    // ... test concurrent lock acquisition
}
```

**Confidence:** MEDIUM - testcontainers-go is industry standard, but adds complexity. Defer until integration tests needed.

## What NOT to Use

| Technology | Why Avoid | Use Instead |
|------------|-----------|-------------|
| ORM (GORM, ent) | Abstracts away SQL, harder to use advisory locks, JSONB ops | database/sql with sqlc or raw queries |
| Docker SDK for Go | Heavy, API changes frequently | Shell out to `docker compose` CLI (current approach works) |
| etcd/Consul | Designed for distributed systems, overkill for local dev | PostgreSQL advisory locks |
| Custom YAML parsers | Error-prone, reinventing wheel | gopkg.in/yaml.v3 |
| JWT for encryption | JWTs are for authentication, not data encryption | crypto/aes + crypto/cipher GCM mode |
| Redis for state | Adds dependency, state already in Postgres | PostgreSQL JSONB for flexibility |
| GraphQL | Adds complexity, REST API sufficient for dashboard | Continue with REST (chi router) |

## Version Compatibility Matrix

| Go Version | PostgreSQL | lib/pq | gopkg.in/yaml.v3 | Notes |
|------------|------------|--------|------------------|-------|
| 1.25.6 (current) | 14+ | v1.10.9 | v3.0.1 | Fully compatible, tested |
| 1.24+ | 14+ | v1.10.9 | v3.0.1 | Compatible (Go 1.24 introduced structured logging) |
| 1.23+ | 12+ | v1.10.8+ | v3.0.1 | Compatible (older Postgres versions lack some JSONB operators) |

**Frontend compatibility:**
- React 19.2.0 requires Node.js 18.17.0+
- Vite 7.2.4 requires Node.js 18+
- TanStack packages are interdependent (router + query should be close in versions)

## Installation Commands

### Backend
```bash
# Core dependencies (already in go.mod)
go get github.com/go-chi/chi/v5@v5.1.0
go get github.com/lib/pq@v1.10.9
go get gopkg.in/yaml.v3@v3.0.1

# No new dependencies needed for stacks milestone
# stdlib handles crypto, advisory locks, JSON merge
```

### Frontend (already installed)
```bash
cd dashboard
npm install @tanstack/react-query@^5.90.19
npm install @tanstack/react-router@^1.154.12
npm install axios@^1.13.2
npm install zod@^4.3.6
```

## Sources

**Confidence levels:**
- HIGH: Based on current codebase, Go stdlib documentation, PostgreSQL official docs
- MEDIUM: Based on ecosystem best practices (Docker Compose patterns, Terraform workflows)
- LOW: Speculative or needing phase-specific validation

**Verification:**
- Go version: `go version` output (1.25.6)
- go.mod dependencies: `/home/fhcadmin/projects/devarch/api/go.mod`
- Dashboard dependencies: `/home/fhcadmin/projects/devarch/dashboard/package.json`
- Database schema: `/home/fhcadmin/projects/devarch/api/migrations/*.up.sql`
- Existing patterns: `/home/fhcadmin/projects/devarch/api/internal/compose/generator.go`

**Advisory locks:**
- PostgreSQL docs: https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS
- lib/pq support: Verified via database/sql interface (no special driver methods needed)

**AES-256-GCM:**
- Go crypto package docs: https://pkg.go.dev/crypto/cipher (GCM interface)
- Pattern from Go blog: https://go.dev/blog/go1.5-crypto (AES-GCM example)

**Copy-on-write config:**
- PostgreSQL JSONB operators: https://www.postgresql.org/docs/current/functions-json.html
- Docker Compose override pattern: https://docs.docker.com/compose/multiple-compose-files/extends/

---

**Research completed:** 2026-02-03
**Confidence:** HIGH for stdlib/existing stack, MEDIUM for new patterns (plan/apply diffing)
**Next steps:** Roadmap creation can proceed with stack decisions locked in
