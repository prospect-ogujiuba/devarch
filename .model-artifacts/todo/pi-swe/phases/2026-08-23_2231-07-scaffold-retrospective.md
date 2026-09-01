# 07-scaffold-retrospective

Created: 2026-08-23T22:31:15.331Z
Purpose: Capture repetitive post-scaffold work only after implementation evidence exists

# Phase 7 — Scaffold retrospective and automation proposals

Created: 2026-08-23
Purpose: Use observed implementation repetition—not guesses—to improve commands, generators, and templates after the application work is complete.

## Goal

Produce a concrete scaffold-friction report and implement only safe, reusable generator improvements with separate verification.

## Scope

- During phases 1–6, log repeated manual edits by file type, intent, count, and whether application-specific or generally reusable.
- Compare generated resource output with final Employee, Department, AssetAssignment, AuditEvent, and recurring-work resources.
- Candidate patterns to measure: actor columns/FKs, soft delete, WordPress user selects, policies/capabilities, audit hooks, tabbed form layout, relationship list filters, migration indexes, factory/test contracts, admin parent menu registration, immutable resources, import command skeletons, and read-only system panels.
- Propose new flags/commands or template updates only where repetition occurs across at least two resources and semantics are stable.
- Keep MakerMaker core changes in its own repository/scope; do not mix application-specific organization fields into generic templates.

## Outputs

- `.model-artifacts/findings/makerdesk-scaffold-friction/...` report with command transcript, repeated edit matrix, time/risk impact, and recommendation priority.
- Proposed command/API design with backward compatibility and fixture updates.
- If approved/in scope, isolated MakerMaker changes with contract tests and docs; otherwise actionable follow-up todos.

## Acceptance criteria

- Findings distinguish generator gaps from deliberate domain customization.
- Every recommendation cites repeated implementation evidence.
- Generated output remains deterministic, safe on existing files, and backward compatible by default.
- Template/command changes have fixtures and contract tests.
- MakerMaker and application commits/scopes remain separable.

## Verification

- Run MakerMaker contract suite and generate into temporary fixtures.
- Compare expected file inventory and content for old/default and new flags.
- Re-run one representative MakerDesk resource scaffold in a disposable location.

## Non-goals

Do not modify generators speculatively before phases 1–6 reveal actual repetition.
