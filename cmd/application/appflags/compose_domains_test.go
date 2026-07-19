package appflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestApplyComposeDomainsFlag_PreservesCommaSeparatedURLsAndEquals(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	BindComposeDomainsFlag(cmd)
	require.NoError(t, cmd.Flags().Set("compose-domain", "litellm=https://litellm.example.com"))
	require.NoError(t, cmd.Flags().Set("compose-domain", "admin=https://admin.example.com,https://admin2.example.com/login?next=/users&role=admin"))

	var domains []models.DockerComposeDomain
	changed, err := ApplyComposeDomainsFlag(cmd, &domains)

	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, []models.DockerComposeDomain{
		{Name: "litellm", Domain: "https://litellm.example.com"},
		{Name: "admin", Domain: "https://admin.example.com,https://admin2.example.com/login?next=/users&role=admin"},
	}, domains)
}

func TestApplyComposeDomainsFlag_AllowsEmptyDomain(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	BindComposeDomainsFlag(cmd)
	require.NoError(t, cmd.Flags().Set("compose-domain", "admin="))

	var domains []models.DockerComposeDomain
	changed, err := ApplyComposeDomainsFlag(cmd, &domains)

	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, []models.DockerComposeDomain{{Name: "admin", Domain: ""}}, domains)
}

func TestApplyComposeDomainsFlag_RejectsMalformedValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "missing separator", value: "litellm", want: "expected <service>=<url>"},
		{name: "empty service", value: " =https://example.com", want: "service name cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			BindComposeDomainsFlag(cmd)
			require.NoError(t, cmd.Flags().Set("compose-domain", tt.value))

			var domains []models.DockerComposeDomain
			changed, err := ApplyComposeDomainsFlag(cmd, &domains)

			assert.False(t, changed)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestApplyComposeDomainsFlag_OmitsUnchangedFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	BindComposeDomainsFlag(cmd)

	var domains []models.DockerComposeDomain
	changed, err := ApplyComposeDomainsFlag(cmd, &domains)

	require.NoError(t, err)
	assert.False(t, changed)
	assert.Nil(t, domains)
}
