package cli

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
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

var zonesCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		zone, err := app.API.CreateZone(context.Background(), name, false, cloudflare.Account{}, "full")
		if err != nil {
			return fmt.Errorf("failed to create zone: %w", err)
		}

		fmt.Printf("Zone created successfully!\n\n")
		fmt.Printf("ID:          %s\n", zone.ID)
		fmt.Printf("Name:        %s\n", zone.Name)
		fmt.Printf("Status:      %s\n", zone.Status)
		if len(zone.NameServers) > 0 {
			fmt.Printf("Nameservers: %v\n", zone.NameServers)
		}
		return nil
	},
}

var zonesDeleteCmd = &cobra.Command{
	Use:   "delete [zone-name-or-id]",
	Short: "Delete a zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		_, err = app.API.DeleteZone(context.Background(), zoneID)
		if err != nil {
			return fmt.Errorf("failed to delete zone: %w", err)
		}

		fmt.Println("Zone deleted successfully.")
		return nil
	},
}

var zonesCheckCmd = &cobra.Command{
	Use:   "check [zone-name-or-id]",
	Short: "Trigger an activation check for a pending zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		_, err = app.API.ZoneActivationCheck(context.Background(), zoneID)
		if err != nil {
			return fmt.Errorf("failed to trigger check: %w", err)
		}

		fmt.Println("Activation check triggered successfully.")
		return nil
	},
}

func init() {
	zonesCmd.AddCommand(zonesListCmd)
	zonesCmd.AddCommand(zonesCreateCmd)
	zonesCmd.AddCommand(zonesDeleteCmd)
	zonesCmd.AddCommand(zonesCheckCmd)
	rootCmd.AddCommand(zonesCmd)
}
