# DevArch V2 Plan + Spec

Status: Draft v0.1  
Date: 2026-04-17  
Audience: product, architecture, implementation, and pi agent operators

---

## 1. Executive Summary

DevArch V2 is a **manifest-first local workspace control plane**.

It replaces the current DB-first, CRUD-heavy, multi-surface architecture with a much smaller system built around three core things:

1. **Catalog templates** — reusable service blueprints stored on disk
2. **Workspace manifests** — the canonical desired state for a local environment
3. **Runtime state** — computed status, plans, logs, and apply history

V2 keeps the strongest parts of the current product:

- isolated environments
- reusable templates
- plan/apply workflow
- project-aware setup
- live status/logs/terminal
- portable exports

V2 intentionally cuts or postpones the heavyweight admin surfaces that turned the current system into a very broad control plane.

This document contains:

- the **product spec**
- the **technical architecture spec**
- the **migration strategy**
- the **pi-managed multi-agent execution plan**
- the **surgeon-subagent work model**

---

## 2. Rewrite Thesis

> Stop storing the product as decomposed compose fragments in a large relational admin model. Store a workspace manifest, resolve it into an effective graph, then compute plan, apply, and runtime state from that graph.

---

## 3. Product Direction

### 3.1 North Star

DevArch V2 should feel like:

- **one local product**
- **one simple mental model**
- **one canonical config file per environment**
- **one clean workflow: define → plan → apply → observe**

### 3.2 Product Identity

DevArch V2 is:

- a **local-first developer environment orchestrator**
- a **workspace-centric manifest runner**
- a **template-powered environment composer**

DevArch V2 is **not**:

- a generic infrastructure CMDB
- a full Docker admin clone
- a registry browser with orchestration attached
- an AI platform with environment tooling attached

---

## 4. Current-State Problems V2 Must Solve

### 4.1 Architecture problems

- Desired state is split across files, DB tables, generated compose, and runtime materialization.
- Service config and instance overrides are modeled as many child tables and mirrored CRUD handlers.
- The API surface is too broad.
- The UI mirrors backend granularity instead of user workflow.
- The system is expensive to reason about and expensive to change.

### 4.2 Product problems

- Too many first-class concepts.
- Too many pages for low-frequency admin concerns.
- Too much ceremony for common operations.
- Too much code dedicated to representation instead of product value.

### 4.3 Rewrite targets

V2 must deliver:

- radically fewer concepts
- radically fewer routes
- radically fewer editable entities
- drastically smaller blast radius per feature
- clear module boundaries
- manifest portability
- simpler onboarding and safer maintenance

---

## 5. Product Principles

1. **Manifest is truth**  
   Desired state lives on disk in versionable files.

2. **Local-first by default**  
   The system must work great for a solo developer on one machine without requiring a DB-backed control plane.

3. **Compute, don’t CRUD**  
   Runtime views, plans, effective config, and graph relationships should be computed from manifests and catalog templates.

4. **Workspace-first UX**  
   The main thing a user works on is a workspace, not scattered admin entities.

5. **Small core, optional extras**  
   AI, registry browsing, image management, CVE scanning, and advanced team features are optional modules or later phases.

6. **One engine, many surfaces**  
   CLI, API, and UI all call the same core Go engine.

7. **Schema-driven forms where possible**  
   Avoid one bespoke editor per override type when metadata can drive the editor.

8. **pi-native delivery**  
   The rewrite process must be decomposable into small plans and executable by pi-managed specialist agents.

---

## 6. Core Domain Model

V2 intentionally reduces the product to the following primary nouns.

### 6.1 Catalog Template

A reusable service or project blueprint stored on disk.

Examples:

- postgres
- redis
- nginx
- laravel-app
- node-api

A template defines defaults such as:

- image or build source
- ports
- volumes
- env defaults
- health checks
- contracts
- mount points
- metadata/tags

### 6.2 Workspace

The top-level environment unit.

A workspace contains:

- metadata
- runtime preferences
- a list or map of resources
- policies
- imports/templates references
- secret refs

A workspace corresponds to what current DevArch splits into:

- stack
- project backing stack
- exported devarch.yml concepts

### 6.3 Resource

An instantiated unit inside a workspace.

A resource may be backed by:

- a catalog template
- a project adapter
- a raw compose fragment adapter for compatibility

A resource contains only what is needed for local operation:

- source/template
- enabled flag
- overrides
- dependencies/imports
- expose rules
- mounts
- env
- health/develop options

### 6.4 Runtime Snapshot

Computed state of a workspace and its resources.

Examples:

- running/stopped
- health
- mapped ports
- container ids
- restart counts
- last apply timestamp

This is **not** canonical desired state.

### 6.5 Plan

A computed diff between:

- desired state from manifest + templates
- current runtime state

A plan is ephemeral and reproducible.

### 6.6 Secret Reference

A reference to a local secret source, not a first-class DB entity.

Examples:

- env var reference
- local file path
- encrypted local store entry

---

## 7. V2 Concept Mapping

| V1 Concept | V2 Mapping |
|---|---|
| Service | Catalog Template |
| Stack | Workspace |
| Instance | Resource |
| Project | Resource with `source.type=project` |
| Categories | Template tags/grouping only |
| Compose export/import | Native manifest workflow |
| DB desired state | Manifest desired state |
| Effective config tab | Computed resource resolution |
| Wiring | Imports/exports contracts resolved into links |
| Runtime pages (images/networks/registries) | Utilities, not core navigation |

---

## 8. V2 Scope

### 8.1 In scope for V2.0

- workspace manifest format
- catalog template format
- template resolution engine
- effective config computation
- plan/apply engine
- Docker/Podman runtime abstraction
- workspace status/logs/terminal
- minimal local API
- minimal polished UI
- project attachment/scanning
- importers from current catalog and exported configs
- parity harness against current compose output where meaningful

### 8.2 Explicitly postponed

- separate AI service as core architecture
- pgvector and embeddings as platform requirement
- top-level images admin page
- top-level registries admin page
- top-level networks admin page
- CVE scanning in core milestone
- soft-delete/trash system
- strict/team auth as first-class local requirement
- giant category CRUD surface

---

## 9. Recommended Tech Stack

### 9.1 Keep

- **Go** for the engine, runtime control, planning, importers, and local API
- **React + TypeScript** for the dashboard
- **Vite** for the web app
- **TanStack Router/Query** for app structure
- **Tailwind** for UI styling

### 9.2 Change

- Replace PostgreSQL as desired-state storage with **manifest files on disk**
- Use **SQLite** only for optional local cache/snapshots/history
- Replace shell-heavy CLI with a **Go CLI**
- Collapse duplicated API/CLI logic into a single engine package
- Make the UI schema-driven where it makes sense

### 9.3 Final stack

#### Core
- Go 1.24+
- Cobra CLI
- YAML + JSON Schema
- SQLite for runtime cache/history only
- Docker/Podman adapters

#### Web
- React 19
- TypeScript
- Vite
- TanStack Router + Query
- Tailwind 4
- shadcn/ui or thin internal component primitives

#### QA
- Go tests for engine and adapters
- golden tests for resolution and plan output
- Playwright for core UI flows
- fixture-based migration tests from V1 examples

---

## 10. Proposed V2 Repository Layout

```text
cmd/
  devarch/                 # CLI
  devarchd/                # local API daemon

internal/
  spec/                    # manifest + template schema parsing/validation
  catalog/                 # template loading, indexing, discovery
  workspace/               # workspace loading, normalization
  resolve/                 # template + overrides -> effective graph
  contracts/               # imports/exports linking
  plan/                    # desired vs runtime diff
  apply/                   # apply orchestration
  runtime/                 # runtime abstraction + status/logs/exec
  projectscan/             # project detection and adapters
  importv1/                # migration/import from current DevArch
  cache/                   # sqlite snapshots/history
  api/                     # thin HTTP handlers
  events/                  # event streams

schemas/
  workspace.schema.json
  template.schema.json
  plan.schema.json

catalog/
  builtin/                 # committed first-party templates

web/
  src/
    app/
    routes/
    features/
    components/
    lib/
    generated/             # schema-derived types/forms if used

.pi/
  settings.json
  extensions/
    subagent/
    plan-mode/
    worklog/
  agents/
    orchestrator.md
    scout.md
    architect.md
    planner.md
    reviewer.md
    verifier.md
    scribe.md
    surgeon-engine.md
    surgeon-catalog.md
    surgeon-api.md
    surgeon-ui.md
    surgeon-import.md
    surgeon-runtime.md
  prompts/
    v2-scout-plan.md
    v2-implement-slice.md
    v2-review-slice.md
    v2-ship-phase.md
  skills/
    devarch-v2-rules/
      SKILL.md

docs/
  adr/
  rfc/
  migration/
```

---

## 11. Manifest Specification

### 11.1 Canonical rule

Each workspace has one canonical manifest:

- `devarch.workspace.yaml`

Optional split files may be supported later, but the engine always resolves them into one normalized in-memory model.

### 11.2 Example workspace manifest

```yaml
apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: shop-local
  displayName: Shop Local
  description: Local development workspace for shop app

runtime:
  provider: auto
  isolatedNetwork: true
  namingStrategy: workspace-resource

catalog:
  sources:
    - ./catalog/builtin

policies:
  autoWire: true
  secretSource: local

resources:
  postgres:
    template: postgres
    enabled: true
    env:
      POSTGRES_DB: shop
      POSTGRES_USER: shop
      POSTGRES_PASSWORD:
        secretRef: postgres_password
    volumes:
      - source: workspace.postgres-data
        target: /var/lib/postgresql/data
    exports:
      - postgres

  redis:
    template: redis
    enabled: true
    exports:
      - redis

  api:
    template: node-api
    source:
      type: project
      path: ./apps/shop-api
    imports:
      - contract: postgres
        from: postgres
      - contract: redis
        from: redis
    env:
      NODE_ENV: development
    ports:
      - host: 8200
        container: 3000

  web:
    template: vite-web
    source:
      type: project
      path: ./apps/shop-web
    imports:
      - contract: api
        from: api
    ports:
      - host: 8202
        container: 5173
```

### 11.3 Workspace schema requirements

Required top-level sections:

- `apiVersion`
- `kind`
- `metadata.name`
- `resources`

Optional sections:

- `runtime`
- `catalog`
- `policies`
- `secrets`
- `profiles`

### 11.4 Resource schema requirements

Each resource supports a limited, normalized set of fields:

- `template`
- `source`
- `enabled`
- `env`
- `ports`
- `volumes`
- `dependsOn`
- `imports`
- `exports`
- `health`
- `domains`
- `develop`
- `overrides`

### 11.5 Contract resolution

Imports/exports replace ad hoc wiring UI as the core model.

Rules:

- exactly one matching provider = auto-link
- zero providers = unresolved warning
- multiple providers = explicit selection required

Resolved links are computed and shown in the workspace graph.

---

## 12. Template Specification

### 12.1 Canonical rule

Templates live on disk as standalone documents under `catalog/`.

Example:

```text
catalog/builtin/database/postgres/template.yaml
```

### 12.2 Template fields

- metadata (name, tags, description)
- source (image/build)
- default env
- default ports
- default volumes
- health
- contracts provided/required
- develop hints
- compatibility tags

### 12.3 Template example

```yaml
apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: postgres
  tags: [database, sql]
  description: PostgreSQL database template

spec:
  runtime:
    image: postgres:16
  env:
    POSTGRES_DB: app
    POSTGRES_USER: app
  ports:
    - container: 5432
  volumes:
    - target: /var/lib/postgresql/data
      kind: data
  exports:
    - contract: postgres
      env:
        DB_HOST: "${resource.host}"
        DB_PORT: "${resource.port.5432}"
  health:
    test: ["CMD-SHELL", "pg_isready -U ${env.POSTGRES_USER}"]
```

### 12.4 Template philosophy

Templates are:

- plain files
- easy to diff
- easy to version
- easy to migrate
- easy to ship with the repo

Templates are **not** stored in the DB as canonical state.

---

## 13. Engine Architecture

### 13.1 Core engine pipeline

```text
workspace manifest
  + catalog templates
  + runtime snapshot
  -> normalize
  -> resolve
  -> validate
  -> contract-link
  -> effective graph
  -> plan
  -> apply
```

### 13.2 Required engine modules

#### `spec`
- parse/validate YAML
- versioned schema support
- normalized in-memory models

#### `catalog`
- discover template files
- index by name/tags/contracts
- support builtin + workspace-local catalogs

#### `resolve`
- merge template defaults with workspace overrides
- attach project sources
- normalize ports/volumes/env/health

#### `contracts`
- resolve imports/exports
- inject derived connection env
- produce warnings for ambiguities

#### `plan`
- compare desired graph vs runtime snapshot
- compute add/modify/remove/restart
- surface user-readable reasoning

#### `apply`
- render runtime payloads
- materialize config if needed
- create network
- run compose or runtime operations
- emit events

#### `runtime`
- adapter interface for Docker and Podman
- logs
- exec/terminal
- inspect/status
- network lifecycle helpers

#### `projectscan`
- detect app type/framework/package manager
- create suggested resource manifests

#### `importv1`
- convert current service-library and exported configs to V2 model
- run parity comparisons where possible

---

## 14. Local State Strategy

### 14.1 Source of truth

- files on disk

### 14.2 Cached/ephemeral state

Optional SQLite DB for:

- recent runtime snapshots
- apply history
- indexed logs metadata
- migration audit data
- last-opened workspaces

### 14.3 Hard rule

No runtime cache may become the canonical desired-state store.

---

## 15. CLI Specification

### 15.1 Primary commands

```bash
devarch workspace list
devarch workspace open <name>
devarch workspace plan <name>
devarch workspace apply <name>
devarch workspace status <name>
devarch workspace logs <name> [resource]
devarch workspace exec <name> <resource>

devarch catalog list
devarch catalog show <template>

devarch import v1-stack <file>
devarch import v1-library <path>
devarch scan project <path>
```

### 15.2 CLI principles

- one binary
- human-readable output first
- JSON output for automation where useful
- same engine as API/UI

---

## 16. API Specification

### 16.1 Minimal API surface

```text
GET    /api/catalog/templates
GET    /api/catalog/templates/:name

GET    /api/workspaces
GET    /api/workspaces/:name
GET    /api/workspaces/:name/manifest
GET    /api/workspaces/:name/graph
GET    /api/workspaces/:name/status
GET    /api/workspaces/:name/plan
POST   /api/workspaces/:name/apply
GET    /api/workspaces/:name/logs
WS/SSE /api/workspaces/:name/events
WS     /api/workspaces/:name/resources/:resource/exec

POST   /api/import/v1-stack
POST   /api/import/v1-library
POST   /api/scan/project
GET    /api/runtime/status
```

### 16.2 API principles

- thin wrappers around engine calls
- no giant entity CRUD surface
- computed responses over child-table mutation endpoints

---

## 17. UI Specification

### 17.1 Primary navigation

Only four first-class areas:

- **Workspaces**
- **Catalog**
- **Activity**
- **Settings**

### 17.2 Workspace view

Tabs:

- Overview
- Resources
- Graph
- Plan
- Logs
- Raw Config

### 17.3 Resource editing model

Preferred flow:

- high-level structured editor for common fields
- raw manifest editor for advanced users
- no separate bespoke page for every override subtype

### 17.4 UI quality bar

V2 should feel:

- smaller
- faster
- more opinionated
- less admin-panel-like
- more like a polished developer tool

---

## 18. Migration Strategy from V1

### 18.1 Keep

Reuse ideas and assets from V1:

- identity/naming conventions
- service library content
- contract/wiring concepts
- runtime abstraction concepts
- project scanning logic ideas
- plan/apply semantics

### 18.2 Do not preserve line-for-line

Do not port directly:

- current DB schema
- current handler structure
- current service/instance override CRUD shape
- current shell CLI architecture
- current separated AI service requirement

### 18.3 Migration path

#### Stage A — extraction
- export sample stacks from V1
- snapshot representative service templates
- build golden fixtures

#### Stage B — template conversion
- convert service library folders into V2 templates
- preserve tags/contracts/defaults

#### Stage C — workspace importer
- convert exported V1 stack files into V2 workspace manifests
- support compatibility mapping for services/projects/instances

#### Stage D — parity harness
- compare V1 effective compose with V2 rendered effective graph for a set of fixtures
- flag intentional deviations explicitly

#### Stage E — rollout
- run V2 in shadow mode for representative workspaces
- switch primary UX to V2 once parity and usability gates pass

---

## 19. Phased Delivery Plan

## Phase 0 — pi Execution Layer + Rewrite Charter

### Goal
Create the repo-local pi operating system that will manage the rewrite.

### Deliverables
- `.pi/settings.json`
- `.pi/extensions/subagent/`
- `.pi/extensions/plan-mode/`
- `.pi/agents/` role definitions
- `.pi/prompts/` workflow prompts
- `docs/rfc/000-devarch-v2-charter.md`
- `docs/adr/0001-v2-manifest-first.md`

### Acceptance
- pi can run project-local agents
- team can execute scout/architect/surgeon/reviewer/verifier workflows
- plan mode works for read-only planning passes

## Phase 1 — Spec + Schema + Repo Skeleton

### Goal
Lock the V2 model before building the engine.

### Deliverables
- workspace schema draft
- template schema draft
- repo/module skeleton
- fixture set from V1
- ADRs for source of truth, runtime boundary, and API shape

### Acceptance
- schemas validate sample workspace and template docs
- team agrees on scope cuts and deferred features

## Phase 2 — Catalog + Workspace Resolver

### Goal
Build the file-based model loader and effective graph resolver.

### Deliverables
- template discovery/indexing
- workspace loader
- resolution engine
- contracts linker
- normalized effective graph
- golden tests

### Acceptance
- effective graph is deterministic
- imports/exports resolve correctly
- ambiguous links are surfaced clearly

## Phase 3 — Plan + Apply + Runtime Adapters

### Goal
Build the operational core.

### Deliverables
- runtime adapter interface
- Docker adapter
- Podman adapter
- planner diff engine
- apply executor
- event stream model
- logs/exec/status primitives

### Acceptance
- a simple workspace can be planned and applied end-to-end
- status/logs/exec work for both runtimes on supported fixtures

## Phase 4 — CLI + Thin Local API

### Goal
Expose the engine through a minimal operator surface.

### Deliverables
- Go CLI
- local API daemon
- structured JSON output
- initial import commands

### Acceptance
- core flows work from CLI without UI
- API surface remains intentionally small

## Phase 5 — Web UI

### Goal
Ship the polished workspace-first interface.

### Deliverables
- workspace list
- workspace detail tabs
- plan view
- graph view
- logs view
- raw manifest editor
- catalog browser

### Acceptance
- user can perform define → plan → apply → observe from the UI
- navigation is reduced to core surfaces only

## Phase 6 — V1 Importers + Parity Harness

### Goal
Make migration real.

### Deliverables
- V1 catalog importer
- V1 stack importer
- fixture parity runner
- migration docs

### Acceptance
- representative V1 workspaces import into valid V2 manifests
- parity harness passes agreed scenarios

## Phase 7 — Polish + Decomposition of Advanced Modules

### Goal
Stabilize V2 and separate postponed concerns.

### Deliverables
- error UX
- performance tuning
- docs/site update
- advanced module boundaries for AI, registry utilities, scanning, etc.

### Acceptance
- V2 is shippable as primary product direction
- deferred features have explicit module/plugin plans, not hidden TODOs

---

## 20. pi-Orchestrated Delivery Model

## 20.1 Important pi fact

Pi core does **not** ship built-in subagents or plan mode by default.

Therefore this rewrite must commit a **repo-local pi operating layer** using project-local `.pi/` resources.

### Required repo-local pi components

```text
.pi/
  settings.json
  extensions/
    subagent/
    plan-mode/
    worklog/
  agents/
    orchestrator.md
    scout.md
    architect.md
    planner.md
    reviewer.md
    verifier.md
    scribe.md
    surgeon-engine.md
    surgeon-catalog.md
    surgeon-runtime.md
    surgeon-api.md
    surgeon-ui.md
    surgeon-import.md
  prompts/
    v2-scout-plan.md
    v2-implement-slice.md
    v2-review-slice.md
    v2-phase-closeout.md
  skills/
    devarch-v2-rules/
      SKILL.md
```

### Why repo-local `.pi/`

Because pi supports:

- project-local extensions
- project-local skills
- project-local prompts
- project-local settings
- project-local agents when enabled by a subagent extension

This keeps the rewrite process reproducible for the whole team.

---

## 21. Agent Topology

## 21.1 Control roles

### `orchestrator`
Runs the overall phase plan, chooses work packets, schedules agents, enforces gates.

### `scribe`
Keeps RFCs, ADRs, changelogs, migration notes, and completion summaries up to date.

### `scout`
Fast repo/doc reconnaissance. Produces compressed context for downstream agents.

### `architect`
Produces explicit implementation plans and risk assessment.

### `planner`
Translates requirements/findings into smaller execution packets where needed.

### `reviewer`
Reviews diffs for correctness, code quality, and scope discipline.

### `verifier`
Runs tests/build/fixture/parity checks and reports pass/fail with evidence.

## 21.2 Surgeon implementation roles

These are all **surgeon-style agents** with the same minimal-diff philosophy but different domain prompts.

### `surgeon-engine`
Owns `internal/spec`, `resolve`, `contracts`, `plan`.

### `surgeon-catalog`
Owns template format, loaders, and builtin catalog conversion.

### `surgeon-runtime`
Owns runtime adapters, apply pipeline, status/logs/exec.

### `surgeon-api`
Owns thin HTTP surface and event endpoints.

### `surgeon-ui`
Owns web UX, schema-driven editors, and workspace pages.

### `surgeon-import`
Owns V1 importers, migration fixtures, and parity tooling.

### `surgeon-tests`
Optional dedicated surgeon for test harnesses, golden fixtures, Playwright, and CI wiring.

---

## 22. Work Packet Rules

Every rewrite task must be decomposed into small work packets that are safe for isolated execution.

### 22.1 Packet size

A packet should usually be:

- 1 clear objective
- 1 domain only
- 2-3 concrete tasks max
- <= 10 touched files preferred
- independently reviewable
- independently revertible

### 22.2 Packet template

```xml
<work_packet>
  <id>P2-RSLV-003</id>
  <phase>2</phase>
  <domain>resolve</domain>
  <owner>surgeon-engine</owner>
  <goal>Implement template default merge for env/ports/volumes</goal>
  <inputs>
    <file>docs/rfc/001-workspace-spec.md</file>
    <file>schemas/workspace.schema.json</file>
  </inputs>
  <tasks>
    <task>Create normalized merge function for env values</task>
    <task>Add port and volume merge semantics</task>
    <task>Add golden tests for override precedence</task>
  </tasks>
  <validation>
    <check>go test ./internal/resolve/...</check>
  </validation>
  <done>
    <item>workspace overrides win over template defaults</item>
    <item>golden tests cover env, ports, and volumes</item>
  </done>
</work_packet>
```

### 22.3 Ownership rule

One packet = one primary surgeon.

Cross-domain work requires either:

- a prior packet that creates the dependency, or
- a new architect plan that coordinates the sequence

---

## 23. Standard pi Workflows

## 23.1 Planning workflow

**Chain:** `scout -> architect -> scribe`

Use when:

- defining a phase
- designing a new subsystem
- preparing a packet batch

### Output
- RFC/ADR or packet batch spec

## 23.2 Implementation workflow

**Chain:** `scout -> architect -> surgeon-* -> reviewer -> verifier -> scribe`

Use when:

- executing a packet
- implementing code with clear completion criteria

### Output
- code changes
- review findings
- verification evidence
- updated docs/worklog

## 23.3 Parallel implementation workflow

**Parallel surgeons:** multiple `surgeon-*` agents on independent packets  
**Then chain:** `reviewer -> verifier -> scribe`

Use when:

- packets are domain-isolated
- dependencies are already satisfied

### Rule
No parallel execution on the same files or same package boundary without an explicit architect-approved split.

---

## 24. Example Subagent Invocation Patterns

## 24.1 Plan a phase

```text
Use the subagent tool with a chain:
1. scout: inspect current V2 resolver code and schemas
2. architect: produce an executable packet batch for Phase 2
3. scribe: write docs/rfc/phase-2-batch-a.md from {previous}
```

## 24.2 Implement a packet

```text
Use the subagent tool with a chain:
1. scout: gather only the files relevant to packet P2-RSLV-003
2. architect: produce a 2-3 task execution plan for packet P2-RSLV-003 using {previous}
3. surgeon-engine: implement the plan from {previous}
4. reviewer: review the changes from {previous}
5. verifier: run validation for packet P2-RSLV-003 using {previous}
6. scribe: update packet status and summary using {previous}
```

## 24.3 Run parallel surgeons

```text
Use the subagent tool in parallel:
- surgeon-catalog on P2-CAT-002
- surgeon-engine on P2-RSLV-003
- surgeon-import on P6-IMP-001

Then run reviewer, verifier, and scribe after all complete.
```

---

## 25. Required pi Prompts

These prompt templates should be committed under `.pi/prompts/`.

### `v2-scout-plan.md`
Purpose: scout + architect only, no code changes.

### `v2-implement-slice.md`
Purpose: scout + architect + surgeon + reviewer + verifier on one packet.

### `v2-review-slice.md`
Purpose: reviewer + verifier on already-implemented work.

### `v2-phase-closeout.md`
Purpose: summarize phase completion, open risks, and remaining backlog.

---

## 26. Agent Prompt Requirements

Every project-local agent must include:

- explicit domain ownership
- allowed tools
- output format
- rules for scope discipline
- required validation behavior
- handoff format for the next agent

### Special requirement for surgeon agents

Surgeon agents must:

- read before editing
- prefer minimal diffs
- avoid opportunistic refactors
- stop and escalate on architectural surprises
- return structured changed-file summaries

---

## 27. Definition of Done Per Packet

A packet is only complete when all are true:

- architect plan exists
- surgeon implementation completed
- reviewer has no unresolved critical findings
- verifier shows passing required checks
- scribe updated the packet log and relevant RFC/ADR docs
- no undocumented scope drift

---

## 28. Definition of Done Per Phase

A phase is complete when:

- all committed packets in the phase are done
- all required acceptance criteria pass
- docs and ADRs are current
- deferred work is recorded explicitly
- next-phase inputs are prepared

---

## 29. Branching and Merge Strategy

### 29.1 Branch model

Recommended:

- `v2/foundation`
- `v2/phase-1-spec`
- `v2/phase-2-resolve`
- `v2/phase-3-runtime`
- etc.

Packets can use short-lived branches:

- `v2/p2-rslv-003`
- `v2/p3-run-004`

### 29.2 Merge rule

Only merge packets that:

- pass verifier checks
- do not conflict with current domain owner work
- have a scribe summary attached

---

## 30. Testing and Verification Strategy

### 30.1 Core required checks

#### Engine
- unit tests
- golden resolution tests
- planner diff tests
- runtime adapter contract tests

#### API
- handler tests
- integration tests for plan/apply/status

#### UI
- route-level tests
- Playwright smoke flows
- raw manifest editor tests

#### Migration
- fixture import tests
- parity comparison tests

### 30.2 Mandatory gates by phase

| Phase | Gate |
|---|---|
| 1 | schemas validate fixtures |
| 2 | resolver golden tests pass |
| 3 | plan/apply/status works on sample workspaces |
| 4 | CLI and API produce same engine outcomes |
| 5 | UI completes core workspace flow |
| 6 | import/parity suite passes agreed fixtures |
| 7 | release checklist complete |

---

## 31. Risk Register

### High risks

1. **Scope creep from V1 parity pressure**  
   Mitigation: keep explicit V2 cut list and deferred features registry.

2. **Rebuilding too much admin surface too early**  
   Mitigation: lock UI nav to four areas only.

3. **Manifest format over-engineering**  
   Mitigation: start small, validate against real fixtures.

4. **Parallel agent collisions**  
   Mitigation: packet ownership and domain boundaries.

5. **Import complexity from V1 data shape**  
   Mitigation: fixture-first importer design and parity harness.

6. **Runtime abstraction drift across Docker/Podman**  
   Mitigation: adapter contract tests and sample matrix.

---

## 32. Immediate Build Order

This is the recommended order for the first ten packets.

1. Create repo-local `.pi/` execution layer
2. Write rewrite charter and ADRs
3. Draft workspace schema
4. Draft template schema
5. Build fixture corpus from V1 representative exports/templates
6. Create repo skeleton and module boundaries
7. Implement template discovery/index
8. Implement workspace loader + validation
9. Implement resolver merge semantics
10. Add golden tests for resolution and contract linking

Do **not** start with UI polish or migration tooling first.

---

## 33. First Packet Batch

### Batch A — Foundation

#### `P0-PI-001`
Bootstrap project-local pi layer from example subagent + plan-mode extensions.

#### `P1-SPEC-001`
Write V2 charter, goals, scope cuts, and ADRs.

#### `P1-SCHEMA-001`
Define `workspace.schema.json` and 3 valid sample manifests.

#### `P1-SCHEMA-002`
Define `template.schema.json` and 5 sample templates.

#### `P1-FIX-001`
Extract representative V1 fixtures for stacks, templates, and projects.

#### `P1-REPO-001`
Create module skeleton for engine, api, cli, and web.

---

## 34. Success Criteria for the Rewrite

The rewrite is successful if all of the following are true:

- DevArch V2 has **one canonical desired-state model** on disk
- the number of first-class concepts is dramatically reduced
- the API surface is drastically smaller than V1
- core flows work through the same engine in CLI, API, and UI
- the UI is workspace-first and materially simpler
- V1 representative fixtures can be imported or intentionally rejected with clear reasons
- the implementation process is repeatable by pi-managed agents using small work packets

---

## 35. Final Recommendation

Build V2 as a **clean break in architecture** with selective reuse of ideas and assets from V1.

Do not attempt to “clean up” the V1 relational CRUD model.

Instead:

- define the V2 manifest
- build the resolver and planner
- expose a thin local API
- ship a smaller UI
- run the rewrite with a pi-managed agent operating system committed into the repo

This is the path most likely to produce a polished, maintainable, radically simpler product.
