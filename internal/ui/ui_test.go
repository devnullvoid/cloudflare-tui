package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/cloudflare/cloudflare-go"
)

func TestUpdate(t *testing.T) {
	m := InitialModel(nil, &DefaultTheme, nil, nil)

	// Test case: FetchedZonesMsg should transition to ZoneListState
	zones := []cloudflare.Zone{
		{ID: "123", Name: "example.com"},
	}
	newModel, _ := m.Update(FetchedZonesMsg(zones))

	updatedModel := newModel.(*Model)
	if updatedModel.State != ZoneListState {
		t.Errorf("Expected state to be ZoneListState, got %v", updatedModel.State)
	}

	if len(updatedModel.ZoneList.Items()) != 1 {
		t.Errorf("Expected 1 zone item, got %d", len(updatedModel.ZoneList.Items()))
	}
}

func TestRecordListTransition(t *testing.T) {
	m := InitialModel(nil, &DefaultTheme, nil, nil)
	m.State = ZoneListState

	// Mock selecting a zone
	z := &ZoneItem{ID: "zone123", Name: "test.com"}
	m.ZoneList.SetItems([]list.Item{z})

	records := []cloudflare.DNSRecord{
		{ID: "rec1", Name: "www", Type: "A", Content: "1.1.1.1"},
	}

	newModel, _ := m.Update(FetchedRecordsMsg(records))
	updatedModel := newModel.(*Model)

	if updatedModel.State != RecordListState {
		t.Errorf("Expected RecordListState, got %v", updatedModel.State)
	}
	if len(updatedModel.RecordList.Items()) != 1 {
		t.Errorf("Expected 1 record item, got %d", len(updatedModel.RecordList.Items()))
	}
}

func TestNewRecordForm(t *testing.T) {
	// Test creating a new form from scratch
	form := NewRecordForm(nil, &DefaultTheme)
	if form.ID != "" {
		t.Errorf("Expected empty ID for new form, got %s", form.ID)
	}
	// Initializing with default A should have Name, Content, TTL, Comment
	if len(form.Inputs) != 4 {
		t.Errorf("Expected 4 inputs, got %d", len(form.Inputs))
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
	form = NewRecordForm(record, &DefaultTheme)
	if form.ID != "rec123" {
		t.Errorf("Expected ID rec123, got %s", form.ID)
	}
	if form.Type != "A" {
		t.Errorf("Expected Type A, got %s", form.Type)
	}
	if form.Inputs[0].Value() != "test" {
		t.Errorf("Expected Name test, got %s", form.Inputs[0].Value())
	}
	if !form.Proxied {
		t.Error("Expected proxied to be true")
	}
}

func TestZoneOperations(t *testing.T) {
	server, api := setupMockCloudflare(t)
	defer server.Close()

	m := InitialModel(api, &DefaultTheme, nil, nil)
	testZone := cloudflare.Zone{ID: "123", Name: "test.com"}

	// Test Zone Creation Message
	newModel, _ := m.Update(ZoneCreatedMsg(testZone))
	updatedModel := newModel.(*Model)
	if updatedModel.State != PendingZoneState {
		t.Errorf("Expected PendingZoneState after creation, got %v", updatedModel.State)
	}

	// Test Zone Deletion Message
	newModel, _ = m.Update(ZoneDeletedMsg{})
	updatedModel = newModel.(*Model)
	if updatedModel.State != LoadingZonesState {
		t.Errorf("Expected LoadingZonesState after deletion, got %v", updatedModel.State)
	}

	// Test Zone Check Message
	newModel, _ = m.Update(ZoneCheckTriggeredMsg(testZone))
	updatedModel = newModel.(*Model)
	if updatedModel.State != PendingZoneState {
		t.Errorf("Expected PendingZoneState after check, got %v", updatedModel.State)
	}
}
