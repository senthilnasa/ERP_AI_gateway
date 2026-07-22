#!/usr/bin/env bash

# Complete Production Rebuild & Redeployment Script for ERP AI Gateway

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "  ERP AI Gateway - Full Production Redeploy"
echo "=========================================="

# 1. Pull latest code from Git
echo "[1/6] Pulling latest updates from Git repository..."
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git pull origin main || echo "⚠️ Warning: git pull skipped or up to date."
else
    echo "Notice: Not a git repo, skipping git pull."
fi

# 2. Stop running service / process
echo "[2/6] Stopping active application..."
if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet erp-ai-gateway.service 2>/dev/null; then
    echo "Stopping systemd service erp-ai-gateway.service..."
    sudo systemctl stop erp-ai-gateway.service || true
else
    PID=$(pgrep -f "bin/ai-gateway" || pgrep -f "ai-gateway" || true)
    if [ -n "$PID" ]; then
        echo "Stopping running process PID: $PID"
        kill -9 $PID 2>/dev/null || true
        sleep 1
    fi
fi

# 3. Clean old binaries and cache
echo "[3/6] Cleaning old binaries..."
rm -rf bin/ai-gateway server

# 4. Verify / Generate Configuration
echo "[4/6] Verifying configuration file..."
./scripts/setup_config.sh -y

# 5. Tidy dependencies and build fresh binary
echo "[5/6] Running tests and building clean production binary..."
go mod tidy
./scripts/build.sh

# 6. Restart Service & Verify Health
echo "[6/6] Launching updated application..."
if command -v systemctl >/dev/null 2>&1 && [ -f "/etc/systemd/system/erp-ai-gateway.service" ]; then
    echo "Restarting systemd service erp-ai-gateway.service..."
    sudo systemctl restart erp-ai-gateway.service
    sleep 2
    sudo systemctl status erp-ai-gateway.service --no-pager
else
    echo "Starting process in background..."
    nohup ./bin/ai-gateway > ai-gateway.log 2>&1 &
    NEW_PID=$!
    sleep 2

    if kill -0 $NEW_PID 2>/dev/null; then
        echo "✔ Application started successfully (PID: $NEW_PID)."
    else
        echo "✖ Process failed to start! Checking logs..."
        tail -n 20 ai-gateway.log
        exit 1
    fi
fi

# 7. Check Health Endpoint
echo ""
echo "Verifying /health endpoint..."
curl -s http://localhost:8080/health || echo "Note: Check health at your domain URL"

echo ""
echo "=========================================="
echo " ✔ Production Redeployment Complete!"
echo " Logs: ai-gateway.log or journalctl -u erp-ai-gateway -f"
echo "=========================================="
