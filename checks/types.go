package checks

import (
	"errors"
)

type HealthCheckStatus string

const (
	StatusUnknown   HealthCheckStatus = "UNKNOWN"
	StatusHealthy   HealthCheckStatus = "HEALTHY"
	StatusUnhealthy HealthCheckStatus = "UNHEALTHY"
)

var (
	ErrCoreHTTPIsNotOnline        error = errors.New("core http endpoint is not online")
	ErrCoreInvalidResponse        error = errors.New("invalid non json response from core")
	ErrFailedToParseCurrentTime   error = errors.New("failed to parse current time from statistics endpoint")
	ErrFailedToParseVegaTime      error = errors.New("failed to parse vega time from statistics endpoint")
	ErrCoreHttpRequestTookTooLong error = errors.New("http to core request took too long")

	ErrDataNodeHTTPIsNotOnline error = errors.New("data node http endpoint is not online")
	ErrDataNodeInvalidResponse error = errors.New("invalid non json response from core")
	ErrDataNodeIsLagging       error = errors.New("data node is lagging behind core")

	ErrBlockExplorerHTTPIsNotOnline   error = errors.New("block explorer http endpoint is not online")
	ErrBlockExplorerInvalidResponse   error = errors.New("invalid response from the block explorer")
	ErrBlockExplorerHasNoTransactions error = errors.New("block explorer returns empty list of the transactions")

	ErrDataNodeHttpRequestTookTooLong error = errors.New("http to data node request took too long")

	ErrBlockDidNotIncreased error = errors.New("block did not increased")
	ErrCoreTimeDiffTooBig   error = errors.New("core time is too far in the past")
)

type Result struct {
	Status  HealthCheckStatus
	Reasons []error
}

type HealthCheckFunc func() error

func NewUnknownResult() Result {
	return Result{
		Status:  StatusUnknown,
		Reasons: []error{},
	}
}

type HealthCheckResponse struct {
	Status  string   `json:"status"`
	Reasons []string `json:"reasons"`
}

type StatisticsResponse struct {
	Statistics struct {
		BlockHeight string `json:"blockHeight"`
		CurrentTime string `json:"currentTime"`
		VegaTime    string `json:"vegaTime"`
	} `json:"statistics"`
}

type TransactionsResponse struct {
	Transactions []struct {
		Block string `json:"block"`
	} `json:"transactions"`
}
