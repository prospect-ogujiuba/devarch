# Phase 09: Conservative cleanup

Created: 2026-08-19
Purpose: Reclaim space and remove stale recovery artifacts without turning convenience into accidental data loss.

## Goal

Prefer direct, documented Podman disk/prune commands; retain a script only for DevArch-owned filesystem artifacts that Podman cannot manage.

## Scope

- Document and expose native `podman system df`, `podman system check`, `podman container prune`, `podman image prune`, `podman volume prune`, and `podman system prune` with their own filters and confirmation behavior.
- Do not precompute Podman's prune candidates, sizes, reachability, or ownership. Podman decides what is unused.
- Keep DevArch filesystem cleanup separate: list old `.devarch-backups`, `.devarch-failed`, resolved recovery guards, and incomplete staging directories using simple path/age rules.
- A retained helper has two explicit modes: `podman -- ARGS...` (executes the chosen native prune command) and `files` (DevArch-owned paths only). Never combine them implicitly.

## Native delegation

Podman owns container/image/volume/storage accounting and deletion. Native confirmation and filters remain visible. DevArch owns only its repository backup/recovery directories and must not infer Podman resource liveness.

## Outputs

- Cleanup script, policy documentation, test fixtures, and optional resource-label recommendations for Compose definitions.

## Acceptance criteria

- Podman cleanup is never run unless the user explicitly supplies a native prune operation; no hidden `--force` is added.
- Native Podman prompts/output identify runtime deletions; DevArch does not claim stronger candidate guarantees than Podman.
- Filesystem cleanup defaults to listing paths and requires explicit apply/confirmation; active recovery guards and newest backups are protected.
- Filesystem apply mode revalidates paths immediately before deletion.
- Partial failures are reported per resource and produce nonzero exit status.

## Verification

- Recording-Podman tests assert exact passthrough and prove no implicit `--force`, `--all`, or `--volumes` flags are added.
- Filesystem tests cover threshold boundaries, active guards, symlinks, changed paths, protected newest backups, and interrupted deletion.
- Optional native prune smoke testing uses disposable resources only.

## Non-goals

No custom Podman candidate inventory, disk accounting, ownership inference, prune algorithm, deletion of arbitrary app repositories, or automatic resolution of recovery guards. `podman system prune` remains available only as an explicit user-selected native command.
