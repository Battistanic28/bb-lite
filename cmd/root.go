package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "bb-lite",
	Short: "A Bloomberg terminal-like CLI tool",
}

func init() {
	rootCmd.AddCommand(newsCmd)
	rootCmd.AddCommand(performanceCmd)
	rootCmd.AddCommand(intradayCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
