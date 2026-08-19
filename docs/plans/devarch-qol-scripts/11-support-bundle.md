# Phase 11: Redacted support bundle

Created: 2026-08-19
Purpose: Capture enough bounded evidence to troubleshoot DevArch without sharing secrets or entire application workspaces.

## Goal

Add `scripts/devarch/support-bundle.sh` that assembles a deterministic, reviewable diagnostic archive.

## Scope

- Capture native `podman version`, `podman info --debug`, `podman ps --all`, `podman inspect`, `podman network inspect`, `podman system df`, selected `podman compose ... config`, bounded `podman logs`, Git summaries, and `openssl x509` metadata.
- Redact environment assignments, credentials, tokens, URL userinfo, authorization headers, and configured custom patterns.
- Exclude `.env`, private keys, application content, database data, volumes, credentials files, and unbounded logs by construction.
- Generate a simple inventory with exact native commands, exit statuses, included files, and checksums; use native `tar` for the archive.

## Native delegation

Each collected fact comes from the owning tool. DevArch only bounds command scope, stores output files, and applies defense-in-depth redaction. It does not consume a custom status schema or reinterpret Podman diagnostics.

## Outputs

- Support-bundle script, redaction tests, plain command/file inventory, native tar archive, and sharing checklist.

## Acceptance criteria

- Collection is read-only and works partially when runtime commands fail.
- Native command output is preserved apart from documented redaction; archive paths cannot escape the staging root.
- Known secret fixtures do not occur in filenames, contents, manifest, or terminal output.
- Log collection has explicit service, line, byte, and time bounds.
- User is shown the staging path and instructed to review it before sharing.

## Verification

- Seed canary secrets in all supported forms and scan the resulting archive byte-for-byte.
- Test command failures, huge logs, symlinks, weird filenames, missing tools, and archive-tool fallback.
- Verify manifest checksums and deterministic file ordering.

## Non-goals

No runtime diagnostic implementation, custom status parser, automatic upload, remote ticket creation, memory dump, database dump, or collection of application source.
