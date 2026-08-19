# Phase 08: Backup and restore

Created: 2026-08-19
Purpose: Make application and database recovery deliberate, verifiable, and independent of bootstrap replacement backups.

## Goal

Add `scripts/devarch/backup.sh` and `scripts/devarch/restore.sh` with manifest-driven backups for supported applications and databases.

## Scope

- Back up app trees plus MariaDB/MySQL, PostgreSQL, and MongoDB using native tools in their containers.
- Support `--app`, `--database`, `--all`, output directory, compression, and retention selection.
- Store a versioned manifest containing types, logical names, timestamps, tool/image versions, file checksums, and redacted provenance.
- Restore only from validated manifests; provide preflight, `--dry-run`, explicit confirmation, and collision policy.
- Keep WordPress `.wpress` workflows distinct and documented rather than pretending they are SQL backups.

## Outputs

- Backup/restore scripts, versioned manifest schema, adapter functions, fixtures, tests, and recovery runbook.

## Acceptance criteria

- Backup writes to a staging directory and atomically publishes only after dumps and checksums succeed.
- Partial backups are marked failed and never presented as restorable.
- Restore verifies every checksum and compatibility rule before the first mutation.
- Existing targets are refused by default; replacement requires explicit confirmation and a pre-restore safety backup.
- Secrets never appear in argv logs, manifests, or console output.
- Retention never deletes the newest successful backup for a selected resource.

## Verification

- Host-only adapter tests use fake database clients and corrupt/truncated archives.
- Round-trip smoke tests for each database type are optional, explicitly destructive, and use disposable names.
- Test interruption, insufficient disk space, checksum mismatch, unknown manifest version, and rollback behavior.

## Open questions

Choose the default backup root and whether app source trees already managed by Git should be included by default. Resolve before implementation.

## Non-goals

No remote/cloud upload, encryption key management, live filesystem snapshotting, or guarantee of cross-major-version database restore.
