---
name: gentic:iterate
description: Fix cycle after NO-SHIP verdict — surgeon fixes → verifier re-checks
argument-hint: "<fix description or path to no-ship report>"
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - Task
---

<objective>
Execute a fix cycle after a NO-SHIP verdict from /gentic:ship or /gentic:build.

Purpose: When verification or review produces a NO-SHIP verdict, this command chains the surgeon and verifier to fix the identified issues and re-verify. Faster than re-running the full build workflow — skips impact analysis, planning, and review.

Output: Fix results with updated verification status.
</objective>

<execution_context>
This is a compound command that chains 2 agents:

1. **surgeon** — Fix the identified issues
2. **verifier** — Re-run checks to confirm fixes

Each agent is spawned via Task tool.
</execution_context>

<context>
Fix target: $ARGUMENTS

If the argument is a path to a report file (.gentic/reports/*), read it to extract the specific failures and findings that need fixing.
</context>

<process>

## 1. Identify Failures

Read the NO-SHIP report (if provided) or the most recent report in `.gentic/reports/`.

Extract:
- Failed verification checks
- Reviewer findings (critical/high severity)
- Stubs or missing artifacts
- Specific files and issues

## 2. Fix (surgeon)

Spawn the **surgeon** agent with:
- The specific issues to fix
- The files that need changes
- Deviation rules apply (same as /gentic:quick)
- Atomic commits per fix

## 3. Re-verify (verifier)

Spawn the **verifier** agent to:
- Re-run all checks that previously failed
- Run full check suite for regression
- Report updated status

## 4. Report

Write iteration report. If verification passes, report FIXED. If still failing, report STILL-FAILING with remaining issues.

</process>

<anti_patterns>
- Don't fix issues outside the NO-SHIP findings
- Don't skip re-verification after fixes
- Don't make architectural changes — those need /gentic:build
- Don't iterate more than 3 times without asking the user
</anti_patterns>

<output_format>
Write a markdown report to `.gentic/reports/iterate-<timestamp>-<slug>.md`.

```markdown
# Iteration Report

**Fixing:** <description>
**Status:** FIXED | STILL-FAILING
**Iteration:** N

## Issues Fixed

| Issue | File | Fix |
|-------|------|-----|
| Test failure in auth | src/auth.js | Fixed null check |

## Commits

| Commit | Description |
|--------|-------------|
| abc1234 | fix: null check in auth handler |

## Verification

| Check | Before | After |
|-------|--------|-------|
| npm test | FAIL | PASS |

## Remaining Issues

None — all checks pass.
```
</output_format>

<success_criteria>
- [ ] NO-SHIP report identified and parsed
- [ ] Specific failures extracted
- [ ] Surgeon fixed identified issues
- [ ] Atomic commits for each fix
- [ ] Verifier re-ran all checks
- [ ] Report written with before/after comparison
- [ ] Clear FIXED or STILL-FAILING status
</success_criteria>
