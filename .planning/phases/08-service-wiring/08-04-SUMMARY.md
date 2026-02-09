---
phase: 08-service-wiring
plan: 04
subsystem: dashboard
tags: [wiring-ui, wire-crud, tanstack-query, radix-ui]
dependency_graph:
  requires: [wire-management-api, stack-detail-tabs]
  provides: [wiring-tab-ui, create-wire-dialog, wire-query-hooks]
  affects: [stack-detail-page]
tech_stack:
  added: [dashboard/src/components/stacks/wiring-tab.tsx, dashboard/src/components/stacks/create-wire-dialog.tsx]
  patterns: [tanstack-query-hooks, radix-dialog, table-ui]
key_files:
  created:
    - dashboard/src/components/stacks/wiring-tab.tsx
    - dashboard/src/components/stacks/create-wire-dialog.tsx
  modified:
    - dashboard/src/features/stacks/queries.ts
    - dashboard/src/routes/stacks/$name.tsx
decisions:
  - Wire query hooks follow existing stack query patterns with 30s refetch interval
  - Wiring tab positioned between Compose and Deploy tabs
  - Create wire dialog pre-populated from unresolved contracts with available providers
  - Active wires table shows consumer/provider/contract with source badges (auto vs explicit)
  - Unresolved contracts displayed with amber warning styling
metrics:
  duration_seconds: 257
  completed_at: "2026-02-09T00:10:18Z"
---

# Phase 08 Plan 04: Dashboard Wiring Tab & Wire Management UI

**One-liner:** Wiring tab with active wires table, unresolved contracts list, and create-wire dialog for explicit wiring.

## What Was Built

### Wire Query Hooks (`dashboard/src/features/stacks/queries.ts`)

Added five TanStack Query hooks for wire management:

**useStackWires(name: string):**
- GET `/stacks/{name}/wires`
- Returns `{ wires: Wire[], unresolved: UnresolvedContract[] }`
- 30-second refetch interval
- queryKey: `['stacks', name, 'wires']`

**useResolveWires(name: string):**
- POST `/stacks/{name}/wires/resolve`
- Triggers auto-wire resolution
- Toast: "Resolved N wires" or "No changes"
- Invalidates wires query on success

**useCreateWire(name: string):**
- POST `/stacks/{name}/wires`
- Body: `{consumer_instance_id, provider_instance_id, import_contract_name}`
- Toast: "Wire created"
- Invalidates wires query

**useDeleteWire(name: string):**
- DELETE `/stacks/{name}/wires/{wireId}`
- Toast: "Wire disconnected"
- Invalidates wires query

**useCleanupOrphanedWires(name: string):**
- POST `/stacks/{name}/wires/cleanup`
- Returns `{deleted: N}`
- Toast: "Cleaned up N orphaned wires"

Added TypeScript types `UnresolvedContract` and `WireListResponse` at top of file.

### Wiring Tab Component (`dashboard/src/components/stacks/wiring-tab.tsx`)

Full-featured wiring UI with three sections:

**Header:**
- "Resolve" button (calls `useResolveWires`)
- "Add Wire" button (opens `CreateWireDialog`)

**Active Wires Table:**
- Columns: Consumer, Arrow (→), Provider, Contract, Source, Actions
- Consumer/provider shown as instance names in monospace font
- Contract displays name + type badge
- Source badge: "explicit" (blue) or "auto" (gray/muted)
- Disconnect button (Unplug icon) calls `useDeleteWire`
- Empty state: "No active wires" centered text

**Unresolved Contracts Section:**
- Only shown if `unresolved.length > 0`
- Card with amber left border (warning styling)
- Each contract shows:
  - Instance name (monospace, bold)
  - Contract name + type badge
  - Required badge (red) if `required === true`
  - Reason: "No provider available" or "N providers available"
- Amber background for cards

**Empty State:**
- Shown when no wires and no unresolved contracts
- Centered card with Cable icon
- Message: "No wiring needed — This stack has no import contracts"

### Create Wire Dialog (`dashboard/src/components/stacks/create-wire-dialog.tsx`)

Dialog for creating explicit wires:

**Form Fields:**
1. Consumer & Contract dropdown:
   - Lists unresolved contracts with available providers
   - Format: `{instance} → {contract_name}`
   - Shows contract type + provider count in sub-text
   - Empty state: "No unresolved contracts with available providers"

2. Provider Instance dropdown (conditional):
   - Only shown after selecting contract
   - Lists `available_providers` from selected contract
   - Shows provider instance names

**Behavior:**
- Resets state on dialog close (`handleOpenChange`)
- Clears provider selection when contract changes (`handleContractChange`)
- Submit calls `useCreateWire` with `{consumer_instance_id, provider_instance_id, import_contract_name}`
- Closes dialog on success
- Cancel button and X close icon both reset state

**Props:**
- `stackName: string` — stack identifier
- `open: boolean` — dialog visibility
- `onOpenChange: (open: boolean) => void` — visibility handler
- `unresolved: UnresolvedContract[]` — pre-populated from parent

### Stack Detail Page Updates (`dashboard/src/routes/stacks/$name.tsx`)

Registered Wiring tab:

- Added `'wiring'` to `stackTabs` array
- Added `'wiring'` to route search validation schema
- Added tab item: `{ value: 'wiring', label: 'Wiring' }`
- Positioned between Compose and Deploy tabs (as specified in locked decision)
- Added `TabsContent` with `<WiringTab stackName={name} />`
- Imported `WiringTab` from `@/components/stacks/wiring-tab`

## Deviations from Plan

None — plan executed exactly as written.

## Implementation Notes

- Wire query hooks follow existing stacks query patterns (30s refetch, toast on mutations, query invalidation)
- Active wires table uses Radix UI Table component (consistent with project patterns)
- Source badges use variant "default" for explicit (blue) and "secondary" for auto (gray)
- Unresolved contracts use amber warning color (`border-l-amber-500`, `bg-amber-50 dark:bg-amber-950/20`)
- Create wire dialog uses plain `label` HTML elements with `text-sm font-medium` class (matches existing dialog patterns)
- Dialog state management avoids `useEffect` setState anti-pattern (uses handler functions instead)
- Wiring tab positioned between Compose and Deploy per Phase 08 locked decision

## Verification Results

- TypeScript compiles without errors: `npx tsc --noEmit` passes
- Lint passes: `npm run lint` passes
- All query hooks imported and used in wiring-tab component
- WiringTab registered in stack detail route file
- Create wire dialog passes unresolved contracts as prop

## Self-Check: PASSED

Created files:
- FOUND: dashboard/src/components/stacks/wiring-tab.tsx
- FOUND: dashboard/src/components/stacks/create-wire-dialog.tsx

Modified files:
- FOUND: dashboard/src/features/stacks/queries.ts (added 5 wire query hooks)
- FOUND: dashboard/src/routes/stacks/$name.tsx (registered Wiring tab)

Commits:
- FOUND: 1df494d0 (Task 1: wire query hooks and wiring tab component)
- FOUND: ee61ba2f (Task 2: create wire dialog for explicit wiring)

Key links verified:
- FOUND: wiring-tab imports useStackWires, useCreateWire, useDeleteWire, useResolveWires
- FOUND: WiringTab component imported and used in $name.tsx route

## Next Steps

Plan 08-04 is complete. Phase 08 (Service Wiring) is now complete. Next phase is Phase 09 (Secret Management).
