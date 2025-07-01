package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tempusbreve/cloud-init-helper/internal/imds"
)

var imdsUserdataCmd = &cobra.Command{
	Use:   "userdata",
	Short: "Get EC2 instance user data",
	Long:  `Get the user data that was passed to the EC2 instance at launch time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := imds.NewClient()

		userdata, err := client.GetUserData(ctx)
		if err != nil {
			return fmt.Errorf("getting user data: %w", err)
		}

		fmt.Print(userdata)
		return nil
	},
}

func init() {
	imdsCmd.AddCommand(imdsUserdataCmd)
}
