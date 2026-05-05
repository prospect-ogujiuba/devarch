# Podman-First Lightweight DevArch Redesign

> **For Hermes:** Use subagent-driven-development or pi-coding-agent to implement this plan packet-by-packet. Do not expand the legacy `api/` monolith; root V2 is the product boundary.

**Goal:** Redesign DevArch so it is no longer a monolithic Go application that does everything, but a lightweight Podman control layer with one shared API/CLI surface for workflows currently handled by shell scripts.

**Architecture:** DevArch V2 remains manifest-first: workspace manifests and catalog templates are canonical desired state; runtime snapshots are derived from Podman. The CLI (`cmd/devarch`) and local daemon/API (`cmd/devarchd`, `internal/api`) must call a shared application service layer (`internal/appsvc`) instead of duplicating behavior. Script workflows move into small Go workflow packages and the legacy V1 `api/` tree is frozen as migration/reference material.

**Tech Stack:** Go, Podman CLI/API via testable command runners, chi-based local API, root V2 packages under `cmd/`, `internal/`, `schemas/`, `catalog/`, `examples/`, and `web/`.

---

## File Map

- [00 Overview](podman-first-lightweight-redesign/00-overview.md) — findings, package shape, boundaries, verification checklist.
- [01 Architecture Decision](podman-first-lightweight-redesign/01-architecture-decision.md) — ADR for Podman-first lightweight wrapper.
- [02 Podman Command Layer](podman-first-lightweight-redesign/02-podman-command-layer.md) — `internal/podmanctl` runner, network helpers, container command builder.
- [03 Podman Runtime Adapter](podman-first-lightweight-redesign/03-podman-runtime-adapter.md) — implement Podman apply/network mutations and smoke fixture.
- [04 Workflow Services](podman-first-lightweight-redesign/04-workflow-services.md) — move script workflows into `internal/workflows` and `internal/appsvc`.
- [05 CLI and API Surfaces](podman-first-lightweight-redesign/05-cli-api-surfaces.md) — expose workflows through thin CLI/API transports.
- [06 Script Retirement](podman-first-lightweight-redesign/06-script-retirement.md) — turn scripts into compatibility shims after parity.
- [07 Pi Delegation Packets](podman-first-lightweight-redesign/07-pi-delegation-packets.md) — ready-to-run `pi` prompts for implementation packets.

## Suggested Execution Order

1. `01-architecture-decision.md`
2. `02-podman-command-layer.md`
3. `03-podman-runtime-adapter.md`
4. `04-workflow-services.md`
5. `05-cli-api-surfaces.md`
6. `06-script-retirement.md`

## Immediate Next Step

Start with **Packet A** from `07-pi-delegation-packets.md`. It is low-risk, establishes the architectural decision, and creates the reusable Podman command layer needed before mutating the runtime adapter.
