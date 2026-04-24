#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8081"
TARGET_IP="127.0.0.1"
PASS=0
FAIL=0

ok()   { echo "[PASS] $1"; ((PASS++)) || true; }
fail() { echo "[FAIL] $1"; ((FAIL++)) || true; }

echo "Starting Bmc resource validation..."
TEST_NAME="test-bmc-target-$(date +%s)"

# Phase 1: Create Resource
echo "Submitting payload..."
set +e
HTTP_STATUS=$(curl -s -o /tmp/create_resp.json -w "%{http_code}" -X POST "$BASE_URL/bmcs/" -H "Content-Type: application/json" -d "{\"apiVersion\":\"fabrica.dev/v1\",\"kind\":\"Bmc\",\"metadata\":{\"name\":\"$TEST_NAME\"},\"spec\":{\"targetIdentifier\":\"bmc-chassis-A\",\"targetIp\":\"$TARGET_IP\",\"currentAuth\":{\"username\":\"root\",\"password\":\"initial0\"},\"desiredAuth\":{\"username\":\"anonymous\",\"password\":\"test_pass\"}}}")
set -e

if [ "$HTTP_STATUS" = "201" ] || [ "$HTTP_STATUS" = "200" ]; then
  BMC_UID=$(cat /tmp/create_resp.json | python3 -c "import sys,json; print(json.load(sys.stdin).get('metadata', {}).get('uid', ''))" 2>/dev/null || echo "")
  ok "Bmc resource created (uid=$BMC_UID)"
else
  fail "Bmc resource creation failed with HTTP $HTTP_STATUS"
  cat /tmp/create_resp.json
  exit 1
fi

# Phase 2: Verify Internal Fabrica Status
MAX_RETRIES=10
SLEEP_INTERVAL=3
FAILED=false

echo "Polling Bmc resource status for expected 'Failed' state (Max time: 30 seconds)..."

for ((i=1; i<=MAX_RETRIES; i++)); do
  READ_RESP=$(curl -sf "$BASE_URL/bmcs/$BMC_UID/")
  PHASE=$(echo "$READ_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('status', {}).get('phase', 'Pending'))" 2>/dev/null || echo "Pending")
  MESSAGE=$(echo "$READ_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('status', {}).get('message', ''))" 2>/dev/null || echo "")

  if [ "$PHASE" == "Failed" ]; then
    if echo "$MESSAGE" | grep -Eiq "timeout|connection refused|no route to host|i/o timeout"; then
      FAILED=true
      ok "Bmc reconciliation attempted network call and failed as expected: $MESSAGE"
      break
    fi
    fail "Bmc transitioned to Failed but message did not contain expected network error"
    echo "Failure Message: $MESSAGE"
    exit 1
  elif [ "$PHASE" == "Aligned" ]; then
    fail "Bmc reconciliation unexpectedly reached 'Aligned' in non-routable environment."
    break
  fi

  echo "Attempt $i/$MAX_RETRIES: Current phase is '$PHASE'. Waiting $SLEEP_INTERVAL seconds..."
  sleep "$SLEEP_INTERVAL"
done

if [ "$FAILED" = true ]; then
  ok "Bmc resource transitioned to expected 'Failed' phase"
else
  fail "Bmc resource did not reach expected 'Failed' phase within 30 seconds (Final phase: $PHASE)"
  exit 1
fi

echo ""
echo "Test Execution Complete. PASS: $PASS, FAIL: $FAIL"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi