package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudflare/cloudflare-go"
)

// RecordItem implements list.Item for the DNS Record view.
type RecordItem struct {
	DNS cloudflare.DNSRecord
}

func (i RecordItem) Title() string {
	return fmt.Sprintf("%-6s %s", i.DNS.Type, i.DNS.Name)
}
func (i RecordItem) Description() string {
	proxied := "No"
	if i.DNS.Proxied != nil && *i.DNS.Proxied {
		proxied = "Yes"
	}
	return fmt.Sprintf("Content: %s | Proxied: %s", i.DNS.Content, proxied)
}
func (i RecordItem) FilterValue() string { return i.DNS.Name }

// NewRecordForm initializes a form for a DNS record.
func NewRecordForm(r *cloudflare.DNSRecord, theme Theme) RecordForm {
	var f RecordForm
	f.Inputs = make([]textinput.Model, 3)

	for i := range f.Inputs {
		f.Inputs[i] = textinput.New()
		f.Inputs[i].PromptStyle = f.Inputs[i].PromptStyle.Foreground(theme.Primary)
		f.Inputs[i].TextStyle = f.Inputs[i].TextStyle.Foreground(theme.Secondary)
	}

	f.Inputs[0].Placeholder = "Type (A, CNAME, etc.)"
	f.Inputs[0].Focus()

	f.Inputs[1].Placeholder = "Name"
	f.Inputs[2].Placeholder = "Content"

	if r != nil {
		f.ID = r.ID
		f.Inputs[0].SetValue(r.Type)
		f.Inputs[1].SetValue(r.Name)
		f.Inputs[2].SetValue(r.Content)
		if r.Proxied != nil {
			f.Proxied = *r.Proxied
		}
	}

	return f
}

// FetchRecords returns a tea.Cmd that fetches DNS records for a specific zone.
func FetchRecords(api *cloudflare.API, zoneID string) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return ErrorMsg(err)
		}
		return FetchedRecordsMsg(records)
	}
}

// SaveRecord returns a tea.Cmd that saves (creates or updates) a DNS record.
func SaveRecord(api *cloudflare.API, zoneID string, f RecordForm) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		proxied := f.Proxied

		var err error
		if f.ID == "" {
			_, err = api.CreateDNSRecord(context.Background(), rc, cloudflare.CreateDNSRecordParams{
				Type:    f.Inputs[0].Value(),
				Name:    f.Inputs[1].Value(),
				Content: f.Inputs[2].Value(),
				Proxied: &proxied,
			})
		} else {
			_, err = api.UpdateDNSRecord(context.Background(), rc, cloudflare.UpdateDNSRecordParams{
				ID:      f.ID,
				Type:    f.Inputs[0].Value(),
				Name:    f.Inputs[1].Value(),
				Content: f.Inputs[2].Value(),
				Proxied: &proxied,
			})
		}

		if err != nil {
			return ErrorMsg(err)
		}
		return RecordSavedMsg{}
	}
}

// DeleteRecord returns a tea.Cmd that deletes a DNS record.
func DeleteRecord(api *cloudflare.API, zoneID string, recordID string) tea.Cmd {
	return func() tea.Msg {
		rc := cloudflare.ZoneIdentifier(zoneID)
		err := api.DeleteDNSRecord(context.Background(), rc, recordID)
		if err != nil {
			return ErrorMsg(err)
		}
		return RecordDeletedMsg{}
	}
}
