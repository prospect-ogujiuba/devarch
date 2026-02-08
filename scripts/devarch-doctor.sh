#!/bin/bash

set -eo pipefail

EXIT_CODE=0
API_BASE="${DEVARCH_API_URL:-http://localhost:8550}"

echo "DevArch Diagnostics"
echo "==================="
echo ""

echo "Check 1: Container Runtime"
RUNTIME_FOUND=0

if command -v podman &>/dev/null; then
  PODMAN_VERSION=$(podman --version 2>/dev/null | cut -d' ' -f3 || echo "unknown")
  echo "  [Pass] Podman installed: $PODMAN_VERSION"
  if podman ps &>/dev/null; then
    echo "  [Pass] Podman is responsive"
    RUNTIME_FOUND=1
  else
    echo "  [Warn] Podman installed but not responsive"
  fi
elif command -v docker &>/dev/null; then
  DOCKER_VERSION=$(docker --version 2>/dev/null | cut -d' ' -f3 | tr -d ',' || echo "unknown")
  echo "  [Pass] Docker installed: $DOCKER_VERSION"
  if docker ps &>/dev/null 2>&1; then
    echo "  [Pass] Docker is responsive"
    RUNTIME_FOUND=1
  else
    echo "  [Warn] Docker installed but not responsive (may need sudo)"
  fi
else
  echo "  [Fail] No container runtime found (podman/docker)"
  EXIT_CODE=1
fi

echo ""
echo "Check 2: API Server"
if curl -s -f -H "X-API-Key: test" "$API_BASE/api/v1/health" &>/dev/null; then
  echo "  [Pass] API server reachable at $API_BASE"
else
  echo "  [Warn] API server unreachable at $API_BASE"
fi

echo ""
echo "Check 3: Disk Space"
AVAILABLE_GB=$(df -BG . | tail -1 | awk '{print $4}' | tr -d 'G')
if [[ "$AVAILABLE_GB" -lt 1 ]]; then
  echo "  [Warn] Low disk space: ${AVAILABLE_GB}GB available"
else
  echo "  [Pass] Disk space: ${AVAILABLE_GB}GB available"
fi

echo ""
echo "Check 4: Required CLI Tools"
TOOLS=("curl" "jq" "grep")
for TOOL in "${TOOLS[@]}"; do
  if command -v "$TOOL" &>/dev/null; then
    echo "  [Pass] $TOOL installed"
  else
    echo "  [Warn] $TOOL not found (recommended)"
  fi
done

echo ""
echo "Check 5: Port Availability"
PORTS=(5432 3306 6379 8080)
for PORT in "${PORTS[@]}"; do
  if command -v nc &>/dev/null; then
    if nc -z localhost "$PORT" 2>/dev/null; then
      echo "  [Warn] Port $PORT is in use"
    else
      echo "  [Pass] Port $PORT available"
    fi
  else
    echo "  [Warn] nc not installed, skipping port checks"
    break
  fi
done

echo ""
echo "==================="
if [[ "$EXIT_CODE" -eq 0 ]]; then
  echo "All critical checks passed"
else
  echo "Some checks failed"
fi

exit $EXIT_CODE
