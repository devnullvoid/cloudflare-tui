package cli

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Manage DNS Records for a zone",
}

func getRecordRow(r *cloudflare.DNSRecord) []string {
	proxied := "No"
	if r.Proxied != nil && *r.Proxied {
		proxied = "Yes"
	}
	return []string{
		r.ID,
		r.Type,
		r.Name,
		r.Content,
		proxied,
		fmt.Sprintf("%v", r.TTL),
	}
}

var recordHeaders = []string{"ID", "Type", "Name", "Content", "Proxied", "TTL"}

var recordsListCmd = &cobra.Command{
	Use:   "list [zone-name-or-id]",
	Short: "List DNS records for a specific zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		rc := cloudflare.ZoneIdentifier(zoneID)
		records, _, err := app.API.ListDNSRecords(context.Background(), rc, cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		rows := make([][]string, len(records))
		for i := range records {
			rows[i] = getRecordRow(&records[i])
		}

		return printOutput(records, app.Config.Format, recordHeaders, rows)
	},
}

var (
	recordType    string
	recordName    string
	recordContent string
	recordProxied bool
)

var recordsCreateCmd = &cobra.Command{
	Use:   "create [zone-name-or-id]",
	Short: "Create a new DNS record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		rc := cloudflare.ZoneIdentifier(zoneID)

		params := cloudflare.CreateDNSRecordParams{
			Type:    recordType,
			Name:    recordName,
			Content: recordContent,
			Proxied: &recordProxied,
		}

		rec, err := app.API.CreateDNSRecord(context.Background(), rc, params)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		rows := [][]string{getRecordRow(&rec)}
		return printOutput(rec, app.Config.Format, recordHeaders, rows)
	},
}

var recordsUpdateCmd = &cobra.Command{
	Use:   "update [zone-name-or-id] [record-id]",
	Short: "Update an existing DNS record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		recordID := args[1]
		rc := cloudflare.ZoneIdentifier(zoneID)

		params := cloudflare.UpdateDNSRecordParams{
			ID:      recordID,
			Type:    recordType,
			Name:    recordName,
			Content: recordContent,
			Proxied: &recordProxied,
		}

		rec, err := app.API.UpdateDNSRecord(context.Background(), rc, params)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}

		rows := [][]string{getRecordRow(&rec)}
		return printOutput(rec, app.Config.Format, recordHeaders, rows)
	},
}

var recordsDeleteCmd = &cobra.Command{
	Use:   "delete [zone-name-or-id] [record-id]",
	Short: "Delete a DNS record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		recordID := args[1]
		rc := cloudflare.ZoneIdentifier(zoneID)

		err = app.API.DeleteDNSRecord(context.Background(), rc, recordID)
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
	_ = recordsCreateCmd.MarkFlagRequired("name")
	_ = recordsCreateCmd.MarkFlagRequired("content")

	recordsUpdateCmd.Flags().StringVarP(&recordType, "type", "t", "A", "DNS record type (A, CNAME, etc.)")
	recordsUpdateCmd.Flags().StringVarP(&recordName, "name", "n", "", "DNS record name")
	recordsUpdateCmd.Flags().StringVarP(&recordContent, "content", "c", "", "DNS record content")
	recordsUpdateCmd.Flags().BoolVarP(&recordProxied, "proxied", "p", false, "Whether the record is proxied through Cloudflare")
	_ = recordsUpdateCmd.MarkFlagRequired("name")
	_ = recordsUpdateCmd.MarkFlagRequired("content")

	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsCreateCmd)
	recordsCmd.AddCommand(recordsUpdateCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)

	// Register dynamic zone name completions
	recordsListCmd.ValidArgsFunction = CompleteZoneNames
	recordsCreateCmd.ValidArgsFunction = CompleteZoneNames
	recordsUpdateCmd.ValidArgsFunction = CompleteZoneNames
	recordsDeleteCmd.ValidArgsFunction = CompleteZoneNames

	rootCmd.AddCommand(recordsCmd)
}
