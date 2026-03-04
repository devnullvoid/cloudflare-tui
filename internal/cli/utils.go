package cli

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cloudflare/cloudflare-go"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// getCloudflareClient initializes the Cloudflare API client using the token from Viper.
func getCloudflareClient() (*cloudflare.API, error) {
	token := viper.GetString("api_token")
	if token == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN environment variable is required")
	}

	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudflare client: %w", err)
	}

	return api, nil
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
		
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ui.DefaultTheme.Primary)).
			Headers(tableHeaders...).
			Rows(tableRows...)

		// Apply styles to headers
		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 { // Header
				return lipgloss.NewStyle().
					Foreground(ui.DefaultTheme.Primary).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

		fmt.Println(t.Render())
	default:
		// Default to table if format is unsupported
		return printOutput(data, "table", tableHeaders, tableRows)
	}
	return nil
}
