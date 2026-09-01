# admin-nesting-sample-data-verification

Created: 2026-08-23T22:01:24.032Z
Purpose: Verification evidence for TypeRocket-native MakerDesk admin nesting and repeatable sample data.

# MakerDesk admin nesting and sample-data verification

Date: 2026-08-23

## Implemented

- Kept the generated Ticket resource index as the single top-level TypeRocket page and branded its WordPress menu title as MakerDesk.
- Nested the other eight generated resource indexes with native `Page::addPage()` parent/child registration.
- Bound every creatable resource index to its own generated add action URL through `Page::addNewButton($addPage->getUrl())`.
- Removed Notification Add New title/admin-bar links because Notifications are system-generated.
- Scaffolded `MakerDeskSeedCommand` with app-level Galaxy, reverted the generator's TypeRocket-core registration edit, and registered the command application-side.
- Added dedicated demo requester/agent/manager users and representative data for all nine resources.
- Added `--reset-passwords` for explicit demo credential rotation.

## Evidence

- Authenticated wp-admin HTML: eight creatable resource indexes each linked to the matching add page; Notification had no create link.
- Runtime contract: native parent relationships, explicit add URLs, sample counts, and demo-user roles passed.
- Seed idempotence: second command run created nothing; stable record counts remained teams=2, SLAs=3, categories=3, assets=2, tickets=6, activities=8, views=2, notifications=1, escalations=2.
- Galaxy registration: `makerdesk:seed` visible through site, MakerMaker, and app launchers.
- Regression: eight contracts passed; PHP lint 105/105.
- Maker audit passed; MakerStarter, MakerBlocks, MakerMaker, and TypeRocket worktrees clean.

## Documentation

Updated the app README and `docs/makerdesk.md` with the native admin hierarchy, seed commands, sample inventory, idempotence behavior, credential rotation, and environment boundary.
