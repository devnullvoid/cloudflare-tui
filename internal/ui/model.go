package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudflare/cloudflare-go"
)

// SessionState tracks which view is currently active.
type SessionState int

const (
	LoadingZonesState SessionState = iota
	ZoneListState
	LoadingRecordsState
	RecordListState
	EditingRecordState
	ConfirmingDeleteState
	ConfirmingSaveState
)

// Styles for the UI.
var (
	DocStyle      = lipgloss.NewStyle().Margin(1, 2)
	ErrStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	FocusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	NoStyle       = lipgloss.NewStyle()
	HelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	ConfirmStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)

// Model represents the application state.
type Model struct {
	State      SessionState
	CfClient   *cloudflare.API
	ZoneList   list.Model
	RecordList list.Model
	Form       RecordForm
	Err        error
	SelectedID string
	
	// Pending deletion info
	PendingDeleteID   string
	PendingDeleteName string
}

// RecordForm manages input fields for adding/editing a record.
type RecordForm struct {
	ID      string
	Inputs  []textinput.Model
	Focused int
	Proxied bool
}

// Message types for async operations.
type FetchedZonesMsg []cloudflare.Zone
type FetchedRecordsMsg []cloudflare.DNSRecord
type RecordSavedMsg struct{}
type RecordDeletedMsg struct{}
type ErrorMsg error
