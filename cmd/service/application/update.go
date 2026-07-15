package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update <service-uuid> <application-uuid>", Short: "Update an application in a service", Args: cli.ExactArgs(2, "<service-uuid> <application-uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			request := &models.ServiceApplicationUpdateRequest{}
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
			setString("url", &request.URL)
			setString("human-name", &request.HumanName)
			setString("description", &request.Description)
			setString("image", &request.Image)
			setBool("exclude-from-status", &request.ExcludeFromStatus)
			setBool("log-drain-enabled", &request.IsLogDrainEnabled)
			setBool("gzip-enabled", &request.IsGzipEnabled)
			setBool("stripprefix-enabled", &request.IsStripprefixEnabled)
			forceDomainOverride, _ := cmd.Flags().GetBool("force-domain-override")
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			application, err := service.NewService(client).UpdateApplication(cmd.Context(), args[0], args[1], forceDomainOverride, request)
			if err != nil {
				return err
			}
			return format(cmd, application)
		},
	}
	cmd.Flags().String("url", "", "Comma-separated public URLs")
	cmd.Flags().String("human-name", "", "Human-readable name")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("image", "", "Container image")
	cmd.Flags().Bool("exclude-from-status", false, "Exclude from service status")
	cmd.Flags().Bool("log-drain-enabled", false, "Enable log drain")
	cmd.Flags().Bool("gzip-enabled", false, "Enable gzip compression")
	cmd.Flags().Bool("stripprefix-enabled", false, "Enable path prefix stripping")
	cmd.Flags().Bool("force-domain-override", false, "Allow conflicting or duplicate domains")
	return cmd
}
