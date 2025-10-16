package service

import "github.com/spf13/cobra"

// NewServiceCommand creates the service parent command with all subcommands
func NewServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Service related commands",
		Long:    `Manage Coolify one-click services (databases, Redis, PostgreSQL, etc.).`,
	}

	// Add main service commands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewRestartCommand())
	cmd.AddCommand(NewDeleteCommand())

	// Add env subcommand (placeholder for now)
	// TODO: Implement env commands
	// envCmd := &cobra.Command{
	// 	Use:     "env",
	// 	Short:   "Manage service environment variables",
	// }
	// envCmd.AddCommand(env.NewListCommand())
	// ... more env commands
	// cmd.AddCommand(envCmd)

	return cmd
}
