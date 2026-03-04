package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
)

func runTUI() {
	api, err := getCloudflareClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	theme := getTheme()
	m := ui.InitialModel(api, theme)
	defer m.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
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
