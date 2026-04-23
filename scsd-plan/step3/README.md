## Build Fabrica workflow

### Step outcome

Test script of all Fabrica commands and expected output to validate service completion.

### What to give

The defined resource models from step 2.

### Prompt

Write a validation bash script using `curl` to simulate the user workflow. The script must create the necessary resources by sending POST requests and then verify that the system processes them by checking their `status` fields. Use the provided sample validation script as a reference for formatting. 

### Context

Fabrica generates standard CRUD endpoints for all defined resources. 
- Create: POST /<resource-name-plural> (Requires "apiVersion", "kind", "metadata.name", and "spec")
- Read: GET /<resource-name-plural>/<uid>
- List: GET /<resource-name-plural>
The background controller will update the resource. The validation script must poll the GET endpoint to check if the 'status.phase' field transitions to the expected terminal state (e.g., 'Ready' or 'Failed').

# Sample validation script structure
```bash
#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

ok()   { echo "[PASS] $1"; ((PASS++)) || true; }
fail() { echo "[FAIL] $1"; ((FAIL++)) || true; }

# Start server logic...

# Phase 1: Create Resource
JOB_RESP=$(curl -sf -X POST "$BASE_URL/updatejobs" -H "Content-Type: application/json" -d '{"apiVersion":"example.fabrica.dev/v1","kind":"UpdateJob","metadata":{"name":"e2e-job"},"spec":{"targetNodes":["nodeA","nodeB"],"firmwareRef":"test-fw"}}')

JOB_UID=$(echo "$JOB_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['metadata']['uid'])" 2>/dev/null || echo "")

if [ -n "$JOB_UID" ]; then
  ok "UpdateJob created (uid=$JOB_UID)"
else
  fail "UpdateJob creation failed"
fi

# Phase 2: Verify Status
sleep 5
TASKS_RESP=$(curl -sf "$BASE_URL/updatetasks")
TASK_COUNT=$(echo "$TASKS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if isinstance(d,list) else 0)" 2>/dev/null || echo 0)

if [ "$TASK_COUNT" -ge 2 ]; then
  ok "UpdateTasks created"
else
  fail "Expected tasks not found"
fi
```

============

================

```bash
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
CREATE_RESP=$(curl -sf -X POST "$BASE_URL/bmcs" -H "Content-Type: application/json" -d '{"apiVersion":"fabrica.dev/v1","kind":"Bmc","metadata":{"name":"test-bmc-target"},"spec":{"targetIdentifier":"bmc-chassis-A","targetIp":"10.0.0.55","currentAuth":{"username":"admin","password":"old_password"},"desiredAuth":{"username":"admin","password":"new_password"}}}')

# Extract UID using python3 to parse the JSON response.
BMC_UID=$(echo "$CREATE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('metadata', {}).get('uid', ''))" 2>/dev/null || echo "")

if [ -n "$BMC_UID" ]; then
  ok "Bmc resource created (uid=$BMC_UID)"
else
  fail "Bmc resource creation failed"
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
```
