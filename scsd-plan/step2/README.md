## Define Fabrica resources

### Step outcome

Resource models split into spec/status fields.

### What to give

The finalized plain English workflow from Step 1.

### Prompt

Using the workflow we defined, identify the required API resources. For each resource, generate the Go structs defining the schema. You must follow the Kubernetes-style resource pattern by splitting the data into two components:
1. `Spec`: The desired state provided by the user.
2. `Status`: The observed state managed by the system in the background.
If resources are hierarchical, use UID string fields to link child resources to their parents.

### Context

Fabrica resources strictly separate data into 'Spec' and 'Status'.
1. Spec: The desired state provided by the user. Must use 'validate' struct tags for input validation.
2. Status: The observed state managed exclusively by the system's reconciliation loop.
3. Relationships: Hierarchical resources are linked using UID strings (e.g., a child resource stores its parent's UID in its Spec, and the parent tracks created child UIDs in its Status).

Example Resource Implementation:
```go
type UserSpec struct {
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"oneof=admin user guest"`
    ParentTeamUID string `json:"parentTeamUid,omitempty"`
}

type UserStatus struct {
    Phase      string     `json:"phase" validate:"oneof=Pending Provisioning Ready Error"`
    Message    string     `json:"message,omitempty"`
    LastLogin  *time.Time `json:"lastLogin,omitempty"`
}
```

======================

```go
package v1

import (
	"time"
)

// Bmc represents the complete Kubernetes-style resource.
type Bmc struct {
	UID    string     `json:"uid" validate:"required"`
	Spec   BmcSpec    `json:"spec" validate:"required"`
	Status BmcStatus  `json:"status,omitempty"`
}

// Credentials defines the authentication payload required to connect to the hardware.
type Credentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// BmcSpec represents the desired state provided by the user via the API payload.
type BmcSpec struct {
	TargetIdentifier string      `json:"targetIdentifier" validate:"required"`
	TargetIP         string      `json:"targetIp" validate:"required,ip"`
	CurrentAuth      Credentials `json:"currentAuth" validate:"required"`
	DesiredAuth      Credentials `json:"desiredAuth" validate:"required"`
}

// BmcStatus represents the observed state managed exclusively by the asynchronous Reconciliation Controller.
type BmcStatus struct {
	Phase       string     `json:"phase" validate:"oneof=Pending Reconciling Aligned Failed"`
	Message     string     `json:"message,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
}
```

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