# Implementation review: ENV-002 Laravel shared dotenv migration

- Mode: implementation-review
- Decision: request changes
- Timestamp: 2026-09-11 14:16 EDT
- Topic: `env-configuration-cleanup`
- Contract: `ENV-002`, plan revision 2, `sha256:2dbae242ddfb0c34dc634ca5279a563f22d8f5a004cf88db791685980d176651`
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1625-initiative-spec-r2.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1625-plan-index-r2.md`
- Verification: `.model-artifacts/initiatives/env-configuration-cleanup/reports/2026-09-11_1318-env-002-verification.md` (`sha256:d02db9fda1be8b787d252e379a7fea0227d740b803c18dd3f134a34db5996b65`)
- Prior review: `.model-artifacts/initiatives/env-configuration-cleanup/reports/2026-09-11_1632-plan-review.md`
- Incorporated findings: TDD, security, and migration/operations/compatibility findings dated 2026-09-11 16:20.

## Contract gate

Manifest approval, active spec/plan r2, canonical contract index, ENV-001 completion, and manifest `activeContract` agree. Active spec, plan, and ENV-002 contract hashes are current. No dependency, blocker, revision, or scope conflict requires return to plan.

## Findings

- **Medium — The positive explicit-file boundary assertion is a false positive.**
  - Affected path: `scripts/laravel/bootstrap.test.sh:71-79`.
  - The fixture sets `LARAVEL_APP_URL=https://explicit-env.test`, invokes the bootstrap with app name `explicit-env`, and expects `https://explicit-env.test`. That is also the bootstrap's default URL, so the assertion passes even if the valid `DEVARCH_ENV_FILE` contents are ignored. The suite therefore does not prove the TDD finding's positive override seam or contract criterion 1 / AC3 file precedence at the Laravel process boundary.
  - The fixture also proves that the unrelated command substitution is not executed, but does not independently assert the unrelated key is absent from child environment/state as required by AC2's consumer evidence.
  - Action: use three distinct values for the selected-file URL, inherited URL, and app-derived default; assert only the selected-file URL is used. Add an integration assertion that the unrelated dotenv key is absent from child environment/state while retaining the sentinel check.

No production-code correctness defect or unrelated churn was found. The six-key allowlist is exact, the shared loader is sourced, invalid explicit paths fail early, private parser symbols are absent, generated application dotenv production code is unchanged, and the implementation remains within the three expected Laravel files.

## Verification implications

The current verification report's manual scenarios demonstrate the implementation behavior, and all recorded commands passed. However, the report overstates the automated Laravel boundary evidence for positive explicit-file selection and AC2 consumer isolation. After the test correction, rerun at minimum:

- `bash scripts/devarch/tests/dotenv_test.sh`
- `bash scripts/laravel/tests/bootstrap_test.sh`
- `bash scripts/laravel/bootstrap.test.sh`
- `bash -n scripts/laravel/bootstrap.sh scripts/laravel/tests/bootstrap_test.sh scripts/laravel/bootstrap.test.sh`
- `git diff --check`

Then replace or supersede the ENV-002 verification artifact with current evidence. The existing approval-eligible verification envelope must not be used after the diff changes.

## Open blockers and residual risk

- Blocking finding: one Medium in-contract regression-evidence defect.
- Residual risk after correction: the accepted bounded custom dotenv grammar remains covered by the shared loader suite.

## Next action

Return to `swe-implement` for the smallest in-contract test-only correction in `scripts/laravel/bootstrap.test.sh`, then rerun `swe-verify` and this implementation review. ENV-002 is not yet eligible for canonical completion.
