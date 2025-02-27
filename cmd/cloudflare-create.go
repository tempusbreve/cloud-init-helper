package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create DNS record",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErr("not implemented")
	},
}

func init() {
	// cloudflareCmd.AddCommand(createCmd)
}
