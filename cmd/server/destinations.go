package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDestinationsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "destinations", Short: "Manage a server's Docker network destinations"}
	cmd.AddCommand(&cobra.Command{Use: "list <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "List server destinations", RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		destinations, err := service.NewDestinationService(client).ListByServer(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list server destinations: %w", err)
		}
		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}
		return formatter.Format(destinations)
	}})
	cmd.AddCommand(newServerDestinationCreateCommand())
	return cmd
}

func newServerDestinationCreateCommand() *cobra.Command {
	var name, network, destinationType string
	cmd := &cobra.Command{Use: "create <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Create a server destination", RunE: func(cmd *cobra.Command, args []string) error {
		if network == "" {
			return fmt.Errorf("--network is required")
		}
		if destinationType != "" && destinationType != "standalone" && destinationType != "swarm" {
			return fmt.Errorf("--type must be standalone or swarm")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		destination, err := service.NewDestinationService(client).CreateForServer(cmd.Context(), args[0], models.DestinationCreateRequest{Name: name, Network: network, Type: destinationType})
		if err != nil {
			return fmt.Errorf("failed to create destination: %w", err)
		}
		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}
		return formatter.Format(destination)
	}}
	cmd.Flags().StringVar(&name, "name", "", "Destination name")
	cmd.Flags().StringVar(&network, "network", "", "Docker network name (required)")
	cmd.Flags().StringVar(&destinationType, "type", "", "Destination type: standalone or swarm")
	return cmd
}
