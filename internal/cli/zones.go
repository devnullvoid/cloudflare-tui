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
		logPath := viper.GetString("log_path")
		debug := viper.GetBool("debug")
		logger, logFile := NewLogger(logPath, debug)
		if logFile != nil {
			defer logFile.Close()
		}

		api, err := getCloudflareClient(logger)
		if err != nil {
			return err
		}

		zones, err := api.ListZones(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list zones: %w", err)
		}

		headers := []string{"ID", "Name", "Status", "Paused"}
		rows := make([][]string, len(zones))
		for i, z := range zones {
			rows[i] = []string{
				z.ID,
				z.Name,
				z.Status,
				fmt.Sprintf("%v", z.Paused),
			}
		}

		format := viper.GetString("format")
		return printOutput(zones, format, headers, rows)
	},
}

func init() {
	zonesCmd.AddCommand(zonesListCmd)
	rootCmd.AddCommand(zonesCmd)
}
