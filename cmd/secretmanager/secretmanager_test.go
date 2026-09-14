package secretmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretManagerCommand_RegistersCreateToken(t *testing.T) {
	cmd := NewCommand()
	create, _, err := cmd.Find([]string{"create-token"})

	require.NoError(t, err)
	assert.Equal(t, "create-token", create.Name())
	for _, flag := range []string{"provider", "name", "provider-token", "base-url", "client-id", "namespace"} {
		assert.NotNil(t, create.Flags().Lookup(flag), "missing --%s", flag)
	}
}

func TestCreateTokenCommand_ValidatesProviderSpecificFlags(t *testing.T) {
	cmd := newCreateTokenCommand()
	cmd.SetArgs([]string{"--provider", "infisical", "--name", "Production", "--provider-token", "secret"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.ErrorContains(t, err, "--base-url and --client-id are required")
}
