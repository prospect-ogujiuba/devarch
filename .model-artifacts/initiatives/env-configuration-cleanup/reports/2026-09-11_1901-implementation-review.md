# implementation-review

Created: 2026-09-11T19:01:33.926Z
Purpose: Durable implementation-review decision for ENV-003 r2.

# Implementation review: ENV-003 template and configuration guidance

- Mode: implementation-review
- Decision: request changes
- Timestamp: 2026-09-11 15:01 EDT
- Topic: `env-configuration-cleanup`
- Contract: `ENV-003`, plan revision 2, `sha256:c1aa17714d5f71b482a0c21ad64173ecf593db9bf85301dbcad492ab0d650704`
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1625-initiative-spec-r2.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1625-plan-index-r2.md`
- Plan review: `.model-artifacts/initiatives/env-configuration-cleanup/reports/2026-09-11_1632-plan-review.md`
- Verification: `.model-artifacts/initiatives/env-configuration-cleanup/reports/2026-09-11_1857-env-003-verification.md` (`sha256:4f3a0fdaf6cfe0dd01d0c927af737024b2d4292dd1fc329236c1da8a8bb8d619`)
- Incorporated findings: TDD, security, and migration/operations/compatibility findings dated 2026-09-11 16:20.

## Contract gate

Manifest approval, active spec/plan r2, canonical contract index, ENV-002 completion, and manifest `activeContract` agree. Active spec, plan, phase, ENV-003 contract, and verification hashes are current. The implementation is confined to the six expected ENV-003 paths; no Compose file changed. No dependency, blocker, revision, or material-scope conflict exists.

## Findings

- **Medium — `.env.example` overrides the documented AIOWM fallback with a non-working repository URL.**
  - Affected path: `.env.example` (`AIOWM_GIT_URL=git@github.com:github-user/all-in-one-wp-migration.git`).
  - Impact: after copying the template, an operator who correctly replaces only `GITHUB_USER` still has an explicit `AIOWM_GIT_URL` pointing at `github-user`. `scripts/wordpress/bootstrap.sh` therefore cannot derive `git@github.com:$GITHUB_USER/all-in-one-wp-migration.git` as documented, and restore can target the placeholder repository.
  - Violated behavior: ENV-003 goal that operators can copy a minimal accurate template; compatibility requirement that `GITHUB_USER`/`AIOWM_GIT_URL` retain documented behavior; documentation in `scripts/wordpress/README.md` stating the AIOWM URL defaults from `GITHUB_USER`.
  - Required action: make `AIOWM_GIT_URL` blank or omit it so the established `GITHUB_USER` fallback remains effective. Prefer blank if the template is intended to expose every supported override.

No other correctness, security, migration, rollback, operations, or scope findings were identified. The allowlist subset check, blank password fields, `ADMIN_*` compatibility guidance, `DEVARCH_ENV_FILE` boundary, obsolete-key categories, Compose distinction, and MariaDB volume warning otherwise match the contract.

## Verification implications

The current verification map is structurally complete and all commands passed, but it does not detect the template's explicit AIOWM override of the documented fallback. After the one-line template correction, rerun the focused template contract, DevArch suite, WordPress regression, and `git diff --check`, then issue a superseding ENV-003 verification artifact. The Laravel suites need rerun only if the correction touches shared test behavior; otherwise their current evidence remains applicable.

## Open blockers and residual risks

- Blocking finding: the Medium template/default mismatch above.
- Residual risk after correction: service-local legacy defaults remain intentionally outside this contract; container-runtime execution remains outside the approved host-only verification scope.

## Next action

Correct `.env.example` within exact contract `ENV-003` r2, refresh affected verification evidence, and repeat implementation review. No plan revision is required because the change is narrow and directly required by the approved contract.
