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

The primary tool is the `devarch` command, a minimal transparent wrapper around podman-compose:

```bash
# Individual service operations
devarch start <service>              # Start a service
devarch stop <service>               # Stop a service
devarch restart <service>            # Restart a service
devarch logs <service> [-f]          # Show logs (use -f to follow)
devarch exec <service> <command>     # Execute command in container

# Convenience operations
devarch start-db                     # Start all database services
devarch start-backend                # Start all backend runtimes
devarch start-all                    # Start all services in dependency order
devarch stop-all                     # Stop all services

# Utility commands
devarch ps                           # List running DevArch containers
devarch status                       # Show status of all services
devarch list                         # List all available services
devarch network                      # Show/inspect network
devarch help                         # Show help

# Examples
devarch start postgres               # Start PostgreSQL
devarch logs nginx-proxy-manager -f  # Follow nginx logs
devarch exec php bash                # Open bash in PHP container
```

**Note:** Each command shows the exact podman operation being executed for transparency. See [docs/SERVICE_MANAGER.md](docs/SERVICE_MANAGER.md) for complete documentation.

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
devarch exec php wp plugin list
devarch exec php wp theme list
devarch exec php wp db query "SELECT * FROM wp_posts LIMIT 5"
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
devarch start-db                     # Start all database services
devarch start nginx-proxy-manager    # Start proxy
devarch start-backend                # Start backend runtimes

# 3. Initialize databases
./scripts/setup-databases.sh -a

# 4. Update hosts file
sudo ./scripts/update-hosts.sh

# 5. Check everything is running
devarch status
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

### Standardized Application Structure (CRITICAL)

**ALL APPLICATIONS** must follow the standardized `public/` directory pattern. This is a fundamental requirement.

#### Core Requirement

The `public/` subdirectory is the **web server document root** for all applications:

```
apps/{app-name}/public/  ← WEB SERVER SERVES FROM HERE
```

**Rationale**:
- Web server (Nginx/Apache) configured to serve from `/var/www/html/{app-name}/public/`
- Source code outside `public/` not accessible via HTTP (security)
- Container mounts expect `public/` as document root
- Consistent behavior across all frameworks

#### Standard Directory Structure

```
apps/{app-name}/
├── public/              # MANDATORY - Web server document root
│   ├── index.html       # Entry point (static/SPA apps)
│   ├── index.php        # Entry point (PHP apps)
│   ├── assets/          # Built assets (CSS, JS, images)
│   ├── api/             # API endpoints (optional)
│   └── .htaccess        # Server config (optional)
├── src/                 # Source code (for compiled apps)
├── config/              # Application configuration
├── scripts/             # Build and deployment scripts
├── tests/               # Test files
├── .env.example         # Environment variables template
├── .gitignore           # Git ignore patterns
├── package.json         # Node.js dependencies (if applicable)
├── composer.json        # PHP dependencies (if applicable)
├── requirements.txt     # Python dependencies (if applicable)
├── go.mod               # Go dependencies (if applicable)
└── README.md            # Application documentation
```

#### Framework Build Configuration

All build processes **MUST** output to `public/`:

**React + Vite**:
```javascript
// vite.config.js
build: {
  outDir: 'public',
  emptyOutDir: false,
}
```

**Next.js**:
```javascript
// next.config.js
module.exports = {
  distDir: 'public/.next',
  output: 'export',  // or 'standalone'
}
```

**Express**:
```javascript
// server.js
app.use(express.static('public'))
```

**Django**:
```python
# settings.py
STATIC_ROOT = os.path.join(BASE_DIR, 'public/static')
```

**Flask**:
```python
app = Flask(__name__, static_folder='public')
```

#### Creating New Applications

DevArch uses **JetBrains IDEs** for project creation (PHPStorm, WebStorm, PyCharm, GoLand, Rider).

**General Workflow**:
1. Open appropriate IDE (PHPStorm, WebStorm, PyCharm, GoLand, Rider)
2. File → New Project → Select framework
3. Location: `/home/fhcadmin/projects/devarch/apps/{app-name}`
4. Follow IDE-specific guides in `docs/jetbrains/`

**WordPress (Special Case)**:
WordPress uses custom installation script for custom repos/templates:

```bash
# Interactive mode
./scripts/wordpress/install-wordpress.sh

# Non-interactive with preset
./scripts/wordpress/install-wordpress.sh my-wp-site \
  --preset clean \
  --title "My WordPress Site"
```

**Why JetBrains IDEs?**
- Superior scaffolding and framework support
- Auto-updated project templates
- Full IDE integration (debugging, linting, testing)
- Better than maintaining custom templates

#### Documentation

- **App Structure Standard**: `docs/APP_STRUCTURE.md` - public/ directory requirement
- **JetBrains IDE Guides**: `docs/jetbrains/` - Framework-specific setup
  - PHPStorm: Laravel, WordPress integration
  - WebStorm: React, Next.js, Vue
  - PyCharm: Django, Flask, FastAPI
  - GoLand: Go applications
  - Rider: ASP.NET Core

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
   devarch start-db
   ```

2. **Start proxy** (for routing):
   ```bash
   devarch start nginx-proxy-manager
   ```

3. **Start backend runtimes**:
   ```bash
   devarch start-backend
   ```

4. **Initialize databases** (if needed):
   ```bash
   ./scripts/setup-databases.sh -a
   ```

5. **Optional: Start observability stack**:
   ```bash
   devarch start node-exporter
   devarch start postgres-exporter
   devarch start redis-exporter
   devarch start prometheus
   devarch start grafana
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
devarch status

# View logs
devarch logs <service> -f

# Restart after config changes
devarch restart <service>

# Rebuild after Dockerfile changes (use podman-compose directly)
podman-compose -f compose/<category>/<service>.yml down
podman-compose -f compose/<category>/<service>.yml build --no-cache
podman-compose -f compose/<category>/<service>.yml up -d

# Access container shell
devarch exec <service> bash
```

## Important Considerations

### Service Dependencies

- The **database** category must always start before other services that need data persistence
- **Exporters** require their corresponding services to be running (e.g., mysql-exporter needs mysql)
- **Analytics** services (Grafana) need exporters and Prometheus to be running for data
- `devarch start-all` handles dependency order automatically, but individual commands do not
- See [docs/SERVICE_MANAGER.md](docs/SERVICE_MANAGER.md) for detailed startup order recommendations

### Volume Persistence

- All databases use **named volumes** for data persistence (e.g., `mariadb_data`, `postgres_data`)
- Application code uses **bind mounts** to `/home/fhcadmin/projects/devarch/apps/`
- Volumes are preserved when stopping services (use podman commands directly to manage volumes)
- To remove volumes: `podman volume rm <volume-name>` or `podman volume prune`

### Container Runtime Detection

The system automatically detects your container runtime setup:
- Checks for Podman first, falls back to Docker
- Detects if running rootless Podman (requires different network handling)
- Automatically determines if sudo is needed based on user groups
- You generally don't need to think about this - it just works

### Troubleshooting

If a service fails to start:
1. Check logs: `devarch logs <service>`
2. Verify dependencies are running: `devarch status`
3. Check if the network exists: `devarch network`
4. Rebuild the service: Use podman-compose directly (see docs/SERVICE_MANAGER.md)
5. Check compose file syntax: `podman-compose -f compose/<category>/<service>.yml config`
