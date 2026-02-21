package cmd

import (
	"fmt"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/chart"
	"bb-lite/internal/web"

	"github.com/spf13/cobra"
)

var intradayInterval string
var intradayWeb bool

var intradayCmd = &cobra.Command{
	Use:   "intraday TICKER",
	Short: "Show a line chart of intraday close prices for a ticker symbol",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := args[0]

		// Using demo API key for the test endpoint.
		// TODO: switch to config.GetAPIKey() once a premium key is available.

		// apiKey, err := config.GetAPIKey()
		// if err != nil {
		// 	return err
		// }
		client := alphavantage.NewClient("demo")
		// client := alphavantage.NewClient(apiKey)

		result, err := client.FetchIntraday(ticker, intradayInterval)
		if err != nil {
			return fmt.Errorf("fetching intraday data: %w", err)
		}

		if intradayWeb {
			return web.OpenChart(result)
		}

		tw := termWidth()
		output := chart.RenderLineChart(result, tw, 24)
		fmt.Print(output)
		return nil
	},
}

func init() {
	intradayCmd.Flags().StringVar(&intradayInterval, "interval", "5min", "data interval (1min, 5min, 15min, 30min, 60min)")
	intradayCmd.Flags().BoolVar(&intradayWeb, "web", false, "open an interactive chart in the browser")
}
