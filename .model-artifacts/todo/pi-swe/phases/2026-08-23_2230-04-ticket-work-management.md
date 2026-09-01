# 04-ticket-work-management

Created: 2026-08-23T22:30:27.732Z
Purpose: Add spreadsheet work-management fields without weakening existing ITSM workflow

# Phase 4 — Ticket work management

Created: 2026-08-23
Purpose: Add the workbook's planning and recurring-work features to the existing robust MakerDesk ticket lifecycle.

## Goal

Support service requests, projects, recurring administrative work, parent/child tasks, due dates, and effort tracking in one auditable ticket model.

## Scope

- Preserve existing number, subject, description, requester/assignee WordPress links, support team, category, asset, SLA, status, priority, impact, urgency, source, due targets, optimistic version, actor stamps, and soft delete.
- Add work type/frequency, optional parent ticket, planned start, user due date, completed date semantics, estimated minutes, actual minutes, and structured completion/work notes.
- Use Employee relationships where useful while retaining WordPress user IDs as authentication/actor and portal ownership identities; document precedence when both exist.
- Parent/child validation prevents self-links and cycles.
- Recurring templates/schedules generate daily, weekly, biweekly, monthly, on-demand, project, administrative, and organization-specific tasks without duplicating active occurrences.
- Archive sheet behavior becomes saved filters for resolved/closed/cancelled/deleted records with retention-safe restore where allowed.
- Ticket form tabs: Request, Classification, Planning, Relationships, Activity, System.
- Queue filters: work type, priority, status, requester, assignee, employee, department, asset, parent, start/due/completed range, overdue, SLA state.

## Outputs

- Additive ticket migration, field rules, relationship/service changes, and recurring-task resource/job if approved by the schema phase.
- Updated list/detail/admin/portal forms.
- Activity events for planning, hierarchy, effort, recurrence, and completion changes.
- Reports for estimated versus actual effort, overdue work, recurring completion, and parent progress.

## Acceptance criteria

- Every Tickets/Ticket Form column is represented.
- Existing transition and SLA rules remain authoritative; completed date is synchronized with resolved/closed transitions rather than freely contradictory.
- Notes are activity entries when historical/auditable; a current summary may exist but cannot replace history.
- Parent cycles and invalid archived-parent operations are rejected.
- Effort values are non-negative and reported consistently in hours.
- Recurring generation is idempotent and actor/source attribution is explicit.

## Verification

- Existing workflow, policy, portal, SLA, and report contracts stay green.
- New hierarchy, date, effort, archive, recurrence, and concurrency contracts pass.
- Form/list contracts cover layout and filters.

## Non-goals

No external calendar or third-party ITSM synchronization yet.
