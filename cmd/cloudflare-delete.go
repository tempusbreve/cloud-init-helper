package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete DNS record",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErr("not implemented")
	},
}

func init() {
	// cloudflareCmd.AddCommand(deleteCmd)
}
