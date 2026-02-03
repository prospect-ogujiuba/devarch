# Pitfalls Research: Service Composition & Orchestration

**Domain:** Local dev orchestration with dynamic compose generation
**Researched:** 2026-02-03
**Confidence:** HIGH (based on existing codebase patterns + domain experience)

## Critical Pitfalls

### Pitfall 1: Container Name Collisions Without Deterministic Naming

**What goes wrong:**
Multiple stacks using the same service template create containers with identical names. Container runtime refuses to start second instance, or worse, operations target wrong container. Example: Two stacks both try to create `postgres` container — second fails with "name already in use" or accidentally stops the first.

**Why it happens:**
- Template service name used directly as container_name in compose (line 79 in generator.go: `ContainerName: service.Name`)
- No stack context in current naming scheme
- Developers assume service names are globally unique when they're only unique within a stack

**How to avoid:**
- Enforce naming: `devarch-{stack_name}-{instance_name}` pattern at compose generation
- Validate uniqueness at plan stage before apply
- Add DB unique constraint on (stack_id, instance_name) — prevents creation, not just runtime failure
- Identity labels on every container: `devarch.stack_id`, `devarch.instance_id`, `devarch.template_service_id`

**Warning signs:**
- "Container name already in use" errors when starting stack
- Operations (stop/restart/logs) affect wrong service
- Dashboard shows one container status for multiple instances
- Port conflicts even when stack configs specify different ports

**Phase to address:**
Phase 1 (Stack Foundation) — this is isolation primitive, everything breaks without it

---

### Pitfall 2: Shared Network Allows Cross-Stack Contamination

**What goes wrong:**
All services across all stacks connect to `microservices-net` (line 34 in config.sh, line 73 in generator.go). Service in stack A can reach service in stack B by container name. Auto-wiring breaks (connects to wrong database), data leaks between environments, "works on my machine" because dev has different stack running.

**Why it happens:**
- Single hardcoded network name assumed for entire DevArch deployment
- Network name not scoped to stack during compose generation
- Compose generator uses `{External: true}` assuming pre-existing network
- No network isolation was needed before stacks existed

**How to avoid:**
- Per-stack bridge network: `devarch-{stack_name}-net`
- Create network in stack compose (not external) so it's stack-managed
- Update compose generator to accept network name parameter
- Plan output must show which network stack uses (prevents invisible dependencies)
- Allow inter-stack networking as explicit opt-in, not default

**Warning signs:**
- Service discovers unexpected containers via DNS
- Database connections work even with wrong config (hitting different stack's DB)
- Environment variables point to wrong service, but connection succeeds
- Security scan flags containers with unexpected network access

**Phase to address:**
Phase 1 (Stack Foundation) — required for stack isolation, couples with naming

---

### Pitfall 3: Hardcoded Runtime Commands Bypass Container Client Abstraction

**What goes wrong:**
`project/controller.go` line 83: `exec.Command("podman", "inspect"...)` hardcodes runtime, ignoring existing container client. When user runs Docker, code fails. When sudo required, command lacks prefix. Stack operations inherit this pattern, break Docker compatibility promise.

**Why it happens:**
- Project controller predates container client abstraction
- Copy-paste from older code without refactor
- inspect/exec operations feel "special" vs compose operations
- No centralized enforcement of runtime abstraction

**How to avoid:**
- Route all runtime commands through container.Client
- Add Inspect() and Exec() methods to client if missing
- Linting rule: grep for `exec.Command.*podman` or `exec.Command.*docker` fails CI
- Code review checklist: "Uses container client for all runtime ops?"
- Stack operations use client exclusively from start

**Warning signs:**
- "podman: command not found" on Docker systems
- Permission errors despite correct DEVARCH_USE_SUDO
- Different behavior between service operations (working) and stack operations (broken)
- Windows/WSL compatibility issues

**Phase to address:**
Phase 0 (Pre-work) — fix before stack implementation inherits the bug

---

### Pitfall 4: Plan vs Actual State Divergence (Race Conditions in Apply)

**What goes wrong:**
User runs plan, sees "will create 3 services." Concurrent user modifies stack. User runs apply with stale plan. Apply creates wrong resources, orphans containers, or fails midway leaving partial stack. Advisory lock on stack prevents concurrent applies but doesn't prevent plan-apply gap.

**Why it happens:**
- Plan operates on DB snapshot at plan time
- Apply operates on DB state at apply time
- Time gap between plan and apply allows state changes
- Advisory lock during apply but not across plan → apply
- No version/checksum to detect plan staleness

**How to avoid:**
- Plan output includes stack state hash (DB version + git-style hash of relevant rows)
- Apply verifies hash matches current state, rejects stale plans with clear error
- Advisory lock acquired at plan start if plan-with-lock flag set (optional strict mode)
- Plan has expiry timestamp (15min default), apply rejects expired plans
- Show diff in apply if state changed: "Plan assumed X, found Y, aborting"

**Warning signs:**
- Apply succeeds but resulting stack differs from plan preview
- Containers with unexpected state (running when plan said would stop)
- "Container not found" during apply when plan showed would modify
- Race condition errors only under concurrent usage

**Phase to address:**
Phase 3 (Plan/Apply) — core feature implementation, not retrofit

---

### Pitfall 5: Copy-on-Write Config Override Ordering Bugs

**What goes wrong:**
Instance overrides healthcheck interval. Compose merges template healthcheck (with test, timeout, retries) + override (just interval). Result: healthcheck test missing, container unhealthy. Or: override ports but template's port mapping still included, causing conflict. JSON merge behavior surprises users.

**Why it happens:**
- Template config and instance overrides live in separate tables
- Merge happens at compose generation time (runtime, not storage)
- Merge semantics unclear: shallow merge, deep merge, or replace?
- Arrays unclear: append or replace? (e.g., volumes, ports)
- Some fields (healthcheck) are objects, others (ports) are arrays

**How to avoid:**
- Explicit merge strategy per override type:
  - Ports/volumes/env: override replaces entire list for that instance
  - Healthcheck: override replaces entire healthcheck object (not field-by-field merge)
  - Labels: merge (template + instance, instance wins on key collision)
  - Dependencies: replace (instance-specific dependencies)
- Document merge behavior in schema/API
- Plan output shows effective config (post-merge) so users verify
- Validation: effective config must pass compose schema validation before apply

**Warning signs:**
- Healthcheck mysteriously absent after override
- Port mapping includes both template and override ports (conflict)
- Environment variable overrides don't take effect
- Volumes from template + instance both mount (filesystem conflicts)

**Phase to address:**
Phase 2 (Instances & Overrides) — define during schema design, test during implementation

---

### Pitfall 6: Secret Redaction Without Encryption Leaks Secrets

**What goes wrong:**
Secrets stored plaintext in DB with `is_secret=true` flag. API redacts in responses. Plan output redacts. Export redacts. But: DB backups expose secrets, DB logs expose secrets, migration scripts see plaintext, developers with DB access see secrets. "Redaction only" is security theater.

**Why it happens:**
- Encryption seems like later optimization ("MVP doesn't need it")
- Flag-based redaction easier to implement than encryption
- Developers assume local dev = trusted environment = no encryption needed
- Secret management scope creeps (initially just dev secrets, then staging, then prod references)

**How to avoid:**
- Encrypt at rest from v1: AES-256-GCM with key in `~/.devarch/secret.key`
- Auto-generate key on first run if missing
- Encrypt before INSERT, decrypt after SELECT (transparent to app logic)
- Redaction applies to already-encrypted values (double protection)
- Export includes encrypted blobs, import decrypts with user's key
- Plan shows `[REDACTED]` for secrets, never plaintext

**Warning signs:**
- DB dump contains actual passwords in `service_env_vars` table
- Migration logs show `INSERT INTO ... VALUES ('API_KEY', 'actual-key-here')`
- Developers ask "should I commit the .env file?" (answer should always be no)
- Security audit flags plaintext secrets in PostgreSQL

**Phase to address:**
Phase 4 (Secrets) — implement early, painful to retrofit after schema locked

---

### Pitfall 7: Service Export/Import Contracts Without Validation

**What goes wrong:**
Template declares `exports: [{name: "db", type: "postgres"}]`. Instance runs MySQL instead (override). Consumer wires to it expecting Postgres protocol. Connection fails at runtime with cryptic error. Or: template exports "http:8080" but instance overrides to 8081, wiring injects wrong port.

**Why it happens:**
- Export contracts declared on template, but instance overrides affect reality
- No validation that instance overrides maintain contract compatibility
- Contract checking happens at wire time, not at override definition time
- Type system can't enforce "if you override image, contract changes"

**How to avoid:**
- Validate contract compatibility during instance configuration:
  - If template exports postgres, instance must still expose postgres-compatible interface
  - If override changes image, must redeclare exports (no inheritance)
- Effective exports = template exports OR instance exports (explicit override required)
- Plan shows effective exports for each instance
- Wire validation uses effective exports, errors clearly: "Instance X exports [Y] but consumer needs [Z]"
- Strong typing: contract types = postgres, mysql, redis, http, grpc (not freeform strings)

**Warning signs:**
- Wiring succeeds but runtime connection fails with protocol error
- Consumer gets "connection refused" because port changed
- Database driver error: "expected postgres, got mysql"
- Ambiguous wire diagnostics: "could not connect" (which side is wrong?)

**Phase to address:**
Phase 5 (Service Wiring) — validation logic during contract implementation

---

### Pitfall 8: Auto-Wiring Ambiguity Causes Silent Misconfigurations

**What goes wrong:**
Stack has postgres-primary and postgres-replica. Backend declares `imports: [{type: "postgres"}]`. Auto-wiring picks replica (first alphabetically). App writes to replica, replication lag hides bug until data inconsistency surfaces. Or: multiple redis instances, wrong one gets wired.

**Why it happens:**
- Auto-wire uses heuristic (first match, alphabetical, etc.)
- Multiple instances satisfy same contract
- No explicit priority/preference mechanism
- "Simple case" (one provider) assumed to be common case, ambiguity edge case

**How to avoid:**
- Auto-wire only when unambiguous (exactly one provider matches contract)
- Ambiguous cases require explicit wiring:
  ```yaml
  wires:
    - from: backend
      to: postgres-primary
      contract: postgres
  ```
- Plan diagnostics surface ambiguity before apply:
  ```
  WARNING: backend.imports[postgres] has 2 providers:
    - postgres-primary
    - postgres-replica
  Action required: explicit wire in stack config
  ```
- Convention: role labels (`devarch.role=primary`) help auto-wire with priority (optional enhancement)

**Warning signs:**
- Service connects to unexpected instance (logs show wrong hostname)
- Data written to read-replica
- Performance issues because wired to wrong cache tier
- "Works sometimes" behavior depending on stack creation order

**Phase to address:**
Phase 5 (Service Wiring) — ambiguity detection during auto-wire implementation

---

### Pitfall 9: Generated Compose YAML Doesn't Match Stored Compose Files

**What goes wrong:**
Service definition changes in DB. Generated compose YAML reflects change. Old compose file still exists in `compose/{category}/{service}/`. User runs `podman compose -f old-file up`, container starts with stale config. Two sources of truth diverge. Confusion about what's actually running.

**Why it happens:**
- Compose files were source of truth, migrated to DB as source of truth
- Old files not cleaned up during migration
- File materialization creates compose files for debugging, persist after use
- Scripts allow direct compose file usage alongside API-driven workflow

**How to avoid:**
- DB is only source of truth for service definitions (already decided)
- Compose YAML is ephemeral: generate to tmpfile, use, delete (already done in container client)
- Compose preview endpoint returns YAML but never writes to disk except temp
- Explicitly prohibit `docker compose -f` in docs: "Use API/CLI only"
- Remove old `compose/{category}/{service}/` dirs or mark as deprecated legacy

**Warning signs:**
- Container config differs from DB state after restart
- Changes in dashboard don't affect running containers
- Service appears to revert to old config on restart
- Conflict between "what API says" and "what's running"

**Phase to address:**
N/A (Already solved in current design) — document as anti-pattern for stack implementation

---

### Pitfall 10: Docker vs Podman Compose YAML Incompatibilities

**What goes wrong:**
Generated compose uses `user: "1000:1000"` (rootless podman convention). Docker interprets as UID 1000 which doesn't exist in container, permission denied. Or: podman-specific volume driver, healthcheck format differences, network driver options. Stack works on dev machine (podman), fails on CI (docker).

**Why it happens:**
- Subtle compose spec interpretation differences between runtimes
- Podman's rootless mode requires different UID mapping
- Volume mount options differ (`:z`, `:Z` for SELinux)
- Network driver features not 1:1 compatible
- Testing only on one runtime during development

**How to avoid:**
- Container client abstracts runtime, but compose generation must too
- Runtime-specific compose generation paths:
  - User spec: omit if default, use numeric UID if specified
  - Volume options: only use portable flags (`:ro`, not `:z`)
  - Network driver: stick to bridge with minimal options
  - Healthcheck: use CMD-SHELL (most compatible)
- Integration tests run against both runtimes in CI
- Plan output includes "Runtime: podman|docker" so users know what's tested

**Warning signs:**
- "Service works on my machine" (different runtime than teammate)
- Permission errors when switching runtime
- Volume mounts fail with SELinux errors on docker but not podman
- Healthchecks don't trigger on one runtime but work on other

**Phase to address:**
Phase 1 (Stack Foundation) — when adapting compose generator for stacks

---

### Pitfall 11: Advisory Lock Held Across Long-Running Operations Deadlocks System

**What goes wrong:**
Apply acquires advisory lock on stack, starts `compose up` (pulls images, builds, 5min+). Second user tries plan on same stack, blocks. API request times out. Lock held until compose finishes. User cancels API request, lock orphaned (depending on implementation). Stack permanently locked.

**Why it happens:**
- Lock granularity too coarse (entire stack vs individual instances)
- Lock held during slow operations (image pull, build)
- No timeout on lock acquisition
- Lock not released on API timeout/cancellation
- Single lock type (no read vs write locks)

**How to avoid:**
- Advisory lock at operation level, not entire apply:
  - Plan: read lock (multiple concurrent plans ok)
  - Apply: write lock (exclusive)
- Lock timeout: acquire with timeout, fail fast if can't get lock
- Graceful lock release: defer unlock, handle cancellation/timeout
- Operation phases: lock → compute plan → unlock → execute (lock only critical section)
- Postgres advisory locks auto-release on connection close (session-based)
- Dashboard shows lock holder: "Stack locked by apply operation started at 14:32 by user@host"

**Warning signs:**
- API timeouts during concurrent stack operations
- Stack "stuck" — operations fail with "could not acquire lock"
- Lock only clears after API restart
- Dashboard freezes waiting for operation that's actually hung

**Phase to address:**
Phase 3 (Plan/Apply) — during concurrency implementation, test with slow operations

---

### Pitfall 12: devarch.yml Export/Import Doesn't Handle Schema Evolution

**What goes wrong:**
User exports stack as devarch.yml. Project adds new field to schema (e.g., resource limits). User imports old export. New field missing, defaults to nil. Stack works differently than when exported. Or: field removed from schema, import fails on validation. No forward/backward compatibility.

**Why it happens:**
- Export uses current schema version implicitly
- Import assumes export matches current schema
- No version field in export format
- Schema changes don't consider existing exports
- No migration path for old exports

**How to avoid:**
- Version field in devarch.yml: `version: "1"` (semver)
- Import detects version, applies transformations to current schema
- Schema changes categorized:
  - Backward compatible: new optional fields (import works, uses defaults)
  - Breaking: renamed/removed fields (import errors with upgrade instructions)
- Export always includes version and full schema (not minimal)
- Validation errors reference version mismatch: "Export v1, current v2, see upgrade guide"

**Warning signs:**
- Import succeeds but stack missing expected features
- Import fails with cryptic YAML validation error
- Old exports no longer importable after upgrade
- Users report "my backup doesn't restore correctly"

**Phase to address:**
Phase 6 (Export/Import) — format design, include version from v1

---

### Pitfall 13: Config File Materialization Race During Stack Start

**What goes wrong:**
Stack starts 3 services concurrently. All call MaterializeConfigFiles (generator.go:309) to write config files to `compose/{category}/{service}/`. Service A writes nginx.conf, service B also writes nginx.conf (different content). File contents race, nondeterministic which wins. Or: parallel writes corrupt file.

**Why it happens:**
- MaterializeConfigFiles called per-service without coordination
- File paths not scoped to instance (should be `compose/{stack}/{instance}/`)
- Concurrent compose up triggers parallel file materialization
- No file locking during write
- Current code assumes one service at a time (single-service workflow)

**How to avoid:**
- Config files scoped to stack + instance: `compose/stacks/{stack_name}/{instance_name}/`
- Materialize all stack configs atomically before compose up
- Use temp dir per apply operation: `compose/tmp/{apply_id}/{instance}/`, symlink to permanent location after success
- Compose generator references instance-scoped paths in volume mounts
- DB stores files per instance (service_config_files needs instance_id or template inheritance)

**Warning signs:**
- Flaky test failures ("worked locally, failed in CI")
- Config file contents randomly wrong on some stack starts
- Race detector flags concurrent file writes
- File corruption errors from services reading mid-write

**Phase to address:**
Phase 2 (Instances & Overrides) — when config file overrides implemented

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip plan staleness check | Faster implementation | Silent divergence between plan and reality | Never — critical for correctness |
| Use global network for all stacks | One network to manage | Cross-stack contamination, no isolation | Never — defeats stack purpose |
| Store secrets plaintext with flag | Easy to implement | DB dumps leak secrets, compliance failure | Never — retrofit painful |
| Auto-wire always picks first match | Simple algorithm | Silent misconfigurations, ambiguous behavior | Only if explicitly documented + plan warnings |
| Skip instance export validation | Trust user input | Invalid stacks importable, break at apply time | Never — validation cheap, recovery expensive |
| Hardcode runtime instead of client | Quick fix for one operation | Docker compatibility breaks, sudo issues | Never — tech debt already present, don't add more |
| Shallow merge for overrides | Simple logic | Surprising behavior (partial healthcheck), user confusion | Never — define merge strategy upfront |
| Lock entire stack during apply | Coarse locking simpler | Contention, timeouts, poor concurrency | MVP only, refine in v2 |
| MaterializeConfigFiles without coordination | Works for single service | Race conditions in concurrent operations | Never in stack context — must coordinate |
| devarch.yml without version field | One less field to manage | Can't evolve format, backward compat impossible | Never — version costs nothing, lack of it fatal |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Container Registry | Hardcode docker.io registry | Respect runtime's configured registry (podman may use multiple) |
| Volume Mounts | Absolute host paths in DB | Resolve relative to $PROJECT_ROOT at compose generation time |
| Env Var Injection | Inject at import time | Inject at wire resolution time (instance names not known at import) |
| Network DNS | Use container name from template | Use actual instance container name (devarch-stack-instance) |
| Healthcheck URLs | Hardcode localhost | Use 127.0.0.1 (localhost resolution differs between runtimes) |
| Depends_on | Reference template service names | Reference instance container names in generated compose |
| Advisory Locks | Use table locks | Use session-based advisory locks (auto-release on disconnect) |
| Compose Generation | Write to fixed path | Write to tmpfile (avoids collisions, auto-cleanup) |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Load all service relations in memory (generator.go:189+) | N+1 queries per service | Batch load with JOINs for stack generation | 50+ services per stack |
| Lock entire stack for plan | Plans block each other | Read locks for plan, write lock for apply | 5+ concurrent users |
| Generate compose on every status check | CPU spike on dashboard refresh | Cache generated compose, invalidate on change | 20+ stacks |
| MaterializeConfigFiles every start | Slow stack startup | Materialize once, reuse until change | 100+ config files per stack |
| Sequential container starts | Stack startup time = sum of service startup times | Parallel starts (compose handles this), ensure config materialized first | 10+ services per stack |
| Full table scan for container status | Slow dashboard with many stacks | Index on container names, filter by stack prefix | 100+ containers |
| Export/import without streaming | OOM on large stacks | Stream YAML generation, SAX-style parsing | 5GB+ config data |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Secret in compose YAML tmpfile unencrypted | Secrets readable from /tmp by other users | tmpfile with 0600 perms, unlink immediately after use |
| Container identity labels not set | Malicious container impersonates devarch container | Always set devarch.stack_id, verify label on operations |
| Plan output includes plaintext secrets | Secrets in terminal history, logs, CI output | Redact before display, even if encrypted at rest |
| Instance can reach other stacks via host network | Data exfiltration, lateral movement | Per-stack networks with no inter-stack routing by default |
| Config files world-readable | Secrets in config files readable by all users | Materialize with mode from DB (0600 for sensitive), validate mode |
| devarch.yml includes unencrypted secrets | Secrets in git, chat, email when sharing | Export with encrypted blobs, redact option for sanitized export |
| No auth on API | Anyone on localhost can modify all stacks | Out of scope for v1 (local tool), document assumption |
| Container with --privileged in template | Privilege escalation, container escape | Warn on privileged containers, block in default policy |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Cryptic plan output | Can't understand what will change | Structured diff: `+ add container X`, `~ modify Y.port: 8080 → 8081`, `- remove Z` |
| Silent auto-wire | Service breaks, user doesn't know why | Always show wiring in plan, explicit confirmation for auto-wired connections |
| Apply failure leaves partial stack | Some containers running, unclear how to recover | Atomic apply: rollback on failure, or clear partial-state error |
| Plan and apply in one command | No preview, scary for production-like stacks | Separate plan and apply, require explicit apply (like terraform) |
| Instance names auto-generated | Can't predict container names, hard to debug | Require explicit instance names, suggest defaults |
| Error: "could not start container" | No context on which container, why | `Instance 'postgres-primary' failed: port 5432 already in use by devarch-other-postgres` |
| Stale lock with no owner info | "Stack locked" forever, no way to understand why | Lock metadata: timestamp, operation, PID, allow admin force-unlock |
| Wire failure at runtime | Service crashes after startup | Validate wiring at plan time, fail early with clear missing-contract error |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Container Naming:** Instance uses deterministic names — verify by starting two stacks with same template, check `podman ps` shows unique names
- [ ] **Network Isolation:** Stacks can't reach each other — verify by attempting cross-stack DNS resolution (should fail)
- [ ] **Runtime Abstraction:** All operations use container client — verify by running on Docker, all operations work
- [ ] **Plan Staleness:** Concurrent modification detected — verify by running plan, modifying DB directly, apply should reject
- [ ] **Override Merge:** Healthcheck override replaces entire healthcheck — verify by overriding interval only, test field still present
- [ ] **Secret Encryption:** Secrets encrypted in DB — verify by querying service_env_vars directly, value should be ciphertext
- [ ] **Secret Redaction:** Plan output redacts secrets — verify by planning stack with secret, output shows [REDACTED]
- [ ] **Contract Validation:** Invalid export override detected — verify by changing image without redeclaring exports, error during config
- [ ] **Auto-Wire Ambiguity:** Multiple providers flagged — verify by creating stack with 2 postgres instances, 1 consumer, plan shows warning
- [ ] **Export Versioning:** Old export importable — verify by importing devarch.yml from previous version (or mock old format)
- [ ] **Advisory Lock Release:** Lock released on timeout — verify by killing API mid-apply, lock clears automatically
- [ ] **Config File Scoping:** Instances don't overwrite each other — verify by starting stack with 2 instances using same config file name
- [ ] **Compose Generation Ephemeral:** No persistent compose files — verify by checking compose/ dir after start, only tmp files

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Container name collision | LOW | Stop conflicting container, rename in DB, regenerate compose, restart |
| Network contamination | MEDIUM | Stop all stacks, recreate networks per-stack, update compose, restart stacks |
| Hardcoded runtime | LOW | Add container client methods, refactor calls, test on both runtimes |
| Stale plan applied | MEDIUM | Detect actual state, compute diff, rollback or roll-forward to desired state |
| Override merge broke config | LOW | Query instance overrides, remove bad override, regenerate compose, restart |
| Plaintext secrets in DB | HIGH | Generate encryption key, migrate all secrets to encrypted, rotate exposed secrets |
| Export contract mismatch | LOW | Reconfigure instance exports, recompute wiring, apply corrected plan |
| Auto-wire picked wrong provider | LOW | Add explicit wire, remove auto-wire, apply corrected plan |
| Compose file divergence | MEDIUM | Delete all old compose files, regenerate from DB, verify via compose preview, redeploy |
| Docker/Podman incompatibility | MEDIUM | Test compose YAML on both runtimes, fix incompatible directives, regenerate |
| Advisory lock stuck | LOW | Force unlock via admin command or DB query: `SELECT pg_advisory_unlock(stack_id)` |
| Config file race corruption | MEDIUM | Stop stack, delete corrupt config dir, rematerialize from DB, restart stack |
| devarch.yml version mismatch | MEDIUM | Run migration script to transform old format to new, re-import |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Container naming collision | Phase 1: Stack Foundation | Two stacks with same template, unique container names |
| Network contamination | Phase 1: Stack Foundation | Cross-stack DNS resolution fails |
| Hardcoded runtime bypass | Phase 0: Pre-work Refactor | All tests pass on Docker and Podman |
| Plan-apply staleness | Phase 3: Plan/Apply Workflow | Concurrent modify causes apply rejection |
| Override merge bugs | Phase 2: Instances & Overrides | Override tests verify effective config correct |
| Plaintext secrets | Phase 4: Secrets Encryption | DB query shows ciphertext, API returns redacted |
| Export contract validation | Phase 5: Service Wiring | Invalid override rejected during configuration |
| Auto-wire ambiguity | Phase 5: Service Wiring | Plan diagnostics flag ambiguous wiring |
| Compose file divergence | N/A (already solved) | Document ephemeral compose pattern |
| Docker/Podman incompatibility | Phase 1: Stack Foundation | CI tests both runtimes |
| Advisory lock deadlock | Phase 3: Plan/Apply Workflow | Lock timeout test, cancellation releases lock |
| Config file materialization race | Phase 2: Instances & Overrides | Concurrent starts produce consistent files |
| Export format evolution | Phase 6: Export/Import | Import old version test, upgrade guide |

## Sources

- Existing DevArch codebase analysis (generator.go, container client, project controller)
- Domain knowledge: Docker Compose, Podman Compose, Terraform plan/apply workflow
- Observed patterns: naming collisions (Kubernetes pods), network isolation (Docker networks), advisory locking (Postgres)
- Anti-patterns: config file races (parallel builds), secret management (plaintext leakage), schema evolution (API versioning)

---
*Pitfalls research for: DevArch Stacks & Instances*
*Researched: 2026-02-03*
*Confidence: HIGH — based on existing codebase + domain patterns*
