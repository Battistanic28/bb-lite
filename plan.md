 Plan: Switch performance chart to ntcharts + Bubble Tea                                                                                                          

 Context

 The current candlestick chart in internal/chart/candlestick.go is hand-rolled with ANSI escape codes. It works but has limited rendering quality (blocky wicks, no
  proper Unicode candlestick runes). Switching to ntcharts gives us proper candlestick rendering via its timeserieslinechart component with DrawCandle(), plus a
 path to interactive features (scroll, zoom, resize) via Bubble Tea.

 How ntcharts candlestick works

 From the ntcharts-ohlc example:
 - Create a timeserieslinechart.Model with time range + Y range
 - Push OHLC data as four separate named data sets ("open", "high", "low", "close"), each using tslc.TimePoint{Time, Value}
 - Call chart.DrawCandle("open", "high", "low", "close", bullStyle, bearStyle) in Update()
 - chart.View() returns the rendered string
 - Runs inside tea.NewProgram() with alt screen, handles resize and quit keys

 Files to Modify

 1. internal/chart/candlestick.go (REWRITE)

 Replace the hand-rolled renderer with a Bubble Tea model wrapping timeserieslinechart:

 - model struct: holds tslc.Model, header string, zone manager
 - newModel(result, width, height): creates the tslc chart with:
   - tslc.WithTimeRange(minTime, maxTime)
   - tslc.WithYRange(minPrice, maxPrice)
   - tslc.WithAxesStyles(axisStyle, labelStyle)
   - Pushes four data sets from result.Candles: "open", "high", "low", "close"
 - Update(): handles tea.WindowSizeMsg (resize chart), tea.KeyMsg (q/ctrl+c to quit), mouse wheel (scroll via tslc.DateNoZoomUpdateHandler)
 - View(): calls DrawCandle(), returns header banner + chart.View()
 - Remove renderCandleCell, aggregateCandles, renderTimeAxis (ntcharts handles all of this)
 - Keep the header banner logic (ticker, open/close/change %) but render it with lipgloss instead of raw ANSI

 2. cmd/performance.go (MODIFY)

 - Instead of calling chart.RenderCandlestick() and printing, call chart.RunChart(result) which launches a tea.NewProgram() with alt screen
 - Remove the fmt.Print(output) pattern
 - Still uses termWidth() for initial sizing

 3. Dependencies (go.mod)

 - Add github.com/NimbleMarkets/ntcharts
 - Add github.com/charmbracelet/bubbletea
 - Add github.com/charmbracelet/lipgloss
 - lipgloss may already be a transitive dep of bubbletea; go mod tidy will sort it out

 No changes needed

 - internal/alphavantage/timeseries.go — data layer unchanged
 - cmd/root.go — already registers the command
 - cmd/news.go — unrelated

 Verification

 go build -o bb-lite . && ./bb-lite performance NVDA --days 30
 Expect:
 - Full-screen alt-screen TUI with candlestick chart
 - Green candles for up days, red for down days
 - Y-axis with price labels, X-axis with dates
 - Header showing ticker, open, close, change %
 - q or ctrl+c exits back to terminal
 - Window resize redraws the chart
