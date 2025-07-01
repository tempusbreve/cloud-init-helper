package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/imds"
)

var imdsInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get common AWS instance information",
	Long:  `Get common AWS instance information from IMDS including instance ID, type, region, and IP addresses.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client := imds.NewClient()

		info := []struct {
			label string
			fn    func(context.Context) (string, error)
		}{
			{"Instance ID", client.GetInstanceID},
			{"Instance Type", client.GetInstanceType},
			{"Region", client.GetRegion},
			{"Availability Zone", client.GetAvailabilityZone},
			{"Local IPv4", client.GetLocalIPv4},
			{"Public IPv4", client.GetPublicIPv4},
		}

		for _, item := range info {
			value, err := item.fn(ctx)
			if err != nil {
				fmt.Printf("%-20s: ERROR - %v\n", item.label, err)
			} else {
				fmt.Printf("%-20s: %s\n", item.label, value)
			}
		}

		return nil
	},
}

func init() {
	imdsCmd.AddCommand(imdsInfoCmd)
}
