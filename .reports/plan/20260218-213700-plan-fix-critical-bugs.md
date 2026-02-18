# Implementation Plan

**Request:** Fix 3 critical bugs (broken column ref in wire resolution, sync tracks wrong containers, AES key not mounted)
**Discovery Level:** 1 — Quick scan, confirming exact lines and column names
**Overall Risk:** medium
**Files affected:** 3
**Context budget:** 3 tasks (~50%)

## Must-Haves

### Observable Truths
- `POST /wires/resolve` no longer crashes with DB error referencing non-existent column
- `loadAllProviders` and `loadAllConsumers` JOIN on `si.template_service_id`, matching sibling functions
- `syncContainerStatus` computes `devarch-{stack}-{instance}` names and looks them up in `runningSet`
- `container_states` rows updated for services whose instances are actually running
- `devarch-api` container mounts `~/.devarch` so AES key survives container recreation

### Required Artifacts
- `api/internal/orchestration/service.go` — Corrected SQL in `loadAllProviders` and `loadAllConsumers`
- `api/internal/sync/manager.go` — `syncContainerStatus` with instance container name resolution
- `compose.yml` — Volume mount for AES key directory

### Key Links
- `loadAllProviders` → `service_exports` via `JOIN on se.service_id = si.template_service_id`
- `loadAllConsumers` → `service_import_contracts` via `JOIN on ic.service_id = si.template_service_id`
- `syncContainerStatus` → `service_instances+stacks` via second query building `devarch-{stack.name}-{si.instance_id}`
- `devarch-api` container → host `~/.devarch/secret.key` via bind-mount in `compose.yml`

## Tasks

### Task 1: Fix wrong column reference in loadAllProviders and loadAllConsumers
- **File:** `api/internal/orchestration/service.go`
- **Risk:** medium — Fixes a crash but changes SQL joins in wire resolution path; incorrect change would silently return empty results
- **Action:**
  - At line 498, change `se.service_id = si.service_id` to `se.service_id = si.template_service_id`
  - At line 555, change `ic.service_id = si.service_id` to `ic.service_id = si.template_service_id`
  - No other changes. Reference: sibling functions `loadProviders` (line 473) and `loadConsumers` (line 523) already use `template_service_id` — match that pattern exactly.
- **Verify:** `grep -n "si.service_id" api/internal/orchestration/service.go` returns no matches; `grep -n "si.template_service_id"` shows 4 matches
- **Done:** Both JOIN clauses reference `si.template_service_id`
- **Depends on:** none

### Task 2: Fix sync manager to look up instance container names
- **File:** `api/internal/sync/manager.go`
- **Risk:** medium — Adds second DB query per sync cycle; upsert on `service_id` means last-write-wins if multiple instances share a template
- **Action:**
  After the existing services query+loop in `syncContainerStatus` (after Part 1 closes its rows), add a second block:
  ```go
  instanceRows, err := m.db.Query(`
      SELECT si.template_service_id, si.instance_id, si.container_name, st.name AS stack_name
      FROM service_instances si
      JOIN stacks st ON st.id = si.stack_id
      WHERE si.deleted_at IS NULL AND si.template_service_id IS NOT NULL
  `)
  if err != nil {
      m.logger.Error("sync: failed to query instances", "error", err)
  } else {
      defer instanceRows.Close()
      for instanceRows.Next() {
          var templateServiceID int
          var instanceID string
          var containerName sql.NullString
          var stackName string
          if err := instanceRows.Scan(&templateServiceID, &instanceID, &containerName, &stackName); err != nil {
              continue
          }
          var cname string
          if containerName.Valid && containerName.String != "" {
              cname = containerName.String
          } else {
              cname = fmt.Sprintf("devarch-%s-%s", stackName, instanceID)
          }
          status := "stopped"
          if state, ok := runningSet[cname]; ok {
              status = state
          }
          _, err := m.db.Exec(`
              INSERT INTO container_states (service_id, status, updated_at)
              VALUES ($1, $2, NOW())
              ON CONFLICT (service_id) DO UPDATE SET status = $2, updated_at = NOW()
          `, templateServiceID, status)
          if err != nil {
              m.logger.Error("sync: failed to update instance status", "container", cname, "error", err)
          }
      }
  }
  ```
  Ensure `"database/sql"` and `"fmt"` are imported (both already present in the file header).
- **Verify:** `go build ./...` from `api/` succeeds; `grep "devarch-%s-%s" api/internal/sync/manager.go` returns a match
- **Done:** `syncContainerStatus` queries `service_instances JOIN stacks`, computes container name, upserts `container_states`
- **Depends on:** none

### Task 3: Mount AES key directory into devarch-api container
- **File:** `compose.yml`
- **Risk:** high — Requires container recreation; if host `~/.devarch` doesn't exist, compose auto-creates it as directory (acceptable — `LoadOrGenerateKey` creates the key inside)
- **Action:**
  In `devarch-api` service volumes list, add:
  ```yaml
  - ${HOME}/.devarch:/root/.devarch
  ```
  Full volumes block becomes:
  ```yaml
      volumes:
        - /run/user/1000/podman/podman.sock:/run/podman/podman.sock:ro
        - ./apps:/workspace/apps:ro
        - ./services-library:/workspace/services-library:ro
        - ./api:/app
        - ${HOME}/.devarch:/root/.devarch
  ```
- **Verify:** `grep ".devarch" compose.yml` returns the mount line
- **Done:** `compose.yml` devarch-api volumes includes `${HOME}/.devarch:/root/.devarch`
- **Depends on:** none

## Verification Plan

| Check | Command | Covers | Expected |
|-------|---------|--------|----------|
| Go build | `cd api && go build ./...` | Task 1, 2 | Success |
| No stale column ref | `grep -n 'si\.service_id' api/internal/orchestration/service.go` | Task 1 | No output |
| Correct column count | `grep -c 'si\.template_service_id' api/internal/orchestration/service.go` | Task 1 | 4 |
| Instance sync pattern | `grep 'devarch-%s-%s' api/internal/sync/manager.go` | Task 2 | Match found |
| Key mount | `grep '\.devarch' compose.yml` | Task 3 | Mount line present |

## XML Plan

```xml
<plan>
  <task type="auto">
    <name>Fix wrong column reference in loadAllProviders and loadAllConsumers</name>
    <files>api/internal/orchestration/service.go</files>
    <action>Line 498: change se.service_id = si.service_id → se.service_id = si.template_service_id. Line 555: same change for loadAllConsumers.</action>
    <verify>grep -n "si.service_id" returns no matches; grep -n "si.template_service_id" shows 4</verify>
    <done>Both JOIN clauses reference si.template_service_id</done>
    <risk level="medium">Changes SQL joins in wire resolution path</risk>
    <needs/>
    <creates/>
  </task>
  <task type="auto">
    <name>Fix sync manager to look up instance container names</name>
    <files>api/internal/sync/manager.go</files>
    <action>Add second query block after services loop: SELECT from service_instances JOIN stacks, compute devarch-{stack}-{instance} name, upsert container_states</action>
    <verify>go build ./... succeeds; grep "devarch-%s-%s" returns match</verify>
    <done>syncContainerStatus resolves instance container names and updates status</done>
    <risk level="medium">Adds second DB query per sync cycle; last-write-wins on shared template_service_id</risk>
    <needs/>
    <creates/>
  </task>
  <task type="auto">
    <name>Mount AES key directory into devarch-api container</name>
    <files>compose.yml</files>
    <action>Add volume mount: ${HOME}/.devarch:/root/.devarch to devarch-api service</action>
    <verify>grep ".devarch" compose.yml returns mount line</verify>
    <done>AES key persists across container recreation</done>
    <risk level="high">Requires container recreation to take effect</risk>
    <needs/>
    <creates/>
  </task>
</plan>
```

## Unresolved Questions

1. **Last-write-wins on shared template:** `container_states` is keyed by `service_id`. Two instances of same template → second upsert overwrites first. Proper fix needs `service_instance_states` table. Acceptable for now?
2. **`${HOME}` in compose.yml:** Resolves at runtime. If compose runs via systemd without `HOME`, mount fails. Hardcode `/home/fhcadmin` instead?
