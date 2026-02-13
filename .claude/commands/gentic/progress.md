---
name: gentic:progress
description: Show current session progress — changes made, checks run, plan completion status
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Show the current session's progress at a glance.

Purpose: Visibility into what's been done. Read the active session's ledger data and any active plans to produce a progress dashboard showing changes made, verification status, plan completion, and what's next.

Output: Structured progress report with current session state.
</objective>

<execution_context>
Spawn the **scribe** agent via Task tool to read the current session data. The scribe has read-only access (Read, Glob, Grep) and produces the progress report.
</execution_context>

<context>
No arguments required. Reports on the current active session.
</context>

<process>

## 1. Load Current Session

Read `.gentic/.current-session` to identify the active session. Read the session's ledger files:
- `changes.jsonl` — What's been modified so far
- `verifications.jsonl` — What checks have been run

## 2. Load Active Plans

Glob for `.gentic/reports/plan-*.md` files. Find the most recent plan. Parse planned steps and their status.

## 3. Compute Progress

For each dimension, compute current status:

**Changes:**
- Total files modified
- Files created vs. edited
- Risk breakdown (how many high/medium/low risk changes)

**Verification:**
- Last verification run timestamp
- Overall status (pass/fail/not-run)
- Per-check breakdown

**Plan completion (if a plan exists):**
- Steps completed vs. total
- Current step (if identifiable)
- Blocked steps

**Session health:**
- Session duration so far
- Changes since last verification (drift indicator)

## 4. Produce Dashboard

Combine all dimensions into a compact progress report.

</process>

<anti_patterns>
- Don't include full file contents or diffs — just counts and summaries
- Don't modify any files
- Don't exceed 2KB for the progress report
- Don't report on previous sessions — only the current active one
- Don't fabricate data — only report what the ledger shows
</anti_patterns>

<output_format>
After computing progress, write a markdown report to `.gentic/reports/progress-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the current git branch name: lowercase, hyphenated, max 40 chars.

```markdown
# Session Progress

**Session:** 20250115T143022-12345-abc
**Started:** 2025-01-15T14:30:22Z
**Duration:** 32 minutes

## Changes

| Metric | Count |
|--------|-------|
| Files modified | 3 |
| Files created | 1 |
| Total edits | 7 |
| Risk: high | 1 |
| Risk: medium | 2 |
| Risk: low | 4 |

## Verification

**Last run:** 5 minutes ago
**Overall:** PASS

| Check | Status |
|-------|--------|
| npm test | pass |
| npm run lint | pass |

## Plan Completion

**Plan:** plan-20250115-143500-add-pagination-admin.md
**Progress:** 3 / 5 steps (60%)

| Step | Status |
|------|--------|
| 1. Add pagination params | completed |
| 2. Create pagination middleware | completed |
| 3. Update response serializer | completed |
| 4. Add admin endpoint pagination | pending |
| 5. Write integration tests | pending |

## Next Steps

- Execute plan step 4: Add admin endpoint pagination
- 2 changes since last verification — consider running /gentic:verify
```

Also output the JSON progress report inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Current session identified from .gentic/.current-session
- [ ] Changes ledger parsed with counts and risk breakdown
- [ ] Verification status reported (or "not yet run")
- [ ] Plan completion tracked if a plan exists
- [ ] Drift indicator (changes since last verification) computed
- [ ] Next steps suggested
- [ ] Output bounded to ~2KB
- [ ] Markdown report written to .gentic/reports/progress-<ts>-<slug>.md
- [ ] No files modified (read-only)
</success_criteria>
