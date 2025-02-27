package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update DNS record",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErr("not implemented")
	},
}

func init() {
	// cloudflareCmd.AddCommand(updateCmd)
}
