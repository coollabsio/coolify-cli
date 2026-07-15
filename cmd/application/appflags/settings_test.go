package appflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyTagsFlag_CombinesRepeatableAndCommaSeparatedTags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	BindTagsFlag(cmd)
	require.NoError(t, cmd.Flags().Set("tag", "production"))
	require.NoError(t, cmd.Flags().Set("tag", "customer-facing"))
	require.NoError(t, cmd.Flags().Set("tags", "api,web"))

	var tags []string
	assert.True(t, ApplyTagsFlag(cmd, &tags))
	assert.Equal(t, []string{"production", "customer-facing", "api", "web"}, tags)
}
