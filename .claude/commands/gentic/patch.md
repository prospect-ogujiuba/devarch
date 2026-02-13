---
name: gentic:patch
description: Apply changes with explicit diff discipline — minimal edits, per-task commits, deviation handling, structured reporting
argument-hint: "[plan or instructions]"
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
Apply changes following a plan with minimal diff discipline and per-task atomic commits.

Purpose: Execute the Architect's plan step by step. Every edit is surgical — smallest possible old_string/new_string. Every task gets its own commit. Deviation rules handle unplanned issues automatically. Self-check verifies claims before reporting.

Output: Structured completion report with commits, files, deviations, and self-check result.
</objective>

<execution_context>
Spawn the **surgeon** agent via Task tool to perform the edits. The surgeon has write access (Read, Write, Edit, Glob, Grep, Bash) and produces a completion report.
</execution_context>

<context>
Plan or instructions: $ARGUMENTS
</context>

<process>

## 1. Load Plan

Parse the plan from prompt context. Identify:
- **Tasks** — Ordered list with files, action, verify, done criteria
- **Must-haves** — Truths, artifacts, key links to satisfy
- **Risk levels** — Which tasks need extra care

Execute tasks in dependency order. Never skip a task.

## 2. Execute Each Task

For each plan task:

1. **Read the target file** — Always read before editing. Understand current state.
2. **Apply the edit** — Use Edit with the smallest possible old_string/new_string.
   - For existing files: always Edit, never Write
   - For new files: Write is acceptable
3. **Verify the edit** — Re-read the file to confirm the change applied correctly.
4. **Check done criteria** — Verify the task's acceptance criteria are met.
5. **Commit the task** — Stage files individually (never `git add .`), commit with type prefix.

**Edit precision rules:**
- Include just enough context in `old_string` to be unique
- Never include unchanged lines in `new_string` unless they're between changed lines
- If multiple edits needed in one file, apply them from bottom to top

## 3. Handle Deviations

While executing, apply deviation rules for unplanned issues:

| Rule | Trigger | Action |
|------|---------|--------|
| **Rule 1** | Bug (broken behavior, type errors) | Auto-fix, track |
| **Rule 2** | Missing critical (validation, error handling, security) | Auto-add, track |
| **Rule 3** | Blocking issue (missing dep, broken import) | Auto-fix, track |
| **Rule 4** | Architectural change (new DB table, schema, breaking API) | STOP, return to user |

Rules 1-3: Fix automatically, document as deviation.
Rule 4: Stop execution, return checkpoint for user decision.

## 4. Per-Task Commit Protocol

After each task's done criteria are met:

```bash
git add src/specific/file.ts    # Stage individually
git commit -m "{type}: {concise description}"
```

Types: feat | fix | test | refactor | chore

Record commit hash for the report.

## 5. Self-Check

After all tasks, verify claims:
- Check created files exist
- Check commits exist in git log
- If self-check fails, document what's missing

## 6. Produce Completion Report

```markdown
## PATCH COMPLETE

**Tasks:** {completed}/{total}
**Deviations:** {count}

### Commits
| Task | Commit | Type | Description |
|------|--------|------|-------------|

### Files Modified
| File | Operation | Lines +/- |
|------|-----------|-----------|

### Deviations
{list or "None — plan executed exactly as written."}

### Self-Check: PASSED
```

</process>

<anti_patterns>
- Don't use Write on existing files — always Edit
- Don't edit a file you haven't read in this session
- Don't change code outside the plan's scope (except deviation Rules 1-3)
- Don't add "improvements" the plan didn't ask for
- Don't combine multiple task edits into one commit — one commit per task
- Don't use `git add .` or `git add -A` — stage files individually
- Don't skip the self-check
</anti_patterns>

<output_format>
The surgeon agent produces the completion report inline. No separate file is written — the report is the command output.
</output_format>

<success_criteria>
- [ ] All plan tasks executed in dependency order
- [ ] Every file read before editing
- [ ] Edit (not Write) used for all existing files
- [ ] Minimal diffs — smallest possible old_string/new_string
- [ ] Per-task atomic commits with proper type prefix
- [ ] Deviation rules applied correctly (1-3 auto-fix, 4 ask)
- [ ] All deviations documented
- [ ] Self-check passed (files exist, commits exist)
- [ ] Structured completion report produced
- [ ] No changes outside plan scope (except documented deviations)
</success_criteria>
</output>
