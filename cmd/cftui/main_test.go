package main

import (
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
)

func TestUpdate(t *testing.T) {
	m := ui.InitialModel(nil)

	// Test case: FetchedZonesMsg should transition to ZoneListState
	zones := []cloudflare.Zone{
		{ID: "123", Name: "example.com"},
	}
	newModel, _ := m.Update(ui.FetchedZonesMsg(zones))
	
	updatedModel := newModel.(ui.Model)
	if updatedModel.State != ui.ZoneListState {
		t.Errorf("Expected state to be ZoneListState, got %v", updatedModel.State)
	}

	if len(updatedModel.ZoneList.Items()) != 1 {
		t.Errorf("Expected 1 zone item, got %d", len(updatedModel.ZoneList.Items()))
	}
}

func TestNewRecordForm(t *testing.T) {
	// Test creating a new form from scratch
	form := ui.NewRecordForm(nil, ui.DefaultTheme)
	if form.ID != "" {
		t.Errorf("Expected empty ID for new form, got %s", form.ID)
	}
	if len(form.Inputs) != 3 {
		t.Errorf("Expected 3 inputs, got %d", len(form.Inputs))
	}

	// Test creating a form from an existing record
	proxied := true
	record := &cloudflare.DNSRecord{
		ID:      "rec123",
		Type:    "A",
		Name:    "test",
		Content: "1.2.3.4",
		Proxied: &proxied,
	}
	form = ui.NewRecordForm(record, ui.DefaultTheme)
	if form.ID != "rec123" {
		t.Errorf("Expected ID rec123, got %s", form.ID)
	}
	if form.Inputs[0].Value() != "A" {
		t.Errorf("Expected Type A, got %s", form.Inputs[0].Value())
	}
	if !form.Proxied {
		t.Error("Expected proxied to be true")
	}
}
