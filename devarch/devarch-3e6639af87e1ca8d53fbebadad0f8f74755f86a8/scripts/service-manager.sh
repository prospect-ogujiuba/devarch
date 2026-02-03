#!/bin/zsh

# =============================================================================
# DEVARCH SERVICE MANAGER - API CLIENT
# =============================================================================
# Thin wrapper around the DevArch API. Requires API to be running.

SCRIPT_NAME="${0##*/}"
API_BASE="http://localhost:8550/api/v1"

# Source config for runtime detection (container commands, network)
. "$(dirname "$0")/config.sh"

# =============================================================================
# OUTPUT HELPERS
# =============================================================================

ok()   { printf "\033[32mOK\033[0m %s\n" "$1"; }
err()  { printf "\033[31mERR\033[0m %s\n" "$1" >&2; }
warn() { printf "\033[33mWARN\033[0m %s\n" "$1"; }

# =============================================================================
# API HELPERS
# =============================================================================

api_get() {
    local response
    response=$(curl -sf "$API_BASE$1" 2>/dev/null)
    if [[ $? -ne 0 ]]; then
        err "API unreachable at $API_BASE — is devarch-api running?"
        exit 1
    fi
    echo "$response"
}

api_post() {
    local response
    response=$(curl -sf -X POST "$API_BASE$1" 2>/dev/null)
    if [[ $? -ne 0 ]]; then
        err "API request failed: POST $1"
        return 1
    fi
    echo "$response"
}

api_post_json() {
    local response
    response=$(curl -sf -X POST -H "Content-Type: application/json" -d "$2" "$API_BASE$1" 2>/dev/null)
    if [[ $? -ne 0 ]]; then
        err "API request failed: POST $1"
        return 1
    fi
    echo "$response"
}

# =============================================================================
# USAGE
# =============================================================================

show_usage() {
    cat << 'EOF'
Usage: service-manager.sh COMMAND [TARGET] [OPTIONS]

COMMANDS:
    up SERVICE              Start a service
    down SERVICE            Stop a service
    restart SERVICE         Restart a service
    rebuild SERVICE         Rebuild and restart (--no-cache)
    logs SERVICE            Show logs (--follow, --tail N)
    status [SERVICE]        Show status
    list                    List all services
    start CATEGORY...       Start category services
    stop CATEGORY...        Stop category services
    compose SERVICE         Show generated compose YAML
    check                   Verify runtime prerequisites

OPTIONS:
    --no-cache              Rebuild without cache
    --follow, -f            Follow logs
    --tail N                Log lines (default: 100)
    -h, --help              Show this help
EOF
}

# =============================================================================
# COMMANDS
# =============================================================================

cmd_up() {
    local name="$1"
    [[ -z "$name" ]] && { err "up requires service name"; exit 1; }
    local result
    result=$(api_post "/services/$name/start")
    if [[ $? -eq 0 ]]; then
        ok "$name started"
    else
        err "$name start failed"
        return 1
    fi
}

cmd_down() {
    local name="$1"
    [[ -z "$name" ]] && { err "down requires service name"; exit 1; }
    local result
    result=$(api_post "/services/$name/stop")
    if [[ $? -eq 0 ]]; then
        ok "$name stopped"
    else
        err "$name stop failed"
        return 1
    fi
}

cmd_restart() {
    local name="$1"
    [[ -z "$name" ]] && { err "restart requires service name"; exit 1; }
    local result
    result=$(api_post "/services/$name/restart")
    if [[ $? -eq 0 ]]; then
        ok "$name restarted"
    else
        err "$name restart failed"
        return 1
    fi
}

cmd_rebuild() {
    local name="$1"
    local no_cache="$2"
    [[ -z "$name" ]] && { err "rebuild requires service name"; exit 1; }
    local url="/services/$name/rebuild"
    [[ "$no_cache" == "true" ]] && url="${url}?no_cache=true"
    local result
    result=$(api_post "$url")
    if [[ $? -eq 0 ]]; then
        ok "$name rebuilt"
    else
        err "$name rebuild failed"
        return 1
    fi
}

cmd_logs() {
    local name="$1"
    local tail="${2:-100}"
    local follow="$3"
    [[ -z "$name" ]] && { err "logs requires service name"; exit 1; }

    if [[ "$follow" == "true" ]]; then
        # Follow via container runtime directly
        eval "$COMPOSE_CMD logs -f --tail $tail $name" 2>/dev/null || \
            eval "$CONTAINER_CMD logs -f --tail $tail $name" 2>/dev/null || \
            err "cannot follow logs for $name"
    else
        curl -sf "$API_BASE/services/$name/logs?tail=$tail" 2>/dev/null || err "failed to get logs for $name"
    fi
}

cmd_status() {
    local name="$1"
    if [[ -n "$name" ]]; then
        local data
        data=$(api_get "/services/$name/status")
        [[ $? -ne 0 ]] && return 1
        local status
        status=$(echo "$data" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status','unknown'))" 2>/dev/null || echo "$data")
        printf "%-20s %s\n" "$name" "$status"
    else
        printf "\033[1m%-20s %-12s %s\033[0m\n" "SERVICE" "STATUS" "CATEGORY"
        local services
        services=$(api_get "/services?include=status&limit=500")
        [[ $? -ne 0 ]] && return 1
        echo "$services" | python3 -c "
import sys, json
services = json.load(sys.stdin)
for s in services:
    name = s.get('name', '')
    cat = s.get('category', {})
    cat_name = cat.get('name', '') if isinstance(cat, dict) else ''
    status_obj = s.get('status')
    status = status_obj.get('status', 'stopped') if status_obj else 'stopped'
    color = '\033[32m' if status == 'running' else '\033[31m' if status in ('stopped', 'exited') else '\033[33m'
    print(f'{color}{name:<20} {status:<12} {cat_name}\033[0m')
" 2>/dev/null
    fi
}

cmd_list() {
    local services
    services=$(api_get "/services?limit=500")
    [[ $? -ne 0 ]] && return 1
    printf "\033[1m%-20s %s\033[0m\n" "SERVICE" "CATEGORY"
    echo "$services" | python3 -c "
import sys, json
services = json.load(sys.stdin)
for s in sorted(services, key=lambda x: (x.get('category',{}).get('name',''), x.get('name',''))):
    cat = s.get('category', {})
    cat_name = cat.get('name', '') if isinstance(cat, dict) else ''
    print(f'{s[\"name\"]:<20} {cat_name}')
" 2>/dev/null
}

cmd_start_category() {
    local targets=("$@")
    [[ ${#targets[@]} -eq 0 ]] && { err "start requires category name(s)"; exit 1; }
    for cat in "${targets[@]}"; do
        local result
        result=$(api_post "/categories/$cat/start")
        if [[ $? -eq 0 ]]; then
            ok "$cat started"
        else
            err "$cat start failed"
        fi
    done
}

cmd_stop_category() {
    local targets=("$@")
    [[ ${#targets[@]} -eq 0 ]] && { err "stop requires category name(s)"; exit 1; }
    for cat in "${targets[@]}"; do
        local result
        result=$(api_post "/categories/$cat/stop")
        if [[ $? -eq 0 ]]; then
            ok "$cat stopped"
        else
            err "$cat stop failed"
        fi
    done
}

cmd_compose() {
    local name="$1"
    [[ -z "$name" ]] && { err "compose requires service name"; exit 1; }
    curl -sf "$API_BASE/services/$name/compose" 2>/dev/null || err "failed to get compose for $name"
}

cmd_check() {
    local exit_code=0
    local has_runtime=false

    command -v podman &>/dev/null && { printf "\033[32m+\033[0m podman %s\n" "$(podman --version 2>/dev/null | cut -d' ' -f3)"; has_runtime=true; }
    command -v docker &>/dev/null && { printf "\033[32m+\033[0m docker %s\n" "$(docker --version 2>/dev/null | cut -d' ' -f3 | tr -d ',')"; has_runtime=true; }
    [[ "$has_runtime" == "false" ]] && { err "No container runtime"; exit_code=1; }

    # Check API
    curl -sf "$API_BASE/services?limit=1" >/dev/null 2>&1 && \
        printf "\033[32m+\033[0m api at %s\n" "$API_BASE" || \
        { printf "\033[31m!\033[0m api unreachable at %s\n" "$API_BASE"; exit_code=1; }

    # Network
    eval "$CONTAINER_CMD network exists $NETWORK_NAME" 2>/dev/null && \
        printf "\033[32m+\033[0m network %s\n" "$NETWORK_NAME" || \
        printf "\033[33m-\033[0m network %s (will create)\n" "$NETWORK_NAME"

    [[ -f "$PROJECT_ROOT/.env" ]] && printf "\033[32m+\033[0m .env\n" || printf "\033[33m-\033[0m .env (optional)\n"

    return $exit_code
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    [[ $# -eq 0 || "$1" == "-h" || "$1" == "--help" ]] && { show_usage; exit 0; }

    local cmd="$1"; shift
    local service_name=""
    local opt_no_cache=false
    local opt_follow=false
    local opt_tail=100
    local -a targets

    # Extract service name or targets
    case "$cmd" in
        up|down|restart|rebuild|logs|compose)
            [[ -n "$1" && "$1" != -* ]] && { service_name="$1"; shift; }
            ;;
        start|stop)
            while [[ -n "$1" && "$1" != -* ]]; do
                targets+=("$1"); shift
            done
            ;;
    esac

    # Parse options
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --no-cache) opt_no_cache=true; shift ;;
            --follow|-f) opt_follow=true; shift ;;
            --tail) opt_tail="$2"; shift 2 ;;
            -h|--help) show_usage; exit 0 ;;
            *) err "Unknown option: $1"; exit 1 ;;
        esac
    done

    case "$cmd" in
        up)       cmd_up "$service_name" ;;
        down)     cmd_down "$service_name" ;;
        restart)  cmd_restart "$service_name" ;;
        rebuild)  cmd_rebuild "$service_name" "$opt_no_cache" ;;
        logs)     cmd_logs "$service_name" "$opt_tail" "$opt_follow" ;;
        status)   cmd_status "$service_name" ;;
        list)     cmd_list ;;
        start)    cmd_start_category "${targets[@]}" ;;
        stop)     cmd_stop_category "${targets[@]}" ;;
        compose)  cmd_compose "$service_name" ;;
        check)    cmd_check ;;
        *)        err "Unknown command: $cmd"; show_usage; exit 1 ;;
    esac
}

if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi
