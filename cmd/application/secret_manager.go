package application

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/secretmanager"
)

// NewSecretManagerCommand creates the app secret-manager command tree.
func NewSecretManagerCommand() *cobra.Command {
	return secretmanager.NewCommand(secretmanager.Resource{APIType: "applications", Command: "app", Arg: "<app_uuid>"})
}
