package github

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCreateCommandDoesNotRequireAPIURL(t *testing.T) {
	cmd := NewCreateCommand()
	flag := cmd.Flags().Lookup("api-url")

	require.NotNil(t, flag)
	require.NotContains(t, flag.Annotations, cobra.BashCompOneRequiredFlag)
}
