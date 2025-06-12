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
