package cmd

import (
	"github.com/spf13/cobra"
)

var maddyCmd = &cobra.Command{
	Use:     "maddy",
	Aliases: []string{"m"},
	Short:   "Maddy related commands",
	GroupID: toolsGroup,
}

func init() {
	rootCmd.AddCommand(maddyCmd)
}
