#!/bin/zsh

# =============================================================================
# TRAEFIK MANAGEMENT SCRIPT
# =============================================================================
# Comprehensive management script for Traefik proxy operations
# Provides commands for configuration, monitoring, and troubleshooting

# Source the central configuration
. "$(dirname "$0")/config.sh"

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_use_sudo="$DEFAULT_SUDO"
opt_show_errors=false
opt_command=""
opt_service=""
opt_domain=""
opt_port=""
opt_format="table"

# =============================================================================
# USAGE & HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 COMMAND [OPTIONS]

DESCRIPTION:
    Comprehensive Traefik management script for microservices architecture.
    Provides commands for configuration, monitoring, and troubleshooting.

COMMANDS:
    status                          Show Traefik status and health
    routes                          List all configured routes
    services                        List all configured services
    dashboard                       Open Traefik dashboard URL
    api                             Show Traefik API information
    reload                          Reload Traefik configuration
    logs [tail|follow]              Show Traefik logs
    add-service SERVICE DOMAIN PORT Add a new service route
    remove-service SERVICE          Remove a service route
    test-route DOMAIN               Test if a route is accessible
    switch-to-nginx                 Switch back to Nginx Proxy Manager
    switch-to-traefik               Switch to Traefik (default)
    backup-config                   Backup current Traefik configuration
    restore-config [FILE]           Restore Traefik configuration

OPTIONS:
    -s, --sudo                      Use sudo for container commands
    -e, --errors                    Show detailed error messages
    -f, --format FORMAT             Output format (table, json, yaml)
    -h, --help                      Show this help message

EXAMPLES:
    $0 status                       # Check Traefik status
    $0 routes                       # List all routes
    $0 dashboard                    # Open dashboard
    $0 add-service myapp myapp.test 3000  # Add new service
    $0 test-route nginx.test        # Test route accessibility
    $0 logs follow                  # Follow logs in real-time
    $0 reload                       # Reload configuration

DASHBOARD:
    Access the Traefik dashboard at: https://traefik.test
    Default credentials: admin / 123456

API ENDPOINTS:
    Dashboard: http://localhost:8080
    API:       http://localhost:8080/api/rawdata
    Health:    http://localhost:8080/ping
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    if [[ $# -eq 0 ]]; then
        print_status "error" "No command specified"
        show_usage
        exit 1
    fi
    
    opt_command="$1"
    shift
    
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
            -f|--format)
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_format="$2"
                    shift 2
                else
                    handle_error "Option $1 requires a format value"
                fi
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                # Handle positional arguments based on command
                case "$opt_command" in
                    "add-service")
                        if [[ -z "$opt_service" ]]; then
                            opt_service="$1"
                        elif [[ -z "$opt_domain" ]]; then
                            opt_domain="$1"
                        elif [[ -z "$opt_port" ]]; then
                            opt_port="$1"
                        fi
                        ;;
                    "remove-service"|"test-route")
                        if [[ -z "$opt_service" ]]; then
                            opt_service="$1"
                        fi
                        ;;
                    "logs")
                        opt_service="$1"  # tail, follow, etc.
                        ;;
                    "restore-config")
                        opt_service="$1"  # backup file path
                        ;;
                    *)
                        handle_error "Unknown argument for command $opt_command: $1"
                        ;;
                esac
                shift
                ;;
        esac
    done
}

# =============================================================================
# TRAEFIK STATUS AND HEALTH FUNCTIONS
# =============================================================================

show_traefik_status() {
    print_status "step" "Checking Traefik status..."
    
    # Check container status
    if ! eval "$CONTAINER_CMD container exists traefik $ERROR_REDIRECT"; then
        print_status "error" "Traefik container does not exist"
        return 1
    fi
    
    local container_status
    container_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Status}}' traefik 2>/dev/null" || echo "unknown")
    
    local health_status=""
    if eval "$CONTAINER_CMD inspect --format='{{.State.Health.Status}}' traefik >/dev/null 2>&1"; then
        health_status=$(eval "$CONTAINER_CMD inspect --format='{{.State.Health.Status}}' traefik 2>/dev/null")
    fi
    
    # Get uptime
    local started_at
    started_at=$(eval "$CONTAINER_CMD inspect --format='{{.State.StartedAt}}' traefik 2>/dev/null" | cut -d'T' -f1)
    
    # Check API accessibility
    local api_status="❌ Not accessible"
    if curl -s "http://localhost:8080/ping" >/dev/null 2>&1; then
        api_status="✅ Accessible"
    fi
    
    # Display status
    echo ""
    echo "🚀 Traefik Status Report"
    echo "======================="
    echo "Container Status: $container_status"
    [[ -n "$health_status" ]] && echo "Health Status: $health_status"
    echo "Started At: $started_at"
    echo "API Status: $api_status"
    echo "Dashboard: https://traefik.test"
    echo "API Endpoint: http://localhost:8080"
    echo ""
    
    # Show ports
    echo "📡 Port Mappings:"
    eval "$CONTAINER_CMD port traefik 2>/dev/null" | while read line; do
        echo "  $line"
    done
    echo ""
    
    # Show network information
    echo "🌐 Network Information:"
    local networks
    networks=$(eval "$CONTAINER_CMD inspect --format='{{range .NetworkSettings.Networks}}{{.NetworkID}} {{end}}' traefik 2>/dev/null")
    echo "  Networks: $networks"
    echo ""
    
    if [[ "$container_status" == "running" ]]; then
        print_status "success" "Traefik is running properly"
        return 0
    else
        print_status "error" "Traefik is not running properly"
        return 1
    fi
}

show_traefik_routes() {
    print_status "step" "Fetching Traefik routes..."
    
    if ! validate_traefik_setup; then
        return 1
    fi
    
    case "$opt_format" in
        "json")
            curl -s "http://localhost:8080/api/http/routers" 2>/dev/null | jq '.' 2>/dev/null || echo "[]"
            ;;
        "yaml")
            curl -s "http://localhost:8080/api/http/routers" 2>/dev/null | yq eval -P '.' 2>/dev/null || echo "No routes found"
            ;;
        *)
            echo ""
            echo "🛣️  Traefik Routes"
            echo "=================="
            printf "%-20s %-30s %-20s %-10s\n" "NAME" "RULE" "SERVICE" "TLS"
            printf "%-20s %-30s %-20s %-10s\n" "----" "----" "-------" "---"
            
            local routes_json
            routes_json=$(curl -s "http://localhost:8080/api/http/routers" 2>/dev/null)
            
            if [[ $? -eq 0 && -n "$routes_json" ]]; then
                echo "$routes_json" | jq -r 'to_entries[] | "\(.key) \(.value.rule // "N/A") \(.value.service // "N/A") \(if .value.tls then "Yes" else "No" end)"' 2>/dev/null | \
                while read name rule service tls; do
                    printf "%-20s %-30s %-20s %-10s\n" "$name" "$rule" "$service" "$tls"
                done
            else
                echo "No routes found or API not accessible"
            fi
            echo ""
            ;;
    esac
}

show_traefik_services() {
    print_status "step" "Fetching Traefik services..."
    
    if ! validate_traefik_setup; then
        return 1
    fi
    
    case "$opt_format" in
        "json")
            curl -s "http://localhost:8080/api/http/services" 2>/dev/null | jq '.' 2>/dev/null || echo "[]"
            ;;
        "yaml")
            curl -s "http://localhost:8080/api/http/services" 2>/dev/null | yq eval -P '.' 2>/dev/null || echo "No services found"
            ;;
        *)
            echo ""
            echo "🔧 Traefik Services"
            echo "==================="
            printf "%-20s %-40s %-15s\n" "NAME" "SERVERS" "STATUS"
            printf "%-20s %-40s %-15s\n" "----" "-------" "------"
            
            local services_json
            services_json=$(curl -s "http://localhost:8080/api/http/services" 2>/dev/null)
            
            if [[ $? -eq 0 && -n "$services_json" ]]; then
                echo "$services_json" | jq -r 'to_entries[] | "\(.key) \(.value.loadBalancer.servers[0].url // "N/A") \(.value.status // "Unknown")"' 2>/dev/null | \
                while read name url status; do
                    printf "%-20s %-40s %-15s\n" "$name" "$url" "$status"
                done
            else
                echo "No services found or API not accessible"
            fi
            echo ""
            ;;
    esac
}

# =============================================================================
# TRAEFIK CONFIGURATION FUNCTIONS
# =============================================================================

add_traefik_service() {
    local service="$opt_service"
    local domain="$opt_domain"
    local port="$opt_port"
    
    if [[ -z "$service" || -z "$domain" || -z "$port" ]]; then
        handle_error "Missing required parameters: SERVICE DOMAIN PORT"
    fi
    
    print_status "step" "Adding service $service to Traefik..."
    
    local dynamic_config="/tmp/traefik-add-service.yml"
    local service_config="$PROJECT_ROOT/config/traefik/dynamic/custom-services.yml"
    
    # Create custom services file if it doesn't exist
    if [[ ! -f "$service_config" ]]; then
        cat > "$service_config" << 'EOF'
# Custom services configuration
http:
  routers: {}
  services: {}
EOF
    fi
    
    # Add the new service configuration
    cat >> "$service_config" << EOF

  # Added by manage-traefik.sh
  routers:
    ${service}:
      rule: "Host(\`${domain}\`)"
      service: "${service}"
      tls: {}
  
  services:
    ${service}:
      loadBalancer:
        servers:
          - url: "http://${service}:${port}"
EOF
    
    print_status "success" "Service $service added successfully"
    print_status "info" "Route: https://$domain -> http://$service:$port"
    
    # Reload Traefik configuration
    reload_traefik_config
}

remove_traefik_service() {
    local service="$opt_service"
    
    if [[ -z "$service" ]]; then
        handle_error "Missing service name"
    fi
    
    print_status "step" "Removing service $service from Traefik..."
    
    local service_config="$PROJECT_ROOT/config/traefik/dynamic/custom-services.yml"
    
    if [[ ! -f "$service_config" ]]; then
        print_status "warning" "Custom services configuration file not found"
        return 1
    fi
    
    # Remove service configuration (simplified - in production use proper YAML parsing)
    sed -i "/${service}:/,+10d" "$service_config" 2>/dev/null || true
    
    print_status "success" "Service $service removed"
    
    # Reload Traefik configuration
    reload_traefik_config
}

test_traefik_route() {
    local domain="$opt_service"  # Using service field for domain
    
    if [[ -z "$domain" ]]; then
        handle_error "Missing domain to test"
    fi
    
    print_status "step" "Testing route accessibility for $domain..."
    
    # Test HTTP redirect
    local http_status
    http_status=$(curl -s -o /dev/null -w "%{http_code}" "http://$domain" 2>/dev/null || echo "000")
    
    # Test HTTPS
    local https_status
    https_status=$(curl -s -o /dev/null -w "%{http_code}" -k "https://$domain" 2>/dev/null || echo "000")
    
    echo ""
    echo "🧪 Route Test Results for $domain"
    echo "=================================="
    echo "HTTP Status:  $http_status"
    echo "HTTPS Status: $https_status"
    
    if [[ "$http_status" == "301" || "$http_status" == "302" ]]; then
        print_status "success" "HTTP redirect working (status: $http_status)"
    elif [[ "$http_status" == "200" ]]; then
        print_status "success" "HTTP accessible (status: $http_status)"
    else
        print_status "warning" "HTTP issues (status: $http_status)"
    fi
    
    if [[ "$https_status" == "200" ]]; then
        print_status "success" "HTTPS accessible (status: $https_status)"
    else
        print_status "warning" "HTTPS issues (status: $https_status)"
    fi
    
    # Test DNS resolution
    if command -v nslookup >/dev/null 2>&1; then
        echo ""
        echo "🌐 DNS Resolution:"
        nslookup "$domain" 2>/dev/null || echo "DNS resolution failed"
    fi
    echo ""
}

# =============================================================================
# TRAEFIK OPERATIONS FUNCTIONS
# =============================================================================

show_traefik_logs() {
    local log_type="${opt_service:-tail}"
    
    print_status "step" "Showing Traefik logs ($log_type)..."
    
    if ! validate_traefik_setup; then
        return 1
    fi
    
    case "$log_type" in
        "follow"|"f")
            eval "$CONTAINER_CMD logs -f traefik"
            ;;
        "tail"|"t"|*)
            eval "$CONTAINER_CMD logs --tail 100 traefik"
            ;;
    esac
}

open_traefik_dashboard() {
    local dashboard_url="https://traefik.test"
    
    print_status "info" "Traefik Dashboard: $dashboard_url"
    print_status "info" "API Endpoint: http://localhost:8080"
    print_status "info" "Default credentials: admin / 123456"
    
    # Try to open in browser
    if command -v open >/dev/null 2>&1; then
        open "$dashboard_url"
    elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open "$dashboard_url"
    elif command -v start >/dev/null 2>&1; then
        start "$dashboard_url"
    else
        echo "Please open $dashboard_url in your browser"
    fi
}

backup_traefik_config() {
    local backup_dir="$PROJECT_ROOT/backups/traefik"
    local backup_file="$backup_dir/traefik-config-$(date +%Y%m%d-%H%M%S).tar.gz"
    
    print_status "step" "Backing up Traefik configuration..."
    
    mkdir -p "$backup_dir"
    
    # Create backup
    tar -czf "$backup_file" \
        -C "$PROJECT_ROOT" \
        config/traefik \
        compose/proxy/traefik.yml \
        2>/dev/null
    
    if [[ -f "$backup_file" ]]; then
        print_status "success" "Configuration backed up to: $backup_file"
    else
        print_status "error" "Backup failed"
        return 1
    fi
}

restore_traefik_config() {
    local backup_file="$opt_service"
    
    if [[ -z "$backup_file" ]]; then
        # List available backups
        local backup_dir="$PROJECT_ROOT/backups/traefik"
        if [[ -d "$backup_dir" ]]; then
            echo "Available backups:"
            ls -la "$backup_dir"/*.tar.gz 2>/dev/null || echo "No backups found"
        else
            print_status "error" "No backup directory found"
        fi
        return 1
    fi
    
    if [[ ! -f "$backup_file" ]]; then
        handle_error "Backup file not found: $backup_file"
    fi
    
    print_status "step" "Restoring Traefik configuration from $backup_file..."
    
    # Stop Traefik
    eval "$CONTAINER_CMD stop traefik $ERROR_REDIRECT" || true
    
    # Restore configuration
    tar -xzf "$backup_file" -C "$PROJECT_ROOT" 2>/dev/null
    
    # Restart Traefik
    eval "$CONTAINER_CMD start traefik $ERROR_REDIRECT"
    
    print_status "success" "Configuration restored successfully"
}

switch_proxy_provider() {
    local provider="$1"
    
    case "$provider" in
        "nginx")
            print_status "step" "Switching to Nginx Proxy Manager..."
            
            # Stop Traefik
            eval "$CONTAINER_CMD stop traefik $ERROR_REDIRECT" || true
            
            # Update config.sh to use nginx
            switch_proxy_provider "nginx"
            
            # Start Nginx Proxy Manager
            start_service_category "proxy"
            
            print_status "success" "Switched to Nginx Proxy Manager"
            print_status "info" "Access dashboard at: http://localhost:81"
            ;;
        "traefik")
            print_status "step" "Switching to Traefik..."
            
            # Stop Nginx Proxy Manager
            eval "$CONTAINER_CMD stop nginx-proxy-manager $ERROR_REDIRECT" || true
            
            # Update config.sh to use traefik
            switch_proxy_provider "traefik"
            
            # Start Traefik
            start_service_category "proxy"
            
            print_status "success" "Switched to Traefik"
            print_status "info" "Access dashboard at: https://traefik.test"
            ;;
        *)
            handle_error "Unknown provider: $provider"
            ;;
    esac
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Set up command context based on options
    setup_command_context "$opt_use_sudo" "$opt_show_errors"
    
    # Execute the requested command
    case "$opt_command" in
        "status")
            show_traefik_status
            ;;
        "routes")
            show_traefik_routes
            ;;
        "services")
            show_traefik_services
            ;;
        "dashboard")
            open_traefik_dashboard
            ;;
        "api")
            get_traefik_api
            ;;
        "reload")
            reload_traefik_config
            ;;
        "logs")
            show_traefik_logs
            ;;
        "add-service")
            add_traefik_service
            ;;
        "remove-service")
            remove_traefik_service
            ;;
        "test-route")
            test_traefik_route
            ;;
        "switch-to-nginx")
            switch_proxy_provider "nginx"
            ;;
        "switch-to-traefik")
            switch_proxy_provider "traefik"
            ;;
        "backup-config")
            backup_traefik_config
            ;;
        "restore-config")
            restore_traefik_config
            ;;
        *)
            print_status "error" "Unknown command: $opt_command"
            show_usage
            exit 1
            ;;
    esac
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi