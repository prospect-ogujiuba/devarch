# DevArch — Project Status & Open-Core Strategy

## Current State (January 2026)

DevArch is a microservices development environment managing **166 containerized services** across **22 categories** via Docker/Podman Compose, with a CLI, Go API, and React dashboard.

### What's Built

| Component | Maturity | Summary |
|---|---|---|
| **Compose Infrastructure** | Production | 166 services, one compose file each, consistent structure |
| **CLI (`service-manager.sh`)** | Production | 45KB — start/stop/rebuild, bulk ops, health checks, parallel startup, dry-run |
| **Go API** | Functional | ~7,200 LOC — REST + WebSocket, service/category/project management, registry integration, Trivy scanning |
| **React Dashboard** | Feature-rich | React 19, TanStack Router, Radix UI, Tailwind — 8 pages, real-time status, bulk actions |
| **Documentation** | Strong | 7 planning docs, CLAUDE.md, AGENTS.md, API dev guide |

### What's Missing

- No authentication or user system
- No RBAC / permissions
- No multi-tenancy or team isolation
- No audit logging
- No LICENSE file or README.md at root
- No contribution guidelines

---

## Open-Core Model

### The Split

**DevArch Community (FOSS)** — the engine that gets adoption.
**DevArch Enterprise** — the layer that gets revenue.

The line is drawn at **individual vs. organizational use**. A solo developer or small team gets a fully functional product. The moment you need user management, access control, compliance, or support — you need Enterprise.

---

## DevArch Community Edition (FOSS)

**License:** Apache 2.0 (permissive, enterprise-friendly, patent grant)

### Included — Everything That Exists Today, Plus Polish

**Infrastructure & CLI**
- All 166 compose service definitions
- Full `service-manager.sh` CLI with all current flags
- Runtime detection and switching (Docker/Podman)
- `microservices-net` bridge networking
- Socket management
- Database initialization scripts

**Go API (single-user)**
- All service CRUD and lifecycle endpoints (start/stop/restart/rebuild)
- Category listing and bulk operations
- Project scanning and management
- WebSocket real-time status
- Container metrics (CPU, memory, network)
- Service versioning and config snapshots
- Registry integration (DockerHub, GHCR)
- Vulnerability scanning via Trivy
- Nginx config generation
- Rate limiting (IP-based)
- Single API key authentication

**React Dashboard (single-user)**
- Overview dashboard with status summary
- Services list — table/grid view, filtering, sorting, bulk actions
- Service detail — metrics, logs, env vars, compose YAML, health status
- Categories — filtering, start/stop controls
- Projects — scanning, framework detection, dependency display
- Settings — runtime switching, socket config, API key
- Real-time WebSocket updates

**Documentation**
- Full setup and usage docs
- Service catalog with categories
- API reference
- Contributing guide

### Community Edition Goals
1. Get on GitHub with a clean README, LICENSE, and CONTRIBUTING.md
2. Establish the project as the go-to local microservices dev environment
3. Build a contributor base around the service catalog (new compose files are easy PRs)
4. Drive adoption in dev teams → creates demand for Enterprise

---

## DevArch Enterprise Edition

**License:** Commercial (proprietary, per-seat or per-organization)

### Tier 1: Team ($X/seat/month)

Target: development teams of 5-50 who share infrastructure.

| Feature | Description |
|---|---|
| **Authentication** | Email/password login, JWT sessions, password reset, account locking |
| **Team Management** | Create teams, invite members, assign roles |
| **RBAC** | Roles: admin, operator, developer, viewer. Per-resource permissions (services, categories, projects) |
| **Audit Log** | Who started/stopped/rebuilt what, when. Before/after config snapshots |
| **User-scoped API Tokens** | Personal access tokens with expiry and scope limits |
| **Per-user Rate Limiting** | Rate limits bound to users, not just IPs |
| **Dashboard Auth UI** | Login page, user profile, team settings, role management |
| **Email Notifications** | Service down alerts, build failures, security advisories |

### Tier 2: Enterprise ($X/org/month)

Target: organizations with multiple teams, compliance requirements, or production-adjacent use.

| Feature | Description |
|---|---|
| **SSO / OIDC** | SAML, OAuth2, LDAP integration (Okta, Azure AD, Keycloak, etc.) |
| **Multi-tenancy** | Isolated workspaces per team/org, row-level data separation |
| **Resource Quotas** | Max services, CPU, memory limits per team/user |
| **Approval Workflows** | Change requests for protected environments, scheduled deployments |
| **Backup & Restore** | Full environment export/import, automated backup scheduling |
| **Advanced Alerting** | Threshold-based rules, Slack/webhook/PagerDuty channels, escalation |
| **Compliance Reporting** | Change history exports, access log reports, SOC 2 evidence |
| **Priority Support** | Dedicated Slack/email, SLA-backed response times |
| **Custom Integrations** | Webhook events for all actions, custom service templates |
| **Rollback** | One-click rollback to previous service config versions |

### Tier 3: Custom / On-Prem

Target: large orgs that need air-gapped or heavily customized deployments.

| Feature | Description |
|---|---|
| **Air-gapped Install** | Offline installation with bundled images |
| **Custom Service Catalog** | Private service templates, org-specific defaults |
| **Dedicated Support** | Named account engineer, quarterly reviews |
| **Custom Development** | Feature development on contract |
| **Training** | Team onboarding sessions |

---

## Implementation Roadmap

### Phase 1: FOSS Launch (Weeks 1-4)

Ship what exists, polished.

- [ ] Add Apache 2.0 LICENSE
- [ ] Write root README.md (hero description, quick start, screenshots, architecture diagram)
- [ ] Write CONTRIBUTING.md (PR process, service catalog contributions, dev setup)
- [ ] Clean up .env.example — remove secrets, add sane defaults
- [ ] Add GitHub Actions CI (lint dashboard, build Go API, validate compose files)
- [ ] Create GitHub issue templates (bug, feature, new service request)
- [ ] Tag v1.0.0-community
- [ ] Publish to relevant channels (Hacker News, Reddit r/selfhosted, Docker community)

### Phase 2: Enterprise Foundation (Weeks 5-12)

Build the auth and team layer.

- [ ] Database migrations: users, roles, permissions, teams, audit_logs, api_tokens tables
- [ ] Auth system: JWT-based login, password hashing (bcrypt), session management
- [ ] RBAC middleware: permission checks on all API handlers
- [ ] Audit logging middleware: automatic action recording
- [ ] User management API endpoints
- [ ] Dashboard login flow and auth context
- [ ] Dashboard permission-gated UI (hide controls based on role)
- [ ] User management page (admin only)
- [ ] API token management page
- [ ] Audit log viewer page

### Phase 3: Team Features (Weeks 13-20)

Ship Tier 1.

- [ ] Team/org model: creation, invites, member management
- [ ] Role assignment UI
- [ ] Per-user rate limiting
- [ ] Email notification system (SMTP integration)
- [ ] Service-down alerting
- [ ] Stripe integration for billing
- [ ] License key validation system
- [ ] Marketing site and docs site
- [ ] Tag v1.0.0-enterprise

### Phase 4: Enterprise Features (Weeks 21-32)

Ship Tier 2.

- [ ] SSO/OIDC integration (generic OpenID Connect provider)
- [ ] Multi-tenancy with workspace isolation
- [ ] Resource quotas and enforcement
- [ ] Approval workflows for protected services
- [ ] Backup/restore system
- [ ] Advanced alerting with multiple channels
- [ ] Compliance report generation
- [ ] Webhook event system

### Phase 5: Scale (Ongoing)

- [ ] Air-gapped installation packaging
- [ ] Kubernetes operator (alternative to Compose for production)
- [ ] Plugin system for custom service templates
- [ ] GraphQL API (optional, alongside REST)
- [ ] Marketplace for community service templates

---

## Revenue Model

### Pricing Anchors (research competitors: Portainer, Docker Desktop Business, Rancher)

| Edition | Price | Target |
|---|---|---|
| **Community** | Free, forever | Individual devs, small teams, OSS projects |
| **Team** | ~$15-25/seat/month | Dev teams 5-50, startups |
| **Enterprise** | ~$40-60/seat/month (or flat org pricing) | Orgs 50+, compliance-heavy |
| **Custom** | Contact sales | Air-gapped, on-prem, custom dev |

### Revenue Streams
1. **Seat licenses** — recurring, predictable
2. **Support contracts** — high margin, builds relationships
3. **Custom development** — one-time, funds roadmap
4. **Training** — low effort after initial material creation
5. **Managed hosting** (future) — DevArch Cloud, fully hosted instance

### Key Metrics to Track
- GitHub stars and forks (adoption signal)
- Community downloads / Docker pulls
- Conversion rate: Community → Team trial
- Monthly recurring revenue (MRR)
- Churn rate per tier
- Support ticket volume and resolution time

---

## Competitive Positioning

| Competitor | What They Do | DevArch Differentiator |
|---|---|---|
| **Portainer** | Container management UI | DevArch is dev-environment-first with 166 pre-configured services, not just a container viewer |
| **Docker Desktop** | Local Docker runtime + UI | DevArch manages the full service catalog, not just containers |
| **Rancher** | Kubernetes management | DevArch targets Compose-based dev environments, simpler on-ramp |
| **Tilt / Skaffold** | Dev workflow for K8s | DevArch is Compose-native with a broader service catalog |
| **Coolify / CapRover** | Self-hosted PaaS | DevArch focuses on development environments, not production hosting |

**DevArch's moat:** The curated, pre-configured catalog of 166 services with a unified management layer. No one else ships a development environment with this breadth out of the box.

---

## Repo Structure for Open-Core

```
devarch/                          ← FOSS (Apache 2.0, public GitHub)
├── compose/                      ← All service definitions
├── config/                       ← All service configs
├── scripts/                      ← CLI tools
├── api/                          ← Go API (community features)
│   ├── internal/enterprise/      ← Build-tag gated (ee/ or separate repo)
├── dashboard/                    ← React dashboard (community features)
├── LICENSE                       ← Apache 2.0
├── README.md
└── CONTRIBUTING.md

devarch-enterprise/               ← Commercial (private repo or build-tag gated)
├── auth/                         ← Authentication system
├── rbac/                         ← Role-based access control
├── teams/                        ← Multi-tenancy
├── audit/                        ← Audit logging
├── billing/                      ← Stripe integration
├── alerts/                       ← Advanced alerting
└── LICENSE                       ← Commercial license
```

Two common approaches:
1. **Separate repo** — enterprise code in a private repo, imported as a Go module. Cleaner separation.
2. **Build tags** — enterprise code in the same repo behind `//go:build enterprise` tags. Simpler development, GitLab/Sourcegraph model.

Recommendation: **Separate repo**. Keeps the FOSS repo clean, avoids accidental leaks of commercial code, and makes licensing unambiguous.

---

## Immediate Next Steps

1. Add LICENSE (Apache 2.0) and root README.md
2. Set up GitHub repo with proper description, topics, and social preview
3. Write CONTRIBUTING.md focused on service catalog contributions (lowest friction)
4. Set up CI (GitHub Actions)
5. Start building auth system in a private enterprise repo
6. Research pricing by surveying target users
