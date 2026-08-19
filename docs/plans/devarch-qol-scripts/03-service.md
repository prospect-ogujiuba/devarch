# Phase 03: Single-service lifecycle helper

Created: 2026-08-19
Purpose: Remove repeated directory navigation and runtime-specific Compose syntax while preserving direct Compose ownership.

## Goal

Add a single thin Compose-file resolver that hands control to `podman compose`; do not create a parallel service-management CLI.

## Scope

- DevArch-owned modes are only `list`, `path`, and `run SERVICE -- COMPOSE_ARGS...` (a shorter `SERVICE -- ...` form may be offered after usability testing).
- Resolve canonical or unambiguous short service IDs through the shared catalog.
- Execute `podman compose -f <resolved-compose.yml> ...` with all native Compose arguments unchanged.
- Network creation remains an explicit native prerequisite (`podman network create microservices-net`) or bootstrap responsibility; this helper does not create it implicitly.
- Native examples cover `config`, `up`, `down`, `restart`, `stop`, `pull`, `ps`, `logs`, `exec`, profiles, and Podman's/provider's own `--dry-run`.

## Native delegation

There are no DevArch implementations of lifecycle subcommands. For example, `service.sh database/postgres -- up -d` becomes exactly `podman compose -f services-library/database/postgres/compose.yml up -d`. Unknown native flags and future provider commands pass through automatically.

## Outputs

- Thin Compose-file resolver/passthrough, recording-Podman tests, and direct native-command documentation.

## Acceptance criteria

- Every runtime invocation uses the selected Compose file explicitly.
- Unknown or ambiguous names fail before runtime execution.
- `list` works without a container runtime.
- The wrapper requires `--` before native arguments and never uses `eval`.
- Native help, prompts, dry-run behavior, output, TTY, signals, and exit status are preserved.
- The wrapper contains no case statement enumerating Compose lifecycle commands.

## Verification

- Tests assert the exact `podman compose -f ...` argument array and unchanged provider arguments.
- Test paths containing spaces and forwarded arguments containing metacharacters.
- Syntax, ShellCheck, and existing bootstrap regression suites pass.

## Non-goals

No lifecycle implementation, automatic network mutation, status/log formatting, dependency graph, stack orchestration, browser opening, or persistent service registry.
