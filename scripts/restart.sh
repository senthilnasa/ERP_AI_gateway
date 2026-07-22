#!/usr/bin/env bash

# Exit on error
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "  OneERP AI Gateway - Redeploy & Restart  "
echo "=========================================="

# 1. Kill any existing running process
echo "[1/5] Stopping any running gateway processes..."
PID=$(pgrep -f "bin/ai-gateway" || pgrep -f "ai-gateway" || true)
if [ -n "$PID" ]; then
    echo "Terminating process PID: $PID"
    kill -9 $PID 2>/dev/null || true
    sleep 1
    echo "✔ Process stopped."
else
    echo "✔ No existing process running."
fi

# 2. Remove old binary
echo "[2/5] Cleaning old binaries..."
rm -f bin/ai-gateway server

# 3. Ensure Config setup
echo "[3/5] Checking configuration file..."
./scripts/setup_config.sh

# 4. Build new app cleanly
echo "[4/5] Building fresh application binary..."
./scripts/build.sh

# 5. Start background app and verify error-free execution
echo "[5/5] Launching new binary in background..."
nohup ./bin/ai-gateway > ai-gateway.log 2>&1 &
NEW_PID=$!
sleep 2

# Check if process is still alive
if kill -0 $NEW_PID 2>/dev/null; then
    echo "✔ Server started successfully (PID: $NEW_PID)."
    echo "Logs writing to: ai-gateway.log"
    echo "=========================================="
    echo " Gateway is live at http://localhost:8080"
    echo " Swagger docs: http://localhost:8080/docs"
    echo "=========================================="
else
    echo "✖ Server failed to start! Checking logs..."
    tail -n 20 ai-gateway.log
    exit 1
fi
