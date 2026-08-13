package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand_DoesNotExposeV5MeshCommands(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, rootCmd.Commands(), "root command must have subcommands")

	for _, name := range []string{"init", "firewall"} {
		var found string
		for _, c := range rootCmd.Commands() {
			if c.Name() == name || c.HasAlias(name) {
				found = c.Name()
				break
			}
		}
		require.Empty(t, found, "%q must not be registered on the public CLI", name)
	}
}
