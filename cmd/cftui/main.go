package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudflare/cloudflare-go"
)

// sessionState tracks which view is currently active.
type sessionState int

const (
	loadingZonesState sessionState = iota
	zoneListState
	loadingRecordsState
	recordListState
)

// Styles for the UI.
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

// Message types for async operations.
type fetchedZonesMsg []cloudflare.Zone
type fetchedRecordsMsg []cloudflare.DNSRecord
type errorMsg error

// zoneItem implements list.Item for the Zone Selection view.
type zoneItem struct {
	id   string
	name string
}

func (i zoneItem) Title() string       { return i.name }
func (i zoneItem) Description() string { return i.id }
func (i zoneItem) FilterValue() string { return i.name }

// recordItem implements list.Item for the DNS Record view.
type recordItem struct {
	dns cloudflare.DNSRecord
}

func (i recordItem) Title() string {
	return fmt.Sprintf("%-6s %s", i.dns.Type, i.dns.Name)
}
func (i recordItem) Description() string {
	proxied := "No"
	if i.dns.Proxied != nil && *i.dns.Proxied {
		proxied = "Yes"
	}
	return fmt.Sprintf("Content: %s | Proxied: %s", i.dns.Content, proxied)
}
func (i recordItem) FilterValue() string { return i.dns.Name }

// model represents the application state.
type model struct {
	state      sessionState
	cfClient   *cloudflare.API
	zoneList   list.Model
	recordList list.Model
	err        error
	selectedID string
}

/*
 * Async Message Passing Explanation:
 * 1. Bubble Tea's `Init()` and `Update()` return a `tea.Cmd`.
 * 2. A `tea.Cmd` is a function that returns a `tea.Msg` (interface{}).
 * 3. These commands are executed concurrently by the Bubble Tea runtime.
 * 4. Once finished, the returned `tea.Msg` is sent back to the `Update()` method.
 * 5. This allows the UI to stay responsive while API calls (like Cloudflare's) 
 *    are being processed in the background.
 */

// fetchZones returns a tea.Cmd that fetches all Cloudflare zones.
// When it completes, it returns either a fetchedZonesMsg or an errorMsg.
func fetchZones(api *cloudflare.API) tea.Cmd {
	return func() tea.Msg {
		// Fetch all zones for the token.
		zones, err := api.ListZones(context.Background())
		if err != nil {
			return errorMsg(err)
		}
		// Return the successful response as a message to be handled in Update().
		return fetchedZonesMsg(zones)
	}
}

// fetchRecords returns a tea.Cmd that fetches DNS records for a specific zone.
// It uses the zone ID and triggers an async update once Cloudflare responds.
func fetchRecords(api *cloudflare.API, zoneID string) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return errorMsg(err)
		}
		// Return the records as a message.
		return fetchedRecordsMsg(records)
	}
}

func (m model) Init() tea.Cmd {
	return fetchZones(m.cfClient)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case fetchedZonesMsg:
		items := make([]list.Item, len(msg))
		for i, z := range msg {
			items[i] = zoneItem{id: z.ID, name: z.Name}
		}
		m.zoneList.SetItems(items)
		m.state = zoneListState
		return m, nil

	case fetchedRecordsMsg:
		items := make([]list.Item, len(msg))
		for i, r := range msg {
			items[i] = recordItem{dns: r}
		}
		m.recordList.SetItems(items)
		m.state = recordListState
		return m, nil

	case errorMsg:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.zoneList.SetSize(msg.Width-h, msg.Height-v)
		m.recordList.SetSize(msg.Width-h, msg.Height-v)
	}

	// Update logic based on current state
	switch m.state {
	case zoneListState:
		m.zoneList, cmd = m.zoneList.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			if i, ok := m.zoneList.SelectedItem().(zoneItem); ok {
				m.selectedID = i.id
				m.state = loadingRecordsState
				m.recordList.Title = "DNS Records: " + i.name
				return m, fetchRecords(m.cfClient, i.id)
			}
		}

	case recordListState:
		m.recordList, cmd = m.recordList.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && (msg.String() == "esc" || msg.String() == "backspace") {
			m.state = zoneListState
			return m, nil
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return docStyle.Render(fmt.Sprintf("%s\n\nPress 'q' to exit.", errStyle.Render("Error: "+m.err.Error())))
	}

	switch m.state {
	case loadingZonesState:
		return docStyle.Render("Loading zones from Cloudflare...")
	case zoneListState:
		return docStyle.Render(m.zoneList.View())
	case loadingRecordsState:
		return docStyle.Render(fmt.Sprintf("Loading DNS records for %s...", m.selectedID))
	case recordListState:
		return docStyle.Render(m.recordList.View())
	default:
		return "Unknown state"
	}
}

func main() {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, errStyle.Render("Error: CLOUDFLARE_API_TOKEN environment variable is required."))
		os.Exit(1)
	}

	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudflare client: %v", err)
	}

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Cloudflare Zones"

	r := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	r.Title = "DNS Records"

	m := model{
		state:      loadingZonesState,
		cfClient:   api,
		zoneList:   l,
		recordList: r,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
