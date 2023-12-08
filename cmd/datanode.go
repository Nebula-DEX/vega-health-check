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
	dataNodeHTTPPort     int
	dataNodeCoreEndpoint string
	dataNodeAPIEndpoint  string
)

var DataNodeCmd = &cobra.Command{
	Use:   "data-node",
	Short: "Start the data node(including core) health-check service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDataNodeHealthCheck(dataNodeHTTPPort, dataNodeCoreEndpoint, dataNodeAPIEndpoint); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	DataNodeCmd.PersistentFlags().IntVar(&dataNodeHTTPPort, "http-port", 8080, "The HTTP Server port, where health-check is hosted")
	DataNodeCmd.PersistentFlags().StringVar(&dataNodeCoreEndpoint, "core-url", "https://localhost:3003", "HTTP URL for the core")
	DataNodeCmd.PersistentFlags().StringVar(&dataNodeAPIEndpoint, "api-url", "https://localhost:3008", "HTTP URL for the data node API")
}

func runDataNodeHealthCheck(vegaHTTPPort int, coreEndpoint, dataNodeAPIEndpoint string) error {
	ctx := context.Background()
	healthCheckServer := checks.NewHealthCheckServer(vegaHTTPPort, []checks.HealthCheckFunc{
		checks.CheckVegaHttpOnlineWrapper(coreEndpoint),
		checks.CheckDataNodeHttpOnlineWrapper(coreEndpoint),
		checks.CheckVegaBlockIncreasedWrapper(coreEndpoint, 3*time.Second),
		checks.CheckDataNodeLagWrapper(coreEndpoint, dataNodeAPIEndpoint),
	})
	healthCheckServer.Start(ctx)

	<-ctx.Done()

	return nil
}
