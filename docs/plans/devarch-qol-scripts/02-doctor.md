# Phase 02: Environment doctor

Created: 2026-08-19
Purpose: Turn common DevArch setup failures into one safe, actionable diagnostic report.

## Goal

Add `scripts/devarch/doctor.sh` to assess host prerequisites, repository configuration, and active runtime health without mutating the system.

## Scope

- Check Bash version, required utilities, runtime/Compose availability, and runtime access.
- Check `.env` existence, file mode, required values, and unchanged placeholder values without printing secrets.
- Check `microservices-net`, wildcard certificate/key presence and basic validity, repository/app permissions, Compose readability, port conflicts, and container health.
- Classify checks as pass, warning, failure, or skipped; support `--json` and focused `--section` execution.

## Outputs

- `doctor.sh`, host-only tests, README reference, and stable check identifiers.

## Acceptance criteria

- Read-only invocation never creates networks, starts containers, edits files, or elevates privileges.
- Missing runtime still yields useful static checks.
- Each failure includes one concrete remediation and returns nonzero only for failures, not warnings.
- Secret values are absent from human and JSON output.
- The known static port collision can be reported with every owning service.

## Verification

- Fixture tests cover no runtime, Podman, Docker, missing network, unhealthy service, placeholder environment, bad certificate, and port collision.
- Scan captured output for fixture secrets.
- Run `bash -n`, ShellCheck, and a manual read-only doctor invocation.

## Non-goals

Doctor does not repair configuration, install dependencies, trust certificates, or start services.
