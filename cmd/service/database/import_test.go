package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceDatabaseCommandRegistersImportCommands(t *testing.T) {
	command := NewCommand()
	for _, name := range []string{"upload", "start", "status"} {
		child, _, err := command.Find([]string{"import", name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
	start, _, err := command.Find([]string{"import", "start"})
	require.NoError(t, err)
	require.Equal(t, "start <service_uuid> <database_uuid>", start.Use)
}
