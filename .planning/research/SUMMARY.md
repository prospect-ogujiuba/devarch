# Project Research Summary

**Project:** DevArch - Stacks/Instances/Wiring Milestone
**Domain:** Local microservices development orchestration
**Researched:** 2026-02-03
**Confidence:** HIGH

## Executive Summary

DevArch extends its database-as-source-of-truth architecture to support multi-stack isolation, enabling developers to run multiple independent environments (e.g., Laravel+MySQL stack alongside Django+Postgres stack) without collision. Research validates the approach: use stdlib patterns (PostgreSQL advisory locks, Go crypto, JSONB overrides) with minimal dependencies, follow Docker Compose conventions for familiarity, and implement Terraform-style plan/apply for safety.

The recommended architecture uses copy-on-write instance overrides (store only deltas from templates), per-stack network isolation, deterministic container naming, and contract-based service wiring. Critical differentiators are plan/apply workflow (preview before execution), encrypted secrets (trust signal), and declarative export/import (team sharing). All core patterns have high-confidence implementations via Go stdlib and PostgreSQL native features.

Key risk: container name collisions and network contamination break isolation. Mitigation: enforce `devarch-{stack}-{instance}` naming from day one, create per-stack networks, validate uniqueness at plan stage. Secondary risks: plan staleness (concurrent modifications), config override merge bugs, advisory lock contention. All addressable via established patterns from the PostgreSQL and Docker ecosystems.

## Key Findings

### Recommended Stack

DevArch's existing stack (Go 1.25, PostgreSQL 14+, gopkg.in/yaml.v3, chi router, React 19, TanStack libraries) requires **zero new dependencies** for stacks milestone. Go stdlib handles AES-256-GCM encryption (crypto/aes, crypto/cipher), advisory locks via database/sql (PostgreSQL native feature), and JSON merge logic. Frontend extends existing TanStack Query patterns for stack state management.

**Core technologies:**
- **Go stdlib crypto**: AES-256-GCM for secrets encryption — FIPS-validated, no supply chain risk
- **PostgreSQL advisory locks**: Concurrency control for plan/apply — native since 8.2, auto-release semantics
- **JSONB copy-on-write**: Instance overrides storage — efficient deltas, stdlib JSON marshaling
- **gopkg.in/yaml.v3**: Compose generation — already in use, proven for complex structs

**Avoid:**
- ORMs (GORM/ent) — abstract away SQL, harder to use advisory locks and JSONB operators
- Docker SDK — heavy, API churn; shell out to compose CLI (current approach works)
- etcd/Consul — overkill for local dev; PostgreSQL advisory locks sufficient
- Redis for state — adds dependency, state already in Postgres

### Expected Features

**Must have (table stakes):**
- Stack CRUD with deterministic naming (`devarch-{stack}-{instance}`)
- Per-stack network isolation (one bridge network per stack)
- Service instances with config overrides (copy-on-write pattern)
- Stack-scoped compose generation
- Plan/Apply workflow with advisory locking — **critical differentiator**
- Encrypted secrets at rest — **trust signal for adoption**
- Export/Import stacks as YAML — **enables team sharing**

**Should have (competitive advantage):**
- Service wiring with contracts (explicit dependencies)
- Auto-wiring for unambiguous cases (single provider per contract)
- Visual service graph in dashboard
- Real-time status updates via WebSocket

**Defer (v2+):**
- Auto-restart on file changes (Tilt-style hot reload)
- Dev-prod parity validation
- Service catalog/recipes (LAMP, MEAN templates)
- Database snapshot/restore per stack

### Architecture Approach

System follows layered architecture: API handlers → service layer (config resolver, wiring graph, plan engine) → database (services templates, stack instances, overrides). Copy-on-write pattern: templates in `services` table, instances in `service_instances` with JSONB `compose_overrides` storing only deltas. Config resolution merges at query time (application-level merge) or uses PostgreSQL JSONB operators for shallow merges.

**Major components:**
1. **Config Resolver** — merges template + instance overrides with precedence rules (override > template)
2. **Wiring Graph Builder** — DAG construction, cycle detection, topological sort for startup order
3. **Plan/Apply Engine** — diffs desired vs current state, executes with PostgreSQL advisory locks
4. **Stack Compose Generator** — generates single YAML per stack with resolved configs + wiring

**Key patterns:**
- Advisory locks: `pg_try_advisory_xact_lock()` for transaction-scoped locking (auto-release)
- Stack-scoped compose: one file for entire stack, includes depends_on from wiring graph
- Config file materialization: write to `compose/stacks/{stack}/{instance}/` to avoid races

### Critical Pitfalls

1. **Container name collisions** — without deterministic naming, multiple stacks create same container names. Prevention: enforce `devarch-{stack}-{instance}` pattern, validate uniqueness in plan stage, add DB constraint.

2. **Network contamination** — shared network allows cross-stack access. Prevention: per-stack bridge network (`devarch-{stack}-net`), create in compose (not external), no inter-stack routing by default.

3. **Plan staleness** — concurrent modifications between plan and apply. Prevention: include stack state hash in plan, validate on apply, reject stale plans with clear error.

4. **Config override merge bugs** — shallow vs deep merge surprises users. Prevention: explicit merge strategy (ports/volumes/env replace entire list, healthcheck replaces entire object), document in API, show effective config in plan.

5. **Advisory lock held during slow operations** — blocks concurrent requests. Prevention: lock only critical sections (not during compose up), use timeouts, release on cancellation, show lock holder in dashboard.

## Implications for Roadmap

Based on research, 6-phase structure with early phases delivering isolation foundation and later phases adding sophistication.

### Phase 0: Pre-work Refactor
**Rationale:** Fix existing hardcoded runtime commands before stack implementation inherits bug
**Delivers:** Container client abstraction enforced, all operations use client (not exec.Command)
**Avoids:** Pitfall #3 (runtime bypass breaks Docker compatibility)
**Research needs:** None (internal refactor)

### Phase 1: Stack Foundation
**Rationale:** Isolation primitives must work before instances, otherwise name/network collisions break everything
**Delivers:** Stack CRUD, deterministic naming, per-stack networks, basic compose generation
**Addresses:** Table stakes features (naming, networking, compose)
**Avoids:** Pitfalls #1 (naming), #2 (network), #10 (runtime compat)
**Stack elements:** PostgreSQL stacks table, chi handlers, gopkg.in/yaml.v3
**Research needs:** None (patterns well-established)

### Phase 2: Instances & Overrides
**Rationale:** Copy-on-write pattern enables reusable templates, must come before wiring (needs instance configs)
**Delivers:** Service instances table, override tables, config resolution, merged config API
**Addresses:** Service instance templates (differentiator)
**Avoids:** Pitfalls #5 (merge bugs), #13 (config file races)
**Stack elements:** JSONB for overrides, application-level merge
**Architecture:** Config Resolver component
**Research needs:** Test merge edge cases during implementation (LOW effort)

### Phase 3: Plan/Apply Workflow
**Rationale:** Safety mechanism differentiates DevArch, depends on instance config resolution
**Delivers:** Plan generation, diff algorithm, advisory locking, apply executor
**Addresses:** Plan/Apply workflow (critical differentiator)
**Avoids:** Pitfalls #4 (staleness), #11 (lock contention)
**Stack elements:** PostgreSQL advisory locks, plan state table
**Architecture:** Plan/Apply Engine component
**Research needs:** Advisory lock timeout patterns, plan diff algorithm edge cases (MEDIUM effort)

### Phase 4: Secrets Encryption
**Rationale:** Difficult to retrofit after schema locked, trust signal for adoption
**Delivers:** AES-256-GCM encryption, key management, redaction in exports
**Addresses:** Encrypted secrets (trust signal)
**Avoids:** Pitfall #6 (plaintext leaks)
**Stack elements:** crypto/aes, crypto/cipher stdlib
**Research needs:** None (stdlib crypto patterns documented in STACK.md)

### Phase 5: Service Wiring
**Rationale:** Complex feature, decoupled from earlier phases, requires instance configs
**Delivers:** Wiring contracts, instance-to-instance wiring, graph builder, auto-wiring rules
**Addresses:** Automatic service wiring (differentiator), contract-based discovery
**Avoids:** Pitfalls #7 (contract validation), #8 (ambiguity)
**Stack elements:** Graph algorithms (DAG, topological sort)
**Architecture:** Wiring Graph Builder component
**Research needs:** Auto-wiring heuristics validation (MEDIUM effort, needs domain testing)

### Phase 6: Export/Import
**Rationale:** Adoption driver, depends on all prior phases (must export complete stack state)
**Delivers:** devarch.yml format with versioning, export with secret handling, import with validation
**Addresses:** Export/Import portable definitions (adoption driver)
**Avoids:** Pitfall #12 (schema evolution)
**Stack elements:** gopkg.in/yaml.v3, format versioning
**Research needs:** None (YAML patterns proven)

### Phase Ordering Rationale

- **Foundation first:** Naming and networking isolation must work before anything else (Phases 0-1)
- **Instances before wiring:** Can't wire services until instances exist with resolved configs (Phase 2 before 5)
- **Plan/Apply early:** Safety mechanism needed before complex features add more state (Phase 3 before 5-6)
- **Secrets before export:** Encryption painful to retrofit, export depends on secret handling (Phase 4 before 6)
- **Wiring independent:** Complex but decoupled, can be Phase 5 or swapped with 4
- **Export last:** Requires all state (instances, secrets, wiring) to export complete stacks

### Research Flags

**Phases needing deeper research during planning:**
- **Phase 3 (Plan/Apply):** Advisory lock patterns under concurrent load need profiling; plan diff edge cases (array merge, nested object changes)
- **Phase 5 (Wiring):** Auto-wiring heuristics need validation with real service types; ambiguity detection algorithm needs testing

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation):** Docker naming conventions, network creation well-documented
- **Phase 2 (Instances):** JSONB merge patterns standard in PostgreSQL apps
- **Phase 4 (Secrets):** AES-256-GCM implementation in STACK.md covers it
- **Phase 6 (Export/Import):** YAML marshal/unmarshal patterns already proven in codebase

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Zero new dependencies, stdlib patterns proven, existing codebase already uses recommended libraries |
| Features | MEDIUM | Table stakes well-known from Docker Compose; differentiators (plan/apply, wiring) based on Terraform/Kubernetes patterns but adapted |
| Architecture | HIGH | Copy-on-write, advisory locks, graph algorithms are established patterns with PostgreSQL/Go implementations |
| Pitfalls | HIGH | Based on existing codebase analysis (generator.go, container client), domain experience (container naming, network isolation) |

**Overall confidence:** HIGH

### Gaps to Address

- **Advisory lock timeout tuning:** Research shows pattern, but optimal timeout (30s? 60s?) needs profiling during Phase 3 implementation
- **Auto-wiring heuristics:** Contract-based wiring pattern clear, but "when to auto-wire vs require explicit" needs real-world service type testing
- **Config override merge edge cases:** Merge strategy defined (replace vs deep merge), but healthcheck/depends_on nested structures need test coverage
- **Performance at scale:** Research assumes <100 stacks; config resolution caching strategy deferred to profiling results

**Resolution strategy:**
- Advisory lock timeout: start with 30s, adjust based on image pull times in Phase 3 testing
- Auto-wiring: Phase 5 begins with conservative "only unambiguous cases" rule, expand based on user feedback
- Override merge: Phase 2 includes comprehensive merge test suite before compose generation integration
- Performance: Phase 1-2 establish baseline metrics, flag if resolution >100ms

## Sources

### Primary (HIGH confidence)
- DevArch codebase: `/home/fhcadmin/projects/devarch/api/go.mod` (dependencies), `internal/compose/generator.go` (existing patterns), `CLAUDE.md` (architecture context)
- PostgreSQL official docs: Advisory locks (https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS)
- Go stdlib docs: crypto/aes, crypto/cipher (AES-256-GCM patterns), database/sql (advisory lock interface)
- Docker Compose specification: https://docs.docker.com/compose/compose-file/ (compose generation patterns)

### Secondary (MEDIUM confidence)
- Terraform plan/apply workflow: Pattern inspiration (https://www.terraform.io/docs/cli/commands/plan.html)
- Kubernetes service discovery: Contract-based wiring analogies
- Docker Compose override files: Copy-on-write pattern validation (https://docs.docker.com/compose/multiple-compose-files/extends/)

### Tertiary (LOW confidence, needs validation)
- Competitor features (Docker Compose profiles, Tilt contexts, DDEV recipes): Training knowledge as of Jan 2025, may have evolved
- Auto-wiring heuristics: Inferred from Kubernetes service mesh patterns, needs domain-specific validation

---
*Research completed: 2026-02-03*
*Ready for roadmap: yes*
