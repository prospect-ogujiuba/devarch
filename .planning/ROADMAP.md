# DevArch Product Roadmap

## Editions

### Community Edition (FOSS — Apache 2.0)

Everything that exists today, polished and documented. Free forever.

**Core Platform**
- 166 pre-configured containerized services across 22 categories
- Unified CLI (`service-manager.sh`) — start, stop, rebuild, logs, bulk ops, health checks
- Docker and Podman support with runtime switching
- Bridge networking with service discovery

**Go API (single-user)**
- Service lifecycle management (CRUD, start/stop/restart/rebuild)
- Category and project management
- Real-time WebSocket status streaming
- Container metrics (CPU, memory, network I/O)
- Service config versioning
- Registry integration and vulnerability scanning (Trivy)
- IP-based rate limiting
- Single API key auth

**React Dashboard (single-user)**
- Overview dashboard with real-time status
- Service management — table/grid, filters, sorting, bulk actions
- Service detail — metrics, logs, env vars, compose, health
- Category management — filter, bulk start/stop
- Project scanning — framework/language detection
- Settings — runtime, socket, API key

**Documentation & Community**
- Full setup guide and API reference
- Service catalog documentation
- Contributing guide (easy path: add new services)
- GitHub issue templates

---

### Team Edition (Commercial — $15-25/seat/month)

For development teams sharing infrastructure. Everything in Community, plus:

**Authentication**
- Email/password login with JWT sessions
- Password reset and account locking
- Multi-factor authentication (TOTP)
- Personal access tokens with expiry and scoping

**Team Management**
- Create and manage teams
- Invite members via email or link
- Assign roles per team

**Role-Based Access Control**
- Built-in roles: Admin, Operator, Developer, Viewer
- Per-resource permissions (services, categories, projects)
- Permission-gated dashboard UI
- API-level enforcement on all endpoints

**Audit & Visibility**
- Full audit log — who did what, when, on which resource
- Config change tracking with before/after diffs
- Audit log viewer and export in dashboard

**Notifications**
- Email alerts for service failures
- Configurable notification preferences per user

---

### Enterprise Edition (Commercial — $40-60/seat/month or org pricing)

For organizations with compliance needs or multiple teams. Everything in Team, plus:

**Single Sign-On**
- OIDC / OAuth2 provider integration
- SAML support
- LDAP / Active Directory
- Works with Okta, Azure AD, Keycloak, Google Workspace

**Multi-Tenancy**
- Isolated workspaces per team or department
- Row-level data separation
- Per-workspace service catalogs

**Governance**
- Resource quotas per team/user (service count, CPU, memory)
- Approval workflows for protected environments
- Scheduled deployment windows

**Operations**
- Full environment backup and restore
- Automated backup scheduling
- One-click rollback to previous service configs

**Advanced Alerting**
- Threshold-based alert rules (CPU, memory, health)
- Multiple channels: Slack, webhook, PagerDuty, email
- Escalation policies and on-call routing

**Compliance**
- Change history reports (CSV/PDF export)
- Access log reports
- SOC 2 evidence generation

**Support**
- Priority email/Slack support
- SLA-backed response times
- Dedicated onboarding

---

### Custom / On-Prem (Contact sales)

For large organizations with specific requirements. Everything in Enterprise, plus:

- Air-gapped installation with bundled container images
- Custom service templates and org-specific defaults
- Named account engineer with quarterly business reviews
- Custom feature development on contract
- Team training and onboarding sessions
- Optional managed hosting (DevArch Cloud)

---

## Implementation Phases

### Phase 1: Community Launch

Polish and ship the FOSS edition.

```
Week 1-2: Foundation
├── Add Apache 2.0 LICENSE
├── Write README.md (hero, quick start, screenshots, architecture)
├── Write CONTRIBUTING.md (PR process, adding services)
├── Clean .env.example (remove secrets, add defaults)
└── Create GitHub issue/PR templates

Week 3-4: CI & Release
├── GitHub Actions: lint dashboard, build API, validate compose
├── Automated release workflow
├── Tag v1.0.0
└── Launch: HN, Reddit r/selfhosted, r/docker, Twitter/X
```

### Phase 2: Auth & RBAC Foundation

Build the enterprise repo and core auth.

```
Week 5-8: Authentication
├── Enterprise repo setup (private, Go module)
├── DB migrations: users, sessions, api_tokens
├── Auth handlers: register, login, logout, refresh, reset
├── JWT middleware with session management
├── Password hashing (bcrypt/argon2)
├── Dashboard: login page, auth context, protected routes
└── API token management (create, revoke, list)

Week 9-12: RBAC
├── DB migrations: roles, permissions, role_permissions, user_roles
├── RBAC middleware for all API routes
├── Built-in roles: admin, operator, developer, viewer
├── Dashboard: role assignment UI, permission-gated components
├── Audit log table and recording middleware
└── Audit log viewer page
```

### Phase 3: Team Edition Launch

Ship the first commercial product.

```
Week 13-16: Teams & Billing
├── DB migrations: teams, team_members, invitations
├── Team CRUD API and dashboard pages
├── Invite flow (email + link)
├── Per-user rate limiting
├── Stripe integration (subscriptions, seat management)
├── License key generation and validation
└── Billing dashboard page

Week 17-20: Notifications & Polish
├── SMTP email integration
├── Service failure notifications
├── User notification preferences
├── Marketing site (landing page, pricing, docs)
├── Tag v1.0.0-team
└── Launch Team Edition
```

### Phase 4: Enterprise Edition

Build governance and compliance features.

```
Week 21-24: SSO & Multi-Tenancy
├── OIDC/OAuth2 provider integration
├── SAML support
├── Workspace model with data isolation
├── Per-workspace service catalogs
└── Resource quotas and enforcement

Week 25-28: Operations & Compliance
├── Backup/restore system
├── Automated backup scheduling
├── Approval workflows for protected services
├── Compliance report generation
├── Advanced alerting (rules, channels, escalation)
└── Webhook event system for all actions

Week 29-32: Launch
├── Enterprise documentation
├── Security audit
├── Tag v1.0.0-enterprise
└── Launch Enterprise Edition
```

### Phase 5: Growth (Ongoing)

```
├── DevArch Cloud (managed hosting)
├── Kubernetes operator
├── Plugin/extension system
├── Service template marketplace
├── GraphQL API
├── Mobile companion app
└── IDE extensions (VS Code, JetBrains)
```

---

## Go-to-Market

### Community Adoption Strategy

1. **GitHub presence** — clean repo, good README, social preview image
2. **Service catalog as growth engine** — each new service is an easy PR, grows the contributor base
3. **Content marketing** — "How to run X locally with DevArch" blog posts / tutorials
4. **Community channels** — Discord or GitHub Discussions
5. **Target communities** — r/selfhosted, r/docker, r/devops, Hacker News, dev.to
6. **Integrations** — VS Code extension for service management, GitHub Action for CI environments

### Enterprise Sales Strategy

1. **Product-led growth** — Community users bring DevArch into their companies
2. **Team trial** — 14-day free trial of Team Edition
3. **Self-serve purchase** — credit card checkout for Team tier
4. **Sales-assisted** — Enterprise tier requires a call / demo
5. **Case studies** — document early adopters, publish success stories
6. **Partner channel** — DevOps consultancies reselling/implementing

### Pricing Research

Study these competitors for pricing anchors:
- **Portainer Business**: $5/node/month
- **Docker Business**: $24/user/month
- **GitLab Premium**: $29/user/month
- **Rancher Prime**: contact sales
- **Coolify**: $5/month self-hosted

DevArch provides more breadth than Portainer (166 services vs container management) but is less mature in enterprise features than GitLab. Price accordingly — start lower to drive adoption, raise as features mature.
