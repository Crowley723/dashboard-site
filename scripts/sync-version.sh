#!/bin/bash
set -e

VERSION_FILE="VERSION"
HELM_CHART="helm/Chart.yaml"
PACKAGE_JSON="web/package.json"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

validate_semver() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        echo -e "${RED}Error: Invalid semantic version format: $version${NC}"
        echo "Expected format: MAJOR.MINOR.PATCH (e.g., 1.2.3)"
        exit 1
    fi
}

get_current_version() {
    if [ ! -f "$VERSION_FILE" ]; then
        echo -e "${RED}Error: VERSION file not found${NC}"
        exit 1
    fi
    cat "$VERSION_FILE" | tr -d '\n'
}

update_version_file() {
    local new_version=$1
    echo "$new_version" > "$VERSION_FILE"
    echo -e "${GREEN}Updated VERSION file to $new_version${NC}"
}

update_helm_chart() {
    local new_version=$1
    if [ -f "$HELM_CHART" ]; then
        sed -i "s/^version: .*/version: $new_version/" "$HELM_CHART"
        echo -e "${GREEN}Updated Helm Chart.yaml to $new_version${NC}"
    else
        echo -e "${YELLOW}Helm Chart.yaml not found, skipping${NC}"
    fi
}

update_package_json() {
    local new_version=$1
    if [ -f "$PACKAGE_JSON" ]; then
        if command -v jq &> /dev/null; then
            jq --arg version "$new_version" '.version = $version' "$PACKAGE_JSON" > "$PACKAGE_JSON.tmp"
            mv "$PACKAGE_JSON.tmp" "$PACKAGE_JSON"
        else
            sed -i "s/\"version\": \".*\"/\"version\": \"$new_version\"/" "$PACKAGE_JSON"
        fi
        echo -e "${GREEN}Updated package.json to $new_version${NC}"
    else
        echo -e "${YELLOW}package.json not found, skipping${NC}"
    fi
}

main() {
    local current_version=$(get_current_version)
    local new_version=""

    if [ $# -eq 0 ]; then
        echo -e "${GREEN}Current version: $current_version${NC}"
        echo ""
        echo "To update version, run:"
        echo "  ./scripts/sync-version.sh <new-version>"
        exit 0
    elif [ $# -eq 1 ]; then
        new_version=$1
    else
        echo -e "${RED}Error: Too many arguments${NC}"
        echo "Usage: $0 [new-version]"
        exit 1
    fi

    validate_semver "$new_version"

    echo -e "${YELLOW}Updating version from $current_version to $new_version${NC}"
    echo ""

    update_version_file "$new_version"
    update_helm_chart "$new_version"
    update_package_json "$new_version"

    echo ""
    echo -e "${GREEN}Version sync complete${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review the changes: git diff"
    echo "2. Commit the changes: git add -A && git commit -m \"chore: bump version to $new_version\""
    echo "3. Tag the release: git tag -a v$new_version -m \"Release v$new_version\""
    echo "4. Push changes: git push && git push --tags"
}

main "$@"
