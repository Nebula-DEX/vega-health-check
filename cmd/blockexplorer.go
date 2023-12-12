package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-health-check/checks"
)

var (
	explorerHTTPPort            int
	explorerEndpoint            string
	explorerCoreEndpoint        string
	explorerDataNodeAPIEndpoint string
)

var BlockExplorerCmd = &cobra.Command{
	Use:   "blockexplorer",
	Short: "Start the block explorer health-check service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runExplorerHealthCheck(explorerHTTPPort, explorerCoreEndpoint, explorerDataNodeAPIEndpoint, explorerEndpoint); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	BlockExplorerCmd.PersistentFlags().IntVar(&explorerHTTPPort, "http-port", 8080, "The HTTP Server port, where health-check is hosted")
	BlockExplorerCmd.PersistentFlags().StringVar(&explorerEndpoint, "blockexplorer-api-url", "http://localhost:1515", "HTTP URL for the explorer service")
	BlockExplorerCmd.PersistentFlags().StringVar(&explorerCoreEndpoint, "core-url", "http://localhost:3003", "HTTP URL for the core")
	BlockExplorerCmd.PersistentFlags().StringVar(&explorerDataNodeAPIEndpoint, "data-node-api-url", "", "HTTP URL for the data node API. If empty We do not check the data node API")
}

func runExplorerHealthCheck(vegaHTTPPort int, coreEndpoint, dataNodeAPIEndpoint, explorerEndpoint string) error {
	healthChecks := []checks.HealthCheckFunc{
		checks.CheckVegaHttpOnlineWrapper(coreEndpoint),
		checks.CompareVegaAndCurrentTime(coreEndpoint),
		checks.CheckDataNodeHttpOnlineWrapper(coreEndpoint),
		checks.CheckVegaBlockIncreasedWrapper(coreEndpoint, 3*time.Second),
		checks.CheckExplorerIsOnlineWrapper(explorerEndpoint),
		checks.CheckExplorerTransactionListIsNotEmptyWrapper(explorerEndpoint),
	}

	if dataNodeAPIEndpoint != "" {
		healthChecks = append(healthChecks, checks.CheckDataNodeLagWrapper(coreEndpoint, dataNodeAPIEndpoint))
	}

	ctx := context.Background()
	healthCheckServer := checks.NewHealthCheckServer(vegaHTTPPort, healthChecks)
	healthCheckServer.Start(ctx)

	<-ctx.Done()

	return nil
}
