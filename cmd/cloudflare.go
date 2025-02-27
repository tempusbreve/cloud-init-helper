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

	cloudflareCmd.PersistentFlags().StringVar(&cfOpts.token, cfTokenKey, "", "Token for Cloudflare auth")
	_ = cloudflareCmd.MarkPersistentFlagRequired(cfTokenKey)

	cloudflareCmd.PersistentFlags().StringVar(&cfOpts.zoneName, cfZoneName, "", "Cloudflare Zone")
	cloudflareCmd.PersistentFlags().StringVar(&cfOpts.zoneID, cfZoneID, "", "Cloudflare Zone ID")
	cloudflareCmd.PersistentFlags().StringVar(&cfOpts.recordType, cfRecordType, "", "Record Type (MX, A, TXT, etc)")
	cloudflareCmd.PersistentFlags().StringVar(&cfOpts.recordName, cfRecordName, "", "Record Name")
}

type cloudflareOpts struct {
	token      string
	zoneName   string
	zoneID     string
	recordType string
	recordName string
}
