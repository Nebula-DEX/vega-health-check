package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type HealthCheckServer struct {
	port int

	checks     []HealthCheckFunc
	lastResult Result
	resultMut  *sync.RWMutex
}

func NewHealthCheckServer(port int, checks []HealthCheckFunc) *HealthCheckServer {
	return &HealthCheckServer{
		port:       port,
		lastResult: NewUnknownResult(),
		resultMut:  &sync.RWMutex{},
		checks:     checks,
	}
}

func (hcs *HealthCheckServer) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hcs.handler())

	httpSerer := &http.Server{
		Addr:    fmt.Sprintf(":%d", hcs.port),
		Handler: mux,
	}

	go func() {
		httpSerer.ListenAndServe()
	}()

	resultChan := make(chan Result)

	go func() {
		HealthCheckLoop(ctx, resultChan, hcs.checks)
	}()

	go func() {
		for {
			select {
			case newResult := <-resultChan:
				hcs.resultMut.Lock()
				hcs.lastResult = newResult
				hcs.resultMut.Unlock()
			case <-ctx.Done():
				log.Printf("Health check result forwarder finished")

				return
			}
		}
	}()
}

func (hcs *HealthCheckServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hcs.resultMut.RLock()
		result := hcs.lastResult
		hcs.resultMut.RUnlock()

		response := HealthCheckResponse{
			Status:  "",
			Reasons: []string{},
		}

		if result.Status == StatusHealthy {
			response.Status = string(StatusHealthy)
		} else {
			response.Status = string(result.Status)

			for _, e := range result.Reasons {
				response.Reasons = append(response.Reasons, e.Error())
			}
		}

		resBytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))

			return
		}

		w.Header().Add("Content-Type", "application/json")
		if result.Status == StatusHealthy {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Write(resBytes)
	}
}
