## Domain specific logic

### Step outcome

A clear picture of what the reconiliation logic will look like.

### What to give

Any domain-specific rules, formulas, external API endpoints, or hardware interactions the service must perform.

### Prompt

Your task is to define the reconciliation logic for the background operations identified in the workflow. For each resource that requires background processing, document the following:
1. Triggering condition: What specific event or state change initiates this logic.
2. Required external interactions: You MUST define the exact technical implementation. If interacting with an external API, provide the exact HTTP methods, precise URI paths, and the literal JSON payload structures required. Do not generalize (e.g., do not say "make an API call"). If creating child resources, define the exact resource payload.
3. Idempotency condition: How to fast-path check if the logic has already been applied, and how to deep-check the external system to verify actual state.
4. Terminal states: The exact string values applied to the `Status` fields upon success or failure, and the progressive states applied during execution.

### Context

The business logic resides in Fabrica Reconcilers. You must design this logic adhering to these constraints:
1. Idempotency: The reconciler may be called multiple times for the same event. It must check the 'Status.Phase' first and return immediately if the work is already done.
2. Progressive Updates: The reconciler should update 'Status.Phase' to intermediate states (e.g., "Provisioning") before starting long-running tasks.
3. State Storage: Any changes to the resource status must be explicitly saved via the storage client before the function returns.
