# Phase 15: Validation & Parity - Research

**Researched:** 2026-02-10
**Domain:** Verification tooling, parity testing, boundary testing
**Confidence:** HIGH

## Summary

Phase 15 is the milestone exit gate for v1.1 Schema Reconciliation. Two verification dimensions: (1) forward-looking — extend verify-parity tool to achieve 100% parity across 173 services and build verify-boundary tool for import size limits, (2) backward-looking — audit all prior phases (10-14) for unresolved gaps that block milestone completion.

**Current state:** verify-parity tool is 821 lines with whitelist/golden-services support already committed (546220d). Golden services JSON exists with 7 entries, whitelist JSON is empty. Network fallback bug from Phase 12 has been FIXED (stack.go no longer has lines 176-178, 195-198 hardcoded fallback). Recent commit 54e41df resolved 8 of 9 original parity failures. Phase 12 verification showed 164/173 passing (95%), recent fixes likely improved this but exact current status unknown without running the tool.

**Primary recommendation:** Run verify-parity against current codebase to get actual failure count, triage remaining failures (likely 1-5 services), fix generator/importer bugs or whitelist with documented reasons. Build verify-boundary command using io.Pipe streaming pattern to test 200MB accept / 300MB rejection. All prior phases (10-14) PASSED verification with Phase 13 having 4 deferred human tests that Phase 15 addresses.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Fix all 9 known failures to reach true 100% parity — no ambiguous "expected differences" left undocumented
- Network fallback gap from Phase 12 (stack.go lines 176-178, 195-198) is fixed in this phase as part of parity work
- Explicit whitelist file (JSON) for any triaged expected differences — parity tool reads it and skips whitelisted checks
- Each whitelist entry must include service name, field, and human-readable reason
- Golden services (php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api) can NEVER be whitelisted — any failure in a golden service is a hard blocker
- Extend existing `api/cmd/verify-parity/main.go` rather than building a separate test suite
- Tool stays DB-dependent — validates the real import-then-generate pipeline, not mocked paths
- Single exit code: 0 = all pass (with whitelisted exceptions), 1 = any failure
- Add `--json` flag for machine-readable structured output; console output stays human-friendly by default
- Whitelist file and golden service list committed to `api/cmd/verify-parity/` — version-controlled and auditable
- Separate command (`api/cmd/verify-boundary/` or similar) — boundary testing is infrastructure limits, not data correctness
- Synthetic YAML multiplier for payload generation: duplicate real compose services with unique names until target size
- Tests require a running API server — exercise full HTTP path (multipart encoding -> route middleware -> size limit -> handler)
- 300MB rejection must return HTTP 413 with JSON body: `{"error": "Import payload exceeds 256MB limit", "max_bytes": 268435456, "received_bytes": ...}`
- Tool output IS the report — no separate markdown report to maintain
- v1.1 milestone complete when: (1) verify-parity exits 0, (2) boundary tests pass
- Re-runnable at any time as living proof of parity

### Claude's Discretion
- Internal structure of the whitelist JSON schema
- Synthetic payload generation algorithm details
- Console output formatting and progress indicators
- How to structure the boundary test command's flags and output
</user_constraints>

## Standard Stack

### Core Tools
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| Go testing stdlib | 1.22+ | Test infrastructure | Built-in table-driven test pattern, no external deps |
| io.Pipe | stdlib | Streaming multipart generation | Prevents memory exhaustion for 200MB+ payloads |
| mime/multipart | stdlib | Multipart form data handling | Standard library multipart writer/reader |
| net/http/httptest | stdlib | HTTP testing | Mock HTTP requests without network |
| encoding/json | stdlib | Whitelist/golden service files | Parse JSON governance files |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| yaml.v3 | gopkg.in/yaml.v3 | YAML generation | Synthetic DevArchFile payload |
| bytes.Buffer | stdlib | Small payload buffering | When payload < 50MB |
| os.CreateTemp | stdlib | Temp file handling | Intermediate storage for generated YAML |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| io.Pipe | bytes.Buffer | Buffer causes OOM for 300MB payloads; Pipe streams without intermediate storage |
| Custom test framework | Go testing stdlib | Adding framework (testify, ginkgo) adds deps without benefit for CLI tools |
| httptest.Server | Real API server | verify-boundary MUST test real HTTP path including middleware, httptest skips chi router |

**Installation:**
No new dependencies required — all tools are Go stdlib.

## Architecture Patterns

### Recommended verify-parity Extension Structure
Tool already has whitelist/golden service support (commit 546220d). Extension pattern:
```
main.go (821 lines current)
├── loadWhitelist()              # Already exists
├── loadGoldenServices()         # Already exists
├── validateWhitelistAgainstGolden()  # Already exists
├── applyWhitelistGovernance()   # Already exists — filters failures through whitelist
├── verifyService()              # Main comparison loop — extend if needed
└── JSON output marshaling        # Add --json flag support
```

**What to add:**
- `--json` flag handling in main()
- JSON output struct with summary, service reports, golden flags
- `--strict` mode that exits 1 when whitelist.entries > 0

### Recommended verify-boundary Structure
New command at `api/cmd/verify-boundary/main.go`:
```go
package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"gopkg.in/yaml.v3"
)

func main() {
	// Parse flags: --api-url, --api-key, --stack-name, --json

	// Pre-test: Create test stack via POST /api/v1/stacks

	// Test 1: 200MB payload (VALD-03)
	if err := testPayloadSize(200 << 20); err != nil {
		// Should succeed (HTTP 200 or 400 for parse error, NOT 413)
	}

	// Test 2: 300MB payload (VALD-04)
	if err := testPayloadSize(300 << 20); err != nil {
		// Should fail with HTTP 413 + JSON error body
	}

	// Cleanup: Delete test stack

	// Exit 0 if both pass, 1 if any fails
}

func generateSyntheticYAML(targetBytes int64) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		enc := yaml.NewEncoder(pw)
		// Generate DevArchFile with N service instances
		// Each instance ~10KB, need targetBytes/10KB instances
		enc.Encode(&devarchFile)
	}()
	return pr
}

func testPayloadSize(bytes int64) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer mw.Close()
		part, _ := mw.CreateFormFile("file", "stack.yml")
		yamlReader := generateSyntheticYAML(bytes)
		io.Copy(part, yamlReader)
	}()

	req, _ := http.NewRequest("POST", apiURL+"/import", pr)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	// Check resp.StatusCode and body
	return nil
}
```

### Pattern: Streaming Multipart with io.Pipe
**What:** Create multipart payload without buffering entire content in memory
**When to use:** Payloads > 100MB where buffer causes OOM
**Example:**
```go
// Source: Go stdlib mime/multipart docs + https://blog.depa.do/post/bufferless-multipart-post-in-go
pr, pw := io.Pipe()
mw := multipart.NewWriter(pw)

go func() {
	defer pw.Close()
	defer mw.Close()

	part, err := mw.CreateFormFile("file", "stack.yml")
	if err != nil {
		return
	}

	// Stream content into part
	io.Copy(part, contentReader)
}()

// pr can now be used as request body
req, _ := http.NewRequest("POST", url, pr)
req.Header.Set("Content-Type", mw.FormDataContentType())
```

**Critical:** Must call `mw.Close()` to write multipart trailer, otherwise request is invalid.

### Pattern: Synthetic YAML Generation via Repetition
**What:** Generate large DevArchFile by repeating service instances with unique names
**When to use:** Need predictable payload size for boundary testing
**Example:**
```go
type DevArchFile struct {
	Version  int              `yaml:"version"`
	Stack    StackDefinition  `yaml:"stack"`
	Services []ServiceInstance `yaml:"services"`
}

func generateSyntheticYAML(targetBytes int64) io.Reader {
	const instanceSize = 10000 // ~10KB per instance
	instanceCount := int(targetBytes / instanceSize)

	file := DevArchFile{
		Version: 1,
		Stack: StackDefinition{Name: "boundary-test", Description: "Synthetic test stack"},
		Services: make([]ServiceInstance, instanceCount),
	}

	for i := 0; i < instanceCount; i++ {
		file.Services[i] = ServiceInstance{
			Name:     fmt.Sprintf("boundary-svc-%04d", i),
			Template: "nginx",
			Overrides: ServiceOverrides{
				Environment: map[string]string{
					"PADDING_1": strings.Repeat("x", 500),
					"PADDING_2": strings.Repeat("y", 500),
					// ... padding to reach ~10KB
				},
			},
		}
	}

	var buf bytes.Buffer
	yaml.NewEncoder(&buf).Encode(&file)
	return &buf
}
```

### Anti-Patterns to Avoid
- **bytes.Buffer for 200MB+ payloads:** Causes memory allocation equal to payload size. Use io.Pipe instead.
- **httptest.Server for boundary tests:** Bypasses chi router and middleware chain. Must test against real API server on localhost:8550.
- **Separate parity test suite:** verify-parity already exists and is 821 lines. Extend it, don't rewrite.
- **Mocking container runtime:** Phase 13 verification warns against this — tools must exercise real import-generate path.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Multipart streaming | Custom buffered writer | io.Pipe + mime/multipart | Edge cases: boundary escaping, trailer formatting, partial write handling |
| HTTP testing with middleware | httptest.Server | Real API server on localhost | httptest bypasses chi router, middleware chain, MaxBodySize logic |
| YAML generation | String concatenation | yaml.v3 Encoder | Handles escaping, indentation, type marshaling correctly |
| Test fixtures | JSON files on disk | In-memory structs | Easier to version, modify, see full context in code |
| JSON output formatting | fmt.Printf | json.MarshalIndent | Type safety, consistent formatting, machine-parseable |

**Key insight:** Go stdlib is production-grade for CLI tools. Reaching for external test frameworks (testify, ginkgo) or HTTP mocking libraries adds complexity without benefit for verification CLIs that exercise real system paths.

## Common Pitfalls

### Pitfall 1: Forgetting multipart.Writer.Close()
**What goes wrong:** HTTP request hangs or server rejects with "unexpected EOF"
**Why it happens:** multipart writer must write closing boundary `--boundary--` via Close()
**How to avoid:** Always defer `mw.Close()` immediately after creating multipart.Writer
**Warning signs:** Test hangs indefinitely, server logs "failed to read multipart"

### Pitfall 2: Assuming verify-parity needs rewrite
**What goes wrong:** Duplicate 821 lines of parity logic instead of extending existing tool
**Why it happens:** Plan says "extend verify-parity" but tool is large and adding --json seems hard
**How to avoid:** Tool already has all comparison logic. Add JSON output struct + marshal at end. < 100 lines.
**Warning signs:** Creating new files like `verify-parity-v2/main.go` or `parity_test.go`

### Pitfall 3: Buffering 300MB payload to measure size
**What goes wrong:** OOM or swap thrashing during boundary test
**Why it happens:** Trying to count bytes by reading entire payload into memory
**How to avoid:** Use io.Pipe streaming — server enforces size limit, test just verifies rejection. Don't need to know exact bytes, server tells you in 413 response.
**Warning signs:** `bytes.Buffer` or `ioutil.ReadAll` in testPayloadSize()

### Pitfall 4: Testing boundary on non-import endpoint
**What goes wrong:** All non-import endpoints have 10MB cap, import has 256MB cap
**Why it happens:** Testing POST /api/v1/services instead of POST /api/v1/stacks/{name}/import
**How to avoid:** Import endpoint is `/api/v1/stacks/{stackName}/import` per routes.go:194
**Warning signs:** 200MB payload rejected with 413 (should succeed on import route)

### Pitfall 5: Network fallback still exists (outdated info)
**What goes wrong:** Assuming Phase 12 gap still needs fixing
**Why it happens:** Phase 12 verification report says lines 176-178, 195-198 have fallback
**How to avoid:** Verification ran 2026-02-09. Commit 54e41df (same day) fixed 8 failures. Check current stack.go state — fallback is GONE.
**Warning signs:** Creating fix for already-fixed issue

### Pitfall 6: Whitelist JSON schema mismatch with existing code
**What goes wrong:** Creating incompatible whitelist format, tool can't parse
**Why it happens:** Not checking existing loadWhitelist() implementation
**How to avoid:** verify-parity main.go lines 307-333 define schema: `{"entries": [{"service": "...", "field": "...", "reason": "..."}]}`
**Warning signs:** Adding `"version"`, `"timestamp"`, or other fields not in existing schema

## Code Examples

Verified patterns from existing codebase and Go stdlib:

### verify-parity Whitelist Governance (existing)
```go
// Source: api/cmd/verify-parity/main.go lines 307-365
func loadWhitelist(path string) (map[string]whitelistEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file whitelistFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	entries := make(map[string]whitelistEntry, len(file.Entries))
	for _, entry := range file.Entries {
		service := strings.TrimSpace(strings.ToLower(entry.Service))
		field := strings.TrimSpace(strings.ToLower(entry.Field))
		reason := strings.TrimSpace(entry.Reason)
		if service == "" || field == "" || reason == "" {
			return nil, fmt.Errorf("invalid whitelist entry: service, field, and reason are required")
		}
		entries[whitelistKey(service, field)] = entry
	}
	return entries, nil
}

func validateWhitelistAgainstGolden(whitelist map[string]whitelistEntry, golden map[string]bool) error {
	for _, entry := range whitelist {
		if golden[entry.Service] {
			return fmt.Errorf("service %q is golden and cannot be whitelisted", entry.Service)
		}
	}
	return nil
}
```

### Streaming Multipart POST with io.Pipe
```go
// Source: https://blog.depa.do/post/bufferless-multipart-post-in-go
// Pattern used for 200MB+ payload without buffering
func uploadLargeFile(url, apiKey string, content io.Reader, contentSize int64) (*http.Response, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer mw.Close()

		part, err := mw.CreateFormFile("file", "stack.yml")
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(part, content); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequest("POST", url, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)

	return http.DefaultClient.Do(req)
}
```

### Stack Import Handler Size Enforcement (existing)
```go
// Source: api/internal/api/handlers/stack_import.go lines 34-104
func (h *StackHandler) ImportStack(w http.ResponseWriter, r *http.Request) {
	importMaxBytes := int64(256 << 20)
	if envVal := os.Getenv("STACK_IMPORT_MAX_BYTES"); envVal != "" {
		if parsed, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			importMaxBytes = parsed
		}
	}

	// Size cap at multipart reader level
	mr := multipart.NewReader(io.LimitReader(r.Body, importMaxBytes+4096), boundary)

	// ... find file part ...

	// Stream to temp file with size check
	receivedBytes, err := io.Copy(tmpFile, filePart)
	if receivedBytes > importMaxBytes {
		writeImportError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("Import payload exceeds %dMB limit", importMaxBytes>>20),
			map[string]interface{}{
				"max_bytes":      importMaxBytes,
				"received_bytes": receivedBytes,
			})
		return
	}
}
```

### Route-Level Size Override (existing)
```go
// Source: api/internal/api/routes.go lines 28-34, 191-194
importMaxBytes := int64(256 << 20)
if envVal := os.Getenv("STACK_IMPORT_MAX_BYTES"); envVal != "" {
	if parsed, err := strconv.ParseInt(envVal, 10, 64); err == nil {
		importMaxBytes = parsed
	}
}

// Later in stack routes:
r.Route("/{name}", func(r chi.Router) {
	r.With(mw.MaxBodySize(importMaxBytes)).Post("/import", stackHandler.ImportStack)
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Phase 12 had 9 parity failures | Commit 54e41df fixed 8 failures (2026-02-09) | Recent fix | Likely 1 failure remains, not 9 |
| stack.go had network fallback (lines 176-178, 195-198) | Fallback removed | Between Phase 12 verification and now | Phase 12 gap #1 is RESOLVED |
| 569-line verify-parity tool | 821-line tool with whitelist/golden support | Commit 546220d (2026-02-10) | Governance infrastructure exists |
| No boundary testing | Phase 15 adds verify-boundary | Current phase | Completes VALD-03, VALD-04 |

**Deprecated/outdated:**
- Phase 12 verification report line numbers (176-178, 195-198): Network fallback already removed, these line numbers no longer apply
- "9 parity failures": Outdated count, commit 54e41df resolved 8 of them

## Milestone Gap Analysis

**Audit of phases 10-14 for unresolved gaps blocking milestone v1.1 completion:**

### Phase 10: Fresh Baseline Migrations
**Status:** PASSED 5/5 must-haves
**Gaps:** None
**Blockers:** None

### Phase 11: Parser & Importer Updates
**Status:** PASSED 6/6 must-haves
**Gaps:** None
**Blockers:** None

### Phase 12: Compose Generator Parity
**Status:** GAPS FOUND 6/8 must-haves
**Gaps identified in verification:**
1. **Network fallback (stack.go lines 176-178, 195-198):** ✅ RESOLVED — fallback removed from stack.go, no longer present in current code
2. **9 parity failures (3 port conflicts, 1 healthcheck, 4 missing files, 1 dependency):** ⚠️ PARTIALLY RESOLVED — commit 54e41df fixed 8 of 9, likely 1 failure remains

**Remaining work for Phase 15:** Run verify-parity to identify remaining failure(s), fix or whitelist with documented reason

### Phase 13: Import Scalability
**Status:** PASSED 5/5 must-haves
**Gaps:** None
**Deferred human verification items (all addressed by Phase 15):**
1. Large payload import test (256MB boundary) → VALD-03 (verify-boundary test 1)
2. Over-limit rejection test (>256MB) → VALD-04 (verify-boundary test 2)
3. Idempotent re-import test → Out of scope for Phase 15 (not a VALD requirement)
4. Non-import endpoint 10MB cap test → Out of scope for Phase 15 (not a VALD requirement)

**Phase 15 addresses:** Items 1-2 via verify-boundary command

### Phase 14: Dashboard Updates
**Status:** PASSED 13/13 must-haves
**Gaps:** None
**Blockers:** None

### Requirements Traceability
All SCHM, PARS, GENR, IMPT, DASH requirements marked satisfied in phase verification reports. Only VALD requirements remain:

| Requirement | Status | Phase 15 Work |
|-------------|--------|---------------|
| VALD-01 | Pending | Run verify-parity, fix remaining failure(s), achieve exit 0 |
| VALD-02 | Pending | Verify golden services pass (likely already passing) |
| VALD-03 | Pending | Build verify-boundary, test 200MB accepts |
| VALD-04 | Pending | Test 300MB rejects with HTTP 413 |

**Critical finding:** Phase 12's 2 gaps have been addressed post-verification. Network fallback is gone, 8 of 9 parity failures fixed. Phase 15 must confirm the final state and close the last gap(s).

## Open Questions

1. **Current parity failure count**
   - What we know: Phase 12 had 9 failures, commit 54e41df fixed 8
   - What's unclear: Which 1 failure remains? Or are all 9 fixed now?
   - Recommendation: Run verify-parity immediately to get current state. If all pass, Phase 15 is trivial (just add --json flag and boundary tests). If 1-2 fail, triage and fix.

2. **Whitelist population**
   - What we know: whitelist.json is empty (commit 546220d created it)
   - What's unclear: Will any services need whitelisting, or will all 173 pass cleanly?
   - Recommendation: After running verify-parity, if failures are legitimate expected differences (e.g., multi-service compose files where service is secondary), whitelist with clear reason. If failures are bugs, fix the generator/importer.

3. **API server availability for boundary tests**
   - What we know: verify-boundary requires running API on localhost:8550
   - What's unclear: Is docker-compose.yml API container stable enough for CI/automated testing?
   - Recommendation: Manual test first, then consider CI integration if stable

## Sources

### Primary (HIGH confidence)
- api/cmd/verify-parity/main.go (821 lines, read directly)
- api/cmd/verify-parity/golden-services.json (read directly)
- api/cmd/verify-parity/whitelist.json (read directly)
- api/internal/api/handlers/stack_import.go (read directly)
- api/internal/api/routes.go (read directly)
- api/internal/compose/stack.go (read directly — verified network fallback removed)
- .planning/phases/10-fresh-baseline-migrations/10-VERIFICATION.md (read directly)
- .planning/phases/11-parser-importer-updates/11-VERIFICATION.md (read directly)
- .planning/phases/12-compose-generator-parity/12-VERIFICATION.md (read directly)
- .planning/phases/13-import-scalability/13-VERIFICATION.md (read directly)
- .planning/phases/14-dashboard-updates/14-VERIFICATION.md (read directly)
- Git commit history (546220d, 54e41df, 454321c — examined via git show)

### Secondary (MEDIUM confidence)
- [Buffer-less Multipart POST in Golang](https://blog.depa.do/post/bufferless-multipart-post-in-go) — io.Pipe pattern for streaming multipart
- [Go stdlib net/http/request_test.go](https://github.com/golang/go/blob/master/src/net/http/request_test.go) — multipart testing patterns
- [Go stdlib mime/multipart/multipart_test.go](https://go.dev/src/mime/multipart/multipart_test.go) — multipart writer/reader examples

### Tertiary (LOW confidence)
- None used — all findings verified against actual codebase

## Metadata

**Confidence breakdown:**
- verify-parity current state: HIGH — read source directly, examined commits
- Boundary testing patterns: HIGH — Go stdlib patterns, verified against existing stack_import.go
- Milestone gap analysis: HIGH — all 5 phase verification reports read, commit history examined
- Parity failure count: MEDIUM — know 54e41df fixed 8 of 9, but haven't run tool to confirm current state

**Research date:** 2026-02-10
**Valid until:** 2026-03-10 (30 days — stable domain, v1.1 milestone close, no rapid API changes expected)
