package ui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	return fmt.Sprintf("Content: %s | Proxied: %s | TTL: %d", i.DNS.Content, proxied, i.DNS.TTL)
}
func (i *RecordItem) FilterValue() string { return i.DNS.Name }

// NewRecordForm initializes a form for a DNS record.
func NewRecordForm(r *cloudflare.DNSRecord, theme *Theme) RecordForm {
	var f RecordForm

	// Initialize Type List
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	f.TypeList = list.New(validRecordTypes, delegate, 0, 0)
	f.TypeList.Title = "Select Record Type"
	f.TypeList.SetFilteringEnabled(true)

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
	// Fields: Name, Content, TTL, Priority (if MX)
	numFields := 3
	if f.Type == "MX" {
		numFields = 4
	}

	f.Inputs = make([]textinput.Model, numFields)
	for i := range f.Inputs {
		f.Inputs[i] = textinput.New()
		f.Inputs[i].PromptStyle = f.Inputs[i].PromptStyle.Foreground(theme.Primary)
		f.Inputs[i].TextStyle = f.Inputs[i].TextStyle.Foreground(theme.Secondary)
	}

	f.Inputs[0].Placeholder = "Name"
	f.Inputs[1].Placeholder = "Content"
	f.Inputs[2].Placeholder = "TTL (1 = auto)"
	f.Inputs[2].CharLimit = 10

	if f.Type == "MX" {
		f.Inputs[3].Placeholder = "Priority"
		f.Inputs[3].CharLimit = 5
	}

	if r != nil {
		f.Inputs[0].SetValue(r.Name)
		f.Inputs[1].SetValue(r.Content)
		f.Inputs[2].SetValue(strconv.Itoa(r.TTL))
		if f.Type == "MX" && r.Priority != nil {
			f.Inputs[3].SetValue(strconv.Itoa(int(*r.Priority)))
		}
	} else {
		f.Inputs[2].SetValue("1") // Default TTL auto
	}

	// Start with Name focused
	f.Inputs[0].Focus()
	f.Focused = 0
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

		ttl, _ := strconv.Atoi(f.Inputs[2].Value())
		params := cloudflare.UpdateDNSRecordParams{
			ID:      f.ID,
			Type:    f.Type,
			Name:    f.Inputs[0].Value(),
			Content: f.Inputs[1].Value(),
			TTL:     ttl,
			Proxied: &f.Proxied,
		}

		if f.Type == "MX" && len(f.Inputs) > 3 {
			p, _ := strconv.Atoi(f.Inputs[3].Value())
			up := uint16(p)
			params.Priority = &up
		}

		var err error
		if f.ID == "" {
			logger.Debug("Initiating CreateDNSRecord API call", "zoneID", zoneID, "name", params.Name)
			_, err = api.CreateDNSRecord(context.Background(), rc, cloudflare.CreateDNSRecordParams{
				Type:     params.Type,
				Name:     params.Name,
				Content:  params.Content,
				TTL:      params.TTL,
				Priority: params.Priority,
				Proxied:  params.Proxied,
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
