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

	cfMaddyCmd.PersistentFlags().StringVar(&cfMailOpts.domain, cfMailDomain, "", "Mail Domain")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfMailDomain)

	cfMaddyCmd.PersistentFlags().StringVar(&cfMailOpts.postmaster, cfPostmaster, "", "Mail Domain Postmaster email address")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfPostmaster)

	cfMaddyCmd.PersistentFlags().StringVar(&cfMailOpts.dkim, cfDKIM, "", "DKIM TXT record value")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfDKIM)

	cfMaddyCmd.PersistentFlags().StringSliceVar(&cfMailOpts.mxHosts, cfMXHost, []string{}, "DKIM TXT record value")
	_ = cfMaddyCmd.MarkPersistentFlagRequired(cfMXHost)
}

var cfMailOpts cfMailOptions

type cfMailOptions struct {
	domain     string
	postmaster string
	dkim       string
	mxHosts    []string
}
