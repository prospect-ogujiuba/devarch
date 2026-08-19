# Phase 04: Consolidated status dashboard

Created: 2026-08-19
Purpose: Make the health and reachability of DevArch services visible without inspecting containers one at a time.

## Goal

Add `scripts/devarch/status.sh` that combines the static catalog with bounded runtime inspection.

## Scope

- Report canonical service ID, container name, lifecycle state, health, published endpoints, image, and uptime when available.
- Support all services, category filtering, named services, `--running`, `--unhealthy`, `--json`, and `--watch INTERVAL`.
- Distinguish not-created, stopped, starting, healthy, unhealthy, and unknown.
- Degrade cleanly when runtime or health data is unavailable.

## Outputs

- Status script, parsers for supported runtime JSON/text responses, tests, and documentation.

## Acceptance criteria

- One bounded runtime inventory is used per refresh rather than one process per service.
- Static services absent from the runtime remain visible unless `--running` is selected.
- Output ordering is deterministic and narrow terminals remain readable.
- Watch mode handles interruption and never accumulates temporary files or child processes.
- JSON schema is versioned and suitable for `open`, `support-bundle`, and completions.

## Verification

- Parser fixtures cover Podman and Docker, duplicate/unrelated containers, missing health, and malformed runtime output.
- Test filters, ordering, terminal-width behavior, JSON validity, and Ctrl-C cleanup.
- Compare a manual status row with the active runtime when available.

## Non-goals

No historical monitoring, metrics storage, alerts, or service mutations.
