#!/usr/bin/env bash

# Test 10 concurrent write requests against AI Gateway

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_FILE="${CONFIG_FILE:-$SCRIPT_DIR/config/config.yaml}"

URL="${URL:-http://localhost:8080/api/v1/write}"

# Dynamically load API Key from config/config.yaml if not set in environment
if [ -z "$API_KEY" ]; then
    if [ -f "$CONFIG_FILE" ]; then
        API_KEY=$(awk '/security:/{f=1} f && /api_key:/{sub(/.*api_key:[ \t]*/, ""); gsub(/^["'\''`]+|["'\''`]+$/, ""); print; exit}' "$CONFIG_FILE")
    fi
fi
API_KEY="${API_KEY:-krea-secret-ai-key-2026}"

MODEL_NAME="${1:-qwen2.5:0.5b}"

echo "=========================================="
echo " Starting 10 Concurrent Requests Benchmark"
echo " Target Model: $MODEL_NAME"
echo "=========================================="

START_TIME=$(date +%s)

run_request() {
    ID=$1
    echo "[Req #$ID] Sending..."
    RESP=$(curl -s -X POST "$URL" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
        "profile": "email",
        "action": "rewrite",
        "tone": "professional",
        "language": "english",
        "text": "please approve the purchase order for 5 laptops",
        "options": {
          "model": "'"$MODEL_NAME"'"
        }
      }')
    
    SUCCESS=$(echo "$RESP" | grep -o '"success":true' || echo "failed")
    TIME_MS=$(echo "$RESP" | grep -o '"processing_ms":[0-9]*' | cut -d':' -f2 || echo "0")
    echo "[Req #$ID] Finished -> Status: $SUCCESS | Duration: ${TIME_MS}ms"
}

export -f run_request
export URL API_KEY MODEL_NAME

# Spawn 10 concurrent subshells
for i in {1..10}; do
    run_request $i &
done

# Wait for all background requests to complete
wait

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

echo "=========================================="
echo " Benchmark Completed!"
echo " Total Wall-Clock Time: ${TOTAL_DURATION} seconds for 10 concurrent requests."
echo "=========================================="
