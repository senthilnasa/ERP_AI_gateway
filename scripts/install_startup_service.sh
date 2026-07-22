#!/usr/bin/env bash

# Install ERP AI Gateway as an automatic startup service (systemd for Linux / launchd for macOS)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "  ERP AI Gateway - Startup Service Setup  "
echo "=========================================="

# 1. Build latest binary
echo "[1/4] Ensuring binary is compiled..."
./scripts/build.sh

# 2. Ensure config exists
./scripts/setup_config.sh -y

BINARY_PATH="$SCRIPT_DIR/bin/ai-gateway"
WORKING_DIR="$SCRIPT_DIR"
USER_NAME="$(whoami)"

OS_TYPE="$(uname -s)"

if [ "$OS_TYPE" = "Linux" ]; then
    echo "[2/4] Detected Linux systemd service environment..."
    SERVICE_FILE="/etc/systemd/system/erp-ai-gateway.service"

    echo "Creating systemd unit file at $SERVICE_FILE..."
    sudo bash -c "cat <<EOF > $SERVICE_FILE
[Unit]
Description=ERP AI Gateway Service
After=network.target

[Service]
Type=simple
User=$USER_NAME
WorkingDirectory=$WORKING_DIR
ExecStart=$BINARY_PATH
Restart=always
RestartSec=5
StandardOutput=append:$WORKING_DIR/ai-gateway.log
StandardError=append:$WORKING_DIR/ai-gateway.log

[Install]
WantedBy=multi-user.target
EOF"

    echo "[3/4] Reloading systemd daemon and enabling service..."
    sudo systemctl daemon-reload
    sudo systemctl enable erp-ai-gateway.service
    sudo systemctl restart erp-ai-gateway.service

    echo "[4/4] Service status:"
    sudo systemctl status erp-ai-gateway.service --no-pager

    echo "=========================================="
    echo " ✔ Installed as Linux Systemd Startup Service!"
    echo " Control with:"
    echo "   sudo systemctl status erp-ai-gateway"
    echo "   sudo systemctl stop erp-ai-gateway"
    echo "   sudo systemctl restart erp-ai-gateway"
    echo "=========================================="

elif [ "$OS_TYPE" = "Darwin" ]; then
    echo "[2/4] Detected macOS launchd environment..."
    PLIST_DIR="$HOME/Library/LaunchAgents"
    PLIST_FILE="$PLIST_DIR/com.senthilnasa.erp-ai-gateway.plist"

    mkdir -p "$PLIST_DIR"

    echo "Creating LaunchAgent plist at $PLIST_FILE..."
    cat <<EOF > "$PLIST_FILE"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.senthilnasa.erp-ai-gateway</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$WORKING_DIR</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$WORKING_DIR/ai-gateway.log</string>
    <key>StandardErrorPath</key>
    <string>$WORKING_DIR/ai-gateway.log</string>
</dict>
</plist>
EOF

    echo "[3/4] Loading LaunchAgent into macOS session..."
    launchctl unload "$PLIST_FILE" 2>/dev/null || true
    launchctl load -w "$PLIST_FILE"

    echo "[4/4] Verifying process status..."
    sleep 1
    launchctl list | grep "com.senthilnasa.erp-ai-gateway" || true

    echo "=========================================="
    echo " ✔ Installed as macOS Startup LaunchAgent!"
    echo " Control with:"
    echo "   launchctl load -w $PLIST_FILE"
    echo "   launchctl unload -w $PLIST_FILE"
    echo "=========================================="

else
    echo "✖ Unsupported operating system: $OS_TYPE"
    exit 1
fi
