#!/bin/zsh
# Source configuration and dependencies
. "$(dirname "$0")/config.sh"

# Colorize output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print section header
print_header() {
  echo ""
  echo -e "${BLUE}=== $1 ===${NC}"
  echo ""
}

# Function to print service info
print_service() {
  local name=$1
  local local_url=$2
  local proxy_url=$3
  
  echo -e "${GREEN}${name}${NC}"
  echo -e "  ${YELLOW}Local:${NC} ${local_url}"
  echo -e "  ${YELLOW}Proxy:${NC} ${proxy_url}"
  echo ""
}

# Display welcome message
echo -e "${BLUE}===============================================${NC}"
echo -e "${BLUE}      MICROSERVICES ARCHITECTURE DIRECTORY     ${NC}"
echo -e "${BLUE}===============================================${NC}"
echo ""
echo -e "Each service can be accessed both ${YELLOW}locally${NC} via port and through the ${YELLOW}proxy${NC} via domain name."

# Front End/Proxy
print_header "FRONT END"
print_service "Nginx Proxy Server" "http://localhost:81" "https://nginx.test"

# Database Administration
print_header "DATABASE ADMINISTRATION TOOLS"
print_service "Adminer" "http://localhost:8082" "https://adminer.test"
print_service "phpMyAdmin" "http://localhost:8083" "https://phpmyadmin.test"
print_service "Mongo Express" "http://localhost:8084" "https://mongodb.test"
print_service "Metabase" "http://localhost:8085" "https://metabase.test"
print_service "NocoDB" "http://localhost:8086" "https://nocodb.test"

# Analytics Services
print_header "ANALYTICS & MONITORING"
print_service "Prometheus" "http://localhost:9090" "https://prometheus.test"
print_service "Grafana" "http://localhost:9001" "https://grafana.test"
print_service "Matomo" "http://localhost:9010" "https://matomo.test"

# AI & Workflow Services
print_header "AI & WORKFLOW SERVICES"
print_service "n8n" "http://localhost:9100" "https://n8n.test"
print_service "Langflow" "http://localhost:9110" "https://langflow.test"
print_service "Kibana" "http://localhost:9120" "https://kibana.test"
print_service "Elasticsearch" "http://localhost:9130" "https://elasticsearch.test"
print_service "Keycloak" "https://localhost:9400" "https://keycloak.test"

# Development Tools
print_header "DEVELOPMENT TOOLS"
print_service "Mailpit" "http://localhost:9200" "https://mailpit.test"
print_service "Gitea" "http://localhost:9210" "https://gitea.test"

# ERP and Business Applications
print_header "BUSINESS APPLICATIONS"
print_service "Odoo ERP" "http://localhost:9300" "https://odoo.test"

# Database Connections
print_header "DATABASE DIRECT CONNECTIONS"
print_service "MariaDB" "localhost:8501" "mariadb:3306 (internal)"
print_service "PostgreSQL" "localhost:8502" "postgres:5432 (internal)"
print_service "MongoDB" "localhost:8503" "mongodb:27017 (internal)"
print_service "Redis" "localhost:8504" "redis:6379 (internal)"

# Print footer with tips
print_header "QUICK TIPS"
echo "1. Use local ports for development and direct database access"
echo "2. Use the proxy URLs for a production-like environment"
echo "3. All services are on the same Docker network: 'microservices-net'"
echo "4. Configure your local hosts file if needed:"
echo "   echo '127.0.0.1 adminer.test phpmyadmin.test mongodb.test' >> /etc/hosts"
echo ""
echo -e "${BLUE}===============================================${NC}"