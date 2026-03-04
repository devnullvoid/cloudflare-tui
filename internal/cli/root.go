package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/cloudflare/cloudflare-go"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CLI stores the shared state for CLI commands.
type CLI struct {
	API    *cloudflare.API
	Logger *log.Logger
	Config *Config
}

// Config stores application configuration.
type Config struct {
	APIToken string
	Theme    string
	Format   string
	LogPath  string
	Debug    bool
}

var (
	app     = &CLI{}
	rootCmd = &cobra.Command{
		Use:   "cftui",
		Short: "A fast, terminal-based user interface for managing Cloudflare DNS records",
		Long: `cftui is a CLI and TUI tool for managing Cloudflare DNS records without logging into the dashboard.

If no command is provided, it will launch the interactive TUI.
You can also use the CLI commands to script and output data in structured formats.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip for completion and help commands
			if cmd.Name() == "completion" || cmd.Name() == "help" {
				return nil
			}

			app.Config = &Config{
				APIToken: viper.GetString("api_token"),
				Theme:    viper.GetString("theme"),
				Format:   viper.GetString("format"),
				LogPath:  viper.GetString("log_path"),
				Debug:    viper.GetBool("debug"),
			}

			var logFile *os.File
			app.Logger, logFile = NewLogger(app.Config.LogPath, app.Config.Debug)
			// Note: We don't close logFile here because we want it open for the duration of the command.
			// For short-lived CLI commands, the OS will close it. For the TUI, m.Close() handles it.
			// To be strictly correct, we could store it in app and use PostRun.
			_ = logFile

			api, err := getCloudflareClient(app.Logger)
			if err != nil {
				return err
			}
			app.API = api

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runTUI()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defaultLogPath := ""
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			stateDir = filepath.Join(home, ".local", "state")
		}
	}
	if stateDir != "" {
		defaultLogPath = filepath.Join(stateDir, "cftui", "cftui.log")
	}

	rootCmd.PersistentFlags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().StringP("theme", "t", "ansi", "Color theme (ansi, mocha, nord, dracula, rose-pine, tokyo-night, gruvbox, everforest)")
	rootCmd.PersistentFlags().String("log", defaultLogPath, "Path to log file")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	_ = viper.BindPFlag("log_path", rootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CFTUI")
	_ = viper.BindEnv("api_token", "CLOUDFLARE_API_TOKEN")
	_ = viper.BindEnv("theme", "CFTUI_THEME")
	_ = viper.BindEnv("log_path", "CFTUI_LOG")
	_ = viper.BindEnv("debug", "CFTUI_DEBUG")
}

func getTheme() *ui.Theme {
	name := viper.GetString("theme")
	if theme, ok := ui.AvailableThemes[name]; ok {
		return &theme
	}
	return &ui.DefaultTheme
}
