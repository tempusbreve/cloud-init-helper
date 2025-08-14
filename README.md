# cloud-init-helper

This tool provides a collection of helper commands for cloud-init scripts.

## Tools

### Cloudflare

The `cloudflare` command provides utilities for managing Cloudflare DNS records.

### IMDS

The `imds` command provides utilities for interacting with the AWS Instance Metadata Service (IMDSv2).

### Chezmoi

The `chezmoi` command provides utilities for installing and applying dotfiles with chezmoi.

### Tailscale

The `tailscale` command provides utilities for managing Tailscale.

## Environment Variables

All command flags can also be set using environment variables. The environment variables are prefixed with `CLOUD_INIT_HELPER_` and are all uppercase. For example, the `--tailscale-api-key` flag can be set with the `CLOUD_INIT_HELPER_TAILSCALE_API_KEY` environment variable.
