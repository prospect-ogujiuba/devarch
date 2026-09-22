#!/usr/bin/env bash
set -Eeuo pipefail

script_dir="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)"
services=(redmine openproject plane glpi leantime vikunja)

usage() {
  cat <<'EOF'
Usage: scripts/project-management/manage.sh <command>

Commands:
  up       Create the shared network and start all six stacks
  down     Stop and remove all six stacks (persistent volumes are retained)
  status   Show the target containers and their state
  urls     Print local URLs and test logins
  help     Show this help
EOF
}

compose() {
  local service="$1"
  shift
  (
    cd "$repo_root/services-library/project/$service"
    podman compose "$@"
  )
}

print_urls() {
  cat <<'EOF'
Local evaluation services:
  Redmine:     https://redmine.test
  OpenProject: https://openproject.test  (login and password in services-library/project/openproject/.env)
  Plane:       https://plane.test
  GLPI:        https://glpi.test         (login glpi; password in services-library/project/glpi/.env)
  Leantime:    https://leantime.test
  Vikunja:     https://vikunja.test
EOF
}

case "${1:-help}" in
  up)
    if ! podman network exists microservices-net; then
      podman network create microservices-net >/dev/null
    fi
    for service in "${services[@]}"; do
      printf 'Starting %s...\n' "$service"
      if ! compose "$service" up -d; then
        printf 'Initial %s start was incomplete; retrying once...\n' "$service" >&2
        sleep 2
        compose "$service" up -d
      fi
    done
    printf '\n'
    print_urls
    ;;
  down)
    for ((i=${#services[@]}-1; i>=0; i--)); do
      service="${services[$i]}"
      printf 'Stopping %s...\n' "$service"
      compose "$service" down
    done
    ;;
  status)
    podman ps -a --filter name=redmine --filter name=openproject --filter name=plane --filter name=glpi --filter name=leantime --filter name=vikunja \
      --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
    ;;
  urls)
    print_urls
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    printf 'Unknown command: %s\n\n' "$1" >&2
    usage >&2
    exit 2
    ;;
esac
