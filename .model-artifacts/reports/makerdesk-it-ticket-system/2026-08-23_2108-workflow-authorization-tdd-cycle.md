# workflow-authorization-tdd-cycle

Created: 2026-08-23T21:08:01.840Z
Purpose: Record Phase 3 workflow, authorization, audit, and optimistic concurrency evidence.

# MakerDesk Workflow and Authorization TDD Cycle

Date: 2026-08-23

## Behavior

MakerDesk accepts only defined ticket transitions, produces stable human-readable numbers, enforces role/ownership policy, records accepted mutations, and rejects stale writes.

## Test level

Pure service/policy contracts plus WordPress/MariaDB integration.

## Red

- Added `tests/workflow-contract.php` before the workflow, number, and capability services existed.
- Result: failed because `TicketWorkflowService.php` did not exist.

## Green

- Used app Galaxy `make:service` for TicketWorkflowService, TicketNumberService, and MakerDeskCapabilityService, then customized the generated services.
- Used app Galaxy `make:fields` for TicketTransitionFields and TicketAssignmentFields, then customized validation.
- Customized generated ticket/activity controllers and policies.
- Workflow, number, policy, and capability contracts passed.

## Refactor

- Fixed workflow represented by a small adjacency map.
- Number allocation uses an ID-derived final number and random transaction-local placeholder instead of a race-prone max query.
- Mutation handlers share the versioned `id + version` update invariant and write activity in the same transaction.
- Ticket activity is immutable; requester access is ownership-limited and internal notes are agent-only.

## Verification

- Schema contract passed.
- Workflow contract passed.
- Ticket role/ownership policy contract passed.
- MariaDB contract proved current update affects one row, stale update affects zero, and accepted transition has an activity record.
- Six policies registered in TypeRocket.
- Role installation: requester 6 capabilities, agent 19, manager 42; administrator receives MakerDesk capabilities.
- PHP lint: 65/65 passed.
- TypeRocket and MakerMaker core worktrees remain clean.

## Follow-up

Phase 4 owns route registration, version-bearing forms, requester pages, agent queues, filters, attachments, and browser smoke tests.
