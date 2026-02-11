# Phase 22: Identity Service & Naming Consolidation - Research

**Researched:** 2026-02-11
**Domain:** Go naming service / identifier management / string formatting consolidation
**Confidence:** HIGH

## Summary

Phase 22 consolidates scattered naming logic into a dedicated identity service. Current state: 14+ files contain `fmt.Sprintf("devarch-...")` calls for generating container names, network names, file names, and label keys. Naming rules exist in `internal/container/labels.go` and `internal/container/validation.go`, but actual name generation is duplicated across handlers, orchestration, export, lock, and project packages.

The goal is centralizing ALL naming logic (stack/instance/network/container/file naming + validation) in a single service. This isn't just refactoring—it's establishing identity as a bounded context. Names ARE identities in DevArch: `devarch-{stack}-{instance}` is how containers are found, networks are referenced, lock files are written. Scattered naming creates collision risk and validation inconsistency.

Go ecosystem patterns: DDD entity identity uses Value Objects for identifiers, encapsulating generation and validation. Service layer approach: "fat service" with explicit dependencies, coordinating validation + generation without HTTP coupling. Kubernetes validation package provides DNS naming validators (`IsDNS1123Label`, `IsDNS1123Subdomain`) as reference standard.

**Primary recommendation:** Create `internal/identity` service with StackID, InstanceID, NetworkID, ContainerID, FileID types (Value Objects). Each type validates on creation, exposes String() for formatted output. Service coordinates naming rules, handles custom overrides (network_name, container_name), validates combinations (stack+instance length limits).

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| database/sql | stdlib | No DB needed for identity service | Identity is pure computation, validates input strings |
| regexp | stdlib | DNS pattern validation | Already used in container/validation.go |
| strings | stdlib | Name manipulation | Slugify, prefix/suffix operations |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/uuid | v1.6+ | Optional: generate random IDs | If adding auto-generated instance IDs (not in scope) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Value Objects | Plain strings | Lose compile-time safety, validation scattered |
| identity package | Keep in container | Conceptual mismatch: identity != container runtime |
| Service methods | Free functions | Harder to test, no dependency injection point |

**Installation:**
```bash
# No new dependencies required
# Existing code already has all validation patterns needed
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/
├── identity/                # NEW: naming and identity service
│   ├── service.go           # Identity service with naming methods
│   ├── types.go             # Value Objects: StackID, InstanceID, NetworkID, etc.
│   ├── validation.go        # DNS validation, reserved names, length checks
│   └── service_test.go      # Validation + generation test suite
├── container/
│   ├── labels.go            # REMOVE: ContainerName, NetworkName (move to identity)
│   ├── validation.go        # REMOVE: ValidateName, ValidateContainerName (move to identity)
│   └── client.go            # Keep: runtime operations unchanged
├── orchestration/
│   └── service.go           # REFACTOR: inject identity service, use identity.NetworkName()
├── export/
│   └── exporter.go          # REFACTOR: use identity.NetworkName(), identity.FileName()
└── lock/
    └── generator.go         # REFACTOR: use identity.ExtractInstanceID()
```

### Pattern 1: Value Objects for Identifiers

**What:** Wrap string identifiers in types that validate on construction, expose String() for output

**When to use:** When identifiers have invariants (DNS-safe, length limits, reserved names) that must be enforced

**Example:**
```go
// internal/identity/types.go
package identity

import "fmt"

// StackID represents a validated stack identifier
type StackID struct {
    value string
}

// NewStackID creates and validates a stack ID
func NewStackID(name string) (*StackID, error) {
    if err := validateDNSName(name); err != nil {
        return nil, fmt.Errorf("invalid stack name: %w", err)
    }
    if err := checkReservedNames(name); err != nil {
        return nil, err
    }
    return &StackID{value: name}, nil
}

// String returns the validated stack name
func (s StackID) String() string {
    return s.value
}

// NetworkName returns the network name for this stack (devarch-{stack}-net)
func (s StackID) NetworkName() string {
    return fmt.Sprintf("devarch-%s-net", s.value)
}

// InstanceID represents a validated instance identifier within a stack
type InstanceID struct {
    stackID   StackID
    instance  string
}

// NewInstanceID validates instance name and stack+instance combination
func NewInstanceID(stackID StackID, instance string) (*InstanceID, error) {
    if err := validateDNSName(instance); err != nil {
        return nil, fmt.Errorf("invalid instance name: %w", err)
    }
    // Validate combined length for container naming
    fullName := fmt.Sprintf("devarch-%s-%s", stackID.String(), instance)
    if len(fullName) > 127 {
        return nil, fmt.Errorf("container name %q (%d chars) exceeds 127-char limit", fullName, len(fullName))
    }
    return &InstanceID{stackID: stackID, instance: instance}, nil
}

// ContainerName returns the container name for this instance
func (i InstanceID) ContainerName() string {
    return fmt.Sprintf("devarch-%s-%s", i.stackID.String(), i.instance)
}
```

### Pattern 2: Identity Service for Coordinated Validation

**What:** Service layer encapsulating naming rules, handling custom overrides (network_name, container_name from DB)

**When to use:** When naming involves business rules beyond simple validation (custom names, prefix conventions, file naming)

**Example:**
```go
// internal/identity/service.go
package identity

import (
    "database/sql"
    "fmt"
)

type Service struct {
    db *sql.DB
}

func NewService(db *sql.DB) *Service {
    return &Service{db: db}
}

// ResolveNetworkName resolves network name: custom override or computed default
func (s *Service) ResolveNetworkName(stackName string, customNetworkName *string) (string, error) {
    stackID, err := NewStackID(stackName)
    if err != nil {
        return "", err
    }

    if customNetworkName != nil && *customNetworkName != "" {
        // Validate custom network name
        if err := validateDNSName(*customNetworkName); err != nil {
            return "", fmt.Errorf("invalid custom network name: %w", err)
        }
        return *customNetworkName, nil
    }

    return stackID.NetworkName(), nil
}

// ResolveContainerName resolves container name: custom override or computed default
func (s *Service) ResolveContainerName(stackName, instanceName string, customContainerName *string) (string, error) {
    stackID, err := NewStackID(stackName)
    if err != nil {
        return "", err
    }

    if customContainerName != nil && *customContainerName != "" {
        // Custom names still validated but don't need stack prefix
        if err := validateDNSName(*customContainerName); err != nil {
            return "", fmt.Errorf("invalid custom container name: %w", err)
        }
        return *customContainerName, nil
    }

    instanceID, err := NewInstanceID(*stackID, instanceName)
    if err != nil {
        return "", err
    }

    return instanceID.ContainerName(), nil
}

// ExportFileName returns the export file name for a stack
func (s *Service) ExportFileName(stackName string) (string, error) {
    stackID, err := NewStackID(stackName)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s-devarch.yml", stackID.String()), nil
}

// LockFileName returns the lock file name for a stack
func (s *Service) LockFileName(stackName string) (string, error) {
    stackID, err := NewStackID(stackName)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s-devarch.lock", stackID.String()), nil
}

// ExtractInstanceName extracts instance ID from container name (reverse operation)
func (s *Service) ExtractInstanceName(stackName, containerName string) (string, error) {
    prefix := fmt.Sprintf("devarch-%s-", stackName)
    if !strings.HasPrefix(containerName, prefix) {
        return "", fmt.Errorf("container name %q does not match stack %q prefix", containerName, stackName)
    }
    return strings.TrimPrefix(containerName, prefix), nil
}
```

### Pattern 3: Label Constants Migration

**What:** Move label key constants from container package to identity package (labels are identity concerns)

**When to use:** When label keys define identity relationships (stack_id, instance_id, managed_by)

**Example:**
```go
// internal/identity/labels.go
package identity

const (
    LabelPrefix            = "devarch."
    LabelStackID           = "devarch.stack_id"
    LabelInstanceID        = "devarch.instance_id"
    LabelTemplateServiceID = "devarch.template_service_id"
    LabelManagedBy         = "devarch.managed_by"
    LabelVersion           = "devarch.version"
    ManagedByValue         = "devarch"
)

// BuildLabels returns label map for a container
func BuildLabels(stackName, instanceName, templateServiceID string) (map[string]string, error) {
    labels := map[string]string{
        LabelManagedBy: ManagedByValue,
        LabelVersion:   "1.0",
    }

    if stackName != "" {
        labels[LabelStackID] = stackName
    }
    if instanceName != "" {
        labels[LabelInstanceID] = instanceName
    }
    if templateServiceID != "" {
        labels[LabelTemplateServiceID] = templateServiceID
    }

    return labels, nil
}

// ValidateLabelKey ensures user-provided label keys don't use reserved prefix
func ValidateLabelKey(key string) error {
    if strings.HasPrefix(key, LabelPrefix) {
        return fmt.Errorf("label key %q cannot start with %q - this prefix is reserved for system labels", key, LabelPrefix)
    }
    return nil
}
```

### Anti-Patterns to Avoid

- **Scattered fmt.Sprintf calls:** Each location inventing own prefix format creates divergence risk
- **Validation without context:** Validating stack name without checking combined stack+instance length fails late
- **String parsing in business logic:** `strings.TrimPrefix("devarch-stack-", name)` duplicates knowledge of naming format
- **Mixed responsibilities:** Container package should handle runtime, not identity rules

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| DNS validation regex | Custom pattern | Kubernetes validation package patterns | Handles edge cases (leading/trailing hyphens, digit-only labels) |
| Name slugification | Ad-hoc string manipulation | Existing Slugify in container/validation.go | Already handles spaces, underscores, collapsing hyphens, truncation |
| Reserved name checks | List in multiple places | Single source in identity/validation.go | One list to maintain, guaranteed consistency |
| Label prefix validation | String checks in handlers | ValidateLabelKey in identity package | Centralized reserved prefix enforcement |

**Key insight:** Naming isn't just strings—it's how DevArch identifies resources across DB, runtime, and filesystem. Centralize before divergence creates collision bugs.

## Common Pitfalls

### Pitfall 1: Forgetting Custom Name Overrides

**What goes wrong:** Code assumes all networks use `devarch-{stack}-net` format, breaks when user specifies custom network_name

**Why it happens:** Default generation is common path, custom overrides are DB-nullable fields often forgotten

**How to avoid:** Identity service has explicit `Resolve*Name` methods accepting optional custom override, handlers always call these

**Warning signs:** `grep "devarch-.*-net" | grep -v ResolveNetworkName` finds direct formatting

### Pitfall 2: Validation Order—Individual Then Combined

**What goes wrong:** Validate stack name (passes), validate instance name (passes), combined container name exceeds 127 chars (fails at runtime)

**Why it happens:** DNS label limit is 63 chars per label, container name limit is 127 chars for full name including prefix

**How to avoid:** `NewInstanceID` constructor validates BOTH instance name individually AND combined stack+instance length

**Warning signs:** CreateInstance handler validates instance name but not combined length

### Pitfall 3: Extracting Instance Names from Container Names

**What goes wrong:** `strings.TrimPrefix` assumes naming format, breaks if format changes or custom names used

**Why it happens:** Lock generator, lifecycle handlers need reverse operation (container name → instance name)

**How to avoid:** Identity service provides `ExtractInstanceName` method encoding format knowledge once

**Warning signs:** Multiple locations doing string manipulation to extract IDs

### Pitfall 4: Label Key Validation Inconsistency

**What goes wrong:** Instance overrides handler validates `devarch.` prefix, other handlers miss it, users confused

**Why it happens:** Validation scattered across handlers, each implementing own check

**How to avoid:** `ValidateLabelKey` in identity package, handlers call it consistently

**Warning signs:** `grep "devarch\." | grep -v ValidateLabelKey` finds direct string checks

## Code Examples

Verified patterns from codebase analysis:

### Current Scattered Pattern (TO ELIMINATE)
```go
// api/internal/orchestration/service.go:201
netName := fmt.Sprintf("devarch-%s-net", stackName)

// api/internal/export/exporter.go:44
netName := fmt.Sprintf("devarch-%s-net", stackName)

// api/internal/lock/generator.go:45
netName := fmt.Sprintf("devarch-%s-net", stackName)

// api/internal/project/controller.go:51
info.networkName = fmt.Sprintf("devarch-%s-net", info.stackName)
```

### Target Consolidated Pattern
```go
// ALL locations become:
netName, err := identityService.ResolveNetworkName(stackName, customNetworkNameOrNil)
if err != nil {
    return nil, fmt.Errorf("resolve network name: %w", err)
}
```

### Current Container Name Pattern (TO ELIMINATE)
```go
// api/internal/container/labels.go:36
func ContainerName(stackID, instanceID string) string {
    return fmt.Sprintf("devarch-%s-%s", stackID, instanceID)
}

// Called from 5+ locations without validation
```

### Target Pattern with Validation
```go
// identity service validates on creation, handlers just call String()
instanceID, err := identity.NewInstanceID(stackID, instanceName)
if err != nil {
    return nil, err
}
containerName := instanceID.ContainerName()
```

### Current Label Building (MIGRATE TO IDENTITY)
```go
// api/internal/container/labels.go:16
func BuildLabels(stackID, instanceID, templateServiceID string) map[string]string {
    // Builds labels without validation
}

// Moves to internal/identity/labels.go with validation
```

### Current Extraction Pattern (TO CONSOLIDATE)
```go
// api/internal/lock/generator.go:84-89
func extractInstanceName(stackName, containerName string) string {
    prefix := fmt.Sprintf("devarch-%s-", stackName)
    if !strings.HasPrefix(containerName, prefix) {
        return ""
    }
    return strings.TrimPrefix(containerName, prefix)
}

// Becomes identity.Service.ExtractInstanceName with proper error handling
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Validation functions scattered in packages | Centralized validation in dedicated package | Kubernetes pkg/util/validation established pattern ~2015 | Single source of truth for naming rules |
| Plain strings for identifiers | Value Objects with validation | DDD patterns gained traction ~2004, Go adoption ~2015+ | Type safety, validation at construction |
| Free functions for formatting | Service layer with dependency injection | Service pattern mainstream in Go ~2017+ | Testable, mockable, coordinated business logic |
| Manual string parsing | Encapsulated reverse operations | API design principle: operations have inverses | Format knowledge in one place |

**Deprecated/outdated:**
- Validation spread across handlers: Modern Go consolidates validation in domain layer
- container package holding naming: Identity ≠ runtime concerns, separate packages since clean architecture popularization

## Open Questions

1. **Should identity service query DB for custom overrides directly?**
   - What we know: Current handlers query network_name/container_name from DB, pass to generators
   - What's unclear: Does identity service accept `*sql.DB` and query itself, or accept custom names as parameters?
   - Recommendation: Accept as parameters (transport-agnostic like orchestration service), handlers query and pass

2. **Do we need StackID/InstanceID as types or just service methods?**
   - What we know: Value Objects provide compile-time safety, validation on construction
   - What's unclear: Does added type ceremony justify benefits in small codebase?
   - Recommendation: Start with service methods only, refactor to types if validation bugs emerge (YAGNI principle)

3. **Should Slugify remain in container/validation.go or move to identity?**
   - What we know: Slugify converts arbitrary strings to DNS-safe names (used in ValidateName error messages)
   - What's unclear: Is slugification identity concern or utility function?
   - Recommendation: Move to identity/validation.go—it's about creating valid identifiers

## Sources

### Primary (HIGH confidence)
- Codebase analysis: 14 files with `fmt.Sprintf("devarch-...")` patterns
- Existing validation: `api/internal/container/validation.go` DNS patterns
- Existing label constants: `api/internal/container/labels.go`
- [Kubernetes validation package](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/validation) - DNS naming standards

### Secondary (MEDIUM confidence)
- [Domain-Driven Design in Go: Designing Entities](https://pkritiotis.io/ddd-entity-in-go/) - Value Object patterns
- [How To Implement Domain-Driven Design (DDD) in Golang](https://programmingpercy.tech/blog/how-to-domain-driven-design-ddd-golang/) - Entity identity
- [Domain Driven Design in Golang - Tactical Design](https://www.damianopetrungaro.com/posts/ddd-using-golang-tactical-design/) - Aggregate identity references
- [The 'fat service' pattern for Go web applications](https://www.alexedwards.net/blog/the-fat-service-pattern) - Service layer structure
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting) - fmt.Sprintf patterns

### Tertiary (LOW confidence)
- Web search results about naming service patterns—generic, not Go-specific
- General refactoring articles—principles apply but lack Go idioms

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Uses stdlib only, existing validation code proves patterns work
- Architecture: HIGH - DDD identity + service layer patterns well-established in Go
- Pitfalls: HIGH - Derived from analyzing 14+ locations with current ad-hoc naming
- Implementation complexity: MEDIUM - Straightforward refactoring but touches many files

**Research date:** 2026-02-11
**Valid until:** 2026-03-11 (30 days - stable domain, stdlib-based)

**Key decision for planning:** Value Objects (StackID/InstanceID types) vs. service methods only. Recommend service methods first (simpler), refactor to types if needed (YAGNI).
