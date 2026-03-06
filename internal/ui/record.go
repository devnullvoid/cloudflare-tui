package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/cloudflare/cloudflare-go"
)

// typeItem implements list.Item for the Record Type Picker.
type typeItem string

func (i typeItem) Title() string       { return string(i) }
func (i typeItem) Description() string { return "" }
func (i typeItem) FilterValue() string { return string(i) }

var validRecordTypes = []list.Item{
	typeItem("A"), typeItem("AAAA"), typeItem("CAA"), typeItem("CERT"),
	typeItem("CNAME"), typeItem("DNSKEY"), typeItem("DS"), typeItem("HTTPS"),
	typeItem("LOC"), typeItem("MX"), typeItem("NAPTR"), typeItem("NS"),
	typeItem("PTR"), typeItem("SMIMEA"), typeItem("SRV"), typeItem("SSHFP"),
	typeItem("SVCB"), typeItem("TLSA"), typeItem("TXT"), typeItem("URI"),
}

// RecordItem implements list.Item for the DNS Record view.
type RecordItem struct {
	DNS cloudflare.DNSRecord
}

func (i *RecordItem) Title() string {
	return fmt.Sprintf("%s [%s]", i.DNS.Name, i.DNS.Type)
}
func (i *RecordItem) Description() string {
	proxied := "No"
	if i.DNS.Proxied != nil && *i.DNS.Proxied {
		proxied = "Yes"
	}
	content := i.DNS.Content
	if i.DNS.Type == "SRV" && i.DNS.Data != nil {
		// Simplified display for SRV data if content is empty
		content = fmt.Sprintf("%v", i.DNS.Data)
	}
	return fmt.Sprintf("Content: %s | Proxied: %s | TTL: %d", content, proxied, i.DNS.TTL)
}
func (i *RecordItem) FilterValue() string { return i.DNS.Name }

// NewRecordForm initializes a form for a DNS record.
func NewRecordForm(r *cloudflare.DNSRecord, theme *Theme) RecordForm {
	var f RecordForm

	// Initialize Type List with theme
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(theme.Primary).
		BorderLeftForeground(theme.Primary)

	f.TypeList = list.New(validRecordTypes, delegate, 0, 0)
	f.TypeList.Title = "Select Record Type"
	f.TypeList.SetFilteringEnabled(true)
	f.TypeList.Styles.Title = f.TypeList.Styles.Title.
		Background(theme.Primary).
		Foreground(lipgloss.Color("#1e1e2e"))

	// Themed Filter Input for Type Picker
	f.TypeList.FilterInput.PromptStyle = f.TypeList.FilterInput.PromptStyle.Foreground(theme.Primary)
	f.TypeList.FilterInput.TextStyle = f.TypeList.FilterInput.TextStyle.Foreground(theme.Secondary)
	f.TypeList.FilterInput.Cursor.Style = f.TypeList.FilterInput.Cursor.Style.Foreground(theme.Primary)

	// Default to A if new
	f.Type = "A"
	if r != nil {
		f.ID = r.ID
		f.Type = r.Type
		f.Proxied = r.Proxied != nil && *r.Proxied
	}

	f.initializeInputs(r, theme)
	return f
}

func (f *RecordForm) initializeInputs(r *cloudflare.DNSRecord, theme *Theme) {
	// Define fields based on type
	type field struct {
		label       string
		placeholder string
		value       string
	}

	var fields []field
	fields = append(fields, field{label: "Name", placeholder: "e.g. @ or www", value: ""})

	switch f.Type {
	case "MX":
		fields = append(fields,
			field{label: "Mail Server", placeholder: "e.g. mail.example.com", value: ""},
			field{label: "Priority", placeholder: "e.g. 10", value: ""},
		)
	case "SRV":
		fields = append(fields,
			field{label: "Service", placeholder: "e.g. _sip", value: ""},
			field{label: "Proto", placeholder: "e.g. _tcp", value: ""},
			field{label: "Priority", placeholder: "e.g. 0", value: ""},
			field{label: "Weight", placeholder: "e.g. 5", value: ""},
			field{label: "Port", placeholder: "e.g. 5060", value: ""},
			field{label: "Target", placeholder: "e.g. sip.example.com", value: ""},
		)
	case "CAA":
		fields = append(fields,
			field{label: "Tag", placeholder: "e.g. issue", value: ""},
			field{label: "Flags", placeholder: "e.g. 0", value: ""},
			field{label: "Value", placeholder: "e.g. letsencrypt.org", value: ""},
		)
	default:
		fields = append(fields, field{label: "Content", placeholder: "e.g. 1.2.3.4", value: ""})
	}

	fields = append(fields, field{label: "TTL", placeholder: "1 = auto", value: "1"})

	f.Inputs = make([]textinput.Model, len(fields))
	for i, fld := range fields {
		f.Inputs[i] = textinput.New()
		f.Inputs[i].Prompt = fld.label + ": "
		f.Inputs[i].Placeholder = fld.placeholder
		f.Inputs[i].SetValue(fld.value)
		f.Inputs[i].PromptStyle = f.Inputs[i].PromptStyle.Foreground(theme.Primary)
		f.Inputs[i].TextStyle = f.Inputs[i].TextStyle.Foreground(theme.Secondary)
		f.Inputs[i].Blur()
	}

	// Fill values if editing
	if r != nil {
		f.Inputs[0].SetValue(r.Name) // Name is always index 0

		switch f.Type {
		case "MX":
			f.Inputs[1].SetValue(r.Content)
			if r.Priority != nil {
				f.Inputs[2].SetValue(strconv.Itoa(int(*r.Priority)))
			}
			f.Inputs[3].SetValue(strconv.Itoa(r.TTL))
		case "SRV":
			if data, ok := r.Data.(map[string]interface{}); ok {
				f.Inputs[1].SetValue(fmt.Sprintf("%v", data["service"]))
				f.Inputs[2].SetValue(fmt.Sprintf("%v", data["proto"]))
				f.Inputs[3].SetValue(fmt.Sprintf("%v", data["priority"]))
				f.Inputs[4].SetValue(fmt.Sprintf("%v", data["weight"]))
				f.Inputs[5].SetValue(fmt.Sprintf("%v", data["port"]))
				f.Inputs[6].SetValue(fmt.Sprintf("%v", data["target"]))
			}
			f.Inputs[7].SetValue(strconv.Itoa(r.TTL))
		case "CAA":
			if data, ok := r.Data.(map[string]interface{}); ok {
				f.Inputs[1].SetValue(fmt.Sprintf("%v", data["tag"]))
				f.Inputs[2].SetValue(fmt.Sprintf("%v", data["flags"]))
				f.Inputs[3].SetValue(fmt.Sprintf("%v", data["value"]))
			}
			f.Inputs[4].SetValue(strconv.Itoa(r.TTL))
		default:
			f.Inputs[1].SetValue(r.Content)
			f.Inputs[2].SetValue(strconv.Itoa(r.TTL))
		}
	}

	f.Inputs[0].Focus()
	f.Focused = 1 // Start focus on the first input (after Type trigger)
}

// FetchRecords returns a tea.Cmd that fetches DNS records for a specific zone.
func FetchRecords(api *cloudflare.API, zoneID string, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		logger.Debug("Initiating ListDNSRecords API call", "zoneID", zoneID)
		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			logger.Error("ListDNSRecords API call failed", "error", err, "zoneID", zoneID)
			return ErrorMsg(err)
		}
		logger.Debug("ListDNSRecords API call successful", "count", len(records), "zoneID", zoneID)
		return FetchedRecordsMsg(records)
	}
}

// SaveRecord returns a tea.Cmd that saves (creates or updates) a DNS record.
func SaveRecord(api *cloudflare.API, zoneID string, f *RecordForm, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)

		// Map inputs back to cloudflare params
		var name, content string
		var ttl int
		var priority *uint16
		var data interface{}

		name = f.Inputs[0].Value()

		switch f.Type {
		case "MX":
			content = f.Inputs[1].Value()
			p, _ := strconv.Atoi(f.Inputs[2].Value())
			up := uint16(p)
			priority = &up
			ttl, _ = strconv.Atoi(f.Inputs[3].Value())
		case "SRV":
			p, _ := strconv.Atoi(f.Inputs[3].Value())
			w, _ := strconv.Atoi(f.Inputs[4].Value())
			po, _ := strconv.Atoi(f.Inputs[5].Value())
			data = map[string]interface{}{
				"service":  f.Inputs[1].Value(),
				"proto":    f.Inputs[2].Value(),
				"priority": p,
				"weight":   w,
				"port":     po,
				"target":   f.Inputs[6].Value(),
				"name":     name,
			}
			ttl, _ = strconv.Atoi(f.Inputs[7].Value())
		case "CAA":
			fl, _ := strconv.Atoi(f.Inputs[2].Value())
			data = map[string]interface{}{
				"tag":   f.Inputs[1].Value(),
				"flags": fl,
				"value": f.Inputs[3].Value(),
			}
			ttl, _ = strconv.Atoi(f.Inputs[4].Value())
		default:
			content = f.Inputs[1].Value()
			ttl, _ = strconv.Atoi(f.Inputs[2].Value())
		}

		params := cloudflare.UpdateDNSRecordParams{
			ID:       f.ID,
			Type:     f.Type,
			Name:     name,
			Content:  content,
			TTL:      ttl,
			Priority: priority,
			Proxied:  &f.Proxied,
			Data:     data,
		}

		var err error
		if f.ID == "" {
			logger.Debug("Initiating CreateDNSRecord API call", "zoneID", zoneID, "name", name)
			_, err = api.CreateDNSRecord(context.Background(), rc, cloudflare.CreateDNSRecordParams{
				Type:     params.Type,
				Name:     params.Name,
				Content:  params.Content,
				TTL:      params.TTL,
				Priority: params.Priority,
				Proxied:  params.Proxied,
				Data:     params.Data,
			})
		} else {
			logger.Debug("Initiating UpdateDNSRecord API call", "zoneID", zoneID, "recordID", f.ID)
			_, err = api.UpdateDNSRecord(context.Background(), rc, params)
		}

		if err != nil {
			logger.Error("SaveRecord API call failed", "error", err, "zoneID", zoneID, "recordID", f.ID)
			return ErrorMsg(err)
		}
		logger.Debug("SaveRecord API call successful", "zoneID", zoneID, "recordID", f.ID)
		return RecordSavedMsg{}
	}
}

// DeleteRecord returns a tea.Cmd that deletes a DNS record.
func DeleteRecord(api *cloudflare.API, zoneID, recordID string, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		logger.Debug("Initiating DeleteDNSRecord API call", "zoneID", zoneID, "recordID", recordID)
		rc := cloudflare.ZoneIdentifier(zoneID)
		err := api.DeleteDNSRecord(context.Background(), rc, recordID)
		if err != nil {
			logger.Error("DeleteDNSRecord API call failed", "error", err, "zoneID", zoneID, "recordID", recordID)
			return ErrorMsg(err)
		}
		logger.Debug("DeleteDNSRecord API call successful", "zoneID", zoneID, "recordID", recordID)
		return RecordDeletedMsg{}
	}
}
