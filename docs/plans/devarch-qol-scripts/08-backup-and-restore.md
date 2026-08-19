# Phase 08: Backup and restore

Created: 2026-08-19
Purpose: Make application and database recovery deliberate, verifiable, and independent of bootstrap replacement backups.

## Goal

Add the thinnest safe selectors around Podman volume transfer and native database dump/restore tools; avoid inventing a backup platform.

## Scope

- Use `podman volume inspect`, `podman volume export`, and `podman volume import` for explicit named-volume snapshots where crash-consistent filesystem copies are acceptable.
- Use native logical tools for stateful databases: `mariadb-dump`/`mariadb`, `pg_dump`/`pg_restore`/`psql`, and `mongodump`/`mongorestore`, invoked through `podman exec` or a purpose-built client container.
- Use `tar` for app trees and `sha256sum`/`shasum` for integrity; do not create a custom archive format.
- If metadata is required, use a short human-readable sidecar recording exact native commands, versions, resource name, timestamp, and checksums—not a general versioned backup schema.
- Restore validates checksums and native tool availability, then invokes the corresponding native restore command with its output/prompts intact.
- Keep WordPress `.wpress` workflows distinct and call the established native WP-CLI command.

## Native delegation

Podman owns volume export/import and process execution. Database vendors own logical consistency and restore semantics. `tar`/checksum tools own archiving and integrity. DevArch only resolves resource names, chooses the documented native method, and prevents accidental target mismatch.

## Outputs

- Thin target selectors, native volume/database command recipes, optional plain metadata sidecar, fixtures, tests, and recovery runbook.

## Acceptance criteria

- Native command failure stops the operation; a staging directory is renamed only after native output and checksums succeed.
- Restore verifies checksums and matching resource/tool type before the first mutation; deeper compatibility remains the native tool's decision.
- Existing targets are refused by default; replacement requires explicit confirmation and a pre-restore safety backup.
- Secrets never appear in argv logs, manifests, or console output.
- Retention is deferred to phase 09 or documented native filesystem tools; backup commands do not implement a retention engine.

## Verification

- Recording-command tests assert exact Podman/database/tar/checksum argv and use corrupt/truncated fixtures.
- Round-trip smoke tests for each database type are optional, explicitly destructive, and use disposable names.
- Test interruption, insufficient disk space, checksum mismatch, mismatched resource/tool metadata, and native restore failure behavior.

## Open questions

Choose the default backup root and whether Git-managed app sources need archives. Confirm per database whether logical dump or Podman volume export is the documented default; logical dumps are recommended for databases.

## Non-goals

No backup daemon, catalog/database, custom archive/manifest protocol, remote/cloud upload, encryption key management, live filesystem snapshotting, retention engine, or guarantee of cross-major-version database restore.
