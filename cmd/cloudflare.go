package cmd

import (
	"github.com/spf13/cobra"
)

var cloudflareCmd = &cobra.Command{
	Use:     "cloudflare",
	Aliases: []string{"cf"},
	Short:   "Cloudflare helper commands",
	GroupID: toolsGroup,
}

var cfOpts = cloudflareOpts{}

func init() {
	rootCmd.AddCommand(cloudflareCmd)

	const (
		cfTokenKey   = "token"
		cfZoneName   = "zone-name"
		cfZoneID     = "zone-id"
		cfRecordType = "record-type"
		cfRecordName = "record-name"
	)

	flags := cloudflareCmd.PersistentFlags()

	flags.StringVarP(&cfOpts.token, cfTokenKey, "t", cfOpts.token, "Token for Cloudflare auth")
	_ = cloudflareCmd.MarkPersistentFlagRequired(cfTokenKey)

	flags.StringVarP(&cfOpts.zoneName, cfZoneName, "z", cfOpts.zoneName, "Cloudflare Zone")
	flags.StringVarP(&cfOpts.zoneID, cfZoneID, "i", cfOpts.zoneID, "Cloudflare Zone ID")
	flags.StringVarP(&cfOpts.recordType, cfRecordType, "y", cfOpts.recordType, "Record Type (MX, A, TXT, etc)")
	flags.StringVarP(&cfOpts.recordName, cfRecordName, "n", cfOpts.recordName, "Record Name")
}

type cloudflareOpts struct {
	token      string
	zoneName   string
	zoneID     string
	recordType string
	recordName string
}
