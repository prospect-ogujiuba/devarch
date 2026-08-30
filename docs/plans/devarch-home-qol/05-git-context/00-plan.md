# Project Git context

Created: 2026-08-29
Status: Planned
Depends on: current app inventory API

## Outcome

App cards and detail pages show concise repository context—branch, clean/modified state, and last commit age—updated only on initial inventory load or explicit refresh.

## Scope

- Detect whether each `apps/*` workspace is a Git worktree.
- Return allowlisted branch, tracked/untracked change counts, detached state, short commit ID, and last commit timestamp.
- Display compact status on cards and fuller context on app details.
- Bound per-repository and total discovery cost.

## Subphases

1. [Git inventory and safety contract](01-inventory-contract.md)
2. [Backend discovery and UI](02-discovery-and-ui.md)
3. [Performance, failure, and verification](03-verification-rollout.md)

## Definition of done

- Non-repositories and Git failures do not affect other inventory.
- No commit messages, remotes, author data, config, diffs, file names, or credentials are exposed.
- Manual refresh is the only update mechanism.
- Large/broken repositories cannot stall the dashboard indefinitely.

## Non-goals

Git operations, staging, commits, diffs, branch switching, remote status, fetch/pull/push, file lists, or repository initialization.