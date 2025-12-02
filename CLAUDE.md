# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**DevArch** is a containerized microservices development environment designed for local WordPress development and testing. It provides a complete infrastructure stack including:

- Multiple WordPress application instances (b2bcnc, playground, flowstate, mediagrowthpartner)
- Full database stack (MariaDB, MySQL, PostgreSQL, MongoDB, Redis, Memcached)
- Database management tools (Adminer, phpMyAdmin, Metabase, NocoDB, pgAdmin, etc.)
- Complete observability stack (Prometheus, Grafana, ELK Stack, OpenTelemetry)
- Message queuing systems (Kafka, RabbitMQ)
- Search engines (Meilisearch, Typesense)
- Reverse proxy with SSL (Nginx Proxy Manager)
- Multiple backend runtimes (PHP 8.3, Node.js, Python, Go)

The project uses Podman (with Docker fallback) and organizes services into modular compose files by category.

## Common Commands

### Service Management

The primary tool is `service-manager.sh` which orchestrates all services:

```bash
# Individual service operations
./scripts/service-manager.sh up <service> [--force]
./scripts/service-manager.sh down <service> [--remove-volumes]
./scripts/service-manager.sh restart <service>
./scripts/service-manager.sh rebuild <service> [--no-cache]
./scripts/service-manager.sh logs <service> [--follow] [--tail 100]
./scripts/service-manager.sh status [service]

# Bulk operations
./scripts/service-manager.sh start-all [--parallel] [--wait-healthy]
./scripts/service-manager.sh stop-all [--cleanup-orphans]
./scripts/service-manager.sh list  # List all available services
./scripts/service-manager.sh ps    # List running services

# Start specific categories (respects dependency order)
./scripts/service-manager.sh start database backend proxy

# Start with exclusions
./scripts/service-manager.sh start-all --exclude analytics,messaging

# Start specific services only
./scripts/service-manager.sh start --services postgres,redis,nginx-proxy-manager

# Cleanup operations
./scripts/service-manager.sh stop-all --cleanup-older-than 7d
./scripts/service-manager.sh stop-all --cleanup-large-volumes --max-volume-size 500
```

### Database Setup

```bash
# Setup all databases
./scripts/setup-databases.sh -a

# Setup specific databases
./scripts/setup-databases.sh -m        # MariaDB only
./scripts/setup-databases.sh -p        # PostgreSQL only
./scripts/setup-databases.sh -m -p     # Both

# Include test data
./scripts/setup-databases.sh -a --test-data
```

### WordPress Development

```bash
# Install WordPress instance
./scripts/wordpress/install-wordpress.sh

# WP-CLI commands (execute inside PHP container)
podman exec -it php wp plugin list
podman exec -it php wp theme list
podman exec -it php wp db query "SELECT * FROM wp_posts LIMIT 5"
```

### Host Management

```bash
# Update /etc/hosts with service domains (.test suffix)
./scripts/update-hosts.sh
```

### Quick Start Sequence

```bash
# 1. Initial setup (first time only)
./scripts/install.sh

# 2. Start essential services
./scripts/service-manager.sh start database proxy backend

# 3. Initialize databases
./scripts/setup-databases.sh -a

# 4. Update hosts file
sudo ./scripts/update-hosts.sh

# 5. Check everything is running
./scripts/service-manager.sh status
```

## Architecture

### Service Organization and Dependencies

Services are organized into **11 categories** with strict dependency ordering:

1. **database** - Core data stores (MariaDB, PostgreSQL, MongoDB, Redis, etc.) - **must start first**
2. **dbms** - Database management tools (Adminer, phpMyAdmin, Metabase, etc.)
3. **proxy** - Nginx Proxy Manager (handles routing and SSL)
4. **management** - Administrative tools (Portainer)
5. **backend** - Application runtimes (PHP, Node.js, Python, Go)
6. **project** - Project management tools
7. **mail** - Email services (Mailpit)
8. **exporters** - Prometheus exporters for metrics collection
9. **analytics** - Monitoring stack (Prometheus, Grafana, ELK)
10. **messaging** - Message queues (Kafka, RabbitMQ)
11. **search** - Search engines (Meilisearch, Typesense)

Each service has its own compose file in `/compose/<category>/<service>.yml`. The service-manager.sh script automatically handles dependency ordering when starting services.

### Configuration Management

**Centralized Configuration**: `/home/fhcadmin/projects/devarch/scripts/config.sh` (898 lines)
- Service discovery and path resolution with fallback search
- Container runtime detection (Podman vs Docker, rootless vs rootful)
- Automatic sudo management based on user groups
- Service categorization and dependency management
- Environment variable loading and validation

**Environment Variables**: `.env` file contains all service credentials and configuration
- Database credentials (usernames, passwords)
- Admin credentials for management tools
- Domain configurations
- Port mappings
- **Note**: The .env file is tracked in git and contains exposed credentials - this should be addressed for production use

**Modular Compose Files**: Each service is defined in its own YAML file for maintainability and selective startup

### Container Runtime

The project intelligently detects and adapts to the available container runtime:
- Auto-detects Podman or Docker
- Supports both rootless and rootful Podman
- Automatically determines if sudo is needed
- All services communicate via the `microservices-net` bridge network (created automatically if missing)

### Network Architecture

- **Shared Network**: All services on `microservices-net` (bridge driver)
- **Service Discovery**: Services communicate via container names
- **Port Binding**: Services exposed on 127.0.0.1 (localhost only) for security
- **Domain Suffix**: `.test` for local development domains

### WordPress Application Structure

Each WordPress instance follows this pattern:
```
/apps/<instance-name>/
├── public/                    # WordPress root
│   ├── wp-content/
│   │   ├── plugins/
│   │   │   ├── makermaker/   # Custom plugin
│   │   │   └── makerblocks/  # Custom blocks plugin
│   │   ├── themes/
│   │   │   └── makerstarter/ # Custom theme
│   │   └── mu-plugins/
│   │       └── typerocket-pro-v6/  # TypeRocket framework
│   └── galaxy/                # Galaxy configuration management
```

Applications share the same PHP container but have:
- Dedicated databases
- Isolated file structures
- Separate domain names (e.g., playground.test, b2bcnc.test)

### PHP Container Configuration

The PHP container (`/home/fhcadmin/projects/devarch/config/php/Dockerfile`) is extensively configured:
- **Base**: PHP 8.3 FPM
- **Extensions**: GD, mbstring, zip, opcache, PDO (MySQL, PostgreSQL), Redis, MongoDB, Memcached, ImageMagick, YAML, APCu, Xdebug
- **Tools**: Composer, WP-CLI, Laravel Installer, Node.js, npm, Oh My Zsh
- **Mailpit Integration**: Configured to send emails to Mailpit for testing
- **Development Optimizations**: Xdebug, increased memory limits

### Observability Strategy

**Metrics Collection**:
- Prometheus (port 9090) scrapes 12+ exporters every 15 seconds
- Exporters for: Node, MySQL, PostgreSQL, MongoDB, Redis, Kafka, RabbitMQ, Memcached, and more
- Configuration: `/home/fhcadmin/projects/devarch/config/prometheus/prometheus.yml`

**Logging**:
- Centralized in ELK Stack (Elasticsearch, Logstash, Kibana)
- Logstash processes logs from all services
- Kibana provides visualization and search

**Tracing**:
- OpenTelemetry Collector for distributed tracing
- Configuration: `/home/fhcadmin/projects/devarch/config/otel-collector/otel-collector-config.yaml`

**Visualization**:
- Grafana dashboards for metrics
- Pre-configured data sources for Prometheus and Elasticsearch

## Key Directories

```
/home/fhcadmin/projects/devarch/
├── apps/                 # Application code (WordPress instances)
├── compose/              # Docker compose files organized by service category
│   ├── analytics/       # Monitoring services (Prometheus, Grafana, ELK)
│   ├── backend/         # Runtime environments (PHP, Node, Python, Go)
│   ├── database/        # Database services
│   ├── dbms/           # Database management tools
│   ├── exporters/      # Prometheus exporters
│   └── [...]           # Other service categories
├── config/              # Service-specific configurations and Dockerfiles
│   ├── php/            # PHP Dockerfile, php.ini, extensions
│   ├── nginx/          # Nginx configs, SSL certs
│   ├── prometheus/     # Prometheus scrape configs
│   └── [...]           # Other service configs
├── scripts/             # Management and automation scripts
│   ├── config.sh       # Central configuration (service discovery, runtime detection)
│   ├── service-manager.sh   # Main orchestration tool
│   ├── setup-databases.sh   # Database initialization
│   └── wordpress/      # WordPress installation scripts
└── logs/               # Application and service logs
```

## Development Workflow

### Standard Startup Sequence

1. **Start database layer** (required first):
   ```bash
   ./scripts/service-manager.sh start database
   ```

2. **Start proxy** (for routing):
   ```bash
   ./scripts/service-manager.sh start proxy
   ```

3. **Start backend runtimes**:
   ```bash
   ./scripts/service-manager.sh start backend
   ```

4. **Initialize databases** (if needed):
   ```bash
   ./scripts/setup-databases.sh -a
   ```

5. **Optional: Start observability stack**:
   ```bash
   ./scripts/service-manager.sh start exporters analytics
   ```

### Accessing Services

#### Infrastructure Services
- **Nginx Proxy Manager**: http://localhost:81 (admin interface)
- **Prometheus**: http://localhost:9090
- **Grafana**: Check .env for configured port
- **Metabase**: http://localhost:8085
- **Mailpit**: http://localhost:8025 (web UI)
- **WordPress apps**: https://<appname>.test (after configuring Nginx Proxy Manager)

#### Backend Service Ports

Each backend runtime has a dedicated 100-port range to eliminate conflicts and allow all services to run simultaneously:

**PHP Applications (8100-8199):**
- http://localhost:8100 - Laravel/PHP apps (internal: 8000)
- http://localhost:8102 - Vite dev server (internal: 5173)

**Node Applications (8200-8299):**
- http://localhost:8200 - Primary Node app (internal: 3000)
- http://localhost:8201 - Secondary Node app (internal: 3001)
- http://localhost:8202 - Vite dev server (internal: 5173)
- http://localhost:8203 - GraphQL server (internal: 4000)
- localhost:9229 - Node debugger (unchanged)

**Python Applications (8300-8399):**
- http://localhost:8300 - Django/FastAPI apps (internal: 8000)
- http://localhost:8301 - Flask apps (internal: 5000)
- http://localhost:8302 - Jupyter Lab (internal: 8888)
- http://localhost:8303 - Flower Celery monitoring (internal: 5555)

**Go Applications (8400-8499):**
- http://localhost:8400 - Go applications (internal: 8080)
- http://localhost:8401 - Metrics endpoint (internal: 8081)
- http://localhost:8402 - Delve debugger (internal: 2345)
- http://localhost:8403 - pprof profiling (internal: 6060)

This port allocation strategy ensures all backend services can run together without conflicts, supporting full microservices development workflows.

### Working with a Single Service

```bash
# Check if service is running
./scripts/service-manager.sh status <service>

# View logs
./scripts/service-manager.sh logs <service> --follow

# Restart after config changes
./scripts/service-manager.sh restart <service>

# Rebuild after Dockerfile changes
./scripts/service-manager.sh rebuild <service> --no-cache

# Access container shell
podman exec -it <service> bash
```

## Important Considerations

### Service Dependencies

- The **database** category must always start before other services that need data persistence
- **Exporters** require their corresponding services to be running (e.g., mysql-exporter needs mysql)
- **Analytics** services (Grafana) need exporters and Prometheus to be running for data
- Use `--wait-healthy` flag with start-all to ensure services are fully initialized before starting dependents

### Volume Persistence

- All databases use **named volumes** for data persistence (e.g., `mariadb_data`, `postgres_data`)
- Application code uses **bind mounts** to `/home/fhcadmin/projects/devarch/apps/`
- Volumes are preserved by default when stopping services
- Use `--remove-volumes` flag to delete volumes when needed

### Container Runtime Detection

The system automatically detects your container runtime setup:
- Checks for Podman first, falls back to Docker
- Detects if running rootless Podman (requires different network handling)
- Automatically determines if sudo is needed based on user groups
- You generally don't need to think about this - it just works

### Parallel vs Sequential Startup

```bash
# Sequential startup (default) - safer, respects dependencies
./scripts/service-manager.sh start-all

# Parallel startup - faster but may cause issues if dependencies aren't met
./scripts/service-manager.sh start-all --parallel

# Best practice: parallel within categories
./scripts/service-manager.sh start database --parallel
./scripts/service-manager.sh start backend --parallel
```

### Troubleshooting

If a service fails to start:
1. Check logs: `./scripts/service-manager.sh logs <service>`
2. Verify dependencies are running: `./scripts/service-manager.sh status`
3. Check if the network exists: `podman network ls | grep microservices-net`
4. Rebuild the service: `./scripts/service-manager.sh rebuild <service>`
5. Check compose file syntax: `podman-compose -f compose/<category>/<service>.yml config`
