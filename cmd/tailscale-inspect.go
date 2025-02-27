package cmd

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"insp", "look", "view"},
	Run: func(cmd *cobra.Command, args []string) {
		nameFilter := newFilter(args)
		cl := tsOpts.MustConnect()

		devices, err := cl.Devices(cmd.Context())
		cobra.CheckErr(err)

		for _, device := range devices {
			hostname := strings.Split(device.Hostname, ".")
			if nameFilter.Match(hostname[0]) && device.Authorized {
				cmd.Printf("%s: created %s; last seen %s\n", device.Name, humanize.Time(device.Created.Time), humanize.Time(device.LastSeen.Time))
			}
		}
	},
}

func init() {
	tailscaleCmd.AddCommand(inspectCmd)
}
