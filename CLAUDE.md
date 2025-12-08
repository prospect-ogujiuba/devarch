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

**Centralized Configuration**: `/home/{user}/projects/devarch/scripts/config.sh` (898 lines)
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

Applications share the same PHP container but have:
- Dedicated databases
- Isolated file structures
- Separate domain names (e.g., playground.test, b2bcnc.test)

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
