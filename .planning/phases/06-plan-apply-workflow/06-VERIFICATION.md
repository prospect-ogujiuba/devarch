---
phase: 06-plan-apply-workflow
verified: 2026-02-07T21:30:00Z
status: passed
score: 19/19 must-haves verified
---

# Phase 6: Plan/Apply Workflow Verification Report

**Phase Goal:** Users preview changes before applying (Terraform-style safety), with advisory locking preventing concurrent modifications

**Verified:** 2026-02-07T21:30:00Z

**Status:** passed

**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Plan types represent add/modify/remove actions with per-field detail | ✓ VERIFIED | types.go defines Action constants, Change struct with Fields map for per-field detail |
| 2 | Differ compares desired instances against running containers and returns structured changes | ✓ VERIFIED | differ.go ComputeDiff stateless function handles add/modify/remove detection |
| 3 | Staleness token is deterministic: same inputs always produce same hash | ✓ VERIFIED | staleness.go GenerateToken sorts instances by ID, uses RFC3339Nano + SHA256 |
| 4 | Staleness validation rejects when any timestamp has changed | ✓ VERIFIED | ValidateToken queries current state, regenerates token, compares, returns ErrStalePlan on mismatch |
| 5 | GET /stacks/{name}/plan returns structured diff with add/modify/remove changes | ✓ VERIFIED | stack_plan.go handler queries DB + runtime, calls ComputeDiff, returns Plan JSON |
| 6 | POST /stacks/{name}/apply acquires advisory lock, validates staleness, executes flow | ✓ VERIFIED | stack_apply.go uses pg_try_advisory_lock, ValidateToken, sequential flow |
| 7 | Apply rejects concurrent operations with HTTP 409 | ✓ VERIFIED | Line 55-58 in stack_apply.go: returns 409 if lock not acquired |
| 8 | Apply rejects stale plans with HTTP 409 | ✓ VERIFIED | Lines 66-70 in stack_apply.go: checks errors.Is(err, plan.ErrStalePlan), returns 409 |
| 9 | Apply flow: lock → ensure network → materialize configs → compose up | ✓ VERIFIED | stack_apply.go lines 54-142: lock (55), network (86), materialize (105), compose (138) |
| 10 | Error handling: network fail = unlock+error, config fail = cleanup+unlock+error, compose fail = leave configs+unlock+error | ✓ VERIFIED | Defer unlock at line 60, config cleanup at 107/115, no cleanup on compose fail |
| 11 | User can click Generate Plan to see structured diff preview | ✓ VERIFIED | Deploy tab line 493: Generate Plan button calls useGeneratePlan, sets currentPlan |
| 12 | Diff shows adds in green, modifications in yellow, removals in red | ✓ VERIFIED | Lines 530-532: border-green-500, border-yellow-500, border-red-500 based on action |
| 13 | User can click Apply to execute the plan | ✓ VERIFIED | Line 568: Apply button calls useApplyPlan with token |
| 14 | Apply errors (409 stale, 409 locked, 500) show appropriate toast messages | ✓ VERIFIED | queries.ts lines 267-271: checks status 409, shows stale/locked message |
| 15 | Deploy tab appears alongside Instances and Compose tabs | ✓ VERIFIED | Line 335: TabsTrigger value="deploy" in TabsList |
| 16 | After successful apply, plan clears and stack data refreshes | ✓ VERIFIED | Line 568 onSuccess clears plan, queries.ts 260-264 invalidates caches |
| 17 | Plan handler queries DB and runtime, computes diff, generates staleness token | ✓ VERIFIED | stack_plan.go lines 21-95: queries stack + instances + containers, calls ComputeDiff + GenerateToken |
| 18 | Apply handler runs 3-step flow with advisory lock | ✓ VERIFIED | stack_apply.go: lock acquisition, network ensure, config materialize, compose up |
| 19 | Dashboard types match API response structure | ✓ VERIFIED | api.ts lines 492-518: StackPlan, PlanChange, PlanFieldChange match Go types |

**Score:** 19/19 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `api/internal/plan/types.go` | Plan, Change, FieldChange, Action types | ✓ VERIFIED | 35 lines, exports Plan/Change/FieldChange/Action, JSON tags present |
| `api/internal/plan/differ.go` | Diff computation between desired and runtime state | ✓ VERIFIED | 76 lines, exports ComputeDiff, DesiredInstance, stateless function |
| `api/internal/plan/staleness.go` | Staleness token generation and validation | ✓ VERIFIED | 78 lines, exports GenerateToken/ValidateToken/ErrStalePlan/InstanceTimestamp |
| `api/internal/api/handlers/stack_plan.go` | Plan endpoint handler | ✓ VERIFIED | 109 lines, Plan method on StackHandler, queries DB+runtime, returns JSON |
| `api/internal/api/handlers/stack_apply.go` | Apply endpoint handler with advisory lock | ✓ VERIFIED | 150 lines, Apply method with pg_try_advisory_lock, staleness check, flow execution |
| `api/internal/api/routes.go` | Route registration for plan and apply | ✓ VERIFIED | Lines 154-155: r.Get("/plan"), r.Post("/apply") under stack group |
| `dashboard/src/types/api.ts` | StackPlan, PlanChange, PlanFieldChange types | ✓ VERIFIED | Lines 492-518: types present with correct structure |
| `dashboard/src/features/stacks/queries.ts` | useGeneratePlan and useApplyPlan hooks | ✓ VERIFIED | Lines 240-274: both hooks present with correct API calls and error handling |
| `dashboard/src/routes/stacks/$name.tsx` | Deploy tab with diff visualization and apply button | ✓ VERIFIED | Lines 335, 486-602: Deploy tab with plan generation, diff display, apply execution |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| differ.go | types.go | imports plan.Change type | ✓ WIRED | Line 12 imports plan, line 29 uses plan.Change |
| staleness.go | database/sql | queries stacks and instances updated_at | ✓ WIRED | Lines 37-77: SQL queries for timestamps |
| stack_plan.go | plan package | imports ComputeDiff and GenerateToken | ✓ WIRED | Line 12 imports plan, lines 94-95 call functions |
| stack_apply.go | plan package | imports ValidateToken | ✓ WIRED | Line 16 imports plan, line 66 calls ValidateToken |
| stack_apply.go | compose package | uses Generator for YAML and config materialization | ✓ WIRED | Lines 15,97,105,112: imports and calls MaterializeStackConfigs, GenerateStack |
| routes.go | handlers | registers plan and apply routes | ✓ WIRED | Lines 154-155: stackHandler.Plan, stackHandler.Apply |
| $name.tsx | queries.ts | imports useGeneratePlan and useApplyPlan | ✓ WIRED | Line 14 imports, lines 147-148 use hooks |
| queries.ts | /api/v1/stacks/{name}/plan | fetch calls to plan endpoint | ✓ WIRED | Line 243: api.get to /stacks/${name}/plan |
| queries.ts | /api/v1/stacks/{name}/apply | fetch calls to apply endpoint | ✓ WIRED | Line 256: api.post to /stacks/${name}/apply with token |
| api.ts | queries.ts | StackPlan type used in query return types | ✓ WIRED | Line 3 imports StackPlan, line 243 uses as generic type |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| PLAN-01: Plan endpoint returns structured preview | ✓ SATISFIED | None — stack_plan.go returns Change array with action field |
| PLAN-02: Apply endpoint with advisory lock per stack | ✓ SATISFIED | None — pg_try_advisory_lock(stackID) at line 55 |
| PLAN-03: Apply flow sequence | ✓ SATISFIED | None — lock → network (86) → materialize (105) → compose (138) |
| PLAN-04: Plan staleness detection | ✓ SATISFIED | None — ValidateToken checks timestamps, returns ErrStalePlan |
| PLAN-05: Structured diff output | ✓ SATISFIED | None — Plan.Changes with Action + InstanceID + Fields |

### Anti-Patterns Found

None detected. All files substantive, no TODO/FIXME markers, no placeholder content, no stub patterns.

### Human Verification Required

1. **Visual Diff Rendering**
   - Test: Generate plan with mixed changes (add/modify/remove), view Deploy tab
   - Expected: Green left border for adds, yellow for modifications, red for removals; +/~/- symbols visible
   - Why human: Visual appearance, color rendering in browser

2. **Concurrent Apply Conflict**
   - Test: Open two browser tabs, generate plan in both, try to apply simultaneously
   - Expected: Second apply shows "Stack is being applied by another session" toast
   - Why human: Requires multi-session coordination

3. **Stale Plan Detection**
   - Test: Generate plan, modify instance (enable/disable), try to apply old plan
   - Expected: Apply rejected with "Plan is stale — stack was modified since plan was generated" toast
   - Why human: Requires coordinated DB modifications between plan generation and apply

4. **Full Apply Flow Success**
   - Test: Create stack, add instance, generate plan, apply
   - Expected: Network created, configs materialized to compose/stacks/{stack}, containers started
   - Why human: End-to-end filesystem + runtime verification

5. **Apply Flow Error Handling**
   - Test: Trigger config materialization error (invalid file path), trigger compose error (invalid YAML)
   - Expected: Config fail cleans up compose/stacks/{stack}, compose fail leaves configs for debugging
   - Why human: Requires artificial error injection

---

_Verified: 2026-02-07T21:30:00Z_
_Verifier: Claude (gsd-verifier)_
