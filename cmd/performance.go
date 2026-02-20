package cmd

import (
	"fmt"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/chart"
	"bb-lite/internal/config"

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

		tw := termWidth()
		output := chart.RenderCandlestick(result, tw, 28)
		fmt.Print(output)
		return nil
	},
}

func init() {
	performanceCmd.Flags().IntVar(&perfDaysBack, "days", 30, "lookback window in days")
}
