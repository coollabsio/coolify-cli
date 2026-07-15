package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update <service-uuid> <database-uuid>", Short: "Update a database in a service", Args: cli.ExactArgs(2, "<service-uuid> <database-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			request := &models.ServiceDatabaseUpdateRequest{}
			setString := func(name string, target **string) {
				if cmd.Flags().Changed(name) {
					value, _ := cmd.Flags().GetString(name)
					*target = &value
				}
			}
			setBool := func(name string, target **bool) {
				if cmd.Flags().Changed(name) {
					value, _ := cmd.Flags().GetBool(name)
					*target = &value
				}
			}
			setInt := func(name string, target **int) {
				if cmd.Flags().Changed(name) {
					value, _ := cmd.Flags().GetInt(name)
					*target = &value
				}
			}

			setString("human-name", &request.HumanName)
			setString("description", &request.Description)
			setString("image", &request.Image)
			setBool("exclude-from-status", &request.ExcludeFromStatus)
			setBool("log-drain-enabled", &request.IsLogDrainEnabled)
			setBool("public", &request.IsPublic)
			setInt("public-port", &request.PublicPort)
			setInt("public-port-timeout", &request.PublicPortTimeout)

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			database, err := service.NewService(client).UpdateDatabase(cmd.Context(), args[0], args[1], request)
			if err != nil {
				return err
			}
			return format(cmd, database)
		},
	}
	cmd.Flags().String("human-name", "", "Human-readable name")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("image", "", "Container image")
	cmd.Flags().Bool("exclude-from-status", false, "Exclude from service status")
	cmd.Flags().Bool("log-drain-enabled", false, "Enable log drain")
	cmd.Flags().Bool("public", false, "Expose the database publicly")
	cmd.Flags().Int("public-port", 0, "Public database port")
	cmd.Flags().Int("public-port-timeout", 0, "Public proxy timeout in seconds")
	return cmd
}
