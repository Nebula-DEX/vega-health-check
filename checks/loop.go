package checks

import (
	"context"
	"time"

	"log"
)

const (
	checkInterval = 30 * time.Second
)

func HealthCheckLoop(ctx context.Context, resultChan chan<- Result, checks []HealthCheckFunc) error {

	tick := time.NewTicker(checkInterval)

	for {
		select {
		case <-tick.C:
			result := Result{
				Status:  StatusHealthy,
				Reasons: []error{},
			}
			for _, checkFunc := range checks {
				if err := checkFunc(); err != nil {
					result.Reasons = append(result.Reasons, err)
					result.Status = StatusUnhealthy
				}
			}
			resultChan <- result

		case <-ctx.Done():
			log.Printf("Health check loop stopped due to context done")

			return nil
		}
	}
}
