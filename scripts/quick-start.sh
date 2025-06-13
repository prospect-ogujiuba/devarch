#!/bin/zsh
# Quick Start Script for Microservices Access
# Run this after your services are up to get immediate access

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "🚀 Microservices Quick Start Setup"
echo "=================================="

# Check if services are running
echo ""
echo "📊 Checking service status..."
if command -v podman >/dev/null 2>&1; then
    RUNTIME="podman"
elif command -v docker >/dev/null 2>&1; then
    RUNTIME="docker"
else
    echo "❌ Neither podman nor docker found!"
    exit 1
fi

# Count running containers
RUNNING_COUNT=$($RUNTIME ps --filter "network=microservices-net" --format "{{.Names}}" | wc -l)
echo "✅ Found $RUNNING_COUNT microservices running"

if [ $RUNNING_COUNT -lt 3 ]; then
    echo "⚠️  Few services detected. You may want to start more services first:"
    echo "   ./scripts/install.sh  # For full setup"
    echo "   ./scripts/start-services.sh -c database -c dbms  # For basic setup"
    echo ""
fi

# Quick domain setup
echo "🌐 Setting up local domain access..."
echo "This will add entries to your /etc/hosts file (requires sudo)"
read -p "Add .test domains to hosts file? (y/N): " add_hosts

if [[ "$add_hosts" =~ ^[Yy]$ ]]; then
    # Backup hosts file
    sudo cp /etc/hosts /etc/hosts.backup.$(date +%Y%m%d_%H%M%S)
    
    # Add microservices entries if they don't exist
    DOMAINS=(
        "nginx.test"
        "adminer.test" 
        "phpmyadmin.test"
        "mongodb.test"
        "metabase.test"
        "nocodb.test"
        "grafana.test"
        "prometheus.test"
        "matomo.test"
        "n8n.test"
        "langflow.test"
        "kibana.test"
        "keycloak.test"
        "mailpit.test"
        "gitea.test"
        "odoo.test"
    )
    
    # Check if microservices section exists
    if ! grep -q "# Microservices Development" /etc/hosts; then
        echo "" | sudo tee -a /etc/hosts > /dev/null
        echo "# Microservices Development" | sudo tee -a /etc/hosts > /dev/null
    fi
    
    # Add missing domains
    ADDED=0
    for domain in "${DOMAINS[@]}"; do
        if ! grep -q "127\.0\.0\.1[[:space:]]*$domain" /etc/hosts; then
            echo "127.0.0.1 $domain" | sudo tee -a /etc/hosts > /dev/null
            ADDED=$((ADDED + 1))
        fi
    done
    
    echo "✅ Added $ADDED domain entries to /etc/hosts"
fi

# Show immediate access options
echo ""
echo "🎯 IMMEDIATE ACCESS OPTIONS"
echo "=========================="
echo ""
echo "📱 Management Interfaces:"
echo "  🌐 Nginx Proxy Manager:  http://localhost:81"
echo "     └─ Login: admin@example.com / changeme"
echo "  🗄️  Database Admin:       http://localhost:8082"  
echo "  📊 Grafana Monitoring:   http://localhost:9001"
echo "     └─ Login: admin / 123456"
echo ""

# Check what's actually running and show relevant URLs
echo "🔍 Available Services:"
SERVICES=(
    "nginx-proxy-manager:81:🌐 Nginx Proxy"
    "adminer:8082:🗄️ Database Admin"
    "pgadmin:8087:🐘 PostgreSQL Admin"
    "phpmyadmin:8083:🐬 MySQL Admin"
    "mongo-express:8084:🍃 MongoDB Admin"
    "metabase:8085:📊 Business Intelligence"
    "nocodb:8086:📋 No-Code Database"
    "grafana:9001:📈 Monitoring"
    "prometheus:9090:📊 Metrics"
    "matomo:9010:📈 Analytics"
    "n8n:9100:🔄 Workflows"
    "langflow:9110:🤖 AI Workflows"
    "kibana:9120:📝 Log Viewer"
    "mailpit:9200:📧 Email Testing"
    "gitea:9210:📁 Git Repos"
    "odoo:9300:💼 ERP System"
)

for service_info in "${SERVICES[@]}"; do
    IFS=':' read -r service port description <<< "$service_info"
    if $RUNTIME ps --format "{{.Names}}" | grep -q "^${service}$"; then
        echo "  ✅ $description: http://localhost:$port"
    fi
done

# Show backend development ports
echo ""
echo "🛠️  Development Environments:"
BACKEND_SERVICES=(
    "php:8000:🐘 PHP Development"
    "node:8030:🟢 Node.js Development" 
    "python:8040:🐍 Python Development"
    "go:8020:🐹 Go Development"
    "dotnet:8010:⚡ .NET Development"
)

for service_info in "${BACKEND_SERVICES[@]}"; do
    IFS=':' read -r service port description <<< "$service_info"
    if $RUNTIME ps --format "{{.Names}}" | grep -q "^${service}$"; then
        echo "  ✅ $description: http://localhost:$port"
        echo "     └─ Mount point: ./apps/ → /var/www/html/ (PHP) or /app/ (others)"
    fi
done

# Database connections
echo ""
echo "🗄️  Database Connections:"
DB_SERVICES=(
    "postgres:8502:🐘 PostgreSQL"
    "mariadb:8501:🐬 MariaDB"
    "mysql:8505:🐬 MySQL"
    "mongodb:8503:🍃 MongoDB"
    "redis:8504:🔴 Redis"
)

for service_info in "${DB_SERVICES[@]}"; do
    IFS=':' read -r service port description <<< "$service_info"
    if $RUNTIME ps --format "{{.Names}}" | grep -q "^${service}$"; then
        echo "  ✅ $description: localhost:$port"
        case $service in
            "postgres") echo "     └─ User: postgres, Pass: 123456" ;;
            "mariadb"|"mysql") echo "     └─ User: root, Pass: 123456" ;;
            "mongodb") echo "     └─ User: root, Pass: 123456" ;;
            "redis") echo "     └─ No authentication required" ;;
        esac
    fi
done

# SSL setup option
echo ""
echo "🔐 HTTPS/SSL Setup (Optional):"
echo "  To enable https:// URLs with trusted certificates:"
echo "  ./scripts/setup-ssl.sh      # Generate certificates"
echo "  ./scripts/trust-host.sh     # Trust certificates system-wide"

# Next steps
echo ""
echo "🎯 NEXT STEPS"
echo "============"
echo ""
echo "1. 📱 Visit Nginx Proxy Manager: http://localhost:81"
echo "   • Set up proxy rules for your custom domains"
echo "   • Configure SSL certificates"
echo ""
echo "2. 🛠️  Deploy Your Applications:"
echo "   • Create folder: mkdir apps/my-awesome-app"
echo "   • Add your code (any framework will be auto-detected)"
echo "   • Start service: podman compose -f compose/node.yml up -d"
echo "   • Access: http://localhost:8030"
echo ""
echo "3. 🗄️  Set Up Databases:"
echo "   • Visit Adminer: http://localhost:8082"
echo "   • Create databases for your applications"
echo "   • Use connection strings from the guide above"
echo ""
echo "4. 📊 Monitor Everything:"
echo "   • ./scripts/show-services.sh    # Service status"
echo "   • ./scripts/show-services.sh -u # Quick URL reference"
echo "   • Visit Grafana: http://localhost:9001"
echo ""

# Browser open option
if command -v xdg-open >/dev/null 2>&1 || command -v open >/dev/null 2>&1; then
    echo "🌐 Quick Access:"
    read -p "Open Nginx Proxy Manager in browser? (y/N): " open_browser
    
    if [[ "$open_browser" =~ ^[Yy]$ ]]; then
        if command -v xdg-open >/dev/null 2>&1; then
            xdg-open "http://localhost:81" >/dev/null 2>&1 &
        elif command -v open >/dev/null 2>&1; then
            open "http://localhost:81" >/dev/null 2>&1 &
        fi
        echo "✅ Opening http://localhost:81 in browser..."
    fi
fi

echo ""
echo "✨ Quick Start Complete!"
echo "📖 For detailed information, see the access guide above."
echo "🆘 Having issues? Run: ./scripts/show-services.sh -v"