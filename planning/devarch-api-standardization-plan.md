# DevArch API Standardization Plan

## Objective
Stabilize DevArch's Go API contract and transport behavior without changing core product behavior, so stack/instance flows stay intact while reducing drift and operational surprises.

## Scope
Focus only on:
- API contract definition for stack/instance endpoints currently used by dashboard
- Unified JSON error envelope
- Middleware correctness fixes (rate limit keying and cache header formatting)
- Dashboard consumption of generated API types for stack/instance flows

Out of scope:
- Large architecture rewrites
- Moving all handlers into service/store layers
- Broad UI redesign

## Current Findings
- Router surface is centralized and extensive in `api/internal/api/routes.go`.
- Handler files mix transport, SQL, and business logic (`stack.go`, `instance.go`, `service.go`).
- Dashboard uses hand-maintained types in `dashboard/src/types/api.ts`, which can drift from backend responses.
- Rate limiting currently keys by `RemoteAddr` (includes port), causing unstable buckets.
- Cache-Control helper builds `max-age` from duration string format, not integer seconds.

## Deliverables
1. OpenAPI contract for stack/instance endpoints used by dashboard.
2. Shared response helper for JSON success/error output.
3. Standardized error format across stack/instance handlers:
   - `code` (stable machine value)
   - `message` (human readable)
   - `details` (optional)
   - `request_id` (from middleware request id)
4. Middleware fixes:
   - client IP extraction from `X-Forwarded-For`/`X-Real-IP`/remote host
   - rate limit keyed by stable client IP
   - valid `Cache-Control: max-age=<seconds>`
5. Generated TypeScript API types/client for stacks/instances and query hook migration.

## Implementation Plan

### Step 1: Contract Baseline
- Add OpenAPI spec file under API directory.
- Define schemas for Stack, Instance, Plan, ApplyResult, and ErrorResponse.
- Capture all stack/instance routes used by dashboard queries/mutations first.

### Step 2: Response Standardization
- Add small HTTP helper package with:
  - `WriteJSON(w, status, payload)`
  - `WriteError(w, status, code, message, details, requestID)`
- Replace ad-hoc `http.Error` usage in stack/instance handlers with helper.
- Preserve current success payload shapes to avoid frontend breakage.

### Step 3: Middleware Hardening
- Fix limiter client identity extraction:
  - first IP from `X-Forwarded-For`
  - fallback `X-Real-IP`
  - fallback remote host parsed from `RemoteAddr`
- Ensure cache middleware writes `max-age` integer seconds.

### Step 4: Dashboard Contract Alignment
- Generate TypeScript types/client from OpenAPI.
- Update `stacks` and `instances` query modules to use generated types.
- Keep existing query keys/hook names to avoid route-level UI churn.
- Remove duplicated manual API contract types where replaced.

### Step 5: Verification
- API tests:
  - middleware extraction and limiter behavior
  - stack/instance error envelope shape
- Full checks:
  - `go test ./...` in `api/`
  - strict TypeScript build in `dashboard/`
- Manual smoke:
  - stack list/get/create/update/delete
  - instance list/get/create/update/delete
  - expected JSON error envelope on not found/conflict cases

## Success Criteria
- OpenAPI exists and covers stack/instance endpoints used by dashboard.
- Stack/instance errors are consistently JSON with stable `code` and `message`.
- Rate limiting no longer fragments by source port.
- Dashboard stacks/instances compile against generated API types.
- No regressions in existing stack/instance user flows.

## Risks and Controls
- Risk: contract mismatch during migration.
  - Control: migrate stacks/instances first, preserve success payload shapes.
- Risk: generated types forcing broad frontend changes.
  - Control: scope conversion to stack/instance modules only initially.
- Risk: hidden dependencies on plaintext errors.
  - Control: add explicit tests for error envelope and status codes.

## Recommended Execution Order
1. OpenAPI draft for stacks/instances
2. Response helper + handler conversion
3. Middleware fixes + tests
4. TS type generation + dashboard hook updates
5. Build/test/smoke verification
