package chart

import (
	"fmt"
	"math"
	"strings"

	"bb-lite/internal/alphavantage"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[1;31m"
	colorGreen = "\033[1;32m"
	colorCyan  = "\033[1;36m"
	colorWhite = "\033[1;37m"
)

// RenderCandlestick produces a terminal candlestick chart string.
func RenderCandlestick(result *alphavantage.TimeSeriesResult, width, height int) string {
	candles := result.Candles
	if len(candles) == 0 {
		return "No data to chart."
	}

	yAxisWidth := 10
	chartWidth := width - yAxisWidth - 1 // 1 for the axis line
	if chartWidth < 10 {
		chartWidth = 10
	}

	// Aggregate candles into buckets if more candles than can fit
	// Each candle needs at least 2 cols (1 body + 1 gap)
	maxCandles := chartWidth / 2
	if maxCandles < 1 {
		maxCandles = 1
	}
	buckets := aggregateCandles(candles, maxCandles)

	// Calculate per-candle width to fill the chart area
	// Each candle slot = bodyWidth + 1 gap, except last candle has no trailing gap
	candleWidth, gap := 1, 1
	if len(buckets) > 0 {
		// Total width = n * bodyWidth + (n-1) * gap
		// bodyWidth = (chartWidth + gap) / n - gap
		candleWidth = (chartWidth + gap) / len(buckets)
		if candleWidth < 2 {
			candleWidth = 2
			gap = 0
		} else {
			gap = 1
		}
		// Cap candle body width for aesthetics
		if candleWidth > 8 {
			candleWidth = 8
		}
	}
	bodyWidth := candleWidth - gap
	if bodyWidth < 1 {
		bodyWidth = 1
	}
	// Actual total columns used
	totalCols := len(buckets)*candleWidth - gap
	if totalCols < 0 {
		totalCols = 0
	}

	// Find price range
	minPrice := math.MaxFloat64
	maxPrice := -math.MaxFloat64
	for _, c := range buckets {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
	}

	// Add a small margin
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}
	margin := priceRange * 0.05
	minPrice -= margin
	maxPrice += margin
	priceRange = maxPrice - minPrice

	// Build header
	var sb strings.Builder
	firstCandle := candles[0]
	lastCandle := candles[len(candles)-1]
	change := lastCandle.Close - firstCandle.Open
	changePct := (change / firstCandle.Open) * 100
	changeColor := colorGreen
	changeSign := "+"
	if change < 0 {
		changeColor = colorRed
		changeSign = ""
	}

	bannerWidth := width - 2
	headerPlain := fmt.Sprintf(" BB-LITE PERFORMANCE | %s | %s | Open: %.2f | Close: %.2f | Change: %s%.2f (%s%.2f%%) ",
		strings.ToUpper(result.Symbol), result.Interval,
		firstCandle.Open, lastCandle.Close,
		changeSign, change, changeSign, changePct)
	headerColored := fmt.Sprintf(" BB-LITE PERFORMANCE | %s%s%s | %s | Open: %.2f | Close: %.2f | Change: %s%s%.2f (%s%.2f%%)%s ",
		colorCyan, strings.ToUpper(result.Symbol), colorReset,
		result.Interval,
		firstCandle.Open, lastCandle.Close,
		changeColor, changeSign, math.Abs(change), changeSign, math.Abs(changePct), colorReset)

	plainLen := len(headerPlain)
	if plainLen > bannerWidth {
		bannerWidth = plainLen
	}

	sb.WriteString("┌" + strings.Repeat("─", bannerWidth) + "┐\n")
	sb.WriteString("│" + headerColored + strings.Repeat(" ", bannerWidth-plainLen) + "│\n")
	sb.WriteString("└" + strings.Repeat("─", bannerWidth) + "┘\n")

	// Render chart rows
	for row := 0; row < height; row++ {
		priceAtRow := maxPrice - (float64(row)/float64(height-1))*priceRange

		// Y-axis label every 4 rows
		if row%4 == 0 {
			sb.WriteString(fmt.Sprintf("%9.2f ", priceAtRow))
		} else {
			sb.WriteString(strings.Repeat(" ", yAxisWidth))
		}

		// Axis line
		sb.WriteString("│")

		// Each bucket gets candleWidth columns
		for i, bucket := range buckets {
			cell := renderCandleCell(bucket, priceAtRow, priceRange, height)
			sb.WriteString(strings.Repeat(cell, bodyWidth))
			if gap > 0 && i < len(buckets)-1 {
				sb.WriteString(" ")
			}
		}

		// Pad remaining space to fill width
		usedCols := totalCols
		if usedCols < chartWidth {
			sb.WriteString(strings.Repeat(" ", chartWidth-usedCols))
		}

		sb.WriteString("\n")
	}

	// X-axis line
	sb.WriteString(strings.Repeat(" ", yAxisWidth) + "└" + strings.Repeat("─", chartWidth) + "\n")

	// X-axis time labels
	sb.WriteString(renderTimeAxis(buckets, yAxisWidth+1, chartWidth, candleWidth))

	return sb.String()
}

func renderCandleCell(c alphavantage.Candle, priceAtRow, priceRange float64, height int) string {
	rowHeight := priceRange / float64(height-1)
	halfRow := rowHeight / 2

	bodyTop := math.Max(c.Open, c.Close)
	bodyBot := math.Min(c.Open, c.Close)

	color := colorGreen
	if c.Close < c.Open {
		color = colorRed
	}

	// Check if this row intersects the candle
	rowTop := priceAtRow + halfRow
	rowBot := priceAtRow - halfRow

	inBody := rowBot < bodyTop && rowTop > bodyBot
	inWickHigh := rowBot < c.High && rowTop > bodyTop
	inWickLow := rowBot < bodyBot && rowTop > c.Low

	if inBody {
		return color + "█" + colorReset
	}
	if inWickHigh || inWickLow {
		return color + "│" + colorReset
	}
	return " "
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

func renderTimeAxis(buckets []alphavantage.Candle, offset, chartWidth, candleWidth int) string {
	if len(buckets) == 0 {
		return ""
	}

	lineLen := offset + chartWidth
	line := make([]byte, lineLen)
	for i := range line {
		line[i] = ' '
	}

	// Place labels evenly, skipping any that would overlap
	labelCount := 7
	if len(buckets) < labelCount {
		labelCount = len(buckets)
	}

	lastEnd := 0
	for i := 0; i < labelCount; i++ {
		idx := i * (len(buckets) - 1) / (labelCount - 1)
		if idx >= len(buckets) {
			idx = len(buckets) - 1
		}

		label := buckets[idx].Timestamp.Format("Jan 02")

		// Position label at center of the candle slot
		pos := offset + idx*candleWidth
		if pos+len(label) > lineLen {
			pos = lineLen - len(label)
		}
		if pos < 0 {
			pos = 0
		}
		if i > 0 && pos < lastEnd+1 {
			continue
		}
		copy(line[pos:], label)
		lastEnd = pos + len(label)
	}

	return string(line) + "\n"
}
