#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <secret-name>"
    echo "Example: $0 authelia/templates/secret"
    exit 1
fi

SECRET_NAME="$1"
ENCRYPTED_FILE="${SECRET_NAME}.enc.yaml"
TEMP_FILE=$(mktemp --suffix=.yaml)

echo "Creating new encrypted secret: $ENCRYPTED_FILE"
echo "Enter your secret YAML content, then press Ctrl+D when done:"
echo "---"

# Read the input to a temp file first
cat > "$TEMP_FILE"

# Encrypt the file directly, explicitly as YAML
sops --input-type yaml --output-type yaml -e "$TEMP_FILE" > "$ENCRYPTED_FILE"

# Clean up
rm "$TEMP_FILE"

echo "Created encrypted secret: $ENCRYPTED_FILE"
