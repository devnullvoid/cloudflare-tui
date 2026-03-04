package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// loggingRoundTripper logs HTTP requests and responses.
type loggingRoundTripper struct {
	next   http.RoundTripper
	logger *log.Logger
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log Request
	dump, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		l.logger.Debug("Cloudflare API Request", "raw", string(dump))
	}

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Log Response
	dumpResp, err := httputil.DumpResponse(resp, true)
	if err == nil {
		l.logger.Debug("Cloudflare API Response", "raw", string(dumpResp))
	}

	return resp, nil
}

// NewLogger creates a themed logger for file output.
func NewLogger(logPath string, debug bool) (*log.Logger, *os.File) {
	if logPath == "" {
		return log.New(os.Stderr), nil
	}

	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	f, _ := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	
	logger := log.New(f)
	logger.SetReportTimestamp(true)
	logger.SetTimeFormat("[2006-01-02 15:04:05]")
	
	styles := log.DefaultStyles()
	levelStyle := func(level string) lipgloss.Style {
		return lipgloss.NewStyle().SetString("[" + level + "]")
	}
	styles.Levels[log.DebugLevel] = levelStyle("DEBUG")
	styles.Levels[log.InfoLevel] = levelStyle("INFO")
	styles.Levels[log.WarnLevel] = levelStyle("WARN")
	styles.Levels[log.ErrorLevel] = levelStyle("ERROR")
	logger.SetStyles(styles)

	if debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	return logger, f
}

// getCloudflareClient initializes the Cloudflare API client.
func getCloudflareClient(logger *log.Logger) (*cloudflare.API, error) {
	token := viper.GetString("api_token")
	if token == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN environment variable is required")
	}

	opts := []cloudflare.Option{}
	
	// If debugging is active, wrap the HTTP client to log raw requests/responses
	if logger != nil && logger.GetLevel() == log.DebugLevel {
		httpClient := &http.Client{
			Transport: &loggingRoundTripper{
				next:   http.DefaultTransport,
				logger: logger,
			},
		}
		opts = append(opts, cloudflare.HTTPClient(httpClient))
	}

	api, err := cloudflare.NewWithAPIToken(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudflare client: %w", err)
	}

	return api, nil
}

// resolveZoneID takes a string that could be either a Zone ID or a Zone Name
func resolveZoneID(api *cloudflare.API, identifier string) (string, error) {
	zones, err := api.ListZones(context.Background(), identifier)
	if err == nil && len(zones) > 0 {
		return zones[0].ID, nil
	}

	zone, err := api.ZoneDetails(context.Background(), identifier)
	if err == nil {
		return zone.ID, nil
	}

	return "", fmt.Errorf("could not find zone with name or ID: %s", identifier)
}

// CompleteZoneNames returns a list of zone names for shell completion.
func CompleteZoneNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// We don't log during completion to keep it fast and clean
	api, err := getCloudflareClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	zones, err := api.ListZones(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, z := range zones {
		names = append(names, z.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// printOutput formats and prints structured data based on the requested format.
func printOutput(data interface{}, format string, tableHeaders []string, tableRows [][]string) error {
	switch format {
	case "json":
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
	case "table":
		if len(tableHeaders) == 0 || len(tableRows) == 0 {
			fmt.Println("No data to display.")
			return nil
		}
		
		theme := getTheme()
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(theme.Primary)).
			Headers(tableHeaders...).
			Rows(tableRows...)

		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Foreground(theme.Primary).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

		fmt.Println(t.Render())
	default:
		return printOutput(data, "table", tableHeaders, tableRows)
	}
	return nil
}
