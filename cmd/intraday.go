package cmd

import (
	"fmt"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/chart"

	"github.com/spf13/cobra"
)

var intradayInterval string

var intradayCmd = &cobra.Command{
	Use:   "intraday TICKER",
	Short: "Show a line chart of intraday close prices for a ticker symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := args[0]

		// Using demo API key for the test endpoint.
		// TODO: switch to config.GetAPIKey() once a premium key is available.
		client := alphavantage.NewClient("demo")
		result, err := client.FetchIntraday(ticker, intradayInterval)
		if err != nil {
			return fmt.Errorf("fetching intraday data: %w", err)
		}

		tw := termWidth()
		output := chart.RenderLineChart(result, tw, 24)
		fmt.Print(output)
		return nil
	},
}

func init() {
	intradayCmd.Flags().StringVar(&intradayInterval, "interval", "5min", "data interval (1min, 5min, 15min, 30min, 60min)")
}
