# Phase 0 — pi Execution Layer + Rewrite Charter

## Goal
Create the repo-local pi operating layer and rewrite charter so the rest of V2 can be executed in small controlled packets.

## Acceptance

- project-local pi settings and extensions are committed
- project-local agents and prompts are usable
- rewrite charter and first ADRs exist
- the team can run read-only planning and implementation workflows consistently

## Recommended sequence

Run these packets in order. `P0-DOC-001` can overlap with the tail of `P0-PI-001` if agent definitions are already stable.

---

### P0-PI-001 — Bootstrap project-local pi layer
- **Owner:** orchestrator
- **Depends on:** none
- **Goal:** add the minimal repo-local pi execution layer
- **Tasks:**
  1. Commit `.pi/settings.json` and enable project-local prompts and skills.
  2. Add `.pi/extensions/subagent`, `.pi/extensions/plan-mode`, and `.pi/extensions/worklog`.
  3. Verify project-local pi resources load cleanly.
- **Validation:**
  - inspect `.pi/` layout
  - reload pi locally without extension errors
- **Done when:**
  - repo-local pi resources exist
  - subagent and plan-mode behaviors are available
  - the team has one documented way to run V2 packets

### P0-AGT-001 — Define core project-local agents
- **Owner:** architect
- **Depends on:** P0-PI-001
- **Goal:** create the role prompts for orchestration, planning, implementation, review, and verification
- **Tasks:**
  1. Add prompts for `orchestrator`, `scout`, `architect`, `planner`, `reviewer`, `verifier`, and `scribe`.
  2. Add surgeon prompts for engine, catalog, runtime, API, UI, and import domains.
  3. Ensure each agent includes ownership, allowed tools, validation rules, and handoff format.
- **Validation:**
  - manual prompt review against the spec
  - one dry-run subagent invocation per role family
- **Done when:**
  - all required agent files exist
  - surgeon roles have explicit domain boundaries
  - prompts align with the packet model

### P0-PRM-001 — Commit standard V2 workflow prompts
- **Owner:** scribe
- **Depends on:** P0-AGT-001
- **Goal:** provide reusable workflow entry points for planning, implementation, review, and phase closeout
- **Tasks:**
  1. Add `v2-scout-plan.md`, `v2-implement-slice.md`, `v2-review-slice.md`, and `v2-phase-closeout.md`.
  2. Encode the expected agent chains and outputs in each prompt.
  3. Add one short usage note for each prompt.
- **Validation:**
  - prompt files present and readable
  - prompt text matches intended workflows
- **Done when:**
  - standard workflows can be started without ad hoc prompting

### P0-DOC-001 — Write rewrite charter and first ADRs
- **Owner:** scribe
- **Depends on:** P0-PI-001
- **Goal:** lock the rewrite thesis and source-of-truth rules in docs
- **Tasks:**
  1. Write the rewrite charter RFC.
  2. Add ADRs for manifest-first desired state, runtime-derived state, and thin local API.
  3. Record explicit scope cuts and deferred concerns.
- **Validation:**
  - docs review against the plan spec
- **Done when:**
  - charter and foundational ADRs exist
  - deferred concerns are named, not implied

### P0-RULE-001 — Commit shared V2 skill and working rules
- **Owner:** architect
- **Depends on:** P0-AGT-001, P0-DOC-001
- **Goal:** make rewrite guardrails reusable across packets
- **Tasks:**
  1. Add the `devarch-v2-rules` skill.
  2. Encode manifest-first, packet sizing, boundary, and completion rules.
  3. Cross-check agent prompts against the skill.
- **Validation:**
  - skill file exists and matches charter/ADRs
- **Done when:**
  - agents can apply one shared V2 rule source

## Parallel-safe packets

After `P0-PI-001`, these can run in parallel if they do not edit the same files:

- `P0-AGT-001`
- `P0-DOC-001`

## Handoff to Phase 1

Phase 1 starts only after:

- project-local pi workflows are usable
- charter and ADRs are committed
- the V2 skill exists and is referenced by operators
