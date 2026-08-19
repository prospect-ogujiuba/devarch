# Phase 02: Environment doctor

Created: 2026-08-19
Purpose: Turn common DevArch setup failures into one safe, actionable diagnostic report.

## Goal

Add `scripts/devarch/doctor.sh` to assess host prerequisites, repository configuration, and active runtime health without mutating the system.

## Scope

- Check Bash version and invoke `podman version`, `podman info`, and `podman compose version` rather than duplicating capability detection.
- Check `.env` existence, file mode, required values, and unchanged placeholder values without printing secrets.
- Use `podman network exists microservices-net`, `podman ps --all`, native health fields/`podman inspect`, `openssl x509`, and `openssl pkey` for runtime/certificate facts.
- Delegate Compose validity to `podman compose -f FILE config`; use `ss` or `lsof` only for host listeners that Podman cannot see.
- Classify DevArch policy checks as pass, warning, failure, or skipped; support focused `--section`. Add JSON only if a real consumer is implemented, otherwise preserve native text.

## Native delegation

Doctor sequences bounded read-only native commands and adds only DevArch expectations (required network, paths, placeholders, and known Compose files). It does not implement container health, certificate parsing, port scanning, or runtime diagnostics itself.

## Outputs

- `doctor.sh`, host-only tests, README reference, and stable check identifiers.

## Acceptance criteria

- Read-only invocation never creates networks, starts containers, edits files, or elevates privileges.
- Missing Podman still yields useful static checks and reports the failed native command.
- Each failure includes one concrete remediation and returns nonzero only for failures, not warnings.
- Secret values are absent from human and JSON output.
- The known static port collision can be reported with every owning service.

## Verification

- Fixture tests cover missing Podman/provider, inaccessible Podman storage, missing network, unhealthy service, placeholder environment, bad certificate, and port collision.
- Scan captured output for fixture secrets.
- Run `bash -n`, ShellCheck, and a manual read-only doctor invocation.

## Non-goals

Doctor does not repair configuration, install dependencies, trust certificates, start services, emulate `podman info`, or maintain its own runtime capability database.
