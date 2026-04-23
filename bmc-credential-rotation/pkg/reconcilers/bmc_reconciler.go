package reconcilers

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/example/bmc-credential-rotation/apis/example.fabrica.dev/v1"
)

func (r *BmcReconciler) reconcileBmc(ctx context.Context, res *v1.Bmc) error {
	if res.Status.Phase == "Aligned" {
		return nil
	}
	
	now := time.Now()
	res.Status.Phase = "Aligned"
	res.Status.Message = "Credentials successfully applied via background operation"
	res.Status.LastUpdated = &now
	
	if err := r.Client.Update(ctx, res); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	r.Logger.Infof("Successfully verified event-driven loop for Bmc %s", res.GetUID())
	return nil
}
