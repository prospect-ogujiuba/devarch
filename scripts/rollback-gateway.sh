#!/bin/zsh
# Rollback Dynamic Gateway Installation

echo "🔄 Rolling back dynamic gateway installation..."

# Stop gateway container
echo "Stopping gateway container..."
/home/priz/projects/devarch/scripts/service-manager.sh down dynamic-gateway 2>/dev/null || true

# Remove gateway files
echo "Removing gateway files..."
rm -rf "/home/priz/projects/devarch/compose/gateway"
rm -rf "/home/priz/projects/devarch/config/gateway"
rm -f "/home/priz/projects/devarch/config/node/project-handler.js"
rm -f "/home/priz/projects/devarch/scripts/create-project.sh"
rm -f "/home/priz/projects/devarch/scripts/rollback-gateway.sh"

# Restore original files
if [[ -d "/home/priz/projects/devarch/backup-20250619-193326" ]]; then
    echo "Restoring original files from /home/priz/projects/devarch/backup-20250619-193326..."
    [[ -f "/home/priz/projects/devarch/backup-20250619-193326/config.sh" ]] && cp "/home/priz/projects/devarch/backup-20250619-193326/config.sh" "/home/priz/projects/devarch/scripts/"
    [[ -f "/home/priz/projects/devarch/backup-20250619-193326/traefik.yml" ]] && cp "/home/priz/projects/devarch/backup-20250619-193326/traefik.yml" "/home/priz/projects/devarch/config/traefik/"
    [[ -d "/home/priz/projects/devarch/backup-20250619-193326/node" ]] && cp -r "/home/priz/projects/devarch/backup-20250619-193326/node" "/home/priz/projects/devarch/config/"
    [[ -d "/home/priz/projects/devarch/backup-20250619-193326/python" ]] && cp -r "/home/priz/projects/devarch/backup-20250619-193326/python" "/home/priz/projects/devarch/config/"
    [[ -d "/home/priz/projects/devarch/backup-20250619-193326/go" ]] && cp -r "/home/priz/projects/devarch/backup-20250619-193326/go" "/home/priz/projects/devarch/config/"
    [[ -d "/home/priz/projects/devarch/backup-20250619-193326/dotnet" ]] && cp -r "/home/priz/projects/devarch/backup-20250619-193326/dotnet" "/home/priz/projects/devarch/config/"
fi

# Remove test projects
rm -rf "/home/priz/projects/devarch/apps/test-node"
rm -rf "/home/priz/projects/devarch/apps/test-python"
rm -rf "/home/priz/projects/devarch/apps/test-static"

# Clean up
rm -f "/home/priz/projects/devarch/.gateway-backup-location"

echo "✅ Rollback completed!"
echo "ℹ️  You may need to restart your services: ./scripts/start-services.sh"
