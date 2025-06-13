#!/bin/zsh
# Get script location portably
PROJECT_ROOT=$(cd "$(dirname "$0")/../" && pwd)
SCRIPT_DIR="${PROJECT_ROOT}/scripts"
COMPOSE_DIR="${PROJECT_ROOT}/compose"
NETWORK_NAME="microservices-net"

MARIADB_ROOT_PASSWORD="123456"
POSTGRES_PASSWORD="123456"

# Export variables for use in other scripts
export SCRIPT_DIR PROJECT_ROOT COMPOSE_DIR NETWORK_NAME MARIADB_ROOT_PASSWORD POSTGRES_PASSWORD
