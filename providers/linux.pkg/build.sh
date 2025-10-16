#!/usr/bin/env bash
# Build script for linux.pkg WASM provider

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Output files
WASM_OUTPUT="plugin.wasm"
MANIFEST="manifest.yaml"

echo -e "${GREEN}Building linux.pkg WASM provider...${NC}"

# Check for TinyGo
if ! command -v tinygo &> /dev/null; then
    echo -e "${RED}Error: TinyGo is not installed${NC}"
    echo "Please install TinyGo from: https://tinygo.org/getting-started/install/"
    exit 1
fi

TINYGO_VERSION=$(tinygo version | head -n1)
echo -e "${GREEN}Using TinyGo: ${TINYGO_VERSION}${NC}"

# Clean previous build
if [ -f "$WASM_OUTPUT" ]; then
    echo "Cleaning previous build..."
    rm -f "$WASM_OUTPUT"
fi

# Build WASM module
echo "Compiling to WASM..."
tinygo build \
    -o "$WASM_OUTPUT" \
    -target=wasi \
    -opt=2 \
    -no-debug \
    -scheduler=none \
    main.go

if [ ! -f "$WASM_OUTPUT" ]; then
    echo -e "${RED}Error: Build failed - $WASM_OUTPUT not created${NC}"
    exit 1
fi

# Get file size
WASM_SIZE=$(wc -c < "$WASM_OUTPUT" | tr -d ' ')
WASM_SIZE_KB=$((WASM_SIZE / 1024))

echo -e "${GREEN}Build successful!${NC}"
echo "WASM module size: ${WASM_SIZE_KB}KB"

# Calculate SHA256 checksum
if command -v sha256sum &> /dev/null; then
    CHECKSUM=$(sha256sum "$WASM_OUTPUT" | cut -d' ' -f1)
elif command -v shasum &> /dev/null; then
    CHECKSUM=$(shasum -a 256 "$WASM_OUTPUT" | cut -d' ' -f1)
else
    echo -e "${YELLOW}Warning: sha256sum or shasum not found, skipping checksum${NC}"
    CHECKSUM=""
fi

if [ -n "$CHECKSUM" ]; then
    echo "SHA256 checksum: $CHECKSUM"

    # Update manifest with checksum
    if command -v sed &> /dev/null; then
        # macOS compatible sed
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/plugin.wasm: \"\"/plugin.wasm: \"$CHECKSUM\"/" "$MANIFEST"
        else
            sed -i "s/plugin.wasm: \"\"/plugin.wasm: \"$CHECKSUM\"/" "$MANIFEST"
        fi
        echo "Updated manifest with checksum"
    fi
fi

# Update build info in manifest
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
    COMMIT=$(git rev-parse --short HEAD)

    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/timestamp: \"\"/timestamp: \"$TIMESTAMP\"/" "$MANIFEST"
        sed -i '' "s/commit: \"\"/commit: \"$COMMIT\"/" "$MANIFEST"
        sed -i '' "s/tinygo_version: \"\"/tinygo_version: \"$TINYGO_VERSION\"/" "$MANIFEST"
    else
        sed -i "s/timestamp: \"\"/timestamp: \"$TIMESTAMP\"/" "$MANIFEST"
        sed -i "s/commit: \"\"/commit: \"$COMMIT\"/" "$MANIFEST"
        sed -i "s/tinygo_version: \"\"/tinygo_version: \"$TINYGO_VERSION\"/" "$MANIFEST"
    fi
fi

# Validate WASM module
echo "Validating WASM module..."
if command -v wasm-validate &> /dev/null; then
    if wasm-validate "$WASM_OUTPUT"; then
        echo -e "${GREEN}WASM module is valid${NC}"
    else
        echo -e "${RED}Error: WASM module validation failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}Warning: wasm-validate not found, skipping validation${NC}"
    echo "Install wabt (WebAssembly Binary Toolkit) for validation"
fi

# Optional: Optimize with wasm-opt
if command -v wasm-opt &> /dev/null; then
    echo "Optimizing WASM module..."
    ORIGINAL_SIZE=$WASM_SIZE
    wasm-opt -O3 -o "${WASM_OUTPUT}.opt" "$WASM_OUTPUT"
    mv "${WASM_OUTPUT}.opt" "$WASM_OUTPUT"

    NEW_SIZE=$(wc -c < "$WASM_OUTPUT" | tr -d ' ')
    NEW_SIZE_KB=$((NEW_SIZE / 1024))
    SAVED=$((ORIGINAL_SIZE - NEW_SIZE))
    SAVED_KB=$((SAVED / 1024))

    echo "Optimized size: ${NEW_SIZE_KB}KB (saved ${SAVED_KB}KB)"
fi

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo "Output: $WASM_OUTPUT"
echo ""
echo "To package the provider:"
echo "  tar czf linux.pkg-1.0.0.tar.gz plugin.wasm manifest.yaml schemas/"
echo ""
echo "To test with OpenFroyo:"
echo "  froyo provider install ./linux.pkg-1.0.0.tar.gz"
