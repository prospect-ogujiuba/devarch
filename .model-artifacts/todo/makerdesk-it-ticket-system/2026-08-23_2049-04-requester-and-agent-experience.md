# Phase 4: Requester and agent experience

Created: 2026-08-23
Purpose: Expose the ticket lifecycle through usable application-owned interfaces.

## Goal

Deliver requester intake/tracking and agent queue/detail workflows using generated views/controllers as the starting point.

## Scope

- Requester create, list, view, reply, attachment, and reopen actions.
- Agent queue, filters, assignment, internal notes, transitions, and resolution.
- Team manager workload view and reassignment.
- Accessible forms, validation feedback, pagination, search, and saved filters.
- WordPress Media integration for attachments.

## Outputs

- Customized generated views and controllers.
- Registered routes/menu entries.
- Minimal application styles/scripts using existing asset pipelines.

## Acceptance criteria

- A requester can create and follow a ticket without wp-admin privileges.
- An agent can process a ticket end to end.
- Internal notes never appear in requester views.
- Queue queries are paginated and use indexed filters.

## Verification

- Controller/integration tests.
- Role-separated browser smoke tests.
- Accessibility and attachment authorization checks.
