---
name: gentic:context
description: Generate a bounded neighborhood context pack for a target file
argument-hint: "<file-path>"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Task
---

<objective>
Generate a bounded context pack for a target file.

Purpose: Before modifying a file, understand its neighborhood — what it imports, what imports it, what tests cover it, and what files tend to change alongside it. This prevents blind edits.

Output: Structured JSON context pack with target summary, imports, test files, co-change neighbors, and references.
</objective>

<execution_context>
Spawn the **librarian** agent via Task tool to perform the actual discovery. The librarian has read-only access (Read, Glob, Grep) and produces the context pack.
</execution_context>

<context>
Target file: $ARGUMENTS
</context>

<process>

## 1. Read Target

Read the target file. Extract: path, line count, exported symbols (functions, classes, types, constants).

## 2. Trace Imports

Parse import/require statements from the target file. Resolve relative paths to actual file paths. List each direct dependency.

## 3. Find Test Files

Search for test files using heuristic patterns:

| Pattern | Example |
|---------|---------|
| `test/<path>/<name>.test.<ext>` | test/core/engine.test.ts |
| `__tests__/<name>.test.<ext>` | __tests__/engine.test.ts |
| `<name>.spec.<ext>` (co-located) | engine.spec.ts |

## 4. Find Co-change Neighbors

Query git log for files that frequently appear in the same commits as the target:

```bash
git log --pretty=format:"%H" -- <target> | head -30
```

For each commit, get changed files. Rank by `sharedCommits / totalCommits`. Max 10 results.

## 5. Find References

Use Grep to find files that import or reference the target's exports. Search for import statements pointing to the target path. Max 10 results.

## 6. Assemble Pack

Combine all sections into a single JSON context pack.

</process>

<anti_patterns>
- Don't include full file contents for neighbors — just paths and summaries
- Don't fail if git history is unavailable — skip co-change analysis gracefully
- Don't return more than 10 neighbors or 10 references
- Don't modify any files
- Don't read files outside the project directory
</anti_patterns>

<output_format>
```json
{
  "target": { "path": "src/core/engine.ts", "lines": 245, "exports": ["Engine", "createEngine"] },
  "imports": ["src/utils/helpers.ts", "src/types/index.ts"],
  "testFiles": ["test/core/engine.test.ts"],
  "cochangeNeighbors": [
    { "file": "src/core/parser.ts", "score": 0.85, "sharedCommits": 12 }
  ],
  "references": [
    { "file": "src/api/handler.ts", "symbol": "Engine", "line": 12 }
  ]
}
```
</output_format>

<success_criteria>
- [ ] Target file read and summarized (path, lines, exports)
- [ ] Direct imports traced and resolved
- [ ] Test files discovered (heuristic patterns)
- [ ] Co-change neighbors ranked by score (max 10)
- [ ] References found via grep (max 10)
- [ ] Output is valid JSON
- [ ] No full file contents in output (paths and summaries only)
- [ ] No files modified
</success_criteria>
