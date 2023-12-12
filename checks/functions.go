package checks

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	tooLongRequestThreshold    = 3 * time.Second
	dataNodeLagBlocksThreshold = 50
	timeDiffThreshold          = 60 * time.Second
)

func CheckDataNodeHttpOnlineWrapper(dataNodeURL string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/api/v2/info", strings.TrimRight(dataNodeURL, "/")), nil)

	return func() error {
		result, err := client.Get(nil)
		if err != nil {
			log.Printf("CheckDataNodeHttpOnlineWrapper: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrDataNodeInvalidResponse
			}

			return ErrDataNodeHTTPIsNotOnline
		}

		if result.Duration > tooLongRequestThreshold {
			return ErrDataNodeHttpRequestTookTooLong
		}

		if result.StatusCode != http.StatusOK {
			return ErrDataNodeHTTPIsNotOnline
		}

		return nil
	}
}

func CheckVegaHttpOnlineWrapper(coreURL string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/statistics", strings.TrimRight(coreURL, "/")), nil)

	return func() error {
		result, err := client.Get(nil)
		if err != nil {
			log.Printf("CheckVegaHttpOnlineWrapper: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrCoreInvalidResponse
			}

			return ErrCoreHTTPIsNotOnline
		}

		if result.Duration > tooLongRequestThreshold {
			return ErrCoreHttpRequestTookTooLong
		}

		if result.StatusCode != http.StatusOK {
			return ErrCoreHTTPIsNotOnline
		}

		return nil
	}
}

func CheckVegaBlockIncreasedWrapper(coreURL string, duration time.Duration) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/statistics", strings.TrimRight(coreURL, "/")), nil)

	return func() error {
		stats1 := &StatisticsResponse{}
		if _, err := client.Get(stats1); err != nil {
			log.Printf("CheckVegaBlockIncreasedWrapper: check1: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrCoreInvalidResponse
			}

			return ErrCoreHTTPIsNotOnline
		}

		time.Sleep(duration)

		stats2 := &StatisticsResponse{}
		if _, err := client.Get(stats2); err != nil {
			log.Printf("CheckVegaBlockIncreasedWrapper: check2: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrCoreInvalidResponse
			}

			return ErrCoreHTTPIsNotOnline
		}

		firstBlock, err := strconv.Atoi(stats1.Statistics.BlockHeight)
		if err != nil {
			log.Printf("CheckVegaBlockIncreasedWrapper: conv1: %s", err.Error())
			return ErrCoreInvalidResponse
		}
		secondBlock, err := strconv.Atoi(stats2.Statistics.BlockHeight)
		if err != nil {
			log.Printf("CheckVegaBlockIncreasedWrapper: conv2: %s", err.Error())
			return ErrCoreInvalidResponse
		}

		if firstBlock < 100 || secondBlock < 100 {
			return ErrBlockDidNotIncreased
		}

		if secondBlock-firstBlock < 1 {
			return ErrBlockDidNotIncreased
		}

		return nil
	}
}

func CheckDataNodeLagWrapper(coreURL string, apiURL string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/statistics", strings.TrimRight(coreURL, "/")), nil)

	return func() error {
		stats := &StatisticsResponse{}
		response, err := client.Get(stats)
		if err != nil {
			log.Printf("CheckDataNodeLagWrapper: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrCoreInvalidResponse
			}

			return ErrCoreHTTPIsNotOnline
		}

		coreBlock, err := strconv.Atoi(stats.Statistics.BlockHeight)
		if err != nil {
			log.Printf("CheckDataNodeLagWrapper: conv core block: %s", err.Error())
			return ErrCoreInvalidResponse
		}

		dataNodeBlockStr := response.Headers.Get("x-block-height")
		if dataNodeBlockStr == "" {
			log.Print("CheckDataNodeLagWrapper: empty response")
			return ErrCoreInvalidResponse
		}
		dataNodeBlock, err := strconv.Atoi(dataNodeBlockStr)

		if err != nil {
			log.Printf("CheckDataNodeLagWrapper: conv data node: %s", err.Error())
			return ErrCoreInvalidResponse
		}

		if coreBlock-dataNodeBlock > dataNodeLagBlocksThreshold {
			return ErrDataNodeIsLagging
		}

		return nil
	}
}

func CheckExplorerIsOnlineWrapper(explorerEndpoint string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/rest/info", strings.TrimRight(explorerEndpoint, "/")), nil)

	return func() error {
		result, err := client.Get(nil)
		if err != nil {
			log.Printf("CheckExplorerIsOnlineWrapper: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrBlockExplorerInvalidResponse
			}

			return ErrBlockExplorerHTTPIsNotOnline
		}

		if result.Duration > tooLongRequestThreshold {
			return ErrBlockExplorerInvalidResponse
		}

		if result.StatusCode != http.StatusOK {
			return ErrBlockExplorerHTTPIsNotOnline
		}

		return nil
	}
}

func CheckExplorerTransactionListIsNotEmptyWrapper(explorerEndpoint string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/rest/transactions", strings.TrimRight(explorerEndpoint, "/")), nil)

	return func() error {
		transactions := &TransactionsResponse{}

		if _, err := client.Get(transactions); err != nil {
			log.Printf("CheckExplorerTransactionListIsNotEmptyWrapper: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrBlockExplorerInvalidResponse
			}

			return ErrBlockExplorerHTTPIsNotOnline
		}

		if len(transactions.Transactions) < 1 {
			return ErrBlockExplorerHasNoTransactions
		}

		return nil
	}
}

func CompareVegaAndCurrentTime(coreURL string) HealthCheckFunc {
	client := NewHTTPChecker(fmt.Sprintf("%s/statistics", strings.TrimRight(coreURL, "/")), nil)

	return func() error {
		stats := &StatisticsResponse{}

		if _, err := client.Get(stats); err != nil {
			log.Printf("CompareVegaAndCurrentTime: %s", err.Error())

			if errors.Is(err, ErrHTTPFailUnmarshal) {
				return ErrCoreInvalidResponse
			}

			return ErrCoreHTTPIsNotOnline
		}

		currentTime, err := time.Parse(time.RFC3339Nano, stats.Statistics.CurrentTime)
		if err != nil {
			log.Printf("CompareVegaAndCurrentTime: convert current time: %s", err.Error())
			return ErrFailedToParseCurrentTime
		}

		vegaTime, err := time.Parse(time.RFC3339Nano, stats.Statistics.VegaTime)
		if err != nil {
			log.Printf("CompareVegaAndCurrentTime: convert vega time: %s", err.Error())
			return ErrFailedToParseVegaTime
		}

		timeDiff := currentTime.Sub(vegaTime)
		if timeDiff > timeDiffThreshold {
			return ErrCoreTimeDiffTooBig
		}

		return nil
	}
}
