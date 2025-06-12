#!/bin/bash

# Podman Project Structure Setup Script
# This script creates the complete directory structure for your Podman-based project

echo "Creating Podman project structure..."

# Create main directories
mkdir -p apps
mkdir -p compose
mkdir -p config/{nginx,dockerfiles,app-configs}
mkdir -p scripts

# Create .env file with common variables
cat > .env << 'EOF'
# Project Configuration
PROJECT_NAME=my-podman-project
COMPOSE_PROJECT_NAME=my-podman-project

# Network Configuration
NETWORK_NAME=podman-network
SUBNET=172.20.0.0/16

# Common Ports (adjust as needed)
HTTP_PORT=80
HTTPS_PORT=443
DB_PORT=5432
REDIS_PORT=6379

# Environment
ENVIRONMENT=development

# Timezone
TZ=America/New_York

# Data Volumes Base Path
DATA_PATH=./data

# Logs Path
LOGS_PATH=./logs
EOF

# Create initial README
cat > README.md << 'EOF'
# Podman Project

This project uses Podman and Docker Compose for container orchestration.

## Project Structure

```
.
├── apps/                 # Application code and services
├── compose/             # Docker Compose files
├── config/              # Configuration files
│   ├── nginx/          # Nginx configurations
│   ├── dockerfiles/    # Custom Dockerfiles
│   └── app-configs/    # Application-specific configs
├── scripts/            # Utility scripts
├── data/              # Persistent data (created at runtime)
├── logs/              # Log files (created at runtime)
└── .env               # Environment variables
```

## Usage

1. Modify `.env` file with your specific configuration
2. Add your compose files to the `compose/` directory
3. Place configuration files in the appropriate `config/` subdirectories
4. Use the scripts in `scripts/` directory for common operations

## Getting Started

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Start services
./scripts/start.sh

# Stop services
./scripts/stop.sh

# View logs
./scripts/logs.sh
```
EOF

# Create basic .gitignore
cat > .gitignore << 'EOF'
# Data and logs
data/
logs/
*.log

# Environment files (keep template, ignore actual configs)
.env.local
.env.production
.env.staging

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# Temporary files
*.tmp
*.temp
temp/
tmp/
EOF

# Create placeholder files to maintain directory structure
touch apps/.gitkeep
touch compose/.gitkeep
touch config/nginx/.gitkeep
touch config/dockerfiles/.gitkeep
touch config/app-configs/.gitkeep
touch scripts/.gitkeep

# Create data and logs directories
mkdir -p data logs

echo "✅ Project structure created successfully!"
echo ""
echo "Directory structure:"
tree -a -I '.git' || ls -la

echo ""
echo "Next steps:"
echo "1. Review and modify the .env file with your specific settings"
echo "2. Tell me what services you need and I'll create the compose files"
echo "3. We'll then create utility scripts for managing your containers"
EOF