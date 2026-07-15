package cloudtoken

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
)

func TestCloudTokenCommand_RegistersFullSurface(t *testing.T) {
	cmd := NewCloudTokenCommand()
	for _, name := range []string{"list", "get", "create", "update", "delete", "validate"} {
		child, _, err := cmd.Find([]string{name})
		require.NoError(t, err)
		assert.Equal(t, name, child.Name())
	}
}

func TestOutputOptions_RespectsShowSensitiveFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("show-sensitive", false, "")
	require.NoError(t, cmd.Flags().Set("show-sensitive", "true"))

	assert.True(t, outputOptions(cmd).ShowSensitive)
}

func TestPrepareOutput_RedactsCloudTokensWithoutMutatingResponse(t *testing.T) {
	secret := "provider-secret"
	token := &models.CloudToken{UUID: "token-1", Token: &secret}

	redacted := prepareOutput(token, false).(*models.CloudToken)

	require.NotNil(t, redacted.Token)
	assert.Equal(t, output.SensitiveOverlay, *redacted.Token)
	assert.Equal(t, secret, *token.Token)
	assert.Same(t, token, prepareOutput(token, true))
}
