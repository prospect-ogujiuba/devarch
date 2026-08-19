# Phase 18: New-service scaffolder

Created: 2026-08-19
Purpose: Make new service definitions consistent with DevArch conventions from their first commit.

## Goal

Copy a small reviewed Compose template, substitute only validated DevArch fields, and hand validation to `podman compose config`.

## Scope

- Keep inputs minimal: category, service name, image, and one selected reviewed template (`http`, `command-health`, or `job`). More complex services are authored directly in Compose.
- Copy the selected template and substitute only strictly validated identifiers/ports using a standard text substitution tool; never generate arbitrary YAML structures in Bash.
- Generated Compose uses restart policy, loopback publishing, `microservices-net`, volume conventions, and a visible placeholder for the native container healthcheck.
- Run `podman compose -f STAGED_FILE config` and DevArch validation/port checks before moving the staged directory into place.

## Native delegation

Compose is the configuration language and `podman compose config` is the structural authority. DevArch provides starter files and repository conventions only. There is no service model, YAML builder, or lifecycle action.
- Provide `--dry-run` and refuse existing/nonempty targets.

## Outputs

- Scaffolder, templates, tests, contributor documentation, and optional endpoint-registry prompt/output.

## Acceptance criteria

- Names and paths are strictly validated and cannot traverse directories.
- Generated YAML parses and passes all applicable validation rules.
- No secrets or default passwords are embedded; required sensitive settings use documented Compose environment references or Podman secrets where appropriate.
- Files are staged in a temporary directory and moved into place only after validation.
- Existing files are never overwritten, including through symlinks or case collisions.
- Generated output remains simple enough to edit directly without the scaffolder.

## Verification

- Golden tests cover HTTP, command, and job templates plus invalid names, port conflicts, existing targets, symlink escapes, and validation failure.
- Run generated fixtures through YAML parsing and `validate.sh`.
- Manually review one generated service rather than starting it.

## Non-goals

No YAML builder/schema, image discovery, service-specific configuration generator, automatic startup, remote registry authentication, or migration of existing services.
