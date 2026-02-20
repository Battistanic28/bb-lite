package cmd

import (
	"fmt"
	"os"
	"strings"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/config"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var hoursBack int

var newsCmd = &cobra.Command{
	Use:   "news TICKER",
	Short: "Fetch recent news for a ticker symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := args[0]

		apiKey, err := config.GetAPIKey()
		if err != nil {
			return err
		}

		client := alphavantage.NewClient(apiKey)
		result, err := client.FetchNews(ticker, hoursBack)
		if err != nil {
			return fmt.Errorf("fetching news: %w", err)
		}

		if len(result.Articles) == 0 {
			fmt.Printf("No news found for %s in the last %d hours.\n", ticker, hoursBack)
			return nil
		}

		printGrid(ticker, result)
		return nil
	},
}

func init() {
	newsCmd.Flags().IntVar(&hoursBack, "hours", 24, "lookback window in hours")
}

// ANSI color helpers
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[1;31m"
	colorGreen  = "\033[1;32m"
	colorYellow = "\033[1;33m"
)

func sentimentColor(label string) string {
	switch label {
	case "Bullish", "Somewhat-Bullish":
		return colorGreen
	case "Bearish", "Somewhat-Bearish":
		return colorRed
	default:
		return colorYellow
	}
}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 80 {
		return 120
	}
	return w
}

// hyperlink generates an OSC 8 clickable terminal hyperlink.
// Displays label as visible text, but clicking opens url.
func hyperlink(url, label string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, label)
}

func printGrid(ticker string, result *alphavantage.NewsResult) {
	tw := termWidth()

	const (
		colDate   = 16
		colSource = 20
		colLink   = 6
	)
	// 4 columns: 5 borders + 4 leading spaces = 9 chrome chars
	chrome := 9
	colTitle := tw - chrome - colDate - colSource - colLink
	if colTitle < 20 {
		colTitle = 20
	}

	hDate := strings.Repeat("─", colDate+1)
	hSource := strings.Repeat("─", colSource+1)
	hTitle := strings.Repeat("─", colTitle+1)
	hLink := strings.Repeat("─", colLink+1)

	divider := func(left, mid, right string) string {
		return left + hDate + mid + hSource + mid + hTitle + mid + hLink + right
	}

	row := func(a, b, c, d string) string {
		return fmt.Sprintf("│ %-*s│ %-*s│ %-*s│ %-*s│", colDate, a, colSource, b, colTitle, c, colLink, d)
	}

	// sentiment display with color
	sentimentText := fmt.Sprintf("%s (%.3f)", result.SentimentLabel, result.OverallScore)
	clr := sentimentColor(result.SentimentLabel)
	coloredSentiment := clr + sentimentText + colorReset

	// header line
	headerPlain := fmt.Sprintf(" BB-LITE NEWS │ %s │ Last %d hours │ %d articles │ Sentiment: %s ",
		ticker, hoursBack, len(result.Articles), sentimentText)
	headerColored := fmt.Sprintf(" BB-LITE NEWS │ %s │ Last %d hours │ %d articles │ Sentiment: %s ",
		ticker, hoursBack, len(result.Articles), coloredSentiment)

	bannerWidth := colDate + 1 + colSource + 1 + colTitle + 1 + colLink + 1 + 3
	plainLen := len(headerPlain)
	if plainLen > bannerWidth {
		headerPlain = headerPlain[:bannerWidth]
		headerColored = headerPlain
	}

	fmt.Println("┌" + strings.Repeat("─", bannerWidth) + "┐")
	fmt.Println("│" + headerColored + strings.Repeat(" ", bannerWidth-plainLen) + "│")

	fmt.Println(divider("├", "┬", "┤"))
	fmt.Println(row("DATE", "SOURCE", "TITLE", "LINK"))
	fmt.Println(divider("├", "┼", "┤"))

	for i, a := range result.Articles {
		link := hyperlink(a.URL, "Link")
		// link contains invisible escape chars, so we pad manually
		// visible width of "Link" is 4, column is colLink
		linkPadded := link + strings.Repeat(" ", colLink-4)
		fmt.Printf("│ %-*s│ %-*s│ %-*s│ %s│\n",
			colDate, a.PublishedAt.Format("2006-01-02 15:04"),
			colSource, truncate(a.Source, colSource),
			colTitle, truncate(a.Title, colTitle),
			linkPadded,
		)
		if i < len(result.Articles)-1 {
			fmt.Println(divider("├", "┼", "┤"))
		}
	}

	fmt.Println(divider("└", "┴", "┘"))
}

func pad(s string, width int, ch rune) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(string(ch), width-len(s))
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return string(runes)
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}
