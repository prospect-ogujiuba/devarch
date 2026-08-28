#!/bin/sh
set -eu

container_user="${DEVARCH_NODE_CONTAINER_USER:-0:0}"
package_manager="${DEVARCH_NODE_PACKAGE_MANAGER:-npm}"
package_script="${DEVARCH_NODE_SCRIPT:-devarch}"

case "$container_user" in
  *[!0-9:]*|:*|*:) printf 'invalid DEVARCH_NODE_CONTAINER_USER\n' >&2; exit 2 ;;
esac
case "$package_manager" in
  npm|pnpm|yarn) ;;
  *) printf 'unsupported package manager: %s\n' "$package_manager" >&2; exit 2 ;;
esac
case "$package_script" in
  *[!A-Za-z0-9:_-]*|'') printf 'invalid package script\n' >&2; exit 2 ;;
esac

mkdir -p /app/node_modules /tmp/devarch-node-home
chown "$container_user" /app/node_modules /tmp/devarch-node-home
export HOME=/tmp/devarch-node-home

case "$package_manager" in
  npm)
    su-exec "$container_user" npm install
    exec su-exec "$container_user" npm run "$package_script"
    ;;
  pnpm)
    su-exec "$container_user" corepack pnpm install
    exec su-exec "$container_user" corepack pnpm run "$package_script"
    ;;
  yarn)
    su-exec "$container_user" corepack yarn install
    exec su-exec "$container_user" corepack yarn run "$package_script"
    ;;
esac
