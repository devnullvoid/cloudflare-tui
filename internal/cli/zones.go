package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage Cloudflare Zones",
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones accessible by the token",
	RunE: func(cmd *cobra.Command, args []string) error {
		api, err := getCloudflareClient()
		if err != nil {
			return err
		}

		zones, err := api.ListZones(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list zones: %w", err)
		}

		format := viper.GetString("format")
		return printOutput(zones, format)
	},
}

func init() {
	zonesCmd.AddCommand(zonesListCmd)
	rootCmd.AddCommand(zonesCmd)
}
