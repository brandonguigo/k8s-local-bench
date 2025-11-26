package create

import (
	"context"

	"localplane/utils/dnsmasq"

	"github.com/spf13/cobra"
)

// updateDnsmasqConfig ensures dnsmasq maps the provided domain to the ip.
// It returns an error if the update or reload fails.
func updateDnsmasqConfig(cmd *cobra.Command, domain, ip string) error {
	client := dnsmasq.NewClient("")
	return client.EnsureDomainIP(context.Background(), domain, ip)
}
