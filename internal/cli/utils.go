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

	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return log.New(os.Stderr), nil
	}

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
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	api, err := completionAPI()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	zones, err := api.ListZones(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := make([]string, len(zones))
	for i := range zones {
		names[i] = zones[i].Name + "\t" + zones[i].Status
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// CompleteRecordTypes returns valid DNS record type values for --type flag completion.
func CompleteRecordTypes(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"A\tIPv4 address",
		"AAAA\tIPv6 address",
		"CAA\tCertification Authority Authorization",
		"CERT\tCertificate",
		"CNAME\tCanonical name",
		"DNSKEY\tDNS Key",
		"DS\tDelegation Signer",
		"MX\tMail exchange",
		"NAPTR\tName Authority Pointer",
		"NS\tName server",
		"PTR\tPointer",
		"SMIMEA\tS/MIME Certificate Association",
		"SPF\tSender Policy Framework",
		"SRV\tService locator",
		"SSHFP\tSSH Fingerprint",
		"SVCB\tService Binding",
		"TLSA\tTLS Authentication",
		"TXT\tText",
		"URI\tUniform Resource Identifier",
	}, cobra.ShellCompDirectiveNoFileComp
}

// CompleteRecordIDs returns DNS record IDs for a zone already provided as args[0].
func CompleteRecordIDs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	api, err := completionAPI()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	zoneID, err := resolveZoneID(api, args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ids := make([]string, len(records))
	for i := range records {
		ids[i] = records[i].ID + "\t" + records[i].Type + " " + records[i].Name
	}

	return ids, cobra.ShellCompDirectiveNoFileComp
}

// CompleteZoneThenRecordID completes zone name for arg[0] and record ID for args[1].
func CompleteZoneThenRecordID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return CompleteZoneNames(cmd, args, toComplete)
	case 1:
		return CompleteRecordIDs(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completionAPI returns the API client for use in completion functions,
// preferring the already-initialized app client when available.
func completionAPI() (*cloudflare.API, error) {
	if app.API != nil {
		return app.API, nil
	}
	return getCloudflareClient(nil)
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
