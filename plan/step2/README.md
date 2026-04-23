## Define Fabrica resources

### Step outcome

Resource models split into spec/status fields.

### What to give

The finalized plain English workflow from Step 1.

### Prompt

Using the workflow we defined, identify the required API resources. For each resource, generate the Go structs defining the schema. You must follow the Kubernetes-style resource pattern by splitting the data into two components:
1. `Spec`: The desired state provided by the user.
2. `Status`: The observed state managed by the system in the background.
If resources are hierarchical, use UID string fields to link child resources to their parents.

### Context

Fabrica resources strictly separate data into 'Spec' and 'Status'.
1. Spec: The desired state provided by the user. Must use 'validate' struct tags for input validation.
2. Status: The observed state managed exclusively by the system's reconciliation loop.
3. Relationships: Hierarchical resources are linked using UID strings (e.g., a child resource stores its parent's UID in its Spec, and the parent tracks created child UIDs in its Status).

Example Resource Implementation:

```go
type UserSpec struct {
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"oneof=admin user guest"`
    ParentTeamUID string `json:"parentTeamUid,omitempty"`
}

type UserStatus struct {
    Phase      string     `json:"phase" validate:"oneof=Pending Provisioning Ready Error"`
    Message    string     `json:"message,omitempty"`
    LastLogin  *time.Time `json:"lastLogin,omitempty"`
}
```