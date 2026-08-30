# Copy-only native commands

Created: 2026-08-29
Status: Planned
Depends on: command palette action registry; runtime relationships

## Outcome

Apps, services, and containers expose a small, consistent set of safely rendered native commands that users can copy and run in their own terminal. The dashboard never executes them.

## Scope

- App workspace `cd` command.
- Service Compose status and bounded logs commands, plus the existing copied start command.
- Container bounded logs commands. Do not add inspect output shortcuts in this program; they drift toward administration and may reveal environment secrets when run.
- Shared command descriptor/rendering contract and action-registry integration.
- Clear read-only versus state-changing copy labels.

## Subphases

1. [Command catalog and shell-safety contract](01-command-contract.md)
2. [Entity actions and responsive UI](02-actions-and-ui.md)
3. [Security, portability, and verification](03-verification-rollout.md)

## Definition of done

- Every command is generated from argv/path components using tested POSIX shell quoting.
- Commands are copied only after an explicit action.
- The dashboard has no command execution endpoint.
- New commands are inspection/navigation commands; no new lifecycle-management catalog is introduced.

## Non-goals

Embedded terminal, command execution, custom scripts, command history, Windows PowerShell generation, start/stop/restart/delete controls, or arbitrary command templates.