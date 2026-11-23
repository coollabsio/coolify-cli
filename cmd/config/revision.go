package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
)

func newRevisionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revisions <link_uuid>",
		Short: "List the revisions of a coolify.json link, newest first",
		Args:  cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			revisions, err := svc.Revisions(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printData(cmd, revisions)
		},
	}
}

func newApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <link_uuid>",
		Short: "Plan and apply coolify.json",
		Long: `Plan and apply coolify.json.

For a link in api mode, give the file with -f. A safe plan applies at once;
a plan with destructive changes waits for approval.
For a link in git mode, do not give a file: Coolify plans the head of the branch.

The command exits with a non-zero code when the revision fails.

Examples:
  coolify config apply <link_uuid> -f coolify.json
  coolify config apply <git_link_uuid>`,
		Args: cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			var file *string
			if cmd.Flags().Changed("file") {
				path, _ := cmd.Flags().GetString("file")
				contents, err := readConfigFile(cmd, path)
				if err != nil {
					return err
				}
				file = &contents
			}
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			revision, err := svc.Apply(cmd.Context(), args[0], file)
			if err != nil {
				return err
			}
			return showRevision(cmd, revision)
		},
	}
	cmd.Flags().StringP("file", "f", defaultFile, "Path to the coolify.json file for an api link (- reads standard input)")
	return cmd
}

func newSuspendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "suspend <link_uuid>",
		Short: "Pause the sync of a coolify.json link",
		Long:  "Pause the sync of a coolify.json link. Paused resources can be edited in Coolify again. Only a team admin can pause.",
		Args:  cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			link, err := svc.Suspend(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printLink(cmd, link)
		},
	}
}

func newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <link_uuid>",
		Short: "Resume the sync of a coolify.json link",
		Long: `Resume the sync of a coolify.json link. Only a team admin can resume.

Without --confirm, show the plan of the latest file. Nothing changes.
With --confirm, resume and apply the file. This needs a token with the deploy ability.

Examples:
  coolify config resume <link_uuid>
  coolify config resume <link_uuid> --confirm`,
		Args: cli.ExactArgs(1, "<link_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			if ok, _ := cmd.Flags().GetBool("confirm"); !ok {
				plan, err := svc.ResumePreview(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				if err := printPlan(cmd, plan); err != nil {
					return err
				}
				if formatOf(cmd) == output.FormatTable {
					fmt.Fprintf(cmd.OutOrStdout(), "\nResume and apply with: coolify config resume %s --confirm\n", args[0])
				}
				return planErr(plan)
			}

			result, err := svc.Resume(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if formatOf(cmd) != output.FormatTable {
				if err := printData(cmd, result); err != nil {
					return err
				}
				return revisionErr(result.Revision)
			}
			if err := printLink(cmd, &result.Link); err != nil {
				return err
			}
			if result.Revision == nil {
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout())
			return showRevision(cmd, result.Revision)
		},
	}
	cmd.Flags().Bool("confirm", false, "Resume and apply the file")
	return cmd
}

func newApproveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <revision_uuid>",
		Short: "Approve a revision that waits for approval",
		Long: `Approve a revision that waits for approval, and apply it.

Removed resources keep their data, except the resources that you name with
--delete-data. Deleting data cannot be undone, so the command asks first
(use --force to skip the question).

Examples:
  coolify config approve <revision_uuid>
  coolify config approve <revision_uuid> --delete-data old-db --delete-data cache`,
		Args: cli.ExactArgs(1, "<revision_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteData, _ := cmd.Flags().GetStringSlice("delete-data")
			if force, _ := cmd.Flags().GetBool("force"); len(deleteData) > 0 && !force {
				ok, err := confirm(cmd, fmt.Sprintf("Delete the data of %s? This cannot be undone.", strings.Join(deleteData, ", ")))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(cmd.OutOrStdout(), "Approval cancelled.")
					return nil
				}
			}
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			revision, err := svc.Approve(cmd.Context(), args[0], deleteData)
			if err != nil {
				return err
			}
			return showRevision(cmd, revision)
		},
	}
	cmd.Flags().StringSlice("delete-data", nil, "Key of a removed resource whose data is deleted too (repeatable)")
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt for --delete-data")
	return cmd
}

func newReleaseCommand() *cobra.Command {
	return newRevisionActionCommand("release", "Release a stuck apply of a revision (team admin)",
		func(cmd *cobra.Command, uuid string) (*models.ConfigRevision, error) {
			svc, err := configService(cmd)
			if err != nil {
				return nil, err
			}
			return svc.Release(cmd.Context(), uuid)
		})
}

func newRetryCommand() *cobra.Command {
	return newRevisionActionCommand("retry", "Plan the file of a revision again",
		func(cmd *cobra.Command, uuid string) (*models.ConfigRevision, error) {
			svc, err := configService(cmd)
			if err != nil {
				return nil, err
			}
			return svc.Retry(cmd.Context(), uuid)
		})
}

func newRevisionActionCommand(name, short string, run func(*cobra.Command, string) (*models.ConfigRevision, error)) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <revision_uuid>",
		Short: short,
		Args:  cli.ExactArgs(1, "<revision_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			revision, err := run(cmd, args[0])
			if err != nil {
				return err
			}
			return showRevision(cmd, revision)
		},
	}
}

// showRevision prints a revision and fails when the revision failed.
func showRevision(cmd *cobra.Command, revision *models.ConfigRevision) error {
	if err := printRevision(cmd, revision); err != nil {
		return err
	}
	return revisionErr(revision)
}
