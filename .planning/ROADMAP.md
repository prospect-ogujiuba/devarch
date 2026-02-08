# Roadmap: DevArch Stacks & Instances Milestone

## Overview

Transform DevArch from single-service orchestration to multi-stack isolation, enabling developers to run multiple independent environments (Laravel+MySQL, Django+Postgres) without collision. Phases progress from foundation (naming, networking) through instances and config resolution, to export/import portability (validated early to prove the "hand it to a teammate" loop), then wiring and secrets. Core primitive: two stacks using the same service template must never collide. Portability promise: devarch.yml + devarch.lock + `devarch init` = reproducible environment on any machine.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation & Guardrails** - Identity system, validation, runtime abstraction
- [x] **Phase 2: Stack CRUD** - Stack management API + dashboard UI
- [x] **Phase 3: Service Instances** - Instance overrides + config resolution
- [x] **Phase 4: Network Isolation** - Per-stack networks, deterministic naming
- [x] **Phase 5: Compose Generation** - Stack-scoped YAML generation
- [ ] **Phase 6: Plan/Apply Workflow** - Safety mechanism with advisory locking
- [ ] **Phase 7: Export/Import & Bootstrap** - Shareable devarch.yml, lockfile, init/doctor
- [ ] **Phase 8: Service Wiring** - Contract-based auto-wiring + explicit wiring
- [ ] **Phase 9: Secrets & Resources** - Encryption + resource limits

## Phase Details

### Phase 1: Foundation & Guardrails
**Goal**: Establish isolation primitives (identity labels, validation, runtime abstraction) that all stack operations depend on
**Depends on**: Nothing (first phase)
**Requirements**: BASE-01, BASE-02, BASE-03
**Success Criteria** (what must be TRUE):
  1. Identity label constants exist and are used consistently (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
  2. Stack and instance names are validated before creation (charset, length, uniqueness)
  3. All container operations route through container.Client (no hardcoded podman exec.Command)
  4. Runtime abstraction works for both Docker and Podman
**Plans**: 2 plans

Plans:
- [x] 01-01-PLAN.md — Labels, validation, types, and DB migration (COMPLETE 2026-02-03)
- [x] 01-02-PLAN.md — Runtime abstraction refactor and hardcoded call elimination (COMPLETE 2026-02-03)

### Phase 2: Stack CRUD
**Goal**: Users can create and manage stacks via API and dashboard
**Depends on**: Phase 1
**Requirements**: STCK-01, STCK-02, STCK-03, STCK-04, STCK-05, STCK-06, STCK-07, STCK-08, MIGR-01
**Success Criteria** (what must be TRUE):
  1. User can create stack with name and description via API and dashboard
  2. User can list all stacks with status summary (instance count, running count)
  3. User can view stack detail page showing instances and network info
  4. User can edit stack metadata (description only; stack name is immutable ID)
  5. User can "rename" via clone: clone stack to new name, then optionally delete old stack
  6. User can delete stack and all resources cascade (containers, instances, network)
  7. User can enable/disable stack without deleting it
  8. User can clone stack with new name (copies instances + overrides)
**Plans**: 5 plans

Plans:
- [x] 02-01-PLAN.md — Migration soft-delete + Stack handler core CRUD (COMPLETE 2026-02-03)
- [x] 02-02-PLAN.md — Advanced ops (enable/disable/clone/rename/trash) + route wiring (COMPLETE 2026-02-03)
- [x] 02-03-PLAN.md — Dashboard types, query hooks, navigation (COMPLETE 2026-02-03)
- [x] 02-04-PLAN.md — Stack list page (grid + table views, empty state, toolbar) (COMPLETE 2026-02-03)
- [x] 02-05-PLAN.md — Stack detail page + all action dialogs (COMPLETE 2026-02-03)

### Phase 3: Service Instances
**Goal**: Users can create service instances from templates with full copy-on-write overrides
**Depends on**: Phase 2
**Requirements**: INST-01, INST-02, INST-03, INST-04, INST-05, INST-06, INST-07, INST-08, INST-09, INST-10, INST-11, INST-12, MIGR-02
**Success Criteria** (what must be TRUE):
  1. User can add service instance to stack from template catalog via dashboard
  2. User can override instance ports, volumes, env vars, dependencies, labels, domains, healthchecks, config files
  3. Effective config API returns merged template + overrides (overrides win)
  4. User can view effective config before applying (no surprises)
  5. User can edit instance overrides after creation
  6. User can remove instance from stack
**Plans**: 5 plans

Plans:
- [x] 03-01-PLAN.md — Migration 014 (override tables) + instance CRUD handler (COMPLETE 2026-02-04)
- [x] 03-02-PLAN.md — Override PUT handlers + effective config resolver (COMPLETE 2026-02-04)
- [x] 03-03-PLAN.md — Dashboard types, hooks, add-instance dialog, instance list (COMPLETE 2026-02-04)
- [x] 03-04-PLAN.md — Instance detail page with tabbed override editors (COMPLETE 2026-02-04)
- [x] 03-05-PLAN.md — Effective config tab + instance lifecycle actions (COMPLETE 2026-02-04)

### Phase 4: Network Isolation
**Goal**: Each stack runs on isolated network with deterministic container naming (no cross-stack contamination). Implementation must be runtime-agnostic: Docker and Podman both get per-stack isolated networks and the same DNS/service-discovery semantics.
**Depends on**: Phase 3
**Requirements**: NETW-01, NETW-02, NETW-03, NETW-04
**Success Criteria** (what must be TRUE):
  1. Containers use deterministic naming: devarch-{stack}-{instance}
  2. Each stack has dedicated bridge network: devarch-{stack}-net
  3. Network is created automatically before containers start (Docker + Podman)
  4. All stack containers have identity labels (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
  5. Two stacks using same template never collide on names or networks
**Plans**: 2 plans

Plans:
- [x] 04-01-PLAN.md — Container client network methods, name validation, network_name auto-population, network status API (COMPLETE 2026-02-06)
- [x] 04-02-PLAN.md — Effective config identity label injection, dashboard network UI (detail card + list badges) (COMPLETE 2026-02-06)

### Phase 5: Compose Generation
**Goal**: Stack compose generator produces single YAML with all instances, replacing per-service generation
**Depends on**: Phase 4
**Requirements**: COMP-01, COMP-02, COMP-03
**Success Criteria** (what must be TRUE):
  1. Stack compose endpoint returns single YAML with N services from effective configs
  2. Config files materialize to compose/stacks/{stack}/{instance}/ (no race conditions)
  3. Existing single-service compose generation still works (backward compatibility)
  4. Generated compose includes proper network references and depends_on
**Plans**: 2 plans

Plans:
- [x] 05-01-PLAN.md — Stack compose generator, config materialization, handler + route wiring (COMPLETE 2026-02-07)
- [x] 05-02-PLAN.md — Dashboard compose tab with CodeMirror preview, warnings, download (COMPLETE 2026-02-07)

### Phase 6: Plan/Apply Workflow
**Goal**: Users preview changes before applying (Terraform-style safety), with advisory locking preventing concurrent modifications
**Depends on**: Phase 5
**Requirements**: PLAN-01, PLAN-02, PLAN-03, PLAN-04, PLAN-05
**Success Criteria** (what must be TRUE):
  1. Plan endpoint returns structured preview (+ add, ~ modify, - remove)
  2. User can review plan before applying (dashboard shows diff)
  3. Apply endpoint executes plan with advisory lock (one apply per stack at a time)
  4. Stale plans are rejected if stack changed since plan creation
  5. Apply flow: lock → ensure network → materialize configs → compose up
**Plans**: 3 plans

Plans:
- [ ] 06-01-PLAN.md — Plan domain package: types, diff computation, staleness detection
- [ ] 06-02-PLAN.md — Plan + Apply API endpoints with advisory locking and route wiring
- [ ] 06-03-PLAN.md — Dashboard Deploy tab with diff visualization and apply execution

### Phase 7: Export/Import & Bootstrap
**Goal**: Users export stacks to devarch.yml (with resolved specifics) for sharing, import with reconciliation, lockfile for deterministic reproduction, and one-command bootstrap + diagnostics
**Depends on**: Phase 6
**Requirements**: EXIM-01, EXIM-02, EXIM-03, EXIM-04, EXIM-05, EXIM-06, BOOT-01, BOOT-02, LOCK-01, LOCK-02, LOCK-03
**Success Criteria** (what must be TRUE):
  1. Export produces devarch.yml with stack, instances, overrides (no wires yet — Phase 8)
  2. Export includes resolved specifics: host ports, image digests/pinned versions, template versions
  3. Import creates or updates stack from devarch.yml (create-update mode)
  4. devarch.yml includes version field for format evolution
  5. Secrets are redacted in exports (placeholders, not plaintext)
  6. Export → import → export round-trip is stable (excluding ids/timestamps)
  7. `devarch init` bootstraps from devarch.yml (pull images, create networks, apply config)
  8. `devarch doctor` checks runtime, permissions, port conflicts, disk, required tools
  9. devarch.lock generated from resolved state (ports, digests, template versions)
  10. Apply warns when runtime diverges from lockfile
**Plans**: TBD

Plans:
- [ ] 07-01-PLAN.md: TBD
- [ ] 07-02-PLAN.md: TBD
- [ ] 07-03-PLAN.md: TBD

### Phase 8: Service Wiring
**Goal**: Services automatically discover dependencies via contracts (auto-wiring for simple cases, explicit wiring for ambiguous)
**Depends on**: Phase 7
**Requirements**: WIRE-01, WIRE-02, WIRE-03, WIRE-04, WIRE-05, WIRE-06, WIRE-07, WIRE-08, MIGR-03
**Success Criteria** (what must be TRUE):
  1. Template services declare exports (name, type, port, protocol)
  2. Template services declare import contracts (what they need)
  3. Auto-wiring connects unambiguous provider-consumer pairs
  4. Explicit wiring UI handles ambiguous cases (multiple PostgreSQL instances)
  5. Plan output shows missing or ambiguous required contracts
  6. Consumer instances receive env vars from wires (DB_HOST, DB_PORT using internal DNS)
  7. Consumer instance env overrides win over injected wire values
  8. Wires included in devarch.yml export (re-export after wiring)
**Plans**: TBD

Plans:
- [ ] 08-01-PLAN.md: TBD
- [ ] 08-02-PLAN.md: TBD
- [ ] 08-03-PLAN.md: TBD
- [ ] 08-04-PLAN.md: TBD

### Phase 9: Secrets & Resources
**Goal**: Secrets encrypted at rest, resource limits per instance, all sensitive data redacted in outputs
**Depends on**: Phase 8
**Requirements**: SECR-01, SECR-02, SECR-03, SECR-04, RESC-01, RESC-02, RESC-03, MIGR-04, MIGR-05
**Success Criteria** (what must be TRUE):
  1. Secrets encrypted at rest using AES-256-GCM
  2. Encryption key auto-generated in ~/.devarch/secret.key on first use
  3. Secrets redacted in all API responses, plan output, compose previews, exports
  4. Encryption is transparent to app logic (encrypt before INSERT, decrypt after SELECT)
  5. User can set resource limits per instance (CPU, memory)
  6. Resource limits appear in compose deploy.resources fields
  7. Plan output shows resource limits for validation
**Plans**: TBD

Plans:
- [ ] 09-01-PLAN.md: TBD
- [ ] 09-02-PLAN.md: TBD
- [ ] 09-03-PLAN.md: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8 -> 9

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Guardrails | 2/2 | Complete | 2026-02-03 |
| 2. Stack CRUD | 5/5 | Complete | 2026-02-03 |
| 3. Service Instances | 5/5 | Complete | 2026-02-04 (verified 2026-02-06) |
| 4. Network Isolation | 2/2 | Complete | 2026-02-06 |
| 5. Compose Generation | 2/2 | Complete | 2026-02-07 |
| 6. Plan/Apply Workflow | 0/3 | Not started | - |
| 7. Export/Import & Bootstrap | 0/TBD | Not started | - |
| 8. Service Wiring | 0/TBD | Not started | - |
| 9. Secrets & Resources | 0/TBD | Not started | - |

---
*Created: 2026-02-03*
*Last updated: 2026-02-07 after phase 6 planning*
