# DevArch Service Manager Documentation

## Overview

The `devarch` command is a minimal, transparent wrapper around `podman-compose` (or `docker compose`) that provides UX-friendly commands while showing exactly what container operations are being executed.

**Philosophy**: Simple, transparent, no magic. Developers should understand what's happening under the hood.

## Installation

The `devarch` script is located at `/home/fhcadmin/projects/devarch/scripts/devarch`.

### Add to PATH (recommended)

```bash
# Add to your ~/.bashrc or ~/.zshrc
export PATH="$PATH:/home/fhcadmin/projects/devarch/scripts"

# Or create a symlink
sudo ln -s /home/fhcadmin/projects/devarch/scripts/devarch /usr/local/bin/devarch
```

## Quick Start

```bash
# Start database services
devarch start-db

# Start proxy
devarch start nginx-proxy-manager

# Start backend runtimes
devarch start-backend

# Check what's running
devarch ps

# View logs
devarch logs postgres -f
```

## Core Commands

All core commands execute a single podman-compose operation on one service.

### `devarch start <service>`

Start a service in detached mode.

**Examples:**
```bash
devarch start postgres
devarch start nginx-proxy-manager
devarch start php
```

**Direct equivalent:**
```bash
podman-compose -f compose/database/postgres.yml up -d
```

**What it does:**
1. Finds the service compose file (e.g., `compose/database/postgres.yml`)
2. Creates the `microservices-net` network if missing
3. Runs `podman-compose -f <compose-file> up -d`
4. Shows the exact command before executing

---

### `devarch stop <service>`

Stop a service and remove its containers.

**Examples:**
```bash
devarch stop postgres
devarch stop nginx-proxy-manager
```

**Direct equivalent:**
```bash
podman-compose -f compose/database/postgres.yml down
```

**What it does:**
1. Finds the service compose file
2. Runs `podman-compose -f <compose-file> down`
3. Containers are stopped and removed
4. Volumes are **preserved** (use podman commands directly to remove volumes)

---

### `devarch restart <service>`

Restart a service (stop then start).

**Examples:**
```bash
devarch restart php
devarch restart nginx-proxy-manager
```

**Direct equivalent:**
```bash
podman-compose -f compose/backend/php.yml down
sleep 1
podman-compose -f compose/backend/php.yml up -d
```

---

### `devarch logs <service> [--follow]`

View service logs.

**Examples:**
```bash
devarch logs postgres                    # Last 100 lines
devarch logs nginx-proxy-manager -f      # Follow logs
devarch logs php --follow                # Follow logs
```

**Direct equivalent:**
```bash
podman-compose -f compose/database/postgres.yml logs --tail 100
podman-compose -f compose/database/postgres.yml logs --tail 100 -f
```

---

### `devarch exec <service> <command>`

Execute a command inside a service container.

**Examples:**
```bash
devarch exec php bash
devarch exec postgres psql -U postgres
devarch exec node npm install
devarch exec php wp plugin list
```

**Direct equivalent:**
```bash
podman exec -it php bash
podman exec -it postgres psql -U postgres
```

---

### `devarch ps`

List all running DevArch containers.

**Examples:**
```bash
devarch ps
```

**Direct equivalent:**
```bash
podman ps --filter network=microservices-net
```

**Output:** Shows all containers connected to the DevArch network.

---

## Convenience Commands

These commands execute multiple podman-compose operations in sequence.

### `devarch start-db`

Start all database services sequentially.

**What it starts:**
- mariadb
- mysql
- postgres
- mongodb
- redis
- mssql
- memcached

**Example:**
```bash
devarch start-db
```

**Direct equivalent:**
```bash
podman-compose -f compose/database/mariadb.yml up -d
podman-compose -f compose/database/mysql.yml up -d
podman-compose -f compose/database/postgres.yml up -d
podman-compose -f compose/database/mongodb.yml up -d
podman-compose -f compose/database/redis.yml up -d
podman-compose -f compose/database/mssql.yml up -d
podman-compose -f compose/database/memcached.yml up -d
```

**Note:** 1 second delay between each service to prevent race conditions.

---

### `devarch start-backend`

Start all backend runtime services.

**What it starts:**
- php
- node
- python
- go
- dotnet

**Example:**
```bash
devarch start-backend
```

---

### `devarch start-all`

Start ALL DevArch services in dependency order.

**Startup Order:**
1. **database** - Data stores (mariadb, mysql, postgres, mongodb, redis, mssql, memcached)
2. **dbms** - DB management tools (adminer, phpmyadmin, metabase, etc.)
3. **proxy** - nginx-proxy-manager
4. **management** - portainer
5. **backend** - Runtime environments (php, node, python, go, dotnet)
6. **project** - Project management (openproject, gitea)
7. **mail** - mailpit
8. **exporters** - Prometheus exporters
9. **analytics** - Monitoring stack (prometheus, grafana, elasticsearch, kibana, etc.)
10. **messaging** - Message queues (kafka, rabbitmq)
11. **search** - Search engines (meilisearch, typesense)

**Example:**
```bash
devarch start-all
```

**Note:**
- 0.5 second delay between services within a category
- 2 second delay between categories
- Sequential execution ensures dependencies are met

---

### `devarch stop-all`

Stop all DevArch services in reverse dependency order.

**Example:**
```bash
devarch stop-all
```

**Note:** Stops in reverse order to prevent dependency issues.

---

## Utility Commands

### `devarch status`

Show status of all DevArch services.

**Example:**
```bash
devarch status
```

**Output:**
```
📊 DevArch Service Status:

📂 database:
  ✅ mariadb
  ✅ mysql
  ✅ postgres
  ❌ mongodb
  ...
```

---

### `devarch network`

Show/inspect the DevArch network.

**Example:**
```bash
devarch network
```

**Direct equivalent:**
```bash
podman network inspect microservices-net
```

---

### `devarch list`

List all available DevArch services.

**Example:**
```bash
devarch list
```

**Output:** Shows all services organized by category with compose file status.

---

### `devarch help`

Show help and command reference.

**Example:**
```bash
devarch help
```

---

## Recommended Startup Sequences

### Minimal Development Environment

```bash
# 1. Start databases
devarch start-db

# 2. Start proxy for SSL/routing
devarch start nginx-proxy-manager

# 3. Start your runtime
devarch start php

# 4. Check everything is running
devarch ps
```

### Full Stack Development

```bash
# 1. Start all databases
devarch start-db

# 2. Start proxy
devarch start nginx-proxy-manager

# 3. Start all backend runtimes
devarch start-backend

# 4. Start DB management tools
devarch start adminer
devarch start phpmyadmin

# 5. Start mail catcher
devarch start mailpit

# 6. Check status
devarch status
```

### Complete Observability Stack

```bash
# 1. Start everything
devarch start-all

# This includes:
# - All databases
# - All backend runtimes
# - Prometheus + exporters
# - Grafana
# - ELK Stack (Elasticsearch, Logstash, Kibana)
# - All management tools
```

---

## Startup Order and Dependencies

**IMPORTANT:** DevArch does NOT automatically manage dependencies. You must start services in the correct order.

### Critical Dependencies

1. **Databases MUST start first**
   - Services like nginx-proxy-manager, metabase, and nocodb need database access
   - Always run `devarch start-db` before other categories

2. **Exporters need their targets**
   - `postgres-exporter` requires postgres running
   - `mysqld-exporter` requires mysql/mariadb running
   - `redis-exporter` requires redis running
   - Start exporters AFTER their target databases

3. **Analytics tools need exporters**
   - Prometheus scrapes exporters for metrics
   - Grafana queries Prometheus
   - Start analytics AFTER exporters

### Recommended Order

```bash
# 1. Core infrastructure
devarch start-db
devarch start nginx-proxy-manager

# 2. Application runtimes
devarch start-backend

# 3. Management and tools
devarch start portainer
devarch start mailpit

# 4. Observability (if needed)
devarch start node-exporter
devarch start postgres-exporter
devarch start redis-exporter
devarch start prometheus
devarch start grafana
```

---

## Direct Podman Equivalents

Every devarch command maps to a direct podman operation:

| DevArch Command | Podman Equivalent |
|-----------------|-------------------|
| `devarch start postgres` | `podman-compose -f compose/database/postgres.yml up -d` |
| `devarch stop postgres` | `podman-compose -f compose/database/postgres.yml down` |
| `devarch restart postgres` | `podman-compose -f compose/database/postgres.yml down && podman-compose -f compose/database/postgres.yml up -d` |
| `devarch logs postgres` | `podman-compose -f compose/database/postgres.yml logs --tail 100` |
| `devarch logs postgres -f` | `podman-compose -f compose/database/postgres.yml logs --tail 100 -f` |
| `devarch exec php bash` | `podman exec -it php bash` |
| `devarch ps` | `podman ps --filter network=microservices-net` |
| `devarch status` | `podman inspect --format='{{.State.Status}}' <service>` (for each service) |

**Note:** The script automatically detects whether to use `podman` or `docker` and whether `sudo` is needed.

---

## Advanced Usage

### Manual Volume Management

DevArch does NOT manage volumes. Use podman commands directly:

```bash
# List volumes
podman volume ls

# Remove a specific volume
podman volume rm postgres_data

# Remove all unused volumes (DANGER: data loss)
podman volume prune
```

### Rebuild a Service

DevArch doesn't have a rebuild command. Use podman-compose directly:

```bash
# Rebuild PHP container
podman-compose -f compose/backend/php.yml down
podman-compose -f compose/backend/php.yml build --no-cache
podman-compose -f compose/backend/php.yml up -d
```

### Custom Compose Files

If you create custom compose files outside the standard structure:

```bash
# Use podman-compose directly
podman-compose -f custom/my-service.yml up -d
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
devarch logs <service>

# Check if port is in use
sudo netstat -tulpn | grep <port>

# Check if network exists
devarch network

# Try manual podman-compose
podman-compose -f compose/<category>/<service>.yml up
```

### Service Not Found

```bash
# List all available services
devarch list

# Check if compose file exists
ls compose/<category>/<service>.yml
```

### Network Issues

```bash
# Check network status
devarch network

# Recreate network manually
podman network rm microservices-net
podman network create --driver bridge microservices-net
```

### Permission Issues

If you get permission errors:

```bash
# Check if you need sudo
groups | grep podman   # Should show podman group

# If not in podman group, use rootful podman
sudo podman ps
```

---

## Migration from Old service-manager.sh

### Command Mapping

| Old Command | New Command |
|-------------|-------------|
| `./scripts/service-manager.sh up postgres` | `devarch start postgres` |
| `./scripts/service-manager.sh down postgres` | `devarch stop postgres` |
| `./scripts/service-manager.sh restart postgres` | `devarch restart postgres` |
| `./scripts/service-manager.sh logs postgres --follow` | `devarch logs postgres -f` |
| `./scripts/service-manager.sh status` | `devarch status` |
| `./scripts/service-manager.sh ps` | `devarch ps` |
| `./scripts/service-manager.sh list` | `devarch list` |
| `./scripts/service-manager.sh start database` | `devarch start-db` |
| `./scripts/service-manager.sh start backend` | `devarch start-backend` |
| `./scripts/service-manager.sh start-all` | `devarch start-all` |
| `./scripts/service-manager.sh stop-all` | `devarch stop-all` |

### Removed Features

The new `devarch` command intentionally removes:

1. **Parallel execution** - All operations are sequential for predictability
2. **Health checking** - Let podman-compose handle this
3. **Automatic dependency resolution** - Document startup order instead
4. **Volume management flags** - Use podman commands directly
5. **Rebuild commands** - Use podman-compose directly
6. **Wait flags** - Services start immediately, check logs if needed
7. **Cleanup operations** - Use podman commands directly

### Why These Were Removed

- **Transparency:** Abstraction hides what's happening. Direct commands are clearer.
- **Simplicity:** Less code means fewer bugs and easier maintenance.
- **Learning:** Developers understand podman better when using direct commands.
- **Flexibility:** Manual podman commands provide more control for complex operations.

### Migrating Scripts

If you have scripts using the old service-manager.sh:

**Before:**
```bash
./scripts/service-manager.sh up postgres --force
./scripts/service-manager.sh start database backend --parallel
./scripts/service-manager.sh stop-all --preserve-volumes
```

**After:**
```bash
devarch start postgres
devarch start-db
devarch start-backend
devarch stop-all
```

For advanced features, use podman-compose directly:

```bash
# Force recreate
podman-compose -f compose/database/postgres.yml up -d --force-recreate

# Remove volumes
podman-compose -f compose/database/postgres.yml down --volumes
```

---

## Configuration

### Service Definitions

Services are defined in the script at `/home/fhcadmin/projects/devarch/scripts/devarch`:

```bash
declare -A CATEGORY_SERVICES=(
    ["database"]="mariadb mysql postgres mongodb redis mssql memcached"
    ["backend"]="php node python go dotnet"
    ...
)
```

To add a new service:
1. Add compose file to appropriate category directory
2. Add service name to the category in the script
3. The service will be available immediately

### Network Name

The network is hardcoded as `microservices-net`. To change:

```bash
# Edit the script
NETWORK="your-network-name"
```

---

## Examples

### WordPress Development Workflow

```bash
# 1. Start database
devarch start mariadb

# 2. Start PHP runtime
devarch start php

# 3. Start proxy for SSL
devarch start nginx-proxy-manager

# 4. Start mail catcher
devarch start mailpit

# 5. Check everything is running
devarch ps

# 6. Access PHP container
devarch exec php bash

# 7. Inside container, install WordPress
wp core download
wp config create --dbname=wordpress --dbuser=root --dbpass=123456 --dbhost=mariadb
wp core install --url=mysite.test --title="My Site" --admin_user=admin --admin_email=admin@test.local
```

### Monitoring Stack Setup

```bash
# 1. Start databases
devarch start-db

# 2. Start exporters
devarch start node-exporter
devarch start postgres-exporter
devarch start redis-exporter

# 3. Start Prometheus
devarch start prometheus

# 4. Start Grafana
devarch start grafana

# 5. Check status
devarch status

# 6. View Prometheus logs
devarch logs prometheus -f
```

### Full Reset

```bash
# 1. Stop everything
devarch stop-all

# 2. Remove all volumes (DANGER: data loss)
podman volume prune

# 3. Start fresh
devarch start-all
```

---

## FAQ

### Q: Why not use the old service-manager.sh?

**A:** The old script had 1600+ lines of complex abstraction. The new script is ~600 lines and shows exactly what it's doing. Simplicity wins.

### Q: What happened to parallel startup?

**A:** Removed for simplicity. Sequential startup is fast enough and more predictable. If you need parallel, run multiple `devarch start` commands in different terminals.

### Q: Can I still use the old service-manager.sh?

**A:** Yes, it's backed up at `/home/fhcadmin/projects/devarch/scripts/service-manager.sh.backup`. But the new devarch command is recommended.

### Q: How do I remove volumes?

**A:** Use podman commands directly:
```bash
podman volume rm <volume-name>
podman volume prune  # Remove all unused volumes
```

### Q: How do I rebuild a service?

**A:** Use podman-compose directly:
```bash
podman-compose -f compose/backend/php.yml down
podman-compose -f compose/backend/php.yml build --no-cache
podman-compose -f compose/backend/php.yml up -d
```

### Q: Does this work with Docker?

**A:** Yes, the script auto-detects docker and uses `docker compose` instead of `podman-compose`.

### Q: Can I customize startup order?

**A:** Yes, edit the `CATEGORIES` array in the script. Services start in the order defined.

### Q: What if I don't want to start all services in a category?

**A:** Use individual `devarch start <service>` commands instead of convenience commands.

---

## See Also

- [CLAUDE.md](/home/fhcadmin/projects/devarch/CLAUDE.md) - Project overview and architecture
- [README.md](/home/fhcadmin/projects/devarch/README.md) - Main project documentation
- [Podman Compose Documentation](https://github.com/containers/podman-compose)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
