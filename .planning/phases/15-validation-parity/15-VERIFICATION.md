---
status: passed
phase: 15
name: validation-parity
verified: 2026-02-10
---

# Phase 15: Validation & Parity — Verification

## Goal
Legacy parity verified across all services, boundary cases tested.

## Must-Haves Verification

### VALD-01: Full Legacy Parity
**Status: PASSED**
- verify-parity exits 0 against full service catalog
- 172/173 services pass, 1 whitelisted (zammad — external deps not in catalog)
- Covers: ports, volumes, env vars, env files, deps, healthchecks, labels, networks, config mappings

### VALD-02: Golden Service Parity
**Status: PASSED**
- All 7 golden services pass with zero whitelisted exceptions:
  php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api
- No golden service appears in whitelist.json
- --json output confirms `golden=true, pass=true` for all 7

### VALD-03: 200MB Import Accepted
**Status: PASSED**
- 200MB streaming payload generated via io.Pipe (no memory buffering)
- Response: not 413 (payload accepted by size limit)
- Verified via verify-boundary tool

### VALD-04: 300MB Import Rejected
**Status: PASSED**
- 300MB payload rejected with HTTP 413
- Response body contains: error message, max_bytes=268435456, received_bytes > 0
- Verified via verify-boundary tool

## Additional Fixes During Execution

1. **Importer warnings polluting stdout** — fixed to stderr, preserving clean JSON output
2. **Global MaxBodySize(10MB) shadowing route-specific 256MB** — restructured middleware scoping in routes.go; import route now registered outside /api/v1 group

## Artifacts

| Tool | Purpose | Status |
|------|---------|--------|
| verify-parity | Full catalog parity proof | exits 0 |
| verify-boundary | Import size boundary proof | exits 0 |

## Score: 4/4 must-haves verified
