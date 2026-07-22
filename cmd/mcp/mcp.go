package mcp

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewMCPCommand creates the mcp parent command
func NewMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage the Coolify MCP server",
		Long:  `Enable or disable the Coolify MCP server. Requires a root team API token.`,
	}

	cmd.AddCommand(NewEnableCommand())
	cmd.AddCommand(NewDisableCommand())

	return cmd
}

// NewEnableCommand creates the mcp enable command
func NewEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable the MCP server",
		Long:  `Enable the Coolify MCP server (API: POST /mcp/enable). Requires a root team API token.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			mcpSvc := service.NewMCPService(client)
			resp, err := mcpSvc.Enable(ctx)
			if err != nil {
				return fmt.Errorf("failed to enable MCP server: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

// NewDisableCommand creates the mcp disable command
func NewDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable the MCP server",
		Long:  `Disable the Coolify MCP server (API: POST /mcp/disable). Requires a root team API token.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			mcpSvc := service.NewMCPService(client)
			resp, err := mcpSvc.Disable(ctx)
			if err != nil {
				return fmt.Errorf("failed to disable MCP server: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}
