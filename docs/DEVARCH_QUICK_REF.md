# DevArch Quick Reference

## Installation
```bash
# Add to PATH
export PATH="$PATH:/home/fhcadmin/projects/devarch/scripts"

# Or create symlink
sudo ln -s /home/fhcadmin/projects/devarch/scripts/devarch /usr/local/bin/devarch
```

## Essential Commands

```bash
devarch start <service>              # Start a service
devarch stop <service>               # Stop a service
devarch restart <service>            # Restart a service
devarch logs <service> [-f]          # View logs (-f to follow)
devarch exec <service> <command>     # Run command in container

devarch start-db                     # Start all databases
devarch start-backend                # Start all backend runtimes
devarch start-all                    # Start everything (dependency order)
devarch stop-all                     # Stop everything

devarch ps                           # List running containers
devarch status                       # Show all service statuses
devarch list                         # List all available services
devarch network                      # Show network info
devarch help                         # Show help
```

## Quick Start
```bash
# Minimal setup
devarch start-db
devarch start nginx-proxy-manager
devarch start php

# Full setup
devarch start-all
```

## Service Categories

**database** - mariadb, mysql, postgres, mongodb, redis, mssql, memcached

**backend** - php, node, python, go, dotnet

**proxy** - nginx-proxy-manager

**dbms** - adminer, phpmyadmin, metabase, nocodb, pgadmin, etc.

**analytics** - prometheus, grafana, elasticsearch, kibana, logstash

**exporters** - node-exporter, postgres-exporter, redis-exporter, etc.

**messaging** - kafka, rabbitmq

**search** - meilisearch, typesense

**mail** - mailpit

**project** - openproject, gitea

**management** - portainer

## Common Workflows

### WordPress Development
```bash
devarch start mariadb
devarch start php
devarch start nginx-proxy-manager
devarch start mailpit
devarch exec php bash
```

### Full Stack App
```bash
devarch start postgres
devarch start redis
devarch start node
devarch start nginx-proxy-manager
devarch logs node -f
```

### Monitoring Stack
```bash
devarch start-db
devarch start node-exporter
devarch start postgres-exporter
devarch start prometheus
devarch start grafana
```

## Direct Podman Equivalents

| DevArch | Podman |
|---------|--------|
| `devarch start postgres` | `podman-compose -f compose/database/postgres.yml up -d` |
| `devarch stop postgres` | `podman-compose -f compose/database/postgres.yml down` |
| `devarch logs postgres` | `podman-compose -f compose/database/postgres.yml logs --tail 100` |
| `devarch exec php bash` | `podman exec -it php bash` |
| `devarch ps` | `podman ps --filter network=microservices-net` |

## Advanced Operations

### Rebuild a Service
```bash
podman-compose -f compose/backend/php.yml down
podman-compose -f compose/backend/php.yml build --no-cache
podman-compose -f compose/backend/php.yml up -d
```

### Force Recreate
```bash
podman-compose -f compose/database/postgres.yml up -d --force-recreate
```

### Volume Management
```bash
podman volume ls                     # List volumes
podman volume inspect postgres_data  # Inspect volume
podman volume rm postgres_data       # Remove volume
podman volume prune                  # Remove unused volumes
```

### Network Management
```bash
devarch network                      # Show network status
podman network ls                    # List all networks
podman network inspect microservices-net
```

### Container Management
```bash
podman ps -a                         # All containers
podman logs postgres                 # Container logs
podman inspect postgres              # Container details
podman exec -it postgres bash        # Access container
```

### Cleanup
```bash
podman container prune               # Remove stopped containers
podman volume prune                  # Remove unused volumes
podman image prune                   # Remove unused images
podman system prune                  # Clean everything
```

## Startup Order (for manual starts)

1. **database** (MUST start first)
2. **proxy** (nginx-proxy-manager)
3. **backend** (php, node, python, go, dotnet)
4. **dbms** (adminer, phpmyadmin, etc.)
5. **mail** (mailpit)
6. **exporters** (requires databases running)
7. **analytics** (requires exporters running)
8. **messaging** (kafka, rabbitmq)
9. **search** (meilisearch, typesense)
10. **project** (openproject, gitea)

**Note:** `devarch start-all` handles this automatically.

## Troubleshooting

```bash
# Check service status
devarch status

# View service logs
devarch logs <service>
devarch logs <service> -f    # Follow logs

# Check if network exists
devarch network

# Inspect container
podman inspect <service>

# Check compose file syntax
podman-compose -f compose/<category>/<service>.yml config

# Restart service
devarch restart <service>

# Full reset (DANGER: data loss)
devarch stop-all
podman volume prune
devarch start-all
```

## Common Issues

### Service won't start
1. Check logs: `devarch logs <service>`
2. Check dependencies: `devarch status`
3. Check network: `devarch network`
4. Rebuild: Use podman-compose directly

### Port already in use
```bash
# Find what's using the port
sudo netstat -tulpn | grep <port>
sudo lsof -i :<port>

# Stop conflicting service
sudo systemctl stop <service>
```

### Permission denied
```bash
# Check if you need sudo
groups | grep podman    # Should show podman group

# If not in group
sudo usermod -aG podman $USER
# Log out and back in
```

### Network issues
```bash
# Recreate network
podman network rm microservices-net
podman network create --driver bridge microservices-net
```

## Environment Files

- `.env` - Service credentials and configuration
- Database passwords, admin users, ports, domains

## Key Directories

- `compose/` - Service compose files by category
- `apps/` - Application code
- `config/` - Service configurations and Dockerfiles
- `scripts/` - Management scripts
- `logs/` - Application logs

## Documentation

- [SERVICE_MANAGER.md](SERVICE_MANAGER.md) - Complete documentation
- [MIGRATION_TO_DEVARCH.md](MIGRATION_TO_DEVARCH.md) - Migration guide
- [CLAUDE.md](../CLAUDE.md) - Project overview

## Tips

- Every command shows the podman operation it runs (transparency)
- Network is created automatically if missing
- Volumes are preserved when stopping (use podman commands to remove)
- Sequential startup is safe and fast enough
- For complex operations, use podman-compose directly
- Check logs if something doesn't work: `devarch logs <service> -f`

## Getting Help

```bash
devarch help                         # Command reference
devarch list                         # See all services
devarch status                       # See what's running
```

## Version Info

**New System:** 565 lines, minimal transparent wrapper

**Old System (backup):** 2,579 lines, complex abstraction

**Reduction:** 78% smaller, infinitely more transparent
