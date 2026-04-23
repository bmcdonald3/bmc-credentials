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

Fabrica uses declarative resources. 

Example Resource Implementation:
type UserSpec struct {
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"oneof=admin user guest"`
}

type UserStatus struct {
    LastLogin  *time.Time `json:"lastLogin,omitempty"`
    LoginCount int        `json:"loginCount"`
    Health     string     `json:"health" validate:"oneof=healthy degraded unhealthy"`
}
