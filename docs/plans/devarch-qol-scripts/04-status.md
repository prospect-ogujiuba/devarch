# Phase 04: Consolidated status dashboard

Created: 2026-08-19
Purpose: Make the health and reachability of DevArch services visible without inspecting containers one at a time.

## Goal

Prefer documented native `podman ps`/`podman compose ps` views; add a script only if DevArch label selection cannot be made ergonomic with a shell alias or documented format.

## Scope

- First add consistent Compose labels/project names if the provider does not already emit sufficient labels.
- Document native views using `podman ps --all --filter label=... --format ...`, `podman ps --watch`, `podman inspect`, `podman port`, and per-service `podman compose -f FILE ps`.
- If retained, `status.sh` may only choose a documented native format/filter and invoke one Podman command; it must not join a static catalog to live state or implement a dashboard schema.
- Use Podman's own states and health fields verbatim. Do not redefine not-created/stopped/starting/healthy semantics.

## Native delegation

Podman owns inventory, filtering, health, ports, formatting, and watch refresh. DevArch may contribute stable labels and a readable default Go template. JSON callers use native JSON output directly.

## Outputs

- Native Podman command recipes and format templates; an optional one-command selector only if labels require it; recording tests and documentation.

## Acceptance criteria

- The final implementation is either documentation/alias guidance or one direct `exec podman ps ...` command.
- `podman ps --watch` owns refresh and interruption; DevArch has no refresh loop.
- Native JSON/Go-template fields are preserved; DevArch publishes no status schema.
- Static services that were never created are obtained with `service.sh list`, not synthesized into runtime status.

## Verification

- Recording-Podman tests assert exact `ps`/`compose ps` argv, filters, and passthrough.
- Verify DevArch labels emitted by representative Compose services are filterable.
- Compare the documented format with direct native output; no parser tests should exist.

## Non-goals

No custom dashboard, status parser, terminal-width renderer, watch loop, JSON schema, historical monitoring, metrics storage, alerts, or service mutations.
