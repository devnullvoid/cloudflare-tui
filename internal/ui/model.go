package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

// Theme defines the color scheme for the application.
type Theme struct {
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Error     lipgloss.TerminalColor
	Warning   lipgloss.TerminalColor
	Inactive  lipgloss.TerminalColor
}

// DefaultTheme provides a standard "Charm" inspired palette.
var DefaultTheme = Theme{
	Primary:   lipgloss.Color("205"), // Pink
	Secondary: lipgloss.Color("86"),  // Aqua
	Error:     lipgloss.Color("9"),   // Red
	Warning:   lipgloss.Color("3"),   // Yellow
	Inactive:  lipgloss.Color("241"), // Gray
}

// Styles are derived from the theme.
var (
	DocStyle     = lipgloss.NewStyle().Margin(1, 2)
	NoStyle      = lipgloss.NewStyle()
)

// Model represents the application state.
type Model struct {
	State      SessionState
	CfClient   *cloudflare.API
	Theme      Theme
	
	ZoneList   list.Model
	RecordList list.Model
	Form       RecordForm
	Spinner    spinner.Model
	
	Err        error
	SelectedID string
	
	// Confirmation info
	OldRecord         *cloudflare.DNSRecord
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
