#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "      OneERP AI Gateway - Build Script    "
echo "=========================================="

echo "[1/3] Running Go unit tests..."
if go test ./...; then
    echo "✔ All tests passed successfully."
else
    echo "✖ Unit tests failed! Aborting build."
    exit 1
fi

echo "[2/3] Building executable binary..."
mkdir -p bin
if CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/ai-gateway ./cmd/server; then
    echo "✔ Build succeeded: bin/ai-gateway created."
else
    echo "✖ Compilation failed!"
    exit 1
fi

echo "[3/3] Verifying binary..."
if [ -f "bin/ai-gateway" ]; then
    SIZE=$(du -h bin/ai-gateway | cut -f1)
    echo "✔ Binary verified (Size: $SIZE)."
    echo "=========================================="
    echo " Build Complete! Run with: ./bin/ai-gateway"
    echo "=========================================="
else
    echo "✖ Binary missing!"
    exit 1
fi
