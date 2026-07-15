package destination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDestinationCommand_RegistersCRUD(t *testing.T) {
	cmd := NewDestinationCommand()
	for _, name := range []string{"list", "get", "create", "delete"} {
		child, _, err := cmd.Find([]string{name})
		require.NoError(t, err)
		assert.Equal(t, name, child.Name())
	}
}
