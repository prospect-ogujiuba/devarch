---
phase: 09-secrets-resources
plan: 01
subsystem: crypto
tags: [encryption, secrets, resources, migrations]
dependency_graph:
  requires: [Phase 08 complete]
  provides: [encrypted env vars, resource limits table, crypto package]
  affects: [service handlers, instance handlers, effective config]
tech_stack:
  added: [AES-256-GCM encryption, crypto/rand nonce generation]
  patterns: [encrypt-on-write, decrypt-on-read, lazy migration, sticky secrets]
key_files:
  created:
    - api/internal/crypto/keymanager.go
    - api/internal/crypto/cipher.go
    - api/migrations/020_encryption.up.sql
    - api/migrations/020_encryption.down.sql
    - api/migrations/021_resources.up.sql
    - api/migrations/021_resources.down.sql
  modified:
    - api/cmd/server/main.go
    - api/internal/api/routes.go
    - api/internal/api/handlers/service.go
    - api/internal/api/handlers/instance.go
    - api/internal/api/handlers/instance_overrides.go
    - api/internal/api/handlers/instance_effective.go
decisions:
  - AES-256-GCM with 12-byte nonce prepended to ciphertext
  - Key auto-generated on first run at ~/.devarch/secret.key with 0600 permissions
  - Sticky secret pattern: *** preserves existing encrypted_value
  - Lazy migration: plaintext secrets encrypted on first read
  - Redaction for read-only endpoints (GET, effective config)
  - encryption_version=1 for AES-256-GCM, version=0 for plaintext/non-secret
metrics:
  duration: 338
  completed: 2026-02-09T03:55:30Z
---

# Phase 09 Plan 01: Encryption Foundation Summary

**One-liner:** AES-256-GCM encryption at rest for secret env vars with transparent encrypt-on-write/decrypt-on-read, sticky secret UX, and lazy migration for existing plaintext secrets.

## What Was Built

### Crypto Package (Task 1)

**keymanager.go:**
- `LoadOrGenerateKey()` auto-generates 32-byte key using `crypto/rand`
- Creates `~/.devarch/secret.key` with mode 0600
- Creates `~/.devarch/` directory with 0755 if missing
- Validates key size exactly 32 bytes on load

**cipher.go:**
- `Cipher` struct wraps 32-byte encryption key
- `Encrypt()` uses AES-256-GCM with random 12-byte nonce prepended to ciphertext
- `Decrypt()` extracts nonce (first 12 bytes), decrypts remainder
- Returns base64-encoded strings
- Uses only stdlib: `crypto/aes`, `crypto/cipher`, `crypto/rand`, `encoding/base64`

**Migrations:**
- **020_encryption:** adds `encrypted_value TEXT` and `encryption_version INTEGER DEFAULT 0` to `service_env_vars` and `instance_env_vars`
- **021_resources:** creates `instance_resource_limits` table with `cpu_limit`, `cpu_reservation`, `memory_limit`, `memory_reservation` columns, FK to `service_instances`, `UNIQUE(instance_id)`

### Integration (Task 2)

**Startup (main.go):**
- Calls `crypto.LoadOrGenerateKey()` after DB connect
- Creates `crypto.NewCipher(key)` and passes to router
- Logs "encryption key loaded successfully"

**Service Handler:**
- Added `cipher *crypto.Cipher` field
- **Create:** encrypts secrets on INSERT, stores in `encrypted_value` with `encryption_version=1`, sets `value=""`.
- **UpdateEnvVars:** sticky secret pattern — when incoming value is `***` or empty string, preserves existing `encrypted_value` and `encryption_version` from current DB row.
- **Get (loadServiceRelations):** decrypts if `encryption_version > 0`, redacts to `***`. Lazy migration: if `is_secret=true` AND `encryption_version=0` AND `value != ""`, encrypts in-place via UPDATE.

**Instance Handler:**
- Same `cipher` field and constructor pattern
- **UpdateEnvVars (instance_overrides.go):** identical encrypt-on-INSERT + sticky secret logic as service handler
- **loadServiceEnvVars + loadInstanceEnvVars (instance_effective.go):** SELECT includes `encrypted_value, encryption_version` columns, redacts to `***` for all secrets (read-only endpoint)

**Query Pattern:**
```sql
-- INSERT (secrets)
INSERT INTO service_env_vars
  (service_id, key, value, is_secret, encrypted_value, encryption_version)
VALUES ($1, $2, '', true, <encrypted>, 1)

-- INSERT (non-secrets)
INSERT INTO service_env_vars
  (service_id, key, value, is_secret, encrypted_value, encryption_version)
VALUES ($1, $2, <plaintext>, false, NULL, 0)

-- SELECT
SELECT key, value, is_secret, encrypted_value, encryption_version
FROM service_env_vars WHERE service_id = $1
```

## Deviations from Plan

None — plan executed exactly as written.

## Verification Results

**Compilation:** ✅ `go build ./...` passes

**Key file generation:** ✅ 32-byte key at `/root/.devarch/secret.key` with mode 0600 in API container

**Database storage:**
```
# Created service with secret
    key     |   value   | is_secret | enc_len | encryption_version
------------+-----------+-----------+---------+--------------------
 PUBLIC_VAR | public123 | f         |         |                  0
 SECRET_KEY |           | t         |      68 |                  1
```

**GET response redaction:** ✅
```json
{
  "key": "SECRET_KEY",
  "value": "***",
  "is_secret": true
}
```

**Sticky secret preservation:** ✅ Updating with `"value": "***"` preserved existing encrypted_value (68 bytes)

**New secret encryption:** ✅ Updating with new plaintext changed ciphertext length (68 → 60 bytes)

**Instance env vars:** ✅ Same encryption behavior for instance overrides

**Effective config:** ✅ Both template and instance secrets redacted to `***`

## Technical Notes

**Nonce generation:** Uses `io.ReadFull(rand.Reader, nonce)` — never `math/rand`

**Lazy migration:** No separate backfill script needed. Plaintext secrets (is_secret=true, encryption_version=0, value != "") are encrypted on first read via SELECT in `loadServiceRelations`.

**Sticky secret UX:** Dashboard can send `"***"` for secrets user hasn't changed, preserving encrypted value without exposing plaintext. Works because UpdateEnvVars checks `if e.Value == "***" || e.Value == ""` before preserving.

**Encryption version:** Future-proofs algorithm changes. Version 0 = plaintext/non-secret, version 1 = AES-256-GCM. Adding new algorithms increments version.

**Error handling:** Encryption failures return 500 with descriptive message. Decryption failures in lazy migration are silent (don't break reads).

## Next Steps (Plan 02)

- Dashboard UI for secret input (password field with show/hide toggle)
- Resource limits CRUD endpoints and UI
- Resource limits integration into compose generation
- Testing edit endpoints return decrypted values (not yet implemented)

## Self-Check: PASSED

All created files exist:
- ✅ api/internal/crypto/keymanager.go
- ✅ api/internal/crypto/cipher.go
- ✅ api/migrations/020_encryption.up.sql
- ✅ api/migrations/020_encryption.down.sql
- ✅ api/migrations/021_resources.up.sql
- ✅ api/migrations/021_resources.down.sql

All commits exist:
- ✅ 6215de2a: add crypto package and encryption migrations
- ✅ cb645e50: integrate encryption into env var handlers
