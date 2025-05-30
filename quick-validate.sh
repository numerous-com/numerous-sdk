#!/bin/bash

# Quick Validation Script - Common Use Cases
# Wrapper around validate-task-workflow.sh for common scenarios

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALIDATION_SCRIPT="${SCRIPT_DIR}/validate-task-workflow.sh"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Numerous Task Workflow - Quick Validation${NC}"
echo "============================================="
echo ""

# Check if validation script exists
if [ ! -f "$VALIDATION_SCRIPT" ]; then
    echo "❌ Validation script not found at $VALIDATION_SCRIPT"
    exit 1
fi

# Build CLI if needed
if [ ! -f "${SCRIPT_DIR}/bin/numerous" ]; then
    echo "🔨 Building CLI binary..."
    cd "$SCRIPT_DIR"
    go build -o bin/numerous .
    echo -e "${GREEN}✅ CLI built successfully${NC}"
    echo ""
fi

# Show common usage options
echo "Select validation scenario:"
echo ""
echo "1. Full validation (interactive org selection)"
echo "2. Full validation with specific org"
echo "3. Quick test (skip login, use existing token)"
echo "4. Development mode (verbose, no cleanup)"
echo "5. CI/CD mode (skip login, minimal output)"
echo "6. Custom options"
echo "7. Show help"
echo ""

read -p "Choose option (1-7): " choice

case $choice in
    1)
        echo -e "${GREEN}Running full validation with interactive org selection...${NC}"
        exec "$VALIDATION_SCRIPT"
        ;;
    2)
        read -p "Enter organization slug: " org_slug
        if [ -z "$org_slug" ]; then
            echo "❌ Organization slug is required"
            exit 1
        fi
        echo -e "${GREEN}Running full validation for org: $org_slug${NC}"
        exec "$VALIDATION_SCRIPT" --org "$org_slug"
        ;;
    3)
        read -p "Enter organization slug: " org_slug
        if [ -z "$org_slug" ]; then
            echo "❌ Organization slug is required"
            exit 1
        fi
        echo -e "${GREEN}Running quick test (skip login) for org: $org_slug${NC}"
        exec "$VALIDATION_SCRIPT" --org "$org_slug" --skip-login
        ;;
    4)
        read -p "Enter organization slug: " org_slug
        if [ -z "$org_slug" ]; then
            echo "❌ Organization slug is required"
            exit 1
        fi
        echo -e "${GREEN}Running development mode (verbose, no cleanup) for org: $org_slug${NC}"
        exec "$VALIDATION_SCRIPT" --org "$org_slug" --verbose --no-cleanup
        ;;
    5)
        read -p "Enter organization slug: " org_slug
        if [ -z "$org_slug" ]; then
            echo "❌ Organization slug is required"
            exit 1
        fi
        echo -e "${GREEN}Running CI/CD mode for org: $org_slug${NC}"
        exec "$VALIDATION_SCRIPT" --org "$org_slug" --skip-login
        ;;
    6)
        echo "Available options:"
        echo "  --org SLUG           Organization slug"
        echo "  --api-url URL        API endpoint URL"
        echo "  --skip-login         Skip login step"
        echo "  --no-cleanup         Don't cleanup resources"
        echo "  --verbose            Enable verbose output"
        echo ""
        read -p "Enter custom options: " custom_options
        echo -e "${GREEN}Running with custom options: $custom_options${NC}"
        exec "$VALIDATION_SCRIPT" $custom_options
        ;;
    7)
        exec "$VALIDATION_SCRIPT" --help
        ;;
    *)
        echo "❌ Invalid option. Please choose 1-7."
        exit 1
        ;;
esac 