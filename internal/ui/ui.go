package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return FetchZones(m.CfClient)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case FetchedZonesMsg:
		items := make([]list.Item, len(msg))
		for i, z := range msg {
			items[i] = ZoneItem{ID: z.ID, Name: z.Name}
		}
		m.ZoneList.SetItems(items)
		m.State = ZoneListState
		return m, nil

	case FetchedRecordsMsg:
		items := make([]list.Item, len(msg))
		for i, r := range msg {
			items[i] = RecordItem{DNS: r}
		}
		m.RecordList.SetItems(items)
		m.State = RecordListState
		return m, nil

	case RecordSavedMsg, RecordDeletedMsg:
		m.State = LoadingRecordsState
		return m, FetchRecords(m.CfClient, m.SelectedID)

	case ErrorMsg:
		m.Err = msg
		return m, nil

	case tea.WindowSizeMsg:
		h, v := DocStyle.GetFrameSize()
		m.ZoneList.SetSize(msg.Width-h, msg.Height-v)
		m.RecordList.SetSize(msg.Width-h, msg.Height-v)
	}

	switch m.State {
	case ZoneListState:
		m.ZoneList, cmd = m.ZoneList.Update(msg)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			if i, ok := m.ZoneList.SelectedItem().(ZoneItem); ok {
				m.SelectedID = i.ID
				m.State = LoadingRecordsState
				m.RecordList.Title = "DNS Records: " + i.Name
				return m, FetchRecords(m.CfClient, i.ID)
			}
		}
		return m, cmd

	case RecordListState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc", "backspace":
				m.State = ZoneListState
				return m, nil
			case "a":
				m.Form = NewRecordForm(nil)
				m.State = EditingRecordState
				return m, nil
			case "enter":
				if i, ok := m.RecordList.SelectedItem().(RecordItem); ok {
					m.Form = NewRecordForm(&i.DNS)
					m.State = EditingRecordState
					return m, nil
				}
			case "d":
				if i, ok := m.RecordList.SelectedItem().(RecordItem); ok {
					return m, DeleteRecord(m.CfClient, m.SelectedID, i.DNS.ID)
				}
			}
		}
		m.RecordList, cmd = m.RecordList.Update(msg)
		return m, cmd

	case EditingRecordState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc":
				m.State = RecordListState
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.Form.Focused--
				} else {
					m.Form.Focused++
				}

				if m.Form.Focused > 4 { // 0,1,2: Inputs, 3: Proxied, 4: Save
					m.Form.Focused = 0
				} else if m.Form.Focused < 0 {
					m.Form.Focused = 4
				}

				cmds = make([]tea.Cmd, len(m.Form.Inputs))
				for i := range m.Form.Inputs {
					if i == m.Form.Focused {
						cmds[i] = m.Form.Inputs[i].Focus()
					} else {
						m.Form.Inputs[i].Blur()
					}
				}
				return m, tea.Batch(cmds...)

			case " ":
				if m.Form.Focused == 3 {
					m.Form.Proxied = !m.Form.Proxied
					return m, nil
				}
			case "enter":
				if m.Form.Focused == 4 {
					return m, SaveRecord(m.CfClient, m.SelectedID, m.Form)
				}
			}
		}

		cmds = make([]tea.Cmd, len(m.Form.Inputs))
		for i := range m.Form.Inputs {
			m.Form.Inputs[i], cmds[i] = m.Form.Inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m Model) View() string {
	if m.Err != nil {
		return DocStyle.Render(fmt.Sprintf("%s\n\nPress 'q' to exit.", ErrStyle.Render("Error: "+m.Err.Error())))
	}

	switch m.State {
	case LoadingZonesState:
		return DocStyle.Render("Loading zones from Cloudflare...")
	case ZoneListState:
		return DocStyle.Render(m.ZoneList.View())
	case LoadingRecordsState:
		return DocStyle.Render(fmt.Sprintf("Loading DNS records for %s...", m.SelectedID))
	case RecordListState:
		view := m.RecordList.View()
		help := HelpStyle.Render("(a) add record, (enter) edit record, (d) delete record, (esc) back")
		return DocStyle.Render(view + "\n" + help)
	case EditingRecordState:
		var b strings.Builder

		for i := range m.Form.Inputs {
			b.WriteString(m.Form.Inputs[i].View())
			if i < len(m.Form.Inputs)-1 {
				b.WriteRune('\n')
			}
		}

		proxiedStr := "[ ] Proxied"
		if m.Form.Proxied {
			proxiedStr = "[x] Proxied"
		}
		
		if m.Form.Focused == 3 {
			b.WriteString("\n\n" + FocusedStyle.Render(proxiedStr))
		} else {
			b.WriteString("\n\n" + proxiedStr)
		}

		saveStr := "Save"
		if m.Form.Focused == 4 {
			b.WriteString("\n\n" + FocusedStyle.Render("["+saveStr+"]"))
		} else {
			b.WriteString("\n\n[" + saveStr + "]")
		}

		b.WriteString("\n\n(esc) cancel")

		return DocStyle.Render(b.String())
	default:
		return "Unknown state"
	}
}
