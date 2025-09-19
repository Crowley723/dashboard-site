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

echo "Decrypting $ENCRYPTED_FILE (output to stdout only):"
echo "---"
sops --input-type yaml --output-type yaml -d "$ENCRYPTED_FILE"
