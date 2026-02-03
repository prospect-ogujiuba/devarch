#!/bin/bash
# Initialize required databases for microservices
# Run this after MariaDB container is started for the first time

set -e

# Load environment variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
else
    echo "Error: .env file not found at $PROJECT_ROOT/.env"
    exit 1
fi

echo "Initializing databases in MariaDB..."

# Check if MariaDB is running
if ! podman ps | grep -q mariadb; then
    echo "Error: MariaDB container is not running. Start it first:"
    echo "  podman-compose -f compose/database/mariadb.yml up -d"
    exit 1
fi

# Wait for MariaDB to be ready
echo "Waiting for MariaDB to be ready..."
for i in {1..30}; do
    if podman exec mariadb mariadb -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SELECT 1" &>/dev/null; then
        echo "MariaDB is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: MariaDB did not become ready in time"
        exit 1
    fi
    sleep 2
done

# Create databases
echo "Creating databases..."
podman exec -i mariadb mariadb -uroot -p"${MYSQL_ROOT_PASSWORD}" <<EOF
-- Databases for Containers
CREATE DATABASE IF NOT EXISTS npm;
CREATE DATABASE IF NOT EXISTS metabase;
CREATE DATABASE IF NOT EXISTS matomo;

-- WordPress application databases
#CREATE DATABASE IF NOT EXISTS playground;

-- Show all databases
SHOW DATABASES;
EOF

echo ""
echo "Database initialization complete!"
echo ""
echo "Databases created:"
echo "  - npm (Nginx Proxy Manager)"
echo "  - b2bcnc (WordPress)"
echo "  - playground (WordPress)"
echo "  - flowstate (WordPress)"
echo "  - mediagrowthpartner (WordPress)"
