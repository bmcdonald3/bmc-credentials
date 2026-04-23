## Domain specific logic

### Step outcome

A final plan for implementing reconciliation logic.

### What to give

Reconciler plan from step 5.

### Prompt

Here is the reconciler plan:
### BMC Reconciliation Logic Design

**1. Triggering Condition**
The background reconciliation process is initiated when the controller receives the `io.fabrica.bmc.updated` CloudEvent from the message broker, indicating a new or updated `Bmc` resource exists in the database.

**2. Required External Interactions**
The reconciler executes two distinct HTTP `PATCH` requests to the physical BMC to align the desired state. 

**Authentication Alignment:**
* **HTTP Method:** `PATCH`
* **URI Path:** `/redfish/v1/AccountService/Accounts/1`
* **JSON Payload:**
    ```json
    {
      "UserName": "<Bmc.Spec.DesiredAuth.Username>",
      "Password": "<Bmc.Spec.DesiredAuth.Password>",
      "Enabled": true,
      "RoleId": "Administrator"
    }
    ```
    *(Note: The `Enabled` and `RoleId` fields are mandatory to ensure the applied credentials can actually interact with the API, as the default state for `Accounts/1` was shown as disabled with `NoAccess`.)*

**Network Interface Alignment:**
* **HTTP Method:** `PATCH`
* **URI Path:** `/redfish/v1/Managers/BMC/EthernetInterfaces/1`
* **JSON Payload:**
    ```json
    {
      "DHCPv4": {
        "DHCPEnabled": false
      },
      "IPv4StaticAddresses": [
        {
          "Address": "<Bmc.Spec.TargetIP>"
        }
      ]
    }
    ```
    *(Note: The Redfish spec often requires `SubnetMask` and `Gateway` to be submitted alongside `Address` within the `IPv4StaticAddresses` array. The reconciler must extract the current `SubnetMask` and `Gateway` from the existing `IPv4Addresses` array during the idempotency check and inject them into this payload.)*

**3. Idempotency Condition**
* **Fast-Path Check:** Upon invocation, evaluate `Bmc.Status.Phase` in the database. If the value is exactly `Aligned`, the reconciler terminates immediately without executing any network calls.
* **Deep-Check Verification:** 1. Issue an HTTP `GET` request to `/redfish/v1/Managers/BMC/EthernetInterfaces/1` using the credentials specified in `Bmc.Spec.DesiredAuth`. 
    2. If the BMC returns HTTP `401 Unauthorized`, the authentication state is unaligned. Proceed to the Authentication Alignment interaction using `Bmc.Spec.CurrentAuth` to gain access.
    3. If the BMC returns HTTP `200 OK`, the authentication state is aligned. Parse the JSON response body.
    4. Iterate over the `IPv4Addresses` array in the response. If any `Address` field perfectly matches `Bmc.Spec.TargetIP`, the external system is already in the desired state. Update the database to reflect `Aligned` and terminate. If no match is found, proceed to the Network Interface Alignment interaction.

**4. Progressive and Terminal States**
Every status update requires an immediate write operation to the storage client before proceeding to the next step.
* **Progressive State:** Prior to initiating the Deep-Check Verification, update `Status.Phase` to `Reconciling` and record the current timestamp in `Status.LastUpdated`.
* **Terminal State (Success):** Once both the credentials and network modifications return HTTP `200 OK` or `204 No Content` from the `PATCH` operations (or if the deep-check proves no modifications are necessary), update `Status.Phase` to `Aligned`, write an empty string to `Status.Message`, and update `Status.LastUpdated`.
* **Terminal State (Failure):** If the initial connection times out, if `CurrentAuth` fails during the initial connection attempt, or if a `PATCH` request returns an error (e.g., HTTP `400 Bad Request` or `500 Internal Server Error`), update `Status.Phase` to `Failed`, write the exact HTTP status code and response body string to `Status.Message`, and update `Status.LastUpdated`.

Generate the Go code for the reconcilers designed in the previous step. The code should be formatted to fit within the `reconcile[Resource]` methods located in the generated `pkg/reconcilers/[resource]_reconciler.go` files. Ensure the code includes idempotency checks, calls to `r.Client.Update` to save status changes, and appropriate error returns to trigger requeuing on failure.

### Context

The custom logic must be implemented in the safe-to-edit user stub files (pkg/reconcilers/<resource>_reconciler.go). The generated orchestration wrapper handles event ingestion and requeuing.

```go
// Example Reconciler Implementation
func (r *RackReconciler) reconcileRack(ctx context.Context, res *rack.Rack) error {
    // 1. Idempotency Check
    if res.Status.Phase == "Ready" {
        return nil
    }

    // 2. Perform domain logic (e.g. create child resources)
    template := r.loadTemplate(ctx, res.Spec.TemplateUID)
    
    // 3. Update Status
    res.Status.Phase = "Ready"
    res.Status.TotalChassis = template.Spec.ChassisCount
    
    // 4. Save state
    if err := r.Client.Update(ctx, res); err != nil {
        return fmt.Errorf("failed to update status: %w", err)
    }

    return nil
}
```

Return 'nil' to stop the loop, or return an 'error' to trigger an exponential backoff retry.
