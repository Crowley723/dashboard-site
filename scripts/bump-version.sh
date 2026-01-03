#!/bin/bash
set -e

VERSION_FILE="VERSION"
BUMP_TYPE=${1:-patch}

if [ ! -f "$VERSION_FILE" ]; then
    echo "Error: VERSION file not found" >&2
    exit 1
fi

CURRENT_VERSION=$(cat "$VERSION_FILE" | tr -d '\n')

IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"

case "$BUMP_TYPE" in
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
        echo "Error: Invalid bump type. Use: major, minor, or patch" >&2
        exit 1
        ;;
esac

echo "$major.$minor.$patch"
