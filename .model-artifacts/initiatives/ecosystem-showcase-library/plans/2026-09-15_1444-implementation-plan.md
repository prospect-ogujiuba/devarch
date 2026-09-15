# Ecosystem showcase library — implementation plan

## How to use this plan with /swe

- `../workflow.json` is the **sole mutable lifecycle authority** for task order, dependencies, assessments, statuses, checks, and execution evidence. This document supplies the full design; it is not a second manifest or runner.
- This preserves the approved spec r2 and plan r3, including all twelve executable task bodies, four phase descriptions, the acceptance-to-evidence matrix, specialist decisions, and review rationale. Read the specification, common verification rules, and the active task's full section before implementing it. Summary acceptance in the JSON does not replace this detail.
- Scope remains planning-only. No implementation, protected verification evidence, or completion is claimed. All twelve tasks remain pending; `SHOW-001` is initially ready. Workflow revision numbers describe migration edits, not legacy plan revisions.
- Historical findings/reviews below are context, not live gates. Their earlier task numbers and dispositions are superseded by the r3 task sections and the workflow. No legacy specialist dispatch, hash approval gate, or per-stage runner is reinstated.
- `REC-001` had a contradictory acceptance sentence about declaration-order ties. The approved spec, DSA decision, contract scope, and tests consistently require **lexical node-ID tie-breaking**; that one sentence is reconciled here. No declaration-order scheduling is introduced.
- The source spec permits explicit recovery/replacement of managed targets with backup first. It never authorizes automatically adopting, replacing, or deleting unmanaged `apps/showcase-laravel` or `apps/showcase-wp`. Do not confuse these names with matrix-generated `apps/showcase-laravel-<profile>` targets.
- Obsolete r1/r2 execution trees are not restored. Their exact contents, and the original approval records, remain recoverable from the source commit and blob inventory at the end of this document.

## Common execution and verification rules

1. Use normal `/swe` start/resume, protected `bash`, `swe_workflow verify`, and complete. Bind each planned command separately immediately after its result. All planned checks are required; after mutation rerun them. Do not use a simulated lifecycle or `allowGap` as implementation evidence.
2. For every implementation task, capture characterization or failing tests before production changes, then Green/refactor results. Apply the detailed task-specific fixtures and rollback checks below, not merely the existing baseline suites.
3. All default tests use temporary repository roots, fake adapters/bootstrap/Composer/Artisan/installers, argv call logs, deterministic clock/hash seams, and test-only failure injection. They must not modify real apps, hosts files, or credentials or perform real network scaffolding. These rules apply from `SHOW-001`, not only at `VER-001`.
4. Run focused fixtures first, adjacent native suites second, `bash -n` on affected shell files, and ShellCheck when installed. Preserve the secret-canary and malicious-input corpus, including symlink/escape, forged output, stale-state, and interrupted-write cases. Absence of ShellCheck is reported, not represented as a passing lint run.
5. Planned future test runners in `workflow.json` must be created/extended by their owning tasks. They are not existing passing tests. The migration selected `scripts/showcase/recipe.test.sh` as the recipe test entry point; it does not mandate a standalone `recipe.sh` implementation. If the implementation uses different files or additional focused checks, revise the full workflow graph and reassess incomplete tasks before continuing, preserving scope and dependencies.
6. The JSON initially records offline entry points. Before completing a task, add and run its concrete additional syntax, lint, dry-run, and applicable smoke checks through `swe_workflow` action `revise` and protected verification. Legacy prose did not specify final CLI flags, installer versions, or all smoke commands; do not guess them or drop their requirements. Keep invocation details and outcomes in execution evidence.
7. Network/runtime golden paths are explicitly opt-in, sequential, uniquely named, disk/network/service-prerequisite checked, bounded by timeouts, and retain logs and cleanup/recovery instructions. Do not run them merely because this migration was authorized. An unavailable prerequisite is an explicit gap, never a success; block work that cannot satisfy its acceptance rather than silently marking it complete.
8. Preserve these task-specific runtime requirements: `LAR-002` one starter `none`/minimal smoke; `LAR-004` the five profiles **minimal, api, react, queues, observability** with HTTPS/Artisan/service checks; `LAR-005` one-profile create/status/verify/stop; `REC-002` two-app HTTPS/API/CORS and second-node failure/resume; `VER-001` integrated five-profile, existing JavaScript runtime, and recipe golden paths. `SHOW-003` also exercises one adapted existing-profile dry-run. `LAR-003` does not provision a real multi-capability profile before `LAR-004`.
9. Independent implementation review, documentation checks, and an acceptance-to-evidence report remain substantive verification obligations, not a reinstatement of legacy workflow machinery. Record AC-01 through AC-13 outcomes and any blocked checks without inventing review or runtime evidence. No new performance/concurrency/UI work is implied; diagnosis, performance, and accessibility-UX remain not applicable unless task discovery changes that assessment. Compatibility is covered through migration, operations, and acceptance, not an unsupported `/swe` approach name.

<a id="specification"></a>

## Approved specification (r2)

### Problem

DevArch has a deterministic JavaScript showcase matrix, but generated applications under `apps/` are ignored local workspaces and Laravel/WordPress have no equivalent reproducible showcase definitions. Users cannot enumerate, recreate, start, stop, or verify showcase applications consistently across ecosystems. Laravel's current `bare`, `standard`, and `loaded` profiles express only Mailpit, Redis, and Composer packages; they cannot safely select official starter kits or reviewed Laravel ecosystem capabilities. Multi-application combinations such as a Laravel API plus TanStack frontend have no tracked recipe or coordinated recovery model.

### Outcomes and observable behavior

1. A top-level showcase command enumerates tracked ecosystem/profile definitions in deterministic order, reports adapter capability negotiation, and delegates supported lifecycle operations to ecosystem-owned adapters without becoming a container-runtime abstraction.
2. Existing JavaScript bootstrap and matrix commands remain compatible while becoming addressable through the top-level showcase surface.
3. Laravel gains a data-only profile/capability model with constrained, reviewed directives; arbitrary shell, PHP, Composer script, or Artisan text is never accepted from profile data.
4. Laravel supports reproducible, current starter profiles for `minimal`, `web`, `api`, `livewire`, `react`, `vue`, `svelte`, `queues`, `realtime`, `admin`, `search`, `observability`, and `testing`.
5. A Laravel matrix can list, create, start, stop, and verify deterministic `apps/showcase-laravel-<profile>` applications, resume bulk creation, isolate failures, and preserve existing applications unless replacement is explicit.
6. A tracked multi-app recipe model can create related applications transactionally enough to recover safely, beginning with Laravel Sanctum API plus TanStack Start frontend.
7. WordPress is exposed through an adapter using its existing mature profiles; this initiative does not invent additional WordPress profiles.
8. Generated apps remain ignored and disposable; tracked definitions, adapters, tests, and documentation are the reproducible library.

### Users

- Developers evaluating frameworks and Laravel stacks locally.
- Maintainers adding ecosystem adapters, profiles, capabilities, or recipes.
- CI and reviewers validating scaffold definitions without committing generated applications.

### Constraints

- Preserve current `scripts/javascript/*`, `scripts/laravel/bootstrap.sh`, and `scripts/wordpress/bootstrap.sh` entry points and documented behavior unless an exact contract explicitly extends them.
- Keep profiles and recipes data-only. Parse allowlisted directives; do not `source` new cross-ecosystem, Laravel capability, or recipe files.
- Do not accept arbitrary command directives. Capability IDs resolve through reviewed code-owned handlers.
- Respect DevArch's shared-script contract: delegate native Podman/Compose behavior unchanged and do not build a runtime feature matrix.
- Use deterministic DNS-safe app names, lexical discovery and DAG ready-node order, bounded retries/timeouts, concise logs, atomic durable recovery markers, backup-before-replace, and secret-safe dry runs.
- Every adapter/profile declares support tier, supported actions, required catalog service IDs/health probes, and compatibility probes. Starter-kit and package compatibility must be resolved against the generated Laravel major version before later capability mutation where feasible, with exact failure reporting and quarantine on partial creation.
- Default tests are offline fixtures/characterization. Opt-in real golden paths require unique run IDs, disk/network prerequisite checks, bounded execution, preserved failure logs, and explicit cleanup.
- External-account capabilities (billing, WorkOS, social providers) may be catalogued as deferred recipes but must not be part of unattended golden paths.
- Existing uncommitted hosts-helper and sample-file changes are outside implementation scope and must be preserved.

### Non-goals

- Committing generated `apps/*` trees.
- Supporting every framework or every Laravel package in the first release.
- Replacing ecosystem-native bootstraps with one universal implementation.
- Adding arbitrary hooks or executable profile commands.
- Running all generated applications concurrently by default.
- Adding new WordPress profiles.
- Provisioning production infrastructure or external SaaS accounts.
- Guaranteeing cross-version compatibility beyond the generated current stable major and explicitly tested constraints.

### Compatibility

- Existing JavaScript matrix names and commands remain functional.
- Existing Laravel `bare`, `standard`, and `loaded` remain canonical public profile files/names with unchanged CLI output and semantics; new profiles are additive.
- Existing WordPress profiles and bootstrap semantics remain unchanged.
- The top-level command is additive. Adapter protocol changes are versioned and fail closed on unsupported versions/actions.
- Profile additions must not change an existing generated app without explicit replacement.
- Existing `apps/showcase-laravel` and `apps/showcase-wp` are unmanaged local apps; the new system never adopts, renames, overwrites, or deletes them automatically.

### Migration and rollback

- Introduce adapters and the top-level command additively.
- Extend Laravel parsing behind explicit new directives while retaining legacy directives.
- Map legacy Laravel profiles to equivalent capability sets and characterize their plans before refactoring internals.
- Generated test/showcase targets use existing backup, failed-target quarantine, and recovery markers.
- Each implementation contract is independently revertible. Removing the top-level orchestrator leaves native entry points usable; removing new profiles does not delete generated apps.
- Recipe rollback preserves completed app trees or moves partial trees to `.devarch-failed` with a versioned, secret-free, atomically replaced recovery record bound to recipe hash and prefix; it never silently deletes pre-existing applications.

### Risks

- Upstream starter-kit/package drift can make supposedly curated profiles incompatible.
- A universal adapter contract can erase important lifecycle differences between ecosystems.
- Profile inheritance or unrestricted hooks can create hidden execution and ordering behavior.
- Bulk scaffolding is network- and disk-heavy and can leave partial targets.
- Multi-app recipe failure can produce mismatched credentials, URLs, or half-created applications.
- External-service profiles may appear usable without required credentials.

### Acceptance criteria

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

### Open blockers

None. Exact third-party package constraints and installer flags must be researched, compatibility-probed, and recorded within their implementation contracts rather than assumed from planning-time versions. External-account, multi-tenancy, billing/social, and Octane ideas remain documented future candidates only.

<a id="execution-map"></a>

## Approved execution map and traceability (r3)

### Ordered phases and contracts

1. `phase-01-contract-and-adapters`
   1. `SHOW-001` — top-level showcase contract and orchestrator.
   2. `SHOW-002` — JavaScript compatibility adapter; depends on `SHOW-001`.
   3. `SHOW-003` — WordPress existing-profile adapter; depends on `SHOW-001`.
2. `phase-02-laravel-profile-system`
   1. `LAR-001` — profile parser and legacy characterization; depends on `SHOW-001`.
   2. `LAR-002` — official starter engine; depends on `LAR-001`.
   3. `LAR-003` — reviewed capability handlers; depends on `LAR-001`.
   4. `LAR-004` — curated Laravel profile catalog; depends on `LAR-002`, `LAR-003`.
   5. `LAR-005` — native Laravel scaffold matrix; depends on `LAR-004`.
   6. `LAR-006` — Laravel showcase adapter; depends on `LAR-005`.
3. `phase-03-recipes`
   1. `REC-001` — recoverable multi-app recipe engine; depends on `SHOW-002`, `LAR-006`.
   2. `REC-002` — Laravel API plus TanStack Start recipe; depends on `REC-001`.
4. `phase-04-verification-and-docs`
   1. `VER-001` — golden paths, compatibility suite, and documentation; depends on `SHOW-003`, `REC-002`.

### Dependency graph

`SHOW-001 -> {SHOW-002, SHOW-003, LAR-001}`

`LAR-001 -> {LAR-002, LAR-003} -> LAR-004 -> LAR-005 -> LAR-006`

`{SHOW-002, LAR-006} -> REC-001 -> REC-002`

`{SHOW-003, REC-002} -> VER-001`

First ready after approval: `SHOW-001`.

### Specialist applicability

- Diagnosis: not required — no unexplained defect is in scope.
- DSA: complete — [preserved finding](#finding-dsa); r2 incorporates arrays/maps, lexical Kahn ordering, no inheritance, and atomic versioned recovery records.
- TDD: complete — [preserved finding](#finding-tdd); r2 contracts incorporate characterization, Red order, seams, and offline/opt-in separation.
- Security: complete — [preserved finding](#finding-security); r2 incorporates fixed roots, non-symlink/data-only parsing, argv boundaries, secret canaries, and trusted-upstream disclosure.
- Migration: complete — [preserved finding](#finding-migration-operations-compatibility); r2 preserves canonical legacy names and unmanaged apps.
- Performance: not required — operations are network/disk bound and sequential by design; measure before adding concurrency.
- Accessibility/UX: not required — no graphical UI changes; CLI clarity is covered by acceptance tests.
- Operations: complete — shared migration/operations/compatibility finding; r2 adds support metadata, prerequisites, timeouts, unique runs, logs, atomic recovery, status/resume, and cleanup.
- Compatibility: complete — shared migration/operations/compatibility finding; r2 adds adapter action negotiation and version compatibility probes/evidence.

### Outcome traceability

- Top-level library and adapters: `SHOW-001`, `SHOW-002`, `SHOW-003` → AC-01, AC-02, AC-03, AC-11.
- Laravel parser/starter/capabilities: `LAR-001`, `LAR-002`, `LAR-003` → AC-04, AC-05, AC-08, AC-09.
- Laravel curated profiles: `LAR-004` → AC-05, AC-06, AC-08, AC-09.
- Laravel matrix lifecycle and adapter: `LAR-005`, `LAR-006` → AC-01, AC-02, AC-07.
- Multi-app recipes/TanStack: `REC-001`, `REC-002` → AC-09, AC-10.
- Integrated verification/docs: `VER-001` → AC-12, AC-13 and regression coverage for all criteria.

### Acceptance-to-evidence matrix

- AC-01/02: top-level CLI fixture tests covering deterministic listing, support tier/prerequisites/supported actions, state, validated dispatch/status/resume, unsupported actions, and exit propagation.
- AC-03/05: characterization tests for existing JavaScript and Laravel direct commands/plans.
- AC-04/08/09: malicious/malformed fixture tests, duplicate/conflict/order/service/health tests, and synthetic canary assertions across argv/stdin/log/recovery/definitions.
- AC-06: offline profile-plan/support/compatibility table plus opt-in uniquely named bounded provisioning for five representative profiles.
- AC-07: offline Laravel matrix fixtures for prerequisite, timeout, resume, retry, independent continuation, unmanaged/partial targets, preserved logs, and summary status.
- AC-10: offline recipe DAG/atomic recovery/status/resume tests plus opt-in local two-app HTTPS/API/CORS connectivity smoke.
- AC-11: WordPress adapter list/dispatch fixtures and one existing-profile dry-run.
- AC-12: documentation review against commands and support tiers.
- AC-13: focused suites, `bash -n`, ShellCheck when installed, all adjacent bootstrap/matrix suites, and runtime smoke report.

### Revision history

- r1: initial draft for specialist review.
- r2: incorporated all required DSA, TDD, security, migration, operations, and compatibility findings; separated offline defaults from opt-in golden paths.
- r3: resolved plan review findings by splitting the oversized Laravel model/starter/capability and matrix/adapter work into six dependency-safe contracts; spec r2 unchanged.

<a id="phase-01-contract-and-adapters"></a>

## Phase 01: contract and adapters

- Children: `SHOW-001`, `SHOW-002`, `SHOW-003`

Establish the versioned adapter boundary and expose existing JavaScript and WordPress behavior without replacing their native implementations.

<a id="show-001"></a>

## SHOW-001: top-level showcase contract and orchestrator

### Goal and observable behavior

Create a versioned, ecosystem-neutral CLI at `scripts/showcase/showcase.sh` that deterministically lists tracked showcase definitions and dispatches validated `create`, `start`, `stop`, `verify`, `status`, and `resume` actions to ecosystem adapters. Listing reports ecosystem, definition, deterministic app names, support tier, prerequisites, adapter-declared supported actions, and local state. Unsupported actions fail clearly without emulation.

### Scope

- Define adapter protocol v1 and status vocabulary.
- Implement data-only adapter registry/discovery, lexical ordering, capability negotiation, validation, argument boundary, dispatch, and exit propagation.
- Add fixture tests and concise README.

### Non-goals

No Podman abstraction, generated apps, ecosystem profile changes, recipes, concurrency, or arbitrary adapter paths from user data.

### Entry inputs

Approved plan r3; existing script conventions; existing ignored `apps/` policy.

### Expected files

`scripts/showcase/showcase.sh`, `scripts/showcase/adapters/`, `scripts/showcase/showcase.test.sh`, `scripts/showcase/README.md`.

### Acceptance criteria

- AC-01/02 listing and dispatch behavior pass fixtures.
- Registry entries are parsed as data; adapters resolve to canonical repository-contained non-symlink regular files under the fixed adapter root, are uniquely named, and protocol-version checked.
- State derives only from adapter-declared completion markers, not from definition authority in generated apps.
- `--` cleanly separates top-level arguments from adapter-owned arguments.
- Invalid IDs/actions and duplicate adapters fail before mutation.

### Verification

Plan-time TDD: first Red tests cover sorted list, supported-action reporting, unknown adapter/action, duplicate registry ID, controls/traversal/symlink rejection, exact argv/exit propagation, and state reporting. Use arrays only, never evaluated strings. Then `bash -n`, focused fixtures, ShellCheck if installed.

### Rollback

Delete the additive `scripts/showcase/` surface; native ecosystem commands remain untouched.

<a id="show-002"></a>

## SHOW-002: JavaScript compatibility adapter

### Goal and observable behavior

Expose every existing JavaScript framework/profile through adapter protocol v1 while preserving direct bootstrap/matrix commands, deterministic names, skip/retry behavior, logs, and exit codes.

### Scope

Implement the JavaScript adapter mapping list/create/start/stop/status/verify to existing scripts; add additive compatibility assertions.

### Non-goals

No JavaScript profile changes, bootstrap rewrite, bulk concurrency, or generated app commits.

### Expected files

`scripts/showcase/adapters/javascript.sh`, focused showcase fixtures, additive `scripts/javascript/scaffold-matrix.test.sh` assertions.

### Acceptance criteria

- AC-02/03 pass.
- Direct and adapted plans/argv are behavior-equivalent for representative profiles.
- Existing app prefix and adapter passthrough remain supported.
- Adapter does not parse or source JavaScript profile contents itself.

### Verification

Red characterization for direct-vs-adapter list/create/start/stop, then focused and full JavaScript bootstrap/matrix suites.

### Rollback

Remove adapter only.

<a id="show-003"></a>

## SHOW-003: WordPress existing-profile adapter

### Goal and observable behavior

List and individually provision WordPress's existing `bare`, `clean`, `custom`, and `loaded` profiles through adapter protocol v1 without changing profile contents or native bootstrap behavior.

### Scope

Add WordPress adapter supported-action declaration and lifecycle mapping, deterministic showcase names, state/verify behavior, and fixture/dry-run tests. Treat existing `apps/showcase-wp` as unmanaged and never adopt or alter it.

### Non-goals

No new WordPress profiles, bulk provision-all default, profile parser changes, or restore automation through the showcase adapter.

### Expected files

`scripts/showcase/adapters/wordpress.sh`, showcase fixtures/tests, documentation.

### Acceptance criteria

- AC-11 passes.
- `list` is mutation-free; `create` delegates one selected profile.
- Restore and credential-bearing options are rejected at the showcase boundary and remain available only through native bootstrap; registry/profile text is never sourced or evaluated.
- Existing WordPress tests remain green.

### Verification

Red fixtures for list/name/unsupported restore and exact dry-run dispatch; WordPress bootstrap suite and one adapted profile dry-run.

### Rollback

Remove adapter only.

<a id="phase-02-laravel-profile-system"></a>

## Phase 02: Laravel profile system

- Children: `LAR-001`, `LAR-002`, `LAR-003`, `LAR-004`, `LAR-005`, `LAR-006`

Characterize and extend safe profile data, add official starters and reviewed capabilities as separate slices, then ship the curated catalog, matrix, and adapter.

<a id="lar-001"></a>

## LAR-001: Laravel profile parser and legacy model

### Goal and observable behavior

Characterize existing `bare`, `standard`, and `loaded` behavior, then extend Laravel's data-only profile parser/model for one allowlisted starter ID, repeatable capability IDs, support tier, compatibility probe ID, and required catalog service/health IDs without installing new starters or capabilities yet.

### Scope

Use indexed arrays for declaration order and associative maps for membership/conflicts; coalesce exact repeatables at first occurrence; reject duplicate singletons, controls, arbitrary commands, traversal, unknown metadata IDs, and declared conflicts before mutation. Preserve all current profile/package/database/migrate/seed/force/no-hosts plans and CLI output.

### Non-goals

No new starter execution, capability handlers, profiles, matrix, services, or generated apps.

### Expected files

`scripts/laravel/bootstrap.sh`, parser/model tests, legacy characterization fixtures, Laravel directive documentation.

### Specialist decisions incorporated

DSA arrays/maps/no inheritance; TDD characterization-first; security data-only parsing; migration preserves canonical profile files/names.

### Acceptance criteria

- AC-04 and parser portion of AC-05/08/09 pass.
- Legacy resolved plans are byte-stable except explicitly additive metadata lines approved by fixtures.
- New metadata resolves into a deterministic secret-safe plan but unknown handler IDs fail preflight.

### Verification

Red legacy characterization first; then valid model, controls/command text, duplicate singleton/repeatable, conflict, missing service/health/compatibility ID, stable order, and synthetic secret-canary fixtures. Run focused and full Laravel bootstrap tests, `bash -n`, ShellCheck if installed.

### Rollback

Revert parser/model while retaining untouched legacy profile files and generated apps.

<a id="lar-002"></a>

## LAR-002: official Laravel starter engine

### Goal and observable behavior

Implement non-interactive current-major scaffolding for starter IDs `none`, `livewire`, `react`, `vue`, and `svelte`, including frontend dependency/build steps only when selected, compatibility probes/resolved-version evidence, timeouts, and existing recovery/quarantine behavior.

### Scope

Research and pin/validate current Laravel installer flags during implementation; provide code-owned starter handlers; expose trusted upstream execution in plans; provide diagnostic no-scripts behavior where feasible; record resolved Laravel/starter/frontend versions in local logs/evidence.

### Non-goals

No capability packages, new profile catalog, matrix, recipes, or external auth variants.

### Expected files

Laravel bootstrap/starter handler, PHP image only if installer availability requires it, tests/docs.

### Specialist decisions incorporated

TDD starter Red order; security argv arrays/fixed handlers/upstream disclosure; compatibility probes before later mutation; operations bounded timeout/log/quarantine.

### Acceptance criteria

- Each starter produces an exact redacted dry-run plan and valid app completion markers.
- Unknown starter or incompatible current-major resolution fails before database/later capability mutation where feasible.
- Failure leaves no ambiguous target and produces preserved actionable logs/recovery state.
- Legacy `none` path remains equivalent.

### Verification

Offline fake-installer fixtures for every starter, incompatibility, timeout, scripts mode, frontend build selection, and failure stages; full Laravel suite; one opt-in uniquely named `none`/minimal real smoke with prerequisite and cleanup checks.

### Rollback

Revert starter handlers; parser remains additive and legacy `none` behavior remains usable.

<a id="lar-003"></a>

## LAR-003: reviewed Laravel capability handlers

### Goal and observable behavior

Implement code-owned plan/install/configure/verify handlers for `sanctum`, `horizon`, `reverb`, `telescope`, `pulse`, `pennant`, `scout-meilisearch`, `filament`, and `pest`, including exact conflicts, required catalog services/health probes, compatibility checks, deterministic ordering, and recovery integration.

### Scope

One registry maps fixed IDs to shell functions and metadata; profile text never supplies commands. Resolve all requested handlers/conflicts/services before mutation. Handler execution uses argv arrays, validated package constraints, stdin/files for secrets, bounded timeouts, local resolved-version evidence, and explicit post-install verification.

### Non-goals

No starter engine changes, profile catalog, external SaaS, billing, social auth, tenancy, Octane, or arbitrary third-party capability extension.

### Expected files

`scripts/laravel/lib/capabilities.sh` or equivalent; Laravel bootstrap integration; handler fixture tests/docs.

### Specialist decisions incorporated

Stable ordered de-duplication; fixed trust roots/handlers; service preflight and health; offline default tests; explicit upstream package trust.

### Acceptance criteria

- AC-08/09 pass for all handlers.
- Exact duplicates coalesce; conflicts/missing services/failed probes/incompatible packages fail before app capability mutation.
- Only selected services start; every handler verifies its installed/configured state.
- Synthetic secrets never enter argv/output/log/recovery/definitions.

### Verification

Red table-driven handler plan/order/conflict/service/health/compatibility tests; fake Composer/Artisan execution and postcondition failures; canary scan; full Laravel suite. No real multi-capability profile smoke until `LAR-004`.

### Rollback

Remove handlers/integration; parser and starter engine remain, legacy profiles remain usable.

<a id="lar-004"></a>

## LAR-004: curated Laravel profile catalog

### Goal and observable behavior

Add 13 explicit, non-inheriting profiles: `minimal`, `web`, `api`, `livewire`, `react`, `vue`, `svelte`, `queues`, `realtime`, `admin`, `search`, `observability`, and `testing`, while retaining canonical `bare`, `standard`, and `loaded` files/names.

### Profile definitions

- `minimal`: SQLite, starter none.
- `web`: MariaDB, Mailpit, starter none.
- `api`: MariaDB, Sanctum.
- `livewire`/`react`/`vue`/`svelte`: matching official starter, MariaDB, Mailpit.
- `queues`: MariaDB, Redis, Mailpit, Horizon.
- `realtime`: MariaDB, Redis, Reverb.
- `admin`: MariaDB, Mailpit, Filament.
- `search`: MariaDB, Meilisearch, Scout.
- `observability`: MariaDB, Redis, Telescope, Pulse.
- `testing`: SQLite, Pest and deterministic test execution.

Each declares support tier, compatibility probe, required catalog services/health probes, and exact description. Existing unmanaged `apps/showcase-laravel` is untouched.

### Non-goals

No external-account, billing/social, tenancy, Octane, inheritance, matrix, or recipes.

### Expected files

Laravel profile files, table-driven catalog tests, README.

### Acceptance criteria

AC-05/06/08/09 pass; all profiles list lexically and resolve exact secret-safe plans with no shipped conflict.

### Verification

Red offline exact-plan/support/compatibility table for every legacy/new profile. Opt-in uniquely named bounded real smokes for `minimal`, `api`, `react`, `queues`, `observability`, each with prerequisite checks, preserved failure log, HTTPS/Artisan/service postconditions, and cleanup instructions.

### Rollback

Remove new files only; legacy profiles and generated apps remain.

<a id="lar-005"></a>

## LAR-005: Laravel scaffold matrix

### Goal and observable behavior

Add native Laravel matrix commands for lexical list, create, create-all, start, stop, status, and verify over deterministic `apps/showcase-laravel-<profile>` targets.

### Scope

DNS-safe prefix override; `artisan`+`composer.json` completion markers; sequential resumable bulk creation; bounded retry/timeouts only for fresh scaffold acquisition; unique run IDs; disk/network/service prerequisites; preserved per-profile failure logs; independent continuation; concise nonzero failure summary. Never adopt/rename/overwrite/delete unmanaged `apps/showcase-laravel`.

### Non-goals

No top-level adapter, concurrency, default start-all, pruning, recipes, or profile internals.

### Expected files

`scripts/laravel/scaffold-matrix.sh`, matrix tests/docs.

### Acceptance criteria

AC-07 passes through native matrix commands; direct Laravel bootstrap semantics remain unchanged.

### Verification

Red offline fixtures for list/name/state, complete skip, incomplete/unmanaged target, prerequisite, timeout, bounded retry, continuation, quarantine/recovery, unique logs, and final summary; focused/full Laravel suites; opt-in one-profile create/status/verify/stop smoke and cleanup.

### Rollback

Remove matrix; bootstrap and apps remain.

<a id="lar-006"></a>

## LAR-006: Laravel showcase adapter

### Goal and observable behavior

Expose the native Laravel matrix through adapter protocol v1 with exact supported-action negotiation, validated dispatch, state, prerequisites, support tier, and exit propagation.

### Scope

Map top-level list/create/start/stop/status/verify to native matrix actions. `resume` is declared unsupported for single-app matrix entries; recipe resume remains recipe-owned. Add direct-versus-adapter parity fixtures.

### Non-goals

No matrix/profile/bootstrap changes, action emulation, or generated apps.

### Expected files

`scripts/showcase/adapters/laravel.sh`, focused adapter tests/docs.

### Acceptance criteria

AC-01/02 pass for Laravel; unsupported actions are reported, not emulated; exact argv/output/exit parity holds.

### Verification

Red protocol/version/action/argv/exit/list-state parity fixtures, then native Laravel matrix/bootstrap and top-level showcase suites.

### Rollback

Remove adapter only.

<a id="phase-03-recipes"></a>

## Phase 03: multi-app recipes

- Children: `REC-001`, `REC-002`

Add constrained multi-application composition after JavaScript and Laravel single-app adapters are stable.

<a id="rec-001"></a>

## REC-001: recoverable multi-app recipe engine

### Goal and observable behavior

Allow tracked data-only recipes to coordinate multiple adapter-owned applications with deterministic names, validated dependency order, explicit configuration links, and durable recovery state.

### Scope

Define recipe protocol v1 with allowlisted directives for app nodes, dependency edges, and named non-secret configuration outputs/inputs; validate unique IDs, acyclic graph, adapter/profile/action support, target collisions, and all inputs before mutation. Execute Kahn topological order with lexical node-ID tie-breaking. Persist a versioned restricted line-record recovery file recording pre-existing, creating, created, failed, and pending nodes plus recipe hash/prefix; replace it atomically through sibling temporary write and rename; never delete pre-existing targets.

### Non-goals

No arbitrary commands/templates, secret values in recipe files/manifests, concurrent creation, distributed transactions, or automatic rollback deletion.

### Expected files

`scripts/showcase/recipes/`, recipe parser/runner and tests, recovery documentation.

### Acceptance criteria

- AC-09 and recipe-engine portion of AC-10 pass.
- Validation rejects cycles, duplicate nodes/outputs, missing adapters/profiles, traversal, unsupported links, and target collisions before mutation.
- Stable Kahn topological ordering uses lexical node-ID tie-breaking, independent of declaration order.
- Failure stops dependent nodes, preserves independent/pre-existing nodes, quarantines partial adapter targets, and writes a secret-free actionable recovery record.
- `status` and `resume` validate recovery schema, recipe hash, prefix, paths, and definition identity before reporting or continuing; stale/corrupt records fail closed.

### Verification

Red parser/graph fixtures including shuffled declarations and lexical tie order; failure injection at every atomically persisted node transition; interrupted-write, resume/stale-definition/prefix/pre-existing-target tests; synthetic canary secret scan; then one offline two-app fixture integration.

### Rollback

Remove recipe runner/definitions; preserve all generated and quarantined trees for manual recovery.

<a id="rec-002"></a>

## REC-002: Laravel API plus TanStack Start recipe

### Goal and observable behavior

Ship the first multi-app recipe as separately routable `apps/showcase-laravel-tanstack-api` and `apps/showcase-laravel-tanstack-web`, using Laravel API/Sanctum and TanStack Start/Router/Query with coordinated local URLs.

### Scope

Add a current TanStack Start JavaScript definition through the existing JavaScript bootstrap model with support tier, compatibility probe, and resolved-version evidence; define recipe nodes/edge; configure frontend API URL and Laravel CORS/Sanctum stateful domains through reviewed allowlisted adapter-owned non-secret outputs; add start/stop/status/resume/verify and documentation.

### Non-goals

No monorepo, SSR authentication guarantee beyond documented Sanctum flow, external identity provider, deployment topology, or generic code templating.

### Expected files

TanStack JavaScript framework/profile definition and tests; one recipe definition; adapter-owned configuration handlers; recipe integration tests/docs.

### Acceptance criteria

- AC-10 passes.
- Both apps are ignored, independently routable, reproducibly named, and reported as one recipe.
- Browser/API origin configuration is explicit and contains no credential.
- Verification proves frontend HTTPS, API health/JSON response, allowed local CORS origin, and recovery after injected second-node failure.

### Verification

Red offline definition/edge/config fixture and hostile-output rejection; scaffold dry-run; injected failure/atomic recovery/resume. Then opt-in uniquely named bounded real creation and HTTPS/API smoke with prerequisites, preserved logs, and cleanup instructions.

### Rollback

Remove definition/config handlers; generated apps remain separable and recoverable.

<a id="phase-04-verification-and-docs"></a>

## Phase 04: verification and documentation

- Children: `VER-001`

Run the integrated compatibility/golden-path suite and document the reproducible showcase library.

<a id="ver-001"></a>

## VER-001: integrated golden paths and documentation

### Goal and observable behavior

Deliver one documented, verified showcase workflow covering top-level discovery, JavaScript compatibility, Laravel profiles/matrix, WordPress adapter, and the Laravel–TanStack recipe.

### Scope

Run default offline focused/adjacent tests, characterization, static shell checks, deterministic list snapshots, WordPress adapted dry-run, secret-canary scans, and failure/recovery exercises. Run explicitly opt-in, uniquely named, prerequisite-checked, bounded golden paths for five representative Laravel profiles, existing JavaScript runtime, and TanStack recipe HTTPS/API/service behavior; preserve failures and document cleanup. Update root/ecosystem/showcase documentation and examples.

### Non-goals

No new features, profile expansion, performance optimization, all-profile concurrent startup, or external SaaS validation.

### Expected files/outputs

Root README updates, `scripts/showcase/README.md`, ecosystem READMEs, test runner integration, durable verification report under this initiative.

### Acceptance criteria

- AC-12/13 and the complete acceptance-to-evidence matrix pass or explicitly record a blocking gap.
- Documentation distinguishes definitions, unmanaged existing apps, and matrix-generated apps; names support tiers, supported actions, prerequisites, trusted-upstream execution, recovery, and cleanup.
- Commands shown in documentation are exercised by tests or the runtime smoke report.
- Existing unrelated worktree changes are preserved.

### Verification

Default: all focused offline suites, characterization, `bash -n`, ShellCheck when installed, existing DevArch/bootstrap/matrix suites, Windows/Unix hosts helpers, and synthetic canary scan of argv/logs/definitions/recovery. Opt-in: bounded uniquely named runtime golden paths with disk/network/service probes, timeouts, preserved logs, and cleanup. Produce the complete acceptance-to-evidence report.

### Rollback

Documentation/test-runner changes revert independently; prior verified implementation contracts remain intact.

## Historical specialist findings and reviews

These are preserved planning rationale, not task statuses. The r3 task boundaries above supersede earlier IDs: old `LAR-001` split into `LAR-001`/`LAR-002`/`LAR-003`; old `LAR-002` became `LAR-004`; old `LAR-003` split into `LAR-005`/`LAR-006`. The r2 change request was resolved by r3 approval. Old hash/path checks describe their original review, not a new verification run.

<a id="finding-dsa"></a>

## Historical finding: dsa

### Problem summary

The system needs deterministic registry discovery, ordered capability execution, duplicate/conflict checks, recipe dependency ordering, and persistent recovery state. These are semantic correctness requirements; high-scale optimization is not.

### Current implementation

JavaScript discovers profile files lexically and loops sequentially. Laravel stores selected features/packages in Bash arrays and has no capability graph. No cross-ecosystem registry or recipe graph exists.

### Workload / constraints

Expected scale is tens of adapters/profiles and fewer than ten nodes per recipe. Reads dominate; writes occur only during create/recovery transitions. Ordering and reproducibility matter more than latency. Facts are inferred from the current catalog and proposed scope.

### Recommendation

- Discover fixed repository adapter/profile paths lexically into indexed arrays.
- Maintain associative maps keyed by validated ID for uniqueness/membership/conflicts and indexed arrays for stable user-declared installation order.
- Coalesce exact repeatable capabilities/packages at first occurrence; reject duplicate singleton directives and declared conflicts.
- Represent recipe adjacency and indegree with associative maps/arrays; validate using Kahn topological sort with lexical node-ID tie-breaking. At expected scale a repeatedly sorted ready array is simplest and adequate.
- Forbid profile inheritance in v1; recipes alone form a DAG.
- Persist secret-free recovery records as a versioned, validated line format with restricted whitespace-free IDs/status/path fields. Write to a sibling temporary file, `fsync` where available, then atomic rename. Bind resume to recipe content hash and app-name prefix.

### Rejected alternatives

- General graph library/SQLite: unjustified dependency and migration cost.
- Recursive DFS in shell: harder cycle diagnostics and recursion state.
- Profile inheritance: creates a second graph and hidden order.
- JSON recovery requiring `jq`/runtime parser: avoid a new prerequisite for simple bounded state.
- Parallel queues: no measured need and worsens recovery semantics.

### Complexity impact

Discovery/sorting: `O(P log P)`. Duplicate/conflict checks: expected `O(P+C)`. Recipe validation: `O(V² + E)` with a repeatedly sorted small ready array; acceptable for `V < 10`. State lookup: expected `O(1)` associative access.

### Memory tradeoff

Arrays/maps duplicate only small validated IDs and paths; bounded negligible memory is preferable to reparsing and ambiguous ordering.

### Migration advice

Characterize current order before internal migration. Keep legacy profile names and direct entry points. Version recovery records from first release; reject stale definition hashes rather than guessing conversion.

### Validation plan

Table-driven duplicate/conflict/order fixtures; shuffled filesystem/declaration fixtures; cycle/missing-node tests; transition failure injection; interrupted-write fixture proving old-or-new recovery record, never truncation.

### Confidence

High for representation and algorithms given small bounded workloads and current shell implementation.

<a id="finding-tdd"></a>

## Historical finding: tdd

### Observable behavior and Red ordering

1. Characterize existing JavaScript and Laravel list/plan/dispatch/recovery behavior before refactoring.
2. `SHOW-001`: Red top-level sorted listing, invalid/duplicate/traversal adapter rejection, protocol mismatch, exact argv and exit propagation, unsupported action, and generated-state reporting.
3. `SHOW-002/003`: Red direct-versus-adapter parity and restricted WordPress action tests.
4. `LAR-001`: Red starter plan, controls/arbitrary directive rejection, singleton duplicate, capability duplicate/conflict, deterministic order, legacy equivalence, secret redaction, and partial-starter recovery.
5. `LAR-002`: Red table of exact resolved plans for all curated profiles before adding definitions; runtime tests remain opt-in.
6. `LAR-003`: Red list/name/state, resumable skip, incomplete target, bounded retry, continue-independent-profile, final nonzero summary, and quarantine cases.
7. `REC-001`: Red parser and DAG cases first, then recovery state transitions in order: pending → creating → created/failed; interruption at each persisted boundary; stale definition/prefix refusal.
8. `REC-002`: Red two-node definition/config edge, then injected frontend failure after backend creation, resume, and end-to-end CORS/API behavior.
9. `VER-001`: no new production behavior; assemble the already-green checks and verify documentation commands.

### Test levels

- Unit/fixture: parsing, validation, ordering, graph, state transitions, argv.
- Characterization: existing direct commands and legacy profiles.
- Integration: adapters, matrix orchestration, recovery/resume.
- End-to-end opt-in: selected real Laravel profiles and Laravel–TanStack HTTPS/API path.

### Required test seams

Use temporary repository roots, fake adapters/bootstrap executables, call logs, deterministic clock/hash injection for recovery files, and explicit failure-stage environment variables available only to tests. Never depend on real `apps/`, real hosts files, or external credentials in default tests.

### Risk-scaled verification

Every contract runs focused tests first and adjacent native suites second. Real network scaffolds are opt-in, bounded, sequential, uniquely named, and leave documented cleanup/recovery evidence. Capture runtime Red/Green/Refactor only during implementation.

<a id="finding-security"></a>

## Historical finding: security

### Trust boundaries

Tracked definitions are repository-controlled data, but may be changed by contributors. Adapters and capability handlers execute package managers and framework installers with network access and write access to `apps/`. Generated package lifecycle scripts are upstream executable code. Local dotenv, Composer/npm credentials, SSH agents, databases, and hosts elevation are sensitive boundaries.

### Required decisions

- Parse new registry/profile/recipe files as data; never `source`, `eval`, shell-expand, or execute definition fields.
- Resolve adapter and capability IDs only through repository-contained, non-symlink regular files under fixed roots; canonicalize paths and reject traversal, duplicates, controls, whitespace ambiguity, and unsupported protocol versions before mutation.
- Use argv arrays and `--` boundaries. Do not construct shell command strings.
- Capability handlers own exact package/Artisan/npm operations and configuration keys. Definitions may select IDs and validated constraints only.
- Treat Composer/npm install scripts as explicit trusted-upstream execution; display package/starter sources in plans, provide a no-scripts diagnostic mode where feasible, and never claim profiles are sandboxed.
- Pass secrets through stdin or permission-restricted files where supported; redact plans/logs; recovery records contain IDs/status/paths/hashes only. Add secret-pattern scans using synthetic canaries.
- Reject credential-bearing Git/HTTPS URLs and external-account profiles from unattended golden paths.
- Prevent symlink writes/target escape beneath `apps/`; preserve pre-existing targets and refuse dangerous cleanup paths.
- Recipe cross-app outputs must be allowlisted non-secret values (origin/URL/name). Secret sharing requires an adapter-owned secure channel and is deferred.

### Verification

Malicious fixture corpus for command substitutions, separators, newlines/controls, traversal, symlinks, duplicate IDs, hostile constraints/URLs, stale recovery hashes, and forged adapter output. Assert no canary appears in argv, stdout/stderr, run logs, recovery state, or tracked definitions.

### Residual risk

Upstream installer/package code executes with the shared PHP/Node container's access. Document this and pin/validate sources; eliminating it is outside scope.

<a id="finding-migration-operations-compatibility"></a>

## Historical finding: migration-operations-compatibility

### Findings

1. **High — compatibility:** r1 does not define a version/support policy for current upstream starter kits and third-party capabilities. Add profile metadata for support tier and compatibility probe, record resolved Laravel/starter/package versions in generated local evidence, and fail a profile as unsupported before later capability mutation where resolution proves incompatible.
2. **High — operations:** real smoke expectations for five profiles are costly and mutable. Separate default fixture/characterization verification from opt-in network/runtime golden paths; require unique run IDs, disk/network prerequisite checks, bounded timeouts, cleanup instructions, and preserved failure logs.
3. **High — migration:** legacy `bare`, `standard`, and `loaded` must remain canonical public names rather than silently becoming aliases to new names. Characterize exact resolved plans, retain files/CLI output, and implement new profiles additively.
4. **Medium — operations:** recipe recovery state needs atomic-write and stale-definition behavior plus an explicit operator `status`/`resume`; r1 mentions resume but the top-level action contract omits it.
5. **Medium — compatibility:** adapter protocol must distinguish unsupported lifecycle actions without pretending all ecosystems have identical create/start semantics. Require capability negotiation in list/status and preserve native entry points.
6. **Medium — migration:** existing ignored `apps/showcase-laravel` and `apps/showcase-wp` do not match new deterministic matrix names. Treat them as unmanaged local apps; do not rename, adopt, overwrite, or delete them automatically.
7. **Medium — operations:** service-rich capabilities must declare required catalog service IDs and health probes. Preflight missing services before app/database mutation.
8. **Low — scope:** catalog external-account ideas (`billing`, WorkOS/social, multi-tenancy, Octane) only as documented future candidates, not executable definitions in this initiative.

### Required r2 incorporation

Update spec acceptance, adapter protocol, Laravel model/catalog, recipe engine, verification contract, and migration section with all findings. No blocker remains after those changes and hash/link reconciliation.

<a id="review-r2"></a>

## Historical plan review r2

### Findings

1. **High — contract granularity/executability:** `LAR-001` combines legacy characterization, parser/model migration, official installer/starter execution, frontend installation/build, nine capability handlers, compatibility probes, service health, security, and recovery. A developer cannot complete/review this as one smallest honest vertical slice. Split into parser/legacy model, official starter engine, and reviewed capability handlers. Update TDD ordering and downstream dependencies.
2. **Medium — contract granularity:** `LAR-003` combines the native Laravel matrix and top-level adapter. Split matrix lifecycle from protocol adapter so matrix compatibility can be proven before integration.
3. **Medium — verification traceability:** `LAR-002` names five runtime smokes, but those depend on starter and capability implementations currently hidden inside oversized `LAR-001`. Reorder curated profile definitions after all required engine/handler contracts, then matrix and adapter.

### Checks

- Canonical paths/topic/revision links: pass.
- Manifest specialist completeness: pass.
- Contract hash integrity, IDs, dependency references: pass.
- Graph first-ready uniqueness (`SHOW-001`): pass.
- Specialist incorporation: pass.
- Outcome/acceptance traceability: pass, subject to granular remap.
- Developer executability: fail for findings 1–3.

### Verification implications

No implementation may start from r2. Preserve the exact offline/opt-in, security-canary, compatibility, and recovery verification decisions while splitting contracts.

### Blockers and next action

Blocking finding: oversized Laravel contracts. Create plan revision r3 with narrow Laravel model, starter engine, capability handler, catalog, matrix, and adapter contracts; synchronize hashes/graph/traceability; then re-review r3. Spec r2 does not require revision.

<a id="review-r3"></a>

## Historical plan review r3

### Findings

No blocking findings.

The r2 granularity findings are resolved: Laravel parser/legacy behavior, official starters, capability handlers, profile definitions, native matrix, and top-level adapter are now six narrow dependency-safe contracts. Recipe and verification contracts consume those stable boundaries.

### Review checks

- Canonical topic, immutable revision paths, predecessor/review links: pass.
- Active spec outcomes/constraints/non-goals and open blockers: pass; no blocker.
- Specialist applicability/completion and incorporation: pass.
- Schema-v2 contract metadata, parent IDs, hashes, dependencies: pass.
- Graph acyclicity and first-ready uniqueness: pass; `SHOW-001` only.
- Architecture boundaries: pass; top-level dispatch remains thin, native ecosystem bootstraps own behavior, and recipe composition is separate.
- Security/data-only/fixed-handler/secret boundary: pass.
- Migration/rollback and unmanaged-app preservation: pass.
- TDD Red ordering and offline versus opt-in verification feasibility: pass.
- Outcome-to-contract and acceptance-to-evidence traceability: pass.
- Developer executability and contract granularity: pass.

### Residual risks

Upstream Laravel starter kits, Composer packages, and TanStack scaffolding may drift. Approved contracts require compatibility probes, resolved-version evidence, bounded failure, and fail-closed behavior; no planning blocker remains.

### Verification implications

Implementation must capture runtime Red/Green/Refactor evidence per contract and must not treat opt-in network golden paths as default offline tests. Completion requires the exact contract-specific verification and implementation review.

### Decision and next action

Approve exact spec r2 and plan r3 with zero blocking findings. First dependency- and gate-satisfied implementation contract: `SHOW-001`.

## Source provenance and inventory

Source commit: `f1e6a4a3200d8d902ead5b7ac0db0cbdcdd17ce5`. The immutable Git blob IDs below identify all 57 retired source files; `git show <blob-id>` retrieves exact pre-migration bytes without a legacy directory or manifest. SHA-256 values identify source bytes only; they do not claim approval of this migrated document. Paths are source-relative historical labels, not live filesystem links. Current design sources are spec r2, plan/index r3, r3 task/phase bodies, four findings, and two reviews. Earlier revisions remain history only.

| Source-relative label | Git blob | SHA-256 |
| --- | --- | --- |
| `findings/2026-09-12_1637-dsa-decision.md` | `4f033412b13779d94d1a8d9140ec33c7226c7636` | `c7aef4693a3945d0681b202a8fceb09947c29a7c4c300ff820477906fe7a7036` |
| `findings/2026-09-12_1637-migration-operations-compatibility-finding.md` | `d1503ed450e041df725068badef5b65ebe2eb773` | `9f6d55186f324da5042284749db630a7380879494ceb26bba90483b9f9ad1755` |
| `findings/2026-09-12_1637-security-finding.md` | `31303dc9018ba8f32c8bcdcaf8d48cfc13dbb721` | `d4f40f028ef8d2d2927c9ad9c00ad6f69652ca936da31255bc3d4851fd79eab2` |
| `findings/2026-09-12_1637-tdd-plan.md` | `0f00a63102e694eb998899337d9835d7b31ae301` | `e7806278a8c054644f2f3a998a42743a026ab96eb0a2b0aa51021bf00024bd3b` |
| `plans/2026-09-12_1637-plan-index-r1.md` | `6fc02f9cc6b3c9310339fd2d914b68fbf4842219` | `a2904fe4dd7cef344fa39f67fc701ba48f2d4d08dac0bc1fd77a3a38bb9c470d` |
| `plans/2026-09-12_1637-plan-index-r2.md` | `4ab264cca7ff59c221633269c59cb44f651e1a71` | `f0e19bf2b9f67d45faefdba3e5d0a4cb715549d7a94137b1dbe6c2177ce9848b` |
| `plans/2026-09-12_1637-plan-index-r3.md` | `077e2fe6f7b872a2c1a0567c34c82dafd5dde998` | `7393e413295038f73cd4d3c496e0f228c1378da60612fd1ed4621227f404940a` |
| `plans/revisions/r1/contracts.json` | `b0cf04b112627da125731b594c0ee58f5827b977` | `8ab14ca5f01ce310a945e945e5653a28f232a9ed6ad9574fc4ee15106997b52e` |
| `plans/revisions/r1/phases/phase-01-contract-and-adapters/phase-01-contract-and-adapters.md` | `a2954f4a8374e94e0df229dee937327b4ba11126` | `ba3f9e17d791b545d07754e8ee595171e54719e0a5ccf99081af636118070acc` |
| `plans/revisions/r1/phases/phase-01-contract-and-adapters/show-001-top-level-showcase-contract.md` | `df7e0ede313ed9b5c9a8addfae86087334ee2808` | `b9ac6df61ea8caa62b44f8ccf6009c93040b25bdb5b3dfd9f5f3bcbd06159437` |
| `plans/revisions/r1/phases/phase-01-contract-and-adapters/show-002-javascript-compatibility-adapter.md` | `0d4cabd40125d287ab32605a3dc708150894437c` | `07e2d4ca5da3a151519082655ee9007446929fda90809c4fa224b38cbd7bc689` |
| `plans/revisions/r1/phases/phase-01-contract-and-adapters/show-003-wordpress-existing-profile-adapter.md` | `94d3a1e78a8be40fb024103ae9b26027440cd8a3` | `495af4f35c5133449a8a0b8b1a8925d2733dcae2041ad6f5d09dc8f3daeb8d88` |
| `plans/revisions/r1/phases/phase-02-laravel-profile-system/lar-001-safe-starter-capability-model.md` | `2ca9180720fa69b0f7e0d92d9352e0fca0a6b175` | `9d287b8127dce0638432b61be72018e25a673aa1dfd606accf0d11a7eb80a934` |
| `plans/revisions/r1/phases/phase-02-laravel-profile-system/lar-002-curated-profile-catalog.md` | `e9f1fbc30a27080b2991eb4db0daca79c00e6465` | `9aa0e7e7533ab1d3a8c48c840e627c1471ed4d00c814fd3ffb1a23c4b9843b28` |
| `plans/revisions/r1/phases/phase-02-laravel-profile-system/lar-003-laravel-matrix-adapter.md` | `e3f78e4de1fb49e3860b83e16965abb512b09661` | `5c0dd3a4b17b56d5d4f632ab2d612b546ab44b136788bd4482b7a6a8af7e90f3` |
| `plans/revisions/r1/phases/phase-02-laravel-profile-system/phase-02-laravel-profile-system.md` | `7f2f76018c4fa407776c8a7ecd165ab350c8c4f2` | `37a5cb6bb6c09844d55a68087c0fe009351aa5a7a039dbe93ff1af932cdfbfa5` |
| `plans/revisions/r1/phases/phase-03-recipes/phase-03-recipes.md` | `c439c38fddfbf4c05ed8b5ffe92e5fbd6c31e9f4` | `f4b6681468651f92b2b09c6a489b30c8d85ca7233b4692e5f4ac84122aec15d4` |
| `plans/revisions/r1/phases/phase-03-recipes/rec-001-recoverable-recipe-engine.md` | `d8714ab5bd68690e240c382162c5d200fe8cde89` | `cf487e49f6af2499e973d2c5bc87344cee13ac286307956e8e81ddc7bd2662c3` |
| `plans/revisions/r1/phases/phase-03-recipes/rec-002-laravel-tanstack-recipe.md` | `ad96f319fbecf404837bf302fa7054ba3203fad0` | `c17080cc896d9d0f31251942d938a6ad9d411e191d79d3f06fba9cdaffdd67ab` |
| `plans/revisions/r1/phases/phase-04-verification-and-docs/phase-04-verification-and-docs.md` | `94a30ef2c5b0efdf729f0c8ad2cb8dbe48ae00d9` | `2e1385aecb7b4372b7e42dc75a1228730d7901fbeccb6b5c9d649d53c8be96d2` |
| `plans/revisions/r1/phases/phase-04-verification-and-docs/ver-001-golden-paths-documentation.md` | `63a67e6cd1ac49170183d8661da8b90aea8430d6` | `18908f2a04558fd0353f899e1096aec3a450e1f81e3b780a9fa4245faa718700` |
| `plans/revisions/r2/contracts.json` | `b85435361759c93738ba1bcc6b0bf5eee54bed93` | `a2566673fbb51c85053165fd362a79350c1b373451991b767d20e5f1c727af1c` |
| `plans/revisions/r2/phases/phase-01-contract-and-adapters/phase-01-contract-and-adapters.md` | `3750e842768c16f135f93aca09a48344b4c5ed0f` | `3056c4a955053c3ad7033bbb90242f930a151616ba0ded880a48b788bdc97746` |
| `plans/revisions/r2/phases/phase-01-contract-and-adapters/show-001-top-level-showcase-contract.md` | `bc19a90ecba585925b1bb25a5783a02b7d1b184b` | `7c08c786a2f88db40bef876aa5abb84216ddd8c3f03855664c17721b1093d3d7` |
| `plans/revisions/r2/phases/phase-01-contract-and-adapters/show-002-javascript-compatibility-adapter.md` | `ee06fb16c34546680beab105c0c02aa8812150c8` | `9d009d956828b51e5131535100678d6f9a351dabe7f718b54af1e35edb5b6ec6` |
| `plans/revisions/r2/phases/phase-01-contract-and-adapters/show-003-wordpress-existing-profile-adapter.md` | `438e178993e269c9e8645e9ce2ab9f773dcc057b` | `7c9ae4dbd2cd98ddfd29f4453e2413df7d90d7406f7da2ba329a53dcd66568f6` |
| `plans/revisions/r2/phases/phase-02-laravel-profile-system/lar-001-safe-starter-capability-model.md` | `ae2774ba469be0a50438c7487c5e1805cdaa9b31` | `ff8d8f2c411d031e1d6aceceab8ebad2e74273d02a54b11512f7aa9719b14d87` |
| `plans/revisions/r2/phases/phase-02-laravel-profile-system/lar-002-curated-profile-catalog.md` | `6854232d92fd34aa45269642e57e67aa8ed4a65e` | `76007a7f7fdefe6f0d82e3b132e0fb3c41ad2da8dffaa9ad740a10b40585b180` |
| `plans/revisions/r2/phases/phase-02-laravel-profile-system/lar-003-laravel-matrix-adapter.md` | `341eabc0b25a0ae297758c68172de0b556f9aa91` | `99f100f348c4c0c2a41b5cf00f7f57aee46b936aec53f29278237c6b523915a9` |
| `plans/revisions/r2/phases/phase-02-laravel-profile-system/phase-02-laravel-profile-system.md` | `d92b126e0b0f32744d4baac8c8944b746c2bae73` | `e8a0ff833cb23e3b4b9fa81835c5307386562eb72bcd61fe38de3b5537fc0461` |
| `plans/revisions/r2/phases/phase-03-recipes/phase-03-recipes.md` | `a55508c54dbae3997718e5e77760ca1b9b0c4c5e` | `542e3b3c7da7366b16c38ac46cfd2cf881afbd9b220a714bd07a56e15fd8c729` |
| `plans/revisions/r2/phases/phase-03-recipes/rec-001-recoverable-recipe-engine.md` | `a3ac1ac3f4c5c13541461d937786caf9bdab53b0` | `dff366bbfadbcbb48db7e0995521229f343a5be44740dc5ecefdcc1d94dde1fb` |
| `plans/revisions/r2/phases/phase-03-recipes/rec-002-laravel-tanstack-recipe.md` | `e7666515f175b8ec1d6a2a9e52388c81d388a921` | `53c897b71cf196525a17c203d028e0a46d622e84770a348debb167edf1af95e1` |
| `plans/revisions/r2/phases/phase-04-verification-and-docs/phase-04-verification-and-docs.md` | `0f5841c47a95382f4e91d2cb05fb6b6b4a469059` | `868bb279043d2128679d1fad93d78ae27adc8af865a0dbfe890f4ff2e855d67e` |
| `plans/revisions/r2/phases/phase-04-verification-and-docs/ver-001-golden-paths-documentation.md` | `504ce6d4f6a5888e0883ce2d1e03c5b96d61fdc4` | `a1f1e14840580970f5ca075833235a48bcc890c523fe915a32130e98005fc25e` |
| `plans/revisions/r3/contracts.json` | `688261324da94c7c40ca7173252512537222d28c` | `7a586286748e7f48a831dc630c9c1805236f1177cbbc94a6873df869fca5ec54` |
| `plans/revisions/r3/phases/phase-01-contract-and-adapters/phase-01-contract-and-adapters.md` | `727d663cfbeb87e3d20c67b38f90f17731beefd3` | `d49a8a78c83d968ee29922ba5474d4b0b4ea7cbe9a1f7a91fd05658eb1628589` |
| `plans/revisions/r3/phases/phase-01-contract-and-adapters/show-001-top-level-showcase-contract.md` | `014e3da48ff4f2522cc3a38dee5ab2105e801500` | `bb4fce396c90962b0e5f75cbf1a0b349659284ee665dfdc57845397758d4eb71` |
| `plans/revisions/r3/phases/phase-01-contract-and-adapters/show-002-javascript-compatibility-adapter.md` | `9da740fa079b5b8c8441159641f8231850a06ebc` | `5823968c2a575d410d64f91b1ccedf1fde824f07fd65f2a3213285546ea63f51` |
| `plans/revisions/r3/phases/phase-01-contract-and-adapters/show-003-wordpress-existing-profile-adapter.md` | `404fd80e667c8a894abb8834544fcf5002e2ace1` | `e84d853f4c6d8db9f2d6581d29a62a29edee3ca16aa2965c9e698792264fc329` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-001-profile-parser-legacy-model.md` | `1874aee79784a56f7368d32579f5804a30bb7451` | `cb0d5bad4c8d3766027fddeed5e37bc929fc7106c89f8c0380d427c4c030709a` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-002-official-starter-engine.md` | `4d26f292a2ee2e7a496d6ab8253599b29e198e4f` | `9d652be4cb740399e02dd27ee67a6a4e7b506e49dc8355c3b959d298c6c7082a` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-003-reviewed-capability-handlers.md` | `140f3812c94f167704d26f3e2876da5a87605385` | `a8af915f4571b1c792dca79cd6f398f5d42fc0a2509ed4e671d27992909c77a6` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-004-curated-profile-catalog.md` | `25baf1a392986a75e480897762849100d33a6284` | `61943cdeb9decff178f56e13bbf09d05e575e6d4f8a6b6666d1eda02cbeff50b` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-005-laravel-matrix.md` | `5b46fc683ad147685c579d988d826e8f40f059d0` | `d13df1da74ffb63e1c478eff02f4018b9d65d509c587ced2a2fcb6bf6ec4fa30` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/lar-006-laravel-showcase-adapter.md` | `81d2c632bfe2d5c44b45690a27aca9b1bff2a6cc` | `a3ea823ec989f142c59361f9de1c7361fb83855fa7b7a770d013cf53153e1c85` |
| `plans/revisions/r3/phases/phase-02-laravel-profile-system/phase-02-laravel-profile-system.md` | `88ed2405758db10bd2b0f665323738e1a7c7f16d` | `eafa212e99ed88fd20f1625821980d295b9683ffc059e84089834a9b61d44d3e` |
| `plans/revisions/r3/phases/phase-03-recipes/phase-03-recipes.md` | `ed569a795d9b8c185c6e80d3a70d9234deafa520` | `1e874fa28b78e4e40d7f77c923cb8d04680c51f01d19ce7e10b611c4e16ad44c` |
| `plans/revisions/r3/phases/phase-03-recipes/rec-001-recoverable-recipe-engine.md` | `183b4fcf897caae2b5d580787242eb742c083ce4` | `b30c1a7d689974eaababb7a3a2222a775cb5be3a5c148cf8cc8caa21d5cca2dc` |
| `plans/revisions/r3/phases/phase-03-recipes/rec-002-laravel-tanstack-recipe.md` | `2338a57cfa7f953f70fe76093cecd5e3d3c1e9e7` | `9111e95f3ba336ef3ee3c0916d05ea00ac0c2c880f5570e92d96ea00977c8d2a` |
| `plans/revisions/r3/phases/phase-04-verification-and-docs/phase-04-verification-and-docs.md` | `180a10386090f92b89bf369738958a99424a47f8` | `e9e70935e695a932515508b9325a84f94d35b484c8ee8ecd401d948e1eb270c1` |
| `plans/revisions/r3/phases/phase-04-verification-and-docs/ver-001-golden-paths-documentation.md` | `9d4fe7a11d8ed0820ac47d4efca789f051b1a991` | `17ab6d237d917e795da7dcc7f0ec713378bc6b2fb71b44c1499dc1f8e64eeab1` |
| `reports/2026-09-12_1637-plan-review-r2.md` | `f2c36178edd7a924692fef36e5efeeae2a42fd37` | `da56d3d2bf75c1ba050574bec33c79607be3ff9ba700027d0e0a994d3a57a872` |
| `reports/2026-09-12_1637-plan-review-r3.md` | `a6b77920b0c11f026fd21aca6fda1048a223f32c` | `44b29a7f08f5ceb97dd7ddc0422a20ea0cbd14853d632fc5520f58f0015f215b` |
| `specs/2026-09-12_1637-initiative-spec-r1.md` | `a8ab49e8f5cb60d6d7d291abfcc1a8543a9a472f` | `8d212d0bd6e0f2770f79080659bfd8b3fc7a81a7bc57b38172487185f3b4ec92` |
| `specs/2026-09-12_1637-initiative-spec-r2.md` | `2e63a9cedc73c84eb75dadea74be6986901484e3` | `0460b38808694ee9ccb7c95b45ad7a824b42be63be82375f53e45e23a5ec3cb7` |
| `specs/manifest.json` | `00c6c6f71a3aa0939b6d009a9b57b94dfb66f34a` | `02f0fbc82d525c888d317a94c624e5ce680811c56a6037a52294350bc77f78b0` |
