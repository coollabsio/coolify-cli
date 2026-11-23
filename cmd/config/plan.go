package config

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

func newPlanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <environment_uuid>",
		Short: "Compare a coolify.json file with an environment",
		Long: `Compare a coolify.json file with an environment and show the changes,
errors, warnings, and missing required values. Nothing changes.

The command exits with a non-zero code when the file has errors, so CI can use it.

Examples:
  coolify config plan <environment_uuid>
  coolify config plan <environment_uuid> -f path/to/coolify.json
  cat coolify.json | coolify config plan <environment_uuid> -f -`,
		Args: cli.ExactArgs(1, "<environment_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("file")
			file, err := readConfigFile(cmd, path)
			if err != nil {
				return err
			}
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			plan, err := svc.Plan(cmd.Context(), args[0], file)
			if err != nil {
				return err
			}
			if err := printPlan(cmd, plan); err != nil {
				return err
			}
			return planErr(plan)
		},
	}
	cmd.Flags().StringP("file", "f", defaultFile, "Path to the coolify.json file (- reads standard input)")
	return cmd
}
