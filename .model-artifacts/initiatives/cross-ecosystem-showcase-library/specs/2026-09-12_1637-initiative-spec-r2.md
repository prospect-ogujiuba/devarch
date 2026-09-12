# Initiative specification r2: cross-ecosystem-showcase-library

- Topic: `cross-ecosystem-showcase-library`
- Status: plan review
- Created: 2026-09-12 16:37 -0400
- Predecessor: `.model-artifacts/initiatives/cross-ecosystem-showcase-library/specs/2026-09-12_1637-initiative-spec-r1.md`
- Incorporated findings: DSA, TDD, security, migration/operations/compatibility findings dated 2026-09-12 16:37 -0400

## Problem

DevArch has a deterministic JavaScript showcase matrix, but generated applications under `apps/` are ignored local workspaces and Laravel/WordPress have no equivalent reproducible showcase definitions. Users cannot enumerate, recreate, start, stop, or verify showcase applications consistently across ecosystems. Laravel's current `bare`, `standard`, and `loaded` profiles express only Mailpit, Redis, and Composer packages; they cannot safely select official starter kits or reviewed Laravel ecosystem capabilities. Multi-application combinations such as a Laravel API plus TanStack frontend have no tracked recipe or coordinated recovery model.

## Outcomes and observable behavior

1. A top-level showcase command enumerates tracked ecosystem/profile definitions in deterministic order, reports adapter capability negotiation, and delegates supported lifecycle operations to ecosystem-owned adapters without becoming a container-runtime abstraction.
2. Existing JavaScript bootstrap and matrix commands remain compatible while becoming addressable through the top-level showcase surface.
3. Laravel gains a data-only profile/capability model with constrained, reviewed directives; arbitrary shell, PHP, Composer script, or Artisan text is never accepted from profile data.
4. Laravel supports reproducible, current starter profiles for `minimal`, `web`, `api`, `livewire`, `react`, `vue`, `svelte`, `queues`, `realtime`, `admin`, `search`, `observability`, and `testing`.
5. A Laravel matrix can list, create, start, stop, and verify deterministic `apps/showcase-laravel-<profile>` applications, resume bulk creation, isolate failures, and preserve existing applications unless replacement is explicit.
6. A tracked multi-app recipe model can create related applications transactionally enough to recover safely, beginning with Laravel Sanctum API plus TanStack Start frontend.
7. WordPress is exposed through an adapter using its existing mature profiles; this initiative does not invent additional WordPress profiles.
8. Generated apps remain ignored and disposable; tracked definitions, adapters, tests, and documentation are the reproducible library.

## Users

- Developers evaluating frameworks and Laravel stacks locally.
- Maintainers adding ecosystem adapters, profiles, capabilities, or recipes.
- CI and reviewers validating scaffold definitions without committing generated applications.

## Constraints

- Preserve current `scripts/javascript/*`, `scripts/laravel/bootstrap.sh`, and `scripts/wordpress/bootstrap.sh` entry points and documented behavior unless an exact contract explicitly extends them.
- Keep profiles and recipes data-only. Parse allowlisted directives; do not `source` new cross-ecosystem, Laravel capability, or recipe files.
- Do not accept arbitrary command directives. Capability IDs resolve through reviewed code-owned handlers.
- Respect DevArch's shared-script contract: delegate native Podman/Compose behavior unchanged and do not build a runtime feature matrix.
- Use deterministic DNS-safe app names, lexical discovery and DAG ready-node order, bounded retries/timeouts, concise logs, atomic durable recovery markers, backup-before-replace, and secret-safe dry runs.
- Every adapter/profile declares support tier, supported actions, required catalog service IDs/health probes, and compatibility probes. Starter-kit and package compatibility must be resolved against the generated Laravel major version before later capability mutation where feasible, with exact failure reporting and quarantine on partial creation.
- Default tests are offline fixtures/characterization. Opt-in real golden paths require unique run IDs, disk/network prerequisite checks, bounded execution, preserved failure logs, and explicit cleanup.
- External-account capabilities (billing, WorkOS, social providers) may be catalogued as deferred recipes but must not be part of unattended golden paths.
- Existing uncommitted hosts-helper and sample-file changes are outside implementation scope and must be preserved.

## Non-goals

- Committing generated `apps/*` trees.
- Supporting every framework or every Laravel package in the first release.
- Replacing ecosystem-native bootstraps with one universal implementation.
- Adding arbitrary hooks or executable profile commands.
- Running all generated applications concurrently by default.
- Adding new WordPress profiles.
- Provisioning production infrastructure or external SaaS accounts.
- Guaranteeing cross-version compatibility beyond the generated current stable major and explicitly tested constraints.

## Compatibility

- Existing JavaScript matrix names and commands remain functional.
- Existing Laravel `bare`, `standard`, and `loaded` remain canonical public profile files/names with unchanged CLI output and semantics; new profiles are additive.
- Existing WordPress profiles and bootstrap semantics remain unchanged.
- The top-level command is additive. Adapter protocol changes are versioned and fail closed on unsupported versions/actions.
- Profile additions must not change an existing generated app without explicit replacement.
- Existing `apps/showcase-laravel` and `apps/showcase-wp` are unmanaged local apps; the new system never adopts, renames, overwrites, or deletes them automatically.

## Migration and rollback

- Introduce adapters and the top-level command additively.
- Extend Laravel parsing behind explicit new directives while retaining legacy directives.
- Map legacy Laravel profiles to equivalent capability sets and characterize their plans before refactoring internals.
- Generated test/showcase targets use existing backup, failed-target quarantine, and recovery markers.
- Each implementation contract is independently revertible. Removing the top-level orchestrator leaves native entry points usable; removing new profiles does not delete generated apps.
- Recipe rollback preserves completed app trees or moves partial trees to `.devarch-failed` with a versioned, secret-free, atomically replaced recovery record bound to recipe hash and prefix; it never silently deletes pre-existing applications.

## Risks

- Upstream starter-kit/package drift can make supposedly curated profiles incompatible.
- A universal adapter contract can erase important lifecycle differences between ecosystems.
- Profile inheritance or unrestricted hooks can create hidden execution and ordering behavior.
- Bulk scaffolding is network- and disk-heavy and can leave partial targets.
- Multi-app recipe failure can produce mismatched credentials, URLs, or half-created applications.
- External-service profiles may appear usable without required credentials.

## Acceptance criteria

- AC-01: `scripts/showcase/showcase.sh list` deterministically reports ecosystem, profile/recipe, generated application name(s), support tier, supported actions, prerequisites, and local state without reading generated app metadata as the source of truth.
- AC-02: Top-level create/start/stop/verify/status/resume dispatch passes validated arguments only to adapter-declared supported actions, preserves exit status/actionable output, and reports unsupported lifecycle actions without emulation.
- AC-03: Existing JavaScript matrix regression tests pass unchanged or with additive assertions; direct commands remain compatible.
- AC-04: Laravel profiles are parsed as data, reject controls, unknown directives, duplicate singleton directives, unknown capabilities, cycles/implicit inheritance, and arbitrary command text before mutation.
- AC-05: Legacy Laravel `bare`, `standard`, and `loaded` remain canonical names/files and produce behavior-equivalent secret-safe plans after internal migration to the capability model.
- AC-06: The 13 curated Laravel profiles list deterministically with support tier, compatibility probe, required service IDs/health probes, and resolved-version evidence; each passes offline syntax/preflight tests, while opt-in representative `minimal`, `api`, `react`, `queues`, and `observability` provisioning smokes pass or report explicit unavailable prerequisites.
- AC-07: Laravel bulk creation is resumable, skips complete targets, refuses incomplete/pre-existing targets without explicit recovery/replacement, retries only bounded transient scaffold failures, continues independent profiles, and summarizes failures nonzero.
- AC-08: Capability installation order is deterministic, duplicates are idempotently coalesced at first occurrence, conflicts and missing catalog services fail preflight, and service dependencies start only when selected and pass declared health probes.
- AC-09: No secrets appear in dry-run, logs, generated tracked definitions, process arguments where stdin/file alternatives exist, or test fixtures.
- AC-10: The Laravel–TanStack recipe creates two separately routable ignored apps with coordinated API URL, CORS/Sanctum configuration, deterministic names, lexical DAG scheduling, and status/resume over atomically persisted recoverable partial-failure state.
- AC-11: WordPress existing profiles can be listed and individually provisioned through its adapter without changing native profile contents.
- AC-12: Documentation distinguishes tracked definitions from ignored generated apps and explains extension, lifecycle, support tiers, cleanup, recovery, and external prerequisites.
- AC-13: Focused offline unit/fixture/characterization tests, shell static checks where available, native bootstrap suites, matrix integration tests, secret-canary scans, and explicitly opt-in bounded runtime HTTPS checks pass.

## Open blockers

None. Exact third-party package constraints and installer flags must be researched, compatibility-probed, and recorded within their implementation contracts rather than assumed from planning-time versions. External-account, multi-tenancy, billing/social, and Octane ideas remain documented future candidates only.
