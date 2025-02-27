package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tempusbreve/cloud-init-helper/internal/dns"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read DNS record",
	Run: func(cmd *cobra.Command, args []string) {
		var api dns.API = dns.NewCloudFlareDNS(
			dns.WithCFToken(cfOpts.token),
			dns.WithCFZoneName(cfOpts.zoneName),
		)

		records, err := api.GetRecords(cmd.Context(), cfOpts.recordName, cfOpts.recordType)
		cobra.CheckErr(err)

		for _, rec := range records {
			cmd.Printf("%s: %s %s :: %s\n", rec.ID(), rec.Type(), rec.Name(), rec.Content())
		}
	},
}

var readOpts = cloudflareReadOpts{}

type cloudflareReadOpts struct {
	content string
}

func init() {
	cloudflareCmd.AddCommand(readCmd)

	const (
		readContent = "content"
	)

	readCmd.PersistentFlags().StringVar(&readOpts.content, readContent, "", "Match Records with this Content")
}
