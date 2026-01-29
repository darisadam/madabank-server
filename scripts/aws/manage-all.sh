#!/bin/bash
# Wrapper to manage ALL environments
ACTION=$1

if [ -z "$ACTION" ]; then
    echo "Usage: ./scripts/manage-all.sh [start|stop]"
    exit 1
fi

echo "========================================"
echo "🌍 Executing '$ACTION' on ALL Environments"
echo "========================================"

# Dev
echo
echo "--- Development ---"
./scripts/aws/cost-saver.sh "$ACTION" dev

# Staging
echo
echo "--- Staging ---"
./scripts/aws/cost-saver.sh "$ACTION" staging

# Prod
echo
echo "--- Production ---"
./scripts/aws/cost-saver.sh "$ACTION" prod

echo
echo "✅ All environments processed."
