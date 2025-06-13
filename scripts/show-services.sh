#!/bin/zsh

# =============================================================================
# MICROSERVICES DIRECTORY SCRIPT
# =============================================================================
# Enhanced service directory with real-time status, health checks, and management

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_show_status=true
opt_show_health=true
opt_show_ports=true
opt_show_urls=true
opt_category_filter=""
opt_status_filter=""
opt_output_format="table"
opt_show_credentials=true
opt_export_file=""

# Color definitions
declare -A COLORS=(
    [RED]='\033[0;31m'
    [GREEN]='\033[0;32m'
    [YELLOW]='\033[0;33m'
    [BLUE]='\033[0;34m'
    [PURPLE]='\033[0;35m'
    [CYAN]='\033[0;36m'
    [WHITE]='\033[0;37m'
    [BOLD]='\033[1m'
    [NC]='\033[0m'
)

# Service status icons
declare -A STATUS_ICONS=(
    [running]="✅"
    [stopped]="❌"
    [paused]="⏸️"
    [restarting]="🔄"
    [exited]="🔴"
    [unknown]="❓"
)

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Enhanced service directory showing real-time status, health, and access information
    for all microservices in the architecture.

OPTIONS:
    -s, --sudo              Use sudo for container commands
    -e, --errors            Show detailed error messages
    --no-status             Hide container status information
    --no-health             Hide health check information
    --no-ports              Hide port information
    --no-urls               Hide URL information
    --no-credentials        Hide default credentials
    -c, --category FILTER   Show only specific categories (comma-separated)
    -S, --status FILTER     Show only containers with specific status
    -f, --format FORMAT     Output format: table, json, csv, simple
    -o, --output FILE       Export output to file
    -h, --help              Show this help message

CATEGORIES:
    database, db-tools, backend, analytics, ai-services, mail, project, erp, auth, proxy

STATUS FILTERS:
    running, stopped, paused, restarting, exited

OUTPUT FORMATS:
    table    - Formatted table (default)
    json     - JSON format
    csv      - CSV format  
    simple   - Simple text list

EXAMPLES:
    $0                                  # Show all services
    $0 -c database,backend             # Show only database and backend services
    $0 -S running                      # Show only running services
    $0 -f json -o services.json        # Export to JSON file
    $0 --no-credentials --no-health    # Minimal display
    $0 -c analytics -f csv             # Analytics services in CSV format

NOTES:
    - Real-time container status and health information
    - Automatic URL generation for both local and proxy access
    - Color-coded status indicators for quick visual scanning
    - Export capabilities for automation and documentation
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--sudo)
                opt_use_sudo=true
                shift
                ;;
            -e|--errors)
                opt_show_errors=true
                shift
                ;;
            --no-status)
                opt_show_status=false
                shift
                ;;
            --no-health)
                opt_show_health=false
                shift
                ;;
            --no-ports)
                opt_show_ports=false
                shift
                ;;
            --no-urls)
                opt_show_urls=false
                shift
                ;;
            --no-credentials)
                opt_show_credentials=false
                shift
                ;;
            -c|--category)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_category_filter="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a category value"
                fi
                ;;
            -S|--status)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_status_filter="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a status value"
                fi
                ;;
            -f|--format)
                if [[ -n "$2" && "$2" != -* ]]; then
                    case "$2" in
                        table|json|csv|simple)
                            opt_output_format="$2"
                            ;;
                        *)
                            handle_error "Invalid format: $2. Use table, json, csv, or simple"
                            ;;
                    esac
                    shift 2
                else
                    handle_error "Option $1 requires a format value"
                fi
                ;;
            -o|--output)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_export_file="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a filename"
                fi
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1. Use -h for help."
                ;;
        esac
    done
}

# =============================================================================
# SERVICE INFORMATION FUNCTIONS
# =============================================================================

# Service configuration with enhanced metadata
get_service_info() {
    local service_name="$1"
    
    case "$service_name" in
        # Database Services
        "mariadb")
            echo "database|MariaDB|SQL Database|8501|mariadb.test|MariaDB/MySQL compatible database"
            ;;
        "mysql")
            echo "database|MySQL|SQL Database|8505|mysql.test|MySQL database server"
            ;;
        "postgres")
            echo "database|PostgreSQL|SQL Database|8502|postgres.test|Advanced SQL database"
            ;;
        "mongodb")
            echo "database|MongoDB|NoSQL Database|8503|mongodb.test|Document-oriented database"
            ;;
        "redis")
            echo "database|Redis|Cache/Queue|8504|redis.test|In-memory data structure store"
            ;;
            
        # Database Tools
        "adminer")
            echo "db-tools|Adminer|DB Management|8082|adminer.test|Database management tool"
            ;;
        "phpmyadmin")
            echo "db-tools|phpMyAdmin|MySQL Admin|8083|phpmyadmin.test|MySQL administration interface"
            ;;
        "mongo-express")
            echo "db-tools|Mongo Express|MongoDB Admin|8084|mongodb.test|MongoDB administration interface"
            ;;
        "metabase")
            echo "db-tools|Metabase|Analytics|8085|metabase.test|Business intelligence and analytics"
            ;;
        "nocodb")
            echo "db-tools|NocoDB|No-Code DB|8086|nocodb.test|No-code database platform"
            ;;
        "pgadmin")
            echo "db-tools|pgAdmin|PostgreSQL Admin|8087|pgadmin.test|PostgreSQL administration"
            ;;
            
        # Backend Services
        "dotnet")
            echo "backend|.NET Core|Web Framework|8010,8011|dotnet.test|Microsoft .NET Core applications"
            ;;
        "go")
            echo "backend|Go|Web Framework|8020|go.test|Go applications and APIs"
            ;;
        "node")
            echo "backend|Node.js|Web Framework|8030|node.test|Node.js applications and APIs"
            ;;
        "php")
            echo "backend|PHP|Web Framework|8000,5173|php.test|PHP applications and frameworks"
            ;;
        "python")
            echo "backend|Python|Web Framework|8040|python.test|Python applications and APIs"
            ;;
            
        # Analytics Services
        "elasticsearch")
            echo "analytics|Elasticsearch|Search Engine|9130,9131|elasticsearch.test|Distributed search and analytics"
            ;;
        "kibana")
            echo "analytics|Kibana|Data Visualization|9120|kibana.test|Elasticsearch data visualization"
            ;;
        "logstash")
            echo "analytics|Logstash|Log Processing|5000,5044,9600|logstash.test|Log collection and processing"
            ;;
        "grafana")
            echo "analytics|Grafana|Monitoring|9001|grafana.test|Metrics and monitoring dashboards"
            ;;
        "prometheus")
            echo "analytics|Prometheus|Metrics|9090|prometheus.test|Metrics collection and alerting"
            ;;
        "matomo")
            echo "analytics|Matomo|Web Analytics|9010|matomo.test|Web analytics platform"
            ;;
            
        # AI Services
        "langflow")
            echo "ai-services|Langflow|AI Workflows|9110|langflow.test|Visual AI workflow builder"
            ;;
        "n8n")
            echo "ai-services|n8n|Automation|9100|n8n.test|Workflow automation platform"
            ;;
            
        # Mail Services
        "mailpit")
            echo "mail|Mailpit|Mail Testing|9200,9201|mailpit.test|Email testing and debugging"
            ;;
            
        # Project Services
        "gitea")
            echo "project|Gitea|Git Platform|9210,2222|gitea.test|Git repository platform"
            ;;
            
        # ERP Services
        "odoo")
            echo "erp|Odoo|ERP System|9300,9301|odoo.test|Enterprise Resource Planning"
            ;;
            
        # Auth Services
        "keycloak")
            echo "auth|Keycloak|Identity Provider|9400,9401|keycloak.test|Identity and access management"
            ;;
            
        # Proxy Services
        "nginx-proxy-manager")
            echo "proxy|Nginx Proxy|Reverse Proxy|80,443,81|nginx.test|Reverse proxy and SSL management"
            ;;
            
        *)
            echo "unknown|Unknown|Unknown|0|unknown.test|Unknown service"
            ;;
    esac
}

get_container_status() {
    local container_name="$1"
    local status
    
    status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' $container_name 2>/dev/null" || echo "not-found")
    echo "$status"
}

get_container_health() {
    local container_name="$1"
    local health
    
    health=$(eval "$CONTAINER_CMD inspect --format='{{.State.Health.Status}}' $container_name 2>/dev/null" || echo "no-healthcheck")
    
    # If no health check, check if container is running
    if [[ "$health" == "no-healthcheck" ]]; then
        local status
        status=$(get_container_status "$container_name")
        if [[ "$status" == "running" ]]; then
            health="running"
        else
            health="unknown"
        fi
    fi
    
    echo "$health"
}

get_container_uptime() {
    local container_name="$1"
    local started_at
    
    started_at=$(eval "$CONTAINER_CMD inspect --format='{{.State.StartedAt}}' $container_name 2>/dev/null" || echo "")
    
    if [[ -n "$started_at" && "$started_at" != "0001-01-01T00:00:00Z" ]]; then
        local current_time started_epoch current_epoch uptime_seconds
        current_time=$(date -u +"%Y-%m-%dT%H:%M:%S.000000000Z")
        
        # Convert to epoch (handling different date command formats)
        if date -d "$started_at" +%s >/dev/null 2>&1; then
            # GNU date
            started_epoch=$(date -d "$started_at" +%s)
            current_epoch=$(date -d "$current_time" +%s)
        else
            # BSD date (macOS)
            started_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "${started_at%.*}" +%s 2>/dev/null || echo "0")
            current_epoch=$(date +%s)
        fi
        
        if [[ "$started_epoch" -gt 0 && "$current_epoch" -gt 0 ]]; then
            uptime_seconds=$((current_epoch - started_epoch))
            
            if [[ $uptime_seconds -lt 60 ]]; then
                echo "${uptime_seconds}s"
            elif [[ $uptime_seconds -lt 3600 ]]; then
                echo "$((uptime_seconds / 60))m"
            elif [[ $uptime_seconds -lt 86400 ]]; then
                echo "$((uptime_seconds / 3600))h"
            else
                echo "$((uptime_seconds / 86400))d"
            fi
        else
            echo "unknown"
        fi
    else
        echo "not-started"
    fi
}

# =============================================================================
# SERVICE COLLECTION FUNCTIONS
# =============================================================================

collect_service_data() {
    local -A service_data
    local -a filtered_services
    
    # Get all running containers
    local containers
    containers=$(eval "$CONTAINER_CMD ps -a --format '{{.Names}}' $ERROR_REDIRECT" || echo "")
    
    # Filter containers based on our known services
    for container in ${(f)containers}; do
        local service_info
        service_info=$(get_service_info "$container")
        
        if [[ "$service_info" != "unknown|"* ]]; then
            local category="${service_info%%|*}"
            
            # Apply category filter
            if [[ -n "$opt_category_filter" ]]; then
                local -a requested_categories
                requested_categories=(${(s:,:)opt_category_filter})
                if [[ ! " ${requested_categories[*]} " =~ " $category " ]]; then
                    continue
                fi
            fi
            
            # Apply status filter
            if [[ -n "$opt_status_filter" ]]; then
                local status
                status=$(get_container_status "$container")
                if [[ "$status" != "$opt_status_filter" ]]; then
                    continue
                fi
            fi
            
            filtered_services+=("$container")
        fi
    done
    
    # Collect data for filtered services
    for service in "${filtered_services[@]}"; do
        local info status health uptime
        info=$(get_service_info "$service")
        
        if [[ "$opt_show_status" == "true" ]]; then
            status=$(get_container_status "$service")
        else
            status=""
        fi
        
        if [[ "$opt_show_health" == "true" ]]; then
            health=$(get_container_health "$service")
            uptime=$(get_container_uptime "$service")
        else
            health=""
            uptime=""
        fi
        
        service_data["$service"]="$info|$status|$health|$uptime"
    done
    
    # Export the associative array data
    for service in "${!service_data[@]}"; do
        echo "$service:${service_data[$service]}"
    done
}

# =============================================================================
# OUTPUT FORMATTING FUNCTIONS
# =============================================================================

format_table_output() {
    local -a service_lines
    service_lines=("$@")
    
    # Print header
    print_header
    
    # Group by category
    local -A categories
    for line in "${service_lines[@]}"; do
        local service="${line%%:*}"
        local data="${line#*:}"
        local category="${data%%|*}"
        
        if [[ -z "${categories[$category]}" ]]; then
            categories["$category"]="$line"
        else
            categories["$category"]="${categories[$category]}"$'\n'"$line"
        fi
    done
    
    # Print services by category
    local -a category_order=("proxy" "database" "db-tools" "backend" "analytics" "ai-services" "mail" "project" "erp" "auth")
    
    for category in "${category_order[@]}"; do
        if [[ -n "${categories[$category]}" ]]; then
            print_category_section "$category" "${categories[$category]}"
        fi
    done
    
    # Print footer with summary information
    print_footer "${service_lines[@]}"
}

print_header() {
    echo ""
    echo -e "${COLORS[BLUE]}${COLORS[BOLD]}===============================================${COLORS[NC]}"
    echo -e "${COLORS[BLUE]}${COLORS[BOLD]}      MICROSERVICES ARCHITECTURE DIRECTORY     ${COLORS[NC]}"
    echo -e "${COLORS[BLUE]}${COLORS[BOLD]}===============================================${COLORS[NC]}"
    echo ""
}

print_category_section() {
    local category="$1"
    local category_data="$2"
    
    # Category headers with emojis
    local category_display
    case "$category" in
        "proxy") category_display="🌐 REVERSE PROXY & SSL" ;;
        "database") category_display="🗄️  DATABASE SERVICES" ;;
        "db-tools") category_display="🔧 DATABASE MANAGEMENT TOOLS" ;;
        "backend") category_display="⚙️  BACKEND SERVICES" ;;
        "analytics") category_display="📊 ANALYTICS & MONITORING" ;;
        "ai-services") category_display="🤖 AI & WORKFLOW SERVICES" ;;
        "mail") category_display="📧 MAIL SERVICES" ;;
        "project") category_display="📋 PROJECT MANAGEMENT" ;;
        "erp") category_display="🏢 BUSINESS APPLICATIONS" ;;
        "auth") category_display="🔐 AUTHENTICATION" ;;
        *) category_display="❓ $category" ;;
    esac
    
    echo -e "${COLORS[BLUE]}=== $category_display ===${COLORS[NC]}"
    echo ""
    
    # Process each service in the category
    local -a lines
    lines=(${(f)category_data})
    
    for line in "${lines[@]}"; do
        print_service_line "$line"
    done
    
    echo ""
}

print_service_line() {
    local line="$1"
    local service="${line%%:*}"
    local data="${line#*:}"
    
    # Parse service data
    local -a fields
    fields=(${(s:|:)data})
    
    local category="${fields[1]}"
    local display_name="${fields[2]}"
    local service_type="${fields[3]}"
    local ports="${fields[4]}"
    local domain="${fields[5]}"
    local description="${fields[6]}"
    local status="${fields[7]}"
    local health="${fields[8]}"
    local uptime="${fields[9]}"
    
    # Color-code based on status
    local status_color="${COLORS[NC]}"
    local status_icon="${STATUS_ICONS[unknown]}"
    
    case "$status" in
        "running")
            status_color="${COLORS[GREEN]}"
            status_icon="${STATUS_ICONS[running]}"
            ;;
        "stopped"|"exited")
            status_color="${COLORS[RED]}"
            status_icon="${STATUS_ICONS[stopped]}"
            ;;
        "paused")
            status_color="${COLORS[YELLOW]}"
            status_icon="${STATUS_ICONS[paused]}"
            ;;
        "restarting")
            status_color="${COLORS[PURPLE]}"
            status_icon="${STATUS_ICONS[restarting]}"
            ;;
    esac
    
    # Print service information
    echo -e "${status_color}${COLORS[BOLD]}$display_name${COLORS[NC]} ${status_icon}"
    
    if [[ "$opt_show_status" == "true" && -n "$status" ]]; then
        echo -e "  ${COLORS[YELLOW]}Status:${COLORS[NC]} ${status_color}$status${COLORS[NC]}"
        if [[ -n "$uptime" && "$uptime" != "not-started" ]]; then
            echo -e "  ${COLORS[YELLOW]}Uptime:${COLORS[NC]} $uptime"
        fi
    fi
    
    if [[ "$opt_show_health" == "true" && -n "$health" && "$health" != "no-healthcheck" ]]; then
        local health_color="${COLORS[NC]}"
        case "$health" in
            "healthy"|"running") health_color="${COLORS[GREEN]}" ;;
            "unhealthy") health_color="${COLORS[RED]}" ;;
            "starting") health_color="${COLORS[YELLOW]}" ;;
        esac
        echo -e "  ${COLORS[YELLOW]}Health:${COLORS[NC]} ${health_color}$health${COLORS[NC]}"
    fi
    
    if [[ "$opt_show_ports" == "true" && "$ports" != "0" ]]; then
        echo -e "  ${COLORS[YELLOW]}Ports:${COLORS[NC]} $ports"
    fi
    
    if [[ "$opt_show_urls" == "true" ]]; then
        # Generate URLs
        local -a port_list
        port_list=(${(s:,:)ports})
        local main_port="${port_list[1]}"
        
        if [[ "$main_port" != "0" ]]; then
            echo -e "  ${COLORS[YELLOW]}Local:${COLORS[NC]} http://localhost:$main_port"
        fi
        
        if [[ "$domain" != "unknown.test" ]]; then
            echo -e "  ${COLORS[YELLOW]}Proxy:${COLORS[NC]} https://$domain"
        fi
    fi
    
    echo -e "  ${COLORS[CYAN]}Description:${COLORS[NC]} $description"
    echo ""
}

format_json_output() {
    local -a service_lines
    service_lines=("$@")
    
    echo "{"
    echo '  "microservices": {'
    echo '    "timestamp": "'$(date -Iseconds)'",'
    echo '    "container_runtime": "'$CONTAINER_RUNTIME'",'
    echo '    "services": ['
    
    local first=true
    for line in "${service_lines[@]}"; do
        [[ "$first" == "false" ]] && echo ","
        first=false
        
        local service="${line%%:*}"
        local data="${line#*:}"
        local -a fields
        fields=(${(s:|:)data})
        
        cat << EOF
      {
        "name": "$service",
        "display_name": "${fields[2]}",
        "category": "${fields[1]}",
        "type": "${fields[3]}",
        "ports": "${fields[4]}",
        "domain": "${fields[5]}",
        "description": "${fields[6]}",
        "status": "${fields[7]}",
        "health": "${fields[8]}",
        "uptime": "${fields[9]}"
      }
EOF
    done
    
    echo ""
    echo "    ]"
    echo "  }"
    echo "}"
}

format_csv_output() {
    local -a service_lines
    service_lines=("$@")
    
    echo "Service,Display_Name,Category,Type,Ports,Domain,Description,Status,Health,Uptime"
    
    for line in "${service_lines[@]}"; do
        local service="${line%%:*}"
        local data="${line#*:}"
        local -a fields
        fields=(${(s:|:)data})
        
        echo "$service,${fields[2]},${fields[1]},${fields[3]},${fields[4]},${fields[5]},\"${fields[6]}\",${fields[7]},${fields[8]},${fields[9]}"
    done
}

format_simple_output() {
    local -a service_lines
    service_lines=("$@")
    
    for line in "${service_lines[@]}"; do
        local service="${line%%:*}"
        local data="${line#*:}"
        local -a fields
        fields=(${(s:|:)data})
        
        local status_icon="${STATUS_ICONS[${fields[7]}]:-❓}"
        echo "$status_icon ${fields[2]} (${fields[1]}) - ${fields[7]}"
        
        if [[ "$opt_show_urls" == "true" ]]; then
            local -a port_list
            port_list=(${(s:,:)${fields[4]}})
            local main_port="${port_list[1]}"
            
            if [[ "$main_port" != "0" ]]; then
                echo "    Local: http://localhost:$main_port"
            fi
            
            if [[ "${fields[5]}" != "unknown.test" ]]; then
                echo "    Proxy: https://${fields[5]}"
            fi
        fi
        echo ""
    done
}

print_footer() {
    local -a service_lines
    service_lines=("$@")
    
    if [[ "$opt_show_credentials" == "true" ]]; then
        echo -e "${COLORS[BLUE]}=== 🔐 DEFAULT CREDENTIALS ===${COLORS[NC]}"
        echo ""
        echo -e "  ${COLORS[YELLOW]}Username:${COLORS[NC]} admin"
        echo -e "  ${COLORS[YELLOW]}Password:${COLORS[NC]} 123456"
        echo -e "  ${COLORS[YELLOW]}Email:${COLORS[NC]} admin@site.test"
        echo ""
    fi
    
    echo -e "${COLORS[BLUE]}=== 📋 QUICK MANAGEMENT COMMANDS ===${COLORS[NC]}"
    echo ""
    echo "  Start all services:    $SCRIPT_DIR/start-services.sh"
    echo "  Stop all services:     $SCRIPT_DIR/stop-services.sh"
    echo "  Setup databases:       $SCRIPT_DIR/setup-databases.sh -a"
    echo "  Setup SSL:            $SCRIPT_DIR/setup-ssl.sh"
    echo "  Install SSL trust:     $SCRIPT_DIR/trust-host.sh"
    echo ""
    
    # Service summary
    local total=0 running=0 stopped=0
    for line in "${service_lines[@]}"; do
        ((total++))
        local status="${line##*|}"
        status="${status%%|*}"
        case "$status" in
            "running") ((running++)) ;;
            "stopped"|"exited") ((stopped++)) ;;
        esac
    done
    
    echo -e "${COLORS[BLUE]}=== 📊 SERVICE SUMMARY ===${COLORS[NC]}"
    echo ""
    echo -e "  ${COLORS[GREEN]}Running:${COLORS[NC]} $running"
    echo -e "  ${COLORS[RED]}Stopped:${COLORS[NC]} $stopped"
    echo -e "  ${COLORS[BLUE]}Total:${COLORS[NC]} $total"
    echo ""
    echo -e "${COLORS[BLUE]}===============================================${COLORS[NC]}"
}

# =============================================================================
# MAIN EXECUTION FUNCTIONS
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Collect service data
    local -a service_data
    service_data=($(collect_service_data))
    
    if [[ ${#service_data[@]} -eq 0 ]]; then
        print_status "warning" "No services found matching the specified criteria"
        exit 0
    fi
    
    # Generate output based on format
    local output=""
    case "$opt_output_format" in
        "table")
            output=$(format_table_output "${service_data[@]}")
            ;;
        "json")
            output=$(format_json_output "${service_data[@]}")
            ;;
        "csv")
            output=$(format_csv_output "${service_data[@]}")
            ;;
        "simple")
            output=$(format_simple_output "${service_data[@]}")
            ;;
    esac
    
    # Display or export output
    if [[ -n "$opt_export_file" ]]; then
        echo "$output" > "$opt_export_file"
        print_status "success" "Service information exported to: $opt_export_file"
    else
        echo "$output"
    fi
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi