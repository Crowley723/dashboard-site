#!/bin/bash
# scripts/validate-k8s.sh
set -e

echo "🔍 Validating Kubernetes YAML files..."

# Check if kubeconform is installed
if ! command -v kubeconform &> /dev/null; then
    echo "📦 Installing kubeconform..."

    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
    esac

    # Download and install kubeconform
    KUBECONFORM_VERSION="v0.6.4"
    curl -L "https://github.com/yannh/kubeconform/releases/download/${KUBECONFORM_VERSION}/kubeconform-${OS}-${ARCH}.tar.gz" | tar xz
    sudo mv kubeconform /usr/local/bin/
    echo "✅ kubeconform installed"
fi

# Validate each file passed to the script
for file in "$@"; do
    echo "📋 Validating: $file"

    # Skip if file doesn't exist or isn't a YAML file
    if [[ ! -f "$file" ]] || [[ ! "$file" =~ \.(yaml|yml)$ ]]; then
        continue
    fi

    # Skip Helm templates (they contain Go template syntax)
    if [[ "$file" == *"/templates/"* ]]; then
        echo "⏭️  Skipping Helm template: $file"
        continue
    fi

    # Validate with kubeconform
    if ! kubeconform -summary -verbose "$file"; then
        echo "❌ Validation failed for: $file"
        exit 1
    fi

    echo "✅ Valid: $file"
done

echo "🎉 All Kubernetes YAML files validated successfully!"
