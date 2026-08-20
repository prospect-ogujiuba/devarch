# phase-01-verification

Created: 2026-08-20T19:07:40.523Z
Purpose: Durable verification evidence for Phase 1 package ownership and extension contracts.

# Verification evidence: Maker package ownership contracts

Timestamp: 2026-08-20 19:07 UTC
Scope: `.model-artifacts/todo/maker-core-workspaces/phases/01-package-contracts [COMPLETE].md` (assigned before completion without the marker); canonical Playground MakerStarter, MakerBlocks, and MakerMaker repositories.

## Compile/typecheck

- `php -l functions.php` and `node --check tests/theme.test.mjs` in MakerStarter
  - Result: PASS, exit 0.
  - Evidence: PHP reported no syntax errors; Node syntax check completed without output.
- `npm run build` and `php -l makerblocks.php` in MakerBlocks
  - Result: PASS, exit 0.
  - Evidence: Webpack 5.109.2 compiled successfully; PHP reported no syntax errors. Build produced no additional worktree drift.
- `composer lint` in MakerMaker
  - Result: PASS, exit 0.
  - Evidence: Application, test, and entrypoint PHP files passed lint.

## Run

- In-memory Node/PHP harness temporarily removed `build/blocks-manifest.php`, loaded `makerblocks.php`, and invoked `makerblocks_build_notice()`; the harness restored the manifest in `finally`.
  - Result: PASS, exit 0.
  - Evidence: Notice instructed installation of a packaged release or `npm ci && npm run build`; manifest restoration confirmed.
- Focused boot/dependency contract paths:
  - MakerStarter: `node --test --test-name-pattern='declares the replaceable|boots and registers' tests/theme.test.mjs`
  - MakerBlocks: `node --test tests/contract.test.mjs`
  - MakerMaker: `php tests/contract.php`
  - Result: PASS, all exit 0.
  - Evidence: MakerStarter 2/2; MakerBlocks 3/3; MakerMaker admin and WP-CLI dependency contract passed.

## Test

- `npm test` in MakerStarter
  - Result: PASS, exit 0; 6 tests, 0 failures.
- `npm test` in MakerBlocks
  - Result: PASS, exit 0; 19 tests, 0 failures.
- `composer test` in MakerMaker
  - Result: PASS, exit 0; 11 existing tests with 0 failures plus the package boundary/dependency contract.

## Manual contract evidence

- Reviewed each `CORE-BOUNDARY.md` and README against `git ls-files --cached --others --exclude-standard` and the current top-level tree.
  - Result: PASS. MakerStarter 23 files, MakerBlocks 96 files, and MakerMaker 37 files are covered by blanket core ownership; each sibling workspace and normal customization set is named.
- Seeded clean temporary Git repositories from the current package worktrees, created representative sibling `verify-site-theme`, `verify-site-blocks`, and `verify-site-app` packages, then removed them.
  - Result: PASS. All three core repositories remained clean after both creation and removal. Scenario used `Template: makerstarter`, project block namespace `verify-site/example`, and a sibling app model.
- `git diff --check` in all three canonical package repositories
  - Result: PASS, exit 0. Verification left only the intended Phase 1 changes and no temporary/build drift.

## Gaps

- No live browser WordPress admin or real WP-CLI session with TypeRocket Pro v6 was used. Admin/WP-CLI dependency behavior is covered by isolated PHP contract harnesses; live integration remains environment-level release verification.
- MakerMaker's existing broken-template negative-path test emits one expected `file_put_contents` warning while the suite exits 0. Warning cleanup is outside this slice.

## Outcome

PASS. Phase 1 acceptance and verification criteria have evidence. The noted live-environment gap does not block this package-contract slice.
