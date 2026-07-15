package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudTokenDecodesPrivilegedSensitiveToken(t *testing.T) {
	var token CloudToken
	require.NoError(t, json.Unmarshal([]byte(`{"uuid":"token-uuid","name":"prod","provider":"vultr","token":"secret"}`), &token))
	require.NotNil(t, token.Token)
	require.Equal(t, "secret", *token.Token)
}
