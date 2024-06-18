package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tailscale/tailscale-client-go/tailscale"
)

var tailscaleCmd = &cobra.Command{
	Use:     "tailscale",
	Aliases: []string{"ts"},
}

var tsOpts = tailscaleOpts{}

func init() {
	rootCmd.AddCommand(tailscaleCmd)

	const (
		apiKeyName  = "tailscale-api-key"
		authKeyName = "tailscale-auth-key"
		domainName  = "tailscale-domain"
	)

	tailscaleCmd.PersistentFlags().StringVar(&tsOpts.apiKey, apiKeyName, "", "API Key for Tailscale")
	tailscaleCmd.MarkPersistentFlagRequired(apiKeyName)

	tailscaleCmd.PersistentFlags().StringVar(&tsOpts.authKey, authKeyName, "", "Auth Key for Tailscale")
	tailscaleCmd.MarkPersistentFlagRequired(authKeyName)

	tailscaleCmd.PersistentFlags().StringVar(&tsOpts.domain, domainName, "", "Domain for Tailscale")
	tailscaleCmd.MarkPersistentFlagRequired(domainName)
}

type tailscaleOpts struct {
	apiKey  string
	authKey string
	domain  string
}

func (t tailscaleOpts) String() string {
	return fmt.Sprintf("Domain: %q / Auth: %q / API: %q", t.domain, t.authKey, t.apiKey)
}

func (t tailscaleOpts) MustConnect() *tailscale.Client {
	cl, err := tailscale.NewClient(t.apiKey, t.domain, tailscale.WithUserAgent("initialize-tool"))
	cobra.CheckErr(err)
	return cl
}

type filter map[string]bool

func newFilter(args []string) filter {
	var f filter = filter{}
	for _, arg := range args {
		f[arg] = true
	}
	return f
}

func (f filter) Match(name string) bool {
	if len(f) == 0 {
		return true
	}

	_, ok := f[name]
	return ok
}
