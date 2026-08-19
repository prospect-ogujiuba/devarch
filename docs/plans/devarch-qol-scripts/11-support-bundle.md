# Phase 11: Redacted support bundle

Created: 2026-08-19
Purpose: Capture enough bounded evidence to troubleshoot DevArch without sharing secrets or entire application workspaces.

## Goal

Add `scripts/devarch/support-bundle.sh` that assembles a deterministic, reviewable diagnostic archive.

## Scope

- Capture runtime/tool versions, Git revision/status summary, doctor/status/ports/validate results, selected Compose configs, bounded recent logs, disk usage, and certificate metadata.
- Redact environment assignments, credentials, tokens, URL userinfo, authorization headers, and configured custom patterns.
- Exclude `.env`, private keys, application content, database data, volumes, credentials files, and unbounded logs by construction.
- Generate a manifest with included files, commands, exit statuses, checksums, and redaction count; support directory output before archive creation.

## Outputs

- Support-bundle script, redaction library/tests, manifest schema, and sharing checklist.

## Acceptance criteria

- Collection is read-only and works partially when runtime commands fail.
- Archive paths cannot escape the staging root and are stable across platforms.
- Known secret fixtures do not occur in filenames, contents, manifest, or terminal output.
- Log collection has explicit service, line, byte, and time bounds.
- User is shown the staging path and instructed to review it before sharing.

## Verification

- Seed canary secrets in all supported forms and scan the resulting archive byte-for-byte.
- Test command failures, huge logs, symlinks, weird filenames, missing tools, and archive-tool fallback.
- Verify manifest checksums and deterministic file ordering.

## Non-goals

No automatic upload, remote ticket creation, memory dump, database dump, or collection of application source.
