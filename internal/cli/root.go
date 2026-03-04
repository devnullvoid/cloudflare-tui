package cli

import (
	"fmt"
	"os"

	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "cftui",
	Short: "A fast, terminal-based user interface for managing Cloudflare DNS records",
	Long: `cftui is a CLI and TUI tool for managing Cloudflare DNS records without logging into the dashboard.

If no command is provided, it will launch the interactive TUI.
You can also use the CLI commands to script and output data in structured formats.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().StringP("theme", "t", "ansi", "Color theme (ansi, mocha, nord, dracula)")
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
}

func initConfig() {
	viper.AutomaticEnv()
	// Map the CLOUDFLARE_API_TOKEN environment variable
	viper.BindEnv("api_token", "CLOUDFLARE_API_TOKEN")
}

// getTheme returns the selected theme based on CLI flags or viper config.
func getTheme() ui.Theme {
	name := viper.GetString("theme")
	if theme, ok := ui.AvailableThemes[name]; ok {
		return theme
	}
	return ui.DefaultTheme
}
