#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <encrypted-secret-file>"
    echo "Example: $0 authelia/templates/secret.enc.yaml"
    exit 1
fi

ENCRYPTED_FILE="$1"
SEALED_FILE="${ENCRYPTED_FILE%.enc.yaml}.sealed.yaml"

if [ ! -f "$ENCRYPTED_FILE" ]; then
    echo "File not found: $ENCRYPTED_FILE"
    exit 1
fi

echo "Converting $ENCRYPTED_FILE to sealed secret: $SEALED_FILE"
sops --input-type yaml --output-type yaml -d "$ENCRYPTED_FILE" | kubeseal -o yaml > "$SEALED_FILE"
echo "Created sealed secret: $SEALED_FILE"
