package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
func InitialModel(api *cloudflare.API) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Cloudflare Zones"

	r := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	r.Title = "DNS Records"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(DefaultTheme.Primary)

	return Model{
		State:      LoadingZonesState,
		CfClient:   api,
		Theme:      DefaultTheme,
		ZoneList:   l,
		RecordList: r,
		Spinner:    s,
	}
}
