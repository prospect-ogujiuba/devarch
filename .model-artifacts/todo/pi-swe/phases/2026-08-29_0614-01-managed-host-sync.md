# 01-managed-host-sync

Created: 2026-08-29T06:14:22.253Z
Purpose: Contract for the managed hosts synchronization implementation.

# Managed host synchronization

Created: 2026-08-23
Purpose: Add a repeatable command that safely synchronizes all DevArch service and app domains into the system hosts file.

## Goal

Generate a deterministic managed block mapping discovered `.test` domains to `127.0.0.1`, with one elevation operation and no changes outside that block.

## Scope

- Add `scripts/hosts/sync-hosts.sh` for discovery, validation, dry-run, Unix updates, and Windows delegation.
- Add `scripts/hosts/sync-hosts.ps1` for Windows managed-block updates and UAC elevation.
- Add fixture-based tests under `scripts/hosts/`.
- Document usage in `scripts/hosts/README.md` and the top-level README where appropriate.

## Outputs

- Stable sorted service and app domain groups plus `devarch.test`.
- `# BEGIN DEVARCH HOSTS` / `# END DEVARCH HOSTS` managed block.
- Idempotent replacement that preserves unrelated content.

## Acceptance criteria

- Service names come from literal `container_name` values in `services-library/**/compose.yml`.
- App names come from non-hidden `apps/*` directories with a routing marker (`index.php`, `public/index.php`, `public/index.html`, or `package.json`).
- Invalid DNS labels fail rather than entering the hosts file.
- `--dry-run` prints the generated block without editing.
- Re-running causes no write when the block is current.
- Unix fixture tests prove insertion, replacement, preservation, discovery, and dry-run behavior.

## Verification

- `bash -n` on shell scripts.
- Focused fixture test.
- Existing single-host registration test.
- Dry-run against this repository and count discovered domains.

## Non-goals

- Removing matching domains outside the managed block.
- Verifying that every catalog service is currently running.
- Replacing local wildcard DNS.
