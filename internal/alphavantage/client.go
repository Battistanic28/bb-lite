package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Article struct {
	Title       string
	Source      string
	URL         string
	PublishedAt time.Time
}

type NewsResult struct {
	Articles       []Article
	OverallScore   float64
	SentimentLabel string
}

type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type apiResponse struct {
	Items string     `json:"items"`
	Feed  []feedItem `json:"feed"`
}

type feedItem struct {
	Title           string            `json:"title"`
	URL             string            `json:"url"`
	Source          string            `json:"source"`
	TimePublished   string            `json:"time_published"`
	TickerSentiment []tickerSentiment `json:"ticker_sentiment"`
}

type tickerSentiment struct {
	Ticker string  `json:"ticker"`
	Score  float64 `json:"ticker_sentiment_score,string"`
	Label  string  `json:"ticker_sentiment_label"`
}

func (c *Client) FetchNews(ticker string, days int) (*NewsResult, error) {
	timeFrom := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	timeFromStr := timeFrom.Format("20060102T1504")

	u := fmt.Sprintf("https://www.alphavantage.co/query?function=NEWS_SENTIMENT&tickers=%s&time_from=%s&apikey=%s",
		url.QueryEscape(ticker),
		timeFromStr,
		url.QueryEscape(c.APIKey),
	)

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

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	upperTicker := strings.ToUpper(ticker)
	articles := make([]Article, 0, len(apiResp.Feed))
	var scoreSum float64
	var scoreCount int

	for _, item := range apiResp.Feed {
		published, err := time.Parse("20060102T150405", item.TimePublished)
		if err != nil {
			continue
		}
		articles = append(articles, Article{
			Title:       item.Title,
			Source:      item.Source,
			URL:         item.URL,
			PublishedAt: published,
		})
		for _, ts := range item.TickerSentiment {
			if strings.ToUpper(ts.Ticker) == upperTicker {
				scoreSum += ts.Score
				scoreCount++
				break
			}
		}
	}

	result := &NewsResult{Articles: articles}
	if scoreCount > 0 {
		result.OverallScore = scoreSum / float64(scoreCount)
		result.SentimentLabel = labelFromScore(result.OverallScore)
	}

	return result, nil
}

func labelFromScore(score float64) string {
	switch {
	case score <= -0.35:
		return "Bearish"
	case score <= -0.15:
		return "Somewhat-Bearish"
	case score < 0.15:
		return "Neutral"
	case score < 0.35:
		return "Somewhat-Bullish"
	default:
		return "Bullish"
	}
}
