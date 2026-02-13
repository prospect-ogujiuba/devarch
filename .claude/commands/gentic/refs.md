---
name: gentic:refs
description: Find all references to a symbol or file path — classified by type and ranked by confidence
argument-hint: "<symbol or file path>"
allowed-tools:
  - Read
  - Glob
  - Grep
  - Write
---

<objective>
Find and classify all references to a symbol or file path across the codebase.

Purpose: Quickly understand how a function, class, variable, or file is used — where it's called, imported, referenced as a type, or mentioned in strings. Results are ranked by confidence so the most relevant matches (direct calls, imports) appear first.

Output: Structured reference table grouped by file, with classification and confidence ranking.
</objective>

<context>
Query: $ARGUMENTS
</context>

<process>

## 1. Search for References

Use the Grep tool to find all occurrences of the query across the codebase:
- Search with the exact query string
- Exclude common noise directories: node_modules, .git, dist, build, coverage, .gentic
- Use `output_mode: "content"` with line numbers enabled
- Capture file path, line number, and match text for each result

## 2. Classify Each Match

For each match, classify it into one of these categories based on the surrounding text:

| Classification | Pattern | Confidence |
|---------------|---------|------------|
| `call` | `query(`, `query.`, `await query` | 1 (highest) |
| `import` | `require('...query')`, `import.*query`, `from.*query` | 2 |
| `definition` | `function query`, `class query`, `const query =`, `def query` | 3 |
| `type-ref` | `query[]`, `: query`, `<query>`, `extends query`, `implements query` | 4 |
| `string` | `'query'`, `"query"`, backtick containing query | 5 (lowest) |

If a match fits multiple classifications, use the highest-confidence one.

## 3. Group and Rank Results

- Group matches by file path
- Within each file, sort by line number
- Sort file groups by best (lowest) confidence score in the group
- Files with direct calls appear before files with only string mentions

## 4. Output Results

Output a JSON summary inline:

```json
{
  "query": "<query>",
  "totalReferences": <count>,
  "files": [
    {
      "path": "src/core/config.js",
      "references": [
        { "line": 45, "classification": "call", "confidence": 1, "text": "loadConfig(cwd)" },
        { "line": 12, "classification": "import", "confidence": 2, "text": "const { loadConfig } = require('./config')" }
      ]
    }
  ]
}
```

Then write a markdown report to `.gentic/reports/refs-<timestamp>-<slug>.md`. Derive `<slug>` from the query: lowercase, hyphenated, max 40 chars. Example: "loadConfig" → `loadconfig`.

</process>

<output_format>
Write the report to `.gentic/reports/refs-<timestamp>-<slug>.md` using the Write tool. Use ISO-8601 date (YYYYMMDD-HHmmss) for `<timestamp>`.

```markdown
# Reference Report: `<query>`

**Total references:** N across M files

## By Classification

| Classification | Count |
|---------------|-------|
| call | 5 |
| import | 3 |
| type-ref | 2 |
| string | 1 |

## References by File

### src/core/config.js (3 refs)

| Line | Type | Text |
|------|------|------|
| 12 | import | `const { loadConfig } = require('./config')` |
| 45 | call | `loadConfig(cwd)` |
| 88 | call | `const cfg = loadConfig(root)` |

### hooks/guard-bash.js (1 ref)

| Line | Type | Text |
|------|------|------|
| 7 | import | `const { loadConfig } = require('./core/config')` |
```

Also output the JSON report inline for programmatic consumption.
</output_format>

<anti_patterns>
- Don't search without excluding noise directories (node_modules, .git, dist)
- Don't return raw grep output — always classify and rank
- Don't truncate results silently — report total count even if display is limited
- Don't conflate definition sites with usage sites
- Don't include matches from binary files
</anti_patterns>

<success_criteria>
- [ ] All references found via Grep with line numbers
- [ ] Each match classified (call, import, definition, type-ref, string)
- [ ] Results grouped by file, sorted by confidence
- [ ] JSON summary output inline
- [ ] Markdown report written to .gentic/reports/refs-<ts>-<slug>.md
- [ ] Noise directories excluded from search
</success_criteria>
