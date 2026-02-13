---
name: gentic:verify
description: Discover and run project checks, verify goal achievement through 3-level artifact checks and stub detection
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Discover and run all available project checks, then verify goal achievement at the code level.

Purpose: Task completion does not equal correctness. After changes, run every available check (tests, lints, type checks) independently, then verify artifacts exist, are substantive (not stubs), and are wired together. A failing test is valuable information — suppressing it is not.

Output: Structured results with pass/fail for checks, artifact verification, stub detection, and goal status.
</objective>

<execution_context>
Spawn the **verifier** agent via Task tool to run the checks and code verification. The verifier uses Bash for commands and Read/Glob/Grep for code inspection.
</execution_context>

<context>
Optional arguments: $ARGUMENTS (specific checks to run, or empty for auto-discovery)

If a plan with must-haves is available, pass it to the verifier for goal-backward verification.
</context>

<process>

## 1. Discover Checks

Auto-detect available checks from project configuration:

| Source | Check | Command |
|--------|-------|---------|
| `package.json` scripts.test | npm test | `npm test` |
| `package.json` scripts.lint | npm lint | `npm run lint` |
| `package.json` scripts.typecheck | typecheck | `npm run typecheck` |
| `Makefile` target "test" | make test | `make test` |
| `Cargo.toml` present | cargo test | `cargo test` |
| `go.mod` present | go test | `go test ./...` |
| `pytest.ini` / `pyproject.toml` [tool.pytest] | pytest | `pytest` |
| `mix.exs` present | mix test | `mix test` |

If the user provides explicit checks in $ARGUMENTS, run those instead.

## 2. Run Checks

Run each discovered check independently. **Never chain them** — a failure in one must not prevent others from running.

For each check:
1. Record start time
2. Execute the command with timeout (60s per check, 120s total)
3. Capture stdout and stderr
4. Record exit code and duration
5. Determine status: `pass` (exit 0), `fail` (exit non-zero), `skip` (not available), `error` (failed to execute)

## 3. Goal-Backward Verification (if must-haves available)

If a plan with must-haves is provided:

**Verify artifacts at 3 levels:**
- Level 1 (exists): Does the file exist?
- Level 2 (substantive): Real implementation, not a stub?
- Level 3 (wired): Connected to the rest of the system?

**Verify key links:**
- Are critical connections wired? (component→API, API→DB, form→handler, state→render)

**Stub detection on modified files:**
- Comment stubs: TODO, FIXME, PLACEHOLDER
- Empty implementations: return null, return {}, => {}
- Wiring red flags: fetch without await, handler without logic, state without render

## 4. Produce Results

Assemble structured results combining check outcomes and code verification.

`overall` is `pass` only if ALL checks pass AND no MISSING/STUB artifacts.

`goalStatus`: `verified` (all clear), `gaps_found` (stubs/missing/unwired), `checks_failed` (test/lint failures)

</process>

<anti_patterns>
- Don't modify files to make checks pass — report failures, don't fix them
- Don't skip checks that fail — run all of them
- Don't chain checks with `&&` — run each independently
- Don't suppress output — capture everything for the report
- Don't run checks that the project doesn't have configured
- Don't trust claims — verify artifacts actually exist and have substance
- Don't skip stub detection
</anti_patterns>

<output_format>
After running checks and verification, write a markdown report to `.gentic/reports/verify-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the current git branch name: lowercase, hyphenated, max 40 chars.

```markdown
# Verification Report

**Overall:** FAIL
**Goal Status:** GAPS_FOUND
**Checks run:** 3 | **Passed:** 2 | **Failed:** 1

## Check Results

| Check | Command | Status | Duration |
|-------|---------|--------|----------|
| npm test | `npm test` | PASS | 4.2s |
| npm lint | `npm run lint` | FAIL | 1.1s |
| typecheck | `npm run typecheck` | PASS | 2.1s |

## Failures

### npm lint
```
src/api.ts:12 error...
```

## Artifact Verification

| Artifact | Status | Level | Issue |
|----------|--------|-------|-------|
| src/api/auth.ts | VERIFIED | 3 | — |
| src/components/Login.tsx | STUB | 2 | return null on line 15 |

## Key Link Verification

| From | To | Status |
|------|-----|--------|
| Login.tsx | /api/auth | NOT_WIRED |

## Stubs Detected

| File | Line | Pattern | Severity |
|------|------|---------|----------|
| src/api/handler.ts | 42 | TODO | WARNING |
```

Also output the JSON results inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Available checks auto-detected from project configuration
- [ ] Each check run independently (not chained)
- [ ] Output captured for each check (stdout + stderr)
- [ ] Timing recorded for each check
- [ ] Status correctly determined from exit codes
- [ ] Artifacts verified at 3 levels if must-haves provided
- [ ] Key links verified if provided
- [ ] Stub detection run on modified files
- [ ] Goal status determined (verified/gaps_found/checks_failed)
- [ ] Markdown report written to .gentic/reports/verify-<ts>-<slug>.md
- [ ] Timeout enforced (60s per check)
</success_criteria>
</output>
