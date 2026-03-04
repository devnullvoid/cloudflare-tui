package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cloudflare/cloudflare-go"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
	"github.com/spf13/cobra"
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

// resolveZoneID takes a string that could be either a Zone ID or a Zone Name
// and returns the actual Zone ID.
func resolveZoneID(api *cloudflare.API, identifier string) (string, error) {
	// Search for zone by name
	zones, err := api.ListZones(context.Background(), identifier)
	if err == nil && len(zones) > 0 {
		return zones[0].ID, nil
	}

	// Try as ID
	zone, err := api.ZoneDetails(context.Background(), identifier)
	if err == nil {
		return zone.ID, nil
	}

	return "", fmt.Errorf("could not find zone with name or ID: %s", identifier)
}

// CompleteZoneNames returns a list of zone names for shell completion.
func CompleteZoneNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	api, err := getCloudflareClient()
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
		
		t := table.New().
			Border(lipgloss.NormalBorder()).
			Headers(tableHeaders...).
			Rows(tableRows...)

		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Foreground(ui.DefaultTheme.Primary).
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
