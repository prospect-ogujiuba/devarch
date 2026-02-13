---
name: gentic:insights
description: Human-readable intelligence report from oracle — fragile files, co-evolution patterns, health metrics
argument-hint: ""
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
---

<objective>
Produce a human-readable intelligence report from mined session data.

Purpose: Show what the oracle has learned from past sessions — which files are fragile, which files always change together, and overall project health trends. This helps inform decisions before starting new work.

Output: Structured intelligence report with actionable insights.
</objective>

<execution_context>
Spawn the **oracle** agent via Task tool. It reads .gentic/intelligence/ and produces the report.
</execution_context>

<context>
Arguments: $ARGUMENTS
</context>

<process>

## 1. Check Intelligence Data

Verify `.gentic/intelligence/` exists and has data. If empty, report "No intelligence data yet — complete at least one session to start mining."

## 2. Analyze

Spawn the **oracle** agent to:
- Read fragile file patterns
- Analyze co-evolution clusters
- Summarize health metrics
- Produce recommendations

## 3. Report

Present the oracle's report directly.

</process>

<output_format>
Present the oracle agent's report. If no intelligence data exists, report that mining hasn't occurred yet and suggest running a session first.
</output_format>

<success_criteria>
- [ ] Intelligence directory checked
- [ ] Oracle agent spawned for analysis
- [ ] Report presented with fragile files, co-evolution, health
- [ ] Graceful message when no data exists
</success_criteria>
