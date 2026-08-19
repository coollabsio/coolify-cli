package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsCommand_RegistersInstanceEmailCommands(t *testing.T) {
	cmd := NewSettingsCommand()
	email, _, err := cmd.Find([]string{"email"})
	require.NoError(t, err)
	assert.Equal(t, "email", email.Name())

	for _, name := range []string{"get", "update"} {
		child, _, err := email.Find([]string{name})
		require.NoError(t, err)
		assert.Equal(t, name, child.Name())
	}

	update, _, err := email.Find([]string{"update"})
	require.NoError(t, err)
	assert.NotNil(t, update.Flags().Lookup("json"))
	assert.Contains(t, update.Long, "smtp_ehlo_domain")
}
