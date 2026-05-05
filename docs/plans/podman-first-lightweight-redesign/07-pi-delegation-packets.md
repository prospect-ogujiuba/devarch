# 07 — Pi Delegation Packets

## Packet A — ADR + command layer

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement phases 01 and 02 from docs/plans/podman-first-lightweight-redesign/.
Do not touch legacy api/ except to read context.
Run go test ./internal/podmanctl/... and report files changed, tests run, and any blockers.
"
```

## Packet B — Podman adapter mutations

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement phase 03 from docs/plans/podman-first-lightweight-redesign/.
Use internal/podmanctl for command construction.
Do not touch legacy api/.
Run go test ./internal/runtime/podman/... ./internal/apply/... ./internal/runtime/... and report files changed, tests run, and blockers.
"
```

## Packet C — workflow services

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement phase 04 from docs/plans/podman-first-lightweight-redesign/.
Port doctor/runtime/socket workflows into internal/workflows and expose them through internal/appsvc.
Do not modify scripts yet except to read behavior.
Run go test ./internal/workflows/... ./internal/appsvc/... and report files changed, tests run, and blockers.
"
```

## Packet D — CLI/API exposure

```bash
cd /home/priz/projects/devarch
pi -p --no-session "
Implement phase 05 from docs/plans/podman-first-lightweight-redesign/.
Keep CLI and API thin over internal/appsvc.
Run go test ./cmd/devarch/... ./internal/api/... ./internal/appsvc/... and report files changed, tests run, and blockers.
"
```
