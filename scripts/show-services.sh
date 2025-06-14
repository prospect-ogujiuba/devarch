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

# Proxy & SSL
print_header "🌐 REVERSE PROXY & SSL"
print_service "Nginx Proxy Manager" "http://localhost:81" "https://nginx.test"

# Database Services
print_header "🗄️ DATABASE SERVICES"
print_service "MariaDB" "localhost:8501" "mariadb:3306 (internal)"
print_service "MySQL" "localhost:8505" "mysql:3306 (internal)"
print_service "PostgreSQL" "localhost:8502" "postgres:5432 (internal)"
print_service "MongoDB" "localhost:8503" "mongodb:27017 (internal)"
print_service "Redis" "localhost:8504" "redis:6379 (internal)"

# Database Management Tools
print_header "🔧 DATABASE MANAGEMENT TOOLS"
print_service "Adminer" "http://localhost:8082" "https://adminer.test"
print_service "phpMyAdmin" "http://localhost:8083" "https://phpmyadmin.test"
print_service "Mongo Express" "http://localhost:8084" "https://mongodb.test"
print_service "Metabase" "http://localhost:8085" "https://metabase.test"
print_service "NocoDB" "http://localhost:8086" "https://nocodb.test"
print_service "pgAdmin" "http://localhost:8087" "https://pgadmin.test"

# Backend Services
print_header "⚙️ BACKEND SERVICES"
print_service ".NET Core" "http://localhost:8010" "https://dotnet.test"
print_service "Go" "http://localhost:8020" "https://go.test"
print_service "Node.js" "http://localhost:8030" "https://node.test"
print_service "PHP" "http://localhost:8000" "https://php.test"
print_service "Python" "http://localhost:8040" "https://python.test"

# Analytics & Monitoring
print_header "📊 ANALYTICS & MONITORING"
print_service "Elasticsearch" "http://localhost:9130" "https://elasticsearch.test"
print_service "Kibana" "http://localhost:9120" "https://kibana.test"
print_service "Grafana" "http://localhost:9001" "https://grafana.test"
print_service "Prometheus" "http://localhost:9090" "https://prometheus.test"
print_service "Matomo Analytics" "http://localhost:9010" "https://matomo.test"

# AI & Workflow Services
print_header "🤖 AI & WORKFLOW SERVICES"
print_service "Langflow" "http://localhost:9110" "https://langflow.test"
print_service "n8n" "http://localhost:9100" "https://n8n.test"

# Mail Services
print_header "📧 MAIL SERVICES"
print_service "Mailpit" "http://localhost:9200" "https://mailpit.test"

# Project Management
print_header "📋 PROJECT MANAGEMENT"
print_service "Gitea" "http://localhost:9210" "https://gitea.test"

# Business Applications
print_header "🏢 BUSINESS APPLICATIONS"
print_service "Odoo ERP" "http://localhost:9300" "https://odoo.test"

# Print footer with credentials and tips
print_header "🔐 DEFAULT CREDENTIALS"
echo -e "${YELLOW}Username:${NC} admin"
echo -e "${YELLOW}Password:${NC} 123456"
echo -e "${YELLOW}Email:${NC} admin@site.test"

print_header "📋 QUICK MANAGEMENT COMMANDS"
echo "Start all services:    $SCRIPT_DIR/start-services.sh"
echo "Stop all services:     $SCRIPT_DIR/stop-services.sh"
echo "Setup databases:       $SCRIPT_DIR/setup-databases.sh -a"
echo "Setup SSL:            $SCRIPT_DIR/setup-ssl.sh"
echo "Install SSL trust:     $SCRIPT_DIR/trust-host.sh"

print_header "💡 QUICK TIPS"
echo "1. Use local ports for development and direct database access"
echo "2. Use the proxy URLs for a production-like environment"
echo "3. All services are on the same Docker network: 'microservices-net'"
echo "4. Configure your local hosts file if needed:"
echo "   echo '127.0.0.1 nginx.test grafana.test metabase.test' >> /etc/hosts"
echo ""
echo -e "${BLUE}===============================================${NC}"