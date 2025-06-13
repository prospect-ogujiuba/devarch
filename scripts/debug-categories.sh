#!/bin/zsh
# =============================================================================
# debug-categories.sh - Debug category parsing issues
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/config.sh"

echo "🔍 Debugging category parsing..."
echo "Script Dir: $SCRIPT_DIR"
echo "Project Root: $PROJECT_ROOT"
echo "Compose Dir: $COMPOSE_DIR"
echo ""

echo "📋 SERVICE_CATEGORIES definition:"
# Use bash-compatible syntax instead of zsh
echo "Keys: ${!SERVICE_CATEGORIES[*]}"
echo ""

echo "📋 SERVICE_STARTUP_ORDER:"
for i in "${!SERVICE_STARTUP_ORDER[@]}"; do
    echo "  [$i] ${SERVICE_STARTUP_ORDER[$i]}"
done
echo ""

echo "📋 Individual category contents:"
for category in "${!SERVICE_CATEGORIES[@]}"; do
    echo "Category: $category"
    # Use bash-compatible array expansion
    eval "files=(\${SERVICE_CATEGORIES[$category]})"
    for file in "${files[@]}"; do
        echo "  - $file"
        if [ -f "$COMPOSE_DIR/$file" ]; then
            echo "    ✅ File exists"
        else
            echo "    ❌ File missing: $COMPOSE_DIR/$file"
        fi
    done
    echo ""
done

echo "📁 Actual compose directory structure:"
if [ -d "$PROJECT_ROOT/compose" ]; then
    find "$PROJECT_ROOT/compose" -name "*.yml" | sort
else
    echo "❌ Compose directory not found at: $PROJECT_ROOT/compose"
    echo "📁 Contents of project root:"
    ls -la "$PROJECT_ROOT"
fi
echo ""

echo "🔍 Testing determine_install_categories function..."

# Simulate the function call
INSTALL_CATEGORIES=()
EXCLUDE_CATEGORIES=()

categories_to_install=()

if [ ${#INSTALL_CATEGORIES[@]} -gt 0 ]; then
    categories_to_install=("${INSTALL_CATEGORIES[@]}")
    echo "Installing only specified categories: ${categories_to_install[*]}"
else
    for category in "${SERVICE_STARTUP_ORDER[@]}"; do
        if [[ ! " ${EXCLUDE_CATEGORIES[@]} " =~ " ${category} " ]]; then
            categories_to_install+=("$category")
        fi
    done
    
    if [ ${#EXCLUDE_CATEGORIES[@]} -gt 0 ]; then
        echo "Installing all categories except: ${EXCLUDE_CATEGORIES[*]}"
    else
        echo "Installing all categories: ${categories_to_install[*]}"
    fi
fi

echo ""
echo "🎯 Final categories to install:"
for category in "${categories_to_install[@]}"; do
    echo "  - $category"
done