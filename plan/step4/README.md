## Bootstrap service

### Step outcome

A script to initialize Fabrica project with all resources and testing to ensure it is working.

### What to give

Defined go structs from step 2.

### Prompt

Generate a bash bootstrap script to initialize the Fabrica project. The script must use the `fabrica init` command to create the project with event-driven reconciliation enabled. It must then use `fabrica add resource` for each of our identified resources. Following the resource creation, the script must write our custom Go structs into the generated files in the `apis/` directory, overriding the default types. Finally, the script must run `fabrica generate`. 

### Context

Sample bootstrap file output

```bash
#!/bin/bash
set -e

PROJECT_NAME="fms"
MODULE_NAME="github.com/bmcdonald3/fms"
GROUP="example.fabrica.dev"
API_VERSION="v1"
API_DIR="apis/$GROUP/$API_VERSION"

rm -rf $PROJECT_NAME

fabrica init $PROJECT_NAME --module $MODULE_NAME --storage-type ent --db sqlite --events --events-bus memory --reconcile

cd $PROJECT_NAME

for res in DeviceProfile FirmwareProfile; do fabrica add resource $res; done

cat << 'EOF' > $API_DIR/deviceprofile_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type DeviceProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       DeviceProfileSpec `json:"spec"`
	Status     DeviceProfileStatus `json:"status,omitempty"`
}
type DeviceProfileSpec struct {
	Manufacturer string `json:"manufacturer" validate:"required"`
	Model        string `json:"model,omitempty"`
	RedfishPath  string `json:"redfishPath,omitempty"`
	ManagementIp string `json:"managementIp,omitempty"`
}
type DeviceProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *DeviceProfile) GetKind() string { return "DeviceProfile" }
func (r *DeviceProfile) GetName() string { return r.Metadata.Name }
func (r *DeviceProfile) GetUID() string { return r.Metadata.UID }
func (r *DeviceProfile) IsHub() {}
EOF

cat << 'EOF' > $API_DIR/firmwareprofile_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type FirmwareProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       FirmwareProfileSpec `json:"spec"`
	Status     FirmwareProfileStatus `json:"status,omitempty"`
}
type FirmwareProfileSpec struct {
	VersionString   string `json:"versionString" validate:"required"`
	VersionNumber   string `json:"versionNumber" validate:"required"`
	TargetComponent string `json:"targetComponent" validate:"required"`
	PreConditions   string `json:"preConditions,omitempty"`
	PostConditions  string `json:"postConditions,omitempty"`
	SoftwareId      string `json:"softwareId,omitempty"`
}
type FirmwareProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *FirmwareProfile) GetKind() string { return "FirmwareProfile" }
func (r *FirmwareProfile) GetName() string { return r.Metadata.Name }
func (r *FirmwareProfile) GetUID() string { return r.Metadata.UID }
func (r *FirmwareProfile) IsHub() {}
EOF

fabrica generate

go mod tidy

cat << 'EOF' > pkg/reconcilers/deviceprofile_reconciler.go
package reconcilers
import (
	"context"
	"fmt"
	v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
)
func (r *DeviceProfileReconciler) reconcileDeviceProfile(ctx context.Context, res *v1.DeviceProfile) error {
	if res.Status.Phase == "Reconciliation Proved!" {
		return nil
	}
	res.Status.Phase = "Reconciliation Proved!"
	if err := r.Client.Update(ctx, res); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	r.Logger.Infof("Successfully verified event-driven loop for DeviceProfile %s", res.GetUID())
	return nil
}
EOF

go run ./cmd/server serve --database-url="file:data.db?cache=shared&_fk=1&_busy_timeout=10000" > server.log 2>&1 &
SERVER_PID=$!

sleep 10

set +e

JOB_RESP=$(curl -s -f -X POST http://localhost:8080/deviceprofiles -H "Content-Type: application/json" -d '{"metadata": {"name": "test-job"}, "spec": {"manufacturer": ["HPE"]}}')

CURL_STATUS=$?

if [ $CURL_STATUS -ne 0 ]; then echo "Failed to connect to the server. Checking logs:"; cat server.log; kill $SERVER_PID 2>/dev/null; exit 1; fi

set -e

JOB_UID=$(echo $JOB_RESP | grep -o '"uid":"[^"]*"' | head -1 | cut -d'"' -f4)

sleep 5

STATUS_RESP=$(curl -s http://localhost:8080/deviceprofiles/$JOB_UID)

PHASE=$(echo $STATUS_RESP | grep -o '"phase":"[^"]*"' | cut -d'"' -f4)

if [ "$PHASE" = "Reconciliation Proved!" ]; then echo "SUCCESS: The event bus and controller successfully executed the reconciler logic."; else echo "FAILURE: The reconciliation loop did not modify the resource. Phase is: $PHASE"; cat server.log; fi

kill $SERVER_PID
```
