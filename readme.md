# Microservices Architecture Platform

A comprehensive, Docker/Podman-based microservices platform that provides a complete development environment with databases, backend services, analytics, AI tools, and development utilities. Built with cross-platform compatibility and smart automation.

## 🚀 Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd microservices-platform

# Full installation (recommended)
./scripts/install.sh

# Or step-by-step installation
./scripts/start-services.sh
./scripts/setup-databases.sh -a
./scripts/setup-ssl.sh
./scripts/trust-host.sh
```

## 🏗️ Architecture Overview

### Service Categories
- **🗄️ Databases**: MariaDB, MySQL, PostgreSQL, MongoDB, Redis
- **🔧 Database Tools**: Adminer, phpMyAdmin, Mongo Express, Metabase, NocoDB, pgAdmin
- **⚙️ Backend Services**: .NET Core, Go, Node.js, PHP, Python (with smart detection)
- **📊 Analytics & Monitoring**: Elasticsearch, Kibana, Grafana, Prometheus, Matomo
- **🤖 AI Services**: Langflow, n8n automation platforms
- **📧 Development Tools**: Mailpit, Gitea
- **🌐 Reverse Proxy**: Nginx Proxy Manager with SSL

### Smart Features
- **Intelligent Project Detection**: Automatically detects and configures Laravel, Django, React, Next.js, etc.
- **Cross-Platform SSL**: Works on Linux, macOS, Windows, and WSL
- **Zero-Configuration**: Smart entrypoints handle dependency installation and setup
- **Health Monitoring**: Built-in health checks and dependency management

## 📁 Project Structure

```
microservices-platform/
├── scripts/           # Management scripts (install, start, stop, etc.)
├── compose/           # Docker Compose files organized by category
│   ├── database/      # Database services
│   ├── dbms/         # Database management tools
│   ├── backend/      # Application runtimes
│   ├── analytics/    # Monitoring and analytics
│   └── ai-services/  # AI and automation tools
├── config/           # Service configurations and Dockerfiles
├── apps/             # Your application code goes here
└── logs/             # Application logs
```

## 🎯 Key Services & Access Points

| Service | Local URL | Proxy URL | Purpose |
|---------|-----------|-----------|---------|
| Nginx Proxy Manager | http://localhost:81 | https://nginx.test | SSL proxy management |
| Adminer | http://localhost:8082 | https://adminer.test | Universal database tool |
| Grafana | http://localhost:9001 | https://grafana.test | Metrics dashboards |
| Metabase | http://localhost:8085 | https://metabase.test | Business intelligence |
| n8n | http://localhost:9100 | https://n8n.test | Workflow automation |
| Langflow | http://localhost:9110 | https://langflow.test | AI workflow builder |
| Mailpit | http://localhost:9200 | https://mailpit.test | Email testing |
| Gitea | http://localhost:9210 | https://gitea.test | Git repository hosting |

## 🔧 Management Commands

```bash
# Service Management
./scripts/start-services.sh [options]     # Start all or specific services
./scripts/stop-services.sh [options]      # Stop services with cleanup options
./scripts/show-services.sh                # Display all services and URLs

# Database Setup
./scripts/setup-databases.sh -a           # Setup all databases
./scripts/setup-databases.sh -m -p        # Setup MariaDB and PostgreSQL only

# SSL Configuration
./scripts/setup-ssl.sh                    # Generate SSL certificates
./scripts/trust-host.sh                   # Install certificates in system trust

# Full Installation
./scripts/install.sh [options]            # Complete setup with options
```

### Command Options
- `-s, --sudo`: Use sudo for container commands
- `-c, --categories LIST`: Target specific service categories
- `-f, --force`: Force recreation/regeneration
- `-q, --quick`: Quick mode (skip SSL setup)
- `-h, --help`: Show detailed help for any script

## 🐳 Container Runtime Support

The platform supports both **Docker** and **Podman** with automatic detection:

```bash
# Configuration is automatic, but you can override:
export CONTAINER_RUNTIME="podman"  # or "docker"
export USE_PODMAN=true             # or false
```

## 💾 Application Development

### Adding Your Applications

1. **Place your code** in the appropriate `apps/` subdirectory:
   ```
   apps/
   ├── dotnet/     # .NET applications
   ├── go/         # Go applications  
   ├── node/       # Node.js applications
   ├── php/        # PHP applications
   └── python/     # Python applications
   ```

2. **Smart Detection**: The platform automatically detects:
   - **Node.js**: React, Next.js, NestJS, Express
   - **Python**: Django, FastAPI, Flask, Streamlit
   - **PHP**: Laravel, WordPress, Symfony, CodeIgniter
   - **Go**: Gin, Echo, Fiber, standard net/http
   - **.NET**: ASP.NET Core, Blazor, Web API, Console apps

3. **Auto-Configuration**: Dependencies are installed and applications are configured automatically.

### Database Connections

Default connection strings:
```bash
# PostgreSQL
postgresql://postgres:123456@postgres:5432/postgres

# MariaDB/MySQL  
mysql://root:123456@mariadb:3306/mariadb

# MongoDB
mongodb://root:123456@mongodb:27017/admin

# Redis
redis://redis:6379
```

## 🔐 Security & Credentials

**Default Credentials** (change for production):
- Username: `admin`
- Password: `123456`
- Email: `admin@example.com`

**SSL Certificates**: Self-signed wildcard certificates for `*.test` domains, automatically trusted by browsers.

## 🌐 Production Deployment

For production use:

1. **Update credentials** in `.env` file
2. **Configure real SSL** certificates via Let's Encrypt
3. **Set appropriate environment variables**:
   ```bash
   ASPNETCORE_ENVIRONMENT=Production
   NODE_ENV=production
   GO_ENV=production
   ```
4. **Review security settings** in compose files
5. **Use Docker secrets** for sensitive data

## 📚 Documentation & Support

- **Service Status**: `./scripts/show-services.sh`
- **Logs**: Check `logs/` directory or use `podman logs <container>`
- **Configuration**: All configs in `config/` directory
- **Troubleshooting**: Use `-e` flag with scripts for detailed errors

## 🎉 Features Highlights

- ✅ **Zero-config setup** for 20+ development tools
- ✅ **Smart project detection** and dependency management  
- ✅ **Cross-platform SSL** with automatic browser trust
- ✅ **Unified networking** with service discovery
- ✅ **Health monitoring** and graceful shutdown
- ✅ **Selective service management** by category
- ✅ **Production-ready** configurations available

---

**Happy coding!** 🚀 This platform gives you a complete microservices environment in minutes, not hours.