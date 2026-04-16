package initcmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestInitFlags_ParseSSHTimeout verifies the duration parsing helper.
func TestInitFlags_ParseSSHTimeout(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"30s", 30 * time.Second},
		{"1m", time.Minute},
		{"invalid", 30 * time.Second}, // default on parse failure
		{"0s", 30 * time.Second},      // default when <= 0
		{"", 30 * time.Second},        // default on empty
	}
	for _, tt := range tests {
		f := &InitFlags{SSHTimeout: tt.input}
		assert.Equal(t, tt.want, f.ParseSSHTimeout(), "input: %q", tt.input)
	}
}

// TestValidatePlanFlags checks required flag validation.
func TestValidatePlanFlags(t *testing.T) {
	t.Run("missing servers", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{SSHKey: "/path/to/key"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--servers")
	})

	t.Run("missing ssh key", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{Servers: []string{"1.1.1.1"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--ssh-key")
	})

	t.Run("valid", func(t *testing.T) {
		err := validatePlanFlags(&InitFlags{
			Servers: []string{"1.1.1.1"},
			SSHKey:  "/path/to/key",
		})
		assert.NoError(t, err)
	})
}

// fakeSSHRunner is a deterministic ssh.Runner for unit tests.
type fakeSSHRunner struct {
	// responses maps a command substring to the stdout to return.
	responses map[string]string
}

func (f *fakeSSHRunner) Run(_ context.Context, _, _, _ string, _ int, cmd string) (string, string, error) {
	for substr, resp := range f.responses {
		if strings.Contains(cmd, substr) {
			return resp, "", nil
		}
	}
	return "", "", nil
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
