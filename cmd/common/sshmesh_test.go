package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHMeshFlags_ParseSSHTimeout(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"30s", 30 * time.Second},
		{"1m", time.Minute},
		{"invalid", 30 * time.Second},
		{"0s", 30 * time.Second},
		{"", 30 * time.Second},
	}
	for _, tt := range tests {
		f := &SSHMeshFlags{SSHTimeout: tt.input}
		assert.Equal(t, tt.want, f.ParseSSHTimeout(), "input: %q", tt.input)
	}
}

func TestSSHMeshFlags_Validate(t *testing.T) {
	t.Run("missing servers", func(t *testing.T) {
		err := (&SSHMeshFlags{SSHKey: "/k"}).Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--servers")
	})
	t.Run("missing ssh key", func(t *testing.T) {
		err := (&SSHMeshFlags{Servers: []string{"1.1.1.1"}}).Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--ssh-key")
	})
	t.Run("valid", func(t *testing.T) {
		err := (&SSHMeshFlags{Servers: []string{"1.1.1.1"}, SSHKey: "/k"}).Validate()
		require.NoError(t, err)
	})
}

func TestSSHMeshFlags_ResolvePassphrase_Env(t *testing.T) {
	t.Setenv("COOLIFY_SSH_PASSPHRASE", "hunter2")
	pass, err := (&SSHMeshFlags{}).ResolvePassphrase()
	require.NoError(t, err)
	assert.Equal(t, []byte("hunter2"), pass)
}

func TestSSHMeshFlags_ResolvePassphrase_NoPrompt(t *testing.T) {
	t.Setenv("COOLIFY_SSH_PASSPHRASE", "")
	pass, err := (&SSHMeshFlags{SSHPassphrasePrompt: false}).ResolvePassphrase()
	require.NoError(t, err)
	assert.Nil(t, pass)
}
