package cmd

import (
	"github.com/spf13/cobra"
)

var imdsCmd = &cobra.Command{
	Use:     "imds",
	Short:   "AWS Instance Metadata Service (IMDSv2) helper commands",
	GroupID: toolsGroup,
}

func init() {
	rootCmd.AddCommand(imdsCmd)
}
