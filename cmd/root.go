package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use: "cloud-init-helper",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var toolsGroup = "tools"

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: toolsGroup, Title: "Installer Tools"})
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viperHack(rootCmd.Commands())
}

func viperHack(commands []*cobra.Command) {
	for _, cmd := range commands {
		_ = viper.BindPFlags(cmd.Flags())
		viperHackEnv(cmd)
		if cmd.HasSubCommands() {
			viperHack(cmd.Commands())
		}
	}
}

func viperHackEnv(cmd *cobra.Command) {
	_ = viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = viper.BindPFlag(f.Name, f)
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			_ = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
