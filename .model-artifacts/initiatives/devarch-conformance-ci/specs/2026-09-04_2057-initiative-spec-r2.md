# Initiative Specification: DevArch Conformance & Golden-Path CI

- Topic: `devarch-conformance-ci`
- Revision: 2
- Created: 2026-09-04T20:57:54Z
- Predecessor: `.model-artifacts/initiatives/devarch-conformance-ci/specs/2026-09-04_2057-initiative-spec-r1.md`
- Initiative state: planning

## Problem

DevArch has 171 tracked Compose definitions containing 191 service entries and multiple WordPress, Laravel, JavaScript, Node routing, dashboard, host-management, and browser-test workflows. Its first-party checks are individually useful and currently pass, but they are discovered through separate READMEs, no tracked CI workflow runs them, current top-level Compose validation proves only YAML shape, and runtime integration checks may skip when local infrastructure is absent. A contributor therefore cannot obtain one authoritative answer that the catalog and representative clean-machine workflows still work.

Baseline observed before planning:

- 321 tracked files, including 171 recognized Compose definitions and 191 Compose service entries.
- 12 discovered local static/unit checks passed in 54.3 seconds.
- 0 tracked files under `.github/workflows/`.
- 169/190 service entries define health checks; 21 without them are concentrated in Plane, Taiga, and one-shot ERPNext jobs and need explicit policy treatment rather than blanket mutation.

## Outcome and observable behavior

A contributor or CI runner invokes one repository-owned verification entry point and receives a deterministic pass/fail result plus a secret-safe structured report. The entry point runs the complete registered static/unit suite, validates every catalog service against explicit conformance metadata and resolved Compose configuration, and can execute isolated runtime golden paths for WordPress, Laravel, a JavaScript framework, and database + application + proxy routing. Every runtime scenario proves readiness at an external boundary and tears down all resources and temporary files on success, failure, signal, or timeout. Required CI checks run automatically and preserve actionable evidence.

## Users

- Contributors changing scripts, Compose definitions, profiles, routing, or dashboard behavior.
- Maintainers reviewing pull requests and diagnosing regressions.
- Operators validating a checkout before using DevArch locally.

## Constraints

- Preserve Podman-first behavior and let `podman compose` select its configured external Compose provider; do not add an implicit Docker fallback to shared DevArch helpers.
- Support rootless execution without mixing container users.
- Do not mutate host DNS or `/etc/hosts` in CI; golden paths use `--no-hosts` or direct `Host` headers.
- Never print real `.env` values, generated credentials, tokens, or unredacted subprocess environments.
- Use unique project, network, container, volume, hostname, application, and temporary-directory names per run.
- Pin the CI operating-system image and tool setup; pin third-party actions by immutable commit SHA.
- The repository remains usable without GitHub Actions; the local verifier is authoritative.
- Phase contracts are implemented with behavior-first TDD and actual Red evidence only during implementation.

## Non-goals

- Proving all 170 services can run simultaneously.
- Replacing service-native health checks or application test suites.
- Supporting Docker as a new fallback where a workflow is currently Podman-only.
- Benchmarking application throughput or load capacity.
- Automatically changing repository branch-protection settings.
- Rewriting existing bootstrap commands, replacing Compose, or introducing a hosted control plane.

## Compatibility

Existing script paths, documented flags, Compose IDs, rootless Podman behavior, `podman-compose` provider compatibility, and direct per-component test commands remain supported. The unified verifier is additive. Existing service definitions may record time-bounded, justified policy exceptions when a health check is structurally inapplicable, such as a one-shot job; an exception is not a silent pass.

## Conformance model

Each `services-library/<category>/<service>/` directory containing one or more recognized Compose definitions receives one local `conformance.yml` that binds every definition in that directory with schema version, canonical service ID, validation tier, required fixture/environment variable names, health/readiness policy, smoke command or probe, timeout, cleanup expectations, and any justified exception with rationale and review date. The verifier rejects missing, duplicate, unknown, stale, path-mismatched, or secret-bearing metadata. New service directories cannot pass verification without metadata and at least a config-level smoke assertion; runtime-tier services require an external readiness probe.

## Migration and rollback

Conformance metadata is additive. First inventory all recognized Compose definitions without duplicate matches, generate reviewed metadata for all containing service directories, and record explicit exceptions for structurally inapplicable checks. Do not bulk-edit Compose files merely to satisfy policy. Rollback removes the additive verifier, metadata, and CI workflows; existing bootstrap and Compose entry points remain functional. CI required-check activation is a separate maintainer action and must occur only after the checks are stable on the default branch.

## Risks

- Upstream scaffolders and container registries can make runtime jobs flaky.
- Fixed host ports and shared `microservices-net` can cause cross-job collisions.
- Cleanup failure can leak rootless containers, volumes, networks, or generated apps.
- CI logs or artifacts can expose secrets unless redaction is fail-closed.
- Blanket health-check rules can misclassify one-shot jobs.
- A catalog-wide metadata migration can drift from the catalog unless completeness is machine-checked.

## Acceptance criteria

- **AC1 — Unified command:** one documented repository entry point supports `static`, `unit`, `golden`, and `all` scopes, returns nonzero on any required failure, and emits deterministic human output plus versioned JSON with check ID, scope, timestamps, duration, result, diagnostic, and artifact paths.
- **AC2 — Complete local suite:** the unit scope registers every existing first-party non-runtime check; a test fails when a registered check disappears or a known test file is unregistered. The current suite remains green.
- **AC3 — Catalog conformance:** every tracked recognized Compose definition (`compose.yml`, `compose.yaml`, `*.compose.yml`, or `*.compose.yaml`) resolves with the configured Compose provider using non-secret fixtures and has valid local conformance metadata. Policy covers service IDs, images/builds, required variables, ports, volumes, networks, health/readiness, smoke tier, timeouts, and exceptions.
- **AC4 — New-service gate:** adding a service directory without valid `conformance.yml` and its declared smoke assertion fails locally and in CI.
- **AC5 — Safe runtime harness:** golden scenarios use unique resource names, bounded readiness and execution timeouts, direct host-header or container-network probes, and idempotent teardown on success, failure, signal, and timeout; a post-run leak assertion passes.
- **AC6 — Representative golden paths:** clean-checkout scenarios prove WordPress HTTP plus database readiness, Laravel HTTP plus migration/database readiness, one pinned JavaScript scaffold through the isolated Node router, and a database + application + wildcard proxy route. No private repository or production secret is required.
- **AC7 — CI enforcement:** pull requests and default-branch pushes run the complete static/unit suite and the four golden paths on a pinned supported runner; required checks fail closed, upload bounded redacted reports, cancel superseded runs, and always execute teardown.
- **AC8 — Diagnostics:** for injected failures in each layer, the JSON report names the first failing check/resource, preserves the relevant bounded log, and uploads or writes actionable evidence within five minutes of failure. Over a rolling 30 completed golden runs, successful clean-run rate is at least 99%; until 30 runs exist the dashboard reports `insufficient-sample` rather than claiming the target.
- **AC9 — Compatibility and documentation:** existing direct commands and tests still pass; README documents prerequisites, scopes, exit codes, metadata authoring, local reproduction, CI ownership, exception review, branch-protection activation, and rollback.

## Revision note

r2 corrects the inventory from 170/190 to 171/191 and explicitly includes `services-library/backend/node/app.compose.yml` and future recognized `*.compose.y*ml` templates in discovery, metadata binding, and verification.

## Open blockers

None for implementation of the first contract. Execution of runtime contracts requires a rootless Podman runner with outbound registry/package access. Final branch-protection activation requires repository administrator action and is deliberately outside automated mutation.
