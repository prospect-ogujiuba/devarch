# Phase 5: SLA, notifications, and reporting

Created: 2026-08-23
Purpose: Add operational service-management behavior after the core workflow is stable.

## Goal

Track service targets, escalate risks, notify participants, and provide useful operational reporting.

## Scope

- SLA target selection by priority/category.
- Response and resolution deadlines with pause states.
- Scheduled warning/breach processing using WordPress scheduling.
- Email and dashboard notifications with deduplication.
- Open queue, aging, SLA breach, workload, resolution-time, category, and CSV reports.

## Outputs

- Scaffold-derived Notification and SavedView resources if needed.
- Application-owned scheduler callbacks and query/report collaborators.
- Dashboard/report views and tests.

## Acceptance criteria

- SLA timers pause/resume deterministically.
- Warning and breach notifications are idempotent.
- Reports respect authorization and bounded date ranges.
- CSV export neutralizes spreadsheet-formula injection.

## Verification

- Clock-controlled SLA tests.
- Repeated scheduler execution tests.
- Reporting authorization and aggregate correctness tests.
