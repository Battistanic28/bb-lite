## Roadmap

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
