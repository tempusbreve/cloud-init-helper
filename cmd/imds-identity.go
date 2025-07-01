package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/imds"
)

var imdsIdentityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Get EC2 instance identity document",
	Long:  `Get the EC2 instance identity document in JSON format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := imds.NewClient()

		document, err := client.GetInstanceIdentityDocument(ctx)
		if err != nil {
			return fmt.Errorf("getting instance identity document: %w", err)
		}

		fmt.Print(document)
		return nil
	},
}

func init() {
	imdsCmd.AddCommand(imdsIdentityCmd)
}
