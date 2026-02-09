# Phase 15: Validation & Parity - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Final verification that v1.1 schema reconciliation produces 1:1 legacy parity across all 173 services, plus import boundary testing. This phase fixes remaining gaps from Phase 12 (9 parity failures + network fallback), extends the parity tool, adds boundary testing, and declares v1.1 complete when tools exit 0.

</domain>

<decisions>
## Implementation Decisions

### Parity failure triage
- Fix all 9 known failures to reach true 100% parity — no ambiguous "expected differences" left undocumented
- Network fallback gap from Phase 12 (stack.go lines 176-178, 195-198) is fixed in this phase as part of parity work
- Explicit whitelist file (JSON) for any triaged expected differences — parity tool reads it and skips whitelisted checks
- Each whitelist entry must include service name, field, and human-readable reason
- Golden services (php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api) can NEVER be whitelisted — any failure in a golden service is a hard blocker

### Verification tooling
- Extend existing `api/cmd/verify-parity/main.go` rather than building a separate test suite
- Tool stays DB-dependent — validates the real import-then-generate pipeline, not mocked paths
- Single exit code: 0 = all pass (with whitelisted exceptions), 1 = any failure
- Add `--json` flag for machine-readable structured output; console output stays human-friendly by default
- Whitelist file and golden service list committed to `api/cmd/verify-parity/` — version-controlled and auditable

### Boundary test strategy
- Separate command (`api/cmd/verify-boundary/` or similar) — boundary testing is infrastructure limits, not data correctness
- Synthetic YAML multiplier for payload generation: duplicate real compose services with unique names until target size
- Tests require a running API server — exercise full HTTP path (multipart encoding -> route middleware -> size limit -> handler)
- 300MB rejection must return HTTP 413 with JSON body: `{"error": "Import payload exceeds 256MB limit", "max_bytes": 268435456, "received_bytes": ...}`

### Deliverable format
- Tool output IS the report — no separate markdown report to maintain
- v1.1 milestone complete when: (1) verify-parity exits 0, (2) boundary tests pass
- Re-runnable at any time as living proof of parity

### Claude's Discretion
- Internal structure of the whitelist JSON schema
- Synthetic payload generation algorithm details
- Console output formatting and progress indicators
- How to structure the boundary test command's flags and output

</decisions>

<specifics>
## Specific Ideas

- Phase 12 verification identified exact gaps: stack.go network fallback (lines 176-178, 195-198) + 9 service failures (3 port conflicts, 1 healthcheck format, 4 missing files, 1 dependency resolution)
- Existing parity tool is 569 lines, compares 14 field groups — extend rather than rewrite
- Phase 13 set the import limit at 256MB via `STACK_IMPORT_MAX_BYTES` with route-level `MaxBodySize(256MB)`
- Golden services chosen because they represent the most complex compose configurations in the project

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 15-validation-parity*
*Context gathered: 2026-02-09*
