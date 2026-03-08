package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

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
	Mock     bool
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
			if cmd.Name() == "completion" || cmd.Name() == "help" {
				return nil
			}

			app.Config = &Config{
				APIToken: viper.GetString("api_token"),
				Theme:    viper.GetString("theme"),
				Format:   viper.GetString("format"),
				LogPath:  viper.GetString("log_path"),
				Debug:    viper.GetBool("debug"),
				Mock:     viper.GetBool("mock"),
			}

			var logFile *os.File
			app.Logger, logFile = NewLogger(app.Config.LogPath, app.Config.Debug)
			_ = logFile

			var api *cloudflare.API
			var err error

			if app.Config.Mock {
				app.Logger.Warn("MOCK MODE ENABLED: Using local mock server")
				server := setupLocalMockServer()
				api, err = cloudflare.New("deadbeef", "test@example.com", cloudflare.BaseURL(server.URL))
			} else {
				api, err = getCloudflareClient(app.Logger)
			}

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

func setupLocalMockServer() *httptest.Server {
	mux := http.NewServeMux()
	
	// Mock Zones
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			resp := map[string]interface{}{
				"success": true,
				"result": []cloudflare.Zone{
					{ID: "123", Name: "mock-zone.com", Status: "active", NameServers: []string{"ns1.cloudflare.com", "ns2.cloudflare.com"}},
					{ID: "456", Name: "pending-mock.io", Status: "pending", NameServers: []string{"alice.ns.cloudflare.com", "bob.ns.cloudflare.com"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.Method == http.MethodPost {
			resp := map[string]interface{}{
				"success": true,
				"result":  cloudflare.Zone{ID: "789", Name: "new-mock.com", Status: "pending", NameServers: []string{"carl.ns.cloudflare.com", "dave.ns.cloudflare.com"}},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	// Mock Single Zone and DNS Records
	mux.HandleFunc("/zones/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Handle dns_records request: /zones/:id/dns_records
		if strings.HasSuffix(r.URL.Path, "/dns_records") {
			resp := map[string]interface{}{
				"success": true,
				"result": []cloudflare.DNSRecord{
					{ID: "r1", Name: "www", Type: "A", Content: "1.2.3.4", TTL: 1, Proxied: cloudflare.BoolPtr(true)},
					{ID: "r2", Name: "api", Type: "CNAME", Content: "target.mock.com", TTL: 1, Proxied: cloudflare.BoolPtr(false)},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle generic zone details
		resp := map[string]interface{}{
			"success": true, 
			"result": map[string]interface{}{
				"id": "123", 
				"name": "mock-zone.com",
				"status": "active",
				"name_servers": []string{"ns1.cloudflare.com", "ns2.cloudflare.com"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	
	return httptest.NewServer(mux)
}

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
	rootCmd.PersistentFlags().Bool("mock", false, "Use a local mock API for testing")

	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	_ = viper.BindPFlag("log_path", rootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("mock", rootCmd.PersistentFlags().Lookup("mock"))
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CFTUI")
	_ = viper.BindEnv("api_token", "CLOUDFLARE_API_TOKEN")
	_ = viper.BindEnv("theme", "CFTUI_THEME")
	_ = viper.BindEnv("log_path", "CFTUI_LOG")
	_ = viper.BindEnv("debug", "CFTUI_DEBUG")
	_ = viper.BindEnv("mock", "CFTUI_MOCK")
}

func getTheme() *ui.Theme {
	name := viper.GetString("theme")
	if theme, ok := ui.AvailableThemes[name]; ok {
		return &theme
	}
	return &ui.DefaultTheme
}
