# Microservices Architecture Management System

## Overview

### System Architecture

This microservices architecture provides a comprehensive development and production environment with 20+ services organized into logical categories. The system features intelligent auto-detection capabilities that automatically configure different project types without manual intervention.

#### Service Categories

🗄️ Database Services
🛠️ Database Management
🚀 Backend Development
📧 Communication
📊 Analytics & Monitoring
📂 Project Management
🤖 AI & Automation
🔐 Security & Proxy

#### Key Features

- **Smart Detection**: Automatically detects and configures Laravel, Django, Next.js, Go modules, and more
- **Modular Architecture**: Each service in separate compose files for independent scaling
- **Cross-Platform**: Works with Docker and Podman on Linux, Windows (WSL), and macOS
- **Development Ready**: Hot reload, debugging, and development tools included
- **Production Ready**: SSL certificates, security hardening, and monitoring
- **Flexible Deployment**: Deploy all services or specific categories

### Directory Structure
microservices/
├── compose/                 # Docker Compose files organized by category
│   ├── analytics/          # Elasticsearch, Kibana, Grafana, etc.
│   ├── backend/            # PHP, Node.js, Python, Go, .NET
│   ├── database/           # PostgreSQL, MySQL, MongoDB, Redis
│   ├── dbms/              # Database management tools
│   ├── mail/              # Email services
│   ├── project/           # Project management tools
│   └── proxy/             # Reverse proxy and security
├── config/                 # Service configuration files
│   ├── nginx/             # Nginx proxy configuration
│   ├── php/               # PHP configuration and Dockerfile
│   ├── python/            # Python configuration and Dockerfile
│   └── ...                # Configuration for each service
├── scripts/               # Management and automation scripts
├── apps/                  # Your application code (auto-detected)
├── logs/                  # Service logs and debugging
├── ssl/                   # SSL certificates and keys
└── .env                   # Environment configuration

### Network Configuration

The system uses localhost-bound ports to avoid conflicts:

- **Database Services**: 8501-8505
- **Management Tools**: 8082-8087
- **Analytics**: 9001, 9090, 9120-9131
- **Proxy & Security**: 80, 81, 443, 9400-9401
- **Other Services**: 9100-9300

---

## Installation Guide

### Installation Methods

#### 1. Full Installation (Recommended)

Install all services with automatic setup:

```bash
# Complete installation with confirmation prompts
./scripts/install.sh

# Silent installation (no prompts)
./scripts/install.sh -y

# Verbose installation with detailed logging
./scripts/install.sh -y -v

# Installation with parallel service startup (faster)
./scripts/install.sh -y -p
```

#### 2. Category-Specific Installation

Install only specific service categories:

```bash
# Install only database services
./scripts/install.sh -c database -y

# Install database and management tools
./scripts/install.sh -c database -c dbms -y

# Install everything except AI services
./scripts/install.sh -x ai -y
```

#### 3. Custom Installation

```bash
# Skip SSL setup (for internal networks)
./scripts/install.sh -S -y

# Skip database initialization
./scripts/install.sh -D -y

# Use Podman instead of Docker
./scripts/install.sh -r podman -y

# Install with custom timeout
./scripts/install.sh -t 600 -y
```

### Installation Process

#### Phase 1: Prerequisites Check

- Validates container runtime (Docker/Podman)
- Checks available disk space (10GB minimum)
- Verifies compose file syntax
- Creates necessary directories

#### Phase 2: Network Setup

- Creates shared `microservices-net` network
- Configures bridge networking for inter-service communication

#### Phase 3: Service Deployment

- Deploys services in dependency order:
  1. **Database** → PostgreSQL, MySQL, MongoDB, Redis
  2. **DBMS** → Adminer, pgAdmin, Metabase
  3. **Backend** → Development environments
  4. **Analytics** → Monitoring and logging
  5. **Proxy** → Nginx Proxy Manager

#### Phase 4: Post-Installation Setup

- Database initialization and user creation
- SSL certificate generation
- Certificate trust store installation
- Health checks and verification

### Troubleshooting Installation

#### Common Issues

**Port Conflicts:**

```bash
# Check for port conflicts
netstat -tlnp | grep -E ':(80|443|8501|8502|9001)'

# Stop conflicting services
sudo systemctl stop apache2 nginx mysql postgresql
```

**Permission Issues:**

```bash
# Fix Docker permissions
sudo usermod -aG docker $USER
newgrp docker

# Fix directory permissions
sudo chown -R $USER:$USER ./microservices
```

**Container Runtime Issues:**

```bash
# Test Docker
docker run hello-world

# Test Podman
podman run hello-world

# Reset container environment
docker system prune -a  # WARNING: Removes all containers/images
```

#### Installation Logs

Check installation logs for detailed error information:

```bash
# View installation logs
tail -f logs/scripts.log

# Search for errors
grep -i error logs/scripts.log

# View specific service logs
./scripts/show-services.sh -L
```

---

## Script Reference

### config.sh - Core Configuration

**Purpose**: Central configuration and utility functions for all scripts.

**Key Functions**:

- Environment setup and validation
- Logging with multiple levels (DEBUG, INFO, WARN, ERROR)
- Container runtime detection (Docker/Podman)
- Network management
- Common utility functions

**Usage**: Automatically sourced by other scripts.

**Environment Variables**:

```bash
# Set log level
export MS_LOG_LEVEL=0  # DEBUG level

# Container runtime override
export CONTAINER_RUNTIME=podman
```

---

### install.sh - System Installation

**Purpose**: Complete system installation with intelligent service deployment.

**Syntax**:

```bash
./scripts/install.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-y` | Skip confirmation prompts | `./scripts/install.sh -y` |
| `-c CATEGORY` | Install specific category only | `./scripts/install.sh -c database` |
| `-x CATEGORY` | Exclude specific category | `./scripts/install.sh -x ai` |
| `-p` | Enable parallel installation | `./scripts/install.sh -p` |
| `-D` | Skip database setup | `./scripts/install.sh -D` |
| `-S` | Skip SSL certificate setup | `./scripts/install.sh -S` |
| `-T` | Skip trust certificate setup | `./scripts/install.sh -T` |
| `-t SECONDS` | Installation timeout | `./scripts/install.sh -t 600` |
| `-v` | Verbose output | `./scripts/install.sh -v` |
| `-d` | Dry run (show commands only) | `./scripts/install.sh -d` |

**Examples**:

```bash
# Complete installation
./scripts/install.sh -y

# Development environment only
./scripts/install.sh -c database -c dbms -c backend -y

# Production without AI services
./scripts/install.sh -x ai -y

# Fast parallel installation
./scripts/install.sh -y -p -t 600

# Debug installation issues
./scripts/install.sh -d -v
```

**Categories Available**:

- `database` - PostgreSQL, MySQL, MariaDB, MongoDB, Redis
- `dbms` - Database management interfaces
- `backend` - Development environments (PHP, Node.js, Python, Go, .NET)
- `analytics` - Monitoring and analytics tools
- `ai` - AI and workflow automation
- `mail` - Email testing services
- `project` - Project management tools
- `erp` - Business applications
- `proxy` - Reverse proxy and security

---

### start-services.sh - Service Startup

**Purpose**: Start microservices with dependency management and health checking.

**Syntax**:

```bash
./scripts/start-services.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-c CATEGORY` | Start specific category | `./scripts/start-services.sh -c database` |
| `-x CATEGORY` | Exclude category | `./scripts/start-services.sh -x ai` |
| `-R` | Restart services instead of start | `./scripts/start-services.sh -R` |
| `-H` | Skip health checks | `./scripts/start-services.sh -H` |
| `-w` | Wait for services to be healthy | `./scripts/start-services.sh -w` |

**Examples**:

```bash
# Start all services
./scripts/start-services.sh

# Start only databases
./scripts/start-services.sh -c database

# Restart all services
./scripts/start-services.sh -R

# Start with health monitoring
./scripts/start-services.sh -w -v

# Start everything except AI services
./scripts/start-services.sh -x ai
```

---

### stop-services.sh - Service Shutdown

**Purpose**: Gracefully stop services with cleanup options.

**Syntax**:

```bash
./scripts/stop-services.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-c CATEGORY` | Stop specific category | `./scripts/stop-services.sh -c backend` |
| `-x CATEGORY` | Exclude category | `./scripts/stop-services.sh -x database` |
| `-f` | Force stop (kill instead of graceful) | `./scripts/stop-services.sh -f` |
| `-V` | Remove volumes after stopping | `./scripts/stop-services.sh -V` |
| `-I` | Remove images after stopping | `./scripts/stop-services.sh -I` |
| `-o` | Remove orphaned containers | `./scripts/stop-services.sh -o` |

**Examples**:

```bash
# Stop all services gracefully
./scripts/stop-services.sh

# Force stop all services
./scripts/stop-services.sh -f

# Stop and cleanup volumes/images
./scripts/stop-services.sh -V -I

# Stop everything except databases
./scripts/stop-services.sh -x database
```

**⚠️ Warning**: The `-V` and `-I` options permanently delete data and images.

---

### show-services.sh - Service Information

**Purpose**: Display comprehensive service status, health, and connectivity information.

**Syntax**:

```bash
./scripts/show-services.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-S` | Show service status and health | `./scripts/show-services.sh -S` |
| `-H` | Show detailed health information | `./scripts/show-services.sh -H` |
| `-L` | Show recent logs | `./scripts/show-services.sh -L` |
| `-T` | Show resource usage statistics | `./scripts/show-services.sh -T` |
| `-R` | Show only running services | `./scripts/show-services.sh -R` |
| `-c CATEGORY` | Filter by category | `./scripts/show-services.sh -c database` |
| `-f FORMAT` | Output format (table/json/csv/markdown) | `./scripts/show-services.sh -f json` |
| `-i SECONDS` | Continuous monitoring | `./scripts/show-services.sh -i 5` |
| `-u` | Show URLs only (quick reference) | `./scripts/show-services.sh -u` |

**Examples**:

```bash
# Basic service overview
./scripts/show-services.sh

# Detailed health dashboard
./scripts/show-services.sh -H

# Resource usage monitoring
./scripts/show-services.sh -T

# Live monitoring (refresh every 5 seconds)
./scripts/show-services.sh -i 5 -S

# Quick URL reference
./scripts/show-services.sh -u

# JSON output for scripting
./scripts/show-services.sh -f json

# Database services only
./scripts/show-services.sh -c database -R
```

**Output Formats**:

- **Table**: Human-readable formatted table (default)
- **JSON**: Machine-readable for scripts and APIs
- **CSV**: Spreadsheet-compatible format
- **Markdown**: Documentation-friendly format

---

### setup-databases.sh - Database Configuration

**Purpose**: Initialize databases, create users, and set up schemas for all services.

**Syntax**:

```bash
./scripts/setup-databases.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-A` | Setup all databases | `./scripts/setup-databases.sh -A` |
| `-P` | Setup PostgreSQL only | `./scripts/setup-databases.sh -P` |
| `-M` | Setup MariaDB only | `./scripts/setup-databases.sh -M` |
| `-Y` | Setup MySQL only | `./scripts/setup-databases.sh -Y` |
| `-O` | Setup MongoDB only | `./scripts/setup-databases.sh -O` |
| `-R` | Setup Redis only | `./scripts/setup-databases.sh -R` |
| `-S` | Create sample data | `./scripts/setup-databases.sh -S` |
| `-B` | Backup existing databases | `./scripts/setup-databases.sh -B` |
| `-F FILE` | Restore from backup | `./scripts/setup-databases.sh -F backup.sql` |
| `-w SECONDS` | Wait time for readiness | `./scripts/setup-databases.sh -w 45` |

**Examples**:

```bash
# Setup all databases
./scripts/setup-databases.sh -A

# Setup with sample data
./scripts/setup-databases.sh -A -S

# Backup and setup PostgreSQL
./scripts/setup-databases.sh -P -B

# Setup only essential databases
./scripts/setup-databases.sh -P -M

# Restore from backup
./scripts/setup-databases.sh -P -F postgres_backup.sql
```

**Created Databases**:

| Service        | Database  | User           | Purpose               |
| -------------- | --------- | -------------- | --------------------- |
| **PostgreSQL** | metabase  | metabase_user  | Business Intelligence |
|                | nocodb    | nocodb_user    | No-code platform      |
| **MariaDB**    | npm       | npm_user       | Nginx Proxy Manager   |
|                | matomo    | matomo_user    | Web analytics         |
| **MySQL**      | backup_db | backup_user    | Backup storage        |
|                | analytics | analytics_user | Analytics data        |
| **MongoDB**    | logs      | logs_user      | Application logs      |
|                | analytics | analytics_user | Analytics data        |
|                | sessions  | sessions_user  | Session storage       |

---

### setup-ssl.sh - SSL Certificate Management

**Purpose**: Generate and manage SSL certificates for secure HTTPS access.

**Syntax**:

```bash
./scripts/setup-ssl.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-t TYPE` | Certificate type (wildcard/individual/letsencrypt) | `./scripts/setup-ssl.sh -t wildcard` |
| `-D DOMAIN` | Domain for certificate | `./scripts/setup-ssl.sh -D "*.test"` |
| `-f` | Force regenerate existing certificates | `./scripts/setup-ssl.sh -f` |
| `-V DAYS` | Certificate validity days | `./scripts/setup-ssl.sh -V 365` |
| `-K SIZE` | RSA key size | `./scripts/setup-ssl.sh -K 4096` |
| `-B` | Skip backup of existing certificates | `./scripts/setup-ssl.sh -B` |
| `-N` | Skip Nginx configuration | `./scripts/setup-ssl.sh -N` |
| `-T` | Skip certificate validation | `./scripts/setup-ssl.sh -T` |

**Certificate Types**:

1. **Wildcard** (Recommended for development):

   ```bash
   ./scripts/setup-ssl.sh -t wildcard -D "*.test"
   ```

2. **Individual** (Separate certificate per service):

   ```bash
   ./scripts/setup-ssl.sh -t individual
   ```

3. **Let's Encrypt** (Production with valid domain):
   ```bash
   ./scripts/setup-ssl.sh -t letsencrypt -D "api.mycompany.com"
   ```

**Examples**:

```bash
# Generate wildcard certificate for *.test
./scripts/setup-ssl.sh

# Force regenerate with custom validity
./scripts/setup-ssl.sh -f -V 365

# Production Let's Encrypt certificate
./scripts/setup-ssl.sh -t letsencrypt -D "myapp.com"

# Individual certificates for each service
./scripts/setup-ssl.sh -t individual -v
```

---

### trust-host.sh - Certificate Trust Management

**Purpose**: Install SSL certificates in system and browser trust stores for seamless HTTPS.

**Syntax**:

```bash
./scripts/trust-host.sh [OPTIONS]
```

**Options**:
| Option | Description | Example |
|--------|-------------|---------|
| `-L` | Trust on Linux/WSL | `./scripts/trust-host.sh -L` |
| `-W` | Trust on Windows | `./scripts/trust-host.sh -W` |
| `-M` | Trust on macOS | `./scripts/trust-host.sh -M` |
| `-F` | Trust in Firefox | `./scripts/trust-host.sh -F` |
| `-C` | Trust in Chrome | `./scripts/trust-host.sh -C` |
| `-H` | Update hosts file | `./scripts/trust-host.sh -H` |
| `-S SOURCE` | Certificate source (container/local/custom) | `./scripts/trust-host.sh -S local` |
| `-P PATH` | Custom certificate path | `./scripts/trust-host.sh -S custom -P /path/cert.pem` |
| `-R` | Remove trusted certificates | `./scripts/trust-host.sh -R` |
| `-l` | List trusted certificates | `./scripts/trust-host.sh -l` |

**Examples**:

```bash
# Auto-detect platform and install certificates
./scripts/trust-host.sh

# Windows + Firefox trust
./scripts/trust-host.sh -W -F

# Complete setup with hosts file
./scripts/trust-host.sh -L -F -C -H

# Use local certificate file
./scripts/trust-host.sh -S local -L

# Remove all trusted certificates
./scripts/trust-host.sh -R -v

# List currently trusted certificates
./scripts/trust-host.sh -l
```

**Platform Support**:

- **Linux**: System CA certificates + browser profiles
- **Windows**: Certificate store + browser configuration
- **macOS**: System keychain + browser profiles
- **WSL**: Both Linux and Windows trust stores

---

## Service Management

### Starting Services

#### Complete System Startup

```bash
# Start all services with health checks
./scripts/start-services.sh -w

# Start with verbose logging
./scripts/start-services.sh -v

# Restart all services
./scripts/start-services.sh -R
```

#### Category-Based Startup

```bash
# Start only database services
./scripts/start-services.sh -c database

# Start development environment
./scripts/start-services.sh -c database -c dbms -c backend

# Start monitoring stack
./scripts/start-services.sh -c analytics
```

#### Individual Service Management

```bash
# Start specific service
podman compose -f compose/postgres.yml up -d

# Start with logs
podman compose -f compose/grafana.yml up -d && podman logs -f grafana

# Start multiple related services
podman compose -f compose/postgres.yml -f compose/pgadmin.yml up -d
```

### Stopping Services

#### Graceful Shutdown

```bash
# Stop all services gracefully
./scripts/stop-services.sh

# Stop specific category
./scripts/stop-services.sh -c backend

# Stop all except databases
./scripts/stop-services.sh -x database
```

#### Force Shutdown and Cleanup

```bash
# Force stop all services
./scripts/stop-services.sh -f

# Stop and remove volumes (⚠️ DATA LOSS)
./scripts/stop-services.sh -V

# Complete cleanup
./scripts/stop-services.sh -f -V -I -o
```

### Health Monitoring

#### Service Status Overview

```bash
# Basic status
./scripts/show-services.sh

# Detailed health dashboard
./scripts/show-services.sh -H

# Only running services
./scripts/show-services.sh -R
```

#### Continuous Monitoring

```bash
# Live monitoring (refresh every 5 seconds)
./scripts/show-services.sh -i 5 -S

# Resource usage monitoring
./scripts/show-services.sh -T

# Health dashboard with auto-refresh
./scripts/show-services.sh -H -i 10
```

#### Individual Service Health

```bash
# Check specific service health
podman healthcheck run postgres

# View service logs
podman logs postgres --tail 50

# Follow logs in real-time
podman logs -f nginx-proxy-manager

# Check service resource usage
podman stats postgres --no-stream
```

### Log Management

#### Centralized Log Viewing

```bash
# View logs for all services
./scripts/show-services.sh -L

# View logs with more lines
./scripts/show-services.sh -L -l 100

# Follow logs in real-time
./scripts/show-services.sh -L -F
```

#### Service-Specific Logs

```bash
# Recent logs
podman logs postgres --tail 50

# Logs with timestamps
podman logs postgres --timestamps

# Follow logs
podman logs -f nginx-proxy-manager

# Logs from specific time
podman logs postgres --since "2024-01-01T00:00:00"
```

#### Log Analysis

```bash
# Search for errors
podman logs postgres 2>&1 | grep -i error

# Search across all services
./scripts/show-services.sh -L | grep -i "error\|fail\|exception"

# Export logs to file
podman logs postgres > logs/postgres-$(date +%Y%m%d).log
```

### Resource Monitoring

#### System Resource Usage

```bash
# Resource dashboard
./scripts/show-services.sh -T

# Live resource monitoring
podman stats

# Specific service resources
podman stats postgres grafana
```

#### Network Monitoring

```bash
# View network information
./scripts/show-services.sh -N

# Check network connectivity
podman network inspect microservices-net

# Test inter-service connectivity
podman exec postgres ping adminer
```

#### Volume Usage

```bash
# View volume information
./scripts/show-services.sh -V

# Check volume usage
podman system df

# Detailed volume inspection
podman volume inspect postgres_data
```

---

## Database Management

### Database Setup and Configuration

#### Initial Database Setup

```bash
# Setup all databases with sample data
./scripts/setup-databases.sh -A -S

# Setup specific databases
./scripts/setup-databases.sh -P -M  # PostgreSQL and MariaDB only

# Setup with backup of existing data
./scripts/setup-databases.sh -A -B
```

#### Database Connection Details

| Database       | Host      | Port | Default Credentials | Management URL          |
| -------------- | --------- | ---- | ------------------- | ----------------------- |
| **PostgreSQL** | localhost | 8502 | postgres / 123456   | https://pgadmin.test    |
| **MariaDB**    | localhost | 8501 | root / 123456       | https://phpmyadmin.test |
| **MySQL**      | localhost | 8505 | root / 123456       | https://phpmyadmin.test |
| **MongoDB**    | localhost | 8503 | root / 123456       | https://mongodb.test    |
| **Redis**      | localhost | 8504 | (no auth)           | (command line only)     |

#### Service-Specific Databases

**PostgreSQL Databases**:

```sql
-- Connect: psql -h localhost -p 8502 -U postgres
\l                          -- List databases
\c metabase                 -- Connect to metabase database
\dt                         -- List tables
\du                         -- List users
```

**MariaDB/MySQL Databases**:

```sql
-- Connect: mysql -h localhost -P 8501 -u root -p
SHOW DATABASES;             -- List databases
USE npm;                    -- Switch to npm database
SHOW TABLES;               -- List tables
SELECT User FROM mysql.user; -- List users
```

**MongoDB Databases**:

```javascript
// Connect: mongosh mongodb://root:123456@localhost:8503
show dbs                    // List databases
use logs                    // Switch to logs database
show collections           // List collections
db.stats()                 // Database statistics
```

### Backup and Restore Procedures

#### Automated Backups

```bash
# Backup before database setup
./scripts/setup-databases.sh -A -B

# Manual backup with timestamp
backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$backup_dir"

# PostgreSQL backup
podman exec postgres pg_dumpall -U postgres > "$backup_dir/postgres_full.sql"

# MariaDB backup
podman exec mariadb mysqldump --all-databases -u root -p123456 > "$backup_dir/mariadb_full.sql"

# MongoDB backup
podman exec mongodb mongodump --authenticationDatabase admin -u root -p 123456 --out "/dump/backup_$(date +%Y%m%d)"
```

#### Individual Database Backups

```bash
# PostgreSQL specific database
podman exec postgres pg_dump -U postgres metabase > backups/metabase_$(date +%Y%m%d).sql

# MariaDB specific database
podman exec mariadb mysqldump -u root -p123456 npm > backups/npm_$(date +%Y%m%d).sql

# MongoDB specific database
podman exec mongodb mongodump --authenticationDatabase admin -u root -p 123456 --db logs --out /dump/logs_backup
```

#### Restore Procedures

```bash
# Restore PostgreSQL
./scripts/setup-databases.sh -P -F backups/postgres_backup.sql

# Manual PostgreSQL restore
podman exec -i postgres psql -U postgres < backups/postgres_full.sql

# Manual MariaDB restore
podman exec -i mariadb mysql -u root -p123456 < backups/mariadb_full.sql

# MongoDB restore
podman exec mongodb mongorestore --authenticationDatabase admin -u root -p 123456 /dump/backup_20240101
```

### Sample Data Management

#### Creating Sample Data

```bash
# Generate sample data during setup
./scripts/setup-databases.sh -A -S

# Add custom sample data
cat << 'EOF' > sample_data.sql
-- PostgreSQL sample data
\c metabase;
INSERT INTO sample_metrics (name, value, created_at) VALUES
('Daily Users', 1500, NOW()),
('Revenue', 25000, NOW()),
('Conversion Rate', 3.5, NOW());
EOF

podman exec -i postgres psql -U postgres < sample_data.sql
```

#### Sample Data Structure

**PostgreSQL Sample Data**:

- `metabase.sample_data` - Sample metrics and KPIs
- `nocodb.projects` - Sample project data

**MariaDB Sample Data**:

- `npm.test_table` - Sample entries for testing
- `matomo.sample_visits` - Web analytics sample data

**MongoDB Sample Data**:

- `logs.application_logs` - Sample application log entries
- `analytics.events` - Sample event tracking data

### Database Administration

#### PostgreSQL Administration

```bash
# Connect to PostgreSQL
podman exec -it postgres psql -U postgres

# Create new database
CREATE DATABASE myapp;
CREATE USER myapp_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE myapp TO myapp_user;

# Performance monitoring
SELECT * FROM pg_stat_activity WHERE state = 'active';
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del
FROM pg_stat_user_tables;

# Backup specific database
podman exec postgres pg_dump -U postgres -d myapp > backups/myapp.sql
```

#### MySQL/MariaDB Administration

```bash
# Connect to MariaDB
podman exec -it mariadb mysql -u root -p

# Create new database
CREATE DATABASE myapp;
CREATE USER 'myapp_user'@'%' IDENTIFIED BY 'secure_password';
GRANT ALL PRIVILEGES ON myapp.* TO 'myapp_user'@'%';
FLUSH PRIVILEGES;

# Performance monitoring
SHOW PROCESSLIST;
SHOW STATUS LIKE 'Threads_%';
SHOW ENGINE INNODB STATUS\G

# Backup specific database
podman exec mariadb mysqldump -u root -p123456 myapp > backups/myapp.sql
```

#### MongoDB Administration

```javascript
// Connect to MongoDB
// mongosh mongodb://root:123456@localhost:8503

// Create new database and user
use myapp;
db.createUser({
  user: "myapp_user",
  pwd: "secure_password",
  roles: [
    { role: "readWrite", db: "myapp" }
  ]
});

// Performance monitoring
db.runCommand({ currentOp: 1 });
db.stats();
db.serverStatus();

// Backup specific database
// podman exec mongodb mongodump --authenticationDatabase admin -u root -p 123456 --db myapp --out /dump/myapp_backup
```

---

## SSL and Security

### SSL Certificate Generation

#### Wildcard Certificate (Recommended for Development)

```bash
# Generate wildcard certificate for *.test domains
./scripts/setup-ssl.sh -t wildcard

# Custom domain wildcard
./scripts/setup-ssl.sh -t wildcard -D "*.dev.local"

# Force regenerate existing certificate
./scripts/setup-ssl.sh -t wildcard -f
```

#### Individual Certificates

```bash
# Generate separate certificate for each service
./scripts/setup-ssl.sh -t individual

# Useful for production environments with specific domain requirements
./scripts/setup-ssl.sh -t individual -V 365
```

#### Let's Encrypt Certificates (Production)

```bash
# For production with valid domain
./scripts/setup-ssl.sh -t letsencrypt -D "api.mycompany.com"

# Multiple domains
./scripts/setup-ssl.sh -t letsencrypt -D "app.mycompany.com,api.mycompany.com"
```

### Certificate Trust Installation

#### Cross-Platform Trust Setup

```bash
# Auto-detect platform and install
./scripts/trust-host.sh

# Linux with browser support
./scripts/trust-host.sh -L -F -C

# Windows (WSL) with full browser support
./scripts/trust-host.sh -W -F -C

# macOS with browser support
./scripts/trust-host.sh -M -F -C
```

#### Platform-Specific Instructions

**Linux (Ubuntu/Debian)**:

```bash
# Install certificate in system trust store
./scripts/trust-host.sh -L

# Manual installation
sudo cp ssl/wildcard.test.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Verify installation
openssl x509 -in /usr/local/share/ca-certificates/wildcard.test.crt -text -noout
```

**Windows (WSL)**:

```bash
# Install in both Linux and Windows trust stores
./scripts/trust-host.sh -W -L

# Manual Windows installation (from PowerShell as Administrator)
Import-Certificate -FilePath "C:\path\to\wildcard.test.crt" -CertStoreLocation "Cert:\LocalMachine\Root"
```

**macOS**:

```bash
# Install in system keychain
./scripts/trust-host.sh -M

# Manual installation
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ssl/wildcard.test.crt

# Verify installation
security find-certificate -c "wildcard.test" /Library/Keychains/System.keychain
```

#### Browser-Specific Configuration

**Firefox**:

```bash
# Install certificate in Firefox profiles
./scripts/trust-host.sh -F

# Manual installation (requires libnss3-tools)
sudo apt install libnss3-tools
certutil -A -n "Microservices Wildcard" -t "TCu,Cu,Tu" -i ssl/wildcard.test.crt -d sql:$HOME/.mozilla/firefox/*.default*/

# Verify installation
certutil -L -d sql:$HOME/.mozilla/firefox/*.default*/
```

**Chrome/Chromium**:

```bash
# Setup Chrome certificate policy
./scripts/trust-host.sh -C

# Chrome uses system trust store on most platforms
# Restart Chrome after installing system certificates
```

### Hosts File Management

#### Update System Hosts File

```bash
# Add all microservice domains to /etc/hosts
./scripts/trust-host.sh -H

# Manual hosts file entries
sudo tee -a /etc/hosts << EOF
# Microservices Development
127.0.0.1 nginx.test
127.0.0.1 adminer.test
127.0.0.1 metabase.test
127.0.0.1 grafana.test
127.0.0.1 n8n.test
EOF
```

#### Service Domains

| Service         | Domain              | Purpose                      |
| --------------- | ------------------- | ---------------------------- |
| nginx.test      | Nginx Proxy Manager | Reverse proxy management     |
| adminer.test    | Adminer             | Universal database tool      |
| pgadmin.test    | pgAdmin             | PostgreSQL administration    |
| phpmyadmin.test | phpMyAdmin          | MySQL/MariaDB administration |
| mongodb.test    | Mongo Express       | MongoDB administration       |
| metabase.test   | Metabase            | Business intelligence        |
| nocodb.test     | NocoDB              | No-code database platform    |
| grafana.test    | Grafana             | Monitoring dashboards        |
| prometheus.test | Prometheus          | Metrics collection           |
| kibana.test     | Kibana              | Log analysis                 |
| n8n.test        | n8n                 | Workflow automation          |
| langflow.test   | Langflow            | AI workflow builder          |
| mailpit.test    | Mailpit             | Email testing                |
| gitea.test      | Gitea               | Git repository management    |

### Security Best Practices

#### Development Environment Security

```bash
# Change default passwords
sed -i 's/ADMIN_PASSWORD=123456/ADMIN_PASSWORD=your_secure_password/g' .env
sed -i 's/MYSQL_ROOT_PASSWORD=123456/MYSQL_ROOT_PASSWORD=your_secure_password/g' .env

# Generate secure tokens
openssl rand -hex 32  # For JWT secrets
openssl rand -base64 32  # For session secrets
```

#### Production Security Hardening

```bash
# 1. Update all default credentials in .env
# 2. Use strong passwords (16+ characters)
# 3. Enable firewall rules
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable

# 4. Restrict database access
# Edit compose files to remove external port mappings for databases

# 5. Enable container security
podman run --security-opt label=enable --security-opt no-new-privileges
```

#### Certificate Management

```bash
# Check certificate expiration
openssl x509 -in ssl/wildcard.test.crt -noout -dates

# Auto-renewal script for Let's Encrypt
cat << 'EOF' > scripts/renew-certificates.sh
#!/bin/bash
# Run monthly via cron
./scripts/setup-ssl.sh -t letsencrypt -f
./scripts/trust-host.sh
systemctl reload nginx
EOF

# Add to crontab
echo "0 2 1 * * /path/to/scripts/renew-certificates.sh" | crontab -
```

---

## Development Workflows

### Smart Backend Detection System

The system automatically detects and configures different project types without manual intervention.

#### PHP Projects

**Supported Frameworks**:

- Laravel (artisan commands)
- WordPress (wp-config detection)
- Symfony (bin/console commands)
- CodeIgniter (spark commands)
- Generic Composer projects

**Auto-Detection Examples**:

```bash
# Laravel Project
mkdir apps/my-laravel-app
cd apps/my-laravel-app
composer create-project laravel/laravel .
cd ../..

# Deploy (auto-detects Laravel, runs composer install, artisan commands)
podman compose -f compose/php.yml up -d

# Check detection logs
podman logs php
```

**Manual PHP Deployment**:

```bash
# Create PHP project structure
mkdir -p apps/my-php-app
cat << 'EOF' > apps/my-php-app/index.php
<?php
echo "Hello from PHP!";
phpinfo();
EOF

# Start PHP service (auto-detects static PHP)
podman compose -f compose/php.yml up -d

# Access at http://localhost:8000
```

#### Node.js Projects

**Supported Frameworks**:

- Next.js (automatic build detection)
- NestJS (TypeScript compilation)
- React (build scripts)
- Express.js
- Fastify
- Generic npm projects

**Auto-Detection Examples**:

```bash
# Next.js Project
mkdir apps/my-nextjs-app
cd apps/my-nextjs-app
npx create-next-app@latest . --typescript --tailwind --eslint
cd ../..

# Deploy (auto-detects Next.js, runs npm install, npm run build)
podman compose -f compose/node.yml up -d

# Express.js Project
mkdir apps/my-api
cat << 'EOF' > apps/my-api/package.json
{
  "name": "my-api",
  "version": "1.0.0",
  "main": "server.js",
  "scripts": {
    "start": "node server.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
EOF

cat << 'EOF' > apps/my-api/server.js
const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello from Express!' });
});

app.listen(port, '0.0.0.0', () => {
  console.log(`Server running on port ${port}`);
});
EOF

# Deploy (auto-detects Express, runs npm install)
podman compose -f compose/node.yml up -d
```

#### Python Projects

**Supported Frameworks**:

- Django (manage.py detection)
- FastAPI (automatic detection)
- Flask (app.py detection)
- Streamlit
- Poetry projects
- Pipenv projects

**Auto-Detection Examples**:

```bash
# Django Project
mkdir apps/my-django-app
cd apps/my-django-app
python -m django startproject myproject .
echo "Django>=4.0" > requirements.txt
cd ../..

# Deploy (auto-detects Django, runs pip install, migrations)
podman compose -f compose/python.yml up -d

# FastAPI Project
mkdir apps/my-fastapi-app
cat << 'EOF' > apps/my-fastapi-app/main.py
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello from FastAPI!"}
EOF

cat << 'EOF' > apps/my-fastapi-app/requirements.txt
fastapi>=0.68.0
uvicorn>=0.15.0
EOF

# Deploy (auto-detects FastAPI)
podman compose -f compose/python.yml up -d
```

#### Go Projects

**Supported Frameworks**:

- Gin framework
- Gorilla Mux
- Echo framework
- Fiber framework
- Standard net/http
- Cobra CLI applications

**Auto-Detection Examples**:

```bash
# Gin Project
mkdir apps/my-go-api
cd apps/my-go-api
go mod init my-go-api

cat << 'EOF' > main.go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Gin!",
        })
    })
    r.Run(":8080")
}
EOF

go mod tidy
cd ../..

# Deploy (auto-detects Gin framework)
podman compose -f compose/go.yml up -d
```

#### .NET Projects

**Supported Project Types**:

- ASP.NET Core (Web API, MVC, Blazor)
- Worker Services
- Console Applications
- Class Libraries

**Auto-Detection Examples**:

```bash
# ASP.NET Core Web API
mkdir apps/my-dotnet-api
cd apps/my-dotnet-api
dotnet new webapi
cd ../..

# Deploy (auto-detects ASP.NET Core, runs dotnet restore, dotnet build)
podman compose -f compose/dotnet.yml up -d

# Access at http://localhost:8010/swagger
```

### Application Deployment Patterns

#### Multi-Application Structure

```bash
# Organize multiple applications
apps/
├── frontend-react/         # React frontend
├── api-nodejs/            # Node.js API
├── admin-php/             # PHP admin panel
├── analytics-python/      # Python analytics service
├── gateway-go/           # Go API gateway
└── shared/               # Shared resources
    ├── configs/
    └── assets/
```

#### Environment-Specific Deployment

```bash
# Development environment
./scripts/start-services.sh -c backend -c database -c dbms

# Staging environment
./scripts/start-services.sh -x ai  # Skip AI services

# Production environment (minimal set)
./scripts/start-services.sh -c database -c proxy -c analytics
```

### Hot Reload and Development

#### Enable Development Mode

```bash
# Start services with development configurations
export ASPNETCORE_ENVIRONMENT=Development
export NODE_ENV=development
export FLASK_ENV=development

./scripts/start-services.sh -c backend
```

#### Volume Mounting for Hot Reload

All backend services automatically mount your application code:

```yaml
# Automatic volume mounting pattern
volumes:
  - ../apps:/var/www/html # PHP
  - ../apps:/app # Node.js, Python, Go, .NET
```

#### Debugging Applications

```bash
# View application logs
podman logs -f node
podman logs -f python
podman logs -f php

# Access container for debugging
podman exec -it node bash
podman exec -it python bash
podman exec -it php bash

# Debug with specific tools
podman exec node npm run debug
podman exec python python -m pdb main.py
```

### IDE Integration

#### VS Code Development

```bash
# Install Docker extension
# Create .vscode/launch.json for debugging

# Example Node.js debug configuration
cat << 'EOF' > .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "attach",
      "name": "Docker: Attach to Node",
      "remoteRoot": "/app",
      "localRoot": "${workspaceFolder}/apps/my-node-app",
      "port": 9229,
      "address": "localhost"
    }
  ]
}
EOF
```

#### Database Integration

```bash
# VS Code database connections
# Install SQLTools extension
# Add connection configurations:

# PostgreSQL connection
{
  "name": "PostgreSQL",
  "driver": "PostgreSQL",
  "server": "localhost",
  "port": 8502,
  "username": "postgres",
  "password": "123456",
  "database": "postgres"
}
```

---

## Production Considerations

### Security Hardening Checklist

#### 1. Credential Management

```bash
# Generate secure passwords
openssl rand -base64 32 > .secrets/db_password
openssl rand -base64 32 > .secrets/admin_password
openssl rand -hex 32 > .secrets/jwt_secret

# Update .env with secure credentials
sed -i "s/ADMIN_PASSWORD=123456/ADMIN_PASSWORD=$(cat .secrets/admin_password)/" .env
sed -i "s/MYSQL_ROOT_PASSWORD=123456/MYSQL_ROOT_PASSWORD=$(cat .secrets/db_password)/" .env
```

#### 2. Network Security

```bash
# Remove external database ports in production
# Edit compose files to remove port mappings:
# ports:
#   - "127.0.0.1:8502:5432"  # Remove this line

# Configure firewall
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw deny 8501     # Block direct database access
sudo ufw deny 8502
sudo ufw enable

# Use internal network communication
networks:
  microservices-net:
    external: true
    # Add network policies for production
```

#### 3. Container Security

```bash
# Run containers as non-root users (already configured)
# Enable security options
podman run --security-opt label=enable --security-opt no-new-privileges

# Use read-only root filesystems where possible
volumes:
  - type: bind
    source: ./apps
    target: /var/www/html
    read_only: true
```

#### 4. SSL/TLS Configuration

```bash
# Use Let's Encrypt for production
./scripts/setup-ssl.sh -t letsencrypt -D "yourdomain.com"

# Configure strong SSL settings in Nginx
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
```

### Performance Optimization

#### 1. Container Resource Limits

```yaml
# Add to compose files for production
services:
  postgres:
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: 2G
        reservations:
          cpus: "1"
          memory: 1G
```

#### 2. Database Optimization

```bash
# PostgreSQL optimization
cat << 'EOF' > config/postgres/postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
EOF

# MySQL optimization
cat << 'EOF' > config/mysql/my.cnf
[mysqld]
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
query_cache_size = 128M
query_cache_type = 1
EOF
```

#### 3. Application Performance

```bash
# Enable production builds
export NODE_ENV=production
export ASPNETCORE_ENVIRONMENT=Production

# PHP OPcache optimization
opcache.memory_consumption=256
opcache.interned_strings_buffer=16
opcache.max_accelerated_files=10000
opcache.revalidate_freq=0
opcache.validate_timestamps=0
```

### Monitoring and Alerting Setup

#### 1. Configure Prometheus Monitoring

```yaml
# config/prometheus/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "microservices"
    static_configs:
      - targets:
          - "postgres-exporter:9187"
          - "nginx-exporter:9113"
          - "node-exporter:9100"
```

#### 2. Grafana Dashboard Setup

```bash
# Import standard dashboards
curl -X POST \
  http://admin:123456@localhost:9001/api/dashboards/import \
  -H 'Content-Type: application/json' \
  -d '{
    "dashboard": {
      "id": 1860,
      "title": "Node Exporter Full"
    },
    "overwrite": true
  }'
```

#### 3. Log Aggregation

```bash
# Configure centralized logging
# Elasticsearch + Kibana + Logstash already included

# Configure application log shipping
# Add to application configurations:
# PHP: error_log = syslog
# Node.js: winston -> syslog transport
# Python: logging -> syslog handler
```

### Backup Strategies

#### 1. Automated Database Backups

```bash
# Create backup script
cat << 'EOF' > scripts/backup-production.sh
#!/bin/bash
BACKUP_DIR="/backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Database backups
podman exec postgres pg_dumpall -U postgres | gzip > "$BACKUP_DIR/postgres.sql.gz"
podman exec mariadb mysqldump --all-databases -u root -p$MYSQL_ROOT_PASSWORD | gzip > "$BACKUP_DIR/mariadb.sql.gz"
podman exec mongodb mongodump --authenticationDatabase admin -u root -p $MONGO_ROOT_PASSWORD --gzip --archive="$BACKUP_DIR/mongodb.archive.gz"

# Volume backups
tar -czf "$BACKUP_DIR/volumes.tar.gz" -C /var/lib/containers/storage/volumes .

# Retention (keep 7 days)
find /backups -type d -mtime +7 -exec rm -rf {} +
EOF

chmod +x scripts/backup-production.sh
```

#### 2. Schedule Regular Backups

```bash
# Add to crontab
echo "0 2 * * * /path/to/scripts/backup-production.sh" | crontab -

# Weekly full system backup
echo "0 3 * * 0 /path/to/scripts/backup-production.sh && rsync -av /backups/ backup-server:/remote/path/" | crontab -
```

### Disaster Recovery Planning

#### 1. Recovery Procedures

```bash
# Document recovery steps
cat << 'EOF' > docs/disaster-recovery.md
# Disaster Recovery Procedures

## Service Recovery
1. Install fresh system with same OS version
2. Install container runtime (Docker/Podman)
3. Restore project files from git repository
4. Restore latest backup
5. Run recovery script

## Database Recovery
1. Deploy database containers
2. Restore from latest backup
3. Verify data integrity
4. Update application configurations
EOF
```

#### 2. Recovery Testing

```bash
# Regular recovery testing
# 1. Deploy to test environment
# 2. Restore from production backup
# 3. Verify all services
# 4. Test application functionality
# 5. Document any issues

# Automated recovery testing
./scripts/install.sh -c database
./scripts/setup-databases.sh -F /backups/latest/postgres.sql.gz
./scripts/show-services.sh -H  # Verify health
```

---

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Port Conflicts

**Symptoms**:

- Services fail to start
- "Port already in use" errors
- Connection refused errors

**Diagnosis**:

```bash
# Check what's using specific ports
netstat -tlnp | grep -E ':(80|443|8501|8502|9001)'
ss -tlnp | grep -E ':(80|443|8501|8502|9001)'

# Check if services are already running
podman ps -a
```

**Solutions**:

```bash
# Stop conflicting services
sudo systemctl stop apache2 nginx mysql postgresql

# Kill processes using ports
sudo lsof -ti:80 | xargs kill -9
sudo lsof -ti:443 | xargs kill -9

# Change ports in compose files if needed
sed -i 's/8501:3306/8506:3306/' compose/mariadb.yml
```

#### 2. Container Startup Failures

**Symptoms**:

- Containers exit immediately
- "Container not found" errors
- Services show as unhealthy

**Diagnosis**:

```bash
# Check container status
podman ps -a

# View container logs
podman logs container_name

# Inspect container configuration
podman inspect container_name

# Check system resources
df -h  # Disk space
free -h  # Memory
```

**Solutions**:

```bash
# Remove and recreate containers
podman stop container_name
podman rm container_name
podman compose -f compose/service.yml up -d

# Clear container cache
podman system prune -a

# Fix permission issues
sudo chown -R $USER:$USER ./microservices
```

#### 3. Database Connection Issues

**Symptoms**:

- Applications can't connect to database
- Authentication failures
- Connection timeouts

**Diagnosis**:

```bash
# Test database connectivity
podman exec postgres pg_isready -U postgres
podman exec mariadb mysqladmin ping -h localhost

# Check database logs
podman logs postgres
podman logs mariadb

# Test from application container
podman exec node ping postgres
podman exec php telnet mariadb 3306
```

**Solutions**:

```bash
# Reset database passwords
./scripts/setup-databases.sh -A

# Recreate database containers
podman compose -f compose/postgres.yml down
podman volume rm postgres_data  # WARNING: Data loss
podman compose -f compose/postgres.yml up -d

# Check network connectivity
podman network inspect microservices-net
```

#### 4. SSL Certificate Issues

**Symptoms**:

- "Certificate not trusted" warnings
- HTTPS connections fail
- Browser security errors

**Diagnosis**:

```bash
# Check certificate validity
openssl x509 -in ssl/wildcard.test.crt -noout -dates -subject

# Test SSL connection
openssl s_client -connect nginx.test:443

# Check certificate installation
./scripts/trust-host.sh -l
```

**Solutions**:

```bash
# Regenerate certificates
./scripts/setup-ssl.sh -f

# Reinstall trust certificates
./scripts/trust-host.sh -R  # Remove old
./scripts/trust-host.sh     # Install new

# Clear browser cache
# Chrome: chrome://settings/certificates
# Firefox: about:preferences#privacy -> Certificates
```

#### 5. Network Connectivity Problems

**Symptoms**:

- Services can't communicate
- DNS resolution failures
- Proxy configuration issues

**Diagnosis**:

```bash
# Check network configuration
podman network ls
podman network inspect microservices-net

# Test inter-service connectivity
podman exec nginx-proxy-manager ping postgres
podman exec adminer nslookup postgres

# Check proxy configuration
podman logs nginx-proxy-manager
```

**Solutions**:

```bash
# Recreate network
podman network rm microservices-net
podman network create --driver bridge microservices-net

# Restart services
./scripts/stop-services.sh
./scripts/start-services.sh

# Reset proxy configuration
# Access http://localhost:81 and reconfigure proxy hosts
```

### Log Analysis Techniques

#### 1. Centralized Log Viewing

```bash
# View all service logs
./scripts/show-services.sh -L

# Follow logs in real-time
./scripts/show-services.sh -L -F

# Filter logs by service category
./scripts/show-services.sh -c database -L
```

#### 2. Error Pattern Analysis

```bash
# Search for errors across all services
podman ps --format "table {{.Names}}" | grep -v NAMES | while read container; do
  echo "=== $container ==="
  podman logs "$container" 2>&1 | grep -i -E "error|fail|exception|panic" | tail -5
done

# Analyze error patterns
grep -r "ERROR" logs/ | cut -d: -f3- | sort | uniq -c | sort -nr

# Export logs for external analysis
./scripts/show-services.sh -f json > logs/service-status-$(date +%Y%m%d).json
```

#### 3. Performance Log Analysis

```bash
# Identify slow queries (PostgreSQL)
podman exec postgres psql -U postgres -c "
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC LIMIT 10;"

# Analyze web server logs
podman logs nginx-proxy-manager | awk '{print $1}' | sort | uniq -c | sort -nr

# Monitor resource usage over time
while true; do
  echo "$(date): $(podman stats --no-stream --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}')"
  sleep 60
done > logs/resource-usage.log
```

### Recovery Procedures

#### 1. Service Recovery

```bash
# Individual service recovery
./scripts/stop-services.sh -c backend
./scripts/start-services.sh -c backend

# Full system recovery
./scripts/stop-services.sh
podman system prune -f
./scripts/start-services.sh

# Recovery with data preservation
./scripts/setup-databases.sh -B  # Backup first
./scripts/stop-services.sh
./scripts/start-services.sh
./scripts/setup-databases.sh -F backup_file.sql  # Restore if needed
```

#### 2. Data Recovery

```bash
# Database recovery from backup
./scripts/setup-databases.sh -F /backups/latest/postgres.sql

# Volume recovery
podman volume create postgres_data
tar -xzf /backups/latest/volumes.tar.gz -C /var/lib/containers/storage/volumes/

# Application code recovery from git
git stash  # Save local changes
git pull origin main  # Get latest code
git stash pop  # Restore local changes
```

#### 3. Network Recovery

```bash
# Recreate network infrastructure
podman network rm microservices-net
podman network create --driver bridge microservices-net

# Restart all services with network
./scripts/stop-services.sh
./scripts/start-services.sh

# Verify network connectivity
podman exec postgres ping adminer
podman exec nginx-proxy-manager ping grafana
```

### Performance Troubleshooting

#### 1. Identify Resource Bottlenecks

```bash
# Monitor system resources
htop  # CPU and memory usage
iotop  # Disk I/O usage
nethogs  # Network usage per process

# Container-specific monitoring
podman stats  # Live container statistics
./scripts/show-services.sh -T  # Service resource dashboard
```

#### 2. Database Performance Issues

```bash
# PostgreSQL performance analysis
podman exec postgres psql -U postgres -c "
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, n_tup_hot_upd
FROM pg_stat_user_tables
ORDER BY n_tup_ins + n_tup_upd + n_tup_del DESC;"

# MySQL performance analysis
podman exec mariadb mysql -u root -p123456 -e "
SHOW PROCESSLIST;
SELECT * FROM information_schema.INNODB_TRX;
SHOW ENGINE INNODB STATUS\G"

# MongoDB performance analysis
podman exec mongodb mongosh --authenticationDatabase admin -u root -p 123456 --eval "
db.runCommand({currentOp: 1});
db.serverStatus().opcounters;
db.stats();"
```

#### 3. Application Performance Optimization

```bash
# Identify slow applications
./scripts/show-services.sh -T | grep -E "CPU|Memory" | sort -k2 -nr

# Profile application performance
podman exec node npm run profile  # Node.js profiling
podman exec python python -m cProfile main.py  # Python profiling

# Optimize container resources
# Edit compose files to add resource limits
services:
  service_name:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

---

## Advanced Usage

### Customizing the System

#### 1. Adding New Services

**Create a new service compose file**:

```yaml
# compose/redis-commander.yml
version: "3.8"

networks:
  microservices-net:
    external: true

services:
  redis-commander:
    container_name: redis-commander
    image: rediscommander/redis-commander:latest
    restart: unless-stopped
    env_file: ../.env
    environment:
      - REDIS_HOSTS=local:redis:6379
    ports:
      - "127.0.0.1:8088:8081"
    networks:
      - microservices-net
```

**Add to service categories**:

```bash
# Edit scripts/config.sh
SERVICE_CATEGORIES=(
    ["database"]="postgres.yml mysql.yml mariadb.yml mongodb.yml redis.yml"
    ["dbms"]="adminer.yml pgadmin.yml phpmyadmin.yml mongo-express.yml metabase.yml nocodb.yml redis-commander.yml"
    # ... other categories
)
```

#### 2. Custom Backend Language Support

**Create new backend service**:

```dockerfile
# config/rust/Dockerfile
FROM rust:1.70

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git curl zsh vim nano \
    && rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install Rust tools
RUN cargo install cargo-watch cargo-edit

WORKDIR /app

# Copy smart entrypoint
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/smart-entrypoint.sh"]
CMD ["cargo", "run"]
```

**Smart entrypoint for Rust**:

```bash
#!/bin/bash
# config/rust/smart-entrypoint.sh

echo "🦀 Rust Smart Entrypoint: Detecting project structure..."

detect_rust_projects() {
    for app_dir in /app/*/; do
        if [ -d "$app_dir" ]; then
            app_name=$(basename "$app_dir")
            echo "📁 Found app: $app_name"

            cd "$app_dir"

            if [ -f "Cargo.toml" ]; then
                echo "📦 Rust project detected in $app_name"

                # Check for web frameworks
                if grep -q "actix-web" Cargo.toml; then
                    echo "🚀 Actix-web framework detected"
                elif grep -q "warp" Cargo.toml; then
                    echo "⚡ Warp framework detected"
                elif grep -q "rocket" Cargo.toml; then
                    echo "🚀 Rocket framework detected"
                fi

                # Build in release mode for production
                if [ "$RUST_ENV" = "production" ]; then
                    cargo build --release
                else
                    cargo check
                fi
            fi

            chown -R app:app "$app_dir"
        fi
    done
}

detect_rust_projects
echo "✅ Rust Smart Entrypoint: Setup complete!"

exec "$@"
```

#### 3. Environment-Specific Configurations

**Create environment overlay files**:

```yaml
# compose/overrides/development.yml
version: "3.8"

services:
  postgres:
    environment:
      - POSTGRES_LOG_STATEMENT=all
    volumes:
      - ../../logs/postgres:/var/log/postgresql

  node:
    environment:
      - NODE_ENV=development
      - DEBUG=*
    command: ["npm", "run", "dev"]

  python:
    environment:
      - FLASK_ENV=development
      - DJANGO_DEBUG=True
    command: ["python", "manage.py", "runserver", "0.0.0.0:8000"]
```

**Use with docker compose**:

```bash
# Deploy with development overrides
podman compose -f compose/postgres.yml -f compose/overrides/development.yml up -d

# Production deployment
podman compose -f compose/postgres.yml -f compose/overrides/production.yml up -d
```

### Automation and Scripting

#### 1. Custom Deployment Scripts

**Application-specific deployment**:

```bash
#!/bin/bash
# scripts/deploy-app.sh

APP_NAME="$1"
APP_TYPE="$2"

if [ -z "$APP_NAME" ] || [ -z "$APP_TYPE" ]; then
    echo "Usage: $0 <app_name> <app_type>"
    echo "Types: react, vue, laravel, django, express, fastapi"
    exit 1
fi

case "$APP_TYPE" in
    "react")
        npx create-react-app "apps/$APP_NAME"
        podman compose -f compose/node.yml up -d
        ;;
    "laravel")
        composer create-project laravel/laravel "apps/$APP_NAME"
        podman compose -f compose/php.yml up -d
        ;;
    "django")
        mkdir -p "apps/$APP_NAME"
        cd "apps/$APP_NAME"
        python -m django startproject . .
        echo "Django>=4.0" > requirements.txt
        cd ../..
        podman compose -f compose/python.yml up -d
        ;;
    *)
        echo "Unknown app type: $APP_TYPE"
        exit 1
        ;;
esac

echo "✅ $APP_NAME ($APP_TYPE) deployed successfully!"
```

#### 2. Health Check Automation

**Advanced health monitoring**:

```bash
#!/bin/bash
# scripts/health-monitor.sh

WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
CHECK_INTERVAL=300  # 5 minutes

monitor_services() {
    while true; do
        echo "$(date): Starting health check cycle..."

        # Get service status
        services_json=$(./scripts/show-services.sh -f json)

        # Check for unhealthy services
        unhealthy=$(echo "$services_json" | jq -r '.microservices.services[] | select(.health == "unhealthy") | .name')

        if [ -n "$unhealthy" ]; then
            message="🚨 Unhealthy services detected: $unhealthy"
            echo "$message"

            # Send alert (Slack, email, etc.)
            curl -X POST -H 'Content-type: application/json' \
                --data "{\"text\":\"$message\"}" \
                "$WEBHOOK_URL"

            # Attempt automatic recovery
            for service in $unhealthy; do
                echo "Attempting to restart $service..."
                podman restart "$service"
            done
        fi

        sleep $CHECK_INTERVAL
    done
}

# Run as daemon
if [ "$1" = "daemon" ]; then
    monitor_services > logs/health-monitor.log 2>&1 &
    echo $! > /tmp/health-monitor.pid
    echo "Health monitor started as daemon (PID: $(cat /tmp/health-monitor.pid))"
else
    monitor_services
fi
```

#### 3. Backup Automation

**Intelligent backup script**:

```bash
#!/bin/bash
# scripts/smart-backup.sh

BACKUP_ROOT="/backups"
RETENTION_DAYS=30
S3_BUCKET="microservices-backups"

create_backup() {
    local backup_type="$1"  # full, incremental, databases-only
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="$BACKUP_ROOT/$backup_type/$timestamp"

    mkdir -p "$backup_dir"

    case "$backup_type" in
        "full")
            echo "Creating full system backup..."

            # Database backups
            backup_databases "$backup_dir"

            # Volume backups
            backup_volumes "$backup_dir"

            # Configuration backups
            tar -czf "$backup_dir/configs.tar.gz" compose/ config/ scripts/ .env

            # Application code (if not in git)
            if [ -d "apps" ]; then
                tar --exclude=node_modules --exclude=.git \
                    -czf "$backup_dir/apps.tar.gz" apps/
            fi
            ;;

        "incremental")
            echo "Creating incremental backup..."
            # Compare with last backup and only backup changes
            ;;

        "databases-only")
            echo "Creating database-only backup..."
            backup_databases "$backup_dir"
            ;;
    esac

    # Compress and encrypt backup
    cd "$backup_dir/.."
    tar -czf "${timestamp}.tar.gz" "$timestamp"
    gpg --symmetric --cipher-algo AES256 "${timestamp}.tar.gz"
    rm -rf "$timestamp" "${timestamp}.tar.gz"

    # Upload to cloud storage
    if command -v aws >/dev/null 2>&1; then
        aws s3 cp "${timestamp}.tar.gz.gpg" "s3://$S3_BUCKET/$backup_type/"
    fi

    echo "✅ Backup completed: $backup_dir"
}

backup_databases() {
    local backup_dir="$1"

    # PostgreSQL
    if podman container exists postgres; then
        podman exec postgres pg_dumpall -U postgres | gzip > "$backup_dir/postgres.sql.gz"
    fi

    # MariaDB
    if podman container exists mariadb; then
        podman exec mariadb mysqldump --all-databases -u root -p$MYSQL_ROOT_PASSWORD | gzip > "$backup_dir/mariadb.sql.gz"
    fi

    # MongoDB
    if podman container exists mongodb; then
        podman exec mongodb mongodump --authenticationDatabase admin -u root -p $MONGO_ROOT_PASSWORD --gzip --archive="$backup_dir/mongodb.archive.gz"
    fi
}

backup_volumes() {
    local backup_dir="$1"

    # Get list of volumes
    volumes=$(podman volume ls --format "{{.Name}}" | grep -E "(postgres|mariadb|mongodb|grafana|prometheus)_data")

    for volume in $volumes; do
        echo "Backing up volume: $volume"
        podman run --rm -v "$volume:/source:ro" -v "$backup_dir:/backup" \
            alpine tar -czf "/backup/$volume.tar.gz" -C /source .
    done
}

cleanup_old_backups() {
    find "$BACKUP_ROOT" -type f -name "*.tar.gz.gpg" -mtime +$RETENTION_DAYS -delete
}

# Schedule different backup types
case "${1:-full}" in
    "full")
        create_backup "full"
        ;;
    "incremental")
        create_backup "incremental"
        ;;
    "databases")
        create_backup "databases-only"
        ;;
    "cleanup")
        cleanup_old_backups
        ;;
    *)
        echo "Usage: $0 {full|incremental|databases|cleanup}"
        exit 1
        ;;
esac
```

### CI/CD Pipeline Integration

#### 1. GitLab CI Configuration

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  CONTAINER_REGISTRY: registry.gitlab.com
  MICROSERVICES_ENV: staging

before_script:
  - echo $CI_REGISTRY_PASSWORD | podman login -u $CI_REGISTRY_USER --password-stdin $CI_REGISTRY

test:
  stage: test
  script:
    - ./scripts/install.sh -c database -c backend -y
    - ./scripts/show-services.sh -H
    -  # Run application tests
  only:
    - merge_requests
    - main

build:
  stage: build
  script:
    -  # Build application containers
    - podman build -t $CONTAINER_REGISTRY/$CI_PROJECT_PATH/app:$CI_COMMIT_SHA apps/
    - podman push $CONTAINER_REGISTRY/$CI_PROJECT_PATH/app:$CI_COMMIT_SHA
  only:
    - main

deploy_staging:
  stage: deploy
  script:
    - ./scripts/stop-services.sh -x database # Keep databases running
    - ./scripts/start-services.sh -c backend
    - ./scripts/show-services.sh -H
  environment:
    name: staging
    url: https://staging.myapp.com
  only:
    - main

deploy_production:
  stage: deploy
  script:
    - ./scripts/smart-backup.sh full
    - ./scripts/start-services.sh
    - ./scripts/show-services.sh -H
  environment:
    name: production
    url: https://myapp.com
  when: manual
  only:
    - main
```

#### 2. GitHub Actions Workflow

```yaml
# .github/workflows/microservices.yml
name: Microservices CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup environment
        run: |
          sudo apt-get update
          sudo apt-get install -y podman

      - name: Install microservices
        run: |
          ./scripts/install.sh -c database -c backend -y

      - name: Run health checks
        run: |
          sleep 30  # Wait for services to start
          ./scripts/show-services.sh -H

      - name: Run application tests
        run: |
          # Add your test commands here
          ./scripts/show-services.sh -c backend -R

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3

      - name: Deploy to production
        env:
          SSH_PRIVATE_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
          PRODUCTION_HOST: ${{ secrets.PRODUCTION_HOST }}
        run: |
          echo "$SSH_PRIVATE_KEY" > /tmp/ssh_key
          chmod 600 /tmp/ssh_key

          ssh -i /tmp/ssh_key -o StrictHostKeyChecking=no user@$PRODUCTION_HOST '
            cd /opt/microservices
            git pull origin main
            ./scripts/smart-backup.sh full
            ./scripts/start-services.sh
            ./scripts/show-services.sh -H
          '
```

#### 3. Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any

    environment {
        MICROSERVICES_PATH = '/opt/microservices'
        BACKUP_PATH = '/backups'
    }

    stages {
        stage('Checkout') {
            steps {
                git branch: 'main', url: 'https://github.com/your-org/microservices.git'
            }
        }

        stage('Test') {
            steps {
                sh '''
                    cd $MICROSERVICES_PATH
                    ./scripts/install.sh -c database -c backend -y
                    sleep 30
                    ./scripts/show-services.sh -H
                '''
            }
        }

        stage('Backup') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    cd $MICROSERVICES_PATH
                    ./scripts/smart-backup.sh full
                '''
            }
        }

        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    cd $MICROSERVICES_PATH
                    ./scripts/start-services.sh
                    ./scripts/show-services.sh -H
                '''
            }
            post {
                success {
                    slackSend(
                        color: 'good',
                        message: "✅ Microservices deployment successful: ${env.BUILD_URL}"
                    )
                }
                failure {
                    slackSend(
                        color: 'danger',
                        message: "❌ Microservices deployment failed: ${env.BUILD_URL}"
                    )
                }
            }
        }
    }

    post {
        always {
            sh '''
                cd $MICROSERVICES_PATH
                ./scripts/show-services.sh -f json > build_${BUILD_NUMBER}_status.json
            '''
            archiveArtifacts artifacts: 'build_*_status.json', allowEmptyArchive: true
        }
    }
}
```

### Multi-Environment Management

#### 1. Environment Separation

**Directory structure for multiple environments**:

```
microservices/
├── environments/
│   ├── development/
│   │   ├── .env
│   │   └── docker-compose.override.yml
│   ├── staging/
│   │   ├── .env
│   │   └── docker-compose.override.yml
│   └── production/
│       ├── .env
│       └── docker-compose.override.yml
├── scripts/
│   └── deploy-env.sh
└── ...
```

**Environment deployment script**:

```bash
#!/bin/bash
# scripts/deploy-env.sh

ENVIRONMENT="$1"
ACTION="${2:-start}"

if [ -z "$ENVIRONMENT" ]; then
    echo "Usage: $0 <environment> [start|stop|status]"
    echo "Environments: development, staging, production"
    exit 1
fi

ENV_DIR="environments/$ENVIRONMENT"

if [ ! -d "$ENV_DIR" ]; then
    echo "Environment '$ENVIRONMENT' not found"
    exit 1
fi

# Load environment-specific configuration
export ENV_FILE="$ENV_DIR/.env"
export COMPOSE_PROJECT_NAME="microservices_$ENVIRONMENT"

case "$ACTION" in
    "start")
        echo "🚀 Starting $ENVIRONMENT environment..."

        # Copy environment configuration
        cp "$ENV_FILE" .env

        # Start services with environment-specific overrides
        if [ -f "$ENV_DIR/docker-compose.override.yml" ]; then
            export COMPOSE_FILE="compose/postgres.yml:compose/mariadb.yml:$ENV_DIR/docker-compose.override.yml"
        fi

        ./scripts/start-services.sh
        ;;

    "stop")
        echo "🛑 Stopping $ENVIRONMENT environment..."
        ./scripts/stop-services.sh
        ;;

    "status")
        echo "📊 Status of $ENVIRONMENT environment..."
        ./scripts/show-services.sh
        ;;

    *)
        echo "Unknown action: $ACTION"
        exit 1
        ;;
esac
```

#### 2. Environment-Specific Configurations

**Development environment**:

```bash
# environments/development/.env
ADMIN_PASSWORD=dev123
MYSQL_ROOT_PASSWORD=dev123
DEBUG=true
LOG_LEVEL=debug

# Development-specific services
ENABLE_HOT_RELOAD=true
ENABLE_DEBUG_PORTS=true
```

**Production environment**:

```bash
# environments/production/.env
ADMIN_PASSWORD=complex_secure_password_here
MYSQL_ROOT_PASSWORD=another_complex_password_here
DEBUG=false
LOG_LEVEL=error

# Production-specific settings
ENABLE_SSL=true
ENABLE_MONITORING=true
BACKUP_ENABLED=true
```

---

## Reference Materials

### Complete Service Listing

#### Database Services

| Service    | Container | Internal Port | External Port | URL | Default Credentials |
| ---------- | --------- | ------------- | ------------- | --- | ------------------- |
| PostgreSQL | postgres  | 5432          | 8502          | N/A | postgres / 123456   |
| MariaDB    | mariadb   | 3306          | 8501          | N/A | root / 123456       |
| MySQL      | mysql     | 3306          | 8505          | N/A | root / 123456       |
| MongoDB    | mongodb   | 27017         | 8503          | N/A | root / 123456       |
| Redis      | redis     | 6379          | 8504          | N/A | No authentication   |

#### Database Management Tools

| Service       | Container     | Internal Port | External Port | URL                     | Default Credentials      |
| ------------- | ------------- | ------------- | ------------- | ----------------------- | ------------------------ |
| Adminer       | adminer       | 8080          | 8082          | https://adminer.test    | N/A                      |
| pgAdmin       | pgadmin       | 80            | 8087          | https://pgadmin.test    | admin@site.test / 123456 |
| phpMyAdmin    | phpmyadmin    | 80            | 8083          | https://phpmyadmin.test | N/A                      |
| Mongo Express | mongo-express | 8081          | 8084          | https://mongodb.test    | admin / 123456           |
| Metabase      | metabase      | 3000          | 8085          | https://metabase.test   | Setup required           |
| NocoDB        | nocodb        | 8080          | 8086          | https://nocodb.test     | Setup required           |

#### Backend Development Environments

| Service | Container | Internal Port | External Port | URL                   | Features                                   |
| ------- | --------- | ------------- | ------------- | --------------------- | ------------------------------------------ |
| PHP     | php       | 8000          | 8000          | http://localhost:8000 | Laravel, WordPress, Symfony auto-detection |
| Node.js | node      | 3000          | 8030          | http://localhost:8030 | Next.js, React, Express auto-detection     |
| Python  | python    | 8000          | 8040          | http://localhost:8040 | Django, FastAPI, Flask auto-detection      |
| Go      | go        | 8080          | 8020          | http://localhost:8020 | Gin, Echo, Fiber auto-detection            |
| .NET    | dotnet    | 80/443        | 8010/8011     | http://localhost:8010 | ASP.NET Core auto-detection                |

#### Analytics & Monitoring

| Service       | Container     | Internal Port  | External Port  | URL                        | Default Credentials |
| ------------- | ------------- | -------------- | -------------- | -------------------------- | ------------------- |
| Grafana       | grafana       | 3000           | 9001           | https://grafana.test       | admin / 123456      |
| Prometheus    | prometheus    | 9090           | 9090           | https://prometheus.test    | N/A                 |
| Elasticsearch | elasticsearch | 9200/9300      | 9130/9131      | https://elasticsearch.test | N/A                 |
| Kibana        | kibana        | 5601           | 9120           | https://kibana.test        | N/A                 |
| Logstash      | logstash      | 5000/5044/9600 | 5000/5044/9600 | N/A                        | N/A                 |
| Matomo        | matomo        | 80             | 9010           | https://matomo.test        | Setup required      |

#### AI & Automation

| Service  | Container | Internal Port | External Port | URL                   | Default Credentials |
| -------- | --------- | ------------- | ------------- | --------------------- | ------------------- |
| n8n      | n8n       | 5678          | 9100          | https://n8n.test      | Setup required      |
| Langflow | langflow  | 7860          | 9110          | https://langflow.test | N/A                 |

#### Communication & Project Management

| Service | Container | Internal Port | External Port | URL                  | Default Credentials |
| ------- | --------- | ------------- | ------------- | -------------------- | ------------------- |
| Mailpit | mailpit   | 8025/1025     | 9200/9201     | https://mailpit.test | N/A                 |
| Gitea   | gitea     | 3000/22       | 9210/2222     | https://gitea.test   | Setup required      |

#### Proxy & Security

| Service             | Container           | Internal Port | External Port | URL                   | Default Credentials          |
| ------------------- | ------------------- | ------------- | ------------- | --------------------- | ---------------------------- |
| Nginx Proxy Manager | nginx-proxy-manager | 80/443/81     | 80/443/81     | https://nginx.test    | admin@example.com / changeme |
| Keycloak            | keycloak            | 8443/9000     | 9400/9401     | https://keycloak.test | Setup required               |

### Environment Variable Reference

#### Global Settings

```bash
# Admin Credentials
ADMIN_USER=admin
ADMIN_PASSWORD=123456
ADMIN_EMAIL=admin@site.test

# Container Runtime
CONTAINER_RUNTIME=podman  # or docker

# Network Configuration
NETWORK_NAME=microservices-net
COMPOSE_PROJECT_NAME=microservices
```

#### Database Configuration

```bash
# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=123456
POSTGRES_MULTIPLE_DATABASES=postgres,metabase,nocodb,keycloak

# MariaDB
MYSQL_HOST=mariadb
MYSQL_PORT=3306
MYSQL_ROOT_PASSWORD=123456
MYSQL_DATABASE=mariadb
MYSQL_USER=root
MYSQL_PASSWORD=123456

# MongoDB
MONGO_INITDB_ROOT_USERNAME=root
MONGO_INITDB_ROOT_PASSWORD=123456
MONGO_INITDB_DATABASE=admin

# Redis (no authentication required)
```

#### Service-Specific Variables

```bash
# Grafana
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=123456

# Nginx Proxy Manager
DB_MYSQL_HOST=mariadb
DB_MYSQL_PORT=3306
DB_MYSQL_USER=root
DB_MYSQL_PASSWORD=123456
DB_MYSQL_NAME=npm

# Matomo
MATOMO_DATABASE_HOST=mariadb
MATOMO_DATABASE_USERNAME=root
MATOMO_DATABASE_PASSWORD=123456
MATOMO_DATABASE_DBNAME=matomo

# Development Environment Variables
ASPNETCORE_ENVIRONMENT=Development
NODE_ENV=development
FLASK_ENV=development
```

### Default Credentials and Security Notes

#### ⚠️ Security Warning

**All default passwords must be changed for production use!**

#### Default Credentials Table

| Service                 | Username          | Password | Notes                      |
| ----------------------- | ----------------- | -------- | -------------------------- |
| **System Admin**        | admin             | 123456   | Used for multiple services |
| **PostgreSQL**          | postgres          | 123456   | Database superuser         |
| **MariaDB/MySQL**       | root              | 123456   | Database root user         |
| **MongoDB**             | root              | 123456   | Database admin user        |
| **Grafana**             | admin             | 123456   | Dashboard admin            |
| **Nginx Proxy Manager** | admin@example.com | changeme | Proxy management           |
| **pgAdmin**             | admin@site.test   | 123456   | PostgreSQL web admin       |
| **Mongo Express**       | admin             | 123456   | MongoDB web admin          |

#### Production Security Checklist

- [ ] Change all default passwords
- [ ] Generate secure JWT secrets
- [ ] Update GitHub tokens
- [ ] Configure proper hostnames
- [ ] Review admin credentials
- [ ] Enable firewall rules
- [ ] Use Docker secrets for sensitive values
- [ ] Configure SSL certificates
- [ ] Set up backup encryption
- [ ] Enable audit logging

### Network Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                    microservices-net (Bridge Network)           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │  Databases  │    │ Management  │    │   Backend   │        │
│  │             │    │    Tools    │    │     Apps    │        │
│  │ postgres    │◄───┤ adminer     │    │ php         │        │
│  │ mariadb     │◄───┤ pgadmin     │    │ node        │        │
│  │ mysql       │◄───┤ phpmyadmin  │    │ python      │        │
│  │ mongodb     │◄───┤ mongo-exp   │    │ go          │        │
│  │ redis       │    │ metabase    │    │ dotnet      │        │
│  └─────────────┘    │ nocodb      │    └─────────────┘        │
│                     └─────────────┘                           │
│                                                                │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │ Monitoring  │    │     AI      │    │   Proxy     │        │
│  │             │    │             │    │             │        │
│  │ grafana     │    │ n8n         │    │ nginx-pm    │◄─── HTTP/HTTPS
│  │ prometheus  │    │ langflow    │    │ keycloak    │        │
│  │ kibana      │    └─────────────┘    └─────────────┘        │
│  │ elastic     │                                              │
│  │ logstash    │    ┌─────────────┐    ┌─────────────┐        │
│  │ matomo      │    │    Mail     │    │   Project   │        │
│  └─────────────┘    │             │    │             │        │
│                     │ mailpit     │    │ gitea       │        │
│                     └─────────────┘    └─────────────┘        │
└─────────────────────────────────────────────────────────────────┘
```

### File and Directory Reference

#### Project Structure

````
microservices/
├── compose/                    # Docker Compose files
│   ├── analytics/             # Monitoring & analytics services
│   │   ├── elasticsearch.yml
│   │   ├── grafana.yml
│   │   ├── kibana.yml
│   │   ├── logstash.yml
│   │   ├── matomo.yml
│   │   └── prometheus.yml
│   ├── backend/               # Development environments
│   │   ├── dotnet.yml
│   │   ├── go.yml
│   │   ├── node.yml
│   │   ├── php.yml
│   │   └── python.yml
│   ├── database/              # Database services
│   │   ├── mariadb.yml
│   │   ├── mongodb.yml
│   │   ├── mysql.yml
│   │   ├── postgres.yml
│   │   └── redis.yml
│   ├── dbms/                  # Database management tools
│   │   ├── adminer.yml
│   │   ├── metabase.yml
│   │   ├── mongo-express.yml
│   │   ├── nocodb.yml
│   │   ├── pgadmin.yml
│   │   └── phpmyadmin.yml
│   ├── mail/                  # Email services
│   │   └── mailpit.yml
│   ├── project/               # Project management
│   │   └── gitea.yml
│   ├── proxy/                 # Reverse proxy & security
│   │   ├── keycloak.yml
│   │   └── nginx-proxy-manager.yml
│   ├── ai/                    # AI & automation
│   │   ├── langflow.yml
│   │   └── n8n.yml
│    
│
├── config/                    # Service configurations
│   ├── dotnet/
│   │   ├── Dockerfile
│   │   └── smart-entrypoint.sh
│   ├── go/
│   │   ├── Dockerfile
│   │   └── smart-entrypoint.sh
│   ├── keycloak/
│   │   └── Dockerfile
│   ├── logstash/
│   │   ├── config/
│   │   │   └── logstash.yml
│   │   └── pipeline/
│   │       └── pipeline.conf
│   ├── nginx/
│   │   ├── custom/
│   │   └── Dockerfile
│   ├── node/
│   │   ├── Dockerfile
│   │   └── smart-entrypoint.sh
│   ├── php/
│   │   ├── Dockerfile
│   │   ├── php.ini
│   │   └── smart-entrypoint.sh
│   ├── prometheus/
│   │   └── prometheus.yml
│   └── python/
│       ├── Dockerfile
│       └── smart-entrypoint.sh
├── scripts/                   # Management scripts
│   ├── config.sh             # Core configuration & utilities
│   ├── install.sh            # System installation
│   ├── setup-databases.sh    # Database initialization
│   ├── setup-ssl.sh          # SSL certificate management
│   ├── show-services.sh      # Service status & monitoring
│   ├── start-services.sh     # Service startup
│   ├── stop-services.sh      # Service shutdown
│   └── trust-host.sh         # Certificate trust management
├── apps/                      # Your application code
│   ├── my-react-app/         # Frontend applications
│   ├── my-api/               # Backend services
│   ├── my-laravel-app/       # PHP applications
│   └── shared/               # Shared resources
├── logs/                      # Service logs
│   ├── scripts.log           # Script execution logs
│   ├── dotnet/              # .NET application logs
│   ├── go/                  # Go application logs
│   ├── node/                # Node.js application logs
│   ├── php/                 # PHP application logs
│   └── python/              # Python application logs
├── ssl/                       # SSL certificates
│   ├── wildcard.test.crt     # Certificate file
│   ├── wildcard.test.key     # Private key
│   └── backups/              # Certificate backups
├── backups/                   # System backups
│   ├── databases/            # Database backups
│   ├── volumes/              # Volume backups
│   └── configs/              # Configuration backups
├── .env                      # Environment configuration
├── .env-sample              # Environment template
#### Configuration Files

**Core Configuration Files**:
- `.env` - Main environment configuration
- `scripts/config.sh` - Core script configuration and utilities
- `compose/*.yml` - Individual service definitions

**Service-Specific Configurations**:
- `config/php/php.ini` - PHP runtime configuration
- `config/prometheus/prometheus.yml` - Monitoring configuration
- `config/logstash/pipeline/pipeline.conf` - Log processing rules
- `config/nginx/custom/` - Custom Nginx configurations

**Smart Entrypoint Scripts**:
- `config/*/smart-entrypoint.sh` - Automatic project detection and setup

#### Log File Locations

**Script Logs**:
- `logs/scripts.log` - All script execution logs
- `logs/health-monitor.log` - Health monitoring logs (if enabled)

**Application Logs**:
- Backend services mount logs to `logs/{service}/` directories
- View with: `./scripts/show-services.sh -L`
- Container logs: `podman logs {container_name}`

**System Logs**:
- Container runtime logs: `journalctl -u podman` or `journalctl -u docker`
- System logs: `/var/log/syslog` or `journalctl`

### Quick Reference Commands

#### Essential Commands
```bash
# Quick status check
./scripts/show-services.sh -u

# Start all services
./scripts/start-services.sh

# Stop all services
./scripts/stop-services.sh

# Health dashboard
./scripts/show-services.sh -H

# Live monitoring
./scripts/show-services.sh -i 5 -S

# View logs
./scripts/show-services.sh -L

# Database setup
./scripts/setup-databases.sh -A

# SSL setup
./scripts/setup-ssl.sh && ./scripts/trust-host.sh
````

#### Troubleshooting Commands

```bash
# Check service status
podman ps -a

# View specific service logs
podman logs {service_name}

# Follow logs in real-time
podman logs -f {service_name}

# Check resource usage
podman stats

# Network inspection
podman network inspect microservices-net

# Volume inspection
podman volume ls
podman volume inspect {volume_name}

# Container inspection
podman inspect {container_name}

# System cleanup
podman system prune -a
```

#### Database Commands

```bash
# PostgreSQL
podman exec -it postgres psql -U postgres

# MariaDB/MySQL
podman exec -it mariadb mysql -u root -p

# MongoDB
podman exec -it mongodb mongosh --authenticationDatabase admin -u root -p

# Redis
podman exec -it redis redis-cli
```

### Support and Resources

#### Getting Help

**Script Help**:

```bash
# General help for any script
./scripts/{script_name}.sh -h

# Examples:
./scripts/install.sh -h
./scripts/setup-ssl.sh -h
./scripts/show-services.sh -h
```

**Common Support Scenarios**:

1. **Service Won't Start**:

   ```bash
   # Check logs
   podman logs {service_name}

   # Check port conflicts
   netstat -tlnp | grep {port}

   # Recreate service
   podman compose -f compose/{service}.yml down
   podman compose -f compose/{service}.yml up -d
   ```

2. **Database Connection Issues**:

   ```bash
   # Test connectivity
   podman exec {app_container} ping {db_container}

   # Check database status
   ./scripts/show-services.sh -c database -H

   # Reset database
   ./scripts/setup-databases.sh -A
   ```

3. **SSL Certificate Problems**:

   ```bash
   # Regenerate certificates
   ./scripts/setup-ssl.sh -f

   # Reinstall trust
   ./scripts/trust-host.sh -R && ./scripts/trust-host.sh

   # Check certificate validity
   openssl x509 -in ssl/wildcard.test.crt -noout -dates
   ```

4. **Performance Issues**:

   ```bash
   # Monitor resources
   ./scripts/show-services.sh -T

   # Check system resources
   htop
   df -h

   # Optimize containers
   podman system prune
   ```

#### Best Practices Summary

**Development**:

- Use category-specific deployments for development
- Enable debug logging with `-v` flags
- Regularly backup development databases
- Keep applications in separate directories under `apps/`

**Production**:

- Change all default passwords before deployment
- Use Let's Encrypt certificates for valid domains
- Enable monitoring and alerting
- Set up automated backups
- Configure firewalls and security policies
- Test disaster recovery procedures

**Maintenance**:

- Regularly update container images
- Monitor disk space and clean up old logs
- Review and rotate certificates
- Backup before major changes
- Keep documentation updated

#### Performance Tips

**Container Optimization**:

- Set appropriate resource limits
- Use multi-stage builds for smaller images
- Enable build caching
- Regularly prune unused resources

**Database Performance**:

- Configure appropriate buffer sizes
- Enable query caching
- Set up proper indexes
- Monitor slow queries

**Network Performance**:

- Use internal container networking
- Optimize proxy configurations
- Enable compression
- Configure appropriate timeouts

**Security Best Practices**:

- Regular security updates
- Network segmentation
- Access control policies
- Audit logging
- Backup encryption

---

## Conclusion

This microservices architecture provides a comprehensive, production-ready development environment with intelligent automation and extensive management capabilities. The system's smart detection features, modular design, and comprehensive tooling make it suitable for both development and production deployments.

### Key Benefits

1. **Intelligent Automation**: Automatic detection and configuration of different project types
2. **Modular Architecture**: Independent service scaling and management
3. **Cross-Platform Support**: Works with Docker and Podman on multiple operating systems
4. **Production Ready**: SSL certificates, monitoring, backup, and security features
5. **Developer Friendly**: Hot reload, debugging support, and comprehensive logging
6. **Extensive Tooling**: Complete set of management scripts for all operations

### Getting Started Checklist

- [ ] Install container runtime (Docker or Podman)
- [ ] Clone/extract the microservices project
- [ ] Copy `.env-sample` to `.env`
- [ ] Run `./scripts/install.sh -y` for full installation
- [ ] Configure SSL certificates with `./scripts/setup-ssl.sh`
- [ ] Install certificates with `./scripts/trust-host.sh`
- [ ] Access services via https://{service}.test URLs
- [ ] Deploy your applications to the `apps/` directory

### Next Steps

After completing the basic setup:

1. **Customize** the system for your specific needs
2. **Deploy** your applications using the smart detection features
3. **Configure** monitoring and alerting for your environment
4. **Set up** automated backups and disaster recovery
5. **Integrate** with your CI/CD pipelines
6. **Secure** the system for production use

This documentation serves as your complete reference for managing and extending the microservices architecture. Keep it updated as you customize the system for your specific requirements.

---

_Last updated: $(date)_
_Version: 1.0_
