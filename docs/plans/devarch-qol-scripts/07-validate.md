# Phase 07: Repository validation suite

Created: 2026-08-19
Purpose: Catch Compose, shell, catalog, port, and configuration regressions before they disrupt local services.

## Goal

Add `scripts/devarch/validate.sh` as the single host-only entrypoint for static checks and existing regression suites.

## Scope

- Make `podman compose -f FILE config` the authoritative Compose parse/normalization check; do not build a competing Compose parser or schema.
- Run repository-specific assertions only where Compose cannot express policy: canonical path/name agreement, required external network, referenced repository files, healthcheck policy, pinning policy, and cross-file host-port uniqueness.
- Invoke native `bash -n`, `shellcheck`, and the existing hosts/Laravel/WordPress test entrypoints unchanged.
- Support `--quick`, `--full`, `--section`, `--json`, and explicit handling when optional tools are unavailable.
- Maintain a small reviewed allowlist only for DevArch policy exceptions, including job-style services without healthchecks.

## Native delegation

`podman compose config` owns Compose validity and interpolation. Bash and ShellCheck own shell analysis. The script is only a test runner plus cross-file DevArch policy checks; it must not normalize YAML or duplicate provider diagnostics.

## Outputs

- Validation script, checks/allowlist files, tests, and development-check documentation.

## Acceptance criteria

- Default validation performs no runtime or host mutation.
- Native diagnostic text and exit status are preserved; DevArch prefixes only the file/check context.
- Duplicate published ports are failures unless explicitly documented in the exception file.
- Missing optional ShellCheck is a warning in quick mode and a failure only in an explicitly strict/full mode.
- Existing test exit statuses are aggregated without hiding later independent failures.

## Verification

- Mutation tests introduce one fixture defect per rule and assert its diagnostic.
- Confirm the current `8091` collision is detected until resolved or explicitly waived with rationale.
- Run quick and full modes; verify JSON and final exit status.

## Non-goals

No Compose/YAML implementation, diagnostic reformatter, automatic rewriting, or downloading of validation dependencies.
