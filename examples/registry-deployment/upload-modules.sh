#!/bin/bash
# Upload WASM modules to the registry server

set -e

REGISTRY_HOST="${1:-192.168.1.60}"
MODULES_DIR="${2:-../../bin/modules}"
REGISTRY_USER="${3:-root}"

echo "==========================================="
echo "Upload Modules to OpenFroyo Registry"
echo "==========================================="
echo "Registry Host: $REGISTRY_HOST"
echo "Modules Directory: $MODULES_DIR"
echo ""

# Check if modules directory exists
if [ ! -d "$MODULES_DIR" ]; then
    echo "Error: Modules directory not found: $MODULES_DIR"
    echo "Build modules first: make build-wasm"
    exit 1
fi

# Count modules
MODULE_COUNT=$(find "$MODULES_DIR" -name "*.wasm" | wc -l)
if [ "$MODULE_COUNT" -eq 0 ]; then
    echo "Error: No WASM modules found in $MODULES_DIR"
    exit 1
fi

echo "Found $MODULE_COUNT modules to upload"
echo ""

# Upload modules via SCP
echo "Uploading modules..."
scp "$MODULES_DIR"/*.wasm "${REGISTRY_USER}@${REGISTRY_HOST}:/var/lib/froyo-registry/modules/"

if [ $? -eq 0 ]; then
    echo "✓ Modules uploaded successfully!"
else
    echo "✗ Upload failed"
    exit 1
fi

# Set permissions
echo "Setting permissions..."
ssh "${REGISTRY_USER}@${REGISTRY_HOST}" "chown froyo-registry:froyo-registry /var/lib/froyo-registry/modules/*.wasm"

# Restart registry to pick up new modules
echo "Restarting registry..."
ssh "${REGISTRY_USER}@${REGISTRY_HOST}" "systemctl restart froyo-registry"

sleep 2

# Verify
echo ""
echo "Verifying registry..."
REGISTRY_URL="http://${REGISTRY_HOST}:8080"
MODULE_LIST=$(curl -s "${REGISTRY_URL}/modules" | grep -o '"count":[0-9]*' | cut -d: -f2)

echo "✓ Registry has $MODULE_LIST modules"
echo ""
echo "================================================"
echo "Registry URL: $REGISTRY_URL"
echo "Health: $REGISTRY_URL/healthz"
echo "List modules: $REGISTRY_URL/modules"
echo "================================================"
