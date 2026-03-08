package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpHeight = 2

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		FetchZones(m.CfClient, m.Logger),
		m.Spinner.Tick,
	)
}

func (m *Model) Close() {
	if m.LogFile != nil {
		_ = m.LogFile.Close()
	}
}

func (m *Model) updateList(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.State {
	case ZoneListState:
		m.ZoneList, cmd = m.ZoneList.Update(msg)
	case RecordListState:
		m.RecordList, cmd = m.RecordList.Update(msg)
	case PickingTypeState:
		m.Form.TypeList, cmd = m.Form.TypeList.Update(msg)
	}
	return cmd
}

func (m *Model) isProxiedSupported() bool {
	t := strings.ToUpper(m.Form.Type)
	return t == "A" || t == "AAAA" || t == "CNAME"
}

func (m *Model) isFlattenSupported() bool {
	return strings.EqualFold(m.Form.Type, "CNAME")
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// Handle states that ignore standard key processing
		if m.State == ConfirmingQuitState {
			switch msg.String() {
			case "y", "Y", "enter":
				return m, tea.Quit
			case "n", "N", "q", "esc":
				if m.SelectedID != "" {
					m.State = RecordListState
				} else {
					m.State = ZoneListState
				}
				return m, nil
			}
		}

		// If we are filtering, we let the list handle EVERYTHING.
		if (m.State == ZoneListState && m.ZoneList.FilterState() == list.Filtering) ||
			(m.State == RecordListState && m.RecordList.FilterState() == list.Filtering) ||
			(m.State == PickingTypeState && m.Form.TypeList.FilterState() == list.Filtering) {
			if msg.String() != "esc" {
				cmd = m.updateList(msg)
				return m, cmd
			}
		}

		// Trigger quit confirmation on 'q' (handled specifically in states below for more precision)

	case FetchedZonesMsg:
		m.Logger.Info("Fetched zones", "count", len(msg))
		items := make([]list.Item, len(msg))
		for i := range msg {
			items[i] = &ZoneItem{ID: msg[i].ID, Name: msg[i].Name}
		}
		m.ZoneList.SetItems(items)
		m.State = ZoneListState
		return m, nil

	case FetchedRecordsMsg:
		m.Logger.Info("Fetched records", "count", len(msg), "zoneID", m.SelectedID)
		items := make([]list.Item, len(msg))
		for i := range msg {
			items[i] = &RecordItem{DNS: msg[i]}
		}
		m.RecordList.SetItems(items)
		m.State = RecordListState
		return m, nil

	case RecordSavedMsg:
		m.Logger.Info("Record saved successfully")
		m.State = LoadingRecordsState
		return m, tea.Batch(FetchRecords(m.CfClient, m.SelectedID, m.Logger), m.Spinner.Tick)

	case RecordDeletedMsg:
		m.Logger.Info("Record deleted successfully", "id", m.PendingDeleteID)
		m.State = LoadingRecordsState
		return m, tea.Batch(FetchRecords(m.CfClient, m.SelectedID, m.Logger), m.Spinner.Tick)

	case ErrorMsg:
		m.Logger.Error("API Error", "error", msg.Error())
		m.Err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		h, v := DocStyle.GetFrameSize()
		m.ZoneList.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		m.RecordList.SetSize(msg.Width-h, msg.Height-v-helpHeight)
		m.Form.TypeList.SetSize(msg.Width-h, msg.Height-v-helpHeight)
	}

	// Update the spinner
	m.Spinner, cmd = m.Spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch m.State {
	case ZoneListState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q", "esc":
				if m.ZoneList.FilterState() == list.Unfiltered {
					m.State = ConfirmingQuitState
					return m, nil
				}
			}
		}
		cmd = m.updateList(msg)
		cmds = append(cmds, cmd)
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			if i, ok := m.ZoneList.SelectedItem().(*ZoneItem); ok {
				m.SelectedID = i.ID
				m.State = LoadingRecordsState
				m.RecordList.Title = "DNS Records: " + i.Name
				m.Logger.Info("Selecting zone", "name", i.Name, "id", i.ID)
				return m, tea.Batch(FetchRecords(m.CfClient, i.ID, m.Logger), m.Spinner.Tick)
			}
		}

	case RecordListState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q":
				if m.RecordList.FilterState() == list.Unfiltered {
					m.State = ConfirmingQuitState
					return m, nil
				}
			case "esc", "backspace":
				if m.RecordList.FilterState() == list.Unfiltered {
					m.State = ZoneListState
					return m, nil
				}
			case "a":
				m.Logger.Debug("Opening add record form")
				m.Form = NewRecordForm(nil, &m.Theme)
				h, v := DocStyle.GetFrameSize()
				m.Form.TypeList.SetSize(m.Width-h, m.Height-v-helpHeight)
				m.OldRecord = nil
				m.State = EditingRecordState
				return m, nil
			case "enter":
				if i, ok := m.RecordList.SelectedItem().(*RecordItem); ok {
					m.Logger.Debug("Opening edit record form", "id", i.DNS.ID)
					m.Form = NewRecordForm(&i.DNS, &m.Theme)
					h, v := DocStyle.GetFrameSize()
					m.Form.TypeList.SetSize(m.Width-h, m.Height-v-helpHeight)
					m.OldRecord = &i.DNS
					m.State = EditingRecordState
					return m, nil
				}
			case "d":
				if i, ok := m.RecordList.SelectedItem().(*RecordItem); ok {
					m.PendingDeleteID = i.DNS.ID
					m.PendingDeleteName = i.DNS.Name
					m.State = ConfirmingDeleteState
					return m, nil
				}
			}
		}
		cmd = m.updateList(msg)
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

				totalElements := len(m.Form.Inputs) + 4
				if m.Form.Focused >= totalElements {
					m.Form.Focused = 0
				} else if m.Form.Focused < 0 {
					m.Form.Focused = totalElements - 1
				}

				if m.Form.Focused == len(m.Form.Inputs)+1 && !m.isProxiedSupported() {
					if s == "up" || s == "shift+tab" {
						m.Form.Focused--
					} else {
						m.Form.Focused++
					}
				}
				if m.Form.Focused == len(m.Form.Inputs)+2 && !m.isFlattenSupported() {
					if s == "up" || s == "shift+tab" {
						m.Form.Focused--
					} else {
						m.Form.Focused++
					}
				}

				for i := range m.Form.Inputs {
					if i == m.Form.Focused-1 {
						cmds = append(cmds, m.Form.Inputs[i].Focus())
					} else {
						m.Form.Inputs[i].Blur()
					}
				}
				return m, tea.Batch(cmds...)

			case "enter":
				if m.Form.Focused == 0 {
					m.State = PickingTypeState
					return m, nil
				}
				if m.Form.Focused == len(m.Form.Inputs)+3 { // Save Button
					if m.Form.Inputs[0].Value() == "" {
						m.Err = fmt.Errorf("name is required")
						return m, nil
					}
					m.State = ConfirmingSaveState
					return m, nil
				}
			case " ":
				if m.Form.Focused == len(m.Form.Inputs)+1 && m.isProxiedSupported() {
					m.Form.Proxied = !m.Form.Proxied
					return m, nil
				}
				if m.Form.Focused == len(m.Form.Inputs)+2 && m.isFlattenSupported() {
					m.Form.FlattenCNAME = !m.Form.FlattenCNAME
					return m, nil
				}
			}
		}

		for i := range m.Form.Inputs {
			m.Form.Inputs[i], cmd = m.Form.Inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}

	case PickingTypeState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc":
				m.State = EditingRecordState
				return m, nil
			case "enter":
				if i, ok := m.Form.TypeList.SelectedItem().(typeItem); ok {
					m.Form.Type = string(i)
					m.Form.initializeInputs(m.OldRecord, &m.Theme)
					m.State = EditingRecordState
					return m, nil
				}
			}
		}
		cmd = m.updateList(msg)
		cmds = append(cmds, cmd)

	case ConfirmingSaveState:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y", "enter":
				m.Logger.Info("Saving record", "id", m.Form.ID)
				return m, SaveRecord(m.CfClient, m.SelectedID, &m.Form, m.Logger)
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
				return m, DeleteRecord(m.CfClient, m.SelectedID, m.PendingDeleteID, m.Logger)
			case "n", "N", "esc":
				m.State = RecordListState
				return m, nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
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

	case PickingTypeState:
		return DocStyle.Render(m.Form.TypeList.View())

	case EditingRecordState:
		var b strings.Builder
		title := "Adding DNS Record"
		if m.Form.ID != "" {
			title = "Editing DNS Record"
		}
		fmt.Fprintf(&b, "%s\n\n", confirmStyle.Render(title))

		typeStr := fmt.Sprintf("Type: %s (Press Enter to change)", m.Form.Type)
		if m.Form.Focused == 0 {
			fmt.Fprintf(&b, "%s\n\n", focusedStyle.Render(typeStr))
		} else {
			fmt.Fprintf(&b, "%s\n\n", typeStr)
		}

		for i := range m.Form.Inputs {
			b.WriteString(m.Form.Inputs[i].View())
			if i < len(m.Form.Inputs)-1 {
				b.WriteRune('\n')
			}
		}

		if m.isProxiedSupported() {
			proxiedStr := "[ ] Proxied"
			if m.Form.Proxied {
				proxiedStr = "[x] Proxied"
			}
			if m.Form.Focused == len(m.Form.Inputs)+1 {
				fmt.Fprintf(&b, "\n\n%s", focusedStyle.Render(proxiedStr))
			} else {
				fmt.Fprintf(&b, "\n\n%s", proxiedStr)
			}
		}

		if m.isFlattenSupported() {
			flattenStr := "[ ] Flatten CNAME"
			if m.Form.FlattenCNAME {
				flattenStr = "[x] Flatten CNAME"
			}
			if m.Form.Focused == len(m.Form.Inputs)+2 {
				fmt.Fprintf(&b, "\n\n%s", focusedStyle.Render(flattenStr))
			} else {
				fmt.Fprintf(&b, "\n\n%s", flattenStr)
			}
		}

		saveStr := "Save"
		if m.Form.Focused == len(m.Form.Inputs)+3 {
			fmt.Fprintf(&b, "\n\n%s", focusedStyle.Render("["+saveStr+"]"))
		} else {
			fmt.Fprintf(&b, "\n\n[%s]", saveStr)
		}

		b.WriteString("\n\n(esc) cancel")
		return DocStyle.Render(b.String())

	case ConfirmingSaveState:
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n\n", confirmStyle.Render("Review Changes"))

		typeOld := ""
		if m.OldRecord != nil {
			typeOld = m.OldRecord.Type
		}
		if typeOld != "" && typeOld != m.Form.Type {
			fmt.Fprintf(&b, "Type:    %s -> %s\n", diffOld.Render(typeOld), diffNew.Render(m.Form.Type))
		} else {
			fmt.Fprintf(&b, "Type:    %s\n", m.Form.Type)
		}

		for i := range m.Form.Inputs {
			fmt.Fprintf(&b, "%s %s\n", m.Form.Inputs[i].Prompt, diffNew.Render(m.Form.Inputs[i].Value()))
		}

		if m.isProxiedSupported() {
			oldProxied := "No"
			if m.OldRecord != nil && m.OldRecord.Proxied != nil && *m.OldRecord.Proxied {
				oldProxied = "Yes"
			}
			newProxied := "No"
			if m.Form.Proxied {
				newProxied = "Yes"
			}
			if m.OldRecord != nil {
				fmt.Fprintf(&b, "Proxied: %s -> %s\n", diffOld.Render(oldProxied), diffNew.Render(newProxied))
			} else {
				fmt.Fprintf(&b, "Proxied: %s\n", diffNew.Render(newProxied))
			}
		}

		if m.isFlattenSupported() {
			oldFlat := "No"
			if m.OldRecord != nil && m.OldRecord.Settings.FlattenCNAME != nil && *m.OldRecord.Settings.FlattenCNAME {
				oldFlat = "Yes"
			}
			newFlat := "No"
			if m.Form.FlattenCNAME {
				newFlat = "Yes"
			}
			if m.OldRecord != nil {
				fmt.Fprintf(&b, "Flatten: %s -> %s\n", diffOld.Render(oldFlat), diffNew.Render(newFlat))
			} else {
				fmt.Fprintf(&b, "Flatten: %s\n", diffNew.Render(newFlat))
			}
		}

		fmt.Fprintf(&b, "\n%s", confirmStyle.Render("Confirm save? (y/n)"))
		return DocStyle.Render(b.String())

	case ConfirmingDeleteState:
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n\n", confirmStyle.Render("Confirm Deletion"))
		fmt.Fprintf(&b, "Record: %s\nID:     %s\n\n", m.PendingDeleteName, m.PendingDeleteID)
		b.WriteString(errStyle.Render("Are you sure you want to delete this record? (y/n)"))
		return DocStyle.Render(b.String())

	default:
		return "Unknown state"
	}
}
