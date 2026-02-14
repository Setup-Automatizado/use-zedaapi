#!/bin/bash
# Version Bump Script
# Usage: ./scripts/bump-version.sh [major|minor|patch|prerelease] [--dry-run]
#
# Examples:
#   ./scripts/bump-version.sh patch           # 2.0.0 -> 2.0.1
#   ./scripts/bump-version.sh minor           # 2.0.0 -> 2.1.0
#   ./scripts/bump-version.sh major           # 2.0.0 -> 3.0.0

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$ROOT_DIR/VERSION"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BUMP_TYPE="${1:-patch}"
DRY_RUN=false

if [[ "$2" == "--dry-run" ]] || [[ "$1" == "--dry-run" ]]; then
    DRY_RUN=true
    if [[ "$1" == "--dry-run" ]]; then
        BUMP_TYPE="patch"
    fi
fi

if [[ ! -f "$VERSION_FILE" ]]; then
    echo -e "${RED}ERROR: VERSION file not found at $VERSION_FILE${NC}"
    exit 1
fi

CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '\n')
echo -e "${YELLOW}Current version: $CURRENT_VERSION${NC}"

if [[ "$CURRENT_VERSION" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-([a-zA-Z]+)\.([0-9]+))?$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    PATCH="${BASH_REMATCH[3]}"
    PRERELEASE_TAG="${BASH_REMATCH[5]}"
    PRERELEASE_NUM="${BASH_REMATCH[6]}"
else
    echo -e "${RED}ERROR: Invalid version format: $CURRENT_VERSION${NC}"
    echo "Expected format: X.Y.Z (e.g., 3.0.2)"
    exit 1
fi

case "$BUMP_TYPE" in
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_VERSION="${NEW_MAJOR}.0.0"
        ;;
    minor)
        NEW_MINOR=$((MINOR + 1))
        NEW_VERSION="${MAJOR}.${NEW_MINOR}.0"
        ;;
    patch)
        NEW_PATCH=$((PATCH + 1))
        NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
        ;;
    *)
        echo -e "${RED}ERROR: Invalid bump type: $BUMP_TYPE${NC}"
        echo "Valid types: major, minor, patch"
        exit 1
        ;;
esac

echo -e "${GREEN}New version: $NEW_VERSION${NC}"

if [[ "$DRY_RUN" == true ]]; then
    echo -e "${YELLOW}DRY RUN: No changes will be made${NC}"
    exit 0
fi

echo "$NEW_VERSION" > "$VERSION_FILE"
echo -e "${GREEN}Updated VERSION file${NC}"

echo ""
echo -e "${GREEN}Version bumped: $CURRENT_VERSION -> $NEW_VERSION${NC}"
echo ""
echo "Next steps:"
echo "  1. Review: git diff"
echo "  2. Commit: git add VERSION && git commit -m 'chore: bump version to $NEW_VERSION'"
