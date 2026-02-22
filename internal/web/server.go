package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"bb-lite/internal/alphavantage"
)

//go:embed chart.html
var chartHTML embed.FS

type candleJSON struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

func OpenChart(result *alphavantage.TimeSeriesResult, chartType string) error {
	candles := make([]candleJSON, len(result.Candles))
	for i, c := range result.Candles {
		candles[i] = candleJSON{
			Time:   c.Timestamp.Unix(),
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
		}
	}

	data, err := json.Marshal(candles)
	if err != nil {
		return fmt.Errorf("marshaling candle data: %w", err)
	}

	tmpl, err := template.ParseFS(chartHTML, "chart.html")
	if err != nil {
		return fmt.Errorf("parsing chart template: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting listener: %w", err)
	}

	addr := listener.Addr().String()
	url := "http://" + addr

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, struct {
			Symbol    string
			DataJSON  template.JS
			ChartType string
		}{
			Symbol:    result.Symbol,
			DataJSON:  template.JS(data),
			ChartType: chartType,
		})
	})

	fmt.Printf("Opening chart at %s\n", url)
	fmt.Println("Press Ctrl+C to stop the server.")

	openBrowser(url)

	return http.Serve(listener, mux)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
