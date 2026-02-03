#!/bin/zsh

# =============================================================================
# DEVARCH CONFIGURATION - RUNTIME & ENVIRONMENT
# =============================================================================
# Provides: runtime detection, network config, env vars, command setup.
# Service discovery and categories are now managed by the database/API.

# =============================================================================
# BACKEND SERVICE PORT ALLOCATION STRATEGY
# =============================================================================
# PHP:    8100-8199 (main: 8100, vite: 8102)
# Node:   8200-8299 (main: 8200, secondary: 8201, vite: 8202, graphql: 8203, debug: 9229)
# Python: 8300-8399 (main: 8300, flask: 8301, jupyter: 8302, flower: 8303)
# Go:     8400-8499 (main: 8400, metrics: 8401, debug: 8402, pprof: 8403)
# .NET:   8600-8699 (main: 8600, secondary: 8601, debug: 8602, hot-reload: 8603)
# Rust:   8700-8799 (main: 8700, secondary: 8701, debug: 8702, metrics: 8703)

# =============================================================================
# PATHS
# =============================================================================

SCRIPT_SOURCE="${(%):-%x}"
export PROJECT_ROOT=$(cd "$(dirname "$SCRIPT_SOURCE")/../" && pwd)
export SCRIPT_DIR="${PROJECT_ROOT}/scripts"
export CONFIG_DIR="${PROJECT_ROOT}/config"
export APPS_DIR="${PROJECT_ROOT}/apps"
export LOGS_DIR="${PROJECT_ROOT}/logs"

# =============================================================================
# NETWORK & CONTAINER CONFIGURATION
# =============================================================================

export NETWORK_NAME="microservices-net"
export CONTAINER_RUNTIME="podman"

# =============================================================================
# OUTPUT HELPERS
# =============================================================================

VERBOSE=${VERBOSE:-0}
QUIET=${QUIET:-0}

print_status() {
    local level="$1" message="$2"
    case "$level" in
        info)    [[ $VERBOSE -eq 1 ]] && printf "\033[36m--\033[0m %s\n" "$message" ;;
        success) [[ $QUIET -eq 0 ]]   && printf "\033[32mOK\033[0m %s\n" "$message" ;;
        warning) printf "\033[33mWARN\033[0m %s\n" "$message" ;;
        error)   printf "\033[31mERR\033[0m %s\n" "$message" >&2 ;;
        *)       echo "$message" ;;
    esac
}

handle_error() {
    printf "\033[31mERR\033[0m %s\n" "$1" >&2
    exit "${2:-1}"
}

# =============================================================================
# ENVIRONMENT VARIABLES
# =============================================================================

if [[ -f "$PROJECT_ROOT/.env" ]]; then
    set -a; source "$PROJECT_ROOT/.env"; set +a
fi

export MARIADB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-123456}"
export MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-123456}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-123456}"
export MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-123456}"
export ADMIN_PASSWORD="${ADMIN_PASSWORD:-123456}"
export ADMIN_USER="${ADMIN_USER:-admin}"
export ADMIN_EMAIL="${ADMIN_EMAIL:-admin@test.local}"

# =============================================================================
# RUNTIME DETECTION & COMMAND SETUP
# =============================================================================

detect_sudo_requirement() {
    if command -v podman >/dev/null 2>&1; then
        export USE_PODMAN=true
        export DEFAULT_SUDO=false
    elif command -v docker >/dev/null 2>&1; then
        export USE_PODMAN=false
        if groups | grep -q docker 2>/dev/null; then
            export DEFAULT_SUDO=false
        else
            export DEFAULT_SUDO=true
        fi
    else
        echo "Error: Neither podman nor docker found!"
        exit 1
    fi
}

setup_command_context() {
    local want_sudo="${1:-$DEFAULT_SUDO}"
    local show_errors="${2:-false}"

    [[ "$show_errors" == "true" ]] && export ERROR_REDIRECT="" || export ERROR_REDIRECT="2>/dev/null"

    if [[ "$USE_PODMAN" == "true" ]]; then
        local rootless_has_net="" rootful_has_net=""
        podman network exists "$NETWORK_NAME" >/dev/null 2>&1 && rootless_has_net="yes"
        sudo -n podman network exists "$NETWORK_NAME" >/dev/null 2>&1 && rootful_has_net="yes"

        if [[ -n "$rootless_has_net" && -z "$rootful_has_net" ]]; then
            export SUDO_PREFIX=""
        elif [[ -z "$rootless_has_net" && -n "$rootful_has_net" ]]; then
            export SUDO_PREFIX="sudo "
        else
            [[ "$want_sudo" == "true" ]] && export SUDO_PREFIX="sudo " || export SUDO_PREFIX=""
        fi

        export CONTAINER_CMD="${SUDO_PREFIX}podman"
        if ${SUDO_PREFIX}podman compose version >/dev/null 2>&1; then
            export COMPOSE_CMD="${SUDO_PREFIX}podman compose"
        else
            export COMPOSE_CMD="${SUDO_PREFIX}podman-compose"
        fi

        export PODMAN_COMPOSE_WARNING_LOGS="false"
        export COMPOSE_IGNORE_ORPHANS="true"
        export PODMAN_USERNS="keep-id"
    else
        [[ "$want_sudo" == "true" ]] && export SUDO_PREFIX="sudo " || export SUDO_PREFIX=""
        export CONTAINER_CMD="${SUDO_PREFIX}docker"
        export COMPOSE_CMD="${SUDO_PREFIX}docker compose"
    fi
}

ensure_network_exists() {
    if [[ "$USE_PODMAN" == "true" ]]; then
        eval "$CONTAINER_CMD network exists \"$NETWORK_NAME\" $ERROR_REDIRECT" && return 0
    else
        eval "$CONTAINER_CMD network ls --format '{{.Name}}'" 2>/dev/null | grep -q "^$NETWORK_NAME$" && return 0
    fi
    eval "$CONTAINER_CMD network create --driver bridge \"$NETWORK_NAME\" $ERROR_REDIRECT" || { print_status "error" "network create failed"; return 1; }
}

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        grep -q Microsoft /proc/version 2>/dev/null && export OS_TYPE="wsl" || export OS_TYPE="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        export OS_TYPE="macos"
    else
        export OS_TYPE="unknown"
    fi
}

# =============================================================================
# INITIALIZATION
# =============================================================================

detect_sudo_requirement
detect_os
setup_command_context "$DEFAULT_SUDO" "false"
