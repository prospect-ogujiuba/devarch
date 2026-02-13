---
name: gentic:build
description: Full development workflow — pathfinder impact analysis → architect plan → surgeon patches → verifier checks → reviewer critique
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
Execute a complete development workflow for a task, chaining 5 agents in sequence.

Purpose: For non-trivial tasks that deserve the full ceremony. Combines blast radius prediction, architectural planning, surgical implementation, verification, and code review into a single command. Each stage gates the next — if verification fails, the workflow stops before review.

Output: Completed task with ship/no-ship verdict.
</objective>

<execution_context>
This is a compound command that chains agents sequentially:

1. **pathfinder** — Predict blast radius of the planned changes
2. **architect** — Design the implementation plan with blast radius awareness
3. **surgeon** — Execute the plan with minimal diffs and atomic commits
4. **verifier** — Run all available checks
5. **reviewer** — Critique the diffs and produce ship verdict

Each agent is spawned via Task tool. Results flow forward: pathfinder → architect uses impact data, architect → surgeon uses plan, etc.
</execution_context>

<context>
Task: $ARGUMENTS
</context>

<process>

## 1. Impact Analysis (pathfinder)

Spawn the **pathfinder** agent to predict which files will be affected by the task.

Use the blast radius to inform the architect about:
- High-confidence impact files that need review
- Test files that must pass
- Fragile files that need extra care

## 2. Plan (architect)

Spawn the **architect** agent with:
- The task description
- Blast radius results from step 1
- Instructions to account for impacted files in the plan

The architect produces a structured plan with tasks, files, and verification criteria.

## 3. Execute (surgeon)

Spawn the **surgeon** agent with:
- The architect's plan
- Deviation rules (auto-fix bugs, ask for arch changes)
- Atomic commit instructions

The surgeon implements each task, committing atomically.

## 4. Verify (verifier)

Spawn the **verifier** agent to:
- Run all project checks (tests, lints, type checks)
- Goal-backward verification against the plan
- Stub detection

**Gate:** If verification fails, stop here. Report NO-SHIP with failures.

## 5. Review (reviewer)

Spawn the **reviewer** agent to:
- Analyze all diffs from the surgeon's commits
- Check for security, performance, correctness issues
- Produce ship/no-ship verdict

## 6. Final Report

Combine all stage results into a build report.

</process>

<anti_patterns>
- Don't skip impact analysis — it informs the plan
- Don't proceed to review if verification fails
- Don't auto-fix reviewer findings — report them
- Don't commit during review — only surgeon commits
</anti_patterns>

<output_format>
Write a markdown report to `.gentic/reports/build-<timestamp>-<slug>.md`. Derive `<slug>` from the task: lowercase, hyphenated, 3-5 key words, max 40 chars.

```markdown
# Build Report

**Task:** <description>
**Verdict:** SHIP | NO-SHIP
**Stages:** 5/5 completed

## Impact Analysis
- Files predicted: N
- High confidence: N

## Plan
- Tasks: N
- Files: N

## Execution
- Commits: N
- Deviations: N

## Verification
**Result:** PASS | FAIL

## Review
**Verdict:** approve | request-changes
**Findings:** N (critical: N, high: N, medium: N, low: N)

## Final Verdict: SHIP
```
</output_format>

<success_criteria>
- [ ] Pathfinder blast radius completed
- [ ] Architect plan created with blast radius awareness
- [ ] Surgeon executed plan with atomic commits
- [ ] Verifier ran all checks
- [ ] Reviewer analyzed diffs (if verification passed)
- [ ] Build report written to .gentic/reports/
- [ ] Verdict correctly computed
</success_criteria>
