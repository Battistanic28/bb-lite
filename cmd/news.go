package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"bb-lite/internal/alphavantage"
	"bb-lite/internal/config"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var newsDaysBack int

var newsCmd = &cobra.Command{
	Use:   "news TICKER",
	Short: "Fetch recent news for a ticker symbol",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := args[0]

		apiKey, err := config.GetAPIKey()
		if err != nil {
			return err
		}

		client := alphavantage.NewClient(apiKey)
		result, err := client.FetchNews(ticker, newsDaysBack)
		if err != nil {
			return fmt.Errorf("fetching news: %w", err)
		}

		if len(result.Articles) == 0 {
			fmt.Printf("No news found for %s in the last %d days.\n", ticker, newsDaysBack)
			return nil
		}

		return runNewsTUI(ticker, result)
	},
}

func init() {
	newsCmd.Flags().IntVar(&newsDaysBack, "days", 7, "lookback window in days")
}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 80 {
		return 120
	}
	return w
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func sentimentStyle(label string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true)
	switch label {
	case "Bullish", "Somewhat-Bullish":
		return base.Foreground(lipgloss.Color("#00C851"))
	case "Bearish", "Somewhat-Bearish":
		return base.Foreground(lipgloss.Color("#FF4444"))
	default:
		return base.Foreground(lipgloss.Color("#FFD700"))
	}
}

type newsModel struct {
	table  table.Model
	banner string
	urls   []string // parallel to rows, for opening on Enter
	status string   // feedback message shown in footer
}

func (m newsModel) Init() tea.Cmd {
	return nil
}

func (m newsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "enter", "o":
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.urls) {
				if err := openBrowser(m.urls[idx]); err != nil {
					m.status = fmt.Sprintf("Failed to open: %s", err)
				} else {
					m.status = fmt.Sprintf("Opened: %s", m.urls[idx])
				}
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var (
	baseStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#666666"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Bold(true)
)

func (m newsModel) View() string {
	help := helpStyle.Render("  q/esc: quit • ↑/↓: navigate • enter/o: open in browser")
	footer := help
	if m.status != "" {
		footer = "  " + statusStyle.Render(m.status) + "\n" + help
	}
	return m.banner + "\n" + baseStyle.Render(m.table.View()) + "\n" + footer + "\n"
}

func runNewsTUI(ticker string, result *alphavantage.NewsResult) error {
	tw := termWidth()

	colDate := 16
	colSource := 20
	colTitle := tw - colDate - colSource - 11
	if colTitle < 20 {
		colTitle = 20
	}

	columns := []table.Column{
		{Title: "Date", Width: colDate},
		{Title: "Source", Width: colSource},
		{Title: "Title", Width: colTitle},
	}

	rows := make([]table.Row, len(result.Articles))
	urls := make([]string, len(result.Articles))
	for i, a := range result.Articles {
		rows[i] = table.Row{
			a.PublishedAt.Format("2006-01-02 15:04"),
			a.Source,
			a.Title,
		}
		urls[i] = a.URL
	}

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(lipgloss.Color("#FFA500")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	height := len(rows)
	if height > 20 {
		height = 20
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
		table.WithStyles(s),
	)

	sentimentText := fmt.Sprintf("%s (%.3f)", result.SentimentLabel, result.OverallScore)
	banner := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf(" BB-LITE NEWS | %s | Last %d days | %d articles | Sentiment: %s",
			ticker, newsDaysBack, len(result.Articles),
			sentimentStyle(result.SentimentLabel).Render(sentimentText)))

	m := newsModel{table: t, banner: banner, urls: urls}
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
