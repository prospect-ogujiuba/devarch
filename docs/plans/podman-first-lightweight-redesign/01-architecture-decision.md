# 01 — Record the Podman-First Architecture Decision

## Goal

Make the new direction explicit so future changes do not drift back into monolith/CRUD/script sprawl.

### Task 1: Add ADR for Podman-first lightweight wrapper

**Objective:** Create an accepted architecture decision for the Podman-first lightweight wrapper direction.

**Files:**
- Create: `docs/adr/0004-podman-first-lightweight-wrapper.md`

**Content outline:**

```markdown
# ADR 0004 — DevArch V2 is a Podman-first lightweight wrapper

- Status: Accepted
- Date: 2026-05-04

## Context

DevArch V1 grew into a DB-backed Go API with many CRUD handlers and separate scripts for operator workflows. The desired V2 product is a small local control plane over Podman, not another monolith.

## Decision

DevArch V2 is Podman-first. Workspace manifests and catalog templates are desired state. The root engine resolves manifests into runtime payloads, plans changes, and applies them through Podman. CLI and local API are transport layers over `internal/appsvc`. Script workflows move into small Go workflow services and scripts become compatibility shims.

## Consequences

- Root V2 packages are the product path.
- `api/` is frozen as V1 reference/import material.
- Docker remains optional compatibility only.
- Adding API endpoints requires an appsvc/workflow method first.
- Adding CLI commands requires the same shared method used by the API.
```

**Validation:**
- Read the ADR and confirm it does not contradict:
  - `docs/rfc/000-devarch-v2-charter.md`
  - `docs/adr/0003-v2-thin-local-api.md`
