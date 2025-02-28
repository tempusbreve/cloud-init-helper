package cmd

import (
	"github.com/spf13/cobra"
)

var cfMaddyCmd = &cobra.Command{
	Use:   "maddy",
	Short: "Cloudflare DNS related commands for Maddy configuration",
}

func init() {
	cloudflareCmd.AddCommand(cfMaddyCmd)

	const (
		cfMailDomain = "mail-domain"
		cfPostmaster = "postmaster"
		cfDKIM       = "dkim"
		cfMXHost     = "mx-host"
	)

	cfMaddyCmd.PersistentFlags().StringVarP(&cfMailOpts.domain, cfMailDomain, "m", cfMailOpts.domain, "Mail Domain")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfMailDomain)

	cfMaddyCmd.PersistentFlags().StringVarP(&cfMailOpts.postmaster, cfPostmaster, "p", cfMailOpts.postmaster, "Mail Domain Postmaster email address")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfPostmaster)

	cfMaddyCmd.PersistentFlags().StringVarP(&cfMailOpts.dkim, cfDKIM, "k", cfMailOpts.dkim, "DKIM TXT record value")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfDKIM)

	cfMaddyCmd.PersistentFlags().StringSliceVarP(&cfMailOpts.mxHosts, cfMXHost, "x", cfMailOpts.mxHosts, "DKIM TXT record value")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfMXHost)
}

var cfMailOpts cfMailOptions

type cfMailOptions struct {
	domain     string
	postmaster string
	dkim       string
	mxHosts    []string
}
