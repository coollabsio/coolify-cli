package cliprivatekeys

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func (c *cliPrivateKeys) newRemoveCommand() *cobra.Command {
	var forceRemove bool

	cmd := &cobra.Command{
		Use:     "remove [uuid]",
		Short:   "Remove a private key",
		Long:    `Remove an SSH private key from your Coolify instance.`,
		Aliases: []string{"delete", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			if !forceRemove {
				fmt.Printf("Are you sure you want to remove the private key with UUID '%s'? [y/N] ", uuid)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Operation canceled")
					return nil
				}
			}

			req, err := c.coolify().NewRequest(cmd.Context(), http.MethodDelete, fmt.Sprintf("security/keys/%s", uuid), nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			_, err = c.coolify().DoRequest(req)
			if err != nil {
				return fmt.Errorf("failed to remove private key: %w", err)
			}

			fmt.Println("Private key removed successfully")
			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.BoolVarP(&forceRemove, "force", "f", false, "Remove without confirmation prompt")

	return cmd
}
