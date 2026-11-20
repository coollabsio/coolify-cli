package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationSecretManagerCommand_RegistersSetCommand(t *testing.T) {
	cmd := NewSecretManagerCommand()
	set, _, err := cmd.Find([]string{"set"})

	require.NoError(t, err)
	assert.Equal(t, "set", set.Name())
	for _, flag := range []string{"integration-token-uuid", "project", "config", "project-id", "environment", "secret-path", "mount", "path"} {
		assert.NotNil(t, set.Flags().Lookup(flag), "missing --%s", flag)
	}
}

func TestApplicationCommand_RegistersSecretManager(t *testing.T) {
	cmd := NewAppCommand()
	child, _, err := cmd.Find([]string{"secret-manager"})

	require.NoError(t, err)
	assert.Equal(t, "secret-manager", child.Name())
}
