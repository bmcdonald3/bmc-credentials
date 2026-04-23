#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

ok()   { echo "[PASS] $1"; ((PASS++)) || true; }
fail() { echo "[FAIL] $1"; ((FAIL++)) || true; }

echo "Starting Bmc resource validation..."

# Phase 1: Create Resource
# Submits the desired state manifest to the API server.
echo "Submitting payload..."
set +e
HTTP_STATUS=$(curl -s -o /tmp/create_resp.json -w "%{http_code}" -X POST "$BASE_URL/bmcs" -H "Content-Type: application/json" -d '{"apiVersion":"fabrica.dev/v1","kind":"Bmc","metadata":{"name":"test-bmc-target"},"spec":{"targetIdentifier":"bmc-chassis-A","targetIp":"10.0.0.55","currentAuth":{"username":"admin","password":"old_password"},"desiredAuth":{"username":"admin","password":"new_password"}}}')
set -e

if [ "$HTTP_STATUS" = "201" ] || [ "$HTTP_STATUS" = "200" ]; then
  BMC_UID=$(cat /tmp/create_resp.json | python3 -c "import sys,json; print(json.load(sys.stdin).get('metadata', {}).get('uid', ''))" 2>/dev/null || echo "")
  ok "Bmc resource created (uid=$BMC_UID)"
else
  fail "Bmc resource creation failed with HTTP $HTTP_STATUS"
  cat /tmp/create_resp.json
  exit 1
fi

# Phase 2: Verify Status
# Poll the GET endpoint to monitor the background controller's progress.
MAX_RETRIES=10
SLEEP_INTERVAL=3
ALIGNED=false

echo "Polling Bmc resource status for 'Aligned' state (Max time: 30 seconds)..."

for ((i=1; i<=MAX_RETRIES; i++)); do
  READ_RESP=$(curl -sf "$BASE_URL/bmcs/$BMC_UID")
  PHASE=$(echo "$READ_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('status', {}).get('phase', 'Pending'))" 2>/dev/null || echo "Pending")

  if [ "$PHASE" == "Aligned" ]; then
    ALIGNED=true
    break
  elif [ "$PHASE" == "Failed" ]; then
    fail "Bmc reconciliation failed. Status phase transitioned to 'Failed'."
    exit 1
  fi

  echo "Attempt $i/$MAX_RETRIES: Current phase is '$PHASE'. Waiting $SLEEP_INTERVAL seconds..."
  sleep "$SLEEP_INTERVAL"
done

if [ "$ALIGNED" = true ]; then
  ok "Bmc resource successfully transitioned to 'Aligned' phase"
else
  fail "Bmc resource did not reach 'Aligned' phase within 30 seconds (Final phase: $PHASE)"
fi

echo ""
echo "Test Execution Complete. PASS: $PASS, FAIL: $FAIL"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi