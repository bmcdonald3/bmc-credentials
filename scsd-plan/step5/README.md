## Domain specific logic

### Step outcome

A clear picture of what the reconiliation logic will look like.

### What to give

Any domain-specific rules, formulas, external API endpoints, or hardware interactions the service must perform.

### Prompt

Your task is to design the reconciliation logic for the background operations identified in the workflow. For each resource that requires background processing, document the following:
1. Triggering condition: What specific event or state change initiates this logic.
2. Required external interactions: You MUST define the exact technical implementation. If interacting with an external API, provide the exact HTTP methods, precise URI paths, and the literal JSON payload structures required. Do not generalize (e.g., do not say "make an API call"). If creating child resources, define the exact resource payload.
3. Idempotency condition: How to fast-path check if the logic has already been applied, and how to deep-check the external system to verify actual state.
4. Terminal states: The exact string values applied to the `Status` fields upon success or failure, and the progressive states applied during execution.

### Context

The business logic resides in Fabrica Reconcilers. You must design this logic adhering to these constraints:
1. Idempotency: The reconciler may be called multiple times for the same event. It must check the 'Status.Phase' first and return immediately if the work is already done.
2. Progressive Updates: The reconciler should update 'Status.Phase' to intermediate states (e.g., "Provisioning") before starting long-running tasks.
3. State Storage: Any changes to the resource status must be explicitly saved via the storage client before the function returns.

=================

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