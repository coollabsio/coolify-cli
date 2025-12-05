package env

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewListEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <app_uuid>",
		Short: "List all environment variables for an application",
		Long:  `List all environment variables for a specific application. By default, only non-preview environment variables are shown. Use --preview to show preview environment variables instead, or --all to show all variables (non-preview first, then preview).`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			envs, err := appSvc.ListEnvs(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to list environment variables: %w", err)
			}

			// Filter by preview/all flags
			showAll, _ := cmd.Flags().GetBool("all")
			showPreview, _ := cmd.Flags().GetBool("preview")

			if showAll {
				// Sort: non-preview first, then preview
				sort.SliceStable(envs, func(i, j int) bool {
					if envs[i].IsPreview != envs[j].IsPreview {
						return !envs[i].IsPreview // non-preview (false) comes before preview (true)
					}
					return false // maintain original order within groups
				})
			} else {
				// Filter by preview flag
				var filtered []models.EnvironmentVariable
				for _, env := range envs {
					if env.IsPreview == showPreview {
						filtered = append(filtered, env)
					}
				}
				envs = filtered
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			if !showSensitive {
				for i := range envs {
					envs[i].Value = "********"
					if envs[i].RealValue != nil {
						masked := "********"
						envs[i].RealValue = &masked
					}
				}
			}

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(envs)
		},
	}

	cmd.Flags().Bool("preview", false, "Show preview environment variables instead of regular ones")
	cmd.Flags().Bool("all", false, "Show all environment variables (non-preview first, then preview)")

	return cmd
}
