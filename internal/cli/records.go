package cli

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Manage DNS Records for a zone",
}

var recordsListCmd = &cobra.Command{
	Use:   "list [zone-id]",
	Short: "List DNS records for a specific zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		api, err := getCloudflareClient()
		if err != nil {
			return err
		}

		zoneID := args[0]
		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := api.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		format := viper.GetString("format")
		return printOutput(records, format)
	},
}

var (
	recordType    string
	recordName    string
	recordContent string
	recordProxied bool
)

var recordsCreateCmd = &cobra.Command{
	Use:   "create [zone-id]",
	Short: "Create a new DNS record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		api, err := getCloudflareClient()
		if err != nil {
			return err
		}

		zoneID := args[0]
		rc := cloudflare.ZoneIdentifier(zoneID)

		params := cloudflare.CreateDNSRecordParams{
			Type:    recordType,
			Name:    recordName,
			Content: recordContent,
			Proxied: &recordProxied,
		}

		rec, err := api.CreateDNSRecord(context.Background(), rc, params)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		format := viper.GetString("format")
		return printOutput(rec, format)
	},
}

var recordsUpdateCmd = &cobra.Command{
	Use:   "update [zone-id] [record-id]",
	Short: "Update an existing DNS record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		api, err := getCloudflareClient()
		if err != nil {
			return err
		}

		zoneID := args[0]
		recordID := args[1]
		rc := cloudflare.ZoneIdentifier(zoneID)

		params := cloudflare.UpdateDNSRecordParams{
			ID:      recordID,
			Type:    recordType,
			Name:    recordName,
			Content: recordContent,
			Proxied: &recordProxied,
		}

		rec, err := api.UpdateDNSRecord(context.Background(), rc, params)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}

		format := viper.GetString("format")
		return printOutput(rec, format)
	},
}

var recordsDeleteCmd = &cobra.Command{
	Use:   "delete [zone-id] [record-id]",
	Short: "Delete a DNS record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		api, err := getCloudflareClient()
		if err != nil {
			return err
		}

		zoneID := args[0]
		recordID := args[1]
		rc := cloudflare.ZoneIdentifier(zoneID)

		err = api.DeleteDNSRecord(context.Background(), rc, recordID)
		if err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}

		fmt.Println("Record deleted successfully.")
		return nil
	},
}

func init() {
	recordsCreateCmd.Flags().StringVarP(&recordType, "type", "t", "A", "DNS record type (A, CNAME, etc.)")
	recordsCreateCmd.Flags().StringVarP(&recordName, "name", "n", "", "DNS record name")
	recordsCreateCmd.Flags().StringVarP(&recordContent, "content", "c", "", "DNS record content")
	recordsCreateCmd.Flags().BoolVarP(&recordProxied, "proxied", "p", false, "Whether the record is proxied through Cloudflare")
	recordsCreateCmd.MarkFlagRequired("name")
	recordsCreateCmd.MarkFlagRequired("content")

	recordsUpdateCmd.Flags().StringVarP(&recordType, "type", "t", "A", "DNS record type (A, CNAME, etc.)")
	recordsUpdateCmd.Flags().StringVarP(&recordName, "name", "n", "", "DNS record name")
	recordsUpdateCmd.Flags().StringVarP(&recordContent, "content", "c", "", "DNS record content")
	recordsUpdateCmd.Flags().BoolVarP(&recordProxied, "proxied", "p", false, "Whether the record is proxied through Cloudflare")
	recordsUpdateCmd.MarkFlagRequired("name")
	recordsUpdateCmd.MarkFlagRequired("content")

	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsCreateCmd)
	recordsCmd.AddCommand(recordsUpdateCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)

	rootCmd.AddCommand(recordsCmd)
}
