package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseCommandRegistersParityCommands(t *testing.T) {
	command := NewDatabaseCommand()
	for _, name := range []string{"logs", "move", "tag"} {
		child, _, err := command.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
	require.NotNil(t, NewCreateCommand().Flags().Lookup("tag"))
	require.NotNil(t, NewCreateCommand().Flags().Lookup("tags"))
	tagCommand, _, err := command.Find([]string{"tag"})
	require.NoError(t, err)
	for _, name := range []string{"list", "add", "remove"} {
		child, _, err := tagCommand.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
}
