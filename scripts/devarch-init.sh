#!/bin/bash

set -eo pipefail

YML_FILE="${1:-devarch.yml}"
API_BASE="${DEVARCH_API_URL:-http://localhost:8550}"
API_KEY="${DEVARCH_API_KEY:-test}"

if [[ ! -f "$YML_FILE" ]]; then
  echo "Error: File not found: $YML_FILE"
  exit 1
fi

STACK_NAME=$(grep -A 5 "^stack:" "$YML_FILE" | grep "name:" | head -1 | sed 's/.*name: *"\?\([^"]*\)"\?.*/\1/')

if [[ -z "$STACK_NAME" ]]; then
  echo "Error: Could not extract stack name from $YML_FILE"
  exit 1
fi

echo "Initializing stack: $STACK_NAME"
echo "Step 1/3: Importing stack configuration..."

IMPORT_RESULT=$(curl -s -w "\n%{http_code}" -X POST \
  -H "X-API-Key: $API_KEY" \
  -F "file=@$YML_FILE" \
  "$API_BASE/api/v1/stacks/import")

HTTP_CODE=$(echo "$IMPORT_RESULT" | tail -1)
RESPONSE_BODY=$(echo "$IMPORT_RESULT" | sed '$d')

if [[ "$HTTP_CODE" -ne 200 ]]; then
  echo "Error: Import failed with HTTP $HTTP_CODE"
  echo "$RESPONSE_BODY"
  exit 1
fi

CREATED_COUNT=$(echo "$RESPONSE_BODY" | jq -r '.created | length')
UPDATED_COUNT=$(echo "$RESPONSE_BODY" | jq -r '.updated | length')
echo "Import complete. Created: $CREATED_COUNT, Updated: $UPDATED_COUNT"

echo "Step 2/3: Pulling container images..."

IMAGES=$(grep -E "^[[:space:]]+image:" "$YML_FILE" | sed 's/.*image: *"\?\([^"]*\)"\?.*/\1/' | sort -u)

if command -v podman &>/dev/null; then
  RUNTIME="podman"
elif command -v docker &>/dev/null; then
  RUNTIME="docker"
else
  echo "Warning: No container runtime found (podman/docker). Skipping image pull."
  RUNTIME=""
fi

if [[ -n "$RUNTIME" ]]; then
  for IMAGE in $IMAGES; do
    echo "Pulling $IMAGE..."
    $RUNTIME pull "$IMAGE" || echo "Warning: Failed to pull $IMAGE"
  done
fi

echo "Step 3/3: Applying stack..."

APPLY_RESULT=$(curl -s -w "\n%{http_code}" -X POST \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"\"}" \
  "$API_BASE/api/v1/stacks/$STACK_NAME/apply")

HTTP_CODE=$(echo "$APPLY_RESULT" | tail -1)
RESPONSE_BODY=$(echo "$APPLY_RESULT" | sed '$d')

if [[ "$HTTP_CODE" -ne 200 ]]; then
  echo "Warning: Apply returned HTTP $HTTP_CODE"
  echo "$RESPONSE_BODY"
else
  echo "Apply complete."
fi

echo ""
echo "Stack '$STACK_NAME' initialized successfully!"
echo "View status: curl -H 'X-API-Key: $API_KEY' $API_BASE/api/v1/stacks/$STACK_NAME"
