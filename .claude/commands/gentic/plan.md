---
name: gentic:plan
description: Produce a step-by-step implementation plan with risk assessment, must-haves, and verification plan
argument-hint: "<description of what to implement>"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Produce an implementation plan for a requested change.

Purpose: Think before cutting. Analyze the codebase, identify affected files, produce ordered tasks with dependencies and risk levels, derive must-haves, and define a verification plan. Plans are prompts — they teach the Surgeon exactly what to do.

Output: XML task format plan with must-haves, context budget awareness, and verification plan.
</objective>

<execution_context>
Spawn the **architect** agent via Task tool to perform the analysis and planning. The architect has read-only access (Read, Glob, Grep) and produces the plan.
</execution_context>

<context>
Request: $ARGUMENTS
</context>

<process>

## 1. Assess Discovery Level

Before planning, determine how much exploration is needed:

| Level | When | Action |
|-------|------|--------|
| 0 — Skip | Pure internal work, existing patterns | Proceed directly |
| 1 — Quick Scan | Known library, confirming syntax | Read files, grep patterns |
| 2 — Standard Research | Choosing between approaches | Deeper analysis |
| 3 — Deep Dive | Architectural decision | Full exploration first |

Announce the level before proceeding.

## 2. Analyze the Codebase

Read relevant files to understand current state. Use Glob and Grep to find related code, patterns, and conventions already in use.

## 3. Identify Affected Files

List every file that needs to change. For each file, note what kind of change is needed (new function, modified function, new file, config change).

## 4. Produce Tasks (XML Format)

Create ordered implementation tasks. Each task specifies:

```xml
<task type="auto">
  <name>Action-oriented title</name>
  <files>exact/path/to/file.ext</files>
  <action>Specific implementation instructions (enough for a different Claude to execute without questions)</action>
  <verify>How to prove this task is complete</verify>
  <done>Measurable acceptance criteria</done>
  <risk level="low|medium|high|critical">Risk explanation</risk>
  <needs>files or tasks this depends on</needs>
  <creates>files this produces</creates>
</task>
```

**Context budget:** 2-3 tasks per plan, targeting ~50% context usage. If more than 3 tasks are needed, split into multiple plans.

## 5. Derive Must-Haves (Goal-Backward)

Work backwards from the goal:

1. **Observable truths** — What must be TRUE from the user's perspective? (3-7 items)
2. **Required artifacts** — What files must EXIST with substantive content?
3. **Key links** — What must be CONNECTED for artifacts to function?

```xml
<must_haves>
  <truths>
    <truth>User can see paginated results</truth>
  </truths>
  <artifacts>
    <artifact path="src/api/pagination.ts" provides="Pagination helper"/>
  </artifacts>
  <key_links>
    <link from="controller.ts" to="pagination.ts" via="import and usage"/>
  </key_links>
</must_haves>
```

## 6. Build Verification Plan

Specify which checks to run after implementation, and which tasks each check covers. Include must-haves verification.

</process>

<anti_patterns>
- Don't produce vague tasks — "update the handler" is not actionable, "add pagination params to UserController.list() method" is
- Don't modify any files — this is planning only
- Don't skip risk assessment — every task needs a level and notes
- Don't assume the codebase structure — read files first
- Don't plan changes to files you haven't read
- Don't plan more than 3 tasks without splitting into separate plans
</anti_patterns>

<output_format>
After generating the plan, write a markdown report to `.gentic/reports/plan-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the request: lowercase, hyphenated, 3-5 key words, max 40 chars, drop articles/prepositions. Example: "Add pagination to the admin endpoint" → `add-pagination-admin-endpoint`.

```markdown
# Implementation Plan

**Request:** <what was requested>
**Discovery Level:** 0
**Overall Risk:** medium
**Files affected:** 3
**Context budget:** 2 tasks (~45%)

## Must-Haves

### Observable Truths
- User can see paginated results
- Response includes total count and page metadata

### Required Artifacts
- `src/api/middleware/pagination.ts` — Pagination helper
- `src/api/controllers/user.ts` — Modified with pagination params

### Key Links
- user.ts → pagination.ts (import and usage)

## Tasks

### Task 1: Add pagination middleware
- **File:** `src/api/middleware/pagination.ts`
- **Risk:** low — New file, no existing code affected
- **Action:** Create pagination helper...
- **Verify:** File exists, exports parsePagination
- **Done:** parsePagination returns correct offset/limit
- **Depends on:** none

### Task 2: Wire pagination into controller
- **File:** `src/api/controllers/user.ts`
- **Risk:** medium — Changes public API response shape
- **Action:** Import parsePagination, apply to list()...
- **Verify:** npm test passes
- **Done:** Response shape is {data, total, page, limit}
- **Depends on:** Task 1

## Verification Plan

| Check | Command | Covers Tasks |
|-------|---------|-------------|
| Type check | `npm run typecheck` | 1, 2 |
| Unit tests | `npm test` | 1, 2 |
| Must-haves | Code inspection | truths, artifacts, key_links |
```

Also output the XML plan inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Discovery level assessed and announced
- [ ] Request parsed and understood
- [ ] Relevant codebase files read (not assumed)
- [ ] All affected files identified
- [ ] Tasks use XML format with name, files, action, verify, done
- [ ] 2-3 tasks per plan (~50% context budget)
- [ ] Dependencies explicit (needs/creates per task)
- [ ] Risk level assigned to every task
- [ ] Must-haves derived (truths, artifacts, key_links)
- [ ] Verification plan covers all tasks + must-haves
- [ ] Markdown report written to .gentic/reports/plan-<ts>-<slug>.md
</success_criteria>
</output>
