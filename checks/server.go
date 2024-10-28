package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
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

func (hcs *HealthCheckServer) Start(ctx context.Context, interval time.Duration) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hcs.handler())

	httpSerer := &http.Server{
		Addr:    fmt.Sprintf(":%d", hcs.port),
		Handler: mux,
	}

	go func() {
		if err := httpSerer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	resultChan := make(chan Result)

	go func() {
		if err := HealthCheckLoop(ctx, resultChan, hcs.checks, interval); err != nil {
			log.Fatal(err)
		}
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
			if _, err := w.Write([]byte(err.Error())); err != nil {
				log.Printf("ERROR: failed writing marshal error: %s", err.Error())
			}

			return
		}

		w.Header().Add("Content-Type", "application/json")
		if result.Status != StatusHealthy {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if _, err := w.Write(resBytes); err != nil {
			log.Printf("ERROR: failed writing http response: %s", err.Error())
		}
	}
}
