package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var WithResources bool

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Server related commands",
}

var listServersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Check API version
		version, err := client.GetVersion(ctx)
		if err != nil {
			return fmt.Errorf("failed to get API version: %w", err)
		}
		Version = version

		// Use service layer
		serverSvc := service.NewServerService(client)
		servers, err := serverSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		// Use output formatter
		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		if err := formatter.Format(servers); err != nil {
			return err
		}

		if !showSensitive && format == output.FormatTable {
			fmt.Println("\nNote: Use -s to show sensitive information.")
		}

		return nil
	},
}

var oneServerCmd = &cobra.Command{
	Use:   "get [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Get server details by uuid",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Use service layer
		serverSvc := service.NewServerService(client)
		uuid := args[0]

		// Get format flags
		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		var data interface{}
		if WithResources {
			resources, err := serverSvc.GetResources(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get server resources: %w", err)
			}
			data = resources.Resources
		} else {
			server, err := serverSvc.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get server: %w", err)
			}
			data = server
		}

		// Use output formatter
		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		if err := formatter.Format(data); err != nil {
			return err
		}

		if !showSensitive && format == output.FormatTable && !WithResources {
			fmt.Println("\nNote: Use -s to show sensitive information.")
		}

		return nil
	},
}

var removeServerCmd = &cobra.Command{
	Use:   "remove [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Remove a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Use service layer
		serverSvc := service.NewServerService(client)
		uuid := args[0]

		if err := serverSvc.Delete(ctx, uuid); err != nil {
			return fmt.Errorf("failed to delete server: %w", err)
		}

		fmt.Printf("Server %s deleted successfully\n", uuid)
		return nil
	},
}

var addServerCmd = &cobra.Command{
	Use:   "add [name] [ip] [private_key_uuid]",
	Args:  cobra.ExactArgs(3),
	Short: "Add a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Parse arguments and flags
		name := args[0]
		ip := args[1]
		privateKeyUuid := args[2]
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		validate, _ := cmd.Flags().GetBool("validate")

		// Create request
		req := models.ServerCreateRequest{
			Name:            name,
			IP:              ip,
			Port:            port,
			User:            user,
			PrivateKeyUUID:  privateKeyUuid,
			InstantValidate: validate,
		}

		// Use service layer
		serverSvc := service.NewServerService(client)
		response, err := serverSvc.Create(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		if validate {
			fmt.Printf("Server added successfully with uuid %s\n", response.UUID)
		} else {
			fmt.Printf("Server added successfully with uuid %s. Server is not validated. Use 'servers validate %s' to validate the server.\n", response.UUID, response.UUID)
		}

		return nil
	},
}

var validateServerCmd = &cobra.Command{
	Use:   "validate [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Validate a server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Use service layer
		serverSvc := service.NewServerService(client)
		uuid := args[0]

		response, err := serverSvc.Validate(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to validate server: %w", err)
		}

		if response.Message != "" {
			fmt.Println(response.Message)
		} else {
			fmt.Printf("Server %s validated successfully\n", uuid)
		}

		return nil
	},
}

func init() {
	// Note: format and show-sensitive flags are inherited from rootCmd.PersistentFlags()

	oneServerCmd.Flags().BoolVarP(&WithResources, "resources", "", false, "With resources")
	rootCmd.AddCommand(serversCmd)
	serversCmd.AddCommand(listServersCmd)
	serversCmd.AddCommand(oneServerCmd)

	addServerCmd.Flags().IntP("port", "p", 22, "Port")
	addServerCmd.Flags().StringP("user", "u", "root", "User")
	addServerCmd.Flags().BoolP("validate", "", false, "Validate the server")
	serversCmd.AddCommand(addServerCmd)
	serversCmd.AddCommand(validateServerCmd)
	serversCmd.AddCommand(removeServerCmd)
}
