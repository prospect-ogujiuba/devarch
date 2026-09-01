# Phase 3: Ticket workflow and authorization

Created: 2026-08-23
Purpose: Make the scaffolded ticket domain safe and behaviorally complete.

## Goal

Implement explicit lifecycle transitions, assignment rules, role capabilities, audit behavior, and validation.

## Scope

- Requester, Agent, Team Manager, and Administrator capabilities.
- New, Triaged, Assigned, In Progress, Pending User, Pending Vendor, Resolved, Closed, Cancelled, and Reopened transitions.
- Deny-by-default resource policies expanded only for required actions.
- Activity/audit creation for mutations and transitions.
- Ticket numbering and optimistic conflict protection appropriate to the existing stack.
- Approved representation: a static status adjacency map; `IT-YYYY-NNNNNN` numbers finalized from the database ID inside the create transaction; versioned updates use `WHERE id = ? AND version = ?` and increment atomically.

## Outputs

- Customized generated policies, controllers, models, and tests.
- Capability installation/cleanup hooks where required.
- Lifecycle and validation services only through an available scaffold or as small application-owned collaborators attached to generated resources.

## Acceptance criteria

- Unauthorized users cannot read or mutate tickets outside policy.
- Invalid transitions are rejected.
- Every accepted transition is auditable.
- Closed tickets are immutable except for authorized reopen behavior.

## Verification

- Behavior-first policy and transition tests.
- Role matrix integration tests.
- Database integration test proving stale version updates affect zero rows.
