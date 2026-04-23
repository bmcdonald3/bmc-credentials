## Domain specific logic

### Step outcome

A final plan for implementing reconciliation logic.

### What to give

Reconciler plan from step 5.

### Prompt

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
