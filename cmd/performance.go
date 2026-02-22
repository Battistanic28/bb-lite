package cmd

import (
	"fmt"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/config"
	"bb-lite/internal/web"

	"github.com/spf13/cobra"
)

var perfDaysBack int

var performanceCmd = &cobra.Command{
	Use:   "performance TICKER",
	Short: "Show a candlestick chart for a ticker symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := args[0]

		apiKey, err := config.GetAPIKey()
		if err != nil {
			return err
		}

		client := alphavantage.NewClient(apiKey)
		result, err := client.FetchTimeSeries(ticker, perfDaysBack)
		if err != nil {
			return fmt.Errorf("fetching time series: %w", err)
		}

		return web.OpenChart(result)
	},
}

func init() {
	performanceCmd.Flags().IntVar(&perfDaysBack, "days", 30, "lookback window in days")
}
