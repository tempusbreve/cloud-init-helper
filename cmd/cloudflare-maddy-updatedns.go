package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/dns"
)

var updateDNSCmd = &cobra.Command{
	Use:     "update-dns",
	Aliases: []string{"update"},
	Short:   "Update DNS Records for Maddy",
	Run: func(cmd *cobra.Command, args []string) {
		var api dns.API = dns.NewCloudFlareDNS(
			dns.WithCFToken(cfOpts.token),
			dns.WithCFZoneName(cfOpts.zoneName),
		)

		mc := dns.NewMailConfig(dns.WithAPI(api))

		options := dns.UpdateMailRecordsParams{
			Domain:      cfMailOpts.domain,
			Postmaster:  cfMailOpts.postmaster,
			DKIM:        cfMailOpts.dkim,
			MXHosts:     map[string]int{},
			Destructive: destructive,
		}

		for _, h := range cfMailOpts.mxHosts {
			options.MXHosts[h] = 10
		}

		cobra.CheckErr(mc.UpdateAllMailRecords(cmd.Context(), options))
	},
}

var destructive bool

func init() {
	cfMaddyCmd.AddCommand(updateDNSCmd)

	updateDNSCmd.Flags().BoolVar(&destructive, "destructive", false, "Cause conflicting DNS records to be deleted")
}
