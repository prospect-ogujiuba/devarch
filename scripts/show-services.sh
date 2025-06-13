#!/bin/zsh
# Enhanced Show Services Script for Microservices Architecture
# Displays comprehensive service information with status, health, and connectivity

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# =============================================================================
# SCRIPT-SPECIFIC CONFIGURATION
# =============================================================================

# Display options
SHOW_STATUS=true
SHOW_HEALTH=false
SHOW_LOGS=false
SHOW_STATS=false
SHOW_NETWORKS=false
SHOW_VOLUMES=false
SHOW_ONLY_RUNNING=false
SHOW_ONLY_CATEGORY=""
OUTPUT_FORMAT="table"           # table, json, csv, markdown
FOLLOW_LOGS=false
LOG_LINES=50
REFRESH_INTERVAL=0
SHOW_URLS_ONLY=false

# Service information structure
declare -A SERVICE_PORTS=(
    # Database services
    ["postgres"]="8502:5432"
    ["mariadb"]="8501:3306"
    ["mysql"]="8505:3306"
    ["mongodb"]="8503:27017"
    ["redis"]="8504:6379"
    
    # Database management tools
    ["adminer"]="8082:8080"
    ["pgadmin"]="8087:80"
    ["phpmyadmin"]="8083:80"
    ["mongo-express"]="8084:8081"
    ["metabase"]="8085:3000"
    ["nocodb"]="8086:8080"
    
    # Backend services
    ["php"]="8000:8000,5173:5173"
    ["node"]="8030:3000"
    ["python"]="8040:8000"
    ["go"]="8020:8080"
    ["dotnet"]="8010:80,8011:443"
    
    # Analytics & monitoring
    ["elasticsearch"]="9130:9200,9131:9300"
    ["kibana"]="9120:5601"
    ["logstash"]="5000:5000,5044:5044,9600:9600"
    ["grafana"]="9001:3000"
    ["prometheus"]="9090:9090"
    ["matomo"]="9010:80"
    
    # AI & automation
    ["langflow"]="9110:7860"
    ["n8n"]="9100:5678"
    
    # Mail & communication
    ["mailpit"]="9201:1025,9200:8025"
    
    # Project management
    ["gitea"]="9210:3000,2222:22"
    
    # ERP
    ["odoo"]="9300:8069,9301:8072"
    
    # Proxy & security
    ["nginx-proxy-manager"]="80:80,443:443,81:81"
    ["keycloak"]="9400:8443,9401:9000"
)

declare -A SERVICE_URLS=(
    # Database management
    ["adminer"]="https://adminer.test"
    ["pgadmin"]="https://pgadmin.test"
    ["phpmyadmin"]="https://phpmyadmin.test"
    ["mongo-express"]="https://mongodb.test"
    ["metabase"]="https://metabase.test"
    ["nocodb"]="https://nocodb.test"
    
    # Analytics & monitoring
    ["grafana"]="https://grafana.test"
    ["prometheus"]="https://prometheus.test"
    ["matomo"]="https://matomo.test"
    ["kibana"]="https://kibana.test"
    ["elasticsearch"]="https://elasticsearch.test"
    
    # AI & automation
    ["n8n"]="https://n8n.test"
    ["langflow"]="https://langflow.test"
    
    # Mail & communication
    ["mailpit"]="https://mailpit.test"
    
    # Project management
    ["gitea"]="https://gitea.test"
    
    # ERP
    ["odoo"]="https://odoo.test"
    
    # Proxy & security
    ["nginx-proxy-manager"]="https://nginx.test"
    ["keycloak"]="https://keycloak.test"
)

declare -A SERVICE_DESCRIPTIONS=(
    # Database services
    ["postgres"]="PostgreSQL Database Server"
    ["mariadb"]="MariaDB Database Server"
    ["mysql"]="MySQL Database Server"
    ["mongodb"]="MongoDB NoSQL Database"
    ["redis"]="Redis Key-Value Store"
    
    # Database management
    ["adminer"]="Universal Database Management Tool"
    ["pgadmin"]="PostgreSQL Administration Tool"
    ["phpmyadmin"]="MySQL/MariaDB Administration Tool"
    ["mongo-express"]="MongoDB Administration Interface"
    ["metabase"]="Business Intelligence & Analytics"
    ["nocodb"]="No-Code Database Platform"
    
    # Backend services
    ["php"]="PHP Development Environment"
    ["node"]="Node.js Development Environment"
    ["python"]="Python Development Environment"
    ["go"]="Go Development Environment"
    ["dotnet"]="ASP.NET Core Development Environment"
    
    # Analytics & monitoring
    ["elasticsearch"]="Search & Analytics Engine"
    ["kibana"]="Data Visualization & Management"
    ["logstash"]="Log Processing Pipeline"
    ["grafana"]="Monitoring & Observability"
    ["prometheus"]="Metrics Collection & Monitoring"
    ["matomo"]="Web Analytics Platform"
    
    # AI & automation
    ["langflow"]="Visual AI Workflow Builder"
    ["n8n"]="Workflow Automation Platform"
    
    # Mail & communication
    ["mailpit"]="Email Testing & SMTP Server"
    
    # Project management
    ["gitea"]="Git Repository Management"
    
    # ERP
    ["odoo"]="Enterprise Resource Planning"
    
    # Proxy & security
    ["nginx-proxy-manager"]="Reverse Proxy & SSL Manager"
    ["keycloak"]="Identity & Access Management"
)

# Colors for output
declare -A COLORS=(
    ["RED"]='\033[0;31m'
    ["GREEN"]='\033[0;32m'
    ["YELLOW"]='\033[0;33m'
    ["BLUE"]='\033[0;34m'
    ["MAGENTA"]='\033[0;35m'
    ["CYAN"]='\033[0;36m'
    ["WHITE"]='\033[0;37m'
    ["BOLD"]='\033[1m'
    ["RESET"]='\033[0m'
)

# =============================================================================
# HELP FUNCTION
# =============================================================================

show_help() {
    cat << EOF
Enhanced Microservices Status Display

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s              Use sudo for container commands
    -e              Show error messages
    -v              Verbose output (debug mode)
    -d              Dry run (show commands without executing)
    -r RUNTIME      Container runtime (docker/podman)
    -S              Show service status and health
    -H              Show detailed health information
    -L              Show recent logs for services
    -T              Show resource usage statistics
    -N              Show network information
    -V              Show volume information
    -R              Show only running services
    -c CATEGORY     Show only specific category
    -f FORMAT       Output format: table, json, csv, markdown (default: table)
    -F              Follow logs in real-time (use with -L)
    -l LINES        Number of log lines to show (default: 50)
    -i SECONDS      Refresh interval for continuous monitoring (0 = no refresh)
    -u              Show URLs only (quick reference)
    -h              Show this help message

CATEGORIES:
    database        Database servers (postgres, mysql, mariadb, mongodb, redis)
    dbms           Database management tools (adminer, pgadmin, phpmyadmin, etc.)
    backend        Development environments (php, node, python, go, dotnet)
    analytics      Monitoring & analytics (grafana, prometheus, kibana, etc.)
    ai             AI & automation (langflow, n8n)
    mail           Email services (mailpit)
    project        Project management (gitea)
    erp            Business applications (odoo)
    proxy          Proxy & security (nginx-proxy-manager, keycloak)

OUTPUT FORMATS:
    table          Formatted table output (default)
    json           JSON format for scripting
    csv            CSV format for spreadsheets
    markdown       Markdown format for documentation

EXAMPLES:
    $0                              # Show all services in table format
    $0 -S -H                        # Show status and health information
    $0 -c database -R               # Show only running database services
    $0 -L -l 100                    # Show last 100 log lines for all services
    $0 -f json                      # Output in JSON format
    $0 -u                           # Quick URL reference
    $0 -i 5 -S                      # Continuous monitoring (refresh every 5 seconds)
    $0 -c analytics -H -T           # Detailed view of analytics services

EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_show_args() {
    local OPTIND=1
    
    while getopts "sevdr:SHLTNVRc:f:Fl:i:uh" opt; do
        case $opt in
            s|e|v|d|r) ;; # Handled by parse_common_args
            S) SHOW_STATUS=true ;;
            H) SHOW_HEALTH=true ;;
            L) SHOW_LOGS=true ;;
            T) SHOW_STATS=true ;;
            N) SHOW_NETWORKS=true ;;
            V) SHOW_VOLUMES=true ;;
            R) SHOW_ONLY_RUNNING=true ;;
            c) SHOW_ONLY_CATEGORY="$OPTARG" ;;
            f) OUTPUT_FORMAT="$OPTARG" ;;
            F) FOLLOW_LOGS=true ;;
            l) LOG_LINES="$OPTARG" ;;
            i) REFRESH_INTERVAL="$OPTARG" ;;
            u) SHOW_URLS_ONLY=true ;;
            h) show_help; exit 0 ;;
            ?) show_help; exit 1 ;;
        esac
    done
    
    # Validate output format
    case "$OUTPUT_FORMAT" in
        "table"|"json"|"csv"|"markdown") ;;
        *)
            log "ERROR" "Invalid output format: $OUTPUT_FORMAT"
            log "INFO" "Valid formats: table, json, csv, markdown"
            exit 1
            ;;
    esac
    
    # Validate category
    if [ -n "$SHOW_ONLY_CATEGORY" ]; then
        if [[ ! " ${(k)SERVICE_CATEGORIES[@]} " =~ " $SHOW_ONLY_CATEGORY " ]]; then
            log "ERROR" "Invalid category: $SHOW_ONLY_CATEGORY"
            log "INFO" "Valid categories: ${(k)SERVICE_CATEGORIES[*]}"
            exit 1
        fi
    fi
}

# =============================================================================
# SERVICE INFORMATION GATHERING
# =============================================================================

get_service_info() {
    local service_name="$1"
    local info_type="$2"  # status, health, logs, stats
    
    case "$info_type" in
        "status")
            get_service_status "$service_name"
            ;;
        "health")
            get_service_health "$service_name"
            ;;
        "logs")
            get_service_logs "$service_name"
            ;;
        "stats")
            get_service_stats "$service_name"
            ;;
        "ports")
            get_service_ports "$service_name"
            ;;
        "uptime")
            get_service_uptime "$service_name"
            ;;
    esac
}

get_service_status() {
    local service_name="$1"
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service_name" 2>/dev/null; then
        echo "not found"
        return 1
    fi
    
    local status=$(${SUDO_PREFIX}${CONTAINER_RUNTIME} container inspect "$service_name" --format '{{.State.Status}}' 2>/dev/null)
    echo "${status:-unknown}"
}

get_service_health() {
    local service_name="$1"
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service_name" 2>/dev/null; then
        echo "container not found"
        return 1
    fi
    
    # Check if container has health check
    local health_status=$(${SUDO_PREFIX}${CONTAINER_RUNTIME} container inspect "$service_name" --format '{{.State.Health.Status}}' 2>/dev/null)
    
    if [ -n "$health_status" ] && [ "$health_status" != "<no value>" ]; then
        echo "$health_status"
    else
        # If no health check, determine health based on status and connectivity
        local status=$(get_service_status "$service_name")
        case "$status" in
            "running")
                # Try to connect to service if we know the port
                if check_service_connectivity "$service_name"; then
                    echo "healthy"
                else
                    echo "unhealthy"
                fi
                ;;
            "exited")
                echo "unhealthy"
                ;;
            *)
                echo "unknown"
                ;;
        esac
    fi
}

check_service_connectivity() {
    local service_name="$1"
    local ports="${SERVICE_PORTS[$service_name]}"
    
    if [ -z "$ports" ]; then
        return 0  # Assume healthy if no ports defined
    fi
    
    # Extract first port for connectivity test
    local first_port=$(echo "$ports" | cut -d',' -f1 | cut -d':' -f1)
    
    # Test connectivity
    if command -v nc >/dev/null 2>&1; then
        nc -z localhost "$first_port" 2>/dev/null
    elif command -v curl >/dev/null 2>&1; then
        curl -s --connect-timeout 2 "http://localhost:$first_port" >/dev/null 2>&1
    else
        return 0  # Assume healthy if no test tools available
    fi
}

get_service_logs() {
    local service_name="$1"
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service_name" 2>/dev/null; then
        echo "Container not found"
        return 1
    fi
    
    local log_cmd="${SUDO_PREFIX}${CONTAINER_RUNTIME} logs"
    
    if [ "$FOLLOW_LOGS" = true ]; then
        log_cmd="$log_cmd -f"
    fi
    
    log_cmd="$log_cmd --tail $LOG_LINES $service_name"
    
    eval "$log_cmd" 2>/dev/null || echo "No logs available"
}

get_service_stats() {
    local service_name="$1"
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service_name" 2>/dev/null; then
        echo "Container not found"
        return 1
    fi
    
    # Get basic stats
    local stats=$(${SUDO_PREFIX}${CONTAINER_RUNTIME} stats "$service_name" --no-stream --format "table {{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}" 2>/dev/null | tail -1)
    echo "$stats"
}

get_service_ports() {
    local service_name="$1"
    echo "${SERVICE_PORTS[$service_name]:-N/A}"
}

get_service_uptime() {
    local service_name="$1"
    
    if ! ${SUDO_PREFIX}${CONTAINER_RUNTIME} container exists "$service_name" 2>/dev/null; then
        echo "N/A"
        return 1
    fi
    
    local started_at=$(${SUDO_PREFIX}${CONTAINER_RUNTIME} container inspect "$service_name" --format '{{.State.StartedAt}}' 2>/dev/null)
    
    if [ -n "$started_at" ]; then
        # Calculate uptime
        local start_timestamp=$(date -d "$started_at" +%s 2>/dev/null || echo "0")
        local current_timestamp=$(date +%s)
        local uptime_seconds=$((current_timestamp - start_timestamp))
        
        # Format uptime
        local days=$((uptime_seconds / 86400))
        local hours=$(((uptime_seconds % 86400) / 3600))
        local minutes=$(((uptime_seconds % 3600) / 60))
        
        if [ $days -gt 0 ]; then
            echo "${days}d ${hours}h ${minutes}m"
        elif [ $hours -gt 0 ]; then
            echo "${hours}h ${minutes}m"
        else
            echo "${minutes}m"
        fi
    else
        echo "N/A"
    fi
}

# =============================================================================
# SERVICE LISTING
# =============================================================================

get_services_to_show() {
    local services_to_show=()
    
    if [ -n "$SHOW_ONLY_CATEGORY" ]; then
        # Show only services from specified category
        local files=(${=SERVICE_CATEGORIES[$SHOW_ONLY_CATEGORY]})
        for file in "${files[@]}"; do
            local service_name=$(basename "$file" .yml)
            services_to_show+=("$service_name")
        done
    else
        # Show all services
        for category in "${(k)SERVICE_CATEGORIES[@]}"; do
            local files=(${=SERVICE_CATEGORIES[$category]})
            for file in "${files[@]}"; do
                local service_name=$(basename "$file" .yml)
                services_to_show+=("$service_name")
            done
        done
    fi
    
    # Filter to running services only if requested
    if [ "$SHOW_ONLY_RUNNING" = true ]; then
        local running_services=()
        for service in "${services_to_show[@]}"; do
            local status=$(get_service_status "$service")
            if [ "$status" = "running" ]; then
                running_services+=("$service")
            fi
        done
        services_to_show=("${running_services[@]}")
    fi
    
    echo "${services_to_show[@]}"
}

# =============================================================================
# OUTPUT FORMATTING
# =============================================================================

format_status_color() {
    local status="$1"
    
    case "$status" in
        "running")
            echo -e "${COLORS[GREEN]}●${COLORS[RESET]} $status"
            ;;
        "exited"|"stopped")
            echo -e "${COLORS[RED]}●${COLORS[RESET]} $status"
            ;;
        "restarting")
            echo -e "${COLORS[YELLOW]}●${COLORS[RESET]} $status"
            ;;
        "paused")
            echo -e "${COLORS[BLUE]}●${COLORS[RESET]} $status"
            ;;
        *)
            echo -e "${COLORS[MAGENTA]}●${COLORS[RESET]} $status"
            ;;
    esac
}

format_health_color() {
    local health="$1"
    
    case "$health" in
        "healthy")
            echo -e "${COLORS[GREEN]}✓${COLORS[RESET]} $health"
            ;;
        "unhealthy")
            echo -e "${COLORS[RED]}✗${COLORS[RESET]} $health"
            ;;
        "starting")
            echo -e "${COLORS[YELLOW]}○${COLORS[RESET]} $health"
            ;;
        *)
            echo -e "${COLORS[MAGENTA]}?${COLORS[RESET]} $health"
            ;;
    esac
}

# =============================================================================
# TABLE OUTPUT
# =============================================================================

show_table_output() {
    local services=($(get_services_to_show))
    
    if [ ${#services[@]} -eq 0 ]; then
        log "INFO" "No services found matching criteria"
        return 0
    fi
    
    # Print header
    print_table_header
    
    # Print service information
    for service in "${services[@]}"; do
        print_service_row "$service"
    done
    
    # Print summary
    print_table_summary "${services[@]}"
}

print_table_header() {
    local header_line="╭─────────────────────┬──────────────┬─────────────────┬─────────────────────────────┬──────────────┬─────────────────────────────────────────────────────────╮"
    local header_text="│ Service             │ Status       │ Health          │ Ports                       │ Uptime       │ URL                                                     │"
    local separator="├─────────────────────┼──────────────┼─────────────────┼─────────────────────────────┼──────────────┼─────────────────────────────────────────────────────────┤"
    
    echo "$header_line"
    echo "$header_text"
    echo "$separator"
}

print_service_row() {
    local service="$1"
    local status=$(get_service_status "$service")
    local health=""
    local ports=$(get_service_ports "$service")
    local uptime=$(get_service_uptime "$service")
    local url="${SERVICE_URLS[$service]:-N/A}"
    local description="${SERVICE_DESCRIPTIONS[$service]:-No description}"
    
    # Get health if showing health info
    if [ "$SHOW_HEALTH" = true ]; then
        health=$(get_service_health "$service")
    else
        health="N/A"
    fi
    
    # Format status and health with colors
    local status_colored=$(format_status_color "$status")
    local health_colored=$(format_health_color "$health")
    
    # Truncate long values
    local service_display=$(printf "%-19s" "${service:0:19}")
    local ports_display=$(printf "%-27s" "${ports:0:27}")
    local uptime_display=$(printf "%-12s" "${uptime:0:12}")
    local url_display=$(printf "%-55s" "${url:0:55}")
    
    printf "│ %-19s │ %-20s │ %-23s │ %-27s │ %-12s │ %-55s │\n" \
        "$service_display" \
        "$status_colored" \
        "$health_colored" \
        "$ports_display" \
        "$uptime_display" \
        "$url_display"
}

print_table_summary() {
    local services=("$@")
    local total=${#services[@]}
    local running=0
    local stopped=0
    local healthy=0
    local unhealthy=0
    
    for service in "${services[@]}"; do
        local status=$(get_service_status "$service")
        local health=$(get_service_health "$service")
        
        case "$status" in
            "running") running=$((running + 1)) ;;
            "exited"|"stopped") stopped=$((stopped + 1)) ;;
        esac
        
        case "$health" in
            "healthy") healthy=$((healthy + 1)) ;;
            "unhealthy") unhealthy=$((unhealthy + 1)) ;;
        esac
    done
    
    local footer="╰─────────────────────┴──────────────┴─────────────────┴─────────────────────────────┴──────────────┴─────────────────────────────────────────────────────────╯"
    echo "$footer"
    echo ""
    echo -e "${COLORS[BOLD]}Summary:${COLORS[RESET]} $total services total │ ${COLORS[GREEN]}$running running${COLORS[RESET]} │ ${COLORS[RED]}$stopped stopped${COLORS[RESET]} │ ${COLORS[GREEN]}$healthy healthy${COLORS[RESET]} │ ${COLORS[RED]}$unhealthy unhealthy${COLORS[RESET]}"
}

# =============================================================================
# JSON OUTPUT
# =============================================================================

show_json_output() {
    local services=($(get_services_to_show))
    
    echo "{"
    echo "  \"microservices\": {"
    echo "    \"timestamp\": \"$(date -Iseconds)\","
    echo "    \"total_services\": ${#services[@]},"
    echo "    \"services\": ["
    
    local first=true
    for service in "${services[@]}"; do
        if [ "$first" = false ]; then
            echo ","
        fi
        first=false
        
        print_service_json "$service"
    done
    
    echo ""
    echo "    ]"
    echo "  }"
    echo "}"
}

print_service_json() {
    local service="$1"
    local status=$(get_service_status "$service")
    local health=$(get_service_health "$service")
    local ports=$(get_service_ports "$service")
    local uptime=$(get_service_uptime "$service")
    local url="${SERVICE_URLS[$service]}"
    local description="${SERVICE_DESCRIPTIONS[$service]}"
    
    cat << EOF
      {
        "name": "$service",
        "status": "$status",
        "health": "$health",
        "ports": "$ports",
        "uptime": "$uptime",
        "url": "$url",
        "description": "$description"
      }
EOF
}

# =============================================================================
# CSV OUTPUT
# =============================================================================

show_csv_output() {
    local services=($(get_services_to_show))
    
    # Print header
    echo "Service,Status,Health,Ports,Uptime,URL,Description"
    
    # Print data
    for service in "${services[@]}"; do
        local status=$(get_service_status "$service")
        local health=$(get_service_health "$service")
        local ports=$(get_service_ports "$service")
        local uptime=$(get_service_uptime "$service")
        local url="${SERVICE_URLS[$service]}"
        local description="${SERVICE_DESCRIPTIONS[$service]}"
        
        echo "$service,$status,$health,$ports,$uptime,$url,$description"
    done
}

# =============================================================================
# MARKDOWN OUTPUT
# =============================================================================

show_markdown_output() {
    local services=($(get_services_to_show))
    
    echo "# Microservices Status Report"
    echo ""
    echo "Generated: $(date)"
    echo ""
    echo "| Service | Status | Health | Ports | Uptime | URL | Description |"
    echo "|---------|--------|--------|-------|--------|-----|-------------|"
    
    for service in "${services[@]}"; do
        local status=$(get_service_status "$service")
        local health=$(get_service_health "$service")
        local ports=$(get_service_ports "$service")
        local uptime=$(get_service_uptime "$service")
        local url="${SERVICE_URLS[$service]}"
        local description="${SERVICE_DESCRIPTIONS[$service]}"
        
        echo "| $service | $status | $health | $ports | $uptime | $url | $description |"
    done
}

# =============================================================================
# URL ONLY OUTPUT
# =============================================================================

show_urls_only() {
    echo -e "${COLORS[BOLD]}${COLORS[CYAN]}🌐 Microservices Quick URL Reference${COLORS[RESET]}"
    echo "══════════════════════════════════════════════════════════════"
    echo ""
    
    # Group URLs by category
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        local files=(${=SERVICE_CATEGORIES[$category]})
        local category_urls=()
        
        for file in "${files[@]}"; do
            local service_name=$(basename "$file" .yml)
            local url="${SERVICE_URLS[$service_name]}"
            
            if [ -n "$url" ]; then
                local status=$(get_service_status "$service_name")
                local status_icon=""
                
                case "$status" in
                    "running") status_icon="${COLORS[GREEN]}●${COLORS[RESET]}" ;;
                    *) status_icon="${COLORS[RED]}●${COLORS[RESET]}" ;;
                esac
                
                category_urls+=("$status_icon $url - ${SERVICE_DESCRIPTIONS[$service_name]}")
            fi
        done
        
        if [ ${#category_urls[@]} -gt 0 ]; then
            echo -e "${COLORS[BOLD]}${category^^}:${COLORS[RESET]}"
            printf "  %s\n" "${category_urls[@]}"
            echo ""
        fi
    done
    
    echo "Direct Database Connections:"
    echo "  🔌 PostgreSQL: localhost:8502 (user: postgres)"
    echo "  🔌 MariaDB: localhost:8501 (user: root)"
    echo "  🔌 MySQL: localhost:8505 (user: root)"
    echo "  🔌 MongoDB: localhost:8503 (user: root)"
    echo "  🔌 Redis: localhost:8504"
    echo ""
    echo -e "${COLORS[YELLOW]}💡 Tip: Add entries to /etc/hosts or use trust-host.sh for HTTPS${COLORS[RESET]}"
}

# =============================================================================
# DETAILED VIEWS
# =============================================================================

show_detailed_logs() {
    local services=($(get_services_to_show))
    
    echo -e "${COLORS[BOLD]}📋 Service Logs${COLORS[RESET]}"
    echo "═══════════════════════════════════════════════════════════"
    
    for service in "${services[@]}"; do
        echo ""
        echo -e "${COLORS[CYAN]}📝 Logs for $service:${COLORS[RESET]}"
        echo "─────────────────────────────────────────────────────────"
        
        get_service_logs "$service"
        
        echo "─────────────────────────────────────────────────────────"
    done
}

show_network_info() {
    echo -e "${COLORS[BOLD]}🌐 Network Information${COLORS[RESET]}"
    echo "═══════════════════════════════════════════════════════════"
    
    # Show network details
    if execute_command "Show network info" \
       "${SUDO_PREFIX}${CONTAINER_RUNTIME} network ls" false false; then
        echo ""
        echo "Network Details for $NETWORK_NAME:"
        execute_command "Show network details" \
            "${SUDO_PREFIX}${CONTAINER_RUNTIME} network inspect $NETWORK_NAME" false false
    fi
}

show_volume_info() {
    echo -e "${COLORS[BOLD]}💾 Volume Information${COLORS[RESET]}"
    echo "═══════════════════════════════════════════════════════════"
    
    execute_command "Show volumes" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} volume ls" false false
    
    echo ""
    echo "Volume Usage:"
    execute_command "Show volume usage" \
        "${SUDO_PREFIX}${CONTAINER_RUNTIME} system df" false false
}

# =============================================================================
# CONTINUOUS MONITORING
# =============================================================================

continuous_monitoring() {
    local interval="$1"
    
    if [ "$interval" -le 0 ]; then
        return 0
    fi
    
    log "INFO" "🔄 Starting continuous monitoring (refresh every ${interval}s). Press Ctrl+C to stop."
    
    trap 'echo -e "\n${COLORS[YELLOW]}Monitoring stopped${COLORS[RESET]}"; exit 0' INT
    
    while true; do
        clear
        echo -e "${COLORS[BOLD]}${COLORS[CYAN]}🔄 Microservices Live Monitor${COLORS[RESET]} - $(date)"
        echo "Refreshing every ${interval}s | Press Ctrl+C to stop"
        echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
        echo ""
        
        # Show services based on selected format
        case "$OUTPUT_FORMAT" in
            "table") show_table_output ;;
            "json") show_json_output ;;
            "csv") show_csv_output ;;
            "markdown") show_markdown_output ;;
        esac
        
        # Add system resource info
        echo ""
        echo -e "${COLORS[BOLD]}System Resources:${COLORS[RESET]}"
        if command -v free >/dev/null 2>&1; then
            echo -n "Memory: "
            free -h | grep '^Mem:' | awk '{printf "Used: %s/%s (%.1f%%)", $3, $2, ($3/$2)*100}'
            echo ""
        fi
        
        if command -v df >/dev/null 2>&1; then
            echo -n "Disk: "
            df -h / | tail -1 | awk '{printf "Used: %s/%s (%s)", $3, $2, $5}'
            echo ""
        fi
        
        sleep "$interval"
    done
}

# =============================================================================
# HEALTH CHECK DASHBOARD
# =============================================================================

show_health_dashboard() {
    local services=($(get_services_to_show))
    
    echo -e "${COLORS[BOLD]}${COLORS[GREEN]}🏥 Service Health Dashboard${COLORS[RESET]}"
    echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
    echo ""
    
    # Categorize services by health
    local healthy_services=()
    local unhealthy_services=()
    local starting_services=()
    local unknown_services=()
    
    for service in "${services[@]}"; do
        local health=$(get_service_health "$service")
        case "$health" in
            "healthy") healthy_services+=("$service") ;;
            "unhealthy") unhealthy_services+=("$service") ;;
            "starting") starting_services+=("$service") ;;
            *) unknown_services+=("$service") ;;
        esac
    done
    
    # Show health summary
    echo -e "${COLORS[BOLD]}Health Summary:${COLORS[RESET]}"
    echo -e "  ${COLORS[GREEN]}✓ Healthy: ${#healthy_services[@]}${COLORS[RESET]}"
    echo -e "  ${COLORS[RED]}✗ Unhealthy: ${#unhealthy_services[@]}${COLORS[RESET]}"
    echo -e "  ${COLORS[YELLOW]}○ Starting: ${#starting_services[@]}${COLORS[RESET]}"
    echo -e "  ${COLORS[MAGENTA]}? Unknown: ${#unknown_services[@]}${COLORS[RESET]}"
    echo ""
    
    # Show detailed health info for problematic services
    if [ ${#unhealthy_services[@]} -gt 0 ]; then
        echo -e "${COLORS[BOLD]}${COLORS[RED]}🚨 Unhealthy Services:${COLORS[RESET]}"
        for service in "${unhealthy_services[@]}"; do
            local status=$(get_service_status "$service")
            echo -e "  ${COLORS[RED]}●${COLORS[RESET]} $service (status: $status)"
            
            # Show recent logs for unhealthy services
            echo "    Recent logs:"
            ${SUDO_PREFIX}${CONTAINER_RUNTIME} logs --tail 3 "$service" 2>/dev/null | sed 's/^/      /' || echo "      No logs available"
        done
        echo ""
    fi
    
    if [ ${#starting_services[@]} -gt 0 ]; then
        echo -e "${COLORS[BOLD]}${COLORS[YELLOW]}⏳ Starting Services:${COLORS[RESET]}"
        for service in "${starting_services[@]}"; do
            local uptime=$(get_service_uptime "$service")
            echo -e "  ${COLORS[YELLOW]}○${COLORS[RESET]} $service (uptime: $uptime)"
        done
        echo ""
    fi
    
    # Show connectivity test results
    echo -e "${COLORS[BOLD]}🔌 Connectivity Tests:${COLORS[RESET]}"
    for service in "${healthy_services[@]}"; do
        local url="${SERVICE_URLS[$service]}"
        if [ -n "$url" ]; then
            echo -n "  Testing $service ($url)... "
            if command -v curl >/dev/null 2>&1; then
                if curl -s --connect-timeout 3 --max-time 5 "$url" >/dev/null 2>&1; then
                    echo -e "${COLORS[GREEN]}✓ OK${COLORS[RESET]}"
                else
                    echo -e "${COLORS[RED]}✗ Failed${COLORS[RESET]}"
                fi
            else
                echo -e "${COLORS[MAGENTA]}? No curl available${COLORS[RESET]}"
            fi
        fi
    done
}

# =============================================================================
# STATISTICS DASHBOARD
# =============================================================================

show_stats_dashboard() {
    local services=($(get_services_to_show))
    
    echo -e "${COLORS[BOLD]}${COLORS[BLUE]}📊 Resource Usage Statistics${COLORS[RESET]}"
    echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
    echo ""
    
    # Check if stats are available
    if ! command -v "${CONTAINER_RUNTIME}" >/dev/null 2>&1; then
        log "ERROR" "Container runtime not available for stats"
        return 1
    fi
    
    # Show stats header
    printf "%-20s %-10s %-15s %-10s %-20s %-15s\n" \
        "Service" "CPU %" "Memory" "Mem %" "Network I/O" "Block I/O"
    echo "─────────────────────────────────────────────────────────────────────────────────────────────────────"
    
    # Get stats for each running service
    for service in "${services[@]}"; do
        local status=$(get_service_status "$service")
        if [ "$status" = "running" ]; then
            local stats=$(get_service_stats "$service")
            if [ "$stats" != "Container not found" ]; then
                # Parse stats output
                local cpu=$(echo "$stats" | awk '{print $1}')
                local mem_usage=$(echo "$stats" | awk '{print $2}')
                local mem_percent=$(echo "$stats" | awk '{print $3}')
                local net_io=$(echo "$stats" | awk '{print $4}')
                local block_io=$(echo "$stats" | awk '{print $5}')
                
                printf "%-20s %-10s %-15s %-10s %-20s %-15s\n" \
                    "$service" "$cpu" "$mem_usage" "$mem_percent" "$net_io" "$block_io"
            else
                printf "%-20s %-10s %-15s %-10s %-20s %-15s\n" \
                    "$service" "N/A" "N/A" "N/A" "N/A" "N/A"
            fi
        fi
    done
    
    echo ""
    
    # Show system-wide container stats
    echo -e "${COLORS[BOLD]}System Overview:${COLORS[RESET]}"
    if execute_command "Show system info" \
       "${SUDO_PREFIX}${CONTAINER_RUNTIME} system df" false false; then
        echo ""
    fi
    
    # Show top resource consumers
    echo -e "${COLORS[BOLD]}Top Resource Consumers:${COLORS[RESET]}"
    echo "Getting live stats..."
    ${SUDO_PREFIX}${CONTAINER_RUNTIME} stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | head -6
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    parse_common_args "$@"
    parse_show_args "$@"
    
    # Show URLs only if requested
    if [ "$SHOW_URLS_ONLY" = true ]; then
        show_urls_only
        exit 0
    fi
    
    # Set up continuous monitoring if requested
    if [ "$REFRESH_INTERVAL" -gt 0 ]; then
        continuous_monitoring "$REFRESH_INTERVAL"
        exit 0
    fi
    
    # Show different views based on options
    if [ "$SHOW_LOGS" = true ]; then
        show_detailed_logs
        exit 0
    fi
    
    if [ "$SHOW_NETWORKS" = true ]; then
        show_network_info
        exit 0
    fi
    
    if [ "$SHOW_VOLUMES" = true ]; then
        show_volume_info
        exit 0
    fi
    
    if [ "$SHOW_STATS" = true ]; then
        show_stats_dashboard
        exit 0
    fi
    
    if [ "$SHOW_HEALTH" = true ]; then
        show_health_dashboard
        exit 0
    fi
    
    # Default: show services in requested format
    case "$OUTPUT_FORMAT" in
        "table")
            show_table_output
            ;;
        "json")
            show_json_output
            ;;
        "csv")
            show_csv_output
            ;;
        "markdown")
            show_markdown_output
            ;;
    esac
    
    # Show quick tips
    if [ "$OUTPUT_FORMAT" = "table" ] && [ "$DRY_RUN" = false ]; then
        echo ""
        echo -e "${COLORS[YELLOW]}💡 Quick Tips:${COLORS[RESET]}"
        echo "  • Use -u for quick URL reference"
        echo "  • Use -H for health dashboard"
        echo "  • Use -T for resource statistics"  
        echo "  • Use -i 5 for live monitoring"
        echo "  • Use -L to view service logs"
        echo "  • Use -c <category> to filter by service type"
    fi
}

main "$@"