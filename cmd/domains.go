package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:     "domain",
	Aliases: []string{"domains"},
	Short:   "Domain related commands",
	Long:    `List all domains configured across your Coolify resources.`,
}

var listDomainsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		domainSvc := service.NewDomainService(client)
		domains, err := domainSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(domains)
	},
}

func init() {
	rootCmd.AddCommand(domainsCmd)
	domainsCmd.AddCommand(listDomainsCmd)
}
