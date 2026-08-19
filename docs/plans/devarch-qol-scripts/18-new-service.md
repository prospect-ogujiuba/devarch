# Phase 18: New-service scaffolder

Created: 2026-08-19
Purpose: Make new service definitions consistent with DevArch conventions from their first commit.

## Goal

Add `scripts/devarch/new-service.sh` to generate a minimal, validated service directory from explicit inputs.

## Scope

- Inputs: category, service name, image, container port, optional host port, healthcheck type/target, volume, and config directory.
- Generate `services-library/<category>/<service>/compose.yml` with pinned image guidance, restart policy, loopback publishing, `microservices-net`, deterministic volume naming, and healthcheck.
- Support safe templates for HTTP, TCP/command, and job-style/no-healthcheck services with required rationale.
- Use the port helper for advisory conflict checks and validation suite before final publication.
- Provide `--dry-run` and refuse existing/nonempty targets.

## Outputs

- Scaffolder, templates, tests, contributor documentation, and optional endpoint-registry prompt/output.

## Acceptance criteria

- Names and paths are strictly validated and cannot traverse directories.
- Generated YAML parses and passes all applicable validation rules.
- No secrets or default passwords are embedded; required sensitive settings use documented environment references.
- Files are staged in a temporary directory and moved into place only after validation.
- Existing files are never overwritten, including through symlinks or case collisions.
- Generated output remains simple enough to edit directly without the scaffolder.

## Verification

- Golden tests cover HTTP, command, and job templates plus invalid names, port conflicts, existing targets, symlink escapes, and validation failure.
- Run generated fixtures through YAML parsing and `validate.sh`.
- Manually review one generated service rather than starting it.

## Non-goals

No image discovery, service-specific configuration generation, automatic startup, remote registry authentication, or schema migration of existing services.
