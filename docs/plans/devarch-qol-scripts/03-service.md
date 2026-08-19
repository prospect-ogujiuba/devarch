# Phase 03: Single-service lifecycle helper

Created: 2026-08-19
Purpose: Remove repeated directory navigation and runtime-specific Compose syntax while preserving direct Compose ownership.

## Goal

Add `scripts/devarch/service.sh` for discoverable, predictable operations on one Compose definition.

## Scope

- Commands: `list`, `path`, `config`, `up`, `down`, `restart`, `stop`, `pull`, `ps`, `logs`, and `exec`.
- Resolve canonical or unambiguous short service IDs through the shared catalog.
- Ensure `microservices-net` only for mutating startup when needed.
- Forward arguments after `--` safely; support `--dry-run` for mutations and `--json` for list/path/ps where practical.

## Outputs

- Lifecycle script, command tests using a recording fake runtime, and usage documentation.

## Acceptance criteria

- Every runtime invocation uses the selected Compose file explicitly.
- Unknown or ambiguous names fail before runtime execution.
- `list` works without a container runtime.
- `exec` requires an explicit service command after `--` and never uses `eval`.
- Dry-run prints shell-escaped intent and causes zero mutations.
- Runtime exit status is preserved.

## Verification

- Tests assert exact Podman, podman-compose, and Docker argument arrays.
- Test paths containing spaces and forwarded arguments containing metacharacters.
- Syntax, ShellCheck, and existing bootstrap regression suites pass.

## Non-goals

No automatic dependency graph, stack orchestration, browser opening, or persistent service registry.
