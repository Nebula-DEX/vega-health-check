package checks

import (
	"context"
	"time"

	"log"
)

func HealthCheckLoop(ctx context.Context, resultChan chan<- Result, checks []HealthCheckFunc, interval time.Duration) error {

	tick := time.NewTicker(interval)

	checkExecution := func(resultChan chan<- Result, checks []HealthCheckFunc) {
		result := Result{
			Status:  StatusHealthy,
			Reasons: []error{},
		}
		for _, checkFunc := range checks {
			if err := checkFunc(); err != nil {
				result.Reasons = append(result.Reasons, err)
				result.Status = StatusUnhealthy

				log.Printf("Endpoint unhealthy because of %s", err.Error())
			}
		}
		resultChan <- result
		if len(result.Reasons) < 1 {
			log.Print("Health check loop execution finished: Node healthy")
		} else {
			log.Print("Health check loop execution finished: Node unhealthy")
		}
	}

	for {
		checkExecution(resultChan, checks)
		select {
		case <-tick.C:

		case <-ctx.Done():
			log.Printf("Health check loop stopped due to context done")

			return nil
		}

	}
}
