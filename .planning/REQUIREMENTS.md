# Requirements: DevArch Stacks & Instances

**Defined:** 2026-02-03
**Core Value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.

## v1 Requirements

### Baseline & Guardrails

- [ ] **BASE-01**: DevArch identity label constants defined (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
- [ ] **BASE-02**: Stack/instance name validation helpers (charset, length, uniqueness)
- [ ] **BASE-03**: Runtime abstraction fix — all container operations route through container.Client, no hardcoded podman exec.Command

### Stack Management

- [ ] **STCK-01**: Create stack with name, description, optional network name
- [ ] **STCK-02**: List stacks with status summary (instance count, running count)
- [ ] **STCK-03**: Get stack detail (instances, network, last applied)
- [ ] **STCK-04**: Update stack metadata (name, description)
- [ ] **STCK-05**: Delete stack (with cascade: stop containers, remove instances, remove network)
- [ ] **STCK-06**: Enable/disable stack without deleting
- [ ] **STCK-07**: Clone stack with new name (copies instances + overrides)
- [ ] **STCK-08**: Dashboard UI for stack CRUD (list, create, detail, edit, delete, clone)

### Service Instances

- [ ] **INST-01**: Create instance from template service within a stack
- [ ] **INST-02**: Override instance ports (copy-on-write from template)
- [ ] **INST-03**: Override instance volumes
- [ ] **INST-04**: Override instance environment variables
- [ ] **INST-05**: Override instance dependencies
- [ ] **INST-06**: Override instance labels
- [ ] **INST-07**: Override instance domains
- [ ] **INST-08**: Override instance healthchecks
- [ ] **INST-09**: Override instance config files
- [ ] **INST-10**: Effective config resolver (template + overrides merged, overrides win)
- [ ] **INST-11**: List/get/update/delete instances within a stack
- [ ] **INST-12**: Dashboard UI for instance management (add to stack, edit overrides, remove)

### Network Isolation

- [ ] **NETW-01**: Deterministic container naming: devarch-{stack}-{instance}
- [ ] **NETW-02**: Per-stack bridge network: devarch-{stack}-net
- [ ] **NETW-03**: EnsureNetwork in container client (create if not exists, Docker + Podman)
- [ ] **NETW-04**: Identity labels injected on all stack containers

### Compose Generation

- [ ] **COMP-01**: Stack compose generator produces one YAML with N services from effective configs
- [ ] **COMP-02**: Stack-scoped materialization path: compose/stacks/{stack}/{instance}/
- [ ] **COMP-03**: Existing single-service generator continues working (backward compat)

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

- [ ] **EXIM-01**: Export stack to devarch.yml (stack + instances + overrides + wires)
- [ ] **EXIM-02**: Import devarch.yml with create-update reconciliation
- [ ] **EXIM-03**: Version field in devarch.yml format
- [ ] **EXIM-04**: Secret redaction in exports (placeholders, not plaintext)
- [ ] **EXIM-05**: Export → import → export round-trip stable (excluding ids/timestamps)

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

- [ ] **MIGR-01**: Migration 013: service_exports, service_import_contracts
- [ ] **MIGR-02**: Migration 014: stacks, service_instances, all instance override tables
- [ ] **MIGR-03**: Migration 015: service_instance_wires
- [ ] **MIGR-04**: Migration 016: secrets encryption fields
- [ ] **MIGR-05**: Migration 017: service_instance_resources

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
| BASE-01 | Phase 1 | Pending |
| BASE-02 | Phase 1 | Pending |
| BASE-03 | Phase 1 | Pending |
| STCK-01 | Phase 2 | Pending |
| STCK-02 | Phase 2 | Pending |
| STCK-03 | Phase 2 | Pending |
| STCK-04 | Phase 2 | Pending |
| STCK-05 | Phase 2 | Pending |
| STCK-06 | Phase 2 | Pending |
| STCK-07 | Phase 2 | Pending |
| STCK-08 | Phase 2 | Pending |
| INST-01 | Phase 2 | Pending |
| INST-02 | Phase 2 | Pending |
| INST-03 | Phase 2 | Pending |
| INST-04 | Phase 2 | Pending |
| INST-05 | Phase 2 | Pending |
| INST-06 | Phase 2 | Pending |
| INST-07 | Phase 2 | Pending |
| INST-08 | Phase 2 | Pending |
| INST-09 | Phase 2 | Pending |
| INST-10 | Phase 2 | Pending |
| INST-11 | Phase 2 | Pending |
| INST-12 | Phase 2 | Pending |
| NETW-01 | Phase 3 | Pending |
| NETW-02 | Phase 3 | Pending |
| NETW-03 | Phase 3 | Pending |
| NETW-04 | Phase 3 | Pending |
| COMP-01 | Phase 4 | Pending |
| COMP-02 | Phase 4 | Pending |
| COMP-03 | Phase 4 | Pending |
| PLAN-01 | Phase 4 | Pending |
| PLAN-02 | Phase 4 | Pending |
| PLAN-03 | Phase 4 | Pending |
| PLAN-04 | Phase 4 | Pending |
| PLAN-05 | Phase 4 | Pending |
| WIRE-01 | Phase 5 | Pending |
| WIRE-02 | Phase 5 | Pending |
| WIRE-03 | Phase 5 | Pending |
| WIRE-04 | Phase 5 | Pending |
| WIRE-05 | Phase 5 | Pending |
| WIRE-06 | Phase 5 | Pending |
| WIRE-07 | Phase 5 | Pending |
| WIRE-08 | Phase 5 | Pending |
| EXIM-01 | Phase 6 | Pending |
| EXIM-02 | Phase 6 | Pending |
| EXIM-03 | Phase 6 | Pending |
| EXIM-04 | Phase 6 | Pending |
| EXIM-05 | Phase 6 | Pending |
| SECR-01 | Phase 7 | Pending |
| SECR-02 | Phase 7 | Pending |
| SECR-03 | Phase 7 | Pending |
| SECR-04 | Phase 7 | Pending |
| RESC-01 | Phase 7 | Pending |
| RESC-02 | Phase 7 | Pending |
| RESC-03 | Phase 7 | Pending |
| MIGR-01 | Phase 2 | Pending |
| MIGR-02 | Phase 2 | Pending |
| MIGR-03 | Phase 5 | Pending |
| MIGR-04 | Phase 7 | Pending |
| MIGR-05 | Phase 7 | Pending |

**Coverage:**
- v1 requirements: 48 total
- Mapped to phases: 48
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-03*
*Last updated: 2026-02-03 after initial definition*
