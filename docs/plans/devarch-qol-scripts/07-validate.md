# Phase 07: Repository validation suite

Created: 2026-08-19
Purpose: Catch Compose, shell, catalog, port, and configuration regressions before they disrupt local services.

## Goal

Add `scripts/devarch/validate.sh` as the single host-only entrypoint for static checks and existing regression suites.

## Scope

- Validate YAML parsing, Compose structure/config, service/container naming, external-network declarations, referenced bind-mount files, healthchecks, image tags, and host-port uniqueness.
- Run Bash syntax, optional ShellCheck, hosts tests, Laravel tests, WordPress tests, and new DevArch tests.
- Support `--quick`, `--full`, `--section`, `--json`, and explicit handling when optional tools are unavailable.
- Maintain a reviewed allowlist for intentional exceptions, including job-style services without healthchecks.

## Outputs

- Validation script, checks/allowlist files, tests, and development-check documentation.

## Acceptance criteria

- Default validation performs no runtime or host mutation.
- Every diagnostic names the file, rule, and remediation.
- Duplicate published ports are failures unless explicitly documented in the exception file.
- Missing optional ShellCheck is a warning in quick mode and a failure only in an explicitly strict/full mode.
- Existing test exit statuses are aggregated without hiding later independent failures.

## Verification

- Mutation tests introduce one fixture defect per rule and assert its diagnostic.
- Confirm the current `8091` collision is detected until resolved or explicitly waived with rationale.
- Run quick and full modes; verify JSON and final exit status.

## Non-goals

No automatic rewriting of Compose files or downloading of validation dependencies.
