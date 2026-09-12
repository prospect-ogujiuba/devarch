# LAR-001: safe Laravel starter/capability model

- Status: planned
- Plan revision: 2
- Spec: r2
- Dependencies: `SHOW-001`

## Goal and observable behavior

Extend Laravel profiles as strictly parsed data supporting one starter (`none`, `livewire`, `react`, `vue`, `svelte`), reviewed capability IDs, existing service features, and Composer packages. Resolve capabilities through code-owned handlers with deterministic plan/install/configure/verify steps and conflicts before mutation.

## Scope

- Characterize legacy profiles and CLI behavior.
- Add allowlisted `starter` and `capability` directives, support-tier/compatibility/service metadata, uniqueness/conflict validation, stable ordered de-duplication, and capability registry.
- Use indexed arrays for declared order and associative maps for membership/conflicts; exact repeats coalesce at first occurrence and singleton duplicates fail.
- Add current Laravel installer/starter-kit execution through the shared PHP image/runtime, frontend dependency/build handling, resolved-version evidence, compatibility probes, bounded timeouts, and recovery integration.
- Initial capabilities: `sanctum`, `horizon`, `reverb`, `telescope`, `pulse`, `pennant`, `scout-meilisearch`, `filament`, `pest`.
- Preserve package/service options and secret-safe plans.

## Non-goals

No arbitrary hooks, profile inheritance/includes, external SaaS credentials, recipes, matrix, Octane, billing, social login, or multi-tenancy.

## Expected files

`scripts/laravel/bootstrap.sh`, `scripts/laravel/lib/capabilities.sh` or equivalent code-owned registry, PHP image only if installer availability requires it, Laravel tests/docs.

## Acceptance criteria

- AC-04/05/08/09 pass.
- Unknown/duplicate/conflicting directives and capability IDs, missing declared catalog services, and failed compatibility probes fail before recovery marker/database/app mutation where dependency resolution permits.
- Ordered arrays preserve declared install order; keyed membership maps coalesce exact duplicates and detect singleton/conflict violations.
- Official starter selection is non-interactive and current-major compatible; frontend install/build occurs only for selected starters. Plans disclose trusted upstream package execution and a diagnostic no-scripts mode where feasible.
- Legacy `bare`, `standard`, `loaded`, package CLI/files, database, migrate/seed, force, and no-hosts behavior remains equivalent.

## Verification

TDD Red order: legacy characterization; valid starter plan; malformed/control/arbitrary command rejection; duplicate singleton/conflict; stable de-duplication/order; missing service/failed health probe; compatibility failure; starter failure quarantine; synthetic secret-canary redaction. Run focused offline parser/bootstrap suites and one opt-in, uniquely named, bounded real minimal starter smoke with prerequisite/cleanup checks.

## Migration/rollback

Keep legacy directives accepted. New handlers are additive and capability calls occur only when requested. Revert parser/handlers without changing generated apps; recovery paths preserve partial targets.
