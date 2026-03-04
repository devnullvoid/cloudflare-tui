package cli

import (
	"fmt"
	"os"

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
		// Default behavior is to run the TUI if no subcommands are provided
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
	rootCmd.PersistentFlags().StringP("format", "f", "json", "Output format (json, yaml)")
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
}

func initConfig() {
	viper.AutomaticEnv()
	// Map the CLOUDFLARE_API_TOKEN environment variable
	viper.BindEnv("api_token", "CLOUDFLARE_API_TOKEN")
}
