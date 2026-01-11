#!/bin/bash
# =============================================================================
# Dayawarga Version Management Script
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$PROJECT_ROOT/VERSION"
CHANGELOG_FILE="$PROJECT_ROOT/CHANGELOG.md"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

get_version() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE" | tr -d '\n'
    else
        echo "0.0.0"
    fi
}

get_git_info() {
    cd "$PROJECT_ROOT"

    local commit_hash=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local commit_date=$(git log -1 --format=%ci 2>/dev/null | cut -d' ' -f1 || echo "unknown")
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    local dirty=""

    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        dirty="-dirty"
    fi

    echo "{\"commit\":\"${commit_hash}${dirty}\",\"date\":\"${commit_date}\",\"branch\":\"${branch}\"}"
}

show_info() {
    local version=$(get_version)
    local git_info=$(get_git_info)

    echo ""
    echo -e "${BLUE}Dayawarga Senyar${NC}"
    echo -e "Version: ${GREEN}v${version}${NC}"
    echo -e "Git: $(echo $git_info | jq -r '.commit')"
    echo -e "Branch: $(echo $git_info | jq -r '.branch')"
    echo -e "Date: $(echo $git_info | jq -r '.date')"
    echo ""
}

show_json() {
    local version=$(get_version)
    local git_info=$(get_git_info)

    echo "{\"version\":\"${version}\",\"git\":${git_info}}"
}

bump_version() {
    local current=$(get_version)
    local type="${1:-patch}"

    IFS='.' read -r major minor patch <<< "$current"

    case "$type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            echo -e "${RED}Invalid version type: $type${NC}"
            echo "Usage: $0 bump [major|minor|patch]"
            exit 1
            ;;
    esac

    local new_version="${major}.${minor}.${patch}"
    echo "$new_version" > "$VERSION_FILE"

    echo -e "${GREEN}Version bumped: v${current} -> v${new_version}${NC}"
}

set_version() {
    local new_version="$1"

    if [[ ! "$new_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Invalid version format: $new_version${NC}"
        echo "Expected format: X.Y.Z (e.g., 1.2.3)"
        exit 1
    fi

    echo "$new_version" > "$VERSION_FILE"
    echo -e "${GREEN}Version set to: v${new_version}${NC}"
}

create_release_tag() {
    local version=$(get_version)
    local tag="v${version}"

    cd "$PROJECT_ROOT"

    # Check if tag already exists
    if git rev-parse "$tag" >/dev/null 2>&1; then
        echo -e "${YELLOW}Tag $tag already exists${NC}"
        return 1
    fi

    # Create annotated tag
    git tag -a "$tag" -m "Release $tag"
    echo -e "${GREEN}Created tag: $tag${NC}"

    echo -e "${YELLOW}To push the tag, run:${NC}"
    echo "  git push origin $tag"
}

show_help() {
    echo "Dayawarga Version Management"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  info          Show current version info (default)"
    echo "  json          Output version info as JSON"
    echo "  get           Get current version number only"
    echo "  bump [type]   Bump version (major|minor|patch)"
    echo "  set VERSION   Set specific version (e.g., 1.2.3)"
    echo "  tag           Create git tag for current version"
    echo "  help          Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 info           # Show version info"
    echo "  $0 bump patch     # 1.0.0 -> 1.0.1"
    echo "  $0 bump minor     # 1.0.1 -> 1.1.0"
    echo "  $0 bump major     # 1.1.0 -> 2.0.0"
    echo "  $0 set 2.0.0      # Set to specific version"
    echo "  $0 tag            # Create git tag v2.0.0"
}

# Main
case "${1:-info}" in
    info)
        show_info
        ;;
    json)
        show_json
        ;;
    get)
        get_version
        ;;
    bump)
        bump_version "${2:-patch}"
        ;;
    set)
        if [ -z "$2" ]; then
            echo -e "${RED}Version required${NC}"
            echo "Usage: $0 set VERSION"
            exit 1
        fi
        set_version "$2"
        ;;
    tag)
        create_release_tag
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        show_help
        exit 1
        ;;
esac
