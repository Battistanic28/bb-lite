package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

type TimeSeriesResult struct {
	Symbol   string
	Interval string
	Candles  []Candle
}

func (c *Client) FetchTimeSeries(ticker string, days int) (*TimeSeriesResult, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)

	params := url.Values{}
	params.Set("function", "TIME_SERIES_DAILY")
	params.Set("symbol", ticker)
	params.Set("apikey", c.APIKey)

	u := "https://www.alphavantage.co/query?" + params.Encode()

	resp, err := c.HTTPClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Alpha Vantage returns: { "Meta Data": {...}, "Time Series (Daily)": { "date": { "1. open": "...", ... } } }
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Find the time series key
	var seriesData map[string]map[string]string
	for key, val := range raw {
		if key == "Meta Data" {
			continue
		}
		if key == "Error Message" || key == "Information" || key == "Note" {
			return nil, fmt.Errorf("API error: %s", string(val))
		}
		if err := json.Unmarshal(val, &seriesData); err == nil {
			break
		}
	}
	fmt.Println(seriesData)

	if seriesData == nil {
		return nil, fmt.Errorf("no time series data found in response")
	}

	result := &TimeSeriesResult{
		Symbol:   ticker,
		Interval: "daily",
	}

	for ts, values := range seriesData {
		t, err := time.Parse("2006-01-02", ts)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			continue
		}

		open, _ := strconv.ParseFloat(values["1. open"], 64)
		high, _ := strconv.ParseFloat(values["2. high"], 64)
		low, _ := strconv.ParseFloat(values["3. low"], 64)
		close_, _ := strconv.ParseFloat(values["4. close"], 64)
		vol, _ := strconv.ParseInt(values["5. volume"], 10, 64)

		result.Candles = append(result.Candles, Candle{
			Timestamp: t,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close_,
			Volume:    vol,
		})
	}

	sort.Slice(result.Candles, func(i, j int) bool {
		return result.Candles[i].Timestamp.Before(result.Candles[j].Timestamp)
	})

	if len(result.Candles) == 0 {
		return nil, fmt.Errorf("no candle data found for %s in the last %d days", ticker, days)
	}

	return result, nil
}
