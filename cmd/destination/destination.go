package destination

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDestinationCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "destination", Aliases: []string{"destinations"}, Short: "Manage Docker network destinations"}
	cmd.AddCommand(newListCommand(), newGetCommand(), newCreateCommand(), newDeleteCommand())
	return cmd
}

func print(cmd *cobra.Command, value any) error {
	format, _ := cmd.Flags().GetString("format")
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	formatter, err := output.NewFormatter(format, output.Options{ShowSensitive: showSensitive})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

func newListCommand() *cobra.Command {
	var serverUUID string
	cmd := &cobra.Command{Use: "list", Short: "List destinations", RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		svc := service.NewDestinationService(client)
		var destinations []models.Destination
		if serverUUID == "" {
			destinations, err = svc.List(cmd.Context())
		} else {
			destinations, err = svc.ListByServer(cmd.Context(), serverUUID)
		}
		if err != nil {
			return fmt.Errorf("failed to list destinations: %w", err)
		}
		return print(cmd, destinations)
	}}
	cmd.Flags().StringVar(&serverUUID, "server", "", "Only destinations belonging to this server UUID")
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{Use: "get <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Get a destination", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		destination, err := service.NewDestinationService(client).Get(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get destination: %w", err)
		}
		return print(cmd, destination)
	}}
}

func newCreateCommand() *cobra.Command {
	var serverUUID, name, network, destinationType string
	cmd := &cobra.Command{Use: "create", Short: "Create a destination for a server", RunE: func(cmd *cobra.Command, _ []string) error {
		if serverUUID == "" || network == "" {
			return fmt.Errorf("--server and --network are required")
		}
		if destinationType != "" && destinationType != "standalone" && destinationType != "swarm" {
			return fmt.Errorf("--type must be standalone or swarm")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		destination, err := service.NewDestinationService(client).CreateForServer(cmd.Context(), serverUUID, models.DestinationCreateRequest{Name: name, Network: network, Type: destinationType})
		if err != nil {
			return fmt.Errorf("failed to create destination: %w", err)
		}
		return print(cmd, destination)
	}}
	cmd.Flags().StringVar(&serverUUID, "server", "", "Server UUID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Destination name (defaults on the server)")
	cmd.Flags().StringVar(&network, "network", "", "Docker network name (required)")
	cmd.Flags().StringVar(&destinationType, "type", "", "Destination type: standalone or swarm")
	return cmd
}

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{Use: "delete <uuid>", Aliases: []string{"remove"}, Args: cli.ExactArgs(1, "<uuid>"), Short: "Delete an unused destination", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		if err := service.NewDestinationService(client).Delete(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete destination: %w", err)
		}
		fmt.Println("Destination deleted successfully.")
		return nil
	}}
}
