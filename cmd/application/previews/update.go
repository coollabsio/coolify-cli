package previews

import (
	"fmt"
	"strconv"

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
		Short: "Update preview deployment domains",
		Long: `Replace the domains for an application preview deployment.

Use --domains for regular applications. Use repeatable --compose-domain flags for Docker Compose applications.
Include an internal port in a URL when needed; Coolify stores the public domain without the port.

Examples:
  coolify app previews update <app_uuid> 42 --domains "https://preview.example.com:3000"
  coolify app previews update <app_uuid> 42 --compose-domain "web=https://web.example.com:8080" --compose-domain "api=https://api.example.com:3000"`,
		Args: cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			prID := args[1]
			parsedPRID, err := strconv.Atoi(prID)
			if err != nil || parsedPRID <= 0 {
				return fmt.Errorf("invalid pr_id: must be a positive integer")
			}

			req := models.ApplicationPreviewUpdateRequest{}
			if cmd.Flags().Changed("domains") {
				domains, _ := cmd.Flags().GetString("domains")
				req.Domains = &domains
			}
			composeDomainsChanged, err := appflags.ApplyComposeDomainsFlag(cmd, &req.DockerComposeDomains)
			if err != nil {
				return err
			}
			if req.Domains == nil && !composeDomainsChanged {
				return fmt.Errorf("one of --domains or --compose-domain is required")
			}
			req.ForceDomainOverride, _ = cmd.Flags().GetBool("force-domain-override")

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			if err := cli.CheckMinimumVersion(cmd.Context(), client, "4.0.0-beta.474"); err != nil {
				return err
			}

			preview, err := service.NewApplicationService(client).UpdatePreview(cmd.Context(), args[0], prID, req)
			if err != nil {
				return fmt.Errorf("failed to update preview deployment: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(preview)
		},
	}

	cmd.Flags().String("domains", "", "Domains for a regular application preview (comma-separated)")
	appflags.BindComposeDomainsFlag(cmd)
	cmd.Flags().Bool("force-domain-override", false, "Allow domains already used by another resource")
	cmd.MarkFlagsMutuallyExclusive("domains", "compose-domain")

	return cmd
}
