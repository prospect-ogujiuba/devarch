# Phase 9: Secrets & Resources - Context

**Gathered:** 2026-02-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Encrypt secret env vars at rest using AES-256-GCM, redact them in all API outputs (effective config, plan diff, compose preview, export), and add optional per-instance resource limits (CPU, memory) that map to compose `deploy.resources` fields. Key auto-generated on first use. No new capabilities — this hardens existing data and adds resource controls to existing instance/compose infrastructure.

</domain>

<decisions>
## Implementation Decisions

### Encryption Key Management
- Auto-generate `~/.devarch/secret.key` on first encrypt operation
- No passphrase — server-side operations (compose gen, plan/apply) must be non-interactive
- Key loss requires re-encrypt migration (acceptable for local dev tool)

### Encryption Scope
- Only `value` column in rows where `is_secret = true`
- Applies to both `service_env_vars` and `instance_env_vars` tables
- Config file contents are NOT encrypted (keyword heuristic in `export/secrets.go` remains as safety net for export redaction)

### Redaction Enforcement
- Redact at API serialization layer (handler responses), not at DB layer
- Decrypt on SELECT (SECR-04), then handler decides: mask for read endpoints, pass real value for edit endpoints
- Covered endpoints: GET effective config, plan diff output, compose preview
- Export already handled by `export/secrets.go` (no change)
- WebSocket status doesn't carry env vars (no change)

### Redaction Format
- API JSON responses: `"***"` for masked secret values
- Dashboard display: `••••••••` (8 bullets) — standardize across all components
- Reveal toggle (eye icon) stays as-is in edit mode
- Resolves current inconsistency: `********` / `••••••••` / `***` all become `••••••••` in dashboard

### Resource Limit Granularity
- CPU (decimal cores, e.g. `0.5`, `2.0`) + Memory (with suffix, e.g. `512m`, `1g`) only
- Maps directly to compose `deploy.resources.limits` and `deploy.resources.reservations`
- No GPU, no disk limits (scope creep for local dev tool)

### Resource Limit Behavior
- Limits are optional — omitting means no `deploy.resources` in compose output
- Instance-level only, not template-level (limits are environment-specific: my laptop ≠ yours)
- Matches Phase 3 override pattern: instance overrides are the customization layer

### Resource Limits in Plan Output
- Show limits as part of instance config in plan diff (value changes diffed like env vars)
- Missing limits = omit from compose (don't set to zero)
- Validation: warn if limits seem unreasonable (e.g. memory < 4m), never block

### Migration Strategy
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

</decisions>

<specifics>
## Specific Ideas

- `encryption_version` integer enables future key rotation without schema changes
- Keyword heuristic (`export/secrets.go`) stays as fallback alongside DB `is_secret` flag — belt and suspenders
- Resource limits as separate table (not instance columns) keeps the override pattern consistent with ports, volumes, env vars

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-secrets-resources*
*Context gathered: 2026-02-08*
