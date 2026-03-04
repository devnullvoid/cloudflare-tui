package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runTUI() {
	if err := executeTUI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func executeTUI() error {
	theme := getTheme()
	logPath := viper.GetString("log_path")
	debug := viper.GetBool("debug")
	logger, logFile := NewLogger(logPath, debug)

	api, err := getCloudflareClient(logger)
	if err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return err
	}

	m := ui.InitialModel(api, theme, logger, logFile)
	defer m.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
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
