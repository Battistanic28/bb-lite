package chart

import (
	"bb-lite/internal/alphavantage"

	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
)

// RenderCandlestick produces a terminal candlestick chart using ntcharts.
func RenderCandlestick(result *alphavantage.TimeSeriesResult, width, height int) string {
	candles := result.Candles
	if len(candles) == 0 {
		return axisStyle.Render("No data to chart.")
	}

	chart := tslc.New(width, height,
		tslc.WithAxesStyles(axisStyle, lblStyle),
		tslc.WithYLabelFormatter(priceLabelFormatter()),
		tslc.WithXLabelFormatter(tslc.DateTimeLabelFormatter()),
	)

	for _, c := range candles {
		chart.PushDataSet("open", tslc.TimePoint{Time: c.Timestamp, Value: c.Open})
		chart.PushDataSet("high", tslc.TimePoint{Time: c.Timestamp, Value: c.High})
		chart.PushDataSet("low", tslc.TimePoint{Time: c.Timestamp, Value: c.Low})
		chart.PushDataSet("close", tslc.TimePoint{Time: c.Timestamp, Value: c.Close})
	}

	chart.DrawCandle("open", "high", "low", "close", bullStyle, bearStyle)

	return renderHeader(result, candles, width) + chart.View()
}
