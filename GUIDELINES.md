# DevArch Project Guidelines

## Table of Contents

1. [Project Overview](#project-overview)
2. [Project Structure](#project-structure)
3. [Service Management](#service-management)
4. [Configuration Management](#configuration-management)
5. [Backend Services](#backend-services)
6. [Development Workflows](#development-workflows)
7. [Docker/Podman Conventions](#dockerpodman-conventions)
8. [Networking](#networking)
9. [Database Management](#database-management)
10. [Best Practices](#best-practices)
11. [Troubleshooting](#troubleshooting)

---

## Project Overview

**DevArch** is a comprehensive microservices development environment that provides a
unified platform for managing multiple backend runtimes, databases, analytics tools, CI/CD
pipelines, and supporting services using Docker Compose or Podman Compose.

### Key Features

- **Multiple Backend Runtimes**: PHP, Node.js, Python, Go, .NET, Rust, Java, Bun, Deno,
  Elixir, Zig
- **Unified Service Manager**: Single command-line tool for managing all services
- **Smart Configuration**: Centralized configuration with dependency management
- **Port Isolation**: Dedicated port ranges for each runtime to prevent conflicts
- **Flexible Deployment**: Support for both Docker and Podman (rootless or rootful)

---

## Project Structure

```
devarch/
├── apps/                       # Application code
│   └── dashboard/             # Management dashboard
├── compose/                   # Docker/Podman Compose files
│   ├── analytics/            # Monitoring & analytics (Prometheus, Grafana, etc.)
│   ├── backend/              # Backend runtime services
│   ├── ci/                   # CI/CD tools (Jenkins, Drone, GitLab Runner, etc.)
│   ├── collaboration/        # Team collaboration (Mattermost, Nextcloud, etc.)
│   ├── database/             # Database services (PostgreSQL, MySQL, MongoDB, etc.)
│   ├── dbms/                 # Database management tools (phpMyAdmin, Adminer, etc.)
│   ├── docs/                 # Documentation platforms (Wiki.js, BookStack, etc.)
│   ├── erp/                  # ERP systems
│   ├── exporters/            # Prometheus exporters
│   ├── gateway/              # API gateways (KrakenD, Kong, Traefik, etc.)
│   ├── mail/                 # Mail services (Mailpit, MailHog, etc.)
│   ├── management/           # Container management (Portainer, DevArch UI, etc.)
│   ├── messaging/            # Message queues (Kafka, RabbitMQ, NATS, etc.)
│   ├── project/              # Project management (GitLab, Gitea, Taiga, etc.)
│   ├── proxy/                # Reverse proxies (Nginx, Caddy, HAProxy, etc.)
│   ├── registry/             # Container registries (Harbor, Nexus, etc.)
│   ├── search/               # Search engines (Meilisearch, Typesense, Solr, etc.)
│   ├── security/             # Security tools (Vault, Keycloak, Authentik, etc.)
│   ├── storage/              # Object storage (MinIO, SeaweedFS, etc.)
│   ├── support/              # Support tools
│   ├── testing/              # Testing tools (Selenium, K6, Playwright, etc.)
│   └── workflow/             # Workflow automation (Airflow, n8n, Prefect, etc.)
├── config/                    # Service-specific configurations
│   ├── node/                 # Node.js Dockerfile and configs
│   ├── python/               # Python Dockerfile and configs
│   ├── php/                  # PHP Dockerfile and configs
│   ├── go/                   # Go Dockerfile and configs
│   ├── dotnet/               # .NET Dockerfile and configs
│   ├── rust/                 # Rust Dockerfile and configs
│   ├── nginx/                # Nginx configurations
│   ├── krakend/              # KrakenD API gateway configs
│   ├── prometheus/           # Prometheus configs and rules
│   ├── kafka/                # Kafka configurations
│   └── [other services]/     # Service-specific configs
├── scripts/                   # Management scripts
│   ├── service-manager.sh    # Unified service management tool
│   ├── config.sh             # Centralized configuration
│   ├── vite-auto-discover.sh # Vite development server automation
│   ├── laravel/              # Laravel-specific scripts
│   └── wordpress/            # WordPress-specific scripts
├── context/                   # Docker build contexts
├── prompts/                   # AI/LLM prompts and templates
├── .env                       # Environment variables (create from .env-sample)
└── CLAUDE.md                  # AI assistant context documentation
```

---

## Service Management

### The Unified Service Manager

The `service-manager.sh` script is your primary tool for managing all services. It
provides a consistent interface for starting, stopping, and managing services across the
entire stack.

#### Basic Commands

```bash
# Individual service commands
./scripts/service-manager.sh up <service>        # Start a service
./scripts/service-manager.sh down <service>      # Stop a service
./scripts/service-manager.sh restart <service>   # Restart a service
./scripts/service-manager.sh rebuild <service>   # Rebuild and restart a service
./scripts/service-manager.sh logs <service>      # View service logs

# System-wide commands
./scripts/service-manager.sh start [targets]     # Start services/categories
./scripts/service-manager.sh stop [targets]      # Stop services/categories
./scripts/service-manager.sh status              # Show all service statuses
./scripts/service-manager.sh ps                  # List running services
./scripts/service-manager.sh list                # List all available services

# System management
./scripts/service-manager.sh list-components     # List all podman/docker components
./scripts/service-manager.sh prune-components    # Remove ALL components (destructive!)
```

#### Common Options

```bash
--sudo                  # Use sudo for Docker/Podman commands
--errors                # Show error messages
--force-recreate        # Force recreation of containers
--no-cache              # Build without cache
--remove-volumes        # Remove volumes when stopping
--follow                # Follow log output
--tail N                # Show last N log lines
--timeout N             # Timeout for stop operations (default: 30s)
--dry-run               # Show what would be executed without running
--verbose               # Verbose output
```

#### Examples

```bash
# Start database and backend services
./scripts/service-manager.sh start database backend

# Stop specific service with volume removal
./scripts/service-manager.sh down postgres --remove-volumes

# Rebuild Node.js service without cache
./scripts/service-manager.sh rebuild node --no-cache

# Follow logs for a service
./scripts/service-manager.sh logs redis --follow --tail 50

# Start all services in analytics category
./scripts/service-manager.sh start --categories analytics

# Stop all services except databases
./scripts/service-manager.sh stop-all --except-categories database

# List all Docker/Podman components
./scripts/service-manager.sh list-components

# Prune all components (use with caution!)
./scripts/service-manager.sh prune-components --force
```

### Service Categories

Services are organized into logical categories with dependency-aware startup ordering:

1. **database** - Core databases (PostgreSQL, MySQL, MongoDB, Redis, etc.)
2. **storage** - Object storage (MinIO, SeaweedFS, etc.)
3. **dbms** - Database management UIs
4. **security** - Authentication & secrets (Vault, Keycloak, etc.)
5. **registry** - Container registries
6. **gateway** - API gateways
7. **proxy** - Reverse proxies
8. **management** - Container management UIs
9. **backend** - Backend runtime services
10. **ci** - CI/CD pipelines
11. **project** - Project management tools
12. **mail** - Email services
13. **exporters** - Metrics exporters
14. **analytics** - Monitoring & analytics
15. **messaging** - Message queues
16. **search** - Search engines
17. **workflow** - Workflow automation
18. **docs** - Documentation platforms
19. **testing** - Testing tools
20. **collaboration** - Team collaboration
21. **ai** - AI/ML services

---

## Configuration Management

### Central Configuration (`scripts/config.sh`)

All project-wide settings are managed through `scripts/config.sh`:

- **Service definitions** and categories
- **Path resolution** for compose files
- **Runtime detection** (Docker vs Podman, rootless vs rootful)
- **Network configuration**
- **Environment variable** defaults
- **Helper functions** for service operations

### Environment Variables (`.env` file)

Create a `.env` file in the project root (copy from `.env-sample` if available):

```bash
# Database credentials
MARIADB_ROOT_PASSWORD=123456
MYSQL_ROOT_PASSWORD=123456
POSTGRES_PASSWORD=123456
MONGO_ROOT_PASSWORD=123456

# Admin credentials
ADMIN_USER=admin
ADMIN_PASSWORD=123456
ADMIN_EMAIL=admin@test.local

# Database users
MARIADB_USER=mariadb_user
MARIADB_PASSWORD=123456
MYSQL_CUSTOM_USER=mysql_user
POSTGRES_CUSTOM_USER=postgres_user
MONGO_CUSTOM_USER=mongodb_user

# Application-specific
DB_MYSQL_NAME=npm
DB_MYSQL_USER=npm_user
MATOMO_DATABASE_DBNAME=matomo
MB_DB_NAME=metabase
NC_DATABASE_NAME=nocodb

# GitHub integration (optional)
GITHUB_TOKEN=
GITHUB_USER=
```

### Service Path Resolution

The configuration system supports flexible path resolution:

1. **Standard paths**: `compose/<category>/<service>.yml`
2. **Category overrides**: Custom directory for entire category
3. **Service overrides**: Specific path for individual services
4. **Fallback search**: Automatic search in subdirectories

---

## Backend Services

### Port Allocation Strategy

Each runtime has a dedicated 100-port range for clean separation:

| Runtime     | Port Range | Primary | Secondary   | Vite/Hot Reload | Additional                              |
|-------------|------------|---------|-------------|-----------------|-----------------------------------------|
| **PHP**     | 8100-8199  | 8100    | -           | 8102            | -                                       |
| **Node.js** | 8200-8299  | 8200    | 8201        | 8202            | GraphQL: 8203, Debug: 9229              |
| **Python**  | 8300-8399  | 8300    | Flask: 8301 | -               | Jupyter: 8302, Flower: 8303             |
| **Go**      | 8400-8499  | 8400    | -           | -               | Metrics: 8401, Debug: 8402, pprof: 8403 |
| **.NET**    | 8600-8699  | 8600    | 8601        | 8603            | Debug: 8602                             |
| **Rust**    | 8700-8799  | 8700    | 8701        | -               | Debug: 8702, Metrics: 8703              |

### Backend Runtime Features

Each backend service includes:

- **Development tools**: Debuggers, linters, formatters
- **Package managers**: npm/yarn/pnpm (Node), pip/poetry/uv (Python), etc.
- **Database clients**: PostgreSQL, MySQL, MongoDB, Redis
- **Testing frameworks**: Jest/Mocha (Node), pytest (Python), etc.
- **Hot reload**: Vite integration for frontend frameworks
- **Non-root user**: Security best practice with matching UID/GID
- **Oh My Zsh**: Enhanced shell experience

---

## Development Workflows

### Starting a New Project

1. **Choose your runtime** (e.g., Node.js, Python, PHP)
2. **Start required services**:
   ```bash
   ./scripts/service-manager.sh start database node
   ```
3. **Access the container**:
   ```bash
   podman exec -it node zsh
   # or
   docker exec -it node zsh
   ```
4. **Create your application** inside `/app` directory
5. **Configure hot reload** if using Vite (automatic discovery enabled)

### Working with Databases

1. **Start database service**:
   ```bash
   ./scripts/service-manager.sh up postgres
   ```
2. **Access database management UI**:
   ```bash
   ./scripts/service-manager.sh up pgadmin
   ```
3. **Connect from backend**: Use service name as hostname
   ```
   Host: postgres
   Port: 5432
   User: postgres
   Password: (from .env)
   ```

### Hot Reload with Vite

The `vite-auto-discover.sh` script automatically discovers and proxies Vite dev servers:

- Scans all backend services for Vite processes
- Configures dynamic proxying through API gateway
- Supports multiple concurrent Vite servers
- No manual configuration required

---

## Docker/Podman Conventions

### Container Runtime Selection

The project supports both Docker and Podman:

- **Podman**: Preferred for rootless operation and security
- **Docker**: Fully supported alternative

Detection is automatic based on available runtime.

### Rootless vs Rootful

**Podman Rootless** (recommended):

- No sudo required
- Better security isolation
- User namespace mapping
- Persistent service via systemd (optional)

**Rootful Mode** (Docker or Podman with sudo):

- Use `--sudo` flag with service-manager.sh
- Required for some legacy services
- Better compatibility with Docker-only images

### User ID Mapping

All custom backend images support UID/GID mapping:

```yaml
services:
  node:
    build:
      context: ../config/node
      args:
        MY_UID: 1000
        MY_GID: 1000
```

This ensures file ownership matches your host user.

### Volume Management

**Best Practices**:

- Use named volumes for persistent data
- Use bind mounts for development code
- Avoid volumes in read-only paths
- Regular backups of important volumes

**Volume Cleanup**:

```bash
# List volumes
./scripts/service-manager.sh list-components

# Remove all volumes (DESTRUCTIVE!)
./scripts/service-manager.sh prune-components --force
```

---

## Networking

### Container Network

All services connect to a shared bridge network:

- **Network Name**: `microservices-net`
- **Driver**: bridge
- **Automatic creation**: On first service start
- **Service discovery**: Use service names as hostnames

### Inter-Service Communication

Services communicate using service names:

```yaml
# In your application configuration
database_host: postgres
redis_host: redis
api_gateway: krakend
```

### Port Mapping

**Internal**: Services communicate on their internal ports
**External**: Selected ports mapped to host for development access

Example:

```yaml
services:
  postgres:
    ports:
      - "5432:5432"  # Host:Container
```

---

## Database Management

### Supported Databases

- **Relational**: PostgreSQL, MySQL, MariaDB, MSSQL
- **NoSQL**: MongoDB, CouchDB, Cassandra
- **In-Memory**: Redis, Memcached
- **Graph**: Neo4j
- **Time-Series**: ClickHouse, VictoriaMetrics
- **Modern**: SurrealDB, EdgeDB, CockroachDB

### Database Tools

- **Universal**: Adminer, CloudBeaver
- **PostgreSQL**: pgAdmin
- **MySQL/MariaDB**: phpMyAdmin
- **MongoDB**: Mongo Express
- **Redis**: Redis Commander
- **General BI**: Metabase, NocoDB

### Backup & Restore

**Manual Backup**:

```bash
# PostgreSQL
podman exec postgres pg_dump -U postgres dbname > backup.sql

# MySQL
podman exec mysql mysqldump -u root -p dbname > backup.sql

# MongoDB
podman exec mongodb mongodump --out /backup
```

**Restore**:

```bash
# PostgreSQL
podman exec -i postgres psql -U postgres dbname < backup.sql

# MySQL
podman exec -i mysql mysql -u root -p dbname < backup.sql
```

---

## Best Practices

### Security

1. **Change default passwords** in `.env` file
2. **Use rootless Podman** when possible
3. **Limit exposed ports** to only necessary services
4. **Regular updates** of base images
5. **Scan images** for vulnerabilities (use Trivy service)
6. **Use secrets management** (Vault service available)

### Performance

1. **Limit concurrent services** - Start only what you need
2. **Resource limits** - Configure CPU/memory limits in compose files
3. **Use volumes wisely** - Bind mounts can be slower on non-Linux hosts
4. **Prune regularly** - Remove unused images/containers/volumes
5. **Monitor resources** - Use Prometheus + Grafana for insights

### Development

1. **Use version control** for your application code
2. **Document service dependencies** in your project README
3. **Keep compose files simple** - One service per file
4. **Environment-specific configs** - Use `.env` for local overrides
5. **Test in isolation** - Start only required services for testing
6. **@apps/dashboard is a first class citizen**
7. The tech stack for the dashboard project should be read form the package.json.
8. **Always Use existing packages i.e., Tailwind Plus via HeadlessUI for components, and
   the icon packages unless specified otherwise.**

### Maintenance

1. **Regular backups** of databases and volumes
2. **Update base images** periodically
3. **Clean up** unused resources
4. **Monitor logs** for errors and warnings
5. **Document custom configurations**

---

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check service status
./scripts/service-manager.sh status

# View logs
./scripts/service-manager.sh logs <service> --tail 100

# Try force recreate
./scripts/service-manager.sh up <service> --force-recreate
```

#### Port Already in Use

```bash
# Find what's using the port (Linux)
sudo lsof -i :8080

# Stop the conflicting service or change port in compose file
```

#### Permission Denied

```bash
# For Podman rootless
podman unshare chown -R 1000:1000 /path/to/volume

# For Docker
sudo chown -R $USER:$USER /path/to/volume

# Or use --sudo flag
./scripts/service-manager.sh up <service> --sudo
```

#### Network Not Found

```bash
# Recreate network
podman network create --driver bridge microservices-net
# or
docker network create --driver bridge microservices-net
```

#### Container Name Conflicts

```bash
# Remove existing container
podman rm -f <container>
# or
docker rm -f <container>

# Or use prune-components (removes ALL)
./scripts/service-manager.sh prune-components --force
```

### Debugging Services

#### Check Container Status

```bash
podman ps -a
# or
docker ps -a
```

#### Inspect Container

```bash
podman inspect <container>
# or
docker inspect <container>
```

#### Access Container Shell

```bash
podman exec -it <container> zsh
# or
docker exec -it <container> zsh
```

#### View Resource Usage

```bash
podman stats
# or
docker stats
```

### Getting Help

1. **Check logs**: Most issues are visible in service logs
2. **Verify configuration**: Ensure `.env` file is properly configured
3. **Check documentation**: Review service-specific documentation
4. **Network issues**: Verify all services are on the same network
5. **Resource constraints**: Check available disk space and memory

---

## Additional Resources

### Useful Commands Reference

```bash
# Service Management
./scripts/service-manager.sh status              # Overall system status
./scripts/service-manager.sh ps                  # Running services
./scripts/service-manager.sh list                # All available services
./scripts/service-manager.sh list-components     # All Docker/Podman components

# Bulk Operations
./scripts/service-manager.sh start-all           # Start all services
./scripts/service-manager.sh stop-all            # Stop all services
./scripts/service-manager.sh start database backend  # Start multiple categories
./scripts/service-manager.sh stop --categories analytics,proxy

# Advanced Options
./scripts/service-manager.sh start --rebuild     # Rebuild before starting
./scripts/service-manager.sh stop --remove-volumes  # Remove data
./scripts/service-manager.sh --dry-run <command> # Preview actions
./scripts/service-manager.sh --verbose <command> # Detailed output
```

### Project Conventions

1. **Service naming**: Use lowercase, hyphen-separated names
2. **Compose files**: One service per file, named `<service>.yml`
3. **Categories**: Group related services in subdirectories
4. **Environment**: Use `.env` for all configurable values
5. **Ports**: Follow the port allocation strategy
6. **Volumes**: Prefix with project name for clarity
7. **Networks**: Single shared network for all services
8. **Images**: Build custom images in `config/<runtime>/`

### File Naming

- **Compose files**: `service-name.yml`
- **Dockerfiles**: `Dockerfile` (one per service in config directory)
- **Scripts**: `kebab-case.sh`
- **Config files**: Match service name or convention

---

## Contributing

When adding new services or modifying existing ones:

1. **Follow existing patterns** in compose files
2. **Update `SERVICE_CATEGORIES`** in `scripts/config.sh`
3. **Document environment variables** in `.env-sample`
4. **Test with both Docker and Podman** if possible
5. **Update this guidelines document** with relevant changes
6. **Use appropriate service category** for organization
7. **Follow port allocation strategy** for backend services

---

*This document should be kept up-to-date as the project evolves. Last updated: 2025-12-13*
