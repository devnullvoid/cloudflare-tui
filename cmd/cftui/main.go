package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
	editingRecordState
)

// Styles for the UI.
var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	noStyle      = lipgloss.NewStyle()
)

// Message types for async operations.
type fetchedZonesMsg []cloudflare.Zone
type fetchedRecordsMsg []cloudflare.DNSRecord
type recordSavedMsg struct{}
type errorMsg error

// recordForm manages input fields for adding/editing a record.
type recordForm struct {
	id      string
	inputs  []textinput.Model
	focused int
	proxied bool
}

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
	form       recordForm
	err        error
	selectedID string
}

// newRecordForm initializes a form for a DNS record.
func newRecordForm(r *cloudflare.DNSRecord) recordForm {
	var f recordForm
	f.inputs = make([]textinput.Model, 3)

	f.inputs[0] = textinput.New()
	f.inputs[0].Placeholder = "Type (A, CNAME, etc.)"
	f.inputs[0].Focus()

	f.inputs[1] = textinput.New()
	f.inputs[1].Placeholder = "Name"

	f.inputs[2] = textinput.New()
	f.inputs[2].Placeholder = "Content"

	if r != nil {
		f.id = r.ID
		f.inputs[0].SetValue(r.Type)
		f.inputs[1].SetValue(r.Name)
		f.inputs[2].SetValue(r.Content)
		if r.Proxied != nil {
			f.proxied = *r.Proxied
		}
	}

	return f
}

// fetchZones returns a tea.Cmd that fetches all Cloudflare zones.
func fetchZones(api *cloudflare.API) tea.Cmd {
	return func() tea.Msg {
		zones, err := api.ListZones(context.Background())
		if err != nil {
			return errorMsg(err)
		}
		return fetchedZonesMsg(zones)
	}
}

// fetchRecords returns a tea.Cmd that fetches DNS records for a specific zone.
func fetchRecords(api *cloudflare.API, zoneID string) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return errorMsg(err)
		}
		return fetchedRecordsMsg(records)
	}
}

// saveRecord returns a tea.Cmd that saves (creates or updates) a DNS record.
func saveRecord(api *cloudflare.API, zoneID string, f recordForm) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		proxied := f.proxied

		var err error
		if f.id == "" {
			_, err = api.CreateDNSRecord(context.Background(), rc, cloudflare.CreateDNSRecordParams{
				Type:    f.inputs[0].Value(),
				Name:    f.inputs[1].Value(),
				Content: f.inputs[2].Value(),
				Proxied: &proxied,
			})
		} else {
			_, err = api.UpdateDNSRecord(context.Background(), rc, cloudflare.UpdateDNSRecordParams{
				ID:      f.id,
				Type:    f.inputs[0].Value(),
				Name:    f.inputs[1].Value(),
				Content: f.inputs[2].Value(),
				Proxied: &proxied,
			})
		}

		if err != nil {
			return errorMsg(err)
		}
		return recordSavedMsg{}
	}
}

func (m model) Init() tea.Cmd {
	return fetchZones(m.cfClient)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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

	case recordSavedMsg:
		m.state = loadingRecordsState
		return m, fetchRecords(m.cfClient, m.selectedID)

	case errorMsg:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.zoneList.SetSize(msg.Width-h, msg.Height-v)
		m.recordList.SetSize(msg.Width-h, msg.Height-v)
	}

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
		return m, cmd

	case recordListState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc", "backspace":
				m.state = zoneListState
				return m, nil
			case "a":
				m.form = newRecordForm(nil)
				m.state = editingRecordState
				return m, nil
			case "enter":
				if i, ok := m.recordList.SelectedItem().(recordItem); ok {
					m.form = newRecordForm(&i.dns)
					m.state = editingRecordState
					return m, nil
				}
			}
		}
		m.recordList, cmd = m.recordList.Update(msg)
		return m, cmd

	case editingRecordState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc":
				m.state = recordListState
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.form.focused--
				} else {
					m.form.focused++
				}

				if m.form.focused > len(m.form.inputs) {
					m.form.focused = 0
				} else if m.form.focused < 0 {
					m.form.focused = len(m.form.inputs)
				}

				cmds = make([]tea.Cmd, len(m.form.inputs))
				for i := 0; i <= len(m.form.inputs)-1; i++ {
					if i == m.form.focused {
						cmds[i] = m.form.inputs[i].Focus()
						continue
					}
					m.form.inputs[i].Blur()
				}
				return m, tea.Batch(cmds...)

			case "enter":
				if m.form.focused == len(m.form.inputs) {
					return m, saveRecord(m.cfClient, m.selectedID, m.form)
				}
			case " ":
				if m.form.focused == len(m.form.inputs)-1 {
					// This logic is a bit flawed because focused is index based.
					// Let's adjust: 0: Type, 1: Name, 2: Content, 3: Proxied toggle, 4: Save button
				}
			}
		}

		// Handle proxied toggle specifically
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == " " && m.form.focused == 3 {
			m.form.proxied = !m.form.proxied
			return m, nil
		}
		
		// Handle save button specifically
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" && m.form.focused == 4 {
			return m, saveRecord(m.cfClient, m.selectedID, m.form)
		}

		cmd = m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.form.inputs))
	for i := range m.form.inputs {
		m.form.inputs[i], cmds[i] = m.form.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
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
		return docStyle.Render(m.recordList.View() + "\n\n(a) add record, (enter) edit record, (esc) back")
	case editingRecordState:
		var b strings.Builder

		for i := range m.form.inputs {
			b.WriteString(m.form.inputs[i].View())
			if i < len(m.form.inputs)-1 {
				b.WriteRune('\n')
			}
		}

		proxiedStr := "[ ] Proxied"
		if m.form.proxied {
			proxiedStr = "[x] Proxied"
		}
		
		if m.form.focused == 3 {
			b.WriteString("\n\n" + focusedStyle.Render(proxiedStr))
		} else {
			b.WriteString("\n\n" + proxiedStr)
		}

		saveStr := "Save"
		if m.form.focused == 4 {
			b.WriteString("\n\n" + focusedStyle.Render("["+saveStr+"]"))
		} else {
			b.WriteString("\n\n[" + saveStr + "]")
		}

		b.WriteString("\n\n(esc) cancel")

		return docStyle.Render(b.String())
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
