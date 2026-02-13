---
name: gentic:ship
description: Ship changes only after verification passes, stubs cleared, and reviewer approves
argument-hint: "[--skip-review] [--force]"
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
---

<objective>
Ship changes through a composite verification, stub-check, and review gate.

Purpose: No changes ship without passing checks, clearing stubs, and passing review. This command orchestrates the full release gate — run all project checks, verify goal achievement, critique the diffs, and produce a ship/no-ship verdict. If anything fails, the verdict blocks shipping.

Output: Structured verdict combining verification results, goal status, and review findings.
</objective>

<execution_context>
Spawn the **verifier** agent first, then the **reviewer** agent via Task tool. Verification runs first; review runs second (if verification passes or --force is set).
</execution_context>

<context>
Arguments: $ARGUMENTS
- `--skip-review` — Skip the review step (verification only)
- `--force` — Proceed to review even if verification fails
</context>

<process>

## 1. Verify

Discover and run all available project checks (tests, lints, type checks). Each check runs independently.

If must-haves are available from a recent plan, also run goal-backward verification:
- 3-level artifact checks (exists, substantive, wired)
- Key link verification
- Stub detection

**Gate logic:**
- All checks must pass before proceeding to review
- No MISSING or STUB artifacts allowed (unless --force)
- If any check fails and `--force` is NOT set: stop and report `no-ship` with verification failures
- If any check fails and `--force` IS set: continue to review with a warning

## 2. Review

Read the files that were modified in the current session and analyze for:

| Category | What to check |
|----------|---------------|
| Security | SQL injection, XSS, command injection, hardcoded secrets, missing input validation, SSRF |
| Performance | N+1 queries, unbounded loops, missing pagination, sync operations that should be async |
| Correctness | Off-by-one errors, null/undefined not checked, missing error handling, race conditions, breaking API changes |
| Stubs | TODO/FIXME, empty implementations, placeholder handlers, unwired components |
| Style | Dead code, inconsistent naming, missing types on public interfaces, duplicate logic |

Skip this step if `--skip-review` is set.

## 3. Verification Loop (if gaps found)

If verification finds gaps (stubs, unwired artifacts, missing files):

1. **Identify gaps** — List what's missing or incomplete
2. **Report structured gaps** — PASS/GAPS_FOUND/HUMAN_NEEDED
3. **If GAPS_FOUND:** Report no-ship with specific gaps and remediation steps

This is NOT auto-fix — the ship command reports gaps, it doesn't fix them. Use `/gentic:patch` to fix, then re-run `/gentic:ship`.

## 4. Must-Haves Final Check

If the plan included must-haves, verify all are satisfied before issuing ship verdict:
- All truths: VERIFIED
- All artifacts: exist, substantive, wired
- All key links: WIRED

Any failed must-have → no-ship.

## 5. Verdict

Combine verification, goal status, and review results:

| Verification | Goal Status | Review | Verdict |
|-------------|-------------|--------|---------|
| pass | verified | approve | **SHIP** |
| pass | verified | comment (medium/low only) | **SHIP** (with notes) |
| pass | verified | request-changes (critical/high) | **NO-SHIP** |
| pass | gaps_found | any | **NO-SHIP** |
| fail | any | any | **NO-SHIP** |
| fail + --force | verified | approve/comment | **SHIP** (with warning) |

Gate output format: **SHIP** | **NO-SHIP** | **HUMAN_NEEDED**

</process>

<anti_patterns>
- Don't ship with failing checks unless --force is explicitly set
- Don't ship with unresolved stubs in critical paths
- Don't approve changes with unaddressed critical security findings
- Don't skip verification — it always runs
- Don't modify files during the ship process
- Don't produce review findings without actionable suggestions
</anti_patterns>

<output_format>
After completing verification and review, write a markdown report to `.gentic/reports/ship-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the current git branch name: lowercase, hyphenated, max 40 chars.

```markdown
# Ship Report

**Verdict:** SHIP
**Gate:** PASS | GAPS_FOUND | HUMAN_NEEDED
**Recommendation:** All checks pass, no stubs, review approves. Safe to ship.

## Verification

**Overall:** PASS
**Checks:** 3 run, 3 passed
**Goal Status:** VERIFIED
**Stubs:** 0

## Must-Haves

| Truth | Status |
|-------|--------|
| User can see paginated results | VERIFIED |

| Artifact | Status |
|----------|--------|
| src/api/pagination.ts | VERIFIED (L3) |

| Key Link | Status |
|----------|--------|
| controller → pagination | WIRED |

## Review

**Verdict:** approve

### Findings

| Severity | Category | Message | Suggestion |
|----------|----------|---------|------------|
| low | style | Consider extracting helper | Move shared logic to utils/format.ts |

**Summary:** Changes look good. Minor style suggestion, non-blocking.
```

Also output the JSON verdict inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Verification ran all available checks
- [ ] Goal-backward verification ran (if must-haves available)
- [ ] Stub detection completed
- [ ] Must-haves final check completed (if available)
- [ ] Review analyzed all changed files (unless --skip-review)
- [ ] Every review finding has severity, category, message, and suggestion
- [ ] Verdict correctly computed from verification + goal status + review matrix
- [ ] --force and --skip-review flags honored
- [ ] Markdown report written to .gentic/reports/ship-<ts>-<slug>.md
- [ ] Structured JSON verdict produced
</success_criteria>
</output>
