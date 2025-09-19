#!/bin/bash
set -e

echo "🔍 Validating Helm charts..."

# Find all Chart.yaml files to identify Helm charts
charts=$(find . -name "Chart.yaml" -exec dirname {} \;)

if [ -z "$charts" ]; then
    echo "ℹ️  No Helm charts found, skipping validation"
    exit 0
fi

for chart in $charts; do
    echo "📊 Validating chart: $chart"

    # Check if Chart.yaml is valid
    if ! helm lint "$chart"; then
        echo "❌ Helm lint failed for $chart"
        exit 1
    fi

    # Try to template the chart (dry-run)
    if ! helm template test-release "$chart" --dry-run > /dev/null; then
        echo "❌ Helm template validation failed for $chart"
        exit 1
    fi

    echo "✅ Chart $chart is valid"
done

echo "🎉 All Helm charts validated successfully!"
