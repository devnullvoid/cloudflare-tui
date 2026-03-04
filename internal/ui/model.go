package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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
	ConfirmingQuitState
)

// Theme defines the color scheme for the application.
type Theme struct {
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Error     lipgloss.TerminalColor
	Warning   lipgloss.TerminalColor
	Inactive  lipgloss.TerminalColor
}

// AvailableThemes is a registry of all supported color schemes.
var AvailableThemes = map[string]Theme{
	"ansi": {
		Primary:   lipgloss.Color("5"), // Magenta
		Secondary: lipgloss.Color("6"), // Cyan
		Error:     lipgloss.Color("1"), // Red
		Warning:   lipgloss.Color("3"), // Yellow
		Inactive:  lipgloss.Color("8"), // Gray
	},
	"mocha": {
		Primary:   lipgloss.Color("#fab387"), // Peach
		Secondary: lipgloss.Color("#89dceb"), // Sky
		Error:     lipgloss.Color("#f38ba8"), // Red
		Warning:   lipgloss.Color("#f9e2af"), // Yellow
		Inactive:  lipgloss.Color("#585b70"), // Surface 2
	},
	"nord": {
		Primary:   lipgloss.Color("#88C0D0"), // Frost (Blue)
		Secondary: lipgloss.Color("#81A1C1"), // Frost (Lighter Blue)
		Error:     lipgloss.Color("#BF616A"), // Aurora (Red)
		Warning:   lipgloss.Color("#EBCB8B"), // Aurora (Yellow)
		Inactive:  lipgloss.Color("#4C566A"), // Polar Night (Gray)
	},
	"dracula": {
		Primary:   lipgloss.Color("#bd93f9"), // Purple
		Secondary: lipgloss.Color("#8be9fd"), // Cyan
		Error:     lipgloss.Color("#ff5555"), // Red
		Warning:   lipgloss.Color("#f1fa8c"), // Yellow
		Inactive:  lipgloss.Color("#6272a4"), // Comment
	},
	"rose-pine": {
		Primary:   lipgloss.Color("#ebbcba"), // Rose
		Secondary: lipgloss.Color("#31748f"), // Pine
		Error:     lipgloss.Color("#eb6f92"), // Love
		Warning:   lipgloss.Color("#f6c177"), // Gold
		Inactive:  lipgloss.Color("#908caa"), // Subtle
	},
	"tokyo-night": {
		Primary:   lipgloss.Color("#7aa2f7"), // Blue
		Secondary: lipgloss.Color("#bb9af7"), // Purple
		Error:     lipgloss.Color("#f7768e"), // Red
		Warning:   lipgloss.Color("#e0af68"), // Orange
		Inactive:  lipgloss.Color("#565f89"), // Gray
	},
	"gruvbox": {
		Primary:   lipgloss.Color("#fabd2f"), // Yellow
		Secondary: lipgloss.Color("#83a598"), // Blue
		Error:     lipgloss.Color("#fb4934"), // Red
		Warning:   lipgloss.Color("#fe8019"), // Orange
		Inactive:  lipgloss.Color("#928374"), // Gray
	},
	"everforest": {
		Primary:   lipgloss.Color("#a7c080"), // Green
		Secondary: lipgloss.Color("#7fbbb3"), // Blue
		Error:     lipgloss.Color("#e67e80"), // Red
		Warning:   lipgloss.Color("#dbbc7f"), // Yellow
		Inactive:  lipgloss.Color("#859289"), // Gray
	},
}

// DefaultTheme provides the fallback ANSI palette.
var DefaultTheme = AvailableThemes["ansi"]

// Styles are derived from the theme.
var (
	DocStyle = lipgloss.NewStyle().Margin(1, 2)
	NoStyle  = lipgloss.NewStyle()
)

// Model represents the application state.
type Model struct {
	State    SessionState
	CfClient *cloudflare.API
	Theme    Theme

	ZoneList   list.Model
	RecordList list.Model
	Form       RecordForm
	Spinner    spinner.Model
	Logger     *log.Logger
	LogFile    *os.File

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
