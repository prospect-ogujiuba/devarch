---
phase: 12-compose-generator-parity
verified: 2026-02-09T21:18:15Z
status: gaps_found
score: 6/8 must-haves verified
gaps:
  - truth: "Generator emits networks from service_networks table instead of hardcoded fallback"
    status: failed
    reason: "Stack generator has hardcoded network fallback on lines 176-178 and 195-198"
    artifacts:
      - path: "api/internal/compose/stack.go"
        issue: "Lines 176-178: if len(networksMap) == 0 fallback to netName"
      - path: "api/internal/compose/stack.go"
        issue: "Lines 195-198: if len(cfg.networks) == 0 fallback to []string{netName}"
    missing:
      - "Remove hardcoded network fallback from stack.go lines 176-178"
      - "Remove hardcoded network fallback from stack.go lines 195-198"
      - "Trust DB-sourced networks exclusively per 12-CONTEXT locked decision"
  - truth: "Generated compose for all 166+ services matches legacy output byte-for-byte (modulo whitespace/ordering)"
    status: partial
    reason: "95% parity achieved (164/173 passed), but 9 known failures remain"
    artifacts:
      - path: "api/cmd/verify-parity/main.go"
        issue: "Tool reports 9 failures: 3 port conflicts, 1 healthcheck format, 4 missing files, 1 dependency resolution"
    missing:
      - "Triage and fix 9 parity failures before claiming full 1:1 match"
---

# Phase 12: Compose Generator Parity Verification Report

**Phase Goal:** Generated compose output matches legacy 1:1 using new schema
**Verified:** 2026-02-09T21:18:15Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Generator emits env_file as YAML list from service_env_files table | ✓ VERIFIED | generator.go:289 queries service_env_files, line 145 sets svc.EnvFile = service.EnvFiles |
| 2 | Generator emits networks from service_networks table instead of hardcoded fallback | ✗ FAILED | generator.go uses DB only (lines 77-91), but stack.go has fallback (lines 176-178, 195-198) |
| 3 | Generator emits config mounts as volume strings merged into volumes section | ✓ VERIFIED | generator.go:366-405 loadConfigMounts queries service_config_mounts, line 152 merges into volumes |
| 4 | Stack generator emits env_file, networks, config mounts with instance override support | ⚠️ PARTIAL | stack.go queries all 3 tables (lines 542, 547, 552) but has network fallback issue |
| 5 | Round-trip tool imports all legacy compose files then generates for each and diffs key fields | ✓ VERIFIED | verify-parity/main.go:170 ParseFileAll original, line 216 gen.Generate, line 232 ParseFileAll generated |
| 6 | Diff covers ports, volumes, env vars, env files, deps, healthchecks, labels, networks, config mounts | ✓ VERIFIED | verify-parity compares 14 field groups per plan spec |
| 7 | Tool reports per-service pass/fail with specific field mismatches | ✓ VERIFIED | 569-line tool with detailed diff output, exit code 0/1 |
| 8 | Generated compose for all 166+ services matches legacy output byte-for-byte (modulo whitespace/ordering) | ⚠️ PARTIAL | 95% parity (164/173 passed), 9 failures remain |

**Score:** 6/8 truths verified (75%)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `api/internal/compose/generator.go` | env_file, networks, config mount emission for single-service | ✓ VERIFIED | 410 lines, queries service_env_files (289), service_networks (305), service_config_mounts (368), emits all fields, compiles |
| `api/internal/compose/stack.go` | env_file, networks, config mount emission for stack instances | ⚠️ PARTIAL | Has all queries (1004, 1024, 1045) but includes network fallback logic violating locked decision |
| `api/cmd/verify-parity/main.go` | Import-then-generate round-trip comparison tool | ✓ VERIFIED | 569 lines, uses ParseFileAll and Generator.Generate, compares 14 fields, compiles |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| generator.go | service_env_files table | SQL query in loadServiceRelations | ✓ WIRED | Line 289: SELECT path FROM service_env_files |
| generator.go | service_networks table | SQL query in loadServiceRelations | ✓ WIRED | Line 305: SELECT network_name FROM service_networks |
| generator.go | service_config_mounts table | SQL query in loadConfigMounts | ✓ WIRED | Line 368: SELECT from service_config_mounts with LEFT JOIN service_config_files |
| stack.go | service_env_files table | SQL query in loadEffectiveEnvFiles | ✓ WIRED | Line 1004: SELECT path FROM service_env_files |
| stack.go | service_networks table | SQL query in loadEffectiveNetworks | ⚠️ WIRED_WITH_FALLBACK | Line 1024: SELECT network_name but lines 176-178, 195-198 add fallback |
| stack.go | service_config_mounts table | SQL query in loadEffectiveConfigMounts | ✓ WIRED | Line 1045: SELECT from service_config_mounts with LEFT JOIN |
| verify-parity | compose/parser.go | Uses ParseFileAll to parse both original and generated | ✓ WIRED | Lines 170, 232: compose.ParseFileAll calls |
| verify-parity | compose/generator.go | Uses Generator.Generate to produce compose from DB | ✓ WIRED | Line 216: gen.Generate(&dbService) call |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| GENR-01: Compose generator emits env_files faithfully from DB model | ✓ SATISFIED | None |
| GENR-02: Compose generator emits explicit network attachments from DB model | ✗ BLOCKED | Stack generator has hardcoded network fallback |
| GENR-03: Compose generator emits config mounts with correct provenance from DB model | ✓ SATISFIED | None |
| GENR-04: Generated compose matches legacy output 1:1 for all 166+ services | ⚠️ PARTIAL | 95% parity, 9 failures remain |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| stack.go | 176-178 | Hardcoded network fallback | 🛑 Blocker | Violates "DB-sourced only, no fallback" locked decision from 12-CONTEXT |
| stack.go | 195-198 | Hardcoded network fallback | 🛑 Blocker | Creates dual source of truth for network configuration |

### Gaps Summary

**2 gaps blocking goal achievement:**

1. **Stack generator network fallback violates architecture decision**
   - Issue: Lines 176-178 and 195-198 in stack.go add hardcoded network fallback when DB returns empty
   - Decision violation: 12-CONTEXT explicitly states "No hardcoded fallback to stack default network" and "Importer is sole data path — fallback would create dual source of truth"
   - Impact: Services without networks in DB will get a network anyway, breaking the DB-as-source-of-truth principle
   - Fix: Remove both fallback checks, trust DB exclusively

2. **95% parity not 100% parity**
   - Issue: 9 services fail parity check (3 port conflicts, 1 healthcheck format, 4 missing files, 1 dependency resolution)
   - Success criteria: "Generated compose for all 166+ services matches legacy output byte-for-byte"
   - Impact: Cannot claim full 1:1 parity with 9 known failures
   - Fix: Triage each failure, determine if generator bug or expected difference, document or fix

**Root cause:** Phase 12 Plan 01 task was marked "done" but implementation included fallback logic not specified in plan. Plan explicitly stated "DB-sourced only — read service_networks table, emit exactly what's stored" with "No hardcoded fallback" as locked decision.

---

*Verified: 2026-02-09T21:18:15Z*
*Verifier: Claude (gsd-verifier)*
