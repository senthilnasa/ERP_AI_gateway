#!/usr/bin/env bash

# Stress test 50 concurrent requests with different payloads against AI Gateway

URL="http://localhost:8080/api/v1/write"
API_KEY="krea-secret-ai-key-2026"
MODEL_NAME="${1:-qwen2.5:0.5b}"

echo "=========================================="
echo " Starting 50 Concurrent Requests Stress Test"
echo " Target Model: $MODEL_NAME"
echo "=========================================="

TONES=("professional" "friendly" "concise" "empathetic" "formal" "polite")
LANGUAGES=("english" "spanish" "french" "german")

# 50 unique text payloads for ERP modules
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
)

START_TIME=$(date +%s)

run_single_request() {
    ID=$1
    
    # Cycle through profiles and actions validly
    if [ $((ID % 3)) -eq 0 ]; then
        PROFILE="email"
        ACTION="rewrite"
    elif [ $((ID % 3)) -eq 1 ]; then
        PROFILE="email"
        ACTION="summarize"
    else
        PROFILE="inline_text"
        ACTION="rewrite"
    fi

    T_IDX=$(( (ID % 6) ))
    L_IDX=$(( (ID % 4) ))
    TEXT_IDX=$(( (ID % 10) ))

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
          "signature": "Department #'"$ID"'"
        }
      }')

    SUCCESS=$(echo "$RESP" | grep -o '"success":true' || echo "failed")
    TIME_MS=$(echo "$RESP" | grep -o '"processing_ms":[0-9]*' | cut -d':' -f2 || echo "0")
    echo "[Req #$ID | $PROFILE | $ACTION] Status: $SUCCESS | Latency: ${TIME_MS}ms"
}

export -f run_single_request
export URL API_KEY MODEL_NAME TONES LANGUAGES PAYLOADS

# Launch 50 concurrent requests simultaneously
for i in {1..50}; do
    run_single_request $i &
done

# Wait for all 50 parallel requests to finish
wait

END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

echo "=========================================="
echo " 50 Concurrent Requests Test Complete!"
echo " Total Wall-Clock Time: ${TOTAL_DURATION} seconds."
echo "=========================================="
