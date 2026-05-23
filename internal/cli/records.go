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
	Use:               "list [zone-name-or-id]",
	Short:             "List DNS records for a specific zone",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: CompleteZoneNames,
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

// Universal flags (all record types).
var (
	recordType         string
	recordName         string
	recordContent      string
	recordProxied      bool
	recordTTL          int
	recordComment      string
	recordFlattenCNAME bool
)

// MX and SRV both use priority.
var recordPriority int

// SRV-specific flags.
var (
	recordService string
	recordProto   string
	recordWeight  int
	recordPort    int
	recordTarget  string
)

// CAA-specific flags.
var (
	recordTag      string
	recordCAAFlags int
	recordValue    string
)

// recordFieldSet holds the type-specific fields derived from flags.
type recordFieldSet struct {
	content  string
	priority *uint16
	data     interface{}
	settings cloudflare.DNSRecordSettings
}

// validateAndBuildFields validates flag combinations and returns type-specific field values.
func validateAndBuildFields() (recordFieldSet, error) {
	var f recordFieldSet

	switch recordType {
	case "MX":
		if recordContent == "" {
			return f, fmt.Errorf("--content is required for MX records (mail server hostname)")
		}
		p := uint16(recordPriority)
		f.content = recordContent
		f.priority = &p

	case "SRV":
		if recordService == "" {
			return f, fmt.Errorf("--service is required for SRV records (e.g. _sip)")
		}
		if recordTarget == "" {
			return f, fmt.Errorf("--target is required for SRV records (e.g. sip.example.com)")
		}
		f.data = map[string]interface{}{
			"service":  recordService,
			"proto":    recordProto,
			"priority": recordPriority,
			"weight":   recordWeight,
			"port":     recordPort,
			"target":   recordTarget,
			"name":     recordName,
		}

	case "CAA":
		if recordTag == "" {
			return f, fmt.Errorf("--tag is required for CAA records (e.g. issue, issuewild, iodef)")
		}
		if recordValue == "" {
			return f, fmt.Errorf("--value is required for CAA records")
		}
		f.data = map[string]interface{}{
			"tag":   recordTag,
			"flags": recordCAAFlags,
			"value": recordValue,
		}

	case "CNAME":
		if recordContent == "" {
			return f, fmt.Errorf("--content is required for CNAME records")
		}
		f.content = recordContent
		f.settings = cloudflare.DNSRecordSettings{FlattenCNAME: &recordFlattenCNAME}

	default:
		if recordContent == "" {
			return f, fmt.Errorf("--content is required for %s records", recordType)
		}
		f.content = recordContent
	}

	return f, nil
}

var recordsCreateCmd = &cobra.Command{
	Use:               "create [zone-name-or-id]",
	Short:             "Create a new DNS record",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: CompleteZoneNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		fields, err := validateAndBuildFields()
		if err != nil {
			return err
		}

		rc := cloudflare.ZoneIdentifier(zoneID)
		params := cloudflare.CreateDNSRecordParams{
			Type:     recordType,
			Name:     recordName,
			Content:  fields.content,
			TTL:      recordTTL,
			Priority: fields.priority,
			Proxied:  &recordProxied,
			Data:     fields.data,
			Comment:  recordComment,
			Settings: fields.settings,
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
	Use:               "update [zone-name-or-id] [record-id]",
	Short:             "Update an existing DNS record",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: CompleteZoneThenRecordID,
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		fields, err := validateAndBuildFields()
		if err != nil {
			return err
		}

		comment := recordComment
		rc := cloudflare.ZoneIdentifier(zoneID)
		params := cloudflare.UpdateDNSRecordParams{
			ID:       args[1],
			Type:     recordType,
			Name:     recordName,
			Content:  fields.content,
			TTL:      recordTTL,
			Priority: fields.priority,
			Proxied:  &recordProxied,
			Data:     fields.data,
			Comment:  &comment,
			Settings: fields.settings,
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
	Use:               "delete [zone-name-or-id] [record-id]",
	Short:             "Delete a DNS record",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: CompleteZoneThenRecordID,
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneID, err := resolveZoneID(app.API, args[0])
		if err != nil {
			return err
		}

		rc := cloudflare.ZoneIdentifier(zoneID)
		err = app.API.DeleteDNSRecord(context.Background(), rc, args[1])
		if err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}

		fmt.Println("Record deleted successfully.")
		return nil
	},
}

func registerRecordFlags(cmd *cobra.Command) {
	// Universal
	cmd.Flags().StringVar(&recordType, "type", "A", "DNS record type (A, AAAA, CNAME, TXT, MX, SRV, CAA, …)")
	cmd.Flags().StringVarP(&recordName, "name", "n", "", "Record name (e.g. www, @, _sip._tcp)")
	cmd.Flags().StringVarP(&recordContent, "content", "c", "", "Record content — IP, hostname, or text value (not used for SRV/CAA)")
	cmd.Flags().BoolVarP(&recordProxied, "proxied", "p", false, "Proxy traffic through Cloudflare")
	cmd.Flags().IntVar(&recordTTL, "ttl", 1, "TTL in seconds (1 = auto)")
	cmd.Flags().StringVar(&recordComment, "comment", "", "Optional description for the record")
	cmd.Flags().BoolVar(&recordFlattenCNAME, "flatten-cname", false, "Flatten CNAME at zone apex (CNAME only)")

	// MX / SRV shared
	cmd.Flags().IntVar(&recordPriority, "priority", 10, "Priority (MX: mail priority; SRV: service priority)")

	// SRV-specific
	cmd.Flags().StringVar(&recordService, "service", "", "SRV service name (e.g. _sip)")
	cmd.Flags().StringVar(&recordProto, "proto", "_tcp", "SRV protocol (e.g. _tcp, _udp)")
	cmd.Flags().IntVar(&recordWeight, "weight", 1, "SRV weight")
	cmd.Flags().IntVar(&recordPort, "port", 0, "SRV port")
	cmd.Flags().StringVar(&recordTarget, "target", "", "SRV target hostname")

	// CAA-specific
	cmd.Flags().StringVar(&recordTag, "tag", "", "CAA tag (issue, issuewild, iodef)")
	cmd.Flags().IntVar(&recordCAAFlags, "caa-flags", 0, "CAA flags (0 or 128)")
	cmd.Flags().StringVar(&recordValue, "value", "", "CAA value (e.g. letsencrypt.org)")

	_ = cmd.MarkFlagRequired("name")

	// Tab completion for --type flag
	_ = cmd.RegisterFlagCompletionFunc("type", CompleteRecordTypes)
}

func init() {
	registerRecordFlags(recordsCreateCmd)
	registerRecordFlags(recordsUpdateCmd)

	recordsCmd.AddCommand(recordsListCmd)
	recordsCmd.AddCommand(recordsCreateCmd)
	recordsCmd.AddCommand(recordsUpdateCmd)
	recordsCmd.AddCommand(recordsDeleteCmd)

	rootCmd.AddCommand(recordsCmd)
}
