package deployment

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewBatchCommand deploys multiple resources by name
func NewBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch <name1,name2,...>",
		Short: "Deploy multiple resources by name",
		Long: `Deploy multiple resources at once.
Provide resource names as comma-separated values.
Example: coolify deploy batch app1,app2,app3`,
		Args: cli.ExactArgs(1, "<names>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			namesStr := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Parse comma-separated names
			names := make([]string, 0)
			for _, name := range strings.Split(namesStr, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					names = append(names, name)
				}
			}

			if len(names) == 0 {
				return fmt.Errorf("no resource names provided")
			}

			// Find resources by name
			resourceSvc := service.NewResourceService(client)
			resources, err := resourceSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			// Build map of name -> UUID
			nameToUUID := make(map[string]string)
			for _, r := range resources {
				nameToUUID[r.Name] = r.UUID
			}

			// Validate all names exist
			var notFound []string
			for _, name := range names {
				if _, exists := nameToUUID[name]; !exists {
					notFound = append(notFound, name)
				}
			}
			if len(notFound) > 0 {
				return fmt.Errorf("resources not found: %v", notFound)
			}

			// Deploy all resources
			force, _ := cmd.Flags().GetBool("force")
			deploySvc := service.NewDeploymentService(client)

			type result struct {
				Name    string
				UUID    string
				Success bool
				Message string
				Error   string
			}

			results := make([]result, 0, len(names))

			for _, name := range names {
				uuid := nameToUUID[name]
				fmt.Printf("Deploying %s...\n", name)

				res, err := deploySvc.Deploy(ctx, uuid, force)
				if err != nil {
					results = append(results, result{
						Name:    name,
						UUID:    uuid,
						Success: false,
						Error:   err.Error(),
					})
					fmt.Printf("  ❌ Failed: %v\n", err)
				} else {
					// Get first deployment message from the array
					message := ""
					if len(res.Deployments) > 0 {
						message = res.Deployments[0].Message
					}
					results = append(results, result{
						Name:    name,
						UUID:    uuid,
						Success: true,
						Message: message,
					})
					fmt.Printf("  ✅ Success: %s\n", message)
				}
			}

			// Summary
			successCount := 0
			for _, r := range results {
				if r.Success {
					successCount++
				}
			}

			fmt.Printf("\nBatch deployment complete: %d/%d succeeded\n", successCount, len(results))

			if successCount < len(results) {
				return fmt.Errorf("some deployments failed")
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force deployment")
	return cmd
}
