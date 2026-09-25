package previews

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatePreviewCommand_RequiresDomains(t *testing.T) {
	cmd := NewUpdatePreviewCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"app-uuid", "42"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "one of --domains or --compose-domain is required")
}

func TestUpdatePreviewCommand_RejectsInvalidPullRequestID(t *testing.T) {
	cmd := NewUpdatePreviewCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"app-uuid", "1.9", "--domains", "https://preview.example.com"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pr_id")
}

func TestUpdatePreviewCommand_RejectsDomainsWithComposeDomain(t *testing.T) {
	cmd := NewUpdatePreviewCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{
		"app-uuid", "42",
		"--domains", "https://preview.example.com",
		"--compose-domain", "web=https://web.example.com",
	})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "group [domains compose-domain]")
}
