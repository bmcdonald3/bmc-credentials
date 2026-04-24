## Build Fabrica workflow

### Step outcome

Test script of all Fabrica commands and expected output to validate service completion.

### What to give

The defined resource models from step 2.

### Prompt

Write a validation bash script using `curl` to simulate the user workflow. The script must create the necessary resources by sending POST requests and then verify that the system processes them by checking their `status` fields. Crucially, the script MUST then perform an external validation step: it must issue direct `curl` commands against the external system (e.g., the hardware or third-party API) to functionally prove that the side-effects were actually applied.

Constraint: The execution environment does not have network access to the physical hardware. Do not attempt external hardware validation. The script must validate that the Fabrica server compiles, starts, accepts the resource payload, and that the reconciliation controller attempts the background process. 
Constraint: Because the hardware is unreachable, the script must assert that the 'status.phase' eventually transitions to 'Failed' and that 'status.message' contains a network timeout or connection refused error. This will prove the network logic was implemented.
Constraint: You must wrap the resource creation `curl` commands with `set +e` and `set -e` to prevent the script from aborting on failure. You must capture the HTTP status code and print the full JSON response if the creation fails to assist with debugging.

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