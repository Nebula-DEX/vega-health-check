package main

import (
	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-health-check/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "vega-health-check",
	Short: "Health check command for vega",
}

func init() {
	rootCmd.AddCommand(cmd.BlockExplorerCmd)
	rootCmd.AddCommand(cmd.VegaCmd)
	rootCmd.AddCommand(cmd.DataNodeCmd)
}

func main() {
	rootCmd.Execute()
}
