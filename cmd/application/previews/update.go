package previews

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/application/appflags"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewUpdatePreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <app_uuid> <pr_id>",
		Short: "Update the domains of a preview deployment",
		Long: `Replace the domains of a preview deployment.
Use --domains for regular applications and --compose-domain for Docker Compose applications.
Docker Compose domains are supplied as the complete service-to-domain mapping.
Redeploy the preview to apply the new domains.`,
		Example: `  coolify app previews update <app_uuid> 42 --domains https://pr-42.example.com
  coolify app previews update <app_uuid> 42 --compose-domain web=https://pr-42.example.com`,
		Args: cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			prID, err := parsePullRequestID(args[1])
			if err != nil {
				return err
			}

			req := models.ApplicationPreviewDomainsUpdateRequest{}
			if cmd.Flags().Changed("domains") {
				domains, _ := cmd.Flags().GetString("domains")
				req.Domains = &domains
			}
			composeChanged, err := appflags.ApplyComposeDomainsFlag(cmd, &req.DockerComposeDomains)
			if err != nil {
				return err
			}
			if req.Domains == nil && !composeChanged {
				return fmt.Errorf("set --domains or --compose-domain")
			}
			if cmd.Flags().Changed("force-domain-override") {
				force, _ := cmd.Flags().GetBool("force-domain-override")
				req.ForceDomainOverride = &force
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.3.15"); err != nil {
				return err
			}

			appSvc := service.NewApplicationService(client)
			resp, err := appSvc.UpdatePreviewDomains(ctx, args[0], prID, req)
			if err != nil {
				return fmt.Errorf("failed to update preview deployment: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(resp)
		},
	}

	cmd.Flags().String("domains", "", "Domains (comma-separated)")
	appflags.BindComposeDomainsFlag(cmd)
	cmd.Flags().Bool("force-domain-override", false, "Save the domains even when another resource uses them")
	cmd.MarkFlagsMutuallyExclusive("domains", "compose-domain")
	return cmd
}
