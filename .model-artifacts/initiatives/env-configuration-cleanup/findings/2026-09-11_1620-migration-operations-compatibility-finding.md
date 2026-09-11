# Migration, Operations, and Compatibility Finding

- Specialists: migration, operations, compatibility
- Status: complete
- Assessed spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1614-initiative-spec-r1.md` (`sha256:16eb04ced8ab44ab1865b2d14f7987fda81644d45d103d0ee047fd3bb006b1e3`)
- Assessed plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1614-plan-index-r1.md` (`sha256:fb6a13277a1f59005a11d28b41f38c84107af91a0dbcea9cf4a8430506b27b48`)

## Decision

All three specialties are required. The staged order is safe; one env-file override contract needs clarification.

## Findings to incorporate

1. **High — Define explicit env-file selection semantics.** Introduce `DEVARCH_ENV_FILE` as a host control variable for both bootstraps: unset means optional `<repo>/.env`; a nonempty explicit path must exist and be a regular file or fail before mutation. Tests use a temporary fixture or `/dev/null`. The control variable is not read from dotenv and does not belong in `.env.example`.
2. **High — Preserve initialized MariaDB continuity.** Template cleanup must not imply rotating `MARIADB_ROOT_PASSWORD`. Guidance must tell operators to retain the password used when the persistent volume was initialized, or intentionally recreate/migrate that volume outside this initiative.
3. **Medium — Preserve key precedence.** File assignments continue to replace inherited values for recognized keys. Existing WordPress `WP_*` over `ADMIN_*` fallback and Laravel `LARAVEL_DB_ROOT_PASSWORD` over `MARIADB_ROOT_PASSWORD` resolution remain unchanged after loading.
4. **Medium — Make migration reversible and manual.** Never edit ignored `.env`. Categorize removed template entries as obsolete in this repository, allow them to remain harmlessly ignored, and provide a copy/compare/remove workflow.
5. **Medium — Keep Compose boundaries explicit.** Service-directory Compose remains self-contained. Bootstraps continue passing only `MARIADB_ROOT_PASSWORD` to MariaDB Compose; host-only admin/GitHub values must not move to container `environment:` sections.
6. **Low — Retain aliases for at least this initiative.** Prefer `WP_ADMIN_*` in the refreshed template, but continue supporting/documenting `ADMIN_*` to avoid breaking existing local files.

## Residual risk

Undocumented external wrappers may have relied on `set -a` exporting arbitrary `.env` keys. That behavior is unsafe and intentionally not preserved; release notes must call it out as a security tightening.
