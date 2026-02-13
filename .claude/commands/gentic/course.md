---
name: gentic:course
description: Manual course correction — break loops, reset approach, get navigator recommendations
argument-hint: "[--reset]"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
---

<objective>
Break out of loops and get actionable course correction recommendations.

Purpose: When the agent is stuck in a loop (re-reading files, retrying failed commands, oscillating edits), this command spawns the navigator agent to analyze the session state and recommend a new approach. Optionally resets navigator warnings with --reset.

Output: Session health report with prioritized recommendations.
</objective>

<execution_context>
Spawn the **navigator** agent via Task tool. Pass it the current session directory path.
</execution_context>

<context>
Arguments: $ARGUMENTS
- `--reset` — Clear navigator warnings and start fresh tracking
</context>

<process>

## 1. Check Session State

Read the current session's `navigator.json` to understand what patterns have been detected.

If `--reset` is set, acknowledge the reset but still report current state first.

## 2. Analyze

Spawn the **navigator** agent to analyze:
- Tool call patterns (what's being repeated)
- Edit history (oscillation detection)
- Failure patterns (what keeps failing)
- Scope drift (how many files touched vs. expected)

## 3. Report

Present the navigator's findings and recommendations.

If `--reset` is set, clear the warnings array in navigator.json after reporting.

</process>

<output_format>
Present the navigator agent's report directly. If session has no navigator state yet, report "No navigator data — session just started."
</output_format>

<success_criteria>
- [ ] Navigator state read from current session
- [ ] Navigator agent spawned for analysis
- [ ] Recommendations presented
- [ ] --reset flag honored (clear warnings after report)
</success_criteria>
