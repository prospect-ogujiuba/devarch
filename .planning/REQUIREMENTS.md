# Requirements: DevArch Stacks & Instances

**Defined:** 2026-02-03
**Core Value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.

## v1 Requirements

### Baseline & Guardrails

- [x] **BASE-01**: DevArch identity label constants defined (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
- [x] **BASE-02**: Stack/instance name validation helpers (charset, length, uniqueness)
- [x] **BASE-03**: Runtime abstraction fix — all container operations route through container.Client, no hardcoded podman exec.Command

### Stack Management

- [x] **STCK-01**: Create stack with name, description, optional network name
- [x] **STCK-02**: List stacks with status summary (instance count, running count)
- [x] **STCK-03**: Get stack detail (instances, network, last applied)
- [x] **STCK-04**: Update stack metadata (description only; stack name immutable)
- [x] **STCK-05**: Delete stack (with cascade: stop containers, remove instances, remove network)
- [x] **STCK-06**: Enable/disable stack without deleting
- [x] **STCK-07**: Clone stack with new name (primary rename mechanism; copies instances + overrides)
- [x] **STCK-08**: Dashboard UI for stack CRUD (list, create, detail, edit, delete, clone)

### Service Instances

- [x] **INST-01**: Create instance from template service within a stack
- [x] **INST-02**: Override instance ports (copy-on-write from template)
- [x] **INST-03**: Override instance volumes
- [x] **INST-04**: Override instance environment variables
- [x] **INST-05**: Override instance dependencies
- [x] **INST-06**: Override instance labels
- [x] **INST-07**: Override instance domains
- [x] **INST-08**: Override instance healthchecks
- [x] **INST-09**: Override instance config files
- [x] **INST-10**: Effective config resolver (template + overrides merged, overrides win)
- [x] **INST-11**: List/get/update/delete instances within a stack
- [x] **INST-12**: Dashboard UI for instance management (add to stack, edit overrides, remove)

### Network Isolation

- [x] **NETW-01**: Deterministic container naming: devarch-{stack}-{instance}
- [x] **NETW-02**: Per-stack bridge network: devarch-{stack}-net
- [x] **NETW-03**: EnsureNetwork in container client (create if not exists, Docker + Podman)
- [x] **NETW-04**: Identity labels injected on all stack containers

### Compose Generation

- [x] **COMP-01**: Stack compose generator produces one YAML with N services from effective configs
- [x] **COMP-02**: Stack-scoped materialization path: compose/stacks/{stack}/{instance}/
- [x] **COMP-03**: Existing single-service generator continues working (backward compat)

### Plan/Apply

- [ ] **PLAN-01**: Plan endpoint returns structured preview of what will change (add/modify/remove)
- [ ] **PLAN-02**: Apply endpoint executes plan with advisory lock per stack
- [ ] **PLAN-03**: Apply flow: lock → ensure network → materialize configs → compose up
- [ ] **PLAN-04**: Plan staleness detection (reject apply if stack changed since plan)
- [ ] **PLAN-05**: Structured diff output in plan (+ add, ~ modify, - remove)

### Service Wiring

- [ ] **WIRE-01**: Service export declarations on template catalog (name, type, port, protocol)
- [ ] **WIRE-02**: Service import contracts on template catalog (what a service needs)
- [ ] **WIRE-03**: Wiring cache table (service_instance_wires with partial unique constraints)
- [ ] **WIRE-04**: Auto-wiring for unambiguous cases (one provider matches one consumer)
- [ ] **WIRE-05**: Explicit contract-based wiring for ambiguous cases
- [ ] **WIRE-06**: Wiring diagnostics in plan (missing/ambiguous required contracts)
- [ ] **WIRE-07**: Env var injection from wires (DB_HOST, DB_PORT using internal DNS + container port)
- [ ] **WIRE-08**: Consumer instance env overrides win over injected wire values

### Export/Import

- [ ] **EXIM-01**: Export stack to devarch.yml (stack + instances + overrides; wires added in Phase 8)
- [ ] **EXIM-02**: Import devarch.yml with create-update reconciliation
- [ ] **EXIM-03**: Version field in devarch.yml format
- [ ] **EXIM-04**: Secret redaction in exports (placeholders, not plaintext)
- [ ] **EXIM-05**: Export → import → export round-trip stable (excluding ids/timestamps)
- [ ] **EXIM-06**: Export includes resolved specifics: chosen host ports, image digests/pinned versions, template versions

### Secrets

- [ ] **SECR-01**: Secrets encrypted at rest (AES-256-GCM)
- [ ] **SECR-02**: Auto-generated encryption key in ~/.devarch/secret.key
- [ ] **SECR-03**: Secret redaction in all API responses, plan output, compose previews
- [ ] **SECR-04**: Encrypt before INSERT, decrypt after SELECT (transparent to app logic)

### Resource Limits

- [ ] **RESC-01**: Resource limits per instance (CPU, memory)
- [ ] **RESC-02**: Limits mapped to compose deploy.resources fields
- [ ] **RESC-03**: Limits validated in plan output

### Database Migrations

- [x] **MIGR-01**: Migration 013: stacks table (Phase 2)
- [x] **MIGR-02**: Migration 014: service_instances + all instance override tables (Phase 3)
- [ ] **MIGR-03**: Migration 015: service_exports, service_import_contracts, service_instance_wires (Phase 8)
- [ ] **MIGR-04**: Migration 016: secrets encryption fields (Phase 9)
- [ ] **MIGR-05**: Migration 017: service_instance_resources (Phase 9)

### Bootstrap & Diagnostics

- [ ] **BOOT-01**: `devarch init` — one-command project bootstrap from devarch.yml (pull images, create networks, apply config)
- [ ] **BOOT-02**: `devarch doctor` — environment diagnostics (runtime running, permissions/SELinux, port conflicts, disk space, required tools)

### Lockfile

- [ ] **LOCK-01**: Generate devarch.lock from resolved state (host ports, image digests, pinned versions, template versions)
- [ ] **LOCK-02**: Lock validation on apply — warn/reject when runtime state diverges from lockfile
- [ ] **LOCK-03**: Lock refresh command to update lockfile from current resolved state

## v2 Requirements

### Multi-Stack Operations

- **MULT-01**: Bulk start/stop across multiple stacks
- **MULT-02**: Cross-stack networking (explicit opt-in)
- **MULT-03**: Stack dependency ordering (stack A depends on stack B)

### Advanced Wiring

- **AWIR-01**: Role-based auto-wire priority (devarch.role=primary)
- **AWIR-02**: Wiring graph visualization in dashboard
- **AWIR-03**: Custom env var naming templates for wire injection

### Export/Import Advanced

- **EXIA-01**: Full-reconcile import mode (delete removed items)
- **EXIA-02**: Merge conflict resolution for concurrent imports
- **EXIA-03**: Stack marketplace (shared definitions)

### Security

- **SECU-01**: Key rotation for encryption keys
- **SECU-02**: Team key sharing patterns
- **SECU-03**: API authentication (when team features arrive)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-machine / remote deployment | Local dev tool only |
| Stack-level CI/CD pipelines | Beyond dev environment scope |
| Full vault/HSM secret management | Simple key file sufficient for local dev |
| Stack versioning / rollback | Export/import covers this simply |
| Kubernetes integration | Compose-based tool, K8s is different ecosystem |
| OAuth/SSO for DevArch API | Single-user/local tool |
| Production deployment support | Anti-feature per research — creates false confidence |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BASE-01 | Phase 1 | Complete |
| BASE-02 | Phase 1 | Complete |
| BASE-03 | Phase 1 | Complete |
| STCK-01 | Phase 2 | Complete |
| STCK-02 | Phase 2 | Complete |
| STCK-03 | Phase 2 | Complete |
| STCK-04 | Phase 2 | Complete |
| STCK-05 | Phase 2 | Complete |
| STCK-06 | Phase 2 | Complete |
| STCK-07 | Phase 2 | Complete |
| STCK-08 | Phase 2 | Complete |
| MIGR-01 | Phase 2 | Complete |
| INST-01 | Phase 3 | Complete |
| INST-02 | Phase 3 | Complete |
| INST-03 | Phase 3 | Complete |
| INST-04 | Phase 3 | Complete |
| INST-05 | Phase 3 | Complete |
| INST-06 | Phase 3 | Complete |
| INST-07 | Phase 3 | Complete |
| INST-08 | Phase 3 | Complete |
| INST-09 | Phase 3 | Complete |
| INST-10 | Phase 3 | Complete |
| INST-11 | Phase 3 | Complete |
| INST-12 | Phase 3 | Complete |
| MIGR-02 | Phase 3 | Complete |
| NETW-01 | Phase 4 | Pending |
| NETW-02 | Phase 4 | Pending |
| NETW-03 | Phase 4 | Pending |
| NETW-04 | Phase 4 | Pending |
| COMP-01 | Phase 5 | Pending |
| COMP-02 | Phase 5 | Pending |
| COMP-03 | Phase 5 | Pending |
| PLAN-01 | Phase 6 | Pending |
| PLAN-02 | Phase 6 | Pending |
| PLAN-03 | Phase 6 | Pending |
| PLAN-04 | Phase 6 | Pending |
| PLAN-05 | Phase 6 | Pending |
| EXIM-01 | Phase 7 | Pending |
| EXIM-02 | Phase 7 | Pending |
| EXIM-03 | Phase 7 | Pending |
| EXIM-04 | Phase 7 | Pending |
| EXIM-05 | Phase 7 | Pending |
| EXIM-06 | Phase 7 | Pending |
| BOOT-01 | Phase 7 | Pending |
| BOOT-02 | Phase 7 | Pending |
| LOCK-01 | Phase 7 | Pending |
| LOCK-02 | Phase 7 | Pending |
| LOCK-03 | Phase 7 | Pending |
| WIRE-01 | Phase 8 | Pending |
| WIRE-02 | Phase 8 | Pending |
| WIRE-03 | Phase 8 | Pending |
| WIRE-04 | Phase 8 | Pending |
| WIRE-05 | Phase 8 | Pending |
| WIRE-06 | Phase 8 | Pending |
| WIRE-07 | Phase 8 | Pending |
| WIRE-08 | Phase 8 | Pending |
| MIGR-03 | Phase 8 | Pending |
| SECR-01 | Phase 9 | Pending |
| SECR-02 | Phase 9 | Pending |
| SECR-03 | Phase 9 | Pending |
| SECR-04 | Phase 9 | Pending |
| RESC-01 | Phase 9 | Pending |
| RESC-02 | Phase 9 | Pending |
| RESC-03 | Phase 9 | Pending |
| MIGR-04 | Phase 9 | Pending |
| MIGR-05 | Phase 9 | Pending |

**Coverage:**
- v1 requirements: 66 total
- Mapped to phases: 66
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-03*
*Last updated: 2026-02-06 after phase 3 execution*
