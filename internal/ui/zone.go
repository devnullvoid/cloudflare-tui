package ui

import (
	"context"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/cloudflare/cloudflare-go"
)

// ZoneItem implements list.Item for the Zone Selection view.
type ZoneItem struct {
	ID   string
	Name string
}

func (i *ZoneItem) Title() string       { return i.Name }
func (i *ZoneItem) Description() string { return i.ID }
func (i *ZoneItem) FilterValue() string { return i.Name }

// FetchZones returns a tea.Cmd that fetches all Cloudflare zones.
func FetchZones(api *cloudflare.API, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		logger.Debug("Initiating ListZones API call")
		zones, err := api.ListZones(context.Background())
		if err != nil {
			logger.Error("ListZones API call failed", "error", err)
			return ErrorMsg(err)
		}
		logger.Debug("ListZones API call successful", "count", len(zones))
		return FetchedZonesMsg(zones)
	}
}

// InitialModel returns the initial state of the application.
func InitialModel(api *cloudflare.API, theme *Theme, logger *log.Logger, logFile *os.File) *Model {
	// Customize list delegate with theme colors
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(theme.Primary).
		BorderLeftForeground(theme.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(theme.Secondary).
		BorderLeftForeground(theme.Primary)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Cloudflare Zones"
	l.Styles.Title = l.Styles.Title.Background(theme.Primary).Foreground(lipgloss.Color("#1e1e2e"))

	// Apply theme to filter input
	l.FilterInput.PromptStyle = l.FilterInput.PromptStyle.Foreground(theme.Primary)
	l.FilterInput.TextStyle = l.FilterInput.TextStyle.Foreground(theme.Secondary)
	l.FilterInput.Cursor.Style = l.FilterInput.Cursor.Style.Foreground(theme.Primary)

	r := list.New([]list.Item{}, delegate, 0, 0)
	r.Title = "DNS Records"
	r.Styles.Title = r.Styles.Title.Background(theme.Primary).Foreground(lipgloss.Color("#1e1e2e"))

	// Apply theme to filter input
	r.FilterInput.PromptStyle = r.FilterInput.PromptStyle.Foreground(theme.Primary)
	r.FilterInput.TextStyle = r.FilterInput.TextStyle.Foreground(theme.Secondary)
	r.FilterInput.Cursor.Style = r.FilterInput.Cursor.Style.Foreground(theme.Primary)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Initialize Logger
	if logger == nil {
		logger = log.New(os.Stderr)
		logger.SetLevel(log.FatalLevel)
	}

	// Initialize a dummy form so that Form.TypeList is not zero-valued during WindowSizeMsg
	dummyForm := NewRecordForm(nil, theme)

	return &Model{
		State:      LoadingZonesState,
		CfClient:   api,
		Theme:      *theme,
		ZoneList:   l,
		RecordList: r,
		Form:       dummyForm,
		Spinner:    s,
		Logger:     logger,
		LogFile:    logFile,
	}
}
