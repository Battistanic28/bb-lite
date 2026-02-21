package chart

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"bb-lite/internal/alphavantage"

	"github.com/charmbracelet/lipgloss"
	"github.com/NimbleMarkets/ntcharts/linechart"
)

// Color palette — Bloomberg terminal inspired
var (
	upColor     = lipgloss.Color("#00C851")
	downColor   = lipgloss.Color("#FF4444")
	axisColor   = lipgloss.Color("#666666")
	labelColor  = lipgloss.Color("#888888")
	symbolColor = lipgloss.Color("#FFA500")
	lineColor   = lipgloss.Color("#00BFFF")

	bullStyle = lipgloss.NewStyle().Foreground(upColor)
	bearStyle = lipgloss.NewStyle().Foreground(downColor)
	axisStyle = lipgloss.NewStyle().Foreground(axisColor)
	lblStyle  = lipgloss.NewStyle().Foreground(labelColor)
	symStyle  = lipgloss.NewStyle().Bold(true).Foreground(symbolColor)
	posStyle  = lipgloss.NewStyle().Bold(true).Foreground(upColor)
	negStyle  = lipgloss.NewStyle().Bold(true).Foreground(downColor)
	linStyle  = lipgloss.NewStyle().Foreground(lineColor)
)

// priceLabelFormatter returns a Y-axis label formatter for prices.
func priceLabelFormatter() linechart.LabelFormatter {
	return func(_ int, v float64) string {
		abs := math.Abs(v)
		switch {
		case abs >= 10000:
			return fmt.Sprintf("%.0f", v)
		case abs >= 100:
			return fmt.Sprintf("%.2f", v)
		case abs >= 1:
			return fmt.Sprintf("%.3f", v)
		default:
			return fmt.Sprintf("%.4f", v)
		}
	}
}

// renderHeader creates a Bloomberg-style OHLC header box.
func renderHeader(result *alphavantage.TimeSeriesResult, candles []alphavantage.Candle, width int) string {
	if len(candles) == 0 {
		return ""
	}

	first := candles[0]
	last := candles[len(candles)-1]

	change := last.Close - first.Open
	changePct := (change / first.Open) * 100

	symbolPlain := strings.ToUpper(result.Symbol)
	openStr := fmt.Sprintf("O:%.2f", first.Open)
	highStr := fmt.Sprintf("H:%.2f", maxHigh(candles))
	lowStr := fmt.Sprintf("L:%.2f", minLow(candles))
	closeStr := fmt.Sprintf("C:%.2f", last.Close)
	changePlain := fmt.Sprintf("%+6.2f (%+.2f%%)", change, changePct)

	plainLine := fmt.Sprintf(" %s │ %s │ %s %s %s %s │ %s",
		symbolPlain, result.Interval, openStr, highStr, lowStr, closeStr, changePlain)

	chgStyle := posStyle
	if change < 0 {
		chgStyle = negStyle
	}

	styledLine := fmt.Sprintf(" %s │ %s │ %s %s %s %s │ %s",
		symStyle.Render(symbolPlain), result.Interval, openStr, highStr, lowStr, closeStr,
		chgStyle.Render(changePlain))

	visibleLen := utf8.RuneCountInString(plainLine)
	boxWidth := width - 2
	if visibleLen > boxWidth {
		boxWidth = visibleLen + 4
	}

	var sb strings.Builder
	sb.WriteString(axisStyle.Render("┌"+strings.Repeat("─", boxWidth)+"┐") + "\n")
	sb.WriteString(axisStyle.Render("│") + styledLine + strings.Repeat(" ", boxWidth-visibleLen) + axisStyle.Render("│") + "\n")
	sb.WriteString(axisStyle.Render("└"+strings.Repeat("─", boxWidth)+"┘") + "\n")

	return sb.String()
}

// renderLineHeader creates a header for intraday line charts showing time range.
func renderLineHeader(result *alphavantage.TimeSeriesResult, candles []alphavantage.Candle, width int) string {
	if len(candles) == 0 {
		return ""
	}

	first := candles[0]
	last := candles[len(candles)-1]

	change := last.Close - first.Close
	changePct := (change / first.Close) * 100

	symbolPlain := strings.ToUpper(result.Symbol)
	openStr := fmt.Sprintf("O:%.2f", first.Open)
	closeStr := fmt.Sprintf("C:%.2f", last.Close)
	changePlain := fmt.Sprintf("%+6.2f (%+.2f%%)", change, changePct)
	timeRange := fmt.Sprintf("%s → %s",
		first.Timestamp.Format("15:04"),
		last.Timestamp.Format("15:04"),
	)

	plainLine := fmt.Sprintf(" %s │ %s │ %s │ %s %s │ %s",
		symbolPlain, result.Interval, timeRange, openStr, closeStr, changePlain)

	chgStyle := posStyle
	if change < 0 {
		chgStyle = negStyle
	}

	styledLine := fmt.Sprintf(" %s │ %s │ %s │ %s %s │ %s",
		symStyle.Render(symbolPlain), result.Interval, timeRange, openStr, closeStr,
		chgStyle.Render(changePlain))

	visibleLen := utf8.RuneCountInString(plainLine)
	boxWidth := width - 2
	if visibleLen > boxWidth {
		boxWidth = visibleLen + 4
	}

	var sb strings.Builder
	sb.WriteString(axisStyle.Render("┌"+strings.Repeat("─", boxWidth)+"┐") + "\n")
	sb.WriteString(axisStyle.Render("│") + styledLine + strings.Repeat(" ", boxWidth-visibleLen) + axisStyle.Render("│") + "\n")
	sb.WriteString(axisStyle.Render("└"+strings.Repeat("─", boxWidth)+"┘") + "\n")

	return sb.String()
}

func maxHigh(candles []alphavantage.Candle) float64 {
	max := 0.0
	for _, c := range candles {
		if c.High > max {
			max = c.High
		}
	}
	return max
}

func minLow(candles []alphavantage.Candle) float64 {
	min := math.MaxFloat64
	for _, c := range candles {
		if c.Low < min {
			min = c.Low
		}
	}
	return min
}
