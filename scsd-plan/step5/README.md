## Domain specific logic

### Step outcome

A clear picture of what the reconiliation logic will look like.

### What to give

Any domain-specific rules, formulas, external API endpoints, or hardware interactions the service must perform.

### Prompt

Design the reconciliation logic for the background operations identified in the workflow. For each resource that requires background processing, document the following:
1. Triggering condition (e.g., what state change initiates this logic).
2. Required external interactions or child resource creation.
3. How to check if the logic has already been applied (idempotency condition).
4. The terminal states applied to the `Status` field upon success or failure.

### Context

The business logic resides in Fabrica Reconcilers. You must design this logic adhering to these constraints:
1. Idempotency: The reconciler may be called multiple times for the same event. It must check the 'Status.Phase' first and return immediately if the work is already done.
2. Progressive Updates: The reconciler should update 'Status.Phase' to intermediate states (e.g., "Provisioning") before starting long-running tasks.
3. State Storage: Any changes to the resource status must be explicitly saved via the storage client before the function returns.

=================

### Reconciliation Logic Design: `Bmc` Resource

**1. Triggering Condition**
The background operation is initiated by the receipt of an asynchronous event from the message broker, specifically the `io.fabrica.bmc.updated` (or corresponding creation) CloudEvent. This event is published by the API server immediately after a user's declarative payload successfully passes schema validation and is persisted as the new desired state in the database. The reconciler consumes this event to begin processing the specific `Bmc` identifier included in the payload.

**2. Required External Interactions**
To align the observed state with the user's desired state, the reconciler must perform specific network and storage interactions. No child Kubernetes-style resources are required for this specific workflow.
* **Database Retrieval:** The reconciler must query the database to retrieve the full `Bmc` object, which contains the unredacted `Spec.CurrentAuth` and `Spec.DesiredAuth` needed for the hardware session.
* **Hardware Interface Execution:** The controller must establish a connection over the network to the physical BMC's management interface (e.g., using the Redfish API) using the `Spec.CurrentAuth` payload. Once authenticated, the controller must issue the necessary HTTP POST/PATCH requests to mutate the physical hardware's IP address to `Spec.TargetIP` and its authentication database to reflect `Spec.DesiredAuth`.
* **Storage Client Persistence:** The reconciler must invoke the storage client to save state updates explicitly back to the database at multiple points during the execution lifecycle.

**3. Idempotency Condition**
The reconciler must ensure safe execution even if the `io.fabrica.bmc.updated` event is delivered multiple times for the same payload.
* **Fast-Path State Check:** Immediately after fetching the `Bmc` object from the database, the reconciler must inspect `Status.Phase`. If the value is strictly `"Aligned"`, the desired state has already been achieved by a previous execution, and the function must return immediately without initiating any network connections to the hardware.
* **Hardware Verification (Deep Idempotency):** If the phase is not `"Aligned"`, the reconciler should verify the actual state of the hardware before mutating it. The system should attempt to authenticate against `Spec.TargetIP` using `Spec.DesiredAuth`. If this connection succeeds, it proves the hardware is already properly configured, allowing the reconciler to bypass the mutation steps, update the phase to `"Aligned"`, and persist the status.

**4. Terminal States and Progressive Updates**
The reconciler must rigorously manage the `Status` sub-resource to provide accurate operational visibility and adhere to the progressive update constraints.
* **Progressive Update:** After clearing the idempotency checks and before establishing the connection to the physical hardware, the reconciler must update `Status.Phase` to `"Reconciling"` and set `Status.LastUpdated` to the current timestamp. This intermediate state must be explicitly saved via the storage client before proceeding to the long-running hardware tasks.
* **Terminal State (Success):** Following a successful `200 OK` or `204 No Content` response from the physical BMC confirming the application of the new IP and credentials, the reconciler must update `Status.Phase` to `"Aligned"`. Any previous error text in `Status.Message` must be cleared, `Status.LastUpdated` must be updated, and the record must be explicitly saved via the storage client.
* **Terminal State (Failure):** If the reconciler encounters a failure condition—such as a network timeout when reaching the BMC, a `401 Unauthorized` when using `Spec.CurrentAuth`, or a `400 Bad Request` during the application of the new configuration—the operation is halted. The reconciler must update `Status.Phase` to `"Failed"`. The `Status.Message` field must be populated with the concrete error string returned by the hardware or network stack, `Status.LastUpdated` must be updated, and the resulting object must be explicitly saved via the storage client.