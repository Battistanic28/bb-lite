# bb-lite

A Bloomberg terminal-inspired CLI tool for stock market data, powered by the [Alpha Vantage](https://www.alphavantage.co/) API.

## Setup

### API Key

Get a free API key from [Alpha Vantage](https://www.alphavantage.co/support/#api-key).

Configure it in one of two ways:

1. **Environment variable:**
   ```bash
   export ALPHA_VANTAGE_API_KEY=your_key_here
   ```

2. **Config file** (`~/.bb-lite.yaml` or `.bb-lite.yaml` in the current directory):
   ```yaml
   api_key: your_key_here
   ```

### Build

```bash
go build -o bb-lite .
```

## Commands

### `news`

Fetch recent news and sentiment for a ticker symbol.

```bash
bb-lite news NVDA
bb-lite news AAPL --days 14
```

| Flag | Default | Description |
|------|---------|-------------|
| `--days` | `7` | Lookback window in days |

Displays a table with date, source, title, and a clickable link for each article. Includes an aggregate sentiment score (Bullish/Bearish/Neutral) in the header.

### `performance`

Show a candlestick chart for a ticker symbol using daily price data.

```bash
bb-lite performance NVDA
bb-lite performance AAPL --days 60
```

| Flag | Default | Description |
|------|---------|-------------|
| `--days` | `30` | Lookback window in days |

Renders a full-width terminal candlestick chart with bull/bear colored candles, braille-resolution wicks, Y-axis price labels, X-axis date labels, and a header banner showing OHLC and percent change. Powered by [ntcharts](https://github.com/NimbleMarkets/ntcharts).

### `intraday`

Show a line chart of intraday close prices for a ticker symbol.

```bash
bb-lite intraday IBM
bb-lite intraday IBM --interval 15min
```

| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `5min` | Data interval (`1min`, `5min`, `15min`, `30min`, `60min`) |

Renders a braille-resolution line chart of close prices with time-of-day X-axis labels. Currently uses the Alpha Vantage demo API key (only works for `IBM` with `5min` interval).

## Future Considerations

- **Interactive BubbleTea mode** — the charts use [ntcharts](https://github.com/NimbleMarkets/ntcharts) which is built on BubbleTea. A future enhancement could wrap the chart in a `tea.Program` for keyboard scrolling, zooming, and mouse support.

## Roadmap

### Known Limitations

- **Volume subplot:** Add an optional volume bar chart below the price chart.
- **Responsive height:** Chart height is currently fixed. Could auto-scale based on terminal height.

### Market Data

- **Quote** — real-time/delayed price, bid/ask, market cap, P/E, 52-week range (`GLOBAL_QUOTE`)
- **Watchlist** — track multiple tickers with periodic refresh in a tabular display
- **Movers** — top gainers, losers, and most active (`TOP_GAINERS_LOSERS`)
- **Sector performance** — heatmap or bar chart of sector returns

### Fundamentals

- **Company overview** — description, sector, industry, market cap, dividend yield (`OVERVIEW`)
- **Financials** — income statement, balance sheet, cash flow (`INCOME_STATEMENT`, `BALANCE_SHEET`, `CASH_FLOW`)
- **Earnings** — EPS history, surprise %, upcoming earnings dates (`EARNINGS`)

### Technical Analysis

- **Indicators** — overlay SMA/EMA on the candlestick chart, standalone RSI/MACD display (Alpha Vantage has dedicated endpoints)
- **Screener** — filter stocks by technical criteria (above 200-day SMA, RSI oversold, etc.)

### Portfolio & Tracking

- **Portfolio** — local config of positions (ticker, shares, cost basis), show P&L against current prices
- **Alerts** — price threshold notifications via polling with desktop notification or terminal bell

### Fixed Income & Macro

- **Economic indicators** — GDP, CPI, unemployment, federal funds rate (`REAL_GDP`, `CPI`, etc.)
- **Treasury yields** — yield curve chart (`TREASURY_YIELD`)
- **Forex & crypto** — exchange rates and charts (`CURRENCY_EXCHANGE_RATE`, `DIGITAL_CURRENCY_DAILY`)

### UX & Terminal Polish

- **Dashboard mode** — single-screen layout combining quote, chart, and news (similar to Bloomberg's `TOP` screen)
- **Ticker autocomplete** — symbol search and autocomplete (`SYMBOL_SEARCH`)
- **Theming** — Bloomberg-style orange-on-black color scheme
- **Keyboard navigation** — interactive TUI (e.g., with bubbletea) for switching between views
- **Interactive charting** — scrolling, zooming, and cursor-based price inspection
