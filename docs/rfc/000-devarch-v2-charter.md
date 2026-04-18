# RFC 000 — DevArch V2 Rewrite Charter

- Status: Accepted
- Date: 2026-04-17
- Owners: DevArch V2 rewrite team

## Summary

DevArch V2 is a manifest-first local workspace control plane.

The rewrite replaces the current DB-first, CRUD-heavy V1 architecture with a smaller system built around:

1. catalog templates stored on disk
2. workspace manifests as canonical desired state
3. runtime, plan, and history data derived from manifests and runtime inspection

## Why this rewrite exists

V1 spreads desired state across database tables, generated compose output, runtime materialization, and broad CRUD APIs. That has created:

- too many first-class concepts
- too many pages and handlers for low-frequency admin tasks
- high change cost for ordinary product work
- too much code dedicated to representation instead of user value

V2 exists to collapse the system into a smaller, easier-to-reason-about control plane.

## Product thesis

Users should feel one primary flow:

`define -> plan -> apply -> observe`

The top-level unit is the workspace, not a collection of independently managed CRUD entities.

## In scope for V2

- workspace manifest format
- catalog template format
- deterministic load / resolve / contract-link pipeline
- plan / apply engine
- Docker and Podman runtime abstractions
- thin Go CLI and thin local API backed by the same engine
- new workspace-first web app under `web/`
- V1 importers, fixtures, and parity tooling

## Explicit scope cuts for the core milestone

The following stay out of the core V2 milestone unless later ADRs explicitly restore them:

- AI as a required platform subsystem
- top-level registry, image, and network admin surfaces
- CVE scanning as a core acceptance gate
- relational desired-state storage as the source of truth
- V1-style service/category/instance CRUD sprawl
- team auth / heavy multi-user control-plane concerns as required local defaults

## Source-of-truth rules

- Workspace manifests are canonical desired state.
- Catalog templates are canonical reusable defaults.
- Runtime snapshots, plans, apply history, and caches are derived or ephemeral.
- The local API, CLI, and UI are transport layers over one engine.

## Repository execution rules

- New V2 code lands under root `cmd/`, `internal/`, `schemas/`, `catalog/`, `examples/`, and `web/` boundaries.
- `api/`, `dashboard/`, and `services-library/` are reference and migration sources, not the place to keep building V2 core logic.
- The rewrite is executed in small packets using repo-local pi agents, prompts, skills, and extensions.

## Success criteria

The rewrite is on track only when all of the following are true:

- project-local pi workflows are runnable from committed repo assets
- the root V2 module builds independently of V1 modules
- schemas validate representative examples
- effective graph resolution is deterministic and golden-tested
- the same engine drives CLI, API, and UI flows
- representative V1 fixtures import into V2 manifests or fail clearly with diagnostics

## Deferred-work policy

Anything deferred from the milestone must be recorded explicitly in ADRs, RFCs, or migration docs. Hidden TODO sprawl is not an acceptable substitute for boundary decisions.
