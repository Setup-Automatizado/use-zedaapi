#!/bin/bash
# Version Bump Script
# Usage: ./scripts/bump-version.sh [major|minor|patch|prerelease] [--dry-run]
#
# Examples:
#   ./scripts/bump-version.sh patch           # 2.0.0 -> 2.0.1
#   ./scripts/bump-version.sh minor           # 2.0.0 -> 2.1.0
#   ./scripts/bump-version.sh major           # 2.0.0 -> 3.0.0
#   ./scripts/bump-version.sh prerelease      # 2.0.0-develop.1 -> 2.0.0-develop.2

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$ROOT_DIR/VERSION"
MANAGER_PACKAGE_JSON="$ROOT_DIR/manager-whatsapp-api-golang/package.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
BUMP_TYPE="${1:-patch}"
DRY_RUN=false

if [[ "$2" == "--dry-run" ]] || [[ "$1" == "--dry-run" ]]; then
    DRY_RUN=true
    if [[ "$1" == "--dry-run" ]]; then
        BUMP_TYPE="patch"
    fi
fi

# Read current version
if [[ ! -f "$VERSION_FILE" ]]; then
    echo -e "${RED}ERROR: VERSION file not found at $VERSION_FILE${NC}"
    exit 1
fi

CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '\n')
echo -e "${YELLOW}Current version: $CURRENT_VERSION${NC}"

# Parse version components
# Handle both standard semver (1.2.3) and prerelease (1.2.3-develop.1)
if [[ "$CURRENT_VERSION" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-([a-zA-Z]+)\.([0-9]+))?$ ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    PATCH="${BASH_REMATCH[3]}"
    PRERELEASE_TAG="${BASH_REMATCH[5]}"
    PRERELEASE_NUM="${BASH_REMATCH[6]}"
else
    echo -e "${RED}ERROR: Invalid version format: $CURRENT_VERSION${NC}"
    echo "Expected format: X.Y.Z or X.Y.Z-tag.N (e.g., 2.0.0 or 2.0.0-develop.1)"
    exit 1
fi

# Calculate new version based on bump type
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
    prerelease)
        if [[ -z "$PRERELEASE_TAG" ]]; then
            # Start prerelease from current version
            NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}-develop.1"
        else
            NEW_PRERELEASE=$((PRERELEASE_NUM + 1))
            NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}-${PRERELEASE_TAG}.${NEW_PRERELEASE}"
        fi
        ;;
    release)
        # Remove prerelease tag to create release version
        NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
        ;;
    *)
        echo -e "${RED}ERROR: Invalid bump type: $BUMP_TYPE${NC}"
        echo "Valid types: major, minor, patch, prerelease, release"
        exit 1
        ;;
esac

echo -e "${GREEN}New version: $NEW_VERSION${NC}"

if [[ "$DRY_RUN" == true ]]; then
    echo -e "${YELLOW}DRY RUN: No changes will be made${NC}"
    exit 0
fi

# Update VERSION file
echo "$NEW_VERSION" > "$VERSION_FILE"
echo -e "${GREEN}Updated VERSION file${NC}"

# Update Manager package.json
if [[ -f "$MANAGER_PACKAGE_JSON" ]]; then
    # Use sed to update version in package.json
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS
        sed -i '' "s/\"version\": \"[^\"]*\"/\"version\": \"$NEW_VERSION\"/" "$MANAGER_PACKAGE_JSON"
    else
        # Linux
        sed -i "s/\"version\": \"[^\"]*\"/\"version\": \"$NEW_VERSION\"/" "$MANAGER_PACKAGE_JSON"
    fi
    echo -e "${GREEN}Updated Manager package.json${NC}"
else
    echo -e "${YELLOW}WARNING: Manager package.json not found${NC}"
fi

echo ""
echo -e "${GREEN}Version bumped successfully!${NC}"
echo ""
echo "Files updated:"
echo "  - $VERSION_FILE"
echo "  - $MANAGER_PACKAGE_JSON"
echo ""
echo "Next steps:"
echo "  1. Review the changes: git diff"
echo "  2. Commit: git add -A && git commit -m 'chore: bump version to $NEW_VERSION'"
echo "  3. Tag (optional): git tag v$NEW_VERSION"
