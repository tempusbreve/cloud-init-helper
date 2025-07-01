package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/imds"
)

var imdsGetCmd = &cobra.Command{
	Use:   "get [path]",
	Short: "Get metadata from AWS IMDS",
	Long: `Get metadata from AWS Instance Metadata Service using IMDSv2.

Examples:
  # Get instance ID
  cloud-init-helper imds get instance-id

  # Get instance type
  cloud-init-helper imds get instance-type

  # Get local IPv4 address
  cloud-init-helper imds get local-ipv4

  # Get custom metadata path
  cloud-init-helper imds get placement/availability-zone`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := imds.NewClient()

		if len(args) == 0 {
			paths, err := client.ListMetadataPaths(ctx, "")
			if err != nil {
				return fmt.Errorf("listing metadata paths: %w", err)
			}

			fmt.Println("Available metadata paths:")
			for _, path := range paths {
				fmt.Printf("  %s\n", path)
			}
			return nil
		}

		path := args[0]
		result, err := client.GetMetadata(ctx, path)
		if err != nil {
			return fmt.Errorf("getting metadata for path %q: %w", path, err)
		}

		fmt.Print(result)
		return nil
	},
}

func init() {
	imdsCmd.AddCommand(imdsGetCmd)
}
