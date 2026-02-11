# v1.1.2 — Podman-Native Dashboard Features

Prerequisite: v1.1.1 internal cleanup complete.

## Scope: exec/attach + image management

### Container Operations — Exec/Attach (Web Terminal)
- Interactive shell into running containers via dashboard
- WebSocket-based (gorilla/websocket already in stack)
- Resize support, ANSI rendering
- Biggest single DX win — eliminates terminal context-switching

### Image Management
- Pull / list / remove / inspect / history
- Build from Dockerfile
- Prune dangling images
- Layer inspection
- Table-stakes CRUD surface, straightforward to implement

## Future (v1.1.3+)

### Kubernetes Integration
- `podman generate kube` from existing stacks
- `podman play kube` to import K8s manifests
- Natural extension of compose generation pattern in `compose/generator.go`
- Differentiator: dev locally with Podman, deploy to K8s

### Volume Management
- Inspect / prune beyond current support
- Volume usage stats, orphan detection
- Prevents "where did my disk space go" issues

## Deprioritized

### Pod Management
- Conceptual overlap with stacks — adds confusion unless users request it
- Revisit if demand surfaces

### Advanced Networking
- Port conflict detection has value; rest is niche for dev tool

### Systemd / Auto-updates
- Production concerns, not dev workflow
