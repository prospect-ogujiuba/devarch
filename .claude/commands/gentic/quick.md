---
name: gentic:quick
description: Execute a small task with gentic safety rails, deviation handling, and atomic commits — without full plan-patch-verify ceremony
argument-hint: "<task description>"
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
Execute a small, well-defined task with gentic safety rails but without the full plan → patch → verify ceremony.

Purpose: Not every change needs a 5-step workflow. Bug fixes, config tweaks, small refactors, and documentation updates can use a streamlined path that still enforces read-before-edit, minimal diffs, deviation rules, atomic commits, and verification. Quick mode is the escape hatch for tasks too small for `/gentic:plan` but too important to do without guardrails.

Output: Completed task with change report, commit hashes, and verification results.
</objective>

<execution_context>
This command does NOT spawn a single agent. Instead, it orchestrates a compressed workflow:

1. **Analyze** (inline) — Read relevant files, understand the task
2. **Execute** — Spawn the **surgeon** agent for edits (with deviation rules and per-task commits)
3. **Verify** — Spawn the **verifier** agent to run checks
</execution_context>

<context>
Task: $ARGUMENTS
</context>

<process>

## 1. Analyze (inline — no separate agent)

Read the files relevant to the task. Understand:
- What needs to change
- Which files are affected (should be ≤ 5 files for quick mode)
- What risk level this represents

**Quick mode gate:** If the task affects more than 5 files or involves critical-risk changes (auth, deployment, infrastructure), recommend using `/gentic:plan` instead.

## 2. Execute with Deviation Rules

Spawn the **surgeon** agent with the task description and the files identified in step 1.

**Deviation rules apply:**

| Rule | Trigger | Action |
|------|---------|--------|
| **Rule 1** | Bug found | Auto-fix, track |
| **Rule 2** | Missing critical (validation, security) | Auto-add, track |
| **Rule 3** | Blocking issue (missing dep, broken import) | Auto-fix, track |
| **Rule 4** | Architectural change needed | STOP, ask user |

**Per-task atomic commit:**
- Stage files individually (never `git add .`)
- Commit with type prefix: `{type}: {concise description}`
- Record commit hash

**Quick mode constraints:**
- Max 5 files modified
- Use Edit for existing files, Write only for new files
- Read every file before editing
- Track changes for the report

## 3. Self-Check

After edits, verify claims:
- Check created files exist
- Check commits exist
- If self-check fails, document what's missing

## 4. Verify

Spawn the **verifier** agent to run all available project checks.

**Quick mode verification:**
- All checks must run
- Report results even if some fail
- Don't fix failures — report them

## 5. Report

Produce a compact change + verification report.

</process>

<anti_patterns>
- Don't use quick mode for large features — redirect to /gentic:plan
- Don't skip the analysis step — read before editing
- Don't skip verification — always run checks after changes
- Don't exceed 5 files — that's the quick mode boundary
- Don't make unrelated changes — scope discipline still applies
- Don't suppress verification failures — report them honestly
- Don't skip the self-check — verify claims before reporting
- Don't use `git add .` — stage files individually
</anti_patterns>

<output_format>
After completing the task, write a markdown report to `.gentic/reports/quick-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the task description: lowercase, hyphenated, 3-5 key words, max 40 chars, drop articles/prepositions. Example: "Fix typo in login error message" → `fix-typo-login-error`.

```markdown
# Quick Task Report

**Task:** Fix typo in error message for login endpoint
**Status:** COMPLETE
**Verification:** PASS

## Commits

| Commit | Type | Description |
|--------|------|-------------|
| abc1234 | fix | Fix typo in login error message |

## Changes

| File | Operation | Risk |
|------|-----------|------|
| src/api/controllers/auth.ts | edit | low |

## Deviations

None — task executed as described.

## Verification

| Check | Status |
|-------|--------|
| npm test | pass |
| npm run lint | pass |

## Self-Check: PASSED

## Notes

Single-line edit to fix "authenication" → "authentication" in error message.
```

Also output the JSON report inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Task analyzed — relevant files identified
- [ ] Quick mode gate checked (≤ 5 files, not critical risk)
- [ ] Every file read before editing
- [ ] Minimal diffs applied (Edit, not Write for existing files)
- [ ] Deviation rules applied (1-3 auto-fix, 4 ask)
- [ ] Per-task atomic commit with type prefix
- [ ] Self-check passed (files exist, commits exist)
- [ ] Verification ran all available checks
- [ ] Change report produced with commits, files, deviations
- [ ] Verification results included in report
- [ ] Markdown report written to .gentic/reports/quick-<ts>-<slug>.md
</success_criteria>
</output>
