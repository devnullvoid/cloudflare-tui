package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/cloudflare/cloudflare-go"
)

func TestUpdate(t *testing.T) {
	// Initialize a mock model
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m := model{
		state:    loadingZonesState,
		zoneList: l,
	}

	// Test case: fetchedZonesMsg should transition to zoneListState
	zones := []cloudflare.Zone{
		{ID: "123", Name: "example.com"},
	}
	newModel, _ := m.Update(fetchedZonesMsg(zones))
	
	updatedModel := newModel.(model)
	if updatedModel.state != zoneListState {
		t.Errorf("Expected state to be zoneListState, got %v", updatedModel.state)
	}

	if len(updatedModel.zoneList.Items()) != 1 {
		t.Errorf("Expected 1 zone item, got %d", len(updatedModel.zoneList.Items()))
	}
}

func TestNewRecordForm(t *testing.T) {
	// Test creating a new form from scratch
	form := newRecordForm(nil)
	if form.id != "" {
		t.Errorf("Expected empty ID for new form, got %s", form.id)
	}
	if len(form.inputs) != 3 {
		t.Errorf("Expected 3 inputs, got %d", len(form.inputs))
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
	form = newRecordForm(record)
	if form.id != "rec123" {
		t.Errorf("Expected ID rec123, got %s", form.id)
	}
	if form.inputs[0].Value() != "A" {
		t.Errorf("Expected Type A, got %s", form.inputs[0].Value())
	}
	if !form.proxied {
		t.Error("Expected proxied to be true")
	}
}
