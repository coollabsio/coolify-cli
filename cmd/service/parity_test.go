package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceCommandRegistersParityCommands(t *testing.T) {
	command := NewServiceCommand()
	for _, name := range []string{"logs", "move", "tag", "application", "database"} {
		child, _, err := command.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
	databaseCommand, _, err := command.Find([]string{"database"})
	require.NoError(t, err)
	for _, name := range []string{"list", "get", "update", "logs", "start", "restart", "stop"} {
		child, _, err := databaseCommand.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
	require.NotNil(t, NewCreateCommand().Flags().Lookup("tag"))
	require.NotNil(t, NewCreateCommand().Flags().Lookup("tags"))
	applicationCommand, _, err := command.Find([]string{"application"})
	require.NoError(t, err)
	for _, name := range []string{"list", "get", "update", "logs", "start", "restart", "stop"} {
		child, _, err := applicationCommand.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
	tagCommand, _, err := command.Find([]string{"tag"})
	require.NoError(t, err)
	for _, name := range []string{"list", "add", "remove"} {
		child, _, err := tagCommand.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
}
