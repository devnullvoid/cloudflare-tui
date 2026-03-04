package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Determine default log path
	defaultLogPath := ""
	stateDir, err := os.UserConfigDir()
	if err == nil {
		defaultLogPath = filepath.Join(stateDir, "cftui", "cftui.log")
	}

	// Global flags
	rootCmd.PersistentFlags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().StringP("theme", "t", "ansi", "Color theme (ansi, mocha, nord, dracula, rose-pine, tokyo-night, gruvbox, everforest)")
	rootCmd.PersistentFlags().String("log", defaultLogPath, "Path to log file")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	viper.BindPFlag("log_path", rootCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CFTUI")
	
	// Map environment variables (Prefix: CFTUI_)
	viper.BindEnv("api_token", "CLOUDFLARE_API_TOKEN") // Keep this one as is for convention
	viper.BindEnv("theme", "CFTUI_THEME")
	viper.BindEnv("log_path", "CFTUI_LOG")
	viper.BindEnv("debug", "CFTUI_DEBUG")
}

func getTheme() ui.Theme {
	name := viper.GetString("theme")
	if theme, ok := ui.AvailableThemes[name]; ok {
		return theme
	}
	return ui.DefaultTheme
}
