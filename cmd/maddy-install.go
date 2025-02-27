package cmd

import (
	"github.com/spf13/cobra"
)

var maddyInstallCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install Maddy",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErr("not implemented")
	},
}

func init() {
	maddyCmd.AddCommand(maddyInstallCmd)
}
