---
phase: 15-validation-parity
plan: 01
subsystem: validation
tags: [compose, parity, verification, testing, quality]

# Dependency graph
requires:
  - phase: 12-compose-generator-parity
    provides: verify-parity tool with --json, --strict, whitelist, golden services
provides:
  - 100% parity verification (172 passed, 1 whitelisted)
  - Clean JSON output for programmatic consumption
  - All 7 golden services passing with zero exceptions
  - Validated whitelist governance and strict mode
affects: [deployment, production-readiness]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Stderr for diagnostics, stdout for structured output
    - Whitelist governance for expected differences

key-files:
  created: []
  modified:
    - api/cmd/verify-parity/whitelist.json
    - api/internal/compose/importer.go

key-decisions:
  - "Whitelist zammad dependencies on external services not in catalog"
  - "Send all importer warnings to stderr to preserve clean JSON output"

patterns-established:
  - "Diagnostic messages (warnings, progress) go to stderr"
  - "Structured output (JSON) goes to stdout uncontaminated"

# Metrics
duration: 7m29s
completed: 2026-02-10
---

# Phase 15 Plan 01: Validation Parity Summary

**Full legacy parity achieved: 172 services pass, 1 whitelisted for external dependencies, all 7 golden services clean with validated JSON/strict modes**

## Performance

- **Duration:** 7m29s
- **Started:** 2026-02-10T15:53:04Z
- **Completed:** 2026-02-10T16:00:33Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Achieved 100% parity verification (172/173 passed, 1 whitelisted)
- All 7 golden services (php, python, nginx-proxy-manager, blackbox-exporter, rabbitmq, traefik, devarch-api) pass with zero exceptions
- Fixed JSON output pollution from importer warnings
- Validated --json and --strict modes working correctly
- Whitelist governance prevents golden services from being whitelisted

## Task Commits

Each task was committed atomically:

1. **Task 1: Run verify-parity, fix or whitelist remaining failures** - `a29490f` (chore)
2. **Task 2: Verify golden services individually and validate --json/--strict output** - `09ec1dd` (fix)

## Files Created/Modified
- `api/cmd/verify-parity/whitelist.json` - Added entry for zammad external dependencies
- `api/internal/compose/importer.go` - Fixed warnings to write to stderr instead of stdout

## Decisions Made

**1. Whitelist zammad external dependencies**
- zammad references zammad-db and zammad-elasticsearch which are not standalone services in our catalog
- These are external dependencies that would be provided at runtime in a different orchestration context
- Whitelisting is appropriate because this is an expected difference, not a bug
- zammad is not a golden service, so whitelisting is permitted

**2. Fix JSON output pollution**
- Importer warnings were going to stdout, contaminating --json output
- Changed all fmt.Printf diagnostic messages to fmt.Fprintf(os.Stderr)
- Ensures clean JSON output for programmatic consumption

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Importer warnings polluting JSON output**
- **Found during:** Task 2 (validating --json output)
- **Issue:** fmt.Printf in importer sent warnings to stdout, making JSON unparseable
- **Fix:** Changed 5 fmt.Printf statements to fmt.Fprintf(os.Stderr) for all warning/info messages
- **Files modified:** api/internal/compose/importer.go
- **Verification:** JSON output now clean, parseable by jq, all validations pass
- **Committed in:** 09ec1dd (separate commit for bug fix)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Bug fix necessary for JSON output correctness. No scope creep.

## Issues Encountered
None - plan execution was straightforward once bug was identified and fixed.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Full legacy parity verified and documented
- verify-parity tool ready for CI/CD integration
- Golden services governance working as designed
- Ready for Phase 15 Plan 02 (dashboard validation integration)

---
*Phase: 15-validation-parity*
*Completed: 2026-02-10*

## Self-Check: PASSED

All files and commits verified:
- ✓ api/cmd/verify-parity/whitelist.json exists
- ✓ api/internal/compose/importer.go exists
- ✓ Commit a29490f exists
- ✓ Commit 09ec1dd exists
