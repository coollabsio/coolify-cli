package initcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/cmd/common"
)

// TestValidatePlanFlags checks required flag validation.
func TestValidatePlanFlags(t *testing.T) {
	t.Run("missing servers", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{
			SSHMeshFlags: common.SSHMeshFlags{SSHKey: "/path/to/key"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--servers")
	})

	t.Run("missing ssh key", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{
			SSHMeshFlags: common.SSHMeshFlags{Servers: []string{"1.1.1.1"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--ssh-key")
	})

	t.Run("valid", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{
			SSHMeshFlags: common.SSHMeshFlags{
				Servers: []string{"1.1.1.1"},
				SSHKey:  "/path/to/key",
			},
			MeshNetFlags: common.MeshNetFlags{
				Namespaces: []string{common.DefaultNamespace},
			},
		})
		require.NoError(t, err)
	})

	t.Run("invalid namespace", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{
			SSHMeshFlags: common.SSHMeshFlags{
				Servers: []string{"1.1.1.1"},
				SSHKey:  "/path/to/key",
			},
			MeshNetFlags: common.MeshNetFlags{
				Namespaces: []string{"Not Valid"},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid namespace")
	})
}

// TestShouldSkipGate verifies the alpha gate bypass logic.
func TestShouldSkipGate(t *testing.T) {
	// --yes flag
	assert.True(t, shouldSkipGate(&InitFlags{Yes: true}))

	// Without --yes and without env var, behaviour depends on TTY.
	// We can't reliably test the TTY path in unit tests, but we can
	// confirm the env-var bypass.
	t.Setenv("COOLIFY_NON_INTERACTIVE", "1")
	assert.True(t, shouldSkipGate(&InitFlags{}))
}
