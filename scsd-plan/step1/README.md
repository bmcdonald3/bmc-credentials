## Define workflow in plain english

### Step outcome

A well-defined picture of what you actually want to do.

### What to give

Describe in as much detail as possible how a user would interact with this service.

### Prompt

Act as a system architect. I am designing a new service. I will describe how a user interacts with this service. Please analyze my description and output a numbered, step-by-step workflow. The workflow must differentiate between explicit user actions (e.g., submitting a payload) and the required background system operations (e.g., updating states, triggering downstream events). If the description lacks sufficient detail to determine the background operations, ask me for clarification.xs

### Context

This service is built using Fabrica, a framework for generating Kubernetes-style declarative APIs. 
Architecture Overview:
1. Declarative Design: Users do not trigger imperative actions. Instead, they declare a "Desired State" by creating or updating a JSON resource via a REST API.
2. Asynchronous Processing: Creating or modifying a resource publishes a CloudEvent (e.g., io.fabrica.resource.created).
3. Reconciliation Controller: A background worker receives the event, compares the "Current State" of the system to the "Desired State", and executes business logic to align them.
When analyzing the user's workflow description, you must separate it into two distinct phases: 
Phase A: What desired state the user submits.
Phase B: What the background reconciliation loop must do asynchronously to fulfill that state.

-----------------

For this service, the workflow is as follows: user has a file that is a mapping between all BMCs that are desired to change and the IPs and the desired credentials to update those BMCs for username and password. Our service would already have a list of all known BMCs in the database. If they need the current credentials to update them, that would also be placed in the file. The user would then make a call to our service, possibly posting the file, which would trigger a reconciler to update the spec of all BMCs to change, then the reconciler would go and do the actual update.

===================

Phase A: Desired State Submission (User Actions)
1. The user creates a declarative JSON or YAML manifest for a single Bmc API resource. For the prototype, this payload directly includes the target BMC identifier, the target IP, the current credentials (needed for authentication), and the new desired credentials.
2. The user submits this payload to the service using a standard HTTP client (e.g., a curl PUT or PATCH request to /apis/v1/bmcs/{bmc-id}).
3. The API server validates the payload against the Bmc resource schema.
4. The API server persists this new "Desired State" in the database.
5. The API server returns an HTTP 202 Accepted or 200 OK to the user, acknowledging the state has been recorded.

Phase B: Asynchronous Reconciliation (Background Operations)
6. The API server publishes a CloudEvent (e.g., io.fabrica.bmc.updated) to the message broker.
7. The Reconciliation Controller receives the event and retrieves the full Bmc resource, including the plaintext credentials, from the database.
8. The controller extracts the current credentials from the Desired State and uses them to authenticate against the physical BMC's management interface (e.g., via Redfish API).
9. The controller executes the business logic to apply the target IP and new credentials to the hardware.
10. The controller receives the response from the hardware.
11. The controller updates the Status sub-resource of the Bmc object in the database to reflect the operational outcome (e.g., State: Aligned, State: Failed, along with any error details).