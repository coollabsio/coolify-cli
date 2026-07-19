package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidateCommandInstallFlag(t *testing.T) {
	cmd := NewValidateCommand()
	flag := cmd.Flags().Lookup("install")

	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}
