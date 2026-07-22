#!/usr/bin/env bash

# ==============================================================================
# OneERP AI Gateway - Dynamic Concurrent Requests Benchmark Tool
# ==============================================================================
# Usage:
#   ./scripts/benchmark_n_concurrent.sh [COUNT] [MODEL_NAME]
#   ./scripts/benchmark_n_concurrent.sh -n 25 -m qwen2.5:0.5b
#
# Default COUNT: 20
# Default MODEL: qwen2.5:0.5b (or configured default model)
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_FILE="${CONFIG_FILE:-$SCRIPT_DIR/config/config.yaml}"
URL="${URL:-http://localhost:8080/api/v1/write}"

# Default parameters
COUNT=20
MODEL_NAME="qwen2.5:0.5b"

# Parse CLI positional or flag arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -n|--count)
                COUNT="$2"
                shift 2
                ;;
            -m|--model)
                MODEL_NAME="$2"
                shift 2
                ;;
            -u|--url)
                URL="$2"
                shift 2
                ;;
            -h|--help)
                echo "Usage: $0 [COUNT] [MODEL_NAME]"
                echo "  or:  $0 -n <count> -m <model> -u <url>"
                echo ""
                echo "Options:"
                echo "  -n, --count  Number of concurrent requests (default: 20)"
                echo "  -m, --model  Target LLM model name (default: qwen2.5:0.5b)"
                echo "  -u, --url    Endpoint URL (default: http://localhost:8080/api/v1/write)"
                echo "  -h, --help   Show this help message"
                exit 0
                ;;
            *)
                if [[ "$1" =~ ^[0-9]+$ ]]; then
                    COUNT="$1"
                else
                    MODEL_NAME="$1"
                fi
                shift
                ;;
        esac
    done
}

parse_args "$@"

# Dynamically load API Key from config/config.yaml if not provided in environment
if [ -z "$API_KEY" ]; then
    if [ -f "$CONFIG_FILE" ]; then
        API_KEY=$(awk '/security:/{f=1} f && /api_key:/{sub(/.*api_key:[ \t]*/, ""); gsub(/^["'\''`]+|["'\''`]+$/, ""); print; exit}' "$CONFIG_FILE")
    fi
fi
API_KEY="${API_KEY:-krea-secret-ai-key-2026}"

echo "=========================================================="
echo " Starting $COUNT Concurrent Requests Benchmark"
echo " Target Endpoint : $URL"
echo " Target Model    : $MODEL_NAME"
echo " Configuration   : $CONFIG_FILE"
echo "=========================================================="

TONES=("professional" "friendly" "concise" "empathetic" "formal" "polite" "urgent")
LANGUAGES=("english" "spanish" "french" "german" "italian")

PAYLOADS=(
  "Please approve purchase order PO-9001 for 10 office monitors."
  "Issue with student enrollment portal. Unable to select course schedule."
  "Server maintenance scheduled for Sunday 2:00 AM UTC. Please notify users."
  "Vendor invoice INV-4402 has been received and verified for payment."
  "Employee onboarding checklist completed for John Doe in Engineering."
  "System backup completed successfully. Total archive size 42GB."
  "Request for password reset on faculty portal account prof_smith."
  "Bug report: Financial report export to Excel hangs on large dataset."
  "Quarterly budget allocation approved for IT Infrastructure department."
  "Support ticket 8841: Wi-Fi connectivity dropping in Library Hall B."
  "Logistics update: Shipment SHIP-7718 dispatched to Central Distribution."
  "HR Alert: Open enrollment for health benefits ends this Friday."
  "Security Audit: Compliance verification complete for SOC2 Type II."
  "Payroll notice: Direct deposit processed for period ending July 15."
  "Facility request: HVAC maintenance needed on 4th floor conference room."
)

RESULTS_FILE=$(mktemp)

run_single_request() {
    ID=$1

    # Valid profile & action pairs based on available prompt templates
    case $((ID % 5)) in
        0) PROFILE="email"; ACTION="rewrite" ;;
        1) PROFILE="email"; ACTION="summarize" ;;
        2) PROFILE="email"; ACTION="translate" ;;
        3) PROFILE="inline_text"; ACTION="rewrite" ;;
        4) PROFILE="support_ticket"; ACTION="rewrite" ;;
    esac

    T_IDX=$(( (ID % ${#TONES[@]}) ))
    L_IDX=$(( (ID % ${#LANGUAGES[@]}) ))
    TEXT_IDX=$(( (ID % ${#PAYLOADS[@]}) ))

    TONE="${TONES[$T_IDX]}"
    LANG="${LANGUAGES[$L_IDX]}"
    TEXT="${PAYLOADS[$TEXT_IDX]} (Req ID #$ID)"

    RESP=$(curl -s -X POST "$URL" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
        "profile": "'"$PROFILE"'",
        "action": "'"$ACTION"'",
        "tone": "'"$TONE"'",
        "language": "'"$LANG"'",
        "text": "'"$TEXT"'",
        "options": {
          "model": "'"$MODEL_NAME"'",
          "signature": "Department Unit #'"$ID"'"
        }
      }')

    SUCCESS=$(echo "$RESP" | grep -o '"success":true' || echo "failed")
    TIME_MS=$(echo "$RESP" | grep -o '"processing_ms":[0-9]*' | cut -d':' -f2 || echo "0")
    
    if [ "$SUCCESS" = '"success":true' ]; then
        echo "SUCCESS $TIME_MS" >> "$RESULTS_FILE"
        echo "[Req #$ID | $PROFILE | $ACTION] Status: SUCCESS | Latency: ${TIME_MS}ms"
    else
        echo "FAILED 0" >> "$RESULTS_FILE"
        echo "[Req #$ID | $PROFILE | $ACTION] Status: FAILED"
    fi
}

export -f run_single_request
export URL API_KEY MODEL_NAME TONES LANGUAGES PAYLOADS RESULTS_FILE

START_TIME=$(date +%s)

# Spawn N parallel background processes
for ((i=1; i<=COUNT; i++)); do
    run_single_request $i &
done

# Wait for all N child processes to finish
wait

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

SUCCESS_COUNT=$(grep -c "^SUCCESS" "$RESULTS_FILE" 2>/dev/null | tr -d '\r\n' || true)
FAILED_COUNT=$(grep -c "^FAILED" "$RESULTS_FILE" 2>/dev/null | tr -d '\r\n' || true)

SUCCESS_COUNT="${SUCCESS_COUNT:-0}"
FAILED_COUNT="${FAILED_COUNT:-0}"

rm -f "$RESULTS_FILE"

echo "=========================================================="
echo " Benchmark Summary"
echo "=========================================================="
echo " Total Concurrent Requests Launched : $COUNT"
echo " Successful Requests                : $SUCCESS_COUNT"
echo " Failed Requests                    : $FAILED_COUNT"
echo " Total Wall-Clock Time              : ${TOTAL_DURATION} seconds"
echo "=========================================================="
