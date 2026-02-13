---
name: gentic:debug
description: Systematically diagnose a bug through hypothesis testing and structured investigation
argument-hint: "<description of the bug or symptom>"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Write
  - Task
---

<objective>
Systematically diagnose a bug through hypothesis-driven investigation.

Purpose: Stop guessing. Given a bug description, systematically reproduce, hypothesize, test, and narrow down to the exact root cause. The output is a diagnosis with file, line, evidence, and suggested fix — not a patch. The Surgeon handles the fix.

Output: Structured diagnosis report with root cause, evidence, and suggested fix.
</objective>

<execution_context>
Spawn the **debugger** agent via Task tool to perform the investigation. The debugger has Read + Bash access for diagnosis but cannot modify source files.
</execution_context>

<context>
Bug description: $ARGUMENTS
</context>

<process>

## 1. Understand the Symptom

Parse the bug report. Extract the observable wrong behavior, expected behavior, trigger conditions, and environment context.

## 2. Reproduce

Attempt to reproduce the bug using available test commands or direct execution. Capture actual output, error messages, and stack traces. Max 3 reproduction attempts with 30s timeout each.

## 3. Form Hypotheses

Based on reproduction results, form ranked hypotheses (max 5). Each hypothesis must be testable with a specific investigation action.

## 4. Test Hypotheses

Test each hypothesis in ranked order:

| Technique | When to use |
|-----------|------------|
| Stack trace analysis | Error with traceback |
| Git bisect | Regression — worked before |
| Input boundary testing | Validation/parsing bugs |
| Log/output inspection | Silent failures |
| State inspection | Unexpected state |
| Dependency check | Version/compat issues |
| Config comparison | Environment-specific bugs |

Stop when a hypothesis is CONFIRMED with evidence.

## 5. Produce Diagnosis

Document the root cause with:
- Exact file and line number
- Evidence that confirms the diagnosis
- Hypotheses that were tested and eliminated
- Suggested fix (description, not implementation)
- List of affected and related files

</process>

<anti_patterns>
- Don't fix the bug — diagnose only
- Don't guess without evidence — every conclusion needs proof
- Don't modify source files — use Read and Bash for investigation
- Don't run destructive commands (rm, drop, reset --hard)
- Don't pursue more than 5 hypotheses without re-evaluating
- Don't skip reproduction — always try to reproduce first
</anti_patterns>

<output_format>
After completing the investigation, write a markdown report to `.gentic/reports/debug-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the bug description: lowercase, hyphenated, 3-5 key words, max 40 chars, drop articles/prepositions. Example: "API returns 500 on user creation" → `api-500-user-creation`.

```markdown
# Debug Report

**Symptom:** API returns 500 on user creation
**Status:** ROOT CAUSE IDENTIFIED

## Root Cause

**File:** `src/api/controllers/user.ts:42`
**Category:** validation
**Description:** Missing null check on email field — throws TypeError when email is undefined

## Evidence

1. Stack trace points to user.ts:42
2. Reproduced with `POST /users` body: `{}`
3. Email field is optional in schema but required by controller

## Hypotheses Tested

| # | Hypothesis | Verdict | Reason |
|---|-----------|---------|--------|
| 1 | Database connection timeout | eliminated | DB logs show successful connections |
| 2 | Missing null check on email | **confirmed** | TypeError at line 42 when email undefined |

## Suggested Fix

Add null check for email before database insert, or make email required in request schema.

## Affected Files

- `src/api/controllers/user.ts` (root cause)
- `src/api/schemas/user.ts` (related — schema definition)
- `test/api/user.test.ts` (related — needs test case)
```

Also output the JSON diagnosis inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Bug symptom clearly stated
- [ ] Reproduction attempted (success or documented failure)
- [ ] At least 2 hypotheses formed and tested
- [ ] Root cause identified with file path and line number
- [ ] Evidence provided for the diagnosis
- [ ] Suggested fix described (not applied)
- [ ] Affected and related files listed
- [ ] Markdown report written to .gentic/reports/debug-<ts>-<slug>.md
- [ ] No source files modified during investigation
</success_criteria>
