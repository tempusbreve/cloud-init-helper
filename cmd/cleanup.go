package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean", "scrub"},
	Run: func(cmd *cobra.Command, args []string) {
		nameFilter := newFilter(args)
		now := time.Now().Add(-5 * time.Second)
		isDryRun := " (dry run)"
		if !cleanupDryRun {
			isDryRun = ""
		}

		cl := tsOpts.MustConnect()

		devices, err := cl.Devices(cmd.Context())
		cobra.CheckErr(err)

		for _, device := range devices {
			hostname := strings.Split(device.Hostname, ".")
			if nameFilter.Match(hostname[0]) {
				if device.LastSeen.Before(now) {
					cmd.Printf("Deleting %s%s\n", device.Name, isDryRun)
					if !cleanupDryRun {
						if err := cl.DeleteDevice(cmd.Context(), device.ID); err != nil {
							cmd.PrintErrf("error deleting %q: %v\n", device.Name, err)
						}
					}
				} else {
					cmd.Printf("Skipping %s%s\n", device.Name, isDryRun)
				}
			}
		}
	},
}

var cleanupDryRun = true

func init() {
	tailscaleCmd.AddCommand(cleanupCmd)
	tailscaleCmd.PersistentFlags().BoolVar(&cleanupDryRun, "dry-run", cleanupDryRun, "is this a dry run")
}
