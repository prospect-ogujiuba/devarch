---
name: gentic:impact
description: Predict blast radius of changing a file — references, co-changes, test mapping
argument-hint: "<file-path>"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
---

<objective>
Predict the blast radius of changing a specific file.

Purpose: Before making changes, understand what other files will be affected. Combines static references, git history, oracle intelligence, and test mapping into a multi-signal impact prediction. Files flagged by multiple signals deserve careful review.

Output: Blast radius report with impact map, test files, and risk assessment.
</objective>

<execution_context>
Spawn the **pathfinder** agent via Task tool. Pass it the target file path.
</execution_context>

<context>
Target file: $ARGUMENTS
</context>

<process>

## 1. Validate Target

Confirm the target file exists. If not provided, ask which file to analyze.

## 2. Analyze

Spawn the **pathfinder** agent to:
- Find rg references to the target
- Look up git co-change history
- Check oracle co-evolution data
- Map test files
- Check target fragility

## 3. Report

Present the pathfinder's blast radius report with:
- Impact map (all affected files with confidence)
- Test files that should run
- Risk assessment
- Recommendations

</process>

<output_format>
Present the pathfinder agent's report directly. Write the report to `.gentic/reports/impact-<timestamp>-<slug>.md`.
</output_format>

<success_criteria>
- [ ] Target file validated
- [ ] Pathfinder agent spawned
- [ ] All 4 signal sources queried
- [ ] Impact map presented with confidence levels
- [ ] Test files identified
- [ ] Recommendations provided
- [ ] Report written to .gentic/reports/
</success_criteria>
