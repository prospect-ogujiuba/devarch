---
name: gentic:audit
description: Validate that implementation matches the plan — find gaps, drift, stubs, and unwired artifacts
argument-hint: "[plan file or path]"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Validate that the implementation matches what was planned.

Purpose: Plans drift. After executing a plan, audit the actual changes against the planned steps to find gaps (planned but not done), drift (done differently than planned), stubs (placeholders instead of real implementation), and scope creep (done but not planned). The audit closes the loop between planning and execution.

Output: Structured audit report with step-by-step comparison, must-haves verification, stub detection, coverage analysis, and findings.
</objective>

<execution_context>
Spawn the **auditor** agent via Task tool to perform the analysis. The auditor has read-only access (Read, Glob, Grep) and compares plans against implementation.
</execution_context>

<context>
Optional plan reference: $ARGUMENTS (if empty, uses the most recent plan in .gentic/reports/)
</context>

<process>

## 1. Load Plan

Find the plan to audit:
- If a specific plan file/path is provided, read it
- Otherwise, find the most recent `.gentic/reports/plan-*.md` file
- Extract all planned tasks with file, action, risk level, dependencies
- Extract the verification plan
- Extract must-haves (truths, artifacts, key_links) if present

## 2. Load Implementation

Read the actual implementation state:
- Session ledger (`changes.jsonl`) for what was actually modified
- Verification results (`verifications.jsonl`) for test outcomes
- The actual files that were supposed to change — verify changes are present

## 3. Compare Plan vs. Reality

For each planned step, determine status:

| Status | Criteria |
|--------|----------|
| `implemented` | File modified as specified, change matches plan |
| `partial` | File modified but change is incomplete |
| `missing` | Step was not executed |
| `blocked` | Step attempted but blocked (per ledger) |
| `drifted` | Implementation differs from plan intent |

Also identify unplanned changes — files modified that aren't in any plan step.

## 4. Verify Must-Haves

If the plan included must-haves:

**Truths:** For each observable truth, check if the codebase enables it.
- Status: VERIFIED | FAILED | UNCERTAIN

**Artifacts:** For each required artifact, check 3 levels:
- Level 1 (exists): File exists?
- Level 2 (substantive): Real implementation, not a stub?
- Level 3 (wired): Connected to the rest of the system?

**Key links:** For each critical connection, verify wiring:
- Status: WIRED | PARTIAL | NOT_WIRED

## 5. Detect Stubs

Scan all files that were supposedly changed for stub patterns:

**Comment stubs:** TODO, FIXME, XXX, PLACEHOLDER
**Empty implementations:** return null, return {}, return [], => {}
**Wiring stubs:** fetch without await, state without render, handler without logic, query result not returned

Report each stub with file, line, pattern, and severity.

## 6. Check Verification Coverage

Map verification results to plan steps:
- Were all planned checks run?
- Which steps have no verification coverage?
- Are there untested changes?

## 7. Produce Audit Report

Combine step comparison, must-haves verification, stub detection, unplanned changes, and coverage analysis into the audit verdict:

| Condition | Verdict |
|-----------|---------|
| All steps implemented + verified + no stubs + must-haves satisfied | `pass` |
| All implemented, some unverified or minor stubs | `pass-with-gaps` |
| Some steps missing or drifted | `incomplete` |
| Critical steps missing, stubs in core paths, must-haves failed | `fail` |

**Structured gap output:**

```markdown
## Gaps

| # | Type | Description | Severity | Remediation |
|---|------|-------------|----------|-------------|
| 1 | MISSING | Pagination middleware not created | high | Execute plan step 3 |
| 2 | STUB | return {} in handler.ts:42 | high | Implement actual logic |
| 3 | NOT_WIRED | Login.tsx not connected to /api/auth | high | Add fetch call in onSubmit |
```

</process>

<anti_patterns>
- Don't approve without reading the actual files — check the code, not just summaries
- Don't flag cosmetic differences as drift — focus on functional intent
- Don't ignore unplanned changes — scope creep is a finding
- Don't accept "tests pass" as proof — verify the code matches the plan
- Don't skip stub detection — stubs are the #1 false-completion signal
- Don't modify any files during audit
- Don't produce findings without file/line evidence
</anti_patterns>

<output_format>
After completing the audit, write a markdown report to `.gentic/reports/audit-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the plan being audited: use the plan file's own slug if present, otherwise the current git branch name. Lowercase, hyphenated, max 40 chars.

```markdown
# Audit Report

**Plan:** plan-20250115-143500-add-pagination-admin.md
**Verdict:** INCOMPLETE
**Steps:** 4/5 implemented | 1 missing
**Stubs:** 2 detected
**Must-Haves:** 3/5 verified

## Step Comparison

| # | Planned Step | Status | Evidence |
|---|-------------|--------|----------|
| 1 | Add pagination params | implemented | user.ts modified, params at line 45 |
| 2 | Create pagination middleware | implemented | pagination.ts created |
| 3 | Update response serializer | implemented | serializer.ts modified |
| 4 | Add admin endpoint | **missing** | admin.ts not modified |
| 5 | Write integration tests | implemented | tests added |

## Must-Haves Verification

### Truths
| Truth | Status | Evidence |
|-------|--------|----------|
| Paginated results returned | VERIFIED | GET /users returns {data, total, page} |
| Admin can list all users | FAILED | Admin endpoint missing |

### Artifacts
| Artifact | Status | Level |
|----------|--------|-------|
| src/api/pagination.ts | VERIFIED | L3 (wired) |
| src/api/admin.ts | MISSING | — |

### Key Links
| From | To | Status |
|------|-----|--------|
| controller → pagination | WIRED | import + usage confirmed |
| admin → user model | NOT_WIRED | admin.ts missing |

## Stubs Detected

| File | Line | Pattern | Severity |
|------|------|---------|----------|
| src/api/handler.ts | 42 | return {} | high |
| src/utils/format.ts | 15 | TODO | low |

## Unplanned Changes

| File | Operation | Concern |
|------|-----------|---------|
| src/utils/format.ts | edit | Not in plan — potential scope creep |

## Verification Coverage

| Step | Covered By | Result |
|------|-----------|--------|
| 1 | npm test | pass |
| 2 | npm test | pass |
| 3 | npm test | pass |
| 4 | — | **NO COVERAGE** |
| 5 | npm test | pass |

## Findings

| Severity | Finding | Recommendation |
|----------|---------|----------------|
| high | Step 4 (admin endpoint) not implemented | Execute remaining step or update plan |
| high | Stub: return {} in handler.ts:42 | Implement actual logic |
| high | Key link admin → user model NOT_WIRED | Create admin endpoint with user model query |
| medium | Unplanned edit to format.ts | Document reason or revert |
```

Also output the JSON audit report inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Plan located and all steps parsed (including must-haves)
- [ ] Implementation state loaded from ledger + actual files
- [ ] Every planned step has a comparison status
- [ ] Must-haves verified (truths, artifacts, key_links) if present
- [ ] Stub detection completed on all modified files
- [ ] Unplanned changes identified
- [ ] Verification coverage mapped per step
- [ ] Findings have severity and actionable recommendations
- [ ] Verdict computed from evidence (including must-haves and stubs)
- [ ] Structured gap output produced
- [ ] Markdown report written to .gentic/reports/audit-<ts>-<slug>.md
- [ ] No files modified during audit
</success_criteria>
</output>
