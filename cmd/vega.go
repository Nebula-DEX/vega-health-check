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
	vegaHTTPPort int
	coreEndpoint string
)

var VegaCmd = &cobra.Command{
	Use:   "vega",
	Short: "Start the vega core only health-check service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runVegaHealthCheck(vegaHTTPPort, coreEndpoint); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	VegaCmd.PersistentFlags().IntVar(&vegaHTTPPort, "http-port", 8080, "The HTTP Server port, where health-check is hosted")
	VegaCmd.PersistentFlags().StringVar(&coreEndpoint, "core-url", "https://localhost:3003", "HTTP URL for the core")
}

func runVegaHealthCheck(vegaHTTPPort int, coreEndpoint string) error {
	ctx := context.Background()
	healthCheckServer := checks.NewHealthCheckServer(vegaHTTPPort, []checks.HealthCheckFunc{
		checks.CheckVegaHttpOnlineWrapper(coreEndpoint),
		checks.CompareVegaAndCurrentTime(coreEndpoint),
		checks.CheckVegaBlockIncreasedWrapper(coreEndpoint, 3*time.Second),
	})
	healthCheckServer.Start(ctx)

	<-ctx.Done()

	return nil
}
