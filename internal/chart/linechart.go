package chart

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"bb-lite/internal/alphavantage"

	"github.com/charmbracelet/lipgloss"
)

var (
	lineColor    = lipgloss.Color("#00BFFF")
	lineStyle    = lipgloss.NewStyle().Foreground(lineColor)
	lineBoldStyle = lipgloss.NewStyle().Bold(true).Foreground(lineColor)
)

// RenderLineChart produces a terminal line chart plotting close prices over time.
func RenderLineChart(result *alphavantage.TimeSeriesResult, width, height int) string {
	candles := result.Candles
	if len(candles) == 0 {
		return styleAxis.Render("No data to chart.")
	}

	const yAxisWidth = 12

	chartWidth := width - yAxisWidth - 1
	if chartWidth < 20 {
		chartWidth = 20
	}

	// Sample candles down to fit chart width
	nPoints := len(candles)
	if nPoints > chartWidth {
		nPoints = chartWidth
	}
	points := make([]alphavantage.Candle, nPoints)
	for i := 0; i < nPoints; i++ {
		srcIdx := i * len(candles) / nPoints
		points[i] = candles[srcIdx]
	}

	// Price range from close prices with padding
	minP, maxP := closePriceRange(points)
	priceRange := maxP - minP
	if priceRange == 0 {
		priceRange = 1
	}

	// Compute interpolated y position for every chart column
	colY := make([]int, chartWidth)
	for x := 0; x < chartWidth; x++ {
		t := float64(x) / float64(chartWidth-1)
		dataIdx := t * float64(nPoints-1)
		lo := int(dataIdx)
		hi := lo + 1
		if hi >= nPoints {
			hi = nPoints - 1
		}
		frac := dataIdx - float64(lo)
		price := points[lo].Close*(1-frac) + points[hi].Close*frac
		colY[x] = priceToY(price, minP, maxP, height)
	}

	var sb strings.Builder

	sb.WriteString(renderLineHeader(result, candles, width))

	for row := 0; row < height; row++ {
		priceAtRow := maxP - (float64(row)/float64(height-1))*priceRange

		if row%4 == 0 || row == height-1 {
			label := formatPrice(priceAtRow)
			sb.WriteString(styleAxis.Render(fmt.Sprintf("%10s ", label)))
		} else {
			sb.WriteString(strings.Repeat(" ", yAxisWidth))
		}

		sb.WriteString(styleAxis.Render("│"))

		for col := 0; col < chartWidth; col++ {
			y := colY[col]
			if row == y {
				sb.WriteString(lineBoldStyle.Render("●"))
			} else if col > 0 {
				prevY := colY[col-1]
				lo, hi := prevY, y
				if lo > hi {
					lo, hi = hi, lo
				}
				if row >= lo && row <= hi {
					sb.WriteString(lineStyle.Render("│"))
				} else {
					sb.WriteString(" ")
				}
			} else {
				sb.WriteString(" ")
			}
		}

		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat(" ", yAxisWidth))
	sb.WriteString(styleAxis.Render("└" + strings.Repeat("─", chartWidth)))
	sb.WriteString("\n")

	sb.WriteString(renderLineTimeAxis(points, chartWidth, yAxisWidth+1))

	return sb.String()
}

// renderLineHeader creates a header bar showing symbol, interval and close change.
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

	// Build plain version for accurate width measurement
	plainLine := fmt.Sprintf(" %s │ %s │ %s │ %s %s │ %s",
		symbolPlain, result.Interval, timeRange, openStr, closeStr, changePlain)

	changeStyle := stylePositive
	if change < 0 {
		changeStyle = styleNegative
	}

	// Build styled version for rendering
	styledLine := fmt.Sprintf(" %s │ %s │ %s │ %s %s │ %s",
		styleSymbol.Render(symbolPlain), result.Interval, timeRange, openStr, closeStr,
		changeStyle.Render(changePlain))

	// Use rune count for visible width (multi-byte chars like │ and → are 1 column wide).
	visibleLen := utf8.RuneCountInString(plainLine)
	boxWidth := width - 2
	if visibleLen > boxWidth {
		boxWidth = visibleLen + 4
	}

	var sb strings.Builder
	// Keep \n outside Render to avoid lipgloss treating it as a multi-line string.
	sb.WriteString(styleAxis.Render("┌"+strings.Repeat("─", boxWidth)+"┐") + "\n")
	sb.WriteString(styleAxis.Render("│") + styledLine + strings.Repeat(" ", boxWidth-visibleLen) + styleAxis.Render("│") + "\n")
	sb.WriteString(styleAxis.Render("└"+strings.Repeat("─", boxWidth)+"┘") + "\n")

	return sb.String()
}

// renderLineTimeAxis creates X-axis time labels for intraday data.
func renderLineTimeAxis(points []alphavantage.Candle, chartWidth, offset int) string {
	if len(points) == 0 {
		return ""
	}

	lineLen := offset + chartWidth
	line := make([]byte, lineLen)
	for i := range line {
		line[i] = ' '
	}

	labelCount := 5
	if len(points) < labelCount {
		labelCount = len(points)
	}

	lastEnd := 0
	for i := 0; i < labelCount; i++ {
		idx := i * (len(points) - 1) / (labelCount - 1)
		if idx >= len(points) {
			idx = len(points) - 1
		}

		label := points[idx].Timestamp.Format("15:04")
		colX := idx * (chartWidth - 1) / (len(points) - 1)
		pos := offset + colX - len(label)/2

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

// closePriceRange returns min/max close prices with 10% padding.
func closePriceRange(candles []alphavantage.Candle) (min, max float64) {
	min = math.MaxFloat64
	max = -math.MaxFloat64
	for _, c := range candles {
		if c.Close < min {
			min = c.Close
		}
		if c.Close > max {
			max = c.Close
		}
	}
	padding := (max - min) * 0.1
	if padding == 0 {
		padding = 1
	}
	return min - padding, max + padding
}
