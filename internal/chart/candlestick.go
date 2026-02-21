package chart

import (
	"fmt"
	"math"
	"strings"

	"bb-lite/internal/alphavantage"

	"github.com/charmbracelet/lipgloss"
)

// Professional color palette - Bloomberg terminal inspired
var (
	upColor       = lipgloss.Color("#00C851")  // Bright green
	upWickColor   = lipgloss.Color("#00FF88")  // Light green
	downColor     = lipgloss.Color("#FF4444")  // Bright red
	downWickColor = lipgloss.Color("#FF6666")  // Light red
	axisColor     = lipgloss.Color("#666666")  // Gray axes
	textColor     = lipgloss.Color("#CCCCCC")  // Light text
	symbolColor   = lipgloss.Color("#FFA500")  // Orange (Bloomberg style)
	headerBgColor = lipgloss.Color("#1E3A5F")  // Deep blue header

	// Lipgloss styles
	styleUp       = lipgloss.NewStyle().Foreground(upColor)
	styleUpWick   = lipgloss.NewStyle().Foreground(upWickColor)
	styleDown     = lipgloss.NewStyle().Foreground(downColor)
	styleDownWick = lipgloss.NewStyle().Foreground(downWickColor)
	styleAxis     = lipgloss.NewStyle().Foreground(axisColor)
	styleText     = lipgloss.NewStyle().Foreground(textColor)
	styleSymbol   = lipgloss.NewStyle().Bold(true).Foreground(symbolColor)
	stylePositive = lipgloss.NewStyle().Bold(true).Foreground(upColor)
	styleNegative = lipgloss.NewStyle().Bold(true).Foreground(downColor)
)

// RenderCandlestick produces a professional terminal candlestick chart.
func RenderCandlestick(result *alphavantage.TimeSeriesResult, width, height int) string {
	candles := result.Candles
	if len(candles) == 0 {
		return styleAxis.Render("No data to chart.")
	}

	// Layout
	const yAxisWidth = 12

	chartWidth := width - yAxisWidth - 1
	if chartWidth < 20 {
		chartWidth = 20
	}

	// Aggregate candles to fit width
	maxCandles := chartWidth / 3
	if maxCandles < 1 {
		maxCandles = 1
	}
	buckets := aggregateCandles(candles, maxCandles)

	candleWidth := chartWidth / len(buckets)
	if candleWidth < 2 {
		candleWidth = 2
	}
	if candleWidth > 10 {
		candleWidth = 10
	}
	gap := 1
	if candleWidth < 3 {
		gap = 0
	}
	bodyWidth := candleWidth - gap

	// Price range
	minPrice, maxPrice := findPriceRange(buckets)
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	// Build output
	var sb strings.Builder

	// Header
	sb.WriteString(renderHeader(result, candles, width))

	// Chart rows
	for row := 0; row < height; row++ {
		priceAtRow := maxPrice - (float64(row)/float64(height-1))*priceRange

		// Y-axis label every 4 rows
		if row%4 == 0 || row == height-1 {
			label := formatPrice(priceAtRow)
			sb.WriteString(styleAxis.Render(fmt.Sprintf("%10s ", label)))
		} else {
			sb.WriteString(strings.Repeat(" ", yAxisWidth))
		}

		// Axis line
		sb.WriteString(styleAxis.Render("│"))

		// Candle row
		for i, bucket := range buckets {
			cell := renderCell(bucket, priceAtRow, minPrice, maxPrice, height)
			sb.WriteString(strings.Repeat(cell, bodyWidth))
			if gap > 0 && i < len(buckets)-1 {
				sb.WriteString(" ")
			}
		}

		// Pad to width
		usedCols := len(buckets)*candleWidth - gap
		if usedCols < chartWidth {
			sb.WriteString(strings.Repeat(" ", chartWidth-usedCols))
		}
		sb.WriteString("\n")
	}

	// X-axis line
	sb.WriteString(strings.Repeat(" ", yAxisWidth))
	sb.WriteString(styleAxis.Render("└" + strings.Repeat("─", chartWidth)))
	sb.WriteString("\n")

	// X-axis labels
	sb.WriteString(renderTimeAxis(buckets, yAxisWidth+1, chartWidth, candleWidth))

	return sb.String()
}

// renderCell renders a single cell for a candle at a given price row
func renderCell(c alphavantage.Candle, priceAtRow, minPrice, maxPrice float64, height int) string {
	ratio := (priceAtRow - minPrice) / (maxPrice - minPrice)
	rowY := height - 1 - int(ratio*float64(height-1))

	bodyTop := priceToY(math.Max(c.Open, c.Close), minPrice, maxPrice, height)
	bodyBot := priceToY(math.Min(c.Open, c.Close), minPrice, maxPrice, height)
	yHigh := priceToY(c.High, minPrice, maxPrice, height)
	yLow := priceToY(c.Low, minPrice, maxPrice, height)

	isUp := c.Close >= c.Open

	// Check what part of candle this row represents
	if rowY >= bodyBot && rowY <= bodyTop {
		if isUp {
			return styleUp.Render("█")
		}
		return styleDown.Render("█")
	}
	if rowY >= yLow && rowY <= yHigh && (rowY < bodyBot || rowY > bodyTop) {
		if isUp {
			return styleUpWick.Render("│")
		}
		return styleDownWick.Render("│")
	}
	return " "
}

// priceToY converts price to Y coordinate (0 = top)
func priceToY(price, minPrice, maxPrice float64, height int) int {
	ratio := (price - minPrice) / (maxPrice - minPrice)
	y := height - 1 - int(ratio*float64(height-1))
	if y < 0 {
		return 0
	}
	if y >= height {
		return height - 1
	}
	return y
}

// renderHeader creates professional OHLC header
func renderHeader(result *alphavantage.TimeSeriesResult, candles []alphavantage.Candle, width int) string {
	if len(candles) == 0 {
		return ""
	}

	first := candles[0]
	last := candles[len(candles)-1]

	change := last.Close - first.Open
	changePct := (change / first.Open) * 100

	changeStyle := stylePositive
	if change < 0 {
		changeStyle = styleNegative
	}

	symbol := styleSymbol.Render(strings.ToUpper(result.Symbol))
	openStr := fmt.Sprintf("O:%.2f", first.Open)
	highStr := fmt.Sprintf("H:%.2f", maxHigh(candles))
	lowStr := fmt.Sprintf("L:%.2f", minLow(candles))
	closeStr := fmt.Sprintf("C:%.2f", last.Close)
	changeStr := changeStyle.Render(fmt.Sprintf("%+6.2f (%+.2f%%)", change, changePct))

	line := fmt.Sprintf(" %s │ %s │ %s %s %s %s │ %s",
		symbol, result.Interval, openStr, highStr, lowStr, closeStr, changeStr)

	boxWidth := width - 2
	if len(line) > boxWidth {
		boxWidth = len(line) + 4
	}

	var sb strings.Builder
	sb.WriteString(styleAxis.Render("┌" + strings.Repeat("─", boxWidth) + "┐\n"))
	sb.WriteString(styleAxis.Render("│") + line + strings.Repeat(" ", boxWidth-len(line)) + styleAxis.Render("│\n"))
	sb.WriteString(styleAxis.Render("└" + strings.Repeat("─", boxWidth) + "┘\n"))

	return sb.String()
}

// renderTimeAxis creates X-axis date labels
func renderTimeAxis(buckets []alphavantage.Candle, offset, chartWidth, candleWidth int) string {
	if len(buckets) == 0 {
		return ""
	}

	lineLen := offset + chartWidth
	line := make([]byte, lineLen)
	for i := range line {
		line[i] = ' '
	}

	labelCount := 5
	if len(buckets) < labelCount {
		labelCount = len(buckets)
	}

	lastEnd := 0
	for i := 0; i < labelCount; i++ {
		idx := i * (len(buckets) - 1) / (labelCount - 1)
		if idx >= len(buckets) {
			idx = len(buckets) - 1
		}

		label := buckets[idx].Timestamp.Format("Jan-02")
		pos := offset + idx*candleWidth + candleWidth/2 - len(label)/2

		if pos < 0 {
			pos = 0
		}
		if pos+len(label) > lineLen {
			pos = lineLen - len(label)
		}
		if i > 0 && pos < lastEnd+1 {
			continue
		}
		copy(line[pos:], []byte(label))
		lastEnd = pos + len(label)
	}

	return styleAxis.Render(string(line)) + "\n"
}

// Helper functions

func findPriceRange(buckets []alphavantage.Candle) (min, max float64) {
	min = math.MaxFloat64
	max = -math.MaxFloat64
	for _, c := range buckets {
		if c.Low < min {
			min = c.Low
		}
		if c.High > max {
			max = c.High
		}
	}
	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}
	padding := range_ * 0.1
	return min - padding, max + padding
}

func formatPrice(price float64) string {
	absPrice := math.Abs(price)
	switch {
	case absPrice >= 10000:
		return fmt.Sprintf("%.0f", price)
	case absPrice >= 100:
		return fmt.Sprintf("%.2f", price)
	case absPrice >= 1:
		return fmt.Sprintf("%.3f", price)
	default:
		return fmt.Sprintf("%.4f", price)
	}
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

func aggregateCandles(candles []alphavantage.Candle, maxCols int) []alphavantage.Candle {
	if len(candles) <= maxCols {
		return candles
	}

	bucketSize := float64(len(candles)) / float64(maxCols)
	result := make([]alphavantage.Candle, 0, maxCols)

	for i := 0; i < maxCols; i++ {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(candles) {
			end = len(candles)
		}
		if start >= end {
			continue
		}

		merged := candles[start]
		for j := start + 1; j < end; j++ {
			if candles[j].High > merged.High {
				merged.High = candles[j].High
			}
			if candles[j].Low < merged.Low {
				merged.Low = candles[j].Low
			}
			merged.Close = candles[j].Close
			merged.Volume += candles[j].Volume
		}
		merged.Timestamp = candles[start].Timestamp
		result = append(result, merged)
	}

	return result
}
