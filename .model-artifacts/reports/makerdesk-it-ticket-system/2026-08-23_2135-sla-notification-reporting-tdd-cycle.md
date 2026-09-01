# sla-notification-reporting-tdd-cycle

Created: 2026-08-23T21:35:40.896Z
Purpose: Record Phase 5 scaffold lineage and verification.

# MakerDesk SLA, Notification, and Reporting TDD Cycle

Date: 2026-08-23

## Behavior

MakerDesk calculates business-hour SLAs, pauses/resumes targets deterministically, performs idempotent warning/breach sweeps, notifies authorized users, and exposes bounded authorized reports with safe CSV export.

## Red

`tests/sla-report-contract.php` failed because the app-Galaxy-generated `SlaClockService` had no `addBusinessMinutes()` implementation.

## Green

- MakerMaker resource scaffolds generated Notification, SavedView, and EscalationRule, each with migration, views, factory, and tests.
- App Galaxy generated the SLA runtime migration, SlaClockService, MakerDeskNotificationService, TicketReportService, MakerDeskSlaSweepJob, and MakerDeskSlaSweepCommand.
- App Galaxy generated the command/job files but attempted to register them in TypeRocket core config; those two core config changes were immediately reverted. The command is registered application-side through the supported `typerocket_galaxy_commands` filter, and scheduling is application-owned WordPress cron.
- Added default weekday business hours, category/priority SLA selection, pause/resume extension, sweep markers, deduplicated dashboard/email notifications, escalation recipients, saved queue filters, aggregate reports, and CSV export.

## Refactor

- Business-time arithmetic advances by daily schedule windows rather than minute-by-minute loops.
- Warning/breach marker columns and a unique notification dedupe key make repeated sweeps idempotent.
- Reports cap date ranges to 366 days and exports to 10,000 rows.
- CSV values beginning with formula/control prefixes receive an apostrophe guard.

## Verification

- Pure SLA/report contract passed, including weekend and pause/resume cases.
- Database integration: first sweep processed four events, wrote eight channel notifications and four audit records; second sweep wrote none.
- Priority selected the active SLA policy and produced the expected business-hour deadline.
- Report aggregate and CSV formula neutralization checks passed.
- Requester notification HTTP 200; requester reports HTTP 403; manager reports HTTP 200.
- Phase 5 migration rollback removed three tables/seven columns; reapply restored them.
- `makerdesk:sla-sweep` runs successfully and is visible through site, MakerMaker, and app Galaxy launchers.
- Cron is scheduled; nine TypeRocket policies are registered.
- Full suite passed; PHP lint 103/103.
- TypeRocket and MakerMaker core worktrees are clean.
