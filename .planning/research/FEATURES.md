# Feature Research: Service Composition & Orchestration Tools

**Domain:** Local development environment orchestration with service composition
**Researched:** 2026-02-03
**Confidence:** MEDIUM (based on training knowledge of Docker Compose, Tilt, DDEV, Lando, Docksal, Laragon)

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Start/stop service groups | Core orchestration primitive | LOW | Basic compose up/down operations |
| Service-to-service networking | Services must communicate | LOW | Named bridge networks with DNS |
| Port mapping management | Access services from host | LOW | Prevent collisions, predictable assignments |
| Environment variable injection | Configuration primitive | LOW | Per-service env vars |
| Volume mounting | Persist data, mount code | LOW | Named volumes + bind mounts |
| Service dependencies | Startup ordering | LOW | depends_on with wait conditions |
| Container lifecycle (restart policies) | Resilience to crashes | LOW | restart: unless-stopped |
| Service logs access | Debugging primitive | LOW | docker/podman logs equivalent |
| Health checks | Know when services are ready | MEDIUM | HTTP/TCP/exec probes |
| Configuration file mounting | App configs, nginx.conf, etc | LOW | Bind mount or copy pattern |
| Resource limits (memory/CPU) | Prevent runaway processes | LOW | Compose mem_limit/cpus fields |
| Named networks | Isolation between groups | LOW | One network per stack |
| Service discovery via DNS | Use service names as hostnames | LOW | Built-in to compose networking |
| Declarative config file | Infrastructure as code | LOW | YAML/config file defining stack |
| CLI for basic operations | Developer workflow | MEDIUM | up/down/logs/status commands |

### Differentiators (Competitive Advantage)

Features that set products apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Plan/Apply workflow (Terraform-style) | Preview before execution, reduces mistakes | MEDIUM | Show diff, require confirmation |
| Service instance templates | Reusable definitions with per-instance overrides | MEDIUM | Copy-on-write pattern |
| Automatic service wiring | Zero-config connectivity for common patterns | HIGH | Contract-based discovery + env injection |
| Stack isolation (multi-instance) | Run multiple copies without collision | MEDIUM | Naming prefixes, network isolation, port shifting |
| Export/Import portable definitions | Share environments, backup state | MEDIUM | Serialization with secret handling |
| Visual service graph | Understand dependencies at a glance | MEDIUM | Directed graph of services + wires |
| Encrypted secrets at rest | Trust signal for adoption | MEDIUM | AES-256-GCM with key management |
| Real-time status updates | Modern UX expectation | MEDIUM | WebSocket/SSE for container events |
| Recipe/template catalog | Quick-start for common stacks | LOW | Pre-defined service bundles (LAMP, MEAN, etc) |
| Smart defaults per stack type | Reduce boilerplate | MEDIUM | Framework-aware config generation |
| Dev-prod parity validation | Catch environment drift early | HIGH | Compare stack config to production |
| Auto-restart on file changes | Tight dev loop | HIGH | Watch filesystem, trigger rebuilds (Tilt specialty) |
| Multi-language runtime support | One tool for polyglot teams | MEDIUM | PHP/Node/Python/Go templates with version switching |
| One-command setup | Zero-friction onboarding | MEDIUM | `devarch init` creates entire stack |
| Proxy/ingress auto-config | Pretty URLs without manual nginx | MEDIUM | Automatic reverse proxy generation |
| Database snapshot/restore | Seed data management | MEDIUM | Quick state resets for testing |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Production deployment support | "Use same tool for prod" | Dev and prod have different constraints; blurs security boundaries | Keep focused on local dev; use proper CD tools for prod |
| Full orchestration (K8s-style) | "Need complex scheduling" | Massive complexity for local dev; slow startup | Simple compose is faster and sufficient for dev |
| Automatic port allocation (dynamic) | "No port conflicts" | Non-deterministic; breaks bookmarks and documentation | Deterministic port shifting per stack (stack1: 8100, stack2: 8200) |
| Shared services across stacks | "Save resources" | Breaks isolation; version conflicts; debugging nightmares | Disk is cheap, isolation is priceless |
| GUI-only configuration | "Easier than YAML" | Not version-controllable; hard to review; team collaboration breaks | GUI for visualization + YAML for source of truth |
| Auto-update containers | "Stay current" | Breaking changes mid-dev; inconsistent team environments | Explicit version pins + opt-in updates |
| Cloud-native features (service mesh, etc) | "Mirrors production" | Complexity explosion; slow; local dev doesn't need it | Use simpler networking; reserve complexity for production |
| Everything over HTTP/REST | "Modern APIs" | WebSocket needed for real-time; plan output needs clarity | HTTP for CRUD, WebSocket for events, CLI for plan/apply |
| Full secret vault integration | "Secure like prod" | Overkill for local single-user tool; adds external dependencies | File-based encryption with local key, optional vault for teams |

## Feature Dependencies

```
Container Orchestration (table stakes)
    └──requires──> Service Definitions
                      └──requires──> Config Storage (DB/files)

Stack Isolation
    └──requires──> Deterministic Naming
    └──requires──> Network Isolation
    └──requires──> Port Management

Service Wiring (auto)
    └──requires──> Contract Definitions (exports/imports)
    └──requires──> Service Discovery
    └──requires──> Env Var Injection

Plan/Apply Workflow
    └──requires──> State Tracking (desired vs actual)
    └──requires──> Diff Engine
    └──requires──> Advisory Locking

Export/Import
    └──requires──> Serialization Format
    └──requires──> Secret Handling (redaction or encryption)
    └──enhances──> Stack Isolation (portable stacks)

Encrypted Secrets
    └──requires──> Key Management
    └──requires──> Encryption/Decryption Layer
    └──conflicts──> Plaintext Exports (must redact)

Visual Service Graph
    └──requires──> Dependency Metadata
    └──enhances──> Wiring Diagnostics

Real-time Status
    └──requires──> Event Stream (WebSocket/SSE)
    └──requires──> Container Monitoring
```

### Dependency Notes

- **Stack Isolation requires Deterministic Naming:** Without predictable container names, isolation breaks (name collisions, wrong container targeted)
- **Service Wiring requires Contract Definitions:** Auto-wiring only works when services declare what they provide/need (postgres exports `database`, app imports `database`)
- **Plan/Apply requires State Tracking:** Must compare desired (DB/config) vs actual (running containers) to show diff
- **Export/Import enhances Stack Isolation:** Portable stacks enable team sharing and backup, but only useful if stacks truly isolate
- **Encrypted Secrets conflicts with Plaintext Exports:** Can't have both; must redact or encrypt in exports

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] Service definitions with full config (ports, volumes, env, deps, health, config files) — DevArch already has this
- [ ] Stack CRUD (create, list, update, delete stacks)
- [ ] Service instances from templates with override tables
- [ ] Deterministic naming: devarch-{stack}-{instance}
- [ ] Per-stack network isolation
- [ ] Stack compose generator (one YAML per stack)
- [ ] Basic CLI: stack up/down/logs/status
- [ ] Dashboard UI for stack management
- [ ] Plan/Apply for stack changes — **critical differentiator**
- [ ] Encrypted secrets (even if simple) — **trust signal**
- [ ] Export/Import for sharing — **adoption driver**

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Auto-wiring (simple cases: one provider, one consumer)
- [ ] Contract-based wiring (explicit for ambiguous cases)
- [ ] Wiring diagnostics in plan output
- [ ] Visual service graph in dashboard
- [ ] Database snapshot/restore per stack
- [ ] Service template catalog with recipes (LAMP, MEAN, microservices)
- [ ] One-command stack initialization (`devarch stack create --recipe=laravel`)

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Auto-restart on file changes (Tilt-style live updates)
- [ ] Dev-prod parity validation
- [ ] Multi-host stacks (experimental, carefully scoped)
- [ ] Plugin system for custom service types
- [ ] Team features (shared stack library, discovery)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Stack CRUD + Isolation | HIGH | MEDIUM | P1 |
| Service instances with overrides | HIGH | MEDIUM | P1 |
| Deterministic naming | HIGH | LOW | P1 |
| Per-stack networks | HIGH | LOW | P1 |
| Plan/Apply workflow | HIGH | MEDIUM | P1 |
| Encrypted secrets | MEDIUM | MEDIUM | P1 |
| Export/Import | HIGH | MEDIUM | P1 |
| Dashboard for stacks | HIGH | HIGH | P1 |
| Auto-wiring | MEDIUM | HIGH | P2 |
| Visual service graph | MEDIUM | MEDIUM | P2 |
| Contract-based wiring | MEDIUM | MEDIUM | P2 |
| Wiring diagnostics | MEDIUM | LOW | P2 |
| Service catalog/recipes | LOW | MEDIUM | P2 |
| Auto-restart on changes | LOW | HIGH | P3 |
| Dev-prod parity checks | LOW | HIGH | P3 |
| Multi-host stacks | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch — core value proposition
- P2: Should have, add when possible — enhances core
- P3: Nice to have, future consideration — not critical

## Competitor Feature Analysis

| Feature | Docker Compose | Tilt | DDEV/Lando/Docksal | Laragon | DevArch Approach |
|---------|---------------|------|-------------------|---------|------------------|
| Multi-stack isolation | Profiles (weak) | Contexts | Projects (per-dir) | Per-project dirs | Strong: per-stack networks + naming |
| Service templates | None | None | Recipes | Presets | Full copy-on-write overrides |
| Auto-wiring | Manual links/networks | None | Some via recipes | Auto for common stacks | Contract-based discovery |
| Plan/Apply | None | None | None | None | **Differentiator** |
| Secrets | Plaintext env/files | Plaintext | Plaintext | Plaintext | **Encrypted at rest** |
| Export/Import | docker-compose.yml | Tiltfile | .ddev/config.yml | Manual | **devarch.yml with secrets** |
| Real-time status | CLI polling | Rich TUI | CLI polling | GUI | WebSocket to dashboard |
| Visual graph | None | None | None | None | Planned (P2) |
| Resource limits | Native compose | Native K8s | Native compose | None | Native compose fields |
| Cross-platform | Yes | Yes (K8s focus) | Yes | Windows only | Yes (Podman/Docker) |

### Competitor Insights

**Docker Compose:** Industry standard for basic orchestration, but weak multi-stack support (profiles are an afterthought), no plan/apply, secrets are plaintext.

**Tilt:** Developer-focused with live updates and rich TUI, but Kubernetes-oriented (heavy for local dev), no isolation between projects, configuration is code not data.

**DDEV/Lando/Docksal:** Project-per-directory model with recipes for common stacks (Drupal, WordPress, Laravel). Good DX but limited to their recipe catalog, no multi-stack in one workspace, secrets plaintext.

**Laragon:** Windows-only GUI tool with excellent auto-configuration for WAMP/LEMP stacks. Great UX but not version-controllable, no containerization (uses native Windows binaries), no team sharing.

**DevArch differentiators:**
1. **Multi-stack in one workspace with true isolation** — run stack1 (Laravel + MySQL) and stack2 (Django + Postgres) simultaneously without collision
2. **Plan/Apply for safety** — see what will change before applying (Terraform-style workflow)
3. **Encrypted secrets** — trust signal even for local dev
4. **Declarative export/import** — share exact stack definitions with team or across machines
5. **Database as source of truth** — UI and CLI both work with same state, no file sync issues

## Adoption Drivers

Features that directly impact adoption (sorted by impact):

1. **Plan/Apply workflow** — reduces fear of breaking working environment
2. **Export/Import** — enables team sharing, onboarding, backup
3. **Encrypted secrets** — trust signal for security-conscious teams
4. **Multi-stack isolation** — solves real pain of juggling multiple projects
5. **Visual service graph** — understanding beats memorization
6. **One-command setup** — zero friction for new users
7. **Service catalog/recipes** — quick-start beats custom config

## Complexity-to-Value Assessment

**High value, low complexity (do first):**
- Deterministic naming
- Per-stack networks
- Stack CRUD operations
- Basic export/import (without reconciliation)

**High value, medium complexity (core MVP):**
- Plan/Apply workflow
- Encrypted secrets
- Service instance overrides
- Stack compose generator
- Dashboard UI

**High value, high complexity (post-MVP):**
- Auto-wiring with contracts
- Visual service graph
- Database snapshot/restore
- Auto-restart on changes

**Low value, any complexity (defer or skip):**
- Production deployment
- K8s-style orchestration
- Multi-host stacks
- Plugin system

## Sources

**Confidence Level:** MEDIUM (unable to verify with current tools)

Research based on:
- Training knowledge of Docker Compose (as of Jan 2025)
- Training knowledge of Tilt (as of Jan 2025)
- Training knowledge of DDEV, Lando, Docksal (PHP-focused dev tools)
- Training knowledge of Laragon (Windows local dev environment)
- DevArch project context from PROJECT.md and CLAUDE.md

**Verification needed:**
- Current feature sets of competitors (WebSearch/WebFetch unavailable)
- Recent additions to Docker Compose (profiles evolution, secrets handling)
- Tilt's latest dev workflow features
- DDEV/Lando multi-project capabilities

**Confidence by category:**
- Table stakes features: HIGH (well-established patterns)
- Differentiators: MEDIUM (competitive landscape may have shifted)
- Anti-features: MEDIUM (based on community patterns and known pitfalls)
- Competitor analysis: LOW-MEDIUM (unable to verify current versions)

---
*Feature research for: DevArch stacks & instances milestone*
*Researched: 2026-02-03*
