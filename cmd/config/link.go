package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func newLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage the link between an environment and coolify.json",
	}
	cmd.AddCommand(newLinkCreateCommand())
	cmd.AddCommand(newLinkGetCommand())
	cmd.AddCommand(newLinkUpdateCommand())
	cmd.AddCommand(newLinkDeleteCommand())
	return cmd
}

func newLinkCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Link an environment to coolify.json",
		Long: `Link an environment to coolify.json. Creates a missing project or environment.
Only a team admin can link.

In api mode, apply files with 'coolify config apply <link_uuid> -f coolify.json'.
In git mode, Coolify reads the file from the repository on each push. The
webhook URLs and secret are shown one time (use --show-sensitive to see the secret).

Examples:
  coolify config link create --project my-app --environment production --mode api
  coolify config link create --project my-app --environment production --mode git \
    --repository https://github.com/acme/api --branch main --source-type public --server-uuid <server_uuid>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f := cmd.Flags()
			req := models.ConfigLinkCreateRequest{}
			req.Project, _ = f.GetString("project")
			req.Environment, _ = f.GetString("environment")
			req.Mode, _ = f.GetString("mode")
			req.Repository, _ = f.GetString("repository")
			req.Branch, _ = f.GetString("branch")
			req.FilePath, _ = f.GetString("file-path")
			req.SourceType, _ = f.GetString("source-type")
			req.SourceUUID, _ = f.GetString("source-uuid")
			req.ServerUUID, _ = f.GetString("server-uuid")
			req.DestinationUUID, _ = f.GetString("destination-uuid")
			if req.Project == "" || req.Environment == "" {
				return fmt.Errorf("--project and --environment are required")
			}
			if f.Changed("auto-apply") {
				autoApply, _ := f.GetBool("auto-apply")
				req.AutoApply = &autoApply
			}

			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			link, err := svc.CreateLink(cmd.Context(), req)
			if err != nil {
				return err
			}
			return printLink(cmd, link)
		},
	}
	f := cmd.Flags()
	f.String("project", "", "Project name (required)")
	f.String("environment", "", "Environment name (required)")
	f.String("mode", "api", "Link mode: api or git")
	f.Bool("auto-apply", false, "Apply safe changes automatically")
	f.String("repository", "", "Repository URL (git mode)")
	f.String("branch", "", "Branch (git mode)")
	f.String("file-path", "", "Path of the file in the repository (default coolify.json)")
	f.String("source-type", "", "Git source: public, github_app, gitea_app, or deploy_key (git mode)")
	f.String("source-uuid", "", "UUID of the GitHub app, Gitea app, or deploy key")
	f.String("server-uuid", "", "Server that reads the file and runs resources that name no server (git mode)")
	f.String("destination-uuid", "", "Destination on that server, when it has more than one")
	return cmd
}

func newLinkGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <link_uuid>",
		Short: "Show a coolify.json link",
		Args:  cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			link, err := svc.GetLink(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printLink(cmd, link)
		},
	}
}

func newLinkUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <link_uuid>",
		Short: "Change auto apply or the server of a coolify.json link",
		Long: `Change auto apply or the default server of a coolify.json link.

A server change plans the file again. Resources that move wait for approval.
It needs a token with the deploy ability.

Examples:
  coolify config link update <link_uuid> --auto-apply=false
  coolify config link update <link_uuid> --server-uuid <server_uuid>`,
		Args: cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := cmd.Flags()
			req := models.ConfigLinkUpdateRequest{}
			req.ServerUUID, _ = f.GetString("server-uuid")
			req.DestinationUUID, _ = f.GetString("destination-uuid")
			if f.Changed("auto-apply") {
				autoApply, _ := f.GetBool("auto-apply")
				req.AutoApply = &autoApply
			}
			if req.AutoApply == nil && req.ServerUUID == "" {
				return fmt.Errorf("set --auto-apply or --server-uuid")
			}
			if req.DestinationUUID != "" && req.ServerUUID == "" {
				return fmt.Errorf("--destination-uuid needs --server-uuid")
			}

			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			link, err := svc.UpdateLink(cmd.Context(), args[0], req)
			if err != nil {
				return err
			}
			return printLink(cmd, link)
		},
	}
	cmd.Flags().Bool("auto-apply", false, "Apply safe changes automatically")
	cmd.Flags().String("server-uuid", "", "New default server of the link")
	cmd.Flags().String("destination-uuid", "", "Destination on the new server, when it has more than one")
	return cmd
}

func newLinkDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link_uuid>",
		Short: "Disconnect an environment from coolify.json",
		Long: `Remove the link. The resources stay and can be edited in Coolify again.

Examples:
  coolify config link delete <link_uuid>
  coolify config link delete <link_uuid> --force`,
		Args: cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if force, _ := cmd.Flags().GetBool("force"); !force {
				ok, err := confirm(cmd, fmt.Sprintf("Disconnect config link %s?", args[0]))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "Disconnect cancelled.")
					return nil
				}
			}
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			if err := svc.DeleteLink(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Config link %s disconnected.\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}
