---
name: gentic:research
description: Deep-dive research on a codebase topic — patterns, conventions, dependencies, prior art
argument-hint: "<topic or question to research>"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Perform deep codebase research on a topic before planning or implementation.

Purpose: Research before you plan. Investigate how the codebase handles a topic — what patterns it uses, what dependencies exist, what conventions are followed, what gaps remain. The output feeds into the Architect's planning process.

Output: Structured research report with findings, patterns, dependencies, risks, and recommendations.
</objective>

<execution_context>
Spawn the **researcher** agent via Task tool to perform the analysis. The researcher has read-only access (Read, Glob, Grep) and produces the research report.
</execution_context>

<context>
Research topic: $ARGUMENTS
</context>

<process>

## 1. Parse the Topic

Understand what needs to be researched. Break broad topics into specific research questions. Determine scope and depth.

## 2. Survey the Landscape

Get a high-level view of the relevant area:
- Glob for relevant file patterns
- Read key entry points (main modules, index files, config)
- Grep for topic keywords across the codebase

Build a map of relevant files before deep-diving.

## 3. Deep Dive

For each relevant area, perform detailed analysis:

| Dimension | Questions to Answer |
|-----------|-------------------|
| Architecture | How are components organized? What's the dependency graph? |
| Conventions | Naming, error handling, logging, config patterns? |
| Data flow | How does data enter, transform, and exit the system? |
| Dependencies | What external libs are used? How are they wrapped? |
| Test coverage | What's tested? What test patterns are used? |
| Technical debt | What's marked TODO/FIXME/HACK? What's fragile? |
| Prior art | Has something similar been done in this codebase before? |

## 4. Synthesize Findings

Organize findings into a structured report:
- Summary (one-paragraph answer)
- Key findings (numbered, with file references)
- Patterns discovered
- Dependencies (internal and external)
- Risks and gaps
- Actionable recommendations

Every finding must cite specific files and line numbers.

</process>

<anti_patterns>
- Don't produce findings without file references — every claim needs a citation
- Don't read every file — survey first, then deep-dive on relevant areas
- Don't speculate about code you haven't read
- Don't modify any files — research is read-only
- Don't produce a wall of text — structure with headers and tables
- Don't skip test analysis — understanding what's tested matters
- Don't ignore TODO/FIXME/HACK comments — they're research signals
</anti_patterns>

<output_format>
After completing the research, write a markdown report to `.gentic/reports/research-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the research topic: lowercase, hyphenated, 3-5 key words, max 40 chars, drop articles/prepositions. Example: "How does authentication work in this project?" → `auth-flow`.

```markdown
# Research Report

**Topic:** How does authentication work in this project?
**Files analyzed:** 12

## Summary

Auth uses JWT tokens issued by /api/auth/login with 24h expiry...

## Key Findings

### 1. JWT token management
- Tokens issued in `src/auth/token.ts:28` with 24h expiry
- Verification middleware in `src/middleware/auth.ts:15`

### 2. ...

## Patterns

| Pattern | Example |
|---------|---------|
| Middleware-based auth checks | `src/middleware/auth.ts:15` |

## Dependencies

| Package | Usage | File |
|---------|-------|------|
| jsonwebtoken | Token signing/verification | `src/auth/token.ts` |

## Risks & Gaps

| Risk | Severity | File |
|------|----------|------|
| No token refresh mechanism | medium | `src/auth/token.ts` |

## Recommendations

1. Add token refresh endpoint to avoid forced re-login after 24h
2. Consider rate-limiting on /api/auth/login
```

Also output the JSON research report inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] Research question clearly stated
- [ ] Relevant files identified via survey (Glob + Grep)
- [ ] Key files read and analyzed in detail
- [ ] Every finding cites specific file and line number
- [ ] Patterns and conventions documented
- [ ] Dependencies identified (internal and external)
- [ ] Risks and gaps surfaced
- [ ] Actionable recommendations provided
- [ ] Markdown report written to .gentic/reports/research-<ts>-<slug>.md
- [ ] No files modified (read-only)
</success_criteria>
