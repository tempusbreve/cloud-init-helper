package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/maddy"
)

var maddyInstallCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install Maddy",
	Run: func(cmd *cobra.Command, args []string) {
		if installParams.Repository+installParams.GitRef != "" {
			installParams.ForceCompile = true
		}
		cobra.CheckErr(maddy.Install(cmd.Context(), installParams))

		for _, domain := range additionalDomains {
			configParams.AdditionalDomains = append(configParams.AdditionalDomains, maddy.AdditionalDomain{MailDomain: domain})
		}
		cobra.CheckErr(maddy.Config(cmd.Context(), configParams))
		cobra.CheckErr(maddy.EnableAndStart())
	},
}

var (
	installParams     maddy.InstallParameters
	configParams      maddy.ConfigParameters
	additionalDomains []string
)

func init() {
	maddyCmd.AddCommand(maddyInstallCmd)

	flags := maddyInstallCmd.Flags()

	flags.StringVarP(&installParams.Repository, "maddy-repo", "r", installParams.Repository, "Repository for Maddy source code")
	flags.StringVarP(&installParams.GitRef, "maddy-ref", "b", installParams.GitRef, "GitRef for Maddy source code")
	flags.StringVarP(&installParams.Version, "maddy-version", "v", installParams.Version, "Version for Maddy source code")
	flags.BoolVarP(&installParams.ForceCompile, "force-compile", "f", installParams.ForceCompile, "Force compilation rather than download of binary artifacts")

	flags.StringVarP(&configParams.Hostname, "hostname", "n", configParams.Hostname, "Hostname running the maddy server, as shown by MX records")
	flags.StringVarP(&configParams.PrimaryMailDomain, "primary-domain", "p", configParams.PrimaryMailDomain, "Primary email domain")
	flags.StringSliceVarP(&additionalDomains, "additional-domain", "a", additionalDomains, "Additional email domain, can repeat")
	flags.StringVarP(&configParams.AcmeRegistrationEmail, "acme-registration-email", "e", configParams.AcmeRegistrationEmail, "Email used to register with ACME host")
}
