package cmd

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tempusbreve/cloud-init-helper/internal/chezmoi"
)

var chezmoiCmd = &cobra.Command{
	Use:     "chezmoi",
	Aliases: []string{"chz"},
	Short:   "Chezmoi Installer",
	GroupID: toolsGroup,
}

var chezmoiInstallCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install Chezmoi",
	Run: func(cmd *cobra.Command, args []string) {
		chOpts.AdditionalKeys = args
		cobra.CheckErr(chezmoi.Install(cmd.Context(), chOpts))
	},
}

var chezmoiApplyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"a"},
	Short:   "Apply Chezmoi",
	Run: func(cmd *cobra.Command, args []string) {
		chOpts.AdditionalKeys = args
		cobra.CheckErr(chezmoi.Apply(cmd.Context(), chOpts))
	},
}

var chOpts = chezmoi.Options{
	ChezmoiRepo: "https://github.com/jw4/min.files",
}

func init() {
	rootCmd.AddCommand(chezmoiCmd)
	chezmoiCmd.AddCommand(chezmoiInstallCmd)
	chezmoiCmd.AddCommand(chezmoiApplyCmd)

	if u, err := user.Current(); err == nil {
		chOpts.Display = u.Name
	}

	if home, err := os.UserHomeDir(); err == nil {
		chOpts.InstallDirectory = home
		chOpts.ChezmoiConfigDir = filepath.Join(home, ".config/chezmoi")
	}

	pf := chezmoiCmd.PersistentFlags()
	pf.StringVarP(&chOpts.Display, "display-name", "d", chOpts.Display, "Set Display Name")
	pf.StringVarP(&chOpts.Email, "email-address", "e", chOpts.Email, "Set Email Address")
	pf.StringVarP(&chOpts.GithubID, "github-id", "g", chOpts.GithubID, "Set Github ID")
	pf.StringVarP(&chOpts.ChezmoiRepo, "chezmoi-repo", "r", chOpts.ChezmoiRepo, "Set Chezmoi Repo URL")
	pf.StringVarP(&chOpts.ChezmoiConfigDir, "chezmoi-config-dir", "t", chOpts.ChezmoiConfigDir, "Set Chezmoi Config Dir")
	pf.StringVarP(&chOpts.InstallDirectory, "install-directory", "i", chOpts.InstallDirectory, "Set Dir where chezmoi will output rendered dotfiles")
	pf.StringVarP(&chOpts.SourceDirectory, "source-directory", "s", chOpts.SourceDirectory, "Set Dir where chezmoi will clone dotfiles source")
	pf.BoolVarP(&chOpts.DryRun, "dry-run", "n", chOpts.DryRun, "Dry Run (don't actually change files)")
}
