package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
)

func runTUI() {
	theme := getTheme()

	// Note: The log file is opened in PersistentPreRunE and stored in app.Logger.
	// InitialModel needs the log file handle to close it properly.
	// We'll refactor NewLogger to return the handle in the app struct for cleaner access.

	// For now, we'll re-obtain it safely.
	_, logFile := NewLogger(app.Config.LogPath, app.Config.Debug)

	m := ui.InitialModel(app.API, theme, app.Logger, logFile)
	defer m.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		app.Logger.Fatal("TUI execution failed", "error", err)
	}
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive Terminal User Interface",
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
