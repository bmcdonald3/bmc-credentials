## Bootstrap service

### Step outcome

A script to initialize Fabrica project with all resources and testing to ensure it is working.

### What to give

Resource model, below script.

### Prompt

...

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

for res in DeviceProfile FirmwareProfile UpdateProfile UpdateJob UpdateTask LookupJob; do fabrica add resource $res; done

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
	Model        string `json:"model" validate:"required"`
	RedfishPath  string `json:"redfishPath" validate:"required"`
	ManagementIp string `json:"managementIp" validate:"required"`
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

cat << 'EOF' > $API_DIR/updateprofile_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateProfile struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateProfileSpec `json:"spec"`
	Status     UpdateProfileStatus `json:"status,omitempty"`
}
type UpdateProfileSpec struct {
	CommandType     string `json:"commandType" validate:"required,oneof=Redfish SSH"`
	PayloadPath     string `json:"payloadPath,omitempty"`
	SuccessCriteria string `json:"successCriteria,omitempty"`
}
type UpdateProfileStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *UpdateProfile) GetKind() string { return "UpdateProfile" }
func (r *UpdateProfile) GetName() string { return r.Metadata.Name }
func (r *UpdateProfile) GetUID() string { return r.Metadata.UID }
func (r *UpdateProfile) IsHub() {}
EOF

cat << 'EOF' > $API_DIR/updatejob_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateJob struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateJobSpec `json:"spec"`
	Status     UpdateJobStatus `json:"status,omitempty"`
}
type UpdateJobSpec struct {
	TargetNodes     []string `json:"targetNodes" validate:"required"`
	TargetComponent string   `json:"targetComponent,omitempty"`
	FirmwareRef     string   `json:"firmwareRef,omitempty"`
	DryRun          bool     `json:"dryRun,omitempty"`
	Force           bool     `json:"force,omitempty"`
}
type UpdateJobStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}
func (r *UpdateJob) GetKind() string { return "UpdateJob" }
func (r *UpdateJob) GetName() string { return r.Metadata.Name }
func (r *UpdateJob) GetUID() string { return r.Metadata.UID }
func (r *UpdateJob) IsHub() {}
EOF

cat << 'EOF' > $API_DIR/updatetask_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type UpdateTask struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       UpdateTaskSpec `json:"spec"`
	Status     UpdateTaskStatus `json:"status,omitempty"`
}
type UpdateTaskSpec struct {
	TargetNode      string `json:"targetNode" validate:"required"`
	TargetComponent string `json:"targetComponent,omitempty"`
	FirmwareRef     string `json:"firmwareRef,omitempty"`
	UpdateJobId     string `json:"updateJobId" validate:"required"`
}
type UpdateTaskStatus struct {
	State     string `json:"state,omitempty" validate:"omitempty,oneof=Pending Running Success Failed"`
	StartTime int64  `json:"startTime,omitempty"`
	TaskUri   string `json:"taskUri,omitempty"`
}
func (r *UpdateTask) GetKind() string { return "UpdateTask" }
func (r *UpdateTask) GetName() string { return r.Metadata.Name }
func (r *UpdateTask) GetUID() string { return r.Metadata.UID }
func (r *UpdateTask) IsHub() {}
EOF

cat << 'EOF' > $API_DIR/lookupjob_types.go
package v1
import "github.com/openchami/fabrica/pkg/fabrica"
type LookupJob struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       LookupJobSpec `json:"spec"`
	Status     LookupJobStatus `json:"status,omitempty"`
}
type LookupJobSpec struct {
	TargetNode string `json:"targetNode" validate:"required"`
}
type LookupJobStatus struct {
	State        string `json:"state,omitempty" validate:"omitempty,oneof=Pending Running Complete Failed"`
	JobId        string `json:"jobId,omitempty"`
	FirmwareData string `json:"firmwareData,omitempty"`
}
func (r *LookupJob) GetKind() string { return "LookupJob" }
func (r *LookupJob) GetName() string { return r.Metadata.Name }
func (r *LookupJob) GetUID() string { return r.Metadata.UID }
func (r *LookupJob) IsHub() {}
EOF

fabrica generate

go mod tidy

cat << 'EOF' > pkg/reconcilers/updatejob_reconciler.go
package reconcilers
import (
	"context"
	"fmt"
	v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
)
func (r *UpdateJobReconciler) reconcileUpdateJob(ctx context.Context, res *v1.UpdateJob) error {
	if res.Status.Phase == "Reconciliation Proved!" {
		return nil
	}
	res.Status.Phase = "Reconciliation Proved!"
	if err := r.Client.Update(ctx, res); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	r.Logger.Infof("Successfully verified event-driven loop for UpdateJob %s", res.GetUID())
	return nil
}
EOF

go run ./cmd/server serve --database-url="file:data.db?cache=shared&_fk=1&_busy_timeout=10000" > server.log 2>&1 &
SERVER_PID=$!

sleep 10

set +e

JOB_RESP=$(curl -s -f -X POST http://localhost:8080/updatejobs -H "Content-Type: application/json" -d '{"metadata": {"name": "test-job"}, "spec": {"targetNodes": ["node1"], "targetComponent": "BIOS", "firmwareRef": "v1"}}')

CURL_STATUS=$?

if [ $CURL_STATUS -ne 0 ]; then echo "Failed to connect to the server. Checking logs:"; cat server.log; kill $SERVER_PID 2>/dev/null; exit 1; fi

set -e

JOB_UID=$(echo $JOB_RESP | grep -o '"uid":"[^"]*"' | head -1 | cut -d'"' -f4)

sleep 5

STATUS_RESP=$(curl -s http://localhost:8080/updatejobs/$JOB_UID)

PHASE=$(echo $STATUS_RESP | grep -o '"phase":"[^"]*"' | cut -d'"' -f4)

if [ "$PHASE" = "Reconciliation Proved!" ]; then echo "SUCCESS: The event bus and controller successfully executed the reconciler logic."; else echo "FAILURE: The reconciliation loop did not modify the resource. Phase is: $PHASE"; cat server.log; fi

kill $SERVER_PID
```
