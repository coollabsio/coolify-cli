package database

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/database/importcmd"
)

func newImportCommand() *cobra.Command {
	return importcmd.NewServiceDatabaseCommand()
}
