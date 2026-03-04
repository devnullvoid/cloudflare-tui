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

func (i ZoneItem) Title() string       { return i.Name }
func (i ZoneItem) Description() string { return i.ID }
func (i ZoneItem) FilterValue() string { return i.Name }

// FetchZones returns a tea.Cmd that fetches all Cloudflare zones.
func FetchZones(api *cloudflare.API) tea.Cmd {
	return func() tea.Msg {
		zones, err := api.ListZones(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}
		return FetchedZonesMsg(zones)
	}
}

// InitialModel returns the initial state of the application.
func InitialModel(api *cloudflare.API, theme Theme) Model {
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

	r := list.New([]list.Item{}, delegate, 0, 0)
	r.Title = "DNS Records"
	r.Styles.Title = r.Styles.Title.Background(theme.Primary).Foreground(lipgloss.Color("#1e1e2e"))

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Initialize Logger
	f, _ := os.OpenFile("cftui.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	logger := log.New(f)

	return Model{
		State:      LoadingZonesState,
		CfClient:   api,
		Theme:      theme,
		ZoneList:   l,
		RecordList: r,
		Spinner:    s,
		Logger:     logger,
		LogFile:    f,
	}
}
