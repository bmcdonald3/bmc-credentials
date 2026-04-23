## Define workflow in plain english

### Step outcome

A well-defined picture of what you actually want to do.

### What to give

Describe in as much detail as possible how a user would interact with this service.

### Prompt

Act as a system architect. I am designing a new service. I will describe how a user interacts with this service. Please analyze my description and output a numbered, step-by-step workflow. The workflow must differentiate between explicit user actions (e.g., submitting a payload) and the required background system operations (e.g., updating states, triggering downstream events). If the description lacks sufficient detail to determine the background operations, ask me for clarification.

### Context

This service is built using Fabrica, a framework for generating Kubernetes-style declarative APIs. 
Architecture Overview:
1. Declarative Design: Users do not trigger imperative actions. Instead, they declare a "Desired State" by creating or updating a JSON resource via a REST API.
2. Asynchronous Processing: Creating or modifying a resource publishes a CloudEvent (e.g., io.fabrica.resource.created).
3. Reconciliation Controller: A background worker receives the event, compares the "Current State" of the system to the "Desired State", and executes business logic to align them.
When analyzing the user's workflow description, you must separate it into two distinct phases: 
Phase A: What desired state the user submits.
Phase B: What the background reconciliation loop must do asynchronously to fulfill that state.

[insert service description here]
