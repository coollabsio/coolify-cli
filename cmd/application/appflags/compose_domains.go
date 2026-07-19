package appflags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
)

// BindComposeDomainsFlag registers service-specific domains for Docker Compose applications.
func BindComposeDomainsFlag(cmd *cobra.Command) {
	cmd.Flags().StringArray("compose-domain", nil, "Docker Compose service domain in <service>=<url>[,<url>] format (repeatable)")
}

// ApplyComposeDomainsFlag copies explicitly supplied service domains to an API request.
func ApplyComposeDomainsFlag(cmd *cobra.Command, target *[]models.DockerComposeDomain) (bool, error) {
	if !cmd.Flags().Changed("compose-domain") {
		return false, nil
	}

	values, _ := cmd.Flags().GetStringArray("compose-domain")
	domains := make([]models.DockerComposeDomain, 0, len(values))
	for _, value := range values {
		name, domain, found := strings.Cut(value, "=")
		if !found {
			return false, fmt.Errorf("invalid --compose-domain %q: expected <service>=<url>", value)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return false, fmt.Errorf("invalid --compose-domain %q: service name cannot be empty", value)
		}
		domains = append(domains, models.DockerComposeDomain{
			Name:   name,
			Domain: strings.TrimSpace(domain),
		})
	}

	*target = domains
	return true, nil
}
