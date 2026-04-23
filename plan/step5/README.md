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
