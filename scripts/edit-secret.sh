#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <encrypted-secret-file>"
    echo "Example: $0 authelia/templates/secret.enc.yaml"
    exit 1
fi

ENCRYPTED_FILE="$1"

if [ ! -f "$ENCRYPTED_FILE" ]; then
    echo "File not found: $ENCRYPTED_FILE"
    exit 1
fi

echo "Editing encrypted secret: $ENCRYPTED_FILE"
sops --input-type yaml --output-type yaml "$ENCRYPTED_FILE"
