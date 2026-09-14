package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditCommandRegistersListAndFilters(t *testing.T) {
	cmd := NewAuditCommand()
	list, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)
	assert.Equal(t, "list", list.Name())
	assert.NotNil(t, list.Flags().Lookup("source"))
	assert.NotNil(t, list.Flags().Lookup("action"))
	assert.NotNil(t, list.Flags().Lookup("search"))
	assert.NotNil(t, list.Flags().Lookup("per-page"))
	assert.NotNil(t, list.Flags().Lookup("page"))
}
