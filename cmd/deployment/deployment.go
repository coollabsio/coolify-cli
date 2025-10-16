package deployment

import "github.com/spf13/cobra"

// NewDeploymentCommand creates the deployment parent command with all subcommands
func NewDeploymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy related commands",
	}

	// Add all deployment subcommands
	cmd.AddCommand(NewUUIDCommand())
	cmd.AddCommand(NewNameCommand())
	cmd.AddCommand(NewBatchCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCancelCommand())

	return cmd
}
