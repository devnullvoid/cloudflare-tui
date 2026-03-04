package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage Cloudflare Zones",
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones accessible by the token",
	RunE: func(cmd *cobra.Command, args []string) error {
		zones, err := app.API.ListZones(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list zones: %w", err)
		}

		headers := []string{"ID", "Name", "Status", "Paused"}
		rows := make([][]string, len(zones))
		for i := range zones {
			rows[i] = []string{
				zones[i].ID,
				zones[i].Name,
				zones[i].Status,
				fmt.Sprintf("%v", zones[i].Paused),
			}
		}

		return printOutput(zones, app.Config.Format, headers, rows)
	},
}

func init() {
	zonesCmd.AddCommand(zonesListCmd)
	rootCmd.AddCommand(zonesCmd)
}
