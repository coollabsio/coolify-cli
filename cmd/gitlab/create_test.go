package gitlab

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCreateCommandRequiresNameAndHTMLURL(t *testing.T) {
	cmd := NewCreateCommand()

	nameFlag := cmd.Flags().Lookup("name")
	htmlFlag := cmd.Flags().Lookup("html-url")
	apiFlag := cmd.Flags().Lookup("api-url")

	require.NotNil(t, nameFlag)
	require.NotNil(t, htmlFlag)
	require.NotNil(t, apiFlag)

	require.Contains(t, nameFlag.Annotations, cobra.BashCompOneRequiredFlag)
	require.Contains(t, htmlFlag.Annotations, cobra.BashCompOneRequiredFlag)
	require.NotContains(t, apiFlag.Annotations, cobra.BashCompOneRequiredFlag)
}

func TestGitLabRootCommandWiresSubcommands(t *testing.T) {
	cmd := NewGitLabCommand()
	require.Equal(t, "gitlab", cmd.Use)

	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}

	for _, expected := range []string{"list", "get", "create", "update", "delete"} {
		require.True(t, names[expected], "missing subcommand %s", expected)
	}
}
