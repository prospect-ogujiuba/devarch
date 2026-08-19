# Phase 09: Conservative cleanup

Created: 2026-08-19
Purpose: Reclaim space and remove stale recovery artifacts without turning convenience into accidental data loss.

## Goal

Add `scripts/devarch/cleanup.sh` with auditable policies and dry-run as the default behavior.

## Scope

- Inventory stopped DevArch containers, dangling images, explicitly labeled unused volumes, old app backups, failed provisions, resolved recovery guards, and incomplete backup staging directories.
- Add selectors, age thresholds, size summaries, `--apply`, and `--yes`.
- Attribute runtime resources to DevArch using labels or exact catalog identities; unknown resources remain untouched.
- Require phase 08 manifests before pruning managed backups.

## Outputs

- Cleanup script, policy documentation, test fixtures, and optional resource-label recommendations for Compose definitions.

## Acceptance criteria

- Invocation without `--apply` cannot mutate anything.
- Volumes, active recovery guards, newest successful backups, running containers, and unowned runtime resources are never removed automatically.
- Every candidate includes reason, age, size when available, and exact proposed action.
- Apply mode revalidates candidates immediately before deletion to reduce races.
- Partial failures are reported per resource and produce nonzero exit status.

## Verification

- Fake filesystem/runtime tests cover threshold boundaries, active guards, symlinks, changed candidates, protected newest backups, and interrupted deletion.
- Assert exact delete calls and verify dry-run records none.
- Optional manual smoke test uses only disposable labeled resources.

## Non-goals

No blanket `system prune`, deletion of arbitrary app repositories, or automatic resolution of recovery guards.
