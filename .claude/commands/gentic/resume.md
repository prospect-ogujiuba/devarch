---
name: gentic:resume
description: Resume work from a previous session — reconstruct context from ledger data
argument-hint: "[session-id]"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Resume work from a previous Claude session using ledger data.

Purpose: Context dies between sessions. This command reconstructs what happened in a previous session — what was changed, verified, planned, and left unfinished — so you can continue without re-analyzing the entire codebase.

Output: Structured handoff document with session summary, changes, verification status, unfinished work, and resume hint.
</objective>

<execution_context>
Spawn the **scribe** agent via Task tool to read session data and produce the handoff document. The scribe has read-only access (Read, Glob, Grep) and produces the resume pack.
</execution_context>

<context>
Optional session ID: $ARGUMENTS (if empty, uses most recent session)
</context>

<process>

## 1. Find Session

Locate the target session:
- If a session ID is provided, look for `.gentic/sessions/<id>/`
- If no ID, read `.gentic/.current-session` for the most recent session
- List available sessions if the target can't be found

## 2. Read Ledger Data

Parse the session's ledger files:
- `changes.jsonl` — File modifications with timestamps and operations
- `verifications.jsonl` — Test/lint/typecheck results
- Plan reports from `.gentic/reports/plan-*.md` in the session timeframe
- Ship reports from `.gentic/reports/ship-*.md` in the session timeframe

## 3. Reconstruct State

Categorize the session's work:

| Status | Meaning |
|--------|---------|
| `completed` | Planned, implemented, and verified |
| `in_progress` | Started but not verified |
| `blocked` | Attempted but could not proceed |
| `planned` | In a plan but not yet started |

Identify the last action taken before the session ended.

## 4. Produce Resume Pack

Generate a bounded (~3KB) resume document with:
- Session metadata (ID, duration, timestamps)
- One-sentence summary of what was accomplished
- Files changed (paths and operation types)
- Verification status (overall + per-check)
- Unfinished work with source references
- Blockers if any
- Clear resume hint: what to do next

</process>

<anti_patterns>
- Don't include full file contents — just paths and summaries
- Don't dump raw JSONL — synthesize into structured state
- Don't exceed 3KB for the resume document
- Don't fabricate session data — only report what's in the ledger
- Don't modify any files
- Don't guess what happened — only report what the ledger shows
</anti_patterns>

<output_format>
After producing the resume pack, write a markdown report to `.gentic/reports/resume-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the current git branch name: lowercase, hyphenated, max 40 chars.

```markdown
# Session Resume

**Previous session:** 20250115T143022-12345-abc
**Duration:** 45 minutes
**Summary:** Added pagination to user listing API. Tests pass. Review pending.

## Changes Made

| File | Operation |
|------|-----------|
| src/api/controllers/user.ts | edit |
| src/api/middleware/pagination.ts | create |

## Verification Status

**Overall:** PASS

| Check | Status |
|-------|--------|
| npm test | pass |
| npm run lint | pass |

## Unfinished Work

| Task | Status | Source |
|------|--------|--------|
| Add pagination to admin endpoint | planned | plan-20250115-143500-add-pagination-admin.md |
| Write integration tests | planned | plan-20250115-143500-add-pagination-admin.md |

## Resume Hint

Continue with remaining planned steps: admin endpoint pagination and integration tests. All prior work is verified and passing.
```

Also output the JSON resume pack inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Previous session identified and located
- [ ] Changes ledger read and parsed
- [ ] Verifications ledger read and parsed
- [ ] Work categorized (completed, in_progress, blocked, planned)
- [ ] Resume hint provides clear next steps
- [ ] Output bounded to ~3KB
- [ ] Markdown report written to .gentic/reports/resume-<ts>-<slug>.md
- [ ] No files modified (read-only analysis)
</success_criteria>
