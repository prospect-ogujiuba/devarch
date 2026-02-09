# Phase 9: Secrets & Resources - Research

**Researched:** 2026-02-08
**Domain:** AES-256-GCM encryption at rest, Docker Compose resource limits
**Confidence:** HIGH

## Summary

Phase 9 adds transparent encryption at rest for secret environment variables using Go's standard `crypto/aes` and `crypto/cipher` packages with AES-256-GCM mode, plus optional per-instance CPU/memory resource limits that map to Docker Compose `deploy.resources` fields.

**Primary recommendation:** Use Go standard library AES-256-GCM with random nonces prepended to ciphertext, store encrypted values in TEXT columns (not BYTEA), redact at API serialization layer, and map resource limits directly to Compose deploy section with validation warnings only.

**Key insight:** Encryption complexity lives in the nonce-prepend pattern, not the algorithm itself. GCM provides authenticated encryption with minimal code. The real work is transparent decrypt-on-SELECT/encrypt-on-INSERT and consistent redaction across all API outputs.

## User Constraints (from CONTEXT.md)

<user_constraints>
### Locked Decisions

**Encryption Key Management:**
- Auto-generate `~/.devarch/secret.key` on first encrypt operation
- No passphrase — server-side operations (compose gen, plan/apply) must be non-interactive
- Key loss requires re-encrypt migration (acceptable for local dev tool)

**Encryption Scope:**
- Only `value` column in rows where `is_secret = true`
- Applies to both `service_env_vars` and `instance_env_vars` tables
- Config file contents are NOT encrypted (keyword heuristic in `export/secrets.go` remains as safety net for export redaction)

**Redaction Enforcement:**
- Redact at API serialization layer (handler responses), not at DB layer
- Decrypt on SELECT (SECR-04), then handler decides: mask for read endpoints, pass real value for edit endpoints
- Covered endpoints: GET effective config, plan diff output, compose preview
- Export already handled by `export/secrets.go` (no change)
- WebSocket status doesn't carry env vars (no change)

**Redaction Format:**
- API JSON responses: `"***"` for masked secret values
- Dashboard display: `••••••••` (8 bullets) — standardize across all components
- Reveal toggle (eye icon) stays as-is in edit mode
- Resolves current inconsistency: `********` / `••••••••` / `***` all become `••••••••` in dashboard

**Resource Limit Granularity:**
- CPU (decimal cores, e.g. `0.5`, `2.0`) + Memory (with suffix, e.g. `512m`, `1g`) only
- Maps directly to compose `deploy.resources.limits` and `deploy.resources.reservations`
- No GPU, no disk limits (scope creep for local dev tool)

**Resource Limit Behavior:**
- Limits are optional — omitting means no `deploy.resources` in compose output
- Instance-level only, not template-level (limits are environment-specific: my laptop ≠ yours)
- Matches Phase 3 override pattern: instance overrides are the customization layer

**Resource Limits in Plan Output:**
- Show limits as part of instance config in plan diff (value changes diffed like env vars)
- Missing limits = omit from compose (don't set to zero)
- Validation: warn if limits seem unreasonable (e.g. memory < 4m), never block

**Migration Strategy:**
- **Migration 020:** Add `encrypted_value` (nullable text) and `encryption_version` (integer, default 0) columns to both `service_env_vars` and `instance_env_vars`
- **Migration 021:** Create `instance_resource_limits` table (instance_id FK, cpu_limit, cpu_reservation, memory_limit, memory_reservation)
- Separate resource table follows Phase 3 override pattern (not columns on instances table)
- `encryption_version` column future-proofs key rotation without another migration

### Claude's Discretion

- AES-256-GCM nonce/IV generation strategy
- Exact key file format and permissions
- Migration backfill approach (encrypt existing plaintext secrets)
- Resource limit validation thresholds
- Compose `deploy` section YAML structure details

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `crypto/aes` | stdlib | AES cipher block creation | Official Go cryptography, no external deps |
| `crypto/cipher` | stdlib | GCM AEAD mode wrapper | Provides authenticated encryption with GCM |
| `crypto/rand` | stdlib | Cryptographically secure random number generation | Required for nonce generation |
| `encoding/base64` | stdlib | Binary to text encoding | Store encrypted data in TEXT column |
| `lib/pq` | existing | Postgres driver | Already in use for DB operations |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `os` | stdlib | File operations, permissions | Key file read/write with mode 0600 |
| `path/filepath` | stdlib | Path manipulation | Home directory expansion for ~/.devarch/secret.key |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib crypto | pgcrypto extension | Moves encryption to DB but requires extension, less portable |
| TEXT column | BYTEA column | BYTEA semantically correct but TEXT simpler with base64, no encoding issues |
| Random nonces | Counter-based nonces | Random nonces simpler, no state tracking, sufficient for local dev scale |

**Installation:**
```bash
# No external dependencies — all standard library
go mod tidy
```

## Architecture Patterns

### Recommended Project Structure
```
api/
├── internal/
│   ├── crypto/
│   │   ├── keymanager.go      # Key file operations, auto-generation
│   │   └── cipher.go           # Encrypt/decrypt with nonce prepend
│   ├── api/handlers/
│   │   └── [existing].go       # Add redaction at serialization
│   └── compose/
│       └── generator.go        # Add deploy.resources section
├── migrations/
│   ├── 020_encryption.up.sql   # Add encrypted_value, encryption_version
│   └── 021_resources.up.sql    # Create instance_resource_limits table
```

### Pattern 1: Nonce-Prepended Ciphertext Storage

**What:** Generate random 12-byte nonce per encryption, prepend to ciphertext, store as single base64-encoded TEXT value

**When to use:** All AES-GCM encryption where nonce must travel with ciphertext

**Example:**
```go
// Source: https://pkg.go.dev/crypto/cipher official docs
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

// Encrypt plaintext, return base64-encoded nonce+ciphertext
func Encrypt(key []byte, plaintext string) (string, error) {
    block, err := aes.NewCipher(key) // key must be 32 bytes for AES-256
    if err != nil {
        return "", fmt.Errorf("create cipher: %w", err)
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("create GCM: %w", err)
    }

    nonce := make([]byte, aesGCM.NonceSize()) // 12 bytes for GCM
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("generate nonce: %w", err)
    }

    // Seal prepends nonce (dst), appends ciphertext+tag
    ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

    // Store as base64 TEXT in database
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt base64-encoded nonce+ciphertext
func Decrypt(key []byte, encoded string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return "", fmt.Errorf("decode base64: %w", err)
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("create cipher: %w", err)
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("create GCM: %w", err)
    }

    nonceSize := aesGCM.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("decrypt: %w", err)
    }

    return string(plaintext), nil
}
```

### Pattern 2: Auto-Generated Key File with Permissions

**What:** Generate 32-byte AES-256 key on first use, store at `~/.devarch/secret.key` with mode 0600

**When to use:** Key management for local dev tools where ease-of-use trumps complex key rotation

**Example:**
```go
// Source: Synthesized from Go stdlib patterns
package crypto

import (
    "crypto/rand"
    "fmt"
    "os"
    "path/filepath"
)

const KeySize = 32 // AES-256

// LoadOrGenerateKey reads key from ~/.devarch/secret.key, generates if missing
func LoadOrGenerateKey() ([]byte, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("get home dir: %w", err)
    }

    devarchDir := filepath.Join(home, ".devarch")
    keyPath := filepath.Join(devarchDir, "secret.key")

    // Try to read existing key
    key, err := os.ReadFile(keyPath)
    if err == nil {
        if len(key) != KeySize {
            return nil, fmt.Errorf("invalid key size: %d bytes (expected %d)", len(key), KeySize)
        }
        return key, nil
    }

    if !os.IsNotExist(err) {
        return nil, fmt.Errorf("read key: %w", err)
    }

    // Generate new key
    key = make([]byte, KeySize)
    if _, err := rand.Read(key); err != nil {
        return nil, fmt.Errorf("generate key: %w", err)
    }

    // Ensure directory exists
    if err := os.MkdirAll(devarchDir, 0755); err != nil {
        return nil, fmt.Errorf("create devarch dir: %w", err)
    }

    // Write key with restrictive permissions (owner read/write only)
    if err := os.WriteFile(keyPath, key, 0600); err != nil {
        return nil, fmt.Errorf("write key: %w", err)
    }

    return key, nil
}
```

### Pattern 3: Transparent Encryption Layer

**What:** Encrypt on INSERT, decrypt on SELECT, application logic never sees plaintext in DB

**When to use:** Compliance requirements for encryption at rest without changing application logic

**Example:**
```go
// Source: Synthesized from database encryption patterns
package handlers

// Before INSERT/UPDATE: encrypt if is_secret=true
func (h *Handler) SetEnvVar(serviceID int, key, value string, isSecret bool) error {
    valueToStore := value
    var encryptionVersion int

    if isSecret {
        encrypted, err := h.cipher.Encrypt(value)
        if err != nil {
            return fmt.Errorf("encrypt secret: %w", err)
        }
        valueToStore = ""  // Clear plaintext value
        encryptionVersion = 1  // Mark as encrypted

        _, err = h.db.Exec(`
            INSERT INTO service_env_vars (service_id, key, value, encrypted_value, encryption_version, is_secret)
            VALUES ($1, $2, $3, $4, $5, $6)
            ON CONFLICT (service_id, key) DO UPDATE
            SET value = EXCLUDED.value,
                encrypted_value = EXCLUDED.encrypted_value,
                encryption_version = EXCLUDED.encryption_version,
                is_secret = EXCLUDED.is_secret
        `, serviceID, key, valueToStore, encrypted, encryptionVersion, isSecret)
    } else {
        _, err := h.db.Exec(`
            INSERT INTO service_env_vars (service_id, key, value, is_secret)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (service_id, key) DO UPDATE
            SET value = EXCLUDED.value, is_secret = EXCLUDED.is_secret
        `, serviceID, key, value, isSecret)
    }

    return nil
}

// After SELECT: decrypt if encrypted_version > 0
func (h *Handler) GetEnvVars(serviceID int, reveal bool) ([]EnvVar, error) {
    rows, err := h.db.Query(`
        SELECT key, value, encrypted_value, encryption_version, is_secret
        FROM service_env_vars
        WHERE service_id = $1
    `, serviceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var vars []EnvVar
    for rows.Next() {
        var key, value string
        var encryptedValue sql.NullString
        var encryptionVersion int
        var isSecret bool

        if err := rows.Scan(&key, &value, &encryptedValue, &encryptionVersion, &isSecret); err != nil {
            return nil, err
        }

        finalValue := value
        if encryptionVersion > 0 && encryptedValue.Valid {
            decrypted, err := h.cipher.Decrypt(encryptedValue.String)
            if err != nil {
                return nil, fmt.Errorf("decrypt %s: %w", key, err)
            }

            // Redaction decision at serialization layer
            if reveal {
                finalValue = decrypted
            } else {
                finalValue = "***"  // API response redaction
            }
        }

        vars = append(vars, EnvVar{Key: key, Value: finalValue, IsSecret: isSecret})
    }

    return vars, nil
}
```

### Pattern 4: Compose Deploy Resources

**What:** Map instance resource limits to Docker Compose v3 `deploy.resources` section

**When to use:** Constraining container resource usage in local dev environment

**Example:**
```go
// Source: https://docs.docker.com/reference/compose-file/deploy/
// Add to compose/generator.go serviceConfig struct
type serviceConfig struct {
    Image         string                 `yaml:"image,omitempty"`
    ContainerName string                 `yaml:"container_name"`
    // ... existing fields ...
    Deploy        *deployConfig          `yaml:"deploy,omitempty"`
    Networks      []string               `yaml:"networks"`
}

type deployConfig struct {
    Resources *resourcesConfig `yaml:"resources,omitempty"`
}

type resourcesConfig struct {
    Limits       *resourceLimits `yaml:"limits,omitempty"`
    Reservations *resourceLimits `yaml:"reservations,omitempty"`
}

type resourceLimits struct {
    CPUs   string `yaml:"cpus,omitempty"`   // e.g. "0.5", "2.0"
    Memory string `yaml:"memory,omitempty"` // e.g. "512m", "1g"
}

// In generator.go Generate() method, after loading instance overrides:
func (g *Generator) loadResourceLimits(instanceID int) (*deployConfig, error) {
    var cpuLimit, cpuReservation, memoryLimit, memoryReservation sql.NullString

    err := g.db.QueryRow(`
        SELECT cpu_limit, cpu_reservation, memory_limit, memory_reservation
        FROM instance_resource_limits
        WHERE instance_id = $1
    `, instanceID).Scan(&cpuLimit, &cpuReservation, &memoryLimit, &memoryReservation)

    if err == sql.ErrNoRows {
        return nil, nil // No limits configured
    }
    if err != nil {
        return nil, fmt.Errorf("query resource limits: %w", err)
    }

    // Only create deploy section if at least one limit exists
    if !cpuLimit.Valid && !cpuReservation.Valid && !memoryLimit.Valid && !memoryReservation.Valid {
        return nil, nil
    }

    deploy := &deployConfig{Resources: &resourcesConfig{}}

    if cpuLimit.Valid || memoryLimit.Valid {
        deploy.Resources.Limits = &resourceLimits{}
        if cpuLimit.Valid {
            deploy.Resources.Limits.CPUs = cpuLimit.String
        }
        if memoryLimit.Valid {
            deploy.Resources.Limits.Memory = memoryLimit.String
        }
    }

    if cpuReservation.Valid || memoryReservation.Valid {
        deploy.Resources.Reservations = &resourceLimits{}
        if cpuReservation.Valid {
            deploy.Resources.Reservations.CPUs = cpuReservation.String
        }
        if memoryReservation.Valid {
            deploy.Resources.Reservations.Memory = memoryReservation.String
        }
    }

    return deploy, nil
}
```

**Generated YAML:**
```yaml
services:
  postgres-prod:
    image: postgres:15
    container_name: postgres-prod
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2g
        reservations:
          cpus: '1.0'
          memory: 1g
    networks:
      - devarch-prod-net
```

### Anti-Patterns to Avoid

- **Storing nonce separately:** Don't create separate DB column for nonce — prepend to ciphertext, single column
- **Using math/rand:** NEVER use `math/rand` for nonce generation — always `crypto/rand`
- **Counter-based nonces:** Don't track nonce state in DB — use random nonces, simpler and sufficient
- **Decrypting at DB layer:** Don't use Postgres functions for decrypt — Go layer controls redaction policy
- **BYTEA column type:** Don't use BYTEA — TEXT with base64 simpler, avoids lib/pq encoding quirks
- **Encrypting non-secrets:** Don't encrypt where `is_secret=false` — unnecessary overhead
- **Zero resource values:** Don't write `cpus: "0"` — omit deploy section entirely when no limits

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| AES encryption | Custom crypto implementation | Go stdlib `crypto/aes` + `crypto/cipher` | Timing attacks, padding oracles, constant-time operations require expert review |
| Random number generation | Pseudo-random with seed | `crypto/rand.Reader` | Predictable RNG breaks encryption completely |
| Base64 encoding | String manipulation | `encoding/base64.StdEncoding` | Handles padding, alphabet, line breaks correctly |
| Nonce management | Counter tracking in DB | Random nonces per encryption | State tracking adds complexity, random sufficient at local dev scale |
| Key derivation from passphrase | Hash function on passphrase | Not needed (no passphrase requirement) | If needed in future, use scrypt/argon2, not SHA256 |

**Key insight:** Cryptography is easy to get subtly wrong. Nonce reuse breaks GCM completely. Timing attacks leak keys. Use stdlib, don't innovate.

## Common Pitfalls

### Pitfall 1: Nonce Reuse

**What goes wrong:** Reusing the same nonce with the same key completely breaks AES-GCM authentication, allows message forgery and plaintext recovery

**Why it happens:** Stateful counter seems "cleaner" than random nonces, or nonce generation happens in a loop without fresh randomness

**How to avoid:** Use `crypto/rand.Reader` to generate fresh random nonce for EVERY encryption operation. With 96-bit (12-byte) nonces, collision probability is negligible up to 2^48 messages.

**Warning signs:**
- Same nonce bytes in multiple database rows
- Predictable nonce patterns (incrementing integers)
- Authentication failures on decrypt

**Code check:**
```go
// WRONG - reusing nonce
nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// CORRECT - fresh random nonce every time
nonce := make([]byte, aesGCM.NonceSize())
if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    return err
}
```

### Pitfall 2: Key File Permissions Too Open

**What goes wrong:** Key file readable by other users allows anyone on system to decrypt all secrets

**Why it happens:** Default file permissions (0644) allow world-read, or permissions not set on initial write

**How to avoid:** Always write key file with mode 0600 (owner read/write only). SSH enforces this for private keys and will refuse to work otherwise.

**Warning signs:**
- `ls -l ~/.devarch/secret.key` shows `-rw-r--r--` instead of `-rw-------`
- Security scanners flag world-readable secret files

**Code check:**
```go
// WRONG - default permissions
os.WriteFile(keyPath, key, 0644)

// CORRECT - restrictive permissions
os.WriteFile(keyPath, key, 0600)
```

### Pitfall 3: Forgetting Decryption in New Query Paths

**What goes wrong:** Adding new API endpoints that query env vars directly, forgetting to decrypt `encrypted_value` column

**Why it happens:** Encryption layer is not centralized, each query location needs decrypt logic

**How to avoid:** Centralize env var loading into helper function that always handles decryption. Every new feature queries through helper, not raw SQL.

**Warning signs:**
- Base64 garbage appearing in API responses
- Edit endpoints work (decrypt implemented) but read-only endpoints show encrypted data

**Pattern:**
```go
// Centralized helper
func (h *Handler) loadEnvVars(serviceID int, reveal bool) ([]EnvVar, error) {
    // Always handles encryption_version check and decrypt
}

// All endpoints use helper
func (h *Handler) GetEffectiveConfig(...) {
    vars, err := h.loadEnvVars(serviceID, false) // redacted
}

func (h *Handler) GetServiceForEdit(...) {
    vars, err := h.loadEnvVars(serviceID, true) // revealed
}
```

### Pitfall 4: Compose Deploy Section on Wrong Podman Version

**What goes wrong:** `deploy` key ignored in standalone Docker/Podman Compose, resource limits not applied

**Why it happens:** `deploy` section is Docker Swarm-specific in Compose v3, requires `--compatibility` flag or Compose v2.4+ support

**How to avoid:** Document that resource limits require Docker Compose v2+ or Podman Compose 1.2+. Validate in plan output, warn if deploy section may be ignored.

**Warning signs:**
- Containers running without memory limits despite config
- `docker-compose` warnings about ignoring deploy key
- Resource exhaustion in local dev despite configured limits

**Validation approach:**
```go
// In plan differ or validation step
func validateResourceLimits(limits *resourceLimits) []string {
    var warnings []string

    if limits.Memory != "" {
        // Parse memory string (e.g. "4m", "512m", "1g")
        mem := parseMemory(limits.Memory)
        if mem < 4*1024*1024 { // 4MB
            warnings = append(warnings, fmt.Sprintf("Memory limit %s is very low, container may fail to start", limits.Memory))
        }
    }

    if limits.CPUs != "" {
        cpu, _ := strconv.ParseFloat(limits.CPUs, 64)
        if cpu < 0.01 {
            warnings = append(warnings, fmt.Sprintf("CPU limit %s is extremely low", limits.CPUs))
        }
    }

    return warnings
}
```

### Pitfall 5: TEXT Column Encoding Confusion

**What goes wrong:** lib/pq interprets base64 string as UTF-8, corrupts data on INSERT/SELECT with unexpected character encoding errors

**Why it happens:** Mixing base64.StdEncoding (with padding `=`) with URL-safe encoding, or database client charset issues

**How to avoid:** Always use `base64.StdEncoding` consistently for encode and decode. Postgres TEXT columns handle UTF-8, base64 alphabet is ASCII subset (safe).

**Warning signs:**
- Decryption fails with "illegal base64 data at input byte X"
- Different encryption output lengths for same plaintext
- Encoding errors on non-ASCII plaintext after decrypt

**Code check:**
```go
// CONSISTENT encoding
encoded := base64.StdEncoding.EncodeToString(ciphertext)
data, _ := base64.StdEncoding.DecodeString(encoded)

// Don't mix with URL-safe or raw encodings
```

## Code Examples

Verified patterns from official sources:

### AES-256 Key Generation
```go
// Source: https://pkg.go.dev/crypto/rand
import "crypto/rand"

key := make([]byte, 32) // AES-256 requires 32-byte key
if _, err := rand.Read(key); err != nil {
    return fmt.Errorf("generate key: %w", err)
}
```

### GCM Encryption with Nonce Prepend
```go
// Source: https://pkg.go.dev/crypto/cipher Example_encrypt
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

block, _ := aes.NewCipher(key)
aesgcm, _ := cipher.NewGCM(block)

nonce := make([]byte, aesgcm.NonceSize())
io.ReadFull(rand.Reader, nonce)

ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
// ciphertext is now: [nonce bytes][encrypted bytes][tag bytes]
```

### GCM Decryption with Nonce Extract
```go
// Source: https://pkg.go.dev/crypto/cipher Example_decrypt
block, _ := aes.NewCipher(key)
aesgcm, _ := cipher.NewGCM(block)

nonceSize := aesgcm.NonceSize()
nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
if err != nil {
    // Authentication failed - data tampered or wrong key
}
```

### File Permissions Check
```go
// Source: Linux file permissions best practices
import "os"

info, _ := os.Stat(keyPath)
mode := info.Mode().Perm()
if mode != 0600 {
    return fmt.Errorf("key file permissions %o too open (expected 0600)", mode)
}
```

### Memory Limit Parsing
```go
// Parse memory string like "512m", "1g", "2G"
func parseMemory(s string) (int64, error) {
    s = strings.ToLower(strings.TrimSpace(s))

    var multiplier int64 = 1
    if strings.HasSuffix(s, "k") {
        multiplier = 1024
        s = strings.TrimSuffix(s, "k")
    } else if strings.HasSuffix(s, "m") {
        multiplier = 1024 * 1024
        s = strings.TrimSuffix(s, "m")
    } else if strings.HasSuffix(s, "g") {
        multiplier = 1024 * 1024 * 1024
        s = strings.TrimSuffix(s, "g")
    }

    val, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid memory value: %s", s)
    }

    return val * multiplier, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `cipher.NewGCM` with manual nonce | `cipher.NewGCMWithRandomNonce` | Go 1.24 (2024) | Automatic nonce generation but increases overhead by 12 bytes |
| BYTEA storage | TEXT with base64 | Current best practice | Avoids lib/pq encoding complexity, simpler queries |
| Separate nonce column | Nonce prepended to ciphertext | Standard since GCM introduction | Single column storage, atomic read/write |
| Counter-based nonces | Random nonces | Current recommendation | Eliminates state tracking, collision probability negligible |
| Compose v2 compatibility mode | Compose v3 deploy section | Docker Compose 2.0+ (2021) | Native support for deploy resources in non-Swarm mode |

**Deprecated/outdated:**
- `cipher.NewGCMWithNonceSize` / `NewGCMWithTagSize`: Only for compatibility with legacy systems, standard GCM preferred
- Compose v2 syntax `mem_limit`, `cpus` at service level: Still works but `deploy.resources` is standard in v3+
- Postgres pgcrypto extension: Viable but moves encryption to DB layer, less flexible than app-layer control

## Open Questions

1. **Migration backfill strategy for existing secrets**
   - What we know: Existing `is_secret=true` rows have plaintext in `value` column
   - What's unclear: Migrate on first API server start (expensive), lazy migrate on first access, or manual migration command?
   - Recommendation: Lazy migration — on SELECT, if `encryption_version=0` AND `is_secret=true`, encrypt and UPDATE. Zero-downtime, gradual.

2. **Resource limit validation thresholds**
   - What we know: Docker allows absurdly low values (1KB memory), Podman similar
   - What's unclear: What's a reasonable minimum? 4MB memory? 0.01 CPU?
   - Recommendation: Warn (don't block) on memory < 4MB (standard page sizes), CPU < 0.01 (1% of one core). Let experts shoot themselves in foot if needed.

3. **Key rotation strategy**
   - What we know: `encryption_version` column exists for future key rotation
   - What's unclear: How to rotate without downtime? Keep old key for decrypt, new key for encrypt?
   - Recommendation: Not in scope for Phase 9. Document that `encryption_version` enables future rotation, phase 9 only implements version 1.

4. **Export redaction interaction**
   - What we know: `export/secrets.go` already has keyword heuristic
   - What's unclear: Does it redact before encryption or after? Does encrypted value bypass heuristic?
   - Recommendation: Test that encrypted secrets are ALSO checked by heuristic (value decrypted, then heuristic applied). Belt and suspenders.

## Sources

### Primary (HIGH confidence)
- [Go crypto/cipher official docs](https://pkg.go.dev/crypto/cipher) - AEAD interface, GCM mode, Seal/Open methods
- [Go crypto/aes official docs](https://pkg.go.dev/crypto/aes) - AES cipher block creation
- [Docker Compose deploy specification](https://docs.docker.com/reference/compose-file/deploy/) - deploy.resources.limits and reservations syntax
- [Go crypto/rand official docs](https://pkg.go.dev/crypto/rand) - Cryptographically secure random generation

### Secondary (MEDIUM confidence)
- [AES-GCM encryption examples (GitHub Gist)](https://gist.github.com/kkirsche/e28da6754c39d5e7ea10) - Community-verified patterns
- [Twilio: Encrypt and Decrypt Data in Go with AES-256](https://www.twilio.com/en-us/blog/developers/community/encrypt-and-decrypt-data-in-go-with-aes-256) - Real-world implementation guide
- [PostgreSQL bytea documentation](https://www.postgresql.org/docs/current/datatype-binary.html) - Binary vs TEXT storage tradeoffs
- [Linux file permissions: chmod 600](https://www.oreateai.com/blog/understanding-chmod-600-a-deep-dive-into-linux-file-permissions/63b07c5bccd72ac9efee65e146265301) - Standard practice for secret key files
- [frereit's blog: AES-GCM and breaking it on nonce reuse](https://frereit.de/aes_gcm/) - Security analysis of nonce reuse
- [Podman Compose compatibility](https://podman-desktop.io/docs/migrating-from-docker/managing-docker-compatibility) - Deploy resources support in Podman
- [Medium: Securing Information in Database using Data Encryption (Go)](https://medium.com/swlh/securing-information-in-database-using-data-encryption-written-in-go-4b2754214050) - Database encryption patterns

### Tertiary (LOW confidence - for context only)
- [DEV Community: Go & AES-GCM Security Deep Dive](https://dev.to/js402/go-aes-gcm-a-security-deep-dive-3ec8) - Implementation pitfalls discussion
- [LinkedIn: Building Secure AES-GCM Library in Golang](https://www.linkedin.com/pulse/building-secure-aes-gcm-library-golang-transferring-nonces-hinch) - Nonce management strategies

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Go stdlib is definitive source, no alternatives needed
- Architecture: HIGH - Patterns verified from official Go docs and Docker Compose spec
- Pitfalls: HIGH - Nonce reuse and key permissions are well-documented security issues
- Resource limits: MEDIUM - Docker/Podman compatibility nuances exist, validation thresholds are judgment calls
- Migration backfill: MEDIUM - Multiple valid approaches, lazy migration recommended but not verified in production

**Research date:** 2026-02-08
**Valid until:** 2026-03-08 (30 days - stable domain, Go crypto API rarely changes)
