package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpHeight = 2

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		FetchZones(m.CfClient),
		m.Spinner.Tick,
	)
}

func (m Model) Close() {
	if m.LogFile != nil {
		m.LogFile.Close()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Err != nil {
			m.Err = nil
			return m, nil
		}
		
		// Handle Ctrl+C globally
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle Quit Confirmation logic first
		if m.State == ConfirmingQuitState {
			switch msg.String() {
			case "y", "Y", "enter":
				return m, tea.Quit
			case "n", "N", "q", "esc":
				m.State = ZoneListState // Fallback, previous state would be better but let's keep it simple
				// To be better, we could track 'previousState', but for now ZoneList is safe.
				// Let's actually check if we have a selected zone to decide.
				if m.SelectedID != "" {
					m.State = RecordListState
				} else {
					m.State = ZoneListState
				}
				return m, nil
			}
		}

		// If we are filtering, we let the list handle EVERYTHING.
		if m.State == ZoneListState && m.ZoneList.FilterState() == list.Filtering {
			m.ZoneList, cmd = m.ZoneList.Update(msg)
			return m, cmd
		}
		if m.State == RecordListState && m.RecordList.FilterState() == list.Filtering {
			m.RecordList, cmd = m.RecordList.Update(msg)
			return m, cmd
		}

		// Trigger quit confirmation on 'q'
		if msg.String() == "q" && m.State != EditingRecordState {
			m.State = ConfirmingQuitState
			return m, nil
		}

	case FetchedZonesMsg:
		m.Logger.Info("Fetched zones", "count", len(msg))
		items := make([]list.Item, len(msg))
		for i, z := range msg {
			items[i] = ZoneItem{ID: z.ID, Name: z.Name}
		}
		m.ZoneList.SetItems(items)
		m.State = ZoneListState
		return m, nil

	case FetchedRecordsMsg:
		m.Logger.Info("Fetched records", "count", len(msg), "zoneID", m.SelectedID)
		items := make([]list.Item, len(msg))
		for i, r := range msg {
			items[i] = RecordItem{DNS: r}
		}
		m.RecordList.SetItems(items)
		m.State = RecordListState
		return m, nil

	case RecordSavedMsg:
		m.Logger.Info("Record saved successfully")
		m.State = LoadingRecordsState
		return m, tea.Batch(FetchRecords(m.CfClient, m.SelectedID), m.Spinner.Tick)

	case RecordDeletedMsg:
		m.Logger.Info("Record deleted successfully", "id", m.PendingDeleteID)
		m.State = LoadingRecordsState
		return m, tea.Batch(FetchRecords(m.CfClient, m.SelectedID), m.Spinner.Tick)

	case ErrorMsg:
		m.Logger.Error("API Error", "error", msg.Error())
		m.Err = msg
		return m, nil

	case tea.WindowSizeMsg:
		h, v := DocStyle.GetFrameSize()
		m.ZoneList.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		m.RecordList.SetSize(msg.Width-h, msg.Height-v-helpHeight)
	}

	// Update the spinner
	m.Spinner, cmd = m.Spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch m.State {
	case ZoneListState:
		m.ZoneList, cmd = m.ZoneList.Update(msg)
		cmds = append(cmds, cmd)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			if i, ok := m.ZoneList.SelectedItem().(ZoneItem); ok {
				m.SelectedID = i.ID
				m.State = LoadingRecordsState
				m.RecordList.Title = "DNS Records: " + i.Name
				m.Logger.Info("Selecting zone", "name", i.Name, "id", i.ID)
				return m, tea.Batch(FetchRecords(m.CfClient, i.ID), m.Spinner.Tick)
			}
		}

	case RecordListState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc", "backspace":
				m.State = ZoneListState
				return m, nil
			case "a":
				m.Logger.Debug("Opening add record form")
				m.Form = NewRecordForm(nil, m.Theme)
				m.OldRecord = nil
				m.State = EditingRecordState
				return m, nil
			case "enter":
				if i, ok := m.RecordList.SelectedItem().(RecordItem); ok {
					m.Logger.Debug("Opening edit record form", "id", i.DNS.ID)
					m.Form = NewRecordForm(&i.DNS, m.Theme)
					m.OldRecord = &i.DNS
					m.State = EditingRecordState
					return m, nil
				}
			case "d":
				if i, ok := m.RecordList.SelectedItem().(RecordItem); ok {
					m.PendingDeleteID = i.DNS.ID
					m.PendingDeleteName = i.DNS.Name
					m.State = ConfirmingDeleteState
					return m, nil
				}
			}
		}
		m.RecordList, cmd = m.RecordList.Update(msg)
		cmds = append(cmds, cmd)

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

				if m.Form.Focused > 4 {
					m.Form.Focused = 0
				} else if m.Form.Focused < 0 {
					m.Form.Focused = 4
				}

				for i := range m.Form.Inputs {
					if i == m.Form.Focused {
						cmds = append(cmds, m.Form.Inputs[i].Focus())
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
					if m.Form.Inputs[0].Value() == "" || m.Form.Inputs[1].Value() == "" || m.Form.Inputs[2].Value() == "" {
						m.Err = fmt.Errorf("all fields (Type, Name, Content) are required")
						return m, nil
					}
					m.State = ConfirmingSaveState
					return m, nil
				}
			}
		}

		for i := range m.Form.Inputs {
			m.Form.Inputs[i], cmd = m.Form.Inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}

	case ConfirmingSaveState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y", "enter":
				m.Logger.Info("Saving record", "id", m.Form.ID)
				return m, SaveRecord(m.CfClient, m.SelectedID, m.Form)
			case "n", "N", "esc":
				m.State = EditingRecordState
				return m, nil
			}
		}

	case ConfirmingDeleteState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y", "enter":
				m.Logger.Info("Deleting record", "id", m.PendingDeleteID)
				return m, DeleteRecord(m.CfClient, m.SelectedID, m.PendingDeleteID)
			case "n", "N", "esc":
				m.State = RecordListState
				return m, nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	errStyle := lipgloss.NewStyle().Foreground(m.Theme.Error).Bold(true)
	focusedStyle := lipgloss.NewStyle().Foreground(m.Theme.Primary)
	confirmStyle := lipgloss.NewStyle().Foreground(m.Theme.Warning).Bold(true)
	diffOld := lipgloss.NewStyle().Foreground(m.Theme.Error).Strikethrough(true)
	diffNew := lipgloss.NewStyle().Foreground(m.Theme.Secondary)

	if m.Err != nil {
		return DocStyle.Render(fmt.Sprintf("%s\n\nPress any key to continue.", errStyle.Render("Error: "+m.Err.Error())))
	}

	switch m.State {
	case ConfirmingQuitState:
		return DocStyle.Render(confirmStyle.Render("Are you sure you want to quit? (y/n)"))

	case LoadingZonesState:
		return DocStyle.Render(fmt.Sprintf("%s Loading zones from Cloudflare...", m.Spinner.View()))
	case ZoneListState:
		return DocStyle.Render(m.ZoneList.View())
	case LoadingRecordsState:
		return DocStyle.Render(fmt.Sprintf("%s Loading DNS records for %s...", m.Spinner.View(), m.SelectedID))
	case RecordListState:
		view := m.RecordList.View()
		help := lipgloss.NewStyle().Foreground(m.Theme.Inactive).MarginTop(1).Render("(a) add record, (enter) edit record, (d) delete record, (esc) back, (q) quit")
		return DocStyle.Render(view + "\n" + help)
	case EditingRecordState:
		var b strings.Builder
		title := "Adding DNS Record"
		if m.Form.ID != "" {
			title = "Editing DNS Record"
		}
		b.WriteString(confirmStyle.Render(title) + "\n\n")

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
			b.WriteString("\n\n" + focusedStyle.Render(proxiedStr))
		} else {
			b.WriteString("\n\n" + proxiedStr)
		}

		saveStr := "Save"
		if m.Form.Focused == 4 {
			b.WriteString("\n\n" + focusedStyle.Render("["+saveStr+"]"))
		} else {
			b.WriteString("\n\n[" + saveStr + "]")
		}

		b.WriteString("\n\n(esc) cancel")

		return DocStyle.Render(b.String())

	case ConfirmingSaveState:
		var b strings.Builder
		b.WriteString(confirmStyle.Render("Review Changes") + "\n\n")
		
		if m.OldRecord != nil {
			// Change View
			b.WriteString(fmt.Sprintf("Type:    %s -> %s\n", diffOld.Render(m.OldRecord.Type), diffNew.Render(m.Form.Inputs[0].Value())))
			b.WriteString(fmt.Sprintf("Name:    %s -> %s\n", diffOld.Render(m.OldRecord.Name), diffNew.Render(m.Form.Inputs[1].Value())))
			b.WriteString(fmt.Sprintf("Content: %s -> %s\n", diffOld.Render(m.OldRecord.Content), diffNew.Render(m.Form.Inputs[2].Value())))
			
			oldProxied := "No"
			if m.OldRecord.Proxied != nil && *m.OldRecord.Proxied { oldProxied = "Yes" }
			newProxied := "No"
			if m.Form.Proxied { newProxied = "Yes" }
			b.WriteString(fmt.Sprintf("Proxied: %s -> %s\n", diffOld.Render(oldProxied), diffNew.Render(newProxied)))
		} else {
			// New Record View
			b.WriteString(fmt.Sprintf("Type:    %s\n", diffNew.Render(m.Form.Inputs[0].Value())))
			b.WriteString(fmt.Sprintf("Name:    %s\n", diffNew.Render(m.Form.Inputs[1].Value())))
			b.WriteString(fmt.Sprintf("Content: %s\n", diffNew.Render(m.Form.Inputs[2].Value())))
			proxied := "No"
			if m.Form.Proxied { proxied = "Yes" }
			b.WriteString(fmt.Sprintf("Proxied: %s\n", diffNew.Render(proxied)))
		}

		b.WriteString("\n" + confirmStyle.Render("Confirm save? (y/n)"))
		return DocStyle.Render(b.String())

	case ConfirmingDeleteState:
		var b strings.Builder
		b.WriteString(confirmStyle.Render("Confirm Deletion") + "\n\n")
		b.WriteString(fmt.Sprintf("Record: %s\nID:     %s\n\n", m.PendingDeleteName, m.PendingDeleteID))
		b.WriteString(errStyle.Render("Are you sure you want to delete this record? (y/n)"))
		return DocStyle.Render(b.String())

	default:
		return "Unknown state"
	}
}
