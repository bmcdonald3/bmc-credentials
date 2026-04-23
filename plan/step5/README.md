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