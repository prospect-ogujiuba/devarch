# Backend Service Port Allocation

## Overview

This document describes the standardized port allocation strategy for all backend runtime services in DevArch. Each backend runtime has a dedicated 100-port range to eliminate conflicts and enable simultaneous operation of all services.

## Port Allocation Strategy

### Design Principles

1. **100-Port Ranges**: Each language runtime gets 100 ports (8X00-8X99)
2. **No Conflicts**: All services can run simultaneously without port collisions
3. **Predictable**: Easy to remember and reference (PHP=8100s, Node=8200s, etc.)
4. **Room for Growth**: Each range has unused ports for future expansion
5. **Container Ports Unchanged**: Only host-side ports changed, no application code changes needed

## Port Mappings

### PHP (8100-8199)

| Host Port | Container Port | Service | Description |
|-----------|---------------|---------|-------------|
| 8100 | 8000 | PHP Built-in Server | Main PHP/Laravel application server |
| 8102 | 5173 | Vite | Vite development server |

**Access URLs:**
- PHP App: http://localhost:8100
- Vite: http://localhost:8102

**Available Ports:** 8101, 8103-8199 (reserved for future use)

---

### Node.js (8200-8299)

| Host Port | Container Port | Service | Description |
|-----------|---------------|---------|-------------|
| 8200 | 3000 | Express/Node | Primary Node.js application |
| 8201 | 3001 | Express/Node | Secondary Node.js application |
| 8202 | 5173 | Vite | Vite development server |
| 8203 | 4000 | Apollo/GraphQL | GraphQL server |
| 9229 | 9229 | Node Inspector | Node.js debugger (unchanged, well-known port) |

**Access URLs:**
- Primary App: http://localhost:8200
- Secondary App: http://localhost:8201
- Vite: http://localhost:8202
- GraphQL: http://localhost:8203
- Debugger: ws://localhost:9229

**Available Ports:** 8204-8299 (reserved for future use)

---

### Python (8300-8399)

| Host Port | Container Port | Service | Description |
|-----------|---------------|---------|-------------|
| 8300 | 8000 | Django/FastAPI | Main Python web application |
| 8301 | 5000 | Flask | Flask application server |
| 8302 | 8888 | Jupyter Lab | Jupyter notebook environment |
| 8303 | 5555 | Flower | Celery task monitoring UI |

**Access URLs:**
- Django/FastAPI: http://localhost:8300
- Flask: http://localhost:8301
- Jupyter Lab: http://localhost:8302
- Flower: http://localhost:8303

**Available Ports:** 8304-8399 (reserved for future use)

---

### Go (8400-8499)

| Host Port | Container Port | Service | Description |
|-----------|---------------|---------|-------------|
| 8400 | 8080 | Go HTTP Server | Main Go application |
| 8401 | 8081 | Metrics | Prometheus metrics endpoint |
| 8402 | 2345 | Delve | Go debugger (Delve) |
| 8403 | 6060 | pprof | Go profiling endpoint |

**Access URLs:**
- Go App: http://localhost:8400
- Metrics: http://localhost:8401/metrics
- Debugger: localhost:8402
- Profiler: http://localhost:8403/debug/pprof/

**Available Ports:** 8404-8499 (reserved for future use)

---

## Migration Notes

### Previous Port Conflicts

**Before this change:**
- Port 8000: Used by PHP, Node, AND Python (conflict!)
- Port 5173: Used by PHP AND Node (conflict!)
- Only ONE backend service could run at a time

**After this change:**
- Zero port conflicts
- All backend services can run simultaneously
- Microservices architecture fully functional

### Application Changes Required

**None!** Container-internal ports remain unchanged. Applications inside containers don't need any code modifications. Only the host-side port mappings changed.

### Developer Workflow Changes

**Before:**
```bash
# Could only run one at a time
./scripts/service-manager.sh start php
# or
./scripts/service-manager.sh start node  # Would conflict if PHP running
```

**After:**
```bash
# Can run all simultaneously
./scripts/service-manager.sh start backend  # Starts all 4 runtimes
# or start them individually
./scripts/service-manager.sh start php node python go
```

---

## Verification

### Check for Port Conflicts

```bash
# View all backend port mappings
grep -h "^\s*-\s*\"" compose/backend/*.yml | grep -E ":[0-9]+:[0-9]+" | sort

# Check if ports are in use
ss -tlnp | grep -E ":(8100|8200|8300|8400|9229)"
```

### Test All Services Together

```bash
# Start all backend services
./scripts/service-manager.sh start backend

# Verify all are running
./scripts/service-manager.sh status

# Check port mappings
podman port php
podman port node
podman port python
podman port go

# Test connectivity
curl http://localhost:8100  # PHP
curl http://localhost:8200  # Node
curl http://localhost:8300  # Python
curl http://localhost:8400  # Go
```

---

## Configuration Files Modified

The following files were updated to implement this strategy:

1. **compose/backend/php.yml** - Updated port mappings (8100, 8102)
2. **compose/backend/node.yml** - Updated port mappings (8200-8203, 9229)
3. **compose/backend/python.yml** - Updated port mappings (8300-8303)
4. **compose/backend/go.yml** - Updated port mappings (8400-8403)
5. **scripts/config.sh** - Added port allocation documentation comments
6. **CLAUDE.md** - Added backend service ports section
7. **docs/BACKEND_PORTS.md** - This comprehensive reference (new file)

---

## Future Expansion

Each runtime has plenty of unused ports in its range:

- **PHP**: 97 unused ports (8101, 8103-8199)
- **Node**: 96 unused ports (8204-8299)
- **Python**: 96 unused ports (8304-8399)
- **Go**: 96 unused ports (8404-8499)

Use these ports for:
- Additional application instances
- Microservices within the same runtime
- Development tools specific to that language
- Hot reload servers
- Test servers

---

## Quick Reference Card

```
╔═══════════════════════════════════════════════════════════╗
║           DevArch Backend Port Quick Reference            ║
╠═══════════════════════════════════════════════════════════╣
║  PHP      8100-8199  │  8100 (app)   8102 (vite)         ║
║  Node     8200-8299  │  8200 (app)   8202 (vite)         ║
║                      │  8201 (app2)  8203 (graphql)      ║
║                      │  9229 (debug)                      ║
║  Python   8300-8399  │  8300 (app)   8302 (jupyter)      ║
║                      │  8301 (flask) 8303 (flower)       ║
║  Go       8400-8499  │  8400 (app)   8402 (debug)        ║
║                      │  8401 (metrics) 8403 (pprof)      ║
╚═══════════════════════════════════════════════════════════╝
```

---

## Troubleshooting

### Port Already in Use

If you get "port already in use" errors:

```bash
# Find what's using the port
ss -tlnp | grep :8100

# Or use lsof
lsof -i :8100

# Stop the conflicting service
./scripts/service-manager.sh stop <service>
```

### Service Not Accessible

If a service won't respond on its port:

```bash
# Check if service is running
./scripts/service-manager.sh status <service>

# Check container logs
./scripts/service-manager.sh logs <service>

# Verify port mapping
podman port <service>

# Restart the service
./scripts/service-manager.sh restart <service>
```

### Localhost vs 127.0.0.1

- **PHP**: Binds to all interfaces (accessible from host network)
- **Node/Python/Go**: Bind to 127.0.0.1 (localhost only)

This is intentional for security. PHP needs broader access for WordPress development.

---

## References

- Main documentation: `/home/fhcadmin/projects/devarch/CLAUDE.md`
- Configuration: `/home/fhcadmin/projects/devarch/scripts/config.sh`
- Service manager: `/home/fhcadmin/projects/devarch/scripts/service-manager.sh`
- Compose files: `/home/fhcadmin/projects/devarch/compose/backend/`
