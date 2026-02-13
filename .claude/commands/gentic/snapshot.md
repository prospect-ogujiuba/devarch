---
name: gentic:snapshot
description: Print the current repo snapshot including file tree, git state, hotspots, and risk zones
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Task
---

<objective>
Build a bounded repo snapshot for situational awareness.

Purpose: Produce a quick, structured overview of any repository — file tree, manifests, git state, hotspots, and risk zones. This is the first thing to run when entering an unfamiliar codebase.

Output: Structured JSON snapshot (<= 5KB) covering tree, manifests, git state, hotspots, risk zones.
</objective>

<execution_context>
Spawn the **cartographer** agent via Task tool to perform the actual discovery. The cartographer has read-only access (Read, Glob, Grep) and produces the snapshot.
</execution_context>

<context>
No arguments required. Operates on the current working directory.
</context>

<process>

## 1. File Tree

Produce directory structure to depth 3. Exclude: node_modules, vendor, .git, .gentic, .claude, __pycache__, dist, build, coverage.

## 2. Manifests

Parse and summarize: package.json, Cargo.toml, go.mod, requirements.txt, pyproject.toml. Extract name, version, dependency count — not full contents.

## 3. Git State

Current branch, uncommitted changes count, last 5 commit messages (one-line each).

## 4. Hotspots

Top 10 most-changed files from git history (last 3 months). Use `git log --since="3 months ago" --name-only` to compute frequency.

## 5. Risk Zones

Identify files matching these criteria:

| Pattern | Reason |
|---------|--------|
| Source files with no corresponding test file | Missing test coverage |
| Files with LOC > 500 | Complexity risk |
| Paths matching config/, deploy/, .env, *.sql | Sensitive path |

## 6. Assemble Output

Produce a single JSON object with all sections. Enforce the 5KB budget — truncate tree branches if needed.

</process>

<anti_patterns>
- Don't include full file contents — just paths and summaries
- Don't output secrets (API keys, tokens, PEM blocks) even if found in files
- Don't fail if git history is unavailable — skip hotspots gracefully
- Don't exceed the 5KB output budget
- Don't modify any files
</anti_patterns>

<output_format>
After generating the snapshot data, write a markdown report to `.gentic/reports/snapshot-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`. Derive `<slug>` from the current git branch name: lowercase, hyphenated, max 40 chars.

```markdown
# Repo Snapshot

Generated: <ISO-8601 timestamp>

## File Tree

<depth-3 tree as indented list>

## Manifests

| File | Details |
|------|---------|
| package.json | name: ..., deps: 12 |

## Git State

- **Branch:** main
- **Uncommitted:** 3 files
- **Recent commits:**
  - abc1234 fix: resolve auth bug
  - def5678 feat: add pagination

## Hotspots (last 3 months)

| File | Changes |
|------|---------|
| src/index.ts | 15 |

## Risk Zones

| File | Reason |
|------|--------|
| config/deploy.yml | sensitive path |
```

Also output the JSON snapshot inline for programmatic consumption.
</output_format>

<success_criteria>
- [ ] File tree produced (depth 3, exclusions applied)
- [ ] Manifests parsed (not raw dumped)
- [ ] Git state captured (branch, changes, recent commits)
- [ ] Hotspots computed from git log (top 10, last 3 months)
- [ ] Risk zones identified (untested, large, sensitive)
- [ ] Output is valid JSON
- [ ] Output <= 5KB
- [ ] No secrets in output
- [ ] Markdown report written to .gentic/reports/snapshot-<ts>-<slug>.md
</success_criteria>
