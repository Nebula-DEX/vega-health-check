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

		log.Print("Health check loop execution finished")
	}

	checkExecution(resultChan, checks)

	for {
		select {
		case <-tick.C:
			checkExecution(resultChan, checks)

		case <-ctx.Done():
			log.Printf("Health check loop stopped due to context done")

			return nil
		}
	}
}
