package chart

import (
	"bb-lite/internal/alphavantage"

	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
)

// RenderLineChart produces a terminal line chart using ntcharts braille rendering.
func RenderLineChart(result *alphavantage.TimeSeriesResult, width, height int) string {
	candles := result.Candles
	if len(candles) == 0 {
		return axisStyle.Render("No data to chart.")
	}

	chart := tslc.New(width, height,
		tslc.WithAxesStyles(axisStyle, lblStyle),
		tslc.WithStyle(linStyle),
		tslc.WithYLabelFormatter(priceLabelFormatter()),
		tslc.WithXLabelFormatter(tslc.HourTimeLabelFormatter()),
	)

	for _, c := range candles {
		chart.Push(tslc.TimePoint{Time: c.Timestamp, Value: c.Close})
	}

	chart.DrawBraille()

	return renderLineHeader(result, candles, width) + chart.View()
}
