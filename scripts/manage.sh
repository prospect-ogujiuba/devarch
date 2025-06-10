#!/bin/bash
# =================================================================
# DEVARCH MANAGEMENT SCRIPT
# =================================================================
# This script provides comprehensive management capabilities for the
# DevArch development environment including service control, monitoring,
# backup/restore, and maintenance operations.

set -euo pipefail

# =================================================================
# CONFIGURATION AND VARIABLES
# =================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_DIR="$PROJECT_ROOT/compose"
CONFIG_DIR="$PROJECT_ROOT/config"
APPS_DIR="$PROJECT_ROOT/apps"
LOGS_DIR="$PROJECT_ROOT/logs"
BACKUP_DIR="$PROJECT_ROOT/backups"

# Load environment variables
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
else
    echo "❌ Error: .env file not found in $PROJECT_ROOT"
    exit 1
fi

# Script options
VERBOSE=false
FORCE=false
DRY_RUN=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Service groups
CORE_SERVICES=("mariadb" "postgres" "redis" "php-runtime" "mailpit")
PROXY_SERVICES=("nginx-proxy-manager")
DEV_SERVICES=("adminer" "phpmyadmin" "metabase" "redis-commander" "swagger-ui")
OPTIONAL_SERVICES=("mongodb" "mongo-express" "sonarqube" "gitiles")

# =================================================================
# UTILITY FUNCTIONS
# =================================================================
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${PURPLE}[$(date +'%Y-%m-%d %H:%M:%S')] DEBUG: $1${NC}"
    fi
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Progress spinner
show_spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while kill -0 $pid 2>/dev/null; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Execute with dry run support
execute_command() {
    local cmd="$1"
    local description="${2:-}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${CYAN}[DRY RUN] Would execute: $cmd${NC}"
        return 0
    fi
    
    if [[ -n "$description" ]]; then
        info "$description"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        debug "Executing: $cmd"
        eval "$cmd"
    else
        eval "$cmd" >/dev/null 2>&1
    fi
}

# Check if service exists
service_exists() {
    local service="$1"
    local compose_file="$2"
    
    cd "$COMPOSE_DIR"
    podman compose -f "$compose_file" config --services | grep -q "^${service}$"
}

# Get service status
get_service_status() {
    local service="$1"
    
    if podman container exists "$service" 2>/dev/null; then
        local status
        status=$(podman inspect "$service" --format '{{.State.Status}}' 2>/dev/null || echo "unknown")
        case "$status" in
            "running") echo -e "${GREEN}Running${NC}" ;;
            "exited") echo -e "${RED}Stopped${NC}" ;;
            "paused") echo -e "${YELLOW}Paused${NC}" ;;
            "restarting") echo -e "${BLUE}Restarting${NC}" ;;
            *) echo -e "${PURPLE}$status${NC}" ;;
        esac
    else
        echo -e "${RED}Not Created${NC}"
    fi
}

# Get service health
get_service_health() {
    local service="$1"
    
    if podman container exists "$service" 2>/dev/null; then
        local health
        health=$(podman inspect "$service" --format '{{.State.Health.Status}}' 2>/dev/null || echo "no-healthcheck")
        case "$health" in
            "healthy") echo -e "${GREEN}Healthy${NC}" ;;
            "unhealthy") echo -e "${RED}Unhealthy${NC}" ;;
            "starting") echo -e "${YELLOW}Starting${NC}" ;;
            "no-healthcheck") echo -e "${BLUE}No Check${NC}" ;;
            *) echo -e "${PURPLE}$health${NC}" ;;
        esac
    else
        echo -e "${RED}N/A${NC}"
    fi
}

# =================================================================
# SERVICE CONTROL FUNCTIONS
# =================================================================
start_services() {
    local services=("$@")
    
    if [[ ${#services[@]} -eq 0 ]]; then
        log "Starting all DevArch services..."
        start_core_services
        start_proxy_services
        start_dev_services
    else
        log "Starting specified services: ${services[*]}"
        for service in "${services[@]}"; do
            start_single_service "$service"
        done
    fi
}

start_core_services() {
    info "Starting core infrastructure services..."
    cd "$COMPOSE_DIR"
    
    # Start databases first
    execute_command "podman compose -f core.docker-compose.yml up -d mariadb postgres redis" "Starting databases"
    sleep 10  # Allow databases to initialize
    
    # Start runtime and mail
    execute_command "podman compose -f core.docker-compose.yml up -d php-runtime mailpit" "Starting runtime services"
    
    success "Core services started"
}

start_proxy_services() {
    info "Starting proxy services..."
    cd "$COMPOSE_DIR"
    
    execute_command "podman compose -f core.docker-compose.yml up -d nginx-proxy-manager" "Starting Nginx proxy"
    
    success "Proxy services started"
}

start_dev_services() {
    info "Starting development tools..."
    cd "$COMPOSE_DIR"
    
    local profiles=()
    if [[ "${ENABLE_MONGODB:-false}" == "true" ]]; then
        profiles+=(--profile mongodb)
    fi
    if [[ "${ENABLE_QUALITY:-false}" == "true" ]]; then
        profiles+=(--profile quality)
    fi
    
    execute_command "podman compose ${profiles[*]} -f development.docker-compose.yml up -d" "Starting development tools"
    
    success "Development tools started"
}

start_single_service() {
    local service="$1"
    
    info "Starting service: $service"
    
    # Determine which compose file contains the service
    local compose_file=""
    if printf '%s\n' "${CORE_SERVICES[@]}" "${PROXY_SERVICES[@]}" | grep -q "^${service}$"; then
        compose_file="core.docker-compose.yml"
    elif printf '%s\n' "${DEV_SERVICES[@]}" "${OPTIONAL_SERVICES[@]}" | grep -q "^${service}$"; then
        compose_file="development.docker-compose.yml"
    else
        error "Unknown service: $service"
        return 1
    fi
    
    cd "$COMPOSE_DIR"
    
    # Handle optional services with profiles
    local profiles=()
    if [[ "$service" == "mongodb" || "$service" == "mongo-express" ]]; then
        profiles+=(--profile mongodb)
    elif [[ "$service" == "sonarqube" ]]; then
        profiles+=(--profile quality)
    elif [[ "$service" == "gitiles" ]]; then
        profiles+=(--profile docs)
    fi
    
    execute_command "podman compose ${profiles[*]} -f $compose_file up -d $service" "Starting $service"
    success "Service $service started"
}

stop_services() {
    local services=("$@")
    
    if [[ ${#services[@]} -eq 0 ]]; then
        log "Stopping all DevArch services..."
        stop_all_services
    else
        log "Stopping specified services: ${services[*]}"
        for service in "${services[@]}"; do
            stop_single_service "$service"
        done
    fi
}

stop_all_services() {
    info "Stopping all services..."
    cd "$COMPOSE_DIR"
    
    # Stop in reverse order
    execute_command "podman compose --profile mongodb --profile quality --profile docs -f development.docker-compose.yml down" "Stopping development tools"
    execute_command "podman compose -f core.docker-compose.yml down" "Stopping core services"
    
    success "All services stopped"
}

stop_single_service() {
    local service="$1"
    
    info "Stopping service: $service"
    
    if podman container exists "$service" 2>/dev/null; then
        execute_command "podman stop $service" "Stopping $service"
        success "Service $service stopped"
    else
        warn "Service $service is not running"
    fi
}

restart_services() {
    local services=("$@")
    
    if [[ ${#services[@]} -eq 0 ]]; then
        log "Restarting all DevArch services..."
        stop_all_services
        sleep 5
        start_services
    else
        log "Restarting specified services: ${services[*]}"
        for service in "${services[@]}"; do
            restart_single_service "$service"
        done
    fi
}

restart_single_service() {
    local service="$1"
    
    info "Restarting service: $service"
    
    if podman container exists "$service" 2>/dev/null; then
        execute_command "podman restart $service" "Restarting $service"
        success "Service $service restarted"
    else
        warn "Service $service does not exist, starting instead"
        start_single_service "$service"
    fi
}

# =================================================================
# STATUS AND MONITORING
# =================================================================
show_status() {
    local services=("$@")
    
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                    📊 DevArch Service Status                 ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    
    if [[ ${#services[@]} -eq 0 ]]; then
        show_all_status
    else
        show_specific_status "${services[@]}"
    fi
    
    echo
    show_resource_usage
}

show_all_status() {
    # Core services
    echo -e "${YELLOW}🏗️  Core Infrastructure:${NC}"
    printf "%-20s %-15s %-15s %-20s\n" "Service" "Status" "Health" "Uptime"
    echo "────────────────────────────────────────────────────────────────"
    
    for service in "${CORE_SERVICES[@]}"; do
        local status health uptime
        status=$(get_service_status "$service")
        health=$(get_service_health "$service")
        
        if podman container exists "$service" 2>/dev/null; then
            uptime=$(podman inspect "$service" --format '{{.State.StartedAt}}' 2>/dev/null | xargs -I {} date -d {} +"%Y-%m-%d %H:%M" 2>/dev/null || echo "Unknown")
        else
            uptime="N/A"
        fi
        
        printf "%-20s %-15s %-15s %-20s\n" "$service" "$status" "$health" "$uptime"
    done
    
    echo
    
    # Proxy services
    echo -e "${YELLOW}🌐 Proxy Services:${NC}"
    printf "%-20s %-15s %-15s %-20s\n" "Service" "Status" "Health" "Uptime"
    echo "────────────────────────────────────────────────────────────────"
    
    for service in "${PROXY_SERVICES[@]}"; do
        local status health uptime
        status=$(get_service_status "$service")
        health=$(get_service_health "$service")
        
        if podman container exists "$service" 2>/dev/null; then
            uptime=$(podman inspect "$service" --format '{{.State.StartedAt}}' 2>/dev/null | xargs -I {} date -d {} +"%Y-%m-%d %H:%M" 2>/dev/null || echo "Unknown")
        else
            uptime="N/A"
        fi
        
        printf "%-20s %-15s %-15s %-20s\n" "$service" "$status" "$health" "$uptime"
    done
    
    echo
    
    # Development services
    echo -e "${YELLOW}🔧 Development Tools:${NC}"
    printf "%-20s %-15s %-15s %-20s\n" "Service" "Status" "Health" "Uptime"
    echo "────────────────────────────────────────────────────────────────"
    
    for service in "${DEV_SERVICES[@]}"; do
        local status health uptime
        status=$(get_service_status "$service")
        health=$(get_service_health "$service")
        
        if podman container exists "$service" 2>/dev/null; then
            uptime=$(podman inspect "$service" --format '{{.State.StartedAt}}' 2>/dev/null | xargs -I {} date -d {} +"%Y-%m-%d %H:%M" 2>/dev/null || echo "Unknown")
        else
            uptime="N/A"
        fi
        
        printf "%-20s %-15s %-15s %-20s\n" "$service" "$status" "$health" "$uptime"
    done
    
    # Optional services (only show if they exist)
    local optional_running=()
    for service in "${OPTIONAL_SERVICES[@]}"; do
        if podman container exists "$service" 2>/dev/null; then
            optional_running+=("$service")
        fi
    done
    
    if [[ ${#optional_running[@]} -gt 0 ]]; then
        echo
        echo -e "${YELLOW}⚡ Optional Services:${NC}"
        printf "%-20s %-15s %-15s %-20s\n" "Service" "Status" "Health" "Uptime"
        echo "────────────────────────────────────────────────────────────────"
        
        for service in "${optional_running[@]}"; do
            local status health uptime
            status=$(get_service_status "$service")
            health=$(get_service_health "$service")
            uptime=$(podman inspect "$service" --format '{{.State.StartedAt}}' 2>/dev/null | xargs -I {} date -d {} +"%Y-%m-%d %H:%M" 2>/dev/null || echo "Unknown")
            
            printf "%-20s %-15s %-15s %-20s\n" "$service" "$status" "$health" "$uptime"
        done
    fi
}

show_specific_status() {
    local services=("$@")
    
    printf "%-20s %-15s %-15s %-20s\n" "Service" "Status" "Health" "Uptime"
    echo "────────────────────────────────────────────────────────────────"
    
    for service in "${services[@]}"; do
        local status health uptime
        status=$(get_service_status "$service")
        health=$(get_service_health "$service")
        
        if podman container exists "$service" 2>/dev/null; then
            uptime=$(podman inspect "$service" --format '{{.State.StartedAt}}' 2>/dev/null | xargs -I {} date -d {} +"%Y-%m-%d %H:%M" 2>/dev/null || echo "Unknown")
        else
            uptime="N/A"
        fi
        
        printf "%-20s %-15s %-15s %-20s\n" "$service" "$status" "$health" "$uptime"
    done
}

show_resource_usage() {
    echo -e "${YELLOW}💻 Resource Usage:${NC}"
    
    # System resources
    local total_memory total_cpu
    total_memory=$(free -h | awk '/^Mem:/ {print $2}')
    total_cpu=$(nproc)
    
    echo "System: ${total_memory} RAM, ${total_cpu} CPU cores"
    
    # Container resources
    if command -v podman >/dev/null 2>&1; then
        local running_containers
        running_containers=$(podman ps --filter "label=com.docker.compose.project=${COMPOSE_PROJECT_NAME}" --format "table {{.Names}}" | tail -n +2 | wc -l)
        echo "Running containers: $running_containers"
        
        # Network usage
        if podman network exists "$NETWORK_NAME" 2>/dev/null; then
            echo -e "Network: ${GREEN}$NETWORK_NAME (Active)${NC}"
        else
            echo -e "Network: ${RED}$NETWORK_NAME (Not Found)${NC}"
        fi
    fi
}

# =================================================================
# LOG MANAGEMENT
# =================================================================
show_logs() {
    local service="${1:-}"
    local lines="${2:-50}"
    local follow="${3:-false}"
    
    if [[ -z "$service" ]]; then
        echo "Available services for log viewing:"
        echo
        echo "Core services: ${CORE_SERVICES[*]}"
        echo "Proxy services: ${PROXY_SERVICES[*]}"
        echo "Development services: ${DEV_SERVICES[*]}"
        echo
        echo "Usage: $0 logs <service> [lines] [follow]"
        echo "Example: $0 logs nginx-proxy-manager 100 true"
        return 0
    fi
    
    if ! podman container exists "$service" 2>/dev/null; then
        error "Service '$service' does not exist"
        return 1
    fi
    
    info "Showing logs for $service (last $lines lines)"
    
    local follow_flag=""
    if [[ "$follow" == "true" || "$follow" == "1" ]]; then
        follow_flag="--follow"
        info "Following logs (Press Ctrl+C to stop)"
    fi
    
    podman logs $follow_flag --tail "$lines" "$service"
}

clear_logs() {
    local service="${1:-}"
    
    if [[ -z "$service" ]]; then
        if [[ "$FORCE" != "true" ]]; then
            read -p "Are you sure you want to clear ALL service logs? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "Operation cancelled"
                return 0
            fi
        fi
        
        log "Clearing all service logs..."
        
        # Clear container logs
        for container in $(podman ps -a --filter "label=com.docker.compose.project=${COMPOSE_PROJECT_NAME}" --format "{{.Names}}"); do
            if [[ "$DRY_RUN" != "true" ]]; then
                podman logs "$container" 2>/dev/null | head -1 >/dev/null || true
                > "$(podman inspect "$container" --format '{{.LogPath}}')" 2>/dev/null || true
            fi
            debug "Cleared logs for $container"
        done
        
        # Clear local log files
        if [[ -d "$LOGS_DIR" ]]; then
            execute_command "find $LOGS_DIR -name '*.log' -type f -exec truncate -s 0 {} +" "Clearing local log files"
        fi
        
        success "All logs cleared"
    else
        if ! podman container exists "$service" 2>/dev/null; then
            error "Service '$service' does not exist"
            return 1
        fi
        
        info "Clearing logs for $service"
        
        if [[ "$DRY_RUN" != "true" ]]; then
            > "$(podman inspect "$service" --format '{{.LogPath}}')" 2>/dev/null || true
        fi
        
        success "Logs cleared for $service"
    fi
}

# =================================================================
# BACKUP AND RESTORE
# =================================================================
backup_data() {
    local backup_name="${1:-devarch-backup-$(date +%Y%m%d-%H%M%S)}"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    log "Creating backup: $backup_name"
    
    # Create backup directory
    mkdir -p "$backup_path"
    
    # Backup environment file
    execute_command "cp $PROJECT_ROOT/.env $backup_path/" "Backing up environment configuration"
    
    # Backup applications
    if [[ -d "$APPS_DIR" ]]; then
        execute_command "tar -czf $backup_path/apps.tar.gz -C $PROJECT_ROOT apps" "Backing up applications"
    fi
    
    # Backup configurations
    if [[ -d "$CONFIG_DIR" ]]; then
        execute_command "tar -czf $backup_path/config.tar.gz -C $PROJECT_ROOT config" "Backing up configurations"
    fi
    
    # Backup database volumes
    info "Backing up database volumes..."
    
    # MariaDB backup
    if podman container exists "mariadb" && [[ "$(get_service_status mariadb)" == *"Running"* ]]; then
        execute_command "podman exec mariadb mysqldump -u root -p\${MYSQL_ROOT_PASSWORD} --all-databases > $backup_path/mariadb-dump.sql" "Backing up MariaDB"
    fi
    
    # PostgreSQL backup
    if podman container exists "postgres" && [[ "$(get_service_status postgres)" == *"Running"* ]]; then
        execute_command "podman exec postgres pg_dumpall -U \${POSTGRES_USER} > $backup_path/postgres-dump.sql" "Backing up PostgreSQL"
    fi
    
    # Create backup metadata
    cat > "$backup_path/backup-info.json" << EOF
{
    "backup_name": "$backup_name",
    "created_at": "$(date -Iseconds)",
    "devarch_version": "1.0.0",
    "services": $(podman ps --filter "label=com.docker.compose.project=${COMPOSE_PROJECT_NAME}" --format "{{.Names}}" | jq -R -s -c 'split("\n")[:-1]')
}
EOF
    
    success "Backup created: $backup_path"
    
    # Show backup size
    local backup_size
    backup_size=$(du -sh "$backup_path" | cut -f1)
    info "Backup size: $backup_size"
}

restore_data() {
    local backup_name="$1"
    
    if [[ -z "$backup_name" ]]; then
        echo "Available backups:"
        if [[ -d "$BACKUP_DIR" ]]; then
            ls -la "$BACKUP_DIR"
        else
            echo "No backups found"
        fi
        echo
        echo "Usage: $0 restore <backup-name>"
        return 0
    fi
    
    local backup_path="$BACKUP_DIR/$backup_name"
    
    if [[ ! -d "$backup_path" ]]; then
        error "Backup not found: $backup_path"
        return 1
    fi
    
    if [[ "$FORCE" != "true" ]]; then
        read -p "Are you sure you want to restore from backup '$backup_name'? This will overwrite current data. (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Operation cancelled"
            return 0
        fi
    fi
    
    log "Restoring from backup: $backup_name"
    
    # Stop services
    info "Stopping services for restore..."
    stop_all_services
    
    # Restore applications
    if [[ -f "$backup_path/apps.tar.gz" ]]; then
        execute_command "tar -xzf $backup_path/apps.tar.gz -C $PROJECT_ROOT" "Restoring applications"
    fi
    
    # Restore configurations
    if [[ -f "$backup_path/config.tar.gz" ]]; then
        execute_command "tar -xzf $backup_path/config.tar.gz -C $PROJECT_ROOT" "Restoring configurations"
    fi
    
    # Start services
    info "Starting services..."
    start_services
    
    # Wait for databases to be ready
    sleep 30
    
    # Restore databases
    if [[ -f "$backup_path/mariadb-dump.sql" ]]; then
        execute_command "podman exec -i mariadb mysql -u root -p\${MYSQL_ROOT_PASSWORD} < $backup_path/mariadb-dump.sql" "Restoring MariaDB"
    fi
    
    if [[ -f "$backup_path/postgres-dump.sql" ]]; then
        execute_command "podman exec -i postgres psql -U \${POSTGRES_USER} < $backup_path/postgres-dump.sql" "Restoring PostgreSQL"
    fi
    
    success "Restore completed from backup: $backup_name"
}

# =================================================================
# MAINTENANCE OPERATIONS
# =================================================================
cleanup_system() {
    log "Performing system cleanup..."
    
    if [[ "$FORCE" != "true" ]]; then
        read -p "This will remove unused containers, images, and volumes. Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Operation cancelled"
            return 0
        fi
    fi
    
    # Remove stopped containers
    execute_command "podman container prune -f" "Removing stopped containers"
    
    # Remove unused images
    execute_command "podman image prune -f" "Removing unused images"
    
    # Remove unused volumes
    execute_command "podman volume prune -f" "Removing unused volumes"
    
    # Remove unused networks
    execute_command "podman network prune -f" "Removing unused networks"
    
    # Clean up log files older than 30 days
    if [[ -d "$LOGS_DIR" ]]; then
        execute_command "find $LOGS_DIR -name '*.log' -type f -mtime +30 -delete" "Cleaning old log files"
    fi
    
    # Clean up old backups (keep last 10)
    if [[ -d "$BACKUP_DIR" ]]; then
        local backup_count
        backup_count=$(ls -1 "$BACKUP_DIR" | wc -l)
        if [[ $backup_count -gt 10 ]]; then
            execute_command "ls -1t $BACKUP_DIR | tail -n +11 | xargs -I {} rm -rf $BACKUP_DIR/{}" "Cleaning old backups"
        fi
    fi
    
    success "System cleanup completed"
}

update_system() {
    log "Updating DevArch system..."
    
    cd "$COMPOSE_DIR"
    
    # Pull latest images
    info "Pulling latest container images..."
    execute_command "podman compose -f core.docker-compose.yml pull" "Updating core images"
    execute_command "podman compose -f development.docker-compose.yml pull" "Updating development images"
    
    # Rebuild custom images
    info "Rebuilding custom images..."
    execute_command "podman compose -f core.docker-compose.yml build --no-cache php-runtime" "Rebuilding PHP runtime"
    execute_command "podman compose -f core.docker-compose.yml build --no-cache nginx-proxy-manager" "Rebuilding Nginx proxy"
    
    success "System update completed"
    warn "Restart services to use updated images: $0 restart"
}

# =================================================================
# USAGE AND HELP
# =================================================================
show_usage() {
    echo -e "${CYAN}DevArch Management Script${NC}"
    echo
    echo "Usage: $0 <command> [options] [arguments]"
    echo
    echo -e "${YELLOW}Service Control:${NC}"
    echo "  start [service...]     Start all services or specific services"
    echo "  stop [service...]      Stop all services or specific services"
    echo "  restart [service...]   Restart all services or specific services"
    echo "  status [service...]    Show service status"
    echo
    echo -e "${YELLOW}Logs:${NC}"
    echo "  logs <service> [lines] [follow]  Show service logs"
    echo "  clear-logs [service]             Clear logs for service or all"
    echo
    echo -e "${YELLOW}Backup & Restore:${NC}"
    echo "  backup [name]          Create system backup"
    echo "  restore <name>         Restore from backup"
    echo "  list-backups           List available backups"
    echo
    echo -e "${YELLOW}Maintenance:${NC}"
    echo "  cleanup                Clean unused containers, images, volumes"
    echo "  update                 Update container images"
    echo "  reset                  Reset entire environment"
    echo
    echo -e "${YELLOW}Options:${NC}"
    echo "  -v, --verbose          Enable verbose output"
    echo "  -f, --force            Skip confirmation prompts"
    echo "  -n, --dry-run          Show what would be done without executing"
    echo "  -h, --help             Show this help message"
    echo
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 start               # Start all services"
    echo "  $0 restart nginx-proxy-manager  # Restart specific service"
    echo "  $0 logs mariadb 100 true        # Follow last 100 lines of MariaDB logs"
    echo "  $0 backup my-backup    # Create named backup"
    echo "  $0 status              # Show all service status"
}

# =================================================================
# COMMAND PARSING
# =================================================================
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -n|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
}

reset_environment() {
    if [[ "$FORCE" != "true" ]]; then
        echo -e "${RED}⚠️  WARNING: This will completely reset the DevArch environment!${NC}"
        echo "This action will:"
        echo "  • Stop and remove all containers"
        echo "  • Remove all volumes (databases will be lost!)"
        echo "  • Remove the network"
        echo "  • Clear all logs"
        echo
        read -p "Are you absolutely sure? Type 'RESET' to confirm: " -r
        if [[ $REPLY != "RESET" ]]; then
            info "Operation cancelled"
            return 0
        fi
    fi
    
    log "Resetting DevArch environment..."
    
    # Stop all services
    stop_all_services
    
    # Remove all containers
    execute_command "podman container prune -f" "Removing all containers"
    
    # Remove all volumes
    execute_command "podman volume prune -f" "Removing all volumes"
    
    # Remove network
    if podman network exists "$NETWORK_NAME" 2>/dev/null; then
        execute_command "podman network rm $NETWORK_NAME" "Removing network"
    fi
    
    # Clear logs
    if [[ -d "$LOGS_DIR" ]]; then
        execute_command "rm -rf $LOGS_DIR/*" "Clearing logs"
    fi
    
    success "Environment reset completed"
    info "Run './scripts/install.sh' to reinstall the environment"
}

list_backups() {
    echo -e "${CYAN}📦 Available Backups:${NC}"
    echo
    
    if [[ ! -d "$BACKUP_DIR" || -z "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]]; then
        echo "No backups found"
        return 0
    fi
    
    printf "%-30s %-20s %-10s\n" "Backup Name" "Created" "Size"
    echo "────────────────────────────────────────────────────────────────"
    
    for backup in "$BACKUP_DIR"/*; do
        if [[ -d "$backup" ]]; then
            local name date size
            name=$(basename "$backup")
            
            if [[ -f "$backup/backup-info.json" ]]; then
                date=$(jq -r '.created_at' "$backup/backup-info.json" 2>/dev/null | cut -d'T' -f1 || echo "Unknown")
            else
                date=$(stat -c %y "$backup" 2>/dev/null | cut -d' ' -f1 || echo "Unknown")
            fi
            
            size=$(du -sh "$backup" 2>/dev/null | cut -f1 || echo "Unknown")
            
            printf "%-30s %-20s %-10s\n" "$name" "$date" "$size"
        fi
    done
}

# =================================================================
# MAIN FUNCTION
# =================================================================
main() {
    # Parse global options first
    parse_arguments "$@"
    
    # Remove parsed options from arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose|-f|--force|-n|--dry-run|-h|--help)
                shift
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Ensure we have a command
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi
    
    local command="$1"
    shift
    
    # Ensure we're in the project directory
    cd "$PROJECT_ROOT"
    
    # Execute command
    case "$command" in
        start)
            start_services "$@"
            ;;
        stop)
            stop_services "$@"
            ;;
        restart)
            restart_services "$@"
            ;;
        status)
            show_status "$@"
            ;;
        logs)
            show_logs "$@"
            ;;
        clear-logs)
            clear_logs "$@"
            ;;
        backup)
            backup_data "$@"
            ;;
        restore)
            restore_data "$@"
            ;;
        list-backups)
            list_backups
            ;;
        cleanup)
            cleanup_system
            ;;
        update)
            update_system
            ;;
        reset)
            reset_environment
            ;;
        *)
            error "Unknown command: $command"
            echo
            show_usage
            exit 1
            ;;
    esac
}

# =================================================================
# ERROR HANDLING
# =================================================================
handle_error() {
    local exit_code=$?
    error "Command failed with exit code $exit_code"
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo "Stack trace:"
        local frame=0
        while caller $frame; do
            ((frame++))
        done
    fi
    
    exit $exit_code
}

# Set up error handling
trap handle_error ERR

# =================================================================
# SCRIPT EXECUTION
# =================================================================
# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi